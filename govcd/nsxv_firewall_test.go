// +build nsxv edge firewall functional ALL

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
		Name: "test-firewall",
		Source: types.EdgeFirewallEndpoint{
			VnicGroupIds: []string{"vnic-0"},
		},
		Destination: types.EdgeFirewallEndpoint{
			Exclude: true,
		},
		Application: types.EdgeFirewallApplication{
			Services: []types.EdgeFirewallApplicationService{
				{
					Protocol:   "tcp",
					Port:       "55",
					SourcePort: "44",
				},
				{
					Protocol: "icmp",
				},
			},
		},
		Enabled:        true,
		LoggingEnabled: false,
		Action:         "accept",
	}

	createdFwRule, err := edge.CreateNsxvFirewallRule(firewallRule, "")
	check.Assert(err, IsNil)

	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(createdFwRule.ID, "nsxvFirewallRule", parentEntity, check.TestName())

	gotFwRule, err := edge.GetNsxvFirewallRuleById(createdFwRule.ID)
	check.Assert(err, IsNil)
	check.Assert(gotFwRule, NotNil)
	check.Assert(gotFwRule, DeepEquals, createdFwRule)
	check.Assert(gotFwRule.ID, Equals, createdFwRule.ID)

	// Set ID and update firewall rule with description
	firewallRule.ID = gotFwRule.ID
	firewallRule.Source = types.EdgeFirewallEndpoint{
		IpAddresses: []string{"any"},
	}
	firewallRule.Destination = types.EdgeFirewallEndpoint{
		IpAddresses: []string{"14.14.14.0/24", "17.17.17.0/24"},
	}
	updatedFwRule, err := edge.UpdateNsxvFirewallRule(firewallRule)
	check.Assert(err, IsNil)
	check.Assert(updatedFwRule, NotNil)

	// Check that boolean 'exclude' value was set to false during update
	check.Assert(updatedFwRule.Destination.Exclude, Equals, false)

	// Check if the objects are deeply equal (except updated 'Source' and 'Destination')
	createdFwRule.Source = firewallRule.Source
	createdFwRule.Destination = firewallRule.Destination

	check.Assert(updatedFwRule, DeepEquals, createdFwRule)

	// Add a second rule above the previous one

	firewallRule2 := &types.EdgeFirewallRule{
		Name: "test-firewall-above",
		Source: types.EdgeFirewallEndpoint{
			VnicGroupIds: []string{"vnic-0"},
		},
		Destination: types.EdgeFirewallEndpoint{
			Exclude: true,
		},
		Application: types.EdgeFirewallApplication{
			Services: []types.EdgeFirewallApplicationService{
				{
					Protocol: "any",
				},
			},
		},
		Enabled:        true,
		LoggingEnabled: false,
		Action:         "deny",
	}

	// Create rule 2 above rule 1
	createdFwRule2, err := edge.CreateNsxvFirewallRule(firewallRule2, createdFwRule.ID)
	check.Assert(err, IsNil)
	AddToCleanupList(createdFwRule2.ID, "nsxvFirewallRule", parentEntity, check.TestName())

	// Check rule order and ensure rule 2 is above rule 1 in the list
	allFirewallRules, err := edge.GetAllNsxvFirewallRules()
	check.Assert(err, IsNil)
	var foundRule1BelowRule2, foundRule2AboveRule1 bool
	for _, rule := range allFirewallRules {
		if rule.ID == createdFwRule2.ID {
			foundRule2AboveRule1 = true
		}
		if foundRule2AboveRule1 && rule.ID == createdFwRule.ID {
			foundRule1BelowRule2 = true
		}
	}
	check.Assert(foundRule1BelowRule2, Equals, true)
	check.Assert(foundRule2AboveRule1, Equals, true)

	// Remove both rules
	err = edge.DeleteNsxvFirewallRuleById(gotFwRule.ID)
	check.Assert(err, IsNil)

	err = edge.DeleteNsxvFirewallRuleById(createdFwRule2.ID)
	check.Assert(err, IsNil)

	// Ensure these rule do not exist anymore
	_, err = edge.GetNsxvFirewallRuleById(createdFwRule.ID)
	check.Assert(err, Equals, ErrorEntityNotFound)

	_, err = edge.GetNsxvFirewallRuleById(createdFwRule2.ID)
	check.Assert(err, Equals, ErrorEntityNotFound)

}
