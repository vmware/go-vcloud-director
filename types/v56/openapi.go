// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/json"
	"fmt"
)

// OpenApiPages unwraps pagination for "Get All" endpoints in OpenAPI. Values kept in json.RawMessage helps to decouple
// marshalling paging related information from exact type related information. Paging can be handled dynamically this
// way while values can be marshaled into exact types.
type OpenApiPages struct {
	// ResultTotal reports total results available
	ResultTotal int `json:"resultTotal,omitempty"`
	// PageCount reports total result pages available
	PageCount int `json:"pageCount,omitempty"`
	// Page reports current page of result
	Page int `json:"page,omitempty"`
	// PageSize reports page size
	PageSize int `json:"pageSize,omitempty"`
	// Associations ...
	Associations interface{} `json:"associations,omitempty"`
	// Values holds types depending on the endpoint therefore `json.RawMessage` is used to dynamically unmarshal into
	// specific type as required
	Values json.RawMessage `json:"values,omitempty"`
}

// OpenApiError helps to marshal and provider meaningful `Error` for
type OpenApiError struct {
	MinorErrorCode string `json:"minorErrorCode"`
	Message        string `json:"message"`
	StackTrace     string `json:"stackTrace"`
}

// Error method implements Go's default `error` interface for CloudAPI errors formats them for human readable output.
func (openApiError OpenApiError) Error() string {
	return fmt.Sprintf("%s - %s", openApiError.MinorErrorCode, openApiError.Message)
}

// ErrorWithStack is the same as `Error()`, but also includes stack trace returned by API which is usually lengthy.
func (openApiError OpenApiError) ErrorWithStack() string {
	return fmt.Sprintf("%s - %s. Stack: %s", openApiError.MinorErrorCode, openApiError.Message,
		openApiError.StackTrace)
}

// Role defines access roles in VCD
type Role struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	BundleKey   string `json:"bundleKey,omitempty"`
	ReadOnly    bool   `json:"readOnly,omitempty"`
}

// NsxtTier0Router defines NSX-T Tier 0 router
type NsxtTier0Router struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
}

// NsxtEdgeCluster is a struct to represent logical grouping of NSX-T Edge virtual machines.
type NsxtEdgeCluster struct {
	// ID contains edge cluster ID (UUID format)
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// NodeCount shows number of nodes in the edge cluster
	NodeCount int `json:"nodeCount"`
	// NodeType usually holds "EDGE_NODE"
	NodeType string `json:"nodeType"`
	// DeploymentType (e.g. "VIRTUAL_MACHINE")
	DeploymentType string `json:"deploymentType"`
}

// ExternalNetworkV2 defines a struct for OpenAPI endpoint which is capable of creating NSX-V or
// NSX-T external network based on provided NetworkBackings.
type ExternalNetworkV2 struct {
	// ID is unique for the network. This field is read-only.
	ID string `json:"id,omitempty"`
	// Name of the network.
	Name string `json:"name"`
	// Description of the network
	Description string `json:"description"`
	// Subnets define one or more subnets and IP allocation pools in edge gateway
	Subnets ExternalNetworkV2Subnets `json:"subnets,omitempty"`
	// NetworkBackings for this external network. Describes if this external network is backed by
	// port groups, vCenter standard switch or an NSX-T Tier-0 router.
	NetworkBackings ExternalNetworkV2Backings `json:"networkBackings"`

	// UsingIpSpace indicates whether the external network is using IP Spaces or not. This field is
	// applicable only to the external networks backed by NSX-T Tier-0 router.
	// This field is only available in VCD 10.4.1+
	UsingIpSpace *bool `json:"usingIpSpace,omitempty"`

	// DedicatedEdgeGateway contains reference to the Edge Gateway that this external network is
	// dedicated to. This is null if this is not a dedicated external network. This field is unset
	// if external network is using IP Spaces.
	DedicatedEdgeGateway *OpenApiReference `json:"dedicatedEdgeGateway,omitempty"`

	// DedicatedOrg specifies the Organization that this external network belongs to. This is unset
	// for the external networks which are available to more than one organization.
	//
	// If this external network is dedicated to an Edge Gateway, this field is read-only and will be
	// set to the Organization of the Edge Gateway.
	//
	// If this external network is using IP Spaces, this field can
	// be used to dedicate this external network to the specified Organization.
	DedicatedOrg *OpenApiReference `json:"dedicatedOrg,omitempty"`

	// NatAndFirewallServiceIntention defines different types of intentions to configure NAT and
	// firewall rules:
	// * PROVIDER_GATEWAY - Allow management of NAT and firewall rules only on Provider Gateways.
	//
	// * EDGE_GATEWAY - Allow management of NAT and firewall rules only on Edge Gateways.
	//
	// * PROVIDER_AND_EDGE_GATEWAY - Allow management of NAT and firewall rules on both the Provider
	// and Edge gateways.
	//
	// This only applies to external networks backed by NSX-T Tier-0 router (i.e. Provider Gateway)
	// and is unset otherwise. Public Provider Gateway supports only EDGE_GATEWAY_ONLY. All other
	// values are ignored. Private Provider Gateway can support all the intentions and if unset, the
	// default is EDGE_GATEWAY.
	//
	// This field requires VCD 10.5.1+ (API 38.1+)
	NatAndFirewallServiceIntention string `json:"natAndFirewallServiceIntention,omitempty"`

	// NetworkRouteAdvertisementIntention configures different types of route advertisement
	// intentions for routed Org VDC network connected to Edge Gateway that is connected to this
	// Provider Gateway. Possible values are:
	//
	// * IP_SPACE_UPLINKS_ADVERTISED_STRICT - All networks within IP Space associated with IP Space
	// Uplink will be advertised by default. This can be changed on an individual network level
	// later, if necessary. All other networks outside of IP Spaces associated with IP Space Uplinks
	// cannot be configured to be advertised.
	//
	// * IP_SPACE_UPLINKS_ADVERTISED_FLEXIBLE - All networks within IP Space associated with IP
	// Space Uplink will be advertised by default. This can be changed on an individual network
	// level later, if necessary. All other networks outside of IP Spaces associated with IP Space
	// Uplinks are not advertised by default but can be configured to be advertised after creation.
	//
	// * ALL_NETWORKS_ADVERTISED - All networks, regardless on whether they fall inside of any IP
	// Spaces associated with IP Space Uplinks, will be advertised by default. This can be changed
	// on an individual network level later, if necessary.
	//
	// This only applies to external networks backed by NSX-T Tier-0 router (i.e. Provider Gateway)
	// and is unset otherwise. Public Provider Gateway supports only
	// IP_SPACE_UPLINKS_ADVERTISED_STRICT. All other values are ignored. Private Provider Gateway
	// can support all the intentions and if unset, the default is also
	// IP_SPACE_UPLINKS_ADVERTISED_STRICT.
	//
	// This field requires VCD 10.5.1+ (API 38.1+)
	NetworkRouteAdvertisementIntention string `json:"networkRouteAdvertisementIntention,omitempty"`

	// TotalIpCount contains the number of IP addresses defined by the static ip pools. If the
	// network contains any IPv6 subnets, the total ip count will be null.
	TotalIpCount *int `json:"totalIpCount,omitempty"`

	// UsedIpCount holds the number of IP address used from the static ip pools.
	UsedIpCount *int `json:"usedIpCount,omitempty"`
}

