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

	// Check that DVPG is not found when queries for vCenter ID that does not exist
	// dvpg, err := vcd.client.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.Dvpg, "urn:vcloud:vimserver:e63f5538-a546-4028-9966-120d017a9bb6")
	// check.Assert(err, NotNil)
	// check.Assert(ContainsNotFound(err), Equals, true)

	// spew.Dump(dvpgs)

	// dump all the DVPGs
	// spew.Dump(dvpg.VcenterImportableDvpg)

	// spew.Dump(err)

	// os.Exit(1)

	// check all DVPG functions without vCenter ID specified
	checkAllDvpgFunctions(check, vcd, "")

	// Get DVPG by name to extract vCenter ID
	dvpgByName, err := vcd.client.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.Dvpg, "")
	check.Assert(err, IsNil)
	check.Assert(dvpgByName, NotNil)

	vCenterId := dvpgByName.VcenterImportableDvpg.VirtualCenter.ID
	printVerbose("vCenter ID: %s\n", vCenterId)

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

	// Check that DVPG is not found when name does not exist
	_, err = vcd.client.GetVcenterImportableDvpgByName("non-existing-dvpg", vcenterId)
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	// Check that DVPG is not found when queries for vCenter ID that does not exist
	_, err = vcd.client.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.Dvpg, "urn:vcloud:vimserver:00000000-0000-0000-0000-000000000000")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	// _, err = nsxtVdc.GetVcenterImportableDvpgByName(vcd.config.VCD.Nsxt.Dvpg, "urn:vcloud:vimserver:00000000-0000-0000-0000-000000000000")
	// check.Assert(err, NotNil)
	// check.Assert(ContainsNotFound(err), Equals, true)
}
