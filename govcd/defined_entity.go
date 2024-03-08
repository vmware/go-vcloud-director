/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"strings"
)

const (
	labelDefinedEntity            = "Defined Entity"
	labelDefinedEntityType        = "Defined Entity Type"
	labelRdeBehavior              = "RDE Behavior"
	labelRdeBehaviorOverride      = "RDE Behavior Override"
	labelRdeBehaviorAccessControl = "RDE Behavior Access Control"
)

// DefinedEntityType is a type for handling Runtime Defined Entity (RDE) Type definitions.
// Note. Running a few of these operations in parallel may corrupt database in VCD (at least <= 10.4.2)
type DefinedEntityType struct {
	DefinedEntityType *types.DefinedEntityType
	client            *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (d DefinedEntityType) wrap(inner *types.DefinedEntityType) *DefinedEntityType {
	d.DefinedEntityType = inner
	return &d
}

// DefinedEntity represents an instance of a Runtime Defined Entity (RDE)
type DefinedEntity struct {
	DefinedEntity *types.DefinedEntity
	Etag          string // Populated by VCDClient.GetRdeById, DefinedEntityType.GetRdeById, DefinedEntity.Update
	client        *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (d DefinedEntity) wrap(inner *types.DefinedEntity) *DefinedEntity {
	d.DefinedEntity = inner
	return &d
}

// CreateRdeType creates a Runtime Defined Entity Type.
// Only a System administrator can create RDE Types.
func (vcdClient *VCDClient) CreateRdeType(rde *types.DefinedEntityType) (*DefinedEntityType, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes,
		entityLabel: labelDefinedEntityType,
	}
	outerType := DefinedEntityType{client: &vcdClient.Client}
	return createOuterEntity(&vcdClient.Client, outerType, c, rde)
}

// GetAllRdeTypes retrieves all Runtime Defined Entity Types. Query parameters can be supplied to perform additional filtering.
func (vcdClient *VCDClient) GetAllRdeTypes(queryParameters url.Values) ([]*DefinedEntityType, error) {
	return getAllRdeTypes(&vcdClient.Client, queryParameters)
}

// getAllRdeTypes retrieves all Runtime Defined Entity Types. Query parameters can be supplied to perform additional filtering.
func getAllRdeTypes(client *Client, queryParameters url.Values) ([]*DefinedEntityType, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes,
		entityLabel:     labelDefinedEntityType,
		queryParameters: queryParameters,
	}

	outerType := DefinedEntityType{client: client}
	return getAllOuterEntities[DefinedEntityType, types.DefinedEntityType](client, outerType, c)
}

// GetRdeType gets a Runtime Defined Entity Type by its unique combination of vendor, nss and version.
func (vcdClient *VCDClient) GetRdeType(vendor, nss, version string) (*DefinedEntityType, error) {
	return getRdeType(&vcdClient.Client, vendor, nss, version)
}

// getRdeType gets a Runtime Defined Entity Type by its unique combination of vendor, nss and version.
func getRdeType(client *Client, vendor, nss, version string) (*DefinedEntityType, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("vendor==%s;nss==%s;version==%s", vendor, nss, version))
	rdeTypes, err := getAllRdeTypes(client, queryParameters)
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
	c := crudConfig{
		entityLabel:    labelDefinedEntityType,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes,
		endpointParams: []string{id},
	}

	outerType := DefinedEntityType{client: &vcdClient.Client}
	return getOuterEntity[DefinedEntityType, types.DefinedEntityType](&vcdClient.Client, outerType, c)
}

// Update updates the receiver Runtime Defined Entity Type with the values given by the input.
// Only a System administrator can create RDE Types.
func (rdeType *DefinedEntityType) Update(rdeTypeToUpdate types.DefinedEntityType) error {
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

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes,
		endpointParams: []string{rdeType.DefinedEntityType.ID},
		entityLabel:    labelDefinedEntityType,
	}

	resultDefinedEntityType, err := updateInnerEntity(rdeType.client, c, &rdeTypeToUpdate)
	if err != nil {
		return err
	}
	// Only if there was no error in request we overwrite pointer receiver as otherwise it would
	// wipe out existing data
	rdeType.DefinedEntityType = resultDefinedEntityType

	return nil
}

