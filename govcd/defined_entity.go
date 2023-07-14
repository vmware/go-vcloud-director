/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"time"
)

// DefinedEntityType is a type for handling Runtime Defined Entity (RDE) Type definitions.
// Note. Running a few of these operations in parallel may corrupt database in VCD (at least <= 10.4.2)
type DefinedEntityType struct {
	DefinedEntityType *types.DefinedEntityType
	client            *Client
}

// DefinedEntity represents an instance of a Runtime Defined Entity (RDE)
type DefinedEntity struct {
	DefinedEntity *types.DefinedEntity
	Etag          string // Populated by VCDClient.GetRdeById, DefinedEntityType.GetRdeById, DefinedEntity.Update
	client        *Client
}

// CreateRdeType creates a Runtime Defined Entity Type.
// Only a System administrator can create RDE Types.
func (vcdClient *VCDClient) CreateRdeType(rde *types.DefinedEntityType) (*DefinedEntityType, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes
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

// GetAllRdeTypes retrieves all Runtime Defined Entity Types. Query parameters can be supplied to perform additional filtering.
func (vcdClient *VCDClient) GetAllRdeTypes(queryParameters url.Values) ([]*DefinedEntityType, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes
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

// GetRdeType gets a Runtime Defined Entity Type by its unique combination of vendor, nss and version.
func (vcdClient *VCDClient) GetRdeType(vendor, nss, version string) (*DefinedEntityType, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("vendor==%s;nss==%s;version==%s", vendor, nss, version))
	rdeTypes, err := vcdClient.GetAllRdeTypes(queryParameters)
	if err != nil {
		return nil, err
	}

	if len(rdeTypes) == 0 {
		return nil, fmt.Errorf("%s could not find the Runtime Defined Entity Type with vendor %s, nss %s and version %s", ErrorEntityNotFound, vendor, nss, version)
	}

	if len(rdeTypes) > 1 {
		return nil, fmt.Errorf("found more than 1 Runtime Defined Entity Type with vendor %s, nss %s and version %s", vendor, nss, version)
	}

	return rdeTypes[0], nil
}

// GetRdeTypeById gets a Runtime Defined Entity Type by its ID.
func (vcdClient *VCDClient) GetRdeTypeById(id string) (*DefinedEntityType, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes
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

// Update updates the receiver Runtime Defined Entity Type with the values given by the input.
// Only a System administrator can create RDE Types.
func (rdeType *DefinedEntityType) Update(rdeTypeToUpdate types.DefinedEntityType) error {
	client := rdeType.client
	if rdeType.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity Type is empty")
	}

	if rdeTypeToUpdate.ID != "" && rdeTypeToUpdate.ID != rdeType.DefinedEntityType.ID {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity and the input ID don't match")
	}

	// Name and schema are mandatory, even when we don't want to update them, so we populate them in this situation to avoid errors
	// and make this method more user friendly.
	if rdeTypeToUpdate.Name == "" {
		rdeTypeToUpdate.Name = rdeType.DefinedEntityType.Name
	}
	if rdeTypeToUpdate.Schema == nil || len(rdeTypeToUpdate.Schema) == 0 {
		rdeTypeToUpdate.Schema = rdeType.DefinedEntityType.Schema
	}
	rdeTypeToUpdate.Version = rdeType.DefinedEntityType.Version
	rdeTypeToUpdate.Nss = rdeType.DefinedEntityType.Nss
	rdeTypeToUpdate.Vendor = rdeType.DefinedEntityType.Vendor

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes
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

// Delete deletes the receiver Runtime Defined Entity Type.
// Only a System administrator can delete RDE Types.
func (rdeType *DefinedEntityType) Delete() error {
	client := rdeType.client
	if rdeType.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity Type is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes
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

// GetAllBehaviors retrieves all the Behaviors of the receiver RDE Type.
func (rdeType *DefinedEntityType) GetAllBehaviors(queryParameters url.Values) ([]*types.Behavior, error) {
	if rdeType.DefinedEntityType.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Entity Type is empty")
	}
	return getAllBehaviors(rdeType.client, rdeType.DefinedEntityType.ID, types.OpenApiEndpointRdeTypeBehaviors, queryParameters)
}

// GetBehaviorById retrieves a unique Behavior that belongs to the receiver RDE Type and is determined by the
// input ID. The ID can be a RDE Interface Behavior ID or a RDE Type overridden Behavior ID.
func (rdeType *DefinedEntityType) GetBehaviorById(id string) (*types.Behavior, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviors
	apiVersion, err := rdeType.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := rdeType.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rdeType.DefinedEntityType.ID), id)
	if err != nil {
		return nil, err
	}

	response := types.Behavior{}
	err = rdeType.client.OpenApiGetItem(apiVersion, urlRef, nil, &response, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetBehaviorByName retrieves a unique Behavior that belongs to the receiver RDE Type and is named after
// the input.
func (rdeType *DefinedEntityType) GetBehaviorByName(name string) (*types.Behavior, error) {
	behaviors, err := rdeType.GetAllBehaviors(nil)
	if err != nil {
		return nil, fmt.Errorf("could not get the Behaviors of the Defined Entity Type with ID '%s': %s", rdeType.DefinedEntityType.ID, err)
	}
	for _, b := range behaviors {
		if b.Name == name {
			return b, nil
		}
	}
	return nil, fmt.Errorf("could not find any Behavior with name '%s' in Defined Entity Type with ID '%s': %s", name, rdeType.DefinedEntityType.ID, ErrorEntityNotFound)
}

// UpdateBehaviorOverride overrides an Interface Behavior. Only Behavior description and execution can be overridden.
// It returns the new Behavior, result of the override (with a new ID).
func (rdeType *DefinedEntityType) UpdateBehaviorOverride(behavior types.Behavior) (*types.Behavior, error) {
	if rdeType.DefinedEntityType.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Entity Type is empty")
	}
	if behavior.ID == "" {
		return nil, fmt.Errorf("ID of the Behavior to override is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviors
	apiVersion, err := rdeType.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := rdeType.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rdeType.DefinedEntityType.ID), behavior.ID)
	if err != nil {
		return nil, err
	}
	response := types.Behavior{}
	err = rdeType.client.OpenApiPutItem(apiVersion, urlRef, nil, behavior, &response, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteBehaviorOverride removes a Behavior specified by its ID from the receiver Defined Entity Type.
// The ID can be the Interface Behavior ID or the Type Behavior ID (the overridden one).
func (rdeType *DefinedEntityType) DeleteBehaviorOverride(behaviorId string) error {
	if rdeType.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Entity Type is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviors
	apiVersion, err := rdeType.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := rdeType.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rdeType.DefinedEntityType.ID), behaviorId)
	if err != nil {
		return err
	}
	err = rdeType.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// SetBehaviorAccessControls sets the given slice of BehaviorAccess to the receiver Defined Entity Type.
func (det *DefinedEntityType) SetBehaviorAccessControls(acls []*types.BehaviorAccess) error {
	if det.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Entity Type is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviorAccessControls
	apiVersion, err := det.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := det.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, det.DefinedEntityType.ID))
	if err != nil {
		return err
	}

	// Wrap it in OpenAPI pages, this endpoint requires it
	rawMessage, err := json.Marshal(acls)
	if err != nil {
		return fmt.Errorf("error setting Access controls in payload: %s", err)
	}
	payload := types.OpenApiPages{
		Values: rawMessage,
	}

	err = det.client.OpenApiPutItem(apiVersion, urlRef, nil, payload, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetAllBehaviorsAccessControls gets all the Behaviors Access Controls from the receiver DefinedEntityType.
// Query parameters can be supplied to modify pagination.
func (det *DefinedEntityType) GetAllBehaviorsAccessControls(queryParameters url.Values) ([]*types.BehaviorAccess, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviorAccessControls
	apiVersion, err := det.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := det.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, det.DefinedEntityType.ID))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.BehaviorAccess{{}}
	err = det.client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetAllRdes gets all the RDE instances of the given vendor, nss and version.
func (vcdClient *VCDClient) GetAllRdes(vendor, nss, version string, queryParameters url.Values) ([]*DefinedEntity, error) {
	return getAllRdes(&vcdClient.Client, vendor, nss, version, queryParameters)
}

// GetAllRdes gets all the RDE instances of the receiver type.
func (rdeType *DefinedEntityType) GetAllRdes(queryParameters url.Values) ([]*DefinedEntity, error) {
	return getAllRdes(rdeType.client, rdeType.DefinedEntityType.Vendor, rdeType.DefinedEntityType.Nss, rdeType.DefinedEntityType.Version, queryParameters)
}

// getAllRdes gets all the RDE instances of the given vendor, nss and version.
// Supports filtering with the given queryParameters.
func getAllRdes(client *Client, vendor, nss, version string, queryParameters url.Values) ([]*DefinedEntity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, fmt.Sprintf("%s/%s/%s", vendor, nss, version))
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
	return getRdesByName(rdeType.client, rdeType.DefinedEntityType.Vendor, rdeType.DefinedEntityType.Nss, rdeType.DefinedEntityType.Version, name)
}

