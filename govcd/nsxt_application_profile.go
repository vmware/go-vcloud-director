/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtAppPortProfile uses OpenAPI endpoint to operate NSX-T Application Port Profiles
// It can have 3 types of scopes:
// * SYSTEM - Read-only (The ones that are provided by SYSTEM). Constant `types.ApplicationPortProfileScopeSystem`
// * PROVIDER - Created by Provider on a particular network provider (NSX-T manager). Constant `types.ApplicationPortProfileScopeProvider`
// * TENANT (Created by Tenant at Org VDC level). Constant `types.ApplicationPortProfileScopeTenant`
//
// More details about scope in documentation for types.NsxtAppPortProfile
type NsxtAppPortProfile struct {
	NsxtAppPortProfile *types.NsxtAppPortProfile
	client             *Client
}

// CreateNsxtAppPortProfile allows users to create NSX-T Application Port Profile definition.
// It can have 3 types of scopes:
// * SYSTEM (The ones that are provided by SYSTEM) Read-only
// * PROVIDER (Created by Provider globally)
// * TENANT (Create by tenant at Org level)
// More details about scope in documentation for types.NsxtAppPortProfile
func (org *Org) CreateNsxtAppPortProfile(appPortProfileConfig *types.NsxtAppPortProfile) (*NsxtAppPortProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles
	minimumApiVersion, err := org.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := org.client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAppPortProfile{
		NsxtAppPortProfile: &types.NsxtAppPortProfile{},
		client:             org.client,
	}

	err = org.client.OpenApiPostItem(minimumApiVersion, urlRef, nil, appPortProfileConfig, returnObject.NsxtAppPortProfile, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Application Port Profile: %s", err)
	}

	return returnObject, nil
}

// GetAllNsxtAppPortProfiles returns all NSX-T Application Port Profiles for specific scope
// More details about scope in documentation for types.NsxtAppPortProfile
func (org *Org) GetAllNsxtAppPortProfiles(queryParameters url.Values, scope string) ([]*NsxtAppPortProfile, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	if scope != "" {
		queryParams = queryParameterFilterAnd("scope=="+scope, queryParams)
	}

	return getAllNsxtAppPortProfiles(org.client, queryParams)
}

// GetNsxtAppPortProfileByName allows users to retrieve Application Port Profiles for specific scope.
// More details in documentation for types.NsxtAppPortProfile
//
// Note. Names are enforced to be unique per scope
func (org *Org) GetNsxtAppPortProfileByName(name, scope string) (*NsxtAppPortProfile, error) {
	queryParameters := url.Values{}
	if scope != "" {
		queryParameters = queryParameterFilterAnd("scope=="+scope, queryParameters)
	}

	return getNsxtAppPortProfileByName(org.client, name, queryParameters)
}

// GetNsxtAppPortProfileById retrieves NSX-T Application Port Profile by ID
func (org *Org) GetNsxtAppPortProfileById(id string) (*NsxtAppPortProfile, error) {
	return getNsxtAppPortProfileById(org.client, id)
}

// Update allows users to update NSX-T Application Port Profile
func (appPortProfile *NsxtAppPortProfile) Update(appPortProfileConfig *types.NsxtAppPortProfile) (*NsxtAppPortProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles
	minimumApiVersion, err := appPortProfile.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if appPortProfileConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T Application Port Profile without ID")
	}

	urlRef, err := appPortProfile.client.OpenApiBuildEndpoint(endpoint, appPortProfileConfig.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtAppPortProfile{
		NsxtAppPortProfile: &types.NsxtAppPortProfile{},
		client:             appPortProfile.client,
	}

	err = appPortProfile.client.OpenApiPutItem(minimumApiVersion, urlRef, nil, appPortProfileConfig, returnObject.NsxtAppPortProfile, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T Application Port Profile : %s", err)
	}

	return returnObject, nil
}

// Delete allows users to delete NSX-T Application Port Profile
func (appPortProfile *NsxtAppPortProfile) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles
	minimumApiVersion, err := appPortProfile.client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	if appPortProfile.NsxtAppPortProfile.ID == "" {
		return fmt.Errorf("cannot delete NSX-T Application Port Profile without ID")
	}

	urlRef, err := appPortProfile.client.OpenApiBuildEndpoint(endpoint, appPortProfile.NsxtAppPortProfile.ID)
	if err != nil {
		return err
	}

	err = appPortProfile.client.OpenApiDeleteItem(minimumApiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting NSX-T Application Port Profile: %s", err)
	}

	return nil
}

func getNsxtAppPortProfileByName(client *Client, name string, queryParameters url.Values) (*NsxtAppPortProfile, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("name=="+name, queryParams)

	allAppPortProfiles, err := getAllNsxtAppPortProfiles(client, queryParams)
	if err != nil {
		return nil, fmt.Errorf("could not find NSX-T Application Port Profile with name '%s': %s", name, err)
	}

	if len(allAppPortProfiles) == 0 {
		return nil, fmt.Errorf("%s: expected exactly one NSX-T Application Port Profile with name '%s'. Got %d", ErrorEntityNotFound, name, len(allAppPortProfiles))
	}

	if len(allAppPortProfiles) > 1 {
		return nil, fmt.Errorf("expected exactly one NSX-T Application Port Profile with name '%s'. Got %d", name, len(allAppPortProfiles))
	}

	return getNsxtAppPortProfileById(client, allAppPortProfiles[0].NsxtAppPortProfile.ID)
}

func getNsxtAppPortProfileById(client *Client, id string) (*NsxtAppPortProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty NSX-T Application Port Profile ID specified")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	appPortProfile := &NsxtAppPortProfile{
		NsxtAppPortProfile: &types.NsxtAppPortProfile{},
		client:             client,
	}

	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, appPortProfile.NsxtAppPortProfile, nil)
	if err != nil {
		return nil, err
	}

	return appPortProfile, nil
}

func getAllNsxtAppPortProfiles(client *Client, queryParameters url.Values) ([]*NsxtAppPortProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtAppPortProfile{{}}
	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtAppPortProfile types with client
	wrappedResponses := make([]*NsxtAppPortProfile, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtAppPortProfile{
			NsxtAppPortProfile: typeResponses[sliceIndex],
			client:             client,
		}
	}

	return wrappedResponses, nil
}
