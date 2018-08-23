/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	types "github.com/vmware/go-vcloud-director/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) TestDeleteOrg(test *C) {
	_, err := CreateOrg(vcd.client, "DELETEORG", "DELETEORG", true, &types.OrgSettings{})
	test.Assert(err, IsNil)
	// fetch newly created org
	org, err := GetAdminOrgByName(vcd.client, "DELETEORG")
	test.Assert(err, IsNil)
	test.Assert(org.AdminOrg.Name, Equals, "DELETEORG")
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	test.Assert(err, IsNil)
	// Check if org still exists
	org, err = GetAdminOrgByName(vcd.client, "DELETEORG")
	test.Assert(err, NotNil)
}

func (vcd *TestVCD) TestUpdateOrg(test *C) {
	_, err := CreateOrg(vcd.client, "UPDATEORG", "UPDATEORG", true, &types.OrgSettings{OrgLdapSettings: &types.OrgLdapSettingsType{OrgLdapMode: "NONE"}})
	test.Assert(err, IsNil)
	// fetch newly created org
	org, err := GetAdminOrgByName(vcd.client, "UPDATEORG")
	test.Assert(err, IsNil)
	test.Assert(org.AdminOrg.Name, Equals, "UPDATEORG")
	org.AdminOrg.OrgSettings.General.DeployedVMQuota = 100
	task, err := org.Update()
	test.Assert(err, IsNil)
	// Wait until update is complete
	err = task.WaitTaskCompletion()
	test.Assert(err, IsNil)
	// Refresh
	org, err = GetAdminOrgByName(vcd.client, "UPDATEORG")
	test.Assert(org.AdminOrg.OrgSettings.General.DeployedVMQuota, Equals, 100)
	// Delete, with force and recursive true
	err = org.Delete(true, true)
	test.Assert(err, IsNil)
	// Check if org still exists
	org, err = GetAdminOrgByName(vcd.client, "UPDATEORG")
	test.Assert(err, NotNil)
}

// Tests org function GetVDCByName
func (vcd *TestVCD) TestGetVdcByName(check *C) {
	vdc, err := vcd.org.GetVdcByName(vcd.config.VCD.Vdc)
	check.Assert(err, IsNil)
	check.Assert(vdc.Vdc.Name, Equals, vcd.config.VCD.Vdc)
}

// Tests FindCatalog with Catalog in config file
func (vcd *TestVCD) Test_FindCatalog(check *C) {
	// Find Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	check.Assert(err, IsNil)
	check.Assert(cat.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)
	// checks if user gave a catalog description in config file
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(cat.Catalog.Description, Equals, vcd.config.VCD.Catalog.Description)
	}
}
