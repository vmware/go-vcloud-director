/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Role uses OpenAPI endpoint to operate user roles
type Role struct {
	Role          *types.Role
	client        *Client
	TenantContext *TenantContext
}

// GetRoleById retrieves role by given ID
func (adminOrg *AdminOrg) GetRoleById(id string) (*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
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

	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	role := &Role{
		Role:          &types.Role{},
		client:        adminOrg.client,
		TenantContext: tenantContext,
	}

	err = adminOrg.client.OpenApiGetItem(minimumApiVersion, urlRef, nil, role.Role, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, err
	}

	return role, nil
}

// GetRoleByName retrieves role by given name
func (adminOrg *AdminOrg) GetRoleByName(name string) (*Role, error) {
	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	roles, err := adminOrg.GetAllRoles(queryParams)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(roles) > 1 {
		return nil, fmt.Errorf("more than one role found with name '%s'", name)
	}
	return roles[0], nil
}

// getAllRoles retrieves all roles using OpenAPI endpoint. Query parameters can be supplied to perform additional
// filtering
func getAllRoles(client *Client, queryParameters url.Values, additionalHeader map[string]string) ([]*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.Role{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, additionalHeader)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into Role types with client
	returnRoles := make([]*Role, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnRoles[sliceIndex] = &Role{
			Role:          typeResponses[sliceIndex],
			client:        client,
			TenantContext: getTenantContextFromHeader(additionalHeader),
		}
	}

	return returnRoles, nil
}

// GetAllRoles retrieves all roles as tenant user. Query parameters can be supplied to perform additional
// filtering
func (adminOrg *AdminOrg) GetAllRoles(queryParameters url.Values) ([]*Role, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return getAllRoles(adminOrg.client, queryParameters, getTenantContextHeader(tenantContext))
}

// GetAllRoles retrieves all roles as System administrator. Query parameters can be supplied to perform additional
// filtering
func (client *Client) GetAllRoles(queryParameters url.Values) ([]*Role, error) {
	return getAllRoles(client, queryParameters, nil)
}

// CreateRole creates a new role as a tenant administrator
func (adminOrg *AdminOrg) CreateRole(newRole *types.Role) (*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := adminOrg.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if newRole.BundleKey == "" {
		newRole.BundleKey = types.VcloudUndefinedKey
	}

	urlRef, err := adminOrg.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	returnRole := &Role{
		Role:          &types.Role{},
		client:        adminOrg.client,
		TenantContext: tenantContext,
	}

	err = adminOrg.client.OpenApiPostItem(minimumApiVersion, urlRef, nil, newRole, returnRole.Role, getTenantContextHeader(tenantContext))
	if err != nil {
		return nil, fmt.Errorf("error creating role: %s", err)
	}

	return returnRole, nil
}

// Update updates existing OpenAPI role
func (role *Role) Update() (*Role, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(endpoint)
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
		Role:          &types.Role{},
		client:        role.client,
		TenantContext: role.TenantContext,
	}

	err = role.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, role.Role, returnRole.Role, getTenantContextHeader(role.TenantContext))
	if err != nil {
		return nil, fmt.Errorf("error updating role: %s", err)
	}

	return returnRole, nil
}

// Delete deletes OpenAPI role
func (role *Role) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	minimumApiVersion, err := role.client.checkOpenApiEndpointCompatibility(endpoint)
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

	err = role.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, getTenantContextHeader(role.TenantContext))

	if err != nil {
		return fmt.Errorf("error deleting role: %s", err)
	}

	return nil
}

// AddRights adds a collection of rights to a role
func (role *Role) AddRights(newRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	return addRightsToRole(role.client, "Role", role.Role.Name, role.Role.ID, endpoint, newRights, getTenantContextHeader(role.TenantContext))
}

// UpdateRights replaces existing rights with the given collection of rights
func (role *Role) UpdateRights(newRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	return updateRightsInRole(role.client, "Role", role.Role.Name, role.Role.ID, endpoint, newRights, getTenantContextHeader(role.TenantContext))
}

// RemoveRights removes specific rights from a role
func (role *Role) RemoveRights(removeRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	return removeRightsFromRole(role.client, "Role", role.Role.Name, role.Role.ID, endpoint, removeRights, getTenantContextHeader(role.TenantContext))
}

// RemoveAllRights removes all rights from a role
func (role *Role) RemoveAllRights() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	return removeAllRightsFromRole(role.client, "Role", role.Role.Name, role.Role.ID, endpoint, getTenantContextHeader(role.TenantContext))
}

// addRightsToRole is a generic function that can add rights to a rights collection (Role, Global Role, or Rights bundle)
// roleType is an informative string (one of "Role", "GlobalRole", or "RightsBundle")
// name and id are the name and ID of the collection
// endpoint is the API endpoint used as a basis for the POST operation
// newRights is a collection of rights (ID+name) to be added
// Note: the API call ignores duplicate rights. If the rights to be added already exist, the call succeeds
// but no changes are recorded
func addRightsToRole(client *Client, roleType, name, id, endpoint string, newRights []types.OpenApiReference, additionalHeader map[string]string) error {
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("cannot update %s without id", roleType)
	}
	if name == "" {
		return fmt.Errorf("empty name given for %s %s", roleType, id)
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id, "/rights")
	if err != nil {
		return err
	}

	var input types.OpenApiItems

	for _, right := range newRights {
		input.Values = append(input.Values, types.OpenApiReference{
			Name: right.Name,
			ID:   right.ID,
		})
	}
	var pages types.OpenApiPages

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, &input, &pages, additionalHeader)

	if err != nil {
		return fmt.Errorf("error adding rights to %s %s: %s", roleType, name, err)
	}

	return nil
}

