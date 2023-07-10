/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtEdgeGatewayStaticRoute represents NSX-T Edge Gateway Static Route
type NsxtEdgeGatewayStaticRoute struct {
	NsxtEdgeGatewayStaticRoute *types.NsxtEdgeGatewayStaticRoute
	client                     *Client
	// edgeGatewayId is stored for usage in NsxtEdgeGatewayStaticRoute receiver functions
	edgeGatewayId string
}

// CreateStaticRoute based on type definition
func (egw *NsxtEdgeGateway) CreateStaticRoute(staticRouteConfig *types.NsxtEdgeGatewayStaticRoute) (*NsxtEdgeGatewayStaticRoute, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayStaticRoutes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtEdgeGatewayStaticRoute{
		client:                     egw.client,
		edgeGatewayId:              egw.EdgeGateway.ID,
		NsxtEdgeGatewayStaticRoute: &types.NsxtEdgeGatewayStaticRoute{},
	}

	// Non standard behavior of entity - `Details` field in task contains ID of newly created object while the Owner is Edge Gateway
	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, staticRouteConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway Static Route: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Edge Gateway Static Route: %s", err)
	}

	// API does not return an ID for created object - we know that it is expected to be in
	// task.Task.Details therefore attempt to find it, but if it is empty - look for an entity by a
	// set of requested parameters
	staticRouteId := task.Task.Details
	if staticRouteId != "" {
		getUrlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), staticRouteId)
		if err != nil {
			return nil, err
		}
		err = client.OpenApiGetItem(apiVersion, getUrlRef, nil, returnObject.NsxtEdgeGatewayStaticRoute, nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway Static Route after creation: %s", err)
		}
	} else {
		// ID was not present in response, therefore Static Route needs to be found manually. Using
		// 'Name', 'Description' and 'NetworkCidr' for finding the entity. Duplicate entries can
		// exist, but but it should be a good enough combination for finding unique entry until VCD API is fixed
		allStaticRoutes, err := egw.GetAllStaticRoutes(nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway Static Route after creation: %s", err)
		}

		var foundStaticRoute bool
		for _, singleStaticRoute := range allStaticRoutes {
			if singleStaticRoute.NsxtEdgeGatewayStaticRoute.Name == staticRouteConfig.Name &&
				singleStaticRoute.NsxtEdgeGatewayStaticRoute.NetworkCidr == staticRouteConfig.NetworkCidr &&
				singleStaticRoute.NsxtEdgeGatewayStaticRoute.Description == staticRouteConfig.Description {
				foundStaticRoute = true
				returnObject = singleStaticRoute
				break
			}
		}

		if !foundStaticRoute {
			return nil, fmt.Errorf("error finding Static Route after creation by Name '%s', NetworkCidr '%s', Description '%s'",
				staticRouteConfig.Name, staticRouteConfig.NetworkCidr, staticRouteConfig.Description)
		}

	}

	return returnObject, nil
}

// GetAllStaticRoutes retrieves all Static Routes for a particular NSX-T Edge Gateway
func (egw *NsxtEdgeGateway) GetAllStaticRoutes(queryParameters url.Values) ([]*NsxtEdgeGatewayStaticRoute, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayStaticRoutes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtEdgeGatewayStaticRoute{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	wrappedResponses := make([]*NsxtEdgeGatewayStaticRoute, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtEdgeGatewayStaticRoute{
			NsxtEdgeGatewayStaticRoute: typeResponses[sliceIndex],
			client:                     client,
			edgeGatewayId:              egw.EdgeGateway.ID,
		}
	}

	return wrappedResponses, nil
}

// GetStaticRouteByNetworkCidr retrieves Static Route by network CIDR
//
// Note. It will return an error if more than one items is found
func (egw *NsxtEdgeGateway) GetStaticRouteByNetworkCidr(networkCidr string) (*NsxtEdgeGatewayStaticRoute, error) {
	if networkCidr == "" {
		return nil, fmt.Errorf("cidr cannot be empty")
	}

	allStaticRoutes, err := egw.GetAllStaticRoutes(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway Static Route: %s", err)
	}

	filteredByNetworkCidr := make([]*NsxtEdgeGatewayStaticRoute, 0)
	for _, sr := range allStaticRoutes {
		if sr.NsxtEdgeGatewayStaticRoute.NetworkCidr == networkCidr {
			filteredByNetworkCidr = append(filteredByNetworkCidr, sr)
		}
	}

	singleResult, err := oneOrError("networkCidr", networkCidr, filteredByNetworkCidr)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetStaticRouteByName retrieves Static Route by name
//
// Note. It will return an error if more than one items is found
func (egw *NsxtEdgeGateway) GetStaticRouteByName(name string) (*NsxtEdgeGatewayStaticRoute, error) {
	if name == "" {
		return nil, fmt.Errorf("cidr cannot be empty")
	}

	allStaticRoutes, err := egw.GetAllStaticRoutes(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway Static Route: %s", err)
	}

	filteredByNetworkName := make([]*NsxtEdgeGatewayStaticRoute, 0)
	// First - filter by name
	for _, sr := range allStaticRoutes {
		if sr.NsxtEdgeGatewayStaticRoute.Name == name {
			filteredByNetworkName = append(filteredByNetworkName, sr)
		}
	}

	singleResult, err := oneOrError("name", name, filteredByNetworkName)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetStaticRouteById retrieves Static Route by given ID
func (egw *NsxtEdgeGateway) GetStaticRouteById(id string) (*NsxtEdgeGatewayStaticRoute, error) {
	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayStaticRoutes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), id)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtEdgeGatewayStaticRoute{
		client:                     egw.client,
		edgeGatewayId:              egw.EdgeGateway.ID,
		NsxtEdgeGatewayStaticRoute: &types.NsxtEdgeGatewayStaticRoute{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject.NsxtEdgeGatewayStaticRoute, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Edge Gateway Static Route: %s", err)
	}

	return returnObject, nil
}

// Update Static Route
func (staticRoute *NsxtEdgeGatewayStaticRoute) Update(StaticRouteConfig *types.NsxtEdgeGatewayStaticRoute) (*NsxtEdgeGatewayStaticRoute, error) {
	client := staticRoute.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayStaticRoutes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, staticRoute.edgeGatewayId), StaticRouteConfig.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtEdgeGatewayStaticRoute{
		client:                     staticRoute.client,
		edgeGatewayId:              staticRoute.edgeGatewayId,
		NsxtEdgeGatewayStaticRoute: &types.NsxtEdgeGatewayStaticRoute{},
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, StaticRouteConfig, returnObject.NsxtEdgeGatewayStaticRoute, nil)
	if err != nil {
		return nil, fmt.Errorf("error setting NSX-T Edge Gateway Static Route: %s", err)
	}

	return returnObject, nil
}

// Delete Static Route
func (staticRoute *NsxtEdgeGatewayStaticRoute) Delete() error {
	client := staticRoute.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayStaticRoutes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	// Insert Edge Gateway ID into endpoint path
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, staticRoute.edgeGatewayId), staticRoute.NsxtEdgeGatewayStaticRoute.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T Edge Gateway Static Route: %s", err)
	}

	return nil
}
