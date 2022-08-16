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
	// OwnerRef defines Org VDC or VDC Group that this network belongs to. If the ownerRef is set to a VDC Group, this
	// network will be available across all the VDCs in the vDC Group. If the VDC Group is backed by a NSX-V network
	// provider, the Org VDC network is automatically connected to the distributed router associated with the VDC Group
	// and the "connection" field does not need to be set. For API version 35.0 and above, this field should be set for
	// network creation.
	OwnerRef *OpenApiReference `json:"ownerRef,omitempty"`
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

// OrgVdcNetworkSubnets
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

	// New fields starting with 36.1

	// DnsServers are the IPs to be assigned by this DHCP service. The IP type must match the IP type of the subnet on
	// which the DHCP config is being created.
	DnsServers []string `json:"dnsServers,omitempty"`
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
	ID string `json:"id,omitempty"`
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

	// VmCriteria (VCD 10.3+) defines list of dynamic criteria that determines whether a VM belongs
	// to a dynamic firewall group. A VM needs to meet at least one criteria to belong to the
	// firewall group. In other words, the logical AND is used for rules within a single criteria
	// and the logical OR is used in between each criteria. This is only applicable for Dynamic
	// Security Groups (VM_CRITERIA Firewall Groups).
	VmCriteria []NsxtFirewallGroupVmCriteria `json:"vmCriteria,omitempty"`

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

	// Type is deprecated starting with API 36.0 (VCD 10.3+)
	Type string `json:"type,omitempty"`

	// TypeValue replaces Type starting with API 36.0 (VCD 10.3+) and can be one of:
	// SECURITY_GROUP, IP_SET, VM_CRITERIA(VCD 10.3+ only)
	// Constants `types.FirewallGroupTypeSecurityGroup`, `types.FirewallGroupTypeIpSet`,
	// `types.FirewallGroupTypeVmCriteria` can be used to set the value.
	TypeValue string `json:"typeValue,omitempty"`
}

// NsxtFirewallGroupVmCriteria defines list of rules where criteria represents boolean OR for
// matching There can be up to 3 criteria
type NsxtFirewallGroupVmCriteria struct {
	// VmCriteria is a list of rules where each rule represents boolean AND for matching VMs
	VmCriteriaRule []NsxtFirewallGroupVmCriteriaRule `json:"rules,omitempty"`
}

