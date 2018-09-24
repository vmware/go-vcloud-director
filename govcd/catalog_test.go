/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/util"
	. "gopkg.in/check.v1"
	"path"
	"runtime"
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
	catalog, err := org.FindAdminCatalog(TestUpdateCatalog)
	if catalog != (AdminCatalog{}) {
		err = catalog.Delete(true, true)
		check.Assert(err, IsNil)
	}
	adminCatalog, err := org.CreateCatalog(TestUpdateCatalog, TestUpdateCatalog, true)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestUpdateCatalog, "catalog", vcd.config.VCD.Org, "Test_UpdateCatalog")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, TestUpdateCatalog)

	adminCatalog.AdminCatalog.Description = TestCreateCatalogDesc
	err = adminCatalog.Update()
	check.Assert(err, IsNil)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)

	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
}

// Creates a Catalog, and then deletes the catalog, and checks if
// the catalog still exists. If it does the assertion fails.
func (vcd *TestVCD) Test_DeleteCatalog(check *C) {
	org, err := GetAdminOrgByName(vcd.client, vcd.config.VCD.Org)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	adminCatalog, err := org.FindAdminCatalog(TestDeleteCatalog)
	if adminCatalog != (AdminCatalog{}) {
		err = adminCatalog.Delete(true, true)
		check.Assert(err, IsNil)
	}
	adminCatalog, err = org.CreateCatalog(TestDeleteCatalog, TestDeleteCatalog, true)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestDeleteCatalog, "catalog", vcd.config.VCD.Org, "Test_DeleteCatalog")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, TestDeleteCatalog)
	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
	catalog, err := org.FindCatalog(TestDeleteCatalog)
	check.Assert(err, IsNil)
	check.Assert(catalog, Equals, Catalog{})

}

func (vcd *TestVCD) Test_UploadOvf(check *C) {
	// Fetching organization
	org, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)

	// creating catalog
	adminCatalog, err := org.CreateCatalog(TestCreateCatalog, TestCreateCatalogDesc, true)
	check.Assert(err, IsNil)
	AddToCleanupList(TestCreateCatalog, "catalog", vcd.org.Org.Name, "Test_UploadOvf")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, TestCreateCatalog)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)
	task := NewTask(&vcd.client.Client)
	task.Task = adminCatalog.AdminCatalog.Tasks.Task[0]
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//find catalog
	org, err = GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	catalog, err := org.FindCatalog(TestCreateCatalog)
	check.Assert(err, IsNil)
	util.Logger.Printf("Catalog found#: %#v", catalog)
	itemName := "myTest"
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name, "Test_UploadOvf")
	uploadTask, err := catalog.UploadOvf(getCurrentPath()+"/template_with_custom_chunk_size.ova", itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	catalog, err = org.FindCatalog(TestCreateCatalog)
	check.Assert(catalog.Catalog.CatalogItems[0].CatalogItem[0].Name, Equals, itemName)
}

//helper function to get current runing dir.
func getCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
