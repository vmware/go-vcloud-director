/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// AnyTypeEdgeGateway is a common structure which fetches any type of Edge Gateway (NSX-T or NSX-V)
// using OpenAPI endpoint. It can be useful to identify type of Edge Gateway or just retrieve common
// fields - like OwnerRef. There is also a function GetNsxtEdgeGateway to convert it to
// NsxtEdgeGateway (if it is an NSX-T one)
type AnyTypeEdgeGateway struct {
	EdgeGateway *types.OpenAPIEdgeGateway
	client      *Client
}

// GetNsxtEdgeGatewayById allows retrieving NSX-T or NSX-V Edge Gateway by ID for Org admins
func (adminOrg *AdminOrg) GetAnyTypeEdgeGatewayById(id string) (*AnyTypeEdgeGateway, error) {
	return getAnyTypeApiEdgeGatewayById(adminOrg.client, id, nil)
}

// GetNsxtEdgeGatewayById allows retrieving NSX-T or NSX-V Edge Gateway by ID for Org users
func (org *Org) GetAnyTypeEdgeGatewayById(id string) (*AnyTypeEdgeGateway, error) {
	return getAnyTypeApiEdgeGatewayById(org.client, id, nil)
}

// getNsxtEdgeGatewayById is a private parent for wrapped functions:
// func (adminOrg *AdminOrg) GetAnyTypeEdgeGatewayById(id string) (*AnyTypeEdgeGateway, error)
// func (org *Org) GetAnyTypeEdgeGatewayById(id string) (*AnyTypeEdgeGateway, error)
func getAnyTypeApiEdgeGatewayById(client *Client, id string, queryParameters url.Values) (*AnyTypeEdgeGateway, error) {
	if id == "" {
		return nil, fmt.Errorf("empty Edge Gateway ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	egw := &AnyTypeEdgeGateway{
		EdgeGateway: &types.OpenAPIEdgeGateway{},
		client:      client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, queryParameters, egw.EdgeGateway, nil)
	if err != nil {
		return nil, err
	}
	return egw, nil
}

// IsNsxt checks if Edge Gateways is NSX-T backed
func (anyTypeGateway *AnyTypeEdgeGateway) IsNsxt() bool {
	if anyTypeGateway != nil && anyTypeGateway.EdgeGateway != nil && anyTypeGateway.EdgeGateway.GatewayBacking != nil {
		return anyTypeGateway.EdgeGateway.GatewayBacking.GatewayType == "NSXT_BACKED"
	}
	return false
}

// IsNsxv checks if Edge Gateways is NSX-V backed
func (anyTypeGateway *AnyTypeEdgeGateway) IsNsxv() bool {
	return !anyTypeGateway.IsNsxt()
}

// GetNsxtEdgeGateway converts `AnyTypeEdgeGateway` to `NsxtEdgeGateway` if it is an NSX-T one
func (anyTypeGateway *AnyTypeEdgeGateway) GetNsxtEdgeGateway() (*NsxtEdgeGateway, error) {
	if !anyTypeGateway.IsNsxt() {
		return nil, fmt.Errorf("this is not an NSX-T backed Edge Gateway")
	}

	nsxtEdgeGateway := &NsxtEdgeGateway{
		EdgeGateway: anyTypeGateway.EdgeGateway,
		client:      anyTypeGateway.client,
	}

	return nsxtEdgeGateway, nil
}
