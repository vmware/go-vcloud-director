//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_RegionStoragePolicy(check *C) {
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

	allRegionStoragePolicies, err := region.GetAllStoragePolicies(nil)
	check.Assert(err, IsNil)

	rspById, err := region.GetStoragePolicyById(allRegionStoragePolicies[0].RegionStoragePolicy.ID)
	check.Assert(err, IsNil)
	check.Assert(*rspById.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	rspById2, err := vcd.client.GetRegionStoragePolicyById(allRegionStoragePolicies[0].RegionStoragePolicy.ID)
	check.Assert(err, IsNil)
	check.Assert(*rspById2.RegionStoragePolicy, DeepEquals, *rspById.RegionStoragePolicy)

	rspByName, err := region.GetStoragePolicyByName(allRegionStoragePolicies[0].RegionStoragePolicy.Name)
	check.Assert(err, IsNil)
	check.Assert(*rspByName.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	// Check ENF errors
	_, err = region.GetStoragePolicyById("urn:vcloud:regionStoragePolicy:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(err, NotNil)
	// TODO:TM: Right now throws a 500 NPE...
	// check.Assert(ContainsNotFound(err), Equals, true)

	_, err = region.GetStoragePolicyByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
}
