// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"net/http"
	"net/url"
)

const labelTmVirtualNetworkApplianceCluster = "TM Virtual Network Appliance Cluster"

// TmVirtualNetworkApplianceCluster manages a Virtual Network Appliance Cluster.
type TmVirtualNetworkApplianceCluster struct {
	TmVirtualNetworkApplianceCluster *types.TmVirtualNetworkApplianceCluster
	vcdClient                        *VCDClient
	options                          []TmOperationOptions
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmVirtualNetworkApplianceCluster) wrap(inner *types.TmVirtualNetworkApplianceCluster) *TmVirtualNetworkApplianceCluster {
	g.TmVirtualNetworkApplianceCluster = inner
	return &g
}

// tmVirtualNetworkApplianceClusterConfig returns the CRUD configuration for Virtual Network Appliance Cluster operations. Optional params are appended to the endpoint.
func tmVirtualNetworkApplianceClusterConfig(options []TmOperationOptions, params ...string) crudConfig {
	return crudConfig{
		entityLabel:    labelTmVirtualNetworkApplianceCluster,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmVirtualNetworkApplianceClusters,
		endpointParams: params,
		requiresTm:     true,
		minimumVersion: "40.1",
	}.withTmOptions(options)
}

// GetAllTmVirtualNetworkApplianceClusters retrieves all Virtual Network Appliance Clusters. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetAllTmVirtualNetworkApplianceClusters(queryParameters url.Values, options ...TmOperationOptions) ([]*TmVirtualNetworkApplianceCluster, error) {
	c := tmVirtualNetworkApplianceClusterConfig(options)
	c.queryParameters = queryParameters
	outerType := TmVirtualNetworkApplianceCluster{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmVirtualNetworkApplianceClusterByName retrieves a Virtual Network Appliance Cluster by name. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmVirtualNetworkApplianceClusterByName(name string, options ...TmOperationOptions) (*TmVirtualNetworkApplianceCluster, error) {
	entities, err := vcdClient.GetAllTmVirtualNetworkApplianceClusters(nil, options...)
	if err != nil {
		return nil, err
	}
	return tmEntityByName(labelTmVirtualNetworkApplianceCluster, name, entities, func(e *TmVirtualNetworkApplianceCluster) string { return e.TmVirtualNetworkApplianceCluster.Name })
}

// GetTmVirtualNetworkApplianceClusterById retrieves a Virtual Network Appliance Cluster by ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmVirtualNetworkApplianceClusterById(id string, options ...TmOperationOptions) (*TmVirtualNetworkApplianceCluster, error) {
	outerType := TmVirtualNetworkApplianceCluster{vcdClient: vcdClient, options: retainTmOptions(options)}
	return getOuterEntity(&vcdClient.Client, outerType, tmVirtualNetworkApplianceClusterConfig(options, id))
}

// GetTmVirtualNetworkApplianceClusterByNameOrId retrieves a Virtual Network Appliance Cluster by name or ID. Requires TM API 40.1+.
func (vcdClient *VCDClient) GetTmVirtualNetworkApplianceClusterByNameOrId(identifier string, options ...TmOperationOptions) (*TmVirtualNetworkApplianceCluster, error) {
	return tmEntityByNameOrId(identifier,
		func(id string) (*TmVirtualNetworkApplianceCluster, error) {
			return vcdClient.GetTmVirtualNetworkApplianceClusterById(id, options...)
		},
		func(name string) (*TmVirtualNetworkApplianceCluster, error) {
			return vcdClient.GetTmVirtualNetworkApplianceClusterByName(name, options...)
		})
}

// GetAllTransportNodeStatuses retrieves the status of all transport nodes associated with the Virtual Network
// Appliance Cluster. Requires TM API 40.1+.
func (o *TmVirtualNetworkApplianceCluster) GetAllTransportNodeStatuses(options ...TmOperationOptions) ([]*types.TmEdgeClusterTransportNodeStatus, error) {
	options = inheritTmOptions(o.options, options)
	c := tmCrudConfig("TM Virtual Network Appliance Cluster transport node status", types.OpenApiEndpointTmVirtualNetworkApplianceClusterTransportNodeStatus, "40.1", options, o.TmVirtualNetworkApplianceCluster.ID)
	page, err := getInnerEntity[types.TmEdgeClusterTransportNodeStatusList](&o.vcdClient.Client, c)
	if err != nil {
		return nil, err
	}
	result := make([]*types.TmEdgeClusterTransportNodeStatus, len(page.Values))
	for i := range page.Values {
		result[i] = &page.Values[i]
	}
	return result, nil
}

// SyncTmVirtualNetworkApplianceClusters synchronizes the Virtual Network Appliance Clusters from NSX Manager to the Provider Management. Waits for task completion. Requires TM API 40.1+.
func (vcdClient *VCDClient) SyncTmVirtualNetworkApplianceClusters(options ...TmOperationOptions) error {
	return runOpenApiOperation(&vcdClient.Client, tmCrudConfig("TM Virtual Network Appliance Cluster sync", types.OpenApiEndpointTmVirtualNetworkApplianceClustersSync, "40.1", options), http.MethodPost, nil)
}

// SyncTmVirtualNetworkApplianceClustersAsync synchronizes the Virtual Network Appliance Clusters from NSX Manager to the Provider Management. Returns the tracking task without waiting. Requires TM API 40.1+.
func (vcdClient *VCDClient) SyncTmVirtualNetworkApplianceClustersAsync(options ...TmOperationOptions) (*Task, error) {
	return runOpenApiOperationAsync(&vcdClient.Client, tmCrudConfig("TM Virtual Network Appliance Cluster sync", types.OpenApiEndpointTmVirtualNetworkApplianceClustersSync, "40.1", options), http.MethodPost, nil)
}
