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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	ProjectResourceName = "project"

	ProjectDefaultValueDomain       = ""
	ProjecDefaultValueSubnetSize    = 26
	ProjectDefaultValueRootPassword = ""
	ProjectDefaultValueMaxInstances = 0
	ProjectDefaultValueMaxMemory    = 0
	ProjectDefaultValueMaxStorage   = 0
	ProjectDefaultValueMaxVCPUs     = 0
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	Data *KowabungaProviderData
}

type ProjectResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
	Name           types.String   `tfsdk:"name"`
	Desc           types.String   `tfsdk:"desc"`
	Domain         types.String   `tfsdk:"domain"`
	SubnetSize     types.Int64    `tfsdk:"subnet_size"`
	RootPassword   types.String   `tfsdk:"root_password"`
	User           types.String   `tfsdk:"bootstrap_user"`
	Pubkey         types.String   `tfsdk:"bootstrap_pubkey"`
	Tags           types.List     `tfsdk:"tags"`
	Metadatas      types.Map      `tfsdk:"metadata"`
	MaxInstances   types.Int64    `tfsdk:"max_instances"`
	MaxMemory      types.Int64    `tfsdk:"max_memory"`
	MaxStorage     types.Int64    `tfsdk:"max_storage"`
	MaxVCPUs       types.Int64    `tfsdk:"max_vcpus"`
	PrivateSubnets types.Map      `tfsdk:"private_subnets"`
	Teams          types.List     `tfsdk:"teams"`
	Regions        types.List     `tfsdk:"regions"`
	VRIDs          types.List     `tfsdk:"vrids"`
}

