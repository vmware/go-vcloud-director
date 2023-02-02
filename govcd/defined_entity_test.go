//go:build functional || openapi || rde || ALL
// +build functional openapi rde ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Test_Rde tests the complete journey of RDE type and RDE instance with creation, reads, updates and finally deletion.
func (vcd *TestVCD) Test_Rde(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	for _, endpoint := range []string{
		types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntityTypes,
		types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntitiesResolve,
		types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntities,
	} {
		skipOpenApiEndpointTest(vcd, check, endpoint)
	}

	unmarshaledRdeTypeSchema, err := loadRdeTypeSchemaFromTestResources()
	check.Assert(err, IsNil)
	check.Assert(true, Equals, len(unmarshaledRdeTypeSchema) > 0)

	rdeTypeToCreate := &types.DefinedEntityType{
		Name:        check.TestName(),
		Namespace:   "namespace", // Can't put check.TestName() due to a bug that causes RDEs to fail on GET once created with special characters like "."
		Version:     "1.2.3",
		Description: "Description of " + check.TestName(),
		Schema:      unmarshaledRdeTypeSchema,
		Vendor:      "vmware",
		Interfaces:  []string{"urn:vcloud:interface:vmware:k8s:1.0.0"},
	}

	allRdeTypes, err := vcd.client.GetAllRdeTypes(nil)
	check.Assert(err, IsNil)
	alreadyPresentRdes := len(allRdeTypes)

	createdRdeType, err := vcd.client.CreateRdeType(rdeTypeToCreate)
	check.Assert(err, IsNil)
	check.Assert(createdRdeType, NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Name, Equals, rdeTypeToCreate.Name)
	check.Assert(createdRdeType.DefinedEntityType.Namespace, Equals, rdeTypeToCreate.Namespace)
	check.Assert(createdRdeType.DefinedEntityType.Version, Equals, rdeTypeToCreate.Version)
	check.Assert(createdRdeType.DefinedEntityType.Schema, NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Schema["type"], Equals, "object")
	check.Assert(createdRdeType.DefinedEntityType.Schema["definitions"], NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Schema["required"], NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Schema["properties"], NotNil)

	AddToCleanupListOpenApi(createdRdeType.DefinedEntityType.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEntityTypes+createdRdeType.DefinedEntityType.ID)

	allRdeTypes, err = vcd.client.GetAllRdeTypes(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allRdeTypes), Equals, alreadyPresentRdes+1)

	obtainedRdeType, err := vcd.client.GetRdeTypeById(createdRdeType.DefinedEntityType.ID)
	check.Assert(err, IsNil)
	check.Assert(*obtainedRdeType.DefinedEntityType, DeepEquals, *createdRdeType.DefinedEntityType)

	obtainedRdeType2, err := vcd.client.GetRdeType(obtainedRdeType.DefinedEntityType.Vendor, obtainedRdeType.DefinedEntityType.Namespace, obtainedRdeType.DefinedEntityType.Version)
	check.Assert(err, IsNil)
	check.Assert(*obtainedRdeType2.DefinedEntityType, DeepEquals, *obtainedRdeType.DefinedEntityType)

	// We don't want to update the name nor the schema. It should populate them from the receiver object automatically
	err = obtainedRdeType.Update(types.DefinedEntityType{
		Description: obtainedRdeType.DefinedEntityType.Description + "Updated",
	})
	check.Assert(err, IsNil)
	check.Assert(obtainedRdeType.DefinedEntityType.Description, Equals, rdeTypeToCreate.Description+"Updated")

	testRdeCrudWithGivenType(check, obtainedRdeType)
	testRdeCrudAsTenant(check, obtainedRdeType.DefinedEntityType.Vendor, obtainedRdeType.DefinedEntityType.Namespace, obtainedRdeType.DefinedEntityType.Version, vcd.client)

	deletedId := createdRdeType.DefinedEntityType.ID
	err = createdRdeType.Delete()
	check.Assert(err, IsNil)
	check.Assert(*createdRdeType.DefinedEntityType, DeepEquals, types.DefinedEntityType{})

	_, err = vcd.client.GetRdeTypeById(deletedId)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)
}

