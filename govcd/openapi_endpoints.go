// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/util"

	"github.com/hashicorp/go-version"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// endpointMinApiVersions holds mapping of OpenAPI endpoints and API versions they were introduced in.
var endpointMinApiVersions = map[string]string{
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRights:                              "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles:                       "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsCategories:                    "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles:                               "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles:                         "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles + types.OpenApiEndpointRights: "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointUsers:                               "40.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAuditTrail:                          "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTier0Routers:              "32.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableDvpgs:                     "36.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTestConnection:                      "34.0",
	// OpenApiEndpointExternalNetworks endpoint support was introduced with version 32.0 however it was still not stable
	// enough to be used. (i.e. it did not support update "PUT")
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks:           "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies:         "32.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcAssignedComputePolicies: "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSessionCurrent:             "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeClusters:               "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointQosProfiles:                "36.2", // VCD 10.3.2+ (NSX-T only)
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayQos:             "36.2", // VCD 10.3.2+ (NSX-T only)
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDhcpForwarder:   "36.1", // VCD 10.3.1+ (NSX-T only)
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDns:             "37.0", // VCD 10.4.0+ (NSX-T only)
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel:     "35.0", // VCD 10.2.0+ (NSX-T only)
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewaySlaacProfile:    "35.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayStaticRoutes:    "37.0", // VCD 10.4.0+ (NSX-T only)
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways:               "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayUsedIpAddresses: "34.0",

	// Static security groups and IP sets in VCD 10.2, Dynamic security groups in VCD 10.3+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups:                     "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules:                       "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtFirewallRules:                  "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks:                     "32.0", // VCD 9.7+ for NSX-V, 10.1+ for NSX-T
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworkSegmentProfiles:       "37.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp:                 "32.0", // VCD 9.7+ for NSX-V, 10.1+ for NSX-T
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings:         "36.1", // VCD 10.3.1+ (NSX-T only)
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcCapabilities:                    "32.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles:                    "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnTunnel:                     "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnTunnelConnectionProperties: "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnTunnelStatus:               "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups:                          "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsCandidateVdcs:             "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwPolicies:               "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwDefaultPolicies:        "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSecurityTags:                       "36.0", // VCD 10.3+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtRouteAdvertisement:             "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointLogicalVmGroups:                    "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces:                      "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors:              "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes:                     "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviors:                   "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviorAccessControls:      "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities:                        "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityAccessControls:            "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesTypes:                   "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesResolve:                 "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesBehaviorsInvocations:    "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalEndpoints:                  "38.1", // VCD 10.5.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointApiFilters:                         "38.1", // VCD 10.5.1+

	// IP Spaces
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces:                     "37.1", // VCD 10.4.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceSummaries:             "37.1", // VCD 10.4.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks:               "37.1", // VCD 10.4.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocate:       "37.1", // VCD 10.4.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceIpAllocations:         "37.1", // VCD 10.4.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceOrgAssignments:        "37.1", // VCD 10.4.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceFloatingIpSuggestions: "37.1", // VCD 10.4.1+

	// NSX-T ALB (Advanced/AVI Load Balancer) support was introduced in 10.2
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController:                      "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbImportableClouds:                "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbImportableServiceEngineGroups:   "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud:                           "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups:             "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbEdgeGateway:                     "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments:   "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools:                           "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPoolSummaries:                   "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServices:                 "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceSummaries:         "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpRequestRules:  "38.0", // VCD 10.5+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpResponseRules: "38.0", // VCD 10.5+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceHttpSecurityRules: "38.0", // VCD 10.5+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTrustedCertificates:                "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibrary:              "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibraryOld:           "35.0", // VCD 10.2+ and deprecated from 10.3
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules:                  "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkContextProfiles:             "35.0", // VCD 10.2+

	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpNeighbor:          "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfigPrefixLists: "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfig:            "35.0", // VCD 10.2+

	types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcAssignedComputePolicies: "35.0",
	types.OpenApiPathVersion2_0_0 + types.OpenApiEndpointVdcComputePolicies:         "35.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile:          "36.0", // VCD 10.3+

	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters:                           "36.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointResourcePools:                            "36.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointResourcePoolsBrowseAll:                   "36.2",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointResourcePoolHardware:                     "36.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools:                             "36.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPoolSummaries:                     "36.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointStorageProfiles:                          "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentProfileTemplates:              "36.2",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtGlobalDefaultSegmentProfileTemplates: "36.2",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentIpDiscoveryProfiles:           "36.2",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentMacDiscoveryProfiles:          "36.2",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSpoofGuardProfiles:            "36.2",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentQosProfiles:                   "36.2",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtSegmentSecurityProfiles:              "36.2",

	// Extensions API endpoints. These are not versioned
	types.OpenApiEndpointExtensionsUi:                    "35.0", // VCD 10.2+
	types.OpenApiEndpointExtensionsUiPlugin:              "35.0", // VCD 10.2+
	types.OpenApiEndpointExtensionsUiTenants:             "35.0", // VCD 10.2+
	types.OpenApiEndpointExtensionsUiTenantsPublishAll:   "35.0", // VCD 10.2+
	types.OpenApiEndpointExtensionsUiTenantsPublish:      "35.0", // VCD 10.2+
	types.OpenApiEndpointExtensionsUiTenantsUnpublishAll: "35.0", // VCD 10.2+
	types.OpenApiEndpointExtensionsUiTenantsUnpublish:    "35.0", // VCD 10.2+

	// Endpoints for managing tokens and service accounts
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTokens:                   "36.1", // VCD 10.3.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts:          "37.0", // VCD 10.4.0+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccountGrant:      "37.0", // VCD 10.4.0+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTransportZones: "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVCenterDistributedSwitch: "33.0",

	// Endpoint for managing vGPU profiles
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVgpuProfile: "36.2",

	// Orgs
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs: "37.0",

	// NSX-T Tier 0 router interfaces that can be used for IP Space uplink assignment
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtTier0RouterInterfaces: "38.0",

	// VCF
	types.OpenApiPathVcf + types.OpenApiEndpointRegionStoragePolicies:                  "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointStorageClasses:                         "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointContentLibraries:                       "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItems:                    "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointContentLibraryItemFiles:                "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointNsxManagers:                            "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointRegions:                                "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointSupervisors:                            "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointSupervisorZones:                        "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointZones:                                  "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs:                                 "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmVmClasses:                            "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmVdcsVmClasses:                        "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStoragePolicies:                   "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces:                             "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways:                     "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations:                  "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClusters:                         "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClusterTransportNodeStatus:       "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClustersSync:                     "40.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTmOrgNetworkingSettings:       "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmOrgSettings:                          "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings:           "40.0",
	types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettingsVpcProfile: "40.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFeatureFlags:                  "40.0",
}

