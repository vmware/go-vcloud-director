/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// OpenApiOrgVdcNetwork uses OpenAPI endpoint to operate both - NSX-T and NSX-V Org VDC network DHCP settings
type OpenApiOrgVdcNetworkDhcp struct {
	OpenApiOrgVdcNetworkDhcp *types.OpenApiOrgVdcNetworkDhcp
	client                   *Client
}

// GetOpenApiOrgVdcNetworkDhcp allows to retrieve DHCP configuration for specific Org VDC network
func (orgVdcNet *OpenApiOrgVdcNetwork) GetOpenApiOrgVdcNetworkDhcp() (*OpenApiOrgVdcNetworkDhcp, error) {
	if orgVdcNet == nil || orgVdcNet.client == nil {
		return nil, fmt.Errorf("error - Org VDC network and client cannot be nil")
	}

	client := orgVdcNet.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp
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

	orgNetDhcp := &OpenApiOrgVdcNetworkDhcp{
		OpenApiOrgVdcNetworkDhcp: &types.OpenApiOrgVdcNetworkDhcp{},
		client:                   client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, orgNetDhcp.OpenApiOrgVdcNetworkDhcp, nil)
	if err != nil {
		return nil, err
	}

	return orgNetDhcp, nil
}

// GetOpenApiOrgVdcNetworkDhcp allows to retrieve DHCP configuration for specific Org VDC network
// ID specified as orgNetworkId using OpenAPI
func (vdc *Vdc) GetOpenApiOrgVdcNetworkDhcp(orgNetworkId string) (*OpenApiOrgVdcNetworkDhcp, error) {

	client := vdc.client
	// Inject Vdc ID filter to perform filtering on server side
	params := url.Values{}
	queryParameters := queryParameterFilterAnd("orgVdc.id=="+vdc.Vdc.ID, params)

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if orgNetworkId == "" {
		return nil, fmt.Errorf("empty Org VDC network ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgNetworkId))
	if err != nil {
		return nil, err
	}

	orgNetDhcp := &OpenApiOrgVdcNetworkDhcp{
		OpenApiOrgVdcNetworkDhcp: &types.OpenApiOrgVdcNetworkDhcp{},
		client:                   client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, queryParameters, orgNetDhcp.OpenApiOrgVdcNetworkDhcp, nil)
	if err != nil {
		return nil, err
	}

	return orgNetDhcp, nil
}

// UpdateOpenApiOrgVdcNetworkDhcp allows to update DHCP configuration for specific Org VDC network
// ID specified as orgNetworkId using OpenAPI
func (vdc *Vdc) UpdateOpenApiOrgVdcNetworkDhcp(orgNetworkId string, orgVdcNetworkDhcpConfig *types.OpenApiOrgVdcNetworkDhcp) (*OpenApiOrgVdcNetworkDhcp, error) {
	return updateOrgNetworkDhcp(vdc.client, orgNetworkId, orgVdcNetworkDhcpConfig)
}

// DeleteOpenApiOrgVdcNetworkDhcp allows to perform HTTP DELETE request on DHCP pool configuration for specified Org VDC
// Network ID
func (vdc *Vdc) DeleteOpenApiOrgVdcNetworkDhcp(orgNetworkId string) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp
	apiVersion, err := vdc.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if orgNetworkId == "" {
		return fmt.Errorf("cannot delete Org VDC network DHCP configuration without ID")
	}

	urlRef, err := vdc.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgNetworkId))
	if err != nil {
		return err
	}

	err = vdc.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting Org VDC network DHCP configuration: %s", err)
	}

	return nil
}

// DeletNetworkDhcp allows to perform HTTP DELETE request on DHCP pool configuration for Org network
func (orgVdcNet *OpenApiOrgVdcNetwork) DeletNetworkDhcp() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp
	apiVersion, err := orgVdcNet.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if orgVdcNet.OpenApiOrgVdcNetwork.ID == "" {
		return fmt.Errorf("cannot delete Org VDC network DHCP configuration without ID")
	}

	urlRef, err := orgVdcNet.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgVdcNet.OpenApiOrgVdcNetwork.ID))
	if err != nil {
		return err
	}

	err = orgVdcNet.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting Org VDC network DHCP configuration: %s", err)
	}

	return nil
}

func updateOrgNetworkDhcp(client *Client, orgNetworkId string, orgVdcNetworkDhcpConfig *types.OpenApiOrgVdcNetworkDhcp) (*OpenApiOrgVdcNetworkDhcp, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworksDhcp
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, orgNetworkId))
	if err != nil {
		return nil, err
	}

	orgNetDhcpResponse := &OpenApiOrgVdcNetworkDhcp{
		OpenApiOrgVdcNetworkDhcp: &types.OpenApiOrgVdcNetworkDhcp{},
		client:                   client,
	}

	// From v35.0 onwards, if orgVdcNetworkDhcpConfig.LeaseTime or orgVdcNetworkDhcpConfig.Mode are not explicitly
	// passed, the API doesn't use any defaults returning an error. Previous API versions were setting
	// LeaseTime to 86400 seconds and Mode to EDGE if these values were not supplied. These two conditional
	// address the situation.
	if orgVdcNetworkDhcpConfig.LeaseTime == nil {
		orgVdcNetworkDhcpConfig.LeaseTime = addrOf(86400)
	}

	if len(orgVdcNetworkDhcpConfig.Mode) == 0 {
		orgVdcNetworkDhcpConfig.Mode = "EDGE"
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, orgVdcNetworkDhcpConfig, orgNetDhcpResponse.OpenApiOrgVdcNetworkDhcp, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Org VDC network DHCP configuration: %s", err)
	}

	return orgNetDhcpResponse, nil
}
