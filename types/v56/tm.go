package types

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

type SupervisorZone struct {
	ID string `json:"id"`
	// The name of this zone.
	Name string `json:"name"`
	// The supervisor this zone belongs to.
	Supervisor *OpenApiReference `json:"supervisor"`
	// The vCenter this supervisor zone is associated with.
	VirtualCenter *OpenApiReference `json:"virtualCenter"`
}

type Region struct {
	ID string `json:"id,omitempty"`
	// The name of the region. It must follow RFC 1123 Label Names to conform with Kubernetes standards.
	Name string `json:"name"`
	// The NSX manager for the region.
	NsxManager *OpenApiReference `json:"nsxManager"`
	// Total CPU resources in MHz available to this Region.
	CPUCapacityMHz int `json:"cpuCapacityMHz,omitempty"`
	// Total CPU reservation resources in MHz available to this Region.
	CPUReservationCapacityMHz int `json:"cpuReservationCapacityMHz,omitempty"`
	// The description of the region.
	Description string `json:"description,omitempty"`
	// Whether the region is enabled or not.
	IsEnabled bool `json:"isEnabled,omitempty"`
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

// TmOrg defines structure for creating
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

// TmVdc defines a structure for creating VDCs using OpenAPI endpoint
type TmVdc struct {
	ID string `json:"id,omitempty"`
	// Name of the VDC
	Name string `json:"name"`
	// Description of the VDC
	Description string `json:"description"`
	// IsEnabled defines if the VDC is enabled
	IsEnabled bool `json:"isEnabled"`
	// Org reference
	Org *OpenApiReference `json:"org"`
	// Region reference
	Region *OpenApiReference `json:"region"`
	// Status contains creation status of the VDC
	Status string `json:"status"`
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
	CPULimitMHz          int `json:"cpuLimitMHz"`
	CPUReservationMHz    int `json:"cpuReservationMHz"`
	MemoryLimitMiB       int `json:"memoryLimitMiB"`
	MemoryReservationMiB int `json:"memoryReservationMiB"`
}
