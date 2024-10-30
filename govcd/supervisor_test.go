//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmSupervisor(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vc, vcCreated, nsxtManager, nsxtManagerCreated := getOrCreateVcAndNsxtManager(vcd, check)

	allSupervisors, err := vcd.client.GetAllSupervisors(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allSupervisors) > 0, Equals, true)

	supervisorById, err := vcd.client.GetSupervisorById(allSupervisors[0].Supervisor.SupervisorID)
	check.Assert(err, IsNil)
	check.Assert(supervisorById.Supervisor, DeepEquals, allSupervisors[0].Supervisor)

	supervisorByName, err := vcd.client.GetSupervisorById(allSupervisors[0].Supervisor.SupervisorID)
	check.Assert(err, IsNil)
	check.Assert(supervisorByName.Supervisor, DeepEquals, allSupervisors[0].Supervisor)

	// vCenter Supervisors
	allVcSupervisors, err := vc.GetAllSupervisors(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allVcSupervisors) > 0, Equals, true)

	vcSupervisorById, err := vc.GetSupervisorByName(allSupervisors[0].Supervisor.Name)
	check.Assert(err, IsNil)
	check.Assert(vcSupervisorById.Supervisor, DeepEquals, allSupervisors[0].Supervisor)

	// Checking Supervisor Zones
	sZones, err := vcSupervisorById.GetAllSupervisorZones(nil)
	check.Assert(err, IsNil)
	check.Assert(len(sZones) > 0, Equals, true)

	zoneById, err := vcSupervisorById.GetSupervisorZoneById(sZones[0].SupervisorZone.ID)
	check.Assert(err, IsNil)
	check.Assert(zoneById, NotNil)
	check.Assert(zoneById.SupervisorZone, DeepEquals, sZones[0].SupervisorZone)

	zoneByName, err := vcSupervisorById.GetSupervisorZoneByName(sZones[0].SupervisorZone.Name)
	check.Assert(err, IsNil)
	check.Assert(zoneByName, NotNil)

	check.Assert(zoneById.SupervisorZone, DeepEquals, zoneByName.SupervisorZone)

	// Cleanup
	if vcCreated {
		err = vc.Disable()
		check.Assert(err, IsNil)
		err = vc.Delete()
		check.Assert(err, IsNil)
	}

	if nsxtManagerCreated {
		err = nsxtManager.Delete()
		check.Assert(err, IsNil)
	}
}