// GetRdesByName gets RDE instances with the given name and the given vendor, nss and version.
// VCD allows to have many RDEs with the same name, hence this function returns a slice.
func (vcdClient *VCDClient) GetRdesByName(vendor, nss, version, name string) ([]*DefinedEntity, error) {
	return getRdesByName(&vcdClient.Client, vendor, nss, version, name)
}

// getRdesByName gets RDE instances with the given name and the given vendor, nss and version.
// VCD allows to have many RDEs with the same name, hence this function returns a slice.
func getRdesByName(client *Client, vendor, nss, version, name string) ([]*DefinedEntity, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("name==%s", name))
	rdeTypes, err := getAllRdes(client, vendor, nss, version, queryParameters)
	if err != nil {
		return nil, err
	}

	if len(rdeTypes) == 0 {
		return nil, fmt.Errorf("%s could not find the Runtime Defined Entity with name '%s'", ErrorEntityNotFound, name)
	}

	return rdeTypes, nil
}

// GetRdeById gets a Runtime Defined Entity by its ID.
// Getting a RDE by ID populates the ETag field in the returned object.
func (rdeType *DefinedEntityType) GetRdeById(id string) (*DefinedEntity, error) {
	return getRdeById(rdeType.client, id)
}

// GetRdeById gets a Runtime Defined Entity by its ID.
// Getting a RDE by ID populates the ETag field in the returned object.
func (vcdClient *VCDClient) GetRdeById(id string) (*DefinedEntity, error) {
	return getRdeById(&vcdClient.Client, id)
}

