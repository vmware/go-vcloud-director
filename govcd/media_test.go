// +build catalog functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"io/ioutil"
	"os"

	. "gopkg.in/check.v1"
)

// Tests System function UploadMediaImage by checking if provided standard iso file uploaded.
func (vcd *TestVCD) Test_UploadMediaImage(check *C) {
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
	results, err := queryMediaItemsWithFilter(vdc, "name=="+itemName)

	check.Assert(err, Equals, nil)
	check.Assert(len(results), Equals, 1)
}

// Tests System function UploadMediaImage by checking UploadTask.GetUploadProgress returns values of progress.
func (vcd *TestVCD) Test_UploadMediaImage_progress_works(check *C) {
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
	skipWhenMediaPathMissing(vcd, check)
	itemName := TestUploadMedia + "3"

	uploadTask, err := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
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

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

	check.Assert(string(result), Matches, ".*Upload progress 100.00%")
	verifyMediaImageUploaded(vcd.vdc, check, itemName)
}

// Tests System function UploadMediaImage by creating media item and expecting specific error
// then trying to create same media item. As vCD returns cryptic error for such case.
func (vcd *TestVCD) Test_UploadMediaImage_error_withSameItem(check *C) {
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
func (vcd *TestVCD) Test_DeleteMediaImage(check *C) {
	skipWhenMediaPathMissing(vcd, check)
	itemName := TestUploadMedia + "5"

	uploadTask, err := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

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

}

// Tests System function FindMediaAsCatalogItem by creating media item and
// and finding it as catalog item after.
func (vcd *TestVCD) Test_FindMediaAsCatalogItem(check *C) {
	skipWhenMediaPathMissing(vcd, check)
	itemName := TestUploadMedia + "6"

	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

	err = vcd.org.Refresh()
	check.Assert(err, IsNil)

	catalogItem, err := FindMediaAsCatalogItem(vcd.org, vcd.config.VCD.Catalog.Name, itemName)
	check.Assert(err, IsNil)
	check.Assert(catalogItem, Not(Equals), CatalogItem{})
	check.Assert(catalogItem.CatalogItem.Name, Equals, itemName)

}

// Tests System function Refresh
func (vcd *TestVCD) Test_RefreshMediaImage(check *C) {
	skipWhenMediaPathMissing(vcd, check)
	itemName := "TestRefreshMedia"

	uploadTask, err := vcd.vdc.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	check.Assert(uploadTask, NotNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaImage", vcd.org.Org.Name+"|"+vcd.vdc.Vdc.Name, "Test_UploadMediaImage")

	mediaItem, err := vcd.vdc.FindMediaImage(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaItem, NotNil)
	check.Assert(mediaItem, Not(Equals), MediaItem{})

	oldMediaItem := mediaItem
	mediaItem.Refresh()

	check.Assert(mediaItem, NotNil)
	check.Assert(oldMediaItem.MediaItem.ID, Equals, mediaItem.MediaItem.ID)
	check.Assert(oldMediaItem.MediaItem.Name, Equals, mediaItem.MediaItem.Name)
	check.Assert(oldMediaItem.MediaItem.HREF, Equals, mediaItem.MediaItem.HREF)
}
