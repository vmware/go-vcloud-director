/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetNsxtRouteAdvertisement retrieves the list of subnets that will be advertised so that the Edge Gateway can route
// out to the connected external network.
func (egw *NsxtEdgeGateway) GetNsxtRouteAdvertisement() (*types.RouteAdvertisement, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtRouteAdvertisement

	highestApiVersion, err := egw.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := egw.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	routeAdvertisement := &types.RouteAdvertisement{}
	err = egw.client.OpenApiGetItem(highestApiVersion, urlRef, nil, routeAdvertisement, nil)

	return nil, nil
}

// UpdateNsxtRouteAdvertisement updates the list of subnets that will be advertised so that the Edge Gateway can route
// out to the connected external network.
func (egw *NsxtEdgeGateway) UpdateNsxtRouteAdvertisement(enable bool, subnets []string) (*types.RouteAdvertisement, error) {
	return nil, nil
}

// DeleteNsxtRouteAdvertisement deletes the list of subnets that will be advertised.
func (egw *NsxtEdgeGateway) DeleteNsxtRouteAdvertisement() error {
	return nil
}

// checkSanityNsxtEdgeGatewayRouteAdvertisement
func checkSanityNsxtEdgeGatewayRouteAdvertisement(vdc *Vdc) error {}
