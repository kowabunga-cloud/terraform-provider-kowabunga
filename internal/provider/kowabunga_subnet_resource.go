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
	"strings"

	sdk "github.com/kowabunga-cloud/kowabunga-go"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	SubnetResourceName = "subnet"

	SubnetDefaultValueDefault     = false
	SubnetDefaultValueApplication = "user"
)

var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

type SubnetResource struct {
	Data *KowabungaProviderData
}

type SubnetResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	Name        types.String   `tfsdk:"name"`
	Desc        types.String   `tfsdk:"desc"`
	VNet        types.String   `tfsdk:"vnet"`
	CIDR        types.String   `tfsdk:"cidr"`
	Gateway     types.String   `tfsdk:"gateway"`
	DNS         types.String   `tfsdk:"dns"`
	Reserved    types.List     `tfsdk:"reserved"`
	GwPool      types.List     `tfsdk:"gw_pool"`
	Routes      types.List     `tfsdk:"routes"`
	Application types.String   `tfsdk:"application"`
	Default     types.Bool     `tfsdk:"default"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, SubnetResourceName)
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a subnet resource",
		Attributes: map[string]schema.Attribute{
			KeyVNet: schema.StringAttribute{
				MarkdownDescription: "Associated virtual network name or ID",
				Required:            true,
			},
			KeyCIDR: schema.StringAttribute{
				MarkdownDescription: "Subnet CIDR",
				Required:            true,
			},
			KeyGateway: schema.StringAttribute{
				MarkdownDescription: "Subnet router/gateway",
				Required:            true,
			},
			KeyDNS: schema.StringAttribute{
				MarkdownDescription: "Subnet DNS server",
				Required:            true,
			},
			KeyReserved: schema.ListAttribute{
				MarkdownDescription: "List of subnet's reserved IPv4 ranges (format: 192.168.0.200-192.168.0.240). IPv4 addresses from these ranges cannot be used by Kowabunga to assign resources.",
				Required:            true,
				ElementType:         types.StringType,
			},
			KeyGwPool: schema.ListAttribute{
				MarkdownDescription: "Subnet's range of IPv4 addresses reserved for local zone's network gateway (format: 192.168.0.200-192.168.0.240). Range size must be at least equal to region's number of zones.",
				Required:            true,
				ElementType:         types.StringType,
			},
			KeyRoutes: schema.ListAttribute{
				MarkdownDescription: "List of extra routes to be access through designated gateway (format: 10.0.0.0/8).",
				Required:            true,
				ElementType:         types.StringType,
			},
			KeyApplication: schema.StringAttribute{
				MarkdownDescription: "Optional application service type (defaults to 'user', possible values: 'user', 'ceph').",
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(SubnetDefaultValueApplication),
			},
			KeyDefault: schema.BoolAttribute{
				MarkdownDescription: "Whether to set subnet as virtual network's default one (default: **false**). The first subnet to be created is always considered as default one.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(SubnetDefaultValueDefault),
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// converts subnet from Terraform model to Kowabunga API model
func subnetResourceToModel(d *SubnetResourceModel) sdk.Subnet {
	reservedRanges := []sdk.IpRange{}
	ranges := []string{}
	d.Reserved.ElementsAs(context.TODO(), &ranges, false)
	for _, item := range ranges {
		split := strings.Split(item, "-")
		if len(split) != 2 {
			continue
		}
		ipr := sdk.IpRange{
			First: split[0],
			Last:  split[1],
		}
		reservedRanges = append(reservedRanges, ipr)
	}

	gwPoolRanges := []sdk.IpRange{}
	gwRanges := []string{}
	d.GwPool.ElementsAs(context.TODO(), &gwRanges, false)
	for _, item := range gwRanges {
		split := strings.Split(item, "-")
		if len(split) != 2 {
			continue
		}
		ipr := sdk.IpRange{
			First: split[0],
			Last:  split[1],
		}
		gwPoolRanges = append(gwPoolRanges, ipr)
	}

	routes := []string{}
	d.Routes.ElementsAs(context.TODO(), &routes, false)

	return sdk.Subnet{
		Name:        d.Name.ValueString(),
		Description: d.Desc.ValueStringPointer(),
		Cidr:        d.CIDR.ValueString(),
		Gateway:     d.Gateway.ValueString(),
		Dns:         d.DNS.ValueStringPointer(),
		Reserved:    reservedRanges,
		GwPool:      gwPoolRanges,
		ExtraRoutes: routes,
		Application: d.Application.ValueStringPointer(),
	}
}

// converts subnet from Kowabunga API model to Terraform model
func subnetModelToResource(s *sdk.Subnet, d *SubnetResourceModel) {
	if s == nil {
		return
	}

	d.Name = types.StringValue(s.Name)
	if s.Description != nil {
		d.Desc = types.StringPointerValue(s.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	d.CIDR = types.StringValue(s.Cidr)
	d.Gateway = types.StringValue(s.Gateway)
	if s.Dns != nil {
		d.DNS = types.StringPointerValue(s.Dns)
	} else {
		d.DNS = types.StringValue("")
	}

	ranges := []attr.Value{}
	for _, item := range s.Reserved {
		ipr := fmt.Sprintf("%s-%s", item.First, item.Last)
		ranges = append(ranges, types.StringValue(ipr))
	}
	d.Reserved, _ = types.ListValue(types.StringType, ranges)

	gwRanges := []attr.Value{}
	for _, item := range s.GwPool {
		ipr := fmt.Sprintf("%s-%s", item.First, item.Last)
		gwRanges = append(gwRanges, types.StringValue(ipr))
	}
	d.GwPool, _ = types.ListValue(types.StringType, gwRanges)

	routes := []attr.Value{}
	for _, r := range s.ExtraRoutes {
		routes = append(routes, types.StringValue(r))
	}
	d.Routes, _ = types.ListValue(types.StringType, routes)

	if s.Application != nil {
		d.Application = types.StringPointerValue(s.Application)
	} else {
		d.Application = types.StringValue(SubnetDefaultValueApplication)
	}
}

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SubnetResourceModel
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

	// find parent vnet
	vnetId, err := getVNetID(ctx, r.Data, data.VNet.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// create a new subnet
	m := subnetResourceToModel(data)
	subnet, _, err := r.Data.K.VnetAPI.CreateSubnet(ctx, vnetId).Subnet(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// set virtual network as default
	if data.Default.ValueBool() {
		_, err = r.Data.K.VnetAPI.SetVNetDefaultSubnet(ctx, vnetId, *subnet.Id).Execute()
		if err != nil {
			errorCreateGeneric(resp, err)
			return
		}
	}

	data.ID = types.StringPointerValue(subnet.Id)
	//subnetModelToResource(vnet, data) // read back resulting object
	tflog.Trace(ctx, "created subnet resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SubnetResourceModel
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

	subnet, _, err := r.Data.K.SubnetAPI.ReadSubnet(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	subnetModelToResource(subnet, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SubnetResourceModel
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

	m := subnetResourceToModel(data)
	_, _, err := r.Data.K.SubnetAPI.UpdateSubnet(ctx, data.ID.ValueString()).Subnet(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SubnetResourceModel
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

	_, err := r.Data.K.SubnetAPI.DeleteSubnet(ctx, data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
