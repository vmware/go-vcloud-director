// +build catalog functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
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
	ctx := context.Background()

	cat, err := vcd.org.GetCatalogByName(ctx, catalogName, false)
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
	err = cat.Refresh(ctx)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)
	check.Assert(cat.Catalog.Name, Equals, catalogName)
	check.Assert(cat.Catalog.DateCreated, Equals, dateCreated)
	check.Assert(len(cat.Catalog.CatalogItems), Equals, numItems)
	check.Assert(cat.Catalog.ID, Equals, catalogId)
}

func (vcd *TestVCD) Test_FindCatalogItem(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	ctx := context.Background()

	// Fetch Catalog
	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_FindCatalogItem: Catalog not found. Test can't proceed")
		return
	}

	// Find Catalog Item
	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_FindCatalogItem: Catalog Item not given. Test can't proceed")
	}
	catalogItem, err := cat.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catalogItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)
	// If given a description in config file then it checks if the descriptions match
	// Otherwise it skips the assert
	if vcd.config.VCD.Catalog.CatalogItemDescription != "" {
		check.Assert(catalogItem.CatalogItem.Description, Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
	// Test non-existent catalog item
	catalogItem, err = cat.GetCatalogItemByName(ctx, "INVALID", false)
	check.Assert(err, NotNil)
	check.Assert(catalogItem, IsNil)
}

// Creates a Catalog, updates the description, and checks the changes against the
// newly updated catalog. Then deletes the catalog
func (vcd *TestVCD) Test_UpdateCatalog(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	ctx := context.Background()

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	catalog, _ := org.GetAdminCatalogByName(ctx, TestUpdateCatalog, false)
	if catalog != nil {
		err = catalog.Delete(ctx, true, true)
		check.Assert(err, IsNil)
	}
	adminCatalog, err := org.CreateCatalog(ctx, TestUpdateCatalog, TestUpdateCatalog)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestUpdateCatalog, "catalog", vcd.config.VCD.Org, "Test_UpdateCatalog")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, TestUpdateCatalog)

	adminCatalog.AdminCatalog.Description = TestCreateCatalogDesc
	err = adminCatalog.Update(ctx)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog.AdminCatalog.Description, Equals, TestCreateCatalogDesc)

	err = adminCatalog.Delete(ctx, true, true)
	check.Assert(err, IsNil)
}

// Creates a Catalog, and then deletes the catalog, and checks if
// the catalog still exists. If it does the assertion fails.
func (vcd *TestVCD) Test_DeleteCatalog(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	ctx := context.Background()

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)
	oldCatalog, _ := org.GetAdminCatalogByName(ctx, TestDeleteCatalog, false)
	if oldCatalog != nil {
		err = oldCatalog.Delete(ctx, true, true)
		check.Assert(err, IsNil)
	}
	adminCatalog, err := org.CreateCatalog(ctx, TestDeleteCatalog, TestDeleteCatalog)
	check.Assert(err, IsNil)
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(TestDeleteCatalog, "catalog", vcd.config.VCD.Org, "Test_DeleteCatalog")
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, TestDeleteCatalog)
	err = adminCatalog.Delete(ctx, true, true)
	check.Assert(err, IsNil)
	doesCatalogExist(ctx, check, org)
}

func doesCatalogExist(ctx context.Context, check *C, org *AdminOrg) {
	var err error
	var catalog *AdminCatalog
	for i := 0; i < 30; i++ {
		catalog, err = org.GetAdminCatalogByName(ctx, TestDeleteCatalog, true)
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
	checkUploadOvf(context.Background(), vcd, check, vcd.config.OVA.OvaPath, vcd.config.VCD.Catalog.Name, TestUploadOvf)
}

// Tests System function UploadOvf by creating catalog and
// checking if provided chunked ova file uploaded.
func (vcd *TestVCD) Test_UploadOvf_chunked(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaChunkedPath, check)
	checkUploadOvf(context.Background(), vcd, check, vcd.config.OVA.OvaChunkedPath, vcd.config.VCD.Catalog.Name, TestUploadOvf+"2")
}

// Tests System function UploadOvf by creating catalog and
// checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_UploadOvf_progress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	ctx := context.Background()

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)
	itemName := TestUploadOvf + "3"

	catalog, org := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(ctx, vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	for {
		if value := uploadTask.GetUploadProgress(); value == "100.00" {
			break
		} else {
			check.Assert(value, Not(Equals), "")
		}
	}
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	catalog, err = org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadOvf by creating catalog and
// checking UploadTask.ShowUploadProgress writes values of progress to stdin.
func (vcd *TestVCD) Test_UploadOvf_ShowUploadProgress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	ctx := context.Background()

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)
	itemName := TestUploadOvf + "4"

	catalog, org := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	//execute
	uploadTask, err := catalog.UploadOvf(ctx, vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)

	//take control of stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = uploadTask.ShowUploadProgress(ctx)
	check.Assert(err, IsNil)
	w.Close()
	//read stdin
	result, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	check.Assert(string(result), Matches, "(?s).*Upload progress 100.00%.*")

	catalog, err = org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, true)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadOvf by creating catalog, creating catalog item
// and expecting specific error then trying to create same catalog item. As vCD returns cryptic error for such case.
func (vcd *TestVCD) Test_UploadOvf_error_withSameItem(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	ctx := context.Background()

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)

	itemName := TestUploadOvf + "5"

	catalog, _ := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	//add item
	uploadTask, err2 := catalog.UploadOvf(ctx, vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err2, IsNil)
	err2 = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err2, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	catalog, _ = findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)
	_, err3 := catalog.UploadOvf(ctx, vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err3.Error(), Matches, ".*already exists. Upload with different name.*")
}

