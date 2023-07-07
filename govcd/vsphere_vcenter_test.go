//go:build vsphere || functional || ALL

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetVcenters(check *C) {

	if !vcd.client.Client.IsSysAdmin {
		check.Skip("this test requires system administrator privileges")
	}
	vcenters, err := vcd.client.GetAllVCenters(nil)
	check.Assert(err, IsNil)

	check.Assert(len(vcenters) > 0, Equals, true)

	for _, vc := range vcenters {
		vcenterById, err := vcd.client.GetVCenterById(vc.VSphereVCenter.VcId)
		check.Assert(err, IsNil)
		check.Assert(vc.VSphereVCenter.VcId, Equals, vcenterById.VSphereVCenter.VcId)
		vcenterByName, err := vcd.client.GetVCenterByName(vc.VSphereVCenter.Name)
		check.Assert(err, IsNil)
		check.Assert(vc.VSphereVCenter.VcId, Equals, vcenterByName.VSphereVCenter.VcId)
	}

}
