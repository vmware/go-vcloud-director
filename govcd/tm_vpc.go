package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVirtualPrivateCloud = "Virtual Private Cloud"

type VirtualPrivateCloud struct {
	VirtualPrivateCloud *types.VirtualPrivateCloud
	client              *Client
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g VirtualPrivateCloud) wrap(inner *types.VirtualPrivateCloud) *VirtualPrivateCloud {
	g.VirtualPrivateCloud = inner
	return &g
}

// CreateVirtualPrivateCloud creates a Virtual Private Cloud
func (vdc *Vdc) CreateVirtualPrivateCloud(config *types.VirtualPrivateCloud) (*VirtualPrivateCloud, error) {
	c := crudConfig{
		entityLabel: labelVirtualPrivateCloud,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointVpcs,
		requiresTm:  true,
	}
	outerType := VirtualPrivateCloud{client: vdc.client}
	return createOuterEntity(vdc.client, outerType, c, config)
}

// GetAllVirtualPrivateClouds retrieves all Virtual Private Clouds with an optional query filter
func (vdc *Vdc) GetAllVirtualPrivateClouds(queryParameters url.Values) ([]*VirtualPrivateCloud, error) {
	c := crudConfig{
		entityLabel:     labelVirtualPrivateCloud,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointVpcs,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := VirtualPrivateCloud{client: vdc.client}
	return getAllOuterEntities(vdc.client, outerType, c)
}

// GetVirtualPrivateCloudByName retrieves Virtual Private Cloud by name
func (vdc *Vdc) GetVirtualPrivateCloudByName(name string) (*VirtualPrivateCloud, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelVirtualPrivateCloud)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vdc.GetAllVirtualPrivateClouds(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vdc.GetVirtualPrivateCloudById(singleEntity.VirtualPrivateCloud.Id)
}

// GetVirtualPrivateCloudById retrieves Virtual Private Cloud by ID
func (vdc *Vdc) GetVirtualPrivateCloudById(id string) (*VirtualPrivateCloud, error) {
	c := crudConfig{
		entityLabel:    labelVirtualPrivateCloud,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointVpcs,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := VirtualPrivateCloud{client: vdc.client}
	return getOuterEntity(vdc.client, outerType, c)
}

// Update Virtual Private Cloud
func (o *VirtualPrivateCloud) Update(VirtualPrivateCloudConfig *types.VirtualPrivateCloud) (*VirtualPrivateCloud, error) {
	c := crudConfig{
		entityLabel:    labelVirtualPrivateCloud,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointVpcs,
		endpointParams: []string{o.VirtualPrivateCloud.Id},
		requiresTm:     true,
	}
	outerType := VirtualPrivateCloud{client: o.client}
	return updateOuterEntity(o.client, outerType, c, VirtualPrivateCloudConfig)
}

// Delete Virtual Private Cloud
func (o *VirtualPrivateCloud) Delete() error {
	c := crudConfig{
		entityLabel:    labelVirtualPrivateCloud,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointVpcs,
		endpointParams: []string{o.VirtualPrivateCloud.Id},
		requiresTm:     true,
	}
	return deleteEntityById(o.client, c)
}
