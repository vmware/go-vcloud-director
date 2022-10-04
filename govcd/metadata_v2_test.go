//go:build vapp || vdc || metadata || functional || ALL
// +build vapp vdc metadata functional ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func init() {
	testingTags["metadata"] = "metadata_v2_test.go"
}

//func (vcd *TestVCD) TestMetadataForVm(check *C) {
//	fmt.Printf("Running: %s\n", check.TestName())
//
//	if vcd.skipVappTests {
//		check.Skip("Skipping test because vApp was not successfully created at setup")
//	}
//
//	// Find VApp
//	if vcd.vapp.VApp == nil {
//		check.Skip("skipping test because no vApp is found in configuration")
//	}
//
//	vApp := vcd.findFirstVapp()
//	vmType, vmName := vcd.findFirstVm(vApp)
//	if vmName == "" {
//		check.Skip("skipping test because no VM is found")
//	}
//
//	vm := NewVM(&vcd.client.Client)
//	vm.VM = &vmType
//
//	testMetadataCRUDActions(vm, check, nil)
//}

type metadataCompatible interface {
	GetMetadata() (*types.Metadata, error)
	AddMetadataEntryWithVisibility(key, value, typedValue, visibility string, isSystem bool) error
	MergeMetadataWithMetadataValues(metadata map[string]types.MetadataValue) error
	DeleteMetadataEntry(key string) error
}

func testMetadataCRUDActions(resource metadataCompatible, check *C, extraCheck func()) {

	type metadataTests struct {
		Key        string
		Value      string
		Type       string
		Visibility string
		IsSystem   bool
	}

	// Check how much metadata exists
	metadata, err := resource.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	existingMetaDataCount := len(metadata.MetadataEntry)

	var testCases = []metadataTests{
		{
			Key:        "stringKey",
			Value:      "stringValue",
			Type:       types.MetadataStringValue,
			Visibility: types.MetadataReadWriteVisibility,
			IsSystem:   true,
		},
		{
			Key:        "boolKey",
			Value:      "true",
			Type:       types.MetadataBooleanValue,
			Visibility: types.MetadataReadWriteVisibility,
			IsSystem:   true,
		},
		{
			Key:        "dateKey",
			Value:      "2022-10-05T13:44:00.000Z",
			Type:       types.MetadataDateTimeValue,
			Visibility: types.MetadataReadWriteVisibility,
			IsSystem:   true,
		},
		{
			Key:        "hidden",
			Value:      "hiddenValue",
			Type:       types.MetadataStringValue,
			Visibility: types.MetadataHiddenVisibility,
			IsSystem:   true,
		},
		{
			Key:        "readOnly",
			Value:      "readOnlyValue",
			Type:       types.MetadataStringValue,
			Visibility: types.MetadataReadOnlyVisibility,
			IsSystem:   true,
		},
		{
			Key:        "notSystem",
			Value:      "notSystemValue",
			Type:       types.MetadataStringValue,
			Visibility: types.MetadataReadWriteVisibility,
			IsSystem:   false,
		},
	}

	for _, testCase := range testCases {

		err = resource.AddMetadataEntryWithVisibility(testCase.Key, testCase.Value, testCase.Type, testCase.Visibility, testCase.IsSystem)
		check.Assert(err, IsNil)

		// Check if metadata was added correctly
		metadata, err = resource.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(metadata, NotNil)
		check.Assert(len(metadata.MetadataEntry), Equals, existingMetaDataCount+1)
		var foundEntry *types.MetadataEntry
		for _, entry := range metadata.MetadataEntry {
			if entry.Key == testCase.Key {
				foundEntry = entry
			}
		}

		check.Assert(foundEntry, NotNil)
		check.Assert(foundEntry.Key, Equals, testCase.Key)
		check.Assert(foundEntry.TypedValue.Value, Equals, testCase.Value)
		check.Assert(foundEntry.TypedValue.XsiType, Equals, testCase.Type)
		check.Assert(foundEntry.Domain.Visibility, Equals, testCase.Visibility)
		if testCase.IsSystem {
			check.Assert(foundEntry.Domain.Domain, Equals, "SYSTEM")
		} else {
			check.Assert(foundEntry.Domain.Domain, Equals, "GENERAL")
		}

		// Perform an extra check that can be passed as a function
		if extraCheck != nil {
			extraCheck()
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
					Value: testCase.Value + "-Merged",
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
