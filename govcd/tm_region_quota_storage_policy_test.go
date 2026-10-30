//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

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
				CPUReservationMHz:    addrOf(int64(100)),
				CPULimitMHz:          addrOf(int64(500)),
				MemoryReservationMiB: addrOf(int64(256)),
				MemoryLimitMiB:       addrOf(int64(512)),
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

	// Rejected batches must return an error rather than an inventory listing.
	invalidPolicies, err := createdRegionQuota.CreateStoragePolicies(&types.VirtualDatacenterStoragePolicies{})
	check.Assert(err, NotNil)
	check.Assert(invalidPolicies, IsNil)

	rqPolicies, err := createdRegionQuota.CreateStoragePolicies(&types.VirtualDatacenterStoragePolicies{
		Values: []types.VirtualDatacenterStoragePolicy{
			{
				RegionStoragePolicy: types.OpenApiReference{
					ID: sp.RegionStoragePolicy.ID,
				},
				StorageLimitMiB: addrOf(int64(100)),
				VirtualDatacenter: types.OpenApiReference{
					ID: createdRegionQuota.TmVdc.ID,
				},
			},
		},
	})
	check.Assert(err, IsNil)
	check.Assert(len(rqPolicies), Equals, 1)
	check.Assert(rqPolicies[0].VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(rqPolicies[0].VirtualDatacenterStoragePolicy.VirtualDatacenter.ID, Equals, createdRegionQuota.TmVdc.ID)
	check.Assert(rqPolicies[0].VirtualDatacenterStoragePolicy.RegionStoragePolicy.ID, Equals, sp.RegionStoragePolicy.ID)
	check.Assert(rqPolicies[0].VirtualDatacenterStoragePolicy.StorageLimitMiB, DeepEquals, addrOf(int64(100)))

	// Create an assignment in another quota to verify that batch creation returns
	// only the requested quota's policies, even when other assignments exist. TM
	// permits only one quota per region and tenant, so use a second organization.
	otherOrg, err := vcd.client.CreateTmOrg(&types.TmOrg{
		Name:        check.TestName() + "-other",
		DisplayName: check.TestName() + "-other",
		IsEnabled:   true,
	})
	check.Assert(err, IsNil)
	check.Assert(otherOrg, NotNil)
	PrependToCleanupListOpenApi(otherOrg.TmOrg.ID, otherOrg.TmOrg.Name, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgs+otherOrg.TmOrg.ID)
	defer func() {
		check.Check(otherOrg.Disable(), IsNil)
		check.Check(otherOrg.Delete(), IsNil)
	}()
	otherConfig := *vdcType
	otherConfig.Name += "-other"
	otherConfig.Org = &types.OpenApiReference{ID: otherOrg.TmOrg.ID}
	otherQuota, err := vcd.client.CreateRegionQuota(&otherConfig)
	check.Assert(err, IsNil)
	check.Assert(otherQuota, NotNil)
	PrependToCleanupListOpenApi(otherQuota.TmVdc.ID, otherConfig.Name, types.OpenApiPathVcf+types.OpenApiEndpointTmVdcs+otherQuota.TmVdc.ID)
	defer func() { check.Check(otherQuota.Delete(), IsNil) }()
	otherPolicies, err := vcd.client.CreateRegionQuotaStoragePolicies(otherQuota.TmVdc.ID, &types.VirtualDatacenterStoragePolicies{
		Values: []types.VirtualDatacenterStoragePolicy{{
			RegionStoragePolicy: types.OpenApiReference{ID: sp.RegionStoragePolicy.ID},
			StorageLimitMiB:     addrOf(int64(100)),
			VirtualDatacenter:   types.OpenApiReference{ID: otherQuota.TmVdc.ID},
		}},
	})
	check.Assert(err, IsNil)
	check.Assert(len(otherPolicies), Equals, 1)
	check.Assert(otherPolicies[0].VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(otherPolicies[0].VirtualDatacenterStoragePolicy.VirtualDatacenter.ID, Equals, otherQuota.TmVdc.ID)
	check.Assert(otherPolicies[0].VirtualDatacenterStoragePolicy.RegionStoragePolicy.ID, Equals, sp.RegionStoragePolicy.ID)
	check.Assert(otherPolicies[0].VirtualDatacenterStoragePolicy.ID, Not(Equals), rqPolicies[0].VirtualDatacenterStoragePolicy.ID)

	// Getting policies. We just test methods with RegionQuota receiver as these wrap the generic ones.
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
		StorageLimitMiB:     addrOf(int64(200)),
		VirtualDatacenter:   rqPolicies[0].VirtualDatacenterStoragePolicy.VirtualDatacenter,
	})
	check.Assert(err, IsNil)
	check.Assert(updatedPolicy, NotNil)
	check.Assert(updatedPolicy.VirtualDatacenterStoragePolicy.StorageLimitMiB, NotNil)
	check.Assert(*updatedPolicy.VirtualDatacenterStoragePolicy.StorageLimitMiB, Equals, int64(200))

	policy, err = vcd.client.GetRegionQuotaStoragePolicyById(rqPolicies[0].VirtualDatacenterStoragePolicy.ID)
	check.Assert(err, IsNil)
	check.Assert(policy.VirtualDatacenterStoragePolicy, NotNil)
	check.Assert(*policy.VirtualDatacenterStoragePolicy, DeepEquals, *updatedPolicy.VirtualDatacenterStoragePolicy)

	// Delete the policy
	err = policy.Delete()
	// TODO: TM: Test with more storage policies when possible, right now the Region Quota always has 1
	if vcd.client.Client.APIVCDMaxVersionIs("< 40.1") && len(rqPolicies) == 1 {
		// If there's only one policy, only in VCFA 9.0, expect an error. There must be always one Storage Policy
		check.Assert(err, NotNil)
		check.Assert(strings.Contains(err.Error(), "At least one storage class must be assigned for the region quota to remain functional"), Equals, true)
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

func (vcd *TestVCD) Test_TmRegionQuotaBulkStoragePolicies(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	region, cleanup := tmTestRegion(vcd, check)
	defer cleanup()
	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()
	quota, quotaCleanup := createRegionQuota(vcd, org, region, check)
	defer quotaCleanup()
	policies, err := quota.GetAllStoragePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(policies), Equals, 1)
	// The server schema requires a non-null storageLimitMiB with a minimum of 0
	for _, limit := range []*int64{addrOf(int64(256)), addrOf(int64(0))} {
		policy := *policies[0].VirtualDatacenterStoragePolicy
		policy.StorageLimitMiB = limit
		updated, err := quota.UpdateStoragePolicies(&types.VirtualDatacenterStoragePolicies{Values: []types.VirtualDatacenterStoragePolicy{policy}})
		check.Assert(err, IsNil)
		check.Assert(len(updated.Values), Equals, 1)
		check.Assert(updated.Values[0].StorageLimitMiB, DeepEquals, limit)
		actual, err := quota.GetStoragePolicyById(updated.Values[0].ID)
		check.Assert(err, IsNil)
		check.Assert(actual.VirtualDatacenterStoragePolicy.StorageLimitMiB, DeepEquals, limit)
	}
}