// Delete deletes the receiver Runtime Defined Entity Type.
// Only a System administrator can delete RDE Types.
func (rdeType *DefinedEntityType) Delete() error {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes,
		endpointParams: []string{rdeType.DefinedEntityType.ID},
		entityLabel:    labelDefinedEntityType,
	}

	if err := deleteEntityById(rdeType.client, c); err != nil {
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
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviors,
		endpointParams: []string{rdeType.DefinedEntityType.ID, id},
		entityLabel:    labelRdeBehavior,
	}
	return getInnerEntity[types.Behavior](rdeType.client, c)
}

// GetBehaviorByName retrieves a unique Behavior that belongs to the receiver RDE Type and is named after
// the input.
func (rdeType *DefinedEntityType) GetBehaviorByName(name string) (*types.Behavior, error) {
	behaviors, err := rdeType.GetAllBehaviors(nil)
	if err != nil {
		return nil, fmt.Errorf("could not get the Behaviors of the Defined Entity Type with ID '%s': %s", rdeType.DefinedEntityType.ID, err)
	}
	label := fmt.Sprintf("Defined Entity Behavior with name '%s' in Defined Entity Type with ID '%s': %s", name, rdeType.DefinedEntityType.ID, ErrorEntityNotFound)
	return localFilterOneOrError(label, behaviors, "Name", name)
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

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviors,
		endpointParams: []string{rdeType.DefinedEntityType.ID, behavior.ID},
		entityLabel:    labelRdeBehaviorOverride,
	}
	return updateInnerEntity(rdeType.client, c, &behavior)
}

// DeleteBehaviorOverride removes a Behavior specified by its ID from the receiver Defined Entity Type.
// The ID can be the Interface Behavior ID or the Type Behavior ID (the overridden one).
func (rdeType *DefinedEntityType) DeleteBehaviorOverride(behaviorId string) error {
	if rdeType.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Entity Type is empty")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviors,
		endpointParams: []string{rdeType.DefinedEntityType.ID, behaviorId},
		entityLabel:    labelRdeBehaviorOverride,
	}
	return deleteEntityById(rdeType.client, c)
}

// SetBehaviorAccessControls sets the given slice of BehaviorAccess to the receiver Defined Entity Type.
// If the input is nil, it removes all access controls from the receiver Defined Entity Type.
func (det *DefinedEntityType) SetBehaviorAccessControls(acls []*types.BehaviorAccess) error {
	if det.DefinedEntityType.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Entity Type is empty")
	}

	sanitizedAcls := acls
	if acls == nil {
		sanitizedAcls = []*types.BehaviorAccess{}
	}

	// Wrap it in OpenAPI pages, this endpoint requires it
	rawMessage, err := json.Marshal(sanitizedAcls)
	if err != nil {
		return fmt.Errorf("error setting Access controls in payload: %s", err)
	}
	payload := types.OpenApiPages{
		Values: rawMessage,
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviorAccessControls,
		endpointParams: []string{det.DefinedEntityType.ID},
		entityLabel:    labelRdeBehaviorAccessControl,
	}
	_, err = updateInnerEntity(det.client, c, &payload)
	if err != nil {
		return err
	}

	return nil
}

// GetAllBehaviorsAccessControls gets all the Behaviors Access Controls from the receiver DefinedEntityType.
// Query parameters can be supplied to modify pagination.
func (det *DefinedEntityType) GetAllBehaviorsAccessControls(queryParameters url.Values) ([]*types.BehaviorAccess, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeTypeBehaviorAccessControls,
		queryParameters: queryParameters,
		endpointParams:  []string{det.DefinedEntityType.ID},
		entityLabel:     labelRdeBehaviorAccessControl,
	}
	return getAllInnerEntities[types.BehaviorAccess](det.client, c)
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
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntitiesTypes,
		entityLabel:     labelDefinedEntityType,
		queryParameters: queryParameters,
		endpointParams:  []string{vendor, "/", nss, "/", version},
	}

	outerType := DefinedEntity{client: client}
	return getAllOuterEntities[DefinedEntity, types.DefinedEntity](client, outerType, c)
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
	c := crudConfig{
		entityLabel:    labelDefinedEntityType,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities,
		endpointParams: []string{id},
	}

	outerType := DefinedEntity{client: client}
	result, headers, err := getOuterEntityWithHeaders(client, outerType, c)
	if err != nil {
		return nil, err
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
	task, err := createRde(rdeType.client, entity, tenantContext)
	if err != nil {
		return nil, err
	}
	return getRdeFromTask(rdeType.client, task)
}

