// +build nsxv functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxvDhcpRelay tests out Edge gateway DHCP relay configuration settings. It does the following:
// 1. Creates an IP set to ensure it can be used in DHCP relay configuration settings.
// 2. Creates a DHCP relay configuration with multiple objects and checks. Adds a task for cleanup.
// Note. Because DHCP relay configuration is not an object, but a setting of edge gateway - there is
// nothing to delete
// 3. Resets the DHCP relay configuration and checks that it was removed
// 4. Creates another DHCP relay with different configuration (using only IP sets) and manually
// specifying IP address on edge gateway interface (vNic)
// 6. Resets the DHCP relay configuration and checks that it was removed
func (vcd *TestVCD) Test_NsxvDhcpRelay(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}

	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	// Setup IP set for testing purposes and add it to cleanup list
	createdIpSet, err := testCreateIpSet("dhcp-relay-test", vcd.vdc)
	check.Assert(err, IsNil)
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name
	AddToCleanupList(createdIpSet.Name, "ipSet", parentEntity, check.TestName())

	// Lookup vNic index for our org network
	vNicIndex, _, err := edge.GetAnyVnicIndexByNetworkName(ctx, vcd.config.VCD.Network.Net1)
	check.Assert(err, IsNil)

	dhcpRelayConfig := &types.EdgeDhcpRelay{
		RelayServer: &types.EdgeDhcpRelayServer{
			IpAddress:        []string{"1.1.1.1"},
			Fqdns:            []string{"servergroups.domainname.com", "servergroups.otherdomainname.com"},
			GroupingObjectId: []string{createdIpSet.ID},
		},
		RelayAgents: &types.EdgeDhcpRelayAgents{
			Agents: []types.EdgeDhcpRelayAgent{
				types.EdgeDhcpRelayAgent{
					VnicIndex: vNicIndex,
				},
			},
		},
	}

	createdRelayConfig, err := edge.UpdateDhcpRelay(ctx, dhcpRelayConfig)
	check.Assert(err, IsNil)

	parentEntity = vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	// DHCP relay config is being prepended (added to beginning) of cleanup list because it must be
	// delete first so that above created dependent IP sets can be cleaned up when their turn comes
	PrependToCleanupList(check.TestName(), "dhcpRelayConfig", parentEntity, check.TestName())

	// Cache Gateway auto-assigned address to try specifying it afterwards
	vNicDefaultGateway := createdRelayConfig.RelayAgents.Agents[0].GatewayInterfaceAddress

	readRelayConfig, err := edge.GetDhcpRelay(ctx)
	check.Assert(err, IsNil)
	check.Assert(readRelayConfig, DeepEquals, createdRelayConfig)

	// Patch auto-assigned IP so that structure comparison works
	dhcpRelayConfig.RelayAgents.Agents[0].GatewayInterfaceAddress = readRelayConfig.RelayAgents.Agents[0].GatewayInterfaceAddress
	check.Assert(readRelayConfig.RelayServer, DeepEquals, dhcpRelayConfig.RelayServer)
	check.Assert(readRelayConfig.RelayAgents, DeepEquals, dhcpRelayConfig.RelayAgents)

	// Reset DHCP relay and ensure no settings are present
	err = edge.ResetDhcpRelay(ctx)
	check.Assert(err, IsNil)
	read, err := edge.GetDhcpRelay(ctx)
	check.Assert(err, IsNil)
	check.Assert(read.RelayServer, IsNil)
	check.Assert(read.RelayAgents, IsNil)

	// Second attempt - insert gateway interface IP address during creation to prevent
	// auto-assignment
	dhcpRelayConfig.RelayAgents.Agents[0].GatewayInterfaceAddress = vNicDefaultGateway
	// Remove IP addresses and FQDNs to only use IP sets for specifying DHCP servers addresses
	dhcpRelayConfig.RelayServer.IpAddress = nil
	dhcpRelayConfig.RelayServer.Fqdns = nil

	createdRelayConfig2, err := edge.UpdateDhcpRelay(ctx, dhcpRelayConfig)
	check.Assert(err, IsNil)
	readRelayConfig2, err := edge.GetDhcpRelay(ctx)
	check.Assert(err, IsNil)
	check.Assert(readRelayConfig2, DeepEquals, createdRelayConfig2)

	// Reset DHCP relay and ensure no settings are present
	err = edge.ResetDhcpRelay(ctx)
	check.Assert(err, IsNil)
	read, err = edge.GetDhcpRelay(ctx)
	check.Assert(err, IsNil)
	check.Assert(read.RelayServer, IsNil)
	check.Assert(read.RelayAgents, IsNil)
}
