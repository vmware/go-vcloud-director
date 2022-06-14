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
func (egw *NsxtEdgeGateway) GetNsxtRouteAdvertisement(useTenantContext bool) (*types.RouteAdvertisement, error) {
	err := checkSanityNsxtEdgeGatewayRouteAdvertisement(egw)
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtRouteAdvertisement

	highestApiVersion, err := egw.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := egw.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	var tenantContextHeaders map[string]string
	if useTenantContext {
		tenantContext, err := egw.getTenantContext()
		if err != nil {
			return nil, err
		}

		tenantContextHeaders = getTenantContextHeader(tenantContext)
	}

	routeAdvertisement := &types.RouteAdvertisement{}
	err = egw.client.OpenApiGetItem(highestApiVersion, urlRef, nil, routeAdvertisement, tenantContextHeaders)
	if err != nil {
		return nil, err
	}

	return routeAdvertisement, nil
}

// GetNsxtRouteAdvertisementWithContext method is the same as GetNsxtRouteAdvertisement but using TenantContext by default
func (egw *NsxtEdgeGateway) GetNsxtRouteAdvertisementWithContext() (*types.RouteAdvertisement, error) {
	return egw.GetNsxtRouteAdvertisement(true)
}

// UpdateNsxtRouteAdvertisement updates the list of subnets that will be advertised so that the Edge Gateway can route
// out to the connected external network.
func (egw *NsxtEdgeGateway) UpdateNsxtRouteAdvertisement(enable bool, subnets []string, useTenantContext bool) (*types.RouteAdvertisement, error) {
	err := checkSanityNsxtEdgeGatewayRouteAdvertisement(egw)
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtRouteAdvertisement

	highestApiVersion, err := egw.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := egw.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	var tenantContextHeaders map[string]string
	if useTenantContext {
		tenantContext, err := egw.getTenantContext()
		if err != nil {
			return nil, err
		}

		tenantContextHeaders = getTenantContextHeader(tenantContext)
	}

	routeAdvertisement := &types.RouteAdvertisement{
		Enable:  enable,
		Subnets: subnets,
	}

	err = egw.client.OpenApiPutItem(highestApiVersion, urlRef, nil, routeAdvertisement, nil, tenantContextHeaders)
	if err != nil {
		return nil, err
	}

	return egw.GetNsxtRouteAdvertisement(useTenantContext)
}

// UpdateNsxtRouteAdvertisementWithContext method is the same as UpdateNsxtRouteAdvertisementWithContext but using TenantContext by default
func (egw *NsxtEdgeGateway) UpdateNsxtRouteAdvertisementWithContext(enable bool, subnets []string) (*types.RouteAdvertisement, error) {
	return egw.UpdateNsxtRouteAdvertisement(enable, subnets, true)
}

// DeleteNsxtRouteAdvertisement deletes the list of subnets that will be advertised.
func (egw *NsxtEdgeGateway) DeleteNsxtRouteAdvertisement(useTenantContext bool) error {
	_, err := egw.UpdateNsxtRouteAdvertisement(false, []string{}, useTenantContext)
	return err
}

// DeleteNsxtRouteAdvertisement method is the same as DeleteNsxtRouteAdvertisement but using TenantContext by default
func (egw *NsxtEdgeGateway) DeleteNsxtRouteAdvertisementWithContext() error {
	return egw.DeleteNsxtRouteAdvertisement(true)
}

// checkSanityNsxtEdgeGatewayRouteAdvertisement function performs some checks to *NsxtEdgeGateway parameter and returns error
// if something is wrong. It is useful with methods NsxtEdgeGateway.[Get/Update/Delete]NsxtRouteAdvertisement
func checkSanityNsxtEdgeGatewayRouteAdvertisement(egw *NsxtEdgeGateway) error {
	if egw.EdgeGateway == nil {
		return fmt.Errorf("the EdgeGateway pointer is nil. Please initialize it first before using this method")
	}

	if egw.EdgeGateway.ID == "" {
		return fmt.Errorf("the EdgeGateway ID is empty. Please initialize it first before using this method")
	}

	return nil
}
