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

	EdgeGatewayId string
}

// GetNsxtEdgeGatewayDns retrieves the DNS configuration for the underlying edge gateway
func (egw *NsxtEdgeGateway) GetNsxtEdgeGatewayDns() (*NsxtEdgeGatewayDns, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDns
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	dnsConfig := &NsxtEdgeGatewayDns{
		client:        client,
		EdgeGatewayId: egw.EdgeGateway.ID,
	}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &dnsConfig.NsxtEdgeGatewayDns, nil)
	if err != nil {
		return nil, err
	}

	return dnsConfig, nil
}

// Update updates the DNS configuration for the Edge Gateway
func (dns *NsxtEdgeGatewayDns) Update(updatedConfig *types.NsxtEdgeGatewayDns) (*NsxtEdgeGatewayDns, error) {
	client := dns.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayDns
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dns.EdgeGatewayId))
	if err != nil {
		return nil, err
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, updatedConfig, &dns.NsxtEdgeGatewayDns, nil)
	if err != nil {
		return nil, err
	}

	return dns, nil
}

// Refresh refreshes the DNS configuration for the Edge Gateway
func (dns *NsxtEdgeGatewayDns) Refresh() error {
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

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &dns.NsxtEdgeGatewayDns, nil)
	if err != nil {
		return err
	}

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
