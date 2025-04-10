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
	RegionsDataSourceName = "regions"
)

var _ datasource.DataSource = &RegionsDataSource{}
var _ datasource.DataSourceWithConfigure = &RegionsDataSource{}

func NewRegionsDataSource() datasource.DataSource {
	return &RegionsDataSource{}
}

type RegionsDataSource struct {
	Data *KowabungaProviderData
}

type RegionsDataSourceModel struct {
	Regions map[string]types.String `tfsdk:"regions"`
}

func (d *RegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	datasourceMetadata(req, resp, RegionsDataSourceName)
}

func (d *RegionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Data = datasourceConfigure(req, resp)
}

func (d *RegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	datasourceFullSchema(resp, RegionsDataSourceName)
}

func (d *RegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RegionsDataSourceModel
	d.Data.Mutex.Lock()
	defer d.Data.Mutex.Unlock()

	regions, _, err := d.Data.K.RegionAPI.ListRegions(ctx).Execute()
	if err != nil {
		errorDataSourceReadGeneric(resp, err)
		return
	}
	data.Regions = map[string]types.String{}
	for _, rg := range regions {
		r, _, err := d.Data.K.RegionAPI.ReadRegion(ctx, rg).Execute()
		if err != nil {
			continue
		}
		data.Regions[r.Name] = types.StringPointerValue(r.Id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