// OpenApiIPRangeValues defines allocated IP pools for a subnet in external network
type OpenApiIPRangeValues struct {
	// StartAddress holds starting IP address in the range
	StartAddress string `json:"startAddress"`
	// EndAddress holds ending IP address in the range
	EndAddress string `json:"endAddress"`
}

// ExternalNetworkV2IPRanges contains slice of ExternalNetworkV2IPRange
type OpenApiIPRanges struct {
	Values []OpenApiIPRangeValues `json:"values"`
}

// ExternalNetworkV2Subnets contains slice of ExternalNetworkV2Subnet
type ExternalNetworkV2Subnets struct {
	Values []ExternalNetworkV2Subnet `json:"values"`
}

// ExternalNetworkV2Subnet defines one subnet for external network with assigned static IP ranges
type ExternalNetworkV2Subnet struct {
	// Gateway for the subnet
	Gateway string `json:"gateway"`
	// PrefixLength holds prefix length of the subnet
	PrefixLength int `json:"prefixLength"`
	// DNSSuffix is the DNS suffix that VMs attached to this network will use (NSX-V only)
	DNSSuffix string `json:"dnsSuffix"`
	// DNSServer1 - first DNS server that VMs attached to this network will use (NSX-V only)
	DNSServer1 string `json:"dnsServer1"`
	// DNSServer2 - second DNS server that VMs attached to this network will use (NSX-V only)
	DNSServer2 string `json:"dnsServer2"`
	// Enabled indicates whether the external network subnet is currently enabled
	Enabled bool `json:"enabled"`
	// UsedIPCount shows number of IP addresses defined by the static IP ranges
	UsedIPCount int `json:"usedIpCount,omitempty"`
	// TotalIPCount shows number of IP address used from the static IP ranges
	TotalIPCount int `json:"totalIpCount,omitempty"`
	// IPRanges define allocated static IP pools allocated from a defined subnet
	IPRanges ExternalNetworkV2IPRanges `json:"ipRanges"`
}

type ExternalNetworkV2Backings struct {
	Values []ExternalNetworkV2Backing `json:"values"`
}

// ExternalNetworkV2Backing defines which networking subsystem is used for external network (NSX-T or NSX-V)
type ExternalNetworkV2Backing struct {
	// BackingID must contain either Tier-0 router ID for NSX-T or PortGroup ID for NSX-V
	BackingID string `json:"backingId"`
	Name      string `json:"name,omitempty"`
	// BackingType can be either ExternalNetworkBackingTypeNsxtTier0Router in case of NSX-T or one
	// of ExternalNetworkBackingTypeNetwork or ExternalNetworkBackingDvPortgroup in case of NSX-V
	// Deprecated in favor of BackingTypeValue in API V35.0
	BackingType string `json:"backingType,omitempty"`

	// BackingTypeValue replaces BackingType in API V35.0 and adds support for additional network backing type
	// ExternalNetworkBackingTypeNsxtSegment
	BackingTypeValue string `json:"backingTypeValue,omitempty"`
	// NetworkProvider defines backing network manager
	NetworkProvider NetworkProvider `json:"networkProvider"`
}

// NetworkProvider can be NSX-T manager or vCenter. ID is sufficient for creation purpose.
type NetworkProvider struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id"`
}

// VdcComputePolicy contains VDC specific configuration for workloads. (version 1.0.0)
// Deprecated: Use VdcComputePolicyV2 instead (version 2.0.0)
type VdcComputePolicy struct {
	ID                         string   `json:"id,omitempty"`
	Description                *string  `json:"description"` // It's a not-omitempty pointer to be able to send "null" values for empty descriptions.
	Name                       string   `json:"name"`
	CPUSpeed                   *int     `json:"cpuSpeed,omitempty"`
	Memory                     *int     `json:"memory,omitempty"`
	CPUCount                   *int     `json:"cpuCount,omitempty"`
	CoresPerSocket             *int     `json:"coresPerSocket,omitempty"`
	MemoryReservationGuarantee *float64 `json:"memoryReservationGuarantee,omitempty"`
	CPUReservationGuarantee    *float64 `json:"cpuReservationGuarantee,omitempty"`
	CPULimit                   *int     `json:"cpuLimit,omitempty"`
	MemoryLimit                *int     `json:"memoryLimit,omitempty"`
	CPUShares                  *int     `json:"cpuShares,omitempty"`
	MemoryShares               *int     `json:"memoryShares,omitempty"`
	ExtraConfigs               *struct {
		AdditionalProp1 string `json:"additionalProp1,omitempty"`
		AdditionalProp2 string `json:"additionalProp2,omitempty"`
		AdditionalProp3 string `json:"additionalProp3,omitempty"`
	} `json:"extraConfigs,omitempty"`
	PvdcComputePolicyRef     *OpenApiReference   `json:"pvdcComputePolicyRef,omitempty"`
	PvdcComputePolicy        *OpenApiReference   `json:"pvdcComputePolicy,omitempty"`
	CompatibleVdcTypes       []string            `json:"compatibleVdcTypes,omitempty"`
	IsSizingOnly             bool                `json:"isSizingOnly,omitempty"`
	PvdcID                   string              `json:"pvdcId,omitempty"`
	NamedVMGroups            []OpenApiReferences `json:"namedVmGroups,omitempty"`
	LogicalVMGroupReferences OpenApiReferences   `json:"logicalVmGroupReferences,omitempty"`
	IsAutoGenerated          bool                `json:"isAutoGenerated,omitempty"`
}

// VdcComputePolicyV2 contains VDC specific configuration for workloads (version 2.0.0)
// https://developer.vmware.com/apis/vmware-cloud-director/latest/data-structures/VdcComputePolicy2/
type VdcComputePolicyV2 struct {
	VdcComputePolicy
	PolicyType             string                   `json:"policyType"` // Required. Can be "VdcVmPolicy" or "VdcKubernetesPolicy"
	IsVgpuPolicy           bool                     `json:"isVgpuPolicy,omitempty"`
	PvdcNamedVmGroupsMap   []PvdcNamedVmGroupsMap   `json:"pvdcNamedVmGroupsMap,omitempty"`
	PvdcLogicalVmGroupsMap []PvdcLogicalVmGroupsMap `json:"pvdcLogicalVmGroupsMap,omitempty"`
	PvdcVgpuClustersMap    []PvdcVgpuClustersMap    `json:"pvdcVgpuClustersMap,omitempty"`
	VgpuProfiles           []VgpuProfile            `json:"vgpuProfiles,omitempty"`
}