type ProjectQuotaModel struct {
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, ProjectResourceName)
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root(KeyPrivateSubnets), req, resp)
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a project resource",
		Attributes: map[string]schema.Attribute{
			KeyDomain: schema.StringAttribute{
				MarkdownDescription: "Internal domain name associated to the project (e.g. myproject.acme.com). (default: none)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(ProjectDefaultValueDomain),
			},
			KeySubnetSize: schema.Int64Attribute{
				MarkdownDescription: "Project requested VPC subnet size (defaults to /26)",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(ProjecDefaultValueSubnetSize),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			KeyRootPassword: schema.StringAttribute{
				MarkdownDescription: "The project default root password, set at cloud-init instance bootstrap phase. Will be randomly auto-generated at each instance creation if unspecified.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(ProjectDefaultValueRootPassword),
			},
			KeyBootstrapUser: schema.StringAttribute{
				MarkdownDescription: "The project default service user name, created at cloud-init instance bootstrap phase. Will use Kowabunga's default configuration one if unspecified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			KeyBootstrapPubkey: schema.StringAttribute{
				MarkdownDescription: "The project default public SSH key, to be associated to bootstrap user. Will use Kowabunga's default configuration one if unspecified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			KeyTags: schema.ListAttribute{
				MarkdownDescription: "List of tags associated with the project",
				ElementType:         types.StringType,
				Required:            true,
			},
			KeyMetadata: schema.MapAttribute{
				MarkdownDescription: "List of metadatas key/value associated with the project",
				ElementType:         types.StringType,
				Required:            true,
			},
			KeyMaxInstances: schema.Int64Attribute{
				MarkdownDescription: "Project maximum deployable instances. Defaults to 0 (unlimited).",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(ProjectDefaultValueMaxInstances),
			},
			KeyMaxMemory: schema.Int64Attribute{
				MarkdownDescription: "Project maximum usable memory (expressed in GB). Defaults to 0 (unlimited).",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(ProjectDefaultValueMaxMemory),
			},
			KeyMaxStorage: schema.Int64Attribute{
				MarkdownDescription: "Project maximum usable storage (expressed in GB). Defaults to 0 (unlimited).",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(ProjectDefaultValueMaxStorage),
			},
			KeyMaxVCPUs: schema.Int64Attribute{
				MarkdownDescription: "Project maximum usable virtual CPUs. Defaults to 0 (unlimited).",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(ProjectDefaultValueMaxVCPUs),
			},
			KeyPrivateSubnets: schema.MapAttribute{
				Computed:            true,
				MarkdownDescription: "List of project's private subnets zones association (read-only)",
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			KeyTeams: schema.ListAttribute{
				MarkdownDescription: "The list of user teams allowed to administrate the project (i.e. capable of managing internal resources)",
				ElementType:         types.StringType,
				Required:            true,
			},
			KeyRegions: schema.ListAttribute{
				MarkdownDescription: "The list of regions the project is managing resources from (subnets will be pre-allocated in all referenced regions)",
				ElementType:         types.StringType,
				Required:            true,
			},
			KeyVRIDs: schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "List of VRRP IDs used by -as-a-service resources within the project virtual network (read-only). Should your application use VRRP for service redundancy, you should use different IDs to prevent issues.",
				ElementType:         types.Int64Type,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts project from Terraform model to Kowabunga API model
func projectResourceToModel(d *ProjectResourceModel) sdk.Project {
	tags := []string{}
	d.Tags.ElementsAs(context.TODO(), &tags, false)

	metas := map[string]string{}
	d.Metadatas.ElementsAs(context.TODO(), &metas, false)
	metadatas := []sdk.Metadata{}
	for k, v := range metas {
		m := sdk.Metadata{
			Key:   k,
			Value: v,
		}
		metadatas = append(metadatas, m)
	}

	instances := int32(d.MaxInstances.ValueInt64())
	memory := int64(d.MaxMemory.ValueInt64()) * HelperGbToBytes
	storage := int64(d.MaxStorage.ValueInt64()) * HelperGbToBytes
	vcpus := int32(d.MaxVCPUs.ValueInt64())
	quotas := &sdk.ProjectResources{
		Instances: &instances,
		Memory:    &memory,
		Storage:   &storage,
		Vcpus:     &vcpus,
	}

	teams := []string{}
	d.Teams.ElementsAs(context.TODO(), &teams, false)
	sort.Strings(teams)

	regions := []string{}
	d.Regions.ElementsAs(context.TODO(), &regions, false)
	sort.Strings(regions)

	return sdk.Project{
		Name:            d.Name.ValueString(),
		Description:     d.Desc.ValueStringPointer(),
		Domain:          d.Domain.ValueStringPointer(),
		RootPassword:    d.RootPassword.ValueStringPointer(),
		BootstrapUser:   d.User.ValueStringPointer(),
		BootstrapPubkey: d.Pubkey.ValueStringPointer(),
		Tags:            tags,
		Metadatas:       metadatas,
		Quotas:          quotas,
		Teams:           teams,
		Regions:         regions,
	}
}

// converts project from Kowabunga API model to Terraform model
func projectModelToResource(r *sdk.Project, d *ProjectResourceModel) {
	if r == nil {
		return
	}

	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	if r.Domain != nil {
		d.Domain = types.StringPointerValue(r.Domain)
	} else {
		d.Domain = types.StringValue(ProjectDefaultValueDomain)
	}
	if r.RootPassword != nil {
		d.RootPassword = types.StringPointerValue(r.RootPassword)
	} else {
		d.RootPassword = types.StringValue(ProjectDefaultValueRootPassword)
	}
	if r.BootstrapUser != nil {
		d.User = types.StringPointerValue(r.BootstrapUser)
	} else {
		d.User = types.StringValue("")
	}
	if r.BootstrapPubkey != nil {
		d.Pubkey = types.StringPointerValue(r.BootstrapPubkey)
	} else {
		d.Pubkey = types.StringValue("")
	}

	tags := []attr.Value{}
	for _, t := range r.Tags {
		tags = append(tags, types.StringValue(t))
	}
	d.Tags, _ = types.ListValue(types.StringType, tags)

	metadatas := map[string]attr.Value{}
	for _, m := range r.Metadatas {
		metadatas[m.Key] = types.StringValue(m.Value)
	}
	d.Metadatas = basetypes.NewMapValueMust(types.StringType, metadatas)

	if r.Quotas.Instances != nil {
		d.MaxInstances = types.Int64Value(int64(*r.Quotas.Instances))
	} else {
		d.MaxInstances = types.Int64Value(ProjectDefaultValueMaxInstances)
	}
	if r.Quotas.Memory != nil {
		d.MaxMemory = types.Int64Value(int64(*r.Quotas.Memory) / HelperGbToBytes)
	} else {
		d.MaxMemory = types.Int64Value(ProjectDefaultValueMaxMemory)
	}
	if r.Quotas.Storage != nil {
		d.MaxStorage = types.Int64Value(int64(*r.Quotas.Storage) / HelperGbToBytes)
	} else {
		d.MaxStorage = types.Int64Value(ProjectDefaultValueMaxStorage)
	}
	if r.Quotas.Vcpus != nil {
		d.MaxVCPUs = types.Int64Value(int64(*r.Quotas.Vcpus))
	} else {
		d.MaxVCPUs = types.Int64Value(ProjectDefaultValueMaxVCPUs)
	}

	privateSubnets := map[string]attr.Value{}
	for _, p := range r.PrivateSubnets {
		if p.Value != nil {
			privateSubnets[*p.Key] = types.StringPointerValue(p.Value)
		} else {
			privateSubnets[*p.Key] = types.StringValue("")
		}
	}
	d.PrivateSubnets = basetypes.NewMapValueMust(types.StringType, privateSubnets)

	teams := []attr.Value{}
	sort.Strings(r.Teams)
	for _, g := range r.Teams {
		teams = append(teams, types.StringValue(g))
	}
	d.Teams, _ = types.ListValue(types.StringType, teams)

	regions := []attr.Value{}
	sort.Strings(r.Regions)
	for _, r := range r.Regions {
		regions = append(regions, types.StringValue(r))
	}
	d.Regions, _ = types.ListValue(types.StringType, regions)

	reserved_vrids := []int{}
	for _, vrid := range r.ReservedVrrpIds {
		reserved_vrids = append(reserved_vrids, int(vrid))
	}
	sort.Ints(reserved_vrids)
	vrids := []attr.Value{}
	for _, vrid := range reserved_vrids {
		vrids = append(vrids, types.Int64Value(int64(vrid)))
	}
	d.VRIDs, _ = types.ListValue(types.Int64Type, vrids)
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProjectResourceModel
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

	// create a new project
	m := projectResourceToModel(data)
	project, _, err := r.Data.K.ProjectAPI.CreateProject(ctx).Project(m).SubnetSize(int32(data.SubnetSize.ValueInt64())).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(project.Id)
	projectModelToResource(project, data) // read back resulting object

	tflog.Trace(ctx, "created project resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ProjectResourceModel
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

	project, _, err := r.Data.K.ProjectAPI.ReadProject(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	projectModelToResource(project, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ProjectResourceModel
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

	m := projectResourceToModel(data)
	_, _, err := r.Data.K.ProjectAPI.UpdateProject(ctx, data.ID.ValueString()).Project(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProjectResourceModel
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

	_, err := r.Data.K.ProjectAPI.DeleteProject(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
