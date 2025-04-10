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
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	KeyAccessType                 = "access_type"
	KeyAdapters                   = "adapters"
	KeyAddress                    = "address"
	KeyAddresses                  = "addresses"
	KeyAgents                     = "agents"
	KeyApp                        = "app"
	KeyApplication                = "application"
	KeyAssign                     = "assign"
	KeyBackendIPs                 = "backend_ips"
	KeyBackendPort                = "backend_port"
	KeyBackends                   = "backends"
	KeyBootstrapPubkey            = "bootstrap_pubkey"
	KeyBootstrapUser              = "bootstrap_user"
	KeyBot                        = "bot"
	KeyCIDR                       = "cidr"
	KeyCpuOvercommit              = "cpu_overcommit"
	KeyCpuPrice                   = "cpu_price"
	KeyCurrency                   = "currency"
	KeyDefault                    = "default"
	KeyDesc                       = "desc"
	KeyDestination                = "destination"
	KeyDisk                       = "disk"
	KeyDNS                        = "dns"
	KeyDomain                     = "domain"
	KeyEgressPolicy               = "egress_policy"
	KeyEgressRules                = "egress_rules"
	KeyEmail                      = "email"
	KeyEndpoint                   = "endpoint"
	KeyEndpoints                  = "endpoints"
	KeyExtraDisk                  = "extra_disk"
	KeyFailover                   = "failover"
	KeyFirst                      = "first"
	KeyFS                         = "fs"
	KeyGateway                    = "gateway"
	KeyGwPool                     = "gw_pool"
	KeyID                         = "id"
	KeyIngressRules               = "ingress_rules"
	KeyInterface                  = "interface"
	KeyIP                         = "ip"
	KeyIPsecConnections           = "ipsec_connections"
	KeyIPsecDpdAction             = "dpd_action"
	KeyIPsecDpdTimeout            = "dpd_timeout"
	KeyIPsecP1DHGroupNumber       = "phase1_dh_group_number"
	KeyIPsecP1EncryptionAlgorithm = "phase1_encryption_algorithm"
	KeyIPsecP1IntegrityAlgorithm  = "phase1_integrity_algorithm"
	KeyIPsecP1Lifetime            = "phase1_lifetime"
	KeyIPsecP2DHGroupNumber       = "phase2_dh_group_number"
	KeyIPsecP2EncryptionAlgorithm = "phase2_encryption_algorithm"
	KeyIPsecP2IntegrityAlgorithm  = "phase2_integrity_algorithm"
	KeyIPsecP2Lifetime            = "phase2_lifetime"
	KeyIPsecRekeyTime             = "rekey"
	KeyIPsecStartAction           = "start_action"
	KeyKawaii                     = "kawaii"
	KeyLast                       = "last"
	KeyMAC                        = "hwaddress"
	KeyMaxInstances               = "max_instances"
	KeyMaxMemory                  = "max_memory"
	KeyMaxStorage                 = "max_storage"
	KeyMaxVCPUs                   = "max_vcpus"
	KeyMemory                     = "mem"
	KeyMemoryOvercommit           = "memory_overcommit"
	KeyMemoryPrice                = "memory_price"
	KeyMetadata                   = "metadata"
	KeyName                       = "name"
	KeyNatRules                   = "nat_rules"
	KeyNetmaskBitSize             = "netmask_bitsize"
	KeyNetmask                    = "netmask"
	KeyNetworkConfig              = "netcfg"
	KeyNfs                        = "nfs"
	KeyNotifications              = "notifications"
	KeyNotify                     = "notify"
	KeyOS                         = "os"
	KeyOwner                      = "owner"
	KeyPolicy                     = "policy"
	KeyPool                       = "pool"
	KeyPort                       = "port"
	KeyPorts                      = "ports"
	KeyPreSharedKey               = "pre_shared_key"
	KeyPrice                      = "price"
	KeyPrivateIP                  = "private_ip"
	KeyPrivateIPs                 = "private_ips"
	KeyPrivate                    = "private"
	KeyPrivateSubnets             = "private_subnets"
	KeyProject                    = "project"
	KeyProtocol                   = "protocol"
	KeyProtocols                  = "protocols"
	KeyPublicIP                   = "public_ip"
	KeyPublicIPs                  = "public_ips"
	KeyPublic                     = "public"
	KeyRegion                     = "region"
	KeyRegions                    = "regions"
	KeyRemotePeer                 = "remote_peer"
	KeyRemoteSubnet               = "remote_subnet"
	KeyReserved                   = "reserved"
	KeyResizable                  = "resizable"
	KeyRole                       = "role"
	KeyRootPassword               = "root_password"
	KeyRoutes                     = "routes"
	KeySecret                     = "secret"
	KeySize                       = "size"
	KeySource                     = "source"
	KeySubnetSize                 = "subnet_size"
	KeySubnet                     = "subnet"
	KeyTags                       = "tags"
	KeyTeams                      = "teams"
	KeyTemplate                   = "template"
	KeyTimeouts                   = "timeouts"
	KeyToken                      = "token"
	KeyType                       = "type"
	KeyURI                        = "uri"
	KeyUsers                      = "users"
	KeyVCPUs                      = "vcpus"
	KeyVLAN                       = "vlan"
	KeyVNet                       = "vnet"
	KeyVolumes                    = "volumes"
	KeyVpcPeerings                = "vpc_peerings"
	KeyVRIDs                      = "vrids"
	KeyZone                       = "zone"
	KeyZones                      = "zones"
)

