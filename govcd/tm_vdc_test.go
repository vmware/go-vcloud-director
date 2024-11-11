//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmVdc(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	regionZones, err := region.GetAllZones(nil)
	check.Assert(err, IsNil)
	check.Assert(len(regionZones) > 0, Equals, true)

	vdcType := &types.TmVdc{
		Name:        check.TestName(),
		Org:         &types.OpenApiReference{ID: org.TmOrg.ID},
		Region:      &types.OpenApiReference{ID: region.Region.ID},
		Supervisors: []types.OpenApiReference{{ID: supervisor.Supervisor.SupervisorID}},
		ZoneResourceAllocation: []*types.TmVdcZoneResourceAllocation{{
			Zone: &types.OpenApiReference{ID: regionZones[0].Zone.ID},
			ResourceAllocation: types.TmVdcResourceAllocation{
				CPUReservationMHz:    100,
				CPULimitMHz:          500,
				MemoryReservationMiB: 256,
				MemoryLimitMiB:       512,
			},
		}},
	}

	vdc, err := vcd.client.CreateTmVdc(vdcType)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	// Cleanup
	err = vdc.Delete()
	check.Assert(err, IsNil)

	// vdc, err := vcd.client.GetTmVdcByName(vcd.config.Tm.Vdc)
	// check.Assert(err, IsNil)
	// check.Assert(vdc, NotNil)

	// // Get by ID
	// vdcById, err := vcd.client.GetTmVdcById(vdc.TmVdc.ID)
	// check.Assert(err, IsNil)
	// check.Assert(vdcById, NotNil)

	// check.Assert(vdc.TmVdc, DeepEquals, vdcById.TmVdc)

	// allTmVdc, err := vcd.client.GetAllTmVdcs(nil)
	// check.Assert(err, IsNil)
	// check.Assert(len(allTmVdc) > 0, Equals, true)
}
