/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ValidatorNetworkAddressDescription = "String must be a valid IPv4 address or CIDR"
	ValidatorNetworkAddressErrInvalid  = "Invalid IPv4 address or CIDR"
)

type stringNetworkAddressValidator struct{}

func (v stringNetworkAddressValidator) Description(ctx context.Context) string {
	return ValidatorNetworkAddressDescription
}

func (v stringNetworkAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringNetworkAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	ip := req.ConfigValue.ValueString()
	valid_ip := (net.ParseIP(ip) != nil)
	valid_cidr := true
	_, _, err := net.ParseCIDR(ip)
	if err != nil {
		valid_cidr = false
	}

	if !valid_ip && !valid_cidr {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorNetworkAddressErrInvalid,
			fmt.Sprintf("%s: %s", ValidatorNetworkAddressErrInvalid, ip),
		)
	}
}
