/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// AnyEdgeGateway is a common structure which fetches any type of Edge Gateway (NSX-T or NSX-V)
// using OpenAPI endpoint. It can be useful to identify type of Edge Gateway or just retrieve common
// fields - like OwnerRef. There is also a function GetNsxtEdgeGateway to convert it to
// NsxtEdgeGateway (if it is an NSX-T one)
type AnyEdgeGateway struct {
	EdgeGateway *types.OpenAPIEdgeGateway
	client      *Client
}

// GetNsxtEdgeGatewayById allows retrieving NSX-T edge gateway by ID for Org admins
func (adminOrg *AdminOrg) GetAnyEdgeGatewayById(id string) (*AnyEdgeGateway, error) {
	return getAnyApiEdgeGatewayById(adminOrg.client, id, nil)
}

// GetNsxtEdgeGatewayById allows retrieving NSX-T edge gateway by ID for Org users
func (org *Org) GetAnyEdgeGatewayById(id string) (*AnyEdgeGateway, error) {
	return getAnyApiEdgeGatewayById(org.client, id, nil)
}

// getNsxtEdgeGatewayById is a private parent for wrapped functions:
// func (adminOrg *AdminOrg) GetNsxtEdgeGatewayByName(id string) (*NsxtEdgeGateway, error)
// func (org *Org) GetNsxtEdgeGatewayByName(id string) (*NsxtEdgeGateway, error)
func getAnyApiEdgeGatewayById(client *Client, id string, queryParameters url.Values) (*AnyEdgeGateway, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty Edge Gateway ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	egw := &AnyEdgeGateway{
		EdgeGateway: &types.OpenAPIEdgeGateway{},
		client:      client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, queryParameters, egw.EdgeGateway, nil)
	if err != nil {
		return nil, err
	}
	return egw, nil
}

// IsNsxv checks if Edge Gateways is NSX-T backed
func (anyGateway *AnyEdgeGateway) IsNsxt() bool {
	if anyGateway != nil && anyGateway.EdgeGateway != nil && anyGateway.EdgeGateway.GatewayBacking != nil {
		return anyGateway.EdgeGateway.GatewayBacking.GatewayType == "NSXT_BACKED"
	}
	return false
}

// IsNsxv checks if Edge Gateways is NSX-V backed
func (anyGateway *AnyEdgeGateway) IsNsxv() bool {
	return !anyGateway.IsNsxt()
}

// GetNsxtEdgeGateway converts `AnyEdgeGateway` to `NsxtEdgeGateway` if it is an NSX-T one
func (anyGateway *AnyEdgeGateway) GetNsxtEdgeGateway() (*NsxtEdgeGateway, error) {
	if !anyGateway.IsNsxt() {
		return nil, fmt.Errorf("this is not an NSX-T backed Edge Gateway")
	}

	nsxtEdgeGateway := &NsxtEdgeGateway{
		EdgeGateway: anyGateway.EdgeGateway,
		client:      anyGateway.client,
	}

	return nsxtEdgeGateway, nil
}
