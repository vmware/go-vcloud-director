package types

import "time"

// Capvcd (Cluster API Provider for VCD), is a type that represents a Kubernetes cluster inside VCD, that is created and managed
// with the Container Service Extension (CSE)
type Capvcd struct {
	Kind string `json:"kind"`
	Spec struct {
		VcdKe struct {
			Secure struct {
				ApiToken string `json:"apiToken"`
			} `json:"secure"`
			IsVCDKECluster             bool `json:"isVCDKECluster"`
			AutoRepairOnErrors         bool `json:"autoRepairOnErrors"`
			DefaultStorageClassOptions struct {
				Filesystem             string `json:"filesystem"`
				K8SStorageClassName    string `json:"k8sStorageClassName"`
				VcdStorageProfileName  string `json:"vcdStorageProfileName"`
				UseDeleteReclaimPolicy bool   `json:"useDeleteReclaimPolicy"`
			} `json:"defaultStorageClassOptions"`
		} `json:"vcdKe"`
		CapiYaml string `json:"capiYaml"`
	} `json:"spec"`
	Status struct {
		Cpi struct {
			Name     string `json:"name"`
			Version  string `json:"version"`
			EventSet []struct {
				Name              string    `json:"name"`
				OccurredAt        time.Time `json:"occurredAt"`
				VcdResourceId     string    `json:"vcdResourceId"`
				AdditionalDetails struct {
					DetailedEvent string `json:"Detailed Event"`
				} `json:"additionalDetails"`
			} `json:"eventSet"`
		} `json:"cpi"`
		Csi struct {
			Name     string `json:"name"`
			Version  string `json:"version"`
			EventSet []struct {
				Name              string    `json:"name"`
				OccurredAt        time.Time `json:"occurredAt"`
				AdditionalDetails struct {
					DetailedDescription string `json:"Detailed Description,omitempty"`
				} `json:"additionalDetails"`
			} `json:"eventSet"`
		} `json:"csi"`
		VcdKe struct {
			State    string `json:"state"`
			EventSet []struct {
				Name              string    `json:"name"`
				OccurredAt        time.Time `json:"occurredAt"`
				VcdResourceId     string    `json:"vcdResourceId"`
				AdditionalDetails struct {
					DetailedEvent string `json:"Detailed Event"`
				} `json:"additionalDetails"`
			} `json:"eventSet"`
			WorkerId       string `json:"workerId"`
			VcdKeVersion   string `json:"vcdKeVersion"`
			VcdResourceSet []struct {
				Id   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"vcdResourceSet"`
			HeartbeatString     string `json:"heartbeatString"`
			VcdKeInstanceId     string `json:"vcdKeInstanceId"`
			HeartbeatTimestamp  string `json:"heartbeatTimestamp"`
			DefaultStorageClass struct {
				FileSystem             string `json:"fileSystem"`
				K8SStorageClassName    string `json:"k8sStorageClassName"`
				VcdStorageProfileName  string `json:"vcdStorageProfileName"`
				UseDeleteReclaimPolicy bool   `json:"useDeleteReclaimPolicy"`
			} `json:"defaultStorageClass"`
		} `json:"vcdKe"`
		Capvcd struct {
			Uid     string `json:"uid"`
			Phase   string `json:"phase"`
			Private struct {
				KubeConfig string `json:"kubeConfig"`
			} `json:"private"`
			Upgrade struct {
				Ready   bool `json:"ready"`
				Current struct {
					TkgVersion        string `json:"tkgVersion"`
					KubernetesVersion string `json:"kubernetesVersion"`
				} `json:"current"`
			} `json:"upgrade"`
			EventSet []struct {
				Name              string    `json:"name"`
				OccurredAt        time.Time `json:"occurredAt"`
				VcdResourceId     string    `json:"vcdResourceId"`
				VcdResourceName   string    `json:"vcdResourceName,omitempty"`
				AdditionalDetails struct {
					Event string `json:"event"`
				} `json:"additionalDetails,omitempty"`
			} `json:"eventSet"`
			NodePool []struct {
				Name       string `json:"name"`
				DiskSizeMb int    `json:"diskSizeMb"`
				NodeStatus struct {
					CseTest1WorkerNodePool1774Bdcdbffxcwc4BG9Nh9 string `json:"cse-test1-worker-node-pool-1-774bdcdbffxcwc4b-g9nh9,omitempty"`
					CseTest1WorkerNodePool1774Bdcdbffxcwc4BRx9Wf string `json:"cse-test1-worker-node-pool-1-774bdcdbffxcwc4b-rx9wf,omitempty"`
					CseTest1ControlPlaneNodePool56Jhv            string `json:"cse-test1-control-plane-node-pool-56jhv,omitempty"`
				} `json:"nodeStatus"`
				SizingPolicy      string `json:"sizingPolicy"`
				StorageProfile    string `json:"storageProfile"`
				DesiredReplicas   int    `json:"desiredReplicas"`
				AvailableReplicas int    `json:"availableReplicas"`
			} `json:"nodePool"`
			ParentUid  string `json:"parentUid"`
			K8SNetwork struct {
				Pods struct {
					CidrBlocks []string `json:"cidrBlocks"`
				} `json:"pods"`
				Services struct {
					CidrBlocks []string `json:"cidrBlocks"`
				} `json:"services"`
			} `json:"k8sNetwork"`
			Kubernetes    string `json:"kubernetes"`
			CapvcdVersion string `json:"capvcdVersion"`
			VcdProperties struct {
				Site    string `json:"site"`
				OrgVdcs []struct {
					Id              string `json:"id"`
					Name            string `json:"name"`
					OvdcNetworkName string `json:"ovdcNetworkName"`
				} `json:"orgVdcs"`
				Organizations []struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"organizations"`
			} `json:"vcdProperties"`
			CapiStatusYaml string `json:"capiStatusYaml"`
			VcdResourceSet []struct {
				Id                string `json:"id"`
				Name              string `json:"name"`
				Type              string `json:"type"`
				AdditionalDetails struct {
					VirtualIP string `json:"virtualIP"`
				} `json:"additionalDetails,omitempty"`
			} `json:"vcdResourceSet"`
			ClusterApiStatus struct {
				Phase        string `json:"phase"`
				ApiEndpoints []struct {
					Host string `json:"host"`
					Port int    `json:"port"`
				} `json:"apiEndpoints"`
			} `json:"clusterApiStatus"`
			CreatedByVersion           string `json:"createdByVersion"`
			ClusterResourceSetBindings []struct {
				Kind                   string `json:"kind"`
				Name                   string `json:"name"`
				Applied                bool   `json:"applied"`
				LastAppliedTime        string `json:"lastAppliedTime"`
				ClusterResourceSetName string `json:"clusterResourceSetName"`
			} `json:"clusterResourceSetBindings"`
		} `json:"capvcd"`
		Projector struct {
			Name     string `json:"name"`
			Version  string `json:"version"`
			EventSet []struct {
				Name              string    `json:"name"`
				OccurredAt        time.Time `json:"occurredAt"`
				VcdResourceName   string    `json:"vcdResourceName"`
				AdditionalDetails struct {
					Event string `json:"event"`
				} `json:"additionalDetails"`
			} `json:"eventSet"`
		} `json:"projector"`
	} `json:"status"`
	Metadata struct {
		Name                  string `json:"name"`
		Site                  string `json:"site"`
		OrgName               string `json:"orgName"`
		VirtualDataCenterName string `json:"virtualDataCenterName"`
	} `json:"metadata"`
	ApiVersion string `json:"apiVersion"`
}
