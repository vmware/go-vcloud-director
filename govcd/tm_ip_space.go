package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmIpSpace = "TM IP Space"

// TmIpSpace provides configuration of mainly the external IP Prefixes that specifies
// the accessible external networks from the data center
type TmIpSpace struct {
	TmIpSpace *types.TmIpSpace
	vcdClient *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmIpSpace) wrap(inner *types.TmIpSpace) *TmIpSpace {
	g.TmIpSpace = inner
	return &g
}

// CreateTmIpSpace with a given configuration
func (vcdClient *VCDClient) CreateTmIpSpace(config *types.TmIpSpace) (*TmIpSpace, error) {
	c := crudConfig{
		entityLabel: labelTmIpSpace,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
	}
	outerType := TmIpSpace{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// GetAllTmIpSpaces fetches all TM IP Spaces with an optional query filter
func (vcdClient *VCDClient) GetAllTmIpSpaces(queryParameters url.Values) ([]*TmIpSpace, error) {
	c := crudConfig{
		entityLabel:     labelTmIpSpace,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		queryParameters: queryParameters,
	}

	outerType := TmIpSpace{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmIpSpaceByName retrieves TM IP Spaces with a given name
func (vcdClient *VCDClient) GetTmIpSpaceByName(name string) (*TmIpSpace, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmIpSpace)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmIpSpaces(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmIpSpaceById(singleEntity.TmIpSpace.ID)
}

// GetTmIpSpaceById retrieves an exact IP Space with a given ID
func (vcdClient *VCDClient) GetTmIpSpaceById(id string) (*TmIpSpace, error) {
	c := crudConfig{
		entityLabel:    labelTmIpSpace,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		endpointParams: []string{id},
	}

	outerType := TmIpSpace{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// GetTmIpSpaceByNameAndRegionId retrieves TM IP Spaces with a given name in a provided Region
func (vcdClient *VCDClient) GetTmIpSpaceByNameAndRegionId(name, regionId string) (*TmIpSpace, error) {
	if name == "" || regionId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Region ID", labelTmIpSpace)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("regionRef.id=="+regionId, queryParams)

	filteredEntities, err := vcdClient.GetAllTmIpSpaces(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmIpSpaceById(singleEntity.TmIpSpace.ID)
}

// Update TM IP Space
func (o *TmIpSpace) Update(TmIpSpaceConfig *types.TmIpSpace) (*TmIpSpace, error) {
	c := crudConfig{
		entityLabel:    labelTmIpSpace,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		endpointParams: []string{o.TmIpSpace.ID},
	}
	outerType := TmIpSpace{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmIpSpaceConfig)
}

// Delete TM IP Space
func (o *TmIpSpace) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmIpSpace,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		endpointParams: []string{o.TmIpSpace.ID},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
