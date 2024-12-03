package types

// RegionStoragePolicy defines a Region storage policy
type RegionStoragePolicy struct {
	ID string `json:"id,omitempty"`
	// Name for the policy. It must follow RFC 1123 Label Names to conform with Kubernetes standards
	Name string `json:"name"`
	// The Region that this policy belongs to
	Region *OpenApiReference `json:"region"`
	// Description of the policy
	Description string `json:"description,omitempty"`
	// The creation status of the region storage policy. Can be [NOT_READY, READY]
	Status string `json:"status,omitempty"`
	// Storage capacity in megabytes for this policy
	StorageCapacityMB int64 `json:"storageCapacityMB,omitempty"`
	// Consumed storage in megabytes for this policy
	StorageConsumedMB int64 `json:"storageConsumedMB,omitempty"`
}

// ContentLibrary is an object representing a VCF Content Library
type ContentLibrary struct {
	// The name of the Content Library
	Name string `json:"name"`
	// A collection of storage class references used by this Content Library
	StorageClasses OpenApiReferences `json:"storageClasses,omitempty"`
	// For Tenant Content Libraries this field represents whether this Content Library should be automatically attached to
	// all current and future namespaces in the tenant organization. If no value is supplied during Tenant Content Library
	// creation then this field will default to true. If a value of false is supplied, then this Tenant Content Library will
	// only be attached to namespaces that explicitly request it. For Provider Content Libraries this field is not needed for
	// creation and will always be returned as true. This field cannot be updated after Content Library creation
	AutoAttach bool `json:"autoAttach,omitempty"`
	// The ISO-8601 timestamp representing when this Content Library was created
	CreationDate string `json:"creationDate,omitempty"`
	// The description of the Content Library
	Description string `json:"description,omitempty"`
	// A unique identifier for the Content library
	ID string `json:"id,omitempty"`
	// Whether this Content Library is shared with other organizations
	IsShared bool `json:"isShared,omitempty"`
	// Whether this Content Library is subscribed from an external published library
	IsSubscribed bool `json:"isSubscribed,omitempty"`
	// The type of content library:
	// - PROVIDER - Content Library that is scoped to a provider
	// - TENANT - Content Library that is scoped to a tenant organization
	LibraryType string `json:"libraryType,omitempty"`
	// The reference to the organization that the Content Library belongs to
	Org *OpenApiReference `json:"org,omitempty"`
	// An object representing subscription settings of a Content Library
	SubscriptionConfig *ContentLibrarySubscriptionConfig `json:"subscriptionConfig,omitempty"`
	// Version number of this Content library
	VersionNumber int64 `json:"versionNumber,omitempty"`
}

// ContentLibrarySubscriptionConfig represents subscription settings of a Content Library
type ContentLibrarySubscriptionConfig struct {
	// Subscription url of this Content Library. It cannot be changed once set for a Content Library
	SubscriptionUrl string `json:"subscriptionUrl"`
	// Whether to eagerly download content from publisher and store it locally
	NeedLocalCopy bool `json:"needLocalCopy,omitempty"`
	// Password to use to authenticate with the publisher
	Password string `json:"password,omitempty"`
}

