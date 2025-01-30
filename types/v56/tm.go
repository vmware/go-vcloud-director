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

// StorageClass defines a Storage Class
type StorageClass struct {
	ID string `json:"id,omitempty"`
	// Name for the storage class
	Name string `json:"name"`
	// The Region that this storage class belongs to
	Region *OpenApiReference `json:"region"`
	// The total storage capacity of the storage class in mebibytes
	StorageCapacityMiB int64 `json:"storageCapacityMiB,omitempty"`
	// For tenants, this represents the total storage given to all namespaces consuming from this storage class in mebibytes.
	// For providers, this represents the total storage given to tenants from this storage class in mebibytes.
	StorageConsumedMiB int64 `json:"storageConsumedMiB,omitempty"`
	// The zones available to the storage class
	Zones OpenApiReferences `json:"zones,omitempty"`
}

// VirtualDatacenterStoragePolicies represents a slice of RegionStoragePolicy
type VirtualDatacenterStoragePolicies struct {
	Values []VirtualDatacenterStoragePolicy `json:"values"`
}

// VirtualDatacenterStoragePolicy describes a Virtual Datacenter Storage Policy
type VirtualDatacenterStoragePolicy struct {
	ID                  string           `json:"id,omitempty"`
	RegionStoragePolicy OpenApiReference `json:"regionStoragePolicy"`
	StorageLimitMiB     int64            `json:"storageLimitMiB"`
	VirtualDatacenter   OpenApiReference `json:"virtualDatacenter"`
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

// ContentLibraryItem is an object representing a VCF Content Library Item
type ContentLibraryItem struct {
	// The reference to the content library that this item belongs to
	ContentLibrary OpenApiReference `json:"contentLibrary"`
	// The name of the content library item
	Name string `json:"name"`
	// The type of content library item. This field is only required for content library upload
	ItemType string `json:"itemType"`

	// The ISO-8601 timestamp representing when this item was created
	CreationDate string `json:"creationDate,omitempty"`
	// The description of the content library item
	Description string `json:"description,omitempty"`
	// A unique identifier for the library item
	ID string `json:"id,omitempty"`
	// Virtual Machine Identifier (VMI) of the item. This is a ReadOnly field
	ImageIdentifier string `json:"imageIdentifier,omitempty"`
	// Whether this item is published
	IsPublished bool `json:"isPublished,omitempty"`
	// Whether this item is subscribed
	IsSubscribed bool `json:"isSubscribed,omitempty"`
	// The ISO-8601 timestamp representing when this item was last synced if subscribed
	LastSuccessfulSync string `json:"lastSuccessfulSync,omitempty"`
	// The reference to the organization that the item belongs to
	Org *OpenApiReference `json:"org,omitempty"`
	// Status of this content library item
	Status string `json:"status,omitempty"`
	// The version of this item. For a subscribed library, this version is same as in publisher library
	Version int `json:"version,omitempty"`
}

// ContentLibraryItemFile specifies a Content Library Item file for uploads
type ContentLibraryItemFile struct {
	ExpectedSizeBytes int64  `json:"expectedSizeBytes"`
	BytesTransferred  int64  `json:"bytesTransferred"`
	Name              string `json:"name"`
	TransferUrl       string `json:"transferUrl"`
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

// TmOrgNetworkingSettings defines structure for managing Org Networking Setttings
type TmOrgNetworkingSettings struct {
	// Whether this Organization has tenancy for the network domain in the backing network provider.
	// If enabled, can only be disabled after all Org VDCs and VDC Groups that have networking
	// tenancy enabled are deleted. Is disabled by default.
	NetworkingTenancyEnabled *bool `json:"networkingTenancyEnabled,omitempty"`

	// A short (8 char) display name to identify this Organization in the logs of the backing
	// network provider. Only applies if the Organization is networking tenancy enabled. This
	// identifier is globally unique.
	OrgNameForLogs string `json:"orgNameForLogs"`
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

// TmVdc defines a structure for creating VDCs using OpenAPI endpoint
type TmVdc struct {
	ID string `json:"id,omitempty"`
	// Name of the VDC
	Name string `json:"name"`
	// Description of the VDC
	Description string `json:"description,omitempty"`
	// Org reference
	Org *OpenApiReference `json:"org"`
	// Region reference
	Region *OpenApiReference `json:"region"`
	// Status contains creation status of the VDC
	Status string `json:"status,omitempty"`
	// Supervisors contain references to Supervisors
	Supervisors []OpenApiReference `json:"supervisors,omitempty"`
	// ZoneResourceAllocation contain references of each zone within Supervisor
	ZoneResourceAllocation []*TmVdcZoneResourceAllocation `json:"zoneResourceAllocation,omitempty"`
}

// TmVdcZoneResourceAllocation defines resource allocation for a single zone
type TmVdcZoneResourceAllocation struct {
	ResourceAllocation TmVdcResourceAllocation `json:"resourceAllocation"`
	Zone               *OpenApiReference       `json:"zone"`
}

// TmVdcResourceAllocation defines compute resources of single VDC
type TmVdcResourceAllocation struct {
	// CPULimitMHz defines maximum CPU consumption limit in MHz
	CPULimitMHz int `json:"cpuLimitMHz"`
	// CPUReservationMHz defines reserved CPU capacity in MHz
	CPUReservationMHz int `json:"cpuReservationMHz"`
	// MemoryLimitMiB defines maximum memory consumption limit in MiB
	MemoryLimitMiB int `json:"memoryLimitMiB"`
	// MemoryReservationMiB defines reserved memory in Mib
	MemoryReservationMiB int `json:"memoryReservationMiB"`
}

// Zone defines a Region Zone structure
type Zone struct {
	ID string `json:"id,omitempty"`
	// Name of the Region Zone
	Name string `json:"name"`
	// Region reference
	Region *OpenApiReference `json:"region"`
	// CPULimitMhz defines the total amount of reserved and unreserved CPU resources allocated in
	// MHz
	CPULimitMhz int `json:"cpuLimitMhz"`
	// CPUReservationMhz contains the total amount of CPU resources reserved in MHz
	CPUReservationMhz int `json:"cpuReservationMhz"`
	// CPUReservationUsedMhz defines the amount of CPU resources used in MHz. For Tenants, this
	// value represents the total given to all of a Tenant's Namespaces. For Providers, this value
	// represents the total given to all Tenants
	CPUReservationUsedMhz int `json:"cpuReservationUsedMhz"`
	// CPUUsedMhz defines the amount of reserved and unreserved CPU resources used in MHz. For
	// Tenants, this value represents the total given to all of a Tenant's Namespaces. For
	// Providers, this value represents the total given to all Tenants
	CPUUsedMhz int `json:"cpuUsedMhz"`
	// MemoryLimitMiB defines the total amount of reserved and unreserved memory resources allocated
	// in MiB
	MemoryLimitMiB int `json:"memoryLimitMiB"`
	// MemoryReservationMiB defines the amount of reserved memory resources reserved in MiB
	MemoryReservationMiB int `json:"memoryReservationMiB"`
	// MemoryReservationUsedMiB defines the amount of reserved memory resources used in MiB. For
	// Tenants, this value represents the total given to all of a Tenant's Namespaces. For
	// Providers, this value represents the total given to all Tenants
	MemoryReservationUsedMiB int `json:"memoryReservationUsedMiB"`
	// MemoryUsedMiB defines the total amount of reserved and unreserved memory resources used in
	// MiB. For Tenants, this value represents the total given to all of a Tenant's Namespaces. For
	// Providers, this value represents the total given to all Tenants
	MemoryUsedMiB int `json:"memoryUsedMiB"`
}

// RegionVirtualMachineClass represents virtual machine sizing configurations information including cpu, memory
type RegionVirtualMachineClass struct {
	ID                   string            `json:"id,omitempty"`
	Region               *OpenApiReference `json:"region,omitempty"`
	Name                 string            `json:"name,omitempty"`
	CpuReservationMHz    int               `json:"cpuReservationMHz,omitempty"`
	MemoryReservationMiB int               `json:"memoryReservationMiB,omitempty"`
	CpuCount             int               `json:"cpuCount,omitempty"`
	MemoryMiB            int               `json:"memoryMiB,omitempty"`
	Reserved             bool              `json:"reserved,omitempty"`
}

// RegionVirtualMachineClasses represents a slice of RegionVirtualMachineClass
type RegionVirtualMachineClasses struct {
	Values OpenApiReferences `json:"values"`
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

// TmTier0Gateway represents NSX-T Tier-0 Gateway that are available for consumption in TM
type TmTier0Gateway struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
	// ParentTier0ID in case this is a Tier 0 Gateway VRF
	ParentTier0ID string `json:"parentTier0Id"`
	// AlreadyImported displays if the Tier 0 Gateway is already consumed by TM
	AlreadyImported bool `json:"alreadyImported"`
}

// TmProviderGateway reflects a TM Provider Gateway
type TmProviderGateway struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// OrgRef contains a reference to Org
	OrgRef     *OpenApiReference `json:"orgRef,omitempty"`
	BackingRef OpenApiReference  `json:"backingRef,omitempty"`
	// BackingType - NSX_TIER0
	BackingType string `json:"backingType,omitempty"`
	// RegionRef contains Region reference
	RegionRef OpenApiReference `json:"regionRef,omitempty"`
	// IPSpaceRefs - a list of IP Space references to create associations with.
	// NOTE. It is used _only_ for creation. Reading will return it empty, and update will not work
	// - one must use `TmIpSpaceAssociation` to update IP Space associations with Provider Gateway
	IPSpaceRefs []OpenApiReference `json:"ipSpaceRefs,omitempty"`
	// Represents current status of the networking entity. Possible values are:
	// * PENDING - Desired entity configuration has been received by system and is pending realization.
	// * CONFIGURING - The system is in process of realizing the entity.
	// * REALIZED - The entity is successfully realized in the system.
	// * REALIZATION_FAILED - There are some issues and the system is not able to realize the entity.
	// * UNKNOWN - Current state of entity is unknown.
	Status string `json:"status,omitempty"`
}

// TmIpSpaceAssociation manages IP Space and Provider Gateway associations
type TmIpSpaceAssociation struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	// IPSpaceRef must contain an IP Space reference that will be associated with Provider Gateway
	IPSpaceRef *OpenApiReference `json:"ipSpaceRef"`
	// ProviderGatewayRef must contain a Provider Gateway reference that will be association with an
	// IP Space
	ProviderGatewayRef *OpenApiReference `json:"providerGatewayRef"`
	// Represents current status of the networking entity. Possible values are:
	// * PENDING - Desired entity configuration has been received by system and is pending realization.
	// * CONFIGURING - The system is in process of realizing the entity.
	// * REALIZED - The entity is successfully realized in the system.
	// * REALIZATION_FAILED - There are some issues and the system is not able to realize the entity.
	// * UNKNOWN - Current state of entity is unknown.
	Status string `json:"status,omitempty"`
}

// TmEdgeCluster defines NSX-T Edge Cluster representation structure within TM
type TmEdgeCluster struct {
	ID string `json:"id,omitempty"`
	// Display name for the Edge Cluster
	Name string `json:"name,omitempty"`
	// Description for the Edge Cluster
	Description string            `json:"description,omitempty"`
	RegionRef   *OpenApiReference `json:"regionRef,omitempty"`
	// Deployment type for transport nodes in the Edge Cluster. Possible values are:
	// * VIRTUAL_MACHINE - If all members are of type VIRTUAL_MACHINE
	// * PHYSICAL_MACHINE - If all members are of type PHYSICAL_MACHINE
	// * UNKNOWN - If there are no members or their type is not known
	DeploymentType string `json:"deploymentType,omitempty"`
	// NodeCount contains number of transport nodes in the Edge Cluster. If this information is not
	// available, nodeCount will be set to -1
	NodeCount int `json:"nodeCount,omitempty"`
	// Number of Organizations using the Edge Cluster
	OrgCount int `json:"orgCount,omitempty"`
	// Number of VPCs using the Edge Cluster
	VpcCount int `json:"vpcCount,omitempty"`
	// Average CPU utilization across all member nodes. This is inclusive of both Data plane and
	// Service CPU cores across all the member nodes
	AvgCPUUsagePercentage float64 `json:"avgCpuUsagePercentage,omitempty"`
	// Average RAM utilization across all member nodes
	AvgMemoryUsagePercentage float64 `json:"avgMemoryUsagePercentage,omitempty"`
	// The current health status of the Edge Cluster. Possible values are:
	// * UP - The Edge Cluster is healthy
	// * DOWN - The Edge Cluster is down
	// * DEGRADED - The Edge Cluster is not operating at capacity. One or more member nodes are down or inactive
	// * UNKNOWN - The Edge Cluster state is unknown. If UNKNOWN, avgMemoryUsagePercentage and avgCpuUsagePercentage will be not be set
	HealthStatus string `json:"healthStatus,omitempty"`
	// The default ingress and egress QoS config associated with this Edge Cluster. This will be
	// used to configure default QoS profiles when the cluster is associated with the organization
	// through a Regional Networking Assignment
	DefaultQosConfig TmEdgeClusterDefaultQosConfig `json:"defaultQosConfig"`
	// BackingRef contains reference to the backing NSX edge cluster
	BackingRef *OpenApiReference `json:"backingRef,omitempty"`
	// Status represents current status of the networking entity. Possible values are:
	// * PENDING - Desired entity configuration has been received by system and is pending realization
	// * CONFIGURING - The system is in process of realizing the entity
	// * REALIZED - The entity is successfully realized in the system
	// * REALIZATION_FAILED - There are some issues and the system is not able to realize the entity
	// * UNKNOWN - Current state of entity is unknown
	Status string `json:"status,omitempty"`
}

// The default ingress and egress QoS config associated with this Edge Cluster. This will be used to
// configure default QoS profiles when the cluster is associated with the organization through a
// Regional Networking Assignment
type TmEdgeClusterDefaultQosConfig struct {
	// Gateway QoS profile applicable to Ingress traffic. Setting this property to NULL results in
	// no QoS being applied for traffic in ingress direction
	IngressProfile *TmEdgeClusterQosProfile `json:"ingressProfile"`
	// Gateway QoS profile applicable to Egress traffic. Setting this property to NULL results in no
	// QoS being applied for traffic in egress direction
	EgressProfile *TmEdgeClusterQosProfile `json:"egressProfile"`
}

// TmEdgeClusterQosProfile represents the Gateway QoS profile for egress/ingress traffic
type TmEdgeClusterQosProfile struct {
	// Unique backing ID of the QoS profile in NSX manager backing the region. This is not a Tenant Manager URN
	ID string `json:"id,omitempty"`
	// Name  of the QoS profile in NSX manager backing the region
	Name string `json:"name,omitempty"`
	// Committed bandwidth specified in Mbps. Bandwidth is limited to line rate when the value
	// configured is greater than line rate. Traffic exceeding bandwidth will be dropped
	// Minimum Value: 1
	CommittedBandwidthMbps int `json:"committedBandwidthMbps,omitempty"`
	// Burst size in bytes
	// Minimum Value - 1
	BurstSizeBytes int `json:"burstSizeBytes,omitempty"`
	// Type of the referenced profile.
	// * DEFAULT: The default profile associated with the Edge Cluster
	// * CUSTOM: Custom profile for Organization workloads running within region
	// To override the values and create new profile, set the type to CUSTOM
	Type string `json:"type,omitempty"`
}

// An object representing the status of a member Transport Node of an Edge Cluster
type TmEdgeClusterTransportNodeStatus struct {
	// Average utilization of all the DPDK CPU cores in the system
	AvgDatapathCPUUsagePercentage float64 `json:"avgDatapathCpuUsagePercentage,omitempty"`
	// Average utilization of all the Service CPU cores in the system
	AvgServiceCPUUsagePercentage float64 `json:"avgServiceCpuUsagePercentage,omitempty"`
	// Number of datapath CPU cores in the system. These cores handle fast path packet processing using DPDK
	DatapathCPUCoreCount int `json:"datapathCpuCoreCount,omitempty"`
	// Percentage of memory in use by datapath processes. It is inclusive of heap memory, memory pool and resident memory
	DatapathMemoryUsagePercentage float64 `json:"datapathMemoryUsagePercentage,omitempty"`
	// The current health status of the Edge Node. Possible values are:
	// * UP - The Edge Node is healthy
	// * DOWN - The Edge Node is down
	// * DEGRADED - The Edge Node is not operating at capacity
	// * UNKNOWN - The Edge Node state is unknown
	HealthStatus string `json:"healthStatus,omitempty"`
	// The display name of this Transport Node
	NodeName string `json:"nodeName,omitempty"`
	// Number of Service CPU cores in the system. These cores handle system and layer-7 related processing
	ServiceCPUCoreCount int `json:"serviceCpuCoreCount,omitempty"`
	// Percentage of RAM in use on the edge node
	SystemMemoryUsagePercentage float64 `json:"systemMemoryUsagePercentage,omitempty"`
	// Number of CPU cores in the system
	TotalCPUCoreCount int `json:"totalCpuCoreCount,omitempty"`
}

// TmRegionalNetworkingSetting describes a Regional Networking Setting
type TmRegionalNetworkingSetting struct {
	// ID of the Regional Networking Setting in URN format.
	ID string `json:"id"`
	// Name for the Regional Networking Setting. Name can be entered manually, but will be
	// autogenerated based on org and region if left unset
	Name string `json:"name"`

	// The Organization this Regional Networking Setting belongs to
	OrgRef OpenApiReference `json:"orgRef"`

	// Reference to the associated Provider Gateway for egress
	ProviderGatewayRef OpenApiReference `json:"providerGatewayRef"`

	// The Region this Regional Networking Setting belongs to
	RegionRef OpenApiReference `json:"regionRef"`

	// Reference to the Edge cluster to use for the networking workloads configured within the
	// Region. If the Edge Cluster is not specified, the system will default to the Edge Cluster of
	// the selected Provider Gateway's backing router.
	ServiceEdgeClusterRef *OpenApiReference `json:"serviceEdgeClusterRef"`

	// Status represents current status of the networking entity. Possible values are:
	// * PENDING - Desired entity configuration has been received by system and is pending realization
	// * CONFIGURING - The system is in process of realizing the entity
	// * REALIZED - The entity is successfully realized in the system
	// * REALIZATION_FAILED - There are some issues and the system is not able to realize the entity
	// * UNKNOWN - Current state of entity is unknown
	Status string `json:"status,omitempty"`
}

// VpcConnectivityProfileQosConfig is a type alias for TmEdgeClusterDefaultQosConfig
//
// Note. The structures are identical at the moment, but they are used in different endpoints and
// might drift in future. Having a type alias will allow having different structures without breaking code.
type VpcConnectivityProfileQosConfig = TmEdgeClusterDefaultQosConfig

// VpcConnectivityProfileQosProfile is a type alias for TmEdgeClusterQosProfile
//
// Note. The structures are identical at the moment, but they are used in different endpoints and
// might drift in future. Having a type alias will allow having different structures without breaking code.
type VpcConnectivityProfileQosProfile = TmEdgeClusterQosProfile

// TmRegionalNetworkingVpcConnectivityProfile represents default VPC connectivity profile for
// networking workloads running within the region and Organization specified by Regional Networking
// Setting
type TmRegionalNetworkingVpcConnectivityProfile struct {
	Name                  string                           `json:"name,omitempty"`
	QosConfig             *VpcConnectivityProfileQosConfig `json:"qosConfig,omitempty"`
	ServiceEdgeClusterRef *OpenApiReference                `json:"serviceEdgeClusterRef,omitempty"`
	// ExternalCidrBlocks is a comma separated list of the external IP CIDRs which are available for use by the VPC.
	ExternalCidrBlocks string `json:"externalCidrBlocks,omitempty"`
}
