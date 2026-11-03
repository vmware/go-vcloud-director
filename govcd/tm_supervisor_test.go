//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmSupervisor(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

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

	singleVcSupervisor, err := vcd.client.GetSupervisorByNameAndVcenterId(allVcSupervisors[0].Supervisor.Name, allVcSupervisors[0].Supervisor.VirtualCenter.ID)
	check.Assert(err, IsNil)
	check.Assert(singleVcSupervisor.Supervisor, DeepEquals, allVcSupervisors[0].Supervisor)

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
}

func (vcd *TestVCD) Test_TmComputeClusters(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	_, cleanup := getOrCreateVCenter(vcd, check)
	defer cleanup()
	clusters, err := vcd.client.GetAllTmComputeClusters(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	check.Assert(len(clusters) > 0, Equals, true)
	seen := map[string]bool{}
	for _, cluster := range clusters {
		check.Assert(cluster.ID, Not(Equals), "")
		check.Assert(seen[cluster.ID], Equals, false)
		seen[cluster.ID] = true
		byID, err := vcd.client.GetTmComputeClusterById(cluster.ID)
		check.Assert(err, IsNil)
		check.Assert(byID, DeepEquals, cluster)
		check.Assert(byID.VirtualCenter, NotNil)
		check.Assert(byID.Moref, Not(Equals), "")
	}
}
