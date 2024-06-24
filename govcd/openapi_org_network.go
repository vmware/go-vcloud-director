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

const labelOrgVdcNetworkSegmentProfile = "Org VDC Network Segment Profile"

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

// GetOpenApiOrgVdcNetworkByNameAndOwnerId allows to retrieve both - NSX-T and NSX-V Org VDC networks
// by network name and Owner (VDC or VDC Group) ID
func (org *Org) GetOpenApiOrgVdcNetworkByNameAndOwnerId(name, ownerId string) (*OpenApiOrgVdcNetwork, error) {
	// Inject Org ID filter to perform filtering on server side
	queryParameters := url.Values{}
	queryParameters = queryParameterFilterAnd(fmt.Sprintf("name==%s;ownerRef.id==%s", name, ownerId), queryParameters)

	allEdges, err := getAllOpenApiOrgVdcNetworks(org.client, queryParameters, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Org VDC network by name '%s' in Owner '%s': %s", name, ownerId, err)
	}

	return returnSingleOpenApiOrgVdcNetwork(name, allEdges)
}

// GetOpenApiOrgVdcNetworkById allows to retrieve both - NSX-T and NSX-V Org VDC networks
func (vdc *Vdc) GetOpenApiOrgVdcNetworkById(id string) (*OpenApiOrgVdcNetwork, error) {
	return getOrgVdcNetworkById(vdc.client, id, vdc.Vdc.ID)
}

// GetOpenApiOrgVdcNetworkById allows to retrieve both - NSX-T and NSX-V Org VDC Group networks
func (vdcGroup *VdcGroup) GetOpenApiOrgVdcNetworkById(id string) (*OpenApiOrgVdcNetwork, error) {
	return getOrgVdcNetworkById(vdcGroup.client, id, vdcGroup.VdcGroup.Id)
}

// getOrgVdcNetworkById allows to retrieve both - NSX-T and NSX-V Org VDC Group networks
func getOrgVdcNetworkById(client *Client, id, ownerId string) (*OpenApiOrgVdcNetwork, error) {
	// Inject Vdc ID filter to perform filtering on server side
	params := url.Values{}
	filterParams := queryParameterFilterAnd("ownerRef.id=="+ownerId, params)
	egw, err := getOpenApiOrgVdcNetworkById(client, id, filterParams)
	if err != nil {
		return nil, err
	}

	return egw, nil
}

// GetOpenApiOrgVdcNetworkByName allows to retrieve both - NSX-T and NSX-V Org VDC networks
func (vdc *Vdc) GetOpenApiOrgVdcNetworkByName(name string) (*OpenApiOrgVdcNetwork, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("name==%s", name))

	allEdges, err := vdc.GetAllOpenApiOrgVdcNetworks(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Org VDC network by name '%s': %s", name, err)
	}

	return returnSingleOpenApiOrgVdcNetwork(name, allEdges)
}

// GetOpenApiOrgVdcNetworkByName allows to retrieve both - NSX-T and NSX-V Org VDC networks
func (vdcGroup *VdcGroup) GetOpenApiOrgVdcNetworkByName(name string) (*OpenApiOrgVdcNetwork, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("name==%s", name))

	allEdges, err := vdcGroup.GetAllOpenApiOrgVdcNetworks(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Org VDC network by name '%s': %s", name, err)
	}

	return returnSingleOpenApiOrgVdcNetwork(name, allEdges)
}

// GetAllOpenApiOrgVdcNetworks allows to retrieve all NSX-T or NSX-V Org VDC networks in Org
//
// Note. If pageSize > 32 it will be limited to maximum of 32 in this function because API validation does not allow for
// higher number
func (org *Org) GetAllOpenApiOrgVdcNetworks(queryParameters url.Values) ([]*OpenApiOrgVdcNetwork, error) {
	filteredQueryParams := queryParameterFilterAnd("orgRef.id=="+org.Org.ID, queryParameters)
	return getAllOpenApiOrgVdcNetworks(org.client, filteredQueryParams, nil)
}

// GetAllOpenApiOrgVdcNetworks allows to retrieve all NSX-T or NSX-V Org VDC networks in Vdc
//
// Note. If pageSize > 32 it will be limited to maximum of 32 in this function because API validation does not allow for
// higher number
func (vdc *Vdc) GetAllOpenApiOrgVdcNetworks(queryParameters url.Values) ([]*OpenApiOrgVdcNetwork, error) {
	filteredQueryParams := queryParameterFilterAnd("ownerRef.id=="+vdc.Vdc.ID, queryParameters)
	return getAllOpenApiOrgVdcNetworks(vdc.client, filteredQueryParams, nil)
}

// GetAllOpenApiOrgVdcNetworks allows to retrieve all NSX-T or NSX-V Org VDC networks in Vdc
//
// Note. If pageSize > 32 it will be limited to maximum of 32 in this function because API validation does not allow for
// higher number
func (vdcGroup *VdcGroup) GetAllOpenApiOrgVdcNetworks(queryParameters url.Values) ([]*OpenApiOrgVdcNetwork, error) {
	filteredQueryParams := queryParameterFilterAnd("ownerRef.id=="+vdcGroup.VdcGroup.Id, queryParameters)
	return getAllOpenApiOrgVdcNetworks(vdcGroup.client, filteredQueryParams, nil)
}

// CreateOpenApiOrgVdcNetwork allows to create NSX-T or NSX-V Org VDC network
func (org *Org) CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork) (*OpenApiOrgVdcNetwork, error) {
	return createOpenApiOrgVdcNetwork(org.client, orgVdcNetworkConfig)
}

