/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"fmt"
	"maps"

	sdk "github.com/kowabunga-cloud/kowabunga-go"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	KawaiiIPsecResourceName = "kawaii_ipsec"

	KawaiiIPsecDefaultValueIngressProtocol = "tcp"
	KawaiiIPsecDefaultValueIngressPolicy   = "accept"

	KawaiiIPsecDefaultDpdTimeout    = "240s"
	KawaiiIPsecDefaultDpdAction     = "restart"
	KawaiiIPsecDefaultStartAction   = "start"
	KawaiiIPsecDefaultRekeyTime     = "2h"
	KawaiiIPsecDefaultPhaseLifetime = "1h"
)

var _ resource.Resource = &KawaiiResource{}
var _ resource.ResourceWithImportState = &KawaiiIPsecConnectionResource{}

func NewKawaiiIPsecResource() resource.Resource {
	return &KawaiiIPsecConnectionResource{}
}

type KawaiiIPsecConnectionResource struct {
	Data *KowabungaProviderData
}

type KawaiiIPsecConnectionResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
	Desc     types.String   `tfsdk:"desc"`

	KawaiiID                  types.String `tfsdk:"kawaii"`
	Name                      types.String `tfsdk:"name"`
	IP                        types.String `tfsdk:"ip"`
	PreSharedKey              types.String `tfsdk:"pre_shared_key"`
	RemotePeer                types.String `tfsdk:"remote_peer"`
	RemoteSubnet              types.String `tfsdk:"remote_subnet"`
	DpdTimeout                types.String `tfsdk:"dpd_timeout"`
	DpdTimeoutAction          types.String `tfsdk:"dpd_action"`
	StartAction               types.String `tfsdk:"start_action"`
	Rekey                     types.String `tfsdk:"rekey"`
	Phase1Lifetime            types.String `tfsdk:"phase1_lifetime"`
	Phase1DHGroupNumber       types.Int64  `tfsdk:"phase1_dh_group_number"`
	Phase1IntegrityAlgorithm  types.String `tfsdk:"phase1_integrity_algorithm"`
	Phase1EncryptionAlgorithm types.String `tfsdk:"phase1_encryption_algorithm"`
	Phase2Lifetime            types.String `tfsdk:"phase2_lifetime"`
	Phase2DHGroupNumber       types.Int64  `tfsdk:"phase2_dh_group_number"`
	Phase2IntegrityAlgorithm  types.String `tfsdk:"phase2_integrity_algorithm"`
	Phase2EncryptionAlgorithm types.String `tfsdk:"phase2_encryption_algorithm"`
	IngressRules              types.List   `tfsdk:"ingress_rules"` // KawaiiForwardRule
}

func (r *KawaiiIPsecConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resourceMetadata(req, resp, KawaiiIPsecResourceName)
}

func (r *KawaiiIPsecConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceImportState(ctx, req, resp)
}

func (r *KawaiiIPsecConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Data = resourceConfigure(req, resp)
}

func (r *KawaiiIPsecConnectionResource) SchemaIngressRule() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			KeySource: schema.StringAttribute{
				MarkdownDescription: "The source IP or CIDR to accept public traffic from (defaults to 0.0.0.0/0).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiDefaultValueSource),
				Validators: []validator.String{
					&stringNetworkAddressValidator{},
				},
			},
			KeyProtocol: schema.StringAttribute{
				MarkdownDescription: "The transport layer protocol to forward public traffic to (defaults to 'tcp')",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiIPsecDefaultValueIngressProtocol),
				Validators: []validator.String{
					&stringNetworkProtocolValidator{},
				},
			},
			KeyPorts: schema.StringAttribute{
				MarkdownDescription: "The ports (or range of ports) allowed to receive traffic. Ranges are accepted. Format is a-b,c-d (e.g. 443; 22,80,443; 80,443,3000-3005).",
				Required:            true,
				Validators: []validator.String{
					&stringNetworkPortRangesValidator{},
				},
			},
		},
	}
}

