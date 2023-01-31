/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

// DefinedEntityType is a type for handling Runtime Defined Entity (RDE) Type definitions
type DefinedEntityType struct {
	DefinedEntityType *types.DefinedEntityType
	client            *Client
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

// GetRdeType gets a Runtime Defined Entity Type by its unique combination of vendor, namespace and version.
func (vcdClient *VCDClient) GetRdeType(vendor, namespace, version string) (*DefinedEntityType, error) {
	queryParameters := url.Values{}
	queryParameters.Add("filter", fmt.Sprintf("vendor==%s;nss==%s;version==%s", vendor, namespace, version))
	rdeTypes, err := vcdClient.GetAllRdeTypes(queryParameters)
	if err != nil {
		return nil, err
	}

	if len(rdeTypes) == 0 {
		return nil, fmt.Errorf("%s could not find the Runtime Defined Entity Type with vendor %s, namespace %s and version %s", ErrorEntityNotFound, vendor, namespace, version)
	}

	if len(rdeTypes) > 1 {
		return nil, fmt.Errorf("found more than 1 Runtime Defined Entity Type with vendor %s, namespace %s and version %s", vendor, namespace, version)
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
