//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

// TODO: TM: Missing Create, Update, Delete
func (vcd *TestVCD) Test_RegionStoragePolicy(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	allRegionStoragePolicies, err := vcd.client.GetAllRegionStoragePolicies(nil)
	check.Assert(err, IsNil)

	// TODO: TM: Once we can create Region SPs, we don't need the TM environment to have pre-existing Storage Policies
	if len(allRegionStoragePolicies) == 0 {
		check.Skip("didn't find any Region Storage Policy")
	}

	rspById, err := vcd.client.GetRegionStoragePolicyById(allRegionStoragePolicies[0].RegionStoragePolicy.Id)
	check.Assert(err, IsNil)
	check.Assert(*rspById.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	rspByName, err := vcd.client.GetRegionStoragePolicyByName(allRegionStoragePolicies[0].RegionStoragePolicy.Name)
	check.Assert(err, IsNil)
	check.Assert(*rspByName.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	// Check ENF errors
	_, err = vcd.client.GetRegionStoragePolicyById("urn:vcloud:regionStoragePolicy:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetRegionStoragePolicyByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
}
