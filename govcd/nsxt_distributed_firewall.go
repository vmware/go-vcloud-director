package govcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// DistributedFirewall contains a types.DistributedFirewallRules which handles Distributed Firewall
// rules in a VDC Group
type DistributedFirewall struct {
	DistributedFirewallRuleContainer *types.DistributedFirewallRules
	client                           *Client
	VdcGroup                         *VdcGroup
}

type DistributedFirewallRule struct {
	Rule     *types.DistributedFirewallRule
	client   *Client
	VdcGroup *VdcGroup
}

// GetDistributedFirewall retrieves Distributed Firewall in a VDC Group which contains all rules
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) GetDistributedFirewall() (*DistributedFirewall, error) {
	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault))
	if err != nil {
		return nil, err
	}

	returnObject := &DistributedFirewall{
		DistributedFirewallRuleContainer: &types.DistributedFirewallRules{},
		client:                           client,
		VdcGroup:                         vdcGroup,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject.DistributedFirewallRuleContainer, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Distributed Firewall rules: %s", err)
	}

	return returnObject, nil
}

// UpdateDistributedFirewall updates Distributed Firewall in a VDC Group
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) UpdateDistributedFirewall(dfwRules *types.DistributedFirewallRules) (*DistributedFirewall, error) {
	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault))
	if err != nil {
		return nil, err
	}

	returnObject := &DistributedFirewall{
		DistributedFirewallRuleContainer: &types.DistributedFirewallRules{},
		client:                           client,
		VdcGroup:                         vdcGroup,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, dfwRules, returnObject.DistributedFirewallRuleContainer, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Distributed Firewall rules: %s", err)
	}

	return returnObject, nil
}

// DeleteAllDistributedFirewallRules removes all Distributed Firewall rules
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) DeleteAllDistributedFirewallRules() error {
	_, err := vdcGroup.UpdateDistributedFirewall(&types.DistributedFirewallRules{})
	return err
}

// DeleteAllRules removes all Distributed Firewall rules
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (firewall *DistributedFirewall) DeleteAllRules() error {
	if firewall.VdcGroup != nil && firewall.VdcGroup.VdcGroup != nil && firewall.VdcGroup.VdcGroup.Id == "" {
		return errors.New("empty VDC Group ID for parent VDC Group")
	}

	return firewall.VdcGroup.DeleteAllDistributedFirewallRules()
}

func (vdcGroup *VdcGroup) GetDistributedFirewallRuleById(id string) (*DistributedFirewallRule, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be specified")
	}

	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault), "/", id)
	if err != nil {
		return nil, err
	}

	returnObject := &DistributedFirewallRule{
		Rule:     &types.DistributedFirewallRule{},
		client:   client,
		VdcGroup: vdcGroup,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject.Rule, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Distributed Firewall rule: %s", err)
	}

	return returnObject, nil
}

func (vdcGroup *VdcGroup) GetDistributedFirewallRuleByName(name string) (*DistributedFirewallRule, error) {
	if name == "" {
		return nil, fmt.Errorf("name must be specified")
	}

	dfw, err := vdcGroup.GetDistributedFirewall()
	if err != nil {
		return nil, fmt.Errorf("error returning distributed firewall rules: %s", err)
	}

	var filteredByName []*types.DistributedFirewallRule
	for _, rule := range dfw.DistributedFirewallRuleContainer.Values {
		if rule.Name == name {
			filteredByName = append(filteredByName, rule)
		}
	}

	oneByName, err := oneOrError("name", name, filteredByName)
	if err != nil {
		return nil, err
	}

	return vdcGroup.GetDistributedFirewallRuleById(oneByName.ID)
}

