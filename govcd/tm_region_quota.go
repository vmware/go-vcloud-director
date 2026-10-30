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
	// If each OpenAPI reference does not have Name populated, set it to avoid backend errors
	for i := range vmClasses.Values {
		if vmClasses.Values[i].Name == "" {
			vmc, err := o.GetRegionVirtualMachineClassById(vmClasses.Values[i].ID)
			if err != nil {
				return err
			}
			vmClasses.Values[i].Name = vmc.RegionVirtualMachineClass.Name
		}
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

// DeleteWithForce deletes the RegionQuota and waits for completion (API 40.0+).
// Setting force to true requests forceful deletion even if the backing resource cannot be removed normally.
func (o *RegionQuota) DeleteWithForce(force bool, options ...TmOperationOptions) error {
	c := crudConfig{
		entityLabel:     types.OpenApiEndpointTmVdcs,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVdcs,
		queryParameters: url.Values{"force": {fmt.Sprintf("%t", force)}},
		requiresTm:      true,
		minimumVersion:  "40.0",
	}
	return runOpenApiOperation(&o.vcdClient.Client, c.item(o.TmVdc.ID).withTmOptions(options), http.MethodDelete, nil)
}

// AddVmClassesToRegionQuota adds VM classes without replacing existing assignments (API 40.0+).
// It returns the server's first result page. Use RegionQuota.GetAllVmClasses to retrieve all assignments.
func (o *VCDClient) AddVmClassesToRegionQuota(regionQuotaId string, vmClasses *types.RegionVirtualMachineClasses, options ...TmOperationOptions) (*types.RegionVirtualMachineClasses, error) {
	c := crudConfig{
		entityLabel:    types.OpenApiEndpointTmVdcsVmClasses,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVdcsVmClasses,
		endpointParams: []string{regionQuotaId},
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return writeInnerEntity(&o.Client, c.withTmOptions(options), http.MethodPost, vmClasses)
}

// AddVmClasses adds VM classes to the receiver region quota and returns the first result page (API 40.0+).
func (o *RegionQuota) AddVmClasses(vmClasses *types.RegionVirtualMachineClasses, options ...TmOperationOptions) (*types.RegionVirtualMachineClasses, error) {
	return o.vcdClient.AddVmClassesToRegionQuota(o.TmVdc.ID, vmClasses, options...)
}

// GetAllVmClasses retrieves the region quota's VM-class assignments across all pages (API 40.0+).
func (o *RegionQuota) GetAllVmClasses(queryParameters url.Values, options ...TmOperationOptions) ([]*types.OpenApiReference, error) {
	c := crudConfig{
		entityLabel:     types.OpenApiEndpointTmVdcsVmClasses,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVdcsVmClasses,
		endpointParams:  []string{o.TmVdc.ID},
		queryParameters: queryParameters,
		requiresTm:      true,
		minimumVersion:  "40.0",
	}
	return getAllInnerEntities[types.OpenApiReference](&o.vcdClient.Client, c.withTmOptions(options))
}

// GetAllReservedVmClasses retrieves the region quota's VM-class reservations across all pages (API 40.0+).
func (o *RegionQuota) GetAllReservedVmClasses(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVdcReservedVmClass, error) {
	c := crudConfig{
		entityLabel:     "virtualDatacenters/reservedVMClasses",
		endpoint:        types.OpenApiPathVcf + "virtualDatacenters/%s/reservedVMClasses",
		endpointParams:  []string{o.TmVdc.ID},
		queryParameters: queryParameters,
		requiresTm:      true,
		minimumVersion:  "40.0",
	}
	return getAllInnerEntities[types.TmVdcReservedVmClass](&o.vcdClient.Client, c.withTmOptions(options))
}

// UpdateReservedVmClasses updates the desired VM-class reservations and returns the first result page (API 40.0+).
// Use GetAllReservedVmClasses to retrieve all reservations after the update.
func (o *RegionQuota) UpdateReservedVmClasses(config *types.TmVdcReservedVmClasses, options ...TmOperationOptions) (*types.TmVdcReservedVmClasses, error) {
	c := crudConfig{
		entityLabel:    "virtualDatacenters/reservedVMClasses",
		endpoint:       types.OpenApiPathVcf + "virtualDatacenters/%s/reservedVMClasses",
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
		minimumVersion: "40.0",
	}
	return updateInnerEntity(&o.vcdClient.Client, c.withTmOptions(options), config)
}

// GetAllNamespaces retrieves namespace summaries for the region quota across all pages (API 40.0+).
// queryParameters can specify page and pageSize.
func (o *RegionQuota) GetAllNamespaces(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmNamespaceSummary, error) {
	c := crudConfig{
		entityLabel:     "virtualDatacenters/namespaces",
		endpoint:        types.OpenApiPathVcf + "virtualDatacenters/%s/namespaces",
		endpointParams:  []string{o.TmVdc.ID},
		queryParameters: queryParameters,
		requiresTm:      true,
		minimumVersion:  "40.0",
	}
	return getAllInnerEntities[types.TmNamespaceSummary](&o.vcdClient.Client, c.withTmOptions(options))
}

// GetAllInfraPolicies retrieves the region quota's infrastructure-policy assignments across all pages (API 40.1+).
func (o *RegionQuota) GetAllInfraPolicies(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmVdcInfraPolicy, error) {
	c := crudConfig{
		entityLabel:     "virtualDatacenters/infraPolicies",
		endpoint:        types.OpenApiPathVcf + "virtualDatacenters/%s/infraPolicies",
		endpointParams:  []string{o.TmVdc.ID},
		queryParameters: queryParameters,
		requiresTm:      true,
		minimumVersion:  "40.1",
	}
	return getAllInnerEntities[types.TmVdcInfraPolicy](&o.vcdClient.Client, c.withTmOptions(options))
}

// UpdateInfraPolicies replaces the region quota's infrastructure-policy assignments (API 40.1+).
// Pass an empty Values slice to remove assignments, subject to server validation.
// The response contains the first page; use GetAllInfraPolicies to retrieve all assignments.
func (o *RegionQuota) UpdateInfraPolicies(config *types.TmVdcInfraPolicies, options ...TmOperationOptions) (*types.TmVdcInfraPolicies, error) {
	c := crudConfig{
		entityLabel:    "virtualDatacenters/infraPolicies",
		endpoint:       types.OpenApiPathVcf + "virtualDatacenters/%s/infraPolicies",
		endpointParams: []string{o.TmVdc.ID},
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	return updateInnerEntity(&o.vcdClient.Client, c.withTmOptions(options), config)
}
