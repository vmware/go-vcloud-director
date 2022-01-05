package govcd

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/hashicorp/go-version"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// endpointMinApiVersions holds mapping of OpenAPI endpoints and API versions they were introduced in.
var endpointMinApiVersions = map[string]string{
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRights:                              "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles:                       "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsCategories:                    "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles:                               "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles:                         "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles + types.OpenApiEndpointRights: "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAuditTrail:                          "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTier0Routers:              "32.0",
	// OpenApiEndpointExternalNetworks endpoint support was introduced with version 32.0 however it was still not stable
	// enough to be used. (i.e. it did not support update "PUT")
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks:                   "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies:                 "32.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcAssignedComputePolicies:         "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSessionCurrent:                     "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeClusters:                       "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways:                       "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups:                     "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules:                       "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtFirewallRules:                  "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks:                     "32.0", // VCD 9.7+ for NSX-V, 10.1+ for NSX-T
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp:                 "32.0", // VCD 9.7+ for NSX-V, 10.1+ for NSX-T
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcCapabilities:                    "32.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles:                    "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnTunnel:                     "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnTunnelConnectionProperties: "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnTunnelStatus:               "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroups:                          "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsCandidateVdcs:             "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwPolicies:               "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwDefaultPolicies:        "35.0", // VCD 10.2+

	// NSX-T ALB (Advanced/AVI Load Balancer) support was introduced in 10.2
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbController:                    "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbImportableClouds:              "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbImportableServiceEngineGroups: "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud:                         "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroups:           "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbEdgeGateway:                   "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments: "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools:                         "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPoolSummaries:                 "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServices:               "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbVirtualServiceSummaries:       "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibrary:            "35.0", // VCD 10.2+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointSSLCertificateLibraryOld:         "35.0", // VCD 10.2+ and deprecated from 10.3
}

// elevateNsxtNatRuleApiVersion helps to elevate API version to consume newer NSX-T NAT Rule features
// API V35.2+ supports new fields FirewallMatch and Priority
// API V36.0+ supports new RuleType - REFLEXIVE

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
	},
}

// checkOpenApiEndpointCompatibility checks if VCD version (to which the client is connected) is sufficient to work with
// specified OpenAPI endpoint and returns either an error or the Api version to use for calling that endpoint. This Api
// version can then be supplied to low level OpenAPI client functions.
// If the system default API version is higher than endpoint introduction version - default system one is used.
func (client *Client) checkOpenApiEndpointCompatibility(endpoint string) (string, error) {
	minimumApiVersion, ok := endpointMinApiVersions[endpoint]
	if !ok {
		return "", fmt.Errorf("minimum API version for endopoint '%s' is not defined", endpoint)
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

// getOpenApiHighestElevatedVersion returns highest supported API version for particular endpoint
// These API versions must be defined in endpointElevatedApiVersions. If none are there - it will return minimum
// supported API versions just like client.checkOpenApiEndpointCompatibility().
//
// The advantage of this functions is that it provides a controlled API elevation instead of just picking the highest
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
		if client.APIVCDMaxVersionIs(fmt.Sprintf(">= %s", elevatedVersion.Original())) {
			util.Logger.Printf("[DEBUG] Elevated version '%s' is supported by VCD instance for endpoint '%s'",
				elevatedVersion.Original(), endpoint)
			// highest version found - store it and abort the loop
			supportedElevatedVersion = elevatedVersion.Original()
			break
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
