/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/util"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	OvaPath        = "/test_vapp_template.ova"
	OvaChunkedPath = "/template_with_custom_chunk_size.ova"
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
	checkUploadOvf(vcd, check, getCurrentPath()+OvaPath, TestCreateCatalog, "item1")
}

func (vcd *TestVCD) Test_UploadOvf_chuncked(check *C) {
	checkUploadOvf(vcd, check, getCurrentPath()+OvaChunkedPath, TestCreateCatalog+"2", "item2")
}

func (vcd *TestVCD) Test_UploadOvf_progress_works(check *C) {
	catalogName := TestCreateCatalog + "3"
	itemName := "item3"
	setupCatalog(vcd, check, catalogName)

	catalog, org := findCatalog(vcd, check, catalogName)
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name, "Test_UploadOvf")

	//execute
	uploadTask, err := catalog.UploadOvf(getCurrentPath()+OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	for {
		if value := uploadTask.GetUploadProgress(); value == "100.00" {
			break
		} else {
			check.Assert(value, Not(Equals), "")
		}
	}
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	catalog, err = org.FindCatalog(catalogName)
	check.Assert(catalog.Catalog.CatalogItems[0].CatalogItem[0].Name, Equals, itemName)
}

func (vcd *TestVCD) Test_UploadOvf_ShowProgress_works(check *C) {
	catalogName := TestCreateCatalog + "4"
	itemName := "item4"
	setupCatalog(vcd, check, catalogName)

	catalog, org := findCatalog(vcd, check, catalogName)
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name, "Test_UploadOvf")

	//execute
	uploadTask, err := catalog.UploadOvf(getCurrentPath()+OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)

	//take control of stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	uploadTask.ShowUploadProgress()

	w.Close()
	//read stdin
	result, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	check.Assert(string(result), Matches, ".*Upload progress 100.00%")

	//verify
	catalog, err = org.FindCatalog(catalogName)
	check.Assert(catalog.Catalog.CatalogItems[0].CatalogItem[0].Name, Equals, itemName)
}

func (vcd *TestVCD) Test_UploadOvf_error_withSameItem(check *C) {
	catalogName := TestCreateCatalog + "5"
	itemName := "item5"
	setupCatalog(vcd, check, catalogName)

	catalog, _ := findCatalog(vcd, check, catalogName)
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name, "Test_UploadOvf")

	//add item
	uploadTask, err := catalog.UploadOvf(getCurrentPath()+OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//execute
	_, err = catalog.UploadOvf(getCurrentPath()+OvaPath, itemName, "upload from test", 1024)
	check.Assert(err.Error(), Matches, ".*API Error: 500.*Catalog Item name is not unique within the catalog.*")
}

func (vcd *TestVCD) Test_UploadOvf_cleaned_extracted_files(check *C) {
	catalogName := TestCreateCatalog + "6"
	itemName := "item6"
	setupCatalog(vcd, check, catalogName)

	//check existing count of folders
	oldFolderCount := countFolders()

	catalog, _ := findCatalog(vcd, check, catalogName)
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name, "Test_UploadOvf")

	//execute
	uploadTask, err := catalog.UploadOvf(getCurrentPath()+OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	check.Assert(oldFolderCount, Equals, countFolders())

}

func countFolders() int {
	files, err := ioutil.ReadDir(os.TempDir())
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for _, f := range files {
		if strings.Contains(f.Name(), util.TmpDirPrefix+".*") {
			count++
		}
	}
	return count
}

func checkUploadOvf(vcd *TestVCD, check *C, ovaFileName, testCreateCatalog, itemName string) {
	setupCatalog(vcd, check, testCreateCatalog)

	catalog, org := findCatalog(vcd, check, testCreateCatalog)
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name, "Test_UploadOvf")

	//execute
	uploadTask, err := catalog.UploadOvf(ovaFileName, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	//verify
	catalog, err = org.FindCatalog(testCreateCatalog)
	check.Assert(catalog.Catalog.CatalogItems[0].CatalogItem[0].Name, Equals, itemName)
}

func findCatalog(vcd *TestVCD, check *C, testCreateCatalog string) (Catalog, AdminOrg) {
	org := getOrg(vcd, check)
	catalog, err := org.FindCatalog(testCreateCatalog)
	check.Assert(err, IsNil)
	return catalog, org
}

func setupCatalog(vcd *TestVCD, check *C, testCreateCatalog string) {
	org := getOrg(vcd, check)

	// creating catalog
	adminCatalog, err := org.CreateCatalog(testCreateCatalog, TestCreateCatalogDesc, true)
	check.Assert(err, IsNil)
	AddToCleanupList(testCreateCatalog, "catalog", vcd.org.Org.Name, "Test_UploadOvf")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, testCreateCatalog)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)
	task := NewTask(&vcd.client.Client)
	task.Task = adminCatalog.AdminCatalog.Tasks.Task[0]
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}

func getOrg(vcd *TestVCD, check *C) AdminOrg {
	// Fetching organization
	org, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	return org
}

//helper function to get current runing dir.
func getCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
