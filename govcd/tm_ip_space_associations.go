package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmIpSpaceAssociation = "TM IP Space Association"

type TmIpSpaceAssociation struct {
	TmIpSpaceAssociation *types.TmIpSpaceAssociation
	vcdClient            *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmIpSpaceAssociation) wrap(inner *types.TmIpSpaceAssociation) *TmIpSpaceAssociation {
	g.TmIpSpaceAssociation = inner
	return &g
}

func (vcdClient *VCDClient) CreateTmIpSpaceAssociation(config *types.TmIpSpaceAssociation) (*TmIpSpaceAssociation, error) {
	c := crudConfig{
		entityLabel: labelTmIpSpaceAssociation,
		endpoint:    types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
	}
	outerType := TmIpSpaceAssociation{vcdClient: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

func (vcdClient *VCDClient) GetAllTmIpSpaceAssociations(queryParameters url.Values) ([]*TmIpSpaceAssociation, error) {
	c := crudConfig{
		entityLabel:     labelTmIpSpaceAssociation,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
		queryParameters: queryParameters,
	}

	outerType := TmIpSpaceAssociation{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetTmIpSpaceAssociationById(id string) (*TmIpSpaceAssociation, error) {
	c := crudConfig{
		entityLabel:    labelTmIpSpaceAssociation,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
		endpointParams: []string{id},
	}

	outerType := TmIpSpaceAssociation{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

func (vcdClient *VCDClient) GetAllTmIpSpaceAssociationsByProviderGatewayId(providerGatewayId string) ([]*TmProviderGateway, error) {
	if providerGatewayId == "" {
		return nil, fmt.Errorf("%s lookup requires %s ID", labelTmIpSpaceAssociation, labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams = queryParameterFilterAnd("providerGatewayRef.id=="+providerGatewayId, queryParams)

	return vcdClient.GetAllTmProviderGateways(queryParams)
}

func (vcdClient *VCDClient) GetAllTmIpSpaceAssociationsByIpSpaceId(ipSpaceId string) ([]*TmProviderGateway, error) {
	if ipSpaceId == "" {
		return nil, fmt.Errorf("%s lookup requires %s ID", labelTmIpSpaceAssociation, labelTmProviderGateway)
	}

	queryParams := url.Values{}
	queryParams = queryParameterFilterAnd("ipSpaceRef.id=="+ipSpaceId, queryParams)

	return vcdClient.GetAllTmProviderGateways(queryParams)
}

func (o *TmIpSpaceAssociation) Delete() error {
	c := crudConfig{
		entityLabel:    labelTmIpSpaceAssociation,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmIpSpaceAssociations,
		endpointParams: []string{o.TmIpSpaceAssociation.ID},
	}
	return deleteEntityById(&o.vcdClient.Client, c)
}
