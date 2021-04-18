package govcd

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Role uses OpenAPI endpoint to operate user roles
type Role struct {
	Role   *types.Role
	client *Client
}

// GetOpenApiRoleById retrieves role by given ID
func (adminOrg *AdminOrg) GetOpenApiRoleById(ctx context.Context, id string) (*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty role id")
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	role := &Role{
		Role:   &types.Role{},
		client: adminOrg.client,
	}

	err = adminOrg.client.OpenApiGetItem(ctx, minimumApiVersion, urlRef, nil, role.Role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// GetAllOpenApiRoles retrieves all roles using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
func (adminOrg *AdminOrg) GetAllOpenApiRoles(ctx context.Context, queryParameters url.Values) ([]*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.Role{{}}
	err = adminOrg.client.OpenApiGetAllItems(ctx, minimumApiVersion, urlRef, queryParameters, &typeResponses)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into Role types with client
	returnRoles := make([]*Role, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnRoles[sliceIndex] = &Role{
			Role:   typeResponses[sliceIndex],
			client: adminOrg.client,
		}
	}

	return returnRoles, nil
}

// CreateRole creates a new role using OpenAPI endpoint
func (adminOrg *AdminOrg) CreateRole(ctx context.Context, newRole *types.Role) (*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnRole := &Role{
		Role:   &types.Role{},
		client: adminOrg.client,
	}

	err = adminOrg.client.OpenApiPostItem(ctx, minimumApiVersion, urlRef, nil, newRole, returnRole.Role)
	if err != nil {
		return nil, fmt.Errorf("error creating role: %s", err)
	}

	return returnRole, nil
}

// Update updates existing OpenAPI role
func (role *Role) Update(ctx context.Context) (*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	if role.Role.ID == "" {
		return nil, fmt.Errorf("cannot update role without id")
	}

	urlRef, err := role.client.OpenApiBuildEndpoint(endpoint, role.Role.ID)
	if err != nil {
		return nil, err
	}

	returnRole := &Role{
		Role:   &types.Role{},
		client: role.client,
	}

	err = role.client.OpenApiPutItem(ctx, minimumApiVersion, urlRef, nil, role.Role, returnRole.Role)
	if err != nil {
		return nil, fmt.Errorf("error updating role: %s", err)
	}

	return returnRole, nil
}

// Delete deletes OpenAPI role
func (role *Role) Delete(ctx context.Context) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(ctx, endpoint)
	if err != nil {
		return err
	}

	if role.Role.ID == "" {
		return fmt.Errorf("cannot delete role without id")
	}

	urlRef, err := role.client.OpenApiBuildEndpoint(endpoint, role.Role.ID)
	if err != nil {
		return err
	}

	err = role.client.OpenApiDeleteItem(ctx, minimumApiVersion, urlRef, nil)

	if err != nil {
		return fmt.Errorf("error deleting role: %s", err)
	}

	return nil
}
