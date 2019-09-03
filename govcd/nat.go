/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// requestEdgeNatRules nests EdgeNatRule as a convenience for unmarshalling POST requests
type requestEdgeNatRules struct {
	XMLName       xml.Name              `xml:"natRules"`
	EdgeNatRules []*types.EdgeNatRule `xml:"natRule"`
}

// responseEdgeNatRules is used to unwrap response when retrieving
type responseEdgeNatRules struct {
	XMLName  xml.Name             `xml:"nat"`
	Version  string               `xml:"version"`
	NatRules requestEdgeNatRules `xml:"natRules"`
}

// CreateNatRule
func (egw *EdgeGateway) CreateNsxvNatRule(natRuleConfig *types.EdgeNatRule) (*types.EdgeNatRule, error) {
	if err := validateCreateNsxvNatRule(natRuleConfig, egw); err != nil {
		return nil, err
	}

	// Wrap the provided rule for POST request
	natRuleRequest := requestEdgeNatRules{
		EdgeNatRules: []*types.EdgeNatRule{natRuleConfig},
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateNatPath)
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

	readNatRule, err := egw.GetNsxvNatRule(&types.EdgeNatRule{ID: natRuleId})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve NAT rule with ID (%s) after creation: %s",
			natRuleId, err)
	}
	return readNatRule, nil
}

func (egw *EdgeGateway) UpdateNsxvNatRule(natRuleConfig *types.EdgeNatRule) (*types.EdgeNatRule, error) {
	err := validateUpdateNsxvNatRule(natRuleConfig, egw)
	if err != nil {
		return nil, err
	}
	
	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateNatPath + "/" + natRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating NAT rule : %s", natRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readNatRule, err := egw.GetNsxvNatRuleById(natRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve NAT rule with ID (%s) after update: %s",
		readNatRule.ID, err)
	}
	return readNatRule, nil
}

func (egw *EdgeGateway) GetNsxvNatRule(natRuleConfig *types.EdgeNatRule) (*types.EdgeNatRule, error) {
	if err := validateGetNsxvNatRule(natRuleConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeNatPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	natRuleResponse := &responseEdgeNatRules{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read NAT rule: %s", nil, natRuleResponse)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("whole struct %+#v\n", natRuleResponse)

	// Search for nat rule by ID or by Name
	// for _, rule := range natRuleResponse.NatRules.EdgeNatRules  {
	for _, rule := range natRuleResponse.NatRules.EdgeNatRules {
		// If ID was specified for lookup - look for the same ID
		// fmt.Printf("checking %+#v\n", rule)
		if rule.ID != "" && rule.ID == natRuleConfig.ID {
			return rule, nil
		}
	}

	// 	// If Name was specified for lookup - look for the same Name
	// 	if lbAppRuleConfig.Name != "" && rule.Name == lbAppRuleConfig.Name {
	// 		// We found it by name. Let's verify if search ID was specified and it matches the lookup object
	// 		if lbAppRuleConfig.ID != "" && rule.ID != lbAppRuleConfig.ID {
	// 			return nil, fmt.Errorf("load balancer application rule was found by name (%s)"+
	// 				", but its ID (%s) does not match specified ID (%s)",
	// 				rule.Name, rule.ID, lbAppRuleConfig.ID)
	// 		}
	// 		return rule, nil
	// 	}
	// }

	return nil, ErrorEntityNotFound
}

func (egw *EdgeGateway) GetNsxvNatRuleById(id string) (*types.EdgeNatRule, error) {
	return egw.GetNsxvNatRule(&types.EdgeNatRule{ID: id})
}

func (egw *EdgeGateway) DeleteNsxvNatRule(natRuleConfig *types.EdgeNatRule) error {
	err := validateDeleteNsxvNatRule(natRuleConfig, egw)
	if err != nil {
		return err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateNatPath + "/" + natRuleConfig.ID)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		"unable to delete nat rule: %s", nil, &types.NSXError{})
	if err != nil {
		return err
	}

	return nil
}

func (egw *EdgeGateway) DeleteNsxvNatRuleById(id string) error {
	return egw.DeleteNsxvNatRule(&types.EdgeNatRule{ID: id})
}

func validateCreateNsxvNatRule(natRuleConfig *types.EdgeNatRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support NAT rules")
	}

	if natRuleConfig.Action == "" {
		return fmt.Errorf("NAT rule must have an action")
	}

	if natRuleConfig.TranslatedAddress == "" {
		return fmt.Errorf("NAT rule must translated address specified")
	}

	return nil
}

func validateUpdateNsxvNatRule(natRuleConfig *types.EdgeNatRule, egw *EdgeGateway) error {
	if natRuleConfig.ID == "" {
		return fmt.Errorf("NAT rule must ID must be set for update")
	}

	return validateCreateNsxvNatRule(natRuleConfig, egw)
}

func validateGetNsxvNatRule(natRuleConfig *types.EdgeNatRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support NAT rules")
	}

	if natRuleConfig.ID == "" {
		return fmt.Errorf("unable to retrieve NAT rule without ID")
	}

	return nil
}

func validateDeleteNsxvNatRule(natRuleConfig *types.EdgeNatRule, egw *EdgeGateway) error {
	return validateGetNsxvNatRule(natRuleConfig, egw)
}
