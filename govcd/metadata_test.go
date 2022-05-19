//go:build vapp || vdc || metadata || functional || ALL
// +build vapp vdc metadata functional ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

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
	vdc, err := vcd.vdc.AddMetadata("key", "value")
	check.Assert(err, IsNil)
	check.Assert(vdc, Not(Equals), Vdc{})

	AddToCleanupList("key", "vdcMetaData", vcd.org.Org.Name+"|"+vcd.config.VCD.Vdc, check.TestName())

	// Check if metadata was added correctly
	metadata, err := vcd.vdc.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 1)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")

	// Remove metadata
	vdc, err = vcd.vdc.DeleteMetadata("key")
	check.Assert(err, IsNil)
	check.Assert(vdc, Not(Equals), Vdc{})

	metadata, err = vcd.vdc.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 0)
}

func (vcd *TestVCD) Test_AddMetadataOnVapp(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	// Add metadata
	task, err := vcd.vapp.AddMetadata("key", "value")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Check if metadata was added correctly
	metadata, err := vcd.vapp.GetMetadata()
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
	task, err := vcd.vapp.AddMetadata("key2", "value2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Remove metadata
	task, err = vcd.vapp.DeleteMetadata("key2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	metadata, err := vcd.vapp.GetMetadata()
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

	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Add metadata
	task, err := vm.AddMetadata("key", "value")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Check if metadata was added correctly
	metadata, err := vm.GetMetadata()
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

	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	// Add metadata
	task, err := vm.AddMetadata("key2", "value2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")

	// Remove metadata
	task, err = vm.DeleteMetadata("key2")
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	check.Assert(task.Task.Status, Equals, "success")
	metadata, err := vm.GetMetadata()
	check.Assert(err, IsNil)
	for _, k := range metadata.MetadataEntry {
		if k.Key == "key2" {
			check.Errorf("metadata.MetadataEntry should not contain key: 'key2'")
		}
	}
}

func (vcd *TestVCD) Test_DeleteMetadataOnVAppTemplate(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog not found. Test can't proceed")
		return
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	vAppTemplate, err := catItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Add metadata
	_, err = vAppTemplate.AddMetadata("key2", "value2")
	check.Assert(err, IsNil)

	// Remove metadata
	err = vAppTemplate.DeleteMetadata("key2")
	check.Assert(err, IsNil)

	metadata, err := vAppTemplate.GetMetadata()
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
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog not found. Test can't proceed")
		return
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	vAppTemplate, err := catItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	// Check how much metaData exist
	metadata, err := vAppTemplate.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	existingMetaDataCount := len(metadata.MetadataEntry)

	// Add metadata
	_, err = vAppTemplate.AddMetadata("key", "value")
	check.Assert(err, IsNil)

	// Check if metadata was added correctly
	metadata, err = vAppTemplate.GetMetadata()
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

	err = vAppTemplate.DeleteMetadata("key")
	check.Assert(err, IsNil)
}

func (vcd *TestVCD) Test_DeleteMetadataOnMediaRecord(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	//prepare mediaRecord item
	skipWhenMediaPathMissing(vcd, check)
	itemName := "TestDeleteMediaMetaData"

	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	check.Assert(uploadTask, NotNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_DeleteMetadataOnMediaRecord")

	err = vcd.org.Refresh()
	check.Assert(err, IsNil)

	mediaRecord, err := catalog.QueryMedia(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	// Add metadata
	_, err = mediaRecord.AddMetadata("key2", "value2")
	check.Assert(err, IsNil)

	// Remove metadata
	err = mediaRecord.DeleteMetadata("key2")
	check.Assert(err, IsNil)

	metadata, err := mediaRecord.GetMetadata()
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
	itemName := "TestAddMediaMetaDataEntry"

	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	check.Assert(catalog.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	check.Assert(uploadTask, NotNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_AddMetadataOnMediaRecord")

	err = vcd.org.Refresh()
	check.Assert(err, IsNil)

	mediaRecord, err := catalog.QueryMedia(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	// Add metadata
	_, err = mediaRecord.AddMetadata("key", "value")
	check.Assert(err, IsNil)

	// Check if metadata was added correctly
	metadata, err := mediaRecord.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 1)
	check.Assert(metadata.MetadataEntry[0].Key, Equals, "key")
	check.Assert(metadata.MetadataEntry[0].TypedValue.Value, Equals, "value")
}

func (vcd *TestVCD) Test_MetadataOnAdminCatalogCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	var catalogName string = check.TestName()

	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := vcd.org.CreateCatalog(catalogName, catalogName)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	AddToCleanupList(catalogName, "catalog", org.AdminOrg.Name, catalogName)

	adminCatalog, err := org.GetAdminCatalogByName(catalogName, true)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)

	testMetadataCRUDActions(adminCatalog, check, func() {
		metadata, err := catalog.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(metadata, NotNil)
		check.Assert(len(metadata.MetadataEntry), Equals, 1)
		var foundEntry *types.MetadataEntry
		for _, entry := range metadata.MetadataEntry {
			if entry.Key == "key" {
				foundEntry = entry
			}
		}
		check.Assert(foundEntry, NotNil)
		check.Assert(foundEntry.Key, Equals, "key")
		check.Assert(foundEntry.TypedValue.Value, Equals, "value")
	})
}

func (vcd *TestVCD) Test_MetadataEntryForVdcCRUD(check *C) {
	if vcd.config.VCD.Vdc == "" {
		check.Skip("skipping test because VDC name is empty")
	}

	fmt.Printf("Running: %s\n", check.TestName())

	testMetadataCRUDActions(vcd.vdc, check, nil)
}

func (vcd *TestVCD) Test_MetadataEntryOnVappCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	testMetadataCRUDActions(vcd.vapp, check, nil)
}

func (vcd *TestVCD) Test_MetadataEntryOnVmCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vapp was not successfully created at setup")
	}

	// Find VApp
	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found")
	}

	vapp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vapp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	testMetadataCRUDActions(vm, check, nil)
}

