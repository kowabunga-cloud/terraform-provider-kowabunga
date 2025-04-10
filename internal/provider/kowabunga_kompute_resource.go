/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"maps"

	sdk "github.com/kowabunga-cloud/kowabunga-go"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	KomputeResourceName = "kompute"

	KomputeDefaultValuePool      = ""
	KomputeDefaultValueTemplate  = ""
	KomputeDefaultValueExtraDisk = 0
	KomputeDefaultValuePublic    = false
)

var _ resource.Resource = &KomputeResource{}
var _ resource.ResourceWithImportState = &KomputeResource{}

func NewKomputeResource() resource.Resource {
	return &KomputeResource{}
}

type KomputeResource struct {
	Data *KowabungaProviderData
}

type KomputeResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
	Name      types.String   `tfsdk:"name"`
	Desc      types.String   `tfsdk:"desc"`
	Project   types.String   `tfsdk:"project"`
	Zone      types.String   `tfsdk:"zone"`
	Pool      types.String   `tfsdk:"pool"`
	Template  types.String   `tfsdk:"template"`
	VCPUs     types.Int64    `tfsdk:"vcpus"`
	Memory    types.Int64    `tfsdk:"mem"`
	Disk      types.Int64    `tfsdk:"disk"`
	ExtraDisk types.Int64    `tfsdk:"extra_disk"`
	Public    types.Bool     `tfsdk:"public"`
	IP        types.String   `tfsdk:"ip"`
}

func (r *KomputeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, KomputeResourceName)
}

func (r *KomputeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root(KeyIP), req, resp)
}

func (r *KomputeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *KomputeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kompute virtual machine resource. **Kompute** is an seamless automated way to create virtual machine resources. It abstract the complexity of manually creating instance, volumes and network adapters resources and binding them together. It is the **RECOMMENDED** way to create and manipulate virtual machine services, unless a specific hwardware configuration is required. Kompute provides 2 network adapters, a public (WAN) and a private (LAN/VPC) one, as well as up to two disks (first one for OS, optional second one for extra data).",
		Attributes: map[string]schema.Attribute{
			KeyProject: schema.StringAttribute{
				MarkdownDescription: "Associated project name or ID",
				Required:            true,
			},
			KeyZone: schema.StringAttribute{
				MarkdownDescription: "Associated zone name or ID",
				Required:            true,
			},
			KeyPool: schema.StringAttribute{
				MarkdownDescription: "Associated storage pool name or ID (zone's default if unspecified)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KomputeDefaultValuePool),
			},
			KeyTemplate: schema.StringAttribute{
				MarkdownDescription: "Associated template name or ID (zone's default storage pool's default if unspecified)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KomputeDefaultValueTemplate),
			},
			KeyVCPUs: schema.Int64Attribute{
				MarkdownDescription: "The Kompute instance number of vCPUs",
				Required:            true,
			},
			KeyMemory: schema.Int64Attribute{
				MarkdownDescription: "The Kompute instance memory size (expressed in GB)",
				Required:            true,
			},
			KeyDisk: schema.Int64Attribute{
				MarkdownDescription: "The Kompute instance OS disk size (expressed in GB)",
				Required:            true,
			},
			KeyExtraDisk: schema.Int64Attribute{
				MarkdownDescription: "The Kompute optional data disk size (expressed in GB, disabled by default, 0 to disable)",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(KomputeDefaultValueExtraDisk),
			},
			KeyPublic: schema.BoolAttribute{
				MarkdownDescription: "Should Kompute instance be exposed over public Internet ? (default: **false**)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(KomputeDefaultValuePublic),
			},
			KeyIP: schema.StringAttribute{
				MarkdownDescription: "IP (read-only)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts kompute from Terraform model to Kowabunga API model
func komputeResourceToModel(d *KomputeResourceModel) sdk.Kompute {
	memSize := d.Memory.ValueInt64() * HelperGbToBytes
	diskSize := d.Disk.ValueInt64() * HelperGbToBytes
	extraDiskSize := d.ExtraDisk.ValueInt64() * HelperGbToBytes

	return sdk.Kompute{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		Vcpus:       d.VCPUs.ValueInt64(),
		Memory:      memSize,
		Disk:        diskSize,
		DataDisk:    &extraDiskSize,
		Ip:          d.IP.ValueStringPointer(),
	}
}

// converts kompute from Kowabunga API model to Terraform model
func komputeModelToResource(r *sdk.Kompute, d *KomputeResourceModel) {
	if r == nil {
		return
	}

	memSize := r.Memory / HelperGbToBytes
	diskSize := r.Disk / HelperGbToBytes
	var extraDiskSize int64 = 0
	if r.DataDisk != nil {
		extraDiskSize = *r.DataDisk / HelperGbToBytes
	}

	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	d.VCPUs = types.Int64Value(r.Vcpus)
	d.Memory = types.Int64Value(memSize)
	d.Disk = types.Int64Value(diskSize)
	d.ExtraDisk = types.Int64Value(extraDiskSize)
	if r.Ip != nil {
		d.IP = types.StringPointerValue(r.Ip)
	} else {
		d.IP = types.StringValue("")
	}
}

func (r *KomputeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *KomputeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Create(ctx, DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	// find parent project
	projectId, err := getProjectID(ctx, r.Data, data.Project.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// find parent zone
	zoneId, err := getZoneID(ctx, r.Data, data.Zone.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// find parent pool (optional)
	poolId, _ := getPoolID(ctx, r.Data, data.Pool.ValueString())

	// find parent template (optional)
	templateId, _ := getTemplateID(ctx, r.Data, data.Template.ValueString(), poolId)

	// create a new Kompute
	m := komputeResourceToModel(data)
	api := r.Data.K.ProjectAPI.CreateProjectZoneKompute(ctx, projectId, zoneId).Kompute(m).Public(data.Public.ValueBool())
	if poolId != "" {
		api = api.PoolId(poolId)
	}
	if templateId != "" {
		api = api.TemplateId(templateId)
	}
	kompute, _, err := api.Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(kompute.Id)
	komputeModelToResource(kompute, data) // read back resulting object
	tflog.Trace(ctx, "created Kompute resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KomputeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *KomputeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := data.Timeouts.Read(ctx, DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	kompute, _, err := r.Data.K.KomputeAPI.ReadKompute(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	komputeModelToResource(kompute, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KomputeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *KomputeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := data.Timeouts.Update(ctx, DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	m := komputeResourceToModel(data)
	_, _, err := r.Data.K.KomputeAPI.UpdateKompute(ctx, data.ID.ValueString()).Kompute(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KomputeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *KomputeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := data.Timeouts.Delete(ctx, DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	_, err := r.Data.K.KomputeAPI.DeleteKompute(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
