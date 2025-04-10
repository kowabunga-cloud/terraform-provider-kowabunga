/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ErrorUnconfiguredDataSource = "Unexpected DataSource Configure Type"
)

const (
	DataSourceIdDescription   = "Datasource object internal identifier"
	DataSourceNameDescription = "Datasource name"
)

type GenericDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func datasourceMetadata(req datasource.MetadataRequest, resp *datasource.MetadataResponse, name string) {
	resp.TypeName = req.ProviderTypeName + "_" + name
}

func errorUnconfiguredDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.AddError(
		ErrorUnconfiguredDataSource,
		fmt.Sprintf(ErrorExpectedProviderData, req.ProviderData),
	)
}

func errorDataSourceReadGeneric(resp *datasource.ReadResponse, err error) {
	resp.Diagnostics.AddError(ErrorGeneric, err.Error())
}

func datasourceConfigure(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *KowabungaProviderData {
	if req.ProviderData == nil {
		return nil
	}

	kd, ok := req.ProviderData.(*KowabungaProviderData)
	if !ok {
		errorUnconfiguredDataSource(req, resp)
		return nil
	}

	return kd
}

func datasourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		KeyID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: DataSourceIdDescription,
		},
		KeyName: schema.StringAttribute{
			MarkdownDescription: DataSourceNameDescription,
			Required:            true,
		},
	}
}

func datasourceFilteredSchema(resp *datasource.SchemaResponse, rs string) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf("Data from a %s resource", rs),
		Attributes:          datasourceAttributes(),
	}
}

func datasourceFullSchema(resp *datasource.SchemaResponse, rs string) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf("Data from %s", rs),
		Attributes: map[string]schema.Attribute{
			rs: schema.MapAttribute{
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("List of Kowabunga %s", rs),
				ElementType:         types.StringType,
			},
		},
	}
}
