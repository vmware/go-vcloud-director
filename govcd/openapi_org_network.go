/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// OpenApiOrgVdcNetwork uses OpenAPI endpoint to operate both - NSX-T and NSX-V Org VDC networks
type OpenApiOrgVdcNetwork struct {
	OpenApiOrgVdcNetwork *types.OpenApiOrgVdcNetwork
	client               *Client
}

// GetOpenApiOrgVdcNetworkById allows to retrieve both - NSX-T and NSX-V Org VDC networks
func (org *Org) GetOpenApiOrgVdcNetworkById(id string) (*OpenApiOrgVdcNetwork, error) {
	// Inject Org ID filter to perform filtering on server side
	params := url.Values{}
	filterParams := queryParameterFilterAnd("orgRef.id=="+org.Org.ID, params)
	return getOpenApiOrgVdcNetworkById(org.client, id, filterParams)
}

// GetOpenApiOrgVdcNetworkById allows to retrieve both - NSX-T and NSX-V Org VDC networks
func (vdc *Vdc) GetOpenApiOrgVdcNetworkById(id string) (*OpenApiOrgVdcNetwork, error) {
	// Inject Vdc ID filter to perform filtering on server side
	params := url.Values{}
	filterParams := queryParameterFilterAnd("orgVdc.id=="+vdc.Vdc.ID, params)
	egw, err := getOpenApiOrgVdcNetworkById(vdc.client, id, filterParams)
	if err != nil {
		return nil, err
	}

	return egw, nil
}

// GetOpenApiOrgVdcNetworkByName allows to retrieve both - NSX-T and NSX-V Org VDC networks
func (vdc *Vdc) GetOpenApiOrgVdcNetworkByName(name string) (*OpenApiOrgVdcNetwork, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	allEdges, err := vdc.GetAllOpenApiOrgVdcNetworks(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Org VDC network by name '%s': %s", name, err)
	}

	return returnSingleOpenApiOrgVdcNetwork(name, allEdges)
}

// GetAllOpenApiOrgVdcNetworks allows to retrieve all NSX-T or NSX-V Org VDC networks
//
// Note. If pageSize > 32 it will be limited to maximum of 32 in this function because API validation does not allow for
// higher number
func (vdc *Vdc) GetAllOpenApiOrgVdcNetworks(queryParameters url.Values) ([]*OpenApiOrgVdcNetwork, error) {
	filteredQueryParams := queryParameterFilterAnd("orgVdc.id=="+vdc.Vdc.ID, queryParameters)
	return getAllOpenApiOrgVdcNetworks(vdc.client, filteredQueryParams)
}

// CreateOpenApiOrgVdcNetwork allows to create NSX-T or NSX-V Org VDC network
func (vdc *Vdc) CreateOpenApiOrgVdcNetwork(OrgVdcNetworkConfig *types.OpenApiOrgVdcNetwork) (*OpenApiOrgVdcNetwork, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks
	minimumApiVersion, err := vdc.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vdc.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnEgw := &OpenApiOrgVdcNetwork{
		OpenApiOrgVdcNetwork: &types.OpenApiOrgVdcNetwork{},
		client:               vdc.client,
	}

	err = vdc.client.OpenApiPostItem(minimumApiVersion, urlRef, nil, OrgVdcNetworkConfig, returnEgw.OpenApiOrgVdcNetwork)
	if err != nil {
		return nil, fmt.Errorf("error creating Org VDC network: %s", err)
	}

	return returnEgw, nil
}

// Update allows to update Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) Update(OrgVdcNetworkConfig *types.OpenApiOrgVdcNetwork) (*OpenApiOrgVdcNetwork, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks
	minimumApiVersion, err := orgVdcNet.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if OrgVdcNetworkConfig.ID == "" {
		return nil, fmt.Errorf("cannot update Org VDC network without ID")
	}

	urlRef, err := orgVdcNet.client.OpenApiBuildEndpoint(endpoint, OrgVdcNetworkConfig.ID)
	if err != nil {
		return nil, err
	}

	returnEgw := &OpenApiOrgVdcNetwork{
		OpenApiOrgVdcNetwork: &types.OpenApiOrgVdcNetwork{},
		client:               orgVdcNet.client,
	}

	err = orgVdcNet.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, OrgVdcNetworkConfig, returnEgw.OpenApiOrgVdcNetwork)
	if err != nil {
		return nil, fmt.Errorf("error updating Org VDC network: %s", err)
	}

	return returnEgw, nil
}

