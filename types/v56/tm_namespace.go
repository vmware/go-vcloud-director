// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmNamespaceSummaries represents a list of Namespace Summary Objects.
type TmNamespaceSummaries struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation      `json:"associations,omitzero"`
	Values       []TmNamespaceSummary `json:"values"`
}

// TmNamespace represents provider Management's view of a vSphere namespace including the lists of storage
// policies and virtual machine classes.
type TmNamespace struct {
	// The identifier of the namespace in URN format. Read-only.
	ID string `json:"id,omitempty"`
	// The name of the namespace. It must follow RFC 1123 Label Names to conform with Kubernetes standards.
	Name string `json:"name"`
	// The description for the namespace
	Description string `json:"description"`
	// The project to which the namespace belongs to.
	ProjectAssignment *OpenApiReference `json:"projectAssignment"`
	// Organization to which the namespace belongs to. Read-only.
	Organization *OpenApiReference `json:"organization,omitempty"`
	// Region to which the namespace belongs to.
	Region *OpenApiReference `json:"region"`
	// The name of the Virtual Private Cloud for this namespace to use.
	// This field is read-only and cannot be updated for a namespace.
	VirtualPrivateCloudName string `json:"virtualPrivateCloudName,omitempty"`
	// Optional existing organization network to use as the workload network when using vSphere Networking.
	WorkloadNetwork *OpenApiReference `json:"workloadNetwork,omitempty"`
	// The map of per zone resource allocation.
	ZonalResourceAllocation []TmNamespaceSummaryZonalResourceAllocation `json:"zonalResourceAllocation"`
	// The URL to the Provider Management Supervisor Proxy Read-only.
	ProxyURL string `json:"proxyURL,omitempty"`
	// The status of the namespace. Can be [FAILED, NOT_READY, READY, DELETING, UPDATING]
	Status string `json:"status,omitempty"`
	// vSphere identifiers of supervisor backing this namespace Read-only.
	VcReference *TmNamespaceVcReference `json:"vcReference,omitempty"`
	// SE group name associated with the namespace. API 40.1+.
	SegName string `json:"segName,omitempty"`
	// The last modified time of the namespace. The value will be null if the namespace was never modified.
	// Read-only. API 40.1+. Nil represents null.
	LastModifiedTime *string `json:"lastModifiedTime"`
	// Optional. The supervisor on which this namespace should be created. When provided, placement is pinned
	// to this supervisor instead of being automatically selected. The supervisor must be part of the Region
	// Quota assigned to the organization for the namespace Region. This privileged field is immutable after
	// creation and requires the permissions defined in the schema. API 9.1.2+.
	Supervisor *OpenApiReference `json:"supervisor,omitempty"`
	// Optional. An identifier for system-reserved namespaces. When set, the namespace is managed by the
	// specified system component and is inaccessible to tenant users. This field is immutable after creation.
	// API 9.1.2+.
	ManagedBy string `json:"managedBy,omitempty"`
	// A list of Storage Classes and their corresponding storage limit (in mebibytes)
	StorageClasses []TmNamespaceStorageClass `json:"storageClasses,omitzero"`
	// A list of virtual machine classes that the namespace has access to
	VirtualMachineClasses []OpenApiReference `json:"virtualMachineClasses,omitzero"`
	// A list of content libraries to attach to the created Namespace
	ContentLibraries []OpenApiReference `json:"contentLibraries,omitzero"`
	// A collection of shared subnets for this namespace. This field is bounded to with limit of 2000 shared
	// subnets for a given namespace. API 9.1.1+.
	SharedSubnets []TmNamespaceSharedSubnet `json:"sharedSubnets,omitzero"`
	// A collection of infra policies for this namespace. These policies must be assigned to the region quota
	// first. API 40.1+.
	InfraPolicies []TmNamespaceInfraPolicy `json:"infraPolicies,omitzero"`
}

// TmCapabilities represents a list of read-only capabilities.
type TmCapabilities struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation      `json:"associations,omitzero"`
	Values       []TmRegionCapability `json:"values"`
}

// TmNamespaceImportDetails represents the details needed to import an existing namespace from vCenter.
type TmNamespaceImportDetails struct {
	// The name of the namespace in vCenter.
	Name string `json:"name,omitempty"`
	// The organization reference where the namespace is to be imported.
	Org *OpenApiReference `json:"org,omitempty"`
	// The reference of the vCenter from where the import is to be done.
	Vcenter *OpenApiReference `json:"vcenter,omitempty"`
	// The reference of the project in which the namespace will belong after import.
	ProjectAssignment *OpenApiReference `json:"projectAssignment,omitempty"`
}

// TmNamespaceStorageClassMetrics represents a list of namespace storage class usage objects.
type TmNamespaceStorageClassMetrics struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation                 `json:"associations,omitzero"`
	Values       []TmNamespaceStorageClassMetric `json:"values"`
}

// TmNamespaceStorageClass represents storage classes backing this namespace.
type TmNamespaceStorageClass struct {
	// EntityReference of storage class.
	StorageClass *OpenApiReference `json:"storageClass"`
	// The desired maximum storage that can be used in this storage class (in mebibytes).
	StorageLimitMiB int64 `json:"storageLimitMiB"`
	// realized maximum storage that can be used in this storage class (in mebibytes) Read-only.
	RealizedStorageLimitMiB int64 `json:"realizedStorageLimitMiB,omitempty"`
}

// TmNamespaceSharedSubnet represents a shared subnet assignment to a namespace with optional default
// settings.
type TmNamespaceSharedSubnet struct {
	// The name of the shared subnet.
	SharedSubnetName string `json:"sharedSubnetName"`
	// Whether this shared subnet is the default for pods.
	PodDefault *bool `json:"podDefault,omitempty"`
	// Whether this shared subnet is the default for VMs.
	VmDefault *bool `json:"vmDefault,omitempty"`
	// The VPC associated with the shared subnet. Read-only.
	Vpc string `json:"vpc,omitempty"`
	// The access mode of the shared subnet. Read-only.
	AccessMode string `json:"accessMode,omitempty"`
	// The unique identifier of the shared subnet. Read-only.
	Uid string `json:"uid,omitempty"`
}

// TmNamespaceInfraPolicy represents an infra policy assignment to a namespace with enforcement details.
type TmNamespaceInfraPolicy struct {
	// Reference to the infra policy.
	InfraPolicy *OpenApiReference `json:"infraPolicy"`
	// Indicates who enforced this policy assignment. Possible values are: USER SYSTEM Read-only.
	EnforcedBy string `json:"enforcedBy,omitempty"`
	// The current enforcement status of this policy assignment. Possible values are: PENDING FAILED SUCCESS
	// PENDING_DELETION Read-only.
	EnforcementStatus string `json:"enforcementStatus,omitempty"`
}

// TmNamespaceStorageClassMetric represents namespace storage class metric object.
type TmNamespaceStorageClassMetric struct {
	// The identifier of the namespace in URN format.
	ID string `json:"id,omitempty"`
	// The name of the namespace. It must follow RFC 1123 Label Names to conform with Kubernetes standards.
	Name string `json:"name,omitempty"`
	// Namespace storage class object.
	NamespaceStorageClass *TmNamespaceStorageClass `json:"namespaceStorageClass,omitempty"`
	// Storage class usage in MiB.
	StorageUsedMiB *int64 `json:"storageUsedMiB,omitempty"`
}
