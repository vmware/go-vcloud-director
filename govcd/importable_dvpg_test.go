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

	if vcd.config.VCD.Nsxt.Dvpg == "" {
		check.Skip("No NSX-T Dvpg provided")
	}

	// check all DVPG functions without vCenter ID specified
	checkAllDvpgFunctions(check, vcd, "")

	// Get DVPG by name to extract vCenter ID
	dvpgByName, err := vcd.client.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.Dvpg, "")
	check.Assert(err, IsNil)
	check.Assert(dvpgByName, NotNil)

	vCenterId := dvpgByName.VcenterImportableDvpg.VirtualCenter.ID

	// check all DVPG functions with vCenter ID specified
	checkAllDvpgFunctions(check, vcd, vCenterId)
}

func checkAllDvpgFunctions(check *C, vcd *TestVCD, vcenterId string) {
	// Get all DVPGs
	dvpgs, err := vcd.client.GetAllVcenterImportableDvpgs(vcenterId, nil)
	check.Assert(err, IsNil)
	check.Assert(len(dvpgs) > 0, Equals, true)

	// Get DVPG by name
	dvpgByName, err := vcd.client.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.Dvpg, vcenterId)
	check.Assert(err, IsNil)
	check.Assert(dvpgByName, NotNil)
	check.Assert(dvpgByName.VcenterImportableDvpg.BackingRef.Name, Equals, vcd.config.VCD.Nsxt.Dvpg)

	// Get all DVPGs withing NSX-T VDC
	nsxtVdc, err := vcd.org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(nsxtVdc, NotNil)

	allDvpgsWithingVdc, err := nsxtVdc.GetAllVcenterImportableDvpgs(vcenterId, nil)
	check.Assert(err, IsNil)
	check.Assert(len(allDvpgsWithingVdc) > 0, Equals, true)

	// Get DVPG by name within NSX-T VDC
	dvpgByNameWithinVdc, err := nsxtVdc.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.Dvpg, vcenterId)
	check.Assert(err, IsNil)
	check.Assert(dvpgByNameWithinVdc, NotNil)
	check.Assert(dvpgByNameWithinVdc.VcenterImportableDvpg.BackingRef.Name, Equals, vcd.config.VCD.Nsxt.Dvpg)
}
