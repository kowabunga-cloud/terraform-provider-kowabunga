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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	RegionResourceName = "region"
)

var _ resource.Resource = &RegionResource{}
var _ resource.ResourceWithImportState = &RegionResource{}

func NewRegionResource() resource.Resource {
	return &RegionResource{}
}

type RegionResource struct {
	Data *KowabungaProviderData
}

type RegionResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
	Name     types.String   `tfsdk:"name"`
	Desc     types.String   `tfsdk:"desc"`
	Domain   types.String   `tfsdk:"domain"`
}

func (r *RegionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, RegionResourceName)
}

func (r *RegionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *RegionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *RegionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a region resource",
		Attributes: map[string]schema.Attribute{
			KeyDomain: schema.StringAttribute{
				MarkdownDescription: "Region domain name (e.g. myregion.kowabunga.acme.com).",
				Required:            true,
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts region from Terraform model to Kowabunga API model
func regionResourceToModel(d *RegionResourceModel) sdk.Region {
	return sdk.Region{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		Domain:      d.Domain.ValueString(),
	}
}

// converts region from Kowabunga API model to Terraform model
func regionModelToResource(r *sdk.Region, d *RegionResourceModel) {
	if r == nil {
		return
	}

	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	d.Domain = types.StringValue(r.Domain)
}

func (r *RegionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *RegionResourceModel
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

	m := regionResourceToModel(data)
	region, _, err := r.Data.K.RegionAPI.CreateRegion(ctx).Region(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(region.Id)
	regionModelToResource(region, data) // read back resulting object
	tflog.Trace(ctx, "created region resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RegionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *RegionResourceModel
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

	region, _, err := r.Data.K.RegionAPI.ReadRegion(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	regionModelToResource(region, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RegionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *RegionResourceModel
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

	m := regionResourceToModel(data)
	_, _, err := r.Data.K.RegionAPI.UpdateRegion(ctx, data.ID.ValueString()).Region(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RegionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *RegionResourceModel
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

	_, err := r.Data.K.RegionAPI.DeleteRegion(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
