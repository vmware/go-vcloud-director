/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtFirewall contains a types.NsxtFirewallRuleContainer which encloses three types of rules -
// system, default and user defined rules. User defined rules are the only ones that can be modified, others are
// read-only.
type NsxtFirewall struct {
	NsxtFirewallRuleContainer *types.NsxtFirewallRuleContainer
	client                    *Client
	// edgeGatewayId is stored for usage in NsxtFirewall receiver functions
	edgeGatewayId string
}

// UpdateNsxtFirewall allows user to set new firewall rules or update existing ones. The API does not have POST endpoint
// and always uses PUT endpoint for creating and updating.
func (egw *NsxtEdgeGateway) UpdateNsxtFirewall(firewallRules *types.NsxtFirewallRuleContainer) (*NsxtFirewall, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtFirewallRules
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path edgeGateways/%s/firewall/rules
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtFirewall{
		NsxtFirewallRuleContainer: &types.NsxtFirewallRuleContainer{},
		client:                    client,
		edgeGatewayId:             egw.EdgeGateway.ID,
	}

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, firewallRules, returnObject.NsxtFirewallRuleContainer, nil)
	if err != nil {
		return nil, fmt.Errorf("error setting NSX-T Firewall: %s", err)
	}

	return returnObject, nil
}

// GetNsxtFirewall retrieves all firewall rules system, default and user defined rules
func (egw *NsxtEdgeGateway) GetNsxtFirewall() (*NsxtFirewall, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtFirewallRules
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	// Insert Edge Gateway ID into endpoint path edgeGateways/%s/firewall/rules
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtFirewall{
		NsxtFirewallRuleContainer: &types.NsxtFirewallRuleContainer{},
		client:                    client,
		edgeGatewayId:             egw.EdgeGateway.ID,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, returnObject.NsxtFirewallRuleContainer, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving NSX-T Firewall rules: %s", err)
	}

	// Store Edge Gateway ID for later operations
	returnObject.edgeGatewayId = egw.EdgeGateway.ID

	return returnObject, nil
}

// DeleteAllRules allows users to delete all NSX-T Firewall rules in a particular Edge Gateway
func (firewall *NsxtFirewall) DeleteAllRules() error {

	if firewall.edgeGatewayId == "" {
		return fmt.Errorf("missing Edge Gateway ID")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtFirewallRules
	minimumApiVersion, err := firewall.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := firewall.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, firewall.edgeGatewayId))
	if err != nil {
		return err
	}

	err = firewall.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting all NSX-T Firewall Rules: %s", err)
	}

	return nil
}

// DeleteRuleById allows users to delete NSX-T Firewall Rule By ID
func (firewall *NsxtFirewall) DeleteRuleById(id string) error {
	if id == "" {
		return fmt.Errorf("empty ID specified")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtFirewallRules
	minimumApiVersion, err := firewall.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := firewall.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, firewall.edgeGatewayId), "/", id)
	if err != nil {
		return err
	}

	err = firewall.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting NSX-T Firewall Rule with ID '%s': %s", id, err)
	}

	return nil
}