func (vcd *TestVCD) Test_MetadataEntryOnVAppTemplateCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog not found. Test can't proceed")
		return
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("Test_DeleteMetadataOnCatalogItem: Catalog Item not given. Test can't proceed")
	}

	catItem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catItem, NotNil)
	check.Assert(catItem.CatalogItem.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	vAppTemplate, err := catItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vAppTemplate, NotNil)
	check.Assert(vAppTemplate.VAppTemplate.Name, Equals, vcd.config.VCD.Catalog.CatalogItem)

	testMetadataCRUDActions(&vAppTemplate, check, nil)
}

func (vcd *TestVCD) Test_MetadataEntryOnMediaRecordCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	//prepare media item
	skipWhenMediaPathMissing(vcd, check)
	itemName := "TestAddMediaMetaData"

	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	check.Assert(catalog.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(itemName, "upload from test", vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	check.Assert(uploadTask, NotNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(itemName, "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_AddMetadataOnMediaRecord")

	err = vcd.org.Refresh()
	check.Assert(err, IsNil)

	mediaRecord, err := catalog.QueryMedia(itemName)
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, itemName)

	testMetadataCRUDActions(mediaRecord, check, nil)
}

func (vcd *TestVCD) Test_MetadataOnAdminOrgCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	adminOrg, err := vcd.client.GetAdminOrgById(vcd.org.Org.ID)
	if err != nil {
		check.Skip("Test_AddMetadataOnAdminOrg: Organization (admin) not found. Test can't proceed")
		return
	}
	org, err := vcd.client.GetOrgById(vcd.org.Org.ID)
	if err != nil {
		check.Skip("Test_AddMetadataOnAdminOrg: Organization not found. Test can't proceed")
		return
	}

	testMetadataCRUDActions(adminOrg, check, func() {
		metadata, err := org.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(metadata, NotNil)
		check.Assert(len(metadata.MetadataEntry), Equals, 1)
		var foundEntry *types.MetadataEntry
		for _, entry := range metadata.MetadataEntry {
			if entry.Key == "key" {
				foundEntry = entry
			}
		}
		check.Assert(foundEntry, NotNil)
		check.Assert(foundEntry.Key, Equals, "key")
		check.Assert(foundEntry.TypedValue.Value, Equals, "value")
	})

}

func (vcd *TestVCD) Test_MetadataOnIndependentDiskCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	// Create disk
	diskCreateParamsDisk := &types.Disk{
		Name:        TestCreateDisk,
		SizeMb:      11,
		Description: TestCreateDisk,
	}

	diskCreateParams := &types.DiskCreateParams{
		Disk: diskCreateParamsDisk,
	}

	task, err := vcd.vdc.CreateDisk(diskCreateParams)
	check.Assert(err, IsNil)

	diskHREF := task.Task.Owner.HREF
	PrependToCleanupList(diskHREF, "disk", "", check.TestName())

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	disk, err := vcd.vdc.GetDiskByHref(diskHREF)
	check.Assert(err, IsNil)

	testMetadataCRUDActions(disk, check, nil)
}

func (vcd *TestVCD) Test_MetadataOnVdcNetworkCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	net, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	if err != nil {
		check.Skip(fmt.Sprintf("%s: Network %s not found. Test can't proceed", check.TestName(), vcd.config.VCD.Network.Net1))
		return
	}

	testMetadataCRUDActions(net, check, nil)
}

