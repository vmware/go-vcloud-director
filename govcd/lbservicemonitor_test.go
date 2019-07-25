// +build lb lbServiceMonitor functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_LBServiceMonitor tests CRUD methods for load balancer service monitor.
// The following things are tested if prerequisite Edge Gateway exists:
// Creation of load balancer service monitor
// Read load balancer by both Id and Name (service monitor name must be unique in single edge gateway)
// Update - change a single field and compare that configuration and result objects are deeply equal
// Update - try and fail to update without mandatory field
// Delete
func (vcd *TestVCD) Test_LBServiceMonitor(check *C) {
	if vcd.config.VCD.EdgeGateway == "" {
		check.Skip("Skipping test because no edge gateway given")
	}
	edge, err := vcd.vdc.FindEdgeGateway(vcd.config.VCD.EdgeGateway)
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

	lbMonitor, err := edge.CreateLbServiceMonitor(lbMon)
	check.Assert(err, IsNil)
	check.Assert(lbMonitor.Id, Not(IsNil))

	// We created monitor successfully therefore let's add it to cleanup list
	parentEntity := vcd.org.Org.Name + "|" + vcd.vdc.Vdc.Name + "|" + vcd.config.VCD.EdgeGateway
	AddToCleanupList(check.TestName(), "lbServiceMonitor", parentEntity, check.TestName())

	// Lookup by both name and Id and compare that these are equal values
	lbMonitorByID, err := edge.GetLbServiceMonitor(&types.LbMonitor{Id: lbMonitor.Id})
	check.Assert(err, IsNil)

	lbMonitorByName, err := edge.GetLbServiceMonitor(&types.LbMonitor{Name: lbMonitor.Name})
	check.Assert(err, IsNil)
	check.Assert(lbMonitor.Id, Equals, lbMonitorByName.Id)
	check.Assert(lbMonitorByID.Id, Equals, lbMonitorByName.Id)
	check.Assert(lbMonitorByID.Name, Equals, lbMonitorByName.Name)

	check.Assert(lbMonitor.Id, Equals, lbMonitorByID.Id)
	check.Assert(lbMonitor.Interval, Equals, lbMonitorByID.Interval)
	check.Assert(lbMonitor.Timeout, Equals, lbMonitorByID.Timeout)
	check.Assert(lbMonitor.MaxRetries, Equals, lbMonitorByID.MaxRetries)

	// Test updating fields
	// Update timeout
	lbMonitorByID.Timeout = 35
	updatedLBMonitor, err := edge.UpdateLbServiceMonitor(lbMonitorByID)
	check.Assert(err, IsNil)
	check.Assert(updatedLBMonitor.Timeout, Equals, 35)

	// Verify that updated monitor and it's configuration are identical
	check.Assert(updatedLBMonitor, DeepEquals, lbMonitorByID)

	// Update should fail without name
	lbMonitorByID.Name = ""
	_, err = edge.UpdateLbServiceMonitor(lbMonitorByID)
	check.Assert(err.Error(), Equals, "load balancer monitor Name cannot be empty")

	// Delete / cleanup
	err = edge.DeleteLbServiceMonitor(&types.LbMonitor{Id: lbMonitorByID.Id})
	check.Assert(err, IsNil)

	_, err = edge.GetLbServiceMonitorById(lbMonitorByID.Id)
	check.Assert(IsNotFound(err), Equals, true)
}
