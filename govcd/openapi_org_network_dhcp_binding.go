/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// OpenApiOrgVdcNetworkDhcpBinding handles IPv4 and IPv6 DHCP bindings for NSX-T Org VDC networks.
// Note. To create DHCP bindings, DHCP must be enabled on the network first (see
// `OpenApiOrgVdcNetworkDhcp`)
type OpenApiOrgVdcNetworkDhcpBinding struct {
	OpenApiOrgVdcNetworkDhcpBinding *types.OpenApiOrgVdcNetworkDhcpBinding
	client                          *Client
	// ParentOrgVdcNetworkId is used to construct the URL for the DHCP binding as it contains Org
	// VDC network ID in the path
	ParentOrgVdcNetworkId string
}

// CreateOpenApiOrgVdcNetworkDhcpBinding allows to create DHCP binding for specific Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) CreateOpenApiOrgVdcNetworkDhcpBinding(dhcpBindingConfig *types.OpenApiOrgVdcNetworkDhcpBinding) (*OpenApiOrgVdcNetworkDhcpBinding, error) {
	client := orgVdcNet.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgVdcNet.OpenApiOrgVdcNetwork.ID))
	if err != nil {
		return nil, err
	}

	// DHCP Binding endpoint returns ID of newly created object in `Details` field of the task,
	// which is not standard, therefore it must be explicitly handled in the code here
	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, dhcpBindingConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating Org VDC Network DHCP Binding: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error waiting for Org VDC Network DHCP Binding to be created: %s", err)
	}

	dhcpBindingId := task.Task.Details
	if dhcpBindingId == "" {
		return nil, fmt.Errorf("could not retrieve ID of newly created DHCP binding for Org VDC network with IP address '%s' and MAC address '%s'",
			dhcpBindingConfig.IpAddress, dhcpBindingConfig.MacAddress)
	}

	// Get the DHCP binding by ID and return it
	createdBinding, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcpBindingById(dhcpBindingId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving DHCP binding for Org VDC network after creation: %s", err)
	}

	return createdBinding, nil
}

// GetAllOpenApiOrgVdcNetworkDhcpBindings allows to retrieve all DHCP binding configurations for
// specific Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) GetAllOpenApiOrgVdcNetworkDhcpBindings(queryParameters url.Values) ([]*OpenApiOrgVdcNetworkDhcpBinding, error) {
	if orgVdcNet == nil || orgVdcNet.client == nil {
		return nil, fmt.Errorf("error - Org VDC network and client cannot be nil")
	}

	if orgVdcNet.OpenApiOrgVdcNetwork == nil || orgVdcNet.OpenApiOrgVdcNetwork.ID == "" {
		return nil, fmt.Errorf("empty Org VDC network ID")
	}

	client := orgVdcNet.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgVdcNet.OpenApiOrgVdcNetwork.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.OpenApiOrgVdcNetworkDhcpBinding{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into OpenApiOrgVdcNetworkDhcpBinding types with client
	wrappedResponses := make([]*OpenApiOrgVdcNetworkDhcpBinding, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &OpenApiOrgVdcNetworkDhcpBinding{
			OpenApiOrgVdcNetworkDhcpBinding: typeResponses[sliceIndex],
			client:                          client,
			ParentOrgVdcNetworkId:           orgVdcNet.OpenApiOrgVdcNetwork.ID,
		}
	}

	return wrappedResponses, nil
}

// GetOpenApiOrgVdcNetworkDhcpBindingById allows to retrieve DHCP binding configuration
func (orgVdcNet *OpenApiOrgVdcNetwork) GetOpenApiOrgVdcNetworkDhcpBindingById(id string) (*OpenApiOrgVdcNetworkDhcpBinding, error) {
	if orgVdcNet == nil || orgVdcNet.client == nil {
		return nil, fmt.Errorf("error - Org VDC network and client cannot be nil")
	}

	if id == "" {
		return nil, fmt.Errorf("empty DHCP binding ID")
	}

	if orgVdcNet.OpenApiOrgVdcNetwork == nil || orgVdcNet.OpenApiOrgVdcNetwork.ID == "" {
		return nil, fmt.Errorf("empty Org VDC network ID")
	}

	client := orgVdcNet.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgVdcNet.OpenApiOrgVdcNetwork.ID), id)
	if err != nil {
		return nil, err
	}

	orgVdcNetDhcpBinding := &OpenApiOrgVdcNetworkDhcpBinding{
		OpenApiOrgVdcNetworkDhcpBinding: &types.OpenApiOrgVdcNetworkDhcpBinding{},
		client:                          client,
		ParentOrgVdcNetworkId:           orgVdcNet.OpenApiOrgVdcNetwork.ID,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, orgVdcNetDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding, nil)
	if err != nil {
		return nil, err
	}

	return orgVdcNetDhcpBinding, nil
}

