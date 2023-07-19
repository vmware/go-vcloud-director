/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/netip"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
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

// GetAllNsxtEdgeGateways allows to retrieve all NSX-T Edge Gateways
func (vcdClient *VCDClient) GetAllNsxtEdgeGateways(queryParameters url.Values) ([]*NsxtEdgeGateway, error) {
	if vcdClient == nil {
		return nil, fmt.Errorf("vcdClient is empty")
	}
	return getAllNsxtEdgeGateways(&vcdClient.Client, queryParameters)
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

// Refresh reloads NSX-T Edge Gateway contents
func (egw *NsxtEdgeGateway) Refresh() error {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return fmt.Errorf("cannot refresh Edge Gateway without ID")
	}

	refreshedEdge, err := getNsxtEdgeGatewayById(egw.client, egw.EdgeGateway.ID, nil)
	if err != nil {
		return fmt.Errorf("error refreshing NSX-T Edge Gateway: %s", err)
	}
	egw.EdgeGateway = refreshedEdge.EdgeGateway
	return nil
}

// Update allows updating NSX-T edge gateway for Org admins
func (egw *NsxtEdgeGateway) Update(edgeGatewayConfig *types.OpenAPIEdgeGateway) (*NsxtEdgeGateway, error) {
	if !egw.client.IsSysAdmin {
		return nil, fmt.Errorf("only System Administrator can update Edge Gateway")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGateways
	apiVersion, err := egw.client.getOpenApiHighestElevatedVersion(endpoint)
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

	err = egw.client.OpenApiPutItem(apiVersion, urlRef, nil, edgeGatewayConfig, returnEgw.EdgeGateway, nil)
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
	apiVersion, err := egw.client.getOpenApiHighestElevatedVersion(endpoint)
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

	err = egw.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

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

// GetUsedIpAddresses uses dedicated endpoint to retrieve used IP addresses in an Edge Gateway
func (egw *NsxtEdgeGateway) GetUsedIpAddresses(queryParameters url.Values) ([]*types.GatewayUsedIpAddress, error) {
	if egw.EdgeGateway == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("edge gateway ID must be set to retrieve used IP addresses")
	}
	client := egw.client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayUsedIpAddresses
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	typeResponse := make([]*types.GatewayUsedIpAddress, 0)
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	return typeResponse, nil
}

// GetUnusedExternalIPAddresses will retrieve a requiredIpCount of unused IP addresses for Edge
// Gateway
// Arguments:
// * `requiredIpCount` (how many unuseds IPs should be returned). It will fail and return an
// error if all IPs specified in 'requiredIpCount' cannot be found.
// * `optionalSubnet` is specified (CIDR notation, e.g. 192.168.1.0/24), it will look for an IP in
// this subnet only.
// * `refresh` defines if Edge Gateway structure should be retrieved with latest data before
// performing IP lookup operation
//
// Input and return arguments are using Go's native 'netip' package for IP addressing. This ensures
// correct support for IPv4 and IPv6 IPs.
// `netip.ParseAddr`, `netip.ParsePrefix`, `netip.Addr.String` functions can be used for conversion
// from/to strings
//
// This function performs below listed steps:
// 1. Retrieves a complete list of IPs in Edge Gateway uplinks (returns error if none are found)
// 2. if 'optionalSubnet' was specified - filter IP addresses to only fall into that subnet
// 3. Retrieves all used IP addresses in Edge Gateway using dedicated API endpoint
// 4. Subtracts used IP addresses from available list of IPs in uplink (optionally filtered by optionalSubnet in step 2)
// 5. Checks if 'requiredIpCount' criteria is met, returns error otherwise
// 6. Returns required amount of unused IPs (as defined in 'requiredIpCount')
//
// Notes:
// * This function uses Go's builtin `netip` package to avoid any string processing of IPs and
// supports IPv4 and IPv6 addressing.
// * If an unused IP is not found it will return 'netip.Addr{}' (not using *netip.Addr{} to match
// library semantics) and an error
// * It will return an error if any of uplink IP ranges End IP address is lower than Start IP
// address
func (egw *NsxtEdgeGateway) GetUnusedExternalIPAddresses(requiredIpCount int, optionalSubnet netip.Prefix, refresh bool) ([]netip.Addr, error) {
	if refresh {
		err := egw.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing Edge Gateway: %s", err)
		}
	}
	usedIpAddresses, err := egw.GetUsedIpAddresses(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting used IP addresses for Edge Gateway: %s", err)
	}

	return getUnusedExternalIPAddress(egw.EdgeGateway.EdgeGatewayUplinks, usedIpAddresses, requiredIpCount, optionalSubnet)
}

