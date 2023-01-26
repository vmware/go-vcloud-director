/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"strings"
)

// DefinedInterface is a type for handling Defined Interfaces in VCD that allow to define new RDEs (DefinedEntityType).
type DefinedInterface struct {
	DefinedInterface *types.DefinedInterface
	client           *Client
}

// CreateDefinedInterface creates a Defined Interface.
// Only System administrator can create Defined Interfaces.
func (vcdClient *VCDClient) CreateDefinedInterface(definedInterface *types.DefinedInterface) (*DefinedInterface, error) {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointInterfaces
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
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointInterfaces
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
		return nil, amendDefinedInterfaceError(&client, err)
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

// GetDefinedInterface retrieves a single Defined Interface defined by its unique combination of vendor, namespace and version.
func (vcdClient *VCDClient) GetDefinedInterface(vendor, namespace, version string) (*DefinedInterface, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("vendor==%s;nss==%s;version==%s", vendor, namespace, version))
	interfaces, err := vcdClient.GetAllDefinedInterfaces(queryParameters)
	if err != nil {
		return nil, err
	}

	if len(interfaces) == 0 {
		return nil, fmt.Errorf("%s could not find the Defined Interface with vendor %s, namespace %s and version %s", ErrorEntityNotFound, vendor, namespace, version)
	}

	if len(interfaces) > 1 {
		return nil, fmt.Errorf("found more than 1 Defined Interface with vendor %s, namespace %s and version %s", vendor, namespace, version)
	}

	return interfaces[0], nil
}

// GetDefinedInterfaceById gets a Defined Interface identified by its unique URN.
func (vcdClient *VCDClient) GetDefinedInterfaceById(id string) (*DefinedInterface, error) {
	client := vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointInterfaces
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
		return nil, amendDefinedInterfaceError(&client, err)
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

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointInterfaces
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
		return amendDefinedInterfaceError(client, err)
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

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointInterfaces
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
		return amendDefinedInterfaceError(client, err)
	}

	di.DefinedInterface = &types.DefinedInterface{}
	return nil
}

// amendDefinedInterfaceError fixes a wrong type of error returned by VCD API <= v36.0 on GET operations
// when the defined interface does not exist.
func amendDefinedInterfaceError(client *Client, err error) error {
	if client.APIClientVersionIs("<= 36.0") && err != nil && strings.Contains(err.Error(), "does not exist") {
		return fmt.Errorf("%s: %s", ErrorEntityNotFound.Error(), err)
	}
	return err
}
