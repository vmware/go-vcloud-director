//go:build catalog || functional || ALL
// +build catalog functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	. "gopkg.in/check.v1"
)

// Tests catalog refresh
func (vcd *TestVCD) Test_CatalogRefresh(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	catalogName := vcd.config.VCD.Catalog.Name
	if catalogName == "" {
		check.Skip("Test_CatalogRefresh: Catalog name not given")
		return
	}
	cat, err := vcd.org.GetCatalogByName(catalogName, false)
	if err != nil {
		check.Skip("Test_CatalogRefresh: Catalog not found")
		return
	}
	catalogId := cat.Catalog.ID
	numItems := len(cat.Catalog.CatalogItems)
	dateCreated := cat.Catalog.DateCreated
	check.Assert(cat, NotNil)
	check.Assert(cat.Catalog.Name, Equals, catalogName)

	// Pollute the catalog structure
	cat.Catalog.Name = INVALID_NAME
	cat.Catalog.ID = invalidEntityId
	cat.Catalog.CatalogItems = nil
	cat.Catalog.DateCreated = ""

	// Get the catalog again from vCD
	err = cat.Refresh()
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)
	check.Assert(cat.Catalog.Name, Equals, catalogName)
	check.Assert(cat.Catalog.DateCreated, Equals, dateCreated)
	check.Assert(len(cat.Catalog.CatalogItems), Equals, numItems)
	check.Assert(cat.Catalog.ID, Equals, catalogId)
}

func (vcd *TestVCD) Test_FindCatalogItem(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// Fetch Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_FindCatalogItem: Catalog not found. Test can't proceed")
		return
	}

	// Find Catalog Item
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_FindCatalogItem: Catalog Item not given. Test can't proceed")
	}
	catalogItem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catalogItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)
	// If given a description in config file then it checks if the descriptions match
	// Otherwise it skips the assert
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(catalogItem.CatalogItem.Description, Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
	// Test non-existent catalog item
	catalogItem, err = cat.GetCatalogItemByName("INVALID", false)
	check.Assert(err, NotNil)
	check.Assert(catalogItem, IsNil)
}

// Creates a Catalog, updates the description, and checks the changes against the
// newly updated catalog. Then deletes the catalog
func (vcd *TestVCD) Test_UpdateCatalog(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalog, _ := org.GetAdminCatalogByName(TestUpdateCatalog, false)
	if catalog != nil {
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
	fmt.Printf("Running: %s\n", check.TestName())

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	oldCatalog, _ := org.GetAdminCatalogByName(TestDeleteCatalog, false)
	if oldCatalog != nil {
		err = oldCatalog.Delete(true, true)
		check.Assert(err, IsNil)
	}
	adminCatalog, err := org.CreateCatalog(TestDeleteCatalog, TestDeleteCatalog)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestDeleteCatalog, "catalog", vcd.config.VCD.Org, "Test_DeleteCatalog")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, TestDeleteCatalog)

	checkUploadOvf(vcd, check, vcd.config.OVA.OvaPath, TestDeleteCatalog, TestUploadOvf)
	err = adminCatalog.Delete(false, false)
	check.Assert(err, NotNil)
	// Catalog is not empty. An attempt to delete without recursion will fail
	check.Assert(strings.Contains(err.Error(), "You must remove"), Equals, true)

	err = adminCatalog.Delete(true, true)
	check.Assert(err, IsNil)
	doesCatalogExist(check, org)
}

func doesCatalogExist(check *C, org *AdminOrg) {
	var err error
	var catalog *AdminCatalog
	for i := 0; i < 30; i++ {
		catalog, err = org.GetAdminCatalogByName(TestDeleteCatalog, true)
		if catalog == nil {
			break
		} else {
			time.Sleep(time.Second)
		}
	}
	check.Assert(err, NotNil)
}

// Tests System function UploadOvf by creating catalog and
// checking if provided standard ova file uploaded.
func (vcd *TestVCD) Test_UploadOvf(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)
	checkUploadOvf(vcd, check, vcd.config.OVA.OvaPath, vcd.config.VCD.Catalog.Name, TestUploadOvf)
}