// endpointElevatedApiVersions endpoint elevated API versions
var endpointElevatedApiVersions = map[string][]string{
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules: {
		//"34.0", // Basic minimum required version
		"35.2", // Introduces support for new fields FirewallMatch and Priority
		"36.0", // Adds support for new NAT Rule Type - REFLEXIVE (field Type must be used instead of RuleType)
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks: {
		//"33.0", // Basic minimum required version
		"35.0", // Deprecates field BackingType in favor of BackingTypeValue
		"36.0", // Adds support new type of BackingTypeValue - IMPORTED_T_LOGICAL_SWITCH (backed by NSX-T segment)
		"37.1", // Adds support for IP Spaces with new fields - UsingIpSpace, DedicatedOrg
		"38.1", // Adds support for NAT, Firewall and Route Advertisement intention configuration
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules: {
		//"35.0", // Basic minimum required version
		"35.2", // Deprecates Action field in favor of ActionValue
		"36.2", // Adds 3 new fields - Comments, SourceGroupsExcluded, and DestinationGroupsExcluded
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp: {
		//"32.0", // Basic minimum required version
		"36.1", // Adds support for dnsServers
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups: {
		//"34.0", // Basic minimum required version
		"36.0", // Adds support for Dynamic Security Groups by deprecating `Type` field in favor of `TypeValue`
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController: {
		//"35.0", // Basic minimum required version
		"37.0", // Deprecates LicenseType in favor of SupportedFeatureSet
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups: {
		//"35.0", // Basic minimum required version
		"37.0", // Adds SupportedFeatureSet
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbEdgeGateway: {
		//"35.0", // Basic minimum required version
		"37.0", // Deprecates LicenseType in favor of SupportedFeatureSet. Adds IPv6 service network definition support
		"37.1", // Adds support for Transparent Mode
	},
	//
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServices: {
		//"35.0", // Basic minimum required version
		"37.0", // Adds IPv6 Virtual Service Support
		"37.1", // Adds support for Transparent Mode
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceSummaries: {
		//"35.0", // Basic minimum required version
		"37.0", // Adds IPv6 Virtual Service Support
		"37.1", // Adds support for Transparent Mode
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcNetworkProfile: {
		//"36.0", // Introduced support
		"36.2", // 2 additional fields vappNetworkSegmentProfileTemplateRef and vdcNetworkSegmentProfileTemplateRef added
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes: {
		//"35.0", // Introduced support
		"37.1", // Added MaxImplicitRight property in DefinedEntityType
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities: {
		//"35.0", // Introduced support
		"37.0", // Added metadata support
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways: {
		//"35.0", // Introduced support
		"37.1", // Exposes computed field `UsingIpSpace` in `types.EdgeGatewayUplinks`
		"39.0", // Adds support for DISTRIBUTED_ONLY `deploymentMode`
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDns: {
		"37.0", // Introduced support
		"38.0", // New field SnatRuleExternalIpAddress
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinksAllocate: {
		//"37.1", // Introduced support
		"37.2", // Adds 'value' field
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces: {
		//"37.1", // Introduced support
		"38.0", // Adds 'DefaultGatewayServiceConfig' structure for firewall and NAT rule creation
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaceUplinks: {
		//"37.1", // Introduced support
		"38.0", // Adds 'Interfaces' structure for associating particular Tier-0 router interfaces
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs: {
		"40.0", // TM Orgs
	},
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTier0Routers: {
		"40.0", // TM Tier 0 Gateways
	},
}

// checkOpenApiEndpointCompatibility checks if VCD version (to which the client is connected) is sufficient to work with
// specified OpenAPI endpoint and returns either an error or the Api version to use for calling that endpoint. This Api
// version can then be supplied to low level OpenAPI client functions.
// If the system default API version is higher than endpoint introduction version - default system one is used.
func (client *Client) checkOpenApiEndpointCompatibility(endpoint string) (string, error) {
	minimumApiVersion, ok := endpointMinApiVersions[endpoint]
	if !ok {
		return "", fmt.Errorf("minimum API version for endpoint '%s' is not defined", endpoint)
	}

	if client.APIVCDMaxVersionIs("< " + minimumApiVersion) {
		maxSupportedVersion, err := client.MaxSupportedVersion()
		if err != nil {
			return "", fmt.Errorf("error reading maximum supported API version: %s", err)
		}
		return "", fmt.Errorf("endpoint '%s' requires API version to support at least '%s'. Maximum supported version in this instance: '%s'",
			endpoint, minimumApiVersion, maxSupportedVersion)
	}

	// If default API version is higher than minimum required API version for endpoint - use the system default one.
	if client.APIClientVersionIs("> " + minimumApiVersion) {
		return client.APIVersion, nil
	}

	return minimumApiVersion, nil
}

// getOpenApiHighestElevatedVersion returns the highest supported API version for particular endpoint
// These API versions must be defined in endpointElevatedApiVersions. If none are there - it will return minimum
// supported API versions just like client.checkOpenApiEndpointCompatibility().
//
// The advantage of this function is that it provides a controlled API elevation instead of just picking the highest
// which could be risky and untested (especially if new API version is released after release of package consuming this
// SDK)
func (client *Client) getOpenApiHighestElevatedVersion(endpoint string) (string, error) {
	util.Logger.Printf("[DEBUG] Checking if elevated API versions are defined for endpoint '%s'", endpoint)

	// At first get minimum API version and check if it can be supported
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return "", fmt.Errorf("error getting minimum required API version: %s", err)
	}

	// If no elevated versions are defined - return minimumApiVersion
	elevatedVersionSlice, elevatedVersionsDefined := endpointElevatedApiVersions[endpoint]
	if !elevatedVersionsDefined {
		util.Logger.Printf("[DEBUG] No elevated API versions are defined for endpoint '%s'. Using minimum '%s'",
			endpoint, minimumApiVersion)
		return minimumApiVersion, nil
	}

	util.Logger.Printf("[DEBUG] Found '%d' (%s) elevated API versions for endpoint '%s' ",
		len(elevatedVersionSlice), strings.Join(elevatedVersionSlice, ", "), endpoint)

	// Reverse sort (highest to lowest) slice of elevated API versions so that we can start by highest supported and go down
	versionsRaw := elevatedVersionSlice
	versions := make([]*version.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, err := version.NewVersion(raw)
		if err != nil {
			return "", fmt.Errorf("error evaluating version %s: %s", raw, err)
		}
		versions[i] = v
	}
	sort.Sort(sort.Reverse(version.Collection(versions)))

	var supportedElevatedVersion string
	// Loop highest to the lowest elevated versions and try to find highest from the list of supported ones
	for _, elevatedVersion := range versions {

		util.Logger.Printf("[DEBUG] Checking if elevated version '%s' is supported by VCD instance for endpoint '%s'",
			elevatedVersion.Original(), endpoint)
		// Check if maximum VCD API version supported is greater or equal to elevated version

		if client.APIVCDMaxVersionIs(fmt.Sprintf(">= %s", elevatedVersion.Original())) &&
			!client.APIClientVersionIs(fmt.Sprintf("> %s", elevatedVersion.Original())) {
			util.Logger.Printf("[DEBUG] Elevated version '%s' is supported by VCD instance for endpoint '%s'",
				elevatedVersion.Original(), endpoint)
			// highest version found - store it and abort the loop
			supportedElevatedVersion = elevatedVersion.Original()
			break
		} else {
			util.Logger.Printf("[DEBUG] Skipped Elevated version '%s' for endpoint '%s', Default minimum version '%s'",
				elevatedVersion.Original(), endpoint, client.APIVersion)
		}

		util.Logger.Printf("[DEBUG] API version '%s' is not supported by VCD instance for endpoint '%s'",
			elevatedVersion.Original(), endpoint)
	}

	if supportedElevatedVersion == "" {
		util.Logger.Printf("[DEBUG] No elevated API versions are supported for endpoint '%s'. Will use minimum "+
			"required version '%s'", endpoint, minimumApiVersion)
		return minimumApiVersion, nil
	}

	util.Logger.Printf("[DEBUG] Will use elevated version '%s for endpoint '%s'",
		supportedElevatedVersion, endpoint)
	return supportedElevatedVersion, nil
}
