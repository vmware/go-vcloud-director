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

// NsxtIpSecVpnTunnel specifies the IPsec VPN tunnel configuration
type NsxtIpSecVpnTunnel struct {
	// ID unique for IPsec VPN tunnel. On updates, the id is required for the tunnel, while for create a new id will be
	// generated.
	ID string `json:"id,omitempty"`
	// Name for the tunnel
	Name string `json:"name"`
	// Description for the tunnel
	Description string `json:"description,omitempty"`
	// Enabled describes whether the tunnel is enabled or not. The default is true.
	Enabled bool `json:"enabled"`
	// LocalEndpoint which corresponds to the Edge Gateway the tunnel is being configured on. Local Endpoint requires an
	// IP. That IP must be suballocated to the edge gateway
	LocalEndpoint NsxtIpSecVpnTunnelLocalEndpoint `json:"localEndpoint"`
	// RemoteEndpoint corresponds to the device on the remote site terminating the VPN tunnel
	RemoteEndpoint NsxtIpSecVpnTunnelRemoteEndpoint `json:"remoteEndpoint"`
	// PreSharedKey is key used for authentication. It must be the same on the other end of IPsec VPN tunnel
	PreSharedKey string `json:"preSharedKey"`
	// SecurityType is the security type used for the IPsec Tunnel. If nothing is specified, this will be set to
	// ‘DEFAULT’ in which the default settings in NSX will be used. For custom settings, one should use the
	// connectionProperties endpoint to specify custom settings. The security type will then appropriately reflect
	// itself as ‘CUSTOM’.
	SecurityType string `json:"securityType,omitempty"`
	// Logging sets whether logging for the tunnel is enabled or not. The default is false.
	Logging bool `json:"logging"`

	// AuthenticationMode is authentication mode this IPsec tunnel will use to authenticate with the peer endpoint. The
	// default is a pre-shared key (PSK).
	// * PSK - A known key is shared between each site before the tunnel is established.
	// * CERTIFICATE ? Incoming connections are required to present an identifying digital certificate, which VCD verifies
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

	// Version of IPsec configuration. Must not be set when creating.
	Version *struct {
		// Version is incremented after each update
		Version *int `json:"version,omitempty"`
	} `json:"version,omitempty"`
}

// NsxtIpSecVpnTunnelLocalEndpoint which corresponds to the Edge Gateway the tunnel is being configured on. Local
// Endpoint requires an IP. That IP must be suballocated to the edge gateway
type NsxtIpSecVpnTunnelLocalEndpoint struct {
	// LocalId is the optional local identifier for the endpoint
	LocalId string `json:"localId,omitempty"`
	// LocalAddress is the IPv4 Address for the endpoint. This has to be a suballocated IP on the Edge Gateway. This is
	// required
	LocalAddress string `json:"localAddress"`
	// LocalNetworks is the list of local networks. These must be specified in normal Network CIDR format. At least one is
	// required
	LocalNetworks []string `json:"localNetworks,omitempty"`
}

