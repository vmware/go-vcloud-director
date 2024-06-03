package govcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

const (
	labelDistributedFirewall     = "NSX-T Distributed Firewall"
	labelDistributedFirewallRule = "NSX-T Distributed Firewall Rule"
)

// DistributedFirewall contains a types.DistributedFirewallRules which handles Distributed Firewall
// rules in a VDC Group
type DistributedFirewall struct {
	DistributedFirewallRuleContainer *types.DistributedFirewallRules
	client                           *Client
	VdcGroup                         *VdcGroup
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (d DistributedFirewall) wrap(inner *types.DistributedFirewallRules) *DistributedFirewall {
	d.DistributedFirewallRuleContainer = inner
	return &d
}

// DistributedFirewallRule is a representation of a single rule
type DistributedFirewallRule struct {
	Rule     *types.DistributedFirewallRule
	client   *Client
	VdcGroup *VdcGroup
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (d DistributedFirewallRule) wrap(inner *types.DistributedFirewallRule) *DistributedFirewallRule {
	d.Rule = inner
	return &d
}

// GetDistributedFirewall retrieves Distributed Firewall in a VDC Group which contains all rules
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) GetDistributedFirewall() (*DistributedFirewall, error) {
	c := crudConfig{
		entityLabel:    labelDistributedFirewall,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules,
		endpointParams: []string{vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault},
	}

	outerType := DistributedFirewall{client: vdcGroup.client, VdcGroup: vdcGroup}
	return getOuterEntity[DistributedFirewall, types.DistributedFirewallRules](vdcGroup.client, outerType, c)
}

// UpdateDistributedFirewall updates Distributed Firewall in a VDC Group
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) UpdateDistributedFirewall(dfwRules *types.DistributedFirewallRules) (*DistributedFirewall, error) {
	c := crudConfig{
		entityLabel:    labelDistributedFirewall,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules,
		endpointParams: []string{vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault},
	}

	outerType := DistributedFirewall{client: vdcGroup.client, VdcGroup: vdcGroup}
	return updateOuterEntity[DistributedFirewall, types.DistributedFirewallRules](vdcGroup.client, outerType, c, dfwRules)
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
	c := crudConfig{
		entityLabel:    labelDistributedFirewallRule,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules,
		endpointParams: []string{vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault, "/", id},
	}

	outerType := DistributedFirewallRule{client: vdcGroup.client, VdcGroup: vdcGroup}
	return getOuterEntity[DistributedFirewallRule, types.DistributedFirewallRule](vdcGroup.client, outerType, c)
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

	singleRuleByName, err := localFilterOneOrError(labelDistributedFirewallRule, dfw.DistributedFirewallRuleContainer.Values, "Name", name)
	if err != nil {
		return nil, err
	}

	return vdcGroup.GetDistributedFirewallRuleById(singleRuleByName.ID)
}

// CreateDistributedFirewallRule is a non-thread safe wrapper around
// "vdcGroups/%s/dfwPolicies/%s/rules" endpoint which handles all distributed firewall (DFW) rules
// at once. While there is no real endpoint to create single firewall rule, it is a requirements for
// some cases (e.g. using in Terraform)
// The code works by doing the following steps:
//
// 1. Getting all Distributed Firewall Rules and storing them in private intermediate
// type`distributedFirewallRulesRaw` which holds a []json.RawMessage (text) instead of exact types.
// This will prevent altering existing rules in any way (for example if a new field appears in
// schema in future VCD versions)
//
// 2. Converting the given `rule` into json.RawMessage so that it is provided in the same format as
// other already retrieved rules
//
// 3. Creating a new structure of []json.RawMessage which puts the new rule into one of places:
// 3.1. to the end of []json.RawMessage - bottom of the list
// 3.2. if `optionalAboveRuleId` argument is specified - identifying the position and placing new
// rule above it
// 4. Perform a PUT (update) call to the "vdcGroups/%s/dfwPolicies/%s/rules" endpoint using the
// newly constructed payload
//
// Note. Running this function concurrently will corrupt firewall rules as it uses an endpoint that
// manages all rules ("vdcGroups/%s/dfwPolicies/%s/rules")
func (vdcGroup *VdcGroup) CreateDistributedFirewallRule(optionalAboveRuleId string, rule *types.DistributedFirewallRule) (*DistributedFirewall, *DistributedFirewallRule, error) {
	// 1. Getting all Distributed Firewall Rules and storing them in private intermediate
	// type`distributedFirewallRulesRaw` which holds a []json.RawMessage (text) instead of exact types.
	// This will prevent altering existing rules in any way (for example if a new field appears in
	// schema in future VCD versions)

	c := crudConfig{
		entityLabel:    labelDistributedFirewallRule,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules,
		endpointParams: []string{vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault},
	}
	rawJsonExistingFirewallRules, err := getInnerEntity[distributedFirewallRulesRaw](vdcGroup.client, c)
	if err != nil {
		return nil, nil, err
	}

	// 2. Converting the given `rule` (*types.DistributedFirewallRule) into json.RawMessage so that
	// it is provided in the same format as other already retrieved rules
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

	// 3. Creating a new structure of []json.RawMessage which puts the new rule into one of places:
	switch {
	// 3.1. to the end of []json.RawMessage - bottom of the list (optionalAboveRuleId is empty)
	case optionalAboveRuleId == "":
		rawJsonExistingFirewallRules.Values = append(rawJsonExistingFirewallRules.Values, newRuleRawJson)
		dfwRuleUpdatePayload = rawJsonExistingFirewallRules.Values
		newRuleSlicePosition = len(dfwRuleUpdatePayload) - 1 // -1 to match for slice index

		// 3.2. if `optionalAboveRuleId` argument is specified - identifying the position and placing new
		// rule above it
	case optionalAboveRuleId != "":
		// 3.2.1 Convert '[]json.Rawmessage' to 'types.DistributedFirewallRules'
		dfwRules, err := convertRawJsonToFirewallRules(rawJsonExistingFirewallRules)
		if err != nil {
			return nil, nil, err
		}
		// 3.2.2 Find index for specified 'optionalAboveRuleId' rule
		newFwRuleSliceIndex, err := getFirewallRuleIndexById(dfwRules, optionalAboveRuleId)
		newRuleSlicePosition = newFwRuleSliceIndex // Set rule position for returning single firewall rule
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
	// 4. Perform a PUT (update) call to the "vdcGroups/%s/dfwPolicies/%s/rules" endpoint using the
	// newly constructed payload
	updateRequestPayload := &distributedFirewallRulesRaw{
		Values: dfwRuleUpdatePayload,
	}

	c2 := crudConfig{
		entityLabel:    labelDistributedFirewallRule,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules,
		endpointParams: []string{vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault},
	}

	updatedFirewallRules, err := updateInnerEntity(vdcGroup.client, c2, updateRequestPayload)
	if err != nil {
		return nil, nil, err
	}

	dfwResults, err := convertRawJsonToFirewallRules(updatedFirewallRules)
	if err != nil {
		return nil, nil, err
	}

	returnObjectSingleRule := &DistributedFirewallRule{
		client:   vdcGroup.client,
		VdcGroup: vdcGroup,
		Rule:     dfwResults.Values[newRuleSlicePosition],
	}

	returnAllFirewallRules := &DistributedFirewall{
		DistributedFirewallRuleContainer: dfwResults,
		client:                           vdcGroup.client,
		VdcGroup:                         vdcGroup,
	}

	return returnAllFirewallRules, returnObjectSingleRule, nil
}

// Update a single Distributed Firewall Rule
func (dfwRule *DistributedFirewallRule) Update(rule *types.DistributedFirewallRule) (*DistributedFirewallRule, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules,
		endpointParams: []string{dfwRule.VdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault, "/", dfwRule.Rule.ID},
		entityLabel:    labelDistributedFirewallRule,
	}
	outerType := DistributedFirewallRule{client: dfwRule.client, VdcGroup: dfwRule.VdcGroup}
	return updateOuterEntity(dfwRule.client, outerType, c, rule)
}

