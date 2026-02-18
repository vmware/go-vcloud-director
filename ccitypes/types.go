// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package ccitypes

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApiError is a structure that matches error interface and is able to const
type ApiError struct {
	v1.Status
}

// Error unwraps the error message to human readable one
func (apiError ApiError) Error() string {
	return fmt.Sprintf("apiVersion: %s code: %d kind %s: reason: %s, message: %s status: %s",
		apiError.APIVersion, apiError.Code, apiError.Kind, apiError.Reason, apiError.Message, apiError.Status.Status)
}

// SupervisorNamespace definition
type SupervisorNamespace struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Spec          SupervisorNamespaceSpec    `json:"spec,omitempty"`
	Status        *SupervisorNamespaceStatus `json:"status,omitempty"`
}

type SupervisorNamespaceSpec struct {
	ClassName            string                                      `json:"className,omitempty"`
	Description          string                                      `json:"description,omitempty"`
	ClassConfigOverrides SupervisorNamespaceSpecClassConfigOverrides `json:"classConfigOverrides,omitempty"`
	InfraPolicyNames     []string                                    `json:"infraPolicyNames,omitempty"`
	RegionName           string                                      `json:"regionName,omitempty"`
	SegName              string                                      `json:"segName,omitempty"`
	SharedSubnetNames    []string                                    `json:"sharedSubnetNames,omitempty"`
	VpcName              string                                      `json:"vpcName,omitempty"`
}

type SupervisorNamespaceSpecClassConfigOverrides struct {
	ContentSources []SupervisorNamespaceSpecClassConfigOverridesContentSources `json:"contentSources,omitempty"`
	StorageClasses []SupervisorNamespaceSpecClassConfigOverridesStorageClass   `json:"storageClasses,omitempty"`
	Zones          []SupervisorNamespaceSpecClassConfigOverridesZone           `json:"zones,omitempty"`
	VmClasses      []SupervisorNamespaceSpecClassConfigOverridesVmClass        `json:"vmClasses,omitempty"`
}

type SupervisorNamespaceSpecClassConfigOverridesContentSources struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SupervisorNamespaceSpecClassConfigOverridesStorageClass struct {
	Limit string `json:"limit"`
	Name  string `json:"name"`
}

type SupervisorNamespaceSpecClassConfigOverridesZone struct {
	CpuLimit            string                                                              `json:"cpuLimit"`
	CpuReservation      string                                                              `json:"cpuReservation"`
	MemoryLimit         string                                                              `json:"memoryLimit"`
	MemoryReservation   string                                                              `json:"memoryReservation"`
	Name                string                                                              `json:"name"`
	VmClassReservations []SupervisorNamespaceSpecClassConfigOverridesZoneVmClassReservation `json:"vmClassReservations"`
}

type SupervisorNamespaceSpecClassConfigOverridesZoneVmClassReservation struct {
	Count       int    `json:"count"`
	VmClassName string `json:"vmClassName"`
}

type SupervisorNamespaceSpecClassConfigOverridesVmClass struct {
	Name string `json:"name"`
}

type SupervisorNamespaceStatus struct {
	Conditions           []SupervisorNamespaceStatusConditions       `json:"conditions,omitempty"`
	ContentLibraries     []SupervisorNamespaceStatusContentLibraries `json:"contentLibraries,omitempty"`
	InfraPolicies        []SupervisorNamespaceStatusInfraPolicies    `json:"infraPolicies,omitempty"`
	NamespaceEndpointURL string                                      `json:"namespaceEndpointURL,omitempty"`
	Phase                string                                      `json:"phase,omitempty"`
	SegName              string                                      `json:"segName,omitempty"`
	SharedSubnetNames    []string                                    `json:"sharedSubnetNames,omitempty"`
	StorageClasses       []SupervisorNamespaceStatusStorageClasses   `json:"storageClasses,omitempty"`
	VMClasses            []SupervisorNamespaceStatusVMClasses        `json:"vmClasses,omitempty"`
	VpcName              string                                      `json:"vpcName,omitempty"`
	Zones                []SupervisorNamespaceStatusZones            `json:"zones,omitempty"`
}

type SupervisorNamespaceStatusConditions struct {
	Message  string `json:"message,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Severity string `json:"severity,omitempty"`
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
}

type SupervisorNamespaceStatusContentLibraries struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type SupervisorNamespaceStatusInfraPolicies struct {
	Mandatory bool   `json:"mandatory,omitempty"`
	Name      string `json:"name,omitempty"`
}

type SupervisorNamespaceStatusStorageClasses struct {
	Limit string `json:"limit,omitempty"`
	Name  string `json:"name,omitempty"`
}

type SupervisorNamespaceStatusVMClasses struct {
	Name string `json:"name,omitempty"`
}

type SupervisorNamespaceStatusZones struct {
	CpuLimit            string                                             `json:"cpuLimit,omitempty"`
	CpuReservation      string                                             `json:"cpuReservation,omitempty"`
	MarkedForRemoval    bool                                               `json:"markedForRemoval,omitempty"`
	MemoryLimit         string                                             `json:"memoryLimit,omitempty"`
	MemoryReservation   string                                             `json:"memoryReservation,omitempty"`
	Name                string                                             `json:"name,omitempty"`
	VmClassReservations []SupervisorNamespaceStatusZonesVmClassReservation `json:"vmClassReservations,omitempty"`
}

type SupervisorNamespaceStatusZonesVmClassReservation struct {
	Count       int    `json:"count,omitempty"`
	VmClassName string `json:"vmClassName,omitempty"`
}

type Project struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Spec          ProjectSpec `json:"spec,omitempty"`
}

type ProjectSpec struct {
	Description string `json:"description,omitempty"`
}
