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

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ValidatorFirewallPolicyDescription    = "Protocol must be one of 'accept', 'drop'"
	ValidatorFirewallPolicyErrUnsupported = "Unsupported policy"
)

var firewallSupportedPolicy = []string{
	"accept",
	"drop",
}

type stringFirewallPolicyValidator struct{}

func (v stringFirewallPolicyValidator) Description(ctx context.Context) string {
	return ValidatorFirewallPolicyDescription
}

func (v stringFirewallPolicyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringFirewallPolicyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	protocol := req.ConfigValue.ValueString()
	if !slices.Contains(firewallSupportedPolicy, protocol) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorFirewallPolicyErrUnsupported,
			fmt.Sprintf("%s: %s", ValidatorFirewallPolicyErrUnsupported, protocol),
		)
	}
}
