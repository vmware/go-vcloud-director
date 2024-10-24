package types

// RegionStoragePolicy defines a Region storage policy
type RegionStoragePolicy struct {
	Id string `json:"id,omitempty"`
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
	// A collection of RegionStoragePolicy or VdcStoragePolicy references used by this Content Library
	StoragePolicies []OpenApiReference `json:"storagePolicies"`
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
	Id string `json:"id,omitempty"`
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