// PvdcNamedVmGroupsMap is a combination of a reference to a Provider VDC and a list of references to Named VM Groups.
// This is used for VM Placement Policies (see VdcComputePolicyV2)
type PvdcNamedVmGroupsMap struct {
	NamedVmGroups []OpenApiReferences `json:"namedVmGroups,omitempty"`
	Pvdc          OpenApiReference    `json:"pvdc,omitempty"`
}

// PvdcLogicalVmGroupsMap is a combination of a reference to a Provider VDC and a list of references to Logical VM Groups.
// This is used for VM Placement Policies (see VdcComputePolicyV2)
type PvdcLogicalVmGroupsMap struct {
	LogicalVmGroups OpenApiReferences `json:"logicalVmGroups,omitempty"`
	Pvdc            OpenApiReference  `json:"pvdc,omitempty"`
}

type PvdcVgpuClustersMap struct {
	Clusters []string         `json:"clusters,omitempty"`
	Pvdc     OpenApiReference `json:"pvdc,omitempty"`
}

// OpenApiReference is a generic reference type commonly used throughout OpenAPI endpoints
type OpenApiReference struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

type OpenApiReferences []OpenApiReference

// VdcCapability can be used to determine VDC capabilities, including such:
// * Is it backed by NSX-T or NSX-V pVdc
// * Does it support BGP routing
type VdcCapability struct {
	// Name of capability
	Name string `json:"name"`
	// Description of capability
	Description string `json:"description"`
	// Value can be any value. Sometimes it is a JSON bool (true, false), sometimes it is a JSON array (["custom", "default"])
	// and sometimes just a string ("NSX_V"). It is up for the consumer to handle values as per the Type field.
	Value interface{} `json:"value"`
	// Type of field (e.g. "Boolean", "String", "List")
	Type string `json:"type"`
	// Category of capability (e.g. "Security", "EdgeGateway", "OrgVdcNetwork")
	Category string `json:"category"`
}

// A Right is a component of a role, a global role, or a rights bundle.
// In this view, roles, global roles, and rights bundles are collections of rights.
// Note that the rights are not stored in the above collection structures, but retrieved separately
type Right struct {
	Name             string             `json:"name"`
	ID               string             `json:"id"`
	Description      string             `json:"description,omitempty"`
	BundleKey        string             `json:"bundleKey,omitempty"`        // key used for internationalization
	Category         string             `json:"category,omitempty"`         // Category ID
	ServiceNamespace string             `json:"serviceNamespace,omitempty"` // Not used
	RightType        string             `json:"rightType,omitempty"`        // VIEW or MODIFY
	ImpliedRights    []OpenApiReference `json:"impliedRights,omitempty"`
}

// RightsCategory defines the category to which the Right belongs
type RightsCategory struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	BundleKey   string `json:"bundleKey"` // key used for internationalization
	Parent      string `json:"parent"`
	RightsCount struct {
		View   int `json:"view"`
		Modify int `json:"modify"`
	} `json:"rightsCount"`
	SubCategories []string `json:"subCategories"`
}

// RightsBundle is a collection of Rights to be assigned to a tenant(= organization).
// Changing a rights bundle and publishing it for a given tenant will limit
// the rights that the global roles implement in such tenant.
type RightsBundle struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	Description string `json:"description,omitempty"`
	BundleKey   string `json:"bundleKey,omitempty"` // key used for internationalization
	ReadOnly    bool   `json:"readOnly"`
	PublishAll  *bool  `json:"publishAll"`
}

// GlobalRole is a Role definition implemented in the provider that is passed on to tenants (=organizations)
// Modifying an existing global role has immediate effect on the corresponding roles in the tenants (no need
// to re-publish) while creating a new GlobalRole is only passed to the tenants if it is published.
type GlobalRole struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	Description string `json:"description,omitempty"`
	BundleKey   string `json:"bundleKey,omitempty"` // key used for internationalization
	ReadOnly    bool   `json:"readOnly"`
	PublishAll  *bool  `json:"publishAll"`
}

// OpenApiItems defines the input when multiple items need to be passed to a POST or PUT operation
// All the fields are optional, except Values
// This structure is the same as OpenApiPages, except for the type of Values, which is explicitly
// defined as a collection of name+ID structures
type OpenApiItems struct {
	ResultTotal  int                `json:"resultTotal,omitempty"`
	PageCount    int                `json:"pageCount,omitempty"`
	Page         int                `json:"page,omitempty"`
	PageSize     int                `json:"pageSize,omitempty"`
	Associations interface{}        `json:"associations,omitempty"`
	Values       []OpenApiReference `json:"values"` // a collection of items defined by an ID + a name
}

// CertificateLibraryItem is a Certificate Library definition of stored Certificate details
type CertificateLibraryItem struct {
	Alias                string `json:"alias"`
	Id                   string `json:"id,omitempty"`
	Certificate          string `json:"certificate"` // PEM encoded certificate
	Description          string `json:"description,omitempty"`
	PrivateKey           string `json:"privateKey,omitempty"`           // PEM encoded private key. Required if providing a certificate chain
	PrivateKeyPassphrase string `json:"privateKeyPassphrase,omitempty"` // passphrase for the private key. Required if the private key is encrypted
}

// CurrentSessionInfo gives information about the current session
type CurrentSessionInfo struct {
	ID                        string            `json:"id"`                        // Session ID
	User                      OpenApiReference  `json:"user"`                      // Name of the user associated with this session
	Org                       OpenApiReference  `json:"org"`                       // Organization for this connection
	Location                  string            `json:"location"`                  // Location ID: unknown usage
	Roles                     []string          `json:"roles"`                     // Roles associated with the session user
	RoleRefs                  OpenApiReferences `json:"roleRefs"`                  // Roles references for the session user
	SessionIdleTimeoutMinutes int               `json:"sessionIdleTimeoutMinutes"` // session idle timeout
}

