/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	sdk "github.com/kowabunga-cloud/kowabunga-go"
)

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		httpResp *http.Response
		err      error
		want     bool
	}{
		{
			name:     "nil response and nil error",
			httpResp: nil,
			err:      nil,
			want:     false,
		},
		{
			name:     "200 OK response",
			httpResp: &http.Response{StatusCode: http.StatusOK},
			err:      nil,
			want:     false,
		},
		{
			name:     "404 Not Found response",
			httpResp: &http.Response{StatusCode: http.StatusNotFound},
			err:      errors.New("not found"),
			want:     true,
		},
		{
			name:     "500 Internal Server Error",
			httpResp: &http.Response{StatusCode: http.StatusInternalServerError},
			err:      errors.New("server error"),
			want:     false,
		},
		{
			name:     "generic openapi error with nil model",
			httpResp: nil,
			err:      &sdk.GenericOpenAPIError{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNotFoundError(tt.httpResp, tt.err)
			if got != tt.want {
				t.Errorf("isNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleReadError(t *testing.T) {
	ctx := context.Background()

	t.Run("404 removes resource without error", func(t *testing.T) {
		resp := &resource.ReadResponse{
			State: tfsdk.State{
				Schema: schema.Schema{},
			},
		}
		httpResp := &http.Response{StatusCode: http.StatusNotFound}
		err := errors.New("404 Not Found")

		handleReadError(ctx, resp, httpResp, err)
		if resp.Diagnostics.HasError() {
			t.Errorf("expected no diagnostic error on 404, got %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("non-404 error adds diagnostic error", func(t *testing.T) {
		resp := &resource.ReadResponse{
			State: tfsdk.State{
				Schema: schema.Schema{},
			},
		}
		httpResp := &http.Response{StatusCode: http.StatusInternalServerError}
		err := errors.New("server error")

		handleReadError(ctx, resp, httpResp, err)
		if !resp.Diagnostics.HasError() {
			t.Errorf("expected diagnostic error on 500, got none")
		}
	})
}

func TestHandleDeleteError(t *testing.T) {
	t.Run("404 is ignored without error", func(t *testing.T) {
		resp := &resource.DeleteResponse{}
		httpResp := &http.Response{StatusCode: http.StatusNotFound}
		err := errors.New("404 Not Found")

		handleDeleteError(resp, httpResp, err)
		if resp.Diagnostics.HasError() {
			t.Errorf("expected no diagnostic error on 404, got %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("non-404 error adds diagnostic error", func(t *testing.T) {
		resp := &resource.DeleteResponse{}
		httpResp := &http.Response{StatusCode: http.StatusInternalServerError}
		err := errors.New("server error")

		handleDeleteError(resp, httpResp, err)
		if !resp.Diagnostics.HasError() {
			t.Errorf("expected diagnostic error on 500, got none")
		}
	})
}
