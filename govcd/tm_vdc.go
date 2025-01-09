package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmOrgVdc = "TM Org Vdc"

// TmVdc defines Tenant Manager Organization Virtual Data Center structure
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
		entityLabel: labelTmOrgVdc,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		requiresTm:  true,
	}
	outerType := TmVdc{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllTmVdcs retrieves all Tenant Manager VDCs
func (vcdClient *VCDClient) GetAllTmVdcs(queryParameters url.Values) ([]*TmVdc, error) {
	c := crudConfig{
		entityLabel:     labelTmOrgVdc,
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
		return nil, fmt.Errorf("%s lookup requires name", labelTmOrgVdc)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))
	filteredEntities, err := vcdClient.GetAllTmVdcs(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetTmVdcByNameAndOrgId retrieves VDC by Name and Org ID
func (vcdClient *VCDClient) GetTmVdcByNameAndOrgId(name, orgId string) (*TmVdc, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID to be present", labelTmOrgVdc)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("org.id==%s;name==%s", orgId, name))

	filteredEntities, err := vcdClient.GetAllTmVdcs(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetTmVdcById retrieves a Tenant Manager VDC by a given ID
func (vcdClient *VCDClient) GetTmVdcById(id string) (*TmVdc, error) {
	c := crudConfig{
		entityLabel:    labelTmOrgVdc,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmVdc{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update Tenant Manager VDC
func (o *TmVdc) Update(tmVdcConfig *types.TmVdc) (*TmVdc, error) {
	c := crudConfig{
		entityLabel:    labelTmOrgVdc,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
	}
	outerType := TmVdc{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, tmVdcConfig)
}

// Delete Tenant Manager VDC
func (o *TmVdc) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmOrgVdc,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// AssignVmClasses assigns VM Classes to the receiver VDC
func (o *TmVdc) AssignVmClasses(vmClasses types.OpenApiReferences) error {
	c := crudConfig{
		entityLabel: labelTmOrgVdc,
		endpoint:    types.OpenApiPathVcf + fmt.Sprintf(types.OpenApiEndpointTmVdcsVmClasses, o.TmVdc.ID),
		requiresTm:  true,
	}
	// It's a PUT call with OpenAPI references, so we reuse generic functions for simplicity
	_, err := updateInnerEntity[types.OpenApiReferences](&o.vcdClient.Client, c, &vmClasses)
	if err != nil {
		return err
	}
	return nil
}

// AssignStorageClasses assigns Storage Classes to the receiver VDC
func (o *TmVdc) AssignStorageClasses(storageClasses []StorageClass) error {
	c := crudConfig{
		entityLabel: labelTmOrgVdc,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmVdcStorageClasses,
		requiresTm:  true,
	}
	// It's a POST call with a list of Storage Classes, so we reuse generic functions for simplicity
	_, err := createInnerEntity[[]StorageClass](&o.vcdClient.Client, c, &storageClasses)
	if err != nil {
		return err
	}
	return nil
}
