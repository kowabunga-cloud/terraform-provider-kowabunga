/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ValidatorDurationDescription    = "Duration must be a number, with or without the following suffix `s`, `m`, `h`, `d`"
	ValidatorDurationErrUnsupported = "Unsupported duration"
)

type stringDurationValidator struct{}

func (v stringDurationValidator) Description(ctx context.Context) string {
	return ValidatorDurationDescription
}

func (v stringDurationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringDurationValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	match := false
	match, err := regexp.MatchString("[0-9]+['s'-'m'-'h'-'d']", req.ConfigValue.ValueString())
	if err != nil || !match {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorDurationErrUnsupported,
			fmt.Sprintf("%s: %s", ValidatorDurationDescription, req.ConfigValue.ValueString()),
		)
		return
	}

}
