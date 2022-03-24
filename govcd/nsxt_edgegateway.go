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

// GetNsxtEdgeGatewayById allows retrieving NSX-T edge gateway by ID for Org admins
func (adminOrg *AdminOrg) GetNsxtEdgeGatewayById(id string) (*NsxtEdgeGateway, error) {
	return getNsxtEdgeGatewayById(adminOrg.client, id, nil)
}

// GetNsxtEdgeGatewayById allows retrieving NSX-T edge gateway by ID for Org users
func (org *Org) GetNsxtEdgeGatewayById(id string) (*NsxtEdgeGateway, error) {
	return getNsxtEdgeGatewayById(org.client, id, nil)
}

// GetNsxtEdgeGatewayById allows retrieving NSX-T edge gateway by ID for specific VDC
func (vdc *Vdc) GetNsxtEdgeGatewayById(id string) (*NsxtEdgeGateway, error) {
	params := url.Values{}
	filterParams := queryParameterFilterAnd("ownerRef.id=="+vdc.Vdc.ID, params)
	egw, err := getNsxtEdgeGatewayById(vdc.client, id, filterParams)
	if err != nil {
		return nil, err
	}

	if egw.EdgeGateway.OwnerRef.ID != vdc.Vdc.ID {
		return nil, fmt.Errorf("%s: no NSX-T Edge Gateway with ID '%s' found in VDC '%s'",
			ErrorEntityNotFound, id, vdc.Vdc.ID)
	}

	return egw, nil
}

// GetNsxtEdgeGatewayByName allows retrieving NSX-T edge gateway by Name for Org admins
func (adminOrg *AdminOrg) GetNsxtEdgeGatewayByName(name string) (*NsxtEdgeGateway, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	allEdges, err := adminOrg.GetAllNsxtEdgeGateways(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Edge Gateway by name '%s': %s", name, err)
	}

	onlyNsxtEdges := filterOnlyNsxtEdges(allEdges)

	return returnSingleNsxtEdgeGateway(name, onlyNsxtEdges)
}

// GetNsxtEdgeGatewayByName allows retrieving NSX-T edge gateway by Name for Org admins
func (org *Org) GetNsxtEdgeGatewayByName(name string) (*NsxtEdgeGateway, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	allEdges, err := org.GetAllNsxtEdgeGateways(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Edge Gateway by name '%s': %s", name, err)
	}

	onlyNsxtEdges := filterOnlyNsxtEdges(allEdges)

	return returnSingleNsxtEdgeGateway(name, onlyNsxtEdges)
}

// GetNsxtEdgeGatewayByNameAndOwnerId looks up NSX-T Edge Gateway by name and its owner ID (owner
// can be VDC or VDC Group).
func (org *Org) GetNsxtEdgeGatewayByNameAndOwnerId(edgeGatewayName, ownerId string) (*NsxtEdgeGateway, error) {
	if edgeGatewayName == "" || ownerId == "" {
		return nil, fmt.Errorf("'edgeGatewayName' and 'ownerId' must both be specified")
	}

	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("ownerRef.id==%s;name==%s", ownerId, edgeGatewayName))

	allEdges, err := org.GetAllNsxtEdgeGateways(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Edge Gateway by name '%s': %s", edgeGatewayName, err)
	}

	onlyNsxtEdges := filterOnlyNsxtEdges(allEdges)

	return returnSingleNsxtEdgeGateway(edgeGatewayName, onlyNsxtEdges)
}

// GetNsxtEdgeGatewayByName allows to retrieve NSX-T edge gateway by Name for specific VDC
func (vdc *Vdc) GetNsxtEdgeGatewayByName(name string) (*NsxtEdgeGateway, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	allEdges, err := vdc.GetAllNsxtEdgeGateways(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Edge Gateway by name '%s': %s", name, err)
	}

	return returnSingleNsxtEdgeGateway(name, allEdges)
}

