// +build lb lbVirtualServer nsxv functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LBVirtualServer tests CRUD methods for load balancer virtual server.
// With help of function buildTestLBVirtualServerPrereqs such prerequisite components are created:
// service monitor, server pool, application profile and application rule.
// The following things are tested if prerequisites are met:
// 1. Creation of load balancer virtual server
// 2. Get load balancer virtual server by both ID and Name (virtual server name must be unique in
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

	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	serviceMonitorId, serverPoolId, appProfileId, appRuleId := buildTestLBVirtualServerPrereqs("1.1.1.1", "2.2.2.2",
		TestLbVirtualServer, check, vcd, *edge)

	// Configure creation object including reference to service monitor
	lbVirtualServerConfig := &types.LbVirtualServer{
		Name:                 TestLbVirtualServer,
		IpAddress:            vcd.config.VCD.ExternalIp, // Load balancer virtual server serves on Edge gw IP
		Enabled:              false,
		AccelerationEnabled:  false,
		Protocol:             "http",
		Port:                 8888,
		ConnectionLimit:      5,
		ConnectionRateLimit:  10,
		ApplicationProfileId: appProfileId,
		ApplicationRuleIds:   []string{appRuleId},
		DefaultPoolId:        serverPoolId,
	}

	err = deleteLbVirtualServerIfExists(*edge, lbVirtualServerConfig.Name)
	check.Assert(err, IsNil)
	createdLbVirtualServer, err := edge.CreateLbVirtualServer(ctx, lbVirtualServerConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbVirtualServer.ID, Not(IsNil))
	check.Assert(createdLbVirtualServer.IpAddress, Equals, lbVirtualServerConfig.IpAddress)
	check.Assert(createdLbVirtualServer.Protocol, Equals, lbVirtualServerConfig.Protocol)
	check.Assert(createdLbVirtualServer.Port, Equals, lbVirtualServerConfig.Port)
	check.Assert(createdLbVirtualServer.ConnectionLimit, Equals, lbVirtualServerConfig.ConnectionLimit)
	check.Assert(createdLbVirtualServer.ConnectionRateLimit, Equals, lbVirtualServerConfig.ConnectionRateLimit)
	check.Assert(createdLbVirtualServer.Enabled, Equals, lbVirtualServerConfig.Enabled)
	check.Assert(createdLbVirtualServer.AccelerationEnabled, Equals, lbVirtualServerConfig.AccelerationEnabled)
	check.Assert(createdLbVirtualServer.ApplicationRuleIds, DeepEquals, lbVirtualServerConfig.ApplicationRuleIds)
	check.Assert(createdLbVirtualServer.DefaultPoolId, Equals, lbVirtualServerConfig.DefaultPoolId)

	// Try to delete child components and expect a well parsed NSX error
	err = edge.DeleteLbServiceMonitorById(ctx, serviceMonitorId)
	check.Assert(err, ErrorMatches, `.*Fail to delete objectId .*\S+.* for it is used by .*`)
	err = edge.DeleteLbServerPoolById(ctx, serverPoolId)
	check.Assert(err, ErrorMatches, `.*Fail to delete objectId .*\S+.* for it is used by .*`)
	err = edge.DeleteLbAppProfileById(ctx, appProfileId)
	check.Assert(err, ErrorMatches, `.*Fail to delete objectId .*\S+.* for it is used by .*`)
	err = edge.DeleteLbAppRuleById(ctx, appRuleId)
	check.Assert(err, ErrorMatches, `.*Fail to delete objectId .*\S+.* for it is used by .*`)

	// We created virtual server successfully therefore let's prepend it to cleanup list so that it
	// is deleted before the child components
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	PrependToCleanupList(TestLbVirtualServer, "lbVirtualServer", parentEntity, check.TestName())

	// Lookup by both name and ID and compare that these are equal values
	lbVirtualServerById, err := edge.getLbVirtualServer(ctx, &types.LbVirtualServer{ID: createdLbVirtualServer.ID})
	check.Assert(err, IsNil)
	check.Assert(lbVirtualServerById, Not(IsNil))

	lbVirtualServerByName, err := edge.getLbVirtualServer(ctx, &types.LbVirtualServer{Name: createdLbVirtualServer.Name})
	check.Assert(err, IsNil)
	check.Assert(lbVirtualServerByName, Not(IsNil))
	check.Assert(createdLbVirtualServer.ID, Equals, lbVirtualServerByName.ID)
	check.Assert(lbVirtualServerById.ID, Equals, lbVirtualServerByName.ID)
	check.Assert(lbVirtualServerById.Name, Equals, lbVirtualServerByName.Name)

	// GetLbVirtualServers should return at least one vs which is ours.
	servers, err := edge.GetLbVirtualServers(ctx)
	check.Assert(err, IsNil)
	check.Assert(servers, Not(HasLen), 0)

	// Test updating fields
	// Update algorithm
	lbVirtualServerById.Port = 8889
	updatedLBPool, err := edge.UpdateLbVirtualServer(ctx, lbVirtualServerById)
	check.Assert(err, IsNil)
	check.Assert(updatedLBPool.Port, Equals, lbVirtualServerById.Port)

	// Update boolean value fields
	lbVirtualServerById.Enabled = true
	lbVirtualServerById.AccelerationEnabled = true
	updatedLBPool, err = edge.UpdateLbVirtualServer(ctx, lbVirtualServerById)
	check.Assert(err, IsNil)
	check.Assert(updatedLBPool.Enabled, Equals, lbVirtualServerById.Enabled)
	check.Assert(updatedLBPool.AccelerationEnabled, Equals, lbVirtualServerById.AccelerationEnabled)

	// Verify that updated pool and its configuration are identical
	check.Assert(updatedLBPool, DeepEquals, lbVirtualServerById)

	// Try to set invalid protocol and expect API to return error:
	// vShield Edge [LoadBalancer] Invalid protocol invalid_protocol. Valid protocols are: HTTP|HTTPS|TCP|UDP. (API error: 14542)
	lbVirtualServerById.Protocol = "invalid_protocol"
	updatedLBPool, err = edge.UpdateLbVirtualServer(ctx, lbVirtualServerById)
	check.Assert(updatedLBPool, IsNil)
	check.Assert(err, ErrorMatches, ".*Invalid protocol.*Valid protocols are:.*")

	// Update should fail without name
	lbVirtualServerById.Name = ""
	_, err = edge.UpdateLbVirtualServer(ctx, lbVirtualServerById)
	check.Assert(err.Error(), Equals, "load balancer virtual server Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLbVirtualServer(ctx, &types.LbVirtualServer{ID: createdLbVirtualServer.ID})
	check.Assert(err, IsNil)

	// Ensure it is deleted
	_, err = edge.GetLbVirtualServerById(ctx, createdLbVirtualServer.ID)
	check.Assert(IsNotFound(err), Equals, true)
}

