package govcd

import (
	. "gopkg.in/check.v1"
)

// Tests System function GetOrgByName
func (vcd *TestVCD) TestGetOrgByName(check *C) {
	org, err := GetOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, vcd.config.VCD.Org)
}
