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

// wrappedEdgeSnatRules is used to unwrap response when retrieving
type wrappedEdgeSnatRules struct {
	XMLName  xml.Name `xml:"nat"`
	Version  string   `xml:"version"`
	NatR types.EdgeSnatRules   `xml:"natRules"`
}

func (egw *EdgeGateway) CreateSnatRule(snatRuleConfig *types.EdgeSnatRule) (*types.EdgeSnatRule, error) {
	if err := validateCreateSnatRule(snatRuleConfig, egw); err != nil {
		return nil, err
	}

	natRuleRequest := &types.EdgeSnatRules{}

	natRules := []*types.EdgeSnatRule{}
	natRules = append(natRules, snatRuleConfig)

	natRuleRequest.EdgeSnatRules = natRules

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
	// [/network/edges/edge-3/loadbalancer/config/applicationrules/applicationRule-4]
	lbAppRuleId, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readAppRule, err := egw.GetSnatRule(&types.EdgeSnatRule{ID: lbAppRuleId})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application rule with ID (%s) after creation: %s",
			lbAppRuleId, err)
	}
	return readAppRule, nil
}

func (egw *EdgeGateway) UpdateSnatRule(SnatRuleConfig *types.EdgeSnatRule) (*types.EdgeSnatRule, error) {
	return nil, nil
}

func (egw *EdgeGateway) GetSnatRule(snatRuleConfig *types.EdgeSnatRule) (*types.EdgeSnatRule, error) {
	if err := validateGetSnatRule(snatRuleConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.EdgeNatPath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	natRuleResponse := &wrappedEdgeSnatRules{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read SNAT rule: %s", nil, natRuleResponse)
	if err != nil {
		return nil, err
	}


	fmt.Printf("whole struct %+#v\n", natRuleResponse)

	// Search for nat rule by ID or by Name
	// for _, rule := range natRuleResponse.NatRules.EdgeSnatRules  {
	for _, rule := range natRuleResponse.NatR.EdgeSnatRules  {
		// If ID was specified for lookup - look for the same ID
		fmt.Printf("checking %+#v\n", rule)
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

func (egw *EdgeGateway) DeleteSnatRule(snatRuleConfig *types.EdgeSnatRule) error {
	return nil
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

func validateGetSnatRule(snatRuleConfig *types.EdgeSnatRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support SNAT rules")
	}

	// if snatRuleConfig.RuleType == "" {
	// 	return fmt.Errorf("SnatRule")
	// }

	return nil
}
