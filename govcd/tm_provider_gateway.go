package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmProviderGateway = "TM Provider Gateway"

// TmProviderGateway manages Provider Gateway creation and configuration.
//
// NOTE. While creation of Provider Gateway requires at least one IP Space (`TmIpSpace`) reference,
// they are not being returned by API after creation. One must use `TmIpSpaceAssociation` for
// managing IP Space associations with Provider gateways after the initial creation is done.
type TmProviderGateway struct {
	TmProviderGateway *types.TmProviderGateway
	vcdClient         *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmProviderGateway) wrap(inner *types.TmProviderGateway) *TmProviderGateway {
	g.TmProviderGateway = inner
	return &g
}

func (vcdClient *VCDClient) CreateTmProviderGateway(config *types.TmProviderGateway) (*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel: labelTmProviderGateway,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
	}
	outerType := TmProviderGateway{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllTmProviderGateways(queryParameters url.Values) ([]*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel:     labelTmProviderGateway,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		queryParameters: queryParameters,
	}

	outerType := TmProviderGateway{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmProviderGatewayByName(name string) (*TmProviderGateway, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmProviderGateways(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmProviderGatewayById(singleEntity.TmProviderGateway.ID)
}

func (vcdClient *VCDClient) GetTmProviderGatewayById(id string) (*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel:    labelTmProviderGateway,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		endpointParams: []string{id},
	}

	outerType := TmProviderGateway{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmProviderGatewayByNameAndOrgId(name, orgId string) (*TmProviderGateway, error) {
	if name == "" || orgId == "" {
		return nil, fmt.Errorf("%s lookup requires name and Org ID", labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd("orgRef.id=="+orgId, queryParams)

	filteredEntities, err := vcdClient.GetAllTmProviderGateways(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmProviderGatewayById(singleEntity.TmProviderGateway.ID)
}

func (o *TmProviderGateway) Update(TmProviderGatewayConfig *types.TmProviderGateway) (*TmProviderGateway, error) {
	c := crudConfig{
		entityLabel:    labelTmProviderGateway,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		endpointParams: []string{o.TmProviderGateway.ID},
	}
	outerType := TmProviderGateway{vcdClient: o.vcdClient}
	return updateOuterEntity(&o.vcdClient.Client, outerType, c, TmProviderGatewayConfig)
}

func (o *TmProviderGateway) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmProviderGateway,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways,
		endpointParams: []string{o.TmProviderGateway.ID},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
