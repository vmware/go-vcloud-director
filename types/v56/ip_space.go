/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

// IpSpace provides structured approach to allocating public and private IP addresses by preventing
// the use of overlapping IP addresses across organizations and organization VDCs.
//
// An IP space consists of a set of defined non-overlapping IP ranges and small CIDR blocks that are
// reserved and used during the consumption aspect of the IP space life cycle. An IP space can be
// either IPv4 or IPv6, but not both.
//
// Every IP space has an internal scope and an external scope. The internal scope of an IP space is
// a list of CIDR notations that defines the exact span of IP addresses in which all ranges and
// blocks must be contained in. The external scope defines the total span of IP addresses to which
// the IP space has access, for example the internet or a WAN.
type IpSpace struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// Type is The type of the IP Space. Possible values are:
	// * PUBLIC - These can be consumed by multiple organizations. These are created by System
	// Administrators only, for managing public IPs. The IP addresses and IP Prefixes from this IP
	// space are allocated to specific organizations for consumption.
	// * PRIVATE - These can be consumed by only a single organization. All the IPs within this IP
	// Space are allocated to the particular organization.
	// * SHARED_SERVICES - These are for internal use only. The IP addresses and IP Prefixes from
	// this IP space can be consumed by multiple organizations but those IP addresses and IP
	// Prefixes will not be not visible to the individual users within the organization. These are
	// created by System Administrators only, typically for a service or for management networks.
	//
	// Note. This project contains convenience constants for defining IP Space
	// types`types.IpSpaceShared`, `types.IpSpacePublic`, `types.IpSpacePrivate`
	//
	// Only SHARED_SERVICES type can be changed to PUBLIC type. No other type changes are allowed.
	Type string `json:"type"`

	// The organization this IP Space belongs to. This property is only applicable and is required
	// for IP Spaces with type PRIVATE.
	OrgRef *OpenApiReference `json:"orgRef,omitempty"`

	// Utilization summary for this IP space.
	Utilization IpSpaceUtilization `json:"utilization,omitempty"`

	// List of IP Prefixes.
	IPSpacePrefixes []IPSpacePrefixes `json:"ipSpacePrefixes"`

	// List of IP Ranges. These are logically treated as a single block of IPs for allocation purpose.
	IPSpaceRanges IPSpaceRanges `json:"ipSpaceRanges"`

	// This defines the exact span of IP addresses in a CIDR format within which all IP Ranges and
	// IP Prefixes of this IP Space must be contained. This typically defines the span of IP
	// addresses used within this Data Center.
	IPSpaceInternalScope []string `json:"ipSpaceInternalScope"`

	// This defines the total span of IP addresses in a CIDR format within which all IP Ranges and
	// IP Prefixes of this IP Space must be contained. This is used by the system for creation of
	// NAT rules and BGP prefixes. This typically defines the span of IP addresses outside the
	// bounds of this Data Center. For the internet this may be 0.0.0.0/0. For a WAN, this could be
	// 10.0.0.0/8.
	IPSpaceExternalScope string `json:"ipSpaceExternalScope,omitempty"`

	// Whether the route advertisement is enabled for this IP Space or not. If true, the routed Org
	// VDC networks which are configured from this IP Space will be advertised from the connected
	// Edge Gateway to the Provider Gateway. Route advertisement must be enabled on a particular
	// network for it to be advertised. Networks from the PRIVATE IP Spaces will only be advertised
	// if the associated Provider Gateway is owned by the Organization.
	RouteAdvertisementEnabled bool `json:"routeAdvertisementEnabled"`

	// Status is one of `PENDING`,   `CONFIGURING`,   `REALIZED`,   `REALIZATION_FAILED`,   `UNKNOWN`
	Status string `json:"status,omitempty"`
}

type FloatingIPs struct {
	// TotalCount holds the number of IP addresses or IP Prefixes defined by the IP Space. If user
	// does not own this IP Space, this is the quota that the user's organization is granted. A '-1'
	// value means that the user's organization has no cap on the quota (for this case,
	// allocatedPercentage is unset)
	TotalCount string `json:"totalCount,omitempty"`
	// AllocatedCount holds the number of allocated IP addresses or IP Prefixes.
	AllocatedCount string `json:"allocatedCount,omitempty"`
	// UsedCount holds the number of used IP addresses or IP Prefixes. An allocated IP address or IP
	// Prefix is considered used if it is being used in network services such as NAT rule or in Org
	// VDC network definition.
	UsedCount string `json:"usedCount,omitempty"`
	// UnusedCount holds the number of unused IP addresses or IP Prefixes. An IP address or an IP
	// Prefix is considered unused if it is allocated but not being used by any network service or
	// any Org vDC network definition.
	UnusedCount string `json:"unusedCount,omitempty"`
	// AllocatedPercentage specifies the percentage of allocated IP addresses or IP Prefixes out of
	// all defined IP addresses or IP Prefixes.
	AllocatedPercentage float32 `json:"allocatedPercentage,omitempty"`
	// UsedPercentage specifies the percentage of used IP addresses or IP Prefixes out of total
	// allocated IP addresses or IP Prefixes.
	UsedPercentage float32 `json:"usedPercentage,omitempty"`
}

type PrefixLengthUtilizations struct {
	PrefixLength int `json:"prefixLength"`
	// TotalCount contains total number of IP Prefixes. If user does not own this IP Space, this is
	// the quota that the user's organization is granted. A '-1' value means that the user's
	// organization has no cap on the quota.
	TotalCount int `json:"totalCount"`
	// AllocatedCount contains the number of allocated IP prefixes.
	AllocatedCount int `json:"allocatedCount"`
}