// TmOrg defines structure for creating TM Organization
type TmOrg struct {
	ID string `json:"id,omitempty"`
	// Name of organization that will be used in the URL slug
	Name string `json:"name"`
	// DisplayName contains a full display name of the organization
	DisplayName string `json:"displayName"`
	// Description of the Org
	Description string `json:"description,omitempty"`

	// CanManageOrgs sets whether or not this org can manage other tenant orgs.
	// This can be toggled to true to automatically perform the following steps:
	// * Publishes the Default Sub-Provider Entitlement Rights Bundle to the org
	// * Publishes the Sub-Provider Administrator global role (if it exists) to the org
	// * Creates a Default Rights Bundle in the org containing all publishable rights that are
	// currently published to the org. Marks that Rights Bundle as publish all.
	// * Clones all default roles currently published to the org into Global Roles in the org. Marks
	// them all publish all
	// Cannot be set to false as there may be any number of Rights Bundles granting sub-provider
	// rights to this org. Instead, unpublish any rights bundles that have the Org Traverse right
	// from this org
	CanManageOrgs bool `json:"canManageOrgs,omitempty"`
	// CanPublish defines whether the organization can publish catalogs externally
	CanPublish bool `json:"canPublish,omitempty"`
	// CatalogCount withing the Org
	CatalogCount int `json:"catalogCount,omitempty"`
	// DirectlyManagedOrgCount contains the count of the orgs this org directly manages
	DirectlyManagedOrgCount int `json:"directlyManagedOrgCount,omitempty"`
	// DiskCount defines the number of disks in the Org
	DiskCount int `json:"diskCount,omitempty"`

	// IsClassicTenant defines whether the organization is a classic VRA-style tenant. This field
	// cannot be updated. Note this style is deprecated and this field exists for the purpose of VRA
	// backwards compatibility.
	IsClassicTenant bool `json:"isClassicTenant,omitempty"`
	// IsEnabled defines if the Org is enabled
	IsEnabled bool `json:"isEnabled,omitempty"`
	// ManagedBy defines the provider Org that manages this Organization
	ManagedBy *OpenApiReference `json:"managedBy,omitempty"`
	// MaskedEventTaskUsername sets username as it appears in the tenant events/tasks. Requires
	// 'Organization Edit Username Mask'
	MaskedEventTaskUsername string `json:"maskedEventTaskUsername,omitempty"`
	// OrgVdcCount contains count of VDCs assigned to the Org
	OrgVdcCount int `json:"orgVdcCount,omitempty"`
	// RunningVMCount contains count of VM running in the Org
	RunningVMCount int `json:"runningVMCount,omitempty"`
	// UserCount contains user count in the Org
	UserCount int `json:"userCount,omitempty"`
	// VappCount contains vApp count in the Org
	VappCount int `json:"vappCount,omitempty"`
}

// Region represents a collection of supervisor clusters across different VCs
type Region struct {
	ID string `json:"id,omitempty"`
	// The name of the region. It must follow RFC 1123 Label Names to conform with Kubernetes standards.
	Name string `json:"name"`
	// The description of the region.
	Description string `json:"description"`
	// The NSX manager for the region.
	NsxManager *OpenApiReference `json:"nsxManager"`
	// Total CPU resources in MHz available to this Region.
	CPUCapacityMHz int `json:"cpuCapacityMHz,omitempty"`
	// Total CPU reservation resources in MHz available to this Region.
	CPUReservationCapacityMHz int `json:"cpuReservationCapacityMHz,omitempty"`
	// Whether the region is enabled or not.
	IsEnabled bool `json:"isEnabled"`
	// Total memory resources (in mebibytes) available to this Region.
	MemoryCapacityMiB int `json:"memoryCapacityMiB,omitempty"`
	// Total memory reservation resources (in mebibytes) available to this Region.
	MemoryReservationCapacityMiB int `json:"memoryReservationCapacityMiB,omitempty"`
	// The creation status of the Provider VDC. Possible values are READY, NOT_READY, ERROR, FAILED.
	// A Region needs to be ready and enabled to be usable.
	Status string `json:"status,omitempty"`
	// A list of supervisors in a region
	Supervisors []OpenApiReference `json:"supervisors,omitempty"`
	// A list of distinct vCenter storage policy names from the vCenters taking part in this region.
	// A storage policy with the given name must exist in all the vCenters of this region otherwise
	// it will not be accepted. Only the storage policies added to a region can be published to the
	// tenant Virtual Datacenters.
	StoragePolicies []string `json:"storagePolicies,omitempty"`
}

// Supervisor represents a single Supervisor within vCenter
type Supervisor struct {
	// The immutable identifier of this supervisor.
	SupervisorID string `json:"supervisorId"`
	// The name of this supervisor.
	Name string `json:"name"`
	// The Region this Supervisor is associated with. If null, it has not been associated with a Region.
	Region *OpenApiReference `json:"region,omitempty"`
	// The vCenter this supervisor is associated with.
	VirtualCenter *OpenApiReference `json:"virtualCenter"`
}

