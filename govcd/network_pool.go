/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"strings"
)

type NetworkPool struct {
	NetworkPool *types.NetworkPool
	vcdClient   *VCDClient
}

var backingUseErrorMessages = map[types.BackingUseConstraint]string{
	types.BackingUseExplicit:       "no element named %s found",
	types.BackingUseWhenOnlyOne:    "no single element found for this backing",
	types.BackingUseFirstAvailable: "no elements found for this backing",
}

// GetOpenApiUrl retrieves the full URL of a network pool
func (np *NetworkPool) GetOpenApiUrl() (string, error) {
	response, err := url.JoinPath(np.vcdClient.sessionHREF.String(), "admin", "extension", "networkPool", np.NetworkPool.Id)
	if err != nil {
		return "", err
	}
	return response, nil
}

// GetNetworkPoolSummaries retrieves the list of all available network pools
func (vcdClient *VCDClient) GetNetworkPoolSummaries(queryParameters url.Values) ([]*types.NetworkPool, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPoolSummaries
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	typeResponse := []*types.NetworkPool{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponse, nil)
	if err != nil {
		return nil, err
	}

	return typeResponse, nil
}

// GetNetworkPoolById retrieves Network Pool with a given ID
func (vcdClient *VCDClient) GetNetworkPoolById(id string) (*NetworkPool, error) {
	if id == "" {
		return nil, fmt.Errorf("network pool lookup requires ID")
	}

	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	response := &NetworkPool{
		vcdClient:   vcdClient,
		NetworkPool: &types.NetworkPool{},
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, response.NetworkPool, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetNetworkPoolByName retrieves a network pool with a given name
// Note. It will return an error if multiple network pools exist with the same name
func (vcdClient *VCDClient) GetNetworkPoolByName(name string) (*NetworkPool, error) {
	if name == "" {
		return nil, fmt.Errorf("network pool lookup requires name")
	}

	queryParameters := url.Values{}
	queryParameters.Add("filter", "name=="+name)

	filteredNetworkPools, err := vcdClient.GetNetworkPoolSummaries(queryParameters)
	if err != nil {
		return nil, fmt.Errorf("error getting network pools: %s", err)
	}

	if len(filteredNetworkPools) == 0 {
		return nil, fmt.Errorf("no network pool found with name '%s' - %s", name, ErrorEntityNotFound)
	}

	if len(filteredNetworkPools) > 1 {
		return nil, fmt.Errorf("more than one network pool found with name '%s'", name)
	}

	return vcdClient.GetNetworkPoolById(filteredNetworkPools[0].Id)
}

// CreateNetworkPool creates a network pool using the given configuration
// It can create any type of network pool
func (vcdClient *VCDClient) CreateNetworkPool(config *types.NetworkPool) (*NetworkPool, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &NetworkPool{
		NetworkPool: &types.NetworkPool{},
		vcdClient:   vcdClient,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, config, result.NetworkPool, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update will change all changeable network pool items
func (np *NetworkPool) Update() error {
	if np == nil || np.NetworkPool == nil || np.NetworkPool.Id == "" {
		return fmt.Errorf("network pool must have ID")
	}
	if np.vcdClient == nil || np.vcdClient.Client.APIVersion == "" {
		return fmt.Errorf("network pool '%s': no client found", np.NetworkPool.Name)
	}

	client := np.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, np.NetworkPool.Id)
	if err != nil {
		return err
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, np.NetworkPool, np.NetworkPool, nil)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("error updating network pool '%s': %s", np.NetworkPool.Name, err)
	}

	return nil
}

// Delete removes a network pool
func (np *NetworkPool) Delete() error {
	if np == nil || np.NetworkPool == nil || np.NetworkPool.Id == "" {
		return fmt.Errorf("network pool must have ID")
	}
	if np.vcdClient == nil || np.vcdClient.Client.APIVersion == "" {
		return fmt.Errorf("network pool '%s': no client found", np.NetworkPool.Name)
	}

	client := np.vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNetworkPools
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, np.NetworkPool.Id)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("error deleting network pool '%s': %s", np.NetworkPool.Name, err)
	}

	return nil
}

func backingErrorMessage(constraint types.BackingUseConstraint, name string) string {
	errorMessage := fmt.Sprintf("[constraint: %s] %s", constraint, backingUseErrorMessages[constraint])
	if strings.Contains(errorMessage, "%s") {
		return fmt.Sprintf(errorMessage, name)
	}
	return errorMessage
}

