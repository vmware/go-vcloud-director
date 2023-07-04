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

// GetDistributedFirewallRuleById retrieves single Distributed Firewall Rule by ID
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

// GetDistributedFirewallRuleByName retrieves single firewall rule by name
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

// CreateDistributedFirewallRule is a wrapper around "vdcGroups/%s/dfwPolicies/%s/rules" endpoint
// which handles all distributed firewall (dfw) rules  at once. While there is no real endpoint to
// create single firewall rule, it is a requirements for some cases (e.g. using in Terraform)
// The code works by doing the following steps:
//
// 1. Getting all Distributed Firewall Rules and storing them in private intermediate
// type`distributedFirewallRulesRaw` which holds a []json.RawMessage (text) instead of exact types.
// This will prevent altering existing rules in any way (for example if a new field appears in
// schema in future VCD versions)
//

// 2. Converting the give `rule` into json.RawMessage so that it is provided in the same format as other already retrieved rules
//
// 3. Creating a new structure of []json.RawMessage which puts the new rule into one of places:
// 3.1. to the end of []json.RawMessage - bottom of the list
// 3.2. if `optionalAboveRuleId` argument is specified - identifying the position and placing new
// rule above it
// 2. Converting these []json.RawMessage into a string and Unmarshalling it into exact type
// `types.DistributedFirewallRules` that will allow checking ID field values (it is important to
// note that the order and quantity of elements in both slices remains the same). It will be used
// for finding and matching `optionalAboveRuleId`
//
// 5. Perform a PUT (update) call to the "vdcGroups/%s/dfwPolicies/%s/rules" endpoint
//
// Note. Running this function concurently will probably corrupt firewall rules as it uses an
// endpoint that manage all rules ("vdcGroups/%s/dfwPolicies/%s/rules")
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

	// 1.
	rawJsonExistingFirewallRules := &distributedFirewallRulesRaw{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, rawJsonExistingFirewallRules, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving Distributed Firewall rules in raw format: %s", err)
	}

	// 2.
	newRuleRawJson, err := firewallRuleToRawJson(rule)
	if err != nil {
		return nil, nil, err
	}

	// dfwRuleUpdatePayload will contain complete request for Distributed Firewall Rule Update
	// operation. Its content will be decided based on whether 'optionalAboveRuleId' parameter was
	// specified or not.
	var dfwRuleUpdatePayload []json.RawMessage
	// newRuleSlicePosition will contain slice index to where new firewall rule will be put
	var newRuleSlicePosition int

	switch {
	// 3.1
	case optionalAboveRuleId == "": // an optionalAboveRuleId is empty - it means that the new firewall rule will be appended to the bottom of the list
		rawJsonExistingFirewallRules.Values = append(rawJsonExistingFirewallRules.Values, newRuleRawJson)
		dfwRuleUpdatePayload = rawJsonExistingFirewallRules.Values
		newRuleSlicePosition = len(dfwRuleUpdatePayload) - 1 // -1 to match for slice index

		// optionalAboveRuleId was given - need to search for a position of a rule with that ID so
		// that its index within []json.RawMessage can be found
		// 3.2
	case optionalAboveRuleId != "": // an optionalAboveRuleId is specified - new rule has to be placed above the specified rule
		// 3.2.1 Convert '[]json.Rawmessage' to 'types.DistributedFirewallRules'
		dfwRules, err := convertRawJsonToFirewallRules(rawJsonExistingFirewallRules)
		if err != nil {
			return nil, nil, err
		}
		// 3.2.2 Find index for specified 'optionalAboveRuleId' rule
		newFwRuleSliceIndex, err := getFirewallRuleIndexById(dfwRules, optionalAboveRuleId)
		if err != nil {
			return nil, nil, err
		}

		// 3.2.3 Compose new update (PUT) payload with all firewall rules and inject
		// 'newRuleRawJson' into position 'newFwRuleSliceIndex' and shift other rules to the bottom
		dfwRuleUpdatePayload, err = composeUpdatePayloadWithNewRulePosition(newFwRuleSliceIndex, rawJsonExistingFirewallRules, newRuleRawJson)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating update payload with optionalAboveRuleId '%s' :%s", optionalAboveRuleId, err)
		}
	}
	// Request payload with new configuration using 'dfwRuleUpdatePayload'
	updateRequestPayload := &distributedFirewallRulesRaw{
		Values: dfwRuleUpdatePayload,
	}

	returnAllFirewallRules := &DistributedFirewall{
		DistributedFirewallRuleContainer: &types.DistributedFirewallRules{},
		client:                           client,
		VdcGroup:                         vdcGroup,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, updateRequestPayload, returnAllFirewallRules.DistributedFirewallRuleContainer, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error updating Distributed Firewall rules: %s", err)
	}

	// Create an entity for single firewall rule (which can be updated and deleted using their own endpoints)
	returnObjectSingleRule := &DistributedFirewallRule{
		Rule:     returnAllFirewallRules.DistributedFirewallRuleContainer.Values[newRuleSlicePosition],
		client:   client,
		VdcGroup: vdcGroup,
	}

	return returnAllFirewallRules, returnObjectSingleRule, nil
}

