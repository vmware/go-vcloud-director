/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtEdgeGatewayDns can be used to configure DNS on NSX-T Edge Gateway.
type NsxtEdgeGatewayDns struct {
	NsxtEdgeGatewayDns *types.NsxtEdgeGatewayDns
	client             *Client
	EdgeGatewayId      string
}

// GetDnsConfig retrieves the DNS configuration for the underlying edge gateway
func (egw *NsxtEdgeGateway) GetDnsConfig() (*NsxtEdgeGatewayDns, error) {
	return getDnsConfig(egw.client, egw.EdgeGateway.ID)
}

// UpdateDnsConfig updates the DNS configuration for the Edge Gateway
func (egw *NsxtEdgeGateway) UpdateDnsConfig(updatedConfig *types.NsxtEdgeGatewayDns) (*NsxtEdgeGatewayDns, error) {
	return updateDnsConfig(updatedConfig, egw.client, egw.EdgeGateway.ID)
}

// Update updates the DNS configuration for the underlying Edge Gateway
func (dns *NsxtEdgeGatewayDns) Update(updatedConfig *types.NsxtEdgeGatewayDns) (*NsxtEdgeGatewayDns, error) {
	return updateDnsConfig(updatedConfig, dns.client, dns.EdgeGatewayId)
}

// Refresh refreshes the DNS configuration for the Edge Gateway
func (dns *NsxtEdgeGatewayDns) Refresh() error {
	updatedDns, err := getDnsConfig(dns.client, dns.EdgeGatewayId)
	if err != nil {
		return err
	}
	dns.NsxtEdgeGatewayDns = updatedDns.NsxtEdgeGatewayDns

	return nil
}

// Delete deletes the DNS configuration for the Edge Gateway
func (dns *NsxtEdgeGatewayDns) Delete() error {
	client := dns.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDns
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dns.EdgeGatewayId))
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func getDnsConfig(client *Client, edgeGatewayId string) (*NsxtEdgeGatewayDns, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDns
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, edgeGatewayId))
	if err != nil {
		return nil, err
	}

	dnsConfig := &NsxtEdgeGatewayDns{
		client:        client,
		EdgeGatewayId: edgeGatewayId,
	}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &dnsConfig.NsxtEdgeGatewayDns, nil)
	if err != nil {
		return nil, err
	}

	return dnsConfig, nil

}

func updateDnsConfig(updatedConfig *types.NsxtEdgeGatewayDns, client *Client, edgeGatewayId string) (*NsxtEdgeGatewayDns, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDns
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, edgeGatewayId))
	if err != nil {
		return nil, err
	}

	dns := &NsxtEdgeGatewayDns{
		client:        client,
		EdgeGatewayId: edgeGatewayId,
	}
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, updatedConfig, &dns.NsxtEdgeGatewayDns, nil)
	if err != nil {
		return nil, err
	}

	return dns, nil
}
