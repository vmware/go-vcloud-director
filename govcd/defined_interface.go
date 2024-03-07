/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const (
	labelDefinedInterface         = "Defined Interface"
	labelDefinedInterfaceBehavior = "Defined Interface Behavior"
)

// DefinedInterface is a type for handling Defined Interfaces, from the Runtime Defined Entities framework, in VCD.
// This is often referred as Runtime Defined Entity Interface or RDE Interface in documentation.
type DefinedInterface struct {
	DefinedInterface *types.DefinedInterface
	client           *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (d DefinedInterface) wrap(inner *types.DefinedInterface) *DefinedInterface {
	d.DefinedInterface = inner
	return &d
}

// CreateDefinedInterface creates a Defined Interface.
// Only System administrator can create Defined Interfaces.
func (vcdClient *VCDClient) CreateDefinedInterface(definedInterface *types.DefinedInterface) (*DefinedInterface, error) {
	c := crudConfig{
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces,
		entityLabel: labelDefinedInterface,
	}
	outerType := DefinedInterface{client: &vcdClient.Client}
	return createOuterEntity(&vcdClient.Client, outerType, c, definedInterface)
}

// GetAllDefinedInterfaces retrieves all Defined Interfaces. Query parameters can be supplied to perform additional filtering.
func (vcdClient *VCDClient) GetAllDefinedInterfaces(queryParameters url.Values) ([]*DefinedInterface, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces,
		entityLabel:     labelDefinedInterface,
		queryParameters: queryParameters,
	}

	outerType := DefinedInterface{client: &vcdClient.Client}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetDefinedInterface retrieves a single Defined Interface defined by its unique combination of vendor, nss and version.
func (vcdClient *VCDClient) GetDefinedInterface(vendor, nss, version string) (*DefinedInterface, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("vendor==%s;nss==%s;version==%s", vendor, nss, version))
	interfaces, err := vcdClient.GetAllDefinedInterfaces(queryParameters)
	if err != nil {
		return nil, err
	}

	if len(interfaces) == 0 {
		return nil, fmt.Errorf("%s could not find the Defined Interface with vendor %s, nss %s and version %s", ErrorEntityNotFound, vendor, nss, version)
	}

	if len(interfaces) > 1 {
		return nil, fmt.Errorf("found more than 1 Defined Interface with vendor %s, nss %s and version %s", vendor, nss, version)
	}

	return interfaces[0], nil
}

// GetDefinedInterfaceById gets a Defined Interface identified by its unique URN.
func (vcdClient *VCDClient) GetDefinedInterfaceById(id string) (*DefinedInterface, error) {
	c := crudConfig{
		entityLabel:    labelDefinedInterface,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces,
		endpointParams: []string{id},
	}

	outerType := DefinedInterface{client: &vcdClient.Client}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update updates the receiver Defined Interface with the values given by the input.
// Only System administrator can update Defined Interfaces.
func (di *DefinedInterface) Update(definedInterface types.DefinedInterface) error {
	if di.DefinedInterface.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	if definedInterface.ID != "" && definedInterface.ID != di.DefinedInterface.ID {
		return fmt.Errorf("ID of the receiver Defined Interface and the input ID don't match")
	}

	// We override these as they need to be always sent on updates
	definedInterface.Version = di.DefinedInterface.Version
	definedInterface.Nss = di.DefinedInterface.Nss
	definedInterface.Vendor = di.DefinedInterface.Vendor

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces,
		endpointParams: []string{di.DefinedInterface.ID},
		entityLabel:    labelDefinedInterface,
	}
	resultDefinedInterface, err := updateInnerEntity(di.client, c, &definedInterface)
	if err != nil {
		return err
	}
	// Only if there was no error in request we overwrite pointer receiver as otherwise it would
	// wipe out existing data
	di.DefinedInterface = resultDefinedInterface
	return err
}

// Delete deletes the receiver Defined Interface.
// Only System administrator can delete Defined Interfaces.
func (di *DefinedInterface) Delete() error {
	if di.DefinedInterface.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces,
		endpointParams: []string{di.DefinedInterface.ID},
		entityLabel:    labelDefinedInterface,
	}

	err := deleteEntityById(di.client, c)
	if err != nil {
		return err
	}

	di.DefinedInterface = &types.DefinedInterface{}
	return nil
}