// SupervisorZone represents a single zone within Supervisor
type SupervisorZone struct {
	ID string `json:"id"`
	// The name of this zone.
	Name string `json:"name"`
	// The supervisor this zone belongs to.
	Supervisor *OpenApiReference `json:"supervisor"`
	// The vCenter this supervisor zone is associated with.
	VirtualCenter *OpenApiReference `json:"virtualCenter"`
	// TotalMemoryCapacityMiB - the memory capacity (in mebibytes) in this zone. Total memory
	// consumption in this zone cannot cross this limit
	TotalMemoryCapacityMiB int64 `json:"totalMemoryCapacityMiB"`
	// TotalCPUCapacityMHz - the CPU capacity (in MHz) in this zone. Total CPU consumption in this
	// zone cannot cross this limit
	TotalCPUCapacityMHz int64 `json:"totalCPUCapacityMHz"`
	// MemoryUsedMiB - total memory used (in mebibytes) in this zone
	MemoryUsedMiB int64 `json:"memoryUsedMiB"`
	// CpuUsedMHz - total CPU used (in MHz) in this zone
	CpuUsedMHz int64 `json:"cpuUsedMHz"`
	// Region contains a reference to parent region
	Region *OpenApiReference `json:"region"`
}

// TmIpSpace provides configuration of mainly the external IP Prefixes that specifies
// the accessible external networks from the data center
type TmIpSpace struct {
	ID string `json:"id,omitempty"`
	// Name of the IP Space
	Name string `json:"name"`
	// Description of the IP Space
	Description string `json:"description,omitempty"`
	// RegionRef is the region that this IP Space belongs in. Only Provider Gateways in the same Region can be
	// associated with this IP Space. This field cannot be updated
	RegionRef OpenApiReference `json:"regionRef"`
	// Default IP quota that applies to all the organizations the Ip Space is assigned to
	DefaultQuota TmIpSpaceDefaultQuota `json:"defaultQuota,omitempty"`
	// ExternalScopeCidr defines the total span of IP addresses to which the IP space has access.
	// This typically defines the span of IP addresses outside the bounds of a Data Center. For the
	// internet, this may be 0.0.0.0/0. For a WAN, this could be 10.0.0.0/8.
	ExternalScopeCidr string `json:"externalScopeCidr,omitempty"`
	// InternalScopeCidrBlocks defines the span of IP addresses used within a Data Center. For new
	// CIDR value not in the existing list, a new IP Block will be created. For existing CIDR value,
	// the IP Block's name can be updated. If an existing CIDR value is removed from the list, the
	// the IP Block is removed from the IP Space.
	InternalScopeCidrBlocks []TmIpSpaceInternalScopeCidrBlocks `json:"internalScopeCidrBlocks,omitempty"`
	// Represents current status of the networking entity. Possible values are:
	// * PENDING - Desired entity configuration has been received by system and is pending realization.
	// * CONFIGURING - The system is in process of realizing the entity.
	// * REALIZED - The entity is successfully realized in the system.
	// * REALIZATION_FAILED - There are some issues and the system is not able to realize the entity.
	// * UNKNOWN - Current state of entity is unknown.
	Status string `json:"status,omitempty"`
}

// IP Space quota defines the maximum number of IPv4 IPs and CIDRs that can be allocated and used by
// the IP Space across all its Internal Scopes
type TmIpSpaceDefaultQuota struct {
	// The maximum number of CIDRs with size maxSubnetSize or less, that can be allocated from all
	// the Internal Scopes of the IP Space. A '-1' value means no cap on the number of the CIDRs
	// used
	MaxCidrCount int `json:"maxCidrCount,omitempty"`
	// The maximum number of single floating IP addresses that can be allocated and used from all
	// the Internal Scopes of the IP Space. A '-1' value means no cap on the number of floating IP
	// Addresses
	MaxIPCount int `json:"maxIpCount,omitempty"`
	// The maximum size of the subnets, represented as a prefix length. The CIDRs that are allocated
	// from the Internal Scopes of the IP Space must be smaller or equal to the specified size. For
	// example, for a maxSubnetSize of 24, CIDRs with prefix length of 24, 28 or 30 can be
	// allocated
	MaxSubnetSize int `json:"maxSubnetSize,omitempty"`
}

// An IP Block represents a named CIDR that is backed by a network provider
type TmIpSpaceInternalScopeCidrBlocks struct {
	// Unique backing ID of the IP Block. This is not a Tenant Manager URN. This field is read-only and is ignored on create/update
	ID string `json:"id,omitempty"`
	// The name of the IP Block. If not set, a random name will be generated that will be prefixed
	// with the name of the IP Space. This property is updatable if there's an existing IP Block
	// with the CIDR value
	Name string `json:"name,omitempty"`
	// The CIDR that represents this IP Block. This property is not updatable
	Cidr string `json:"cidr,omitempty"`
}

type TmProviderGateway struct {
	ID string
}

type TmTier0Gateway struct {
	ID string
}