// VdcGroup is a VDC group definition
type VdcGroup struct {
	Description                string                 `json:"description,omitempty"`                // The description of this group.
	DfwEnabled                 bool                   `json:"dfwEnabled,omitempty"`                 // Whether Distributed Firewall is enabled for this vDC Group. Only applicable for NSX_T vDC Groups.
	ErrorMessage               string                 `json:"errorMessage,omitempty"`               // If the group has an error status, a more detailed error message is set here.
	Id                         string                 `json:"id,omitempty"`                         // The unique ID for the vDC Group (read-only).
	LocalEgress                bool                   `json:"localEgress,omitempty"`                // Determines whether local egress is enabled for a universal router belonging to a universal vDC group. This value is used on create if universalNetworkingEnabled is set to true. This cannot be updated. This value is always false for local vDC groups.
	Name                       string                 `json:"name"`                                 // The name of this group. The name must be unique.
	NetworkPoolId              string                 `json:"networkPoolId,omitempty"`              // ID of network pool to use if creating a local vDC group router. Must be set if creating a local group. Ignored if creating a universal group.
	NetworkPoolUniversalId     string                 `json:"networkPoolUniversalId,omitempty"`     // The network provider’s universal id that is backing the universal network pool. This field is read-only and is derived from the list of participating vDCs if a universal vDC group is created. For universal vDC groups, each participating vDC should have a universal network pool that is backed by this same id.
	NetworkProviderType        string                 `json:"networkProviderType,omitempty"`        // The values currently supported are NSX_V and NSX_T. Defines the networking provider backing the vDC Group. This is used on create. If not specified, NSX_V value will be used. NSX_V is used for existing vDC Groups and vDC Groups where Cross-VC NSX is used for the underlying technology. NSX_T is used when the networking provider type for the Organization vDCs in the group is NSX-T. NSX_T only supports groups of type LOCAL (single site).
	OrgId                      string                 `json:"orgId"`                                // The organization that this group belongs to.
	ParticipatingOrgVdcs       []ParticipatingOrgVdcs `json:"participatingOrgVdcs"`                 // The list of organization vDCs that are participating in this group.
	Status                     string                 `json:"status,omitempty"`                     // The status that the group can be in. Possible values are: SAVING, SAVED, CONFIGURING, REALIZED, REALIZATION_FAILED, DELETING, DELETE_FAILED, OBJECT_NOT_FOUND, UNCONFIGURED
	Type                       string                 `json:"type,omitempty"`                       // Defines the group as LOCAL or UNIVERSAL. This cannot be changed. Local vDC Groups can have networks stretched across multiple vDCs in a single Cloud Director instance. Local vDC Groups share the same broadcast domain/transport zone and network provider scope. Universal vDC groups can have networks stretched across multiple vDCs in a single or multiple Cloud Director instance(s). Universal vDC groups are backed by a broadcast domain/transport zone that strectches across a single or multiple Cloud Director instance(s). Local vDC groups are supported for both NSX-V and NSX-T Network Provider Types. Universal vDC Groups are supported for only NSX_V Network Provider Type. Possible values are: LOCAL , UNIVERSAL
	UniversalNetworkingEnabled bool                   `json:"universalNetworkingEnabled,omitempty"` // True means that a vDC group router has been created. If set to true for vdc group creation, a universal router will also be created.
}

// ParticipatingOrgVdcs is a participating Org VDCs definition
type ParticipatingOrgVdcs struct {
	FaultDomainTag       string           `json:"faultDomainTag,omitempty"`       // Represents the fault domain of a given organization vDC. For NSX_V backed organization vDCs, this is the network provider scope. For NSX_T backed organization vDCs, this can vary (for example name of the provider vDC or compute provider scope).
	NetworkProviderScope string           `json:"networkProviderScope,omitempty"` // Read-only field that specifies the network provider scope of the vDC.
	OrgRef               OpenApiReference `json:"orgRef,omitempty"`               // Read-only field that specifies what organization this vDC is in.
	RemoteOrg            bool             `json:"remoteOrg,omitempty"`            // Read-only field that specifies whether the vDC is local to this VCD site.
	SiteRef              OpenApiReference `json:"siteRef,omitempty"`              // The site ID that this vDC belongs to. Required for universal vDC groups.
	Status               string           `json:"status,omitempty"`               // The status that the vDC can be in. An example is if the vDC has been deleted from the system but is still part of the group. Possible values are: SAVING, SAVED, CONFIGURING, REALIZED, REALIZATION_FAILED, DELETING, DELETE_FAILED, OBJECT_NOT_FOUND, UNCONFIGURED
	VdcRef               OpenApiReference `json:"vdcRef"`                         // The reference to the vDC that is part of this a vDC group.
}

// CandidateVdc defines possible candidate VDCs for VDC group
type CandidateVdc struct {
	FaultDomainTag       string           `json:"faultDomainTag"`
	Id                   string           `json:"id"`
	Name                 string           `json:"name"`
	NetworkProviderScope string           `json:"networkProviderScope"`
	OrgRef               OpenApiReference `json:"orgRef"`
	SiteRef              OpenApiReference `json:"siteRef"`
}

// DfwPolicies defines Distributed firewall policies
type DfwPolicies struct {
	Enabled       bool           `json:"enabled"`
	DefaultPolicy *DefaultPolicy `json:"defaultPolicy,omitempty"`
}

// DefaultPolicy defines Default policy for Distributed firewall
type DefaultPolicy struct {
	Description string        `json:"description,omitempty"` // Description for the security policy.
	Enabled     *bool         `json:"enabled,omitempty"`     // Whether this security policy is enabled.
	Id          string        `json:"id,omitempty"`          // The unique id of this security policy. On updates, the id is required for the policy, while for create a new id will be generated. This id is not a VCD URN.
	Name        string        `json:"name"`                  // Name for the security policy.
	Version     *VersionField `json:"version,omitempty"`     // This property describes the current version of the entity. To prevent clients from overwriting each other’s changes, update operations must include the version which can be obtained by issuing a GET operation. If the version number on an update call is missing, the operation will be rejected. This is only needed on update calls.
}

// VersionField defines Version
type VersionField struct {
	Version int `json:"version"`
}

// TestConnection defines the parameters used when testing a connection, including SSL handshake and hostname verification.
type TestConnection struct {
	Host                          string               `json:"host"`                                    // The host (or IP address) to connect to.
	Port                          int                  `json:"port"`                                    // The port to use when connecting.
	Secure                        *bool                `json:"secure,omitempty"`                        // If the connection should use https.
	Timeout                       int                  `json:"timeout,omitempty"`                       // Maximum time (in seconds) any step in the test should wait for a response.
	HostnameVerificationAlgorithm string               `json:"hostnameVerificationAlgorithm,omitempty"` // Endpoint/Hostname verification algorithm to be used during SSL/TLS/DTLS handshake.
	AdditionalCAIssuers           []string             `json:"additionalCAIssuers,omitempty"`           // A list of URLs being authorized by the user to retrieve additional CA certificates from, if necessary, to complete the certificate chain to its trust anchor.
	ProxyConnection               *ProxyTestConnection `json:"proxyConnection,omitempty"`               // Proxy connection to use for test. Only one of proxyConnection and preConfiguredProxy can be specified.
	PreConfiguredProxy            string               `json:"preConfiguredProxy,omitempty"`            // The URN of a ProxyConfiguration to use for the test. Only one of proxyConnection or preConfiguredProxy can be specified. If neither is specified then no proxy is used to test the connection.
}

