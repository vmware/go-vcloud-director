// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmDistributedVxlanConnections represents a list of Distributed VXLAN Connections.
type TmDistributedVxlanConnections struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation                `json:"associations,omitzero"`
	Values       []TmDistributedVxlanConnection `json:"values"`
}

// TmDistributedVxlanConnection provides EVPN-based external
// connectivity.
type TmDistributedVxlanConnection struct {
	// Represents current status of the networking entity. Possible values are: PENDING - Desired entity
	// configuration has been received by system and is pending realization. CONFIGURING - The system is in
	// process of realizing the entity. REALIZED - The entity is successfully realized in the system.
	// REALIZATION_FAILED - There are some issues and the system is not able to realize the entity. UNKNOWN -
	// Current state of entity is unknown. Read-only.
	Status string `json:"status,omitempty"`
	// The unique identifier for this connection. Read-only.
	ID string `json:"id,omitempty"`
	// The name of the connection.
	Name string `json:"name"`
	// The description of the connection.
	Description string `json:"description"`
	// The Region this Distributed VXLAN Connection belongs to.
	RegionRef *OpenApiReference `json:"regionRef"`
	// The NSX Route Controller ID.
	RouteControllerId string `json:"routeControllerId"`
	// Ethernet Virtual Private Network (EVPN) connectivity type, which defaults to L3 network layer. Read-
	// only.
	ConnectivityType string `json:"connectivityType,omitempty"`
	// Layer 3 VNI for EVPN (1-16777215).
	L3Vni int `json:"l3Vni"`
	// Route distinguisher for EVPN (format ASN:nn or IP:nn).
	RouteDistinguisher string `json:"routeDistinguisher"`
	// List of VRF route targets.
	RouteTargets []TmVrfRouteTargets `json:"routeTargets"`
	// This field is optional and allows the connection to be dedicated to a specific organization. Set
	// automatically when dedicated connection is assigned to an org. Read-only.
	DedicatedOrgRef *OpenApiReference `json:"dedicatedOrgRef,omitempty"`
	// The initial set of IP Blocks associated with this Distributed VXLAN Connection. This property is only
	// applicable and required on creation. After creation, the associations can be managed by IP Block
	// Associations API. This property will be set to null on retrieval and ignored on update.
	IpSpaceRefs []OpenApiReference `json:"ipSpaceRefs,omitzero"`
	// ID of the matching Distributed VXLAN Connection in NSX. This field is read-only and is ignored on
	// create/update.
	BackingId string `json:"backingId,omitempty"`
}

// TmVrfRouteTargets represents VRF route targets for EVPN configuration.
type TmVrfRouteTargets struct {
	// Address family.
	AddressFamily string `json:"addressFamily,omitempty"`
	// Export route targets (e.g., ["65000:100"]).
	ExportRouteTargets []string `json:"exportRouteTargets,omitzero"`
	// Import route targets (e.g., ["65000:100"]).
	ImportRouteTargets []string `json:"importRouteTargets,omitzero"`
}
