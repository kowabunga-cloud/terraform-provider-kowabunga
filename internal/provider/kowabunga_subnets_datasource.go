/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	SubnetsDataSourceName = "subnets"
)

var _ datasource.DataSource = &SubnetsDataSource{}
var _ datasource.DataSourceWithConfigure = &SubnetsDataSource{}

func NewSubnetsDataSource() datasource.DataSource {
	return &SubnetsDataSource{}
}

type SubnetsDataSource struct {
	Data *KowabungaProviderData
}

type SubnetsDataSourceModel struct {
	Subnets map[string]types.String `tfsdk:"subnets"`
}

func (d *SubnetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	datasourceMetadata(req, resp, SubnetsDataSourceName)
}

func (d *SubnetsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Data = datasourceConfigure(req, resp)
}

func (d *SubnetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	datasourceFullSchema(resp, SubnetsDataSourceName)
}

func (d *SubnetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetsDataSourceModel
	d.Data.Mutex.Lock()
	defer d.Data.Mutex.Unlock()

	subnets, _, err := d.Data.K.SubnetAPI.ListSubnets(ctx).Execute()
	if err != nil {
		errorDataSourceReadGeneric(resp, err)
		return
	}
	data.Subnets = map[string]types.String{}
	for _, rg := range subnets {
		r, _, err := d.Data.K.SubnetAPI.ReadSubnet(ctx, rg).Execute()
		if err != nil {
			continue
		}
		data.Subnets[r.Name] = types.StringPointerValue(r.Id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
