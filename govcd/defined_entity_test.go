//go:build functional || openapi || rde || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
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

// Test_RdeAndRdeType tests the CRUD operations for the RDE Type with both System administrator and a tenant user.
func (vcd *TestVCD) Test_RdeAndRdeType(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	for _, endpoint := range []string{
		types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes,
		types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesResolve,
		types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities,
	} {
		skipOpenApiEndpointTest(vcd, check, endpoint)
	}

	if len(vcd.config.Tenants) == 0 {
		check.Skip("skipping as there is no configured tenant users")
	}

	// Creates the clients for the System admin and the Tenant user
	systemAdministratorClient := vcd.client
	tenantUserClient := NewVCDClient(vcd.client.Client.VCDHREF, true)
	err := tenantUserClient.Authenticate(vcd.config.Tenants[0].User, vcd.config.Tenants[0].Password, vcd.config.Tenants[0].SysOrg)
	check.Assert(err, IsNil)

	unmarshaledRdeTypeSchema, err := loadRdeTypeSchemaFromTestResources()
	check.Assert(err, IsNil)
	check.Assert(true, Equals, len(unmarshaledRdeTypeSchema) > 0)

	// First, it checks how many exist already, as VCD contains some pre-defined ones.
	allRdeTypesBySystemAdmin, err := systemAdministratorClient.GetAllRdeTypes(nil)
	check.Assert(err, IsNil)
	alreadyPresentRdes := len(allRdeTypesBySystemAdmin)

	// For the tenant, it returns 0 RDE Types, but no error.
	allRdeTypesByTenant, err := tenantUserClient.GetAllRdeTypes(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allRdeTypesByTenant), Equals, 0)

	// Then we create a new RDE Type with System administrator.
	// Can't put check.TestName() in nss due to a bug in VCD 10.4.1 that causes RDEs to fail on GET once created with special characters like "."
	vendor := "vmware"
	nss := strings.ReplaceAll(check.TestName()+"name", ".", "")
	version := "1.2.3"
	rdeTypeToCreate := &types.DefinedEntityType{
		Name:        check.TestName(),
		Nss:         nss,
		Version:     version,
		Description: "Description of " + check.TestName(),
		Schema:      unmarshaledRdeTypeSchema,
		Vendor:      vendor,
		Interfaces:  []string{"urn:vcloud:interface:vmware:k8s:1.0.0"},
	}
	createdRdeType, err := systemAdministratorClient.CreateRdeType(rdeTypeToCreate)
	check.Assert(err, IsNil)
	check.Assert(createdRdeType, NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Name, Equals, rdeTypeToCreate.Name)
	check.Assert(createdRdeType.DefinedEntityType.Nss, Equals, rdeTypeToCreate.Nss)
	check.Assert(createdRdeType.DefinedEntityType.Version, Equals, rdeTypeToCreate.Version)
	check.Assert(createdRdeType.DefinedEntityType.Schema, NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Schema["type"], Equals, "object")
	check.Assert(createdRdeType.DefinedEntityType.Schema["definitions"], NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Schema["required"], NotNil)
	check.Assert(createdRdeType.DefinedEntityType.Schema["properties"], NotNil)
	AddToCleanupListOpenApi(createdRdeType.DefinedEntityType.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntityTypes+createdRdeType.DefinedEntityType.ID)

	// Tenants can't create RDE Types
	nilRdeType, err := tenantUserClient.CreateRdeType(&types.DefinedEntityType{
		Name:    check.TestName(),
		Nss:     "notworking",
		Version: "4.5.6",
		Schema:  unmarshaledRdeTypeSchema,
		Vendor:  "willfail",
	})
	check.Assert(err, NotNil)
	check.Assert(nilRdeType, IsNil)
	check.Assert(strings.Contains(err.Error(), "ACCESS_TO_RESOURCE_IS_FORBIDDEN"), Equals, true)

	// Assign rights to the tenant user, so it can perform following operations.
	// We don't need to clean the rights afterwards as deleting the RDE Type deletes the associated bundle
	// with its rights.
	role, err := systemAdministratorClient.Client.GetGlobalRoleByName("Organization Administrator")
	check.Assert(err, IsNil)
	check.Assert(role, NotNil)

	rightsBundleName := fmt.Sprintf("%s:%s Entitlement", vendor, nss)
	rightsBundle, err := systemAdministratorClient.Client.GetRightsBundleByName(rightsBundleName)
	check.Assert(err, IsNil)
	check.Assert(rightsBundle, NotNil)

	err = rightsBundle.PublishAllTenants()
	check.Assert(err, IsNil)

	rights, err := rightsBundle.GetRights(nil)
	check.Assert(err, IsNil)
	check.Assert(len(rights), Not(Equals), 0)

	var rightsToAdd []types.OpenApiReference
	for _, right := range rights {
		if strings.Contains(right.Name, fmt.Sprintf("%s:%s", vendor, nss)) {
			rightsToAdd = append(rightsToAdd, types.OpenApiReference{
				Name: right.Name,
				ID:   right.ID,
			})
		}
	}
	check.Assert(rightsToAdd, NotNil)
	check.Assert(len(rightsToAdd), Not(Equals), 0)

	err = role.AddRights(rightsToAdd)
	check.Assert(err, IsNil)

	// As we created a new RDE Type, we check the new count is correct in both System admin and Tenant user
	allRdeTypesBySystemAdmin, err = systemAdministratorClient.GetAllRdeTypes(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allRdeTypesBySystemAdmin), Equals, alreadyPresentRdes+1)

	// Count is 1 for tenant user as it can only retrieve the created type as per the assigned rights above.
	allRdeTypesByTenant, err = tenantUserClient.GetAllRdeTypes(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allRdeTypesByTenant), Equals, 1)

	// Test the multiple ways of getting a RDE Types in both users.
	obtainedRdeTypeBySysAdmin, err := systemAdministratorClient.GetRdeTypeById(createdRdeType.DefinedEntityType.ID)
	check.Assert(err, IsNil)
	check.Assert(*obtainedRdeTypeBySysAdmin.DefinedEntityType, DeepEquals, *createdRdeType.DefinedEntityType)

	// The RDE Type retrieved by the tenant should be the same as the retrieved by Sysadmin
	obtainedRdeTypeByTenant, err := tenantUserClient.GetRdeTypeById(createdRdeType.DefinedEntityType.ID)
	check.Assert(err, IsNil)
	check.Assert(*obtainedRdeTypeByTenant.DefinedEntityType, DeepEquals, *obtainedRdeTypeBySysAdmin.DefinedEntityType)

	obtainedRdeTypeBySysAdmin, err = systemAdministratorClient.GetRdeType(createdRdeType.DefinedEntityType.Vendor, createdRdeType.DefinedEntityType.Nss, createdRdeType.DefinedEntityType.Version)
	check.Assert(err, IsNil)
	check.Assert(*obtainedRdeTypeBySysAdmin.DefinedEntityType, DeepEquals, *obtainedRdeTypeBySysAdmin.DefinedEntityType)

	// The RDE Type retrieved by the tenant should be the same as the retrieved by Sysadmin
	obtainedRdeTypeByTenant, err = tenantUserClient.GetRdeType(createdRdeType.DefinedEntityType.Vendor, createdRdeType.DefinedEntityType.Nss, createdRdeType.DefinedEntityType.Version)
	check.Assert(err, IsNil)
	check.Assert(*obtainedRdeTypeByTenant.DefinedEntityType, DeepEquals, *obtainedRdeTypeBySysAdmin.DefinedEntityType)

	// We don't want to update the name nor the schema. It should populate them from the receiver object automatically
	err = obtainedRdeTypeBySysAdmin.Update(types.DefinedEntityType{
		Description: rdeTypeToCreate.Description + "UpdatedByAdmin",
	})
	check.Assert(err, IsNil)
	check.Assert(obtainedRdeTypeBySysAdmin.DefinedEntityType.Description, Equals, rdeTypeToCreate.Description+"UpdatedByAdmin")

	testRdeCrudWithGivenType(check, obtainedRdeTypeBySysAdmin)
	testRdeCrudAsTenant(check, obtainedRdeTypeByTenant.DefinedEntityType.Vendor, obtainedRdeTypeByTenant.DefinedEntityType.Nss, obtainedRdeTypeByTenant.DefinedEntityType.Version, vcd.client)

	// We delete it with Sysadmin
	deletedId := createdRdeType.DefinedEntityType.ID
	err = createdRdeType.Delete()
	check.Assert(err, IsNil)
	check.Assert(*createdRdeType.DefinedEntityType, DeepEquals, types.DefinedEntityType{})

	_, err = systemAdministratorClient.GetRdeTypeById(deletedId)
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
	// The RDE can be automatically deleted now as rde.Resolve() was called successfully
	AddToCleanupListOpenApi(rde.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+rde.DefinedEntity.ID)

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
	AddToCleanupListOpenApi(rde.DefinedEntity.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeEntities+rde.DefinedEntity.ID)

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
