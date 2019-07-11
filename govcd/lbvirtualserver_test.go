// +build lb lbVirtualServer functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LBVirtualServer tests CRUD methods for load balancer virtual server.
// It uses helper function buildTestLBVirtualServerPrereqs to create prerequisite components:
// service monitor, server pool, application profile and application rule.
// The following things are tested if prerequisite Edge Gateway exists:
// Creation of load balancer virtual server
// Read load balancer virtual server by both ID and Name (virtual server name must be unique in single edge gateway)
// Update - change a single field and compare that configuration and result objects are deeply equal
// Update - try and fail to update without mandatory field
// Delete
func (vcd *TestVCD) Test_LBVirtualServer(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	if vcd.config.VCD.ExternalIp == "" {
		check.Skip("Skipping test because no edge gateway external IP given")
	}

	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	_, serverPoolId, appProfileId := buildTestLBVirtualServerPrereqs(check, vcd, edge)

	// Configure creation object including reference to service monitor
	lbVirtualServerConfig := &types.LBVirtualServer{
		Name:                 TestLBVirtualServer,
		IpAddress:            vcd.config.VCD.ExternalIp, // Load balancer virtual server serves on Edge gw IP
		Enabled:              true,
		AccelerationEnabled:  true,
		Protocol:             "http",
		Port:                 8888,
		ConnectionLimit:      5,
		ConnectionRateLimit:  10,
		ApplicationProfileId: appProfileId,
		DefaultPoolId:        serverPoolId,
	}

	createdLbVirtualServer, err := edge.CreateLBVirtualServer(lbVirtualServerConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbVirtualServer.ID, Not(IsNil))
	check.Assert(createdLbVirtualServer.IpAddress, Equals, lbVirtualServerConfig.IpAddress)
	check.Assert(createdLbVirtualServer.Protocol, Equals, lbVirtualServerConfig.Protocol)
	check.Assert(createdLbVirtualServer.Port, Equals, lbVirtualServerConfig.Port)
	check.Assert(createdLbVirtualServer.ConnectionLimit, Equals, lbVirtualServerConfig.ConnectionLimit)
	check.Assert(createdLbVirtualServer.ConnectionRateLimit, Equals, lbVirtualServerConfig.ConnectionRateLimit)
	check.Assert(createdLbVirtualServer.Enabled, Equals, lbVirtualServerConfig.Enabled)
	check.Assert(createdLbVirtualServer.AccelerationEnabled, Equals, lbVirtualServerConfig.AccelerationEnabled)
	// check.Assert(createdLbVirtualServer.ApplicationRuleId, Equals, lbVirtualServerConfig.ApplicationRuleId)
	check.Assert(createdLbVirtualServer.DefaultPoolId, Equals, lbVirtualServerConfig.DefaultPoolId)

	// We created virtual server successfully therefore let's prepend it to cleanup list so that it
	// is deleted before the child components
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	PrependToCleanupList(TestLBVirtualServer, "lbVirtualServer", parentEntity, check.TestName())

	// Lookup by both name and ID and compare that these are equal values
	lbVirtualServerByID, err := edge.ReadLBVirtualServer(&types.LBVirtualServer{ID: createdLbVirtualServer.ID})
	check.Assert(err, IsNil)

	lbVirtualServerByName, err := edge.ReadLBVirtualServer(&types.LBVirtualServer{Name: createdLbVirtualServer.Name})
	check.Assert(err, IsNil)
	check.Assert(createdLbVirtualServer.ID, Equals, lbVirtualServerByName.ID)
	check.Assert(lbVirtualServerByID.ID, Equals, lbVirtualServerByName.ID)
	check.Assert(lbVirtualServerByID.Name, Equals, lbVirtualServerByName.Name)

	// Test updating fields
	// Update algorithm
	lbVirtualServerByID.Port = 8889
	updatedLBPool, err := edge.UpdateLBVirtualServer(lbVirtualServerByID)
	check.Assert(err, IsNil)
	check.Assert(updatedLBPool.Port, Equals, lbVirtualServerByID.Port)

	// Verify that updated pool and it's configuration are identical
	check.Assert(updatedLBPool, DeepEquals, lbVirtualServerByID)

	// Try to set invalid protocol and expect API to return error:
	// vShield Edge [LoadBalancer] Invalid protocol invalid_protocol. Valid protocols are: HTTP|HTTPS|TCP|UDP. (API error: 14542)
	lbVirtualServerByID.Protocol = "invalid_protocol"
	updatedLBPool, err = edge.UpdateLBVirtualServer(lbVirtualServerByID)
	check.Assert(err, ErrorMatches, ".*Invalid protocol.*Valid protocols are:.*")

	// Update should fail without name
	lbVirtualServerByID.Name = ""
	_, err = edge.UpdateLBVirtualServer(lbVirtualServerByID)
	check.Assert(err.Error(), Equals, "load balancer virtual server Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLBVirtualServer(&types.LBVirtualServer{ID: createdLbVirtualServer.ID})
	check.Assert(err, IsNil)

	// Ensure it is deleted
	_, err = edge.ReadLBVirtualServerByID(createdLbVirtualServer.ID)
	check.Assert(IsNotFound(err), Equals, true)
}

// TODO create application rule once it is merged
func buildTestLBVirtualServerPrereqs(check *C, vcd *TestVCD, edge EdgeGateway) (serviceMonitorId, serverPoolId, appProfileId string) {
	// Establish prerequisite - service monitor
	lbMon := &types.LBMonitor{
		Name:       TestLBVirtualServer,
		Interval:   10,
		Timeout:    10,
		MaxRetries: 3,
		Type:       "http",
	}
	lbMonitor, err := edge.CreateLBServiceMonitor(lbMon)
	check.Assert(err, IsNil)

	lbPoolConfig := &types.LBPool{
		Name:      TestLBVirtualServer,
		Algorithm: "round-robin",
		MonitorId: lbMonitor.ID,
		Members: types.LBPoolMembers{
			types.LBPoolMember{
				Name:      "Server_one",
				IpAddress: "1.1.1.1",
				Port:      8443,
				Weight:    1,
				Condition: "enabled",
			},
			types.LBPoolMember{
				Name:      "Server_two",
				IpAddress: "2.2.2.2",
				Port:      8443,
				Weight:    2,
				Condition: "enabled",
			},
		},
	}

	lbPool, err := edge.CreateLBServerPool(lbPoolConfig)
	check.Assert(err, IsNil)
	// Used for creating
	lbAppProfileConfig := &types.LBAppProfile{
		Name: TestLBVirtualServer,
		Persistence: &types.LBAppProfilePersistence{
			Method: "sourceip",
			Expire: 13,
		},
		Template: "HTTP",
	}

	lbAppProfile, err := edge.CreateLBAppProfile(lbAppProfileConfig)
	check.Assert(err, IsNil)
	// ToDo Add application rule once it is in master

	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(TestLBVirtualServer, "lbServerPool", parentEntity, check.TestName())
	AddToCleanupList(TestLBVirtualServer, "lbServiceMonitor", parentEntity, check.TestName())
	AddToCleanupList(TestLBVirtualServer, "lbAppProfile", parentEntity, check.TestName())

	return lbMonitor.ID, lbPool.ID, lbAppProfile.ID
}
