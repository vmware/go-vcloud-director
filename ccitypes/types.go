package ccitypes

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CciEntityStatus struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`

	Spec   any                                      `json:"spec,omitempty"`
	Status CciEntityStatusSupervisorNamespaceStatus `json:"status,omitempty"`
}

type CciEntityStatusSupervisorNamespaceStatus struct {
	// NamespaceEndpointURL string                                    `json:"namespaceEndpointURL,omitempty"`
	Phase      string                                   `json:"phase,omitempty"`
	Conditions []CciSupervisorNamespaceStatusConditions `json:"conditions,omitempty"`
}

type CciSupervisorNamespaceStatusConditions struct {
	Message  string `json:"message,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Severity string `json:"severity,omitempty"`
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
}

type CciApiError struct {
	APIVersion    string `json:"apiVersion"`
	Code          int    `json:"code"`
	Kind          string `json:"kind"`
	Message       string `json:"message"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Reason        string `json:"reason"`
	Status        string `json:"status"`
}

func (cciApiError CciApiError) Error() string {
	return fmt.Sprintf("error %d: reason: %s, message: %s, status: %s",
		cciApiError.Code, cciApiError.Reason, cciApiError.Message, cciApiError.Status)
}

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
	LimitMiB int64  `json:"limitMiB"`
	Name     string `json:"name"`
}

type SupervisorNamespaceSpecInitialClassConfigOverridesZone struct {
	CpuLimitMHz          int64  `json:"cpuLimitMHz"`
	CpuReservationMHz    int64  `json:"cpuReservationMHz"`
	MemoryLimitMiB       int64  `json:"memoryLimitMiB"`
	MemoryReservationMiB int64  `json:"memoryReservationMiB"`
	Name                 string `json:"name"`
}

type SupervisorNamespaceStatus struct {
	Conditions           []SupervisorNamespaceStatusConditions     `json:"conditions,omitempty"`
	NamespaceEndpointURL string                                    `json:"namespaceEndpointURL,omitempty"`
	Phase                string                                    `json:"phase,omitempty"`
	StorageClasses       []SupervisorNamespaceStatusStorageClasses `json:"storageClasses,omitempty"`
	VMClasses            []SupervisorNamespaceStatusVMClasses      `json:"vmClasses,omitempty"`
	Zones                []SupervisorNamespaceStatusZones          `json:"zones,omitempty"`
}

type SupervisorNamespaceStatusConditions struct {
	Message  string `json:"message,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Severity string `json:"severity,omitempty"`
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
}

type SupervisorNamespaceStatusStorageClasses struct {
	LimitMiB int64  `json:"limitMiB,omitempty"`
	Name     string `json:"name,omitempty"`
}

type SupervisorNamespaceStatusVMClasses struct {
	Name string `json:"name,omitempty"`
}

type SupervisorNamespaceStatusZones struct {
	CpuLimitMHz          int64  `json:"cpuLimitMHz,omitempty"`
	CpuReservationMHz    int64  `json:"cpuReservationMHz,omitempty"`
	MemoryLimitMiB       int64  `json:"memoryLimitMiB,omitempty"`
	MemoryReservationMiB int64  `json:"memoryReservationMiB,omitempty"`
	Name                 string `json:"name,omitempty"`
}
