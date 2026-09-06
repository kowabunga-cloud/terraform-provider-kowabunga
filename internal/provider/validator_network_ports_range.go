/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ValidatorNetworkPortsDescription       = "Ports format must follow the following rules : comma separated, ranges ordered with a \"-\" char. e.g : 1234, 5678-5690"
	ValidatorNetworkPortsErrInvalidPort    = "Invalid port"
	ValidatorNetworkPortsErrInvalidRange   = "Invalid range"
	ValidatorNetworkPortsErrTooManyEntries = "Too many entries in range"
	ValidatorNetworkPortsErrOutsideRange   = "Port outside range (0-65535) for port"
	ValidatorNetworkPortsErrBogusRange     = "Left hand side is superior than right hand side"
)

type stringNetworkPortRangesValidator struct{}

func (v stringNetworkPortRangesValidator) Description(ctx context.Context) string {
	return ValidatorNetworkPortsDescription
}

func (v stringNetworkPortRangesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v stringNetworkPortRangesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	raw := strings.TrimSpace(req.ConfigValue.ValueString())
	if raw == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorNetworkPortsErrInvalidPort,
			fmt.Sprintf("%s: empty port list", ValidatorNetworkPortsErrInvalidPort),
		)
		return
	}

	portList := strings.Split(raw, ",")
	for _, port := range portList {
		trimmedPort := strings.TrimSpace(port)
		if trimmedPort == "" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				ValidatorNetworkPortsErrInvalidPort,
				fmt.Sprintf("%s: empty port entry", ValidatorNetworkPortsErrInvalidPort),
			)
			return
		}
		portRanges := strings.Split(trimmedPort, "-") //returns at least 1 entry
		if len(portRanges) > 2 {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				ValidatorNetworkPortsErrTooManyEntries,
				fmt.Sprintf("%s: %s", ValidatorNetworkPortsErrTooManyEntries, trimmedPort),
			)
			return
		}
		p0, err := strconv.ParseUint(strings.TrimSpace(portRanges[0]), 10, 16)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				ValidatorNetworkPortsErrInvalidPort,
				fmt.Sprintf("%s: %s", ValidatorNetworkPortsErrOutsideRange, strings.TrimSpace(portRanges[0])),
			)
			return
		}
		if len(portRanges) == 2 {
			p1, err := strconv.ParseUint(strings.TrimSpace(portRanges[1]), 10, 16)
			if err != nil {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					ValidatorNetworkPortsErrInvalidPort,
					fmt.Sprintf("%s: %s", ValidatorNetworkPortsErrOutsideRange, strings.TrimSpace(portRanges[1])),
				)
				return
			}
			if p0 > p1 {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					ValidatorNetworkPortsErrInvalidRange,
					fmt.Sprintf("%s: %s", ValidatorNetworkPortsErrBogusRange, trimmedPort),
				)
				return
			}
		}
	}
}
