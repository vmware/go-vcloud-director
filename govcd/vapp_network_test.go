// +build vapp functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (vcd *TestVCD) Test_UpdateNetworkFirewallRules(check *C) {
	vapp, networkName, vappNetworkConfig, err := vcd.prepareVappWithVppNetwork(check, "Test_UpdateNetworkFirewallRulesVapp")
	check.Assert(err, IsNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}
	check.Assert(networkFound, Not(Equals), types.VAppNetworkConfiguration{})

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)

	result, err := vapp.UpdateNetworkFirewallRules(uuid, []*types.FirewallRule{&types.FirewallRule{Description: "myFirstRule1", IsEnabled: true, Policy: "allow",
		DestinationPortRange: "Any", DestinationIP: "Any", SourcePortRange: "Any", SourceIP: "Any", Protocols: &types.FirewallRuleProtocols{TCP: true}},
		&types.FirewallRule{Description: "myFirstRule2", IsEnabled: false, Policy: "drop", DestinationPortRange: "Any",
			DestinationIP: "Any", SourcePortRange: "Any", SourceIP: "Any", Protocols: &types.FirewallRuleProtocols{Any: true}}}, "drop", true)
	check.Assert(err, IsNil)
	check.Assert(result, NotNil)

	// verify
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Description, Equals, "myFirstRule1")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].IsEnabled, Equals, true)
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Policy, Equals, "allow")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].DestinationPortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].DestinationIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].SourcePortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].SourceIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Protocols.TCP, Equals, true)
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[0].Protocols.TCP, Equals, true)

	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].Description, Equals, "myFirstRule2")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].IsEnabled, Equals, false)
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].Policy, Equals, "drop")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].DestinationPortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].DestinationIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].SourcePortRange, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].SourceIP, Equals, "Any")
	check.Assert(result.Configuration.Features.FirewallService.FirewallRule[1].Protocols.Any, Equals, true)

	check.Assert(result.Configuration.Features.FirewallService.DefaultAction, Equals, "drop")
	check.Assert(result.Configuration.Features.FirewallService.LogDefaultAction, Equals, true)

	//cleanup
	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}

func (vcd *TestVCD) prepareVappWithVppNetwork(check *C, vappName string) (*VApp, string, *types.NetworkConfigSection, error) {
	fmt.Printf("Running: %s\n", check.TestName())

	vapp, err := createVappForTest(vcd, vappName)
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)

	networkName := "Test_UpdateNetworkFirewallRules"
	description := "Created in test"
	const gateway = "192.168.0.1"
	const netmask = "255.255.255.0"
	const dns1 = "8.8.8.8"
	const dns2 = "1.1.1.1"
	const dnsSuffix = "biz.biz"
	const startAddress = "192.168.0.10"
	const endAddress = "192.168.0.20"
	const dhcpStartAddress = "192.168.0.30"
	const dhcpEndAddress = "192.168.0.40"
	const maxLeaseTime = 3500
	const defaultLeaseTime = 2400
	var guestVlanAllowed = true
	var fwEnabled = false
	var natEnabled = false
	var retainIpMacEnabled = true

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork, NotNil)

	vappNetworkSettings := &VappNetworkSettings{
		Name:               networkName,
		Gateway:            gateway,
		NetMask:            netmask,
		DNS1:               dns1,
		DNS2:               dns2,
		DNSSuffix:          dnsSuffix,
		StaticIPRanges:     []*types.IPRange{{StartAddress: startAddress, EndAddress: endAddress}},
		DhcpSettings:       &DhcpSettings{IsEnabled: true, MaxLeaseTime: maxLeaseTime, DefaultLeaseTime: defaultLeaseTime, IPRange: &types.IPRange{StartAddress: dhcpStartAddress, EndAddress: dhcpEndAddress}},
		GuestVLANAllowed:   &guestVlanAllowed,
		Description:        description,
		FirewallEnabled:    &fwEnabled,
		NatEnabled:         &natEnabled,
		RetainIpMacEnabled: &retainIpMacEnabled,
	}

	vappNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, orgVdcNetwork.OrgVDCNetwork)
	check.Assert(err, IsNil)
	check.Assert(vappNetworkConfig, NotNil)
	return vapp, networkName, vappNetworkConfig, err
}

func (vcd *TestVCD) Test_GetVappNetworkByNameOrId(check *C) {
	vapp, networkName, vappNetworkConfig, err := vcd.prepareVappWithVppNetwork(check, "Test_GetVappNetworkByNameOrId")
	check.Assert(err, IsNil)

	networkFound := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vappNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			networkFound = networkConfig
		}
	}
	check.Assert(networkFound, Not(Equals), types.VAppNetworkConfiguration{})

	uuid, err := GetUuidFromHref(networkFound.Link.HREF, false)
	check.Assert(err, IsNil)

	vappNetwork, err := vapp.GetVappNetworkById(uuid, true)
	check.Assert(err, IsNil)
	check.Assert(vappNetwork, NotNil)

	vappNetwork, err = vapp.GetVappNetworkByName(networkName, false)
	check.Assert(err, IsNil)
	check.Assert(vappNetwork, NotNil)

	//cleanup
	task, err := vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
}
