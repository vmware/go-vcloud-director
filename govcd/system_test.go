package govcd

import (
	types "github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

// Tests System function GetOrgByName
func (vcd *TestVCD) TestGetOrgByName(test *C) {
	org, err := GetOrgByName(vcd.client, vcd.config.VCD.Org)
	test.Assert(err, IsNil)
	test.Assert(org.Org.Name, Equals, vcd.config.VCD.Org)
}

func (vcd *TestVCD) TestGetAdminOrgByName(test *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	test.Assert(err, IsNil)
	test.Assert(org.AdminOrg.Name, Equals, vcd.config.VCD.Org)
}

func (vcd *TestVCD) TestCreateOrg(test *C) {
	_, err := CreateOrg(vcd.client, "CREATEORG", "CREATEORG", true, &types.OrgSettings{})
	test.Assert(err, IsNil)
	// fetch newly created org
	org, err := GetAdminOrgByName(vcd.client, "CREATEORG")
	test.Assert(err, IsNil)
	test.Assert(org.AdminOrg.Name, Equals, "CREATEORG")
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	test.Assert(err, IsNil)
	// Check if org still exists
	org, err = GetAdminOrgByName(vcd.client, "CREATEORG")
	test.Assert(err, NotNil)

}
