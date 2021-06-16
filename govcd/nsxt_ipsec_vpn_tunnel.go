/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// NsxtIpSecVpnTunnel supports site-to-site policy-based IPsec VPN between an NSX-T Data Center Edge Gateway instance
// and a remote site.
// IPsec VPN offers site-to-site connectivity between an Edge Gateway and remote sites which also use NSX-T Data Center
// or which have either third-party hardware routers or VPN gateways that support IPsec.
// Policy-based IPsec VPN requires a VPN policy to be applied to packets to determine which traffic is to be protected
// by IPsec before being passed through a VPN tunnel. This type of VPN is considered static because when a local network
// topology and configuration change, the VPN policy settings must also be updated to accommodate the changes.
// NSX-T Data Center Edge Gateways support split tunnel configuration, with IPsec traffic taking routing precedence.
// VMware Cloud Director supports automatic route redistribution when you use IPsec VPN on an NSX-T edge gateway.
type NsxtIpSecVpnTunnel struct {
	NsxtIpSecVpn *types.NsxtIpSecVpnTunnel
	client       *Client
	// edgeGatewayId is stored here so that pointer receiver functions can embed edge gateway ID into path
	edgeGatewayId string
}

// GetAllIpSecVpns returns all IPsec VPN configurations
func (egw *NsxtEdgeGateway) GetAllIpSecVpns(queryParameters url.Values) ([]*NsxtIpSecVpnTunnel, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpn
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtIpSecVpnTunnel{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtIpSecVpnTunnel types with client
	wrappedResponses := make([]*NsxtIpSecVpnTunnel, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtIpSecVpnTunnel{
			NsxtIpSecVpn:  typeResponses[sliceIndex],
			client:        client,
			edgeGatewayId: egw.EdgeGateway.ID,
		}
	}

	return wrappedResponses, nil
}

func (egw *NsxtEdgeGateway) GetIpSecVpnById(id string) (*NsxtIpSecVpnTunnel, error) {
	if id == "" {
		return nil, fmt.Errorf("canot find NSX-T IPsec VPN configuration without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpn
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), id)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtIpSecVpnTunnel{
		NsxtIpSecVpn:  &types.NsxtIpSecVpnTunnel{},
		client:        client,
		edgeGatewayId: egw.EdgeGateway.ID,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, returnObject.NsxtIpSecVpn)
	if err != nil {
		return nil, err
	}

	return returnObject, nil
}

func (egw *NsxtEdgeGateway) GetIpSecVpnByName(name string) (*NsxtIpSecVpnTunnel, error) {
	if name == "" {
		return nil, fmt.Errorf("canot find NSX-T IPsec VPN configuration without Name")
	}

	allVpns, err := egw.GetAllIpSecVpns(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all NSX-T IPsec VPN configurations: %s", err)
	}

	var allResults []*NsxtIpSecVpnTunnel

	for _, vpnConfig := range allVpns {
		if vpnConfig.NsxtIpSecVpn.Name == name {
			allResults = append(allResults, vpnConfig)
		}
	}

	if len(allResults) > 1 {
		return nil, fmt.Errorf("error - found %d NSX-T IPsec VPN configuratios with Name '%s'. Expected 1", len(allResults), name)
	}

	if len(allResults) == 0 {
		return nil, ErrorEntityNotFound
	}

	// Retrieving the object by ID, because only it includes Pre-shared Key
	return egw.GetIpSecVpnById(allResults[0].NsxtIpSecVpn.ID)
}

func (egw *NsxtEdgeGateway) CreateIpSecVpn(ipSecVpnConfig *types.NsxtIpSecVpnTunnel) (*NsxtIpSecVpnTunnel, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpn
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	task, err := client.OpenApiPostItemAsync(minimumApiVersion, urlRef, nil, ipSecVpnConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T IPsec VPN configuration: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("task failed while creating NSX-T IPsec VPN configuration: %s", err)
	}

	// filtering even by Name is not supported
	allVpns, err := egw.GetAllIpSecVpns(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all NSX-T IPsec VPN configuration after creation: %s", err)
	}

	for index, singleConfig := range allVpns {
		if singleConfig.IsEqualTo(ipSecVpnConfig) {
			// retrieve exact value by ID, because only this endpoint includes private key
			ipSecVpn, err := egw.GetIpSecVpnById(allVpns[index].NsxtIpSecVpn.ID)
			if err != nil {
				return nil, fmt.Errorf("error retrieving NSX-T IPsec VPN configuration: %s", err)
			}

			return ipSecVpn, nil
		}
	}

	return nil, fmt.Errorf("error finding NSX-T IPsec VPN configuration after creation: %s", ErrorEntityNotFound)
}

func (ipSecVpn *NsxtIpSecVpnTunnel) Update(ipSecVpnConfig *types.NsxtIpSecVpnTunnel) (*NsxtIpSecVpnTunnel, error) {
	client := ipSecVpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpn
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if ipSecVpn.NsxtIpSecVpn.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T IPsec VPN configuration without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSecVpn.edgeGatewayId), ipSecVpn.NsxtIpSecVpn.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtIpSecVpnTunnel{
		NsxtIpSecVpn:  &types.NsxtIpSecVpnTunnel{},
		client:        client,
		edgeGatewayId: ipSecVpn.edgeGatewayId,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, ipSecVpnConfig, returnObject.NsxtIpSecVpn)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T IPsec VPN configuration: %s", err)
	}

	return returnObject, nil
}

