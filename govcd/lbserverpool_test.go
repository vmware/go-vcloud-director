// +build lb lbServerPool nsxv functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Test_LBServerPool tests CRUD methods for load balancer server pool.
// The following things are tested if prerequisite Edge Gateway exists:
// 1. Creation of load balancer server pool
// 2. Get load balancer server pool by both ID and Name (server pool name must be unique in single edge gateway)
// 3. Update - change a single field and compare that configuration and result objects are deeply equal
// 4. Update - try and fail to update without mandatory field
// 5. Delete
func (vcd *TestVCD) Test_LBServerPool(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	// Establish prerequisite - service monitor
	lbMon := &types.LbMonitor{
		Name:       check.TestName(),
		Interval:   10,
		Timeout:    10,
		MaxRetries: 3,
		Type:       "http",
	}
	err = deleteLbServiceMonitorIfExists(ctx, *edge, lbMon.Name)
	check.Assert(err, IsNil)
	lbMonitor, err := edge.CreateLbServiceMonitor(ctx, lbMon)
	check.Assert(err, IsNil)
	check.Assert(lbMonitor.ID, NotNil)

	// Add service monitor to cleanup
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(check.TestName(), "lbServiceMonitor", parentEntity, check.TestName())

	// Configure creation object including reference to service monitor
	lbPoolConfig := &types.LbPool{
		Name:        TestLbServerPool,
		Transparent: false,
		Algorithm:   "round-robin",
		MonitorId:   lbMonitor.ID,
		Members: types.LbPoolMembers{
			types.LbPoolMember{
				Name:      "Server_one",
				IpAddress: "1.1.1.1",
				Port:      8443,
				Weight:    1,
				Condition: "enabled",
			},
			types.LbPoolMember{
				Name:      "Server_two",
				IpAddress: "2.2.2.2",
				Port:      8443,
				Weight:    2,
				Condition: "enabled",
			},
		},
	}

	err = deleteLbServerPoolIfExists(ctx, *edge, lbMon.Name)
	check.Assert(err, IsNil)
	createdLbPool, err := edge.CreateLbServerPool(ctx, lbPoolConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbPool.ID, Not(IsNil))
	check.Assert(createdLbPool.Transparent, Equals, lbPoolConfig.Transparent)
	check.Assert(createdLbPool.MonitorId, Equals, lbMonitor.ID)
	check.Assert(len(createdLbPool.Members), Equals, 2)
	check.Assert(createdLbPool.Members[0].Condition, Equals, "enabled")
	check.Assert(createdLbPool.Members[1].Condition, Equals, "enabled")
	check.Assert(createdLbPool.Members[0].Weight, Equals, 1)
	check.Assert(createdLbPool.Members[1].Weight, Equals, 2)
	check.Assert(createdLbPool.Members[0].Name, Equals, "Server_one")
	check.Assert(createdLbPool.Members[1].Name, Equals, "Server_two")

	// We created server pool successfully therefore let's add it to cleanup list
	AddToCleanupList(TestLbServerPool, "lbServerPool", parentEntity, check.TestName())

	// Try to delete used service monitor and expect it to fail with nice error
	err = edge.DeleteLbServiceMonitor(ctx, lbMon)
	check.Assert(err, ErrorMatches, `.*Fail to delete objectId .*\S+.* for it is used by .*`)

	// Lookup by both name and ID and compare that these are equal values
	lbPoolByID, err := edge.getLbServerPool(ctx, &types.LbPool{ID: createdLbPool.ID})
	check.Assert(err, IsNil)
	check.Assert(lbPoolByID, Not(IsNil))

	lbPoolByName, err := edge.getLbServerPool(ctx, &types.LbPool{Name: createdLbPool.Name})
	check.Assert(err, IsNil)
	check.Assert(lbPoolByName, Not(IsNil))
	check.Assert(createdLbPool.ID, Equals, lbPoolByName.ID)
	check.Assert(lbPoolByID.ID, Equals, lbPoolByName.ID)
	check.Assert(lbPoolByID.Name, Equals, lbPoolByName.Name)

	check.Assert(createdLbPool.Algorithm, Equals, lbPoolConfig.Algorithm)

	// GetLbServerPools should return at least one pool which is ours.
	pools, err := edge.GetLbServerPools(ctx)
	check.Assert(err, IsNil)
	check.Assert(pools, Not(HasLen), 0)

	// Test updating fields
	// Update algorithm
	lbPoolByID.Algorithm = "ip-hash"
	updatedLBPool, err := edge.UpdateLbServerPool(ctx, lbPoolByID)
	check.Assert(err, IsNil)
	check.Assert(updatedLBPool.Algorithm, Equals, lbPoolByID.Algorithm)

	// Update boolean value fields
	lbPoolByID.Transparent = true
	updatedLBPool, err = edge.UpdateLbServerPool(ctx, lbPoolByID)
	check.Assert(err, IsNil)
	check.Assert(updatedLBPool.Transparent, Equals, lbPoolByID.Transparent)

	// Verify that updated pool and its configuration are identical
	check.Assert(updatedLBPool, DeepEquals, lbPoolByID)

	// Try to set invalid algorithm hash and expect API to return error
	// Invalid algorithm hash. Valid algorithms are: IP-HASH|ROUND-ROBIN|URI|LEASTCONN|URL|HTTP-HEADER.
	lbPoolByID.Algorithm = "invalid_algorithm"
	updatedLBPool, err = edge.UpdateLbServerPool(ctx, lbPoolByID)
	check.Assert(updatedLBPool, IsNil)
	check.Assert(err, ErrorMatches, ".*Invalid algorithm.*Valid algorithms are:.*")

	// Update should fail without name
	lbPoolByID.Name = ""
	_, err = edge.UpdateLbServerPool(ctx, lbPoolByID)
	check.Assert(err.Error(), Equals, "load balancer server pool Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLbServerPool(ctx, &types.LbPool{ID: createdLbPool.ID})
	check.Assert(err, IsNil)

	_, err = edge.GetLbServerPoolById(ctx, createdLbPool.ID)
	check.Assert(IsNotFound(err), Equals, true)
}