// Tests System function UploadOvf by creating catalog, uploading file and verifying
// that extracted files were deleted.s
func (vcd *TestVCD) Test_UploadOvf_cleaned_extracted_files(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	ctx := context.Background()

	skipWhenOvaPathMissing(vcd.config.OVA.OvaPath, check)

	itemName := TestUploadOvf + "6"

	//check existing count of folders
	oldFolderCount := countFolders()

	catalog, _ := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(ctx, vcd.config.OVA.OvaPath, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	check.Assert(oldFolderCount, Equals, countFolders())

}

// Tests System function UploadOvf by creating catalog and
// checking if provided standard ovf file uploaded.
func (vcd *TestVCD) Test_UploadOvfFile(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvfPath, check)
	checkUploadOvf(context.Background(), vcd, check, vcd.config.OVA.OvfPath, vcd.config.VCD.Catalog.Name, TestUploadOvf+"7")
}

// Tests System function UploadOvf by creating catalog and
// checking that ova file without vmdk size specified can be uploaded.
func (vcd *TestVCD) Test_UploadOvf_withoutVMDKSize(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenOvaPathMissing(vcd.config.OVA.OvaWithoutSizePath, check)
	checkUploadOvf(context.Background(), vcd, check, vcd.config.OVA.OvaWithoutSizePath, vcd.config.VCD.Catalog.Name, TestUploadOvf+"8")
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

func checkUploadOvf(ctx context.Context, vcd *TestVCD, check *C, ovaFileName, catalogName, itemName string) {
	catalog, org := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadOvf(ctx, ovaFileName, itemName, "upload from test", 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "catalogItem", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_UploadOvf")

	catalog, err = org.GetCatalogByName(ctx, catalogName, false)
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

func findCatalog(ctx context.Context, vcd *TestVCD, check *C, catalogName string) (*Catalog, *AdminOrg) {
	org := getOrg(ctx, vcd, check)
	catalog, err := org.GetCatalogByName(ctx, catalogName, false)
	check.Assert(err, IsNil)
	return catalog, org
}

func getOrg(ctx context.Context, vcd *TestVCD, check *C) *AdminOrg {
	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
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

	ctx := context.Background()
	catalog, org := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(ctx, TestCatalogUploadMedia, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(TestCatalogUploadMedia, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage")

	//verifyMediaImageUploaded(vcd.vdc.client, check, TestUploadMedia)
	catalog, err = org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	verifyCatalogItemUploaded(check, catalog, TestCatalogUploadMedia)
}

// Tests System function UploadMediaImage by checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_progress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "2"

	ctx := context.Background()
	catalog, org := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	for {
		if value := uploadTask.GetUploadProgress(); value == "100.00" {
			break
		} else {
			check.Assert(value, Not(Equals), "")
		}
	}
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage_progress_works")

	catalog, err = org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	verifyCatalogItemUploaded(check, catalog, itemName)
}

// Tests System function UploadMediaImage by checking UploadTask.ShowUploadProgress writes values of progress to stdin.
func (vcd *TestVCD) Test_CatalogUploadMediaImage_ShowUploadProgress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "3"

	ctx := context.Background()
	catalog, org := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)

	//take control of stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = uploadTask.ShowUploadProgress(ctx)
	check.Assert(err, IsNil)
	w.Close()
	//read stdin
	result, _ := ioutil.ReadAll(r)
	os.Stdout = oldStdout

	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage_ShowUploadProgress_works")

	check.Assert(string(result), Matches, "(?s).*Upload progress 100.00%.*")
	catalog, err = org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
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

	ctx := context.Background()
	catalog, _ := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogUploadMediaImage_error_withSameItem")
}

// Tests System function Delete by creating media item and
// deleting it after.
func (vcd *TestVCD) Test_CatalogDeleteMediaRecord(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestCatalogUploadMedia + "5"

	ctx := context.Background()
	catalog, org := findCatalog(ctx, vcd, check, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_CatalogDeleteMediaImage")

	mediaRecord, err := catalog.QueryMedia(ctx, itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	task, err := mediaRecord.Delete(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	mediaRecord, err = catalog.QueryMedia(ctx, itemName)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(mediaRecord, IsNil)

	//addition check
	// check through existing catalogItems
	catalog, err = org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
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
	ctx := context.Background()
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		return catalog.GetCatalogItemByName(ctx, name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) {
		return catalog.GetCatalogItemById(ctx, id, refresh)
	}
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return catalog.GetCatalogItemByNameOrId(ctx, id, refresh)
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

	ctx := context.Background()
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	catalogItem, err := catalog.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catalogItem, NotNil)
	check.Assert(catalogItem.CatalogItem.Entity, NotNil)

	vappTemplate, err := catalog.GetVappTemplateByHref(ctx, catalogItem.CatalogItem.Entity.HREF)
	check.Assert(err, IsNil)
	check.Assert(vappTemplate, NotNil)
	check.Assert(vappTemplate.VAppTemplate.ID, Not(Equals), catalogItem.CatalogItem.ID)
	check.Assert(vappTemplate.VAppTemplate.Type, Equals, types.MimeVAppTemplate)
	check.Assert(vappTemplate.VAppTemplate.Name, Equals, catalogItem.CatalogItem.Name)
}
