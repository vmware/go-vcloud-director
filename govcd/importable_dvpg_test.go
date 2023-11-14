//go:build network || nsxt || functional || openapi || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
	"strings"
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

	var compatibleDVPG []*VcenterImportableDvpg
	for _, dvpg := range dvpgs {
		// make a list of the port groups created with 'count' during the VCD configuration
		if strings.HasPrefix(dvpg.VcenterImportableDvpg.BackingRef.Name, vcd.config.VCD.Nsxt.NsxtDvpg) {
			compatibleDVPG = append(compatibleDVPG, dvpg)
		}
	}

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

	check.Assert(len(compatibleDVPG) > 1, Equals, true,
		Commentf("The VCD %s was not configured with multiple vsphere_distributed_port_group", vcd.config.Provider.Url))
	// test that port group created with the same switch ID are compatible
	foundSameParent := false
	for _, dvpg := range dvpgs {
		for _, other := range dvpgs {
			if other == dvpg {
				continue
			}
			if dvpg.Parent().ID == other.Parent().ID {
				foundSameParent = true
				break
			}
		}
		if foundSameParent {
			break
		}
	}
	check.Assert(foundSameParent, Equals, true)
	foundCompatible := false
	for _, dvpg := range dvpgs {
		if dvpg.UsableWith(compatibleDVPG...) {
			foundCompatible = true
			break
		}
	}
	check.Assert(len(compatibleDVPG), Equals, 3)
	check.Assert(foundCompatible, Equals, true)
}
