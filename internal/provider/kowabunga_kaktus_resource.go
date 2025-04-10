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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	KaktusResourceName = "kaktus"

	KaktusDefaultValueCurrent          = "EUR"
	KaktusDefaultValueCpuOverCommit    = 3
	KaktusDefaultValueMemoryOverCommit = 2
)

var _ resource.Resource = &KaktusResource{}
var _ resource.ResourceWithImportState = &KaktusResource{}

func NewKaktusResource() resource.Resource {
	return &KaktusResource{}
}

type KaktusResource struct {
	Data *KowabungaProviderData
}

type KaktusResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
	Name             types.String   `tfsdk:"name"`
	Desc             types.String   `tfsdk:"desc"`
	Zone             types.String   `tfsdk:"zone"`
	CpuPrice         types.Float64  `tfsdk:"cpu_price"`
	MemoryPrice      types.Float64  `tfsdk:"memory_price"`
	Currency         types.String   `tfsdk:"currency"`
	CpuOvercommit    types.Int64    `tfsdk:"cpu_overcommit"`
	MemoryOvercommit types.Int64    `tfsdk:"memory_overcommit"`
	Agents           types.List     `tfsdk:"agents"`
}

func (r *KaktusResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, KaktusResourceName)
}

func (r *KaktusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *KaktusResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *KaktusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a kaktus node resource",
		Attributes: map[string]schema.Attribute{
			KeyZone: schema.StringAttribute{
				MarkdownDescription: "Associated zone name or ID",
				Required:            true,
			},
			KeyCpuPrice: schema.Float64Attribute{
				MarkdownDescription: "Kaktus node monthly CPU price value (default: 0)",
				Computed:            true,
				Optional:            true,
				Default:             float64default.StaticFloat64(0),
			},
			KeyMemoryPrice: schema.Float64Attribute{
				MarkdownDescription: "Kaktus node monthly Memory price value (default: 0)",
				Computed:            true,
				Optional:            true,
				Default:             float64default.StaticFloat64(0),
			},
			KeyCurrency: schema.StringAttribute{
				MarkdownDescription: "Kaktus node monthly price currency (default: **EUR**)",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(KaktusDefaultValueCurrent),
			},
			KeyCpuOvercommit: schema.Int64Attribute{
				MarkdownDescription: "Kaktus node CPU over-commit factor (default: 3)",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(KaktusDefaultValueCpuOverCommit),
			},
			KeyMemoryOvercommit: schema.Int64Attribute{
				MarkdownDescription: "Kaktus node memory over-commit factor (default: 2)",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(KaktusDefaultValueMemoryOverCommit),
			},
			KeyAgents: schema.ListAttribute{
				MarkdownDescription: "The list of Kowabunga remote agents to be associated with the kaktus node",
				ElementType:         types.StringType,
				Required:            true,
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts kaktus from Terraform model to Kowabunga API model
func kaktusResourceToModel(d *KaktusResourceModel) sdk.Kaktus {
	agents := []string{}
	d.Agents.ElementsAs(context.TODO(), &agents, false)

	return sdk.Kaktus{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		CpuCost: &sdk.Cost{
			Price:    float32(d.CpuPrice.ValueFloat64()),
			Currency: d.Currency.ValueString(),
		},
		MemoryCost: &sdk.Cost{
			Price:    float32(d.MemoryPrice.ValueFloat64()),
			Currency: d.Currency.ValueString(),
		},
		OvercommitCpuRatio:    d.CpuOvercommit.ValueInt64Pointer(),
		OvercommitMemoryRatio: d.MemoryOvercommit.ValueInt64Pointer(),
		Agents:                agents,
	}
}

// converts kaktus from Kowabunga API model to Terraform model
func kaktusModelToResource(r *sdk.Kaktus, d *KaktusResourceModel) {
	if r == nil {
		return
	}

	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	d.CpuPrice = types.Float64Value(float64(r.CpuCost.Price))
	d.Currency = types.StringValue(r.CpuCost.Currency)
	d.MemoryPrice = types.Float64Value(float64(r.MemoryCost.Price))
	if r.OvercommitCpuRatio != nil {
		d.CpuOvercommit = types.Int64PointerValue(r.OvercommitCpuRatio)
	} else {
		d.CpuOvercommit = types.Int64Value(KaktusDefaultValueCpuOverCommit)
	}
	if r.OvercommitMemoryRatio != nil {
		d.MemoryOvercommit = types.Int64PointerValue(r.OvercommitMemoryRatio)
	} else {
		d.MemoryOvercommit = types.Int64Value(KaktusDefaultValueMemoryOverCommit)
	}
	agents := []attr.Value{}
	for _, a := range r.Agents {
		agents = append(agents, types.StringValue(a))
	}
	d.Agents, _ = types.ListValue(types.StringType, agents)
}

func (r *KaktusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *KaktusResourceModel
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

	// find parent zone
	zoneId, err := getZoneID(ctx, r.Data, data.Zone.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// create a new kaktus
	m := kaktusResourceToModel(data)
	kaktus, _, err := r.Data.K.ZoneAPI.CreateKaktus(ctx, zoneId).Kaktus(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(kaktus.Id)
	kaktusModelToResource(kaktus, data) // read back resulting object
	tflog.Trace(ctx, "created kaktus resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaktusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *KaktusResourceModel
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

	kaktus, _, err := r.Data.K.KaktusAPI.ReadKaktus(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	kaktusModelToResource(kaktus, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaktusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *KaktusResourceModel
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

	m := kaktusResourceToModel(data)
	_, _, err := r.Data.K.KaktusAPI.UpdateKaktus(ctx, data.ID.ValueString()).Kaktus(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KaktusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *KaktusResourceModel
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

	_, err := r.Data.K.KaktusAPI.DeleteKaktus(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
