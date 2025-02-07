package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelRegionQuota = "Org Region Quota"

// RegionQuota defines Region Quota structure in Tenant Manager
type RegionQuota struct {
	TmVdc     *types.TmVdc
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g RegionQuota) wrap(inner *types.TmVdc) *RegionQuota {
	g.TmVdc = inner
	return &g
}

// CreateRegionQuota sets up a new Region Quota
func (vcdClient *VCDClient) CreateRegionQuota(config *types.TmVdc) (*RegionQuota, error) {
	c := crudConfig{
		entityLabel: labelRegionQuota,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		requiresTm:  true,
	}
	outerType := RegionQuota{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllRegionQuotas retrieves all Region Quotas
func (vcdClient *VCDClient) GetAllRegionQuotas(queryParameters url.Values) ([]*RegionQuota, error) {
	c := crudConfig{
		entityLabel:     labelRegionQuota,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := RegionQuota{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetRegionQuotaByName retrieves a Region Quota by a given name
func (vcdClient *VCDClient) GetRegionQuotaByName(name string) (*RegionQuota, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelRegionQuota)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))
	filteredEntities, err := vcdClient.GetAllRegionQuotas(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetRegionQuotaByNameAndOrgId retrieves a Region Quota by Name and Org ID
func (vcdClient *VCDClient) GetRegionQuotaByNameAndOrgId(name, orgId string) (*RegionQuota, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID to be present", labelRegionQuota)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("org.id==%s;name==%s", orgId, name))

	filteredEntities, err := vcdClient.GetAllRegionQuotas(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetRegionQuotaById retrieves a Region Quota by a given ID
func (vcdClient *VCDClient) GetRegionQuotaById(id string) (*RegionQuota, error) {
	c := crudConfig{
		entityLabel:    labelRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := RegionQuota{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update updates the receiver Region Quota
func (o *RegionQuota) Update(tmVdcConfig *types.TmVdc) (*RegionQuota, error) {
	c := crudConfig{
		entityLabel:    labelRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
	}
	outerType := RegionQuota{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, tmVdcConfig)
}

// Delete deletes the receiver Region Quota
func (o *RegionQuota) Delete() error {
	c := crudConfig{
		entityLabel:    labelRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// AssignVmClassesToRegionQuota assigns VM Classes to the receiver Region Quota
func (o *VCDClient) AssignVmClassesToRegionQuota(regionQuotaId string, vmClasses *types.RegionVirtualMachineClasses) error {
	c := crudConfig{
		entityLabel:    labelRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcsVmClasses,
		endpointParams: []string{regionQuotaId},
		requiresTm:     true,
	}
	// It's a PUT call with OpenAPI references, so we reuse generic functions for simplicity
	_, err := updateInnerEntity[types.RegionVirtualMachineClasses](&o.Client, c, vmClasses)
	if err != nil {
		return err
	}
	return nil
}

// AssignVmClasses assigns VM Classes to the receiver Region Quota
func (o *RegionQuota) AssignVmClasses(vmClasses *types.RegionVirtualMachineClasses) error {
	return o.vcdClient.AssignVmClassesToRegionQuota(o.TmVdc.ID, vmClasses)
}

// GetVmClassesFromRegionQuota returns all VM Classes of the given Region Quota
func (o *VCDClient) GetVmClassesFromRegionQuota(regionQuotaId string) (*types.RegionVirtualMachineClasses, error) {
	c := crudConfig{
		entityLabel:    labelRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcsVmClasses,
		endpointParams: []string{regionQuotaId},
		requiresTm:     true,
	}
	// It's a GET call with OpenAPI references, so we reuse generic functions for simplicity
	result, err := getInnerEntity[types.RegionVirtualMachineClasses](&o.Client, c)
	if err != nil {
		return nil, err
	}
	return result, nil
}
