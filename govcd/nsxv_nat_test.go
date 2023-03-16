//go:build edge || nat || nsxv || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxvSnatRule(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	vnicIndex, err := edge.GetVnicIndexByNetworkNameAndType(vcd.config.VCD.Network.Net1, "internal")
	check.Assert(err, IsNil)

	natRule := &types.EdgeNatRule{
		Action:            "snat",
		Vnic:              vnicIndex,
		OriginalAddress:   vcd.config.VCD.InternalIp,
		TranslatedAddress: vcd.config.VCD.ExternalIp,
		Enabled:           true,
		LoggingEnabled:    true,
		Description:       "my description",
	}
	if testVerbose {
		fmt.Printf("# %s %s %s -> %s\n", natRule.Action, natRule.Protocol, natRule.OriginalAddress,
			natRule.TranslatedAddress)
	}
	testNsxvNat(natRule, vcd, check, *edge)
}
func (vcd *TestVCD) Test_NsxvDnatRule(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	vnicIndex, err := edge.GetVnicIndexByNetworkNameAndType(vcd.config.VCD.ExternalNetwork, "uplink")
	check.Assert(err, IsNil)

	natRule := &types.EdgeNatRule{
		Action:            "dnat",
		Vnic:              vnicIndex,
		Protocol:          "tcp",
		OriginalAddress:   vcd.config.VCD.ExternalIp,
		OriginalPort:      "443",
		TranslatedAddress: vcd.config.VCD.InternalIp,
		TranslatedPort:    "8443",
		Enabled:           true,
		LoggingEnabled:    true,
		Description:       "my description",
	}
	if testVerbose {
		fmt.Printf("# %s %s %s:%s -> %s:%s\n", natRule.Action, natRule.Protocol, natRule.OriginalAddress,
			natRule.OriginalPort, natRule.TranslatedAddress, natRule.TranslatedPort)
	}

	testNsxvNat(natRule, vcd, check, *edge)

	natRule = &types.EdgeNatRule{
		Action:            "dnat",
		Vnic:              vnicIndex,
		Protocol:          "icmp",
		IcmpType:          "router-advertisement",
		OriginalAddress:   vcd.config.VCD.ExternalIp,
		TranslatedAddress: vcd.config.VCD.InternalIp,
		Enabled:           true,
		LoggingEnabled:    true,
		Description:       "my description",
	}
	if testVerbose {
		fmt.Printf("# %s %s:%s %s -> %s\n", natRule.Action, natRule.Protocol, natRule.IcmpType,
			natRule.OriginalAddress, natRule.TranslatedAddress)
	}
	testNsxvNat(natRule, vcd, check, *edge)

	natRule = &types.EdgeNatRule{
		Action:            "dnat",
		Vnic:              vnicIndex,
		Protocol:          "any",
		OriginalAddress:   vcd.config.VCD.ExternalIp,
		TranslatedAddress: vcd.config.VCD.InternalIp,
		Enabled:           true,
		LoggingEnabled:    true,
		Description:       "my description",
	}
	if testVerbose {
		fmt.Printf("# %s %s %s -> %s\n", natRule.Action, natRule.Protocol, natRule.OriginalAddress,
			natRule.TranslatedAddress)
	}
	testNsxvNat(natRule, vcd, check, *edge)
}

// testNsxvNat is a helper to test multiple configurations of NAT rules. It does the following
// 1. Creates NAT rule with provided config
// 2. Checks that it can be retrieve and verifies if IDs match
// 3. Tries to update description field and validates that nothing else except description changes
// 4. Deletes the rule
// 5. Validates that the rule was deleted
func testNsxvNat(natRule *types.EdgeNatRule, vcd *TestVCD, check *C, edge EdgeGateway) {
	createdNatRule, err := edge.CreateNsxvNatRule(natRule)
	check.Assert(err, IsNil)

	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(createdNatRule.ID, "nsxvNatRule", parentEntity, check.TestName())

	gotNatRule, err := edge.GetNsxvNatRuleById(createdNatRule.ID)
	check.Assert(err, IsNil)
	check.Assert(gotNatRule, NotNil)
	check.Assert(gotNatRule, DeepEquals, createdNatRule)
	check.Assert(gotNatRule.ID, Equals, createdNatRule.ID)

	// Set ID and update nat rule with description
	natRule.ID = gotNatRule.ID
	natRule.Description = "Description for NAT rule"
	updatedNatRule, err := edge.UpdateNsxvNatRule(natRule)
	check.Assert(err, IsNil)
	check.Assert(updatedNatRule, NotNil)

	check.Assert(updatedNatRule.Description, Equals, natRule.Description)

	// Test that we can extract a list of NSXV NAT rules, and that one of them is the rule we have got when searching by ID
	natRules, err := edge.GetNsxvNatRules()
	check.Assert(err, IsNil)
	check.Assert(natRules, NotNil)
	foundRule := false
	for _, rule := range natRules {
		if rule.ID == natRule.ID {
			foundRule = true
		}
	}
	check.Assert(foundRule, Equals, true)

	// Check if the objects are deeply equal (except updated 'Description' field)
	createdNatRule.Description = natRule.Description
	check.Assert(updatedNatRule, DeepEquals, createdNatRule)

	err = edge.DeleteNsxvNatRuleById(gotNatRule.ID)
	check.Assert(err, IsNil)

	// Ensure the rule does not exist anymore
	_, err = edge.GetNsxvNatRuleById(createdNatRule.ID)
	check.Assert(IsNotFound(err), Equals, true)
}
