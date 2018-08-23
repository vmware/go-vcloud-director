package govcd

import (
	types "github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

// Tests System function GetOrgByName
func (vcd *TestVCD) Test_GetOrgByName(check *C) {
	org, err := GetOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, vcd.config.VCD.Org)
}

func (vcd *TestVCD) Test_GetAdminOrgByName(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, vcd.config.VCD.Org)
}

func (vcd *TestVCD) Test_CreateOrg(check *C) {
	_, err := CreateOrg(vcd.client, "CREATEORG", "CREATEORG", true, &types.OrgSettings{})
	check.Assert(err, IsNil)
	// fetch newly created org
	org, err := GetAdminOrgByName(vcd.client, "CREATEORG")
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, "CREATEORG")
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	// Check if org still exists
	org, err = GetAdminOrgByName(vcd.client, "CREATEORG")
	check.Assert(err, NotNil)

}