const (
	HelperGbToBytes = 1073741824
)

const (
	DefaultCreateTimeout = 30 * time.Minute // large enough for template upload
	DefaultDeleteTimeout = 5 * time.Minute
	DefaultReadTimeout   = 2 * time.Minute
	DefaultUpdateTimeout = 5 * time.Minute
)

const (
	ErrorGeneric              = "Kowabunga Error"
	ErrorUnconfiguredResource = "Unexpected Resource Configure Type"
	ErrorExpectedProviderData = "Expected *KowabungaProviderData, got: %T."
	ErrorUnknownKaktus        = "Unknown kaktus node"
	ErrorUnknownKawaii        = "Unknown kawaii instance"
	ErrorUnknownNfs           = "Unknown NFS storage"
	ErrorUnknownProject       = "Unknown project"
	ErrorUnknownRegion        = "Unknown region"
	ErrorUnknownPool          = "Unknown storage pool"
	ErrorUnknownSubnet        = "Unknown subnet"
	ErrorUnknownVNet          = "Unknown virtual network"
	ErrorUnknownTemplate      = "Unknown volume template"
	ErrorUnknownZone          = "Unknown zone"
)

const (
	ResourceIdDescription   = "Resource object internal identifier"
	ResourceNameDescription = "Resource name"
	ResourceDescDescription = "Resource extended description"
)

type ResourceBaseModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func errorUnconfiguredResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.AddError(
		ErrorUnconfiguredResource,
		fmt.Sprintf(ErrorExpectedProviderData, req.ProviderData),
	)
}

func errorCreateGeneric(resp *resource.CreateResponse, err error) {
	resp.Diagnostics.AddError(ErrorGeneric, err.Error())
}

func errorReadGeneric(resp *resource.ReadResponse, err error) {
	resp.Diagnostics.AddError(ErrorGeneric, err.Error())
}

func errorUpdateGeneric(resp *resource.UpdateResponse, err error) {
	resp.Diagnostics.AddError(ErrorGeneric, err.Error())
}

func errorDeleteGeneric(resp *resource.DeleteResponse, err error) {
	resp.Diagnostics.AddError(ErrorGeneric, err.Error())
}

func resourceAttributes(ctx *context.Context) map[string]schema.Attribute {
	defaultAttr := map[string]schema.Attribute{
		KeyName: schema.StringAttribute{
			MarkdownDescription: ResourceNameDescription,
			Required:            true,
		},
	}
	maps.Copy(defaultAttr, resourceAttributesWithoutName(ctx))

	return defaultAttr
}

func resourceAttributesWithoutName(ctx *context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		KeyID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: ResourceIdDescription,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		KeyDesc: schema.StringAttribute{
			MarkdownDescription: ResourceDescDescription,
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		KeyTimeouts: timeouts.Attributes(*ctx, timeouts.Opts{
			Create:            true,
			Read:              true,
			Update:            true,
			Delete:            true,
			CreateDescription: DefaultCreateTimeout.String(),
			ReadDescription:   DefaultReadTimeout.String(),
			UpdateDescription: DefaultUpdateTimeout.String(),
			DeleteDescription: DefaultDeleteTimeout.String(),
		}),
	}
}

func resourceMetadata(req resource.MetadataRequest, resp *resource.MetadataResponse, name string) {
	resp.TypeName = req.ProviderTypeName + "_" + name
}

func resourceImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(KeyID), req, resp)
}

