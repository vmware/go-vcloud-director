// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmProjectAccessLevels represents a list of Project Access Levels.
type TmProjectAccessLevels struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation        `json:"associations,omitzero"`
	Values       []TmProjectAccessLevel `json:"values"`
}

// TmProjectAssignments represents a list of Project Assignments.
type TmProjectAssignments struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation       `json:"associations,omitzero"`
	Values       []TmProjectAssignment `json:"values"`
}

// TmProjectAssignment represents a VCF Project Assignment.
type TmProjectAssignment struct {
	// The identifier of the project in URN format.
	ID string `json:"id"`
	// The name of the project.
	Name string `json:"name"`
	// The description of the project.
	Description string `json:"description"`
	// The number of namespaces assigned to the project. Read-only.
	NamespaceCount int `json:"namespaceCount,omitempty"`
}

// TmProjectAssignmentMembers represents a list of Project Assignment Members.
type TmProjectAssignmentMembers struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation             `json:"associations,omitzero"`
	Values       []TmProjectAssignmentMember `json:"values"`
}

// TmProjectAccessLevel represents a Project Access Level.
type TmProjectAccessLevel struct {
	// The identifier of the project access level in URN format.
	ID string `json:"id"`
	// The name of the project access level.
	Name string `json:"name,omitempty"`
}

// TmProjectAssignmentMember represents a Project Assignment Member.
type TmProjectAssignmentMember struct {
	// The entity reference for the project member.
	Member *OpenApiReference `json:"member"`
	// The access levels as entity references for the project member.
	AccessLevels []OpenApiReference `json:"accessLevels"`
}