type IPPrefixes struct {
	// TotalCount holds the number of IP addresses or IP Prefixes defined by the IP Space. If user
	// does not own this IP Space, this is the quota that the user's organization is granted. A '-1'
	// value means that the user's organization has no cap on the quota; for this case,
	// allocatedPercentage is unset.
	TotalCount string `json:"totalCount,omitempty"`
	// TAllocatedCount holds the number of allocated IP addresses or IP Prefixes.
	AllocatedCount string `json:"allocatedCount,omitempty"`
	// UsedCount holds the number of used IP addresses or IP Prefixes. An allocated IP address or IP
	// Prefix is considered used if it is being used in network services such as NAT rule or in Org
	// VDC network definition.
	UsedCount string `json:"usedCount,omitempty"`
	// UnusedCount holds the number of unused IP addresses or IP Prefixes. An IP address or an IP
	// Prefix is considered unused if it is allocated but not being used by any network service or
	// any Org vDC network definition.
	UnusedCount string `json:"unusedCount,omitempty"`
	// AllocatedPercentage specifies the percentage of allocated IP addresses or IP Prefixes out of
	// all defined IP addresses or IP Prefixes.
	AllocatedPercentage float32 `json:"allocatedPercentage,omitempty"`
	// UsedPercentage specifies the percentage of used IP addresses or IP Prefixes out of total
	// allocated IP addresses or IP Prefixes.
	UsedPercentage float32 `json:"usedPercentage,omitempty"`
	// PrefixLengthUtilizations contains utilization summary grouped by IP Prefix's prefix length.
	// This information will only be returned for an individual IP Prefix.
	PrefixLengthUtilizations []PrefixLengthUtilizations `json:"prefixLengthUtilizations,omitempty"`
}

type IpSpaceUtilization struct {
	// FloatingIPs holds utilization summary for floating IPs within the IP space.
	FloatingIPs FloatingIPs `json:"floatingIPs,omitempty"`
	// IPPrefixes holds Utilization summary for IP prefixes within the IP space.
	IPPrefixes IPPrefixes `json:"ipPrefixes,omitempty"`
}

type IPSpaceRanges struct {
	IPRanges []IpSpaceRangeValues `json:"ipRanges"`
	// This specifies the default number of IPs from the specified ranges which can be consumed by
	// each organization using this IP Space. This is typically set for IP Spaces with type PUBLIC
	// or SHARED_SERVICES. A Quota of -1 means there is no cap to the number of IP addresses that
	// can be allocated. A Quota of 0 means that the IP addresses cannot be allocated. If not
	// specified, all PUBLIC or SHARED_SERVICES IP Spaces have a default quota of 1 for Floating IP
	// addresses and all PRIVATE IP Spaces have a default quota of -1 for Floating IP addresses.
	DefaultFloatingIPQuota int `json:"defaultFloatingIpQuota"`
}

type IpSpaceRangeValues struct {
	ID string `json:"id,omitempty"`
	// Starting IP address in the range.
	StartIPAddress string `json:"startIpAddress"`
	// endIpAddress
	EndIPAddress string `json:"endIpAddress"`

	// The number of IP addresses defined by the IP range.
	TotalIPCount string `json:"totalIpCount,omitempty"`
	// The number of allocated IP addresses.
	AllocatedIPCount string `json:"allocatedIpCount,omitempty"`
	// allocatedIpPercentage
	AllocatedIPPercentage float32 `json:"allocatedIpPercentage,omitempty"`
}

type IPSpacePrefixes struct {
	// IPPrefixSequence A sequence of IP prefixes with same prefix length. All the IP Prefix
	// sequences with the same prefix length are treated as one logical unit for allocation purpose.
	IPPrefixSequence []IPPrefixSequence `json:"ipPrefixSequence"`

	// This specifies the number of prefixes from the specified sequence which can be consumed by
	// each organization using this IP Space. All the IP Prefix sequences with the same prefix
	// length are treated as one logical unit for allocation purpose. This is typically set for IP
	// Spaces with type PUBLIC or SHARED_SERVICES. A Quota of -1 means there is no cap to the number
	// of IP Prefixes that can be allocated. A Quota of 0 means that the IP Prefixes cannot be
	// allocated. If not specified, all PUBLIC or SHARED_SERVICES IP Spaces have a default quota of
	// 0 for IP Prefixes and all PRIVATE IP Spaces have a default quota of -1 for IP Prefixes.
	DefaultQuotaForPrefixLength int `json:"defaultQuotaForPrefixLength"`
}

type IPPrefixSequence struct {
	ID string `json:"id,omitempty"`
	// Starting IP address for the IP prefix. Note that if the IP is a host IP and not the network
	// definition IP for the specific prefix length, VCD will automatically modify starting IP to
	// the network definition's IP for the specified host IP. An example is that for prefix length
	// 30, the starting IP of 192.169.0.2 will be automatically modified to 192.169.0.0. 192.169.0.6
	// will be modified to 192.169.0.4. 192.169.0.0/30 and 192.169.0.4/30 are network definition
	// CIDRs for host IPs 192.169.0.2 and 192.169.0.6, respectively.
	StartingPrefixIPAddress string `json:"startingPrefixIpAddress"`
	// The prefix length.
	PrefixLength int `json:"prefixLength"`
	// The number of prefix blocks defined by this IP prefix.
	TotalPrefixCount int `json:"totalPrefixCount"`
	// The number of allocated IP prefix blocks.
	AllocatedPrefixCount int `json:"allocatedPrefixCount,omitempty"`
	// Specifies the percentage of allocated IP prefix blocks out of total specified IP prefix blocks.
	AllocatedPrefixPercentage float32 `json:"allocatedPrefixPercentage,omitempty"`
}
