package types

// OpenAPIEdgeGateway structure supports marshalling both - NSX-V and NSX-T edge gateways as returned by OpenAPI
// endpoint (cloudapi/1.0.0edgeGateways/), but the endpoint only allows users to create NSX-T edge gateways.
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
	// QuickAddAllocatedIPCount allows users to allocate additional IPs during update
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

// OpenAPIEdgeGatewayEdgeCluster allows users to specify edge cluster reference
type OpenAPIEdgeGatewayEdgeCluster struct {
	EdgeClusterRef OpenApiReference `json:"edgeClusterRef"`
	BackingID      string           `json:"backingId"`
}

type OpenAPIEdgeGatewayEdgeClusterConfig struct {
	PrimaryEdgeCluster   OpenAPIEdgeGatewayEdgeCluster `json:"primaryEdgeCluster,omitempty"`
	SecondaryEdgeCluster OpenAPIEdgeGatewayEdgeCluster `json:"secondaryEdgeCluster,omitempty"`
}

// OpenApiOrgVdcNetwork allows users to manage Org Vdc networks
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

// OpenApiOrgVdcNetworkDhcp allows users to manage DHCP configuration for Org VDC networks by using OpenAPI endpoint
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

// NsxtFirewallGroup allows users to set either SECURITY_GROUP or IP_SET which is defined by Type field.
// SECURITY_GROUP (constant types.FirewallGroupTypeSecurityGroup) is a dynamic structure which
// allows users to add Routed Org VDC networks
//
// IP_SET (constant FirewallGroupTypeIpSet) allows users to enter static IPs and later on firewall rules
// can be created both of these objects
//
// When the type is SECURITY_GROUP 'Members' field is used to specify Org VDC networks
// When the type is IP_SET 'IpAddresses' field is used to specify IP addresses or ranges
// field is used
type NsxtFirewallGroup struct {
	// ID contains Firewall Group ID (URN format)
	// e.g. urn:vcloud:firewallGroup:d7f4e0b4-b83f-4a07-9f22-d242c9c0987a
	ID string `json:"id"`
	// Name of Firewall Group. Name are unique per 'Type'. There cannot be two SECURITY_GROUP or two
	// IP_SET objects with the same name, but there can be one object of Type SECURITY_GROUP and one
	// of Type IP_SET named the same.
	Name        string `json:"name"`
	Description string `json:"description"`
	// IP Addresses included in the group. This is only applicable for IP_SET Firewall Groups. This
	// can support IPv4 and IPv6 addresses in single, range, and CIDR formats.
	// E.g [
	//     "12.12.12.1",
	//     "10.10.10.0/24",
	//     "11.11.11.1-11.11.11.2",
	//     "2001:db8::/48",
	//	   "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
	// ],
	IpAddresses []string `json:"ipAddresses,omitempty"`

	// Members define list of Org VDC networks belonging to this Firewall Group (only for Security
	// groups )
	Members []OpenApiReference `json:"members,omitempty"`

	// OwnerRef replaces EdgeGatewayRef in API V35.0+ and can accept both - NSX-T Edge Gateway or a
	// VDC group ID
	// Sample VDC Group URN - urn:vcloud:vdcGroup:89a53000-ef41-474d-80dc-82431ff8a020
	// Sample Edge Gateway URN - urn:vcloud:gateway:71df3e4b-6da9-404d-8e44-0865751c1c38
	//
	// Note. Using API V34.0 Firewall Groups can be created for VDC groups, but on a GET operation
	// there will be no VDC group ID returned.
	OwnerRef *OpenApiReference `json:"ownerRef,omitempty"`

	// EdgeGatewayRef is a deprecated field (use OwnerRef) for setting value, but during read the
	// value is only populated in this field (not OwnerRef)
	EdgeGatewayRef *OpenApiReference `json:"edgeGatewayRef,omitempty"`

	// Type is either SECURITY_GROUP or IP_SET
	Type string `json:"type"`
}

// NsxtFirewallGroupMemberVms is a structure to read NsxtFirewallGroup associated VMs when its type
// is SECURITY_GROUP
type NsxtFirewallGroupMemberVms struct {
	VmRef *OpenApiReference `json:"vmRef"`
	// VappRef will be empty if it is a standalone VM (although hidden vApp exists)
	VappRef *OpenApiReference `json:"vappRef"`
	VdcRef  *OpenApiReference `json:"vdcRef"`
	OrgRef  *OpenApiReference `json:"orgRef"`
}

// NsxtAppPortProfile allows user to set custom application port definitions so that these can later be used
// in NSX-T Firewall rules in combination with IP Sets and Security Groups.
type NsxtAppPortProfile struct {
	ID string `json:"id,omitempty"`
	// Name must be unique per Scope
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// ApplicationPorts contains one or more protocol and port definitions
	ApplicationPorts []NsxtAppPortProfilePort `json:"applicationPorts,omitempty"`
	// OrgRef must contain at least Org ID when SCOPE==TENANT
	OrgRef *OpenApiReference `json:"orgRef,omitempty"`
	// ContextEntityId must contain:
	// * NSX-T Manager URN (when scope==PROVIDER)
	// * VDC or VDC Group ID (when scope==TENANT)
	ContextEntityId string `json:"contextEntityId,omitempty"`
	// Scope can be one of the following:
	// * SYSTEM - Read-only (The ones that are provided by SYSTEM). Constant `types.ApplicationPortProfileScopeSystem`
	// * PROVIDER - Created by Provider on a particular network provider (NSX-T manager). Constant `types.ApplicationPortProfileScopeProvider`
	// * TENANT (Created by Tenant at Org VDC level). Constant `types.ApplicationPortProfileScopeTenant`
	//
	// When scope==PROVIDER:
	//   OrgRef is not required
	//   ContextEntityId must have NSX-T Managers URN
	// When scope==TENANT
	//   OrgRef ID must be specified
	//   ContextEntityId must be set to VDC or VDC group URN
	Scope string `json:"scope,omitempty"`
}

