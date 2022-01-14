//go:build vapp || functional || ALL
// +build vapp functional ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

// TODO: Write test for InstantiateVAppTemplate

func (vcd *TestVCD) Test_RefreshVAppTemplate(check *C) {

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

	catItem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Get VAppTemplate
	vAppTemplate, err := catItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	oldVAppTemplate := vAppTemplate

	err = vAppTemplate.Refresh()
	check.Assert(err, IsNil)
	check.Assert(oldVAppTemplate.VAppTemplate.ID, Equals, vAppTemplate.VAppTemplate.ID)
	check.Assert(oldVAppTemplate.VAppTemplate.Name, Equals, vAppTemplate.VAppTemplate.Name)
	check.Assert(oldVAppTemplate.VAppTemplate.HREF, Equals, vAppTemplate.VAppTemplate.HREF)
}

func (vcd *TestVCD) Test_UpdateVAppTemplate(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_UpdateVAppTemplate: Catalog not found. Test can't proceed")
		return
	}
	check.Assert(catalog, NotNil)

	itemName := check.TestName()

	description := "upload from test"
	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, description, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, check.TestName())

	catItem, err := catalog.GetCatalogItemByName(itemName, true)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, itemName)

	// Get VAppTemplate
	vAppTemplate, err := catItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, itemName)

	nameForUpdate := itemName + "updated"
	descriptionForUpdate := description + "updated"

	AddToCleanupList(nameForUpdate, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, check.TestName())

	vAppTemplate.VAppTemplate.Name = nameForUpdate
	vAppTemplate.VAppTemplate.Description = descriptionForUpdate
	vAppTemplate.VAppTemplate.GoldMaster = true

	_, err = vAppTemplate.Update()
	check.Assert(err, IsNil)
	err = vAppTemplate.Refresh()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, nameForUpdate)
	check.Assert(vAppTemplate.VAppTemplate.Description, Equals, descriptionForUpdate)
	check.Assert(vAppTemplate.VAppTemplate.GoldMaster, Equals, true)
}