// Delete allows users to delete NSX-T Application Port Profile
func (ipSecVpn *NsxtIpSecVpnTunnel) Delete() error {
	client := ipSecVpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpn
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if ipSecVpn.NsxtIpSecVpn.ID == "" {
		return fmt.Errorf("cannot delete NSX-T IPsec VPN configuration without ID")
	}

	urlRef, err := ipSecVpn.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSecVpn.edgeGatewayId), ipSecVpn.NsxtIpSecVpn.ID)
	if err != nil {
		return err
	}

	err = ipSecVpn.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil)

	if err != nil {
		return fmt.Errorf("error deleting NSX-T IPsec VPN configuration: %s", err)
	}

	return nil
}

// GetStatus returns status of IPsec VPN Tunnel.
//
// Note. This is not being immediately populated and may appear after some time
func (ipSecVpn *NsxtIpSecVpnTunnel) GetStatus() (*types.NsxtIpSecVpnTunnelStatus, error) {
	client := ipSecVpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnStatus
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if ipSecVpn.NsxtIpSecVpn.ID == "" {
		return nil, fmt.Errorf("cannot get NSX-T IPsec VPN status without ID")
	}

	urlRef, err := ipSecVpn.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSecVpn.edgeGatewayId, ipSecVpn.NsxtIpSecVpn.ID))
	if err != nil {
		return nil, err
	}

	ipSecVpnTunnelStatus := &types.NsxtIpSecVpnTunnelStatus{}

	err = ipSecVpn.client.OpenApiGetItem(minimumApiVersion, urlRef, nil, ipSecVpnTunnelStatus)

	if err != nil {
		return nil, fmt.Errorf("error deleting NSX-T IPsec VPN configuration: %s", err)
	}

	return ipSecVpnTunnelStatus, nil
}

func (ipSecVpn *NsxtIpSecVpnTunnel) UpdateTunnelConnectionProperties(ipSecVpnTunnelConnectionProperties *types.NsxtIpSecVpnTunnelSecurityProfile) (*types.NsxtIpSecVpnTunnelSecurityProfile, error) {
	client := ipSecVpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnConnectionProperties
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if ipSecVpn.NsxtIpSecVpn.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T IPsec VPN Connection Properties without ID")
	}

	urlRef, err := ipSecVpn.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSecVpn.edgeGatewayId, ipSecVpn.NsxtIpSecVpn.ID))
	if err != nil {
		return nil, err
	}

	ipSecVpnTunnelProfile := &types.NsxtIpSecVpnTunnelSecurityProfile{}
	err = ipSecVpn.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, ipSecVpnTunnelConnectionProperties, ipSecVpnTunnelProfile)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T IPsec VPN Connection Properties: %s", err)
	}

	return ipSecVpnTunnelProfile, nil
}

func (ipSecVpn *NsxtIpSecVpnTunnel) GetTunnelConnectionProperties() (*types.NsxtIpSecVpnTunnelSecurityProfile, error) {
	client := ipSecVpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSecVpnConnectionProperties
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if ipSecVpn.NsxtIpSecVpn.ID == "" {
		return nil, fmt.Errorf("cannot get NSX-T IPsec VPN Connection Properties without ID")
	}

	urlRef, err := ipSecVpn.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, ipSecVpn.edgeGatewayId, ipSecVpn.NsxtIpSecVpn.ID))
	if err != nil {
		return nil, err
	}

	ipSecVpnTunnelProfile := &types.NsxtIpSecVpnTunnelSecurityProfile{}
	err = ipSecVpn.client.OpenApiGetItem(minimumApiVersion, urlRef, nil, ipSecVpnTunnelProfile)
	if err != nil {
		return nil, fmt.Errorf("error removing NSX-T IPsec VPN Connection Properties: %s", err)
	}

	return ipSecVpnTunnelProfile, nil
}

// IsEqualTo helps to find NSX-T IPsec Configuration
// Combination of LocalAddress and RemoteAddress has to be unique. This is a list of fields compared:
// * Name
// * Description
// * Enabled
// * LocalEndpoint.LocalAddress
// * RemoteEndpoint.RemoteAddress
func (ipSecVpn *NsxtIpSecVpnTunnel) IsEqualTo(vpnConfig *types.NsxtIpSecVpnTunnel) bool {
	return ipSetVpnRulesEqual(ipSecVpn.NsxtIpSecVpn, vpnConfig)
}

// ipSetVpnRulesEqual performs comparison of two rules to ease lookup. This is a list of fields compared:
//// * Name
//// * Description
//// * Enabled
//// * LocalEndpoint.LocalAddress
//// * RemoteEndpoint.RemoteAddress
func ipSetVpnRulesEqual(first, second *types.NsxtIpSecVpnTunnel) bool {
	util.Logger.Println("comparing NSX-T IP Sev VPN configuration:")
	util.Logger.Printf("%+v\n", first)
	util.Logger.Println("against:")
	util.Logger.Printf("%+v\n", second)

	// These fields should be enough to cover uniqueness
	if first.Name == second.Name &&
		first.Description == second.Description &&
		first.Enabled == second.Enabled &&
		first.LocalEndpoint.LocalAddress == second.LocalEndpoint.LocalAddress &&
		first.RemoteEndpoint.RemoteAddress == second.RemoteEndpoint.RemoteAddress {
		return true
	}

	return false
}
