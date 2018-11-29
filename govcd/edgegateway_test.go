/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	types "github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_Refresh(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edgegatway given")
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
		check.Skip("Skipping test because no edgegatway given")
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
		check.Skip("Skipping test because no edgegatway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)
	task, err := edge.AddNATPortMapping("DNAT", vcd.config.VCD.ExternalIp, "1177", vcd.config.VCD.InternalIp, "77", "TCP", "any")
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

// TODO: Add a check checking whether the IPsec VPN was added
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
	tunnel := &types.GatewayIpsecVpnTunnel{
		Name:        "Test VPN",
		Description: "Testing VPN Creation",
		IpsecVpnLocalPeer: &types.IpsecVpnLocalPeer{
			ID:   "",
			Name: "",
		},
		EncryptionProtocol: "AES",
		LocalIPAddress:     vcd.config.VCD.ExternalIp,
		LocalID:            vcd.config.VCD.ExternalIp,
		IsEnabled:          true,
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
	_, err = edge.AddIpsecVPN(ipsecVPNConfig)
	check.Assert(err, IsNil)
}
