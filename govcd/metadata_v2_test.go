//go:build vapp || vdc || metadata || functional || ALL
// +build vapp vdc metadata functional ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["metadata"] = "metadata_v2_test.go"
}

func (vcd *TestVCD) TestVmMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}

	if vcd.vapp.VApp == nil {
		check.Skip("skipping test because no vApp is found in configuration")
	}

	vApp := vcd.findFirstVapp()
	vmType, vmName := vcd.findFirstVm(vApp)
	if vmName == "" {
		check.Skip("skipping test because no VM is found")
	}

	vm := NewVM(&vcd.client.Client)
	vm.VM = &vmType

	testMetadataCRUDActions(vm, check, nil)
}

func (vcd *TestVCD) TestVdcMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.config.VCD.Nsxt.Vdc == "" {
		check.Skip("skipping test because VDC name is empty")
	}
	testMetadataCRUDActions(vcd.nsxtVdc, check, nil)
}

func (vcd *TestVCD) TestProviderVdcMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	providerVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	if err != nil {
		check.Skip(fmt.Sprintf("%s: Provider VDC %s not found. Test can't proceed", check.TestName(), vcd.config.VCD.NsxtProviderVdc.Name))
		return
	}
	testMetadataCRUDActions(providerVdc, check, nil)
}

func (vcd *TestVCD) TestVAppMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipVappTests {
		check.Skip("Skipping test because vApp was not successfully created at setup")
	}
	testMetadataCRUDActions(vcd.vapp, check, nil)
}

// TODO: Change to new vApp Template functions
func (vcd *TestVCD) TestVAppTemplateMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip("Skipping test because Catalog was not found. Test can't proceed")
		return
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

func (vcd *TestVCD) TestMediaRecordMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)

	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	check.Assert(catalog.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	uploadTask, err := catalog.UploadMediaImage(check.TestName(), check.TestName(), vcd.config.Media.MediaPath, 1024)
	check.Assert(err, IsNil)
	check.Assert(uploadTask, NotNil)
	err = uploadTask.WaitTaskCompletion()
	check.Assert(err, IsNil)

	AddToCleanupList(check.TestName(), "mediaCatalogImage", vcd.org.Org.Name+"|"+vcd.config.VCD.Catalog.Name, "Test_AddMetadataOnMediaRecord")

	err = vcd.org.Refresh()
	check.Assert(err, IsNil)

	mediaRecord, err := catalog.QueryMedia(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(mediaRecord, NotNil)
	check.Assert(mediaRecord.MediaRecord.Name, Equals, check.TestName())

	testMetadataCRUDActions(mediaRecord, check, nil)
}

func (vcd *TestVCD) TestMediaMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	skipWhenMediaPathMissing(vcd, check)

	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	check.Assert(catalog.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	media, err := catalog.GetMediaByName(vcd.config.Media.Media, false)
	check.Assert(err, IsNil)

	testMetadataCRUDActions(media, check, nil)
}

func (vcd *TestVCD) TestAdminCatalogMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	adminCatalog, err := org.GetAdminCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(adminCatalog, NotNil)
	check.Assert(adminCatalog.AdminCatalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	testMetadataCRUDActions(adminCatalog, check, func(testCase metadataTest) {
		testCatalogMetadata(vcd, check, testCase)
	})
}

func testCatalogMetadata(vcd *TestVCD, check *C, testCase metadataTest) {
	org, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	check.Assert(catalog.Catalog.Name, Equals, vcd.config.VCD.Catalog.Name)

	metadata, err := catalog.GetMetadata()
	check.Assert(err, IsNil)
	assertMetadata(check, metadata, testCase, 1)
}

func (vcd *TestVCD) TestAdminOrgMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	testMetadataCRUDActions(adminOrg, check, func(testCase metadataTest) {
		testOrgMetadata(vcd, check, testCase)
	})
}

func testOrgMetadata(vcd *TestVCD, check *C, testCase metadataTest) {
	org, err := vcd.client.GetOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	metadata, err := org.GetMetadata()
	check.Assert(err, IsNil)
	assertMetadata(check, metadata, testCase, 1)
}