// ProxyTestConnection defines the proxy connection to use for TestConnection (if any).
type ProxyTestConnection struct {
	ProxyHost     string `json:"proxyHost"`               // The host (or IP address) of the proxy.
	ProxyPort     int    `json:"proxyPort"`               // The port to use when connecting to the proxy.
	ProxyUsername string `json:"proxyUsername,omitempty"` // Username to authenticate to the proxy.
	ProxyPassword string `json:"proxyPassword,omitempty"` // Password to authenticate to the proxy.
	ProxySecure   *bool  `json:"proxySecure,omitempty"`   // If the connection to the proxy should use https.
}

// TestConnectionResult is the result of a connection test.
type TestConnectionResult struct {
	TargetProbe *ProbeResult `json:"targetProbe,omitempty"` // Results of a connection test to a specific endpoint.
	ProxyProbe  *ProbeResult `json:"proxyProbe,omitempty"`  // Results of a connection test to a specific endpoint.
}

// ProbeResult results of a connection test to a specific endpoint.
type ProbeResult struct {
	Result              string   `json:"result,omitempty"`              // Localized message describing the connection result stating success or an error message with a brief summary.
	ResolvedIp          string   `json:"resolvedIp,omitempty"`          // The IP address the host was resolved to, if not going through a proxy.
	CanConnect          bool     `json:"canConnect,omitempty"`          // If vCD can establish a connection on the specified port.
	SSLHandshake        bool     `json:"sslHandshake,omitempty"`        // If an SSL Handshake succeeded (secure requests only).
	ConnectionResult    string   `json:"connectionResult,omitempty"`    // A code describing the result of establishing a connection. It can be either SUCCESS, ERROR_CANNOT_RESOLVE_IP or ERROR_CANNOT_CONNECT.
	SSLResult           string   `json:"sslResult,omitempty"`           // A code describing the result of the SSL handshake. It can be either SUCCESS, ERROR_SSL_ERROR, ERROR_UNTRUSTED_CERTIFICATE, ERROR_CANNOT_VERIFY_HOSTNAME or null.
	CertificateChain    string   `json:"certificateChain,omitempty"`    // The SSL certificate chain presented by the server if a secure connection was made.
	AdditionalCAIssuers []string `json:"additionalCAIssuers,omitempty"` // URLs supplied by Certificate Authorities to retrieve signing certificates, when those certificates are not included in the chain.
}

// LogicalVmGroup is used to create VM Placement Policies in VCD.
type LogicalVmGroup struct {
	Name                   string            `json:"name,omitempty"` // Display name
	Description            string            `json:"description,omitempty"`
	ID                     string            `json:"id,omitempty"`                     // UUID for LogicalVmGroup. This is immutable
	NamedVmGroupReferences OpenApiReferences `json:"namedVmGroupReferences,omitempty"` // List of named VM Groups associated with LogicalVmGroup.
	PvdcID                 string            `json:"pvdcId,omitempty"`                 // URN for Provider VDC
}

// DefinedInterface defines a interface for a defined entity. The combination of nss+version+vendor should be unique
type DefinedInterface struct {
	ID         string `json:"id,omitempty"`       // The id of the defined interface type in URN format
	Name       string `json:"name,omitempty"`     // The name of the defined interface
	Nss        string `json:"nss,omitempty"`      // A unique namespace associated with the interface
	Version    string `json:"version,omitempty"`  // The interface's version. The version should follow semantic versioning rules
	Vendor     string `json:"vendor,omitempty"`   // The vendor name
	IsReadOnly bool   `json:"readonly,omitempty"` // True if the entity type cannot be modified
}

// Behavior defines a concept similar to a "procedure" that lives inside Defined Interfaces or Defined Entity Types as overrides.
type Behavior struct {
	ID          string                 `json:"id,omitempty"`          // The Behavior ID is generated and is an output-only property
	Description string                 `json:"description,omitempty"` // A description specifying the contract of the Behavior
	Execution   map[string]interface{} `json:"execution,omitempty"`   // The Behavior execution mechanism. Can be defined both in an Interface and in a Defined Entity Type as an override
	Ref         string                 `json:"ref,omitempty"`         // The Behavior invocation reference to be used for polymorphic behavior invocations. It is generated and is an output-only property
	Name        string                 `json:"name,omitempty"`
}

// BehaviorAccess defines the access control configuration of a Behavior.
type BehaviorAccess struct {
	AccessLevelId string `json:"accessLevelId,omitempty"` // The ID of an AccessLevel
	BehaviorId    string `json:"behaviorId,omitempty"`    // The ID of the Behavior. It can be both a behavior-interface or an overridden behavior-type ID
}

