//go:build catalog || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
	"os"
	"path"
	"runtime"
)

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

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, check.TestName())

	media, err := catalog.GetMediaByName(itemName, true)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)
	check.Assert(media.Media.Name, Equals, itemName)

	task, err := media.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	_, err = catalog.GetMediaByName(itemName, true)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
}

func (vcd *TestVCD) Test_UploadAnyMediaFile(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_UploadAnyMediaFile: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_UploadAnyMediaFile: Catalog name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	_, sourceFile, _, _ := runtime.Caller(0)
	sourceFile = path.Clean(sourceFile)
	itemName := check.TestName()
	itemPath := sourceFile

	// Upload the source file of the current test as a media item
	uploadTask, err := catalog.UploadMediaFile(itemName, "Text file uploaded from test", itemPath, 1024, false)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, check.TestName())

	// Retrieve the media item
	media, err := catalog.GetMediaByName(itemName, true)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)
	check.Assert(media.Media.Name, Equals, itemName)

	// Repeat the download a few times. Make sure that a repeated download works as well as the first one
	for i := 0; i < 2; i++ {
		// Download the media item from VCD as a byte slice
		contents, err := media.Download()
		check.Assert(err, IsNil)
		check.Assert(len(contents), Not(Equals), 0)
		check.Assert(media.Media.Files, NotNil)
		check.Assert(media.Media.Files.File, NotNil)
		check.Assert(media.Media.Files.File[0].Name, Not(Equals), "")
		check.Assert(len(media.Media.Files.File[0].Link), Not(Equals), 0)

		// Read the source file from disk
		fromFile, err := os.ReadFile(path.Clean(sourceFile))
		check.Assert(err, IsNil)
		// Make sure that what we downloaded from VCD corresponds to the file contents.
		check.Assert(len(fromFile), Equals, len(contents))
		check.Assert(fromFile, DeepEquals, contents)
	}

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
	if vcd.config.Media.Media == "" {
		check.Skip("Test_RefreshMediaRecord: Media name not given")
		return
	}
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := vcd.config.Media.Media

	media, err := catalog.GetMediaByName(itemName, true)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)
	check.Assert(media.Media.Name, Equals, itemName)

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
		return catalog.GetMediaById(id)
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

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_QueryMedia: Org name not given")
		return
	}
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Test_QueryMedia: Catalog name not given")
		return
	}
	if vcd.config.Media.Media == "" {
		check.Skip("Test_RefreshMediaRecord: Media name not given")
		return
	}
	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	testQueryMediaName := vcd.config.Media.Media

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

	if vcd.config.Media.Media == "" {
		check.Skip("Test_RefreshMediaRecord: Media name not given")
		return
	}

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := vcd.config.Media.Media

	mediaRecord, err := catalog.QueryMedia(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	oldMediaRecord := mediaRecord
	err = mediaRecord.Refresh()
	check.Assert(err, IsNil)

	check.Assert(mediaRecord, NotNil)
	check.Assert(oldMediaRecord.MediaRecord.ID, Equals, mediaRecord.MediaRecord.ID)
	check.Assert(oldMediaRecord.MediaRecord.Name, Equals, mediaRecord.MediaRecord.Name)
	check.Assert(oldMediaRecord.MediaRecord.HREF, Equals, mediaRecord.MediaRecord.HREF)
}

func (vcd *TestVCD) Test_QueryAllMedia(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_QueryMedia: Org name not given")
		return
	}

	if vcd.config.Media.Media == "" {
		check.Skip("Test_RefreshMediaRecord: Media name not given")
		return
	}
	// Fetching organization
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	testQueryMediaName := vcd.config.Media.Media

	mediaList, err := vcd.vdc.QueryAllMedia(testQueryMediaName)
	check.Assert(err, IsNil)
	check.Assert(mediaList, Not(Equals), nil)

	check.Assert(mediaList[0].MediaRecord.Name, Equals, testQueryMediaName)
	check.Assert(mediaList[0].MediaRecord.HREF, Not(Equals), "")

	// find Invalid media
	mediaList, err = vcd.vdc.QueryAllMedia("INVALID")
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(mediaList, IsNil)
}
