// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmAviController represents an Avi Load Balancer Controller.
type TmAviController struct {
	// Represents current status of the networking entity. Possible values are: PENDING - Desired entity
	// configuration has been received by system and is pending realization. CONFIGURING - The system is in
	// process of realizing the entity. REALIZED - The entity is successfully realized in the system.
	// REALIZATION_FAILED - There are some issues and the system is not able to realize the entity. UNKNOWN -
	// Current state of entity is unknown. Read-only.
	Status string `json:"status,omitempty"`
	// The id of the Avi Load Balancer Controller in URN format. Read-only.
	ID string `json:"id,omitempty"`
	// The name of the Avi Load Balancer Controller. Names for Avi Load Balancer Controllers must be unique
	// across the system.
	Name string `json:"name"`
	// Description for the registered Avi Load Balancer Controller.
	Description string `json:"description"`
	// The URL of the Avi Load Balancer Controller. URLs for Avi Load Balancer Controllers must be unique
	// across the system.
	URL string `json:"url"`
	// Username to connect to the Avi Load Balancer Controller.
	Username string `json:"username"`
	// Cleartext password to connect to the Avi Load Balancer Controller.
	Password string `json:"password"`
	// The version of the Avi Load Balancer Controller. Read-only.
	Version string `json:"version,omitempty"`
	// If set to true, dedicates this Avi Load Balancer Controller for classic tenancy. Read-only.
	IsDedicatedForClassicTenants bool `json:"isDedicatedForClassicTenants,omitempty"`
	// Status of the Avi Load Balancer Controller. Read-only.
	ControllerStatus string `json:"controllerStatus,omitempty"`
	// State of the Avi Load Balancer Controller license. Read-only.
	License string `json:"license,omitempty"`
	// The region associated with the Avi used in Provider Management. Read-only.
	RegionRef *OpenApiReference `json:"regionRef,omitempty"`
}

// TmAviControllers represents a list of Avi Load Balancer Controllers.
type TmAviControllers struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation   `json:"associations,omitzero"`
	Values       []TmAviController `json:"values"`
}
