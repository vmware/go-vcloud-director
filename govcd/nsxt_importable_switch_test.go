// +build network nsxt functional ALL

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

	if vcd.client.Client.APIVCDMaxVersionIs("< 34") {
		check.Skip("At least VCD 10.1 is required")
	}
	skipNoNsxtConfiguration(vcd, check)

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	allSwitches, err := nsxtVdc.GetAllNsxtImportableSwitches()
	check.Assert(err, IsNil)
	check.Assert(len(allSwitches) > 1, Equals, true)
	// spew.Dump(allSwitches)

}

func (vcd *TestVCD) Test_GetNsxtImportableSwitchByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	if vcd.client.Client.APIVCDMaxVersionIs("< 34") {
		check.Skip("At least VCD 10.1 is required")
	}
	skipNoNsxtConfiguration(vcd, check)

	nsxtVdc, err := vcd.org.GetVDCByNameOrId(vcd.config.VCD.Nsxt.Vdc, true)
	check.Assert(err, IsNil)

	logicalSwitch, err := nsxtVdc.GetNsxtImportableSwitchByName(vcd.config.VCD.Nsxt.UnusedSegment)
	check.Assert(err, IsNil)
	check.Assert(logicalSwitch.NsxtImportableSwitch.Name, Equals, vcd.config.VCD.Nsxt.UnusedSegment)
}
