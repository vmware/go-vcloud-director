package types

// OpenAPIEdgeGateway structure supports marshalling both - NSX-V and NSX-T edge gateways as returned by OpenAPI
// endpoint (cloudapi/1.0.0edgeGateways/), but the endpoint only allows to create NSX-T edge gateways.
type OpenAPIEdgeGateway struct {
	Status string `json:"status,omitempty"`
	ID     string `json:"id,omitempty"`
	// Name of edge gateway
	Name string `json:"name"`
	// Description of edge gateway
	Description string `json:"description"`
	// OrgVdc holds the organization vDC or vDC Group that this edge gateway belongs to. If the ownerRef is set to a VDC
	// Group, this gateway will be available across all the participating Organization vDCs in the VDC Group.
	OrgVdc *OpenApiReference `json:"orgVdc,omitempty"`
	// Org holds the organization to which the gateway belongs.
	Org *OpenApiReference `json:"orgRef,omitempty"`
	// EdgeGatewayUplink defines uplink connections for the edge gateway.
	EdgeGatewayUplinks []EdgeGatewayUplinks `json:"edgeGatewayUplinks"`
	// DistributedRoutingEnabled is a flag indicating whether distributed routing is enabled or not. The default is false.
	DistributedRoutingEnabled *bool `json:"distributedRoutingEnabled,omitempty"`
	// EdgeClusterConfig holds Edge Cluster Configuration for the Edge Gateway. Can be specified if a gateway needs to be
	// placed on a specific set of Edge Clusters. For NSX-T Edges, user should specify the ID of the NSX-T edge cluster as
	// the value of primaryEdgeCluster's backingId. The gateway defaults to the Edge Cluster of the connected External
	// Network's backing Tier-0 router, if nothing is specified.
	//
	// Note. The value of secondaryEdgeCluster will be set to NULL for NSX-T edge gateways. For NSX-V Edges, this is
	// read-only and the legacy API must be used for edge specific placement.
	EdgeClusterConfig *OpenAPIEdgeGatewayEdgeClusterConfig `json:"edgeClusterConfig,omitempty"`
	// OrgVdcNetworkCount holds the number of Org VDC networks connected to the gateway.
	OrgVdcNetworkCount *int `json:"orgVdcNetworkCount,omitempty"`
	// GatewayBacking must contain backing details of the edge gateway only if importing an NSX-T router.
	GatewayBacking *OpenAPIEdgeGatewayBacking `json:"gatewayBacking,omitempty"`

	// ServiceNetworkDefinition holds network definition in CIDR form that DNS and DHCP service on an NSX-T edge will run
	// on. The subnet prefix length must be 27. This property applies to creating or importing an NSX-T Edge. This is not
	// supported for VMC. If nothing is set, the default is 192.168.255.225/27. The DHCP listener IP network is on
	// 192.168.255.225/30. The DNS listener IP network is on 192.168.255.228/32. This field cannot be updated.
	ServiceNetworkDefinition string `json:"serviceNetworkDefinition,omitempty"`
}

// EdgeGatewayUplink defines uplink connections for the edge gateway.
type EdgeGatewayUplinks struct {
	// UplinkID contains ID of external network
	UplinkID string `json:"uplinkId,omitempty"`
	// UplinkID contains Name of external network
	UplinkName string `json:"uplinkName,omitempty"`
	// Subnets contain subnets to be used on edge gateway
	Subnets   OpenAPIEdgeGatewaySubnets `json:"subnets,omitempty"`
	Connected bool                      `json:"connected,omitempty"`
	// QuickAddAllocatedIPCount allows to allocate additional IPs during update
	QuickAddAllocatedIPCount int `json:"quickAddAllocatedIpCount,omitempty"`
	// Dedicated defines if the external network is dedicated. Dedicating the External Network will enable Route
	// Advertisement for this Edge Gateway
	Dedicated bool `json:"dedicated,omitempty"`
}

// OpenApiIPRanges is a type alias to reuse the same definitions with appropriate names
type OpenApiIPRanges = ExternalNetworkV2IPRanges

// OpenApiIPRangeValues is a type alias to reuse the same definitions with appropriate names
type OpenApiIPRangeValues = ExternalNetworkV2IPRange

// OpenAPIEdgeGatewaySubnets lists slice of OpenAPIEdgeGatewaySubnetValue values
type OpenAPIEdgeGatewaySubnets struct {
	Values []OpenAPIEdgeGatewaySubnetValue `json:"values"`
}

// OpenAPIEdgeGatewaySubnetValue holds one subnet definition in external network
type OpenAPIEdgeGatewaySubnetValue struct {
	// Gateway specified subnet gateway
	Gateway string `json:"gateway"`
	// PrefixLength from CIDR format (e.g. 24 from 192.168.1.1/24)
	PrefixLength int `json:"prefixLength"`
	// DNSSuffix can only be used for reading NSX-V edge gateway
	DNSSuffix string `json:"dnsSuffix,omitempty"`
	// DNSServer1 can only be used for reading NSX-V edge gateway
	DNSServer1 string `json:"dnsServer1,omitempty"`
	// DNSServer2 can only be used for reading NSX-V edge gateway
	DNSServer2 string `json:"dnsServer2,omitempty"`
	// IPRanges contain IP allocations
	IPRanges *OpenApiIPRanges `json:"ipRanges,omitempty"`
	// Enabled toggles if the subnet is enabled
	Enabled              bool   `json:"enabled"`
	TotalIPCount         int    `json:"totalIpCount,omitempty"`
	UsedIPCount          int    `json:"usedIpCount,omitempty"`
	PrimaryIP            string `json:"primaryIp,omitempty"`
	AutoAllocateIPRanges bool   `json:"autoAllocateIpRanges,omitempty"`
}

// OpenAPIEdgeGatewayBacking specifies edge gateway backing details
type OpenAPIEdgeGatewayBacking struct {
	BackingID       string          `json:"backingId,omitempty"`
	GatewayType     string          `json:"gatewayType,omitempty"`
	NetworkProvider NetworkProvider `json:"networkProvider"`
}

// OpenAPIEdgeGatewayEdgeCluster allows to specify edge cluster reference
type OpenAPIEdgeGatewayEdgeCluster struct {
	EdgeClusterRef OpenApiReference `json:"edgeClusterRef"`
	BackingID      string           `json:"backingId"`
}

type OpenAPIEdgeGatewayEdgeClusterConfig struct {
	PrimaryEdgeCluster   OpenAPIEdgeGatewayEdgeCluster `json:"primaryEdgeCluster,omitempty"`
	SecondaryEdgeCluster OpenAPIEdgeGatewayEdgeCluster `json:"secondaryEdgeCluster,omitempty"`
}
