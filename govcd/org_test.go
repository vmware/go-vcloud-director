/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	types "github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
	"time"
)

// Tests Refresh for Org by updating the org and then asserting if the
// variable is updated.
func (vcd *TestVCD) Test_RefreshOrg(check *C) {

	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	adminOrg, err := GetAdminOrgByName(vcd.client, TestRefreshOrg)
	if adminOrg != (AdminOrg{}) {
		err = adminOrg.Delete(true, true)
		check.Assert(err, IsNil)
	}
	_, err = CreateOrg(vcd.client, TestRefreshOrg, TestRefreshOrg, true, &types.OrgSettings{
		OrgLdapSettings: &types.OrgLdapSettingsType{OrgLdapMode: "NONE"},
	})
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestRefreshOrg, "org", "", "Test_RefreshOrg")
	// fetch newly created org
	org, err := GetOrgByName(vcd.client, TestRefreshOrg)
	check.Assert(err, IsNil)
	check.Assert(org.Org.Name, Equals, TestRefreshOrg)
	// fetch admin version of org for updating
	adminOrg, err = GetAdminOrgByName(vcd.client, TestRefreshOrg)
	check.Assert(err, IsNil)
	check.Assert(adminOrg.AdminOrg.Name, Equals, TestRefreshOrg)
	adminOrg.AdminOrg.FullName = TestRefreshOrgFullName
	task, err := adminOrg.Update()
	check.Assert(err, IsNil)
	// Wait until update is complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// Test Refresh on normal org
	err = org.Refresh()
	check.Assert(err, IsNil)
	check.Assert(org.Org.FullName, Equals, TestRefreshOrgFullName)
	// Test Refresh on admin org
	err = adminOrg.Refresh()
	check.Assert(err, IsNil)
	check.Assert(adminOrg.AdminOrg.FullName, Equals, TestRefreshOrgFullName)
	// Delete, with force and recursive true
	err = adminOrg.Delete(true, true)
	check.Assert(err, IsNil)
}

// Creates an org DELETEORG and then deletes it to test functionality of
// delete org. Fails if org still exists
func (vcd *TestVCD) Test_DeleteOrg(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	org, err := GetAdminOrgByName(vcd.client, TestDeleteOrg)
	if org != (AdminOrg{}) {
		err = org.Delete(true, true)
		check.Assert(err, IsNil)
	}
	_, err = CreateOrg(vcd.client, TestDeleteOrg, TestDeleteOrg, true, &types.OrgSettings{})
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestDeleteOrg, "org", "", "Test_DeleteOrg")
	// fetch newly created org
	org, err = GetAdminOrgByName(vcd.client, TestDeleteOrg)
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, TestDeleteOrg)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	doesOrgExist(check, vcd)
}

// Creates a org UPDATEORG, changes the deployed vm quota on the org,
// and tests the update functionality of the org. Then it deletes the org.
// Fails if the deployedvmquota variable is not changed when the org is
// refetched.
func (vcd *TestVCD) Test_UpdateOrg(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	org, err := GetAdminOrgByName(vcd.client, TestUpdateOrg)
	if org != (AdminOrg{}) {
		err = org.Delete(true, true)
		check.Assert(err, IsNil)
	}
	_, err = CreateOrg(vcd.client, TestUpdateOrg, TestUpdateOrg, true, &types.OrgSettings{
		OrgLdapSettings: &types.OrgLdapSettingsType{OrgLdapMode: "NONE"},
	})
	check.Assert(err, IsNil)
	AddToCleanupList(TestUpdateOrg, "org", "", "TestUpdateOrg")
	// fetch newly created org
	org, err = GetAdminOrgByName(vcd.client, TestUpdateOrg)
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, TestUpdateOrg)
	org.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota = 100
	task, err := org.Update()
	check.Assert(err, IsNil)
	// Wait until update is complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// Refresh
	err = org.Refresh()
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota, Equals, 100)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	doesOrgExist(check, vcd)
}