// updateRightsInRole is a generic function that can change rights in a Role or Global Role
// roleType is an informative string (either "Role" or "GlobalRole")
// name and id are the name and ID of the role
// endpoint is the API endpoint used as a basis for the PUT operation
// newRights is a collection of rights (ID+name) to be added
func updateRightsInRole(client *Client, roleType, name, id, endpoint string, newRights []types.OpenApiReference, additionalHeader map[string]string) error {
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("cannot update %s without id", roleType)
	}
	if name == "" {
		return fmt.Errorf("empty name given for %s %s", roleType, id)
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id, "/rights")
	if err != nil {
		return err
	}

	var input = types.OpenApiItems{
		Values: []types.OpenApiReference{},
	}

	for _, right := range newRights {
		input.Values = append(input.Values, types.OpenApiReference{
			Name: right.Name,
			ID:   right.ID,
		})
	}
	var pages types.OpenApiPages

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, &input, &pages, additionalHeader)

	if err != nil {
		return fmt.Errorf("error updating rights in %s %s: %s", roleType, name, err)
	}

	return nil
}

// removeRightsFromRole is a generic function that can remove rights from a Role or Global Role
// roleType is an informative string (either "Role" or "GlobalRole")
// name and id are the name and ID of the role
// endpoint is the API endpoint used as a basis for the PUT operation
// removeRights is a collection of rights (ID+name) to be removed
func removeRightsFromRole(client *Client, roleType, name, id, endpoint string, removeRights []types.OpenApiReference, additionalHeader map[string]string) error {
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if id == "" {
		return fmt.Errorf("cannot update %s without id", roleType)
	}
	if name == "" {
		return fmt.Errorf("empty name given for %s %s", roleType, id)
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id, "/rights")
	if err != nil {
		return err
	}

	var input = types.OpenApiItems{
		Values: []types.OpenApiReference{},
	}
	var pages types.OpenApiPages

	currentRights, err := getRights(client, id, endpoint, nil, additionalHeader)
	if err != nil {
		return err
	}

	var foundToRemove = make(map[string]bool)

	// Set the items to be removed as not found by default
	for _, rr := range removeRights {
		foundToRemove[rr.Name] = false
	}

	// Search the current rights for items to delete
	for _, cr := range currentRights {
		for _, rr := range removeRights {
			if cr.ID == rr.ID {
				foundToRemove[cr.Name] = true
			}
		}
	}

	for _, cr := range currentRights {
		_, found := foundToRemove[cr.Name]
		if !found {
			input.Values = append(input.Values, types.OpenApiReference{Name: cr.Name, ID: cr.ID})
		}
	}

	// Check that all the items to be removed were found in the current rights list
	notFoundNames := ""
	for name, found := range foundToRemove {
		if !found {
			if notFoundNames != "" {
				notFoundNames += ", "
			}
			notFoundNames += `"` + name + `"`
		}
	}

	if notFoundNames != "" {
		return fmt.Errorf("rights in %s %s not found for deletion: [%s]", roleType, name, notFoundNames)
	}

	err = client.OpenApiPutItem(minimumApiVersion, urlRef, nil, &input, &pages, additionalHeader)

	if err != nil {
		return fmt.Errorf("error updating rights in %s %s: %s", roleType, name, err)
	}

	return nil
}

// removeAllRightsFromRole removes all rights from the given role
func removeAllRightsFromRole(client *Client, roleType, name, id, endpoint string, additionalHeader map[string]string) error {
	return updateRightsInRole(client, roleType, name, id, endpoint, []types.OpenApiReference{}, additionalHeader)
}

// FindMissingImpliedRights returns a list of the rights that are implied in the rights provided as input
func FindMissingImpliedRights(client *Client, rights []types.OpenApiReference) ([]types.OpenApiReference, error) {
	var (
		impliedRights       []types.OpenApiReference
		uniqueInputRights   = make(map[string]types.OpenApiReference)
		uniqueImpliedRights = make(map[string]types.OpenApiReference)
	)

	// Make a searchable collection of unique rights from the input
	// This operation removes duplicates from the list
	for _, right := range rights {
		uniqueInputRights[right.Name] = right
	}

	// Find the implied rights
	for _, right := range rights {
		fullRight, err := client.GetRightByName(right.Name)
		if err != nil {
			return nil, err
		}
		for _, ir := range fullRight.ImpliedRights {
			_, seenAsInput := uniqueInputRights[ir.Name]
			_, seenAsImplied := uniqueImpliedRights[ir.Name]
			// If the right has already been added either as explicit ro as implied right, we skip it
			if seenAsInput || seenAsImplied {
				continue
			}
			// Add to the unique collection of implied rights
			uniqueImpliedRights[ir.Name] = types.OpenApiReference{
				Name: ir.Name,
				ID:   ir.ID,
			}
		}
	}

	// Create the output list from the implied rights collection
	if len(uniqueImpliedRights) > 0 {
		for _, right := range uniqueImpliedRights {
			impliedRights = append(impliedRights, right)
		}
	}

	return impliedRights, nil
}
