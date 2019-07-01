// +build lb lbServerPool functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LBServerPool tests CRUD methods for load balancer server pool.
// The following things are tested if prerequisite Edge Gateway exists:
// Creation of load balancer server pool
// Read load balancer server pool by both ID and Name (server pool name must be unique in single edge gateway)
// Update - change a single field and compare that configuration and result objects are deeply equal
// Update - try and fail to update without mandatory field
// Delete
func (vcd *TestVCD) Test_LBServerPool(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	// Establish prerequisite - service monitor
	lbMon := &types.LBMonitor{
		Name:       check.TestName(),
		Interval:   10,
		Timeout:    10,
		MaxRetries: 3,
		Type:       "http",
	}
	lbMonitor, err := edge.CreateLBServiceMonitor(lbMon)
	check.Assert(err, IsNil)
	check.Assert(lbMonitor.ID, NotNil)

	// Add service monitor to cleanup
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(check.TestName(), "lbServiceMonitor", parentEntity, check.TestName())

	// Configure creation object including reference to service monitor
	lbPoolConfig := &types.LBPool{
		Name:      TestLBServerPool,
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

	createdLbPool, err := edge.CreateLBServerPool(lbPoolConfig)
	check.Assert(err, IsNil)
	check.Assert(createdLbPool.ID, Not(IsNil))
	check.Assert(createdLbPool.MonitorId, Equals, lbMonitor.ID)
	check.Assert(len(createdLbPool.Members), Equals, 2)
	check.Assert(createdLbPool.Members[0].Condition, Equals, "enabled")
	check.Assert(createdLbPool.Members[1].Condition, Equals, "enabled")
	check.Assert(createdLbPool.Members[0].Weight, Equals, 1)
	check.Assert(createdLbPool.Members[1].Weight, Equals, 2)
	check.Assert(createdLbPool.Members[0].Name, Equals, "Server_one")
	check.Assert(createdLbPool.Members[1].Name, Equals, "Server_two")

	// We created server pool successfully therefore let's add it to cleanup list
	AddToCleanupList(TestLBServerPool, "lbServerPool", parentEntity, check.TestName())

	// Lookup by both name and ID and compare that these are equal values
	lbPoolByID, err := edge.ReadLBServerPool(&types.LBPool{ID: createdLbPool.ID})
	check.Assert(err, IsNil)

	lbPoolByName, err := edge.ReadLBServerPool(&types.LBPool{Name: createdLbPool.Name})
	check.Assert(err, IsNil)
	check.Assert(createdLbPool.ID, Equals, lbPoolByName.ID)
	check.Assert(lbPoolByID.ID, Equals, lbPoolByName.ID)
	check.Assert(lbPoolByID.Name, Equals, lbPoolByName.Name)

	check.Assert(createdLbPool.Algorithm, Equals, lbPoolConfig.Algorithm)

	// Test updating fields
	// Update algorithm
	lbPoolByID.Algorithm = "ip-hash"
	updatedLBPool, err := edge.UpdateLBServerPool(lbPoolByID)
	check.Assert(err, IsNil)
	check.Assert(updatedLBPool.Algorithm, Equals, lbPoolByID.Algorithm)

	// Verify that updated pool and it's configuration are identical
	check.Assert(updatedLBPool, DeepEquals, lbPoolByID)

	// Try to set invalid algorithm hash and expect API to return error
	// Invalid algorithm hash. Valid algorithms are: IP-HASH|ROUND-ROBIN|URI|LEASTCONN|URL|HTTP-HEADER.
	lbPoolByID.Algorithm = "invalid_algorithm"
	updatedLBPool, err = edge.UpdateLBServerPool(lbPoolByID)
	check.Assert(err, ErrorMatches, ".*Invalid algorithm.*Valid algorithms are:.*")

	// Update should fail without name
	lbPoolByID.Name = ""
	_, err = edge.UpdateLBServerPool(lbPoolByID)
	check.Assert(err.Error(), Equals, "load balancer server pool Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLBServerPool(&types.LBPool{ID: createdLbPool.ID})
	check.Assert(err, IsNil)
}
