/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// EdgeBgpNeighbor represents NSX-T Edge Gateway BGP Neighbor
type EdgeBgpNeighbor struct {
	EdgeBgpNeighbor *types.EdgeBgpNeighbor
	client          *Client
	// edgeGatewayId is stored for usage in EdgeBgpNeighbor receiver functions
	edgeGatewayId string
}

// CreateBgpNeighbor creates BGP Neighbor with the given configuration
func (egw *NsxtEdgeGateway) CreateBgpNeighbor(bgpNeighborConfig *types.EdgeBgpNeighbor) (*EdgeBgpNeighbor, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpNeighbor
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &EdgeBgpNeighbor{
		client:          egw.client,
		edgeGatewayId:   egw.EdgeGateway.ID,
		EdgeBgpNeighbor: &types.EdgeBgpNeighbor{},
	}

	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, bgpNeighborConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	// API has problems therefore explicit manual handling is required to lookup newly created object
	// VCD 10.2 -> no ID for newly created object is returned at all
	// VCD 10.3 -> `Details` field in task contains ID of newly created object
	// To cover all cases this code will at first look for ID in `Details` field and fall back to
	// lookup by name if `Details` field is empty.
	//
	// The drawback of this is that it is possible to create duplicate records with the same name on VCDs that don't
	// return IDs, but there is no better way for VCD versions that don't return API code

	bgpNeighborId := task.Task.Details
	if bgpNeighborId != "" {
		getUrlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), bgpNeighborId)
		if err != nil {
			return nil, err
		}
		err = client.OpenApiGetItem(apiVersion, getUrlRef, nil, returnObject.EdgeBgpNeighbor, nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP Neighbor after creation: %s", err)
		}
	} else {
		// ID after object creation was not returned therefore retrieving the entity by Name to lookup ID
		// This has a risk of duplicate items, but is the only way to find the object when ID is not returned
		bgpNeighbor, err := egw.GetBgpNeighborByIp(bgpNeighborConfig.NeighborAddress)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP Neighbor after creation: %s", err)
		}
		returnObject = bgpNeighbor
	}

	return returnObject, nil
}

// GetAllBgpNeighbors retrieves all BGP Neighbors with an optional filter
func (egw *NsxtEdgeGateway) GetAllBgpNeighbors(queryParameters url.Values) ([]*EdgeBgpNeighbor, error) {
	queryParams := copyOrNewUrlValues(queryParameters)

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpNeighbor
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.EdgeBgpNeighbor{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*EdgeBgpNeighbor, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &EdgeBgpNeighbor{
			EdgeBgpNeighbor: typeResponses[sliceIndex],
			client:          client,
			edgeGatewayId:   egw.EdgeGateway.ID,
		}
	}

	return wrappedResponses, nil
}

// GetBgpNeighborByIp retrieves BGP Neighbor by Neighbor IP address
// It is meant to retrieve exactly one entry:
// * Will fail if more than one entry with the same Neighbor IP found (should not happen as uniqueness is
// enforced by API)
// * Will return an error containing `ErrorEntityNotFound` if no entries are found
//
// Note. API does not support filtering by 'neighborIpAddress' field therefore filtering is performed on client
// side
func (egw *NsxtEdgeGateway) GetBgpNeighborByIp(neighborIpAddress string) (*EdgeBgpNeighbor, error) {
	if neighborIpAddress == "" {
		return nil, fmt.Errorf("neighborIpAddress cannot be empty")
	}

	allBgpNeighbors, err := egw.GetAllBgpNeighbors(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	var filteredBgpNeighbors []*EdgeBgpNeighbor
	for _, bgpNeighbor := range allBgpNeighbors {
		if bgpNeighbor.EdgeBgpNeighbor.NeighborAddress == neighborIpAddress {
			filteredBgpNeighbors = append(filteredBgpNeighbors, bgpNeighbor)
		}
	}

	if len(filteredBgpNeighbors) > 1 {
		return nil, fmt.Errorf("more than one NSX-T Edge Gateway BGP Neighbor found with IP Address '%s'", neighborIpAddress)
	}

	if len(filteredBgpNeighbors) == 0 {
		return nil, fmt.Errorf("%s: no NSX-T Edge Gateway BGP Neighbor found with IP Address '%s'", ErrorEntityNotFound, neighborIpAddress)
	}

	return filteredBgpNeighbors[0], nil
}

// GetBgpNeighborById retrieves BGP Neighbor By ID
func (egw *NsxtEdgeGateway) GetBgpNeighborById(id string) (*EdgeBgpNeighbor, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpNeighbor
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), id)
	if err != nil {
		return nil, err
	}

	returnObject := &EdgeBgpNeighbor{
		client:          egw.client,
		edgeGatewayId:   egw.EdgeGateway.ID,
		EdgeBgpNeighbor: &types.EdgeBgpNeighbor{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject.EdgeBgpNeighbor, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	return returnObject, nil
}

// Update updates existing BGP Neighbor with new configuration and returns it
func (bgpNeighbor *EdgeBgpNeighbor) Update(bgpNeighborConfig *types.EdgeBgpNeighbor) (*EdgeBgpNeighbor, error) {
	client := bgpNeighbor.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpNeighbor
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, bgpNeighbor.edgeGatewayId), bgpNeighborConfig.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &EdgeBgpNeighbor{
		client:          bgpNeighbor.client,
		edgeGatewayId:   bgpNeighbor.edgeGatewayId,
		EdgeBgpNeighbor: &types.EdgeBgpNeighbor{},
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, bgpNeighborConfig, returnObject.EdgeBgpNeighbor, nil)
	if err != nil {
		return nil, fmt.Errorf("error setting NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	return returnObject, nil
}

// Delete deletes existing BGP Neighbor
func (bgpNeighbor *EdgeBgpNeighbor) Delete() error {
	client := bgpNeighbor.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeBgpNeighbor
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, bgpNeighbor.edgeGatewayId), bgpNeighbor.EdgeBgpNeighbor.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T Edge Gateway BGP Neighbor: %s", err)
	}

	return nil
}
