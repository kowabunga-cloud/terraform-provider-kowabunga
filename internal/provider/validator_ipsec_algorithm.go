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
	ValidatorEncryptionAlgorithmDescription = "Encryption Algorithm only supports the following "
	ValidatorIntegrityAlgorithmDescription  = "Integrity Algorithm only supports the following : "
	ValidatorDHAlgorithmDescription         = "Diffie Hellman Algorithm only supports the following : "
	ValidatorAlgorithmErrUnsupported        = "Unsupported algorithm"
)

var encryptionSupportedTypes = []string{
	"AES128",
	"AES256",
	"CAMELLIA128",
	"CAMELLIA256",
}

var integritySupportedTypes = []string{
	"SHA1",
	"SHA256",
	"SHA384",
	"SHA512",
}

var diffieHellmanSupportedTypes = []int64{
	2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
}

type diffieHellmanAlgorithmTypeValidator struct{}

func (v diffieHellmanAlgorithmTypeValidator) Description(ctx context.Context) string {
	return ValidatorIntegrityAlgorithmDescription + strings.Join(integritySupportedTypes, ", ")
}

func (v diffieHellmanAlgorithmTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v diffieHellmanAlgorithmTypeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if !slices.Contains(diffieHellmanSupportedTypes, req.ConfigValue.ValueInt64()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorAlgorithmErrUnsupported,
			fmt.Sprintf("%s: %d", ValidatorDHAlgorithmDescription, req.ConfigValue.ValueInt64()),
		)
		return
	}
}

type integrityAlgorithmTypeValidator struct{}

func (v integrityAlgorithmTypeValidator) Description(ctx context.Context) string {
	return ValidatorIntegrityAlgorithmDescription + strings.Join(integritySupportedTypes, ", ")
}

func (v integrityAlgorithmTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v integrityAlgorithmTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if !slices.Contains(integritySupportedTypes, req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorAlgorithmErrUnsupported,
			fmt.Sprintf("%s. Got : %s", v.Description(ctx), req.ConfigValue.ValueString()),
		)
		return
	}
}

type encryptionAlgorithmTypeValidator struct{}

func (v encryptionAlgorithmTypeValidator) Description(ctx context.Context) string {
	return ValidatorAgentTypeDescription + strings.Join(encryptionSupportedTypes, ", ")
}

func (v encryptionAlgorithmTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v encryptionAlgorithmTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if !slices.Contains(encryptionSupportedTypes, req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			ValidatorAlgorithmErrUnsupported,
			fmt.Sprintf("%s. Got : %s", v.Description(ctx), req.ConfigValue.ValueString()),
		)
		return
	}
}
