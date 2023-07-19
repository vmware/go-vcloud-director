/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"strings"
)

// DefinedInterface is a type for handling Defined Interfaces, from the Runtime Defined Entities framework, in VCD.
// This is often referred as Runtime Defined Entity Interface or RDE Interface in documentation.
type DefinedInterface struct {
	DefinedInterface *types.DefinedInterface
	client           *Client
}

// CreateDefinedInterface creates a Defined Interface.
// Only System administrator can create Defined Interfaces.
func (vcdClient *VCDClient) CreateDefinedInterface(definedInterface *types.DefinedInterface) (*DefinedInterface, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &DefinedInterface{
		DefinedInterface: &types.DefinedInterface{},
		client:           &vcdClient.Client,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, definedInterface, result.DefinedInterface, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetAllDefinedInterfaces retrieves all Defined Interfaces. Query parameters can be supplied to perform additional filtering.
func (vcdClient *VCDClient) GetAllDefinedInterfaces(queryParameters url.Values) ([]*DefinedInterface, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.DefinedInterface{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, amendRdeApiError(&client, err)
	}

	// Wrap all typeResponses into DefinedEntityType types with client
	returnRDEs := make([]*DefinedInterface, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnRDEs[sliceIndex] = &DefinedInterface{
			DefinedInterface: typeResponses[sliceIndex],
			client:           &vcdClient.Client,
		}
	}

	return returnRDEs, nil
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
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	result := &DefinedInterface{
		DefinedInterface: &types.DefinedInterface{},
		client:           &vcdClient.Client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, result.DefinedInterface, nil)
	if err != nil {
		return nil, amendRdeApiError(&client, err)
	}

	return result, nil
}

// Update updates the receiver Defined Interface with the values given by the input.
// Only System administrator can update Defined Interfaces.
func (di *DefinedInterface) Update(definedInterface types.DefinedInterface) error {
	client := di.client

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

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, di.DefinedInterface.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, definedInterface, di.DefinedInterface, nil)
	if err != nil {
		return amendRdeApiError(client, err)
	}

	return nil
}

// Delete deletes the receiver Defined Interface.
// Only System administrator can delete Defined Interfaces.
func (di *DefinedInterface) Delete() error {
	client := di.client

	if di.DefinedInterface.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaces
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, di.DefinedInterface.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return amendRdeApiError(client, err)
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

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors
	apiVersion, err := di.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := di.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, di.DefinedInterface.ID))
	if err != nil {
		return nil, err
	}

	result := &types.Behavior{}
	err = di.client.OpenApiPostItem(apiVersion, urlRef, nil, behavior, result, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
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
	endpoint := types.OpenApiPathVersion1_0_0 + openApiEndpoint
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, objectId))
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.Behavior{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses, nil
}

// GetBehaviorById retrieves a unique Behavior that belongs to the receiver Defined Interface and is determined by the
// input ID.
func (di *DefinedInterface) GetBehaviorById(id string) (*types.Behavior, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors
	apiVersion, err := di.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := di.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, di.DefinedInterface.ID), id)
	if err != nil {
		return nil, err
	}

	response := types.Behavior{}
	err = di.client.OpenApiGetItem(apiVersion, urlRef, nil, &response, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetBehaviorByName retrieves a unique Behavior that belongs to the receiver Defined Interface and is named after
// the input.
func (di *DefinedInterface) GetBehaviorByName(name string) (*types.Behavior, error) {
	behaviors, err := di.GetAllBehaviors(nil)
	if err != nil {
		return nil, fmt.Errorf("could not get the Behaviors of the Defined Interface with ID '%s': %s", di.DefinedInterface.ID, err)
	}
	for _, b := range behaviors {
		if b.Name == name {
			return b, nil
		}
	}
	return nil, fmt.Errorf("could not find any Behavior with name '%s' in Defined Interface with ID '%s': %s", name, di.DefinedInterface.ID, ErrorEntityNotFound)
}

// UpdateBehavior updates a Behavior specified by the input.
func (di *DefinedInterface) UpdateBehavior(behavior types.Behavior) (*types.Behavior, error) {
	if di.DefinedInterface.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Interface is empty")
	}
	if behavior.ID == "" {
		return nil, fmt.Errorf("ID of the Behavior to update is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors
	apiVersion, err := di.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := di.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, di.DefinedInterface.ID), behavior.ID)
	if err != nil {
		return nil, err
	}
	response := types.Behavior{}
	err = di.client.OpenApiPutItem(apiVersion, urlRef, nil, behavior, &response, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteBehavior removes a Behavior specified by its ID from the receiver Defined Interface.
func (di *DefinedInterface) DeleteBehavior(behaviorId string) error {
	if di.DefinedInterface.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors
	apiVersion, err := di.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := di.client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, di.DefinedInterface.ID), behaviorId)
	if err != nil {
		return err
	}
	err = di.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// amendRdeApiError fixes a wrong type of error returned by VCD API <= v36.0 on GET operations
// when the defined interface does not exist.
func amendRdeApiError(client *Client, err error) error {
	if client.APIClientVersionIs("<= 36.0") && err != nil && strings.Contains(err.Error(), "does not exist") {
		return fmt.Errorf("%s: %s", ErrorEntityNotFound.Error(), err)
	}
	return err
}