func (vcd *TestVCD) TestDiskMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	diskCreateParams := &types.DiskCreateParams{
		Disk: &types.Disk{
			Name:        TestCreateDisk,
			SizeMb:      11,
			Description: TestCreateDisk,
		},
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

func (vcd *TestVCD) TestOrgVDCNetworkMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	net, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	if err != nil {
		check.Skip(fmt.Sprintf("network %s not found. Test can't proceed", vcd.config.VCD.Network.Net1))
		return
	}
	testMetadataCRUDActions(net, check, nil)
}

func (vcd *TestVCD) TestCatalogItemMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	catalog, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		check.Skip(fmt.Sprintf("Catalog %s not found. Test can't proceed", vcd.config.VCD.Catalog.Name))
		return
	}

	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	if err != nil {
		check.Skip(fmt.Sprintf("Catalog item %s not found. Test can't proceed", vcd.config.VCD.Catalog.CatalogItem))
		return
	}

	testMetadataCRUDActions(catalogItem, check, nil)
}

// metadataCompatible allows centralizing and generalizing the tests for metadata compatible resources.
type metadataCompatible interface {
	GetMetadata() (*types.Metadata, error)
	AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error
	MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error
	DeleteMetadataEntry(key string) error
}

type metadataTest struct {
	Key                   string
	Value                 string
	Type                  string
	Visibility            string
	IsSystem              bool
	ExpectErrorOnFirstAdd bool
}