// GetAllUnusedExternalIPAddresses will retrieve all unassigned IP addresses for Edge Gateway It is
// similar to GetUnusedExternalIPAddresses but returns all unused IPs instead of a specific amount
func (egw *NsxtEdgeGateway) GetAllUnusedExternalIPAddresses(refresh bool) ([]netip.Addr, error) {
	if refresh {
		err := egw.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing Edge Gateway: %s", err)
		}
	}
	usedIpAddresses, err := egw.GetUsedIpAddresses(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting used IP addresses for Edge Gateway: %s", err)
	}

	return getAllUnusedExternalIPAddresses(egw.EdgeGateway.EdgeGatewayUplinks, usedIpAddresses, netip.Prefix{})
}

// GetAllocatedIpCount traverses all subnets in Edge Gateway and returns a count of allocated IP
// count for each subnet in each uplink
func (egw *NsxtEdgeGateway) GetAllocatedIpCount(refresh bool) (int, error) {
	if refresh {
		err := egw.Refresh()
		if err != nil {
			return 0, fmt.Errorf("error refreshing Edge Gateway: %s", err)
		}
	}

	allocatedIpCount := 0

	for _, uplink := range egw.EdgeGateway.EdgeGatewayUplinks {
		for _, subnet := range uplink.Subnets.Values {
			if subnet.TotalIPCount != nil {
				allocatedIpCount += *subnet.TotalIPCount
			}
		}
	}

	return allocatedIpCount, nil
}

// GetUsedIpAddressSlice retrieves a list of used IP addresses in an Edge Gateway and returns it
// using native Go type '[]netip.Addr'
func (egw *NsxtEdgeGateway) GetUsedIpAddressSlice(refresh bool) ([]netip.Addr, error) {
	if refresh {
		err := egw.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing Edge Gateway: %s", err)
		}
	}
	usedIpAddresses, err := egw.GetUsedIpAddresses(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting used IP addresses for Edge Gateway: %s", err)
	}

	return flattenGatewayUsedIpAddressesToIpSlice(usedIpAddresses)
}

// QuickDeallocateIpCount refreshes Edge Gateway structure and deallocates specified ipCount from it
// by modifying Uplink structure and calling Update() on it.
//
// Notes:
// * This is a reverse operation to QuickAllocateIpCount and is provided for convenience as the API
// does not support negative values for QuickAddAllocatedIPCount field
// * This function modifies Edge Gateway structure and calls update. To only modify structure,
// please use `NsxtEdgeGateway.DeallocateIpCount` function
func (egw *NsxtEdgeGateway) QuickDeallocateIpCount(ipCount int) (*NsxtEdgeGateway, error) {
	if egw.EdgeGateway == nil {
		return nil, fmt.Errorf("edge gateway is not initialized")
	}

	err := egw.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing Edge Gateway: %s", err)
	}

	err = egw.DeallocateIpCount(ipCount)
	if err != nil {
		return nil, fmt.Errorf("error deallocating IP count: %s", err)
	}

	return egw.Update(egw.EdgeGateway)
}