// testRdeCrudWithGivenType is a sub-section of Test_Rde that is focused on testing all RDE instances casuistics.
// This would be the viewpoint of a System admin as they can retrieve and manipulate RDE types.
func testRdeCrudWithGivenType(check *C, rdeType *DefinedEntityType) {

	// We are missing the mandatory field "foo" on purpose
	rdeEntityJson := []byte(`
	{
		"bar": "stringValue1",
		"prop2": {
			"subprop1": "stringValue2",
			"subprop2": [
				"stringValue3",
				"stringValue4"
			]
		}
	}`)

	var unmarshaledRdeEntityJson map[string]interface{}
	err := json.Unmarshal(rdeEntityJson, &unmarshaledRdeEntityJson)
	check.Assert(err, IsNil)

	rde, err := rdeType.CreateRde(types.DefinedEntity{
		Name:       check.TestName(),
		ExternalId: "123",
		Entity:     unmarshaledRdeEntityJson,
	}, nil)
	check.Assert(err, IsNil)
	check.Assert(rde.DefinedEntity.Name, Equals, check.TestName())
	check.Assert(*rde.DefinedEntity.State, Equals, "PRE_CREATED")

	// If we don't resolve the RDE, we cannot delete it
	err = rde.Delete()
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "RDE_ENTITY_NOT_RESOLVED"))

	// Resolution should fail as we missed to add a mandatory field
	err = rde.Resolve()
	eTag := rde.Etag
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity.State, Equals, "RESOLUTION_ERROR")
	check.Assert(eTag, Not(Equals), "")

	// We amend it
	unmarshaledRdeEntityJson["foo"] = map[string]interface{}{"key": "stringValue5"}
	err = rde.Update(types.DefinedEntity{
		Entity: unmarshaledRdeEntityJson,
	})
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity.State, Equals, "RESOLUTION_ERROR")
	check.Assert(rde.Etag, Not(Equals), "")
	check.Assert(rde.Etag, Not(Equals), eTag)
	eTag = rde.Etag

	// This time it should resolve
	err = rde.Resolve()
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity.State, Equals, "RESOLVED")
	check.Assert(rde.Etag, Not(Equals), "")
	check.Assert(rde.Etag, Not(Equals), eTag)

	// The RDE can't be deleted until rde.Resolve() is called
	AddToCleanupListOpenApi(rde.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEntities+rde.DefinedEntity.ID)

	// Delete the RDE instance now that it's resolved
	deletedId := rde.DefinedEntity.ID
	err = rde.Delete()
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity, DeepEquals, types.DefinedEntity{})

	// RDE should not exist anymore
	_, err = rdeType.GetRdeById(deletedId)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)
}

// testRdeCrudAsTenant is a sub-section of Test_Rde that is focused on testing all RDE instances casuistics without specifying the
// RDE type. This would be the viewpoint of a tenant as they can't get RDE types.
func testRdeCrudAsTenant(check *C, vendor string, namespace string, version string, vcdClient *VCDClient) {
	// We are missing the mandatory field "foo" on purpose
	rdeEntityJson := []byte(`
	{
		"bar": "stringValue1",
		"prop2": {
			"subprop1": "stringValue2",
			"subprop2": [
				"stringValue3",
				"stringValue4"
			]
		}
	}`)

	var unmarshaledRdeEntityJson map[string]interface{}
	err := json.Unmarshal(rdeEntityJson, &unmarshaledRdeEntityJson)
	check.Assert(err, IsNil)

	rde, err := vcdClient.CreateRde(vendor, namespace, version, types.DefinedEntity{
		Name:       check.TestName(),
		ExternalId: "123",
		Entity:     unmarshaledRdeEntityJson,
	}, nil)
	check.Assert(err, IsNil)
	check.Assert(rde.DefinedEntity.Name, Equals, check.TestName())
	check.Assert(*rde.DefinedEntity.State, Equals, "PRE_CREATED")

	// If we don't resolve the RDE, we cannot delete it
	err = rde.Delete()
	check.Assert(err, NotNil)
	check.Assert(true, Equals, strings.Contains(err.Error(), "RDE_ENTITY_NOT_RESOLVED"))

	// Resolution should fail as we missed to add a mandatory field
	err = rde.Resolve()
	eTag := rde.Etag
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity.State, Equals, "RESOLUTION_ERROR")
	check.Assert(eTag, Not(Equals), "")

	// We amend it
	unmarshaledRdeEntityJson["foo"] = map[string]interface{}{"key": "stringValue5"}
	err = rde.Update(types.DefinedEntity{
		Entity: unmarshaledRdeEntityJson,
	})
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity.State, Equals, "RESOLUTION_ERROR")
	check.Assert(rde.Etag, Not(Equals), "")
	check.Assert(rde.Etag, Not(Equals), eTag)
	eTag = rde.Etag

	// This time it should resolve
	err = rde.Resolve()
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity.State, Equals, "RESOLVED")
	check.Assert(rde.Etag, Not(Equals), "")
	check.Assert(rde.Etag, Not(Equals), eTag)

	// The RDE can't be deleted until rde.Resolve() is called
	AddToCleanupListOpenApi(rde.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointEntities+rde.DefinedEntity.ID)

	// Delete the RDE instance now that it's resolved
	deletedId := rde.DefinedEntity.ID
	err = rde.Delete()
	check.Assert(err, IsNil)
	check.Assert(*rde.DefinedEntity, DeepEquals, types.DefinedEntity{})
	check.Assert(rde.Etag, Equals, "")

	// RDE should not exist anymore
	_, err = vcdClient.GetRdeById(deletedId)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)
}

// loadRdeTypeSchemaFromTestResources loads the RDE schema present in the test-resources folder and unmarshals it
// into a map. Returns an error if something fails along the way.
func loadRdeTypeSchemaFromTestResources() (map[string]interface{}, error) {
	// Load the RDE type schema
	rdeFilePath := "../test-resources/rde_type.json"
	_, err := os.Stat(rdeFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to find RDE type file '%s': %s", rdeFilePath, err)
	}

	rdeFile, err := os.OpenFile(filepath.Clean(rdeFilePath), os.O_RDONLY, 0400)
	if err != nil {
		return nil, fmt.Errorf("unable to open RDE type file '%s': %s", rdeFilePath, err)
	}
	defer safeClose(rdeFile)

	rdeSchema, err := io.ReadAll(rdeFile)
	if err != nil {
		return nil, fmt.Errorf("error reading RDE type file %s: %s", rdeFilePath, err)
	}

	var unmarshaledJson map[string]interface{}
	err = json.Unmarshal(rdeSchema, &unmarshaledJson)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal RDE type file %s: %s", rdeFilePath, err)
	}

	return unmarshaledJson, nil
}