// GetOpenApiOrgVdcNetworkDhcpBindingByName allows to retrieve DHCP binding configuration by name
func (orgVdcNet *OpenApiOrgVdcNetwork) GetOpenApiOrgVdcNetworkDhcpBindingByName(name string) (*OpenApiOrgVdcNetworkDhcpBinding, error) {
	// TODO: uncomment when filtering by name is supported in VCD API (It was not supported up to
	// VCD10.4.1)
	// Perform filtering by name in VCD API
	// queryParameters := url.Values{}
	// queryParameters.Add("filter", fmt.Sprintf("name==%s", name))

	// allBindings, err := orgVdcNet.GetAllOpenApiOrgVdcNetworkDhcpBindings(queryParameters)

	allDhcpBindings, err := orgVdcNet.GetAllOpenApiOrgVdcNetworkDhcpBindings(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Org VDC network by name '%s': %s", name, err)
	}

	// Bindings do not support name filtering, so we need to filter them manually
	var foundBinding *OpenApiOrgVdcNetworkDhcpBinding
	for _, binding := range allDhcpBindings {
		if binding.OpenApiOrgVdcNetworkDhcpBinding.Name == name {
			foundBinding = binding
			break
		}
	}

	if foundBinding == nil {
		return nil, fmt.Errorf("%s: could not find NSX-T Org Network Binding by Name %s", ErrorEntityNotFound, name)
	}

	return foundBinding, nil
}

// Update allows to update DHCP configuration
//
// Note. This API requires `Version` field to be sent in the request and this function does it
// automatically
func (dhcpBinding *OpenApiOrgVdcNetworkDhcpBinding) Update(orgVdcNetworkDhcpConfig *types.OpenApiOrgVdcNetworkDhcpBinding) (*OpenApiOrgVdcNetworkDhcpBinding, error) {
	client := dhcpBinding.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if orgVdcNetworkDhcpConfig.ID == "" {
		return nil, fmt.Errorf("empty Org VDC network DHCP binding ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dhcpBinding.ParentOrgVdcNetworkId), orgVdcNetworkDhcpConfig.ID)
	if err != nil {
		return nil, err
	}

	result := &OpenApiOrgVdcNetworkDhcpBinding{
		OpenApiOrgVdcNetworkDhcpBinding: &types.OpenApiOrgVdcNetworkDhcpBinding{ID: orgVdcNetworkDhcpConfig.ID},
		client:                          client,
		ParentOrgVdcNetworkId:           dhcpBinding.ParentOrgVdcNetworkId,
	}

	// load latest binding information to fetch Version value which is required for updates
	err = result.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing Org VDC network DHCP binding configuration with ID '%s': %s", orgVdcNetworkDhcpConfig.ID, err)
	}
	orgVdcNetworkDhcpConfig.Version = result.OpenApiOrgVdcNetworkDhcpBinding.Version

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, orgVdcNetworkDhcpConfig, result.OpenApiOrgVdcNetworkDhcpBinding, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Org VDC network DHCP configuration with ID '%s': %s", orgVdcNetworkDhcpConfig.ID, err)
	}

	return result, nil
}

// Refresh DHCP binding configuration. Mainly useful for retrieving latest `Version` field` of DHCP
// binding before performing update
func (dhcpBinding *OpenApiOrgVdcNetworkDhcpBinding) Refresh() error {
	if dhcpBinding.ParentOrgVdcNetworkId == "" {
		return fmt.Errorf("empty parent Org VDC network ID")
	}

	if dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID == "" {
		return fmt.Errorf("empty DHCP binding ID")
	}

	client := dhcpBinding.client
	orgVdcNet, err := getOpenApiOrgVdcNetworkById(client, dhcpBinding.ParentOrgVdcNetworkId, nil)
	if err != nil {
		return fmt.Errorf("error refreshing Org VDC network DHCP binding configuration with ID '%s': %s", dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID, err)
	}

	newDhcpBinding, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcpBindingById(dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)
	if err != nil {
		return fmt.Errorf("error refreshing Org VDC network DHCP binding configuration with ID '%s': %s", dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID, err)
	}

	// Explicitly reassign the body
	dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding = newDhcpBinding.OpenApiOrgVdcNetworkDhcpBinding

	return nil
}

// Delete removes DHCP binding by performing HTTP DELETE request on DHCP binding
func (dhcpBinding *OpenApiOrgVdcNetworkDhcpBinding) Delete() error {
	if dhcpBinding.ParentOrgVdcNetworkId == "" {
		return fmt.Errorf("empty parent Org VDC network ID")
	}

	if dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID == "" {
		return fmt.Errorf("empty DHCP binding ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := dhcpBinding.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := dhcpBinding.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dhcpBinding.ParentOrgVdcNetworkId), dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)
	if err != nil {
		return err
	}

	err = dhcpBinding.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting Org VDC network DHCP configuration: %s", err)
	}

	return nil
}
