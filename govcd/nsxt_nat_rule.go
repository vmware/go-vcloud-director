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

// NsxtNatRule describes a single NAT rule of 4 different RuleTypes - DNAT`, `NO_DNAT`, `SNAT`, `NO_SNAT`.
//
// A SNAT or a DNAT rule on an Edge Gateway in the VMware Cloud Director environment is always configured from the
// perspective of your organization VDC.
// DNAT and NO_DNAT - outside traffic going inside
// SNAT and NO_SNAT - inside traffic going outside
// More docs in https://docs.vmware.com/en/VMware-Cloud-Director/10.2/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-9E43E3DC-C028-47B3-B7CA-59F0ED40E0A6.html
//
// Note. This structure and all its API calls will require at least API version 34.0, but will elevate it to 35.2 if
// possible because API 35.2 introduces support for 2 new fields FirewallMatch and Priority.
type NsxtNatRule struct {
	NsxtNatRule *types.NsxtNatRule
	client      *Client
	// edgeGatewayId is stored here so that pointer receiver functions can embed edge gateway ID into path
	edgeGatewayId string
}

// GetAllNatRules retrieves all NAT rules with an optional queryParameters filter.
func (egw *NsxtEdgeGateway) GetAllNatRules(queryParameters url.Values) ([]*NsxtNatRule, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	// Fields FirewallMatch and Priority require API version 35.2 to be set therefore version is elevated if API supports
	if client.APIVCDMaxVersionIs(">= 35.2") {
		apiVersion = "35.2"
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtNatRule{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtNatRule types with client
	wrappedResponses := make([]*NsxtNatRule, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtNatRule{
			NsxtNatRule:   typeResponses[sliceIndex],
			client:        client,
			edgeGatewayId: egw.EdgeGateway.ID,
		}
	}

	return wrappedResponses, nil
}

// GetNatRuleByName finds a NAT rule by Name and returns it
//
// Note. API does not enforce name uniqueness therefore an error will be thrown if two rules with the same name exist
func (egw *NsxtEdgeGateway) GetNatRuleByName(name string) (*NsxtNatRule, error) {
	// Ideally this function would use OpenAPI filters to perform server side filtering, but this endpoint does not
	// support any filters - even ID. Therefore one must retrieve all items and look if there is an item with the same ID
	allNatRules, err := egw.GetAllNatRules(nil)
	if err != nil {
		return nil, fmt.Errorf("error retriving all NSX-T NAT rules: %s", err)
	}

	var allResults []*NsxtNatRule

	for _, natRule := range allNatRules {
		if natRule.NsxtNatRule.Name == name {
			allResults = append(allResults, natRule)
		}
	}

	if len(allResults) > 1 {
		return nil, fmt.Errorf("error - found %d NSX-T NAT rules with name '%s'. Expected 1", len(allResults), name)
	}

	if len(allResults) == 0 {
		return nil, ErrorEntityNotFound
	}

	return allResults[0], nil
}

// GetNatRuleById finds a NAT rule by ID and returns it
func (egw *NsxtEdgeGateway) GetNatRuleById(id string) (*NsxtNatRule, error) {
	// Ideally this function would use OpenAPI filters to perform server side filtering, but this endpoint does not
	// support any filters - even ID. Therefore one must retrieve all items and look if there is an item with the same ID
	allNatRules, err := egw.GetAllNatRules(nil)
	if err != nil {
		return nil, fmt.Errorf("error retriving all NSX-T NAT rules: %s", err)
	}

	for _, natRule := range allNatRules {
		if natRule.NsxtNatRule.ID == id {
			return natRule, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// CreateNatRule creates a NAT rule and returns it.
//
// Note. API has a limitation, that it does not return ID for created rule. To work around it this function creates
// a NAT rule, fetches all rules and finds a rule with exactly the same field values and returns it (including ID)
// There is still a slight risk to retrieve wrong ID if exactly the same rule already exists.
func (egw *NsxtEdgeGateway) CreateNatRule(natRuleConfig *types.NsxtNatRule) (*NsxtNatRule, error) {
	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	// Fields FirewallMatch and Priority require API version 35.2 to be set therefore version is elevated if API supports
	if client.APIVCDMaxVersionIs(">= 35.2") {
		apiVersion = "35.2"
	}

	// Insert Edge Gateway ID into endpoint path edgeGateways/%s/nat/rules
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	// Creating NAT rule must follow different way than usual OpenAPI one because this item has an API bug and
	// NAT rule ID is not returned after this object is created. The only way to find its ID afterwards is to GET all
	// items, and manually match it based on rule name, etc.
	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, natRuleConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T NAT rule: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("task failed while creating NSX-T NAT rule: %s", err)
	}

	// queryParameters (API side filtering) are not used because pretty much nothing is accepted as filter (such fields as
	// name, description, ruleType and even ID are not allowed
	allNatRules, err := egw.GetAllNatRules(nil)

	for index, singleRule := range allNatRules {
		// Look for a matching rule
		if singleRule.IsEqualTo(natRuleConfig) {
			return allNatRules[index], nil

		}
	}
	return nil, fmt.Errorf("rule '%s' of type '%s' not found after creation", natRuleConfig.Name, natRuleConfig.RuleType)
}

// Update allows users to update NSX-T NAT rule
func (nsxtNat *NsxtNatRule) Update(natRuleConfig *types.NsxtNatRule) (*NsxtNatRule, error) {
	client := nsxtNat.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	// Fields FirewallMatch and Priority require API version 35.2 to be set therefore version is elevated if API supports
	if client.APIVCDMaxVersionIs(">= 35.2") {
		apiVersion = "35.2"
	}

	if nsxtNat.NsxtNatRule.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T NAT Rule without ID")
	}

	urlRef, err := nsxtNat.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, nsxtNat.edgeGatewayId), nsxtNat.NsxtNatRule.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtNatRule{
		NsxtNatRule:   &types.NsxtNatRule{},
		client:        client,
		edgeGatewayId: nsxtNat.edgeGatewayId,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, natRuleConfig, returnObject.NsxtNatRule)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T NAT Rule: %s", err)
	}

	return returnObject, nil
}

// Delete deletes NSX-T NAT rule
func (nsxtNat *NsxtNatRule) Delete() error {
	client := nsxtNat.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules
	apiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	// Fields FirewallMatch and Priority require API version 35.2 to be set therefore version is elevated if API supports
	if client.APIVCDMaxVersionIs(">= 35.2") {
		apiVersion = "35.2"
	}

	if nsxtNat.NsxtNatRule.ID == "" {
		return fmt.Errorf("cannot delete NSX-T NAT rule without ID")
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, nsxtNat.edgeGatewayId), nsxtNat.NsxtNatRule.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil)

	if err != nil {
		return fmt.Errorf("error deleting NSX-T NAT Rule: %s", err)
	}

	return nil
}

