/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/netip"
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

// GetUsedIpAddresses uses dedicated endpoint to retrieve Used IP addresses in an Edge Gateway
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

// GetUnassignedExternalIPAddresses will retrieve a requiredIpCount of unassigned IP addresses for
// Edge Gateway
// Arguments:
// * `requiredIpCount` (how many unallocated IPs should be returned). It will fail and return an
// error if IP all IPs specified in 'requiredIpCount' cannot be found.
// * `optionalSubnet` is specified (CIDR notation, e.g. 192.168.1.0/24) - it will look for an IP in
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
// 3. Retrieves all used IP addresses in Edge Gateway
// 4. Subtracts used IP addresses from available list of IPs in uplink (optionally filtered by optionalSubnet in step 2)
// 5. Checks if 'requiredIpCount' criteria is met, returns error otherwise
// 6. Returns required amount of unallocated IPs (as defined in 'requiredIpCount')
//
// Notes:
// * This function uses Go's builtin `netip` package to avoid any string processing of IPs and
// supports IPv4 and IPv6 addressing.
// * If an unused IP is not found it will return 'netip.Addr{}' (not using *netip.Addr{} to match
// library semantics) and an error
// * It will return an error if any of uplink IP ranges End IP address is lower than Start IP
// address
func (egw *NsxtEdgeGateway) GetUnassignedExternalIPAddresses(requiredIpCount int, optionalSubnet netip.Prefix, refresh bool) ([]netip.Addr, error) {
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

	return getNusedExternalIPAddress(egw.EdgeGateway.EdgeGatewayUplinks, usedIpAddresses, requiredIpCount, optionalSubnet)
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

// getNusedExternalIPAddress kept separate from data lookup in
// GetUnusedExternalIPAddresses to aid testing. It performs actions which are documented in
// public function.
func getNusedExternalIPAddress(uplinks []types.EdgeGatewayUplinks, usedIpAddresses []*types.GatewayUsedIpAddress, requiredIpCount int, optionalSubnet netip.Prefix) ([]netip.Addr, error) {
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

	// 3. Get Used IP addresses in Edge Gateway in the same slice
	usedIpSlice, err := flattenGatewayUsedIpAddressesToIpSlice(usedIpAddresses)
	if err != nil {
		return nil, fmt.Errorf("could not flatten Edge Gateway used IP addresses: %s", err)
	}

	// 4. Get all unallocated IPs
	// (allIPs - allUsedIPs) = allUnallocatedIPs
	unallocatedIps := ipSliceDifference(assignedIpSlice, usedIpSlice)

	// 5. Check if 'requiredIpCount' criteria is met
	if len(unallocatedIps) < requiredIpCount {
		return nil, fmt.Errorf("not enough unallocated IPs found. Expected %d, got %d", requiredIpCount, len(unallocatedIps))
	}

	// 6. Return required amount of unallocated IPs
	return unallocatedIps[:requiredIpCount], nil
}

// flattenEdgeGatewayUplinkToIpSlice processes Edge Gateway Uplink structure and creates a slice of all
// available IPs
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

				// if we have end address specified - a range of IPs must be expanded into the slice
				if r.EndAddress != "" {
					endIp, err := netip.ParseAddr(r.EndAddress)
					if err != nil {
						return nil, fmt.Errorf("error parsing end IP address in range '%s': %s", r.EndAddress, err)
					}

					// Check if EndAddress is lower than StartAddress ant report an error if so
					if endIp.Less(startIp) {
						return nil, fmt.Errorf("end IP is lower that start IP (%s < %s)", r.EndAddress, r.StartAddress)
					}

					// loop over IPs in range
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
	// Having an empty subtrahendSlice results in minuendSlice (persisting )
	if len(subtrahendSlice) == 0 {
		return minuendSlice
	}

	var difference []netip.Addr

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

		// Store the IP in difference when subtrahend does not contain IP of minuend
		if !foundSubtrahend {
			// Add IP to the resulting difference slice
			difference = append(difference, minuendIp)
		}
	}

	return difference
}

// filterIpSlicesBySubnet accepts 'ipRange' and returns a slice of IPs only that fall into given
// 'subnet' leaving everything out
//
// Special behavior:
// * Passing empty 'subnet' will return `nil` and an error
// * Pasing empty 'ipRange' will return 'nil' and an error
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

// flattenGatewayUsedIpAddressesToIpSlice accepts a slice of `GatewayUsedIpAddress` comming directly
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
