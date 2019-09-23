// +build edge firewall functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxvFirewallRule(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	firewallRule := &types.EdgeFirewallRule{
		Name:        "test-firewall",
		Description: "test-firewall description",
		Source: types.EdgeFirewallObject{
			VnicGroupId: "vnic-0",
		},
		Destination: types.EdgeFirewallObject{
			Exclude: true,
		},
		Application: types.EdgeFirewallApplication{
			Service: types.EdgeFirewallApplicationService{
				Protocol:   "tcp",
				Port:       "55",
				SourcePort: "44",
			},
		},
		Enabled:        true,
		LoggingEnabled: false,
		Action:         "accept",
	}

	createdFwRule, err := edge.CreateNsxvFirewall(firewallRule)
	check.Assert(err, IsNil)

	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(createdFwRule.ID, "nsxvFirewallRule", parentEntity, check.TestName())

	gotFwRule, err := edge.GetNsxvFirewallById(createdFwRule.ID)
	check.Assert(err, IsNil)
	check.Assert(gotFwRule, NotNil)
	check.Assert(gotFwRule, DeepEquals, createdFwRule)
	check.Assert(gotFwRule.ID, Equals, createdFwRule.ID)

	// Set ID and update nat rule with description
	firewallRule.ID = gotFwRule.ID
	firewallRule.Description = "Description for NAT rule"
	updatedFwRule, err := edge.UpdateNsxvFirewall(firewallRule)
	check.Assert(err, IsNil)
	check.Assert(updatedFwRule, NotNil)

	check.Assert(updatedFwRule.Description, Equals, firewallRule.Description)

	// Check if the objects are deeply equal (except updated 'Description' field)
	createdFwRule.Description = firewallRule.Description
	check.Assert(updatedFwRule, DeepEquals, createdFwRule)

	err = edge.DeleteNsxvNatRuleById(gotFwRule.ID)
	check.Assert(err, IsNil)

	// Ensure the rule does not exist anymore
	_, err = edge.GetNsxvNatRuleById(createdFwRule.ID)
	check.Assert(err, Equals, ErrorEntityNotFound)

}
