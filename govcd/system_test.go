/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
	"time"
)

// Tests System function GetOrgByName by checking if the org object
// return has the same name as the one provided in the config file.
// Asserts an error if the names don't match or if the function returned
// an error. Also tests an org that doesn't exist. Asserts an error
// if the function finds it or if the error is not nil.
func (vcd *TestVCD) Test_GetOrgByName(check *C) {
	org, err := GetOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(org, Not(Equals), Org{})
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, vcd.config.VCD.Org)
	// Tests Org That doesn't exist
	org, err = GetOrgByName(vcd.client, INVALID_NAME)
	check.Assert(org, Equals, Org{})
	check.Assert(err, IsNil)
}

// Tests System function GetAdminOrgByName by checking if the AdminOrg object
// return has the same name as the one provided in the config file. Asserts
// an error if the names don't match or if the function returned an error.
// Also tests an org that doesn't exist. Asserts an error
// if the function finds it or if the error is not nil.
func (vcd *TestVCD) Test_GetAdminOrgByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, vcd.config.VCD.Org)
	// Tests Org That doesn't exist
	org, err = GetAdminOrgByName(vcd.client, INVALID_NAME)
	check.Assert(org, Equals, AdminOrg{})
	check.Assert(err, IsNil)
}

// Tests the creation of an org with general settings,
// org vapp template settings, and orgldapsettings. Asserts an
// error if the task, fetching the org, or deleting the org fails
func (vcd *TestVCD) Test_CreateOrg(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	org, err := GetAdminOrgByName(vcd.client, TestCreateOrg)
	if org != (AdminOrg{}) {
		err = org.Delete(true, true)
		check.Assert(err, IsNil)
	}
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
		OrgVAppLeaseSettings: &types.VAppLeaseSettings{
			PowerOffOnRuntimeLeaseExpiration: true,
			DeploymentLeaseSeconds:           1000000,
			DeleteOnStorageLeaseExpiration:   true,
			StorageLeaseSeconds:              1000000,
		},
		OrgLdapSettings: &types.OrgLdapSettingsType{
			OrgLdapMode: "NONE",
		},
	}
	task, err := CreateOrg(vcd.client, TestCreateOrg, TestCreateOrg, true, settings)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestCreateOrg, "org", "", "TestCreateOrg")
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// fetch newly created org
	org, err = GetAdminOrgByName(vcd.client, TestCreateOrg)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, TestCreateOrg)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	// Check if org still exists
	for i := 0; i < 30; i++ {
		org, err = GetAdminOrgByName(vcd.client, TestCreateOrg)
		if org == (AdminOrg{}) {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}
	check.Assert(org, Equals, AdminOrg{})
	check.Assert(err, IsNil)

}

// longer than the 128 characters so nothing can be named this
var INVALID_NAME = `*******************************************INVALID
					****************************************************
					************************`
