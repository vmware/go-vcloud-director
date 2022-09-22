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
	Name        string `json:"name"`
	Description string `json:"description"`
	BundleKey   string `json:"bundleKey"`
	ReadOnly    bool   `json:"readOnly"`
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
	Subnets ExternalNetworkV2Subnets `json:"subnets"`
	// NetworkBackings for this external network. Describes if this external network is backed by
	// port groups, vCenter standard switch or an NSX-T Tier-0 router.
	NetworkBackings ExternalNetworkV2Backings `json:"networkBackings"`
}

// ExternalNetworkV2IPRange defines allocated IP pools for a subnet in external network
type ExternalNetworkV2IPRange struct {
	// StartAddress holds starting IP address in the range
	StartAddress string `json:"startAddress"`
	// EndAddress holds ending IP address in the range
	EndAddress string `json:"endAddress"`
}

// ExternalNetworkV2IPRanges contains slice of ExternalNetworkV2IPRange
type ExternalNetworkV2IPRanges struct {
	Values []ExternalNetworkV2IPRange `json:"values"`
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
