// +build lb lbAppRule nsxv functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Test_LBAppRule tests CRUD methods for load balancer application rule.
// The following things are tested if prerequisite Edge Gateway exists:
// 1. Creation of load balancer application rule
// 2. Get load balancer application rule by both ID and Name (application rule name must be unique in single edge gateway)
// 3. Update - change a field and compare that configuration and result objects are deeply equal
// 4. Update - try and fail to update without mandatory field
// 5. Delete
func (vcd *TestVCD) Test_LBAppRule(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	// Used for creating
	lbAppRuleConfig := &types.LbAppRule{
		Name:   TestLbAppRule,
		Script: "acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page",
	}

	err = deleteLbAppRuleIfExists(ctx, *edge, lbAppRuleConfig.Name)
	check.Assert(err, IsNil)
	createdLbAppRule, err := edge.CreateLbAppRule(ctx, lbAppRuleConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbAppRule.ID, Not(IsNil))

	// // We created application rule successfully therefore let's add it to cleanup list
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(TestLbAppRule, "lbAppRule", parentEntity, check.TestName())

	// // Lookup by both name and ID and compare that these are equal values
	lbAppRuleByID, err := edge.getLbAppRule(ctx, &types.LbAppRule{ID: createdLbAppRule.ID})
	check.Assert(err, IsNil)
	check.Assert(lbAppRuleByID, Not(IsNil))

	lbAppRuleByName, err := edge.getLbAppRule(ctx, &types.LbAppRule{Name: createdLbAppRule.Name})
	check.Assert(err, IsNil)
	check.Assert(lbAppRuleByName, Not(IsNil))
	check.Assert(createdLbAppRule.ID, Equals, lbAppRuleByName.ID)
	check.Assert(lbAppRuleByID.ID, Equals, lbAppRuleByName.ID)
	check.Assert(lbAppRuleByID.Name, Equals, lbAppRuleByName.Name)

	check.Assert(createdLbAppRule.Script, Equals, lbAppRuleConfig.Script)

	// Test that we can extract a list of LB app rules, and that one of them is the rule we have got when searching by name
	lbAppRules, err := edge.GetLbAppRules(ctx)
	check.Assert(err, IsNil)
	check.Assert(lbAppRules, NotNil)
	foundRule := false
	for _, rule := range lbAppRules {
		if rule.Name == lbAppRuleByName.Name && rule.ID == lbAppRuleByName.ID {
			foundRule = true
		}
	}
	check.Assert(foundRule, Equals, true)

	// Test updating fields
	// Update script to be multi-line
	lbAppRuleByID.Script = "acl other_page url_beg / other redirect location https://www.other.com/ ifother_page\n" +
		"acl other_page2 url_beg / other2 redirect location https://www.other2.com/ ifother_page2"
	updatedAppProfile, err := edge.UpdateLbAppRule(ctx, lbAppRuleByID)
	check.Assert(err, IsNil)
	check.Assert(updatedAppProfile.Script, Equals, lbAppRuleByID.Script)

	// Verify that updated pool and its configuration are identical
	check.Assert(updatedAppProfile, DeepEquals, lbAppRuleByID)

	// Try to set invalid script expect API to return error
	// invalid applicationRule script, invalid script line : invalid_script, error details :
	// Unknown keyword 'invalid_script'
	lbAppRuleByID.Script = "invalid_script"
	updatedAppProfile, err = edge.UpdateLbAppRule(ctx, lbAppRuleByID)
	check.Assert(updatedAppProfile, IsNil)
	check.Assert(err, NotNil)

	// Update should fail without name
	lbAppRuleByID.Name = ""
	_, err = edge.UpdateLbAppRule(ctx, lbAppRuleByID)
	check.Assert(err.Error(), Equals, "load balancer application rule Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLbAppRule(ctx, &types.LbAppRule{ID: createdLbAppRule.ID})
	check.Assert(err, IsNil)

	// Ensure it is deleted
	_, err = edge.GetLbAppRuleById(ctx, createdLbAppRule.ID)
	check.Assert(IsNotFound(err), Equals, true)
}