// getRdeById gets a Runtime Defined Entity by its ID.
// Getting a RDE by ID populates the ETag field in the returned object.
func getRdeById(client *Client, id string) (*DefinedEntity, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities
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

	headers, err := client.OpenApiGetItemAndHeaders(apiVersion, urlRef, nil, result.DefinedEntity, nil)
	if err != nil {
		return nil, amendRdeApiError(client, err)
	}
	result.Etag = headers.Get("Etag")

	return result, nil
}

// CreateRde creates an entity of the type of the receiver Runtime Defined Entity (RDE) type.
// The input doesn't need to specify the type ID, as it gets it from the receiver RDE type.
// The input tenant context allows to create the RDE in a given org if the creator is a System admin.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func (rdeType *DefinedEntityType) CreateRde(entity types.DefinedEntity, tenantContext *TenantContext) (*DefinedEntity, error) {
	entity.EntityType = rdeType.DefinedEntityType.ID
	err := createRde(rdeType.client, entity, tenantContext)
	if err != nil {
		return nil, err
	}
	return pollPreCreatedRde(rdeType.client, rdeType.DefinedEntityType.Vendor, rdeType.DefinedEntityType.Nss, rdeType.DefinedEntityType.Version, entity.Name, 5)
}

// CreateRde creates an entity of the type of the given vendor, nss and version.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func (vcdClient *VCDClient) CreateRde(vendor, nss, version string, entity types.DefinedEntity, tenantContext *TenantContext) (*DefinedEntity, error) {
	entity.EntityType = fmt.Sprintf("urn:vcloud:type:%s:%s:%s", vendor, nss, version)
	err := createRde(&vcdClient.Client, entity, tenantContext)
	if err != nil {
		return nil, err
	}
	return pollPreCreatedRde(&vcdClient.Client, vendor, nss, version, entity.Name, 5)
}

