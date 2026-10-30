// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/http"
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
	queryparams := copyOrNewUrlValues(queryParameters)
	queryparams.Set("pageSize", "32")
	c := crudConfig{
		entityLabel:     labelTmRegionalNetworkingSetting,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmRegionalNetworkingSettings,
		queryParameters: queryparams,
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

func (o *TmRegionalNetworkingSetting) operation(endpoint, minimumVersion string) crudConfig {
	return crudConfig{
		entityLabel:    "regionalNetworkingSettings/" + endpoint,
		endpoint:       types.OpenApiPathVcf + ("regionalNetworkingSettings/%s/" + endpoint),
		endpointParams: []string{o.TmRegionalNetworkingSetting.ID},
		requiresTm:     true,
		minimumVersion: minimumVersion,
	}
}

// GetAllProviderGatewayConnections retrieves all providerGatewayConnections assigned to this regional networking setting.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) GetAllProviderGatewayConnections(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmRegionalProviderGatewayConnection, error) {
	c := o.operation("providerGatewayConnections", "40.1")
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmRegionalProviderGatewayConnection](&o.vcdClient.Client, c.withTmOptions(options))
}

// GetProviderGatewayConnectionById retrieves one providerGatewayConnections assignment by its ID.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) GetProviderGatewayConnectionById(id string, options ...TmOperationOptions) (*types.TmRegionalProviderGatewayConnection, error) {
	return getInnerEntity[types.TmRegionalProviderGatewayConnection](&o.vcdClient.Client, o.operation("providerGatewayConnections", "40.1").item(id).withTmOptions(options))
}

// CreateProviderGatewayConnection adds an assignment and waits for its task to complete.
// Use GetAllProviderGatewayConnections to retrieve the resulting assignments.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) CreateProviderGatewayConnection(config *types.TmRegionalProviderGatewayConnection, options ...TmOperationOptions) error {
	task, err := o.CreateProviderGatewayConnectionAsync(config, options...)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// CreateProviderGatewayConnectionAsync adds an assignment and returns its tracking task.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) CreateProviderGatewayConnectionAsync(config *types.TmRegionalProviderGatewayConnection, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("providerGatewayConnections config must not be nil")
	}
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("providerGatewayConnections", "40.1").withTmOptions(options), http.MethodPost, config)
}

// DeleteProviderGatewayConnection removes an assignment and waits for its task to complete.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteProviderGatewayConnection(id string, options ...TmOperationOptions) error {
	return runOpenApiOperation(&o.vcdClient.Client, o.operation("providerGatewayConnections", "40.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// DeleteProviderGatewayConnectionAsync removes an assignment and returns its tracking task.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteProviderGatewayConnectionAsync(id string, options ...TmOperationOptions) (*Task, error) {
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("providerGatewayConnections", "40.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// GetAllSharedSubnets retrieves all sharedSubnets assigned to this regional networking setting.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) GetAllSharedSubnets(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmRegionalSharedSubnet, error) {
	c := o.operation("sharedSubnets", "40.1")
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmRegionalSharedSubnet](&o.vcdClient.Client, c.withTmOptions(options))
}

// GetSharedSubnetById retrieves one sharedSubnets assignment by its ID.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) GetSharedSubnetById(id string, options ...TmOperationOptions) (*types.TmRegionalSharedSubnet, error) {
	return getInnerEntity[types.TmRegionalSharedSubnet](&o.vcdClient.Client, o.operation("sharedSubnets", "40.1").item(id).withTmOptions(options))
}

// CreateSharedSubnet adds an assignment and waits for its task to complete.
// Use GetAllSharedSubnets to retrieve the resulting assignments.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) CreateSharedSubnet(config *types.TmRegionalSharedSubnet, options ...TmOperationOptions) error {
	task, err := o.CreateSharedSubnetAsync(config, options...)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// CreateSharedSubnetAsync adds an assignment and returns its tracking task.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) CreateSharedSubnetAsync(config *types.TmRegionalSharedSubnet, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("sharedSubnets config must not be nil")
	}
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("sharedSubnets", "40.1").withTmOptions(options), http.MethodPost, config)
}

// DeleteSharedSubnet removes an assignment and waits for its task to complete.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteSharedSubnet(id string, options ...TmOperationOptions) error {
	return runOpenApiOperation(&o.vcdClient.Client, o.operation("sharedSubnets", "40.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// DeleteSharedSubnetAsync removes an assignment and returns its tracking task.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteSharedSubnetAsync(id string, options ...TmOperationOptions) (*Task, error) {
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("sharedSubnets", "40.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// GetAllDistributedVlanConnections retrieves all distributedVlanConnections assigned to this regional networking setting.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) GetAllDistributedVlanConnections(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmRegionalDistributedVlanConnection, error) {
	c := o.operation("distributedVlanConnections", "40.1")
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmRegionalDistributedVlanConnection](&o.vcdClient.Client, c.withTmOptions(options))
}

// GetDistributedVlanConnectionById retrieves one distributedVlanConnections assignment by its ID.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) GetDistributedVlanConnectionById(id string, options ...TmOperationOptions) (*types.TmRegionalDistributedVlanConnection, error) {
	return getInnerEntity[types.TmRegionalDistributedVlanConnection](&o.vcdClient.Client, o.operation("distributedVlanConnections", "40.1").item(id).withTmOptions(options))
}