func (vcd *TestVCD) Test_MetadataOnCatalogItemCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip(fmt.Sprintf("%s: Catalog %s not found. Test can't proceed", check.TestName(), vcd.config.VCD.Catalog.Name))
		return
	}

	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	if err != nil {
		check.Skip(fmt.Sprintf("%s: Catalog item %s not found. Test can't proceed", check.TestName(), vcd.config.VCD.Catalog.CatalogItem))
		return
	}

	testMetadataCRUDActions(catalogItem, check, nil)
}


func (vcd *TestVCD) Test_MetadataOnOpenApiOrgVdcNetworkCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	net, err := vcd.vdc.GetOpenApiOrgVdcNetworkByName(vcd.config.VCD.Network.Net1)
	if err != nil {
		check.Skip(fmt.Sprintf("%s: Network %s not found. Test can't proceed", check.TestName(), vcd.config.VCD.Network.Net1))
		return
	}

	testMetadataCRUDActions(net, check, nil)
}

func (vcd *TestVCD) Test_MetadataByHrefCRUD(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	storageProfileRef, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	if err != nil {
		check.Skip(fmt.Sprintf("%s: Storage Profile %s not found. Test can't proceed", check.TestName(), vcd.config.VCD.StorageProfile.SP1))
		return
	}

	storageProfileAdminHref := strings.ReplaceAll(storageProfileRef.HREF, "api", "api/admin")

	// Check how much metadata exists
	metadata, err := vcd.client.GetMetadataByHref(storageProfileAdminHref)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	existingMetaDataCount := len(metadata.MetadataEntry)

	// Add metadata
	err = vcd.client.AddMetadataEntryByHref(storageProfileAdminHref, types.MetadataStringValue, "key", "value")
	check.Assert(err, IsNil)

	// Check if metadata was added correctly
	metadata, err = vcd.client.GetMetadataByHref(storageProfileAdminHref)
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

	// Check the same without admin privileges
	metadata, err = vcd.client.GetMetadataByHref(storageProfileRef.HREF)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount+1)
	for _, entry := range metadata.MetadataEntry {
		if entry.Key == "key" {
			foundEntry = entry
		}
	}
	check.Assert(foundEntry, NotNil)
	check.Assert(foundEntry.Key, Equals, "key")
	check.Assert(foundEntry.TypedValue.Value, Equals, "value")

	// Merge updated metadata with a new entry
	err = vcd.client.MergeMetadataByHref(storageProfileAdminHref, types.MetadataStringValue, map[string]interface{}{
		"key":  "valueUpdated",
		"key2": "value2",
	})
	check.Assert(err, IsNil)

	// Check that the first key was updated and the second, created
	metadata, err = vcd.client.GetMetadataByHref(storageProfileRef.HREF)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount+2)
	for _, entry := range metadata.MetadataEntry {
		switch entry.Key {
		case "key":
			check.Assert(entry.TypedValue.Value, Equals, "valueUpdated")
		case "key2":
			check.Assert(entry.TypedValue.Value, Equals, "value2")
		}
	}

	// Delete the metadata
	err = vcd.client.DeleteMetadataEntryByHref(storageProfileAdminHref, "key")
	check.Assert(err, IsNil)
	err = vcd.client.DeleteMetadataEntryByHref(storageProfileAdminHref, "key2")
	check.Assert(err, IsNil)
	// Check if metadata was deleted correctly
	metadata, err = vcd.client.GetMetadataByHref(storageProfileAdminHref)
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, 0)
}

func testMetadataCRUDActions(resource MetadataCompatible, check *C, extraCheck func()) {
	// Check how much metadata exists
	metadata, err := resource.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	existingMetaDataCount := len(metadata.MetadataEntry)

	// Add metadata
	err = resource.AddMetadataEntry(types.MetadataStringValue, "key", "value")
	check.Assert(err, IsNil)

	// Check if metadata was added correctly
	metadata, err = resource.GetMetadata()
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

	if extraCheck != nil {
		extraCheck()
	}

	// Merge updated metadata with a new entry
	err = resource.MergeMetadata(types.MetadataStringValue, map[string]interface{}{
		"key":  "valueUpdated",
		"key2": "value2",
	})
	check.Assert(err, IsNil)

	// Check that the first key was updated and the second, created
	metadata, err = resource.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount+2)
	for _, entry := range metadata.MetadataEntry {
		switch entry.Key {
		case "key":
			check.Assert(entry.TypedValue.Value, Equals, "valueUpdated")
		case "key2":
			check.Assert(entry.TypedValue.Value, Equals, "value2")
		}
	}

	err = resource.DeleteMetadataEntry("key")
	check.Assert(err, IsNil)
	err = resource.DeleteMetadataEntry("key2")
	check.Assert(err, IsNil)
	// Check if metadata was deleted correctly
	metadata, err = resource.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount)
}

