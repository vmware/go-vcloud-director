// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmVCenterComputePolicies represents a list of vCenter compute policy objects.
type TmVCenterComputePolicies struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation          `json:"associations,omitzero"`
	Values       []TmVCenterComputePolicy `json:"values"`
}

// TmVCenterComputePolicy represents a vCenter compute policy object discovered in VCFA.
type TmVCenterComputePolicy struct {
	// The URN of the vCenter compute policy. Read-only.
	ID string `json:"id,omitempty"`
	// The name of the vCenter compute policy. Read-only.
	Name string `json:"name,omitempty"`
	// The description of the vCenter compute policy. Read-only.
	Description string `json:"description"`
	// The category of the vCenter compute policy. Read-only.
	Category string `json:"category,omitempty"`
	// The capability of the vCenter compute policy. Read-only.
	Capability string `json:"capability,omitempty"`
	// A reference to the vCenter server. Read-only.
	Vcenter *OpenApiReference `json:"vcenter,omitempty"`
}

// TmVCenterComputePolicyTags represents a list of vCenter tag objects.
type TmVCenterComputePolicyTags struct {
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
	Values       []TmVCenterComputePolicyTag `json:"values"`
}

// TmVCenterComputePolicyNames represents a list of unique vCenter compute policy names.
type TmVCenterComputePolicyNames struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation              `json:"associations,omitzero"`
	Values       []TmVCenterComputePolicyName `json:"values"`
}

// TmAssociation represents site association information for an entity.
type TmAssociation struct {
	// ID of the entity.
	EntityID string `json:"entityId,omitempty"`
	// ID of the association.
	AssociationId string `json:"associationId,omitempty"`
}

// TmVCenterComputePolicyTag represents a vCenter tag used by the vCenter compute policy.
type TmVCenterComputePolicyTag struct {
	// The URN of the tag. Read-only.
	ID string `json:"id,omitempty"`
	// The name of the tag. Read-only.
	TagName string `json:"tagName,omitempty"`
	// The type of the tag. Read-only.
	TagType string `json:"tagType,omitempty"`
}

// TmVCenterComputePolicyName represents a vCenter compute policy name object.
type TmVCenterComputePolicyName struct {
	// The name of the vCenter compute policy. Read-only.
	Name string `json:"name,omitempty"`
}
