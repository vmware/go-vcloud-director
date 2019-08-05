// +build api functional catalog org extnetwork vm vdc system user ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"reflect"

	. "gopkg.in/check.v1"
)

// An interface used to test the result of Get* methods
type GenericEntity interface {
	name() string // returns the entity name
	id() string   // returns the entity ID
}

// Defines a generic getter to test all Get* methods
type GetterTestDefinition struct {
	parentName       string                                    // Name of the parent entity
	parentType       string                                    // Type of the parent entity
	entityType       string                                    // Type of the entity to retrieve (Must match the type name)
	entityName       string                                    // Name of the entity to retrieve
	getterPrefix     string                                    // Base name for getter functions
	getterByName     func(string, bool) (GenericEntity, error) // A function that retrieves the entity by name
	getterById       func(string, bool) (GenericEntity, error) // A function that retrieves the entity by ID
	getterByNameOrId func(string, bool) (GenericEntity, error) // A function that retrieves the entity by name or ID
}

// Satisfy interface GenericEntity
func (adminCat *AdminCatalog) name() string { return adminCat.AdminCatalog.Name }
func (adminCat *AdminCatalog) id() string   { return adminCat.AdminCatalog.ID }

func (adminOrg *AdminOrg) name() string { return adminOrg.AdminOrg.Name }
func (adminOrg *AdminOrg) id() string   { return adminOrg.AdminOrg.ID }

func (vdc *AdminVdc) name() string { return vdc.AdminVdc.Name }
func (vdc *AdminVdc) id() string   { return vdc.AdminVdc.ID }

func (cat *Catalog) name() string { return cat.Catalog.Name }
func (cat *Catalog) id() string   { return cat.Catalog.ID }

func (catItem *CatalogItem) name() string { return catItem.CatalogItem.Name }
func (catItem *CatalogItem) id() string   { return catItem.CatalogItem.ID }

func (extNet *ExternalNetwork) name() string { return extNet.ExternalNetwork.Name }
func (extNet *ExternalNetwork) id() string   { return extNet.ExternalNetwork.ID }

func (org *Org) name() string { return org.Org.Name }
func (org *Org) id() string   { return org.Org.ID }

func (orgUser *OrgUser) name() string { return orgUser.User.Name }
func (orgUser *OrgUser) id() string   { return orgUser.User.ID }

func (vdc *Vdc) name() string { return vdc.Vdc.Name }
func (vdc *Vdc) id() string   { return vdc.Vdc.ID }

// Semi-generic tests that check the complete set of Get methods for an entity
// GetEntityByName
// GetEntityById
// GetEntityByNameOrId (using name or Id)
// Get invalid name or ID
// To use this function, the entity must satisfy the interface GenericEntity
// and within the caller it must define the getter functions
func (vcd *TestVCD) testFinderGetGenericEntity(def GetterTestDefinition, check *C) {
	entityName := def.entityName
	if entityName == "" {
		check.Skip(fmt.Sprintf("testFinderGetGenericEntity: %s name not given.", def.entityType))
		return
	}

	if def.getterPrefix == "" {
		def.getterPrefix = def.entityType
	}
	if def.parentType == "" {
		check.Skip("Field parentType was not filled.")
	}
	if testVerbose {
		fmt.Printf("testFinderGetGenericEntity: %s %s getting %s \n", def.parentType, def.parentName, def.entityType)
	}

	// 1. Get the entity by name

	if testVerbose {
		fmt.Printf("#Testing %s.Get%sByName\n", def.parentType, def.getterPrefix)
	}
	ge, err := def.getterByName(entityName, false)
	entity1 := ge.(GenericEntity)
	if err != nil {
		check.Skip(fmt.Sprintf("testFinderGetGenericEntity: %s %s not found.", def.entityType, def.entityName))
		return
	}

	wantedType := fmt.Sprintf("*govcd.%s", def.entityType)
	if testVerbose {
		fmt.Printf("# Detected entity type %s\n", reflect.TypeOf(entity1))
	}

	check.Assert(fmt.Sprintf("%s", reflect.TypeOf(entity1)), Equals, wantedType)

	check.Assert(entity1, NotNil)
	check.Assert(entity1.name(), Equals, entityName)
	entityId := entity1.id()

	// 2. Get the entity by ID
	if testVerbose {
		fmt.Printf("#Testing %s.Get%sById\n", def.parentType, def.getterPrefix)
	}
	ge, err = def.getterById(entityId, false)
	entity2 := ge.(GenericEntity)
	check.Assert(err, IsNil)
	check.Assert(entity2, NotNil)
	check.Assert(entity2.name(), Equals, entityName)
	check.Assert(entity2.id(), Equals, entityId)
	check.Assert(fmt.Sprintf("%s", reflect.TypeOf(entity2)), Equals, wantedType)

	// 3. Get the entity by Name or ID, using a known ID
	if testVerbose {
		fmt.Printf("#Testing %s.Get%sByNameOrId\n", def.parentType, def.getterPrefix)
	}
	ge, err = def.getterByNameOrId(entityId, false)
	entity3 := ge.(GenericEntity)
	check.Assert(err, IsNil)
	check.Assert(entity3, NotNil)
	check.Assert(entity3.name(), Equals, entityName)
	check.Assert(entity3.id(), Equals, entityId)
	check.Assert(fmt.Sprintf("%s", reflect.TypeOf(entity3)), Equals, wantedType)

	// 4. Get the entity by Name or ID, using the entity name
	if testVerbose {
		fmt.Printf("#Testing %s.Get%sByNameOrId\n", def.parentType, def.getterPrefix)
	}
	ge, err = def.getterByNameOrId(entityName, false)
	entity4 := ge.(GenericEntity)
	check.Assert(err, IsNil)
	check.Assert(entity4, NotNil)
	check.Assert(entity4.name(), Equals, entityName)
	check.Assert(entity4.id(), Equals, entityId)
	check.Assert(fmt.Sprintf("%s", reflect.TypeOf(entity4)), Equals, wantedType)

	// 5. Attempting a search by name with an invalid name
	if testVerbose {
		fmt.Printf("#Testing %s.Get%sByName (invalid name)\n", def.parentType, def.getterPrefix)
	}
	ge, err = def.getterByName(INVALID_NAME, false)
	entity5 := ge.(GenericEntity)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(entity5, IsNil)

	// 6. Attempting a search by name or ID with an invalid name
	if testVerbose {
		fmt.Printf("#Testing %s.Get%sByNameOrId (invalid name)\n", def.parentType, def.getterPrefix)
	}
	ge, err = def.getterByNameOrId(INVALID_NAME, false)
	entity6 := ge.(GenericEntity)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(entity6, IsNil)

	// 7. Attempting a search by ID with an invalid ID
	if testVerbose {
		fmt.Printf("#Testing %s.Get%sById (invalid ID)\n", def.parentType, def.getterPrefix)
	}
	ge, err = def.getterById(invalidEntityId, false)
	entity7 := ge.(GenericEntity)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(entity7, IsNil)

	// 8. Attempting a search by name or ID with an invalid ID
	if testVerbose {
		fmt.Printf("#Testing %s.Get%sByNameOrId (invalid ID)\n", def.parentType, def.getterPrefix)
	}
	ge, err = def.getterByNameOrId(invalidEntityId, false)
	entity8 := ge.(GenericEntity)
	check.Assert(err, NotNil)
	check.Assert(IsNotFound(err), Equals, true)
	check.Assert(entity8, IsNil)
}
