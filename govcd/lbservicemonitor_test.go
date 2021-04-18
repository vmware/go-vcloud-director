// +build lb lbServiceMonitor nsxv functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Test_LBServiceMonitor tests CRUD methods for load balancer service monitor.
// The following things are tested if prerequisite Edge Gateway exists:
// 1. Creation of load balancer service monitor
// 2. Get load balancer by both ID and Name (service monitor name must be unique in single edge gateway)
// 3. Update - change a single field and compare that configuration and result objects are deeply equal
// 4. Update - try and fail to update without mandatory field
// 5. Delete
func (vcd *TestVCD) Test_LBServiceMonitor(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	check.Assert(edge.EdgeGateway.Name, Equals, vcd.config.VCD.EdgeGateway)

	if !edge.HasAdvancedNetworking() {
		check.Skip("Skipping test because the edge gateway does not have advanced networking enabled")
	}

	// Used for creating
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
	check.Assert(lbMonitor.ID, Not(IsNil))

	// We created monitor successfully therefore let's add it to cleanup list
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(check.TestName(), "lbServiceMonitor", parentEntity, check.TestName())

	// Lookup by both name and ID and compare that these are equal values
	lbMonitorByID, err := edge.getLbServiceMonitor(ctx, &types.LbMonitor{ID: lbMonitor.ID})
	check.Assert(err, IsNil)
	check.Assert(lbMonitorByID, Not(IsNil))

	lbMonitorByName, err := edge.getLbServiceMonitor(ctx, &types.LbMonitor{Name: lbMonitor.Name})
	check.Assert(err, IsNil)
	check.Assert(lbMonitorByName, Not(IsNil))
	check.Assert(lbMonitor.ID, Equals, lbMonitorByName.ID)
	check.Assert(lbMonitorByID.ID, Equals, lbMonitorByName.ID)
	check.Assert(lbMonitorByID.Name, Equals, lbMonitorByName.Name)

	check.Assert(lbMonitor.ID, Equals, lbMonitorByID.ID)
	check.Assert(lbMonitor.Interval, Equals, lbMonitorByID.Interval)
	check.Assert(lbMonitor.Timeout, Equals, lbMonitorByID.Timeout)
	check.Assert(lbMonitor.MaxRetries, Equals, lbMonitorByID.MaxRetries)

	// GetLbServiceMonitors should return at least one vs which is ours.
	lbMonitors, err := edge.GetLbServiceMonitors(ctx)
	check.Assert(err, IsNil)
	check.Assert(lbMonitors, Not(HasLen), 0)

	// Test updating fields
	// Update timeout
	lbMonitorByID.Timeout = 35
	updatedLBMonitor, err := edge.UpdateLbServiceMonitor(ctx, lbMonitorByID)
	check.Assert(err, IsNil)
	check.Assert(updatedLBMonitor.Timeout, Equals, 35)

	// Verify that updated monitor and its configuration are identical
	check.Assert(updatedLBMonitor, DeepEquals, lbMonitorByID)

	// Update should fail without name
	lbMonitorByID.Name = ""
	_, err = edge.UpdateLbServiceMonitor(ctx, lbMonitorByID)
	check.Assert(err.Error(), Equals, "load balancer monitor Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLbServiceMonitor(ctx, &types.LbMonitor{ID: lbMonitorByID.ID})
	check.Assert(err, IsNil)

	_, err = edge.GetLbServiceMonitorById(ctx, lbMonitorByID.ID)
	check.Assert(IsNotFound(err), Equals, true)
}