// IsEqualTo allows to check if a rule has exactly the same fields (except ID) to the supplied rule
func (nsxtNat *NsxtNatRule) IsEqualTo(rule *types.NsxtNatRule) bool {
	return natRulesEqual(nsxtNat.NsxtNatRule, rule)
}

// natRulesEqual is a helper to check if first and second supplied rules are exactly the same (except ID)
func natRulesEqual(first, second *types.NsxtNatRule) bool {
	util.Logger.Println("comparing NAT rule:")
	util.Logger.Printf("%+v\n", first)
	util.Logger.Println("against:")
	util.Logger.Printf("%+v\n", second)

	if first.Name == second.Name &&
		// Being an org user always returns logging as false - therefore cannot compare it.
		//first.Logging == second.Logging &&
		first.Enabled == second.Enabled &&
		first.Description == second.Description &&
		first.DnatExternalPort == second.DnatExternalPort &&
		first.SnatDestinationAddresses == second.SnatDestinationAddresses &&
		first.ExternalAddresses == second.ExternalAddresses &&
		first.InternalAddresses == second.InternalAddresses &&
		// Match either both application profiles being empty/nil, or both having the same value
		((first.ApplicationPortProfile == second.ApplicationPortProfile) ||
			(first.ApplicationPortProfile != nil && second.ApplicationPortProfile != nil && first.ApplicationPortProfile.ID == second.ApplicationPortProfile.ID)) {

		return true
	}

	return false
}