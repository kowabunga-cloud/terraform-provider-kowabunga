/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"maps"
	"sort"

	sdk "github.com/kowabunga-cloud/kowabunga-go"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	InstanceResourceName = "instance"
)

var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

type InstanceResource struct {
	Data *KowabungaProviderData
}

type InstanceResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
	Name     types.String   `tfsdk:"name"`
	Desc     types.String   `tfsdk:"desc"`
	Project  types.String   `tfsdk:"project"`
	Zone     types.String   `tfsdk:"zone"`
	VCPUs    types.Int64    `tfsdk:"vcpus"`
	Memory   types.Int64    `tfsdk:"mem"`
	Adapters types.List     `tfsdk:"adapters"`
	Volumes  types.List     `tfsdk:"volumes"`
}

func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, InstanceResourceName)
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a raw virtual machine instance resource. Usage of instance resource requires preliminary creation of network adapters and storage volumes to be associated with the instance. It comes handy when one wants to deploy a specifically tuned virtual machine's configuration. For common usage, it is recommended to use the **kce** resource instead, which provides a standard ready-to-be-used virtual machine, offloading much of the complexity.",
		Attributes: map[string]schema.Attribute{
			KeyProject: schema.StringAttribute{
				MarkdownDescription: "Associated project name or ID",
				Required:            true,
			},
			KeyZone: schema.StringAttribute{
				MarkdownDescription: "Associated zone name or ID",
				Required:            true,
			},
			KeyVCPUs: schema.Int64Attribute{
				MarkdownDescription: "The instance number of vCPUs",
				Required:            true,
			},
			KeyMemory: schema.Int64Attribute{
				MarkdownDescription: "The instance memory size (expressed in GB)",
				Required:            true,
			},
			KeyAdapters: schema.ListAttribute{
				MarkdownDescription: "The list of network adapters to be associated with the instance",
				ElementType:         types.StringType,
				Required:            true,
			},
			KeyVolumes: schema.ListAttribute{
				MarkdownDescription: "The list of storage volumes to be associated with the instance",
				ElementType:         types.StringType,
				Required:            true,
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts instance from Terraform model to Kowabunga API model
func instanceResourceToModel(d *InstanceResourceModel) sdk.Instance {
	memSize := d.Memory.ValueInt64() * HelperGbToBytes
	adapters := []string{}
	d.Adapters.ElementsAs(context.TODO(), &adapters, false)
	volumes := []string{}
	d.Volumes.ElementsAs(context.TODO(), &volumes, false)
	sort.Strings(volumes)

	return sdk.Instance{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		Vcpus:       d.VCPUs.ValueInt64(),
		Memory:      memSize,
		Adapters:    adapters,
		Volumes:     volumes,
	}
}

// converts instance from Kowabunga API model to Terraform model
func instanceModelToResource(r *sdk.Instance, d *InstanceResourceModel) {
	if r == nil {
		return
	}

	memSize := r.Memory / HelperGbToBytes
	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	d.VCPUs = types.Int64Value(r.Vcpus)
	d.Memory = types.Int64Value(memSize)
	adapters := []attr.Value{}
	for _, a := range r.Adapters {
		adapters = append(adapters, types.StringValue(a))
	}
	d.Adapters, _ = types.ListValue(types.StringType, adapters)
	volumes := []attr.Value{}
	sort.Strings(r.Volumes)
	for _, v := range r.Volumes {
		volumes = append(volumes, types.StringValue(v))
	}
	d.Volumes, _ = types.ListValue(types.StringType, volumes)
}

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *InstanceResourceModel
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
	// create a new instance
	m := instanceResourceToModel(data)
	instance, _, err := r.Data.K.ProjectAPI.CreateProjectZoneInstance(ctx, projectId, zoneId).Instance(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(instance.Id)
	instanceModelToResource(instance, data) // read back resulting object
	tflog.Trace(ctx, "created instance resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *InstanceResourceModel
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

	instance, _, err := r.Data.K.InstanceAPI.ReadInstance(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}
	instanceModelToResource(instance, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *InstanceResourceModel
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

	m := instanceResourceToModel(data)
	_, _, err := r.Data.K.InstanceAPI.UpdateInstance(ctx, data.ID.ValueString()).Instance(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *InstanceResourceModel
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

	_, err := r.Data.K.InstanceAPI.DeleteInstance(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
