package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelOrganization = "Organization"

type TmOrg struct {
	TmOrg     *types.TmOrg
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmOrg) wrap(inner *types.TmOrg) *TmOrg {
	g.TmOrg = inner
	return &g
}

// CreateTmOrg creates a TM Organization
func (vcdClient *VCDClient) CreateTmOrg(config *types.TmOrg) (*TmOrg, error) {
	c := crudConfig{
		entityLabel: labelOrganization,
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs,
		requiresTm:  true,
	}
	outerType := TmOrg{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllTmOrgs retrieves all TM Organization with an optional query filter
func (vcdClient *VCDClient) GetAllTmOrgs(queryParameters url.Values) ([]*TmOrg, error) {
	c := crudConfig{
		entityLabel:     labelOrganization,
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmOrg{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmOrgByName retrieves TM Organization by name
func (vcdClient *VCDClient) GetTmOrgByName(name string) (*TmOrg, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelOrganization)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmOrgs(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmOrgById(singleEntity.TmOrg.ID)
}

// GetTmOrgById retrieves TM Organization by ID
func (vcdClient *VCDClient) GetTmOrgById(id string) (*TmOrg, error) {
	c := crudConfig{
		entityLabel:    labelOrganization,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmOrg{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update TM Organization
func (o *TmOrg) Update(TmOrgConfig *types.TmOrg) (*TmOrg, error) {
	c := crudConfig{
		entityLabel:    labelOrganization,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs,
		endpointParams: []string{o.TmOrg.ID},
		requiresTm:     true,
	}
	outerType := TmOrg{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmOrgConfig)
}

// Delete TM Organization
func (o *TmOrg) Delete() error {
	c := crudConfig{
		entityLabel:    labelOrganization,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgs,
		endpointParams: []string{o.TmOrg.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// Disable is a shortcut to disable TM Organization
func (o *TmOrg) Disable() error {
	o.TmOrg.IsEnabled = false
	_, err := o.Update(o.TmOrg)
	return err
}
