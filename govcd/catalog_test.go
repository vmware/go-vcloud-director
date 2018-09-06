/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_FindCatalogItem(check *C) {
	// Fetch Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	if err != nil {
		check.Skip("Test_FindCatalogItem: Catalog not found. Test can't proceed")
	}

	// Find Catalog Item
	if vcd.config.VCD.Catalog.Catalogitem == "" {
		check.Skip("Test_FindCatalogItem: Catalog Item not given. Test can't proceed")
	}
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
	check.Assert(catitem, Equals, CatalogItem{})
	check.Assert(err, IsNil)
}

// Creates a Catalog, updates the description, and checks the changes against the
// newly updated catalog. Then deletes the catalog
func (vcd *TestVCD) Test_UpdateCatalog(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(org, Not(Equals), AdminOrg{})
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

// Creates a Catalog, and then deletes the catalog, and checks if
// the catalog still exists. If it does the assertion fails.
func (vcd *TestVCD) Test_DeleteCatalog(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	adminCatalog, err := org.CreateCatalog("DeleteCatalogTest", "DeleteCatalogTest", true)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, "DeleteCatalogTest")
	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
	catalog, err := org.FindCatalog("DeleteCatalogTest")
	check.Assert(err, IsNil)
	check.Assert(catalog, Equals, Catalog{})

}
