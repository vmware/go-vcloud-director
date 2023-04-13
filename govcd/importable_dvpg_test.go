//go:build network || nsxt || functional || openapi || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VcenterImportableDvpg(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtConfiguration(vcd, check)

	if vcd.config.VCD.Nsxt.NsxtDvpg == "" {
		check.Skip("No NSX-T Dvpg provided")
	}

	// Get all DVPGs
	dvpgs, err := vcd.client.GetAllVcenterImportableDvpgs(nil)
	check.Assert(err, IsNil)
	check.Assert(len(dvpgs) > 0, Equals, true)

	// Get DVPG by name
	dvpgByName, err := vcd.client.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.NsxtDvpg)
	check.Assert(err, IsNil)
	check.Assert(dvpgByName, NotNil)
	check.Assert(dvpgByName.VcenterImportableDvpg.BackingRef.Name, Equals, vcd.config.VCD.Nsxt.NsxtDvpg)

	// Get all DVPGs withing NSX-T VDC
	nsxtVdc, err := vcd.org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(nsxtVdc, NotNil)

	allDvpgsWithingVdc, err := nsxtVdc.GetAllVcenterImportableDvpgs(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allDvpgsWithingVdc) > 0, Equals, true)

	// Get DVPG by name within NSX-T VDC
	dvpgByNameWithinVdc, err := nsxtVdc.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.NsxtDvpg)
	check.Assert(err, IsNil)
	check.Assert(dvpgByNameWithinVdc, NotNil)
	check.Assert(dvpgByNameWithinVdc.VcenterImportableDvpg.BackingRef.Name, Equals, vcd.config.VCD.Nsxt.NsxtDvpg)
}