// GetNsxtEdgeGatewayByName allows to retrieve NSX-T edge gateway by Name for specific VDC Group
func (vdcGroup *VdcGroup) GetNsxtEdgeGatewayByName(name string) (*NsxtEdgeGateway, error) {
	if name == "" {
		return nil, fmt.Errorf("'name' must be specified")
	}

	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	allEdges, err := vdcGroup.GetAllNsxtEdgeGateways(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Edge Gateway by name '%s': %s", name, err)
	}

	return returnSingleNsxtEdgeGateway(name, allEdges)
}

// GetAllNsxtEdgeGateways allows to retrieve all NSX-T edge gateways for Org Admins
func (adminOrg *AdminOrg) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	return getAllNsxtEdgeGateways(adminOrg.client, queryParameters)
}

// GetAllNsxtEdgeGateways allows to retrieve all NSX-T edge gateways for Org users
func (org *Org) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	return getAllNsxtEdgeGateways(org.client, queryParameters)
}

// GetAllNsxtEdgeGateways allows to retrieve all NSX-T edge gateways for specific VDC
func (vdc *Vdc) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	filteredQueryParams := queryParameterFilterAnd("ownerRef.id=="+vdc.Vdc.ID, queryParameters)
	return getAllNsxtEdgeGateways(vdc.client, filteredQueryParams)
}

// GetAllNsxtEdgeGateways allows to retrieve all NSX-T edge gateways for specific VDC
func (vdcGroup *VdcGroup) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	filteredQueryParams := queryParameterFilterAnd("ownerRef.id=="+vdcGroup.VdcGroup.Id, queryParameters)
	return getAllNsxtEdgeGateways(vdcGroup.client, filteredQueryParams)
}

