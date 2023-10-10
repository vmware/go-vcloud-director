package types

// NsxtSegmentProfileTemplate allows management of templates that define the segment profiles that
// will be applied during network creation.
type NsxtSegmentProfileTemplate struct {
	ID string `json:"id,omitempty"`
	// Name for Segment Profile template
	Name string `json:"name"`
	// Description for Segment Profile template
	Description string `json:"description,omitempty"`

	// SourceNsxTManagerRef points to NSX-T manager providing the source segment profiles
	SourceNsxTManagerRef   *OpenApiReference `json:"sourceNsxTManagerRef,omitempty"`
	IPDiscoveryProfile     *Reference        `json:"ipDiscoveryProfile,omitempty"`
	MacDiscoveryProfile    *Reference        `json:"macDiscoveryProfile,omitempty"`
	QosProfile             *Reference        `json:"qosProfile,omitempty"`
	SegmentSecurityProfile *Reference        `json:"segmentSecurityProfile,omitempty"`
	SpoofGuardProfile      *Reference        `json:"spoofGuardProfile,omitempty"`

	LastModified string `json:"lastModified,omitempty"`
}

// NsxtSegmentProfileCommonFields contains common fields that are used in all NSX-T Segment
// Profiles
type NsxtSegmentProfileCommonFields struct {
	ID string `json:"id,omitempty"`
	// Description of the segment profile.
	Description string `json:"description,omitempty"`
	// DisplayName represents name of the segment profile. This corresponds to the name used in
	// NSX-T managers logs or GUI.
	DisplayName string `json:"displayName"`
	// NsxTManagerRef where this segment profile is configured.
	NsxTManagerRef *OpenApiReference `json:"nsxTManagerRef"`
}

// NsxtSegmentProfileIpDiscovery contains information about NSX-T IP Discovery Segment Profile
// It is a read-only construct in VCD
type NsxtSegmentProfileIpDiscovery struct {
	NsxtSegmentProfileCommonFields
	// ArpBindingLimit indicates the number of arp snooped IP addresses to be remembered per
	// LogicalPort.
	ArpBindingLimit int `json:"arpBindingLimit"`
	// ArpNdBindingTimeout indicates ARP and ND cache timeout (in minutes).
	ArpNdBindingTimeout int `json:"arpNdBindingTimeout"`
	// IsArpSnoopingEnabled defines whether ARP snooping is enabled.
	IsArpSnoopingEnabled bool `json:"isArpSnoopingEnabled"`
	// IsDhcpSnoopingV4Enabled defines whether DHCP snooping for IPv4 is enabled.
	IsDhcpSnoopingV4Enabled bool `json:"isDhcpSnoopingV4Enabled"`
	// IsDhcpSnoopingV6Enabled defines whether DHCP snooping for IPv6 is enabled.
	IsDhcpSnoopingV6Enabled bool `json:"isDhcpSnoopingV6Enabled"`
	// IsDuplicateIPDetectionEnabled indicates whether duplicate IP detection is enabled. Duplicate
	// IP detection is used to determine if there is any IP conflict with any other port on the same
	// logical switch. If a conflict is detected, then the IP is marked as a duplicate on the port
	// where the IP was discovered last.
	IsDuplicateIPDetectionEnabled bool `json:"isDuplicateIpDetectionEnabled"`
	// IsNdSnoopingEnabled indicates whether ND snooping is enabled. If true, this method will snoop
	// the NS (Neighbor Solicitation) and NA (Neighbor Advertisement) messages in the ND (Neighbor
	// Discovery Protocol) family of messages which are transmitted by a VM. From the NS messages,
	// we will learn about the source which sent this NS message. From the NA message, we will learn
	// the resolved address in the message which the VM is a recipient of. Addresses snooped by this
	// method are subject to TOFU.
	IsNdSnoopingEnabled bool `json:"isNdSnoopingEnabled"`
	// IsTofuEnabled defines whether 'Trust on First Use(TOFU)' paradigm is enabled.
	IsTofuEnabled bool `json:"isTofuEnabled"`
	// IsVMToolsV4Enabled indicates whether fetching IPv4 address using vm-tools is enabled. This
	// option is only supported on ESX where vm-tools is installed.
	IsVMToolsV4Enabled bool `json:"isVmToolsV4Enabled"`
	// IsVMToolsV6Enabled indicates whether fetching IPv6 address using vm-tools is enabled. This
	// will learn the IPv6 addresses which are configured on interfaces of a VM with the help of the
	// VMTools software.
	IsVMToolsV6Enabled bool `json:"isVmToolsV6Enabled"`
	// NdSnoopingLimit defines maximum number of ND (Neighbor Discovery Protocol) snooped IPv6
	// addresses.
	NdSnoopingLimit int `json:"ndSnoopingLimit"`
}