// Delete allows to delete Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks
	minimumApiVersion, err := orgVdcNet.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if orgVdcNet.OpenApiOrgVdcNetwork.ID == "" {
		return fmt.Errorf("cannot delete Org VDC network without ID")
	}

	urlRef, err := orgVdcNet.client.OpenApiBuildEndpoint(endpoint, orgVdcNet.OpenApiOrgVdcNetwork.ID)
	if err != nil {
		return err
	}

	err = orgVdcNet.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil)

	if err != nil {
		return fmt.Errorf("error deleting Org VDC network: %s", err)
	}

	return nil
}

// GetType returns type of Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) GetType() string {
	return orgVdcNet.OpenApiOrgVdcNetwork.NetworkType
}

// IsIsolated returns true if the network type is isolated (NSX-V and NSX-T)
func (orgVdcNet *OpenApiOrgVdcNetwork) IsIsolated() bool {
	return orgVdcNet.GetType() == types.OrgVdcNetworkTypeIsolated
}

// IsRouted returns true if the network type is isolated (NSX-V and NSX-T)
func (orgVdcNet *OpenApiOrgVdcNetwork) IsRouted() bool {
	return orgVdcNet.GetType() == types.OrgVdcNetworkTypeRouted
}

// IsImported returns true if the network type is imported (NSX-T only)
func (orgVdcNet *OpenApiOrgVdcNetwork) IsImported() bool {
	return orgVdcNet.GetType() == types.OrgVdcNetworkTypeOpaque
}

// IsDirect returns true if the network type is direct (NSX-V only)
func (orgVdcNet *OpenApiOrgVdcNetwork) IsDirect() bool {
	return orgVdcNet.GetType() == types.OrgVdcNetworkTypeDirect
}

// getOpenApiOrgVdcNetworkById is a private parent for wrapped functions:
// func (org *Org) GetOpenApiOrgVdcNetworkById(id string) (*OpenApiOrgVdcNetwork, error)
// func (vdc *Vdc) GetOpenApiOrgVdcNetworkById(id string) (*OpenApiOrgVdcNetwork, error)
func getOpenApiOrgVdcNetworkById(client *Client, id string, queryParameters url.Values) (*OpenApiOrgVdcNetwork, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty Org VDC network ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	egw := &OpenApiOrgVdcNetwork{
		OpenApiOrgVdcNetwork: &types.OpenApiOrgVdcNetwork{},
		client:               client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, queryParameters, egw.OpenApiOrgVdcNetwork)
	if err != nil {
		return nil, err
	}

	return egw, nil
}

// returnSingleOpenApiOrgVdcNetwork helps to reduce code duplication for `GetOpenApiOrgVdcNetworkByName` functions with different
// receivers
func returnSingleOpenApiOrgVdcNetwork(name string, allEdges []*OpenApiOrgVdcNetwork) (*OpenApiOrgVdcNetwork, error) {
	if len(allEdges) > 1 {
		return nil, fmt.Errorf("got more than one Org VDC network by name '%s' %d", name, len(allEdges))
	}

	if len(allEdges) < 1 {
		return nil, fmt.Errorf("%s: got zero Org VDC networks by name '%s'", ErrorEntityNotFound, name)
	}

	return allEdges[0], nil
}

// getAllOpenApiOrgVdcNetworks is a private parent for wrapped functions:
// func (vdc *Vdc) GetAllOpenApiOrgVdcNetworks(queryParameters url.Values) ([]*OpenApiOrgVdcNetwork, error)
//
// Note. If pageSize > 32 it will be limited to maximum of 32 in this function because API validation does not allow
// higher number
func getAllOpenApiOrgVdcNetworks(client *Client, queryParameters url.Values) ([]*OpenApiOrgVdcNetwork, error) {

	// Enforce maximum pageSize to be 32 as API endpoint throws error if it is > 32
	pageSizeString := queryParameters.Get("pageSize")

	switch pageSizeString {
	// If no pageSize is specified it must be set to 32 as by default low level API function OpenApiGetAllItems sets 128
	case "":
		queryParameters.Set("pageSize", "32")

	// If pageSize is specified ensure it is not >32
	default:
		pageSizeValue, err := strconv.Atoi(pageSizeString)
		if err != nil {
			return nil, fmt.Errorf("error parsing pageSize value: %s", err)
		}
		if pageSizeString != "" && pageSizeValue > 32 {
			queryParameters.Set("pageSize", "32")
		}
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.OpenApiOrgVdcNetwork{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into OpenApiOrgVdcNetwork types with client
	wrappedResponses := make([]*OpenApiOrgVdcNetwork, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &OpenApiOrgVdcNetwork{
			OpenApiOrgVdcNetwork: typeResponses[sliceIndex],
			client:               client,
		}
	}

	return wrappedResponses, nil
}
