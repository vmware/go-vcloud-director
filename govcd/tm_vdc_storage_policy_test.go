//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
)

func (vcd *TestVCD) Test_TmVdcStoragePolicy(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
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

	sp, err := region.GetStoragePolicyByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sp, NotNil)

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

	check.Assert(err, IsNil)
	vdcPolicies, err := createdVdc.CreateStoragePolicies(&types.VirtualDatacenterStoragePolicies{
		Values: []types.VirtualDatacenterStoragePolicy{
			{
				RegionStoragePolicy: types.OpenApiReference{
					ID: sp.RegionStoragePolicy.ID,
				},
				StorageLimitMiB: 100,
				VirtualDatacenter: types.OpenApiReference{
					ID: createdVdc.TmVdc.ID,
				},
			},
		},
	})
	check.Assert(err, IsNil)
	// TODO: TM: Does not work
	//defer func() {
	//	err = vdcPolicies[0].Delete()
	//	check.Assert(err, IsNil)
	//}()
	check.Assert(len(vdcPolicies), Equals, 1)
	check.Assert(vdcPolicies[0].VirtualDatacenterStoragePolicy, NotNil)

	// Getting policies by VDC (the parent)
	allPolicies, err := createdVdc.GetAllStoragePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allPolicies), Equals, len(vdcPolicies))
	check.Assert(allPolicies[0].VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(*allPolicies[0].VirtualDatacenterStoragePolicy, DeepEquals, *vdcPolicies[0].VirtualDatacenterStoragePolicy)

	// TODO: TM: Does not work
	/*	policy, err := createdVdc.GetStoragePolicyById(vdcPolicies[0].VirtualDatacenterStoragePolicy.ID)
		check.Assert(err, IsNil)
		check.Assert(policy.VirtualDatacenterStoragePolicy, NotNil)
		check.Assert(*policy.VirtualDatacenterStoragePolicy, DeepEquals, *vdcPolicies[0].VirtualDatacenterStoragePolicy)*/

	// Getting policies with general client
	params := url.Values{}
	params.Add("filter", "virtualDatacenter.id=="+createdVdc.TmVdc.ID)
	filteredPolicies, err := vcd.client.GetAllTmVdcStoragePolicies(params)
	check.Assert(err, IsNil)
	check.Assert(len(filteredPolicies), Equals, len(vdcPolicies))
	check.Assert(filteredPolicies[0].VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(*filteredPolicies[0].VirtualDatacenterStoragePolicy, DeepEquals, *vdcPolicies[0].VirtualDatacenterStoragePolicy)

	// TODO: TM: Does not work
	/*
		// Update policy
		updatedPolicy, err := vdcPolicies[0].Update(&types.VirtualDatacenterStoragePolicy{
			RegionStoragePolicy: vdcPolicies[0].VirtualDatacenterStoragePolicy.RegionStoragePolicy,
			StorageLimitMiB:     200,
			VirtualDatacenter:   vdcPolicies[0].VirtualDatacenterStoragePolicy.VirtualDatacenter,
		})
		check.Assert(err, IsNil)
		check.Assert(updatedPolicy, NotNil)
		check.Assert(updatedPolicy.VirtualDatacenterStoragePolicy.StorageLimitMiB, Equals, int64(200))


			policy, err = vcd.client.GetTmVdcStoragePolicyById(vdcPolicies[0].VirtualDatacenterStoragePolicy.ID)
				check.Assert(err, IsNil)
				check.Assert(policy.VirtualDatacenterStoragePolicy, NotNil)
				check.Assert(*policy.VirtualDatacenterStoragePolicy, DeepEquals, *vdcPolicies[0].VirtualDatacenterStoragePolicy)

			// Not Found tests
			byNameInvalid, err := vcd.client.GetTmVdcStoragePolicyById("urn:vcloud:virtualDatacenterStoragePolicy:5344b964-0000-0000-0000-d554913db643")
			check.Assert(ContainsNotFound(err), Equals, true)
			check.Assert(byNameInvalid, IsNil)

			byIdInvalid, err := createdVdc.GetStoragePolicyById("urn:vcloud:virtualDatacenter:5344b964-0000-0000-0000-d554913db643")
			check.Assert(ContainsNotFound(err), Equals, true)
			check.Assert(byIdInvalid, IsNil)*/

}
