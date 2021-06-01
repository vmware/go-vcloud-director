/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type GlobalRole struct {
	GlobalRole *types.GlobalRole
	client     *Client
	rights     []*types.Right
}

// GetAllGlobalRoles retrieves all global roles. Query parameters can be supplied to perform additional filtering
// Only System administrator can handle global roles
func (client *Client) GetAllGlobalRoles(queryParameters url.Values) ([]*GlobalRole, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("only system administrator can handle global roles")
	}
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.GlobalRole{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into GlobalRole types with client
	returnGlobalRoles := make([]*GlobalRole, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnGlobalRoles[sliceIndex] = &GlobalRole{
			GlobalRole: typeResponses[sliceIndex],
			client:     client,
		}
	}

	return returnGlobalRoles, nil
}

// GetGlobalRoleById retrieves global role by given ID
func (client *Client) GetGlobalRoleById(id string) (*GlobalRole, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("only system administrator can handle global roles")
	}
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty GlobalRole id")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	globalRole := &GlobalRole{
		GlobalRole: &types.GlobalRole{},
		client:     client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, globalRole.GlobalRole, nil)
	if err != nil {
		return nil, err
	}

	return globalRole, nil
}

// CreateGlobalRole creates a new global role as a system administrator
func (client *Client) CreateGlobalRole(newRole *types.GlobalRole) (*GlobalRole, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("only system administrator can handle global roles")
	}
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnGlobalRole := &GlobalRole{
		GlobalRole: &types.GlobalRole{},
		client:     client,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, newRole, returnGlobalRole.GlobalRole, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating global role: %s", err)
	}

	return returnGlobalRole, nil
}

// Update updates existing global role
func (globalRole *GlobalRole) Update() (*GlobalRole, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	minimumApiVersion, err := globalRole.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if globalRole.GlobalRole.Id == "" {
		return nil, fmt.Errorf("cannot update role without id")
	}

	urlRef, err := globalRole.client.OpenApiBuildEndpoint(endpoint, globalRole.GlobalRole.Id)
	if err != nil {
		return nil, err
	}

	returnGlobalRole := &GlobalRole{
		GlobalRole: &types.GlobalRole{},
		client:     globalRole.client,
	}

	err = globalRole.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, globalRole.GlobalRole, returnGlobalRole.GlobalRole, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating global role: %s", err)
	}

	return returnGlobalRole, nil
}

// Delete deletes global role
func (globalRole *GlobalRole) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	minimumApiVersion, err := globalRole.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if globalRole.GlobalRole.Id == "" {
		return fmt.Errorf("cannot delete global role without id")
	}

	urlRef, err := globalRole.client.OpenApiBuildEndpoint(endpoint, globalRole.GlobalRole.Id)
	if err != nil {
		return err
	}

	err = globalRole.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting global role: %s", err)
	}

	return nil
}

// AddRights adds a collection of rights to a global role
func (globalRole *GlobalRole) AddRights(newRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return addRightsToRole(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, newRights, nil)
}

// UpdateRights replaces existing rights with the given collection of rights
func (globalRole *GlobalRole) UpdateRights(newRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return updateRightsInRole(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, newRights, nil)
}

// RemoveRights removes specific rights from a global role
func (globalRole *GlobalRole) RemoveRights(newRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return removeRightsFromRole(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, newRights, nil)
}

// RemoveAllRights removes specific rights from a global role
func (globalRole *GlobalRole) RemoveAllRights() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return removeAllRightsFromRole(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, nil)
}

// GetGlobalRoleRights retrieves all rights belonging to a given Global Role. Query parameters can be supplied to perform additional
// filtering
func (globalRole *GlobalRole) GetGlobalRoleRights(queryParameters url.Values) ([]*types.Right, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return getRoleRights(globalRole.client, globalRole.GlobalRole.Id, endpoint, queryParameters, nil)
}