func (r *KawaiiIPsecConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kawaii list of Kowabunga IPsec Connections",
		Attributes: map[string]schema.Attribute{
			KeyIP: schema.StringAttribute{
				MarkdownDescription: "The local IPsec IP (read-only)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			KeyKawaii: schema.StringAttribute{
				MarkdownDescription: "Associated Kawaii name or ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			KeyName: schema.StringAttribute{
				MarkdownDescription: "Kowabunga IPsec Connection Name",
				Required:            true,
			},
			KeyRemotePeer: schema.StringAttribute{
				MarkdownDescription: "Remote VPN Gateway",
				Required:            true,
				Validators: []validator.String{
					&stringNetworkAddressValidator{},
				},
			},
			KeyPreSharedKey: schema.StringAttribute{
				MarkdownDescription: "The Pre-Shared Key (PSK) to authenticate the VPN tunnel to your peer VPN gateway",
				Required:            true,
			},
			KeyRemoteSubnet: schema.StringAttribute{
				MarkdownDescription: "Remote Subnet CIDR",
				Required:            true,
				Validators: []validator.String{
					&stringNetworkAddressValidator{},
				},
			},
			KeyIPsecDpdTimeout: schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Dead Peer Detection Timeout. Default is `%s`", KawaiiIPsecDefaultDpdTimeout),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiIPsecDefaultDpdTimeout),
			},
			KeyIPsecDpdAction: schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Dead Peer Detection Timeout Action. Default is `%s`", KawaiiIPsecDefaultDpdAction),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiIPsecDefaultDpdAction),
			},
			KeyIPsecStartAction: schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("IPsec Default Start Action. Default is `%s`", KawaiiIPsecDefaultStartAction),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiIPsecDefaultStartAction),
			},
			KeyIPsecRekeyTime: schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("IPsec Rekey time in seconds. Default is `%s`", KawaiiIPsecDefaultRekeyTime),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiIPsecDefaultRekeyTime),
				Validators: []validator.String{
					&stringDurationValidator{},
				},
			},
			KeyIPsecP1Lifetime: schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("IPsec Phase 1 Lifetime. Use s, m, h and d suffixes. Default is `%s`", KawaiiIPsecDefaultPhaseLifetime),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiIPsecDefaultPhaseLifetime),
				Validators: []validator.String{
					&stringDurationValidator{},
				},
			},
			KeyIPsecP1DHGroupNumber: schema.Int64Attribute{
				MarkdownDescription: "IPsec phase 1 Diffie Hellman IANA Group Number. Valid values are `2 | 5 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`",
				Required:            true,
				Validators: []validator.Int64{
					&diffieHellmanAlgorithmTypeValidator{},
				},
			},
			KeyIPsecP1IntegrityAlgorithm: schema.StringAttribute{
				MarkdownDescription: "IPsec phase 1 Integrity Algorithm. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`",
				Required:            true,
				Validators: []validator.String{
					&integrityAlgorithmTypeValidator{},
				},
			},
			KeyIPsecP1EncryptionAlgorithm: schema.StringAttribute{
				MarkdownDescription: "IPsec phase 1 Encryption Algorithm. Valid values are `AES128 | AES256 | CAMELLIA128 | CAMELLIA256`",
				Required:            true,
				Validators: []validator.String{
					&encryptionAlgorithmTypeValidator{},
				},
			},
			KeyIPsecP2Lifetime: schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("IPsec Phase 2 Lifetime. Use s, m, h and d suffixes. Default is `%s`", KawaiiIPsecDefaultPhaseLifetime),
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(KawaiiIPsecDefaultPhaseLifetime),
				Validators: []validator.String{
					&stringDurationValidator{},
				},
			},
			KeyIPsecP2DHGroupNumber: schema.Int64Attribute{
				MarkdownDescription: "IPsec phase 2 Diffie Hellman IANA Group Number. Valid values are `2 | 5 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`",
				Required:            true,
				Validators: []validator.Int64{
					&diffieHellmanAlgorithmTypeValidator{},
				},
			},
			KeyIPsecP2IntegrityAlgorithm: schema.StringAttribute{
				MarkdownDescription: "IPsec phase 1 Integrity Algorithm. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`",
				Required:            true,
				Validators: []validator.String{
					&integrityAlgorithmTypeValidator{},
				},
			},
			KeyIPsecP2EncryptionAlgorithm: schema.StringAttribute{
				MarkdownDescription: "IPsec phase 1 Encryption Algorithm. Valid values are `AES128 | AES256 | CAMELLIA128 | CAMELLIA256`",
				Required:            true,
				Validators: []validator.String{
					&encryptionAlgorithmTypeValidator{},
				},
			},
			KeyIngressRules: schema.ListNestedAttribute{
				MarkdownDescription: "The firewall list of Ingress Rules. Default will accept all. Egress is allow all",
				Optional:            true,
				Computed:            true,
				NestedObject:        r.SchemaIngressRule(),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
	maps.Copy(resp.Schema.Attributes, resourceAttributes(&ctx))
}

