package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmIpSpace = "TM IP Space"

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

func (vcdClient *VCDClient) CreateTmIpSpace(config *types.TmIpSpace) (*TmIpSpace, error) {
	c := crudConfig{
		entityLabel: labelTmIpSpace,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
	}
	outerType := TmIpSpace{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllTmIpSpaces(queryParameters url.Values) ([]*TmIpSpace, error) {
	c := crudConfig{
		entityLabel:     labelTmIpSpace,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		queryParameters: queryParameters,
	}

	outerType := TmIpSpace{vcdClient: vcdClient}
	return getAllOuterEntities[TmIpSpace, types.TmIpSpace](&vcdClient.Client, outerType, c)
}

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

func (vcdClient *VCDClient) GetTmIpSpaceById(id string) (*TmIpSpace, error) {
	c := crudConfig{
		entityLabel:    labelTmIpSpace,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		endpointParams: []string{id},
	}

	outerType := TmIpSpace{vcdClient: vcdClient}
	return getOuterEntity[TmIpSpace, types.TmIpSpace](&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmIpSpaceByNameAndOrgId(name, orgId string) (*TmIpSpace, error) {
	if name == "" || orgId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID", labelTmIpSpace)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("orgRef.id=="+orgId, queryParams)

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

func (o *TmIpSpace) Update(TmIpSpaceConfig *types.TmIpSpace) (*TmIpSpace, error) {
	c := crudConfig{
		entityLabel:    labelTmIpSpace,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		endpointParams: []string{o.TmIpSpace.ID},
	}
	outerType := TmIpSpace{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmIpSpaceConfig)
}

func (o *TmIpSpace) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmIpSpace,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaces,
		endpointParams: []string{o.TmIpSpace.ID},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
