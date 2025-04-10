/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"

	emailverifier "github.com/AfterShip/email-verifier"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	ValidatorUserEmailDescription    = "Kowabunga user email is malformed"
	ValidatorUserEmailErrUnsupported = "Unsupported user email"
	ValidatorUserEmailErrMalformed   = "Malformed user email"
)

type stringUserEmailValidator struct{}

func (v stringUserEmailValidator) Description(ctx context.Context) string {
	return ValidatorUserEmailDescription
}

func (v stringUserEmailValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringUserEmailValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	verifier := emailverifier.NewVerifier()
	ret, err := verifier.Verify(req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorUserEmailErrUnsupported,
			fmt.Sprintf("%s: %s", ValidatorUserEmailErrUnsupported, req.ConfigValue.ValueString()),
		)
		return
	}
	if !ret.Syntax.Valid {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorUserEmailErrMalformed,
			fmt.Sprintf("%s: %s", ValidatorUserEmailErrMalformed, req.ConfigValue.ValueString()),
		)
		return
	}
}