// NsxtFirewallGroupVmCriteriaRule defines a single rule for matching VM
// There can be up to 4 rules in a single criteria
type NsxtFirewallGroupVmCriteriaRule struct {
	AttributeType  string `json:"attributeType,omitempty"`
	AttributeValue string `json:"attributeValue,omitempty"`
	Operator       string `json:"operator,omitempty"`
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

// NsxtFirewallRule defines single NSX-T Firewall Rule
type NsxtFirewallRule struct {
	// ID contains UUID (e.g. d0bf5d51-f83a-489a-9323-1661024874b8)
	ID string `json:"id,omitempty"`
	// Name - API does not enforce uniqueness
	Name string `json:"name"`
	// Action 'ALLOW', 'DROP'
	Action string `json:"action"`
	// Enabled allows to enable or disable the rule
	Enabled bool `json:"enabled"`
	// SourceFirewallGroups contains a list of references to Firewall Groups. Empty list means 'Any'
	SourceFirewallGroups []OpenApiReference `json:"sourceFirewallGroups,omitempty"`
	// DestinationFirewallGroups contains a list of references to Firewall Groups. Empty list means 'Any'
	DestinationFirewallGroups []OpenApiReference `json:"destinationFirewallGroups,omitempty"`
	// ApplicationPortProfiles contains a list of references to Application Port Profiles. Empty list means 'Any'
	ApplicationPortProfiles []OpenApiReference `json:"applicationPortProfiles,omitempty"`
	// IpProtocol 'IPV4', 'IPV6', 'IPV4_IPV6'
	IpProtocol string `json:"ipProtocol"`
	Logging    bool   `json:"logging"`
	// Direction 'IN_OUT', 'OUT', 'IN'
	Direction string `json:"direction"`
	// Version of firewall rule. Must not be set when creating.
	Version *struct {
		// Version is incremented after each update
		Version *int `json:"version,omitempty"`
	} `json:"version,omitempty"`
}

// NsxtFirewallRuleContainer wraps NsxtFirewallRule for user-defined and default and system Firewall Rules suitable for
// API. Only UserDefinedRules are writeable. Others are read-only.
type NsxtFirewallRuleContainer struct {
	// SystemRules contain ordered list of system defined edge firewall rules. System rules are applied before user
	// defined rules in the order in which they are returned.
	SystemRules []*NsxtFirewallRule `json:"systemRules"`
	// DefaultRules contain ordered list of user defined edge firewall rules. Users are allowed to add/modify/delete rules
	// only to this list.
	DefaultRules []*NsxtFirewallRule `json:"defaultRules"`
	// UserDefinedRules ordered list of default edge firewall rules. Default rules are applied after the user defined
	// rules in the order in which they are returned.
	UserDefinedRules []*NsxtFirewallRule `json:"userDefinedRules"`
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
	// Deprecated in API V36.0
	RuleType string `json:"ruleType,omitempty"`
	// Type replaces RuleType in V36.0 and adds a new Rule - REFLEXIVE
	Type string `json:"type,omitempty"`

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

	// InternalPort specifies port number or port range for incoming network traffic. If Any Traffic is selected for the
	// Application Port Profile, the default internal port is "ANY".
	// Deprecated since API V35.0 and is replaced by DnatExternalPort
	InternalPort string `json:"internalPort,omitempty"`

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

	// Below two fields are only supported in VCD 10.2.2+ (API v35.2)

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

	// Version of NAT rule. Must not be set when creating.
	Version *struct {
		// Version is incremented after each update
		Version *int `json:"version,omitempty"`
	} `json:"version,omitempty"`
}

// NsxtIpSecVpnTunnel defines the IPsec VPN Tunnel configuration
// Some of the fields like AuthenticationMode and ConnectorInitiationMode are meant for future, because they have only
// one default value at the moment.
type NsxtIpSecVpnTunnel struct {
	// ID unique for IPsec VPN tunnel. On updates, the ID is required for the tunnel, while for create a new ID will be
	// generated.
	ID string `json:"id,omitempty"`
	// Name for the IPsec VPN Tunnel
	Name string `json:"name"`
	// Description for the IPsec VPN Tunnel
	Description string `json:"description,omitempty"`
	// Enabled describes whether the IPsec VPN Tunnel is enabled or not. The default is true.
	Enabled bool `json:"enabled"`
	// LocalEndpoint which corresponds to the Edge Gateway the IPsec VPN Tunnel is being configured on. Local Endpoint
	// requires an IP. That IP must be sub-allocated to the edge gateway
	LocalEndpoint NsxtIpSecVpnTunnelLocalEndpoint `json:"localEndpoint"`
	// RemoteEndpoint corresponds to the device on the remote site terminating the VPN tunnel
	RemoteEndpoint NsxtIpSecVpnTunnelRemoteEndpoint `json:"remoteEndpoint"`
	// PreSharedKey is key used for authentication. It must be the same on the other end of IPsec VPN Tunnel
	PreSharedKey string `json:"preSharedKey"`
	// SecurityType is the security type used for the IPsec VPN Tunnel. If nothing is specified, this will be set to
	// DEFAULT in which the default settings in NSX will be used. For custom settings, one should use the
	// NsxtIpSecVpnTunnelSecurityProfile and UpdateTunnelConnectionProperties(), GetTunnelConnectionProperties() endpoint to
	// specify custom settings. The security type will then appropriately reflect itself as CUSTOM.
	// To revert back to system default, this field must be set to "DEFAULT"
	SecurityType string `json:"securityType,omitempty"`
	// Logging sets whether logging for the tunnel is enabled or not. The default is false.
	Logging bool `json:"logging"`

	// AuthenticationMode is authentication mode this IPsec tunnel will use to authenticate with the peer endpoint. The
	// default is a pre-shared key (PSK).
	// * PSK - A known key is shared between each site before the tunnel is established.
	// * CERTIFICATE - Incoming connections are required to present an identifying digital certificate, which VCD verifies
	// has been signed by a trusted certificate authority.
	//
	// Note. Up to version 10.3 VCD only supports PSK
	AuthenticationMode string `json:"authenticationMode,omitempty"`

	// ConnectorInitiationMode is the mode used by the local endpoint to establish an IKE Connection with the remote site.
	// The default is INITIATOR.
	// Possible values are: INITIATOR , RESPOND_ONLY , ON_DEMAND
	//
	// Note. Up to version 10.3 VCD only supports INITIATOR
	ConnectorInitiationMode string `json:"connectorInitiationMode,omitempty"`

	// Version of IPsec VPN Tunnel configuration. Must not be set when creating, but required for updates
	Version *struct {
		// Version is incremented after each update
		Version *int `json:"version,omitempty"`
	} `json:"version,omitempty"`
}

// NsxtIpSecVpnTunnelLocalEndpoint which corresponds to the Edge Gateway the IPsec VPN Tunnel is being configured on.
// Local Endpoint requires an IP. That IP must be sub-allocated to the edge gateway
type NsxtIpSecVpnTunnelLocalEndpoint struct {
	// LocalId is the optional local identifier for the endpoint. It is usually the same as LocalAddress
	LocalId string `json:"localId,omitempty"`
	// LocalAddress is the IPv4 Address for the endpoint. This has to be a sub-allocated IP on the Edge Gateway. This is
	// required
	LocalAddress string `json:"localAddress"`
	// LocalNetworks is the list of local networks. These must be specified in normal Network CIDR format. At least one is
	// required
	LocalNetworks []string `json:"localNetworks,omitempty"`
}

// NsxtIpSecVpnTunnelRemoteEndpoint corresponds to the device on the remote site terminating the VPN tunnel
type NsxtIpSecVpnTunnelRemoteEndpoint struct {
	// RemoteId is needed to uniquely identify the peer site. If this tunnel is using PSK authentication,
	// the Remote ID is the public IP Address of the remote device terminating the VPN Tunnel. When NAT is configured on
	// the Remote ID, enter the private IP Address of the Remote Site. If the remote ID is not set, VCD will set the
	// remote ID to the remote address.
	RemoteId string `json:"remoteId,omitempty"`
	// RemoteAddress is IPv4 Address of the remote endpoint on the remote site. This is the Public IPv4 Address of the
	// remote device terminating the IPsec VPN Tunnel connection. This is required
	RemoteAddress string `json:"remoteAddress"`
	// RemoteNetworks is the list of remote networks. These must be specified in normal Network CIDR format.
	// Specifying no value is interpreted as 0.0.0.0/0
	RemoteNetworks []string `json:"remoteNetworks,omitempty"`
}

// NsxtIpSecVpnTunnelStatus helps to read IPsec VPN Tunnel Status
type NsxtIpSecVpnTunnelStatus struct {
	// TunnelStatus gives the overall IPsec VPN Tunnel Status. If IKE is properly set and the tunnel is up, the tunnel
	// status will be UP
	TunnelStatus string `json:"tunnelStatus"`
	IkeStatus    struct {
		// IkeServiceStatus status for the actual IKE Session for the given tunnel.
		IkeServiceStatus string `json:"ikeServiceStatus"`
		// FailReason contains more details of failure if the IKE service is not UP
		FailReason string `json:"failReason"`
	} `json:"ikeStatus"`
}

// NsxtIpSecVpnTunnelSecurityProfile specifies the given security profile/connection properties of a given IP Sec VPN
// Tunnel, such as Dead Probe Interval and IKE settings. If a security type is set to 'CUSTOM', then ike, tunnel, and/or
// dpd configurations can be specified. Otherwise, those fields are read only and are set to the values based on the
// specific security type.
type NsxtIpSecVpnTunnelSecurityProfile struct {
	// SecurityType is the security type used for the IPSec Tunnel. If nothing is specified, this will be set to DEFAULT
	// in which the default settings in NSX will be used. If CUSTOM is specified, then IKE, Tunnel, and DPD
	// configurations can be set.
	// To "RESET" configuration to DEFAULT, the NsxtIpSecVpnTunnel.SecurityType field should be changed instead of this
	SecurityType string `json:"securityType,omitempty"`
	// IkeConfiguration is the IKE Configuration to be used for the tunnel. If nothing is explicitly set, the system
	// defaults will be used.
	IkeConfiguration NsxtIpSecVpnTunnelProfileIkeConfiguration `json:"ikeConfiguration,omitempty"`
	// TunnelConfiguration contains parameters such as encryption algorithm to be used. If nothing is explicitly set,
	// the system defaults will be used.
	TunnelConfiguration NsxtIpSecVpnTunnelProfileTunnelConfiguration `json:"tunnelConfiguration,omitempty"`
	// DpdConfiguration contains Dead Peer Detection configuration. If nothing is explicitly set, the system defaults
	// will be used.
	DpdConfiguration NsxtIpSecVpnTunnelProfileDpdConfiguration `json:"dpdConfiguration,omitempty"`
}

// NsxtIpSecVpnTunnelProfileIkeConfiguration is the Internet Key Exchange (IKE) profiles provide information about the
// algorithms that are used to authenticate, encrypt, and establish a shared secret between network sites when you
// establish an IKE tunnel.
//
// Note. While quite a few fields accepts a []string it actually supports single values only.
type NsxtIpSecVpnTunnelProfileIkeConfiguration struct {
	// IkeVersion IKE Protocol Version to use.
	// The default is IKE_V2.
	//
	// Possible values are: IKE_V1 , IKE_V2 , IKE_FLEX
	IkeVersion string `json:"ikeVersion"`
	// EncryptionAlgorithms contains list of Encryption algorithms for IKE. This is used during IKE negotiation.
	// Default is AES_128.
	//
	// Possible values are: AES_128 , AES_256 , AES_GCM_128 , AES_GCM_192 , AES_GCM_256
	EncryptionAlgorithms []string `json:"encryptionAlgorithms"`
	// DigestAlgorithms contains list of Digest algorithms - secure hashing algorithms to use during the IKE negotiation.
	//
	// Default is SHA2_256.
	//
	// Possible values are: SHA1 , SHA2_256 , SHA2_384 , SHA2_512
	DigestAlgorithms []string `json:"digestAlgorithms"`
	// DhGroups contains list of Diffie-Hellman groups to be used if Perfect Forward Secrecy is enabled. These are
	// cryptography schemes that allows the peer site and the edge gateway to establish a shared secret over an insecure
	// communications channel
	//
	// Default is GROUP14.
	//
	// Possible values are: GROUP2, GROUP5, GROUP14, GROUP15, GROUP16, GROUP19, GROUP20, GROUP21
	DhGroups []string `json:"dhGroups"`
	// SaLifeTime is the Security Association life time in seconds. It is number of seconds before the IPsec tunnel needs
	// to reestablish
	//
	// Default is 86400 seconds (1 day).
	SaLifeTime *int `json:"saLifeTime"`
}

// NsxtIpSecVpnTunnelProfileTunnelConfiguration adjusts IPsec VPN Tunnel settings
//
// Note. While quite a few fields accepts a []string it actually supports single values only.
type NsxtIpSecVpnTunnelProfileTunnelConfiguration struct {
	// PerfectForwardSecrecyEnabled enabled or disabled. PFS (Perfect Forward Secrecy) ensures the same key will not be
	// generated and used again, and because of this, the VPN peers negotiate a new Diffie-Hellman key exchange. This
	// would ensure if a hacker\criminal was to compromise the private key, they would only be able to access data in
	// transit protected by that key. Any future data will not be compromised, as future data would not be associated
	// with that compromised key. Both sides of the VPN must be able to support PFS in order for PFS to work.
	//
	// The default value is true.
	PerfectForwardSecrecyEnabled bool `json:"perfectForwardSecrecyEnabled"`
	// DfPolicy Policy for handling defragmentation bit. The default is COPY.
	//
	// Possible values are: COPY, CLEAR
	// * COPY Copies the defragmentation bit from the inner IP packet to the outer packet.
	// * CLEAR Ignores the defragmentation bit present in the inner packet.
	DfPolicy string `json:"dfPolicy"`

	// EncryptionAlgorithms contains list of Encryption algorithms to use in IPSec tunnel establishment.
	// Default is AES_GCM_128.
	// * NO_ENCRYPTION_AUTH_AES_GMAC_XX (XX is 128, 192, 256) enables authentication on input data without encryption.
	// If one of these options is used, digest algorithm should be empty.
	//
	// Possible values are: AES_128, AES_256, AES_GCM_128, AES_GCM_192, AES_GCM_256, NO_ENCRYPTION_AUTH_AES_GMAC_128,
	// NO_ENCRYPTION_AUTH_AES_GMAC_192, NO_ENCRYPTION_AUTH_AES_GMAC_256, NO_ENCRYPTION
	EncryptionAlgorithms []string `json:"encryptionAlgorithms"`

	// DigestAlgorithms contains list of Digest algorithms to be used for message digest. The default digest algorithm is
	// implicitly covered by default encryption algorithm AES_GCM_128.
	//
	// Possible values are: SHA1 , SHA2_256 , SHA2_384 , SHA2_512
	// Note. Only one value can be set inside the slice
	DigestAlgorithms []string `json:"digestAlgorithms"`

	// DhGroups contains list of Diffie-Hellman groups to be used is PFS is enabled. Default is GROUP14.
	//
	// Possible values are: GROUP2, GROUP5, GROUP14, GROUP15, GROUP16, GROUP19, GROUP20, GROUP21
	// Note. Only one value can be set inside the slice
	DhGroups []string `json:"dhGroups"`

	// SaLifeTime is the Security Association life time in seconds.
	//
	// Default is 3600 seconds.
	SaLifeTime *int `json:"saLifeTime"`
}

// NsxtIpSecVpnTunnelProfileDpdConfiguration specifies the Dead Peer Detection Profile. This configurations determines
// the number of seconds to wait in time between probes to detect if an IPSec peer is alive or not. The default value
// for the DPD probe interval is 60 seconds.
type NsxtIpSecVpnTunnelProfileDpdConfiguration struct {
	// ProbeInternal is value of the probe interval in seconds. This defines a periodic interval for DPD probes. The
	// minimum is 3 seconds and the maximum is 60 seconds.
	ProbeInterval int `json:"probeInterval"`
}

// NsxtAlbController helps to integrate VMware Cloud Director with NSX-T Advanced Load Balancer deployment.
// Controller instances are registered with VMware Cloud Director instance. Controller instances serve as a central
// control plane for the load-balancing services provided by NSX-T Advanced Load Balancer.
// To configure an NSX-T ALB one needs to supply AVI Controller endpoint, credentials and license to be used.
type NsxtAlbController struct {
	// ID holds URN for load balancer controller (e.g. urn:vcloud:loadBalancerController:aa23ef66-ba32-48b2-892f-7acdffe4587e)
	ID string `json:"id,omitempty"`
	// Name as shown in VCD
	Name string `json:"name"`
	// Description as shown in VCD
	Description string `json:"description,omitempty"`
	// Url of ALB controller
	Url string `json:"url"`
	// Username of user
	Username string `json:"username"`
	// Password (will not be returned on read)
	Password string `json:"password,omitempty"`
	// LicenseType By enabling this feature, the provider acknowledges that they have independently licensed the
	// enterprise version of the NSX AVI LB.
	// Possible options: 'BASIC', 'ENTERPRISE'
	// This field was removed since VCD 10.4.0 (v37.0) in favor of NsxtAlbServiceEngineGroup.SupportedFeatureSet
	LicenseType string `json:"licenseType,omitempty"`
	// Version of ALB (e.g. 20.1.3). Read-only
	Version string `json:"version,omitempty"`
}

// NsxtAlbImportableCloud allows user to list importable NSX-T ALB Clouds. Each importable cloud can only be imported
// once. It has a flag AlreadyImported which hints if it is already consumed or not.
type NsxtAlbImportableCloud struct {
	// ID (e.g. 'cloud-43726181-f73e-41f2-bf1d-8a9609502586')
	ID string `json:"id"`

	DisplayName string `json:"displayName"`
	// AlreadyImported shows if this ALB Cloud is already imported
	AlreadyImported bool `json:"alreadyImported"`

	// NetworkPoolRef contains a reference to NSX-T network pool
	NetworkPoolRef OpenApiReference `json:"networkPoolRef"`

	// TransportZoneName contains transport zone name
	TransportZoneName string `json:"transportZoneName"`
}

// NsxtAlbCloud helps to use the virtual infrastructure provided by NSX Advanced Load Balancer, register NSX-T Cloud
// instances with VMware Cloud Director by consuming NsxtAlbImportableCloud.
type NsxtAlbCloud struct {
	// ID (e.g. 'urn:vcloud:loadBalancerCloud:947ea2ba-e448-4249-91f7-1432b3d2fcbf')
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	// Name of NSX-T ALB Cloud
	Name string `json:"name"`
	// Description of NSX-T ALB Cloud
	Description string `json:"description,omitempty"`
	// LoadBalancerCloudBacking uniquely identifies a Load Balancer Cloud configured within a Load Balancer Controller. At
	// the present, VCD only supports NSX-T Clouds configured within an NSX-ALB Controller deployment.
	LoadBalancerCloudBacking NsxtAlbCloudBacking `json:"loadBalancerCloudBacking"`
	// NetworkPoolRef for the Network Pool associated with this Cloud
	NetworkPoolRef *OpenApiReference `json:"networkPoolRef"`
	// HealthStatus contains status of the Load Balancer Cloud. Possible values are:
	// UP - The cloud is healthy and ready to enable Load Balancer for an Edge Gateway.
	// DOWN - The cloud is in a failure state. Enabling Load balancer on an Edge Gateway may not be possible.
	// RUNNING - The cloud is currently processing. An example is if it's enabling a Load Balancer for an Edge Gateway.
	// UNAVAILABLE - The cloud is unavailable.
	// UNKNOWN - The cloud state is unknown.
	HealthStatus string `json:"healthStatus,omitempty"`
	// DetailedHealthMessage contains detailed message on the health of the Cloud.
	DetailedHealthMessage string `json:"detailedHealthMessage,omitempty"`
}

// NsxtAlbCloudBacking is embedded into NsxtAlbCloud
type NsxtAlbCloudBacking struct {
	// BackingId is the ID of NsxtAlbImportableCloud
	BackingId string `json:"backingId"`
	// BackingType contains type of ALB (The only supported now is 'NSXALB_NSXT')
	BackingType string `json:"backingType,omitempty"`
	// LoadBalancerControllerRef contains reference to NSX-T ALB Controller
	LoadBalancerControllerRef OpenApiReference `json:"loadBalancerControllerRef"`
}

// NsxtAlbServiceEngineGroup provides virtual service management capabilities for tenants. This entity can be created
// by referencing a backing importable service engine group - NsxtAlbImportableServiceEngineGroups.
//
// A service engine group is an isolation domain that also defines shared service engine properties, such as size,
// network access, and failover. Resources in a service engine group can be used for different virtual services,
// depending on your tenant needs. These resources cannot be shared between different service engine groups.
type NsxtAlbServiceEngineGroup struct {
	// ID of the Service Engine Group
	ID string `json:"id,omitempty"`
	// Name of the Service Engine Group
	Name string `json:"name"`
	// Description of the Service Engine Group
	Description string `json:"description"`
	// ServiceEngineGroupBacking holds backing details that uniquely identifies a Load Balancer Service Engine Group
	// configured within a load balancer cloud.
	ServiceEngineGroupBacking ServiceEngineGroupBacking `json:"serviceEngineGroupBacking"`
	// HaMode defines High Availability Mode for Service Engine Group
	// * ELASTIC_N_PLUS_M_BUFFER - Service Engines will scale out to N active nodes with M nodes as buffer.
	// * ELASTIC_ACTIVE_ACTIVE - Active-Active with scale out.
	// * LEGACY_ACTIVE_STANDBY - Traditional single Active-Standby configuration
	HaMode string `json:"haMode,omitempty"`
	// ReservationType can be `DEDICATED` or `SHARED`
	// * DEDICATED - Dedicated to a single Edge Gateway and can only be assigned to a single Edge Gateway
	// * SHARED - Shared between multiple Edge Gateways. Can be assigned to multiple Edge Gateways
	ReservationType string `json:"reservationType"`
	// MaxVirtualServices holds  maximum number of virtual services supported on the Load Balancer Service Engine Group
	MaxVirtualServices *int `json:"maxVirtualServices,omitempty"`
	// NumDeployedVirtualServices shows number of virtual services currently deployed on the Load Balancer Service Engine
	// Group
	NumDeployedVirtualServices *int `json:"numDeployedVirtualServices,omitempty"`
	// ReservedVirtualServices holds number of virtual services already reserved on the Load Balancer Service Engine Group.
	// This value is the sum of the guaranteed virtual services given to Edge Gateways assigned to the Load Balancer
	// Service Engine Group.
	ReservedVirtualServices *int `json:"reservedVirtualServices,omitempty"`
	// OverAllocated indicates whether the maximum number of virtual services supported on the Load Balancer Service
	// Engine Group has been surpassed by the current number of reserved virtual services.
	OverAllocated *bool `json:"overAllocated,omitempty"`
	// SupportedFeatureSet was added in VCD 10.4.0 (v37.0) as substitute of NsxtAlbController.LicenseType.
	// Possible values are: "STANDARD", "PREMIUM".
	SupportedFeatureSet string `json:"supportedFeatureSet,omitempty"`
}

type ServiceEngineGroupBacking struct {
	BackingId            string            `json:"backingId"`
	BackingType          string            `json:"backingType,omitempty"`
	LoadBalancerCloudRef *OpenApiReference `json:"loadBalancerCloudRef"`
}

// NsxtAlbImportableServiceEngineGroups provides capability to list all Importable Service Engine Groups available in
// ALB Controller so that they can be consumed by NsxtAlbServiceEngineGroup
//
// Note. The API does not return Importable Service Engine Group once it is consumed.
type NsxtAlbImportableServiceEngineGroups struct {
	// ID (e.g. 'serviceenginegroup-b633f16f-2733-4bf5-b552-3a6c4949caa4')
	ID string `json:"id"`
	// DisplayName is the name of
	DisplayName string `json:"displayName"`
	// HaMode (e.g. 'ELASTIC_N_PLUS_M_BUFFER')
	HaMode string `json:"haMode"`
}

// NsxtAlbConfig describes Load Balancer Service configuration on an NSX-T Edge Gateway
type NsxtAlbConfig struct {
	// Enabled is a mandatory flag indicating whether Load Balancer Service is enabled or not
	Enabled bool `json:"enabled"`
	// LicenseType of the backing Load Balancer Cloud.
	// * BASIC - Basic edition of the NSX Advanced Load Balancer.
	// * ENTERPRISE - Full featured edition of the NSX Advanced Load Balancer.
	// This field was removed since VCD 10.4.0 (v37.0) in favor of NsxtAlbConfig.SupportedFeatureSet
	LicenseType string `json:"licenseType,omitempty"`
	// SupportedFeatureSet was added in VCD 10.4.0 (v37.0) as substitute of NsxtAlbConfig.LicenseType.
	// Possible values are: "STANDARD", "PREMIUM".
	SupportedFeatureSet string `json:"supportedFeatureSet,omitempty"`
	// LoadBalancerCloudRef
	LoadBalancerCloudRef *OpenApiReference `json:"loadBalancerCloudRef,omitempty"`
	// ServiceNetworkDefinition in Gateway CIDR format which will be used by Load Balancer service. All the load balancer
	// service engines associated with the Service Engine Group will be attached to this network. The subnet prefix length
	// must be 25. If nothing is set, the default is 192.168.255.1/25. Default CIDR can be configured. This field cannot
	// be updated.
	ServiceNetworkDefinition string `json:"serviceNetworkDefinition,omitempty"`
}

// NsxtAlbServiceEngineGroupAssignment configures Service Engine Group assignments to Edge Gateway. The only mandatory
// fields are `GatewayRef` and `ServiceEngineGroupRef`. `MinVirtualServices` and `MaxVirtualServices` are only available
// for SHARED Service Engine Groups.
type NsxtAlbServiceEngineGroupAssignment struct {
	ID string `json:"id,omitempty"`
	// GatewayRef contains reference to Edge Gateway
	GatewayRef *OpenApiReference `json:"gatewayRef"`
	// ServiceEngineGroupRef contains a reference to Service Engine Group
	ServiceEngineGroupRef *OpenApiReference `json:"serviceEngineGroupRef"`
	// GatewayOrgRef optional Org reference for gateway
	GatewayOrgRef *OpenApiReference `json:"gatewayOrgRef,omitempty"`
	// GatewayOwnerRef can be a VDC or VDC group
	GatewayOwnerRef    *OpenApiReference `json:"gatewayOwnerRef,omitempty"`
	MaxVirtualServices *int              `json:"maxVirtualServices,omitempty"`
	MinVirtualServices *int              `json:"minVirtualServices,omitempty"`
	// NumDeployedVirtualServices is a read only value
	NumDeployedVirtualServices int `json:"numDeployedVirtualServices,omitempty"`
}

// NsxtAlbPool defines configuration of a single NSX-T ALB Pool. Pools maintain the list of servers assigned to them and
// perform health monitoring, load balancing, persistence. A pool may only be used or referenced by only one virtual
// service at a time.
type NsxtAlbPool struct {
	ID string `json:"id,omitempty"`
	// Name is mandatory
	Name string `json:"name"`
	// Description is optional
	Description string `json:"description,omitempty"`

	// GatewayRef is mandatory and associates NSX-T Edge Gateway with this Load Balancer Pool.
	GatewayRef OpenApiReference `json:"gatewayRef"`

	// Enabled defines if the Pool is enabled
	Enabled *bool `json:"enabled,omitempty"`

	// Algorithm for choosing a member within the pools list of available members for each new connection.
	// Default value is LEAST_CONNECTIONS
	// Supported algorithms are:
	// * LEAST_CONNECTIONS
	// * ROUND_ROBIN
	// * CONSISTENT_HASH (uses Source IP Address hash)
	// * FASTEST_RESPONSE
	// * LEAST_LOAD
	// * FEWEST_SERVERS
	// * RANDOM
	// * FEWEST_TASKS
	// * CORE_AFFINITY
	Algorithm string `json:"algorithm,omitempty"`

	// DefaultPort defines destination server port used by the traffic sent to the member.
	DefaultPort *int `json:"defaultPort,omitempty"`

	// GracefulTimeoutPeriod sets maximum time (in minutes) to gracefully disable a member. Virtual service waits for the
	// specified time before terminating the existing connections to the pool members that are disabled.
	//
	// Special values: 0 represents Immediate, -1 represents Infinite.
	GracefulTimeoutPeriod *int `json:"gracefulTimeoutPeriod,omitempty"`

	// PassiveMonitoringEnabled sets if client traffic should be used to check if pool member is up or down.
	PassiveMonitoringEnabled *bool `json:"passiveMonitoringEnabled,omitempty"`

	// HealthMonitors check member servers health. It can be monitored by using one or more health monitors. Active
	// monitors generate synthetic traffic and mark a server up or down based on the response.
	HealthMonitors []NsxtAlbPoolHealthMonitor `json:"healthMonitors,omitempty"`

	// Members field defines list of destination servers which are used by the Load Balancer Pool to direct load balanced
	// traffic.
	Members []NsxtAlbPoolMember `json:"members,omitempty"`

	// CaCertificateRefs point to root certificates to use when validating certificates presented by the pool members.
	CaCertificateRefs []OpenApiReference `json:"caCertificateRefs,omitempty"`

	// CommonNameCheckEnabled specifies whether to check the common name of the certificate presented by the pool member.
	// This cannot be enabled if no caCertificateRefs are specified.
	CommonNameCheckEnabled *bool `json:"commonNameCheckEnabled,omitempty"`

	// DomainNames holds a list of domain names which will be used to verify the common names or subject alternative
	// names presented by the pool member certificates. It is performed only when common name check
	// (CommonNameCheckEnabled) is enabled. If common name check is enabled, but domain names are not specified then the
	// incoming host header will be used to check the certificate.
	DomainNames []string `json:"domainNames,omitempty"`

	// PersistenceProfile of a Load Balancer Pool. Persistence profile will ensure that the same user sticks to the same
	// server for a desired duration of time. If the persistence profile is unmanaged by Cloud Director, updates that
	// leave the values unchanged will continue to use the same unmanaged profile. Any changes made to the persistence
	// profile will cause Cloud Director to switch the pool to a profile managed by Cloud Director.
	PersistenceProfile *NsxtAlbPoolPersistenceProfile `json:"persistenceProfile,omitempty"`

	// MemberCount is a read only value that reports number of members added
	MemberCount int `json:"memberCount,omitempty"`

	// EnabledMemberCount is a read only value that reports number of enabled members
	EnabledMemberCount int `json:"enabledMemberCount,omitempty"`

	// UpMemberCount is a read only value that reports number of members that are serving traffic
	UpMemberCount int `json:"upMemberCount,omitempty"`

	// HealthMessage shows a pool health status (e.g. "The pool is unassigned.")
	HealthMessage string `json:"healthMessage,omitempty"`

	// VirtualServiceRefs holds list of Load Balancer Virtual Services associated with this Load balancer Pool.
	VirtualServiceRefs []OpenApiReference `json:"virtualServiceRefs,omitempty"`
}

// NsxtAlbPoolHealthMonitor checks member servers health. Active monitor generates synthetic traffic and mark a server
// up or down based on the response.
type NsxtAlbPoolHealthMonitor struct {
	Name string `json:"name,omitempty"`
	// SystemDefined is a boolean value
	SystemDefined bool `json:"systemDefined,omitempty"`
	// Type
	// * HTTP - HTTP request/response is used to validate health.
	// * HTTPS - Used against HTTPS encrypted web servers to validate health.
	// * TCP - TCP connection is used to validate health.
	// * UDP - A UDP datagram is used to validate health.
	// * PING - An ICMP ping is used to validate health.
	Type string `json:"type"`
}

// NsxtAlbPoolMember defines a single destination server which is used by the Load Balancer Pool to direct load balanced
// traffic.
type NsxtAlbPoolMember struct {
	// Enabled defines if member is enabled (will receive incoming requests) or not
	Enabled bool `json:"enabled"`
	// IpAddress of the Load Balancer Pool member.
	IpAddress string `json:"ipAddress"`

	// Port number of the Load Balancer Pool member. If unset, the port that the client used to connect will be used.
	Port int `json:"port,omitempty"`

	// Ratio of selecting eligible servers in the pool.
	Ratio *int `json:"ratio,omitempty"`

	// MarkedDownBy gives the names of the health monitors that marked the member as down when it is DOWN. If a monitor
	// cannot be determined, the value will be UNKNOWN.
	MarkedDownBy []string `json:"markedDownBy,omitempty"`

	// HealthStatus of the pool member. Possible values are:
	// * UP - The member is operational
	// * DOWN - The member is down
	// * DISABLED - The member is disabled
	// * UNKNOWN - The state is unknown
	HealthStatus string `json:"healthStatus,omitempty"`

	// DetailedHealthMessage contains non-localized detailed message on the health of the pool member.
	DetailedHealthMessage string `json:"detailedHealthMessage,omitempty"`
}

// NsxtAlbPoolPersistenceProfile holds Persistence Profile of a Load Balancer Pool. Persistence profile will ensure that
// the same user sticks to the same server for a desired duration of time. If the persistence profile is unmanaged by
// Cloud Director, updates that leave the values unchanged will continue to use the same unmanaged profile. Any changes
// made to the persistence profile will cause Cloud Director to switch the pool to a profile managed by Cloud Director.
type NsxtAlbPoolPersistenceProfile struct {
	// Name field is tricky. It remains empty in some case, but if it is sent it can become computed.
	// (e.g. setting 'CUSTOM_HTTP_HEADER' results in value being
	// 'VCD-LoadBalancer-3510eae9-53bb-49f1-b7aa-7aedf5ce3a77-CUSTOM_HTTP_HEADER')
	Name string `json:"name,omitempty"`

	// Type of persistence strategy to use. Supported values are:
	//  * CLIENT_IP - The clients IP is used as the identifier and mapped to the server
	//  * HTTP_COOKIE - Load Balancer inserts a cookie into HTTP responses. Cookie name must be provided as value
	//  * CUSTOM_HTTP_HEADER - Custom, static mappings of header values to specific servers are used. Header name must be
	// provided as value
	//  * APP_COOKIE - Load Balancer reads existing server cookies or URI embedded data such as JSessionID. Cookie name
	// must be provided as value
	//  * TLS - Information is embedded in the client's SSL/TLS ticket ID. This will use default system profile
	// System-Persistence-TLS
	Type string `json:"type,omitempty"`

	// Value of attribute based on selected persistence type.
	// This is required for HTTP_COOKIE, CUSTOM_HTTP_HEADER and APP_COOKIE persistence types.
	//
	// HTTP_COOKIE, APP_COOKIE must have cookie name set as the value and CUSTOM_HTTP_HEADER must have header name set as
	// the value.
	Value string `json:"value,omitempty"`
}

// NsxtAlbVirtualService combines Load Balancer Pools with Service Engine Groups and exposes a virtual service on
// defined VIP (virtual IP address) while optionally allowing to use encrypted traffic
type NsxtAlbVirtualService struct {
	ID string `json:"id,omitempty"`

	// Name contains meaningful name
	Name string `json:"name,omitempty"`

	// Description is optional
	Description string `json:"description,omitempty"`

	// Enabled defines if the virtual service is enabled to accept traffic
	Enabled *bool `json:"enabled"`

	// ApplicationProfile sets protocol for load balancing by using NsxtAlbVirtualServiceApplicationProfile
	ApplicationProfile NsxtAlbVirtualServiceApplicationProfile `json:"applicationProfile"`

	// GatewayRef contains NSX-T Edge Gateway reference
	GatewayRef OpenApiReference `json:"gatewayRef"`
	//LoadBalancerPoolRef contains Pool reference
	LoadBalancerPoolRef OpenApiReference `json:"loadBalancerPoolRef"`
	// ServiceEngineGroupRef points to service engine group (which must be assigned to NSX-T Edge Gateway)
	ServiceEngineGroupRef OpenApiReference `json:"serviceEngineGroupRef"`

	// CertificateRef contains certificate reference if serving encrypted traffic
	CertificateRef *OpenApiReference `json:"certificateRef,omitempty"`

	// ServicePorts define one or more ports (or port ranges) of the virtual service
	ServicePorts []NsxtAlbVirtualServicePort `json:"servicePorts"`

	// VirtualIpAddress to be used for exposing this virtual service
	VirtualIpAddress string `json:"virtualIpAddress"`

	// HealthStatus contains status of the Load Balancer Cloud. Possible values are:
	// UP - The cloud is healthy and ready to enable Load Balancer for an Edge Gateway.
	// DOWN - The cloud is in a failure state. Enabling Load balancer on an Edge Gateway may not be possible.
	// RUNNING - The cloud is currently processing. An example is if it's enabling a Load Balancer for an Edge Gateway.
	// UNAVAILABLE - The cloud is unavailable.
	// UNKNOWN - The cloud state is unknown.
	HealthStatus string `json:"healthStatus,omitempty"`

	// HealthMessage shows a pool health status (e.g. "The pool is unassigned.")
	HealthMessage string `json:"healthMessage,omitempty"`

	// DetailedHealthMessage containes a more in depth health message
	DetailedHealthMessage string `json:"detailedHealthMessage,omitempty"`
}

// NsxtAlbVirtualServicePort port (or port ranges) of the virtual service
type NsxtAlbVirtualServicePort struct {
	// PortStart is always required
	PortStart *int `json:"portStart"`
	// PortEnd is only required if a port range is specified. For single port cases PortStart is sufficient
	PortEnd *int `json:"portEnd,omitempty"`
	// SslEnabled defines if traffic is served as secure. CertificateRef must be specified in NsxtAlbVirtualService when
	// true
	SslEnabled *bool `json:"sslEnabled,omitempty"`
	// TcpUdpProfile defines
	TcpUdpProfile *NsxtAlbVirtualServicePortTcpUdpProfile `json:"tcpUdpProfile,omitempty"`
}

// NsxtAlbVirtualServicePortTcpUdpProfile profile determines the type and settings of the network protocol that a
// subscribing virtual service will use. It sets a number of parameters, such as whether the virtual service is a TCP
// proxy versus a pass-through via fast path. A virtual service can have both TCP and UDP enabled, which is useful for
// protocols such as DNS or syslog.
type NsxtAlbVirtualServicePortTcpUdpProfile struct {
	SystemDefined bool `json:"systemDefined"`
	// Type defines L4 or L4_TLS profiles:
	// * TCP_PROXY (the only possible type when L4_TLS is used). Enabling TCP Proxy causes ALB to terminate an inbound
	// connection from a client. Any application data from the client that is destined for a server is forwarded to that
	// server over a new TCP connection. Separating (or proxying) the client-to-server connections enables ALB to provide
	// enhanced security, such as TCP protocol sanitization or DoS mitigation. It also provides better client and server
	// performance, such as maximizing client and server TCP MSS or window sizes independently and buffering server
	// responses. One must use a TCP/UDP profile with the type set to Proxy for application profiles such as HTTP.
	//
	// * TCP_FAST_PATH profile does not proxy TCP connections - rather, it directly connects clients to the
	// destination server and translates the client's destination virtual service address with the chosen destination
	// server's IP address. The client's source IP address is still translated to the Service Engine address to ensure
	// that server response traffic returns symmetrically.
	//
	// * UDP_FAST_PATH profile enables a virtual service to support UDP. Avi Vantage translates the client's destination
	// virtual service address to the destination server and rewrites the client's source IP address to the Service
	// Engine's address when forwarding the packet to the server. This ensures that server response traffic traverses
	// symmetrically through the original SE.
	Type string `json:"type"`
}

// NsxtAlbVirtualServiceApplicationProfile sets protocol for load balancing. Type field defines possible options.
type NsxtAlbVirtualServiceApplicationProfile struct {
	SystemDefined bool `json:"systemDefined,omitempty"`
	// Type defines Traffic
	// * HTTP
	// * HTTPS (certificate reference is mandatory)
	// * L4
	// * L4 TLS (certificate reference is mandatory)
	Type string `json:"type"`
}

// DistributedFirewallRule represents a single Distributed Firewall rule
type DistributedFirewallRule struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`

	// Action field. Deprecated in favor of ActionValue in VCD 10.2.2+ (API V35.2)
	Action string `json:"action,omitempty"`

	// Description field is not shown in UI. 'Comments' field was introduced in 10.3.2 and is shown
	// in UI.
	Description string `json:"description,omitempty"`

	// ApplicationPortProfiles contains a list of references to Application Port Profiles. Empty
	// list means 'Any'
	ApplicationPortProfiles []OpenApiReference `json:"applicationPortProfiles,omitempty"`

	// SourceFirewallGroups contains a list of references to Firewall Groups. Empty list means 'Any'
	SourceFirewallGroups []OpenApiReference `json:"sourceFirewallGroups,omitempty"`
	// DestinationFirewallGroups contains a list of references to Firewall Groups. Empty list means
	// 'Any'
	DestinationFirewallGroups []OpenApiReference `json:"destinationFirewallGroups,omitempty"`

	// Direction 'IN_OUT', 'OUT', 'IN'
	Direction string `json:"direction"`
	Enabled   bool   `json:"enabled"`

	// IpProtocol 'IPV4', 'IPV6', 'IPV4_IPV6'
	IpProtocol string `json:"ipProtocol"`

	Logging bool `json:"logging"`

	// NetworkContextProfiles sets  list of layer 7 network context profiles where this firewall
	// rule is applicable. Null value or an empty list will be treated as 'ANY' which means rule
	// applies to all applications and domains.
	NetworkContextProfiles []OpenApiReference `json:"networkContextProfiles,omitempty"`

	// Version describes the current version of the entity. To prevent clients from overwriting each
	// other's changes, update operations must include the version which can be obtained by issuing
	// a GET operation. If the version number on an update call is missing, the operation will be
	// rejected. This is only needed on update calls.
	Version *DistributedFirewallRuleVersion `json:"version,omitempty"`

	// New fields starting with 35.2

	// ActionValue replaces deprecated field Action and defines action to be applied to all the
	// traffic that meets the firewall rule criteria. It determines if the rule permits or blocks
	// traffic. Property is required if action is not set. Below are valid values:
	// * ALLOW permits traffic to go through the firewall.
	// * DROP blocks the traffic at the firewall. No response is sent back to the source.
	// * REJECT blocks the traffic at the firewall. A response is sent back to the source.
	ActionValue string `json:"actionValue,omitempty"`

	// New fields starting with 36.2

	// Comments permits setting text for user entered comments on the firewall rule. Length cannot
	// exceed 2048 characters. Comments are shown in UI for 10.3.2+.
	Comments string `json:"comments,omitempty"`

	// SourceGroupsExcluded reverses the list specified in SourceFirewallGroups and the rule gets
	// applied on all the groups that are NOT part of the SourceFirewallGroups. If false, the rule
	// applies to the all the groups including the source groups.
	SourceGroupsExcluded *bool `json:"sourceGroupsExcluded,omitempty"`

	// DestinationGroupsExcluded reverses the list specified in DestinationFirewallGroups and the
	// rule gets applied on all the groups that are NOT part of the DestinationFirewallGroups. If
	// false, the rule applies to the all the groups in DestinationFirewallGroups.
	DestinationGroupsExcluded *bool `json:"destinationGroupsExcluded,omitempty"`
}

type DistributedFirewallRules struct {
	Values []*DistributedFirewallRule `json:"values"`
}

type DistributedFirewallRuleVersion struct {
	Version int `json:"version"`
}

type NsxtNetworkContextProfile struct {
	OrgRef               *OpenApiReference `json:"orgRef"`
	ContextEntityID      interface{}       `json:"contextEntityId"`
	NetworkProviderScope interface{}       `json:"networkProviderScope"`
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`

	// Scope of NSX-T Network Context Profile
	// SYSTEM profiles are available to all tenants. They are default profiles from the backing networking provider.
	// PROVIDER profiles are available to all tenants. They are defined by the provider at a system level.
	// TENANT profiles are available only to the specific tenant organization. They are defined by the tenant or by a provider on behalf of a tenant.
	Scope      string                                `json:"scope"`
	Attributes []NsxtNetworkContextProfileAttributes `json:"attributes"`
}
type NsxtNetworkContextProfileAttributes struct {
	Type          string      `json:"type"`
	Values        []string    `json:"values"`
	SubAttributes interface{} `json:"subAttributes"`
}