// NsxtSegmentProfileMacDiscovery contains information about NSX-T MAC Discovery Segment Profile
// It is a read-only construct in VCD
type NsxtSegmentProfileMacDiscovery struct {
	NsxtSegmentProfileCommonFields
	// IsMacChangeEnabled indcates whether source MAC address change is enabled.
	IsMacChangeEnabled bool `json:"isMacChangeEnabled"`
	// IsMacLearningEnabled indicates whether source MAC address learning is enabled.
	IsMacLearningEnabled bool `json:"isMacLearningEnabled"`
	// IsUnknownUnicastFloodingEnabled indicates whether unknown unicast flooding rule is enabled.
	// This allows flooding for unlearned MAC for ingress traffic.
	IsUnknownUnicastFloodingEnabled bool `json:"isUnknownUnicastFloodingEnabled"`
	// MacLearningAgingTime indicates aging time in seconds for learned MAC address. Indicates how
	// long learned MAC address remain.
	MacLearningAgingTime int `json:"macLearningAgingTime"`
	// MacLimit indicates the maximum number of MAC addresses that can be learned on this port.
	MacLimit int `json:"macLimit"`
	// MacPolicy defines the policy after MAC Limit is exceeded. It can be either 'ALLOW' or 'DROP'.
	MacPolicy string `json:"macPolicy"`
}

// NsxtSegmentProfileSegmentSpoofGuard contains information about NSX-T Spoof Guard Segment Profile
// It is a read-only construct in VCD
type NsxtSegmentProfileSegmentSpoofGuard struct {
	NsxtSegmentProfileCommonFields
	// IsAddressBindingWhitelistEnabled indicates whether Spoof Guard is enabled. If true, it only
	// allows VM sending traffic with the IPs in the whitelist.
	IsAddressBindingWhitelistEnabled bool `json:"isAddressBindingWhitelistEnabled"`
}

// NsxtSegmentProfileSegmentQosProfile contains information about NSX-T QoS Segment Profile
// It is a read-only construct in VCD
type NsxtSegmentProfileSegmentQosProfile struct {
	NsxtSegmentProfileCommonFields
	// ClassOfService groups similar types of traffic in the network and each type of traffic is
	// treated as a class with its own level of service priority. The lower priority traffic is
	// slowed down or in some cases dropped to provide better throughput for higher priority
	// traffic.
	ClassOfService int `json:"classOfService"`
	// DscpConfig contains a Differentiated Services Code Point (DSCP) Configuration for this
	// Segment QoS Profile.
	DscpConfig struct {
		Priority  int    `json:"priority"`
		TrustMode string `json:"trustMode"`
	} `json:"dscpConfig"`
	// EgressRateLimiter indicates egress rate properties in Mb/s.
	EgressRateLimiter NsxtSegmentProfileSegmentQosProfileRateLimiter `json:"egressRateLimiter"`
	// IngressBroadcastRateLimiter indicates broadcast rate properties in Mb/s.
	IngressBroadcastRateLimiter NsxtSegmentProfileSegmentQosProfileRateLimiter `json:"ingressBroadcastRateLimiter"`
	// IngressRateLimiter indicates ingress rate properties in Mb/s.
	IngressRateLimiter NsxtSegmentProfileSegmentQosProfileRateLimiter `json:"ingressRateLimiter"`
}

// NsxtSegmentProfileIpDiscovery contains information about NSX-T IP Discovery Segment Profile
// It is a read-only construct in VCD
type NsxtSegmentProfileSegmentQosProfileRateLimiter struct {
	// Average bandwidth in Mb/s.
	AvgBandwidth int `json:"avgBandwidth"`
	// Burst size in bytes.
	BurstSize int `json:"burstSize"`
	// Peak bandwidth in Mb/s.
	PeakBandwidth int `json:"peakBandwidth"`
}

