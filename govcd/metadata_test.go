// +build vapp vdc metadata functional ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

func init() {
	testingTags["metadata"] = "metadata_test.go"
}

func (vcd *TestVCD) Test_AddMetadataForVdc(check *C) {
	if vcd.config.VCD.Vdc == "" {
		check.Skip("skipping test because VDC name is empty")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	// Add metadata
	vdc, err := vcd.vdc.AddMetadata(ctx, "key", "value")
	check.Assert(err, IsNil)
	check.Assert(vdc, Not(Equals), Vdc{})

	AddToCleanupList("key", "vdcMetaData", vcd.org.Org.Name+"|"+vcd.config.VCD.Vdc, check.TestName())

	// Check if metadata was added correctly
	metadata, err := vcd.vdc.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 1)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")
}

func (vcd *TestVCD) Test_DeleteMetadataForVdc(check *C) {
	if vcd.config.VCD.Vdc == "" {
		check.Skip("skipping test because VDC name is empty")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	// Add metadata
	vdc, err := vcd.vdc.AddMetadata(ctx, "key", "value")
	check.Assert(err, IsNil)
	check.Assert(vdc, Not(Equals), Vdc{})

	AddToCleanupList("key", "vdcMetaData", vcd.org.Org.Name+"|"+vcd.config.VCD.Vdc, check.TestName())

	// Remove metadata
	vdc, err = vcd.vdc.DeleteMetadata(ctx, "key2")
	check.Assert(err, IsNil)
	check.Assert(vdc, Not(Equals), Vdc{})

	metadata, err := vcd.vdc.GetMetadata(ctx)
	check.Assert(err, IsNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: 'key2'")
		}
	}
}

func (vcd *TestVCD) Test_AddMetadataOnVapp(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	// Add metadata
	task, err := vcd.vapp.AddMetadata(ctx, "key", "value")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Check if metadata was added correctly
	metadata, err := vcd.vapp.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 1)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")
}

func (vcd *TestVCD) Test_DeleteMetadataOnVapp(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	// Add metadata
	task, err := vcd.vapp.AddMetadata(ctx, "key2", "value2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Remove metadata
	task, err = vcd.vapp.DeleteMetadata(ctx, "key2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	metadata, err := vcd.vapp.GetMetadata(ctx)
	check.Assert(err, IsNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: 'key2'")
		}
	}
}

func (vcd *TestVCD) Test_AddMetadataOnVm(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Add metadata
	task, err := vm.AddMetadata(ctx, "key", "value")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Check if metadata was added correctly
	metadata, err := vm.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 1)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")
}

func (vcd *TestVCD) Test_DeleteMetadataOnVm(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp(ctx)
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Add metadata
	task, err := vm.AddMetadata(ctx, "key2", "value2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Remove metadata
	task, err = vm.DeleteMetadata(ctx, "key2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	metadata, err := vm.GetMetadata(ctx)
	check.Assert(err, IsNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: 'key2'")
		}
	}
}

func (vcd *TestVCD) Test_DeleteMetadataOnVAppTemplate(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog not found. Test can't proceed")
		return
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	vAppTemplate, err := catItem.GetVAppTemplate(ctx)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Add metadata
	_, err = vAppTemplate.AddMetadata(ctx, "key2", "value2")
	check.Assert(err, IsNil)

	// Remove metadata
	err = vAppTemplate.DeleteMetadata(ctx, "key2")
	check.Assert(err, IsNil)

	metadata, err := vAppTemplate.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: 'key2'")
		}
	}
}

func (vcd *TestVCD) Test_AddMetadataOnVAppTemplate(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog not found. Test can't proceed")
		return
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	vAppTemplate, err := catItem.GetVAppTemplate(ctx)
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Check how much metaData exist
	metadata, err := vAppTemplate.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	existingMetaDataCount := len(metadata.MetadataEntry)

	// Add metadata
	_, err = vAppTemplate.AddMetadata(ctx, "key", "value")
	check.Assert(err, IsNil)

	// Check if metadata was added correctly
	metadata, err = vAppTemplate.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount+1)
	var foundEntry *types.MetadataEntry
	for _, entry := range metadata.MetadataEntry {
		if entry.Key == "key" {
			foundEntry = entry
		}
	}
	check.Assert(foundEntry, NotNil)
	check.Assert(foundEntry.Key, Equals, "key")
	check.Assert(foundEntry.TypedValue.Value, Equals, "value")

	err = vAppTemplate.DeleteMetadata(ctx, "key")
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_DeleteMetadataOnMediaRecord(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	//prepare mediaRecord item
	skipWhenMediaPathMissing(vcd, check)
	itemName := "TestDeleteMediaMetaData"

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	check.Assert(uploadTask, NotNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_DeleteMetadataOnMediaRecord")

	err = vcd.org.Refresh(ctx)
	check.Assert(err, IsNil)

	mediaRecord, err := catalog.QueryMedia(ctx, itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	// Add metadata
	_, err = mediaRecord.AddMetadata(ctx, "key2", "value2")
	check.Assert(err, IsNil)

	// Remove metadata
	err = mediaRecord.DeleteMetadata(ctx, "key2")
	check.Assert(err, IsNil)

	metadata, err := mediaRecord.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: 'key2'")
		}
	}
}

func (vcd *TestVCD) Test_AddMetadataOnMediaRecord(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	//prepare media item
	skipWhenMediaPathMissing(vcd, check)
	itemName := "TestAddMediaMetaData"

	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	check.Assert(catalog.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(ctx, itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	check.Assert(uploadTask, NotNil)
	err = uploadTask.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_AddMetadataOnMediaRecord")

	err = vcd.org.Refresh(ctx)
	check.Assert(err, IsNil)

	mediaRecord, err := catalog.QueryMedia(ctx, itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	// Add metadata
	_, err = mediaRecord.AddMetadata(ctx, "key", "value")
	check.Assert(err, IsNil)

	// Check if metadata was added correctly
	metadata, err := mediaRecord.GetMetadata(ctx)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 1)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")
}