// Delete a single Distributed Firewall Rule
func (dfwRule *DistributedFirewallRule) Delete() error {
	c := crudConfig{
		entityLabel:    labelDistributedFirewallRule,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules,
		endpointParams: []string{dfwRule.VdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault, "/", dfwRule.Rule.ID},
	}
	return deleteEntityById(dfwRule.client, c)
}

// getFirewallRuleIndexById searches for 'firewallRuleId' going through a list of available firewall
// rules and returns its index or error if the firewall rule is not found
func getFirewallRuleIndexById(dfwRules *types.DistributedFirewallRules, firewallRuleId string) (int, error) {
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule 'optionalAboveRuleId=%s'. Searching within '%d' items",
		firewallRuleId, len(dfwRules.Values))
	var fwRuleSliceIndex *int
	for index := range dfwRules.Values {
		if dfwRules.Values[index].ID == firewallRuleId {
			// using function `addrOf` to get copy of `index` value as taking a direct address
			// of `&index` will shift before it is used in later code due to how Go range works
			fwRuleSliceIndex = addrOf(index)
			util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule found existing Firewall Rule with ID '%s' at position '%d'",
				firewallRuleId, index)
			continue
		}
	}

	if fwRuleSliceIndex == nil {
		return 0, fmt.Errorf("specified above rule ID '%s' does not exist in current Distributed Firewall Rule list", firewallRuleId)
	}

	return *fwRuleSliceIndex, nil
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
	// Create a new slice with additional capacity of 1 to add new firewall rule into existing list
	newFwRuleSlice := make([]json.RawMessage, len(rawBodyStructure.Values)+1)
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule new container slice of size '%d' with previous element count '%d'", len(newFwRuleSlice), len(rawBodyStructure.Values))
	// if newRulePosition is not 0 (at the top), then previous rules need to be copied to the beginning of new slice
	if newRuleSlicePosition != 0 {
		util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule copying first '%d' slice [:%d]", newRuleSlicePosition, newRuleSlicePosition)
		copy(newFwRuleSlice[:newRuleSlicePosition], rawBodyStructure.Values[:newRuleSlicePosition])
	}

	// Insert the new element at specified index
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule inserting new element into position %d", newRuleSlicePosition)
	newFwRuleSlice[newRuleSlicePosition] = newRuleJsonMessage

	// Copy the remaining elements after new rule
	copy(newFwRuleSlice[newRuleSlicePosition+1:], rawBodyStructure.Values[newRuleSlicePosition:])
	util.Logger.Printf("[DEBUG] CreateDistributedFirewallRule copying remaining items '%d'", newRuleSlicePosition)

	return newFwRuleSlice, nil
}

// distributedFirewallRulesRaw is a copy of `types.DistributedFirewallRules` so that values can be
// unmarshalled into json.RawMessage (as strings) instead of exact types `DistributedFirewallRule`
// It has Public field Values so that marshalling can work, but is not exported itself as it is only
// an intermediate type used in `VdcGroup.CreateDistributedFirewallRule`
type distributedFirewallRulesRaw struct {
	Values []json.RawMessage `json:"values"`
}
