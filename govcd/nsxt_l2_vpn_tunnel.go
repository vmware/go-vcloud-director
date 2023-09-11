package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtL2VpnTunnel extend your organization VDC by enabling virtual machines to
// maintain their network connectivity across geographical boundaries while keeping
// the same IP address. The connection is secured with a route-based IPSec tunnel between the two sides of the tunnel.
// You can configure the L2 VPN service on an NSX-T edge gateway in your VMware Cloud Director environment
// and create a L2 VPN tunnel. Virtual machines remain on the same subnet, which enables you to extend
// your organization VDC by stretching its network. This way, an edge gateway at one site can provide
// all services to virtual machines on the other site.
type NsxtL2VpnTunnel struct {
	NsxtL2VpnTunnel *types.NsxtL2VpnTunnel
	client          *Client
	// edgeGatewayId is stored for usage in NsxtFirewall receiver functions
	edgeGatewayId string
}

func (egw *NsxtEdgeGateway) CreateL2VpnTunnel(tunnel *types.NsxtL2VpnTunnel) (*NsxtL2VpnTunnel, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot create L2 VPN tunnel for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	// When creating a L2 VPN tunnel, its ID is stored in the creation task Details section,
	// so we need to fetch the newly created tunnel manually
	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, tunnel)
	if err != nil {
		return nil, fmt.Errorf("error creating L2 VPN tunnel: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error waiting for L2 VPN tunnel to be created: %s", err)
	}

	newTunnel, err := egw.GetL2VpnTunnelById(task.Task.Details)
	if err != nil {
		return nil, fmt.Errorf("error getting L2 VPN tunnel with id %s: %s", task.Task.Details, err)
	}

	return newTunnel, nil
}

// GetAllL2VpnTunnels fetches all L2 VPN tunnels that are created on the edge gateway.
func (egw *NsxtEdgeGateway) GetAllL2VpnTunnels(queryParameters url.Values) ([]*NsxtL2VpnTunnel, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot get L2 VPN tunnels for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtL2VpnTunnel{{}}
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID))
	if err != nil {
		return nil, err
	}

	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into IpSpaceOrgAssignment types with client
	results := make([]*NsxtL2VpnTunnel, len(typeResponses))
	for sliceIndex := range typeResponses {
		results[sliceIndex] = &NsxtL2VpnTunnel{
			NsxtL2VpnTunnel: typeResponses[sliceIndex],
			edgeGatewayId:   egw.EdgeGateway.ID,
			client:          egw.client,
		}
	}

	return results, nil
}

func (egw *NsxtEdgeGateway) GetL2VpnTunnelById(id string) (*NsxtL2VpnTunnel, error) {
	if egw.EdgeGateway == nil || egw.client == nil || egw.EdgeGateway.ID == "" {
		return nil, fmt.Errorf("cannot get L2 VPN tunnel for NSX-T Edge Gateway without ID")
	}

	client := egw.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	tunnelResponse := &types.NsxtL2VpnTunnel{}
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), id)
	if err != nil {
		return nil, err
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &tunnelResponse, nil)
	if err != nil {
		return nil, err
	}

	return &NsxtL2VpnTunnel{
		NsxtL2VpnTunnel: tunnelResponse,
		edgeGatewayId:   egw.EdgeGateway.ID,
		client:          egw.client,
	}, nil
}