type getElementFunc[B any] func(*B) string
type validElementFunc[B any] func(*B) bool

// chooseBackingElement will select a backing element from a list, using the given constraint
// * constraint is the type of choice we are looking for
// * wantedName is the name of the element we want. If we use a constraint other than types.BackingUseExplicit, it can be empty
// * elements is the list of backing elements we want to choose from
// * getEl is a function that, given an element, returns its name
// * validateEl is an optional function that tells whether a given element is valid or not. If missing, we assume all elements are valid
func chooseBackingElement[B any](constraint types.BackingUseConstraint, wantedName string, elements []*B, getEl getElementFunc[B], validateEl validElementFunc[B]) (*B, error) {
	var searchedElement *B
	if validateEl == nil {
		validateEl = func(*B) bool { return true }
	}
	numberOfValidElements := 0
	// We need to pre-calculate the number of valid elements, to use it when constraint == BackingUseWhenOnlyOne
	for _, element := range elements {
		if validateEl(element) {
			numberOfValidElements++
		}
	}
	// availableElements will contain the list of available elements, to be used in error messages
	var availableElements []string
	for _, element := range elements {
		elementName := getEl(element)
		if !validateEl(element) {
			continue
		}
		availableElements = append(availableElements, elementName)

		switch constraint {
		case types.BackingUseExplicit:
			// When asking for a specific element explicitly, we return it only if the name matches the request)
			if wantedName == elementName {
				searchedElement = element
			}
		case types.BackingUseWhenOnlyOne:
			// With BackingUseWhenOnlyOne, we return the element only if there is a single *valid* element in the list
			if wantedName == "" && numberOfValidElements == 1 {
				searchedElement = element
			}
		case types.BackingUseFirstAvailable:
			// This is the most permissive constraint: we get the first available element
			if wantedName == "" {
				searchedElement = element
			}
		}
		if searchedElement != nil {
			break
		}
	}
	// If no item was retrieved, we build an error message appropriate for the current constraint, and add
	// the list of available elements to it
	if searchedElement == nil {
		return nil, fmt.Errorf(backingErrorMessage(constraint, wantedName)+" - available elements: %v", availableElements)
	}

	// When we reach this point, we have found what was requested, and return the element
	return searchedElement, nil
}

// CreateNetworkPoolGeneve creates a network pool of GENEVE type
// The function retrieves the given NSX-T manager and corresponding transport zone names
// If the transport zone name is empty, the first available will be used
func (vcdClient *VCDClient) CreateNetworkPoolGeneve(name, description, nsxtManagerName, transportZoneName string, constraint types.BackingUseConstraint) (*NetworkPool, error) {
	managers, err := vcdClient.QueryNsxtManagerByName(nsxtManagerName)
	if err != nil {
		return nil, err
	}

	if len(managers) == 0 {
		return nil, fmt.Errorf("no manager '%s' found", nsxtManagerName)
	}
	if len(managers) > 1 {
		return nil, fmt.Errorf("more than one manager '%s' found", nsxtManagerName)
	}
	manager := managers[0]

	managerId := "urn:vcloud:nsxtmanager:" + extractUuid(managers[0].HREF)
	transportZones, err := vcdClient.GetAllNsxtTransportZones(managerId, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving transport zones for manager '%s': %s", manager.Name, err)
	}
	transportZone, err := chooseBackingElement[types.TransportZone](
		constraint,
		transportZoneName,
		transportZones,
		func(tz *types.TransportZone) string { return tz.Name },
		func(tz *types.TransportZone) bool { return !tz.AlreadyImported },
	)

	if err != nil {
		return nil, err
	}
	if transportZone.AlreadyImported {
		return nil, fmt.Errorf("transport zone '%s' is already imported", transportZone.Name)
	}

	// Note: in this type of network pool, the managing owner is the NSX-T manager
	managingOwner := types.OpenApiReference{
		Name: manager.Name,
		ID:   managerId,
	}
	var config = &types.NetworkPool{
		Name:             name,
		Description:      description,
		PoolType:         types.NetworkPoolGeneveType,
		ManagingOwnerRef: managingOwner,
		Backing: types.NetworkPoolBacking{
			TransportZoneRef: types.OpenApiReference{
				ID:   transportZone.Id,
				Name: transportZone.Name,
			},
			ProviderRef: managingOwner,
		},
	}
	return vcdClient.CreateNetworkPool(config)
}

