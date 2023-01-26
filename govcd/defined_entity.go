/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"time"
)

// DefinedEntityType is a type for handling Runtime Defined Entity (RDE) type definitions
type DefinedEntityType struct {
	DefinedEntityType *types.DefinedEntityType
	client            *Client
}

// DefinedEntity represents an instance of a Runtime Defined Entity (RDE)
type DefinedEntity struct {
	DefinedEntity *types.DefinedEntity
	client        *Client
}

// CreateRdeType creates a Runtime Defined Entity type.
// Only System administrator can create RDE types.
func (vcdClient *VCDClient) CreateRdeType(rde *types.DefinedEntityType) (*DefinedEntityType, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("creating Runtime Defined Entity types requires System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &DefinedEntityType{
		DefinedEntityType: &types.DefinedEntityType{},
		client:            &vcdClient.Client,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, rde, result.DefinedEntityType, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetAllRdeTypes retrieves all Runtime Defined Entity types. Query parameters can be supplied to perform additional filtering.
// Only System administrator can retrieve RDE types.
func (vcdClient *VCDClient) GetAllRdeTypes(queryParameters url.Values) ([]*DefinedEntityType, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("getting Runtime Defined Entity types requires System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.DefinedEntityType{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into DefinedEntityType types with client
	returnRDEs := make([]*DefinedEntityType, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnRDEs[sliceIndex] = &DefinedEntityType{
			DefinedEntityType: typeResponses[sliceIndex],
			client:            &vcdClient.Client,
		}
	}

	return returnRDEs, nil
}

// GetRdeType gets a Runtime Defined Entity type by its unique combination of vendor, namespace and version.
// Only System administrator can retrieve RDE types.
func (vcdClient *VCDClient) GetRdeType(vendor, namespace, version string) (*DefinedEntityType, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("getting Runtime Defined Entity types requires System user")
	}

	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("vendor==%s;nss==%s;version==%s", vendor, namespace, version))
	rdeTypes, err := vcdClient.GetAllRdeTypes(queryParameters)
	if err != nil {
		return nil, err
	}

	if len(rdeTypes) == 0 {
		return nil, fmt.Errorf("%s could not find the Runtime Defined Entity type with vendor %s, namespace %s and version %s", ErrorEntityNotFound, vendor, namespace, version)
	}

	if len(rdeTypes) > 1 {
		return nil, fmt.Errorf("found more than 1 Runtime Defined Entity type with vendor %s, namespace %s and version %s", vendor, namespace, version)
	}

	return rdeTypes[0], nil
}

// GetRdeTypeById gets a Runtime Defined Entity type by its ID
// Only System administrator can retrieve RDE types.
func (vcdClient *VCDClient) GetRdeTypeById(id string) (*DefinedEntityType, error) {
	client := vcdClient.Client
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("getting Runtime Defined Entity types requires System user")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	result := &DefinedEntityType{
		DefinedEntityType: &types.DefinedEntityType{},
		client:            &vcdClient.Client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, result.DefinedEntityType, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Update updates the receiver Runtime Defined Entity type with the values given by the input.
// Only System administrator can update RDE types.
func (rdeType *DefinedEntityType) Update(rdeTypeToUpdate types.DefinedEntityType) error {
	client := rdeType.client
	if !client.IsSysAdmin {
		return fmt.Errorf("updating Runtime Defined Entity types requires System user")
	}

	if rdeType.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity type is empty")
	}

	if rdeTypeToUpdate.ID != "" && rdeTypeToUpdate.ID != rdeType.DefinedEntityType.ID {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity and the input ID don't match")
	}

	// Name and schema are mandatory, despite we don't want to update them, so we populate them in this situation to avoid errors
	// and make this method more user friendly.
	if rdeTypeToUpdate.Name == "" {
		rdeTypeToUpdate.Name = rdeType.DefinedEntityType.Name
	}
	if rdeTypeToUpdate.Schema == nil || len(rdeTypeToUpdate.Schema) == 0 {
		rdeTypeToUpdate.Schema = rdeType.DefinedEntityType.Schema
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, rdeType.DefinedEntityType.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, rdeTypeToUpdate, rdeType.DefinedEntityType, nil)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the receiver Runtime Defined Entity type.
// Only System administrator can delete RDE types.
func (rdeType *DefinedEntityType) Delete() error {
	client := rdeType.client
	if !client.IsSysAdmin {
		return fmt.Errorf("deleting Runtime Defined Entity types requires System user")
	}

	if rdeType.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity type is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, rdeType.DefinedEntityType.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	rdeType.DefinedEntityType = &types.DefinedEntityType{}
	return nil
}

// GetAllRdes gets all the RDE instances of the given vendor, namespace and version.
func (vcdClient *VCDClient) GetAllRdes(vendor, namespace, version string, queryParameters url.Values) ([]*DefinedEntity, error) {
	return getAllRdes(&vcdClient.Client, vendor, namespace, version, queryParameters)
}

// GetAllRdes gets all the RDE instances of the receiver type.
func (rdeType *DefinedEntityType) GetAllRdes(queryParameters url.Values) ([]*DefinedEntity, error) {
	return getAllRdes(rdeType.client, rdeType.DefinedEntityType.Vendor, rdeType.DefinedEntityType.Namespace, rdeType.DefinedEntityType.Version, queryParameters)
}

// getAllRdes gets all the RDE instances of the given vendor, namespace and version.
// Supports filtering with the given queryParameters.
func getAllRdes(client *Client, vendor, namespace, version string, queryParameters url.Values) ([]*DefinedEntity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntitiesTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, fmt.Sprintf("%s/%s/%s", vendor, namespace, version))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.DefinedEntity{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into DefinedEntityType types with client
	returnRDEs := make([]*DefinedEntity, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnRDEs[sliceIndex] = &DefinedEntity{
			DefinedEntity: typeResponses[sliceIndex],
			client:        client,
		}
	}

	return returnRDEs, nil
}

// GetRdesByName gets RDE instances with the given name that belongs to the receiver type.
// VCD allows to have many RDEs with the same name, hence this function returns a slice.
func (rdeType *DefinedEntityType) GetRdesByName(name string) ([]*DefinedEntity, error) {
	return getRdesByName(rdeType.client, rdeType.DefinedEntityType.Vendor, rdeType.DefinedEntityType.Namespace, rdeType.DefinedEntityType.Version, name)
}

// GetRdesByName gets RDE instances with the given name and the given vendor, namespace and version.
// VCD allows to have many RDEs with the same name, hence this function returns a slice.
func (vcdClient *VCDClient) GetRdesByName(vendor, namespace, version, name string) ([]*DefinedEntity, error) {
	return getRdesByName(&vcdClient.Client, vendor, namespace, version, name)
}

// getRdesByName gets RDE instances with the given name and the given vendor, namespace and version.
// VCD allows to have many RDEs with the same name, hence this function returns a slice.
func getRdesByName(client *Client, vendor, namespace, version, name string) ([]*DefinedEntity, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("name==%s", name))
	rdeTypes, err := getAllRdes(client, vendor, namespace, version, queryParameters)
	if err != nil {
		return nil, err
	}

	if len(rdeTypes) == 0 {
		return nil, fmt.Errorf("%s could not find the Runtime Defined Entity with name '%s'", ErrorEntityNotFound, name)
	}

	return rdeTypes, nil
}

// GetRdeById gets a Runtime Defined Entity by its ID
func (rdeType *DefinedEntityType) GetRdeById(id string) (*DefinedEntity, error) {
	return getRdeById(rdeType.client, id)
}

// GetRdeById gets a Runtime Defined Entity by its ID
func (vcdClient *VCDClient) GetRdeById(id string) (*DefinedEntity, error) {
	return getRdeById(&vcdClient.Client, id)
}

// getRdeById gets a Runtime Defined Entity by its ID
func getRdeById(client *Client, id string) (*DefinedEntity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntities
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	result := &DefinedEntity{
		DefinedEntity: &types.DefinedEntity{},
		client:        client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, result.DefinedEntity, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateRde creates an entity of the type of the receiver Runtime Defined Entity (RDE) type.
// The input doesn't need to specify the type ID, as it gets it from the receiver RDE type. If it is specified anyway,
// it must match the type ID of the receiver RDE type.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func (rdeType *DefinedEntityType) CreateRde(entity types.DefinedEntity) (*DefinedEntity, error) {
	err := createRde(rdeType.client, rdeType.DefinedEntityType.ID, entity)
	if err != nil {
		return nil, err
	}
	return pollPreCreatedRde(rdeType.client, rdeType.DefinedEntityType.Vendor, rdeType.DefinedEntityType.Namespace, rdeType.DefinedEntityType.ID, entity.Name, 5)
}

// CreateRde creates an entity of the type of the given vendor, namespace and version.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func (vcdClient *VCDClient) CreateRde(vendor, namespace, version string, entity types.DefinedEntity) (*DefinedEntity, error) {
	rdeTypeId := fmt.Sprintf("urn:vcloud:type:%s:%s:%s", vendor, namespace, version)
	err := createRde(&vcdClient.Client, rdeTypeId, entity)
	if err != nil {
		return nil, err
	}
	return pollPreCreatedRde(&vcdClient.Client, vendor, namespace, version, entity.Name, 5)
}

// CreateRde creates an entity of the type of the receiver Runtime Defined Entity (RDE) type.
// The input doesn't need to specify the type ID, as it gets it from the receiver RDE type. If it is specified anyway,
// it must match the type ID of the receiver RDE type.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func createRde(client *Client, rdeTypeId string, entity types.DefinedEntity) error {
	if rdeTypeId == "" {
		return fmt.Errorf("ID of the Runtime Defined Entity type is empty")
	}

	if entity.EntityType != "" && entity.EntityType != rdeTypeId {
		return fmt.Errorf("ID of the Runtime Defined Entity type '%s' doesn't match with the one to create '%s'", rdeTypeId, entity.EntityType)
	}

	if entity.Entity == nil || len(entity.Entity) == 0 {
		return fmt.Errorf("the entity JSON is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, rdeTypeId)
	if err != nil {
		return err
	}

	_, err = client.OpenApiPostItemAsync(apiVersion, urlRef, nil, entity)
	if err != nil {
		return err
	}
	return nil
}

// pollPreCreatedRde polls VCD for a given amount of tries, to search for the RDE in state PRE_CREATED
// that corresponds to the given vendor, namespace, version and name.
// This function can be useful on RDE creation, as VCD just returns a task that remains at 1% until the RDE is resolved,
// hence one needs to re-fetch the recently created RDE manually.
func pollPreCreatedRde(client *Client, vendor, namespace, version, name string, tries int) (*DefinedEntity, error) {
	var rdes []*DefinedEntity
	var err error
	for i := 0; i < tries; i++ {
		rdes, err = getRdesByName(client, vendor, namespace, version, name)
		if err == nil {
			for _, rde := range rdes {
				// This doesn't really guarantee that the chosen RDE is the one we want, but there's no other way of
				// fine-graining
				if rde.DefinedEntity.State != nil && *rde.DefinedEntity.State == "PRE_CREATED" {
					return rde, nil
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("could not create RDE, failed during retrieval after creation: %s", err)
}

// Resolve needs to be called after an RDE is successfully created. It makes the receiver RDE usable if the JSON entity
// is valid, reaching a state of RESOLVED. If it fails, the state will be RESOLUTION_ERROR,
// and it will need to Update the JSON entity.
func (rde *DefinedEntity) Resolve() error {
	client := rde.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntitiesResolve
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rde.DefinedEntity.ID))
	if err != nil {
		return err
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, nil, rde.DefinedEntity, nil)
	if err != nil {
		return err
	}

	return nil
}

// Update updates the receiver Runtime Defined Entity with the values given by the input. This method is useful
// if rde.Resolve() failed and a JSON entity change is needed.
// Only System administrator can update RDEs.
func (rde *DefinedEntity) Update(rdeToUpdate types.DefinedEntity) error {
	client := rde.client

	if rde.DefinedEntity.ID == "" {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity is empty")
	}

	// Name is mandatory, despite we don't want to update it, so we populate it in this situation to avoid errors
	// and make this method more user friendly.
	if rdeToUpdate.Name == "" {
		rdeToUpdate.Name = rde.DefinedEntity.Name
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntities
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, rde.DefinedEntity.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, rdeToUpdate, rde.DefinedEntity, nil)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the receiver Runtime Defined Entity.
// Only System administrator can delete RDEs.
func (rde *DefinedEntity) Delete() error {
	client := rde.client

	if rde.DefinedEntity.ID == "" {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEntities
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, rde.DefinedEntity.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	rde.DefinedEntity = &types.DefinedEntity{}
	return nil
}
