//go:build tm || functional || ALL

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_RegionStoragePolicy(check *C) {
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

	allRegionStoragePolicies, err := vcd.client.GetAllRegionStoragePolicies(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allRegionStoragePolicies), Not(Equals), 0)

	rspById, err := vcd.client.GetRegionStoragePolicyById(allRegionStoragePolicies[0].RegionStoragePolicy.ID)
	check.Assert(err, IsNil)
	check.Assert(rspById, NotNil)
	check.Assert(*rspById.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	rspByName, err := vcd.client.GetRegionStoragePolicyByName(allRegionStoragePolicies[0].RegionStoragePolicy.Name)
	check.Assert(err, IsNil)
	check.Assert(rspByName, NotNil)
	check.Assert(*rspByName.RegionStoragePolicy, DeepEquals, *rspById.RegionStoragePolicy)

	rspByName2, err := region.GetStoragePolicyByName(allRegionStoragePolicies[0].RegionStoragePolicy.Name)
	check.Assert(err, IsNil)
	check.Assert(rspByName2, NotNil)
	check.Assert(*rspByName2.RegionStoragePolicy, DeepEquals, *allRegionStoragePolicies[0].RegionStoragePolicy)

	// Check ENF errors
	_, err = vcd.client.GetRegionStoragePolicyById("urn:vcloud:regionStoragePolicy:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(err, NotNil)
	// TODO:TM: Right now throws a 500 NPE...
	// check.Assert(ContainsNotFound(err), Equals, true)

	_, err = region.GetStoragePolicyByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetRegionStoragePolicyByName("NotExists")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)
}