// DeallocateIpCount modifies the structure to deallocate IP addresses from the Edge Gateway
// uplinks.
//
// Notes:
// * This function does not call Update() on the Edge Gateway and it is up to the caller to perform
// this operation (or use NsxtEdgeGateway.QuickDeallocateIpCount which wraps this function and
// performs API call)
// * Use `QuickAddAllocatedIPCount` field in the uplink structure to leverage VCD API directly for
// allocating IP addresses.
func (egw *NsxtEdgeGateway) DeallocateIpCount(deallocateIpCount int) error {
	if deallocateIpCount < 0 {
		return fmt.Errorf("deallocateIpCount must be greater than 0")
	}

	if egw == nil || egw.EdgeGateway == nil {
		return fmt.Errorf("edge gateway structure cannot be nil")
	}

	edgeGatewayType := egw.EdgeGateway

	for uplinkIndex, uplink := range edgeGatewayType.EdgeGatewayUplinks {
		for subnetIndex, subnet := range uplink.Subnets.Values {

			// TotalIPCount is an address of a variable so it needs to be dereferenced for easier arithmetic
			// operations. In the end of processing the value is set back to the original location.
			singleSubnetTotalIpCount := *edgeGatewayType.EdgeGatewayUplinks[uplinkIndex].Subnets.Values[subnetIndex].TotalIPCount

			if singleSubnetTotalIpCount > 0 {
				util.Logger.Printf("[DEBUG] Edge Gateway deallocating IPs from subnet '%s', TotalIPCount '%d', deallocate IP count '%d'",
					subnet.Gateway, subnet.TotalIPCount, deallocateIpCount)

				// If a subnet contains more allocated IPs than we need to deallocate - deallocate only what we need
				if singleSubnetTotalIpCount >= deallocateIpCount {
					singleSubnetTotalIpCount -= deallocateIpCount

					// To make deallocation work one must set this to true
					edgeGatewayType.EdgeGatewayUplinks[uplinkIndex].Subnets.Values[subnetIndex].AutoAllocateIPRanges = true

					deallocateIpCount = 0
				} else { // If we have less IPs allocated than we need to deallocate - deallocate all of them
					deallocateIpCount -= singleSubnetTotalIpCount
					singleSubnetTotalIpCount = 0
					edgeGatewayType.EdgeGatewayUplinks[uplinkIndex].Subnets.Values[subnetIndex].AutoAllocateIPRanges = true // To make deallocation work one must set this to true
					util.Logger.Printf("[DEBUG] Edge Gateway IP count after partial deallocation %d", edgeGatewayType.EdgeGatewayUplinks[uplinkIndex].Subnets.Values[subnetIndex].TotalIPCount)
				}
			}

			// Setting value back to original location after all operations
			edgeGatewayType.EdgeGatewayUplinks[uplinkIndex].Subnets.Values[subnetIndex].TotalIPCount = &singleSubnetTotalIpCount
			util.Logger.Printf("[DEBUG] Edge Gateway IP count after complete deallocation %d", edgeGatewayType.EdgeGatewayUplinks[uplinkIndex].Subnets.Values[subnetIndex].TotalIPCount)

			if deallocateIpCount == 0 {
				break
			}
		}
	}

	if deallocateIpCount > 0 {
		return fmt.Errorf("not enough IPs allocated to deallocate requested '%d' IPs", deallocateIpCount)
	}

	return nil
}

// GetQoS retrieves QoS (rate limiting) configuration for an NSX-T Edge Gateway
func (egw *NsxtEdgeGateway) GetQoS() (*types.NsxtEdgeGatewayQos, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot get QoS for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayQos
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	qos := &types.NsxtEdgeGatewayQos{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, qos, nil)
	if err != nil {
		return nil, err
	}

	return qos, nil
}

// UpdateQoS updates QoS (rate limiting) configuration for an NSX-T Edge Gateway
func (egw *NsxtEdgeGateway) UpdateQoS(qosConfig *types.NsxtEdgeGatewayQos) (*types.NsxtEdgeGatewayQos, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot update QoS for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayQos
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	// update QoS with given qosConfig
	updatedQos := &types.NsxtEdgeGatewayQos{}
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, qosConfig, updatedQos, nil)
	if err != nil {
		return nil, err
	}

	return updatedQos, nil
}

// GetDhcpForwarder gets DHCP forwarder configuration for an NSX-T Edge Gateway
func (egw *NsxtEdgeGateway) GetDhcpForwarder() (*types.NsxtEdgeGatewayDhcpForwarder, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot get DHCP forwarder for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDhcpForwarder
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	dhcpForwarder := &types.NsxtEdgeGatewayDhcpForwarder{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, dhcpForwarder, nil)
	if err != nil {
		return nil, err
	}

	return dhcpForwarder, nil
}

// UpdateDhcpForwarder updates DHCP forwarder configuration for an NSX-T Edge Gateway
func (egw *NsxtEdgeGateway) UpdateDhcpForwarder(dhcpForwarderConfig *types.NsxtEdgeGatewayDhcpForwarder) (*types.NsxtEdgeGatewayDhcpForwarder, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot update DHCP forwarder for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDhcpForwarder
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	// update DHCP forwarder with given dhcpForwarderConfig
	updatedDhcpForwarder, err := egw.GetDhcpForwarder()
	if err != nil {
		return nil, err
	}
	dhcpForwarderConfig.Version = updatedDhcpForwarder.Version

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, dhcpForwarderConfig, updatedDhcpForwarder, nil)
	if err != nil {
		return nil, err
	}

	return updatedDhcpForwarder, nil
}

// GetSlaacProfile gets SLAAC (Stateless Address Autoconfiguration) Profile configuration for an
// NSX-T Edge Gateway.
// Note. It represents DHCPv6 Edge Gateway configuration in UI
func (egw *NsxtEdgeGateway) GetSlaacProfile() (*types.NsxtEdgeGatewaySlaacProfile, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot get SLAAC Profile for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewaySlaacProfile
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	slaacProfile := &types.NsxtEdgeGatewaySlaacProfile{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, slaacProfile, nil)
	if err != nil {
		return nil, err
	}

	return slaacProfile, nil
}

