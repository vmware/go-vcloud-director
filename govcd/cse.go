package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"strings"
	"time"
)

// CseCreateKubernetesCluster creates a Kubernetes cluster with the data given as input (CseClusterSettings). If the given
// timeout is 0, it waits forever for the cluster creation.
//
// If the timeout is reached and the cluster is not available (in "provisioned" state), it will return a non-nil CseKubernetesCluster
// with only the cluster ID and an error. This means that the cluster will be left in VCD in any state, and it can be retrieved with
// Org.CseGetKubernetesClusterById manually.
//
// If the cluster is created correctly, returns all the available data in CseKubernetesCluster or an error if some of the fields
// of the created cluster cannot be calculated or retrieved.
func (org *Org) CseCreateKubernetesCluster(clusterData CseClusterSettings, timeoutMinutes time.Duration) (*CseKubernetesCluster, error) {
	clusterId, err := org.CseCreateKubernetesClusterAsync(clusterData)
	if err != nil {
		return nil, err
	}

	err = waitUntilClusterIsProvisioned(org.client, clusterId, timeoutMinutes)
	if err != nil {
		return &CseKubernetesCluster{ID: clusterId}, err
	}

	return org.CseGetKubernetesClusterById(clusterId)
}

// CseCreateKubernetesClusterAsync creates a Kubernetes cluster with the data given as input (CseClusterSettings), but does not
// wait for the creation process to finish, so it doesn't monitor for any errors during the process. It returns just the ID of
// the created cluster. One can manually check the status of the cluster with Org.CseGetKubernetesClusterById and the result of this method.
func (org *Org) CseCreateKubernetesClusterAsync(clusterData CseClusterSettings) (string, error) {
	if org == nil {
		return "", fmt.Errorf("receiver Organization is nil")
	}

	goTemplateContents, err := cseClusterSettingsToInternal(clusterData, *org)
	if err != nil {
		return "", err
	}

	rdeContents, err := getCseKubernetesClusterCreationPayload(goTemplateContents)
	if err != nil {
		return "", err
	}

	cseSubcomponents, err := getCseComponentsVersions(clusterData.CseVersion)
	if err != nil {
		return "", err
	}

	rde, err := createRdeAndPoll(org.client, "vmware", "capvcdCluster", cseSubcomponents.CapvcdRdeTypeVersion, types.DefinedEntity{
		EntityType: goTemplateContents.RdeType.ID,
		Name:       goTemplateContents.Name,
		Entity:     rdeContents,
	}, &TenantContext{
		OrgId:   org.Org.ID,
		OrgName: org.Org.Name,
	})
	if err != nil {
		return "", err
	}

	return rde.DefinedEntity.ID, nil
}

// CseGetKubernetesClusterById retrieves a CSE Kubernetes cluster from VCD by its unique ID
func (org *Org) CseGetKubernetesClusterById(id string) (*CseKubernetesCluster, error) {
	return getCseKubernetesCluster(org.client, id)
}

// getCseKubernetesCluster retrieves a CSE Kubernetes cluster from VCD by its unique ID
func getCseKubernetesCluster(client *Client, clusterId string) (*CseKubernetesCluster, error) {
	rde, err := getRdeById(client, clusterId)
	if err != nil {
		return nil, err
	}
	return cseConvertToCseKubernetesClusterType(rde)
}

// Refresh gets the latest information about the receiver cluster and updates its properties.
func (cluster *CseKubernetesCluster) Refresh() error {
	refreshed, err := getCseKubernetesCluster(cluster.client, cluster.ID)
	if err != nil {
		return fmt.Errorf("failed refreshing the CSE Kubernetes Cluster: %s", err)
	}
	*cluster = *refreshed
	return nil
}

