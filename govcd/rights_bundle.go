/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type RightsBundle struct {
	RightsBundle *types.RightsBundle
	client       *Client
}

// CreateRightsBundle creates a new rights bundle as a system administrator
func (client *Client) CreateRightsBundle(newRightsBundle *types.RightsBundle) (*RightsBundle, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("only system administrator can handle rights bundles")
	}
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	if newRightsBundle.BundleKey == "" {
		newRightsBundle.BundleKey = types.VcloudUndefinedKey
	}
	if newRightsBundle.PublishAll == nil {
		newRightsBundle.PublishAll = addrOf(false)
	}
	returnBundle := &RightsBundle{
		RightsBundle: &types.RightsBundle{},
		client:       client,
	}

	err = client.OpenApiPostItem(minimumApiVersion, urlRef, nil, newRightsBundle, returnBundle.RightsBundle, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating rights bundle: %s", err)
	}

	return returnBundle, nil
}

// Update updates existing rights bundle
func (rb *RightsBundle) Update() (*RightsBundle, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	minimumApiVersion, err := rb.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if rb.RightsBundle.Id == "" {
		return nil, fmt.Errorf("cannot update role without id")
	}

	urlRef, err := rb.client.OpenApiBuildEndpoint(endpoint, rb.RightsBundle.Id)
	if err != nil {
		return nil, err
	}

	returnRightsBundle := &RightsBundle{
		RightsBundle: &types.RightsBundle{},
		client:       rb.client,
	}

	err = rb.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, rb.RightsBundle, returnRightsBundle.RightsBundle, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating rights bundle: %s", err)
	}

	return returnRightsBundle, nil
}

// getAllRightsBundles retrieves all rights bundles. Query parameters can be supplied to perform additional
// filtering
func getAllRightsBundles(client *Client, queryParameters url.Values, additionalHeader map[string]string) ([]*RightsBundle, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("only system administrator can handle rights bundles")
	}
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.RightsBundle{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, additionalHeader)
	if err != nil {
		return nil, err
	}
	if len(typeResponses) == 0 {
		return []*RightsBundle{}, nil
	}
	var results = make([]*RightsBundle, len(typeResponses))
	for i, r := range typeResponses {
		results[i] = &RightsBundle{
			RightsBundle: r,
			client:       client,
		}
	}

	return results, nil
}

// GetAllRightsBundles retrieves all rights bundles. Query parameters can be supplied to perform additional
// filtering
func (client *Client) GetAllRightsBundles(queryParameters url.Values) ([]*RightsBundle, error) {
	return getAllRightsBundles(client, queryParameters, nil)
}

// GetTenants retrieves all tenants associated to a given Rights Bundle.
// Query parameters can be supplied to perform additional filtering
func (rb *RightsBundle) GetTenants(queryParameters url.Values) ([]types.OpenApiReference, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return getContainerTenants(rb.client, rb.RightsBundle.Id, endpoint, queryParameters)
}

func (rb *RightsBundle) GetRights(queryParameters url.Values) ([]*types.Right, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return getRights(rb.client, rb.RightsBundle.Id, endpoint, queryParameters, nil)
}

// AddRights adds a collection of rights to a rights bundle
func (rb *RightsBundle) AddRights(newRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return addRightsToRole(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, newRights, nil)
}

// UpdateRights replaces existing rights with the given collection of rights
func (rb *RightsBundle) UpdateRights(newRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return updateRightsInRole(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, newRights, nil)
}

// RemoveRights removes specific rights from a rights bundle
func (rb *RightsBundle) RemoveRights(removeRights []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return removeRightsFromRole(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, removeRights, nil)
}

// RemoveAllRights removes all rights from a rights bundle
func (rb *RightsBundle) RemoveAllRights() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return removeAllRightsFromRole(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, nil)
}

// PublishTenants publishes a rights bundle to one or more tenants
func (rb *RightsBundle) PublishTenants(tenants []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return publishContainerToTenants(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, tenants, "add")
}

// UnpublishTenants removes publication status in rights bundle from one or more tenants
func (rb *RightsBundle) UnpublishTenants(tenants []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return publishContainerToTenants(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, tenants, "remove")
}

// ReplacePublishedTenants publishes a rights bundle to one or more tenants, removing the tenants already present
func (rb *RightsBundle) ReplacePublishedTenants(tenants []types.OpenApiReference) error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return publishContainerToTenants(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, tenants, "replace")
}

// PublishAllTenants removes publication status in rights bundle from one or more tenants
func (rb *RightsBundle) PublishAllTenants() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return publishContainerToAllTenants(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, true)
}

// UnpublishAllTenants removes publication status in rights bundle from one or more tenants
func (rb *RightsBundle) UnpublishAllTenants() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	return publishContainerToAllTenants(rb.client, "RightsBundle", rb.RightsBundle.Name, rb.RightsBundle.Id, endpoint, false)
}

// GetRightsBundleByName retrieves rights bundle by given name
func (client *Client) GetRightsBundleByName(name string) (*RightsBundle, error) {
	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	rightsBundles, err := client.GetAllRightsBundles(queryParams)
	if err != nil {
		return nil, err
	}
	if len(rightsBundles) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(rightsBundles) > 1 {
		return nil, fmt.Errorf("more than one rights bundle found with name '%s'", name)
	}
	return rightsBundles[0], nil
}

// GetRightsBundleById retrieves rights bundle by given ID
func (client *Client) GetRightsBundleById(id string) (*RightsBundle, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty rights bundle id")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	rightsBundle := &RightsBundle{
		RightsBundle: &types.RightsBundle{},
		client:       client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, rightsBundle.RightsBundle, nil)
	if err != nil {
		return nil, err
	}

	return rightsBundle, nil
}

// Delete deletes rights bundle
func (rb *RightsBundle) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRightsBundles
	minimumApiVersion, err := rb.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if rb.RightsBundle.Id == "" {
		return fmt.Errorf("cannot delete rights bundle without id")
	}

	urlRef, err := rb.client.OpenApiBuildEndpoint(endpoint, rb.RightsBundle.Id)
	if err != nil {
		return err
	}

	err = rb.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting rights bundle: %s", err)
	}

	return nil
}
