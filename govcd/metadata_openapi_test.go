//go:build metadata || openapi || rde || functional || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"regexp"
	"strings"
)

func (vcd *TestVCD) TestRdeMetadata(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	// This RDE type comes out of the box in VCD
	rdeType, err := vcd.client.GetRdeType("vmware", "tkgcluster", "1.0.0")
	check.Assert(err, IsNil)
	check.Assert(rdeType, NotNil)

	rde, err := rdeType.CreateRde(types.DefinedEntity{
		Name:   check.TestName(),
		Entity: map[string]interface{}{"foo": "bar"}, // We don't care about schema correctness here
	}, nil)
	check.Assert(err, IsNil)
	check.Assert(rde, NotNil)

	err = rde.Resolve() // State will be RESOLUTION_ERROR, but we don't care. We resolve to be able to delete it later.
	check.Assert(err, IsNil)

	// The RDE can't be deleted until rde.Resolve() is called
	AddToCleanupListOpenApi(rde.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+rde.DefinedEntity.ID)

	testOpenApiMetadataCRUDActions(rde, check)
	vcd.testOpenApiMetadataIgnore(rde, "entity", rde.DefinedEntity.Name, check)

	err = rde.Delete()
	check.Assert(err, IsNil)
}

// openApiMetadataCompatible allows centralizing and generalizing the tests for OpenAPI metadata compatible resources.
type openApiMetadataCompatible interface {
	GetMetadata() ([]*OpenApiMetadataEntry, error)
	GetMetadataByKey(domain, namespace, key string) (*OpenApiMetadataEntry, error)
	GetMetadataById(id string) (*OpenApiMetadataEntry, error)
	AddMetadata(metadataEntry types.OpenApiMetadataEntry) (*OpenApiMetadataEntry, error)
}

type openApiMetadataTest struct {
	Key                               string
	Value                             interface{} // The type depends on the Type attribute
	UpdateValue                       interface{}
	Namespace                         string
	Type                              string
	IsReadOnly                        bool
	IsPersistent                      bool
	Domain                            string
	ExpectErrorOnFirstAddMatchesRegex string
}

