/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	SubnetDataSourceName           = "subnet"
	SubnetDataSourceAppDescription = "Datasource application"

	SubnetDataSourceErrTooFewArguments  = "either 'name' or 'app' field is required"
	SubnetDataSourceErrTooManyArguments = "one can't ask for both name and app fields"
)

type SubnetDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	App  types.String `tfsdk:"app"`
}

func subnetDatasourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		KeyID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: DataSourceIdDescription,
		},
		KeyName: schema.StringAttribute{
			MarkdownDescription: DataSourceNameDescription,
			Optional:            true,
			Computed:            true,
		},
		KeyApp: schema.StringAttribute{
			MarkdownDescription: SubnetDataSourceAppDescription,
			Optional:            true,
			Computed:            true,
		},
	}
}

var _ datasource.DataSource = &SubnetDataSource{}
var _ datasource.DataSourceWithConfigure = &SubnetDataSource{}

func NewSubnetDataSource() datasource.DataSource {
	return &SubnetDataSource{}
}

type SubnetDataSource struct {
	Data *KowabungaProviderData
}

func (d *SubnetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	datasourceMetadata(req, resp, SubnetDataSourceName)
}

func (d *SubnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.Data = datasourceConfigure(req, resp)
}

func (d *SubnetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf("Data from a %s resource", SubnetDataSourceName),
		Attributes:          subnetDatasourceAttributes(),
	}
}

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.Data.Mutex.Lock()
	defer d.Data.Mutex.Unlock()

	// check that at least one argument has been passed over
	if data.Name.ValueString() == "" && data.App.ValueString() == "" {
		resp.Diagnostics.AddError(ErrorGeneric, SubnetDataSourceErrTooFewArguments)
		return
	}

	// check that no all arguments has been passed over
	if data.Name.ValueString() != "" && data.App.ValueString() != "" {
		resp.Diagnostics.AddError(ErrorGeneric, SubnetDataSourceErrTooManyArguments)
		return
	}

	subnets, _, err := d.Data.K.SubnetAPI.ListSubnets(ctx).Execute()
	if err != nil {
		errorDataSourceReadGeneric(resp, err)
		return
	}
	for _, rg := range subnets {
		r, _, err := d.Data.K.SubnetAPI.ReadSubnet(ctx, rg).Execute()
		if err != nil {
			errorDataSourceReadGeneric(resp, err)
			return
		}

		// request by name
		if r.Name == data.Name.ValueString() {
			data.ID = types.StringPointerValue(r.Id)
			if r.Application != nil {
				data.App = types.StringPointerValue(r.Application)
			} else {
				data.App = types.StringValue("")
			}
			break
		}

		// request by app
		if r.Application != nil && *r.Application == data.App.ValueString() {
			data.ID = types.StringPointerValue(r.Id)
			data.Name = types.StringValue(r.Name)
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
