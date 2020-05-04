// +build catalog functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetVAppTemplate(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_GetVAppTemplate: Catalog not found. Test can't proceed")
		return
	}
	check.Assert(cat, NotNil)

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_GetVAppTemplate: Catalog Item not given. Test can't proceed")
	}

	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)

	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()

	check.Assert(err, IsNil)
	check.Assert(vapptemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(vapptemplate.VAppTemplate.Description, Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
}

// Tests System function Delete by creating catalog item and
// deleting it after.
func (vcd *TestVCD) Test_Delete(check *C) {
	skipWhenOvaPathMissing(vcd, check)
	AddToCleanupList(TestDeleteCatalogItem, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_Delete")

	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	// add catalogItem
	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, TestDeleteCatalogItem, "upload from delete catalog item test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	catalogItem, err := catalog.GetCatalogItemByName(TestDeleteCatalogItem, false)
	check.Assert(err, IsNil)

	err = catalogItem.Delete()
	check.Assert(err, IsNil)

	// check through existing catalogItems
	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	entityFound := false
	for _, catalogItems := range catalog.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == TestDeleteCatalogItem {
				entityFound = true
			}
		}
	}
	check.Assert(entityFound, Equals, false)
}
