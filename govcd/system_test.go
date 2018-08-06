package govcd

import (
	. "gopkg.in/check.v1"
)

// Tests System function GetOrgByName
func (vcd *TestVCD) TestGetOrgByName(test *C) {
	org, err := GetOrgByName(vcd.client, vcd.config.VCD.Org)
	test.Assert(err, IsNil)
	test.Assert(org.Org.Name, Equals, vcd.config.VCD.Org)
}
