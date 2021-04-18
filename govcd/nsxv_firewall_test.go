// +build nsxv edge firewall functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxvFirewallRuleGatewayInterfaces tests firewall rule creation based on vNics (gateway
// interfaces in UI)
func (vcd *TestVCD) Test_NsxvFirewallRuleGatewayInterfaces(check *C) {
	firewallRuleConfig := &types.EdgeFirewallRule{
		Name: "test-firewall-interface",
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

	test_NsxvFirewallRule(check, vcd, firewallRuleConfig)
}

// Test_NsxvFirewallRuleIpAddresses tests firewall rule creation based on IP addresses
func (vcd *TestVCD) Test_NsxvFirewallRuleIpAddresses(check *C) {
	firewallRuleConfig := &types.EdgeFirewallRule{
		Name: "test-firewall-ips",
		Source: types.EdgeFirewallEndpoint{
			IpAddresses: []string{"1.1.1.1", "2.2.2.2/24"},
		},
		Destination: types.EdgeFirewallEndpoint{
			// Excludes works like a boolean ! (not) for a specified list of objects
			Exclude:     true,
			IpAddresses: []string{"any"}, // "any" is a keyword that matches all
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

	test_NsxvFirewallRule(check, vcd, firewallRuleConfig)
}

// Test_NsxvFirewallRuleIpSets tests firewall rule creation based on IP set IDs
func (vcd *TestVCD) Test_NsxvFirewallRuleIpSets(check *C) {
	if vcd.config.VCD.Org == "" {
		check.Skip(check.TestName() + ": Org name not given")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip(check.TestName() + ": VDC name not given")
		return
	}

	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(ctx, vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// IP set based firewall rule
	// Create two IP sets
	ipSetConfig1 := &types.EdgeIpSet{
		Name:               "test-ipset-1",
		IPAddresses:        "10.10.10.1",
		InheritanceAllowed: takeBoolPointer(true), // Must be true to allow using it in firewall rule
	}

	ipSetConfig2 := &types.EdgeIpSet{
		Name:               "test-ipset-2",
		IPAddresses:        "192.168.1.1-192.168.1.200",
		InheritanceAllowed: takeBoolPointer(true), // Must be true to allow using it in firewall rule
	}

	// Set parent entity and create two IP sets for usage in firewall rule
	ipSetParentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name
	ipSet1, err := vdc.CreateNsxvIpSet(ctx, ipSetConfig1)
	check.Assert(err, IsNil)
	AddToCleanupList(ipSet1.Name, "ipSet", ipSetParentEntity, check.TestName())
	check.Assert(ipSet1.ID, Matches, `.*:ipset-\d.*`)

	ipSet2, err := vdc.CreateNsxvIpSet(ctx, ipSetConfig2)
	check.Assert(err, IsNil)
	AddToCleanupList(ipSet2.Name, "ipSet", ipSetParentEntity, check.TestName())
	check.Assert(ipSet2.ID, Matches, `.*:ipset-\d.*`)

	firewallRuleConfig := &types.EdgeFirewallRule{
		Name:           "test-firewall-ipsets",
		Enabled:        true,
		LoggingEnabled: false,
		Action:         "deny",
		Source: types.EdgeFirewallEndpoint{
			GroupingObjectIds: []string{ipSet1.ID},
		},
		Destination: types.EdgeFirewallEndpoint{
			GroupingObjectIds: []string{ipSet1.ID},
		},
	}

	test_NsxvFirewallRule(check, vcd, firewallRuleConfig)
}

// Test_NsxvFirewallRuleVms tests firewall rule creation based on VM
func (vcd *TestVCD) Test_NsxvFirewallRuleVms(check *C) {
	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp wasn't properly created")
	}

	vapp := vcd.findFirstVapp(ctx)
	if vapp.VApp.Name == "" {
		check.Skip("Disabled: No suitable vApp found in vDC")
	}
	vm, _ := vcd.findFirstVm(vapp)
	if vm.Name == "" {
		check.Skip("Disabled: No suitable VM found in vDC")
	}

	firewallRuleConfig := &types.EdgeFirewallRule{
		Name: "test-firewall-vms",
		Source: types.EdgeFirewallEndpoint{
			GroupingObjectIds: []string{vm.ID},
		},
		Destination: types.EdgeFirewallEndpoint{
			Exclude:     true,
			IpAddresses: []string{"any"},
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

	test_NsxvFirewallRule(check, vcd, firewallRuleConfig)
}

// test_NsxvFirewallRule does the test work for a given fwConfig of *types.EdgeFirewallRule
// The tests performed consist of the following steps:
// 1. Creates a firewall rule (as per specified config)
// 2. Retrieve the rule by ID
// 3. Perform an update for the firewall rule
// 4. Creates a second firewall rule above the first one
// 5. Ensures the positions of both rules are correct (second is above first one)
// 6. Deletes both firewall rules
// 7. Validates that none are found anymore
func test_NsxvFirewallRule(check *C, vcd *TestVCD, fwConfig *types.EdgeFirewallRule) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip(check.TestName() + ": no edge gateway given")
	}
	if vcd.config.VCD.Org == "" {
		check.Skip(check.TestName() + ": Org name not given")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip(check.TestName() + ": VDC name not given")
		return
	}

	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(ctx, vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	firewallRule := fwConfig

	createdFwRule, err := edge.CreateNsxvFirewallRule(ctx, firewallRule, "")
	check.Assert(err, IsNil)

	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(createdFwRule.ID, "nsxvFirewallRule", parentEntity, check.TestName())

	gotFwRule, err := edge.GetNsxvFirewallRuleById(ctx, createdFwRule.ID)
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
	updatedFwRule, err := edge.UpdateNsxvFirewallRule(ctx, firewallRule)
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
	createdFwRule2, err := edge.CreateNsxvFirewallRule(ctx, firewallRule2, createdFwRule.ID)
	check.Assert(err, IsNil)
	AddToCleanupList(createdFwRule2.ID, "nsxvFirewallRule", parentEntity, check.TestName())

	// Check rule order and ensure rule 2 is above rule 1 in the list
	allFirewallRules, err := edge.GetAllNsxvFirewallRules(ctx)
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

	// Remove all rules
	err = edge.DeleteNsxvFirewallRuleById(ctx, gotFwRule.ID)
	check.Assert(err, IsNil)

	err = edge.DeleteNsxvFirewallRuleById(ctx, createdFwRule2.ID)
	check.Assert(err, IsNil)

	// Ensure these rule do not exist anymore
	_, err = edge.GetNsxvFirewallRuleById(ctx, createdFwRule.ID)
	check.Assert(err, Equals, ErrorEntityNotFound)

	_, err = edge.GetNsxvFirewallRuleById(ctx, createdFwRule2.ID)
	check.Assert(err, Equals, ErrorEntityNotFound)

}
