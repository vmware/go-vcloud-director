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

// GetGlobalRoleByName retrieves a global role by given name
func (client *Client) GetGlobalRoleByName(name string) (*GlobalRole, error) {
	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	globalRoles, err := client.GetAllGlobalRoles(queryParams)
	if err != nil {
		return nil, err
	}
	if len(globalRoles) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(globalRoles) > 1 {
		return nil, fmt.Errorf("more than one global role found with name '%s'", name)
	}
	return globalRoles[0], nil
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
func (client *Client) CreateGlobalRole(newGlobalRole *types.GlobalRole) (*GlobalRole, error) {
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

	if newGlobalRole.BundleKey == "" {
		newGlobalRole.BundleKey = types.VcloudUndefinedKey
	}
	if newGlobalRole.PublishAll == nil {
		newGlobalRole.PublishAll = addrOf(false)
	}
	returnGlobalRole := &GlobalRole{
		GlobalRole: &types.GlobalRole{},
		client:     client,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, newGlobalRole, returnGlobalRole.GlobalRole, nil)
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

// getContainerTenants retrieves all tenants associated with a given rights container (Global Role, Rights Bundle).
// Query parameters can be supplied to perform additional filtering
func getContainerTenants(client *Client, rightsContainerId, endpoint string, queryParameters url.Values) ([]types.OpenApiReference, error) {
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint + rightsContainerId + "/tenants")
	if err != nil {
		return nil, err
	}

	typeResponses := types.OpenApiItems{
		Values: []types.OpenApiReference{},
	}

	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses.Values, nil)
	if err != nil {
		return nil, err
	}

	return typeResponses.Values, nil
}

// publishContainerToTenants is a generic function that publishes or unpublishes a rights collection (Global Role, or Rights bundle) to tenants
// containerType is an informative string (one of "GlobalRole", "RightsBundle")
// name and id are the name and ID of the collection
// endpoint is the API endpoint used as a basis for the POST operation
// tenants is a collection of tenants (ID+name) to be added
// publishType can be one of "add", "remove", "replace"
func publishContainerToTenants(client *Client, containerType, name, id, endpoint string, tenants []types.OpenApiReference, publishType string) error {
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("cannot update %s without id", containerType)
	}
	if name == "" {
		return fmt.Errorf("empty name given for %s %s", containerType, id)
	}

	var operation string

	var action func(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}, additionalHeader map[string]string) error

	switch publishType {
	case "add":
		operation = "/tenants/publish"
		action = client.OpenApiPostItem
	case "replace":
		operation = "/tenants"
		action = client.OpenApiPutItem
	case "remove":
		operation = "/tenants/unpublish"
		action = client.OpenApiPostItem
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id, operation)
	if err != nil {
		return err
	}

	var input types.OpenApiItems

	for _, tenant := range tenants {
		input.Values = append(input.Values, types.OpenApiReference{
			Name: tenant.Name,
			ID:   tenant.ID,
		})
	}
	var pages types.OpenApiPages

	err = action(minimumApiVersion, urlRef, nil, &input, &pages, nil)

	if err != nil {
		return fmt.Errorf("error publishing %s %s to tenants: %s", containerType, name, err)
	}

	return nil
}

// publishContainerToAllTenants is a generic function that publishes or unpublishes a rights collection ( Global Role, or Rights bundle) to all tenants
// containerType is an informative string (one of "GlobalRole", "RightsBundle")
// name and id are the name and ID of the collection
// endpoint is the API endpoint used as a basis for the POST operation
// If "publish" is false, it will revert the operation
func publishContainerToAllTenants(client *Client, containerType, name, id, endpoint string, publish bool) error {
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("cannot update %s without id", containerType)
	}
	if name == "" {
		return fmt.Errorf("empty name given for %s %s", containerType, id)
	}

	operation := "/tenants/publishAll"
	if !publish {
		operation = "/tenants/unpublishAll"
	}
	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id, operation)
	if err != nil {
		return err
	}

	var pages types.OpenApiPages

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, &pages, &pages, nil)

	if err != nil {
		return fmt.Errorf("error publishing %s %s to tenants: %s", containerType, name, err)
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
func (globalRole *GlobalRole) RemoveRights(removeRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return removeRightsFromRole(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, removeRights, nil)
}

// RemoveAllRights removes all rights from a global role
func (globalRole *GlobalRole) RemoveAllRights() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return removeAllRightsFromRole(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, nil)
}

// GetRights retrieves all rights belonging to a given Global Role. Query parameters can be supplied to perform additional
// filtering
func (globalRole *GlobalRole) GetRights(queryParameters url.Values) ([]*types.Right, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return getRights(globalRole.client, globalRole.GlobalRole.Id, endpoint, queryParameters, nil)
}

// GetTenants retrieves all tenants associated to a given Global Role. Query parameters can be supplied to perform additional
// filtering
func (globalRole *GlobalRole) GetTenants(queryParameters url.Values) ([]types.OpenApiReference, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return getContainerTenants(globalRole.client, globalRole.GlobalRole.Id, endpoint, queryParameters)
}

// PublishTenants publishes a global role to one or more tenants, adding to tenants that may already been there
func (globalRole *GlobalRole) PublishTenants(tenants []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return publishContainerToTenants(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, tenants, "add")
}

// ReplacePublishedTenants publishes a global role to one or more tenants, removing the tenants already present
func (globalRole *GlobalRole) ReplacePublishedTenants(tenants []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return publishContainerToTenants(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, tenants, "replace")
}

// UnpublishTenants remove tenats from a global role
func (globalRole *GlobalRole) UnpublishTenants(tenants []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return publishContainerToTenants(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, tenants, "remove")
}

// PublishAllTenants publishes a global role to all tenants
func (globalRole *GlobalRole) PublishAllTenants() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return publishContainerToAllTenants(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, true)
}

// UnpublishAllTenants remove publication status of a global role from all tenants
func (globalRole *GlobalRole) UnpublishAllTenants() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGlobalRoles
	return publishContainerToAllTenants(globalRole.client, "GlobalRole", globalRole.GlobalRole.Name, globalRole.GlobalRole.Id, endpoint, false)
}
