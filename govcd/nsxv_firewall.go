/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// requestEdgeNatRules nests EdgeNatRule as a convenience for unmarshalling POST requests
type requestEdgeFirewallRules struct {
	XMLName           xml.Name                  `xml:"firewallRules"`
	EdgeFirewallRules []*types.EdgeFirewallRule `xml:"firewallRule"`
}

// responseEdgeNatRules is used to unwrap response when retrieving
type responseEdgeFirewallRules struct {
	XMLName           xml.Name                 `xml:"firewall"`
	Version           string                   `xml:"version"`
	EdgeFirewallRules requestEdgeFirewallRules `xml:"firewallRules"`
}

// CreateNsxvFirewall creates NAT rule using proxied NSX-V API. It is a synchronuous operation.
// It returns an object with all fields populated (including ID)
func (egw *EdgeGateway) CreateNsxvFirewall(natRuleConfig *types.EdgeFirewallRule) (*types.EdgeFirewallRule, error) {
	if err := validateCreateNsxvFirewall(natRuleConfig, egw); err != nil {
		return nil, err
	}

	// Wrap the provided rule for POST request
	natRuleRequest := requestEdgeFirewallRules{
		EdgeFirewallRules: []*types.EdgeFirewallRule{natRuleConfig},
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateFirewallPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating NAT rule: %s", natRuleRequest, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-1/nat/config/rules/197157]
	natRuleId, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readNatRule, err := egw.GetNsxvFirewallById(natRuleId)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve NAT rule with ID (%s) after creation: %s",
			natRuleId, err)
	}
	return readNatRule, nil
}

// UpdateNsxvFirewall updates types.EdgeFirewallRule with all fields using proxied NSX-V API. ID is
// mandatory to perform the update.
func (egw *EdgeGateway) UpdateNsxvFirewall(natRuleConfig *types.EdgeFirewallRule) (*types.EdgeFirewallRule, error) {
	err := validateUpdateNsxvFirewall(natRuleConfig, egw)
	if err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateFirewallPath + "/" + natRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating NAT rule : %s", natRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readNatRule, err := egw.GetNsxvFirewallById(natRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve NAT rule with ID (%s) after update: %s",
			readNatRule.ID, err)
	}
	return readNatRule, nil
}

// GetNsxvFirewallById retrieves types.EdgeFirewallRule by NAT rule ID as shown in the UI using proxied
// NSX-V API.
// It returns and error `ErrorEntityNotFound` if the NAT rule is now found.
func (egw *EdgeGateway) GetNsxvFirewallById(id string) (*types.EdgeFirewallRule, error) {
	if err := validateGetNsxvFirewall(id, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeFirewallPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	natRuleResponse := &responseEdgeFirewallRules{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read NAT rule: %s", nil, natRuleResponse)
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("[DEBUG] Searching for firewall rule with ID: %s", id)
	for _, rule := range natRuleResponse.EdgeFirewallRules.EdgeFirewallRules {
		util.Logger.Printf("[DEBUG] Checking rule: %#+v", rule)
		if rule.ID != "" && rule.ID == id {
			return rule, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// DeleteNsxvFirewallById deletes types.EdgeFirewallRule by NAT rule ID as shown in the UI using proxied
// NSX-V API.
// It returns and error `ErrorEntityNotFound` if the NAT rule is now found.
func (egw *EdgeGateway) DeleteNsxvFirewallById(id string) error {
	err := validateDeleteNsxvFirewall(id, egw)
	if err != nil {
		return err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateFirewallPath + "/" + id)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// check if the rule exists and pass back the error at it may be 'ErrorEntityNotFound'
	_, err = egw.GetNsxvFirewallById(id)
	if err != nil {
		return err
	}

	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete nat rule: %s", nil, &types.NSXError{})
	if err != nil {
		return err
	}

	return nil
}

func validateCreateNsxvFirewall(natRuleConfig *types.EdgeFirewallRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support NAT rules")
	}

	if natRuleConfig.Action == "" {
		return fmt.Errorf("NAT rule must have an action")
	}

	return nil
}

func validateUpdateNsxvFirewall(natRuleConfig *types.EdgeFirewallRule, egw *EdgeGateway) error {
	if natRuleConfig.ID == "" {
		return fmt.Errorf("NAT rule must ID must be set for update")
	}

	return validateCreateNsxvFirewall(natRuleConfig, egw)
}

func validateGetNsxvFirewall(id string, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support NAT rules")
	}

	if id == "" {
		return fmt.Errorf("unable to retrieve NAT rule without ID")
	}

	return nil
}

func validateDeleteNsxvFirewall(id string, egw *EdgeGateway) error {
	return validateGetNsxvFirewall(id, egw)
}
