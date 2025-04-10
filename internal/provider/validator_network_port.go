/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ValidatorNetworkPortDescription     = "Port number must be an integer between 0 and 65535"
	ValidatorNetworkPortErrOutsideRange = "Port number is outside range (0-65535)"
)

type intNetworkPortValidator struct{}

func (v intNetworkPortValidator) Description(ctx context.Context) string {
	return ValidatorNetworkPortDescription
}

func (v intNetworkPortValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v intNetworkPortValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	port := req.ConfigValue.ValueInt64()
	if port < 0 || port > 65535 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorNetworkPortErrOutsideRange,
			fmt.Sprintf("%s: %d", ValidatorNetworkPortsErrOutsideRange, port),
		)
		return
	}
}