// Tests System function UploadOvf by creating catalog and
// checking if provided chunked ova file uploaded.
func (vcd *TestVCD) Test_UploadOvf_chunked(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaChunkedPath, check)
	checkUploadOvf(vcd, check, vcd.config.OVA.OvaChunkedPath, vcd.config.VCD.Catalog.Name, TestUploadOvf+"2")
}

// Tests System function UploadOvf by creating catalog and
// checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_UploadOvf_progress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)
	itemName := TestUploadOvf + "3"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
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

	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadOvf by creating catalog and
// checking UploadTask.ShowUploadProgress writes values of progress to stdin.
func (vcd *TestVCD) Test_UploadOvf_ShowUploadProgress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)
	itemName := TestUploadOvf + "4"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	//execute
	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)

	//take control of stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = uploadTask.ShowUploadProgress()
	check.Assert(err, IsNil)
	w.Close()
	//read stdin
	result, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	check.Assert(string(result), Matches, "(?s).*Upload progress 100.00%.*")

	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadOvf by creating catalog, creating catalog item
// and expecting specific error then trying to create same catalog item. As vCD returns cryptic error for such case.
func (vcd *TestVCD) Test_UploadOvf_error_withSameItem(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)

	itemName := TestUploadOvf + "5"

	catalog, _ := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	//add item
	uploadTask, err2 := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err2, IsNil)
	err2 = uploadTask.WaitTaskCompletion()
	check.Assert(err2, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	catalog, _ = findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)
	_, err3 := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err3.Error(), Matches, ".*already exists. Upload with different name.*")
}

// Tests System function UploadOvf by creating catalog, uploading file and verifying
// that extracted files were deleted.s
func (vcd *TestVCD) Test_UploadOvf_cleaned_extracted_files(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)

	itemName := TestUploadOvf + "6"

	//check existing count of folders
	oldFolderCount := countFolders()

	catalog, _ := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	check.Assert(oldFolderCount, Equals, countFolders())

}

// Tests System function UploadOvf by creating catalog and
// checking if provided standard ovf file uploaded.
func (vcd *TestVCD) Test_UploadOvfFile(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvfPath, check)
	checkUploadOvf(vcd, check, vcd.config.OVA.OvfPath, vcd.config.VCD.Catalog.Name, TestUploadOvf+"7")
}

