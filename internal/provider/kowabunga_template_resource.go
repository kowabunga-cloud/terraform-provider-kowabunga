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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	TemplateResourceName = "template"

	TemplateDefaultValueOS      = "linux"
	TemplateDefaultValueDefault = false
)

var _ resource.Resource = &TemplateResource{}
var _ resource.ResourceWithImportState = &TemplateResource{}

func NewTemplateResource() resource.Resource {
	return &TemplateResource{}
}

type TemplateResource struct {
	Data *KowabungaProviderData
}

type TemplateResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
	Name     types.String   `tfsdk:"name"`
	Desc     types.String   `tfsdk:"desc"`
	Pool     types.String   `tfsdk:"pool"`
	OS       types.String   `tfsdk:"os"`
	Source   types.String   `tfsdk:"source"`
	Default  types.Bool     `tfsdk:"default"`
}

func (r *TemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, TemplateResourceName)
}

func (r *TemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *TemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *TemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a storage pool's template resource",
		Attributes: map[string]schema.Attribute{
			KeyPool: schema.StringAttribute{
				MarkdownDescription: "Associated storage pool name or ID",
				Required:            true,
			},
			KeyOS: schema.StringAttribute{
				MarkdownDescription: "The template type (valid options: 'linux', 'windows'). Defaults to **linux**.",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(TemplateDefaultValueOS),
			},
			KeySource: schema.StringAttribute{
				MarkdownDescription: "The template HTTP(S) source URL.",
				Required:            true,
			},
			KeyDefault: schema.BoolAttribute{
				MarkdownDescription: "Whether to set template as zone's default one (default: **false**). The first template to be created is always considered as default.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(TemplateDefaultValueDefault),
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts template from Terraform model to Kowabunga API model
func templateResourceToModel(d *TemplateResourceModel) sdk.Template {
	return sdk.Template{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		Os:          d.OS.ValueStringPointer(),
		Source:      d.Source.ValueString(),
	}
}

// converts template from Kowabunga API model to Terraform model
func templateModelToResource(r *sdk.Template, d *TemplateResourceModel) {
	if r == nil {
		return
	}

	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	if r.Os != nil {
		d.OS = types.StringPointerValue(r.Os)
	} else {
		d.OS = types.StringValue(TemplateDefaultValueOS)
	}
	d.Source = types.StringValue(r.Source)
}

func (r *TemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TemplateResourceModel
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

	// find parent pool
	poolId, err := getPoolID(ctx, r.Data, data.Pool.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// create a new template
	m := templateResourceToModel(data)
	template, _, err := r.Data.K.PoolAPI.CreateTemplate(ctx, poolId).Template(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// set template as default
	if data.Default.ValueBool() {
		_, err = r.Data.K.PoolAPI.SetStoragePoolDefaultTemplate(ctx, poolId, *template.Id).Execute()
		if err != nil {
			errorCreateGeneric(resp, err)
			return
		}
	}

	data.ID = types.StringPointerValue(template.Id)
	templateModelToResource(template, data) // read back resulting object
	tflog.Trace(ctx, "created template resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TemplateResourceModel
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

	template, _, err := r.Data.K.TemplateAPI.ReadTemplate(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	templateModelToResource(template, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TemplateResourceModel
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

	m := templateResourceToModel(data)
	_, _, err := r.Data.K.TemplateAPI.UpdateTemplate(ctx, data.ID.ValueString()).Template(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TemplateResourceModel
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

	_, err := r.Data.K.TemplateAPI.DeleteTemplate(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