// BehaviorInvocation is an invocation of a Behavior on a Defined Entity instance. Currently, the Behavior interfaces are key-value maps specified in the Behavior description.
type BehaviorInvocation struct {
	Arguments interface{} `json:"arguments,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// DefinedEntityType describes what a Defined Entity Type should look like.
type DefinedEntityType struct {
	ID               string                 `json:"id,omitempty"`               // The id of the defined entity type in URN format
	Name             string                 `json:"name,omitempty"`             // The name of the defined entity type
	Nss              string                 `json:"nss,omitempty"`              // A unique namespace specific string. The combination of nss and version must be unique
	Version          string                 `json:"version,omitempty"`          // The version of the defined entity. The combination of nss and version must be unique. The version string must follow semantic versioning rules
	Description      string                 `json:"description,omitempty"`      // Description of the defined entity
	ExternalId       string                 `json:"externalId,omitempty"`       // An external entity’s id that this definition may apply to
	Hooks            map[string]string      `json:"hooks,omitempty"`            // A mapping defining which behaviors should be invoked upon specific lifecycle events, like PostCreate, PostUpdate, PreDelete. For example: "hooks": { "PostCreate": "urn:vcloud:behavior-interface:postCreateHook:vendorA:containerCluster:1.0.0" }. Added in 36.0
	InheritedVersion string                 `json:"inheritedVersion,omitempty"` // To be used when creating a new version of a defined entity type. Specifies the version of the type that will be the template for the authorization configuration of the new version. The Type ACLs and the access requirements of the Type Behaviors of the new version will be copied from those of the inherited version. If the value of this property is ‘0’, then the new type version will not inherit another version and will have the default authorization settings, just like the first version of a new type. Added in 36.0
	Interfaces       []string               `json:"interfaces,omitempty"`       // List of interface IDs that this defined entity type is referenced by
	MaxImplicitRight string                 `json:"maxImplicitRight,omitempty"` // The maximum Type Right level that will be implied from the user’s Type ACLs if this field is defined. For example, “maxImplicitRight”: “urn:vcloud:accessLevel:ReadWrite” would mean that a user with RO , RW, and FC ACLs to the Type would implicitly get the “Read: ” and “Write: ” rights, but not the “Full Control: ” right. The valid values are “urn:vcloud:accessLevel:ReadOnly”, “urn:vcloud:accessLevel:ReadWrite”, “urn:vcloud:accessLevel:FullControl”
	IsReadOnly       bool                   `json:"readonly,omitempty"`         // `true` if the entity type cannot be modified
	Schema           map[string]interface{} `json:"schema,omitempty"`           // The JSON-Schema valid definition of the defined entity type. If no JSON Schema version is specified, version 4 will be assumed
	Vendor           string                 `json:"vendor,omitempty"`           // The vendor name
}

// DefinedEntity describes an instance of a defined entity type.
type DefinedEntity struct {
	ID         string                 `json:"id,omitempty"`         // The id of the defined entity in URN format
	EntityType string                 `json:"entityType,omitempty"` // The URN ID of the defined entity type that the entity is an instance of. This is a read-only field
	Name       string                 `json:"name,omitempty"`       // The name of the defined entity
	ExternalId string                 `json:"externalId,omitempty"` // An external entity's id that this entity may have a relation to.
	Entity     map[string]interface{} `json:"entity,omitempty"`     // A JSON value representation. The JSON will be validated against the schema of the DefinedEntityType that the entity is an instance of
	State      *string                `json:"state,omitempty"`      // Every entity is created in the "PRE_CREATED" state. Once an entity is ready to be validated against its schema, it will transition in another state - RESOLVED, if the entity is valid according to the schema, or RESOLUTION_ERROR otherwise. If an entity in an "RESOLUTION_ERROR" state is updated, it will transition to the inital "PRE_CREATED" state without performing any validation. If its in the "RESOLVED" state, then it will be validated against the entity type schema and throw an exception if its invalid
	Owner      *OpenApiReference      `json:"owner,omitempty"`      // The owner of the defined entity
	Org        *OpenApiReference      `json:"org,omitempty"`        // The organization of the defined entity.
	Message    string                 `json:"message,omitempty"`    // A message field that might be populated in case entity Resolution fails
}

// DefinedEntityAccess describes Access Control structure for an RDE
type DefinedEntityAccess struct {
	Id            string           `json:"id,omitempty"`
	Tenant        OpenApiReference `json:"tenant"`
	GrantType     string           `json:"grantType"`
	ObjectId      string           `json:"objectId,omitempty"`
	AccessLevelID string           `json:"accessLevelId"`
	MemberID      string           `json:"memberId"`
}

// VSphereVirtualCenter represents a vCenter server
type VSphereVirtualCenter struct {
	// VcId contains the URN of vCenter server
	VcId string `json:"vcId,omitempty"`
	// Name of the vCenter server
	Name string `json:"name"`
	// Optional description
	Description string `json:"description,omitempty"`
	// Username to connect to the server
	Username string `json:"username"`
	// Password in cleartext format to connect to the server
	Password string `json:"password,omitempty"`
	// Url of the server
	Url string `json:"url"`
	// True if the vCenter server is enabled for use
	IsEnabled bool `json:"isEnabled"`
	// VsphereWebClientServerUrl contains the URL of vCenter web client server.
	VsphereWebClientServerUrl string `json:"vsphereWebClientServerUrl,omitempty"`
	// HasProxy indicates that a proxy exists that proxies this vCenter server for access by
	// authorized end-users. Setting this field to true when registering a vCenter server will
	// result in a proxy being created for the vCenter server, and another for the corresponding SSO
	// endpoint (if different from the vCenter server's endpoint). This field is immutable after the
	// vCenter Server is registered, and will be updated by the system when/if the proxy is removed.
	HasProxy bool `json:"hasProxy,omitempty"`
	// vCenter root folder in which the vCloud Director system folder will be created. This parameter only takes the folder name and not directory structure.
	RootFolder string `json:"rootFolder,omitempty"`
	// Network in Vcenter to be used as 'NONE' network by vCD.
	VcNoneNetwork string `json:"vcNoneNetwork,omitempty"`
	// TenantVisibleName contains public label of this vCenter server visible to all tenants
	TenantVisibleName string `json:"tenantVisibleName,omitempty"`
	// IsConnected True if the vCenter server is connected.
	IsConnected bool `json:"isConnected,omitempty"`
	// 	The vCenter mode. One of
	// * NONE - undetermined
	// * IAAS - provider scoped
	// * SDDC - tenant scoped
	// * MIXED
	// IAAS indicates this vCenter server is scoped to the provider. SDDC indicates that this
	// vCenter server is scoped to tenants, while MIXED indicates mixed mode, where both uses are
	// allowed in this vCenter server. Possible values are: NONE , IAAS , SDDC , MIXED
	Mode string `json:"mode,omitempty"`
	// The vCenter listener state. One of:
	// * INITIAL
	// * INVALID_SETTINGS
	// * UNSUPPORTED
	// * DISCONNECTED
	// * CONNECTING
	// * CONNECTED_SYNCING
	// * CONNECTED
	// * STOP_REQ
	// * STOP_AND_PURGE_REQ
	// * STOP_ACK
	ListenerState string `json:"listenerState,omitempty"`
	// ClusterHealthStatus shows the overall health status of clusters in this vCenter server. One
	// of GRAY, RED, YELLOW, GREEN
	ClusterHealthStatus string `json:"clusterHealthStatus,omitempty"`
	// The version of the VIM server.
	VcVersion string `json:"vcVersion,omitempty"`
	// The build number of the VIM server.
	BuildNumber string `json:"buildNumber,omitempty"`
	// The instance UUID property of the vCenter server.
	Uuid string `json:"uuid,omitempty"`
	// NsxVManager stores the NSX-V attached to this Virtual Center server, when present.
	NsxVManager *VSphereVirtualCenterNsxvManager `json:"nsxVManager,omitempty"`
	// ProxyConfigurationUrn is Deprecated
	ProxyConfigurationUrn string `json:"proxyConfigurationUrn,omitempty"`
}

type VSphereVirtualCenterNsxvManager struct {
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	Url             string `json:"url,omitempty"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
}

type ResourcePoolSummary struct {
	Associations []struct {
		EntityId      string `json:"entityId"`
		AssociationId string `json:"associationId"`
	} `json:"associations"`
	Values []ResourcePool `json:"values"`
}

