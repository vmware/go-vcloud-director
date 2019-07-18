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
// With help of function buildTestLBVirtualServerPrereqs such prerequisite components are created:
// service monitor, server pool, application profile and application rule.
// The following things are tested if prerequisites are met:
// 1. Creation of load balancer virtual server
// 2. Read load balancer virtual server by both Id and Name (virtual server name must be unique in
// single edge gateway)
// 3. Update - change a single field and compare that configuration and result objects are deeply
// equal
// 4. Update - try and fail to update without mandatory field
// 5. Delete
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

	_, serverPoolId, appProfileId, appRuleId := buildTestLBVirtualServerPrereqs("1.1.1.1", "2.2.2.2",
		TestLBVirtualServer, check, vcd, edge)

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
		ApplicationRuleId:    appRuleId,
		DefaultPoolId:        serverPoolId,
	}

	createdLbVirtualServer, err := edge.CreateLBVirtualServer(lbVirtualServerConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbVirtualServer.Id, Not(IsNil))
	check.Assert(createdLbVirtualServer.IpAddress, Equals, lbVirtualServerConfig.IpAddress)
	check.Assert(createdLbVirtualServer.Protocol, Equals, lbVirtualServerConfig.Protocol)
	check.Assert(createdLbVirtualServer.Port, Equals, lbVirtualServerConfig.Port)
	check.Assert(createdLbVirtualServer.ConnectionLimit, Equals, lbVirtualServerConfig.ConnectionLimit)
	check.Assert(createdLbVirtualServer.ConnectionRateLimit, Equals, lbVirtualServerConfig.ConnectionRateLimit)
	check.Assert(createdLbVirtualServer.Enabled, Equals, lbVirtualServerConfig.Enabled)
	check.Assert(createdLbVirtualServer.AccelerationEnabled, Equals, lbVirtualServerConfig.AccelerationEnabled)
	check.Assert(createdLbVirtualServer.ApplicationRuleId, Equals, lbVirtualServerConfig.ApplicationRuleId)
	check.Assert(createdLbVirtualServer.DefaultPoolId, Equals, lbVirtualServerConfig.DefaultPoolId)

	// We created virtual server successfully therefore let's prepend it to cleanup list so that it
	// is deleted before the child components
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	PrependToCleanupList(TestLBVirtualServer, "lbVirtualServer", parentEntity, check.TestName())

	// Lookup by both name and Id and compare that these are equal values
	lbVirtualServerById, err := edge.ReadLBVirtualServer(&types.LBVirtualServer{Id: createdLbVirtualServer.Id})
	check.Assert(err, IsNil)

	lbVirtualServerByName, err := edge.ReadLBVirtualServer(&types.LBVirtualServer{Name: createdLbVirtualServer.Name})
	check.Assert(err, IsNil)
	check.Assert(createdLbVirtualServer.Id, Equals, lbVirtualServerByName.Id)
	check.Assert(lbVirtualServerById.Id, Equals, lbVirtualServerByName.Id)
	check.Assert(lbVirtualServerById.Name, Equals, lbVirtualServerByName.Name)

	// Test updating fields
	// Update algorithm
	lbVirtualServerById.Port = 8889
	updatedLBPool, err := edge.UpdateLBVirtualServer(lbVirtualServerById)
	check.Assert(err, IsNil)
	check.Assert(updatedLBPool.Port, Equals, lbVirtualServerById.Port)

	// Verify that updated pool and its configuration are identical
	check.Assert(updatedLBPool, DeepEquals, lbVirtualServerById)

	// Try to set invalid protocol and expect API to return error:
	// vShield Edge [LoadBalancer] Invalid protocol invalid_protocol. Valid protocols are: HTTP|HTTPS|TCP|UDP. (API error: 14542)
	lbVirtualServerById.Protocol = "invalid_protocol"
	updatedLBPool, err = edge.UpdateLBVirtualServer(lbVirtualServerById)
	check.Assert(err, ErrorMatches, ".*Invalid protocol.*Valid protocols are:.*")

	// Update should fail without name
	lbVirtualServerById.Name = ""
	_, err = edge.UpdateLBVirtualServer(lbVirtualServerById)
	check.Assert(err.Error(), Equals, "load balancer virtual server Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLBVirtualServer(&types.LBVirtualServer{Id: createdLbVirtualServer.Id})
	check.Assert(err, IsNil)

	// Ensure it is deleted
	_, err = edge.ReadLBVirtualServerById(createdLbVirtualServer.Id)
	check.Assert(IsNotFound(err), Equals, true)
}

// buildTestLBVirtualServerPrereqs creates all load balancer components which are consumed by
// load balanver virtual server and ads them to cleanup in correct order to avoid deletion of used
// resources
func buildTestLBVirtualServerPrereqs(node1Ip, node2Ip, componentsName string, check *C, vcd *TestVCD, edge EdgeGateway) (serviceMonitorId, serverPoolId, appProfileId, appRuleId string) {
	// Create prerequisites - service monitor
	lbMon := &types.LBMonitor{
		Name:       componentsName,
		Interval:   10,
		Timeout:    10,
		MaxRetries: 3,
		Type:       "http",
	}
	lbMonitor, err := edge.CreateLBServiceMonitor(lbMon)
	check.Assert(err, IsNil)

	// Create prerequisites - server pool
	lbPoolConfig := &types.LBPool{
		Name:      componentsName,
		Algorithm: "round-robin",
		MonitorId: lbMonitor.ID,
		Members: types.LBPoolMembers{
			types.LBPoolMember{
				Name:      "Server_one",
				IpAddress: node1Ip,
				Port:      8000,
				Weight:    1,
				Condition: "enabled",
			},
			types.LBPoolMember{
				Name:      "Server_two",
				IpAddress: node2Ip,
				Port:      8000,
				Weight:    1,
				Condition: "enabled",
			},
		},
	}

	lbPool, err := edge.CreateLBServerPool(lbPoolConfig)
	check.Assert(err, IsNil)

	// Create prerequisites - application profile
	lbAppProfileConfig := &types.LBAppProfile{
		Name:     componentsName,
		Template: "HTTP",
	}

	lbAppProfile, err := edge.CreateLBAppProfile(lbAppProfileConfig)
	check.Assert(err, IsNil)

	lbAppRuleConfig := &types.LBAppRule{
		Name:   componentsName,
		Script: "acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page",
	}

	// Create prerequisites - application rule
	lbAppRule, err := edge.CreateLBAppRule(lbAppRuleConfig)
	check.Assert(err, IsNil)

	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(componentsName, "lbServerPool", parentEntity, check.TestName())
	AddToCleanupList(componentsName, "lbServiceMonitor", parentEntity, check.TestName())
	AddToCleanupList(componentsName, "lbAppProfile", parentEntity, check.TestName())
	AddToCleanupList(componentsName, "lbAppRule", parentEntity, check.TestName())

	// return lbMonitor.ID, lbPool.ID, lbAppProfile.ID, lbAppRule.ID
	return lbMonitor.ID, lbPool.ID, lbAppProfile.ID, lbAppRule.ID
}
