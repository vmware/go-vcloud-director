package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (egw *NsxtEdgeGateway) GetL2VpnTunnels(queryParameters url.Values) ([]*types.NsxtL2VpnTunnel, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot get L2 VPN tunnels for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	responses := []*types.NsxtL2VpnTunnel{{}}
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, responses, nil)
	if err != nil {
		return nil, err
	}

	for _, response := range responses {
		fmt.Printf("%+v", response)
	}

	return nil, nil
}
