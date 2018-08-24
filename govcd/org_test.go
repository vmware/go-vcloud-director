/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	types "github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_DeleteOrg(check *C) {
	_, err := CreateOrg(vcd.client, "DELETEORG", "DELETEORG", true, &types.OrgSettings{})
	check.Assert(err, IsNil)
	// fetch newly created org
	org, err := GetAdminOrgByName(vcd.client, "DELETEORG")
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, "DELETEORG")
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	// Check if org still exists
	org, err = GetAdminOrgByName(vcd.client, "DELETEORG")
	check.Assert(err, NotNil)
}

func (vcd *TestVCD) Test_UpdateOrg(check *C) {
	_, err := CreateOrg(vcd.client, "UPDATEORG", "UPDATEORG", true, &types.OrgSettings{
		OrgLdapSettings: &types.OrgLdapSettingsType{OrgLdapMode: "NONE"},
	})
	check.Assert(err, IsNil)
	// fetch newly created org
	org, err := GetAdminOrgByName(vcd.client, "UPDATEORG")
	check.Assert(err, IsNil)
	check.Assert(org.AdminOrg.Name, Equals, "UPDATEORG")
	org.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota = 100
	task, err := org.Update()
	check.Assert(err, IsNil)
	// Wait until update is complete
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	// Refresh
	org, err = GetAdminOrgByName(vcd.client, "UPDATEORG")
	check.Assert(org.AdminOrg.OrgSettings.OrgGeneralSettings.DeployedVMQuota, Equals, 100)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	check.Assert(err, IsNil)
	// Check if org still exists
	org, err = GetAdminOrgByName(vcd.client, "UPDATEORG")
	check.Assert(err, NotNil)
}

// Tests org function GetVDCByName
func (vcd *TestVCD) Test_GetVdcByName(check *C) {
	vdc, err := vcd.org.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
}

// Tests FindCatalog with Catalog in config file
func (vcd *TestVCD) Test_FindCatalog(check *C) {
	// Find Catalog
	cat, err := vcd.org.GetCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
}

func (vcd *TestVCD) Test_CreateCatalog(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	catalog, err := org.CreateCatalog("Test", "Test123", true)
	check.Assert(err, IsNil)
	check.Assert(catalog.AdminCatalog.Name, Equals, "Test")
	check.Assert(catalog.AdminCatalog.Description, Equals, "Test123")
	err = catalog.Delete(true, true)
	check.Assert(err, IsNil)
}
// Test for AdminOrg version of FindCatalog
func (vcd *TestVCD) Test_AdminFindCatalog(check *C) {
	// Fetch admin org version of current test org
	adminOrg, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	// Find Catalog
	cat, err := adminOrg.GetCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
}
