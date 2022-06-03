/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (egw *NsxtEdgeGateway) GetBgpConfiguration() (*types.EdgeBgpConfig, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfig
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path edgeGateways/%s/firewall/rules
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &types.EdgeBgpConfig{}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	return returnObject, nil
}

func (egw *NsxtEdgeGateway) UpdateBgpConfiguration(bgpConfig *types.EdgeBgpConfig) (*types.EdgeBgpConfig, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpConfig
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	// Update of BGP config requires version to be specified. This function automatically handles it.
	existingBgpConfig, err := egw.GetBgpConfiguration()
	if err != nil {
		return nil, fmt.Errorf("error getting NSX-T Edge Gateway BGP Configuration: %s", err)
	}
	bgpConfig.Version = existingBgpConfig.Version

	returnObject := &types.EdgeBgpConfig{}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, bgpConfig, returnObject, nil)
	if err != nil {
		return nil, fmt.Errorf("error setting NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	return returnObject, nil
}

func (egw *NsxtEdgeGateway) DisableBgpConfiguration() error {
	// Get existring BGP configuration so that when disabling it - other settings remain as they are
	bgpConfig, err := egw.GetBgpConfiguration()
	if err != nil {
		return fmt.Errorf("error retrieving BGP configuration: %s", err)
	}
	bgpConfig.Enabled = false

	_, err = egw.UpdateBgpConfiguration(bgpConfig)
	return err
}
