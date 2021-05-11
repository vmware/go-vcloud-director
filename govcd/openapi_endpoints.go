package govcd

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// endpointMinApiVersions holds mapping of OpenAPI endpoints and API versions they were introduced in.
var endpointMinApiVersions = map[string]string{
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles:                  "31.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAuditTrail:             "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTier0Routers: "32.0",
	// OpenApiEndpointExternalNetworks endpoint support was introduced with version 32.0 however it was still not stable
	// enough to be used. (i.e. it did not support update "PUT")
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks:           "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcComputePolicies:         "32.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcAssignedComputePolicies: "33.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeClusters:               "34.0", // VCD 10.1+
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways:               "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups:             "34.0",
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks:             "32.0", // VCD 9.7+ for NSX-V, 10.1+ for NSX-T
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp:         "32.0", // VCD 9.7+ for NSX-V, 10.1+ for NSX-T
	types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcCapabilities:            "32.0",
}

// checkOpenApiEndpointCompatibility checks if VCD version (to which the client is connected) is sufficient to work with
// specified OpenAPI endpoint and returns either error, either Api version to use for calling that endpoint. This Api
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
