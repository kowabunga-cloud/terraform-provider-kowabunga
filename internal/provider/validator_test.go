/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNetworkPortRangesValidator(t *testing.T) {
	v := stringNetworkPortRangesValidator{}
	ctx := context.Background()

	tests := []struct {
		name    string
		input   types.String
		wantErr bool
	}{
		{"valid single port", types.StringValue("80"), false},
		{"valid port range numeric order", types.StringValue("80-1000"), false},
		{"valid comma separated ports", types.StringValue("22, 80, 443"), false},
		{"valid mixed ports and ranges", types.StringValue("80, 443, 8000-9000"), false},
		{"null string ignored", types.StringNull(), false},
		{"unknown string ignored", types.StringUnknown(), false},
		{"empty string rejected", types.StringValue(""), true},
		{"spaces only rejected", types.StringValue("   "), true},
		{"empty token in list rejected", types.StringValue("80,,443"), true},
		{"inverted range rejected", types.StringValue("1000-80"), true},
		{"port above 65535 rejected", types.StringValue("70000"), true},
		{"non-numeric port rejected", types.StringValue("http"), true},
		{"too many range delimiters", types.StringValue("1-2-3"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("ports"),
				ConfigValue: tt.input,
			}
			var resp validator.StringResponse
			v.ValidateString(ctx, req, &resp)
			if resp.Diagnostics.HasError() != tt.wantErr {
				t.Errorf("got error %v (%v), wantErr %v", resp.Diagnostics.HasError(), resp.Diagnostics.Errors(), tt.wantErr)
			}
		})
	}
}

func TestNetworkAddressValidator(t *testing.T) {
	v := stringNetworkAddressValidator{}
	ctx := context.Background()

	tests := []struct {
		name    string
		input   types.String
		wantErr bool
	}{
		{"valid IPv4 address", types.StringValue("192.168.1.1"), false},
		{"valid IPv4 CIDR", types.StringValue("10.0.0.0/16"), false},
		{"valid IPv4 host CIDR", types.StringValue("192.168.1.1/32"), false},
		{"null ignored", types.StringNull(), false},
		{"unknown ignored", types.StringUnknown(), false},
		{"IPv6 address rejected", types.StringValue("2001:db8::1"), true},
		{"IPv6 CIDR rejected", types.StringValue("2001:db8::/32"), true},
		{"invalid string rejected", types.StringValue("invalid-ip"), true},
		{"out of range IPv4 rejected", types.StringValue("256.0.0.1"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("address"),
				ConfigValue: tt.input,
			}
			var resp validator.StringResponse
			v.ValidateString(ctx, req, &resp)
			if resp.Diagnostics.HasError() != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", resp.Diagnostics.HasError(), tt.wantErr)
			}
		})
	}
}

func TestNetworkPortValidator(t *testing.T) {
	v := intNetworkPortValidator{}
	ctx := context.Background()

	tests := []struct {
		name    string
		input   types.Int64
		wantErr bool
	}{
		{"port 0 valid", types.Int64Value(0), false},
		{"port 80 valid", types.Int64Value(80), false},
		{"port 65535 valid", types.Int64Value(65535), false},
		{"null ignored", types.Int64Null(), false},
		{"unknown ignored", types.Int64Unknown(), false},
		{"negative port rejected", types.Int64Value(-1), true},
		{"port above 65535 rejected", types.Int64Value(65536), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.Int64Request{
				Path:        path.Root("port"),
				ConfigValue: tt.input,
			}
			var resp validator.Int64Response
			v.ValidateInt64(ctx, req, &resp)
			if resp.Diagnostics.HasError() != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", resp.Diagnostics.HasError(), tt.wantErr)
			}
		})
	}
}

func TestNetworkProtocolValidator(t *testing.T) {
	v := stringNetworkProtocolValidator{}
	ctx := context.Background()

	tests := []struct {
		name    string
		input   types.String
		wantErr bool
	}{
		{"tcp valid", types.StringValue("tcp"), false},
		{"udp valid", types.StringValue("udp"), false},
		{"case insensitive TCP valid", types.StringValue("TCP"), false},
		{"null ignored", types.StringNull(), false},
		{"unknown ignored", types.StringUnknown(), false},
		{"icmp rejected", types.StringValue("icmp"), true},
		{"random string rejected", types.StringValue("http"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("protocol"),
				ConfigValue: tt.input,
			}
			var resp validator.StringResponse
			v.ValidateString(ctx, req, &resp)
			if resp.Diagnostics.HasError() != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", resp.Diagnostics.HasError(), tt.wantErr)
			}
		})
	}
}

