package govcd

import (
	types "github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

// Tests System function GetOrgByName by checking if the org object
// return has the same name as the one provided in the config file.
// Asserts an error if the names don't match or if the function returned
// an error.
func (vcd *TestVCD) Test_GetOrgByName(check *C) {
	org, err := GetOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, vcd.config.VCD.Org)
}

// Tests System function GetAdminOrgByName by checking if the AdminOrg object
// return has the same name as the one provided in the config file. Asserts
// an error if the names don't match or if the function returned an error.
func (vcd *TestVCD) Test_GetAdminOrgByName(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, vcd.config.VCD.Org)
}

// Tests the creation of a an org with general settings,
// org vapp template settings, and orgldapsettings. Asserts an
// error if the task, fetching the org, or deleting the org fails
func (vcd *TestVCD) Test_CreateOrg(check *C) {
	settings := &types.OrgSettings{
		OrgGeneralSettings: &types.OrgGeneralSettings{
			CanPublishCatalogs:       true,
			DeployedVMQuota:          10,
			StoredVMQuota:            10,
			UseServerBootSequence:    true,
			DelayAfterPowerOnSeconds: 3,
		},
		OrgVAppTemplateSettings: &types.VAppTemplateLeaseSettings{
			DeleteOnStorageLeaseExpiration: true,
			StorageLeaseSeconds:            10,
		},
		OrgLdapSettings: &types.OrgLdapSettingsType{
			OrgLdapMode: "NONE",
		},
	}
	task, err := CreateOrg(vcd.client, "CREATEORG", "CREATEORG", true, settings)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
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
