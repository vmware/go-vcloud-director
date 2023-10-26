package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtL2VpnTunnel extends an organization VDC by enabling virtual machines to
// maintain their network connectivity across geographical boundaries while keeping
// the same IP address. The connection is secured with a route-based IPSec tunnel between the two sides of the tunnel.
// The L2 VPN service can be configured on an NSX-T edge gateway in a VMware Cloud Director environment
// to create a L2 VPN tunnel. Virtual machines remain on the same subnet, which extends
// the organization VDC by stretching its network. This way, an edge gateway at one site can provide
// all services to virtual machines on the other site.
type NsxtL2VpnTunnel struct {
	NsxtL2VpnTunnel *types.NsxtL2VpnTunnel
	client          *Client
	// edgeGatewayId is stored for usage in NsxtFirewall receiver functions
	edgeGatewayId string
}

// CreateL2VpnTunnel creates a L2 VPN Tunnel on the provided NSX-T Edge Gateway and returns
// the tunnel
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

// Refresh updates the provided NsxtL2VpnTunnel and returns an error if it failed
func (l2Vpn *NsxtL2VpnTunnel) Refresh() error {
	client := l2Vpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, l2Vpn.edgeGatewayId), l2Vpn.NsxtL2VpnTunnel.ID)
	if err != nil {
		return err
	}

	refreshedTunnel := &types.NsxtL2VpnTunnel{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &refreshedTunnel, nil)
	if err != nil {
		return err
	}
	l2Vpn.NsxtL2VpnTunnel = refreshedTunnel

	return nil
}

// GetAllL2VpnTunnels fetches all L2 VPN tunnels that are created on the Edge Gateway.
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

	// Wrap all typeResponses into NsxtL2VpnTunnel types with client
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

// GetL2VpnTunnelByName gets the L2 VPN Tunnel by name
func (egw *NsxtEdgeGateway) GetL2VpnTunnelByName(name string) (*NsxtL2VpnTunnel, error) {
	results, err := egw.GetAllL2VpnTunnels(nil)
	if err != nil {
		return nil, err
	}

	foundTunnels := make([]*NsxtL2VpnTunnel, 0)
	for _, tunnel := range results {
		if tunnel.NsxtL2VpnTunnel.Name == name {
			foundTunnels = append(foundTunnels, tunnel)
		}
	}

	return oneOrError("name", name, foundTunnels)
}

// GetL2VpnTunnelById gets the L2 VPN Tunnel by its ID
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

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, egw.EdgeGateway.ID), id)
	if err != nil {
		return nil, err
	}

	tunnel := &NsxtL2VpnTunnel{
		client:        egw.client,
		edgeGatewayId: egw.EdgeGateway.ID,
	}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &tunnel.NsxtL2VpnTunnel, nil)
	if err != nil {
		return nil, err
	}

	return tunnel, nil
}

// Statistics retrieves connection statistics for a given L2 VPN Tunnel configured on an Edge Gateway.
func (l2Vpn *NsxtL2VpnTunnel) Statistics() (*types.EdgeL2VpnTunnelStatistics, error) {
	client := l2Vpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnelStatistics
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, l2Vpn.edgeGatewayId, l2Vpn.NsxtL2VpnTunnel.ID))
	if err != nil {
		return nil, err
	}

	statistics := &types.EdgeL2VpnTunnelStatistics{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &statistics, nil)
	if err != nil {
		return nil, err
	}

	return statistics, nil
}

// Status retrieves status of a given L2 VPN Tunnel.
func (l2Vpn *NsxtL2VpnTunnel) Status() (*types.EdgeL2VpnTunnelStatus, error) {
	client := l2Vpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnelStatus
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, l2Vpn.edgeGatewayId, l2Vpn.NsxtL2VpnTunnel.ID))
	if err != nil {
		return nil, err
	}

	status := &types.EdgeL2VpnTunnelStatus{}
	err = client.OpenApiGetItem(apiVersion, urlRef, nil, &status, nil)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// Update updates the L2 VPN tunnel with the provided parameters as the argument
func (l2Vpn *NsxtL2VpnTunnel) Update(tunnelParams *types.NsxtL2VpnTunnel) (*NsxtL2VpnTunnel, error) {
	if l2Vpn.NsxtL2VpnTunnel.SessionMode != tunnelParams.SessionMode {
		return nil, fmt.Errorf("error updating the L2 VPN Tunnel: session mode can't be changed after creation")
	}

	if tunnelParams.SessionMode == "CLIENT" && !tunnelParams.Enabled {
		// There is a known bug up to 10.5.0, the CLIENT sessions can't be
		// disabled and can result in unexpected behaviour for the following
		// operations
		if l2Vpn.client.APIVCDMaxVersionIs("<= 38.0") {
			return nil, fmt.Errorf("client sessions can't be disabled on VCD versions up to 10.5.0")
		}
	}

	client := l2Vpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, l2Vpn.edgeGatewayId), l2Vpn.NsxtL2VpnTunnel.ID)
	if err != nil {
		return nil, err
	}

	tunnelParams.Version.Version = l2Vpn.NsxtL2VpnTunnel.Version.Version

	newTunnel := &NsxtL2VpnTunnel{
		client:        l2Vpn.client,
		edgeGatewayId: l2Vpn.edgeGatewayId,
	}
	err = client.OpenApiPutItem(apiVersion, urlRef, nil, tunnelParams, &newTunnel.NsxtL2VpnTunnel, nil)
	if err != nil {
		return nil, err
	}

	return newTunnel, nil
}

// Delete deletes the L2 VPN Tunnel
// On versions up to 10.5.0 (as of writing) there is a bug with deleting
// CLIENT tunnels. If there are any networks attached to the tunnel, the
// DELETE call will fail the amount of times the resource was updated,
// so the best choice is to remove the networks and then call Delete(), or
// call Delete() in a loop until it's successful.
func (l2Vpn *NsxtL2VpnTunnel) Delete() error {
	client := l2Vpn.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointEdgeGatewayL2VpnTunnel
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, l2Vpn.edgeGatewayId), l2Vpn.NsxtL2VpnTunnel.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
