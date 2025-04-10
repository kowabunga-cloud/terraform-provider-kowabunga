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
	ValidatorAgentTypeDescription = "Kowabunga remote agent type must be one of the following: "
	ValidatorAgentErrUnsupported  = "Unsupported agent type"
)

var agentSupportedTypes = []string{
	"Kaktus",
	"Kiwi",
}

type stringAgentTypeValidator struct{}

func (v stringAgentTypeValidator) Description(ctx context.Context) string {
	return ValidatorAgentTypeDescription + strings.Join(agentSupportedTypes, ", ")
}

func (v stringAgentTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringAgentTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if !slices.Contains(agentSupportedTypes, req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorAgentErrUnsupported,
			fmt.Sprintf("%s: %s", ValidatorAgentErrUnsupported, req.ConfigValue.ValueString()),
		)
		return
	}
}