// testMetadataCRUDActions performs a complete test of all use cases that metadata can have, for a metadata compatible resource.
// The function parameter extraReadStep performs an extra read step that can be passed as a function. Useful to perform a test
// on "admin+not admin" resource combinations, where the "not admin" only has a GetMetadata function.
// For example, AdminOrg and Org, where Org only has GetMetadata.
func testMetadataCRUDActions(resource metadataCompatible, check *C, extraReadStep func(testCase metadataTest)) {
	// Check how much metadata exists
	metadata, err := resource.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	existingMetaDataCount := len(metadata.MetadataEntry)

	var testCases = []metadataTest{
		{
			Key:                   "stringKey",
			Value:                 "stringValue",
			Type:                  types.MetadataStringValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: false,
		},
		{
			Key:                   "numberKey",
			Value:                 "notANumber",
			Type:                  types.MetadataNumberValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: true,
		},
		{
			Key:                   "numberKey",
			Value:                 "1",
			Type:                  types.MetadataNumberValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: false,
		},
		{
			Key:                   "boolKey",
			Value:                 "notABool",
			Type:                  types.MetadataBooleanValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: true,
		},
		{
			Key:                   "boolKey",
			Value:                 "true",
			Type:                  types.MetadataBooleanValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: false,
		},
		{
			Key:                   "dateKey",
			Value:                 "notADate",
			Type:                  types.MetadataDateTimeValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: true,
		},
		{
			Key:                   "dateKey",
			Value:                 "2022-10-05T13:44:00.000Z",
			Type:                  types.MetadataDateTimeValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: false,
		},
		{
			Key:                   "hidden",
			Value:                 "hiddenValue",
			Type:                  types.MetadataStringValue,
			Visibility:            types.MetadataHiddenVisibility,
			IsSystem:              true,
			ExpectErrorOnFirstAdd: false,
		},
		{
			Key:                   "readOnly",
			Value:                 "readOnlyValue",
			Type:                  types.MetadataStringValue,
			Visibility:            types.MetadataReadOnlyVisibility,
			IsSystem:              true,
			ExpectErrorOnFirstAdd: false,
		},
		{
			Key:                   "readWriteKey",
			Value:                 "butPlacedInSystem",
			Type:                  types.MetadataStringValue,
			Visibility:            types.MetadataReadWriteVisibility,
			IsSystem:              true,
			ExpectErrorOnFirstAdd: true, // types.MetadataReadWriteVisibility can't have isSystem=true
		},
		{
			Key:                   "readOnlyKey",
			Value:                 "butPlacedInGeneral",
			Type:                  types.MetadataStringValue,
			Visibility:            types.MetadataReadOnlyVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: true, // types.MetadataReadOnlyVisibility can't have isSystem=false
		},
		{
			Key:                   "hiddenKey",
			Value:                 "butPlacedInGeneral",
			Type:                  types.MetadataStringValue,
			Visibility:            types.MetadataHiddenVisibility,
			IsSystem:              false,
			ExpectErrorOnFirstAdd: true, // types.MetadataHiddenVisibility can't have isSystem=false
		},
	}

	for _, testCase := range testCases {

		err = resource.AddMetadataEntryWithVisibility(testCase.Key, testCase.Value, testCase.Type, testCase.Visibility, testCase.IsSystem)
		if testCase.ExpectErrorOnFirstAdd {
			check.Assert(err, NotNil)
			return
		}
		check.Assert(err, IsNil)

		// Check if metadata was added correctly
		metadata, err = resource.GetMetadata()
		check.Assert(err, IsNil)
		assertMetadata(check, metadata, testCase, existingMetaDataCount+1)

		// Perform an extra read step that can be passed as a function. Useful to perform a test
		// on resources that only have a GetMetadata function. For example, AdminOrg and Org, where Org only has GetMetadata.
		if extraReadStep != nil {
			extraReadStep(testCase)
		}

		// Merge updated metadata with a new entry
		err = resource.MergeMetadataWithMetadataValues(map[string]types.MetadataValue{
			"mergedKey": {
				TypedValue: &types.MetadataTypedValue{
					Value:   "mergedValue",
					XsiType: types.MetadataStringValue,
				},
			},
			testCase.Key: {
				TypedValue: &types.MetadataTypedValue{
					Value:   testCase.Value + "-Merged",
					XsiType: testCase.Type,
				},
			},
		})
		check.Assert(err, IsNil)

		// Check that the first key was updated and the second, created
		metadata, err = resource.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(metadata, NotNil)
		check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount+2)
		for _, entry := range metadata.MetadataEntry {
			switch entry.Key {
			case "mergedKey":
				check.Assert(entry.TypedValue.Value, Equals, "mergedValue")
			case testCase.Key:
				check.Assert(entry.TypedValue.Value, Equals, testCase.Value+"-Merged")
			}
		}

		err = resource.DeleteMetadataEntry("mergedKey")
		check.Assert(err, IsNil)
		err = resource.DeleteMetadataEntry(testCase.Key)
		check.Assert(err, IsNil)

		// Check if metadata was deleted correctly
		metadata, err = resource.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(metadata, NotNil)
		check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount)
	}
}

func assertMetadata(check *C, given *types.Metadata, expected metadataTest, expectedMetadataEntries int) {
	check.Assert(given, NotNil)
	check.Assert(len(given.MetadataEntry), Equals, expectedMetadataEntries)
	var foundEntry *types.MetadataEntry
	for _, entry := range given.MetadataEntry {
		if entry.Key == expected.Key {
			foundEntry = entry
		}
	}
	check.Assert(foundEntry, NotNil)
	check.Assert(foundEntry.Key, Equals, expected.Key)
	check.Assert(foundEntry.TypedValue.Value, Equals, expected.Value)
	check.Assert(foundEntry.TypedValue.XsiType, Equals, expected.Type)
	if expected.IsSystem {
		check.Assert(foundEntry.Domain, NotNil)
		check.Assert(foundEntry.Domain.Domain, Equals, "SYSTEM")
		check.Assert(foundEntry.Domain.Visibility, Equals, expected.Visibility)
	} else {
		if expected.Visibility == types.MetadataReadWriteVisibility {
			check.Assert(foundEntry.Domain, IsNil)
		} else {
			check.Assert(foundEntry.Domain.Domain, Equals, "GENERAL")
			check.Assert(foundEntry.Domain.Visibility, Equals, types.MetadataReadWriteVisibility)
		}
	}
}
