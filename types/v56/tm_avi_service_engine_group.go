// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmAviServiceEngineGroups represents a list of Avi Load Balancer Service Engine Groups.
type TmAviServiceEngineGroups struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation           `json:"associations,omitzero"`
	Values       []TmAviServiceEngineGroup `json:"values"`
}

// TmAviServiceEngineGroup represents an Avi Load Balancer Service Engine Group.
type TmAviServiceEngineGroup struct {
	// The identifier of the Avi Load Balancer Service Engine Group in URN format. Read-only.
	ID string `json:"id,omitempty"`
	// The name of the Avi Load Balancer Service Engine Group. Read-only.
	Name string `json:"name,omitempty"`
	// The Service Engine Group's High Availability mode. ELASTIC_N_PLUS_M_BUFFER - Service Engines will scale
	// out to N active nodes with M nodes as buffer. ELASTIC_ACTIVE_ACTIVE - Active-Active with scale out.
	// LEGACY_ACTIVE_STANDBY - Traditional single Active-Standby configuration Read-only.
	HaMode string `json:"haMode,omitempty"`
	// Number of vCPUs for each of the Service Engine virtual machines. Read-only.
	VCpusPerServiceEngine int `json:"vCpusPerServiceEngine,omitempty"`
	// Amount of memory in MB for each of the Service Engine virtual machines. Read-only.
	MemoryPerServiceEngine int `json:"memoryPerServiceEngine,omitempty"`
	// Number of Service Engines in this group. Read-only.
	NumberOfServiceEngines int `json:"numberOfServiceEngines,omitempty"`
	// The Region this Service Engine Group belongs to. This will be populated when the corresponding NSX
	// Manager is associated with some Region on Provider Management. Read-only.
	RegionRef *OpenApiReference `json:"regionRef,omitempty"`
	// Reference to the backing Avi Load Balancer Service Engine Group. Read-only.
	BackingRef *OpenApiReference `json:"backingRef,omitempty"`
}

// TmAviServiceEngineGroupAssignment represents an Avi Load Balancer Service Engine Group Assignment. This
// assigns a PROVIDER_MANAGED Avi Service Engine Group to an Organization.
type TmAviServiceEngineGroupAssignment struct {
	// The identifier of the Avi Load Balancer Service Engine Group Assignment in URN format. Read-only.
	ID string `json:"id,omitempty"`
	// The Service Engine Group in this assignment.
	ServiceEngineGroupRef *OpenApiReference `json:"serviceEngineGroupRef"`
	// The Organization in this assignment.
	OrgRef *OpenApiReference `json:"orgRef"`
	// Reference to the backing Avi Load Balancer Organization Binding. An Organization binding in Avi is an
	// assignment of Service Engine Group to an Organization. Read-only.
	BackingRef *OpenApiReference `json:"backingRef,omitempty"`
}

// TmAviServiceEngineGroupAssignments represents a list of Avi Load Balancer Service Engine Group Assignments.
type TmAviServiceEngineGroupAssignments struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation                     `json:"associations,omitzero"`
	Values       []TmAviServiceEngineGroupAssignment `json:"values"`
}