// testOpenApiMetadataCRUDActions performs a complete test of all use cases that metadata in OpenAPI can have,
// for an OpenAPI metadata compatible resource.
func testOpenApiMetadataCRUDActions(resource openApiMetadataCompatible, check *C) {
	// Check how much metadata exists
	metadata, err := resource.GetMetadata()
	check.Assert(err, IsNil)
	check.Assert(metadata, NotNil)
	existingMetaDataCount := len(metadata)

	var testCases = []openApiMetadataTest{
		{
			Key:         "stringKey",
			Value:       "stringValue",
			UpdateValue: "stringValueUpdated",
			Type:        types.OpenApiMetadataStringEntry,
			IsReadOnly:  false,
			Domain:      "TENANT",
			Namespace:   "foo",
		},
		{
			Key:                               "numberKey",
			Value:                             "notANumber",
			Type:                              types.OpenApiMetadataNumberEntry,
			IsReadOnly:                        false,
			Domain:                            "TENANT",
			Namespace:                         "foo",
			ExpectErrorOnFirstAddMatchesRegex: "notANumber",
		},
		{
			Key:         "numberKey",
			Value:       float64(1),
			UpdateValue: float64(42),
			Type:        types.OpenApiMetadataNumberEntry,
			IsReadOnly:  false,
			Domain:      "TENANT",
			Namespace:   "foo",
		},
		{
			Key:         "negativeNumberKey",
			Value:       float64(-1),
			UpdateValue: float64(-42),
			Type:        types.OpenApiMetadataNumberEntry,
			IsReadOnly:  false,
			Domain:      "TENANT",
			Namespace:   "foo",
		},
		{
			Key:                               "boolKey",
			Value:                             "notABool",
			Type:                              types.OpenApiMetadataBooleanEntry,
			IsReadOnly:                        false,
			Domain:                            "TENANT",
			Namespace:                         "foo",
			ExpectErrorOnFirstAddMatchesRegex: "notABool",
		},
		{
			Key:         "boolKey",
			Value:       true,
			UpdateValue: false,
			Type:        types.OpenApiMetadataBooleanEntry,
			IsReadOnly:  false,
			Domain:      "TENANT",
			Namespace:   "foo",
		},
		{
			Key:         "providerKey",
			Value:       "providerValue",
			UpdateValue: "providerValueUpdated",
			Type:        types.OpenApiMetadataStringEntry,
			IsReadOnly:  false,
			Domain:      "PROVIDER",
			Namespace:   "foo",
		},
		{
			Key:                               "readOnlyProviderKey",
			Value:                             "readOnlyProviderValue",
			Type:                              types.OpenApiMetadataStringEntry,
			IsReadOnly:                        true,
			Domain:                            "PROVIDER",
			Namespace:                         "foo",
			ExpectErrorOnFirstAddMatchesRegex: "VCD_META_CRUD_INVALID_FLAG",
		},
		{
			Key:        "readOnlyTenantKey",
			Value:      "readOnlyTenantValue",
			Type:       types.OpenApiMetadataStringEntry,
			IsReadOnly: true,
			Domain:     "TENANT",
			Namespace:  "foo",
		},
		{
			Key:          "persistentKey",
			Value:        "persistentValue",
			Type:         types.OpenApiMetadataStringEntry,
			IsReadOnly:   false,
			IsPersistent: true,
			Domain:       "TENANT",
			Namespace:    "foo",
		},
	}

	for _, testCase := range testCases {

		var createdEntry *OpenApiMetadataEntry
		createdEntry, err = resource.AddMetadata(types.OpenApiMetadataEntry{
			KeyValue: types.OpenApiMetadataKeyValue{
				Domain:    testCase.Domain,
				Key:       testCase.Key,
				Namespace: testCase.Namespace,
				Value: types.OpenApiMetadataTypedValue{
					Type:  testCase.Type,
					Value: testCase.Value,
				},
			},
			IsPersistent: testCase.IsPersistent,
			IsReadOnly:   testCase.IsReadOnly,
		})
		if testCase.ExpectErrorOnFirstAddMatchesRegex != "" {
			p := regexp.MustCompile("(?s)" + testCase.ExpectErrorOnFirstAddMatchesRegex)
			check.Assert(p.MatchString(err.Error()), Equals, true)
			continue
		}
		check.Assert(err, IsNil)
		check.Assert(createdEntry, NotNil)
		check.Assert(createdEntry.href, Not(Equals), "")
		check.Assert(createdEntry.Etag, Not(Equals), "")
		check.Assert(createdEntry.parentEndpoint, Not(Equals), "")
		check.Assert(createdEntry.MetadataEntry, NotNil)
		check.Assert(createdEntry.MetadataEntry.ID, Not(Equals), "")

		// Check if metadata was added correctly
		metadata, err = resource.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(len(metadata), Equals, existingMetaDataCount+1)
		found := false
		for _, entry := range metadata {
			if entry.MetadataEntry.ID == createdEntry.MetadataEntry.ID {
				check.Assert(*entry.MetadataEntry, DeepEquals, *createdEntry.MetadataEntry)
				found = true
				break
			}
		}
		check.Assert(found, Equals, true)

		metadataByKey, err := resource.GetMetadataByKey(createdEntry.MetadataEntry.KeyValue.Domain, createdEntry.MetadataEntry.KeyValue.Namespace, createdEntry.MetadataEntry.KeyValue.Key)
		check.Assert(err, IsNil)
		check.Assert(metadataByKey, NotNil)
		check.Assert(metadataByKey.MetadataEntry, NotNil)
		check.Assert(*metadataByKey.MetadataEntry, DeepEquals, *createdEntry.MetadataEntry)
		check.Assert(metadataByKey.Etag, Equals, createdEntry.Etag)
		check.Assert(metadataByKey.parentEndpoint, Equals, createdEntry.parentEndpoint)
		check.Assert(metadataByKey.href, Equals, createdEntry.href)

		metadataById, err := resource.GetMetadataById(metadataByKey.MetadataEntry.ID)
		check.Assert(err, IsNil)
		check.Assert(metadataById, NotNil)
		check.Assert(metadataById.MetadataEntry, NotNil)
		check.Assert(*metadataById.MetadataEntry, DeepEquals, *metadataById.MetadataEntry)
		check.Assert(metadataById.Etag, Equals, metadataByKey.Etag)
		check.Assert(metadataById.parentEndpoint, Equals, metadataByKey.parentEndpoint)
		check.Assert(metadataById.href, Equals, metadataByKey.href)

		if testCase.UpdateValue != nil {
			oldEtag := metadataById.Etag
			err = metadataById.Update(testCase.UpdateValue, !metadataById.MetadataEntry.IsPersistent)
			check.Assert(err, IsNil)
			check.Assert(metadataById, NotNil)
			check.Assert(metadataById.MetadataEntry, NotNil)
			// Changed fields
			check.Assert(metadataById.MetadataEntry.IsPersistent, Equals, !metadataByKey.MetadataEntry.IsPersistent)
			check.Assert(metadataById.MetadataEntry.KeyValue.Value.Value, Equals, testCase.UpdateValue)
			// Non-changed fields
			check.Assert(metadataById.MetadataEntry.ID, Equals, metadataByKey.MetadataEntry.ID)
			check.Assert(metadataById.MetadataEntry.KeyValue.Value.Type, Equals, metadataByKey.MetadataEntry.KeyValue.Value.Type)
			check.Assert(metadataById.MetadataEntry.KeyValue.Namespace, Equals, metadataByKey.MetadataEntry.KeyValue.Namespace)
			check.Assert(metadataById.MetadataEntry.IsReadOnly, Equals, metadataByKey.MetadataEntry.IsReadOnly)
			check.Assert(metadataById.Etag, Not(Equals), oldEtag) // ETag should be refreshed as we did an update
			check.Assert(metadataById.parentEndpoint, Equals, metadataByKey.parentEndpoint)
			check.Assert(metadataById.href, Equals, metadataByKey.href)
		}

		err = metadataById.Delete()
		check.Assert(err, IsNil)
		check.Assert(*metadataById.MetadataEntry, DeepEquals, types.OpenApiMetadataEntry{})
		check.Assert(metadataById.Etag, Equals, "")
		check.Assert(metadataById.href, Equals, "")
		check.Assert(metadataById.parentEndpoint, Equals, "")

		// Check if metadata was deleted correctly
		deletedMetadata, err := resource.GetMetadataById(metadataByKey.MetadataEntry.ID)
		check.Assert(err, NotNil)
		check.Assert(deletedMetadata, IsNil)
		check.Assert(true, Equals, ContainsNotFound(err))
	}
}

