/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetDhcpPoolsAndBindings retrieves a structure of *types.EdgeDhcp with all DHCP pool and binding settings present on a
// particular edge gateway.
func (egw *EdgeGateway) GetDhcpPoolsAndBindings() (*types.EdgeDhcp, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateways support DHCP pools")
	}
	response := &types.EdgeDhcp{}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// This query Edge gateway DHCP pool using proxied NSX-V API
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read edge gateway DHCP pool configuration: %s", nil, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// UpdateDhcpPoolsAndBindings allows to update DHCP pool and binding settings for a particular edge gateway
//
// Note. Update must contain all settings to avoid removing previously set settings
func (egw *EdgeGateway) UpdateDhcpPoolsAndBindings(dhcpPoolConfig *types.EdgeDhcp) (*types.EdgeDhcp, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateways support DHCP pools")
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusNoContent or if not an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error setting DHCP pool settings: %s", dhcpPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	return egw.GetDhcpPoolsAndBindings()
}

// UpdateDhcpPools allows to update DHCP pool settings for a particular edge gateway
//
// Note. There is no endpoint to update DHCP pools only as static DHCP bindings must go together always. This function
// retrieves latest static DHCP bindings and feeds them back into dhcpPoolConfig to avoid overwriting/removing existing
// static DHCP bindings
func (egw *EdgeGateway) UpdateDhcpPools(dhcpPoolConfig *types.EdgeDhcp) (*types.EdgeDhcp, error) {
	if !egw.HasAdvancedNetworking() {
		return nil, fmt.Errorf("only advanced edge gateways support DHCP pools")
	}

	// To allow updating DHCP pools only
	currentDhcpSettings, err := egw.GetDhcpPoolsAndBindings()
	if err != nil {
		return nil, fmt.Errorf("error reading existing DHCP pools and bindings: %s", err)
	}
	dhcpPoolConfig.StaticBindings = currentDhcpSettings.StaticBindings

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusNoContent or if not an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error setting DHCP pool settings: %s", dhcpPoolConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	return egw.GetDhcpPoolsAndBindings()
}

// ResetDhcpPools internally performs an update with empty DHCP pool definition while persisting DHCP bindings
//
// Note. It will disable DHCP service as well because that is how native ResetDhcpPoolsAndBindings() works
func (egw *EdgeGateway) ResetDhcpPools() error {
	dhcpConfig, err := egw.GetDhcpPoolsAndBindings()
	if err != nil {
		return fmt.Errorf("error getting DHCP pools: %s", err)
	}

	dhcpPoolConfig := &types.EdgeDhcp{
		Enabled: false,
		// Feed in the same static bindings to not override existing customer settings
		StaticBindings: dhcpConfig.StaticBindings,
	}

	_, err = egw.UpdateDhcpPoolsAndBindings(dhcpPoolConfig)
	return err
}

// ResetDhcpPoolsAndBindings removes all configuration for DHCP Pools and Bindings by sending a DELETE request for DHCP
// pool and binding configuration endpoint
func (egw *EdgeGateway) ResetDhcpPoolsAndBindings() error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support DHCP pools")
	}
	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeDhcpPath)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Send a DELETE request to DHCP pool configuration endpoint
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to reset edge gateway DHCP pool configuration: %s", nil, &types.NSXError{})
	return err
}
