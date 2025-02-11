//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"strings"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegionQuotaStoragePolicy(check *C) {
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

	createdRegionQuota, err := vcd.client.CreateRegionQuota(vdcType)
	check.Assert(err, IsNil)
	check.Assert(createdRegionQuota, NotNil)
	// Add to cleanup list
	PrependToCleanupListOpenApi(createdRegionQuota.TmVdc.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmVdcs+createdRegionQuota.TmVdc.ID)
	defer func() {
		err = createdRegionQuota.Delete()
		check.Assert(err, IsNil)
	}()

	check.Assert(err, IsNil)
	rqPolicies, err := createdRegionQuota.CreateStoragePolicies(&types.VirtualDatacenterStoragePolicies{
		Values: []types.VirtualDatacenterStoragePolicy{
			{
				RegionStoragePolicy: types.OpenApiReference{
					ID: sp.RegionStoragePolicy.ID,
				},
				StorageLimitMiB: 100,
				VirtualDatacenter: types.OpenApiReference{
					ID: createdRegionQuota.TmVdc.ID,
				},
			},
		},
	})
	check.Assert(err, IsNil)
	check.Assert(len(rqPolicies), Equals, 1)
	check.Assert(rqPolicies[0].VirtualDatacenterStoragePolicy, NotNil)

	// Getting policies
	allPolicies, err := createdRegionQuota.GetAllStoragePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allPolicies), Equals, len(rqPolicies))
	check.Assert(allPolicies[0].VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(*allPolicies[0].VirtualDatacenterStoragePolicy, DeepEquals, *rqPolicies[0].VirtualDatacenterStoragePolicy)

	policy, err := createdRegionQuota.GetStoragePolicyById(rqPolicies[0].VirtualDatacenterStoragePolicy.ID)
	check.Assert(err, IsNil)
	check.Assert(policy.VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(*policy.VirtualDatacenterStoragePolicy, DeepEquals, *rqPolicies[0].VirtualDatacenterStoragePolicy)

	policy, err = createdRegionQuota.GetStoragePolicyByName(rqPolicies[0].VirtualDatacenterStoragePolicy.Name)
	check.Assert(err, IsNil)
	check.Assert(policy.VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(*policy.VirtualDatacenterStoragePolicy, DeepEquals, *rqPolicies[0].VirtualDatacenterStoragePolicy)

	// Update policy
	updatedPolicy, err := rqPolicies[0].Update(&types.VirtualDatacenterStoragePolicy{
		RegionStoragePolicy: rqPolicies[0].VirtualDatacenterStoragePolicy.RegionStoragePolicy,
		StorageLimitMiB:     200,
		VirtualDatacenter:   rqPolicies[0].VirtualDatacenterStoragePolicy.VirtualDatacenter,
	})
	check.Assert(err, IsNil)
	check.Assert(updatedPolicy, NotNil)
	check.Assert(updatedPolicy.VirtualDatacenterStoragePolicy.StorageLimitMiB, Equals, int64(200))

	policy, err = vcd.client.GetRegionQuotaStoragePolicyById(rqPolicies[0].VirtualDatacenterStoragePolicy.ID)
	check.Assert(err, IsNil)
	check.Assert(policy.VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(*policy.VirtualDatacenterStoragePolicy, DeepEquals, *updatedPolicy.VirtualDatacenterStoragePolicy)

	// Delete the policy
	err = policy.Delete()
	if len(rqPolicies) == 1 {
		// If there's only one policy, expect an error. There must be always one Storage Policy
		check.Assert(err, NotNil)
		check.Assert(strings.Contains(err.Error(), "Storage policy is not set for the entity in VDC"), Equals, true)
	} else {
		check.Assert(err, IsNil)
	}

	// Not Found tests
	byNameInvalid, err := vcd.client.GetRegionQuotaStoragePolicyById("urn:vcloud:virtualDatacenterStoragePolicy:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byNameInvalid, IsNil)

	byIdInvalid, err := createdRegionQuota.GetStoragePolicyById("urn:vcloud:virtualDatacenterStoragePolicy:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdInvalid, IsNil)

}