// CreateRde creates an entity of the type of the receiver Runtime Defined Entity (RDE) type.
// The input doesn't need to specify the type ID, as it gets it from the receiver RDE type. If it is specified anyway,
// it must match the type ID of the receiver RDE type.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func createRde(client *Client, entity types.DefinedEntity, tenantContext *TenantContext) error {
	if entity.EntityType == "" {
		return fmt.Errorf("ID of the Runtime Defined Entity type is empty")
	}

	if entity.Entity == nil || len(entity.Entity) == 0 {
		return fmt.Errorf("the entity JSON is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, entity.EntityType)
	if err != nil {
		return err
	}

	_, err = client.OpenApiPostItemAsyncWithHeaders(apiVersion, urlRef, nil, entity, getTenantContextHeader(tenantContext))
	if err != nil {
		return err
	}
	return nil
}

// pollPreCreatedRde polls VCD for a given amount of tries, to search for the RDE in state PRE_CREATED
// that corresponds to the given vendor, nss, version and name.
// This function can be useful on RDE creation, as VCD just returns a task that remains at 1% until the RDE is resolved,
// hence one needs to re-fetch the recently created RDE manually.
func pollPreCreatedRde(client *Client, vendor, nss, version, name string, tries int) (*DefinedEntity, error) {
	var rdes []*DefinedEntity
	var err error
	for i := 0; i < tries; i++ {
		rdes, err = getRdesByName(client, vendor, nss, version, name)
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
// Resolving a RDE populates the ETag field in the receiver object.
func (rde *DefinedEntity) Resolve() error {
	client := rde.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesResolve
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rde.DefinedEntity.ID))
	if err != nil {
		return err
	}

	headers, err := client.OpenApiPostItemAndGetHeaders(apiVersion, urlRef, nil, nil, rde.DefinedEntity, nil)
	if err != nil {
		return amendRdeApiError(client, err)
	}
	rde.Etag = headers.Get("Etag")

	return nil
}

// Update updates the receiver Runtime Defined Entity with the values given by the input. This method is useful
// if rde.Resolve() failed and a JSON entity change is needed.
// Updating a RDE populates the ETag field in the receiver object.
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

	if rde.Etag == "" {
		// We need to get an Etag to perform the update
		retrievedRde, err := getRdeById(rde.client, rde.DefinedEntity.ID)
		if err != nil {
			return err
		}
		if retrievedRde.Etag == "" {
			return fmt.Errorf("could not retrieve a valid Etag to perform an update to RDE %s", retrievedRde.DefinedEntity.ID)
		}
		rde.Etag = retrievedRde.Etag
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, rde.DefinedEntity.ID)
	if err != nil {
		return amendRdeApiError(client, err)
	}

	headers, err := client.OpenApiPutItemAndGetHeaders(apiVersion, urlRef, nil, rdeToUpdate, rde.DefinedEntity, map[string]string{"If-Match": rde.Etag})
	if err != nil {
		return err
	}
	rde.Etag = headers.Get("Etag")

	return nil
}

// Delete deletes the receiver Runtime Defined Entity.
func (rde *DefinedEntity) Delete() error {
	client := rde.client

	if rde.DefinedEntity.ID == "" {
		return fmt.Errorf("ID of the receiver Runtime Defined Entity is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities
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
		return amendRdeApiError(client, err)
	}

	rde.DefinedEntity = &types.DefinedEntity{}
	rde.Etag = ""
	return nil
}

// InvokeBehavior calls a Behavior identified by the given ID with the given execution parameters.
// Returns the invocation result as a raw string.
func (rde *DefinedEntity) InvokeBehavior(behaviorId string, invocation types.BehaviorInvocation) (string, error) {
	client := rde.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesBehaviorsInvocations
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return "", err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, rde.DefinedEntity.ID, behaviorId))
	if err != nil {
		return "", err
	}

	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, invocation)
	if err != nil {
		return "", err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return "", err
	}

	if task.Task.Result == nil {
		return "", fmt.Errorf("the Task '%s' returned an empty Result content", task.Task.ID)
	}

	return task.Task.Result.ResultContent.Text, nil
}

// InvokeBehaviorAndMarshal calls a Behavior identified by the given ID with the given execution parameters.
// Returns the invocation result marshaled with the input object.
func (rde *DefinedEntity) InvokeBehaviorAndMarshal(behaviorId string, invocation types.BehaviorInvocation, output interface{}) error {
	result, err := rde.InvokeBehavior(behaviorId, invocation)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(result), &output)
	if err != nil {
		return fmt.Errorf("error marshaling the invocation result '%s': %s", result, err)
	}

	return nil
}
