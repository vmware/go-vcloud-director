// +build catalog functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
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
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := "TestDeleteMedia"
	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_DeleteMediaImage")

	media, err := catalog.GetMediaByName(ctx, itemName, true)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)
	check.Assert(media.Media.Name, Equals, itemName)

	task, err := media.Delete(ctx)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	_, err = catalog.GetMediaByName(ctx, itemName, true)
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
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := vcd.config.Media.Media

	media, err := catalog.GetMediaByName(ctx, itemName, true)
	check.Assert(err, IsNil)
	check.Assert(media, NotNil)
	check.Assert(media.Media.Name, Equals, itemName)

	oldMedia := media
	err = media.Refresh(ctx)
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
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	getByName := func(name string, refresh bool) (genericEntity, error) {
		err = catalog.Refresh(ctx)
		check.Assert(err, IsNil)
		return catalog.GetMediaByName(ctx, name, refresh)
	}
	getById := func(id string, refresh bool) (genericEntity, error) {
		return catalog.GetMediaById(ctx, id)
	}
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		return catalog.GetMediaByNameOrId(ctx, id, refresh)
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
	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)

	testQueryMediaName := vcd.config.Media.Media

	media, err := catalog.QueryMedia(ctx, testQueryMediaName)
	check.Assert(err, IsNil)
	check.Assert(media, Not(Equals), nil)

	check.Assert(media.MediaRecord.Name, Equals, testQueryMediaName)
	check.Assert(media.MediaRecord.HREF, Not(Equals), "")

	// find Invalid media
	media, err = catalog.QueryMedia(ctx, "INVALID")
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

	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	itemName := vcd.config.Media.Media

	mediaRecord, err := catalog.QueryMedia(ctx, itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	oldMediaRecord := mediaRecord
	err = mediaRecord.Refresh(ctx)
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
	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	testQueryMediaName := vcd.config.Media.Media

	mediaList, err := vcd.vdc.QueryAllMedia(ctx, testQueryMediaName)
	check.Assert(err, IsNil)
	check.Assert(mediaList, Not(Equals), nil)

	check.Assert(mediaList[0].MediaRecord.Name, Equals, testQueryMediaName)
	check.Assert(mediaList[0].MediaRecord.HREF, Not(Equals), "")

	// find Invalid media
	mediaList, err = vcd.vdc.QueryAllMedia(ctx, "INVALID")
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(mediaList, IsNil)
}