func TestDurationValidator(t *testing.T) {
	v := stringDurationValidator{}
	ctx := context.Background()

	tests := []struct {
		name    string
		input   types.String
		wantErr bool
	}{
		{"bare number valid", types.StringValue("3600"), false},
		{"seconds suffix valid", types.StringValue("30s"), false},
		{"minutes suffix valid", types.StringValue("15m"), false},
		{"hours suffix valid", types.StringValue("8h"), false},
		{"days suffix valid", types.StringValue("1d"), false},
		{"null ignored", types.StringNull(), false},
		{"unknown ignored", types.StringUnknown(), false},
		{"negative number rejected", types.StringValue("-10s"), true},
		{"unsupported suffix rejected", types.StringValue("10w"), true},
		{"letters only rejected", types.StringValue("fast"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("duration"),
				ConfigValue: tt.input,
			}
			var resp validator.StringResponse
			v.ValidateString(ctx, req, &resp)
			if resp.Diagnostics.HasError() != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", resp.Diagnostics.HasError(), tt.wantErr)
			}
		})
	}
}

func TestFirewallPolicyValidator(t *testing.T) {
	v := stringFirewallPolicyValidator{}
	ctx := context.Background()

	tests := []struct {
		name    string
		input   types.String
		wantErr bool
	}{
		{"accept valid", types.StringValue("accept"), false},
		{"drop valid", types.StringValue("drop"), false},
		{"null ignored", types.StringNull(), false},
		{"unknown ignored", types.StringUnknown(), false},
		{"reject unsupported", types.StringValue("reject"), true},
		{"deny unsupported", types.StringValue("deny"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("policy"),
				ConfigValue: tt.input,
			}
			var resp validator.StringResponse
			v.ValidateString(ctx, req, &resp)
			if resp.Diagnostics.HasError() != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", resp.Diagnostics.HasError(), tt.wantErr)
			}
		})
	}
}

func TestIPsecAlgorithmValidators(t *testing.T) {
	ctx := context.Background()

	t.Run("Diffie-Hellman validator", func(t *testing.T) {
		v := diffieHellmanAlgorithmTypeValidator{}
		validGroups := []int64{2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
		for _, g := range validGroups {
			var resp validator.Int64Response
			v.ValidateInt64(ctx, validator.Int64Request{
				Path:        path.Root("dh"),
				ConfigValue: types.Int64Value(g),
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("group %d should be valid", g)
			}
		}

		var resp validator.Int64Response
		v.ValidateInt64(ctx, validator.Int64Request{
			Path:        path.Root("dh"),
			ConfigValue: types.Int64Value(99),
		}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("group 99 should be invalid")
		}
	})

	t.Run("Integrity validator", func(t *testing.T) {
		v := integrityAlgorithmTypeValidator{}
		for _, alg := range []string{"SHA1", "SHA256", "SHA384", "SHA512"} {
			var resp validator.StringResponse
			v.ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("integrity"),
				ConfigValue: types.StringValue(alg),
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("alg %s should be valid", alg)
			}
		}

		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:        path.Root("integrity"),
			ConfigValue: types.StringValue("MD5"),
		}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("MD5 should be invalid")
		}
	})

	t.Run("Encryption validator", func(t *testing.T) {
		v := encryptionAlgorithmTypeValidator{}
		for _, alg := range []string{"AES128", "AES256", "CAMELLIA128", "CAMELLIA256"} {
			var resp validator.StringResponse
			v.ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("encryption"),
				ConfigValue: types.StringValue(alg),
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("alg %s should be valid", alg)
			}
		}

		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:        path.Root("encryption"),
			ConfigValue: types.StringValue("DES"),
		}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("DES should be invalid")
		}
	})
}

func TestUserValidators(t *testing.T) {
	ctx := context.Background()

	t.Run("User email validator", func(t *testing.T) {
		v := stringUserEmailValidator{}
		validEmails := []string{"user@example.com", "admin@sub.domain.org", "test.name+tag@company.co.uk"}
		for _, email := range validEmails {
			var resp validator.StringResponse
			v.ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("email"),
				ConfigValue: types.StringValue(email),
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("email %s should be valid", email)
			}
		}

		invalidEmails := []string{"plainaddress", "@missinguser.com", "user@", "user@.com"}
		for _, email := range invalidEmails {
			var resp validator.StringResponse
			v.ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("email"),
				ConfigValue: types.StringValue(email),
			}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Errorf("email %s should be invalid", email)
			}
		}
	})

	t.Run("User role validator", func(t *testing.T) {
		v := stringUserRoleValidator{}
		validRoles := []string{"superAdmin", "projectAdmin", "user"}
		for _, role := range validRoles {
			var resp validator.StringResponse
			v.ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("role"),
				ConfigValue: types.StringValue(role),
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("role %s should be valid", role)
			}
		}

		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:        path.Root("role"),
			ConfigValue: types.StringValue("guest"),
		}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("role guest should be invalid")
		}
	})
}

func TestAgentTypeValidator(t *testing.T) {
	v := stringAgentTypeValidator{}
	ctx := context.Background()

	for _, agent := range []string{"Kaktus", "Kiwi"} {
		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:        path.Root("type"),
			ConfigValue: types.StringValue(agent),
		}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("agent %s should be valid", agent)
		}
	}

	var resp validator.StringResponse
	v.ValidateString(ctx, validator.StringRequest{
		Path:        path.Root("type"),
		ConfigValue: types.StringValue("UnknownAgent"),
	}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Errorf("UnknownAgent should be invalid")
	}
}