// CreateDistributedFirewallRule is a convenience method that represents no real endpoint in the API
// While the API has only a mechanism to create all rules in one API go, there is a need in some
// cases to handle firewall rules one by one.
// It works by:
// 1. Getting all existing firewall rules
// 2. Checking that
//
// Note. Multiple instances of this function cannot be run in parallel as it will cause corruption.
// It is up to the consumer to handle locks
//
// 1. Retrieve all firewall rules in []json.RawMessage
// 2. Convert this result into DistributedFirewallRules.Values so that it can be searched based on
// fields (the order of slice remains the same as in []json.RawMessage, which is important)
func (vdcGroup *VdcGroup) CreateDistributedFirewallRule(optionalAboveRuleId string, rule *types.DistributedFirewallRule) (*DistributedFirewall, *DistributedFirewallRule, error) {
	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, nil, err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault))
	if err != nil {
		return nil, nil, err
	}

	// We're retrieving a []json.RawMessage so that the configuration is not altered at all
	// (even if it has fields, that are missing in types.DistributedFirewallRules{}, which can
	// happen as new versions of VCD are introduced)
	rawBodyStructure := &types.DistributedFirewallRulesRaw{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, rawBodyStructure, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving Distributed Firewall rules in raw format: %s", err)
	}

	// Make a string out of
	var rawJsonBodies []string
	for _, singleObject := range rawBodyStructure.Values {
		rawJsonBodies = append(rawJsonBodies, string(singleObject))
	}
	// rawJsonBodies contains a slice of all response objects and they must be formatted as a JSON slice (wrapped
	// into `[]`, separated with semicolons) so that unmarshalling to specified `outType` works in one go
	allResponses := `[` + strings.Join(rawJsonBodies, ",") + `]`

	// Convert the retrieved []json.RawMessage to *types.DistributedFirewallRules.Values so that IDs can be searched for
	// Note. The main goal here is to have 2 slices - one with []json.RawMessage and other
	// []*DistributedFirewallRule. One can look for IDs and capture firewall rule index
	dfwRules := &types.DistributedFirewallRules{}
	// Unmarshal all accumulated responses into `dfwRules`
	if err = json.Unmarshal([]byte(allResponses), &dfwRules.Values); err != nil {
		return nil, nil, fmt.Errorf("error decoding values into type types.DistributedFirewallRules{}: %s", err)
	}

	//////// convert given firewall rule config parameter to raw JSON message so that it can be
	//injected into original rules listed in []json.RawMessage at a selected position
	ruleByteSlice, err := json.Marshal(rule)
	if err != nil {
		return nil, nil, fmt.Errorf("error converting 'rule' to text: %s", err)
	}
	ruleString := string(ruleByteSlice)
	newRuleJsonMessage := json.RawMessage(ruleString)

	var dfwRulePayload []json.RawMessage

	var newRulePosition int

	// decide the position of new rule within a list of existing ones
	switch {
	case optionalAboveRuleId == "": // an optionalAboveRuleId - it means that the new firewall rule will go to the bottom of the list by default
		rawBodyStructure.Values = append(rawBodyStructure.Values, newRuleJsonMessage)
		dfwRulePayload = rawBodyStructure.Values
		newRulePosition = len(dfwRulePayload) - 1 // -1 to match for slice index

		// optionalAboveRuleId was given - need to search for a position of a rule with that ID so
		// that its index within []json.RawMessage can be found
	case optionalAboveRuleId != "":
		util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule 'optionalAboveRuleId=%s'. Searching within '%d' items", optionalAboveRuleId, len(dfwRules.Values))
		// Search above rule ID in the list of responses
		var aboveRuleSliceIndex *int
		for index := range dfwRules.Values {
			if dfwRules.Values[index].ID == optionalAboveRuleId {
				// using function `addrOf` to get copy of `*index` value as taking a direct address
				// of `&index` will shift before it is used in later code
				aboveRuleSliceIndex = addrOf(index)

				util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule found existing Firewall Rule with ID '%s' at position '%d'", optionalAboveRuleId, index)
				continue
			}
		}

		if aboveRuleSliceIndex == nil {
			return nil, nil, fmt.Errorf("specified above rule ID '%s' does not exist in current firewall rule list", optionalAboveRuleId)
		}

		newRulePosition = *aboveRuleSliceIndex // next position
		util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule Creating new structure with above rule ID '%s' in new position '%d'", optionalAboveRuleId, newRulePosition)

		// Create a new slice with additional capacity
		newSlice := make([]json.RawMessage, len(rawBodyStructure.Values)+1)

		util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule new container slice of size '%d' with previous element count '%d'", len(newSlice), len(rawBodyStructure.Values))

		// if newRulePosition is not 0 (at the top), then previous rules need to be copied
		// Copy the elements before the insertion point
		if newRulePosition != 0 {
			util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule Copying first '%d' slice [:%d]", newRulePosition, newRulePosition)
			copy(newSlice[:newRulePosition], rawBodyStructure.Values[:newRulePosition])
		}

		//

		// Insert the new element at the specified position
		util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule Inserting new element into position %d", newRulePosition)
		newSlice[newRulePosition] = newRuleJsonMessage

		// Copy the elements after the insertion point
		copy(newSlice[newRulePosition+1:], rawBodyStructure.Values[newRulePosition:])

		util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule Copying remaining items '%d'", newRulePosition)

		dfwRulePayload = newSlice
	}

	// Make the API call
	// for _, singleObject := range rawBodyStructure.Values {
	// 	rawJsonBodies = append(rawJsonBodies, string(singleObject))
	// }
	// rawJsonBodies contains a slice of all response objects and they must be formatted as a JSON slice (wrapped
	// into `[]`, separated with semicolons) so that unmarshalling to specified `outType` works in one go
	// allResponses = `[` + strings.Join(rawJsonBodies, ",") + `]`

	returnObject := &DistributedFirewall{
		DistributedFirewallRuleContainer: &types.DistributedFirewallRules{},
		client:                           client,
		VdcGroup:                         vdcGroup,
	}

	wrappedPayload := &types.DistributedFirewallRulesRaw{
		Values: dfwRulePayload,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, wrappedPayload, returnObject.DistributedFirewallRuleContainer, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error updating Distributed Firewall rules: %s", err)
	}

	// Single Rule
	returnObjectSingleRule := &DistributedFirewallRule{
		Rule:     returnObject.DistributedFirewallRuleContainer.Values[newRulePosition],
		client:   client,
		VdcGroup: vdcGroup,
	}
	// singleReturnedRule := returnObjectSingleRule.DistributedFirewallRuleContainer.Values[newRulePosition]

	return returnObject, returnObjectSingleRule, nil
}

func (dfwRule *DistributedFirewallRule) Update(rule *types.DistributedFirewallRule) (*DistributedFirewallRule, error) {
	if dfwRule.Rule.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T Distribute Firewall Rule without ID")
	}

	client := dfwRule.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dfwRule.VdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault), dfwRule.Rule.ID)
	if err != nil {
		return nil, err
	}

	returnObjectSingleRule := &DistributedFirewallRule{
		Rule: &types.DistributedFirewallRule{},
		// DistributedFirewallRuleContainer: &types.DistributedFirewallRules{},
		client:   client,
		VdcGroup: dfwRule.VdcGroup,
	}

	rule.ID = dfwRule.Rule.ID
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, rule, returnObjectSingleRule.Rule, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Distributed Firewall rules: %s", err)
	}

	return returnObjectSingleRule, nil
}

func (dfwRule *DistributedFirewallRule) Delete() error {
	if dfwRule.Rule.ID == "" {
		return fmt.Errorf("cannot delete NSX-T Distribute Firewall Rule without ID")
	}

	client := dfwRule.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, dfwRule.VdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault), "/", dfwRule.Rule.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting NSX-T Distribute Firewall Rule with ID '%s': %s", dfwRule.Rule.ID, err)
	}

	return nil
}
