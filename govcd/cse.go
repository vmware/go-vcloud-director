package govcd

import (
	"encoding/json"
	"fmt"
	semver "github.com/hashicorp/go-version"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"strings"
	"time"
)

// CseCreateKubernetesCluster creates a Kubernetes cluster with the data given as input (CseClusterSettings). If the given
// timeout is 0, it waits forever for the cluster creation.
//
// If the timeout is reached and the cluster is not available (in "provisioned" state), it will return a non-nil CseKubernetesCluster
// with only the cluster ID and an error. This means that the cluster will be left in VCD in any state, and it can be retrieved afterward
// with Org.CseGetKubernetesClusterById and the returned ID.
//
// If the cluster is created correctly, returns all the available data in CseKubernetesCluster or an error if some of the fields
// of the created cluster cannot be calculated or retrieved.
func (org *Org) CseCreateKubernetesCluster(clusterData CseClusterSettings, timeout time.Duration) (*CseKubernetesCluster, error) {
	clusterId, err := org.CseCreateKubernetesClusterAsync(clusterData)
	if err != nil {
		return nil, err
	}

	err = waitUntilClusterIsProvisioned(org.client, clusterId, timeout)
	if err != nil {
		return &CseKubernetesCluster{
			client: org.client,
			ID:     clusterId,
		}, err
	}

	return getCseKubernetesClusterById(org.client, clusterId)
}

// CseCreateKubernetesClusterAsync creates a Kubernetes cluster with the data given as input (CseClusterSettings), but does not
// wait for the creation process to finish, so it doesn't monitor for any errors during the process. It returns just the ID of
// the created cluster. One can manually check the status of the cluster with VCDClient.CseGetKubernetesClusterById and the result of this method.
func (org *Org) CseCreateKubernetesClusterAsync(clusterSettings CseClusterSettings) (string, error) {
	if org == nil {
		return "", fmt.Errorf("CseCreateKubernetesClusterAsync cannot be called on a nil Organization receiver")
	}

	tenantContext, err := org.getTenantContext()
	if err != nil {
		return "", fmt.Errorf("error creating the CSE Kubernetes cluster: %s", err)
	}

	cseSubcomponents, err := getCseComponentsVersions(clusterSettings.CseVersion)
	if err != nil {
		return "", err
	}

	internalSettings, err := clusterSettings.toCseClusterSettingsInternal(*org)
	if err != nil {
		return "", fmt.Errorf("error creating the CSE Kubernetes cluster: %s", err)
	}

	payload, err := internalSettings.getUnmarshalledRdePayload()
	if err != nil {
		return "", err
	}

	rde, err := createRdeAndGetFromTask(org.client, cseKubernetesClusterVendor, cseKubernetesClusterNamespace, cseSubcomponents.CapvcdRdeTypeVersion,
		types.DefinedEntity{
			EntityType: internalSettings.RdeType.ID,
			Name:       internalSettings.Name,
			Entity:     payload,
		}, tenantContext)
	if err != nil {
		return "", fmt.Errorf("error creating the CSE Kubernetes cluster: %s", err)
	}

	return rde.DefinedEntity.ID, nil
}

// CseGetKubernetesClusterById retrieves a CSE Kubernetes cluster from VCD by its unique ID
func (vcdClient *VCDClient) CseGetKubernetesClusterById(id string) (*CseKubernetesCluster, error) {
	return getCseKubernetesClusterById(&vcdClient.Client, id)
}

