package govcd

import (
	"fmt"
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
	}
	outerType := Region{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllRegions retrieves all Regions with an optional query filter
func (vcdClient *VCDClient) GetAllRegions(queryParameters url.Values) ([]*Region, error) {
	c := crudConfig{
		entityLabel:     labelRegion,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointRegions,
		queryParameters: queryParameters,
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
	}
	return deleteEntityById(&r.vcdClient.Client, c)
}
