// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TmVirtualNetworkApplianceCluster manages a Tenant Manager VirtualNetworkApplianceCluster.
type TmVirtualNetworkApplianceCluster struct {
	TmVirtualNetworkApplianceCluster *types.TmVirtualNetworkApplianceCluster
	vcdClient                        *VCDClient
	options                          []TmOperationOptions
}

// GetAllTmVirtualNetworkApplianceClusters retrieves all Virtual Network Appliance Clusters in the system.
// Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmVirtualNetworkApplianceClusters(queryParameters url.Values, options ...TmOperationOptions) ([]*TmVirtualNetworkApplianceCluster, error) {
	c := crudConfig{
		entityLabel:     "virtualNetworkApplianceClusters",
		endpoint:        types.OpenApiPathVcf + "virtualNetworkApplianceClusters",
		requiresTm:      true,
		minimumVersion:  "40.1",
		queryParameters: queryParameters,
	}
	result, err := getAllInnerEntities[types.TmVirtualNetworkApplianceCluster](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	wrapped := make([]*TmVirtualNetworkApplianceCluster, len(result))
	for i := range result {
		wrapped[i] = &TmVirtualNetworkApplianceCluster{TmVirtualNetworkApplianceCluster: result[i], vcdClient: vcdClient, options: retainTmOptions(options)}
	}
	return wrapped, nil
}

// GetTmVirtualNetworkApplianceClusterById retrieves a specific Virtual Network Appliance Cluster. Requires
// TM API 40.1+.
func (vcdClient *VCDClient) GetTmVirtualNetworkApplianceClusterById(virtualNetworkApplianceClusterId string, options ...TmOperationOptions) (*TmVirtualNetworkApplianceCluster, error) {
	c := crudConfig{
		entityLabel:    "virtualNetworkApplianceClusters",
		endpoint:       types.OpenApiPathVcf + "virtualNetworkApplianceClusters/%s",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{virtualNetworkApplianceClusterId},
	}
	result, err := getInnerEntity[types.TmVirtualNetworkApplianceCluster](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	return &TmVirtualNetworkApplianceCluster{TmVirtualNetworkApplianceCluster: result, vcdClient: vcdClient, options: retainTmOptions(options)}, nil
}

// GetAllTransportNodeStatuses retrieves status of all transport nodes associated with the Virtual Network
// Appliance Cluster. Requires TM API 40.1+.
func (o *TmVirtualNetworkApplianceCluster) GetAllTransportNodeStatuses(options ...TmOperationOptions) ([]*types.TmEdgeClusterTransportNodeStatus, error) {
	vcdClient := o.vcdClient
	options = inheritTmOptions(o.options, options)
	c := crudConfig{
		entityLabel:    "virtualNetworkApplianceClusters/transportNodesStatus",
		endpoint:       types.OpenApiPathVcf + "virtualNetworkApplianceClusters/%s/transportNodesStatus",
		requiresTm:     true,
		minimumVersion: "40.1",
		endpointParams: []string{o.TmVirtualNetworkApplianceCluster.ID},
	}
	page, err := getInnerEntity[types.TmEdgeClusterTransportNodeStatusList](&vcdClient.Client, c.withTmOptions(options))
	if err != nil {
		return nil, err
	}
	result := make([]*types.TmEdgeClusterTransportNodeStatus, len(page.Values))
	for i := range page.Values {
		result[i] = &page.Values[i]
	}
	return result, nil
}

// SyncTmVirtualNetworkApplianceClusters synchronizes the Virtual Network Appliance Clusters from NSX
// Manager to the Provider Management. Waits for task completion. Requires TM API 40.1+.
func (vcdClient *VCDClient) SyncTmVirtualNetworkApplianceClusters(options ...TmOperationOptions) error {
	c := crudConfig{
		entityLabel:    "virtualNetworkApplianceClusters/sync",
		endpoint:       types.OpenApiPathVcf + "virtualNetworkApplianceClusters/sync",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	return runOpenApiOperation(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, nil)
}

// SyncTmVirtualNetworkApplianceClustersAsync synchronizes the Virtual Network Appliance Clusters from NSX
// Manager to the Provider Management. Returns the tracking task without waiting. Requires TM API 40.1+.
func (vcdClient *VCDClient) SyncTmVirtualNetworkApplianceClustersAsync(options ...TmOperationOptions) (*Task, error) {
	c := crudConfig{
		entityLabel:    "virtualNetworkApplianceClusters/sync",
		endpoint:       types.OpenApiPathVcf + "virtualNetworkApplianceClusters/sync",
		requiresTm:     true,
		minimumVersion: "40.1",
	}
	return runOpenApiOperationAsync(&vcdClient.Client, c.withTmOptions(options), http.MethodPost, nil)
}