// SecurityTag represents An individual security tag
type SecurityTag struct {
	// Entities are the list of entities to tag in urn format.
	Entities []string `json:"entities"`
	// Tag is the tag name to use.
	Tag string `json:"tag"`
}

// SecurityTaggedEntity is an entity that has a tag.
type SecurityTaggedEntity struct {
	// EntityType is the type of entity. Currently, only “vm” is supported.
	EntityType string `json:"entityType"`
	// ID is the unique identifier of the entity in URN format.
	ID string `json:"id"`
	// Name of the entity.
	Name string `json:"name"`
	// OwnerRef is the owner of the specified entity such as vDC or vDC Group. If not applicable, field is not set.
	OwnerRef *OpenApiReference `json:"ownerRef"`
	// ParentRef is the parent of the entity such as vApp if the entity is a VM. If not applicable, field is not set.
	ParentRef *OpenApiReference `json:"parentRef"`
}

// SecurityTagValue describes the most basic tag structure: its value.
type SecurityTagValue struct {
	// Tag is the value of the tag. The value is case-agnostic and will be converted to lower-case.
	Tag string `json:"tag"`
}

// EntitySecurityTags is a list of a tags assigned to a specific entity
type EntitySecurityTags struct {
	// Tags is the list of tags. The value is case-agnostic and will be converted to lower-case.
	Tags []string `json:"tags"`
}

