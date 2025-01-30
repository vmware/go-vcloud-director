// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0
package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmRegionalNetworkingSetting = "Regional Networking Setting"
const labelTmRegionalNetworkingVpcConnectivityProfile = "Regional Networking VPC Connectivity Profile"

type TmRegionalNetworkingSetting struct {
	TmRegionalNetworkingSetting *types.TmRegionalNetworkingSetting
	vcdClient                   *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmRegionalNetworkingSetting) wrap(inner *types.TmRegionalNetworkingSetting) *TmRegionalNetworkingSetting {
	g.TmRegionalNetworkingSetting = inner
	return &g
}

// CreateTmRegionalNetworkingSetting creates a new Regional Networking Setting with a given configuration
func (vcdClient *VCDClient) CreateTmRegionalNetworkingSetting(config *types.TmRegionalNetworkingSetting) (*TmRegionalNetworkingSetting, error) {
	c := crudConfig{
		entityLabel: labelTmRegionalNetworkingSetting,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings,
		requiresTm:  true,
	}
	outerType := TmRegionalNetworkingSetting{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateTmRegionalNetworkingSettingAsync creates a new Regional Networking Setting with a given configuration and returns tracking task
func (vcdClient *VCDClient) CreateTmRegionalNetworkingSettingAsync(config *types.TmRegionalNetworkingSetting) (*Task, error) {
	c := crudConfig{
		entityLabel: labelTmRegionalNetworkingSetting,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings,
		requiresTm:  true,
	}
	return createInnerEntityAsync(&vcdClient.Client, c, config)
}

// GetAllTmRegionalNetworkingSettings retrieves all Regional Networking Settings with an optional filter
func (vcdClient *VCDClient) GetAllTmRegionalNetworkingSettings(queryParameters url.Values) ([]*TmRegionalNetworkingSetting, error) {
	c := crudConfig{
		entityLabel:     labelTmRegionalNetworkingSetting,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmRegionalNetworkingSetting{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmRegionalNetworkingSettingByName retrieves Regional Networking Setting by Name
func (vcdClient *VCDClient) GetTmRegionalNetworkingSettingByName(name string) (*TmRegionalNetworkingSetting, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmRegionalNetworkingSetting)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmRegionalNetworkingSettings(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmRegionalNetworkingSettingById(singleEntity.TmRegionalNetworkingSetting.ID)
}

// GetTmRegionalNetworkingSettingById retrieves Regional Networking Setting by ID
func (vcdClient *VCDClient) GetTmRegionalNetworkingSettingById(id string) (*TmRegionalNetworkingSetting, error) {
	c := crudConfig{
		entityLabel:    labelTmRegionalNetworkingSetting,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmRegionalNetworkingSetting{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetTmRegionalNetworkingSettingByNameAndOrgId retrieves Regional Networking Setting by Name and Org ID
func (vcdClient *VCDClient) GetTmRegionalNetworkingSettingByNameAndOrgId(name, orgId string) (*TmRegionalNetworkingSetting, error) {
	return getTmRegionalNetworkingSettingByNameAndRefId(vcdClient, name, "orgRef", orgId)
}

// GetTmRegionalNetworkingSettingByNameAndOrgId retrieves Regional Networking Setting by Name and Region ID
func (vcdClient *VCDClient) GetTmRegionalNetworkingSettingByNameAndRegionId(name, regionId string) (*TmRegionalNetworkingSetting, error) {
	return getTmRegionalNetworkingSettingByNameAndRefId(vcdClient, name, "regionRef", regionId)
}

// Update Regional Networking Setting with a given config
// Note. Only Name and Edge Cluster fields are updateable
func (o *TmRegionalNetworkingSetting) Update(TmRegionalNetworkingSettingConfig *types.TmRegionalNetworkingSetting) (*TmRegionalNetworkingSetting, error) {
	c := crudConfig{
		entityLabel:    labelTmRegionalNetworkingSetting,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings,
		endpointParams: []string{o.TmRegionalNetworkingSetting.ID},
		requiresTm:     true,
	}
	outerType := TmRegionalNetworkingSetting{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmRegionalNetworkingSettingConfig)
}

// Delete Regional Networking Setting
func (o *TmRegionalNetworkingSetting) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmRegionalNetworkingSetting,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings,
		endpointParams: []string{o.TmRegionalNetworkingSetting.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}

func getTmRegionalNetworkingSettingByNameAndRefId(vcdClient *VCDClient, name, refName, refId string) (*TmRegionalNetworkingSetting, error) {
	if name == "" || refId == "" {
		return nil, fmt.Errorf("%s lookup requires name and refName ID", labelTmRegionalNetworkingSetting)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("%s.id==%s", refName, refId), queryParams)

	filteredEntities, err := vcdClient.GetAllTmRegionalNetworkingSettings(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmRegionalNetworkingSettingById(singleEntity.TmRegionalNetworkingSetting.ID)
}

// GetDefaultVpcConnectivityProfile retrieves default VPC Connectivity profile for Org Regional Networking
func (o *TmRegionalNetworkingSetting) GetDefaultVpcConnectivityProfile() (*types.TmRegionalNetworkingVpcConnectivityProfile, error) {
	c := crudConfig{
		entityLabel:    labelTmRegionalNetworkingVpcConnectivityProfile,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettingsVpcProfile,
		endpointParams: []string{o.TmRegionalNetworkingSetting.ID},
		requiresTm:     true,
	}
	return getInnerEntity[types.TmRegionalNetworkingVpcConnectivityProfile](&o.vcdClient.Client, c)
}

// UpdateDefaultVpcConnectivityProfile changes default VPC Connectivity profile for Org Regional Networking
func (o *TmRegionalNetworkingSetting) UpdateDefaultVpcConnectivityProfile(regNetVpcProfileConfig *types.TmRegionalNetworkingVpcConnectivityProfile) (*types.TmRegionalNetworkingVpcConnectivityProfile, error) {
	c := crudConfig{
		entityLabel:    labelTmRegionalNetworkingVpcConnectivityProfile,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettingsVpcProfile,
		endpointParams: []string{o.TmRegionalNetworkingSetting.ID},
		requiresTm:     true,
	}
	return updateInnerEntity(&o.vcdClient.Client, c, regNetVpcProfileConfig)
}