// CreateNetworkPoolPortGroup creates a network pool of PORTGROUP_BACKED type
// The function retrieves the given vCenter and corresponding port group names
// If the port group name is empty, the first available will be used
func (vcdClient *VCDClient) CreateNetworkPoolPortGroup(name, description, vCenterName string, portgroupNames []string, constraint types.BackingUseConstraint) (*NetworkPool, error) {
	vCenter, err := vcdClient.GetVCenterByName(vCenterName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vCenter '%s': %s", vCenterName, err)
	}
	var params = make(url.Values)
	params.Set("filter", "virtualCenter.id=="+vCenter.VSphereVCenter.VcId)
	portgroups, err := vcdClient.GetAllVcenterImportableDvpgs(params)
	if err != nil {
		return nil, fmt.Errorf("error retrieving portgroups for vCenter '%s': %s", vCenterName, err)
	}

	var chosenPortgroups []*VcenterImportableDvpg
	var chosenReferences []types.OpenApiReference
	for _, portgroupName := range portgroupNames {
		portGroup, err := chooseBackingElement[VcenterImportableDvpg](
			constraint,
			portgroupName,
			portgroups,
			func(v *VcenterImportableDvpg) string {
				return v.VcenterImportableDvpg.BackingRef.Name
			},
			nil,
		)

		if err != nil {
			return nil, err
		}
		chosenPortgroups = append(chosenPortgroups, portGroup)
		chosenReferences = append(chosenReferences, types.OpenApiReference{
			Name: portGroup.VcenterImportableDvpg.BackingRef.Name,
			ID:   portGroup.VcenterImportableDvpg.BackingRef.ID,
		})
	}

	if len(chosenPortgroups) == 0 {
		return nil, fmt.Errorf("no suitable portgroups found for names %v", portgroupNames)
	}
	if len(chosenPortgroups) > 1 {
		if !chosenPortgroups[0].UsableWith(chosenPortgroups...) {
			return nil, fmt.Errorf("portgroups %v should all belong to the same host", portgroupNames)
		}
	}

	// Note: in this type of network pool, the managing owner is the vCenter
	managingOwner := types.OpenApiReference{
		Name: vCenter.VSphereVCenter.Name,
		ID:   vCenter.VSphereVCenter.VcId,
	}
	config := types.NetworkPool{
		Name:             name,
		Description:      description,
		PoolType:         types.NetworkPoolPortGroupType,
		ManagingOwnerRef: managingOwner,
		Backing: types.NetworkPoolBacking{
			PortGroupRefs: chosenReferences,
			ProviderRef:   managingOwner,
		},
	}
	return vcdClient.CreateNetworkPool(&config)
}

// CreateNetworkPoolVlan creates a network pool of VLAN type
// The function retrieves the given vCenter and corresponding distributed switch names
// If the distributed switch name is empty, the first available will be used
func (vcdClient *VCDClient) CreateNetworkPoolVlan(name, description, vCenterName, dsName string, ranges []types.VlanIdRange, constraint types.BackingUseConstraint) (*NetworkPool, error) {
	vCenter, err := vcdClient.GetVCenterByName(vCenterName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vCenter '%s': %s", vCenterName, err)
	}

	dswitches, err := vcdClient.GetAllVcenterDistributedSwitches(vCenter.VSphereVCenter.VcId, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving distributed switches for vCenter '%s': %s", vCenterName, err)
	}

	dswitch, err := chooseBackingElement[types.VcenterDistributedSwitch](
		constraint,
		dsName,
		dswitches,
		func(t *types.VcenterDistributedSwitch) string { return t.BackingRef.Name },
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Note: in this type of network pool, the managing owner is the vCenter
	managingOwner := types.OpenApiReference{
		Name: vCenter.VSphereVCenter.Name,
		ID:   vCenter.VSphereVCenter.VcId,
	}
	config := types.NetworkPool{
		Name:             name,
		Description:      description,
		PoolType:         types.NetworkPoolVlanType,
		ManagingOwnerRef: managingOwner,
		Backing: types.NetworkPoolBacking{
			VlanIdRanges: types.VlanIdRanges{
				Values: ranges,
			},
			VdsRefs: []types.OpenApiReference{
				{
					Name: dswitch.BackingRef.Name,
					ID:   dswitch.BackingRef.ID,
				},
			},
			ProviderRef: managingOwner,
		},
	}
	return vcdClient.CreateNetworkPool(&config)
}
