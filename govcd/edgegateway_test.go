/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_Refresh(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
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
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	task, err := edge.AddNATMapping("DNAT", vcd.config.VCD.ExternalIp, vcd.config.VCD.InternalIp)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = edge.RemoveNATMapping("DNAT", vcd.config.VCD.ExternalIp, vcd.config.VCD.InternalIp, "77")
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
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)
	task, err := edge.AddNATPortMapping("DNAT", vcd.config.VCD.ExternalIp, "1177", vcd.config.VCD.InternalIp, "77", "TCP", "")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	task, err = edge.RemoveNATPortMapping("DNAT", vcd.config.VCD.ExternalIp, "1177", vcd.config.VCD.InternalIp, "77")
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
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
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
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
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
		Xmlns: "http://www.vmware.com/vcloud/v1.5",
		GatewayIpsecVpnService: &types.GatewayIpsecVpnService{
			IsEnabled: true,
			Tunnel:    tunnels,
		},
	}

	// Configures VPN service
	task, err := edge.AddIpsecVPN(ipsecVPNConfig)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// To check the effects of service configuration, we need to reload the edge gateway entity
	err = edge.Refresh()
	check.Assert(err, IsNil)

	// We expect an enabled service, and non-null tunnel and endpoint
	newConf := edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration
	newConfState := newConf.GatewayIpsecVpnService.IsEnabled
	newConfTunnel := newConf.GatewayIpsecVpnService.Tunnel
	newConfEndpoint := newConf.GatewayIpsecVpnService.Endpoint
	check.Assert(newConfState, Equals, true)
	check.Assert(newConfTunnel, NotNil)
	check.Assert(newConfEndpoint, NotNil)

	// Removes VPN service
	task, err = edge.RemoveIpsecVPN()
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	// To check the effects of service configuration, we need to reload the edge gateway entity
	err = edge.Refresh()
	check.Assert(err, IsNil)

	// We expect a disabled service, and null tunnel and endpoint
	afterDeletionConf := edge.EdgeGateway.Configuration.EdgeGatewayServiceConfiguration
	newConfState = afterDeletionConf.GatewayIpsecVpnService.IsEnabled
	newConfTunnel = afterDeletionConf.GatewayIpsecVpnService.Tunnel
	newConfEndpoint = afterDeletionConf.GatewayIpsecVpnService.Endpoint
	check.Assert(newConfState, Equals, false)
	check.Assert(newConfTunnel, IsNil)
	check.Assert(newConfEndpoint, IsNil)
}
