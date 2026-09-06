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
	TeamsDataSourceName = "teams"
)

var _ datasource.DataSource = &TeamsDataSource{}
var _ datasource.DataSourceWithConfigure = &TeamsDataSource{}

func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

type TeamsDataSource struct {
	Data *KowabungaProviderData
}

type TeamsDataSourceModel struct {
	Teams map[string]types.String `tfsdk:"teams"`
}

func (d *TeamsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	datasourceMetadata(req, resp, TeamsDataSourceName)
}

func (d *TeamsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Data = datasourceConfigure(req, resp)
}

func (d *TeamsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	datasourceFullSchema(resp, TeamsDataSourceName)
}

func (d *TeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamsDataSourceModel
	d.Data.Mutex.Lock()
	defer d.Data.Mutex.Unlock()

	teams, _, err := d.Data.K.TeamAPI.ListTeams(ctx).Execute()
	if err != nil {
		errorDataSourceReadGeneric(resp, err)
		return
	}
	data.Teams = map[string]types.String{}
	for _, tn := range teams {
		r, _, err := d.Data.K.TeamAPI.ReadTeam(ctx, tn).Execute()
		if err != nil || r == nil || r.Id == nil {
			continue
		}
		data.Teams[r.Name] = types.StringPointerValue(r.Id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
