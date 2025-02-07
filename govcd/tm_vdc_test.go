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

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()

	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	// Required information: Zones and classes
	regionZones, err := region.GetAllZones(nil)
	check.Assert(err, IsNil)
	check.Assert(len(regionZones) > 0, Equals, true)

	vmClasses, err := region.GetAllVmClasses(nil)
	check.Assert(err, IsNil)
	check.Assert(len(vmClasses) > 0, Equals, true)

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
	PrependToCleanupListOpenApi(createdVdc.TmVdc.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmVdcs+createdVdc.TmVdc.ID)
	defer func() {
		err = createdVdc.Delete()
		check.Assert(err, IsNil)
	}()

	// Assignment of VM Classes
	err = createdVdc.AssignVmClasses(&types.RegionVirtualMachineClasses{
		Values: types.OpenApiReferences{{Name: vmClasses[0].Name, ID: vmClasses[0].ID}},
	})

	// Get TM VDC By Name
	byName, err := vcd.client.GetTmVdcByName(vdcType.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Get TM VDC By Id
	byId, err := vcd.client.GetTmVdcById(createdVdc.TmVdc.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Get By Name and Org ID
	byNameAndOrgId, err := vcd.client.GetTmVdcByNameAndOrgId(createdVdc.TmVdc.Name, org.TmOrg.ID)
	check.Assert(err, IsNil)
	check.Assert(byNameAndOrgId.TmVdc, DeepEquals, createdVdc.TmVdc)

	// Get By Name and Org ID in non existent Org
	byNameAndInvalidOrgId, err := vcd.client.GetTmVdcByNameAndOrgId(createdVdc.TmVdc.Name, "urn:vcloud:org:a93c9db9-0000-0000-0000-a8f7eeda85f9")
	check.Assert(err, NotNil)
	check.Assert(byNameAndInvalidOrgId, IsNil)

	// Not Found tests
	byNameInvalid, err := vcd.client.GetTmVdcByName("fake-name")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byNameInvalid, IsNil)

	byIdInvalid, err := vcd.client.GetTmVdcById("urn:vcloud:virtualDatacenter:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdInvalid, IsNil)

	// Update
	createdVdc.TmVdc.Name = check.TestName() + "-update"
	updatedVdc, err := createdVdc.Update(createdVdc.TmVdc)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc.TmVdc, DeepEquals, createdVdc.TmVdc)
}