// Update a single Distributed Firewall Rule
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
		Rule:     &types.DistributedFirewallRule{},
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

// Delete a single Distributed Firewall Rule
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

// getFirewallRuleIndexById searches for 'firewallRuleId' going through a list of available firewall
// rules and returns its index or error if the firewall rule is not found
func getFirewallRuleIndexById(dfwRules *types.DistributedFirewallRules, firewallRuleId string) (int, error) {
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule 'optionalAboveRuleId=%s'. Searching within '%d' items",
		firewallRuleId, len(dfwRules.Values))
	var aboveRuleSliceIndex *int
	for index := range dfwRules.Values {
		if dfwRules.Values[index].ID == firewallRuleId {
			// using function `addrOf` to get copy of `index` value as taking a direct address
			// of `&index` will shift before it is used in later code due to how Go range works
			aboveRuleSliceIndex = addrOf(index)
			util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule found existing Firewall Rule with ID '%s' at position '%d'",
				firewallRuleId, index)
			continue
		}
	}

	if aboveRuleSliceIndex == nil {
		return 0, fmt.Errorf("specified above rule ID '%s' does not exist in current Distributed Firewall Rule list", firewallRuleId)
	}

	return *aboveRuleSliceIndex, nil
}

// firewallRuleToRawJson Marshal a single `types.DistributedFirewallRule` into `json.RawMessage`
// representation
func firewallRuleToRawJson(rule *types.DistributedFirewallRule) (json.RawMessage, error) {
	ruleByteSlice, err := json.Marshal(rule)
	if err != nil {
		return nil, fmt.Errorf("error marshalling 'rule': %s", err)
	}
	ruleJsonMessage := json.RawMessage(string(ruleByteSlice))
	return ruleJsonMessage, nil
}

// convertRawJsonToFirewallRules converts []json.RawMessage to
// types.DistributedFirewallRules.Values so that entries can be filtered by ID or other fields.
// Note. Slice order remains the same
func convertRawJsonToFirewallRules(rawBodyStructure *distributedFirewallRulesRaw) (*types.DistributedFirewallRules, error) {
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
	if err := json.Unmarshal([]byte(allResponses), &dfwRules.Values); err != nil {
		return nil, fmt.Errorf("error decoding values into type types.DistributedFirewallRules: %s", err)
	}

	return dfwRules, nil
}

// composeUpdatePayloadWithNewRulePosition takes a slice of existing firewall rules and injects new
// firewall rule at a given position `newRuleSlicePosition`
func composeUpdatePayloadWithNewRulePosition(newRuleSlicePosition int, rawBodyStructure *distributedFirewallRulesRaw, newRuleJsonMessage json.RawMessage) ([]json.RawMessage, error) {
	// Create a new slice with 1 additional capacity to add new firewall rule
	newSlice := make([]json.RawMessage, len(rawBodyStructure.Values)+1)
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule new container slice of size '%d' with previous element count '%d'", len(newSlice), len(rawBodyStructure.Values))
	// if newRulePosition is not 0 (at the top), then previous rules need to be copied to the beginning of new slice
	if newRuleSlicePosition != 0 {
		util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule copying first '%d' slice [:%d]", newRuleSlicePosition, newRuleSlicePosition)
		copy(newSlice[:newRuleSlicePosition], rawBodyStructure.Values[:newRuleSlicePosition])
	}

	// Insert the new element at the specified position
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule inserting new element into position %d", newRuleSlicePosition)
	newSlice[newRuleSlicePosition] = newRuleJsonMessage

	// Copy the remaining elements after new rule
	copy(newSlice[newRuleSlicePosition+1:], rawBodyStructure.Values[newRuleSlicePosition:])
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule copying remaining items '%d'", newRuleSlicePosition)

	return newSlice, nil
}

// distributedFirewallRulesRaw is a copy of `types.DistributedFirewallRules` so that values can be
// unmarshalled into json.RawMessage (as strings) instead of exact types `DistributedFirewallRule`
// It has Public field Values so that marshalling can work, but is not exported itself as it is only
// an intermediate type
type distributedFirewallRulesRaw struct {
	Values []json.RawMessage `json:"values"`
}
