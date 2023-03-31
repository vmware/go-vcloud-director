/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// OpenApiOrgVdcNetworkDhcpBinding handles IPv4 and IPv6 DHCP bindings for NSX-T Org VDC networks.
// Note. For this to work, DHCP must be enabled on the network (see `OpenApiOrgVdcNetworkDhcp`)
type OpenApiOrgVdcNetworkDhcpBinding struct {
	OpenApiOrgVdcNetworkDhcpBinding *types.OpenApiOrgVdcNetworkDhcpBinding
	client                          *Client
	// parentOrgVdcNetworkId is used to construct the URL for the DHCP binding as it contains Org
	// VDC network ID in the path
	parentOrgVdcNetworkId string
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

// GetOpenApiOrgVdcNetworkDhcpBinding allows to retrieve all DHCP binding configurations for
// specific Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) GetAllOpenApiOrgVdcNetworkDhcpBindings() ([]*OpenApiOrgVdcNetworkDhcpBinding, error) {
	if orgVdcNet == nil || orgVdcNet.client == nil {
		return nil, fmt.Errorf("error - Org VDC network and client cannot be nil")
	}

	client := orgVdcNet.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if orgVdcNet.OpenApiOrgVdcNetwork.ID == "" {
		return nil, fmt.Errorf("empty Org VDC network ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgVdcNet.OpenApiOrgVdcNetwork.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.OpenApiOrgVdcNetworkDhcpBinding{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, nil, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into OpenApiOrgVdcNetwork types with client
	wrappedResponses := make([]*OpenApiOrgVdcNetworkDhcpBinding, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &OpenApiOrgVdcNetworkDhcpBinding{
			OpenApiOrgVdcNetworkDhcpBinding: typeResponses[sliceIndex],
			client:                          client,
			parentOrgVdcNetworkId:           orgVdcNet.OpenApiOrgVdcNetwork.ID,
		}
	}

	return wrappedResponses, nil
}

// GetOpenApiOrgVdcNetworkDhcpBinding allows to retrieve DHCP binding configuration
func (orgVdcNet *OpenApiOrgVdcNetwork) GetOpenApiOrgVdcNetworkDhcpBindingById(id string) (*OpenApiOrgVdcNetworkDhcpBinding, error) {
	if orgVdcNet == nil || orgVdcNet.client == nil {
		return nil, fmt.Errorf("error ")
	}

	client := orgVdcNet.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if orgVdcNet.OpenApiOrgVdcNetwork.ID == "" {
		return nil, fmt.Errorf("empty Org VDC network ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgVdcNet.OpenApiOrgVdcNetwork.ID), id)
	if err != nil {
		return nil, err
	}

	orgNetDhcp := &OpenApiOrgVdcNetworkDhcpBinding{
		OpenApiOrgVdcNetworkDhcpBinding: &types.OpenApiOrgVdcNetworkDhcpBinding{},
		client:                          client,
		parentOrgVdcNetworkId:           orgVdcNet.OpenApiOrgVdcNetwork.ID,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, orgNetDhcp.OpenApiOrgVdcNetworkDhcpBinding, nil)
	if err != nil {
		return nil, err
	}

	return orgNetDhcp, nil
}

// UpdateOpenApiOrgVdcNetworkDhcpBinding allows to update DHCP configuration
func (dhcpBinding *OpenApiOrgVdcNetworkDhcpBinding) UpdateOpenApiOrgVdcNetworkDhcpBinding(orgVdcNetworkDhcpConfig *types.OpenApiOrgVdcNetworkDhcpBinding) (*OpenApiOrgVdcNetworkDhcpBinding, error) {
	client := dhcpBinding.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if orgVdcNetworkDhcpConfig.ID == "" {
		return nil, fmt.Errorf("empty Org VDC network DHCP binding ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dhcpBinding.parentOrgVdcNetworkId), orgVdcNetworkDhcpConfig.ID)
	if err != nil {
		return nil, err
	}

	result := &OpenApiOrgVdcNetworkDhcpBinding{
		OpenApiOrgVdcNetworkDhcpBinding: &types.OpenApiOrgVdcNetworkDhcpBinding{},
		client:                          client,
		parentOrgVdcNetworkId:           dhcpBinding.parentOrgVdcNetworkId,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, orgVdcNetworkDhcpConfig, result.OpenApiOrgVdcNetworkDhcpBinding, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Org VDC network DHCP configuration with ID '%s': %s", orgVdcNetworkDhcpConfig.ID, err)
	}

	return result, nil
}

// DeleteOpenApiOrgVdcNetworkDhcpBinding allows to perform HTTP DELETE request on DHCP binding
func (dhcpBinding *OpenApiOrgVdcNetworkDhcpBinding) DeleteOpenApiOrgVdcNetworkDhcpBinding() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcpBindings
	apiVersion, err := dhcpBinding.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := dhcpBinding.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dhcpBinding.parentOrgVdcNetworkId), dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)
	if err != nil {
		return err
	}

	err = dhcpBinding.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting Org VDC network DHCP configuration: %s", err)
	}

	return nil
}
