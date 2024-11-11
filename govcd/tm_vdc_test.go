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

	createdVdc, err := vcd.client.CreateTmVdc(vdcType)
	check.Assert(err, IsNil)
	check.Assert(createdVdc, NotNil)
	// Add to cleanup list
	PrependToCleanupListOpenApi(createdVdc.TmVdc.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTmVdcs+createdVdc.TmVdc.ID)
	defer func() {
		err = createdVdc.Delete()
		check.Assert(err, IsNil)
	}()

	// Get TM VDC By Name
	byName, err := vcd.client.GetTmVdcByName(vdcType.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Get TM VDC By Id
	byId, err := vcd.client.GetTmVdcById(createdVdc.TmVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Update
	createdVdc.TmVdc.Name = check.TestName() + "-update"
	updatedVdc, err := createdVdc.Update(createdVdc.TmVdc)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc.TmVdc, DeepEquals, createdVdc.TmVdc)
}
