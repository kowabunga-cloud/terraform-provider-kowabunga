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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	KonveyResourceName = "konvey"

	KonveyDefaultValueFailover = true
	KonveyDefaultValueProtocol = "tcp"
)

var _ resource.Resource = &KonveyResource{}
var _ resource.ResourceWithImportState = &KonveyResource{}

func NewKonveyResource() resource.Resource {
	return &KonveyResource{}
}

type KonveyResource struct {
	Data *KowabungaProviderData
}

type KonveyResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
	Name      types.String   `tfsdk:"name"`
	Desc      types.String   `tfsdk:"desc"`
	Project   types.String   `tfsdk:"project"`
	Region    types.String   `tfsdk:"region"`
	PrivateIP types.String   `tfsdk:"private_ip"` // read-only
	Failover  types.Bool     `tfsdk:"failover"`
	Endpoints types.List     `tfsdk:"endpoints"` // []KonveyEndpoint
}

type KonveyEndpoint struct {
	Name        types.String `tfsdk:"name"`
	Protocol    types.String `tfsdk:"protocol"`
	Port        types.Int64  `tfsdk:"port"`
	BackendPort types.Int64  `tfsdk:"backend_port"`
	BackendIPs  types.List   `tfsdk:"backend_ips"` // []string
}

func (r *KonveyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, KonveyResourceName)
}

func (r *KonveyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *KonveyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *KonveyResource) SchemaEndpoints() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Konvey list of load-balanced endpoints.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				KeyName: schema.StringAttribute{
					MarkdownDescription: "Konvey endpoint name",
					Required:            true,
				},
				KeyProtocol: schema.StringAttribute{
					MarkdownDescription: "The endpoint's transport layer protocol to be exposed (defaults to 'tcp').",
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(KonveyDefaultValueProtocol),
					Validators: []validator.String{
						&stringNetworkProtocolValidator{},
					},
				},
				KeyPort: schema.Int64Attribute{
					MarkdownDescription: "The endpoint's port to be exposed.",
					Required:            true,
					Validators: []validator.Int64{
						&intNetworkPortValidator{},
					},
				},
				KeyBackendPort: schema.Int64Attribute{
					MarkdownDescription: "The endpoint's backend service port.",
					Required:            true,
					Validators: []validator.Int64{
						&intNetworkPortValidator{},
					},
				},
				KeyBackendIPs: schema.ListAttribute{
					MarkdownDescription: "The endpoint's list of load-balanced backend hosts.",
					Required:            true,
					ElementType:         types.StringType,
				},
			},
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	}
}

