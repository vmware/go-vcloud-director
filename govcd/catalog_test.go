/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_FindCatalogItem(check *C) {
	// Fetch Catalog
	cat, err := vcd.org.GetCatalog(vcd.config.VCD.Catalog.Name)
	// Find Catalog Item
	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.Catalogitem)
	check.Assert(err, IsNil)
	check.Assert(catitem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.Catalogitem)
	// If given a description in config file then it checks if the descriptions match
	// Otherwise it skips the assert
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(catitem.CatalogItem.Description, Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
	// Test non-existant catalog item
	catitem, err = cat.FindCatalogItem("INVALID")
	check.Assert(err, NotNil)
}

// Testing UpdateCatalog
func (vcd *TestVCD) Test_UpdateCatalog(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	adminCatalog, err := org.CreateCatalog("UpdateCatalogTest", "UpdateCatalogTest", true)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, "UpdateCatalogTest")

	adminCatalog.AdminCatalog.Description = "Test123"
	task, err := adminCatalog.Update()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, "Test123")

	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_DeleteCatalog(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	adminCatalog, err := org.CreateCatalog("DeleteCatalogTest", "DeleteCatalogTest", true)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, "DeleteCatalogTest")
	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
	_, err = org.GetCatalog("DeleteCatalogTest")
	check.Assert(err, NotNil)

}