// CreateOpenApiOrgVdcNetwork allows to create NSX-T or NSX-V Org VDC network
func (vdc *Vdc) CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork) (*OpenApiOrgVdcNetwork, error) {
	return createOpenApiOrgVdcNetwork(vdc.client, orgVdcNetworkConfig)
}

// CreateOpenApiOrgVdcNetwork allows to create NSX-T or NSX-V Org VDC network
func (vdcGroup *VdcGroup) CreateOpenApiOrgVdcNetwork(orgVdcNetworkConfig *types.OpenApiOrgVdcNetwork) (*OpenApiOrgVdcNetwork, error) {
	return createOpenApiOrgVdcNetwork(vdcGroup.client, orgVdcNetworkConfig)
}

// UpdateDhcp updates DHCP configuration for specific Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) UpdateDhcp(orgVdcNetworkDhcpConfig *types.OpenApiOrgVdcNetworkDhcp) (*OpenApiOrgVdcNetworkDhcp, error) {
	if orgVdcNet.client == nil || orgVdcNet.OpenApiOrgVdcNetwork == nil || orgVdcNet.OpenApiOrgVdcNetwork.ID == "" {
		return nil, fmt.Errorf("error - Org VDC network structure must be set and have ID field available")
	}
	return updateOrgNetworkDhcp(orgVdcNet.client, orgVdcNet.OpenApiOrgVdcNetwork.ID, orgVdcNetworkDhcpConfig)
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

	err = orgVdcNet.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, OrgVdcNetworkConfig, returnEgw.OpenApiOrgVdcNetwork, nil)
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

	err = orgVdcNet.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

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

// IsNsxt returns true if the network is backed by NSX-T
func (orgVdcNet *OpenApiOrgVdcNetwork) IsNsxt() bool {

	// orgVdcNet.OpenApiOrgVdcNetwork.OrgVdcIsNsxTBacked returns `true` only if network is a member
	// of VDC (not VDC Group) therefore an additional check for `BackingNetworkType` is required

	return orgVdcNet.OpenApiOrgVdcNetwork.OrgVdcIsNsxTBacked ||
		orgVdcNet.OpenApiOrgVdcNetwork.BackingNetworkType == types.OpenApiOrgVdcNetworkBackingTypeNsxt
}

// IsDhcpEnabled returns true if DHCP is enabled for NSX-T Org VDC network, false otherwise
func (orgVdcNet *OpenApiOrgVdcNetwork) IsDhcpEnabled() bool {
	if !orgVdcNet.IsNsxt() {
		return false
	}

	dhcpConfig, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcp()
	if err != nil {
		return false
	}

	if dhcpConfig == nil || dhcpConfig.OpenApiOrgVdcNetworkDhcp == nil || dhcpConfig.OpenApiOrgVdcNetworkDhcp.Enabled == nil || !*dhcpConfig.OpenApiOrgVdcNetworkDhcp.Enabled {
		return false
	}

	return true
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

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, queryParameters, egw.OpenApiOrgVdcNetwork, nil)
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
func getAllOpenApiOrgVdcNetworks(client *Client, queryParameters url.Values, additionalHeader map[string]string) ([]*OpenApiOrgVdcNetwork, error) {

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
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, additionalHeader)
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

// createOpenApiOrgVdcNetwork is wrapped by public CreateOpenApiOrgVdcNetwork methods
func createOpenApiOrgVdcNetwork(client *Client, OrgVdcNetworkConfig *types.OpenApiOrgVdcNetwork) (*OpenApiOrgVdcNetwork, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnEgw := &OpenApiOrgVdcNetwork{
		OpenApiOrgVdcNetwork: &types.OpenApiOrgVdcNetwork{},
		client:               client,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, OrgVdcNetworkConfig, returnEgw.OpenApiOrgVdcNetwork, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Org VDC network: %s", err)
	}

	return returnEgw, nil
}

// GetSegmentProfile retrieves Segment Profile configuration for a single Org VDC Network
func (orgVdcNet *OpenApiOrgVdcNetwork) GetSegmentProfile() (*types.OrgVdcNetworkSegmentProfiles, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworkSegmentProfiles,
		endpointParams: []string{orgVdcNet.OpenApiOrgVdcNetwork.ID},
		entityLabel:    labelOrgVdcNetworkSegmentProfile,
	}
	return getInnerEntity[types.OrgVdcNetworkSegmentProfiles](orgVdcNet.client, c)
}

// UpdateSegmentProfile updates a Segment Profile with a given configuration
func (orgVdcNet *OpenApiOrgVdcNetwork) UpdateSegmentProfile(entityConfig *types.OrgVdcNetworkSegmentProfiles) (*types.OrgVdcNetworkSegmentProfiles, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworkSegmentProfiles,
		endpointParams: []string{orgVdcNet.OpenApiOrgVdcNetwork.ID},
		entityLabel:    labelOrgVdcNetworkSegmentProfile,
	}
	return updateInnerEntity(orgVdcNet.client, c, entityConfig)
}

// GetAllOpenApiOrgVdcNetworks checks all Org VDC networks available to the current user
// When 'multiSite' is set, it will also check the networks available from associated organizations
func (adminOrg *AdminOrg) GetAllOpenApiOrgVdcNetworks(queryParameters url.Values, multiSite bool) ([]*OpenApiOrgVdcNetwork, error) {
	var additionalHeader map[string]string
	if multiSite {
		additionalHeader = map[string]string{"Accept": "{{MEDIA_TYPE}};version={{API_VERSION}};multisite=global"}
	}
	if queryParameters == nil {
		queryParameters = make(url.Values)
	}
	result, err := getAllOpenApiOrgVdcNetworks(adminOrg.client, queryParameters, additionalHeader)
	return result, err
}
