package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// This is a template of how a "standard" outer entity implementation can be done when using generic
// functions. It might not cover all scenarios, but is a skeleton for quicker bootstraping of a new
// entity.
//
// "Search and replace the following entries"
//
// TmProviderGateway - outer type (e.g. IpSpace2)
// This should be a non existing new type to create in 'govcd' package
//
// types.TmProviderGateway - inner type (e.g. types.IpSpace)
// This should be an already existing inner type in `types` package
//
// TM Provider Gateway - constant name for entity label (the lower case prefix 'label' prefix is hardcoded)
// The 'label' prefix is hardcoded in the example so that we have autocompletion working for all labelXXXX. (e.g. IpSpace2)
//
// TM Provider Gateway - text for entity label (e.g. Ip Space 2)
// This will be the entity label (used for logging purposes in generic functions)
//
// types.OpenApiPathVcf + types.OpenApiEndpointTmProviderGateways (e.g. types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointIpSpaces)
// An OpenAPI endpoint that is defined in `endpointMinApiVersions` map and in `constants.go`
// NOTE. While this example REPLACES ALL ENDPOINTS to be THE SAME, in reality they can be DIFFERENT

const labelTmProviderGateway = "TM Provider Gateway"

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
	return getOuterEntity[TmProviderGateway, types.TmProviderGateway](&vcdClient.Client, outerType, c)
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
