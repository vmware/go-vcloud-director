//go:build vsphere || functional || ALL

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetVcenters(check *C) {

	vcenters, err := vcd.client.GetAllVcenters()
	check.Assert(err, IsNil)

	check.Assert(len(vcenters) > 0, Equals, true)

	for _, vc := range vcenters {
		vcenterById, err := vcd.client.GetVcenterById(vc.VSphereVcenter.VcId)
		check.Assert(err, IsNil)
		check.Assert(vc.VSphereVcenter.VcId, Equals, vcenterById.VSphereVcenter.VcId)
		vcenterByName, err := vcd.client.GetVcenterByName(vc.VSphereVcenter.Name)
		check.Assert(err, IsNil)
		check.Assert(vc.VSphereVcenter.VcId, Equals, vcenterByName.VSphereVcenter.VcId)
	}

}