// RouteAdvertisement lists the subnets that will be advertised so that the Edge Gateway can route out to the
// connected external network.
type RouteAdvertisement struct {
	// Enable if true, means that the subnets will be advertised.
	Enable bool `json:"enable"`
	// Subnets is the list of subnets that will be advertised so that the Edge Gateway can route out to the connected
	// external network.
	Subnets []string `json:"subnets"`
}

// EdgeBgpNeighbor represents a BGP neighbor on the NSX-T Edge Gateway
type EdgeBgpNeighbor struct {
	ID string `json:"id,omitempty"`

	// NeighborAddress holds IP address of the BGP neighbor. Both IPv4 and IPv6 formats are supported.
	//
	// Note. Uniqueness is enforced by NeighborAddress
	NeighborAddress string `json:"neighborAddress"`

	// RemoteASNumber specified Autonomous System (AS) number of a BGP neighbor in ASPLAIN format.
	RemoteASNumber string `json:"remoteASNumber"`

	// KeepAliveTimer specifies the time interval (in seconds) between keep alive messages sent to
	// peer.
	KeepAliveTimer int `json:"keepAliveTimer,omitempty"`

	// HoldDownTimer specifies the time interval (in seconds) before declaring a peer dead.
	HoldDownTimer int `json:"holdDownTimer,omitempty"`

	// NeighborPassword for BGP neighbor authentication. Empty string ("") clears existing password.
	// Not specifying a value will be treated as "no password".
	NeighborPassword string `json:"neighborPassword"`

	// AllowASIn is a flag indicating whether BGP neighbors can receive routes with same AS.
	AllowASIn bool `json:"allowASIn,omitempty"`

	// GracefulRestartMode Describes Graceful Restart configuration Modes for BGP configuration on
	// an Edge Gateway.
	//
	// Possible values are: DISABLE , HELPER_ONLY , GRACEFUL_AND_HELPER
	// * DISABLE - Both graceful restart and helper modes are disabled.
	// * HELPER_ONLY - Only helper mode is enabled. (ability for a BGP speaker to indicate its ability to preserve
	//   forwarding state during BGP restart
	// * GRACEFUL_AND_HELPER - Both graceful restart and helper modes are enabled.  Ability of a BGP
	//	 speaker to advertise its restart to its peers.
	GracefulRestartMode string `json:"gracefulRestartMode,omitempty"`

	// IpAddressTypeFiltering specifies IP address type based filtering in each direction. Setting
	// the value to "DISABLED" will disable address family based filtering.
	//
	// Possible values are: IPV4 , IPV6 , DISABLED
	IpAddressTypeFiltering string `json:"ipAddressTypeFiltering,omitempty"`

	// InRoutesFilterRef specifies route filtering configuration for the BGP neighbor in 'IN'
	// direction. It is the reference to the prefix list, indicating which routes to filter for IN
	// direction. Not specifying a value will be treated as "no IN route filters".
	InRoutesFilterRef *OpenApiReference `json:"inRoutesFilterRef,omitempty"`

	// OutRoutesFilterRef specifies route filtering configuration for the BGP neighbor in 'OUT'
	// direction. It is the reference to the prefix list, indicating which routes to filter for OUT
	// direction. Not specifying a value will be treated as "no OUT route filters".
	OutRoutesFilterRef *OpenApiReference `json:"outRoutesFilterRef,omitempty"`

	// Specifies the BFD (Bidirectional Forwarding Detection) configuration for failure detection. Not specifying a value
	// results in default behavior.
	Bfd *EdgeBgpNeighborBfd `json:"bfd,omitempty"`
}

