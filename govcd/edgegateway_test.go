// +build gateway functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"regexp"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_RefreshEdgeGateway(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)
	copyEdge := edge
	err = edge.Refresh()
	check.Assert(err, IsNil)
	check.Assert(copyEdge.EdgeGateway.Name, Equals, edge.EdgeGateway.Name)
	check.Assert(copyEdge.EdgeGateway.HREF, Equals, edge.EdgeGateway.HREF)
}

// TODO: Add a check for the final state of the mapping
func (vcd *TestVCD) Test_NATMapping(check *C) {
	if vcd.config.VCD.ExternalIp == "" || vcd.config.VCD.InternalIp == "" {
		check.Skip("Skipping test because no valid ip given")
	}
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	task, err := edge.AddNATRule(orgVdcNetwork.OrgVDCNetwork, "DNAT", vcd.config.VCD.ExternalIp, vcd.config.VCD.InternalIp)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	err = edge.Refresh()
	check.Assert(err, IsNil)
	found := false
	var rule *types.NatRule
	for _, r := range edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
		if r.RuleType == "DNAT" && r.GatewayNatRule.Interface.Name == orgVdcNetwork.OrgVDCNetwork.Name {
			found = true
			rule = r
		}
	}

	check.Assert(found, Equals, true)
	check.Assert(rule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(rule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)

	//task, err = edge.Remove1to1Mapping(vcd.config.VCD.InternalIp, vcd.config.VCD.ExternalIp)
	// Cause Remove1to1Mapping isn't working correctly we will use new function
	err = edge.RemoveNATRule(rule.ID)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// TODO: Add a check for the final state of the mapping
func (vcd *TestVCD) Test_NATPortMapping(check *C) {
	if vcd.config.VCD.ExternalIp == "" || vcd.config.VCD.InternalIp == "" {
		check.Skip("Skipping test because no valid ip given")
	}
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	task, err := edge.AddNATPortMappingWithUplink(orgVdcNetwork.OrgVDCNetwork, "DNAT", vcd.config.VCD.ExternalIp, "1177", vcd.config.VCD.InternalIp, "77", "TCP", "")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	err = edge.Refresh()
	check.Assert(err, IsNil)
	found := false
	var rule *types.NatRule
	for _, r := range edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule {
		if r.RuleType == "DNAT" && r.GatewayNatRule.Interface.Name == orgVdcNetwork.OrgVDCNetwork.Name {
			found = true
			rule = r
		}
	}

	check.Assert(found, Equals, true)
	check.Assert(rule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(rule.GatewayNatRule.TranslatedPort, Equals, "77")
	check.Assert(rule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(rule.GatewayNatRule.OriginalPort, Equals, "1177")
	check.Assert(rule.GatewayNatRule.Protocol, Equals, "tcp")
	check.Assert(rule.GatewayNatRule.IcmpSubType, Equals, "")

	//task, err = edge.Remove1to1Mapping(vcd.config.VCD.InternalIp, vcd.config.VCD.ExternalIp)
	// Cause Remove1to1Mapping isn't working correctly we will use new function
	err = edge.RemoveNATRule(rule.ID)

	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

// TODO: Add a check for the final state of the mapping
func (vcd *TestVCD) Test_1to1Mappings(check *C) {
	if vcd.config.VCD.ExternalIp == "" || vcd.config.VCD.InternalIp == "" {
		check.Skip("Skipping test because no valid ip given")
	}
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edgegatway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)
	task, err := edge.Create1to1Mapping(vcd.config.VCD.InternalIp, vcd.config.VCD.ExternalIp, "description")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = edge.Remove1to1Mapping(vcd.config.VCD.InternalIp, vcd.config.VCD.ExternalIp)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_AddIpsecVPN(check *C) {
	if vcd.config.VCD.ExternalIp == "" {
		check.Skip("Skipping test because no valid ip given")
	}
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edgegatway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	// Check that the minimal input is included
	check.Assert(vcd.config.VCD.InternalIp, Not(Equals), "")
	check.Assert(vcd.config.VCD.InternalNetmask, Not(Equals), "")
	check.Assert(vcd.config.VCD.ExternalIp, Not(Equals), "")
	check.Assert(vcd.config.VCD.ExternalNetmask, Not(Equals), "")

	tunnel := &types.GatewayIpsecVpnTunnel{
		Name:               "TestVPN_API",
		Description:        "Testing VPN Creation",
		EncryptionProtocol: "AES",
		SharedSecret:       "MadeUpWords",             // MANDATORY
		LocalIPAddress:     vcd.config.VCD.ExternalIp, // MANDATORY
		LocalID:            vcd.config.VCD.ExternalIp, // MANDATORY
		PeerIPAddress:      vcd.config.VCD.InternalIp, // MANDATORY
		PeerID:             vcd.config.VCD.InternalIp, // MANDATORY
		IsEnabled:          true,
		LocalSubnet: []*types.IpsecVpnSubnet{
			&types.IpsecVpnSubnet{
				Name:    vcd.config.VCD.ExternalIp,
				Gateway: vcd.config.VCD.ExternalIp,      // MANDATORY
				Netmask: vcd.config.VCD.ExternalNetmask, // MANDATORY
			},
		},
		PeerSubnet: []*types.IpsecVpnSubnet{
			&types.IpsecVpnSubnet{
				Name:    vcd.config.VCD.InternalIp,
				Gateway: vcd.config.VCD.InternalIp,      // MANDATORY
				Netmask: vcd.config.VCD.InternalNetmask, // MANDATORY
			},
		},
	}
	tunnels := make([]*types.GatewayIpsecVpnTunnel, 1)
	tunnels[0] = tunnel
	ipsecVPNConfig := &types.EdgeGatewayServiceConfiguration{
		GatewayIpsecVpnService: &types.GatewayIpsecVpnService{
			IsEnabled: true,
			Tunnel:    tunnels,
		},
	}

	// Configures VPN service
	task, err := edge.AddIpsecVPN(ipsecVPNConfig)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// To check the effects of service configuration, we need to reload the edge gateway entity
	err = edge.Refresh()
	check.Assert(err, IsNil)

	// We expect an enabled service, and non-null tunnel and endpoint
	newConf := edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration
	newConfState := newConf.GatewayIpsecVpnService.IsEnabled
	newConfTunnel := newConf.GatewayIpsecVpnService.Tunnel

	// TODO: assumption about not nil endpoints doesn't hold for all vCD versions and configurations
	// Needs research
	//newConfEndpoint := newConf.GatewayIpsecVpnService.Endpoint
	check.Assert(newConfState, Equals, true)
	check.Assert(newConfTunnel, NotNil)
	// check.Assert(newConfEndpoint, NotNil)

	// Removes VPN service
	task, err = edge.RemoveIpsecVPN()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// To check the effects of service configuration, we need to reload the edge gateway entity
	err = edge.Refresh()
	check.Assert(err, IsNil)

	// We expect a disabled service, and null tunnel and endpoint
	afterDeletionConf := edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration
	newConfState = afterDeletionConf.GatewayIpsecVpnService.IsEnabled
	newConfTunnel = afterDeletionConf.GatewayIpsecVpnService.Tunnel
	newConfEndpoint := afterDeletionConf.GatewayIpsecVpnService.Endpoint
	check.Assert(newConfState, Equals, false)
	check.Assert(newConfTunnel, IsNil)
	check.Assert(newConfEndpoint, IsNil)
}

func (vcd *TestVCD) TestEdgeGateway_GetNetworks(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gatway given")
	}
	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("Skipping test because no external network given")
	}
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)
	network, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	isRouted := false
	// If the network is not linked to the edge gateway, we won't check for its name in the network list
	if network.OrgVDCNetwork.EdgeGateway != nil {
		isRouted = true
	}

	var networkList []SimpleNetworkIdentifier
	networkList, err = edge.GetNetworks()
	check.Assert(err, IsNil)
	foundExternalNetwork := false
	foundNetwork := false
	for _, net := range networkList {
		if net.Name == vcd.config.VCD.ExternalNetwork && net.InterfaceType == "uplink" {
			foundExternalNetwork = true
		}
		if net.Name == vcd.config.VCD.Network.Net1 && net.InterfaceType == "internal" {
			foundNetwork = true
		}
	}
	check.Assert(foundExternalNetwork, Equals, true)
	if isRouted {
		check.Assert(foundNetwork, Equals, true)
	}

}

func (vcd *TestVCD) Test_AddSNATRule(check *C) {
	if vcd.config.VCD.ExternalIp == "" || vcd.config.VCD.InternalIp == "" {
		check.Skip("Skipping test because no valid ip given")
	}
	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("Skipping test because no external network given")
	}
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	description1 := "my Description 1"
	description2 := "my Description 2"

	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(externalNetwork, NotNil)
	check.Assert(externalNetwork.ExternalNetwork.Name, Equals, vcd.config.VCD.ExternalNetwork)

	beforeChangeNatRulesNumber := len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule)

	natRule, err := edge.AddSNATRule(orgVdcNetwork.OrgVDCNetwork.HREF, vcd.config.VCD.ExternalIp, vcd.config.VCD.InternalIp, description1)
	check.Assert(err, IsNil)

	check.Assert(natRule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(natRule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(natRule.Description, Equals, description1)
	check.Assert(natRule.RuleType, Equals, "SNAT")
	check.Assert(strings.Split(natRule.GatewayNatRule.Interface.HREF, "network/")[1], Equals, strings.Split(orgVdcNetwork.OrgVDCNetwork.HREF, "network/")[1])

	err = edge.RemoveNATRule(natRule.ID)
	check.Assert(err, IsNil)

	// verify delete
	err = edge.Refresh()
	check.Assert(err, IsNil)

	check.Assert(len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule), Equals, beforeChangeNatRulesNumber)

	// check with external network
	natRule, err = edge.AddSNATRule(externalNetwork.ExternalNetwork.HREF, vcd.config.VCD.InternalIp, vcd.config.VCD.ExternalIp, description2)
	check.Assert(err, IsNil)

	check.Assert(natRule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(natRule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(natRule.Description, Equals, description2)
	check.Assert(natRule.RuleType, Equals, "SNAT")
	check.Assert(strings.Split(natRule.GatewayNatRule.Interface.HREF, "network/")[1], Equals, strings.Split(externalNetwork.ExternalNetwork.HREF, "externalnet/")[1])

	err = edge.RemoveNATRule(natRule.ID)
	check.Assert(err, IsNil)

	// verify delete
	err = edge.Refresh()
	check.Assert(err, IsNil)

	check.Assert(len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule), Equals, beforeChangeNatRulesNumber)

}

func (vcd *TestVCD) Test_AddDNATRule(check *C) {
	if vcd.config.VCD.ExternalIp == "" || vcd.config.VCD.InternalIp == "" {
		check.Skip("Skipping test because no valid ip given")
	}
	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("Skipping test because no external network given")
	}
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(externalNetwork, NotNil)
	check.Assert(externalNetwork.ExternalNetwork.Name, Equals, vcd.config.VCD.ExternalNetwork)

	beforeChangeNatRulesNumber := len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule)

	description1 := "my Dnat Description 1"
	description2 := "my Dnatt Description 2"

	natRule, err := edge.AddDNATRule(NatRule{NetworkHref: orgVdcNetwork.OrgVDCNetwork.HREF, ExternalIP: vcd.config.VCD.ExternalIp,
		ExternalPort: "1177", InternalIP: vcd.config.VCD.InternalIp, InternalPort: "77", Protocol: "TCP", Description: description1})
	check.Assert(err, IsNil)

	check.Assert(natRule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(natRule.GatewayNatRule.TranslatedPort, Equals, "77")
	check.Assert(natRule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(natRule.GatewayNatRule.OriginalPort, Equals, "1177")
	check.Assert(natRule.GatewayNatRule.Protocol, Equals, "tcp")
	check.Assert(natRule.GatewayNatRule.IcmpSubType, Equals, "")
	check.Assert(natRule.Description, Equals, description1)
	check.Assert(natRule.RuleType, Equals, "DNAT")
	check.Assert(strings.Split(natRule.GatewayNatRule.Interface.HREF, "network/")[1], Equals, strings.Split(orgVdcNetwork.OrgVDCNetwork.HREF, "network/")[1])

	err = edge.RemoveNATRule(natRule.ID)
	check.Assert(err, IsNil)

	// verify delete
	err = edge.Refresh()
	check.Assert(err, IsNil)

	check.Assert(len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule), Equals, beforeChangeNatRulesNumber)

	// check with external network
	natRule, err = edge.AddDNATRule(NatRule{NetworkHref: externalNetwork.ExternalNetwork.HREF, ExternalIP: vcd.config.VCD.ExternalIp,
		ExternalPort: "1188", InternalIP: vcd.config.VCD.InternalIp, InternalPort: "88", Protocol: "TCP", Description: description2})
	check.Assert(err, IsNil)

	check.Assert(natRule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(natRule.GatewayNatRule.TranslatedPort, Equals, "88")
	check.Assert(natRule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(natRule.GatewayNatRule.OriginalPort, Equals, "1188")
	check.Assert(natRule.GatewayNatRule.Protocol, Equals, "tcp")
	check.Assert(natRule.GatewayNatRule.IcmpSubType, Equals, "")
	check.Assert(natRule.Description, Equals, description2)
	check.Assert(natRule.RuleType, Equals, "DNAT")
	check.Assert(strings.Split(natRule.GatewayNatRule.Interface.HREF, "network/")[1], Equals, strings.Split(externalNetwork.ExternalNetwork.HREF, "externalnet/")[1])

	err = edge.RemoveNATRule(natRule.ID)
	check.Assert(err, IsNil)

	// verify delete
	err = edge.Refresh()
	check.Assert(err, IsNil)

	check.Assert(len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule), Equals, beforeChangeNatRulesNumber)
}

func (vcd *TestVCD) Test_UpdateNATRule(check *C) {
	if vcd.config.VCD.ExternalIp == "" || vcd.config.VCD.InternalIp == "" {
		check.Skip("Skipping test because no valid ip given")
	}
	if vcd.config.VCD.ExternalNetwork == "" {
		check.Skip("Skipping test because no external network given")
	}
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	if vcd.config.VCD.Network.Net1 == "" {
		check.Skip("Skipping test because no network was given")
	}

	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	orgVdcNetwork, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	check.Assert(orgVdcNetwork.OrgVDCNetwork.Name, Equals, vcd.config.VCD.Network.Net1)

	externalNetwork, err := vcd.client.GetExternalNetworkByName(vcd.config.VCD.ExternalNetwork)
	check.Assert(err, IsNil)
	check.Assert(externalNetwork, NotNil)
	check.Assert(externalNetwork.ExternalNetwork.Name, Equals, vcd.config.VCD.ExternalNetwork)

	beforeChangeNatRulesNumber := len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule)

	description1 := "my Dnat Description 1"
	description2 := "my Dnatt Description 2"

	natRule, err := edge.AddDNATRule(NatRule{NetworkHref: orgVdcNetwork.OrgVDCNetwork.HREF, ExternalIP: vcd.config.VCD.ExternalIp,
		ExternalPort: "1177", InternalIP: vcd.config.VCD.InternalIp, InternalPort: "77", Protocol: "TCP", Description: description1})
	check.Assert(err, IsNil)

	check.Assert(natRule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(natRule.GatewayNatRule.TranslatedPort, Equals, "77")
	check.Assert(natRule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(natRule.GatewayNatRule.OriginalPort, Equals, "1177")
	check.Assert(natRule.GatewayNatRule.Protocol, Equals, "tcp")
	check.Assert(natRule.GatewayNatRule.IcmpSubType, Equals, "")
	check.Assert(natRule.Description, Equals, description1)
	check.Assert(natRule.RuleType, Equals, "DNAT")
	check.Assert(strings.Split(natRule.GatewayNatRule.Interface.HREF, "network/")[1], Equals, strings.Split(orgVdcNetwork.OrgVDCNetwork.HREF, "network/")[1])

	err = edge.RemoveNATRule(natRule.ID)
	check.Assert(err, IsNil)

	// verify delete
	err = edge.Refresh()
	check.Assert(err, IsNil)

	check.Assert(len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule), Equals, beforeChangeNatRulesNumber)

	// check with external network
	natRule, err = edge.AddDNATRule(NatRule{NetworkHref: externalNetwork.ExternalNetwork.HREF, ExternalIP: vcd.config.VCD.ExternalIp,
		ExternalPort: "1188", InternalIP: vcd.config.VCD.InternalIp, InternalPort: "88", Protocol: "TCP", Description: description2})
	check.Assert(err, IsNil)

	check.Assert(natRule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(natRule.GatewayNatRule.TranslatedPort, Equals, "88")
	check.Assert(natRule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(natRule.GatewayNatRule.OriginalPort, Equals, "1188")
	check.Assert(natRule.GatewayNatRule.Protocol, Equals, "tcp")
	check.Assert(natRule.GatewayNatRule.IcmpSubType, Equals, "")
	check.Assert(natRule.Description, Equals, description2)
	check.Assert(natRule.RuleType, Equals, "DNAT")
	check.Assert(strings.Split(natRule.GatewayNatRule.Interface.HREF, "network/")[1], Equals, strings.Split(externalNetwork.ExternalNetwork.HREF, "externalnet/")[1])

	err = edge.RemoveNATRule(natRule.ID)
	check.Assert(err, IsNil)

	// update test
	natRule, err = edge.AddDNATRule(NatRule{NetworkHref: orgVdcNetwork.OrgVDCNetwork.HREF, ExternalIP: vcd.config.VCD.ExternalIp,
		ExternalPort: "1177", InternalIP: vcd.config.VCD.InternalIp, InternalPort: "77", Protocol: "TCP", Description: description1})
	check.Assert(err, IsNil)

	natRule.GatewayNatRule.OriginalPort = "1166"
	natRule.GatewayNatRule.TranslatedPort = "66"
	natRule.GatewayNatRule.Protocol = "udp"
	natRule.Description = description2
	natRule.GatewayNatRule.Interface.HREF = externalNetwork.ExternalNetwork.HREF

	updateNatRule, err := edge.UpdateNatRule(natRule)

	check.Assert(err, IsNil)
	check.Assert(updateNatRule.GatewayNatRule.TranslatedIP, Equals, vcd.config.VCD.InternalIp)
	check.Assert(updateNatRule.GatewayNatRule.TranslatedPort, Equals, "66")
	check.Assert(updateNatRule.GatewayNatRule.OriginalIP, Equals, vcd.config.VCD.ExternalIp)
	check.Assert(updateNatRule.GatewayNatRule.OriginalPort, Equals, "1166")
	check.Assert(updateNatRule.GatewayNatRule.Protocol, Equals, "udp")
	check.Assert(updateNatRule.GatewayNatRule.IcmpSubType, Equals, "")
	check.Assert(updateNatRule.Description, Equals, description2)
	check.Assert(updateNatRule.RuleType, Equals, "DNAT")
	check.Assert(strings.Split(updateNatRule.GatewayNatRule.Interface.HREF, "network/")[1], Equals, strings.Split(externalNetwork.ExternalNetwork.HREF, "externalnet/")[1])

	err = edge.RemoveNATRule(updateNatRule.ID)
	check.Assert(err, IsNil)

	// verify delete
	err = edge.Refresh()
	check.Assert(err, IsNil)

	check.Assert(len(edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration.NatService.NatRule), Equals, beforeChangeNatRulesNumber)
}

// TestEdgeGateway_UpdateLBGeneralParams main point is to test that no load balancer configuration
// xml tags are lost during changes of load balancer main settings (enable, logging)
// The test does following steps:
// 1. Cache raw XML body and marshaled struct in variables before running the test
// 2. Toggle the settings of load balancer in various ways and ensure no err is returned
// 3. Set the settings back as they originally were and again get raw XML body and marshaled struct
// 4. Compare the XML text and structs before configuration and after configuration - they should be
// identical except <version></version> tag which is versioning the configuration
func (vcd *TestVCD) TestEdgeGateway_UpdateLBGeneralParams(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	// Cache current load balancer settings for change validation in the end
	beforeLb, beforeLbXml := testCacheLoadBalancer(*edge, check)

	_, err = edge.UpdateLBGeneralParams(true, true, true, "critical")
	check.Assert(err, IsNil)

	_, err = edge.UpdateLBGeneralParams(false, false, false, "emergency")
	check.Assert(err, IsNil)

	// Try to set invalid loglevel to get validation error
	_, err = edge.UpdateLBGeneralParams(false, false, false, "invalid_loglevel")
	check.Assert(err, ErrorMatches, ".*Valid log levels are.*")

	// Restore to initial settings and validate that it
	_, err = edge.UpdateLBGeneralParams(beforeLb.Enabled, beforeLb.AccelerationEnabled,
		beforeLb.Logging.Enable, beforeLb.Logging.LogLevel)
	check.Assert(err, IsNil)

	// Validate load balancer configuration against initially cached version
	testCheckLoadBalancerConfig(beforeLb, beforeLbXml, *edge, check)
}

// TestEdgeGateway_UpdateFwGeneralParams main point is to test that no firewall configuration
// xml tags are lost during changes of firewall main settings (enable, logging)
// The test does following steps:
// 1. Cache raw XML body and marshaled struct in variables before running the test
// 2. Toggle the settings of firewall in various ways and ensure no err is returned
// 3. Set the settings back as they originally were and again get raw XML body and marshaled struct
// 4. Compare the XML text and structs before configuration and after configuration - they should be
// identical except <version></version> tag which is versioning the configuration
func (vcd *TestVCD) TestEdgeGateway_UpdateFwGeneralParams(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	// Cache current firewall settings for change validation in the end
	beforeFw, beforeFwXml := testCacheFirewall(*edge, check)

	_, err = edge.UpdateFirewallConfig(false, false, "deny")
	check.Assert(err, IsNil)

	_, err = edge.UpdateFirewallConfig(true, true, "accept")
	check.Assert(err, IsNil)

	// Try to set invalid loglevel to get validation error
	_, err = edge.UpdateFirewallConfig(false, false, "invalid_action")
	check.Assert(err, ErrorMatches, ".*default action must be either 'accept' or 'deny'.*")

	// Restore to initial settings and validate that it
	_, err = edge.UpdateFirewallConfig(beforeFw.Enabled, beforeFw.DefaultPolicy.LoggingEnabled, beforeFw.DefaultPolicy.Action)
	check.Assert(err, IsNil)

	// Validate configuration against initially cached version
	testCheckFirewallConfig(beforeFw, beforeFwXml, *edge, check)
}

func (vcd *TestVCD) TestEdgeGateway_GetVdcNetworks(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	vnics, err := edge.getVdcNetworks()
	check.Assert(err, IsNil)

	foundExtNet := false
	foundOrgNet := false

	check.Assert(len(vnics.EdgeInterface) > 1, Equals, true)
	// Look for both - external and Org networks in returned edge gateway vNics
	for _, vnic := range vnics.EdgeInterface {
		// Look for external network attached to Edge gateway
		if vnic.Name == vcd.config.VCD.ExternalNetwork && vnic.PortgroupName == vcd.config.VCD.ExternalNetwork {
			check.Assert(vnic.AddressGroups.AddressGroup.PrimaryAddress, Equals, vcd.config.VCD.ExternalIp)
			check.Assert(vnic.Type, Equals, "uplink")
			foundExtNet = true
		}

		// Look for org network 1 attached
		if vnic.PortgroupName == vcd.config.VCD.Network.Net1 {
			check.Assert(vnic.Type, Equals, "internal")
			foundOrgNet = true
		}
	}
	check.Assert(foundExtNet, Equals, true)
	check.Assert(foundOrgNet, Equals, true)
}

// testCacheFirewall is meant to store firewall settings before any operations so that all
// configuration can be checked after manipulation
func testCacheFirewall(edge EdgeGateway, check *C) (*types.FirewallConfigWithXml, string) {
	beforeFw, err := edge.GetFirewallConfig()
	check.Assert(err, IsNil)
	beforeFwbXml := testGetEdgeEndpointXML(types.EdgeFirewallPath, edge, check)
	return beforeFw, beforeFwbXml
}

// testCheckFirewallConfig validates if both raw XML string and firewall struct remain
// identical after settings manipulation.
func testCheckFirewallConfig(beforeFw *types.FirewallConfigWithXml, beforeFwXml string, edge EdgeGateway, check *C) {
	afterFw, err := edge.GetFirewallConfig()
	check.Assert(err, IsNil)

	afterFwXml := testGetEdgeEndpointXML(types.EdgeFirewallPath, edge, check)

	// remove `<version></version>` tag from both XML represntation and struct for deep comparison
	// because this version changes with each update and will never be the same after a few
	// operations
	reVersion := regexp.MustCompile(`<version>\w*<\/version>`)
	beforeFwXml = reVersion.ReplaceAllLiteralString(beforeFwXml, "")
	afterFwXml = reVersion.ReplaceAllLiteralString(afterFwXml, "")

	beforeFw.Version = ""
	afterFw.Version = ""

	// Because the test enables and disables firewall configuration, main firewall rule has its ID
	// and ruleTag changed during the test and we must ignore this change while comparing. This is
	// always the first rule kept at the top so we are replacing `id` and `ruletag` fields in
	// "before" and "after" values.
	beforeFw.FirewallRules.Text = replaceFirstMatch(beforeFw.FirewallRules.Text, `<id>\d*</id>`, "<id>99999</id>")
	beforeFw.FirewallRules.Text = replaceFirstMatch(beforeFw.FirewallRules.Text, `<ruleTag>\d*</ruleTag>`, "<ruleTag>99999</ruleTag>")
	afterFw.FirewallRules.Text = replaceFirstMatch(afterFw.FirewallRules.Text, `<id>\d*</id>`, "<id>99999</id>")
	afterFw.FirewallRules.Text = replaceFirstMatch(afterFw.FirewallRules.Text, `<ruleTag>\d*</ruleTag>`, "<ruleTag>99999</ruleTag>")

	beforeFwXml = replaceFirstMatch(beforeFwXml, `<id>\d*</id>`, "<id>99999</id>")
	beforeFwXml = replaceFirstMatch(beforeFwXml, `<ruleTag>\d*</ruleTag>`, "<ruleTag>99999</ruleTag>")
	afterFwXml = replaceFirstMatch(afterFwXml, `<id>\d*</id>`, "<id>99999</id>")
	afterFwXml = replaceFirstMatch(afterFwXml, `<ruleTag>\d*</ruleTag>`, "<ruleTag>99999</ruleTag>")

	check.Assert(beforeFw, DeepEquals, afterFw)
	check.Assert(beforeFwXml, DeepEquals, afterFwXml)
}

// replaceFirstMatch replaces first `regex` matched in `text` with `replacement` and returns it
// It will panic if regex is invalid.
func replaceFirstMatch(text, regex, replacement string) string {
	re := regexp.MustCompile(regex)
	// Replace leftmost found string
	found := re.FindString(text)
	if found != "" {
		return strings.Replace(text, found, replacement, 1)
	}
	return ""
}

// TestListEdgeGateway tests that at least one edge gateway is found,
// and the list contains the element defined in the configuration file
func (vcd *TestVCD) TestListEdgeGateway(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge, NotNil)
	edgeGatewayList, err := vcd.vdc.QueryEdgeGatewayList()
	check.Assert(err, IsNil)
	check.Assert(edgeGatewayList, NotNil)
	check.Assert(len(edgeGatewayList) > 0, Equals, true)
	foundName := false
	foundHref := false
	for _, ref := range edgeGatewayList {
		if ref.Name == edge.EdgeGateway.Name {
			foundName = true
			if ref.HREF == edge.EdgeGateway.HREF {
				foundHref = true
			}
		}
	}
	check.Assert(foundName, Equals, true)
	check.Assert(foundHref, Equals, true)
}

func (vcd *TestVCD) Test_UpdateEdgeGateway(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)
	var saveEGW = types.EdgeGateway{
		Name:        edge.EdgeGateway.Name,
		ID:          edge.EdgeGateway.ID,
		Status:      edge.EdgeGateway.Status,
		HREF:        edge.EdgeGateway.HREF,
		Description: edge.EdgeGateway.Description,
		Configuration: &types.GatewayConfiguration{
			AdvancedNetworkingEnabled: edge.EdgeGateway.Configuration.AdvancedNetworkingEnabled,
		},
	}

	newName := "UpdatedEGWName"
	newDescription := "Updated description"

	edge.EdgeGateway.Name = newName
	edge.EdgeGateway.Description = newDescription

	err = edge.Update()
	check.Assert(err, IsNil)

	// The edge gateway should be updated in place
	check.Assert(edge.EdgeGateway.HREF, Equals, saveEGW.HREF)
	check.Assert(edge.EdgeGateway.Name, Equals, newName)
	check.Assert(edge.EdgeGateway.Description, Equals, newDescription)

	// Check that a new copy of the edge gateway contains the expected data
	edge, err = vcd.vdc.GetEdgeGatewayById(saveEGW.ID, true)
	check.Assert(err, IsNil)

	check.Assert(edge.EdgeGateway.HREF, Equals, saveEGW.HREF)
	check.Assert(edge.EdgeGateway.Name, Equals, newName)
	check.Assert(edge.EdgeGateway.Description, Equals, newDescription)

	edge.EdgeGateway.Name = saveEGW.Name
	edge.EdgeGateway.Description = saveEGW.Description

	err = edge.Update()
	check.Assert(err, IsNil)

	// checking the in-place values
	check.Assert(saveEGW.Name, Equals, edge.EdgeGateway.Name)
	check.Assert(saveEGW.Description, Equals, edge.EdgeGateway.Description)
	check.Assert(saveEGW.HREF, Equals, edge.EdgeGateway.HREF)

	// Checking values in a fresh copy of the edge gateway
	edge, err = vcd.vdc.GetEdgeGatewayById(saveEGW.ID, true)
	check.Assert(err, IsNil)

	check.Assert(saveEGW.Name, Equals, edge.EdgeGateway.Name)
	check.Assert(saveEGW.Description, Equals, edge.EdgeGateway.Description)
	check.Assert(saveEGW.HREF, Equals, edge.EdgeGateway.HREF)
}