// CseGetKubernetesClustersByName retrieves all the CSE Kubernetes clusters from VCD with the given name that belong to the receiver Organization.
// Note: The clusters retrieved won't have a valid ETag to perform operations on them. Use VCDClient.CseGetKubernetesClusterById for that instead.
func (org *Org) CseGetKubernetesClustersByName(cseVersion semver.Version, name string) ([]*CseKubernetesCluster, error) {
	cseSubcomponents, err := getCseComponentsVersions(cseVersion)
	if err != nil {
		return nil, err
	}

	rdes, err := getRdesByName(org.client, cseKubernetesClusterVendor, cseKubernetesClusterNamespace, cseSubcomponents.CapvcdRdeTypeVersion, name)
	if err != nil {
		return nil, err
	}
	var clusters []*CseKubernetesCluster
	for _, rde := range rdes {
		if rde.DefinedEntity.Org != nil && rde.DefinedEntity.Org.ID == org.Org.ID {
			cluster, err := cseConvertToCseKubernetesClusterType(rde)
			if err != nil {
				return nil, err
			}
			clusters = append(clusters, cluster)
		}
	}
	return clusters, nil
}

// getCseKubernetesClusterById retrieves a CSE Kubernetes cluster from VCD by its unique ID
func getCseKubernetesClusterById(client *Client, clusterId string) (*CseKubernetesCluster, error) {
	rde, err := getRdeById(client, clusterId)
	if err != nil {
		return nil, err
	}
	return cseConvertToCseKubernetesClusterType(rde)
}

// Refresh gets the latest information about the receiver CSE Kubernetes cluster and updates its properties.
// All cached fields such as the supported OVAs list (from CseKubernetesCluster.GetSupportedUpgrades) are also cleared.
func (cluster *CseKubernetesCluster) Refresh() error {
	refreshed, err := getCseKubernetesClusterById(cluster.client, cluster.ID)
	if err != nil {
		return fmt.Errorf("failed refreshing the CSE Kubernetes Cluster: %s", err)
	}
	*cluster = *refreshed
	return nil
}

// GetKubeconfig retrieves the Kubeconfig from an existing CSE Kubernetes cluster that is in provisioned state.
// If refresh=true, it retrieves the latest state of the cluster from VCD before requesting the Kubeconfig.
func (cluster *CseKubernetesCluster) GetKubeconfig(refresh bool) (string, error) {
	if refresh {
		err := cluster.Refresh()
		if err != nil {
			return "", err
		}
	}

	if cluster.State == "" {
		return "", fmt.Errorf("cannot get a Kubeconfig of a Kubernetes cluster that does not have a state (expected 'provisioned')")
	}

	if cluster.State != "provisioned" {
		return "", fmt.Errorf("cannot get a Kubeconfig of a Kubernetes cluster that is not in 'provisioned' state. It is '%s'", cluster.State)
	}

	rde, err := getRdeById(cluster.client, cluster.ID)
	if err != nil {
		return "", err
	}
	versions, err := getCseComponentsVersions(cluster.CseVersion)
	if err != nil {
		return "", err
	}

	// Auxiliary wrapper of the result, as the invocation returns the RDE and
	// what we need is inside of it.
	type invocationResult struct {
		Capvcd types.Capvcd `json:"entity,omitempty"`
	}
	result := invocationResult{}

	err = rde.InvokeBehaviorAndMarshal(fmt.Sprintf("urn:vcloud:behavior-interface:getFullEntity:cse:capvcd:%s", versions.CseInterfaceVersion), types.BehaviorInvocation{}, &result)
	if err != nil {
		return "", fmt.Errorf("could not retrieve the Kubeconfig, the Behavior invocation failed: %s", err)
	}
	if result.Capvcd.Status.Capvcd.Private == nil {
		return "", fmt.Errorf("could not retrieve the Kubeconfig, the Behavior invocation succeeded but the Kubeconfig is nil")
	}
	if result.Capvcd.Status.Capvcd.Private.KubeConfig == "" {
		return "", fmt.Errorf("could not retrieve the Kubeconfig, the Behavior invocation succeeded but the Kubeconfig is empty")
	}
	return result.Capvcd.Status.Capvcd.Private.KubeConfig, nil
}

// UpdateWorkerPools executes an update on the receiver cluster to change the existing Worker Pools.
// The input is a map where the key is the Worker pool unique name, and the value is the update payload for that Worker Pool.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating.
// WARNING: At least one worker pool must have one or more nodes running, otherwise the cluster will be left in an unusable state.
func (cluster *CseKubernetesCluster) UpdateWorkerPools(input map[string]CseWorkerPoolUpdateInput, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		WorkerPools: &input,
	}, refresh)
}

