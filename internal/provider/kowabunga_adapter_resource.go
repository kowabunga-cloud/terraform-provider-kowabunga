/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"
	"maps"

	"github.com/3th1nk/cidr"

	sdk "github.com/kowabunga-cloud/kowabunga-go"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	AdapterResourceName = "adapter"

	AdapterDefaultValueAssign   = true
	AdapterDefaultValueReserved = false
)

var _ resource.Resource = &AdapterResource{}
var _ resource.ResourceWithImportState = &AdapterResource{}

func NewAdapterResource() resource.Resource {
	return &AdapterResource{}
}

type AdapterResource struct {
	Data *KowabungaProviderData
}

type AdapterResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
	Name           types.String   `tfsdk:"name"`
	Desc           types.String   `tfsdk:"desc"`
	Subnet         types.String   `tfsdk:"subnet"`
	MAC            types.String   `tfsdk:"hwaddress"`
	Addresses      types.List     `tfsdk:"addresses"`
	Assign         types.Bool     `tfsdk:"assign"`
	Reserved       types.Bool     `tfsdk:"reserved"`
	CIDR           types.String   `tfsdk:"cidr"`
	Netmask        types.String   `tfsdk:"netmask"`
	NetmaskBitSize types.Int64    `tfsdk:"netmask_bitsize"`
	Gateway        types.String   `tfsdk:"gateway"`
}

func (r *AdapterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, AdapterResourceName)
}

func (r *AdapterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root(KeyCIDR), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root(KeyNetmask), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root(KeyNetmaskBitSize), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root(KeyGateway), req, resp)
}

func (r *AdapterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *AdapterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a network adapter resource",
		Attributes: map[string]schema.Attribute{
			KeySubnet: schema.StringAttribute{
				MarkdownDescription: "Associated subnet name or ID",
				Required:            true,
			},
			KeyMAC: schema.StringAttribute{
				MarkdownDescription: "Network adapter hardware MAC address (e.g. 00:11:22:33:44:55). AUto-generated if unspecified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			KeyAddresses: schema.ListAttribute{
				MarkdownDescription: "Network adapter list of associated IPv4 addresses",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			KeyAssign: schema.BoolAttribute{
				MarkdownDescription: "Whether an IP address should be automatically assigned to the adapter (default: **true). Useless if addresses have been specified",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(AdapterDefaultValueAssign),
			},

			KeyReserved: schema.BoolAttribute{
				MarkdownDescription: "Whether the network adapter is reserved (e.g. router), i.e. where the same hardware address can be reused over several subnets (default: **false**)",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(AdapterDefaultValueReserved),
			},
			KeyCIDR: schema.StringAttribute{
				MarkdownDescription: "Network mask CIDR (read-only), e.g. 192.168.0.0/24",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			KeyNetmask: schema.StringAttribute{
				MarkdownDescription: "Network mask (read-only), e.g. 255.255.255.0",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			KeyNetmaskBitSize: schema.Int64Attribute{
				MarkdownDescription: "Network mask size (read-only), e.g 24",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			KeyGateway: schema.StringAttribute{
				MarkdownDescription: "Network Gateway (read-only)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts adapter from Terraform model to Kowabunga API model
func adapterResourceToModel(d *AdapterResourceModel) sdk.Adapter {
	addresses := []string{}
	d.Addresses.ElementsAs(context.TODO(), &addresses, false)
	return sdk.Adapter{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		Mac:         d.MAC.ValueStringPointer(),
		Addresses:   addresses,
		Reserved:    d.Reserved.ValueBoolPointer(),
	}
}

// converts adapter from Kowabunga API model to Terraform model
func adapterModelToResource(r *sdk.Adapter, d *AdapterResourceModel) {
	if r == nil {
		return
	}

	d.Name = types.StringValue(r.Name)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	if r.Mac != nil {
		d.MAC = types.StringPointerValue(r.Mac)
	} else {
		d.MAC = types.StringValue("")
	}
	addresses := []attr.Value{}
	for _, a := range r.Addresses {
		addresses = append(addresses, types.StringValue(a))
	}
	d.Addresses, _ = types.ListValue(types.StringType, addresses)
	if r.Reserved != nil {
		d.Reserved = types.BoolPointerValue(r.Reserved)
	} else {
		d.Reserved = types.BoolValue(AdapterDefaultValueReserved)
	}
}

func ipv4MaskString(m []byte) string {
	if len(m) != 4 {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
}

func (r *AdapterResource) GetSubnetData(ctx context.Context, data *AdapterResourceModel) error {
	// find real subnet id if a string was provided
	subnetId, err := getSubnetID(ctx, r.Data, data.Subnet.ValueString())
	if err != nil {
		return err
	}

	subnet, _, err := r.Data.K.SubnetAPI.ReadSubnet(ctx, subnetId).Execute()
	if err != nil {
		return err
	}

	data.CIDR = types.StringValue(subnet.Cidr)

	c, err := cidr.Parse(subnet.Cidr)
	if err != nil {
		return err
	}
	data.Netmask = types.StringValue(ipv4MaskString(c.Mask()))
	size, _ := c.Mask().Size()
	data.NetmaskBitSize = types.Int64Value(int64(size))
	data.Gateway = types.StringValue(subnet.Gateway)

	return nil
}

func (r *AdapterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AdapterResourceModel
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

	// find parent subnet
	subnetId, err := getSubnetID(ctx, r.Data, data.Subnet.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}

	// create a new adapter
	m := adapterResourceToModel(data)
	api := r.Data.K.SubnetAPI.CreateAdapter(ctx, subnetId).Adapter(m)
	if data.Assign.ValueBool() && len(m.Addresses) == 0 {
		api = api.AssignIP(data.Assign.ValueBool())
	}

	adapter, _, err := api.Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}

	data.ID = types.StringPointerValue(adapter.Id)
	adapterModelToResource(adapter, data) // read back resulting object
	err = r.GetSubnetData(ctx, data)
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}

	tflog.Trace(ctx, "created adapter resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdapterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AdapterResourceModel
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

	adapter, _, err := r.Data.K.AdapterAPI.ReadAdapter(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}
	adapterModelToResource(adapter, data)

	err = r.GetSubnetData(ctx, data)
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdapterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AdapterResourceModel
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

	m := adapterResourceToModel(data)
	_, _, err := r.Data.K.AdapterAPI.UpdateAdapter(ctx, data.ID.ValueString()).Adapter(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AdapterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AdapterResourceModel
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

	_, err := r.Data.K.AdapterAPI.DeleteAdapter(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