// UpdateSlaacProfile creates a SLAAC (Stateless Address Autoconfiguration) profile or updates the
// existing one if it already exists.
// Note. It represents DHCPv6 Edge Gateway configuration in UI
func (egw *NsxtEdgeGateway) UpdateSlaacProfile(slaacProfileConfig *types.NsxtEdgeGatewaySlaacProfile) (*types.NsxtEdgeGatewaySlaacProfile, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot update SLAAC Profile for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewaySlaacProfile
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	updatedSlaacProfile := &types.NsxtEdgeGatewaySlaacProfile{}
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, slaacProfileConfig, updatedSlaacProfile, nil)
	if err != nil {
		return nil, err
	}

	return updatedSlaacProfile, nil
}

func getAllUnusedExternalIPAddresses(uplinks []types.EdgeGatewayUplinks, usedIpAddresses []*types.GatewayUsedIpAddress, optionalSubnet netip.Prefix) ([]netip.Addr, error) {
	// 1. Flatten all IP ranges in Edge Gateway using Go's native 'netip.Addr' IP container instead
	// of plain strings because it is more robust (supports IPv4 and IPv6 and also comparison
	// operator)
	assignedIpSlice, err := flattenEdgeGatewayUplinkToIpSlice(uplinks)
	if err != nil {
		return nil, fmt.Errorf("error listing all IPs in Edge Gateway: %s", err)
	}

	if len(assignedIpSlice) == 0 {
		return nil, fmt.Errorf("no IPs found in Edge Gateway configuration")
	}

	// 2. Optionally filter given IP ranges by optionalSubnet value (if specified)
	if optionalSubnet != (netip.Prefix{}) {
		assignedIpSlice, err = filterIpSlicesBySubnet(assignedIpSlice, optionalSubnet)
		if err != nil {
			return nil, fmt.Errorf("error filtering ranges for given subnet '%s': %s", optionalSubnet, err)
		}
	}

	// 3. Get Used IP addresses in Edge Gateway in the same slice format
	usedIpSlice, err := flattenGatewayUsedIpAddressesToIpSlice(usedIpAddresses)
	if err != nil {
		return nil, fmt.Errorf("could not flatten Edge Gateway used IP addresses: %s", err)
	}

	// 4. Get all unused IPs
	// (allIPs - allUsedIPs) = allUnusedIPs
	unusedIps := ipSliceDifference(assignedIpSlice, usedIpSlice)

	return unusedIps, nil
}

func getUnusedExternalIPAddress(uplinks []types.EdgeGatewayUplinks, usedIpAddresses []*types.GatewayUsedIpAddress, requiredIpCount int, optionalSubnet netip.Prefix) ([]netip.Addr, error) {
	unusedIps, err := getAllUnusedExternalIPAddresses(uplinks, usedIpAddresses, optionalSubnet)
	if err != nil {
		return nil, fmt.Errorf("error getting all unused IPs: %s", err)
	}

	// 5. Check if 'requiredIpCount' criteria is met
	if len(unusedIps) < requiredIpCount {
		return nil, fmt.Errorf("not enough unused IPs found. Expected %d, got %d", requiredIpCount, len(unusedIps))
	}

	// 6. Return required amount of unused IPs
	return unusedIps[:requiredIpCount], nil
}

// flattenEdgeGatewayUplinkToIpSlice processes Edge Gateway Uplink structure and creates a slice of
// all available IPs
func flattenEdgeGatewayUplinkToIpSlice(uplinks []types.EdgeGatewayUplinks) ([]netip.Addr, error) {
	assignedIpSlice := make([]netip.Addr, 0)

	for _, edgeGatewayUplink := range uplinks {
		for _, edgeGatewayUplinkSubnet := range edgeGatewayUplink.Subnets.Values {
			for _, r := range edgeGatewayUplinkSubnet.IPRanges.Values {
				// Convert IPs to netip.Addr
				startIp, err := netip.ParseAddr(r.StartAddress)
				if err != nil {
					return nil, fmt.Errorf("error parsing start IP address in range '%s': %s", r.StartAddress, err)
				}

				// if we have end address specified - a range of IPs must be expanded into slice
				// with all IPs in that range
				if r.EndAddress != "" {
					endIp, err := netip.ParseAddr(r.EndAddress)
					if err != nil {
						return nil, fmt.Errorf("error parsing end IP address in range '%s': %s", r.EndAddress, err)
					}

					// Check if EndAddress is lower than StartAddress ant report an error if so
					if endIp.Less(startIp) {
						return nil, fmt.Errorf("end IP is lower that start IP (%s < %s)", r.EndAddress, r.StartAddress)
					}

					// loop over IPs in range from startIp to endIp and add them to the slice one by one
					// Expression 'ip.Compare(endIp) == 1'  means that 'ip > endIp' and the loop should stop
					for ip := startIp; ip.Compare(endIp) != 1; ip = ip.Next() {
						assignedIpSlice = append(assignedIpSlice, ip)
					}
				} else { // if there is no end address in the range, then it is only a single IP - startIp
					assignedIpSlice = append(assignedIpSlice, startIp)
				}
			}
		}
	}

	return assignedIpSlice, nil
}

