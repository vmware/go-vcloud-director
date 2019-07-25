/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// CreateLbAppRule creates a load balancer application rule based on mandatory fields. It is a
// synchronous operation. It returns created object with all fields (including Id) populated or an error.
func (egw *EdgeGateway) CreateLbAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	if err := validateCreateLbAppRule(lbAppRuleConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}
	// We expect to get http.StatusCreated or if not an error of type types.NSXError
	resp, err := egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPost, types.AnyXMLMime,
		"error creating load balancer application rule: %s", lbAppRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	// Location header should look similar to:
	// [/network/edges/edge-3/loadbalancer/config/applicationrules/applicationRule-4]
	lbAppRuleId, err := extractNsxObjectIdFromPath(resp.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	readAppRule, err := egw.GetLbAppRule(&types.LbAppRule{Id: lbAppRuleId})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application rule with Id (%s) after creation: %s",
			readAppRule.Id, err)
	}
	return readAppRule, nil
}

// GetLbAppRule is able to find the types.LbAppRule type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) GetLbAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	if err := validateGetLbAppRule(lbAppRuleConfig, egw); err != nil {
		return nil, err
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Anonymous struct to unwrap response
	lbAppRuleResponse := &struct {
		LbAppRules []*types.LbAppRule `xml:"applicationRule"`
	}{}

	// This query returns all application rules as the API does not have filtering options
	_, err = egw.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime,
		"unable to read load balancer application rule: %s", nil, lbAppRuleResponse)
	if err != nil {
		return nil, err
	}

	// Search for application rule by Id or by Name
	for _, rule := range lbAppRuleResponse.LbAppRules {
		// If Id was specified for lookup - look for the same Id
		if lbAppRuleConfig.Id != "" && rule.Id == lbAppRuleConfig.Id {
			return rule, nil
		}

		// If Name was specified for lookup - look for the same Name
		if lbAppRuleConfig.Name != "" && rule.Name == lbAppRuleConfig.Name {
			// We found it by name. Let's verify if search Id was specified and it matches the lookup object
			if lbAppRuleConfig.Id != "" && rule.Id != lbAppRuleConfig.Id {
				return nil, fmt.Errorf("load balancer application rule was found by name (%s)"+
					", but its Id (%s) does not match specified Id (%s)",
					rule.Name, rule.Id, lbAppRuleConfig.Id)
			}
			return rule, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// ReadLBAppRuleById wraps GetLbAppRule and needs only an Id for lookup
func (egw *EdgeGateway) GetLbAppRuleById(id string) (*types.LbAppRule, error) {
	return egw.GetLbAppRule(&types.LbAppRule{Id: id})
}

// GetLbAppRuleByName wraps GetLbAppRule and needs only a Name for lookup
func (egw *EdgeGateway) GetLbAppRuleByName(name string) (*types.LbAppRule, error) {
	return egw.GetLbAppRule(&types.LbAppRule{Name: name})
}

// UpdateLbAppRule updates types.LbAppRule with all fields. At least name or Id must be specified.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) UpdateLbAppRule(lbAppRuleConfig *types.LbAppRule) (*types.LbAppRule, error) {
	err := validateUpdateLbAppRule(lbAppRuleConfig, egw)
	if err != nil {
		return nil, err
	}

	lbAppRuleConfig.Id, err = egw.getLbAppRuleIdByNameId(lbAppRuleConfig.Name, lbAppRuleConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot update load balancer application rule: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath + lbAppRuleConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	// Result should be 204, if not we expect an error of type types.NSXError
	_, err = egw.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		"error while updating load balancer application rule : %s", lbAppRuleConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	readAppRule, err := egw.GetLbAppRule(&types.LbAppRule{Id: lbAppRuleConfig.Id})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve application rule with Id (%s) after update: %s",
			readAppRule.Id, err)
	}
	return readAppRule, nil
}

// DeleteLbAppRule is able to delete the types.LbAppRule type by Name and/or Id.
// If both - Name and Id are specified it performs a lookup by Id and returns an error if the specified name and found
// name do not match.
func (egw *EdgeGateway) DeleteLbAppRule(lbAppRuleConfig *types.LbAppRule) error {
	err := validateDeleteLbAppRule(lbAppRuleConfig, egw)
	if err != nil {
		return err
	}

	lbAppRuleConfig.Id, err = egw.getLbAppRuleIdByNameId(lbAppRuleConfig.Name, lbAppRuleConfig.Id)
	if err != nil {
		return fmt.Errorf("cannot update load balancer application rule: %s", err)
	}

	httpPath, err := egw.buildProxiedEdgeEndpointURL(types.LbAppRulePath + lbAppRuleConfig.Id)
	if err != nil {
		return fmt.Errorf("could not get Edge Gateway API endpoint: %s", err)
	}

	return egw.client.ExecuteRequestWithoutResponse(httpPath, http.MethodDelete, "application/xml",
		"unable to delete application rule: %s", nil)
}

// DeleteLBAppRuleById wraps DeleteLbAppRule and requires only Id for deletion
func (egw *EdgeGateway) DeleteLbAppRuleById(id string) error {
	return egw.DeleteLbAppRule(&types.LbAppRule{Id: id})
}

// DeleteLbAppRuleByName wraps DeleteLbAppRule and requires only Name for deletion
func (egw *EdgeGateway) DeleteLbAppRuleByName(name string) error {
	return egw.DeleteLbAppRule(&types.LbAppRule{Name: name})
}

func validateCreateLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbAppRuleConfig.Name == "" {
		return fmt.Errorf("load balancer application rule Name cannot be empty")
	}

	return nil
}

func validateGetLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	if !egw.HasAdvancedNetworking() {
		return fmt.Errorf("only advanced edge gateways support load balancers")
	}

	if lbAppRuleConfig.Id == "" && lbAppRuleConfig.Name == "" {
		return fmt.Errorf("to read load balancer application rule at least one of `Id`, `Name`" +
			" fields must be specified")
	}

	return nil
}

func validateUpdateLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	// Update and create have the same requirements for now
	return validateCreateLbAppRule(lbAppRuleConfig, egw)
}

func validateDeleteLbAppRule(lbAppRuleConfig *types.LbAppRule, egw *EdgeGateway) error {
	// Read and delete have the same requirements for now
	return validateGetLbAppRule(lbAppRuleConfig, egw)
}

// getLbAppRuleIdByNameId checks if at least name or Id is set and returns the Id.
// If the Id is specified - it passes through the Id. If only name was specified
// it will lookup the object by name and return the Id.
func (egw *EdgeGateway) getLbAppRuleIdByNameId(name, id string) (string, error) {
	if name == "" && id == "" {
		return "", fmt.Errorf("at least Name or Id must be specific to find load balancer "+
			"application rule got name (%s) Id (%s)", name, id)
	}
	if id != "" {
		return id, nil
	}

	// if only name was specified, Id must be found, because only Id can be used in request path
	readlbAppRule, err := egw.GetLbAppRuleByName(name)
	if err != nil {
		return "", fmt.Errorf("unable to find load balancer application rule by name: %s", err)
	}
	return readlbAppRule.Id, nil
}