// EdgeBgpNeighborBfd describes BFD (Bidirectional Forwarding Detection) configuration for failure detection.
type EdgeBgpNeighborBfd struct {
	// A flag indicating whether BFD configuration is enabled or not.
	Enabled bool `json:"enabled"`

	// BfdInterval specifies the time interval (in milliseconds) between heartbeat packets.
	BfdInterval int `json:"bfdInterval,omitempty"`

	// DeclareDeadMultiple specifies number of times heartbeat packet is missed before BFD declares
	// that the neighbor is down.
	DeclareDeadMultiple int `json:"declareDeadMultiple,omitempty"`
	// EdgeBgpIpPrefixList holds BGP IP Prefix List configuration for NSX-T Edge Gateways

}

type EdgeBgpIpPrefixList struct {
	// ID is the unique identifier of the entity in URN format.
	ID string `json:"id,omitempty"`

	// Name of the entity
	Name string `json:"name"`

	// Description of the entity
	Description string `json:"description,omitempty"`

	// Prefixes is the list of prefixes that will be advertised so that the Edge Gateway can route out to the
	// connected external network.
	Prefixes []EdgeBgpConfigPrefixListPrefixes `json:"prefixes,omitempty"`
}

// EdgeBgpConfigPrefixListPrefixes is a list of prefixes that will be advertised so that the Edge Gateway can route out to the
// connected external network.
type EdgeBgpConfigPrefixListPrefixes struct {
	// Network is the network address of the prefix
	Network string `json:"network,omitempty"`

	// Action is the action to be taken on the prefix. Can be 'PERMIT' or 'DENY'
	Action string `json:"action,omitempty"`

	// GreateerThan is the the value which the prefix length must be greater than or equal to. Must
	// be less than or equal to 'LessThanEqualTo'
	GreaterThanEqualTo int `json:"greaterThanEqualTo,omitempty"`

	// The value which the prefix length must be less than or equal to. Must be greater than or
	// equal to 'GreaterThanEqualTo'
	LessThanEqualTo int `json:"lessThanEqualTo,omitempty"`
}

