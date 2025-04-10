/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ValidatorUserRoleDescription    = "Kowabunga user role type must be one of the following: "
	ValidatorUserRoleErrUnsupported = "Unsupported user role"
)

var userSupportedRoles = []string{
	"superAdmin",
	"projectAdmin",
	"user",
}

type stringUserRoleValidator struct{}

func (v stringUserRoleValidator) Description(ctx context.Context) string {
	return ValidatorUserRoleDescription + strings.Join(userSupportedRoles, ", ")
}

func (v stringUserRoleValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringUserRoleValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if !slices.Contains(userSupportedRoles, req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorUserRoleErrUnsupported,
			fmt.Sprintf("%s: %s", ValidatorUserRoleErrUnsupported, req.ConfigValue.ValueString()),
		)
		return
	}
}