// ////////////////////////////////////////////////////////////////////
// converts kawaii Ipsec from Terraform model to Kowabunga API model //
// ////////////////////////////////////////////////////////////////////
func kawaiiIPsecResourceModel(ctx *context.Context, d *KawaiiIPsecConnectionResourceModel) sdk.KawaiiIpSec {

	return sdk.KawaiiIpSec{
		Name:                      d.Name.ValueString(),
		Ip:                        d.IP.ValueStringPointer(),
		Description:               d.Desc.ValueStringPointer(),
		RemoteIp:                  d.RemotePeer.ValueString(),
		RemoteSubnet:              d.RemoteSubnet.ValueString(),
		PreSharedKey:              d.PreSharedKey.ValueString(),
		DpdTimeoutAction:          d.DpdTimeoutAction.ValueStringPointer(),
		DpdTimeout:                d.DpdTimeout.ValueStringPointer(),
		StartAction:               d.StartAction.ValueStringPointer(),
		RekeyTime:                 d.Rekey.ValueStringPointer(),
		Phase1Lifetime:            d.Phase1Lifetime.ValueStringPointer(),
		Phase1DhGroupNumber:       d.Phase1DHGroupNumber.ValueInt64(),
		Phase1IntegrityAlgorithm:  d.Phase1IntegrityAlgorithm.ValueString(),
		Phase1EncryptionAlgorithm: d.Phase1EncryptionAlgorithm.ValueString(),
		Phase2Lifetime:            d.Phase2Lifetime.ValueStringPointer(),
		Phase2DhGroupNumber:       d.Phase2DHGroupNumber.ValueInt64(),
		Phase2IntegrityAlgorithm:  d.Phase2IntegrityAlgorithm.ValueString(),
		Phase2EncryptionAlgorithm: d.Phase2EncryptionAlgorithm.ValueString(),
		Firewall:                  kawaiiIPsecFirewallModel(ctx, d),
	}
}

func kawaiiIPsecFirewallModel(ctx *context.Context, d *KawaiiIPsecConnectionResourceModel) *sdk.KawaiiFirewall {
	fwModel := sdk.KawaiiFirewall{
		Ingress: []sdk.KawaiiFirewallIngressRule{},
	}

	// Ingress Rules
	ingressRules := make([]types.Object, 0, len(d.IngressRules.Elements()))
	ingressDiags := d.IngressRules.ElementsAs(*ctx, &ingressRules, false)
	if ingressDiags.HasError() {
		for _, err := range ingressDiags.Errors() {
			tflog.Debug(*ctx, err.Detail())
		}
	}
	for _, ir := range ingressRules {
		rule := KawaiiIngressRule{}
		diags := ir.As(*ctx, &rule, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			for _, err := range diags.Errors() {
				tflog.Error(*ctx, err.Detail())
			}
		}

		fwModel.Ingress = append(fwModel.Ingress, sdk.KawaiiFirewallIngressRule{
			Source:   rule.Source.ValueStringPointer(),
			Protocol: rule.Protocol.ValueStringPointer(),
			Ports:    rule.Ports.ValueString(),
		})
	}
	return &fwModel
}

/////////////////////////////////////////////////////////////////
// converts Kawaii from Kowabunga API model to Terraform model //
/////////////////////////////////////////////////////////////////

func kawaiiIPsecModelToIngressRules(ctx *context.Context, r *sdk.KawaiiIpSec, d *KawaiiIPsecConnectionResourceModel) {
	// ingress rules
	ingressRules := []attr.Value{}
	ingressRuleType := map[string]attr.Type{
		KeySource:   types.StringType,
		KeyProtocol: types.StringType,
		KeyPorts:    types.StringType,
	}
	for _, ir := range r.Firewall.Ingress {
		source := KawaiiDefaultValueSource
		if ir.Source != nil {
			source = *ir.Source
		}
		protocol := KawaiiDefaultValueProtocol
		if ir.Protocol != nil {
			protocol = *ir.Protocol
		}
		r := map[string]attr.Value{
			KeySource:   types.StringValue(source),
			KeyProtocol: types.StringValue(protocol),
			KeyPorts:    types.StringValue(ir.Ports),
		}
		object, _ := types.ObjectValue(ingressRuleType, r)
		ingressRules = append(ingressRules, object)
	}

	if len(r.Firewall.Ingress) == 0 {
		d.IngressRules = types.ListNull(types.ObjectType{AttrTypes: ingressRuleType})
	} else {

		d.IngressRules, _ = types.ListValue(types.ObjectType{AttrTypes: ingressRuleType}, ingressRules)
	}
}

