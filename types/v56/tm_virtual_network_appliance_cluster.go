// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmVirtualNetworkApplianceClusters represents a list of Virtual Network Appliance Clusters.
type TmVirtualNetworkApplianceClusters struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation `json:"associations,omitzero"`
	// The list of Virtual Network Appliance Clusters.
	Values []TmVirtualNetworkApplianceCluster `json:"values"`
}

// TmVirtualNetworkApplianceCluster represents a Virtual Network Appliance Cluster.
type TmVirtualNetworkApplianceCluster struct {
	// The identifier of the Virtual Network Appliance Cluster in URN format. Read-only.
	ID string `json:"id,omitempty"`
	// Display name for the Virtual Network Appliance Cluster. Read-only.
	Name string `json:"name,omitempty"`
	// Description for the Virtual Network Appliance Cluster. Read-only.
	Description string `json:"description"`
	// The Region this Virtual Network Appliance Cluster belongs to. This will be populated when the
	// corresponding NSX Manager is associated with some Region on the Provider Management. Read-only.
	RegionRef *OpenApiReference `json:"regionRef,omitempty"`
	// Number of Virtual Network Appliances in the Virtual Network Appliance Cluster. If this information is
	// not available, nodeCount will be set to -1. Read-only.
	NodeCount int `json:"nodeCount,omitempty"`
	// Number of Organizations that are using this Virtual Network Appliance Cluster. Read-only.
	OrgCount int `json:"orgCount,omitempty"`
	// The current health status of the Virtual Network Appliance Cluster. Possible values are: UP - The
	// Virtual Network Appliance Cluster is healthy. DOWN - The Virtual Network Appliance Cluster is down.
	// DEGRADED - The Virtual Network Appliance Cluster is not operating at capacity. One or more member nodes
	// are down or inactive. UNKNOWN - The Virtual Network Appliance Cluster state is unknown. Read-only.
	HealthStatus string `json:"healthStatus,omitempty"`
	// The service type of the Virtual Network Appliance Cluster. Possible values are: VPC_SERVICES - The
	// Virtual Network Appliance Cluster is used to connect VPC workloads. ROUTE_CONTROLLER - The Virtual
	// Network Appliance Cluster is exclusively for route controller. Read-only. API 9.1.1+.
	ServiceType string `json:"serviceType,omitempty"`
	// Reference to the backing NSX Virtual Network Appliance Cluster. Read-only.
	BackingRef *OpenApiReference `json:"backingRef,omitempty"`
}

// TmEdgeClusterTransportNodeStatusList represents list containing statuses of all member Transport Nodes of
// an Edge Cluster.
type TmEdgeClusterTransportNodeStatusList struct {
	Values []TmEdgeClusterTransportNodeStatus `json:"values"`
}
