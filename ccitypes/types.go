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
	ClassName                   string                                             `json:"className,omitempty"`
	Description                 string                                             `json:"description,omitempty"`
	InitialClassConfigOverrides SupervisorNamespaceSpecInitialClassConfigOverrides `json:"initialClassConfigOverrides,omitempty"`
	RegionName                  string                                             `json:"regionName,omitempty"`
	VpcName                     string                                             `json:"vpcName,omitempty"`
}

type SupervisorNamespaceSpecInitialClassConfigOverrides struct {
	StorageClasses []SupervisorNamespaceSpecInitialClassConfigOverridesStorageClass `json:"storageClasses,omitempty"`
	Zones          []SupervisorNamespaceSpecInitialClassConfigOverridesZone         `json:"zones,omitempty"`
}

type SupervisorNamespaceSpecInitialClassConfigOverridesStorageClass struct {
	Limit string `json:"limit"`
	Name  string `json:"name"`
}

type SupervisorNamespaceSpecInitialClassConfigOverridesZone struct {
	CpuLimit          string `json:"cpuLimit"`
	CpuReservation    string `json:"cpuReservation"`
	MemoryLimit       string `json:"memoryLimit"`
	MemoryReservation string `json:"memoryReservation"`
	Name              string `json:"name"`
}

type SupervisorNamespaceStatus struct {
	Conditions           []SupervisorNamespaceStatusConditions       `json:"conditions,omitempty"`
	ContentLibraries     []SupervisorNamespaceStatusContentLibraries `json:"contentLibraries,omitempty"`
	NamespaceEndpointURL string                                      `json:"namespaceEndpointURL,omitempty"`
	Phase                string                                      `json:"phase,omitempty"`
	StorageClasses       []SupervisorNamespaceStatusStorageClasses   `json:"storageClasses,omitempty"`
	VMClasses            []SupervisorNamespaceStatusVMClasses        `json:"vmClasses,omitempty"`
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
}

type SupervisorNamespaceStatusStorageClasses struct {
	Limit string `json:"limit,omitempty"`
	Name  string `json:"name,omitempty"`
}

type SupervisorNamespaceStatusVMClasses struct {
	Name string `json:"name,omitempty"`
}

type SupervisorNamespaceStatusZones struct {
	CpuLimit          string `json:"cpuLimit,omitempty"`
	CpuReservation    string `json:"cpuReservation,omitempty"`
	MemoryLimit       string `json:"memoryLimit,omitempty"`
	MemoryReservation string `json:"memoryReservation,omitempty"`
	Name              string `json:"name,omitempty"`
}

type Project struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Spec          ProjectSpec `json:"spec,omitempty"`
}

type ProjectSpec struct {
	Description string `json:"description,omitempty"`
}