// GetKubeconfig retrieves the Kubeconfig from an available cluster.
func (cluster *CseKubernetesCluster) GetKubeconfig() (string, error) {
	rde, err := getRdeById(cluster.client, cluster.ID)
	if err != nil {
		return "", err
	}
	versions, err := getCseComponentsVersions(cluster.CseVersion)
	if err != nil {
		return "", err
	}

	// Auxiliary wrapper of the result, as the invocation returns the full RDE.
	type invocationResult struct {
		Capvcd types.Capvcd `json:"entity,omitempty"`
	}
	result := invocationResult{}
	err = rde.InvokeBehaviorAndMarshal(fmt.Sprintf("urn:vcloud:behavior-interface:getFullEntity:cse:capvcd:%s", versions.CseInterfaceVersion), types.BehaviorInvocation{}, &result)
	if err != nil {
		return "", err
	}
	if result.Capvcd.Status.Capvcd.Private.KubeConfig == "" {
		return "", fmt.Errorf("could not retrieve the Kubeconfig from the invocation of the Behavior")
	}
	return result.Capvcd.Status.Capvcd.Private.KubeConfig, nil
}

// UpdateWorkerPools executes an update on the receiver cluster to change the existing worker pools.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating (recommended).
func (cluster *CseKubernetesCluster) UpdateWorkerPools(input map[string]CseWorkerPoolUpdateInput, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		WorkerPools: &input,
	}, refresh)
}

// AddWorkerPools executes an update on the receiver cluster to add new worker pools.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating (recommended).
func (cluster *CseKubernetesCluster) AddWorkerPools(input []CseWorkerPoolSettings, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		NewWorkerPools: &input,
	}, refresh)
}

// UpdateControlPlane executes an update on the receiver cluster to change the existing control plane.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating (recommended).
func (cluster *CseKubernetesCluster) UpdateControlPlane(input CseControlPlaneUpdateInput, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		ControlPlane: &input,
	}, refresh)
}

// ChangeKubernetesTemplate executes an update on the receiver cluster to change the Kubernetes template of the cluster.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating (recommended).
func (cluster *CseKubernetesCluster) ChangeKubernetesTemplate(kubernetesTemplateOvaId string, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		KubernetesTemplateOvaId: &kubernetesTemplateOvaId,
	}, refresh)
}

// SetHealthCheck executes an update on the receiver cluster to enable or disable the machine health check capabilities.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating (recommended).
func (cluster *CseKubernetesCluster) SetHealthCheck(healthCheckEnabled bool, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		NodeHealthCheck: &healthCheckEnabled,
	}, refresh)
}

// SetAutoRepairOnErrors executes an update on the receiver cluster to change the flag that controls the auto-repair
// capabilities of CSE. If refresh=true, it retrieves the latest state of the cluster from VCD before updating (recommended).
func (cluster *CseKubernetesCluster) SetAutoRepairOnErrors(autoRepairOnErrors bool, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		AutoRepairOnErrors: &autoRepairOnErrors,
	}, refresh)
}

