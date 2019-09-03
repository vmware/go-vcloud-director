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

// requestEdgeSnatRules nests EdgeSnatRule as a convenience for unmarshalling POST requests
type requestEdgeSnatRules struct {
	XMLName       xml.Name              `xml:"natRules"`
	EdgeSnatRules []*types.EdgeSnatRule `xml:"natRule"`
}

// responseEdgeSnatRules is used to unwrap response when retrieving
type responseEdgeSnatRules struct {
	XMLName  xml.Name             `xml:"nat"`
	Version  string               `xml:"version"`
	NatRules requestEdgeSnatRules `xml:"natRules"`
}

// CreateSnatRule
func (egw *EdgeGateway) CreateSnatRule(snatRuleConfig *types.EdgeSnatRule) (*types.EdgeSnatRule, error) {
	if err := validateCreateSnatRule(snatRuleConfig, egw); err != nil {
		return nil, err
	}

	// Wrap the provided rule for POST request
	natRuleRequest := requestEdgeSnatRules{
		EdgeSnatRules: []*types.EdgeSnatRule{snatRuleConfig},
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateNatPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating SNAT rule: %s", natRuleRequest, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-1/nat/config/rules/197157]
	snatRuleId, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readSnatRule, err := egw.GetSnatRule(&types.EdgeSnatRule{ID: snatRuleId})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve SNAT rule with ID (%s) after creation: %s",
			snatRuleId, err)
	}
	return readSnatRule, nil
}

func (egw *EdgeGateway) UpdateSnatRule(snatRuleConfig *types.EdgeSnatRule) (*types.EdgeSnatRule, error) {
	err := validateUpdateSnatRule(snatRuleConfig, egw)
	if err != nil {
		return nil, err
	}
	
	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateNatPath + "/" + snatRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating NAT rule : %s", snatRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readSnatRule, err := egw.GetSnatRuleById(snatRuleConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve NAT rule with ID (%s) after update: %s",
		readSnatRule.ID, err)
	}
	return readSnatRule, nil
}

func (egw *EdgeGateway) GetSnatRule(snatRuleConfig *types.EdgeSnatRule) (*types.EdgeSnatRule, error) {
	if err := validateGetSnatRule(snatRuleConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeNatPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	natRuleResponse := &responseEdgeSnatRules{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read SNAT rule: %s", nil, natRuleResponse)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("whole struct %+#v\n", natRuleResponse)

	// Search for nat rule by ID or by Name
	// for _, rule := range natRuleResponse.NatRules.EdgeSnatRules  {
	for _, rule := range natRuleResponse.NatRules.EdgeSnatRules {
		// If ID was specified for lookup - look for the same ID
		// fmt.Printf("checking %+#v\n", rule)
		if rule.ID != "" && rule.ID == snatRuleConfig.ID {
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

func (egw *EdgeGateway) GetSnatRuleById(id string) (*types.EdgeSnatRule, error) {
	return egw.GetSnatRule(&types.EdgeSnatRule{ID: id})
}

func (egw *EdgeGateway) DeleteSnatRule(snatRuleConfig *types.EdgeSnatRule) error {
	err := validateDeleteSnatRule(snatRuleConfig, egw)
	if err != nil {
		return err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeCreateNatPath + "/" + snatRuleConfig.ID)
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

func (egw *EdgeGateway) DeleteSnatRuleById(id string) error {
	return egw.DeleteSnatRule(&types.EdgeSnatRule{ID: id})
}

func validateCreateSnatRule(snatRuleConfig *types.EdgeSnatRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support SNAT rules")
	}

	if snatRuleConfig.Action == "" {
		return fmt.Errorf("NAT rule must have an action")
	}

	if snatRuleConfig.TranslatedAddress == "" {
		return fmt.Errorf("NAT rule must translated address specified")
	}

	return nil
}

func validateUpdateSnatRule(snatRuleConfig *types.EdgeSnatRule, egw *EdgeGateway) error {
	if snatRuleConfig.ID == "" {
		return fmt.Errorf("NAT rule must ID must be set for update")
	}

	return validateCreateSnatRule(snatRuleConfig, egw)
}

func validateGetSnatRule(snatRuleConfig *types.EdgeSnatRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support SNAT rules")
	}

	if snatRuleConfig.ID == "" {
		return fmt.Errorf("unable to retrieve SNAT rule without ID")
	}

	return nil
}

func validateDeleteSnatRule(snatRuleConfig *types.EdgeSnatRule, egw *EdgeGateway) error {
	return validateGetSnatRule(snatRuleConfig, egw)
}
