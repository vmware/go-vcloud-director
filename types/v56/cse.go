package types

import "time"

// Capvcd (Cluster API Provider for VCD), is a type that represents a Kubernetes cluster inside VCD, that is created and managed
// with the Container Service Extension (CSE)
type Capvcd struct {
	Kind string `json:"kind,omitempty"`
	Spec struct {
		VcdKe struct {
			// NOTE: "Secure" struct needs to be a pointer to avoid overriding with empty values by mistake, as VCD doesn't return RDE fields
			// marked with "x-vcloud-restricted: secure"
			Secure *struct {
				ApiToken string `json:"apiToken,omitempty"`
			} `json:"secure,omitempty"`
			IsVCDKECluster             bool `json:"isVCDKECluster,omitempty"`
			AutoRepairOnErrors         bool `json:"autoRepairOnErrors,omitempty"`
			DefaultStorageClassOptions struct {
				Filesystem             string `json:"filesystem,omitempty"`
				K8SStorageClassName    string `json:"k8sStorageClassName,omitempty"`
				VcdStorageProfileName  string `json:"vcdStorageProfileName,omitempty"`
				UseDeleteReclaimPolicy bool   `json:"useDeleteReclaimPolicy,omitempty"`
			} `json:"defaultStorageClassOptions,omitempty"`
		} `json:"vcdKe,omitempty"`
		CapiYaml string `json:"capiYaml,omitempty"`
	} `json:"spec,omitempty"`
	Status struct {
		Cpi struct {
			Name     string `json:"name,omitempty"`
			Version  string `json:"version,omitempty"`
			EventSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedEvent string `json:"Detailed Event,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"eventSet,omitempty"`
			ErrorSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedError string `json:"Detailed Error,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"errorSet,omitempty"`
		} `json:"cpi,omitempty"`
		Csi struct {
			Name     string `json:"name,omitempty"`
			Version  string `json:"version,omitempty"`
			EventSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedDescription string `json:"Detailed Description,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"eventSet,omitempty"`
			ErrorSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedError string `json:"Detailed Error,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"errorSet,omitempty"`
		} `json:"csi,omitempty"`
		VcdKe struct {
			State    string `json:"state,omitempty"`
			EventSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedEvent string `json:"Detailed Event,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"eventSet,omitempty"`
			ErrorSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedError string `json:"Detailed Error,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"errorSet,omitempty"`
			WorkerId       string `json:"workerId,omitempty"`
			VcdKeVersion   string `json:"vcdKeVersion,omitempty"`
			VcdResourceSet []struct {
				Id   string `json:"id,omitempty"`
				Name string `json:"name,omitempty"`
				Type string `json:"type,omitempty"`
			} `json:"vcdResourceSet,omitempty"`
			HeartbeatString     string `json:"heartbeatString,omitempty"`
			VcdKeInstanceId     string `json:"vcdKeInstanceId,omitempty"`
			HeartbeatTimestamp  string `json:"heartbeatTimestamp,omitempty"`
			DefaultStorageClass struct {
				FileSystem             string `json:"fileSystem,omitempty"`
				K8SStorageClassName    string `json:"k8sStorageClassName,omitempty"`
				VcdStorageProfileName  string `json:"vcdStorageProfileName,omitempty"`
				UseDeleteReclaimPolicy bool   `json:"useDeleteReclaimPolicy,omitempty"`
			} `json:"defaultStorageClass,omitempty"`
		} `json:"vcdKe,omitempty"`
		Capvcd struct {
			Uid   string `json:"uid,omitempty"`
			Phase string `json:"phase,omitempty"`
			// NOTE: "Private" struct needs to be a pointer to avoid overriding with empty values by mistake, as VCD doesn't return RDE fields
			// marked with "x-vcloud-restricted: secure"
			Private *struct {
				KubeConfig string `json:"kubeConfig,omitempty"`
			} `json:"private,omitempty"`
			Upgrade struct {
				Ready   bool `json:"ready,omitempty"`
				Current struct {
					TkgVersion        string `json:"tkgVersion,omitempty"`
					KubernetesVersion string `json:"kubernetesVersion,omitempty"`
				} `json:"current,omitempty"`
			} `json:"upgrade,omitempty"`
			EventSet []struct {
				Name            string    `json:"name,omitempty"`
				OccurredAt      time.Time `json:"occurredAt,omitempty"`
				VcdResourceId   string    `json:"vcdResourceId,omitempty"`
				VcdResourceName string    `json:"vcdResourceName,omitempty"`
			} `json:"eventSet,omitempty"`
			ErrorSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedError string `json:"Detailed Error,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"errorSet,omitempty"`
			NodePool []struct {
				Name              string `json:"name,omitempty"`
				DiskSizeMb        int    `json:"diskSizeMb,omitempty"`
				SizingPolicy      string `json:"sizingPolicy,omitempty"`
				StorageProfile    string `json:"storageProfile,omitempty"`
				DesiredReplicas   int    `json:"desiredReplicas,omitempty"`
				AvailableReplicas int    `json:"availableReplicas,omitempty"`
			} `json:"nodePool,omitempty"`
			ParentUid  string `json:"parentUid,omitempty"`
			K8SNetwork struct {
				Pods struct {
					CidrBlocks []string `json:"cidrBlocks,omitempty"`
				} `json:"pods,omitempty"`
				Services struct {
					CidrBlocks []string `json:"cidrBlocks,omitempty"`
				} `json:"services,omitempty"`
			} `json:"k8sNetwork,omitempty"`
			Kubernetes    string `json:"kubernetes,omitempty"`
			CapvcdVersion string `json:"capvcdVersion,omitempty"`
			VcdProperties struct {
				Site    string `json:"site,omitempty"`
				OrgVdcs []struct {
					Id              string `json:"id,omitempty"`
					Name            string `json:"name,omitempty"`
					OvdcNetworkName string `json:"ovdcNetworkName,omitempty"`
				} `json:"orgVdcs,omitempty"`
				Organizations []struct {
					Id   string `json:"id,omitempty"`
					Name string `json:"name,omitempty"`
				} `json:"organizations,omitempty"`
			} `json:"vcdProperties,omitempty"`
			CapiStatusYaml string `json:"capiStatusYaml,omitempty"`
			VcdResourceSet []struct {
				Id                string `json:"id,omitempty"`
				Name              string `json:"name,omitempty"`
				Type              string `json:"type,omitempty"`
				AdditionalDetails struct {
					VirtualIP string `json:"virtualIP,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"vcdResourceSet,omitempty"`
			ClusterApiStatus struct {
				Phase        string `json:"phase,omitempty"`
				ApiEndpoints []struct {
					Host string `json:"host,omitempty"`
					Port int    `json:"port,omitempty"`
				} `json:"apiEndpoints,omitempty"`
			} `json:"clusterApiStatus,omitempty"`
			CreatedByVersion           string `json:"createdByVersion,omitempty"`
			ClusterResourceSetBindings []struct {
				Kind                   string `json:"kind,omitempty"`
				Name                   string `json:"name,omitempty"`
				Applied                bool   `json:"applied,omitempty"`
				LastAppliedTime        string `json:"lastAppliedTime,omitempty"`
				ClusterResourceSetName string `json:"clusterResourceSetName,omitempty"`
			} `json:"clusterResourceSetBindings,omitempty"`
		} `json:"capvcd,omitempty"`
		Projector struct {
			Name     string `json:"name,omitempty"`
			Version  string `json:"version,omitempty"`
			EventSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					Event string `json:"event,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"eventSet,omitempty"`
			ErrorSet []struct {
				Name              string    `json:"name,omitempty"`
				OccurredAt        time.Time `json:"occurredAt,omitempty"`
				VcdResourceId     string    `json:"vcdResourceId,omitempty"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					DetailedError string `json:"Detailed Error,omitempty"`
				} `json:"additionalDetails,omitempty"`
			} `json:"errorSet,omitempty"`
		} `json:"projector,omitempty"`
	} `json:"status,omitempty"`
	Metadata struct {
		Name                  string `json:"name,omitempty"`
		Site                  string `json:"site,omitempty"`
		OrgName               string `json:"orgName,omitempty"`
		VirtualDataCenterName string `json:"virtualDataCenterName,omitempty"`
	} `json:"metadata,omitempty"`
	ApiVersion string `json:"apiVersion,omitempty"`
}
