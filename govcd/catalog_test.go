/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/vmware/go-vcloud-director/util"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_FindCatalogItem(check *C) {
	// Fetch Catalog
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
	if err != nil {
		check.Skip("Test_FindCatalogItem: Catalog not found. Test can't proceed")
	}

	// Find Catalog Item
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_FindCatalogItem: Catalog Item not given. Test can't proceed")
	}
	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.CatalogItem)
	check.Assert(err, IsNil)
	check.Assert(catitem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)
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
	adminCatalog, err := org.CreateCatalog(TestUpdateCatalog, TestUpdateCatalog)
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
	adminCatalog, err = org.CreateCatalog(TestDeleteCatalog, TestDeleteCatalog)
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

// Tests System function UploadOvf by creating catalog and
// checking if provided standard ova file uploaded.
func (vcd *TestVCD) Test_UploadOvf(check *C) {
	skipWhenOvaPathMissing(vcd, check)
	checkUploadOvf(vcd, check, vcd.config.OVA.OVAPath, vcd.config.VCD.Catalog.Name, TestUploadOvf)
}

// Tests System function UploadOvf by creating catalog and
// checking if provided chunked ova file uploaded.
func (vcd *TestVCD) Test_UploadOvf_chunked(check *C) {
	skipWhenOvaPathMissing(vcd, check)
	checkUploadOvf(vcd, check, vcd.config.OVA.OVAChunkedPath, vcd.config.VCD.Catalog.Name, TestUploadOvf+"2")
}

// Tests System function UploadOvf by creating catalog and
// checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_UploadOvf_progress_works(check *C) {
	skipWhenOvaPathMissing(vcd, check)
	itemName := TestUploadOvf + "3"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OVAPath, itemName, "upload from test", 1024)
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

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	catalog, err = org.FindCatalog(vcd.config.VCD.Catalog.Name)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadOvf by creating catalog and
// checking UploadTask.ShowUploadProgress writes values of progress to stdin.
func (vcd *TestVCD) Test_UploadOvf_ShowUploadProgress_works(check *C) {
	skipWhenOvaPathMissing(vcd, check)
	itemName := TestUploadOvf + "4"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	//execute
	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OVAPath, itemName, "upload from test", 1024)
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
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	check.Assert(string(result), Matches, ".*Upload progress 100.00%")

	catalog, err = org.FindCatalog(vcd.config.VCD.Catalog.Name)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadOvf by creating catalog, creating catalog item
// and expecting specific error then trying to create same catalog item. As vCD returns cryptic error for such case.
func (vcd *TestVCD) Test_UploadOvf_error_withSameItem(check *C) {
	skipWhenOvaPathMissing(vcd, check)

	itemName := TestUploadOvf + "5"

	catalog, _ := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	//add item
	uploadTask, err2 := catalog.UploadOvf(vcd.config.OVA.OVAPath, itemName, "upload from test", 1024)
	check.Assert(err2, IsNil)
	err2 = uploadTask.WaitTaskCompletion()
	check.Assert(err2, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	catalog, _ = findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)
	_, err3 := catalog.UploadOvf(vcd.config.OVA.OVAPath, itemName, "upload from test", 1024)
	check.Assert(err3.Error(), Matches, ".*already exists. Upload with different name.*")
}

// Tests System function UploadOvf by creating catalog, uploading file and verifying
// that extracted files were deleted.s
func (vcd *TestVCD) Test_UploadOvf_cleaned_extracted_files(check *C) {
	skipWhenOvaPathMissing(vcd, check)

	itemName := TestUploadOvf + "6"

	//check existing count of folders
	oldFolderCount := countFolders()

	catalog, _ := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OVAPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

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

func checkUploadOvf(vcd *TestVCD, check *C, ovaFileName, catalogName, itemName string) {
	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(ovaFileName, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	catalog, err = org.FindCatalog(catalogName)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

func verifyCatalogItemUploaded(check *C, catalog Catalog, itemName string) {
	entityFound := false
	for _, catalogItems := range catalog.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == itemName {
				entityFound = true
			}
		}
	}
	check.Assert(entityFound, Equals, true)
}

func findCatalog(vcd *TestVCD, check *C, catalogName string) (Catalog, AdminOrg) {
	org := getOrg(vcd, check)
	catalog, err := org.FindCatalog(catalogName)
	check.Assert(err, IsNil)
	return catalog, org
}

func getOrg(vcd *TestVCD, check *C) AdminOrg {
	// Fetching organization
	org, err := GetAdminOrgByName(vcd.client, vcd.org.Org.Name)
	check.Assert(org, Not(Equals), AdminOrg{})
	check.Assert(err, IsNil)
	return org
}

func skipWhenOvaPathMissing(vcd *TestVCD, check *C) {
	if vcd.config.OVA.OVAPath == "" || vcd.config.OVA.OVAChunkedPath == "" {
		check.Skip("Skipping test because no ova path given")
	}
}

// Tests System function UploadMediaImage by checking if provided standard iso file uploaded.
func (vcd *TestVCD) Test_CatalogUploadMediaImage(check *C) {
	skipWhenMediaPathMissing(vcd, check)

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(TestCatalogUploadMedia, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(TestCatalogUploadMedia, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadCatalogMediaImage")

	//verifyMediaImageUploaded(vcd.vdc.client, check, TestUploadMedia)
	catalog, err = org.FindCatalog(vcd.config.VCD.Catalog.Name)
	verifyCatalogItemUploaded(check, catalog, TestCatalogUploadMedia)
}

// Tests System function UploadMediaImage by checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_progress_works(check *C) {
	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "2"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadCatalogMediaImage")

	catalog, err = org.FindCatalog(vcd.config.VCD.Catalog.Name)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadMediaImage by checking UploadTask.ShowUploadProgress writes values of progress to stdin.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_ShowUploadProgress_works(check *C) {
	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "3"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadCatalogMediaImage")

	check.Assert(string(result), Matches, ".*Upload progress 100.00%")
	catalog, err = org.FindCatalog(vcd.config.VCD.Catalog.Name)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadMediaImage by creating media item and expecting specific error
// then trying to create same media item. As vCD returns cryptic error for such case.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_error_withSameItem(check *C) {
	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "4"

	catalog, _ := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadCatalogMediaImage")

	_, err2 := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err2.Error(), Matches, ".*already exists. Upload with different name.*")
}

// Tests System function Delete by creating media item and
// deleting it after.
func (vcd *TestVCD) Test_CatalogDeleteMediaImage(check *C) {
	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "5"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadCatalogMediaImage")

	mediaItem, err := vcd.vdc.FindMediaImage(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaItem, Not(Equals), MediaItem{})

	task, err := mediaItem.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	mediaItem, err = vcd.vdc.FindMediaImage(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaItem, Equals, MediaItem{})

	//addition check
	// check through existing catalogItems
	catalog, err = org.FindCatalog(vcd.config.VCD.Catalog.Name)
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
