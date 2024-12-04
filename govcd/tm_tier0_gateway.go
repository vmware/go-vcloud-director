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

func (vcdClient *VCDClient) GetAllTmTier0GatewaysWithContext(contextEntity string, listImported bool) ([]*TmTier0Gateway, error) {
	if contextEntity == "" {
		return nil, fmt.Errorf("empty region provided")
	}
	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", contextEntity), queryParams)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("alreadyImported==%t", listImported), queryParams)

	return vcdClient.getAllTmTier0Gateways(queryParams)
}

func (vcdClient *VCDClient) GetTmTier0GatewayWithContextByName(displayName, contextEntity string, listImported bool) (*TmTier0Gateway, error) {
	if contextEntity == "" {
		return nil, fmt.Errorf("empty region provided")
	}
	if displayName == "" {
		return nil, fmt.Errorf("empty name provided")
	}

	queryParams := copyOrNewUrlValues(nil)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("_context==%s", contextEntity), queryParams)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("alreadyImported==%t", listImported), queryParams)

	allT0s, err := vcdClient.getAllTmTier0Gateways(queryParams)
	if err != nil {
		return nil, err
	}

	var foundValue *TmTier0Gateway
	for _, v := range allT0s {
		if v.TmTier0Gateway.DisplayName == displayName {
			if foundValue != nil {
				return nil, fmt.Errorf("found more than one %s by Name '%s'", labelTmTier0Gateway, displayName)
			}
			foundValue = v
		}
	}

	if foundValue == nil {
		return nil, fmt.Errorf("no %s found by Name '%s'", labelTmTier0Gateway, displayName)
	}

	return foundValue, nil
}

// getAllTmTier0Gateways is kept private as one receives 0 entries when querying without filters,
// but it is useful construct in higher level functions
func (vcdClient *VCDClient) getAllTmTier0Gateways(queryParameters url.Values) ([]*TmTier0Gateway, error) {
	c := crudConfig{
		entityLabel:     labelTmTier0Gateway,
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTier0Routers,
		queryParameters: queryParameters,
	}

	outerType := TmTier0Gateway{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}
