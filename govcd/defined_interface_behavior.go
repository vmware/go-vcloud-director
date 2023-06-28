/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

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

// GetAllBehaviors retrieves all the Behaviors of the receiver Defined Interface. Query parameters can be supplied to modify pagination.
func (di *DefinedInterface) GetAllBehaviors(queryParameters url.Values) ([]*types.Behavior, error) {
	if di.DefinedInterface.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Interface is empty")
	}
	return getAllBehaviors(di.client, di.DefinedInterface.ID, types.OpenApiEndpointRdeInterfaceBehaviors, queryParameters)
}

// GetAllBehaviors retrieves all the Behaviors of the receiver RDE Type. Query parameters can be supplied to modify pagination.
func (det *DefinedEntityType) GetAllBehaviors(queryParameters url.Values) ([]*types.Behavior, error) {
	if det.DefinedEntityType.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Entity Type is empty")
	}
	return getAllBehaviors(det.client, det.DefinedEntityType.ID, types.OpenApiEndpointRdeTypeBehaviors, queryParameters)
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
		return nil, fmt.Errorf("could not get the Behavior of the Defined Interface with ID '%s': %s", di.DefinedInterface.ID, err)
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

// AddBehaviorAccessControl adds a new Behavior to the receiver DefinedInterface.
// Only allowed if the Interface is not in use.
func (det *DefinedEntityType) AddBehaviorAccessControl(ac types.BehaviorAccess) error {
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

	err = det.client.OpenApiPostItem(apiVersion, urlRef, nil, ac, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetAllBehaviorsAccessControls gets all the Behaviors Access Controls from the receiver DefinedEntityType. Query parameters can be supplied to modify pagination.
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