// AddWorkerPools executes an update on the receiver cluster to add new Worker Pools.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating.
func (cluster *CseKubernetesCluster) AddWorkerPools(input []CseWorkerPoolSettings, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		NewWorkerPools: &input,
	}, refresh)
}

// UpdateControlPlane executes an update on the receiver cluster to change the existing control plane.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating.
func (cluster *CseKubernetesCluster) UpdateControlPlane(input CseControlPlaneUpdateInput, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		ControlPlane: &input,
	}, refresh)
}

// GetSupportedUpgrades queries all vApp Templates from VCD, one by one, and returns those that can be used for upgrading the cluster.
// As retrieving all OVAs one by one from VCD is expensive, the first time this method is called the returned OVAs are
// cached to avoid querying VCD again multiple times.
// If refreshOvas=true, this cache is cleared out and this method will query VCD for every vApp Template again.
// Therefore, the refreshOvas flag should be set to true only when VCD has new OVAs that need to be considered or after a cluster upgrade.
// NOTE: Any refresh operation from other methods will cause the cache to be cleared.
func (cluster *CseKubernetesCluster) GetSupportedUpgrades(refreshOvas bool) ([]*types.VAppTemplate, error) {
	if refreshOvas {
		cluster.supportedUpgrades = make([]*types.VAppTemplate, 0)
	}
	if cluster.State != "provisioned" {
		cluster.supportedUpgrades = make([]*types.VAppTemplate, 0)
		return cluster.supportedUpgrades, nil
	}
	if len(cluster.supportedUpgrades) > 0 {
		return cluster.supportedUpgrades, nil
	}

	vAppTemplates, err := queryVappTemplateListWithFilter(cluster.client, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get vApp Templates: %s", err)
	}
	for _, template := range vAppTemplates {
		// We can only know if the vApp Template is a TKGm OVA by inspecting its internals, hence we need to retrieve every one
		// of them one by one. This is an expensive operation, hence the cache.
		vAppTemplate, err := getVAppTemplateById(cluster.client, fmt.Sprintf("urn:vcloud:vapptemplate:%s", extractUuid(template.HREF)))
		if err != nil {
			continue // This means we cannot retrieve it (maybe due to some rights missing), so we cannot use it. We skip it
		}
		targetVersions, err := getTkgVersionBundleFromVAppTemplate(vAppTemplate.VAppTemplate)
		if err != nil {
			continue // This means it's not a TKGm OVA, or it is not supported, so we skip it
		}
		// The OVA can be used if the TKG version is equal to the actual or higher, and the Kubernetes version is at most 1 minor higher.
		if targetVersions.compareTkgVersion(cluster.TkgVersion.String()) >= 0 && targetVersions.kubernetesVersionIsUpgradeableFrom(cluster.KubernetesVersion.String()) {
			cluster.supportedUpgrades = append(cluster.supportedUpgrades, vAppTemplate.VAppTemplate)
		}
	}
	return cluster.supportedUpgrades, nil
}

// UpgradeCluster executes an update on the receiver cluster to upgrade the Kubernetes template of the cluster.
// If the cluster is not upgradeable or the OVA is incorrect, this method will return an error.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating.
func (cluster *CseKubernetesCluster) UpgradeCluster(kubernetesTemplateOvaId string, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		KubernetesTemplateOvaId: &kubernetesTemplateOvaId,
	}, refresh)
}

// SetNodeHealthCheck executes an update on the receiver cluster to enable or disable the machine health check capabilities.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating.
func (cluster *CseKubernetesCluster) SetNodeHealthCheck(healthCheckEnabled bool, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		NodeHealthCheck: &healthCheckEnabled,
	}, refresh)
}