// ipSliceDifference performs mathematical subtraction for two slices of IPs
// The formula is (minuend − subtrahend = difference)
//
// Special behavior:
// * Passing nil minuend results in nil
// * Passing nil subtrahend will return minuendSlice
//
// NOTE. This function will mutate minuendSlice to save memory and avoid having a copy of all values
// which can become expensive if there are a lot of items
func ipSliceDifference(minuendSlice, subtrahendSlice []netip.Addr) []netip.Addr {
	if minuendSlice == nil {
		return nil
	}

	if subtrahendSlice == nil {
		return minuendSlice
	}

	// Removal of elements from an empty slice results in an empty slice
	if len(minuendSlice) == 0 {
		return []netip.Addr{}
	}
	// Having an empty subtrahendSlice results in minuendSlice
	if len(subtrahendSlice) == 0 {
		return minuendSlice
	}

	resultIpCount := 0 // count of IPs after removing items from subtrahendSlice

	// Loop over minuend IPs
	for _, minuendIp := range minuendSlice {

		// Check if subtrahend has minuend element listed
		var foundSubtrahend bool
		for _, subtrahendIp := range subtrahendSlice {
			if subtrahendIp == minuendIp {
				// IP found in subtrahend, therefore breaking inner loop early
				foundSubtrahend = true
				break
			}
		}

		// Store the IP in `minuendSlice` at `resultIpCount` index and increment the index itself
		if !foundSubtrahend {
			// Add IP to the 'resultIpCount' index position
			minuendSlice[resultIpCount] = minuendIp
			resultIpCount++
		}
	}

	// if all elements are removed - return nil
	if resultIpCount == 0 {
		return nil
	}

	// cut off all values, greater than `resultIpCount`
	minuendSlice = minuendSlice[:resultIpCount]

	return minuendSlice
}

// filterIpSlicesBySubnet accepts 'ipRange' and returns a slice of IPs only that fall into given
// 'subnet' leaving everything out
//
// Special behavior:
// * Passing empty 'subnet' will return `nil` and an error
// * Passing empty 'ipRange' will return 'nil' and an error
//
// Note. This function does not enforce uniqueness of IPs in 'ipRange' and if there are duplicate
// IPs matching 'subnet' they will be in the resulting slice
func filterIpSlicesBySubnet(ipRange []netip.Addr, subnet netip.Prefix) ([]netip.Addr, error) {
	if subnet == (netip.Prefix{}) {
		return nil, fmt.Errorf("empty subnet specified")
	}

	if len(ipRange) == 0 {
		return nil, fmt.Errorf("empty IP Range specified")
	}

	filteredRange := make([]netip.Addr, 0)

	for _, ip := range ipRange {
		if subnet.Contains(ip) {
			filteredRange = append(filteredRange, ip)
		}
	}

	return filteredRange, nil
}

// flattenGatewayUsedIpAddressesToIpSlice accepts a slice of `GatewayUsedIpAddress` coming directly
// from the API and converts it to slice of Go's native '[]netip.Addr' which supports IPv4 and IPv6
func flattenGatewayUsedIpAddressesToIpSlice(usedIpAddresses []*types.GatewayUsedIpAddress) ([]netip.Addr, error) {
	usedIpSlice := make([]netip.Addr, len(usedIpAddresses))
	for usedIpIndex := range usedIpAddresses {
		ip, err := netip.ParseAddr(usedIpAddresses[usedIpIndex].IPAddress)
		if err != nil {
			return nil, fmt.Errorf("error parsing IP '%s' in Edge Gateway used IP address list: %s", usedIpAddresses[usedIpIndex].IPAddress, err)
		}
		usedIpSlice[usedIpIndex] = ip
	}

	return usedIpSlice, nil
}
