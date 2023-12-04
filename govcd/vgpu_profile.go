package govcd

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VgpuProfile defines a vGPU profile which is fetched from vCenter
type VgpuProfile struct {
	VgpuProfile *types.VgpuProfile
	client      *Client
}

// GetAllVgpuProfiles gets all vGPU profiles that are available to VCD
func (client *VCDClient) GetAllVgpuProfiles(queryParameters url.Values) ([]*VgpuProfile, error) {
	return getAllVgpuProfiles(queryParameters, &client.Client)
}

// GetVgpuProfilesByProviderVdc gets all vGPU profiles that are available to a specific provider VDC
func (client *VCDClient) GetVgpuProfilesByProviderVdc(providerVdcUrn string) ([]*VgpuProfile, error) {
	queryParameters := url.Values{}
	queryParameters = queryParameterFilterAnd(fmt.Sprintf("pvdcId==%s", providerVdcUrn), queryParameters)
	return client.GetAllVgpuProfiles(queryParameters)
}

// GetVgpuProfileById gets a vGPU profile by ID
func (client *VCDClient) GetVgpuProfileById(vgpuProfileId string) (*VgpuProfile, error) {
	return getVgpuProfileById(vgpuProfileId, &client.Client)
}

// GetVgpuProfileByName gets a vGPU profile by name
func (client *VCDClient) GetVgpuProfileByName(vgpuProfileName string) (*VgpuProfile, error) {
	return getVgpuProfileByFilter("name", vgpuProfileName, &client.Client)
}

// GetVgpuProfileByTenantFacingName gets a vGPU profile by its tenant facing name
func (client *VCDClient) GetVgpuProfileByTenantFacingName(tenantFacingName string) (*VgpuProfile, error) {
	return getVgpuProfileByFilter("tenantFacingName", tenantFacingName, &client.Client)
}

// Update updates a vGPU profile with new parameters
func (profile *VgpuProfile) Update(newProfile *types.VgpuProfile) error {
	client := profile.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVgpuProfile
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, "/", profile.VgpuProfile.Id)
	if err != nil {
		return err
	}

	err = client.OpenApiPutItemSync(minimumApiVersion, urlRef, nil, newProfile, nil, nil)
	if err != nil {
		return err
	}

	// We need to refresh here, as PUT returns the original struct instead of the updated one
	err = profile.Refresh()
	if err != nil {
		return err
	}

	return nil
}

// Refresh updates the current state of the vGPU profile
func (profile *VgpuProfile) Refresh() error {
	var err error
	newProfile, err := getVgpuProfileById(profile.VgpuProfile.Id, profile.client)
	if err != nil {
		return err
	}
	profile.VgpuProfile = newProfile.VgpuProfile

	return nil
}

func getVgpuProfileByFilter(filter, filterValue string, client *Client) (*VgpuProfile, error) {
	queryParameters := url.Values{}
	queryParameters = queryParameterFilterAnd(fmt.Sprintf("%s==%s", filter, filterValue), queryParameters)
	vgpuProfiles, err := getAllVgpuProfiles(queryParameters, client)
	if err != nil {
		return nil, err
	}

	vgpuProfile, err := oneOrError(filter, filterValue, vgpuProfiles)
	if err != nil {
		return nil, err
	}

	return vgpuProfile, nil
}

func getVgpuProfileById(vgpuProfileId string, client *Client) (*VgpuProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVgpuProfile
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, "/", vgpuProfileId)
	if err != nil {
		return nil, err
	}

	profile := &VgpuProfile{
		client: client,
	}
	err = client.OpenApiGetItem(minimumApiVersion, urlRef, nil, &profile.VgpuProfile, nil)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func getAllVgpuProfiles(queryParameters url.Values, client *Client) ([]*VgpuProfile, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVgpuProfile
	minimumApiVersion, err := client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	responses := []*types.VgpuProfile{{}}

	err = client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParameters, &responses, nil)
	if err != nil {
		return nil, err
	}

	wrappedVgpuProfiles := make([]*VgpuProfile, len(responses))
	for index, response := range responses {
		wrappedVgpuProfile := &VgpuProfile{
			client:      client,
			VgpuProfile: response,
		}
		wrappedVgpuProfiles[index] = wrappedVgpuProfile
	}

	return wrappedVgpuProfiles, nil
}
