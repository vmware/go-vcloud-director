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
// timeout is 0, it waits forever for the cluster creation. Otherwise, if the timeout is reached and the cluster is not available
// (in "provisioned" state), it will return an error (the cluster will be left in VCD in any state) and the latest status
// of the cluster in the returned CseKubernetesCluster.
// If the cluster is created correctly, returns all the data in CseKubernetesCluster.
func (org *Org) CseCreateKubernetesCluster(clusterData CseClusterSettings, timeoutMinutes time.Duration) (*CseKubernetesCluster, error) {
	clusterId, err := org.CseCreateKubernetesClusterAsync(clusterData)
	if err != nil {
		return nil, err
	}

	cluster, err := waitUntilClusterIsProvisioned(org.client, clusterId, timeoutMinutes)
	if err != nil {
		return cluster, err // Returns the latest status of the cluster
	}

	return cluster, nil
}

// CseCreateKubernetesClusterAsync creates a Kubernetes cluster with the data given as input (CseClusterSettings), but does not
// wait for the creation process to finish, so it doesn't monitor for any errors during the process. It returns just the ID of
// the created cluster. One can manually check the status of the cluster with Org.CseGetKubernetesClusterById and the result of this method.
func (org *Org) CseCreateKubernetesClusterAsync(clusterData CseClusterSettings) (string, error) {
	goTemplateContents, err := cseClusterSettingsToInternal(clusterData, org)
	if err != nil {
		return "", err
	}

	rdeContents, err := getCseKubernetesClusterCreationPayload(goTemplateContents)
	if err != nil {
		return "", err
	}

	rde, err := createRdeAndPoll(org.client, "vmware", "capvcdCluster", supportedCseVersions[clusterData.CseVersion][1], types.DefinedEntity{
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
	rde, err := getRdeById(org.client, id)
	if err != nil {
		return nil, err
	}
	// This should be guaranteed by the proper rights, but just in case
	if rde.DefinedEntity.Org.ID != org.Org.ID {
		return nil, fmt.Errorf("could not find any Kubernetes cluster with ID '%s' in Organization '%s': %s", id, org.Org.Name, ErrorEntityNotFound)
	}
	return cseConvertToCseClusterApiProviderClusterType(rde)
}

// Refresh gets the latest information about the receiver cluster and updates its properties.
func (cluster *CseKubernetesCluster) Refresh() error {
	rde, err := getRdeById(cluster.client, cluster.ID)
	if err != nil {
		return err
	}
	refreshed, err := cseConvertToCseClusterApiProviderClusterType(rde)
	if err != nil {
		return err
	}
	cluster.capvcdType = refreshed.capvcdType
	cluster.Etag = refreshed.Etag
	return nil
}

// UpdateWorkerPools executes an update on the receiver cluster to change the existing worker pools.
func (cluster *CseKubernetesCluster) UpdateWorkerPools(input map[string]CseWorkerPoolUpdateInput) error {
	return cluster.Update(CseClusterUpdateInput{
		WorkerPools: &input,
	})
}

// AddWorkerPools executes an update on the receiver cluster to add new worker pools.
func (cluster *CseKubernetesCluster) AddWorkerPools(input []CseWorkerPoolSettings) error {
	return cluster.Update(CseClusterUpdateInput{
		NewWorkerPools: &input,
	})
}

// UpdateControlPlane executes an update on the receiver cluster to change the existing control plane.
func (cluster *CseKubernetesCluster) UpdateControlPlane(input CseControlPlaneUpdateInput) error {
	return cluster.Update(CseClusterUpdateInput{
		ControlPlane: &input,
	})
}

// ChangeKubernetesTemplate executes an update on the receiver cluster to change the Kubernetes template of the cluster.
func (cluster *CseKubernetesCluster) ChangeKubernetesTemplate(kubernetesTemplateOvaId string) error {
	return cluster.Update(CseClusterUpdateInput{
		KubernetesTemplateOvaId: &kubernetesTemplateOvaId,
	})
}

// SetHealthCheck executes an update on the receiver cluster to enable or disable the machine health check capabilities.
func (cluster *CseKubernetesCluster) SetHealthCheck(healthCheckEnabled bool) error {
	return cluster.Update(CseClusterUpdateInput{
		NodeHealthCheck: &healthCheckEnabled,
	})
}

// SetAutoRepairOnErrors executes an update on the receiver cluster to change the flag that controls the auto-repair
// capabilities of CSE.
func (cluster *CseKubernetesCluster) SetAutoRepairOnErrors(autoRepairOnErrors bool) error {
	return cluster.Update(CseClusterUpdateInput{
		AutoRepairOnErrors: &autoRepairOnErrors,
	})
}

// Update executes a synchronous update on the receiver cluster to perform a update on any of the allowed parameters of the cluster. If the given
// timeout is 0, it waits forever for the cluster update to finish. Otherwise, if the timeout is reached and the cluster is not available,
// it will return an error (the cluster will be left in VCD in any state) and the latest status of the cluster will be available in the
// receiver CseKubernetesCluster.
func (cluster *CseKubernetesCluster) Update(input CseClusterUpdateInput) error {
	err := cluster.Refresh()
	if err != nil {
		return err
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

	logHttpResponse := util.LogHttpResponse
	// The following loop is constantly polling VCD to retrieve the RDE, which has a big JSON inside, so we avoid filling
	// the log with these big payloads. We use defer to be sure that we restore the initial logging state.
	defer func() {
		util.LogHttpResponse = logHttpResponse
	}()

	// We do this loop to increase the chances that the Kubernetes cluster is successfully created, as the Go SDK is
	// "fighting" with the CSE Server
	retries := 0
	maxRetries := 5
	updated := false
	for retries <= maxRetries {
		util.LogHttpResponse = false
		rde, err := getRdeById(cluster.client, cluster.ID)
		util.LogHttpResponse = logHttpResponse
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
	logHttpResponse := util.LogHttpResponse

	// The following loop is constantly polling VCD to retrieve the RDE, which has a big JSON inside, so we avoid filling
	// the log with these big payloads. We use defer to be sure that we restore the initial logging state.
	defer func() {
		util.LogHttpResponse = logHttpResponse
	}()

	var elapsed time.Duration
	start := time.Now()
	vcdKe := map[string]interface{}{}
	for elapsed <= timeoutMinutes*time.Minute || timeoutMinutes == 0 { // If the user specifies timeoutMinutes=0, we wait forever
		util.LogHttpResponse = false
		rde, err := getRdeById(cluster.client, cluster.ID)
		util.LogHttpResponse = logHttpResponse
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
				if strings.Contains(strings.ToLower(err.Error()), "etag") {
					continue // We ignore any ETag error. This just means a clash with the CSE Server, we just try again
					// FIXME: No sleep here
				}
				return fmt.Errorf("could not mark the Kubernetes cluster with ID '%s' to be deleted: %s", cluster.ID, err)
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