// CreateNsxtEdgeGateway allows to create NSX-T edge gateway for Org admins
func (adminOrg *AdminOrg) CreateNsxtEdgeGateway(edgeGatewayConfig *types.OpenAPIEdgeGateway) (*NsxtEdgeGateway, error) {
	if !adminOrg.client.IsSysAdmin {
		return nil, fmt.Errorf("only System Administrator can create Edge Gateway")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := adminOrg.client.getOpenApiHighestElevatedVersion(endpoint)
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

	err = adminOrg.client.OpenApiPostItem(minimumApiVersion, urlRef, nil, edgeGatewayConfig, returnEgw.EdgeGateway, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Edge Gateway: %s", err)
	}

	return returnEgw, nil
}

// Update allows updating NSX-T edge gateway for Org admins
func (egw *NsxtEdgeGateway) Update(edgeGatewayConfig *types.OpenAPIEdgeGateway) (*NsxtEdgeGateway, error) {
	if !egw.client.IsSysAdmin {
		return nil, fmt.Errorf("only System Administrator can update Edge Gateway")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := egw.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if edgeGatewayConfig.ID == "" {
		return nil, fmt.Errorf("cannot update Edge Gateway without ID")
	}

	urlRef, err := egw.client.OpenApiBuildEndpoint(endpoint, edgeGatewayConfig.ID)
	if err != nil {
		return nil, err
	}

	returnEgw := &NsxtEdgeGateway{
		EdgeGateway: &types.OpenAPIEdgeGateway{},
		client:      egw.client,
	}

	err = egw.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, edgeGatewayConfig, returnEgw.EdgeGateway, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Edge Gateway: %s", err)
	}

	return returnEgw, nil
}

// Delete allows deleting NSX-T edge gateway for sysadmins
func (egw *NsxtEdgeGateway) Delete() error {
	if !egw.client.IsSysAdmin {
		return fmt.Errorf("only Provider can delete Edge Gateway")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := egw.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if egw.EdgeGateway.ID == "" {
		return fmt.Errorf("cannot delete Edge Gateway without ID")
	}

	urlRef, err := egw.client.OpenApiBuildEndpoint(endpoint, egw.EdgeGateway.ID)
	if err != nil {
		return err
	}

	err = egw.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting Edge Gateway: %s", err)
	}

	return nil
}

// MoveToVdcOrVdcGroup moves NSX-T Edge Gateway to another VDC. This can cover such scenarios:
// * Move from VDC to VDC Group
// * Move from VDC Group to VDC (which is part of that VDC Group)
//
// This function is just an Update operation with OwnerRef changed to vdcGroupId, but it is more
// convenient to use it.
// Note. NSX-T Edge Gateway cannot be moved directly from one VDC to another
func (egw *NsxtEdgeGateway) MoveToVdcOrVdcGroup(vdcOrVdcGroupId string) (*NsxtEdgeGateway, error) {
	edgeGatewayConfig := egw.EdgeGateway
	edgeGatewayConfig.OwnerRef = &types.OpenApiReference{ID: vdcOrVdcGroupId}
	// Explicitly unset VDC field because using it fails
	edgeGatewayConfig.OrgVdc = nil

	return egw.Update(edgeGatewayConfig)
}

// getNsxtEdgeGatewayById is a private parent for wrapped functions:
// func (adminOrg *AdminOrg) GetNsxtEdgeGatewayByName(id string) (*NsxtEdgeGateway, error)
// func (org *Org) GetNsxtEdgeGatewayByName(id string) (*NsxtEdgeGateway, error)
// func (vdc *Vdc) GetNsxtEdgeGatewayById(id string) (*NsxtEdgeGateway, error)
func getNsxtEdgeGatewayById(client *Client, id string, queryParameters url.Values) (*NsxtEdgeGateway, error) {
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

	egw := &NsxtEdgeGateway{
		EdgeGateway: &types.OpenAPIEdgeGateway{},
		client:      client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, queryParameters, egw.EdgeGateway, nil)
	if err != nil {
		return nil, err
	}

	if egw.EdgeGateway.GatewayBacking.GatewayType != "NSXT_BACKED" {
		return nil, fmt.Errorf("%s: this is not NSX-T Edge Gateway (%s)",
			ErrorEntityNotFound, egw.EdgeGateway.GatewayBacking.GatewayType)
	}

	return egw, nil
}

// returnSingleNsxtEdgeGateway helps to reduce code duplication for `GetNsxtEdgeGatewayByName` functions with different
// receivers
func returnSingleNsxtEdgeGateway(name string, allEdges []*NsxtEdgeGateway) (*NsxtEdgeGateway, error) {
	if len(allEdges) > 1 {
		return nil, fmt.Errorf("got more than 1 Edge Gateway by name '%s' %d", name, len(allEdges))
	}

	if len(allEdges) < 1 {
		return nil, fmt.Errorf("%s: got 0 Edge Gateways by name '%s'", ErrorEntityNotFound, name)
	}

	return allEdges[0], nil
}

// getAllNsxtEdgeGateways is a private parent for wrapped functions:
// func (adminOrg *AdminOrg) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error)
// func (org *Org) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error)
// func (vdc *Vdc) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error)
func getAllNsxtEdgeGateways(client *Client, queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	minimumApiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.OpenAPIEdgeGateway{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtEdgeGateway types with client
	wrappedResponses := make([]*NsxtEdgeGateway, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtEdgeGateway{
			EdgeGateway: typeResponses[sliceIndex],
			client:      client,
		}
	}

	onlyNsxtEdges := filterOnlyNsxtEdges(wrappedResponses)

	return onlyNsxtEdges, nil
}

// filterOnlyNsxtEdges filters our list of edge gateways only for NSXT_BACKED ones because original endpoint can
// return NSX-V and NSX-T backed edge gateways.
func filterOnlyNsxtEdges(allEdges []*NsxtEdgeGateway) []*NsxtEdgeGateway {
	filteredEdges := make([]*NsxtEdgeGateway, 0)

	for index := range allEdges {
		if allEdges[index] != nil && allEdges[index].EdgeGateway != nil &&
			allEdges[index].EdgeGateway.GatewayBacking != nil &&
			allEdges[index].EdgeGateway.GatewayBacking.GatewayType == "NSXT_BACKED" {
			filteredEdges = append(filteredEdges, allEdges[index])
		}
	}

	return filteredEdges
}