// CreateDistributedVlanConnection adds an assignment and waits for its task to complete.
// Use GetAllDistributedVlanConnections to retrieve the resulting assignments.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) CreateDistributedVlanConnection(config *types.TmRegionalDistributedVlanConnection, options ...TmOperationOptions) error {
	task, err := o.CreateDistributedVlanConnectionAsync(config, options...)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// CreateDistributedVlanConnectionAsync adds an assignment and returns its tracking task.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) CreateDistributedVlanConnectionAsync(config *types.TmRegionalDistributedVlanConnection, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("distributedVlanConnections config must not be nil")
	}
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("distributedVlanConnections", "40.1").withTmOptions(options), http.MethodPost, config)
}

// DeleteDistributedVlanConnection removes an assignment and waits for its task to complete.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteDistributedVlanConnection(id string, options ...TmOperationOptions) error {
	return runOpenApiOperation(&o.vcdClient.Client, o.operation("distributedVlanConnections", "40.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// DeleteDistributedVlanConnectionAsync removes an assignment and returns its tracking task.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteDistributedVlanConnectionAsync(id string, options ...TmOperationOptions) (*Task, error) {
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("distributedVlanConnections", "40.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// GetAllDistributedVxlanConnections retrieves all distributedVxlanConnections assigned to this regional networking setting.
// Requires TM API 9.1.1 or later.
func (o *TmRegionalNetworkingSetting) GetAllDistributedVxlanConnections(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmRegionalDistributedVxlanConnection, error) {
	c := o.operation("distributedVxlanConnections", "9.1.1")
	c.queryParameters = queryParameters
	return getAllInnerEntities[types.TmRegionalDistributedVxlanConnection](&o.vcdClient.Client, c.withTmOptions(options))
}

// GetDistributedVxlanConnectionById retrieves one distributedVxlanConnections assignment by its ID.
// Requires TM API 9.1.1 or later.
func (o *TmRegionalNetworkingSetting) GetDistributedVxlanConnectionById(id string, options ...TmOperationOptions) (*types.TmRegionalDistributedVxlanConnection, error) {
	return getInnerEntity[types.TmRegionalDistributedVxlanConnection](&o.vcdClient.Client, o.operation("distributedVxlanConnections", "9.1.1").item(id).withTmOptions(options))
}

// CreateDistributedVxlanConnection adds an assignment and waits for its task to complete.
// Use GetAllDistributedVxlanConnections to retrieve the resulting assignments.
// Requires TM API 9.1.1 or later.
func (o *TmRegionalNetworkingSetting) CreateDistributedVxlanConnection(config *types.TmRegionalDistributedVxlanConnection, options ...TmOperationOptions) error {
	task, err := o.CreateDistributedVxlanConnectionAsync(config, options...)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// CreateDistributedVxlanConnectionAsync adds an assignment and returns its tracking task.
// Requires TM API 9.1.1 or later.
func (o *TmRegionalNetworkingSetting) CreateDistributedVxlanConnectionAsync(config *types.TmRegionalDistributedVxlanConnection, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("distributedVxlanConnections config must not be nil")
	}
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("distributedVxlanConnections", "9.1.1").withTmOptions(options), http.MethodPost, config)
}

// DeleteDistributedVxlanConnection removes an assignment and waits for its task to complete.
// Requires TM API 9.1.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteDistributedVxlanConnection(id string, options ...TmOperationOptions) error {
	return runOpenApiOperation(&o.vcdClient.Client, o.operation("distributedVxlanConnections", "9.1.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// DeleteDistributedVxlanConnectionAsync removes an assignment and returns its tracking task.
// Requires TM API 9.1.1 or later.
func (o *TmRegionalNetworkingSetting) DeleteDistributedVxlanConnectionAsync(id string, options ...TmOperationOptions) (*Task, error) {
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("distributedVxlanConnections", "9.1.1").item(id).withTmOptions(options), http.MethodDelete, nil)
}

// GetAviSetting retrieves this regional networking setting's aviSetting.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) GetAviSetting(options ...TmOperationOptions) (*types.TmRegionalAviSetting, error) {
	return getInnerEntity[types.TmRegionalAviSetting](&o.vcdClient.Client, o.operation("aviSetting", "40.1").withTmOptions(options))
}

// UpdateAviSetting updates the configuration, waits for completion and retrieves the result.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) UpdateAviSetting(config *types.TmRegionalAviSetting, options ...TmOperationOptions) (*types.TmRegionalAviSetting, error) {
	task, err := o.UpdateAviSettingAsync(config, options...)
	if err != nil {
		return nil, err
	}
	if err := task.WaitTaskCompletion(); err != nil {
		return nil, err
	}
	return o.GetAviSetting(options...)
}

// UpdateAviSettingAsync updates the configuration and returns its tracking task.
// Requires TM API 40.1 or later.
func (o *TmRegionalNetworkingSetting) UpdateAviSettingAsync(config *types.TmRegionalAviSetting, options ...TmOperationOptions) (*Task, error) {
	if config == nil {
		return nil, fmt.Errorf("aviSetting config must not be nil")
	}
	return runOpenApiOperationAsync(&o.vcdClient.Client, o.operation("aviSetting", "40.1").withTmOptions(options), http.MethodPut, config)
}