// buildTestLBVirtualServerPrereqs creates all load balancer components which are consumed by
// load balanver virtual server and ads them to cleanup in correct order to avoid deletion of used
// resources
func buildTestLBVirtualServerPrereqs(node1Ip, node2Ip, componentsName string, check *C, vcd *TestVCD, edge EdgeGateway) (serviceMonitorId, serverPoolId, appProfileId, appRuleId string) {
	// Create prerequisites - service monitor
	lbMon := &types.LbMonitor{
		Name:       componentsName,
		Interval:   10,
		Timeout:    10,
		MaxRetries: 3,
		Type:       "http",
	}
	err := deleteLbServiceMonitorIfExists(ctx, edge, lbMon.Name)
	check.Assert(err, IsNil)
	lbMonitor, err := edge.CreateLbServiceMonitor(ctx, lbMon)
	check.Assert(err, IsNil)

	// Create prerequisites - server pool
	lbPoolConfig := &types.LbPool{
		Name:      componentsName,
		Algorithm: "round-robin",
		MonitorId: lbMonitor.ID,
		Members: types.LbPoolMembers{
			types.LbPoolMember{
				Name:      "Server_one",
				IpAddress: node1Ip,
				Port:      8000,
				Weight:    1,
				Condition: "enabled",
			},
			types.LbPoolMember{
				Name:      "Server_two",
				IpAddress: node2Ip,
				Port:      8000,
				Weight:    1,
				Condition: "enabled",
			},
		},
	}

	err = deleteLbServerPoolIfExists(ctx, edge, lbPoolConfig.Name)
	check.Assert(err, IsNil)
	lbPool, err := edge.CreateLbServerPool(ctx, lbPoolConfig)
	check.Assert(err, IsNil)

	// Create prerequisites - application profile
	lbAppProfileConfig := &types.LbAppProfile{
		Name:     componentsName,
		Template: "HTTP",
	}

	err = deleteLbAppProfileIfExists(ctx, edge, lbAppProfileConfig.Name)
	check.Assert(err, IsNil)
	lbAppProfile, err := edge.CreateLbAppProfile(ctx, lbAppProfileConfig)
	check.Assert(err, IsNil)

	lbAppRuleConfig := &types.LbAppRule{
		Name:   componentsName,
		Script: "acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page",
	}

	// Create prerequisites - application rule
	err = deleteLbAppRuleIfExists(ctx, edge, lbAppRuleConfig.Name)
	check.Assert(err, IsNil)
	lbAppRule, err := edge.CreateLbAppRule(ctx, lbAppRuleConfig)
	check.Assert(err, IsNil)

	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(lbAppRule.Name, "lbAppRule", parentEntity, check.TestName())
	AddToCleanupList(lbAppProfile.Name, "lbAppProfile", parentEntity, check.TestName())
	AddToCleanupList(lbPool.Name, "lbServerPool", parentEntity, check.TestName())
	AddToCleanupList(lbMon.Name, "lbServiceMonitor", parentEntity, check.TestName())

	return lbMonitor.ID, lbPool.ID, lbAppProfile.ID, lbAppRule.ID
}

// deleteLbVirtualServerIfExists is used to cleanup before creation of component. It returns error only if there was
// other error than govcd.ErrorEntityNotFound
func deleteLbVirtualServerIfExists(edge EdgeGateway, name string) error {
	err := edge.DeleteLbVirtualServerByName(ctx, name)
	if err != nil && !ContainsNotFound(err) {
		return err
	}
	if err != nil && ContainsNotFound(err) {
		return nil
	}

	fmt.Printf("# Removed leftover LB virtual server '%s'\n", name)
	return nil
}