// ResourcePool defines a vSphere Resource Pool
type ResourcePool struct {
	Moref             string `json:"moref"`
	ClusterMoref      string `json:"clusterMoref"`
	Name              string `json:"name"`
	VcId              string `json:"vcId"`
	Eligible          bool   `json:"eligible"`
	KubernetesEnabled bool   `json:"kubernetesEnabled"`
	VgpuEnabled       bool   `json:"vgpuEnabled"`
}

// OpenApiSupportedHardwareVersions is the list of versions supported by a given resource
type OpenApiSupportedHardwareVersions struct {
	Versions          []string `json:"versions"`
	SupportedVersions []struct {
		IsDefault bool   `json:"isDefault"`
		Name      string `json:"name"`
	} `json:"supportedVersions"`
}

// NetworkPool is the full data retrieved for a provider network pool
type NetworkPool struct {
	Status             string             `json:"status,omitempty"`
	Id                 string             `json:"id,omitempty"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	PoolType           string             `json:"poolType"`
	PromiscuousMode    bool               `json:"promiscuousMode,omitempty"`
	TotalBackingsCount int                `json:"totalBackingsCount,omitempty"`
	UsedBackingsCount  int                `json:"usedBackingsCount,omitempty"`
	ManagingOwnerRef   OpenApiReference   `json:"managingOwnerRef"`
	Backing            NetworkPoolBacking `json:"backing"`
}

// NetworkPoolBacking is the definition of the objects supporting the network pool
type NetworkPoolBacking struct {
	VlanIdRanges     VlanIdRanges       `json:"vlanIdRanges,omitempty"`
	VdsRefs          []OpenApiReference `json:"vdsRefs,omitempty"`
	PortGroupRefs    []OpenApiReference `json:"portGroupRefs,omitempty"`
	TransportZoneRef OpenApiReference   `json:"transportZoneRef,omitempty"`
	ProviderRef      OpenApiReference   `json:"providerRef"`
}

type VlanIdRanges struct {
	Values []VlanIdRange `json:"values"`
}

// VlanIdRange is a component of a network pool of type VLAN
type VlanIdRange struct {
	StartId int `json:"startId"`
	EndId   int `json:"endId"`
}

// OpenApiStorageProfile defines a storage profile before it is assigned to a provider VDC
type OpenApiStorageProfile struct {
	Moref string `json:"moref"`
	Name  string `json:"name"`
}

// UIPluginMetadata gives meta information about a UI Plugin
type UIPluginMetadata struct {
	ID             string `json:"id,omitempty"`
	Vendor         string `json:"vendor,omitempty"`
	License        string `json:"license,omitempty"`
	Link           string `json:"link,omitempty"`
	PluginName     string `json:"pluginName,omitempty"`
	Version        string `json:"version,omitempty"`
	Description    string `json:"description,omitempty"`
	ProviderScoped bool   `json:"provider_scoped,omitempty"`
	TenantScoped   bool   `json:"tenant_scoped,omitempty"`
	Enabled        bool   `json:"enabled,omitempty"`
	PluginStatus   string `json:"plugin_status,omitempty"`
}

// UploadSpec gives information about an upload
type UploadSpec struct {
	FileName     string `json:"fileName,omitempty"`
	Size         int64  `json:"size,omitempty"`
	Checksum     string `json:"checksum,omitempty"`
	ChecksumAlgo string `json:"checksumAlgo,omitempty"`
}

type TransportZones struct {
	Values []*TransportZone `json:"values"`
}

// TransportZone is a backing component of a network pool of type 'GENEVE' (NSX-T backed)
type TransportZone struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type,omitempty"`
	AlreadyImported bool   `json:"alreadyImported"`
}

// VcenterDistributedSwitch is a backing component of a network pool of type VLAN
type VcenterDistributedSwitch struct {
	BackingRef    OpenApiReference `json:"backingRef"`
	VirtualCenter OpenApiReference `json:"virtualCenter"`
}

// OpenApiMetadataEntry represents a metadata entry in VCD.
type OpenApiMetadataEntry struct {
	ID           string                  `json:"id,omitempty"`         // UUID for OpenApiMetadataEntry. This is immutable
	IsPersistent bool                    `json:"persistent,omitempty"` // Persistent entries can be copied over on some entity operation, for example: Creating a copy of an Org VDC, capturing a vApp to a template, instantiating a catalog item as a VM, etc.
	IsReadOnly   bool                    `json:"readOnly,omitempty"`   // The kind of level of access organizations of the entry’s domain have
	KeyValue     OpenApiMetadataKeyValue `json:"keyValue,omitempty"`   // Contains core metadata entry data
}

// OpenApiMetadataKeyValue contains core metadata entry data.
type OpenApiMetadataKeyValue struct {
	Domain    string                    `json:"domain,omitempty"`    // Only meaningful for providers. Allows them to share entries with their tenants. Currently, accepted values are: `TENANT`, `PROVIDER`, where that is the ascending sort order of the enumeration.
	Key       string                    `json:"key,omitempty"`       // Key of the metadata entry
	Value     OpenApiMetadataTypedValue `json:"value,omitempty"`     // Value of the metadata entry
	Namespace string                    `json:"namespace,omitempty"` // Namespace of the metadata entry
}

// OpenApiMetadataTypedValue the type and value of the metadata entry.
type OpenApiMetadataTypedValue struct {
	Value interface{} `json:"value,omitempty"` // The Value is anything because it depends on the Type field.
	Type  string      `json:"type,omitempty"`
}

// VgpuProfile uniquely represents a type of vGPU
// vGPU Profiles are fetched from your NVIDIA GRID GPU enabled Clusters in vCenter.
type VgpuProfile struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	TenantFacingName   string `json:"tenantFacingName"`
	Instructions       string `json:"instructions,omitempty"`
	AllowMultiplePerVm bool   `json:"allowMultiplePerVm"`
	Count              int    `json:"count,omitempty"`
}

type OpenApiOrg struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	DisplayName    string `json:"displayName"`
	Description    string `json:"description"`
	IsEnabled      bool   `json:"isEnabled"`
	OrgVdcCount    int    `json:"orgVdcCount"`
	CatalogCount   int    `json:"catalogCount"`
	VappCount      int    `json:"vappCount"`
	RunningVMCount int    `json:"runningVMCount"`
	UserCount      int    `json:"userCount"`
	DiskCount      int    `json:"diskCount"`
	CanPublish     bool   `json:"canPublish"`
}