// NsxtSegmentProfileSegmentSecurity contains information about NSX-T Segment Security Profile
// It is a read-only construct in VCD
type NsxtSegmentProfileSegmentSecurity struct {
	NsxtSegmentProfileCommonFields
	// BpduFilterAllowList indicates pre-defined list of allowed MAC addresses to be excluded from
	// BPDU filtering.
	BpduFilterAllowList []string `json:"bpduFilterAllowList"`
	// IsBpduFilterEnabled indicates whether BPDU filter is enabled.
	IsBpduFilterEnabled bool `json:"isBpduFilterEnabled"`
	// IsDhcpClientBlockV4Enabled indicates whether DHCP Client block IPv4 is enabled. This filters
	// DHCP Client IPv4 traffic.
	IsDhcpClientBlockV4Enabled bool `json:"isDhcpClientBlockV4Enabled"`
	// IsDhcpClientBlockV6Enabled indicates whether DHCP Client block IPv6 is enabled. This filters
	// DHCP Client IPv4 traffic.
	IsDhcpClientBlockV6Enabled bool `json:"isDhcpClientBlockV6Enabled"`
	// IsDhcpServerBlockV4Enabled indicates whether DHCP Server block IPv4 is enabled. This filters
	// DHCP Server IPv4 traffic.
	IsDhcpServerBlockV4Enabled bool `json:"isDhcpServerBlockV4Enabled"`
	// IsDhcpServerBlockV6Enabled indicates whether DHCP Server block IPv6 is enabled. This filters
	// DHCP Server IPv6 traffic.
	IsDhcpServerBlockV6Enabled bool `json:"isDhcpServerBlockV6Enabled"`
	// IsNonIPTrafficBlockEnabled indicates whether non IP traffic block is enabled. If true, it
	// blocks all traffic except IP/(G)ARP/BPDU.
	IsNonIPTrafficBlockEnabled bool `json:"isNonIpTrafficBlockEnabled"`
	// IsRaGuardEnabled indicates whether Router Advertisement Guard is enabled. This filters DHCP
	// Server IPv6 traffic.
	IsRaGuardEnabled bool `json:"isRaGuardEnabled"`
	// IsRateLimitingEnabled indicates whether Rate Limiting is enabled.
	IsRateLimitingEnabled bool `json:"isRateLimitingEnabled"`
	RateLimits            struct {
		// Incoming broadcast traffic limit in packets per second.
		RxBroadcast int `json:"rxBroadcast"`
		// Incoming multicast traffic limit in packets per second.
		RxMulticast int `json:"rxMulticast"`
		// Outgoing broadcast traffic limit in packets per second.
		TxBroadcast int `json:"txBroadcast"`
		// Outgoing multicast traffic limit in packets per second.
		TxMulticast int `json:"txMulticast"`
	} `json:"rateLimits"`
}

// NsxtGlobalDefaultSegmentProfileTemplate is a structure that sets VCD global default Segment
// Profile Templates
type NsxtGlobalDefaultSegmentProfileTemplate struct {
	VappNetworkSegmentProfileTemplateRef *OpenApiReference `json:"vappNetworkSegmentProfileTemplateRef,omitempty"`
	VdcNetworkSegmentProfileTemplateRef  *OpenApiReference `json:"vdcNetworkSegmentProfileTemplateRef,omitempty"`
}

// OrgVdcNetworkSegmentProfiles defines Segment Profile configuration structure for Org VDC networks
// An Org VDC network may have a Segment Profile Template assigned, or individual Segment Profiles
type OrgVdcNetworkSegmentProfiles struct {
	// SegmentProfileTemplate contains a read-only reference to Segment Profile Template
	// To update Segment Profile Template for a particular Org VDC network, one must use
	// `OpenApiOrgVdcNetwork.SegmentProfileTemplate` field and `OpenApiOrgVdcNetwork.Update()`
	SegmentProfileTemplate *SegmentProfileTemplateRef `json:"segmentProfileTemplate,omitempty"`

	IPDiscoveryProfile     *Reference `json:"ipDiscoveryProfile"`
	MacDiscoveryProfile    *Reference `json:"macDiscoveryProfile"`
	QosProfile             *Reference `json:"qosProfile"`
	SegmentSecurityProfile *Reference `json:"segmentSecurityProfile"`
	SpoofGuardProfile      *Reference `json:"spoofGuardProfile"`
}

// SegmentProfileTemplateRef contains reference to segment profile
type SegmentProfileTemplateRef struct {
	Source      string            `json:"source"`
	TemplateRef *OpenApiReference `json:"templateRef"`
}
