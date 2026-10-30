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

const labelRegion = "Region"

type Region struct {
	Region    *types.Region
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (r Region) wrap(inner *types.Region) *Region {
	r.Region = inner
	return &r
}

// CreateRegion creates a new region
func (vcdClient *VCDClient) CreateRegion(config *types.Region) (*Region, error) {
	c := crudConfig{
		entityLabel: labelRegion,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		requiresTm:  true,
	}
	outerType := Region{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateRegionAsync creates a new region and returns its tracking task
func (vcdClient *VCDClient) CreateRegionAsync(config *types.Region) (*Task, error) {
	c := crudConfig{
		entityLabel: labelRegion,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		requiresTm:  true,
	}
	return createInnerEntityAsync(&vcdClient.Client, c, config)
}

// GetAllRegions retrieves all Regions with an optional query filter
func (vcdClient *VCDClient) GetAllRegions(queryParameters url.Values) ([]*Region, error) {
	c := crudConfig{
		entityLabel:     labelRegion,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := Region{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetRegionByName retrieves a region by name
func (vcdClient *VCDClient) GetRegionByName(name string) (*Region, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelRegion)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllRegions(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetRegionById(singleEntity.Region.ID)
}

// GetRegionById retrieves a region by ID
func (vcdClient *VCDClient) GetRegionById(id string) (*Region, error) {
	c := crudConfig{
		entityLabel:    labelRegion,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := Region{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update Region with new configuration
func (r *Region) Update(RegionConfig *types.Region) (*Region, error) {
	c := crudConfig{
		entityLabel:    labelRegion,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		endpointParams: []string{r.Region.ID},
		requiresTm:     true,
	}
	outerType := Region{vcdClient: r.vcdClient}
	return updateOuterEntity(&r.vcdClient.Client, outerType, c, RegionConfig)
}

// Delete Region
func (r *Region) Delete() error {
	c := crudConfig{
		entityLabel:    labelRegion,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		endpointParams: []string{r.Region.ID},
		requiresTm:     true,
	}
	return deleteEntityById(&r.vcdClient.Client, c)
}

// GetAllVmClasses retrieves all VM Classes within a particular Region
func (r *Region) GetAllVmClasses(queryParameters url.Values) ([]*types.RegionVirtualMachineClass, error) {
	c := crudConfig{
		entityLabel:     labelRegion,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmVmClasses,
		queryParameters: queryParameterFilterAnd("region.id=="+r.Region.ID, queryParameters),
		requiresTm:      true,
	}
	return getAllInnerEntities[types.RegionVirtualMachineClass](&r.vcdClient.Client, c)
}

// DeleteWithForce deletes the Region and waits for completion (API 40.0+).
// Setting force to true requests forceful deletion even if the backing resource cannot be removed normally.
func (o *Region) DeleteWithForce(force bool, options ...TmOperationOptions) error {
	c := crudConfig{
		entityLabel:     types.OpenApiEndpointRegions,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		queryParameters: url.Values{"force": {fmt.Sprintf("%t", force)}},
		requiresTm:      true,
		minimumVersion:  "40.0",
	}
	return runOpenApiOperation(&o.vcdClient.Client, c.item(o.Region.ID).withTmOptions(options), http.MethodDelete, nil)
}

// GetAllNamespaceCompatibleZoneGroups retrieves groups of zones compatible for namespace placement (API 40.1+).
// queryParameters can specify page and pageSize; all result pages are retrieved.
func (o *Region) GetAllNamespaceCompatibleZoneGroups(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmCompatibleZoneGroup, error) {
	c := crudConfig{
		entityLabel:     "regions/namespaceCompatibleZones",
		endpoint:        types.OpenApiPathVcf + "regions/%s/namespaceCompatibleZones",
		endpointParams:  []string{o.Region.ID},
		queryParameters: queryParameters,
		requiresTm:      true,
		minimumVersion:  "40.1",
	}
	return getAllInnerEntities[types.TmCompatibleZoneGroup](&o.vcdClient.Client, c.withTmOptions(options))
}

// GetAllCapabilities retrieves the region's capabilities across all result pages (API 40.1+).
// queryParameters can specify page and pageSize.
func (o *Region) GetAllCapabilities(queryParameters url.Values, options ...TmOperationOptions) ([]*types.TmRegionCapability, error) {
	c := crudConfig{
		entityLabel:     "regions/capabilities",
		endpoint:        types.OpenApiPathVcf + "regions/%s/capabilities",
		endpointParams:  []string{o.Region.ID},
		queryParameters: queryParameters,
		requiresTm:      true,
		minimumVersion:  "40.1",
	}
	return getAllInnerEntities[types.TmRegionCapability](&o.vcdClient.Client, c.withTmOptions(options))
}
