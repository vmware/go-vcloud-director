package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmTier0Gateway = "TM Tier0 Gateway"

type TmTier0Gateway struct {
	TmTier0Gateway *types.TmTier0Gateway
	vcdClient      *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmTier0Gateway) wrap(inner *types.TmTier0Gateway) *TmTier0Gateway {
	g.TmTier0Gateway = inner
	return &g
}

func (vcdClient *VCDClient) CreateTmTier0Gateway(config *types.TmTier0Gateway) (*TmTier0Gateway, error) {
	c := crudConfig{
		entityLabel: labelTmTier0Gateway,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmTier0Gateways,
	}
	outerType := TmTier0Gateway{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllTmTier0Gateways(queryParameters url.Values) ([]*TmTier0Gateway, error) {
	c := crudConfig{
		entityLabel:     labelTmTier0Gateway,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmTier0Gateways,
		queryParameters: queryParameters,
	}

	outerType := TmTier0Gateway{vcdClient: vcdClient}
	return getAllOuterEntities[TmTier0Gateway, types.TmTier0Gateway](&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmTier0GatewayByName(name string) (*TmTier0Gateway, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmTier0Gateway)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmTier0Gateways(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmTier0GatewayById(singleEntity.TmTier0Gateway.ID)
}

func (vcdClient *VCDClient) GetTmTier0GatewayById(id string) (*TmTier0Gateway, error) {
	c := crudConfig{
		entityLabel:    labelTmTier0Gateway,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmTier0Gateways,
		endpointParams: []string{id},
	}

	outerType := TmTier0Gateway{vcdClient: vcdClient}
	return getOuterEntity[TmTier0Gateway, types.TmTier0Gateway](&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmTier0GatewayByNameAndOrgId(name, orgId string) (*TmTier0Gateway, error) {
	if name == "" || orgId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID", labelTmTier0Gateway)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("orgRef.id=="+orgId, queryParams)

	filteredEntities, err := vcdClient.GetAllTmTier0Gateways(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmTier0GatewayById(singleEntity.TmTier0Gateway.ID)
}