// EdgeBgpConfig defines BGP configuration on NSX-T Edge Gateways (Tier1 NSX-T Gateways)
type EdgeBgpConfig struct {
	// A flag indicating whether BGP configuration is enabled or not.
	Enabled bool `json:"enabled"`

	// Ecmp A flag indicating whether ECMP is enabled or not.
	Ecmp bool `json:"ecmp"`

	// BGP AS (Autonomous system) number to advertise to BGP peers. BGP AS number can be specified
	// in either ASPLAIN or ASDOT formats, like ASPLAIN format :- '65546', ASDOT format :- '1.10'.
	//
	// Read only if using a VRF-Lite backed external network.
	LocalASNumber string `json:"localASNumber,omitempty"`

	// BGP Graceful Restart configuration. Not specifying a value results in default bahavior.
	//
	// Read only if using a VRF-Lite backed external network.
	GracefulRestart *EdgeBgpGracefulRestartConfig `json:"gracefulRestart,omitempty"`

	// This property describes the current version of the entity. To prevent clients from
	// overwriting each other's changes, update operations must include the version which can be
	// obtained by issuing a GET operation. If the version number on an update call is missing, the
	// operation will be rejected. This is only needed on update calls.
	Version EdgeBgpConfigVersion `json:"version"`
}

// EdgeBgpGracefulRestartConfig describes current graceful restart configuration mode and timer for
// BGP configuration on an edge gateway.
type EdgeBgpGracefulRestartConfig struct {
	// Mode describes Graceful Restart configuration Modes for BGP configuration on an edge gateway.
	// HELPER_ONLY mode is the ability for a BGP speaker to indicate its ability to preserve
	// forwarding state during BGP restart. GRACEFUL_RESTART mode is the ability of a BGP speaker to
	// advertise its restart to its peers.
	//
	// DISABLE - Both graceful restart and helper modes are disabled.
	// HELPER_ONLY - Only helper mode is enabled.
	// GRACEFUL_AND_HELPER - Both graceful restart and helper modes are enabled.
	//
	// Possible values are: DISABLE , HELPER_ONLY , GRACEFUL_AND_HELPER
	Mode string `json:"mode"`

	// RestartTimer specifies maximum time taken (in seconds) for a BGP session to be established
	// after a restart. If the session is not re-established within this timer, the receiving
	// speaker will delete all the stale routes from that peer.
	RestartTimer int `json:"restartTimer"`

	// StaleRouteTimer defines maximum time (in seconds) before stale routes are removed when BGP
	// restarts.
	StaleRouteTimer int `json:"staleRouteTimer"`
}

// EdgeBgpConfigVersion is part of EdgeBgpConfig type and describes current version of the entity
// being modified
type EdgeBgpConfigVersion struct {
	Version int `json:"version"`
}