// Tests System function UploadOvf by creating catalog and
// checking that ova file without vmdk size specified can be uploaded.
func (vcd *TestVCD) Test_UploadOvf_withoutVMDKSize(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaWithoutSizePath, check)
	checkUploadOvf(vcd, check, vcd.config.OVA.OvaWithoutSizePath, vcd.config.VCD.Catalog.Name, TestUploadOvf+"8")
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
	catalog, org := findCatalog(vcd, check, catalogName)

	uploadTask, err := catalog.UploadOvf(ovaFileName, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+catalogName, "checkUploadOvf")

	catalog, err = org.GetCatalogByName(catalogName, false)
	check.Assert(err, IsNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

func verifyCatalogItemUploaded(check *C, catalog *Catalog, itemName string) {
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

func findCatalog(vcd *TestVCD, check *C, catalogName string) (*Catalog, *AdminOrg) {
	org := getOrg(vcd, check)
	catalog, err := org.GetCatalogByName(catalogName, false)
	check.Assert(err, IsNil)
	return catalog, org
}

func getOrg(vcd *TestVCD, check *C) *AdminOrg {
	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	return org
}

func skipWhenOvaPathMissing(ovaPath string, check *C) {
	if ovaPath == "" {
		check.Skip("Skipping test because no OVA/OVF path given")
	}
}

// Tests System function UploadMediaImage by checking if provided standard iso file uploaded.
func (vcd *TestVCD) Test_CatalogUploadMediaImage(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(TestCatalogUploadMedia, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(TestCatalogUploadMedia, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage")

	//verifyMediaImageUploaded(vcd.vdc.client, check, TestUploadMedia)
	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	verifyCatalogItemUploaded(check, catalog, TestCatalogUploadMedia)
}

// Tests System function UploadMediaImage by checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_progress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

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

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage_progress_works")

	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadMediaImage by checking UploadTask.ShowUploadProgress writes values of progress to stdin.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_ShowUploadProgress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "3"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)

	//take control of stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = uploadTask.ShowUploadProgress()
	check.Assert(err, IsNil)
	w.Close()
	//read stdin
	result, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage_ShowUploadProgress_works")

	check.Assert(string(result), Matches, "(?s).*Upload progress 100.00%.*")
	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadMediaImage by creating media item and expecting specific error
// then trying to create same media item. As vCD returns cryptic error for such case.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_error_withSameItem(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "4"

	catalog, _ := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage_error_withSameItem")
}

// Tests System function Delete by creating media item and
// deleting it after.
func (vcd *TestVCD) Test_CatalogDeleteMediaRecord(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "5"

	catalog, org := findCatalog(vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogDeleteMediaImage")

	mediaRecord, err := catalog.QueryMedia(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	task, err := mediaRecord.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	mediaRecord, err = catalog.QueryMedia(itemName)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(mediaRecord, IsNil)

	//addition check
	// check through existing catalogItems
	catalog, err = org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
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

func init() {
	testingTags["catalog"] = "catalog_test.go"
}

// Tests CatalogItem retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_CatalogGetItem(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_CatalogGetItem: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_CatalogGetItem: Catalog name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return catalog.GetCatalogItemByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) { return catalog.GetCatalogItemById(id, refresh) }
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return catalog.GetCatalogItemByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "Catalog",
		parentName:    vcd.config.VCD.Catalog.Name,
		entityType:    "CatalogItem",
		entityName:    vcd.config.VCD.Catalog.CatalogItem,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

// TestGetVappTemplateByHref tests that we can find a vApp template using
// the HREF from the Entity section of a known Catalog Item
func (vcd *TestVCD) TestGetVappTemplateByHref(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_CatalogGetItem: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_CatalogGetItem: Catalog name not given")
		return
	}
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_CatalogGetItem: Catalog item name not given")
		return
	}

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catalogItem, NotNil)
	check.Assert(catalogItem.CatalogItem.Entity, NotNil)

	vappTemplate, err := catalog.GetVappTemplateByHref(catalogItem.CatalogItem.Entity.HREF)
	check.Assert(err, IsNil)
	check.Assert(vappTemplate, NotNil)
	check.Assert(vappTemplate.VAppTemplate.ID, Not(Equals), catalogItem.CatalogItem.ID)
	check.Assert(vappTemplate.VAppTemplate.Type, Equals, types.MimeVAppTemplate)
	check.Assert(vappTemplate.VAppTemplate.Name, Equals, catalogItem.CatalogItem.Name)
}

// Test_GetCatalogByNameSharedCatalog creates a separate Org and VDC just to create Catalog and share it with main Org
// One should be able to find shared catalogs from different Organizations
func (vcd *TestVCD) Test_GetCatalogByNameSharedCatalog(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	newOrg1, vdc, sharedCatalog := createSharedCatalogInNewOrg(vcd, check, check.TestName())

	// Try to find the catalog inside Org which owns it - newOrg1
	catalogByName, err := newOrg1.GetCatalogByName(sharedCatalog.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalogByName.Catalog.Name, Equals, sharedCatalog.Catalog.Name)

	// Try to find the catalog in another Org with which this catalog is shared (vcd.Org)
	sharedCatalogByName, err := vcd.org.GetCatalogByName(sharedCatalog.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(sharedCatalogByName.Catalog.Name, Equals, sharedCatalog.Catalog.Name)

	cleanupCatalogOrgVdc(check, sharedCatalog, vdc, vcd, newOrg1)
}

// Test_GetCatalogByIdSharedCatalog creates a separate Org and VDC just to create Catalog and share it with main Org
// One should be able to find shared catalogs from different Organizations
func (vcd *TestVCD) Test_GetCatalogByIdSharedCatalog(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	newOrg1, vdc, sharedCatalog := createSharedCatalogInNewOrg(vcd, check, check.TestName())

	// Try to find the sharedCatalog inside Org which owns it - newOrg1
	catalogById, err := newOrg1.GetCatalogById(sharedCatalog.Catalog.ID, true)
	check.Assert(err, IsNil)
	check.Assert(catalogById.Catalog.Name, Equals, sharedCatalog.Catalog.Name)

	// Try to find the sharedCatalog in another Org with which this sharedCatalog is shared (vcd.Org)
	sharedCatalogById, err := vcd.org.GetCatalogById(sharedCatalog.Catalog.ID, false)
	check.Assert(err, IsNil)
	check.Assert(sharedCatalogById.Catalog.Name, Equals, sharedCatalog.Catalog.Name)

	cleanupCatalogOrgVdc(check, sharedCatalog, vdc, vcd, newOrg1)
}

// Test_GetCatalogByNamePrefersLocal tests that local catalog (in the same Org) is prioritised against shared catalogs
// in other Orgs. It does so by creating another Org with shared Catalog named just like the one in testing catalog
func (vcd *TestVCD) Test_GetCatalogByNamePrefersLocal(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// Create a catalog  in new org with exactly the same name as in vcd.Org
	newOrg1, vdc, sharedCatalog := createSharedCatalogInNewOrg(vcd, check, vcd.config.VCD.Catalog.Name)

	// Make sure that the Owner Org HREF is the local one for vcd.Org catalog named vcd.config.VCD.Catalog.Name
	catalogByNameInTestOrg, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalogByNameInTestOrg.parent.orgName(), Equals, vcd.org.Org.Name)

	// Make sure that the Owner Org HREF is the local one for vcd.Org catalog named vcd.config.VCD.Catalog.Name
	catalogByNameInNewOrg, err := newOrg1.GetCatalogByName(vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalogByNameInNewOrg.parent.orgName(), Equals, newOrg1.Org.Name)

	cleanupCatalogOrgVdc(check, sharedCatalog, vdc, vcd, newOrg1)
}

// Test_GetCatalogByNameSharedCatalogOrgUser additionally tests GetOrgByName and GetOrgById using a custom created Org
// Admin user. It tests the following cases:
// * System user must be able to retrieve any catalog - shared or unshared from another Org
// * Org Admin user must be able to retrieve catalog in his own Org
// * Org Admin user must be able to retrieve shared catalog from another Org
// * Org admin user must not be able to retrieve unshared catalog from another Org
func (vcd *TestVCD) Test_GetCatalogByXSharedCatalogOrgUser(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	newOrg1, vdc, sharedCatalog := createSharedCatalogInNewOrg(vcd, check, check.TestName())

	// Create one more additional catalog which is not shared
	unsharedCatalog, err := newOrg1.CreateCatalog("unshared-catalog", check.TestName())
	check.Assert(err, IsNil)
	AddToCleanupList(unsharedCatalog.Catalog.Name, "catalog", newOrg1.Org.Name, check.TestName())

	// Try to find the catalog inside Org which owns it - newOrg1
	catalogByName, err := newOrg1.GetCatalogByName(sharedCatalog.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalogByName.Catalog.Name, Equals, sharedCatalog.Catalog.Name)

	// Try to find the catalog in another Org with which this catalog is shared (vcd.Org)
	sharedCatalogByName, err := vcd.org.GetCatalogByName(sharedCatalog.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(sharedCatalogByName.Catalog.Name, Equals, sharedCatalog.Catalog.Name)

	// Try to find unshared catalog from another Org with System user
	systemUnsharedCatalogByName, err := vcd.org.GetCatalogByName(unsharedCatalog.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(systemUnsharedCatalogByName.Catalog.ID, Equals, unsharedCatalog.Catalog.ID)

	// Create an Org Admin user and test that it can find catalog as well
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	orgAdminClient, err := newOrgUserConnection(adminOrg, "test-user", "CHANGE-ME", vcd.config.Provider.Url, true)
	check.Assert(err, IsNil)
	orgAsOrgUser, err := orgAdminClient.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	// Find a catalog in the same Org using Org Admin user
	orgAdminCatalogByNameSameOrg, err := orgAsOrgUser.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(orgAdminCatalogByNameSameOrg.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	orgAdminCatalogByIdSameOrg, err := orgAsOrgUser.GetCatalogById(orgAdminCatalogByNameSameOrg.Catalog.ID, false)
	check.Assert(err, IsNil)
	check.Assert(orgAdminCatalogByIdSameOrg.Catalog.Name, Equals, orgAdminCatalogByNameSameOrg.Catalog.Name)
	check.Assert(orgAdminCatalogByIdSameOrg.Catalog.ID, Equals, orgAdminCatalogByNameSameOrg.Catalog.ID)

	// Find a shared catalog from another Org using Org Admin user
	orgAdminCatalogByName, err := orgAsOrgUser.GetCatalogByName(sharedCatalog.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(orgAdminCatalogByName.Catalog.Name, Equals, sharedCatalog.Catalog.Name)
	check.Assert(orgAdminCatalogByName.Catalog.ID, Equals, sharedCatalog.Catalog.ID)

	orgAdminCatalogById, err := orgAsOrgUser.GetCatalogById(sharedCatalog.Catalog.ID, false)
	check.Assert(err, IsNil)
	check.Assert(orgAdminCatalogById.Catalog.Name, Equals, sharedCatalog.Catalog.Name)
	check.Assert(orgAdminCatalogById.Catalog.ID, Equals, sharedCatalog.Catalog.ID)

	// Try to find unshared catalog from another Org with Org admin user and expect an ErrorEntityNotFound
	_, err = orgAsOrgUser.GetCatalogByName(unsharedCatalog.Catalog.Name, true)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = orgAsOrgUser.GetCatalogById(unsharedCatalog.Catalog.ID, true)
	check.Assert(ContainsNotFound(err), Equals, true)

	// Cleanup
	err = unsharedCatalog.Delete(true, true)
	check.Assert(err, IsNil)

	cleanupCatalogOrgVdc(check, sharedCatalog, vdc, vcd, newOrg1)
}

func createSharedCatalogInNewOrg(vcd *TestVCD, check *C, newCatalogName string) (*Org, *Vdc, Catalog) {
	newOrgName1 := spawnTestOrg(vcd, check, "org")

	newOrg1, err := vcd.client.GetOrgByName(newOrgName1)
	check.Assert(err, IsNil)

	// Spawn a VDC inside newly created Org so that there is storage to create new catalog
	vdc := spawnTestVdc(vcd, check, newOrgName1)

	catalog, err := newOrg1.CreateCatalog(newCatalogName, "Catalog for testing")
	check.Assert(err, IsNil)
	AddToCleanupList(newCatalogName, "catalog", newOrgName1, check.TestName())

	// Share new Catalog in newOrgName1 with default test Org vcd.Org
	readOnly := "ReadOnly"
	accessControl := &types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: &readOnly,
		AccessSettings: &types.AccessSettingList{
			AccessSetting: []*types.AccessSetting{&types.AccessSetting{
				Subject: &types.LocalSubject{
					HREF: vcd.org.Org.HREF,
					Name: vcd.org.Org.Name,
					Type: types.MimeOrg,
				},
				AccessLevel: "ReadOnly",
			}},
		},
	}
	err = catalog.SetAccessControl(accessControl, false)
	check.Assert(err, IsNil)

	return newOrg1, vdc, catalog
}

func cleanupCatalogOrgVdc(check *C, sharedCatalog Catalog, vdc *Vdc, vcd *TestVCD, newOrg1 *Org) {
	// Cleanup catalog, vdc and org
	err := sharedCatalog.Delete(true, true)
	check.Assert(err, IsNil)

	err = vdc.DeleteWait(true, true)
	check.Assert(err, IsNil)

	adminOrg, err := vcd.client.GetAdminOrgByName(newOrg1.Org.Name)
	check.Assert(err, IsNil)
	err = adminOrg.Delete(true, true)
	check.Assert(err, IsNil)
}