func kawaiiIPsecModelToResource(ctx *context.Context, r *sdk.KawaiiIpSec, d *KawaiiIPsecConnectionResourceModel) {
	if r == nil {
		return
	}
	d.Name = types.StringValue(r.Name)
	if r.Ip != nil {
		d.IP = types.StringPointerValue(r.Ip)
	} else {
		d.IP = types.StringValue("")
	}
	d.RemotePeer = types.StringValue(r.RemoteIp)
	d.RemoteSubnet = types.StringValue(r.RemoteSubnet)
	d.PreSharedKey = types.StringValue(r.PreSharedKey)
	if r.Description != nil {
		d.Desc = types.StringPointerValue(r.Description)
	} else {
		d.Desc = types.StringValue("")
	}
	if r.DpdTimeoutAction != nil {
		d.DpdTimeoutAction = types.StringPointerValue(r.DpdTimeoutAction)
	} else {
		d.DpdTimeoutAction = types.StringValue(KawaiiIPsecDefaultDpdTimeout)
	}
	if r.DpdTimeout != nil {
		d.DpdTimeout = types.StringPointerValue(r.DpdTimeout)
	} else {
		d.DpdTimeout = types.StringValue(KawaiiIPsecDefaultDpdAction)
	}
	if r.StartAction != nil {
		d.StartAction = types.StringPointerValue(r.StartAction)
	} else {
		d.StartAction = types.StringValue(KawaiiIPsecDefaultStartAction)
	}
	if r.RekeyTime != nil {
		d.Rekey = types.StringPointerValue(r.RekeyTime)
	} else {
		d.Rekey = types.StringValue(KawaiiIPsecDefaultRekeyTime)
	}
	if r.Phase1Lifetime != nil {
		d.Phase1Lifetime = types.StringPointerValue(r.Phase1Lifetime)
	} else {
		d.Phase1Lifetime = types.StringValue(KawaiiIPsecDefaultPhaseLifetime)
	}
	d.Phase1DHGroupNumber = types.Int64Value(r.Phase1DhGroupNumber)
	d.Phase1IntegrityAlgorithm = types.StringValue(r.Phase1IntegrityAlgorithm)
	d.Phase1EncryptionAlgorithm = types.StringValue(r.Phase1EncryptionAlgorithm)
	if r.Phase2Lifetime != nil {
		d.Phase2Lifetime = types.StringPointerValue(r.Phase2Lifetime)
	} else {
		d.Phase2Lifetime = types.StringValue(KawaiiIPsecDefaultPhaseLifetime)
	}
	d.Phase2DHGroupNumber = types.Int64Value(r.Phase2DhGroupNumber)
	d.Phase2IntegrityAlgorithm = types.StringValue(r.Phase2IntegrityAlgorithm)
	d.Phase2EncryptionAlgorithm = types.StringValue(r.Phase2EncryptionAlgorithm)
	kawaiiIPsecModelToIngressRules(ctx, r, d)
}

//////////////////////////////
// Terraform CRUD Functions //
//////////////////////////////

func (r *KawaiiIPsecConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *KawaiiIPsecConnectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Create(ctx, DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	// find parent project
	kawaiiId, err := getKawaiiID(ctx, r.Data, data.KawaiiID.ValueString())
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	// create a new Kawaii IPsec Connection
	m := kawaiiIPsecResourceModel(&ctx, data)
	kawaiiIpSec, _, err := r.Data.K.KawaiiAPI.CreateKawaiiIpSec(ctx, kawaiiId).KawaiiIpSec(m).Execute()
	if err != nil {
		errorCreateGeneric(resp, err)
		return
	}
	data.ID = types.StringPointerValue(kawaiiIpSec.Id)
	kawaiiIPsecModelToResource(&ctx, kawaiiIpSec, data) // read back resulting object
	tflog.Trace(ctx, "created Kawaii IPsec Tunnel resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KawaiiIPsecConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *KawaiiIPsecConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Read(ctx, DefaultReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	kawaiiIpSec, _, err := r.Data.K.KawaiiAPI.ReadKawaiiIpSec(ctx, data.KawaiiID.ValueString(), data.ID.ValueString()).Execute()
	if err != nil {
		errorReadGeneric(resp, err)
		return
	}

	kawaiiIPsecModelToResource(&ctx, kawaiiIpSec, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KawaiiIPsecConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *KawaiiIPsecConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Update(ctx, DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	m := kawaiiIPsecResourceModel(&ctx, data)
	_, _, err := r.Data.K.KawaiiAPI.UpdateKawaiiIpSec(ctx, data.KawaiiID.ValueString(), data.ID.ValueString()).KawaiiIpSec(m).Execute()
	if err != nil {
		errorUpdateGeneric(resp, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KawaiiIPsecConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *KawaiiIPsecConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Delete(ctx, DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.Data.Mutex.Lock()
	defer r.Data.Mutex.Unlock()

	_, err := r.Data.K.KawaiiAPI.DeleteKawaiiIpSec(ctx, data.KawaiiID.ValueString(), data.ID.ValueString()).Execute()
	if err != nil {
		errorDeleteGeneric(resp, err)
		return
	}
	tflog.Trace(ctx, "Deleted "+data.ID.ValueString())
}