// AddBehavior adds a new Behavior to the receiver DefinedInterface.
// Only allowed if the Interface is not in use.
func (di *DefinedInterface) AddBehavior(behavior types.Behavior) (*types.Behavior, error) {
	if di.DefinedInterface.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors,
		endpointParams: []string{di.DefinedInterface.ID},
		entityLabel:    labelDefinedInterfaceBehavior,
	}
	return createInnerEntity(di.client, c, &behavior)
}

// GetAllBehaviors retrieves all the Behaviors of the receiver Defined Interface.
func (di *DefinedInterface) GetAllBehaviors(queryParameters url.Values) ([]*types.Behavior, error) {
	if di.DefinedInterface.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Interface is empty")
	}
	return getAllBehaviors(di.client, di.DefinedInterface.ID, types.OpenApiEndpointRdeInterfaceBehaviors, queryParameters)
}

// getAllBehaviors gets all the Behaviors from the object referenced by the input Object ID with the given OpenAPI endpoint.
func getAllBehaviors(client *Client, objectId, openApiEndpoint string, queryParameters url.Values) ([]*types.Behavior, error) {
	c := crudConfig{
		endpoint:        types.OpenApiPathVersion1_0_0 + openApiEndpoint,
		entityLabel:     labelDefinedInterfaceBehavior,
		endpointParams:  []string{objectId},
		queryParameters: queryParameters,
	}
	return getAllInnerEntities[types.Behavior](client, c)
}

// GetBehaviorById retrieves a unique Behavior that belongs to the receiver Defined Interface and is determined by the
// input ID.
func (di *DefinedInterface) GetBehaviorById(id string) (*types.Behavior, error) {
	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors,
		endpointParams: []string{di.DefinedInterface.ID, id},
		entityLabel:    labelDefinedInterfaceBehavior,
	}
	return getInnerEntity[types.Behavior](di.client, c)
}

// GetBehaviorByName retrieves a unique Behavior that belongs to the receiver Defined Interface and is named after
// the input.
func (di *DefinedInterface) GetBehaviorByName(name string) (*types.Behavior, error) {
	behaviors, err := di.GetAllBehaviors(nil)
	if err != nil {
		return nil, fmt.Errorf("could not get the Behaviors of the Defined Interface with ID '%s': %s", di.DefinedInterface.ID, err)
	}
	label := fmt.Sprintf("Defined Interface Behavior with name '%s' in Defined Interface with ID '%s': %s", name, di.DefinedInterface.ID, ErrorEntityNotFound)
	return localFilterOneOrError(label, behaviors, "Name", name)
}

// UpdateBehavior updates a Behavior specified by the input.
func (di *DefinedInterface) UpdateBehavior(behavior types.Behavior) (*types.Behavior, error) {
	if di.DefinedInterface.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Interface is empty")
	}
	if behavior.ID == "" {
		return nil, fmt.Errorf("ID of the Behavior to update is empty")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors,
		endpointParams: []string{di.DefinedInterface.ID, behavior.ID},
		entityLabel:    labelDefinedInterfaceBehavior,
	}
	return updateInnerEntity(di.client, c, &behavior)
}

// DeleteBehavior removes a Behavior specified by its ID from the receiver Defined Interface.
func (di *DefinedInterface) DeleteBehavior(behaviorId string) error {
	if di.DefinedInterface.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	c := crudConfig{
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors,
		endpointParams: []string{di.DefinedInterface.ID, behaviorId},
		entityLabel:    labelDefinedInterfaceBehavior,
	}
	return deleteEntityById(di.client, c)
}

// amendRdeApiError fixes a wrong type of error returned by VCD API <= v36.0 on GET operations
// when the defined interface does not exist.
func amendRdeApiError(client *Client, err error) error {
	if client.APIClientVersionIs("<= 36.0") && err != nil && strings.Contains(err.Error(), "does not exist") {
		return fmt.Errorf("%s: %s", ErrorEntityNotFound.Error(), err)
	}
	return err
}