func doesOrgExist(check *C, vcd *TestVCD) {
	var org AdminOrg
	var err error
	for i := 0; i < 30; i++ {
		org, err = GetAdminOrgByName(vcd.client, TestDeleteOrg)
		if org == (AdminOrg{}) {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}
	check.Assert(org, Equals, AdminOrg{})
	check.Assert(err, IsNil)
}

// Tests org function GetVDCByName with the vdc specified
// in the config file. Then tests with a vdc that doesn't exist.
// Fails if the config file name doesn't match with the found vdc, or
// if the invalid vdc is found by the function.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_GetVdcByName(check *C) {
	vdc, err := vcd.org.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(vdc, Not(Equals), Vdc{})
	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
	// Try a vdc that doesn't exist
	vdc, err = vcd.org.GetVdcByName(INVALID_NAME)
	check.Assert(vdc, Equals, Vdc{})
	check.Assert(err, IsNil)
}

// Tests org function Admin version of GetVDCByName with the vdc
// specified in the config file. Fails if the names don't match
// or the function returns an error.  Also tests an vdc
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_Admin_GetVdcByName(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})
	vdc, err := adminOrg.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(vdc, Not(Equals), Vdc{})
	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
	// Try a vdc that doesn't exist
	vdc, err = adminOrg.GetVdcByName(INVALID_NAME)
	check.Assert(vdc, Equals, Vdc{})
	check.Assert(err, IsNil)
}

// Tests FindCatalog with Catalog in config file. Fails if the name and
// description don't match the catalog elements in the config file or if
// function returns an error.  Also tests an catalog
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_FindCatalog(check *C) {
	// Find Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(cat, Not(Equals), Catalog{})
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
	// Check Invalid Catalog
	cat, err = vcd.org.FindCatalog(INVALID_NAME)
	check.Assert(cat, Equals, Catalog{})
	check.Assert(err, IsNil)
}

// Tests Admin version of FindCatalog with Catalog in config file. Asserts an
// error if the name and description don't match the catalog elements in
// the config file or if function returns an error.  Also tests an catalog
// that doesn't exist. Asserts an error if the function finds it or
// if the error is not nil.
func (vcd *TestVCD) Test_Admin_FindCatalog(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	// Fetch admin org version of current test org
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(adminOrg, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	// Find Catalog
	cat, err := adminOrg.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(cat, Not(Equals), Catalog{})
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
	// Check Invalid Catalog
	cat, err = adminOrg.FindCatalog(INVALID_NAME)
	check.Assert(cat, Equals, Catalog{})
	check.Assert(err, IsNil)
}

// Tests CreateCatalog by creating a catalog named CatalogCreationTest and
// asserts that the catalog returned contains the right contents or if it fails.
// Then Deletes the catalog.
func (vcd *TestVCD) Test_CreateCatalog(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	org, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	catalog, err := org.FindAdminCatalog(TestCreateCatalog)
	if catalog != (AdminCatalog{}) {
		err = catalog.Delete(true, true)
		check.Assert(err, IsNil)
	}
	catalog, err = org.CreateCatalog(TestCreateCatalog, TestCreateCatalogDesc)
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateCatalog, "catalog", vcd.org.Org.Name, "Test_CreateCatalog")
	check.Assert(catalog.AdminCatalog.Name, Equals, TestCreateCatalog)
	check.Assert(catalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)
	task := NewTask(&vcd.client.Client)
	task.Task = catalog.AdminCatalog.Tasks.Task[0]
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	org, err = GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	copyCatalog, err := org.FindAdminCatalog(TestCreateCatalog)
	check.Assert(copyCatalog, Not(Equals), AdminCatalog{})
	check.Assert(err, IsNil)
	check.Assert(catalog.AdminCatalog.Name, Equals, copyCatalog.AdminCatalog.Name)
	check.Assert(catalog.AdminCatalog.Description, Equals, copyCatalog.AdminCatalog.Description)
	check.Assert(catalog.AdminCatalog.IsPublished, Equals, false)
	err = catalog.Delete(true, true)
	check.Assert(err, IsNil)
}

// Test for GetAdminCatalog. Gets admin version of Catalog, and asserts that
// the names and description match, and that no error is returned
func (vcd *TestVCD) Test_GetAdminCatalog(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	// Fetch admin org version of current test org
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	// Find Catalog
	cat, err := adminOrg.FindAdminCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	check.Assert(cat.AdminCatalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.AdminCatalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
}