func resourceConfigure(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *KowabungaProviderData {
	if req.ProviderData == nil {
		return nil
	}

	kd, ok := req.ProviderData.(*KowabungaProviderData)
	if !ok {
		errorUnconfiguredResource(req, resp)
		return nil
	}

	return kd
}

func getRegionID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	// let's suppose param is a proper region ID
	region, _, err := data.K.RegionAPI.ReadRegion(ctx, id).Execute()
	if err == nil {
		return *region.Id, nil
	}

	// fall back, it may be a region name then, finds its associated ID
	regions, _, err := data.K.RegionAPI.ListRegions(ctx).Execute()
	if err == nil {
		for _, rg := range regions {
			r, _, err := data.K.RegionAPI.ReadRegion(ctx, rg).Execute()
			if err == nil && r.Name == id {
				return *r.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownRegion)
}

func getZoneID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	// let's suppose param is a proper zone ID
	zone, _, err := data.K.ZoneAPI.ReadZone(ctx, id).Execute()
	if err == nil {
		return *zone.Id, nil
	}

	// fall back, it may be a zone name then, finds its associated ID
	zones, _, err := data.K.ZoneAPI.ListZones(ctx).Execute()
	if err == nil {
		for _, zn := range zones {
			z, _, err := data.K.ZoneAPI.ReadZone(ctx, zn).Execute()
			if err == nil && z.Name == id {
				return *z.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownZone)
}

func getVNetID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	// let's suppose param is a proper virtual network ID
	vnet, _, err := data.K.VnetAPI.ReadVNet(ctx, id).Execute()
	if err == nil {
		return *vnet.Id, nil
	}

	// fall back, it may be a virtual network name then, finds its associated ID
	vnets, _, err := data.K.VnetAPI.ListVNets(ctx).Execute()
	if err == nil {
		for _, vn := range vnets {
			v, _, err := data.K.VnetAPI.ReadVNet(ctx, vn).Execute()
			if err == nil && v.Name == id {
				return *v.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownVNet)
}

func getSubnetID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	// let's suppose param is a proper subnet ID
	subnet, _, err := data.K.SubnetAPI.ReadSubnet(ctx, id).Execute()
	if err == nil {
		return *subnet.Id, nil
	}

	// fall back, it may be a subnet name then, finds its associated ID
	subnets, _, err := data.K.SubnetAPI.ListSubnets(ctx).Execute()
	if err == nil {
		for _, sn := range subnets {
			s, _, err := data.K.SubnetAPI.ReadSubnet(ctx, sn).Execute()
			if err == nil && s.Name == id {
				return *s.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownSubnet)
}

func getProjectID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	// let's suppose param is a proper project ID
	project, _, err := data.K.ProjectAPI.ReadProject(ctx, id).Execute()
	if err == nil {
		return *project.Id, nil
	}

	// fall back, it may be a project name then, finds its associated ID
	projects, _, err := data.K.ProjectAPI.ListProjects(ctx).Execute()
	if err == nil {
		for _, pn := range projects {
			prj, _, err := data.K.ProjectAPI.ReadProject(ctx, pn).Execute()
			if err == nil && prj.Name == id {
				return *prj.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownProject)
}

func getPoolID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	// let's suppose param is a proper pool ID
	pool, _, err := data.K.PoolAPI.ReadStoragePool(ctx, id).Execute()
	if err == nil {
		return *pool.Id, nil
	}

	// fall back, it may be a pool name then, finds its associated ID
	pools, _, err := data.K.PoolAPI.ListStoragePools(ctx).Execute()
	if err == nil {
		for _, pn := range pools {
			pl, _, err := data.K.PoolAPI.ReadStoragePool(ctx, pn).Execute()
			if err == nil && pl.Name == id {
				return *pl.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownPool)
}

func getNfsID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	// let's suppose param is a proper NFS storage ID
	nfs, _, err := data.K.NfsAPI.ReadStorageNFS(ctx, id).Execute()
	if err == nil {
		return *nfs.Id, nil
	}

	// fall back, it may be a NFS storage name then, finds its associated ID
	storages, _, err := data.K.NfsAPI.ListStorageNFSs(ctx).Execute()
	if err == nil {
		for _, s := range storages {
			ns, _, err := data.K.NfsAPI.ReadStorageNFS(ctx, s).Execute()
			if err == nil && ns.Name == id {
				return *ns.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownNfs)
}

func getTemplateID(ctx context.Context, data *KowabungaProviderData, id, poolId string) (string, error) {
	// let's suppose param is a proper template ID
	template, _, err := data.K.TemplateAPI.ReadTemplate(ctx, id).Execute()
	if err == nil {
		return *template.Id, nil
	}

	// fall back, it may be a template name then, finds its associated ID from pool's templates
	templates, _, err := data.K.PoolAPI.ListStoragePoolTemplates(ctx, poolId).Execute()
	if err == nil {
		for _, tn := range templates {
			t, _, err := data.K.TemplateAPI.ReadTemplate(ctx, tn).Execute()
			if err == nil && t.Name == id {
				return *t.Id, nil
			}
		}
	}

	return "", fmt.Errorf("%s", ErrorUnknownTemplate)
}

func getKawaiiID(ctx context.Context, data *KowabungaProviderData, id string) (string, error) {
	kawaii, _, err := data.K.KawaiiAPI.ReadKawaii(ctx, id).Execute()
	if err == nil {
		return *kawaii.Id, nil
	}

	// fall back, it may be a Kawaii name then, finds its associated ID
	kawaiis, _, err := data.K.KawaiiAPI.ListKawaiis(ctx).Execute()
	if err == nil {
		for _, kw := range kawaiis {
			t, _, err := data.K.KawaiiAPI.ReadKawaii(ctx, kw).Execute()
			if err == nil && *t.Name == id {
				return *t.Id, nil
			}
		}
	}
	return "", fmt.Errorf("%s", ErrorUnknownKawaii)
}
