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
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	VNetResourceName = "vnet"

	VNetDefaultValueVlan    = 0
	VNetDefaultValuePrivate = true
)

var _ resource.Resource = &VNetResource{}
var _ resource.ResourceWithImportState = &VNetResource{}

func NewVNetResource() resource.Resource {
	return &VNetResource{}
}

type VNetResource struct {
	Data *KowabungaProviderData
}

type VNetResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
	Name      types.String   `tfsdk:"name"`
	Desc      types.String   `tfsdk:"desc"`
	Region    types.String   `tfsdk:"region"`
	VLAN      types.Int64    `tfsdk:"vlan"`
	Interface types.String   `tfsdk:"interface"`
	Private   types.Bool     `tfsdk:"private"`
}

func (r *VNetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, VNetResourceName)
}

func (r *VNetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *VNetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *VNetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a virtual network resource",
		Attributes: map[string]schema.Attribute{
			KeyRegion: schema.StringAttribute{
				MarkdownDescription: "Associated region name or ID",
				Required:            true,
			},
			KeyVLAN: schema.Int64Attribute{
				MarkdownDescription: "VLAN ID",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
					int64validator.AtMost(4095),
				},
			},
			KeyInterface: schema.StringAttribute{
				MarkdownDescription: "Host bridge network interface",
				Required:            true,
			},
			KeyPrivate: schema.BoolAttribute{
				MarkdownDescription: "Whether the virtual network is private or public (default: **true**, i.e. private). The first virtual network to be created is always considered to be the default one.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(VNetDefaultValuePrivate),
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts virtual network from Terraform model to Kowabunga API model
func vnetResourceToModel(d *VNetResourceModel) sdk.VNet {
	return sdk.VNet{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		Vlan:        d.VLAN.ValueInt64Pointer(),
		Interface:   d.Interface.ValueString(),
		Private:     d.Private.ValueBoolPointer(),
	}
}

// converts virtual network from Kowabunga API model to Terraform model
func vnetModelToResource(r *sdk.VNet, d *VNetResourceModel) {
	if r == nil {
		return
	}

	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	if r.Vlan != nil {
		d.VLAN = types.Int64PointerValue(r.Vlan)
	} else {
		d.VLAN = types.Int64Value(VNetDefaultValueVlan)
	}
	d.Interface = types.StringValue(r.Interface)
	if r.Private != nil {
		d.Private = types.BoolPointerValue(r.Private)
	} else {
		d.Private = types.BoolValue(VNetDefaultValuePrivate)
	}
}

func (r *VNetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *VNetResourceModel
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

	// find parent region
	regionId, err := getRegionID(ctx, r.Data, data.Region.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// create a new virtual network
	m := vnetResourceToModel(data)
	vnet, _, err := r.Data.K.RegionAPI.CreateVNet(ctx, regionId).VNet(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(vnet.Id)
	vnetModelToResource(vnet, data) // read back resulting object
	tflog.Trace(ctx, "created vnet resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VNetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *VNetResourceModel
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

	vnet, _, err := r.Data.K.VnetAPI.ReadVNet(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	vnetModelToResource(vnet, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VNetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *VNetResourceModel
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

	m := vnetResourceToModel(data)
	_, _, err := r.Data.K.VnetAPI.UpdateVNet(ctx, data.ID.ValueString()).VNet(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VNetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *VNetResourceModel
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

	_, err := r.Data.K.VnetAPI.DeleteVNet(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
