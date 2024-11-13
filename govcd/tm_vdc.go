package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmVdc = "Vdc"

// TmVdc defines Tenant Manager Virtual Data Center structure
type TmVdc struct {
	TmVdc     *types.TmVdc
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmVdc) wrap(inner *types.TmVdc) *TmVdc {
	g.TmVdc = inner
	return &g
}

// CreateTmVdc sets up a new Tenant Manager VDC
func (vcdClient *VCDClient) CreateTmVdc(config *types.TmVdc) (*TmVdc, error) {
	c := crudConfig{
		entityLabel: labelTmVdc,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		requiresTm:  true,
	}
	outerType := TmVdc{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllTmVdcs retrieves all Tenant Manager VDCs
func (vcdClient *VCDClient) GetAllTmVdcs(queryParameters url.Values) ([]*TmVdc, error) {
	c := crudConfig{
		entityLabel:     labelTmVdc,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmVdc{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmVdcByName retrieves Tenant Manager by a given name
func (vcdClient *VCDClient) GetTmVdcByName(name string) (*TmVdc, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmVdc)
	}

	// TODO - revisit filtering as filtering by name returns an error
	filteredEntities, err := vcdClient.GetAllTmVdcs(nil)
	if err != nil {
		return nil, err
	}

	for i := range filteredEntities {
		if filteredEntities[i].TmVdc.Name == name {
			return filteredEntities[i], nil
		}
	}

	return nil, fmt.Errorf("%s no VDC found by name '%s'", ErrorEntityNotFound, name)
}

// GetTmVdcById retrieves a Tenant Manager VDC by a given ID
func (vcdClient *VCDClient) GetTmVdcById(id string) (*TmVdc, error) {
	c := crudConfig{
		entityLabel:    labelTmVdc,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmVdc{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update Tenant Manager VDC
func (o *TmVdc) Update(TmVdcConfig *types.TmVdc) (*TmVdc, error) {
	c := crudConfig{
		entityLabel:    labelTmVdc,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
	}
	outerType := TmVdc{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmVdcConfig)
}

// Delete Tenant Manager VDC
func (o *TmVdc) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmVdc,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
