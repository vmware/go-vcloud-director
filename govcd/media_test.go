// +build catalog functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"io/ioutil"
	"os"

	. "gopkg.in/check.v1"
)

// Tests System function UploadMediaImage by checking if provided standard iso file uploaded.
func (vcd *TestVCD) Test_UploadMediaImage(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)

	uploadTask, err := vcd.vdc.UploadMediaImage(TestUploadMedia, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(TestUploadMedia, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

	verifyMediaImageUploaded(vcd.vdc, check, TestUploadMedia)
}

func skipWhenMediaPathMissing(vcd *TestVCD, check *C) {
	if vcd.config.Media.MediaPath == "" {
		check.Skip("Skipping test because no iso path given")
	}
}

func verifyMediaImageUploaded(vdc *Vdc, check *C, itemName string) {
	results, err := queryMediaWithFilter(vdc, "name=="+itemName)

	check.Assert(err, Equals, nil)
	check.Assert(len(results), Equals, 1)
}

// Tests System function UploadMediaImage by checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_UploadMediaImage_progress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestUploadMedia + "2"

	uploadTask, err := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

	verifyMediaImageUploaded(vcd.vdc, check, itemName)
}

// Tests System function UploadMediaImage by checking UploadTask.ShowUploadProgress writes values of progress to stdin.
func (vcd *TestVCD) Test_UploadMediaImage_ShowUploadProgress_works(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestUploadMedia + "3"

	uploadTask, err := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

	check.Assert(string(result), Matches, ".*Upload progress 100.00%")
	verifyMediaImageUploaded(vcd.vdc, check, itemName)
}

// Tests System function UploadMediaImage by creating media item and expecting specific error
// then trying to create same media item. As vCD returns cryptic error for such case.
func (vcd *TestVCD) Test_UploadMediaImage_error_withSameItem(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	itemName := TestUploadMedia + "4"

	uploadTask, err := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

	_, err2 := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err2, NotNil)
	check.Assert(err2.Error(), Matches, ".*already exists. Upload with different name.*")
}

// Tests System function Delete by creating media item and
// deleting it after.
func (vcd *TestVCD) Test_DeleteMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_DeleteMedia: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_DeleteMedia: Catalog name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := "TestDeleteMedia"
	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_DeleteMediaImage")

	media, err := catalog.GetMediaByName(itemName, true)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)
	check.Assert(media.Media.Name, Not(Equals), itemName)

	task, err := media.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	_, err = catalog.GetMediaByName(itemName, true)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
}

// Tests System function Refresh
func (vcd *TestVCD) Test_RefreshMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_RefreshMedia: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_RefreshMedia: Catalog name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := "TestRefreshMedia"
	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_RefreshMediaImage")

	media, err := catalog.GetMediaByName(itemName, true)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)
	check.Assert(media.Media.Name, Not(Equals), itemName)

	oldMedia := media
	err = media.Refresh()
	check.Assert(err, IsNil)

	check.Assert(media, NotNil)
	check.Assert(oldMedia.Media.ID, Equals, media.Media.ID)
	check.Assert(oldMedia.Media.Name, Equals, media.Media.Name)
	check.Assert(oldMedia.Media.HREF, Equals, media.Media.HREF)
}

// Tests Media retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) Test_GetMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_GetMedia: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_GetMedia: Catalog name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		err = catalog.Refresh()
		check.Assert(err, IsNil)
		return catalog.GetMediaByName(name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) {
		return catalog.GetMediaById(id, refresh)
	}
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return catalog.GetMediaByNameOrId(id, refresh)
	}

	var def = getterTestDefinition{
		parentType:    "Catalog",
		parentName:    vcd.config.VCD.Catalog.Name,
		entityType:    "Media",
		entityName:    vcd.config.Media.Media,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

func (vcd *TestVCD) Test_QueryMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)

	testQueryMediaName := "testQueryMedia"
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_QueryMedia: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_QueryMedia: Catalog name not given")
		return
	}

	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	uploadTask, err := catalog.UploadMediaImage(testQueryMediaName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(testQueryMediaName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_RefreshMediaImage")

	media, err := catalog.QueryMedia(testQueryMediaName)
	check.Assert(err, IsNil)
	check.Assert(media, Not(Equals), nil)

	check.Assert(media.MediaRecord.Name, Equals, testQueryMediaName)
	check.Assert(media.MediaRecord.HREF, Not(Equals), "")

	// find Invalid media
	media, err = catalog.QueryMedia("INVALID")
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(media, IsNil)
}

// Tests System function Refresh
func (vcd *TestVCD) Test_RefreshMediaRecord(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)
	if vcd.config.VCD.Org == "" {
		check.Skip("Test_RefreshMediaRecord: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_RefreshMediaRecord: Catalog name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := "TestRefreshMediaRecord"
	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_RefreshMediaImage")

	mediaRecord, err := catalog.QueryMedia(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Not(Equals), itemName)

	oldMediaRecord := mediaRecord
	err = mediaRecord.Refresh()
	check.Assert(err, IsNil)

	check.Assert(mediaRecord, NotNil)
	check.Assert(oldMediaRecord.MediaRecord.ID, Equals, mediaRecord.MediaRecord.ID)
	check.Assert(oldMediaRecord.MediaRecord.Name, Equals, mediaRecord.MediaRecord.Name)
	check.Assert(oldMediaRecord.MediaRecord.HREF, Equals, mediaRecord.MediaRecord.HREF)
}