func (r *KonveyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Konvey resource. **Konvey** provides network load-balancer capabilities for a given project.",
		Attributes: map[string]schema.Attribute{
			KeyProject: schema.StringAttribute{
				MarkdownDescription: "Associated project name or ID",
				Required:            true,
			},
			KeyRegion: schema.StringAttribute{
				MarkdownDescription: "Associated region name or ID",
				Required:            true,
			},
			KeyPrivateIP: schema.StringAttribute{
				MarkdownDescription: "Konvey assigned private virtual IP address (read-only).",
				Required:            false,
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			KeyFailover: schema.BoolAttribute{
				MarkdownDescription: "Whether Konvey must be deployed in a highly-available replicated state to support service failover.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(KonveyDefaultValueFailover),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			KeyEndpoints: r.SchemaEndpoints(),
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

//////////////////////////////////////////////////////////////
// converts konvey from Terraform model to Kowabunga API model //
//////////////////////////////////////////////////////////////

func konveyEndpointsModel(ctx *context.Context, d *KonveyResourceModel) []sdk.KonveyEndpoint {
	epModel := []sdk.KonveyEndpoint{}

	endpoints := make([]types.Object, 0, len(d.Endpoints.Elements()))
	diags := d.Endpoints.ElementsAs(*ctx, &endpoints, false)
	if diags.HasError() {
		for _, err := range diags.Errors() {
			tflog.Debug(*ctx, err.Detail())
		}
	}

	for _, e := range endpoints {
		ep := KonveyEndpoint{}
		diags := e.As(*ctx, &ep, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			for _, err := range diags.Errors() {
				tflog.Error(*ctx, err.Detail())
			}
		}

		// backends
		hosts := make([]string, 0, len(ep.BackendIPs.Elements()))
		diags = ep.BackendIPs.ElementsAs(*ctx, &hosts, false)
		if diags.HasError() {
			for _, err := range diags.Errors() {
				tflog.Debug(*ctx, err.Detail())
			}
		}

		backendsModel := sdk.KonveyBackends{
			Hosts: hosts,
			Port:  ep.BackendPort.ValueInt64(),
		}

		epModel = append(epModel, sdk.KonveyEndpoint{
			Name:     ep.Name.ValueString(),
			Port:     ep.Port.ValueInt64(),
			Protocol: ep.Protocol.ValueString(),
			Backends: backendsModel,
		})
	}

	return epModel
}

func konveyResourceToModel(ctx *context.Context, d *KonveyResourceModel) sdk.Konvey {
	return sdk.Konvey{
		Name:        d.Name.ValueStringPointer(),
		Description: d.Desc.ValueStringPointer(),
		Failover:    d.Failover.ValueBoolPointer(),
		Endpoints:   konveyEndpointsModel(ctx, d),
	}
}

//////////////////////////////////////////////////////////////
// converts konvey from Kowabunga API model to Terraform model //
//////////////////////////////////////////////////////////////

func konveyModelToEndpoints(ctx *context.Context, r *sdk.Konvey, d *KonveyResourceModel) {
	endpoints := []attr.Value{}
	endpointsType := map[string]attr.Type{
		KeyName:        types.StringType,
		KeyProtocol:    types.StringType,
		KeyPort:        types.Int64Type,
		KeyBackendPort: types.Int64Type,
		KeyBackendIPs: types.ListType{
			ElemType: types.StringType,
		},
	}

	// empty endpoints ?
	if len(r.Endpoints) == 0 {
		d.Endpoints = types.ListNull(types.ObjectType{AttrTypes: endpointsType})
		return
	}

	for _, ep := range r.Endpoints {
		r := map[string]attr.Value{
			KeyName:        types.StringValue(ep.Name),
			KeyProtocol:    types.StringValue(ep.Protocol),
			KeyPort:        types.Int64Value(ep.Port),
			KeyBackendPort: types.Int64Value(ep.Backends.Port),
		}

		hosts := []attr.Value{}
		for _, h := range ep.Backends.Hosts {
			hosts = append(hosts, types.StringValue(h))
		}
		r[KeyBackendIPs], _ = types.ListValue(types.StringType, hosts)

		object, _ := types.ObjectValue(endpointsType, r)
		endpoints = append(endpoints, object)
	}

	d.Endpoints, _ = types.ListValue(types.ObjectType{AttrTypes: endpointsType}, endpoints)
}

func konveyModelToResource(ctx *context.Context, r *sdk.Konvey, d *KonveyResourceModel) {
	if r == nil {
		return
	}

	if r.Name != nil {
		d.Name = types.StringPointerValue(r.Name)
	} else {
		d.Name = types.StringValue("")
	}

	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}

	if r.Vip != nil {
		d.PrivateIP = types.StringPointerValue(r.Vip)
	} else {
		d.PrivateIP = types.StringValue("")
	}

	if r.Failover != nil {
		d.Failover = types.BoolPointerValue(r.Failover)
	} else {
		d.Failover = types.BoolValue(KonveyDefaultValueFailover)
	}

	konveyModelToEndpoints(ctx, r, d)
}

func (r *KonveyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *KonveyResourceModel

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
	// find parent region
	regionId, err := getRegionID(ctx, r.Data, data.Region.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	m := konveyResourceToModel(&ctx, data)

	// create a new Konvey
	konvey, _, err := r.Data.K.ProjectAPI.CreateProjectRegionKonvey(ctx, projectId, regionId).Konvey(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(konvey.Id)
	konveyModelToResource(&ctx, konvey, data) // read back resulting object
	tflog.Trace(ctx, "created Konvey resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KonveyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *KonveyResourceModel
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

	konvey, _, err := r.Data.K.KonveyAPI.ReadKonvey(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	konveyModelToResource(&ctx, konvey, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KonveyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *KonveyResourceModel
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

	m := konveyResourceToModel(&ctx, data)
	_, _, err := r.Data.K.KonveyAPI.UpdateKonvey(ctx, data.ID.ValueString()).Konvey(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KonveyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *KonveyResourceModel
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

	_, err := r.Data.K.KonveyAPI.DeleteKonvey(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
