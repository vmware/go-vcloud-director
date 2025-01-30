package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmEdgeCluster = "TM Edge Cluster"
const labelTmEdgeClusterSync = "TM Edge Cluster Sync"
const labelTmEdgeClusterTransportNodeStatus = "TM Edge Cluster Transport Node Status"

// TmEdgeCluster manages read operations for NSX-T Edge Clusters and their QoS settings
type TmEdgeCluster struct {
	TmEdgeCluster *types.TmEdgeCluster
	vcdClient     *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (g TmEdgeCluster) wrap(inner *types.TmEdgeCluster) *TmEdgeCluster {
	g.TmEdgeCluster = inner
	return &g
}

// GetAllTmEdgeClusters retrieves all TM Edge Clusters with an optional filter
func (vcdClient *VCDClient) GetAllTmEdgeClusters(queryParameters url.Values) ([]*TmEdgeCluster, error) {
	c := crudConfig{
		entityLabel:     labelTmEdgeCluster,
		endpoint:        types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClusters,
		queryParameters: queryParameters,
		requiresTm:      true,
	}

	outerType := TmEdgeCluster{vcdClient: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetTmEdgeClusterByName retrieves TM Edge Cluster by Name
func (vcdClient *VCDClient) GetTmEdgeClusterByName(name string) (*TmEdgeCluster, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmEdgeCluster)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	filteredEntities, err := vcdClient.GetAllTmEdgeClusters(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmEdgeClusterById(singleEntity.TmEdgeCluster.ID)
}

// GetTmEdgeClusterByNameAndRegionId retrieves TM Edge Cluster by Name and a Region ID
func (vcdClient *VCDClient) GetTmEdgeClusterByNameAndRegionId(name, regionId string) (*TmEdgeCluster, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelTmEdgeCluster)
	}

	if regionId == "" {
		return nil, fmt.Errorf("%s lookup requires %s ID", labelTmEdgeCluster, labelRegion)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)
	queryParams = queryParameterFilterAnd(fmt.Sprintf("regionRef.id==%s", regionId), queryParams)

	filteredEntities, err := vcdClient.GetAllTmEdgeClusters(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, filteredEntities)
	if err != nil {
		return nil, err
	}

	return vcdClient.GetTmEdgeClusterById(singleEntity.TmEdgeCluster.ID)
}

// GetTmEdgeClusterById retrieves TM Edge Cluster by ID
func (vcdClient *VCDClient) GetTmEdgeClusterById(id string) (*TmEdgeCluster, error) {
	c := crudConfig{
		entityLabel:    labelTmEdgeCluster,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClusters,
		endpointParams: []string{id},
		requiresTm:     true,
	}

	outerType := TmEdgeCluster{vcdClient: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// TmSyncEdgeClusters triggers a global sync operation that re-reads available Edge Clusters in all
// configured NSX-T Managers
func (vcdClient *VCDClient) TmSyncEdgeClusters() error {
	client := vcdClient.Client
	endpoint := types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClustersSync

	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return err
	}

	task, err := client.OpenApiPostItemAsync(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error performing %s: %s", labelTmEdgeClusterSync, err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error performing %s: %s", labelTmEdgeClusterSync, err)
	}

	return nil
}

// Update TM Edge Cluster with a given config
// Note. Only `DefaultQosConfig` structure is updatable
func (e *TmEdgeCluster) Update(TmEdgeClusterConfig *types.TmEdgeCluster) (*TmEdgeCluster, error) {
	c := crudConfig{
		entityLabel:    labelTmEdgeCluster,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClusters,
		endpointParams: []string{e.TmEdgeCluster.ID},
		requiresTm:     true,
	}
	outerType := TmEdgeCluster{vcdClient: e.vcdClient}
	return updateOuterEntity(&e.vcdClient.Client, outerType, c, TmEdgeClusterConfig)
}

// Delete removes the QoS configuration for a given TM Edge Cluster as the Edge Cluster itself is
// not removable
func (e *TmEdgeCluster) Delete() error {
	if e.TmEdgeCluster == nil {
		return fmt.Errorf("nil %s", labelTmEdgeCluster)
	}
	e.TmEdgeCluster.DefaultQosConfig.EgressProfile = &types.TmEdgeClusterQosProfile{
		Type:                   "DEFAULT",
		CommittedBandwidthMbps: -1,
		BurstSizeBytes:         -1,
	}
	e.TmEdgeCluster.DefaultQosConfig.IngressProfile = &types.TmEdgeClusterQosProfile{
		Type:                   "DEFAULT",
		CommittedBandwidthMbps: -1,
		BurstSizeBytes:         -1,
	}

	_, err := e.Update(e.TmEdgeCluster)
	if err != nil {
		return fmt.Errorf("error removing QoS configuration for  %s: %s", labelTmEdgeCluster, err)
	}

	return nil
}

// GetTransportNodeStatus retrieves status of all member transport nodes of specified Edge Cluster
func (e *TmEdgeCluster) GetTransportNodeStatus() ([]*types.TmEdgeClusterTransportNodeStatus, error) {
	c := crudConfig{
		entityLabel:    labelTmEdgeClusterTransportNodeStatus,
		endpoint:       types.OpenApiPathVcf + types.OpenApiEndpointTmEdgeClusterTransportNodeStatus,
		endpointParams: []string{e.TmEdgeCluster.ID},
		requiresTm:     true,
	}

	return getAllInnerEntities[types.TmEdgeClusterTransportNodeStatus](&e.vcdClient.Client, c)
}
