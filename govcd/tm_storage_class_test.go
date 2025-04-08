//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_StorageClass(check *C) {
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

	allStorageClasses, err := vcd.client.GetAllStorageClasses(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allStorageClasses), Not(Equals), 0)

	rspById, err := vcd.client.GetStorageClassById(allStorageClasses[0].StorageClass.ID)
	check.Assert(err, IsNil)
	check.Assert(rspById, NotNil)
	check.Assert(*rspById.StorageClass, DeepEquals, *allStorageClasses[0].StorageClass)

	rspByName, err := vcd.client.GetStorageClassByName(allStorageClasses[0].StorageClass.Name)
	check.Assert(err, IsNil)
	check.Assert(rspByName, NotNil)
	check.Assert(*rspByName.StorageClass, DeepEquals, *rspById.StorageClass)

	rspByName2, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(rspByName2, NotNil)
	check.Assert(rspByName2.StorageClass.Name, Equals, vcd.config.Tm.StorageClass)

	// Check ENF errors
	_, err = vcd.client.GetStorageClassById("urn:vcloud:storageClass:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(err, NotNil)
	// TODO:TM: Right now throws a 500 NPE...
	// check.Assert(ContainsNotFound(err), Equals, true)

	_, err = region.GetStorageClassByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetStorageClassByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
}