// NsxtAppPortProfilePort allows user to set protocol and one or more ports
type NsxtAppPortProfilePort struct {
	// Protocol can be one of the following:
	// * "ICMPv4"
	// * "ICMPv6"
	// * "TCP"
	// * "UDP"
	Protocol string `json:"protocol"`
	// DestinationPorts is optional, but can define list of ports ("1000", "1500") or port ranges ("1200-1400")
	DestinationPorts []string `json:"destinationPorts,omitempty"`
}

// NsxtNatRule describes a single NAT rule of 4 diferent RuleTypes - DNAT`, `NO_DNAT`, `SNAT`, `NO_SNAT`.
//
// A SNAT or a DNAT rule on an Edge Gateway in the VMware Cloud Director environment, you always configure the rule
// from the perspective of your organization VDC.
// DNAT and NO_DNAT - outside traffic going inside
// SNAT and NO_SNAT - inside traffic going outside
// More docs in https://docs.vmware.com/en/VMware-Cloud-Director/10.2/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-9E43E3DC-C028-47B3-B7CA-59F0ED40E0A6.html
type NsxtNatRule struct {
	ID string `json:"id,omitempty"`
	// Name holds a meaningful name for the rule. (API does not enforce uniqueness)
	Name string `json:"name"`
	// Description holds optional description for the rule
	Description string `json:"description"`
	// Enabled defines if the rule is active
	Enabled bool `json:"enabled"`

	// RuleType - one of the following: `DNAT`, `NO_DNAT`, `SNAT`, `NO_SNAT`
	// * An SNAT rule translates an internal IP to an external IP and is used for outbound traffic
	// * A NO SNAT rule prevents the translation of the internal IP address of packets sent from an organization VDC out
	// to an external network or to another organization VDC network.
	// * A DNAT rule translates the external IP to an internal IP and is used for inbound traffic.
	// * A NO DNAT rule prevents the translation of the external IP address of packets received by an organization VDC
	// from an external network or from another organization VDC network.
	RuleType string `json:"ruleType"`

	// ExternalAddresses
	// * SNAT - enter the public IP address of the edge gateway for which you are configuring the SNAT rule.
	// * NO_SNAT - leave empty (but field cannot be skipped at all, therefore it does not have 'omitempty' tag)
	//
	// * DNAT - public IP address of the edge gateway for which you are configuring the DNAT rule. The IP
	// addresses that you enter must belong to the suballocated IP range of the edge gateway.
	// * NO_DNAT - leave empty
	ExternalAddresses string `json:"externalAddresses"`

	// InternalAddresses
	// * SNAT - the IP address or a range of IP addresses of the virtual machines for which you are configuring SNAT, so
	// that they can send traffic to the external network.
	//
	// * DNAT - enter the IP address or a range of IP addresses of the virtual machines for which you are configuring
	// DNAT, so that they can receive traffic from the external network.
	// * NO_DNAT - leave empty
	InternalAddresses      string            `json:"internalAddresses"`
	ApplicationPortProfile *OpenApiReference `json:"applicationPortProfile,omitempty"`

	// DnatExternalPort can set a port into which the DNAT rule is translating for the packets inbound to the virtual
	// machines.
	DnatExternalPort string `json:"dnatExternalPort,omitempty"`

	// SnatDestinationAddresses applies only for RuleTypes `SNAT`, `NO_SNAT`
	// If you want the rule to apply only for traffic to a specific domain, enter an IP address for this domain or an IP
	// address range in CIDR format. If you leave this text box blank, the SNAT rule applies to all destinations outside
	// of the local subnet.
	SnatDestinationAddresses string `json:"snatDestinationAddresses,omitempty"`

	// Logging enabled or disabled logging of that rule
	Logging bool `json:"logging"`

	// Below two fields are only supported in VCD 10.2.2+

	// FirewallMatch determines how the firewall matches the address during NATing if firewall stage is not skipped.
	// * MATCH_INTERNAL_ADDRESS indicates the firewall will be applied to internal address of a NAT rule. For SNAT, the
	// internal address is the original source address before NAT is done. For DNAT, the internal address is the translated
	// destination address after NAT is done. For REFLEXIVE, to egress traffic, the internal address is the original
	// source address before NAT is done; to ingress traffic, the internal address is the translated destination address
	// after NAT is done.
	// * MATCH_EXTERNAL_ADDRESS indicates the firewall will be applied to external address of a NAT rule. For SNAT, the
	// external address is the translated source address after NAT is done. For DNAT, the external address is the original
	// destination address before NAT is done. For REFLEXIVE, to egress traffic, the external address is the translated
	// internal address after NAT is done; to ingress traffic, the external address is the original destination address
	// before NAT is done.
	// * BYPASS firewall stage will be skipped.
	FirewallMatch string `json:"firewallMatch,omitempty"`
	// Priority helps to select rule with highest priority if an address has multiple NAT rules. A lower value means a
	// higher precedence for this rule. Maximum value 2147481599
	Priority *int `json:"priority,omitempty"`
}