// SetAutoRepairOnErrors executes an update on the receiver cluster to change the flag that controls the auto-repair
// capabilities of CSE. If refresh=true, it retrieves the latest state of the cluster from VCD before updating.
// NOTE: This method can only be used in CSE versions < 4.1.1
func (cluster *CseKubernetesCluster) SetAutoRepairOnErrors(autoRepairOnErrors bool, refresh bool) error {
	return cluster.Update(CseClusterUpdateInput{
		AutoRepairOnErrors: &autoRepairOnErrors,
	}, refresh)
}

// Update executes an update on the receiver CSE Kubernetes Cluster on any of the allowed parameters defined in the input type.
// If refresh=true, it retrieves the latest state of the cluster from VCD before updating.
func (cluster *CseKubernetesCluster) Update(input CseClusterUpdateInput, refresh bool) error {
	if refresh {
		err := cluster.Refresh()
		if err != nil {
			return err
		}
	}

	if cluster.State == "" {
		return fmt.Errorf("can't update a Kubernetes cluster that does not have any state")
	}
	if cluster.State != "provisioned" {
		return fmt.Errorf("can't update a Kubernetes cluster that is not in 'provisioned' state, as it is in '%s'", cluster.capvcdType.Status.VcdKe.State)
	}

	if input.AutoRepairOnErrors != nil && *input.AutoRepairOnErrors != cluster.AutoRepairOnErrors {
		// Since CSE 4.1.1, the AutoRepairOnError toggle can't be modified and is turned off
		// automatically by the CSE Server.

		v411, err := semver.NewVersion("4.1.1")
		if err != nil {
			return err
		}
		if cluster.CseVersion.GreaterThanOrEqual(v411) {
			return fmt.Errorf("the 'Auto Repair on Errors' flag can't be changed after the cluster is created since CSE 4.1.1")
		}
		cluster.capvcdType.Spec.VcdKe.AutoRepairOnErrors = *input.AutoRepairOnErrors
	}

	updatedCapiYaml, err := cluster.updateCapiYaml(input)
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

	// We do this loop to increase the chances that the Kubernetes cluster is successfully updated, as the update operation
	// can clash with the CSE Server updates on the same RDE. If the CSE Server does an update just before we do, the ETag
	// verification will fail, so we must retry.
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
			// If it's an ETag error, we just retry without waiting
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

// Delete deletes a CSE Kubernetes cluster, waiting the specified amount of time. If the timeout is reached, this method
// returns an error, even if the cluster is already marked for deletion.
func (cluster *CseKubernetesCluster) Delete(timeout time.Duration) error {
	var elapsed time.Duration
	start := time.Now()
	markForDelete := false
	forceDelete := false
	for elapsed <= timeout || timeout == 0 { // If the user specifies timeout=0, we wait forever
		rde, err := getRdeById(cluster.client, cluster.ID)
		if err != nil {
			if ContainsNotFound(err) {
				return nil // The RDE is gone, so the process is completed and there's nothing more to do
			}
			return fmt.Errorf("could not retrieve the Kubernetes cluster with ID '%s': %s", cluster.ID, err)
		}

		markForDelete = traverseMapAndGet[bool](rde.DefinedEntity.Entity, "spec.vcdKe.markForDelete", ".")
		forceDelete = traverseMapAndGet[bool](rde.DefinedEntity.Entity, "spec.vcdKe.forceDelete", ".")

		if !markForDelete || !forceDelete {
			// Mark the cluster for deletion
			rde.DefinedEntity.Entity["spec"].(map[string]interface{})["vcdKe"].(map[string]interface{})["markForDelete"] = true
			rde.DefinedEntity.Entity["spec"].(map[string]interface{})["vcdKe"].(map[string]interface{})["forceDelete"] = true
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
	if markForDelete && forceDelete {
		return fmt.Errorf("timeout of %s reached, the cluster was successfully marked for deletion but was not removed in time", timeout)
	}
	return fmt.Errorf("timeout of %s reached, the cluster was not marked for deletion, please try again", timeout)
}