// CreateRde creates an entity of the type of the given vendor, nss and version.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func (vcdClient *VCDClient) CreateRde(vendor, nss, version string, entity types.DefinedEntity, tenantContext *TenantContext) (*DefinedEntity, error) {
	return createRdeAndGetFromTask(&vcdClient.Client, vendor, nss, version, entity, tenantContext)
}

// createRdeAndGetFromTask creates an entity of the type of the given vendor, nss and version.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func createRdeAndGetFromTask(client *Client, vendor, nss, version string, entity types.DefinedEntity, tenantContext *TenantContext) (*DefinedEntity, error) {
	entity.EntityType = fmt.Sprintf("urn:vcloud:type:%s:%s:%s", vendor, nss, version)
	task, err := createRde(client, entity, tenantContext)
	if err != nil {
		return nil, err
	}
	return getRdeFromTask(client, task)
}

// CreateRde creates an entity of the type of the receiver Runtime Defined Entity (RDE) type.
// The input doesn't need to specify the type ID, as it gets it from the receiver RDE type. If it is specified anyway,
// it must match the type ID of the receiver RDE type.
// NOTE: After RDE creation, some actor should Resolve it, otherwise the RDE state will be "PRE_CREATED"
// and the generated VCD task will remain at 1% until resolved.
func createRde(client *Client, entity types.DefinedEntity, tenantContext *TenantContext) (*Task, error) {
	if entity.EntityType == "" {
		return nil, fmt.Errorf("ID of the Runtime Defined Entity type is empty")
	}

	if entity.Entity == nil || len(entity.Entity) == 0 {
		return nil, fmt.Errorf("the entity JSON is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntityTypes
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, entity.EntityType)
	if err != nil {
		return nil, err
	}

	task, err := client.OpenApiPostItemAsyncWithHeaders(apiVersion, urlRef, nil, entity, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}
	// The refresh is needed as the task only has the HREF at the moment
	err = task.Refresh()
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// getRdeFromTask gets the Runtime Defined Entity from a given Task. This method is useful after RDE creation, as
// the API just returns a Task with the RDE details inside.
func getRdeFromTask(client *Client, task *Task) (*DefinedEntity, error) {
	if task.Task == nil {
		return nil, fmt.Errorf("could not retrieve the RDE from task, as it is nil")
	}
	rdeId := ""
	if task.Task.Owner == nil {
		// Try to retrieve the ID from the "Operation" field
		beginning := strings.LastIndex(task.Task.Operation, "(")
		end := strings.LastIndex(task.Task.Operation, ")")
		if beginning < 0 || end < 0 || beginning >= end {
			return nil, fmt.Errorf("could not retrieve the RDE from the task with ID '%s'", task.Task.ID)
		}
		rdeId = task.Task.Operation[beginning+1 : end]
	} else {
		rdeId = task.Task.Owner.ID
	}

	return getRdeById(client, rdeId)
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

	c := crudConfig{
		entityLabel:      labelDefinedEntity,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities,
		endpointParams:   []string{rde.DefinedEntity.ID},
		additionalHeader: map[string]string{"If-Match": rde.Etag},
	}

	resultDefinedEntity, headers, err := updateInnerEntityWithHeaders(rde.client, c, &rdeToUpdate)
	if err != nil {
		return err
	}
	// Only if there was no error in request we overwrite pointer receiver as otherwise it would
	// wipe out existing data
	rde.DefinedEntity = resultDefinedEntity
	rde.Etag = headers.Get("Etag")

	return nil
}

// Delete deletes the receiver Runtime Defined Entity.
func (rde *DefinedEntity) Delete() error {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeEntities,
		endpointParams: []string{rde.DefinedEntity.ID},
		entityLabel:    labelDefinedEntity,
	}

	if err := deleteEntityById(rde.client, c); err != nil {
		return amendRdeApiError(rde.client, err)
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
