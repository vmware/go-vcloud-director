//go:build network || nsxt || functional || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetAllNsxtImportableSwitches(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	skipNoNsxtConfiguration(vcd, check)

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	allSwitches, err := nsxtVdc.GetAllNsxtImportableSwitches()
	check.Assert(err, IsNil)
	check.Assert(len(allSwitches) > 0, Equals, true)
}

func (vcd *TestVCD) Test_GetNsxtImportableSwitchByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	skipNoNsxtConfiguration(vcd, check)

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	logicalSwitch, err := nsxtVdc.GetNsxtImportableSwitchByName(vcd.config.VCD.Nsxt.NsxtImportSegment)
	check.Assert(err, IsNil)
	check.Assert(logicalSwitch.NsxtImportableSwitch.Name, Equals, vcd.config.VCD.Nsxt.NsxtImportSegment)
}

func (vcd *TestVCD) Test_GetFilteredNsxtImportableSwitches(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	skipNoNsxtConfiguration(vcd, check)

	// Check that nil filter returns error. This will work as a safeguard to also detect if future versions start accepting
	// empty filter value
	results, err := vcd.client.GetFilteredNsxtImportableSwitches(nil)
	check.Assert(err, Not(IsNil))
	check.Assert(results, IsNil)

	// Filter by VDC ID
	bareVdcId, err := getBareEntityUuid(vcd.nsxtVdc.Vdc.ID)
	check.Assert(err, IsNil)
	filter := map[string]string{"orgVdc": bareVdcId}
	results, err = vcd.client.GetFilteredNsxtImportableSwitches(filter)
	check.Assert(err, IsNil)
	check.Assert(len(results) > 0, Equals, true)

	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers) > 0, Equals, true)

	uuid := extractUuid(nsxtManagers[0].HREF)
	filter = map[string]string{"nsxTManager": uuid}
	results, err = vcd.client.GetFilteredNsxtImportableSwitches(filter)
	check.Assert(err, IsNil)
	check.Assert(len(results) > 0, Equals, true)

	switchByName, err := vcd.client.GetFilteredNsxtImportableSwitchesByName(filter, vcd.config.VCD.Nsxt.NsxtImportSegment)
	check.Assert(err, IsNil)
	check.Assert(switchByName.NsxtImportableSwitch.Name, Equals, vcd.config.VCD.Nsxt.NsxtImportSegment)

}