// Update executes an update on the receiver CSE Kubernetes Cluster on any of the allowed parameters defined in the input type. If refresh=true,
// it retrieves the latest state of the cluster from VCD before updating (recommended).
func (cluster *CseKubernetesCluster) Update(input CseClusterUpdateInput, refresh bool) error {
	if refresh {
		err := cluster.Refresh()
		if err != nil {
			return err
		}
	}

	if cluster.capvcdType.Status.VcdKe.State == "" {
		return fmt.Errorf("can't update a Kubernetes cluster that does not have any state")
	}
	if cluster.capvcdType.Status.VcdKe.State != "provisioned" {
		return fmt.Errorf("can't update a Kubernetes cluster that is not in 'provisioned' state, as it is in '%s'", cluster.capvcdType.Status.VcdKe.State)
	}

	if input.AutoRepairOnErrors != nil {
		cluster.capvcdType.Spec.VcdKe.AutoRepairOnErrors = *input.AutoRepairOnErrors
	}

	// Computed attributes that are required, such as the VcdKeConfig version
	input.clusterName = cluster.Name
	input.vcdKeConfigVersion = cluster.capvcdType.Status.VcdKe.VcdKeVersion
	input.cseVersion = cluster.CseVersion
	updatedCapiYaml, err := cseUpdateCapiYaml(cluster.client, cluster.capvcdType.Spec.CapiYaml, input)
	if err != nil {
		return err
	}
	cluster.capvcdType.Spec.CapiYaml = updatedCapiYaml

	marshaledPayload, err := json.Marshal(cluster.capvcdType)
	if err != nil {
		return err
	}
	entityContent := map[string]interface{}{}
	err = json.Unmarshal(marshaledPayload, &entityContent)
	if err != nil {
		return err
	}

	// We do this loop to increase the chances that the Kubernetes cluster is successfully created, as the Go SDK is
	// "fighting" with the CSE Server
	retries := 0
	maxRetries := 5
	updated := false
	for retries <= maxRetries {
		rde, err := getRdeById(cluster.client, cluster.ID)
		if err != nil {
			return err
		}

		rde.DefinedEntity.Entity = entityContent
		err = rde.Update(*rde.DefinedEntity)
		if err == nil {
			updated = true
			break
		}
		if err != nil {
			// If it's an ETag error, we just retry
			if !strings.Contains(strings.ToLower(err.Error()), "etag") {
				return err
			}
		}
		retries++
		util.Logger.Printf("[DEBUG] The request to update the Kubernetes cluster '%s' failed due to a ETag lock. Trying again", cluster.ID)
	}

	if !updated {
		return fmt.Errorf("could not update the Kubernetes cluster '%s' after %d retries, due to an ETag lock blocking the operations", cluster.ID, maxRetries)
	}

	return cluster.Refresh()
}

// Delete deletes a CSE Kubernetes cluster, waiting the specified amount of minutes. If the timeout is reached, this method
// returns an error, even if the cluster is already marked for deletion.
func (cluster *CseKubernetesCluster) Delete(timeoutMinutes time.Duration) error {
	var elapsed time.Duration
	start := time.Now()
	vcdKe := map[string]interface{}{}
	for elapsed <= timeoutMinutes*time.Minute || timeoutMinutes == 0 { // If the user specifies timeoutMinutes=0, we wait forever
		rde, err := getRdeById(cluster.client, cluster.ID)
		if err != nil {
			if ContainsNotFound(err) {
				return nil // The RDE is gone, so the process is completed and there's nothing more to do
			}
			return fmt.Errorf("could not retrieve the Kubernetes cluster with ID '%s': %s", cluster.ID, err)
		}

		spec, ok := rde.DefinedEntity.Entity["spec"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("JSON object 'spec' is not correct in the RDE '%s': %s", cluster.ID, err)
		}

		vcdKe, ok = spec["vcdKe"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("JSON object 'spec.vcdKe' is not correct in the RDE '%s': %s", cluster.ID, err)
		}

		if !vcdKe["markForDelete"].(bool) || !vcdKe["forceDelete"].(bool) {
			// Mark the cluster for deletion
			vcdKe["markForDelete"] = true
			vcdKe["forceDelete"] = true
			rde.DefinedEntity.Entity["spec"].(map[string]interface{})["vcdKe"] = vcdKe
			err = rde.Update(*rde.DefinedEntity)
			if err != nil {
				// We ignore any ETag error. This just means a clash with the CSE Server, we just try again
				if !strings.Contains(strings.ToLower(err.Error()), "etag") {
					return fmt.Errorf("could not mark the Kubernetes cluster with ID '%s' to be deleted: %s", cluster.ID, err)
				}
			}
		}

		util.Logger.Printf("[DEBUG] Cluster '%s' is still not deleted, will check again in 10 seconds", cluster.ID)
		time.Sleep(10 * time.Second)
		elapsed = time.Since(start)
	}

	// We give a hint to the user about the deletion process result
	if len(vcdKe) >= 2 && vcdKe["markForDelete"].(bool) && vcdKe["forceDelete"].(bool) {
		return fmt.Errorf("timeout of %v minutes reached, the cluster was successfully marked for deletion but was not removed in time", timeoutMinutes)
	}
	return fmt.Errorf("timeout of %v minutes reached, the cluster was not marked for deletion, please try again", timeoutMinutes)
}
