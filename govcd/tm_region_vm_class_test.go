//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_RegionVmClass(check *C) {
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

	vmClasses, err := region.GetAllVmClasses(nil)
	check.Assert(err, IsNil)

	vmClasses2, err := vcd.client.GetAllRegionVirtualMachineClasses(nil)
	check.Assert(err, IsNil)
	check.Assert(len(vmClasses2) >= len(vmClasses), Equals, true)

	// Remaining test requires at least one Region VM Class in VCFA
	if len(vmClasses) == 0 {
		check.Skip("there are not enough Region VM Classes to continue with test")
	}

	vmClass, err := vcd.client.GetRegionVirtualMachineClassByNameAndRegionId(vmClasses[0].Name, region.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(*vmClass.RegionVirtualMachineClass, DeepEquals, *vmClasses[0])

	vmClass, err = vcd.client.GetRegionVirtualMachineClassById(vmClasses[0].ID)
	check.Assert(err, IsNil)
	check.Assert(*vmClass.RegionVirtualMachineClass, DeepEquals, *vmClasses[0])

}

func (vcd *TestVCD) Test_TmReservedVmClasses(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	region, cleanup := tmTestRegion(vcd, check)
	defer cleanup()
	rows, err := vcd.client.GetAllTmReservedVmClasses(url.Values{"filter": {"region.id==" + region.Region.ID}, "pageSize": {"1"}})
	check.Assert(err, IsNil)
	for _, row := range rows {
		check.Assert(row.Region, NotNil)
		check.Assert(row.Region.ID, Equals, region.Region.ID)
		check.Assert(row.Zone, NotNil)
		check.Assert(row.VirtualMachineClass, NotNil)
		check.Assert(row.TotalCount >= row.AvailableCount, Equals, true)
	}
}
