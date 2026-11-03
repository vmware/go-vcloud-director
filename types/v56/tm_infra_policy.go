// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

// TmInfraPolicies represents a list of infra policy objects.
type TmInfraPolicies struct {
	// How many results there are in total (i.e., considering all pages). Read-only.
	ResultTotal int `json:"resultTotal,omitempty"`
	// How many pages there are in total. Read-only.
	PageCount int `json:"pageCount,omitempty"`
	// The page that was fetched, 1-indexed. Read-only.
	Page int `json:"page,omitempty"`
	// Result count for page that was fetched. Read-only.
	PageSize int `json:"pageSize,omitempty"`
	// Association info for each result. API 33.0+.
	Associations []TmAssociation `json:"associations,omitzero"`
	Values       []TmInfraPolicy `json:"values"`
}

// TmInfraPolicy represents an infra policy object that defines vm to host and other types of affinities.
type TmInfraPolicy struct {
	// The URN of the infra policy. Read-only.
	ID string `json:"id,omitempty"`
	// The name of the infra policy. Must adhere to Kubernetes naming standards (RFC 1123 DNS label format) -
	// must be no longer than 63 characters, start and end with a lowercase letter or number, and may contain
	// lowercase letters, numbers, and hyphens.
	Name string `json:"name"`
	// The description of the infra policy.
	Description string `json:"description"`
	// The name of the vCenter compute policy to be used.
	VcComputePolicyName string `json:"vcComputePolicyName,omitempty"`
	// Indicates if the policy is mandatory.
	IsMandatory *bool `json:"isMandatory,omitempty"`
	// The policy rule object. This cannot be updated after policy creation. Any changes made in this field
	// during update of an existing policy will be ignored.
	PolicyRule *TmPolicyRule `json:"policyRule,omitempty"`
	// The list of region-zone mappings that are compatible with this infra policy. Read-only.
	CompatibleRegionZones []TmRegionZones `json:"compatibleRegionZones,omitzero"`
	// The creation status of the infra policy. Possible values are: PENDING SUCCESS FAILED Read-only.
	CreationStatus string `json:"creationStatus,omitempty"`
	// The ISO-8601 timestamp representing when this infra policy was created. Read-only.
	CreationTime string `json:"creationTime,omitempty"`
	// Indicates whether this VCFA infra policy definition is in sync across all vCenters that participate in
	// this policy. A vCenter participates if it has a matching vCenter compute policy name defined in the VCFA
	// infra policy. Read-only.
	SyncedToVCenters bool `json:"syncedToVCenters,omitempty"`
}

// TmVCenterInfraPolicies represents a list of vCenter infra policy objects.
type TmVCenterInfraPolicies struct {
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
	Values       []TmVCenterInfraPolicy `json:"values"`
}

// TmPolicyRule represents a compute policy rule that defines the criteria against which the policy will be
// evaluated for a VM. A PolicyRule can contain either a workloadPolicyRule or a list of nested PolicyRules
// combined with a boolean operator. The nested rules in the 'rules' field are evaluated using the boolean
// operation specified in the 'operator' field. The workloadPolicyRule is always AND'd together with the
// results of the nested rules.
type TmPolicyRule struct {
	// The boolean operation used to evaluate the elements from the rules field. This field does not apply to
	// the workloadPolicyRule field. The workloadPolicyRule is always AND'd together with the results of the
	// rules field. If unset, defaults to AND.
	Operator string `json:"operator,omitempty"`
	// A list of additional PolicyRules that are evaluated using the boolean operation specified in the
	// operator field. If unset, no additional matching is performed.
	Rules []TmPolicyRule `json:"rules,omitzero"`
	// A workload policy rule that defines the criteria for matching workloads. The workload policy rule is a
	// composite rule that is used to match workloads based on the guest OS, guest OS family, and label
	// selectors. All provided rules must match for the workload to be matched.
	WorkloadPolicyRule *TmWorkloadPolicyRule `json:"workloadPolicyRule,omitempty"`
}

// TmRegionZones represents a mapping of a region to its compatible zones.
type TmRegionZones struct {
	// The region reference.
	Region *OpenApiReference `json:"region,omitempty"`
	// The list of zones in this region that are compatible with the infra policy.
	Zones []OpenApiReference `json:"zones,omitzero"`
}

// TmVCenterInfraPolicy represents the status of an infra policy in a specific vCenter.
type TmVCenterInfraPolicy struct {
	// The reference to the vCenter.
	VirtualCenterRef *OpenApiReference `json:"virtualCenterRef,omitempty"`
	// The status of an infra policy in a specific vCenter.
	Status string `json:"status,omitempty"`
}

// TmWorkloadPolicyRule represents a workload policy rule that defines the criteria for matching workloads.
// The workload policy rule is a composite rule that is used to match workloads based on the guest OS, guest
// OS family, and label selectors. All provided rules must match for the workload to be matched.
type TmWorkloadPolicyRule struct {
	// A rule for matching guest OS.
	GuestOsRule *TmGuestOsRule `json:"guestOsRule,omitempty"`
	// A rule for matching guest OS family.
	GuestOsFamilyRule *TmGuestOsFamilyRule `json:"guestOsFamilyRule,omitempty"`
	// The list of label selector rules for matching labels.
	LabelSelectorRules []TmLabelSelectorRule `json:"labelSelectorRules,omitzero"`
}

// TmGuestOsRule represents a rule for matching guest OS.
type TmGuestOsRule struct {
	// The operator for the guest OS rule.
	Operator string `json:"operator,omitempty"`
	// The guest OS value to match.
	Value string `json:"value,omitempty"`
}

// TmGuestOsFamilyRule represents a rule for matching guest OS family.
type TmGuestOsFamilyRule struct {
	// The operator for the guest OS family rule.
	Operator string `json:"operator,omitempty"`
	// The guest OS family value to match. Supported values are: DARWIN LINUX NETWARE SOLARIS NETWARE WINDOWS
	// OTHER
	Value string `json:"value,omitempty"`
}

// TmLabelSelectorRule represents a rule for matching labels.
type TmLabelSelectorRule struct {
	// The operator for the label selector rule.
	Operator string `json:"operator,omitempty"`
	// The label key to match.
	RuleKey string `json:"ruleKey,omitempty"`
	// The list of label values to match.
	Values []string `json:"values"`
}
