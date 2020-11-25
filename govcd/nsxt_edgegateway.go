/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtEdgeGateway uses OpenAPI endpoint to operate NSX-T Edge Gateways
type NsxtEdgeGateway struct {
	EdgeGateway *types.OpenAPIEdgeGateway
	client      *Client
}

// GetNsxtEdgeGatewayById allows to retrieve NSX-T edge gateway by ID for Org admins
func (adminOrg *AdminOrg) GetNsxtEdgeGatewayById(id string) (*NsxtEdgeGateway, error) {
	return getNsxtEdgeGatewayById(adminOrg.client, id)
}

// GetNsxtEdgeGatewayById allows to retrieve NSX-T edge gateway by ID for Org users
func (org *Org) GetNsxtEdgeGatewayById(id string) (*NsxtEdgeGateway, error) {
	return getNsxtEdgeGatewayById(org.client, id)
}

// GetNsxtEdgeGatewayByName allows to retrieve NSX-T edge gateway by Name for Org admins
func (adminOrg *AdminOrg) GetNsxtEdgeGatewayByName(name string) (*NsxtEdgeGateway, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	allEdges, err := adminOrg.GetAllNsxtEdgeGateways(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve edge gateway by name '%s': %s", name, err)
	}

	return returnSingleNsxtEdgeGateway(name, allEdges)
}

// GetNsxtEdgeGatewayByName allows to retrieve NSX-T edge gateway by Name for Org admins
func (org *Org) GetNsxtEdgeGatewayByName(name string) (*NsxtEdgeGateway, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	allEdges, err := org.GetAllNsxtEdgeGateways(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve edge gateway by name '%s': %s", name, err)
	}

	return returnSingleNsxtEdgeGateway(name, allEdges)
}

// GetAllNsxtEdgeGateways allows to retrieve all NSX-T edge gateways for Org Admins
func (adminOrg *AdminOrg) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	return getAllNsxtEdgeGateways(adminOrg.client, queryParameters)
}

// GetAllNsxtEdgeGateways  allows to retrieve all NSX-T edge gateways for Org users
func (org *Org) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	return getAllNsxtEdgeGateways(org.client, queryParameters)
}

// CreateNsxtEdgeGateway allows to create NSX-T edge gateway for Org admins
func (adminOrg *AdminOrg) CreateNsxtEdgeGateway(e *types.OpenAPIEdgeGateway) (*NsxtEdgeGateway, error) {
	if !adminOrg.client.IsSysAdmin {
		return nil, fmt.Errorf("only Provider can create Edge Gateway")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnEgw := &NsxtEdgeGateway{
		EdgeGateway: &types.OpenAPIEdgeGateway{},
		client:      adminOrg.client,
	}

	err = adminOrg.client.OpenApiPostItem(minimumApiVersion, urlRef, nil, e, returnEgw.EdgeGateway)
	if err != nil {
		return nil, fmt.Errorf("error creating Edge Gateway: %s", err)
	}

	return returnEgw, nil
}

// Update allows to update NSX-T edge gateway for Org admins
func (egw *NsxtEdgeGateway) Update(e *types.OpenAPIEdgeGateway) (*NsxtEdgeGateway, error) {
	if !egw.client.IsSysAdmin {
		return nil, fmt.Errorf("only Provider can update Edge Gateway")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := egw.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if e.ID == "" {
		return nil, fmt.Errorf("cannot update Edge Gateway without id")
	}

	urlRef, err := egw.client.OpenApiBuildEndpoint(endpoint, e.ID)
	if err != nil {
		return nil, err
	}

	returnEgw := &NsxtEdgeGateway{
		EdgeGateway: &types.OpenAPIEdgeGateway{},
		client:      egw.client,
	}

	err = egw.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, e, returnEgw.EdgeGateway)
	if err != nil {
		return nil, fmt.Errorf("error updating Edge Gateway: %s", err)
	}

	return returnEgw, nil
}

// Update allows to delete NSX-T edge gateway for Org admins
func (egw *NsxtEdgeGateway) Delete() error {
	if !egw.client.IsSysAdmin {
		return fmt.Errorf("only Provider can update Edge Gateway")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := egw.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if egw.EdgeGateway.ID == "" {
		return fmt.Errorf("cannot delete Edge Gateway without id")
	}

	urlRef, err := egw.client.OpenApiBuildEndpoint(endpoint, egw.EdgeGateway.ID)
	if err != nil {
		return err
	}

	err = egw.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil)

	if err != nil {
		return fmt.Errorf("error deleting Edge Gateway: %s", err)
	}

	return nil
}

// getNsxtEdgeGatewayById is a private parent for wrapped functions:
// func (adminOrg *AdminOrg) GetNsxtEdgeGatewayByName(id string) (*NsxtEdgeGateway, error)
// func (org *Org) GetNsxtEdgeGatewayByName(id string) (*NsxtEdgeGateway, error)
func getNsxtEdgeGatewayById(client *Client, id string) (*NsxtEdgeGateway, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty Edge Gateway id")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	egw := &NsxtEdgeGateway{
		EdgeGateway: &types.OpenAPIEdgeGateway{},
		client:      client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, egw.EdgeGateway)
	if err != nil {
		return nil, err
	}

	return egw, nil
}

// returnSingleNsxtEdgeGateway helps to reduce code duplication for `GetNsxtEdgeGatewayByName` functions with different
// receivers
func returnSingleNsxtEdgeGateway(name string, allEdges []*NsxtEdgeGateway) (*NsxtEdgeGateway, error) {
	if len(allEdges) > 1 {
		return nil, fmt.Errorf("got more than 1 edge gateway by name '%s' %d", name, len(allEdges))
	}

	if len(allEdges) < 1 {
		return nil, fmt.Errorf("%s: got 0 edge gateways by name '%s'", ErrorEntityNotFound, name)
	}

	return allEdges[0], nil
}

// getAllNsxtEdgeGateways is a private parent for wrapped functions:
// func (adminOrg *AdminOrg) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error)
// func (org *Org) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error)
func getAllNsxtEdgeGateways(client *Client, queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.OpenAPIEdgeGateway{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into Role types with client
	wrappedResponses := make([]*NsxtEdgeGateway, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtEdgeGateway{
			EdgeGateway: typeResponses[sliceIndex],
			client:      client,
		}
	}

	return wrappedResponses, nil
}
