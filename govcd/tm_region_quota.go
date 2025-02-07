package govcd

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmOrgRegionQuota = "Org Region Quota"

// OrgRegionQuota defines Tenant Manager Organization Virtual Data Center structure
type OrgRegionQuota struct {
	OrgRegionQuota *types.TmVdc
	vcdClient      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g OrgRegionQuota) wrap(inner *types.TmVdc) *OrgRegionQuota {
	g.OrgRegionQuota = inner
	return &g
}

// CreateOrgRegionQuota sets up a new Org Region Quota
func (vcdClient *VCDClient) CreateOrgRegionQuota(config *types.TmVdc) (*OrgRegionQuota, error) {
	c := crudConfig{
		entityLabel: labelTmOrgRegionQuota,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		requiresTm:  true,
	}
	outerType := OrgRegionQuota{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllOrgRegionQuotas retrieves all Org Region Quotas
func (vcdClient *VCDClient) GetAllOrgRegionQuotas(queryParameters url.Values) ([]*OrgRegionQuota, error) {
	c := crudConfig{
		entityLabel:     labelTmOrgRegionQuota,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := OrgRegionQuota{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetOrgRegionQuotaByName retrieves an Org Region Quota by a given name
func (vcdClient *VCDClient) GetOrgRegionQuotaByName(name string) (*OrgRegionQuota, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmOrgRegionQuota)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))
	filteredEntities, err := vcdClient.GetAllOrgRegionQuotas(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetOrgRegionQuotaByNameAndOrgId retrieves an Org Region Quota by Name and Org ID
func (vcdClient *VCDClient) GetOrgRegionQuotaByNameAndOrgId(name, orgId string) (*OrgRegionQuota, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID to be present", labelTmOrgRegionQuota)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("org.id==%s;name==%s", orgId, name))

	filteredEntities, err := vcdClient.GetAllOrgRegionQuotas(queryParams)
	if err != nil {
		return nil, err
	}

	singleResult, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleResult, nil
}

// GetOrgRegionQuotaById retrieves an Org Region Quota by a given ID
func (vcdClient *VCDClient) GetOrgRegionQuotaById(id string) (*OrgRegionQuota, error) {
	c := crudConfig{
		entityLabel:    labelTmOrgRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := OrgRegionQuota{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update updates the receiver Org Region Quota
func (o *OrgRegionQuota) Update(tmVdcConfig *types.TmVdc) (*OrgRegionQuota, error) {
	c := crudConfig{
		entityLabel:    labelTmOrgRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.OrgRegionQuota.ID},
		requiresTm:     true,
	}
	outerType := OrgRegionQuota{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, tmVdcConfig)
}

// Delete deletes the receiver Org Region Quota
func (o *OrgRegionQuota) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmOrgRegionQuota,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		endpointParams: []string{o.OrgRegionQuota.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}

// AssignVmClassesToOrgRegionQuota assigns VM Classes to the receiver Org Region Quota
func (o *VCDClient) AssignVmClassesToOrgRegionQuota(regionQuotaId string, vmClasses *types.RegionVirtualMachineClasses) error {
	c := crudConfig{
		entityLabel:    labelTmOrgRegionQuota,
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

// AssignVmClasses assigns VM Classes to the receiver Org Region Quota
func (o *OrgRegionQuota) AssignVmClasses(vmClasses *types.RegionVirtualMachineClasses) error {
	return o.vcdClient.AssignVmClassesToOrgRegionQuota(o.OrgRegionQuota.ID, vmClasses)
}

// GetVmClassesFromOrgRegionQuota returns all VM Classes of the given Org Region Quota
func (o *VCDClient) GetVmClassesFromOrgRegionQuota(regionQuotaId string) (*types.RegionVirtualMachineClasses, error) {
	c := crudConfig{
		entityLabel:    labelTmOrgRegionQuota,
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
