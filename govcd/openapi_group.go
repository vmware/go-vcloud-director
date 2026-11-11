// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelOpenApiGroup = "Group"

type OpenApiGroup struct {
	Group         *types.OpenApiGroup
	vcdClient     *VCDClient
	TenantContext *TenantContext
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g OpenApiGroup) wrap(inner *types.OpenApiGroup) *OpenApiGroup {
	g.Group = inner
	return &g
}

// CreateGroup creates a new Group with a given configuration
func (vcdClient *VCDClient) CreateGroup(config *types.OpenApiGroup, ctx *TenantContext) (*OpenApiGroup, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiGroup,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGroups,
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}
	outerType := OpenApiGroup{vcdClient: vcdClient, TenantContext: ctx}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllGroups retrieves all Groups with an optional filter
func (vcdClient *VCDClient) GetAllGroups(queryParameters url.Values, ctx *TenantContext) ([]*OpenApiGroup, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiGroup,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGroups,
		queryParameters:  queryParameters,
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := OpenApiGroup{vcdClient: vcdClient, TenantContext: ctx}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetGroupByName retrieves Group by Name
func (vcdClient *VCDClient) GetGroupByName(name string, ctx *TenantContext) (*OpenApiGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelOpenApiGroup)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllGroups(queryParams, ctx)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetGroupById(singleEntity.Group.ID, ctx)
}

// GetGroupById retrieves Group by ID
func (vcdClient *VCDClient) GetGroupById(id string, ctx *TenantContext) (*OpenApiGroup, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiGroup,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGroups,
		endpointParams:   []string{id},
		additionalHeader: getTenantContextHeader(ctx),
		requiresTm:       true,
	}

	outerType := OpenApiGroup{vcdClient: vcdClient, TenantContext: ctx}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update Group with a given config
func (o *OpenApiGroup) Update(cfg *types.OpenApiGroup) (*OpenApiGroup, error) {
	c := crudConfig{
		entityLabel:      labelOpenApiGroup,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGroups,
		endpointParams:   []string{o.Group.ID},
		additionalHeader: getTenantContextHeader(o.TenantContext),
		requiresTm:       true,
	}
	outerType := OpenApiGroup{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, cfg)
}

// Delete Group
func (o *OpenApiGroup) Delete() error {
	c := crudConfig{
		entityLabel:      labelOpenApiGroup,
		endpoint:         types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointGroups,
		endpointParams:   []string{o.Group.ID},
		additionalHeader: getTenantContextHeader(o.TenantContext),
		requiresTm:       true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