// NsxtIpSecVpnTunnelRemoteEndpoint corresponds to the device on the remote site terminating the VPN tunnel
type NsxtIpSecVpnTunnelRemoteEndpoint struct {
	// RemoteId is This Remote ID is needed to uniquely identify the peer site. If this tunnel is using PSK authentication,
	// the Remote ID is the public IP Address of the remote device terminating the VPN Tunnel. When NAT is configured on
	// the Remote ID, enter the private IP Address of the Remote Site. If the remote ID is not set, VCD will set the
	// remote id to the remote address. If this tunnel is using certificate authentication, enter the distinguished
	// name of the certificate used to secure the
	// remote endpoint (for example, C=US,ST=Massachusetts,O=VMware,OU=VCD,CN=Edge1). The remote id must be provided in
	// this case
	RemoteId string `json:"remoteId,omitempty"`
	// RemoteAddress is IPv4 Address of the remote endpoint on the remote site. This is the Public IPv4 Address of the
	// remote device terminating the VPN connection. This is required
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

// NsxtIpSecVpnTunnelSecurityProfile specifies the given security profile/connection properties of a given IP Sec VPN Tunnel,
// such as Dead Probe Interval and IKE settings. If a security type is set to 'CUSTOM', then ike, tunnel, and/or dpd
// configurations can be specified. Otherwise, those fields are read only and are set to the values based on the
// specific security type.
type NsxtIpSecVpnTunnelSecurityProfile struct {
	// SecurityType is the security type used for the IPSec Tunnel. If nothing is specified, this will be set to ‘DEFAULT’
	// in which the default settings in NSX will be used. If ‘CUSTOM’ is specified, then IKE, Tunnel, and DPD
	// configurations can be set.
	// To "RESET" configuration to DEFAULT, the NsxtIpSecVpnTunnel.SecurityType field should be changed instead of this
	SecurityType string `json:"securityType"`
	// IkeConfiguration is the IKE Configuration to be used for the tunnel. If nothing is explicitly set, the system
	// defaults will be used.
	IkeConfiguration NsxtIpSecVpnTunnelProfileIkeConfiguration `json:"ikeConfiguration"`
	// TunnelConfiguration contains parameters such as encryption algorithm to be used. If nothing is explicitly set,
	// the system defaults will be used.
	TunnelConfiguration NsxtIpSecVpnTunnelProfileTunnelConfiguration `json:"tunnelConfiguration"`
	// DpdConfiguration contains Dead Peer Detection configuration. If nothing is explicitly set, the system defaults
	// will be used.
	DpdConfiguration NsxtIpSecVpnTunnelProfileDpdConfiguration `json:"dpdConfiguration"`
}

// NsxtIpSecVpnTunnelProfileIkeConfiguration is the Internet Key Exchange (IKE) profiles provide information about the
// algorithms that are used to authenticate, encrypt, and establish a shared secret between network sites when you
// establish an IKE tunnel.
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
	// Note. Only one value can be set inside the slice
	EncryptionAlgorithms []string `json:"encryptionAlgorithms"`
	// DigestAlgorithms contains list of Digest algorithms - secure hashing algorithms to use during the IKE negotiation.
	//
	// Default is SHA2_256.
	//
	// Possible values are: SHA1 , SHA2_256 , SHA2_384 , SHA2_512
	// Note. Only one value can be set inside the slice
	DigestAlgorithms []string `json:"digestAlgorithms"`
	// DhGroups contains list of Diffie-Hellman groups to be used if Perfect Forward Secrecy is enabled. These are
	// cryptography schemes that allows the peer site and the edge gateway to establish a shared secret over an insecure
	// communications channel
	//
	// Default is GROUP14.
	//
	// Possible values are: GROUP2, GROUP5, GROUP14, GROUP15, GROUP16, GROUP19, GROUP20, GROUP21
	// Note. Only one value can be set inside the slice
	DhGroups []string `json:"dhGroups"`
	// SaLifeTime is the Security Association life time in seconds. It is number of seconds before the IPsec tunnel needs
	// to reestablish
	//
	// Default is 86400 seconds (1 day).
	SaLifeTime int `json:"saLifeTime"`
}

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
	// Note. Only one value can be set inside the slice
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
	SaLifeTime int `json:"saLifeTime"`
}

// NsxtIpSecVpnTunnelProfileDpdConfiguration specifies the Dead Peer Detection Profile. This configurations determines
// the number of seconds to wait in time between probes to detect if an IPSec peer is alive or not. The default value
// for the DPD probe interval is 60 seconds.
type NsxtIpSecVpnTunnelProfileDpdConfiguration struct {
	// ProbeInternal is value of the probe interval in seconds. This defines a periodic interval for DPD probes. The
	// minimum is 3 seconds and the maximum is 60 seconds.
	ProbeInterval int `json:"probeInterval"`
}