// ExternalEndpoint is part of the API extensibility framework.
// They allow requests to be directly proxied over HTTP to an external endpoint.
type ExternalEndpoint struct {
	Name        string `json:"name,omitempty"`        // The name of the external endpoint
	ID          string `json:"id,omitempty"`          // The unique id of the external endpoint
	Version     string `json:"version,omitempty"`     // The external endpoint's version. The version should follow semantic versioning rules. Versions with pre-release extension are not allowed. The combination of vendor-namespace-version must be unique
	Vendor      string `json:"vendor,omitempty"`      // The vendor name. The combination of vendor-namespace-version must be unique
	Enabled     bool   `json:"enabled"`               // Whether the external endpoint is enabled or not
	Description string `json:"description,omitempty"` // Description of the defined entity
	RootUrl     string `json:"rootUrl,omitempty"`     // The external endpoint which requests will be redirected to. The rootUrl must be a valid URL of https protocol
}

// ApiFilter is part of the API extensibility framework.
// They allow external systems (external services and external endpoints) to extend the standard API included with VCD
// with custom URLs or custom processing of request's responses.
type ApiFilter struct {
	ID             string            `json:"id,omitempty"`             // The unique id of the API filter
	ExternalSystem *OpenApiReference `json:"externalSystem,omitempty"` // Entity reference used to describe VCD entities
	UrlMatcher     *UrlMatcher       `json:"urlMatcher,omitempty"`

	// The responseContentType is expressed as a MIME Content-Type string. Responses whose Content-Type attribute has a value
	// that matches this string are routed to the service. responseContentType is mutually exclusive with urlMatcher.
	ResponseContentType *string `json:"responseContentType,omitempty"`
}

// UrlMatcher consists of urlPattern and urlScope which together identify a URL which will be serviced by an external system.
// For example, if you want the external system to service all requests matching '/ext-api/custom/.', the URL Matcher object should be:
// { "urlPattern": "/custom/.", "urlScope": "EXT_API" }.
// It is important to note that in the case of EXT_UI_TENANT urlScope, the tenant name is not part of the urlPattern.
// The urlPattern will match the request after the tenant name - if request is "/ext-ui/tenant/testOrg/custom/test",
// the pattern will match against "/custom/test".
type UrlMatcher struct {
	UrlPattern string `json:"urlPattern,omitempty"`
	UrlScope   string `json:"urlScope,omitempty"` // EXT_API, EXT_UI_PROVIDER, EXT_UI_TENANT corresponding to /ext-api, /ext-ui/provider, /ext-ui/tenant/<tenant-name>
}

// TrustedCertificate manages certificate trust
type TrustedCertificate struct {
	ID string `json:"id,omitempty"`
	// Alias contains case insensitive name
	Alias string `json:"alias"`
	// Certificate contains PEM encoded certificate. All extraneous whitespace and other information
	// is removed
	Certificate string `json:"certificate"`
}

// NsxtManagerOpenApi reflects NSX-T manager configuration for OpenAPI endpoint
type NsxtManagerOpenApi struct {
	ID string `json:"id,omitempty"`
	// Name of NSX-T Manager
	Name string `json:"name"`
	// Description of NSX-T Manager
	Description string `json:"description"` // This one is documented as optional (omitempty), but it is mandatory (it fails if not sent)
	// Username for authenticating to NSX-T Manager
	Username string `json:"username"`
	// Password for authenticating to NSX-T Manager
	Password string `json:"password,omitempty"`
	// Url for authenticating to NSX-T Manager
	Url string `json:"url"`
	// Active indicates whether the NSX Manager can or cannot be used to manage networking constructs within Tenant Manager
	Active bool `json:"active,omitempty"`
	// ClusterId of the NSX Manager. Each NSX installation has a single cluster. This is not a Tenant Manager URN
	ClusterId string `json:"clusterId,omitempty"`
	// IsDedicatedForClassicTenants whether this NSX Manager is dedicated for legacy VRA-style tenants only and unable to
	// participate in modern constructs such as Regions and Zones. Legacy VRA-style is deprecated and this field exists for
	// the purpose of VRA backwards compatibility only
	IsDedicatedForClassicTenants bool `json:"isDedicatedForClassicTenants,omitempty"`
	// Status of NSX-T Manager. Possible values are:
	// PENDING - Desired entity configuration has been received by system and is pending realization.
	// CONFIGURING - The system is in process of realizing the entity.
	// REALIZED - The entity is successfully realized in the system.
	// REALIZATION_FAILED - There are some issues and the system is not able to realize the entity.
	// UNKNOWN - Current state of entity is unknown.
	Status string `json:"status,omitempty"`
}

// OpenApiUser defines structure for User management using OpenAPI based endpoint
type OpenApiUser struct {
	ID string `json:"id,omitempty"`
	// Username for the user
	Username string `json:"username"`
	// Password for the user. Must be null for external users
	Password string `json:"password,omitempty"`
	// Enabled state of the user. Defaults to true
	Enabled *bool `json:"enabled,omitempty"`
	// Description of the user
	Description string `json:"description,omitempty"`
	// A read-only list of all of a user's roles, both directly assigned and inherited from the user's groups
	EffectiveRoleEntityRefs []*OpenApiReference `json:"effectiveRoleEntityRefs,omitempty"`
	// A user's email address. Based on org email preferences, notifications can be sent to the user via email
	Email string `json:"email,omitempty"`
	// Family name of the user (e.g. last name in most Western languages)
	FamilyName string `json:"familyName,omitempty"`
	// Full name (display name) of the user
	FullName string `json:"fullName,omitempty"`
	// The directly assigned role(s) of the user
	RoleEntityRefs []OpenApiReference `json:"roleEntityRefs,omitempty"`
	// Given name of the user (e.g. first name in most Western languages)
	GivenName string `json:"givenName,omitempty"`
	// Determines if this user can inherit roles from groups. Defaults to false
	InheritGroupRoles bool `json:"inheritGroupRoles,omitempty"`
	// Determines if this user's role is inherited from a group. Defaults to false
	IsGroupRole bool `json:"isGroupRole,omitempty"`
	// True if the user account has been locked due to too many invalid login attempts. An
	// administrator can unlock a locked user account by setting this flag to false. A user may not
	// be explicitly locked. Instead, disable the user, if user's access must be revoked
	// temporarily
	Locked *bool `json:"locked,omitempty"`
	// Name of the user in its source
	NameInSource string `json:"nameInSource,omitempty"`
	// orgEntityRefOptional
	OrgEntityRef *OpenApiReference `json:"orgEntityRef,omitempty"`
	// Phone number of the user.
	Phone string `json:"phone,omitempty"`
	// Provider type of the user. It must be one of: LOCAL, LDAP, SAML, OAUTH
	ProviderType string `json:"providerType,omitempty"`
	// True if the user account has been stranded, meaning it is unable to be accessed due to its
	// original identity source being removed
	Stranded bool `json:"stranded,omitempty"`
}
