/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// DefinedInterfaceBehavior is a type for handling Defined Interface Behaviors, from the Runtime Defined Entities framework, in VCD.
type DefinedInterfaceBehavior struct {
	Behavior *types.Behavior
	client   *Client
}

// AddBehavior adds a new Behavior to the receiver DefinedInterface.
// Only allowed if the Interface is not in use.
func (di *DefinedInterface) AddBehavior(behavior types.Behavior) (*DefinedInterfaceBehavior, error) {
	client := di.client

	if di.DefinedInterface.ID == "" {
		return nil, fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, di.DefinedInterface.ID))
	if err != nil {
		return nil, err
	}

	result := &DefinedInterfaceBehavior{
		Behavior: &types.Behavior{},
		client:   client,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, behavior, result.Behavior, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteBehavior removes a Behavior specified by its ID from the receiver Defined Interface.
func (di *DefinedInterface) DeleteBehavior(behaviorId string) error {
	client := di.client

	if di.DefinedInterface.ID == "" {
		return fmt.Errorf("ID of the receiver Defined Interface is empty")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRdeInterfaceBehaviors
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, di.DefinedInterface.ID), behaviorId)
	if err != nil {
		return err
	}
	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
