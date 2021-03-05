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

// OpenApiOrgVdcNetwork allows to manage Org Vdc networks
type OpenApiOrgVdcNetwork struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status,omitempty"`
	// OwnerRef defines Org VDC or VDC Group that this network belongs to. If the ownerRef is set to a VDC Group, this
	// network will be available across all the VDCs in the vDC Group. If the VDC Group is backed by a NSX-V network
	// provider, the Org VDC network is automatically connected to the distributed router associated with the VDC Group
	// and the "connection" field does not need to be set. For API version 35.0 and above, this field should be set for
	// network creation.
	//
	// Note. In lower API versions (i.e. 32.0) this field is not recognized and OrgVdc should be used instead
	OwnerRef *OpenApiReference `json:"ownerRef,omitempty"`

	OrgVdc *OpenApiReference `json:"orgVdc,omitempty"`

	// NetworkType describes type of Org Vdc network. ('NAT_ROUTED', 'ISOLATED')
	NetworkType string `json:"networkType"`

	// Connection specifies the edge gateway this network is connected to.
	//
	// Note. When NetworkType == ISOLATED, there is no uplink connection.
	Connection *Connection `json:"connection,omitempty"`

	// backingNetworkId contains the NSX ID of the backing network.
	BackingNetworkId string `json:"backingNetworkId,omitempty"`
	// backingNetworkType contains object type of the backing network. ('VIRTUAL_WIRE' for NSX-V, 'NSXT_FLEXIBLE_SEGMENT'
	// for NSX-T)
	BackingNetworkType string `json:"backingNetworkType,omitempty"`

	// ParentNetwork should have external network ID specified when creating NSX-V direct network
	ParentNetwork *OpenApiReference `json:"parentNetworkId"`

	// GuestVlanTaggingAllowed specifies whether guest VLAN tagging is allowed
	GuestVlanTaggingAllowed *bool `json:"guestVlanTaggingAllowed"`

	// Subnets contains list of subnets defined on
	Subnets OrgVdcNetworkSubnets `json:"subnets"`

	// SecurityGroups defines a list of firewall groups of type SECURITY_GROUP that are assigned to the Org VDC Network.
	// These groups can then be used in firewall rules to protect the Org VDC Network and allow/deny traffic.
	SecurityGroups *OpenApiReferences `json:"securityGroups,omitempty"`

	// RouteAdvertised reports if this network is advertised so that it can be routed out to the external networks. This
	// applies only to network backed by NSX-T. Value will be unset if route advertisement is not applicable.
	RouteAdvertised *bool `json:"routeAdvertised,omitempty"`

	// TotalIpCount is a read only attribute reporting total number of IPs available in network
	TotalIpCount *int `json:"totalIpCount"`

	// UsedIpCount is a read only attribute reporting number of used IPs in network
	UsedIpCount *int `json:"usedIpCount"`

	// Shared shares network with other VDCs in the organization
	Shared *bool `json:"shared,omitempty"`
}

// OrgVdcNetworkSubnetIPRanges is a type alias to reuse the same definitions with appropriate names
type OrgVdcNetworkSubnetIPRanges = ExternalNetworkV2IPRanges

// OrgVdcNetworkSubnetIPRangeValues is a type alias to reuse the same definitions with appropriate names
type OrgVdcNetworkSubnetIPRangeValues = ExternalNetworkV2IPRange

//OrgVdcNetworkSubnets
type OrgVdcNetworkSubnets struct {
	Values []OrgVdcNetworkSubnetValues `json:"values"`
}

type OrgVdcNetworkSubnetValues struct {
	Gateway      string                      `json:"gateway"`
	PrefixLength int                         `json:"prefixLength"`
	DNSServer1   string                      `json:"dnsServer1"`
	DNSServer2   string                      `json:"dnsServer2"`
	DNSSuffix    string                      `json:"dnsSuffix"`
	IPRanges     OrgVdcNetworkSubnetIPRanges `json:"ipRanges"`
}

// Connection specifies the edge gateway this network is connected to
type Connection struct {
	RouterRef      OpenApiReference `json:"routerRef"`
	ConnectionType string           `json:"connectionType,omitempty"`
}

// NsxtImportableSwitch is a type alias with better name for holding NSX-T Segments (Logical Switches) which can be used
// to back NSX-T imported Org VDC network
type NsxtImportableSwitch = OpenApiReference

// OpenApiOrgVdcNetworkDhcp allows to manage DHCP configuration for Org VDC networks by using OpenAPI endpoint
type OpenApiOrgVdcNetworkDhcp struct {
	Enabled   *bool                           `json:"enabled,omitempty"`
	LeaseTime *int                            `json:"leaseTime,omitempty"`
	DhcpPools []OpenApiOrgVdcNetworkDhcpPools `json:"dhcpPools,omitempty"`
	// Mode describes how the DHCP service is configured for this network. Once a DHCP service has been created, the mode
	// attribute cannot be changed. The mode field will default to 'EDGE' if it is not provided. This field only applies
	// to networks backed by an NSX-T network provider.
	//
	// The supported values are EDGE (default) and NETWORK.
	// * If EDGE is specified, the DHCP service of the edge is used to obtain DHCP IPs.
	// * If NETWORK is specified, a DHCP server is created for use by this network. (To use NETWORK
	//
	// In order to use DHCP for IPV6, NETWORK mode must be used. Routed networks which are using NETWORK DHCP services can
	// be disconnected from the edge gateway and still retain their DHCP configuration, however network using EDGE DHCP
	// cannot be disconnected from the gateway until DHCP has been disabled.
	Mode string `json:"mode,omitempty"`
	// IPAddress is only applicable when mode=NETWORK. This will specify IP address of DHCP server in network.
	IPAddress string `json:"ipAddress,omitempty"`
}

// OpenApiOrgVdcNetworkDhcpIpRange is a type alias to fit naming
type OpenApiOrgVdcNetworkDhcpIpRange = ExternalNetworkV2IPRange

type OpenApiOrgVdcNetworkDhcpPools struct {
	// Enabled defines if the DHCP pool is enabled or not
	Enabled *bool `json:"enabled,omitempty"`
	// IPRange holds IP ranges
	IPRange OpenApiOrgVdcNetworkDhcpIpRange `json:"ipRange"`
	// MaxLeaseTime is the maximum lease time that can be accepted on clients request
	// This applies for NSX-V Isolated network
	MaxLeaseTime *int `json:"maxLeaseTime,omitempty"`
	// DefaultLeaseTime is the lease time that clients get if they do not specify particular lease time
	// This applies for NSX-V Isolated network
	DefaultLeaseTime *int `json:"defaultLeaseTime,omitempty"`
}