func (vcd *TestVCD) testOpenApiMetadataIgnore(resource openApiMetadataCompatible, objectType, objectName string, check *C) {
	existingMetadata, err := resource.GetMetadata()
	check.Assert(err, IsNil)

	_, err = resource.AddMetadata(types.OpenApiMetadataEntry{
		IsPersistent: false,
		IsReadOnly:   false,
		KeyValue: types.OpenApiMetadataKeyValue{
			Domain: "TENANT",
			Key:    "foo",
			Value: types.OpenApiMetadataTypedValue{
				Value: "bar",
				Type:  types.OpenApiMetadataStringEntry,
			},
			Namespace: "",
		},
	})
	check.Assert(err, IsNil)
	_, err = resource.AddMetadata(types.OpenApiMetadataEntry{
		IsPersistent: false,
		IsReadOnly:   false,
		KeyValue: types.OpenApiMetadataKeyValue{
			Domain: "TENANT",
			Key:    "not_ignored",
			Value: types.OpenApiMetadataTypedValue{
				Value: "bar2",
				Type:  types.OpenApiMetadataStringEntry,
			},
			Namespace: "",
		},
	})
	check.Assert(err, IsNil)

	cleanup := func() {
		vcd.client.Client.IgnoredMetadata = nil
		metadata, err := resource.GetMetadata()
		check.Assert(err, IsNil)
		for _, entry := range metadata {
			itWasAlreadyPresent := false
			for _, existingEntry := range existingMetadata {
				if existingEntry.MetadataEntry.KeyValue.Namespace == entry.MetadataEntry.KeyValue.Namespace &&
					existingEntry.MetadataEntry.KeyValue.Key == entry.MetadataEntry.KeyValue.Key &&
					existingEntry.MetadataEntry.KeyValue.Value.Value == entry.MetadataEntry.KeyValue.Value.Value &&
					existingEntry.MetadataEntry.KeyValue.Value.Type == entry.MetadataEntry.KeyValue.Value.Type {
					itWasAlreadyPresent = true
				}
			}
			if !itWasAlreadyPresent {
				toDelete, err := resource.GetMetadataById(entry.MetadataEntry.ID)
				check.Assert(err, IsNil)
				err = toDelete.Delete()
				check.Assert(err, IsNil)
			}
		}
		metadata, err = resource.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(len(metadata), Equals, len(existingMetadata))
	}
	defer cleanup()

	tests := []struct {
		ignoredMetadata   []IgnoredMetadata
		metadataIsIgnored bool
	}{
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectType: &objectType, KeyRegex: regexp.MustCompile(`^foo$`)}},
			metadataIsIgnored: true,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectType: &objectType, ValueRegex: regexp.MustCompile(`^bar$`)}},
			metadataIsIgnored: true,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectType: &objectType, KeyRegex: regexp.MustCompile(`^fizz$`)}},
			metadataIsIgnored: false,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectType: &objectType, ValueRegex: regexp.MustCompile(`^buzz$`)}},
			metadataIsIgnored: false,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectName: &objectName, KeyRegex: regexp.MustCompile(`^foo$`)}},
			metadataIsIgnored: true,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectName: &objectName, ValueRegex: regexp.MustCompile(`^bar$`)}},
			metadataIsIgnored: true,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectName: &objectName, KeyRegex: regexp.MustCompile(`^fizz$`)}},
			metadataIsIgnored: false,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectName: &objectName, ValueRegex: regexp.MustCompile(`^buzz$`)}},
			metadataIsIgnored: false,
		},
		{
			ignoredMetadata:   []IgnoredMetadata{{ObjectType: &objectType, ObjectName: &objectName, KeyRegex: regexp.MustCompile(`foo`), ValueRegex: regexp.MustCompile(`bar`)}},
			metadataIsIgnored: true,
		},
	}

	for _, tt := range tests {
		vcd.client.Client.IgnoredMetadata = tt.ignoredMetadata

		// Tests getting a simple metadata entry by its key
		singleMetadata, err := resource.GetMetadataByKey("", "", "foo")
		if tt.metadataIsIgnored {
			check.Assert(err, NotNil)
			check.Assert(true, Equals, strings.Contains(err.Error(), "could not find the metadata associated to object"))
		} else {
			check.Assert(err, IsNil)
			check.Assert(singleMetadata, NotNil)
			check.Assert(singleMetadata.MetadataEntry.KeyValue.Value.Value, Equals, "bar")
		}

		// Retrieve all metadata
		allMetadata, err := resource.GetMetadata()
		check.Assert(err, IsNil)
		check.Assert(allMetadata, NotNil)
		if tt.metadataIsIgnored {
			// If metadata is ignored, there should be an offset of 1 entry (with key "test")
			check.Assert(len(allMetadata), Equals, len(existingMetadata)+1)
			for _, entry := range allMetadata {
				if tt.metadataIsIgnored {
					check.Assert(entry.MetadataEntry.KeyValue.Key, Not(Equals), "foo")
					check.Assert(entry.MetadataEntry.KeyValue.Value.Value, Not(Equals), "bar")
				}
			}
		} else {
			// If metadata is NOT ignored, there should be an offset of 2 entries (with key "foo" and "test")
			check.Assert(len(allMetadata), Equals, len(existingMetadata)+2)
		}
	}
}
