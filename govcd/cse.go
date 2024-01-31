package govcd

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"strings"
	"time"
)

// supportedCseVersions is a map that contains only the supported CSE versions as keys,
// and its corresponding components versions as a slice of strings. The first string is the VCDKEConfig RDE Type version,
// then the CAPVCD RDE Type version and finally the CAPVCD Behavior version.
// TODO: Is this really necessary? What happens in UI if I have a 1.1.0-1.2.0-1.0.0 (4.2) cluster and then CSE is updated to 4.3?
var supportedCseVersions = map[string][]string{
	"4.2": {
		"1.1.0", // VCDKEConfig RDE Type version
		"1.2.0", // CAPVCD RDE Type version
		"1.0.0", // CAPVCD Behavior version
	},
}

// CseClusterApiProviderCluster is a type for handling ClusterApiProviderVCD (CAPVCD) cluster instances created
// by the Container Service Extension (CSE)
type CseClusterApiProviderCluster struct {
	Capvcd *types.Capvcd
	ID     string
	Owner  string
	Etag   string
	client *Client
	// TODO: Updated fields are inside the YAML file, like if you update the Control Plane replicas, you need to inspect
	// the YAML to get the updated value. Inspecting Capvcd fields will do nothing. So I need to put here this information
	// for convenience.
}

// CseClusterCreateInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to create a Kubernetes cluster.
type CseClusterCreateInput struct {
	Name                    string
	OrganizationId          string
	VdcId                   string
	NetworkId               string
	KubernetesTemplateOvaId string
	CseVersion              string
	ControlPlane            CseControlPlaneCreateInput
	WorkerPools             []CseWorkerPoolCreateInput
	DefaultStorageClass     *CseDefaultStorageClassCreateInput // Optional
	Owner                   string                             // Optional, if not set will pick the current user present in the VCDClient
	ApiToken                string
	NodeHealthCheck         bool
	PodCidr                 string
	ServiceCidr             string
	SshPublicKey            string
	VirtualIpSubnet         string
	AutoRepairOnErrors      bool
}

// CseControlPlaneCreateInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify the Control Plane inside a CseClusterCreateInput object.
type CseControlPlaneCreateInput struct {
	MachineCount      int
	DiskSizeGi        int
	SizingPolicyId    string // Optional
	PlacementPolicyId string // Optional
	StorageProfileId  string // Optional
	Ip                string // Optional
}

// CseWorkerPoolCreateInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify one Worker Pool inside a CseClusterCreateInput object.
type CseWorkerPoolCreateInput struct {
	Name              string
	MachineCount      int
	DiskSizeGi        int
	SizingPolicyId    string // Optional
	PlacementPolicyId string // Optional
	VGpuPolicyId      string // Optional
	StorageProfileId  string // Optional
}

// CseDefaultStorageClassCreateInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify a Default Storage Class inside a CseClusterCreateInput object.
type CseDefaultStorageClassCreateInput struct {
	StorageProfileId string
	Name             string
	ReclaimPolicy    string
	Filesystem       string
}

// CseClusterUpdateInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to update a Kubernetes cluster.
type CseClusterUpdateInput struct {
	KubernetesTemplateOvaId *string
	ControlPlane            *CseControlPlaneUpdateInput
	WorkerPools             *map[string]CseWorkerPoolUpdateInput // Maps a node pool name with its contents
	NewWorkerPools          *[]CseWorkerPoolCreateInput
	NodeHealthCheck         *bool
	AutoRepairOnErrors      *bool
}

// CseControlPlaneUpdateInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify the Control Plane inside a CseClusterUpdateInput object.
type CseControlPlaneUpdateInput struct {
	MachineCount int
}

// CseWorkerPoolUpdateInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify one Worker Pool inside a CseClusterCreateInput object.
type CseWorkerPoolUpdateInput struct {
	MachineCount int
}

//go:embed cse
var cseFiles embed.FS

// CseCreateKubernetesCluster creates a Kubernetes cluster with the data given as input (CseClusterCreateInput). If the given
// timeout is 0, it waits forever for the cluster creation. Otherwise, if the timeout is reached and the cluster is not available,
// it will return an error (the cluster will be left in VCD in any state) and the latest status of the cluster in the returned CseClusterApiProviderCluster.
// If the cluster is created correctly, returns all the data in CseClusterApiProviderCluster.
func (org *Org) CseCreateKubernetesCluster(clusterData CseClusterCreateInput, timeoutMinutes time.Duration) (*CseClusterApiProviderCluster, error) {
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

// CseCreateKubernetesClusterAsync creates a Kubernetes cluster with the data given as input (CseClusterCreateInput), but does not
// wait for the creation process to finish, so it doesn't monitor for any errors during the process. It returns just the ID of
// the created cluster. One can manually check the status of the cluster with GetKubernetesClusterById and the result of this method.
func (org *Org) CseCreateKubernetesClusterAsync(clusterData CseClusterCreateInput) (string, error) {
	goTemplateContents, err := clusterData.toCseClusterCreationGoTemplateContents(org)
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
func (org *Org) CseGetKubernetesClusterById(id string) (*CseClusterApiProviderCluster, error) {
	rde, err := getRdeById(org.client, id)
	if err != nil {
		return nil, err
	}
	// This should be guaranteed by the proper rights, but just in case
	if rde.DefinedEntity.Org.ID != org.Org.ID {
		return nil, fmt.Errorf("could not find any Kubernetes cluster with ID '%s' in Organization '%s': %s", id, org.Org.Name, ErrorEntityNotFound)
	}
	return rde.cseConvertToCapvcdCluster()
}

// Refresh gets the latest information about the receiver cluster and updates its properties.
func (cluster *CseClusterApiProviderCluster) Refresh() error {
	rde, err := getRdeById(cluster.client, cluster.ID)
	if err != nil {
		return err
	}
	refreshed, err := rde.cseConvertToCapvcdCluster()
	if err != nil {
		return err
	}
	cluster.Capvcd = refreshed.Capvcd
	cluster.Etag = refreshed.Etag
	return nil
}

// UpdateWorkerPools executes an update on the receiver cluster to change the existing worker pools.
func (cluster *CseClusterApiProviderCluster) UpdateWorkerPools(input map[string]CseWorkerPoolUpdateInput) error {
	return cluster.Update(CseClusterUpdateInput{
		WorkerPools: &input,
	})
}

// AddWorkerPools executes an update on the receiver cluster to add new worker pools.
func (cluster *CseClusterApiProviderCluster) AddWorkerPools(input []CseWorkerPoolCreateInput) error {
	return cluster.Update(CseClusterUpdateInput{
		NewWorkerPools: &input,
	})
}

// UpdateControlPlane executes an update on the receiver cluster to change the existing control plane.
func (cluster *CseClusterApiProviderCluster) UpdateControlPlane(input CseControlPlaneUpdateInput) error {
	return cluster.Update(CseClusterUpdateInput{
		ControlPlane: &input,
	})
}

// ChangeKubernetesTemplate executes an update on the receiver cluster to change the Kubernetes template of the cluster.
func (cluster *CseClusterApiProviderCluster) ChangeKubernetesTemplate(kubernetesTemplateOvaId string) error {
	return cluster.Update(CseClusterUpdateInput{
		KubernetesTemplateOvaId: &kubernetesTemplateOvaId,
	})
}

// SetHealthCheck executes an update on the receiver cluster to enable or disable the machine health check capabilities.
func (cluster *CseClusterApiProviderCluster) SetHealthCheck(healthCheckEnabled bool) error {
	return cluster.Update(CseClusterUpdateInput{
		NodeHealthCheck: &healthCheckEnabled,
	})
}

// SetAutoRepairOnErrors executes an update on the receiver cluster to change the flag that controls the auto-repair
// capabilities of CSE.
func (cluster *CseClusterApiProviderCluster) SetAutoRepairOnErrors(autoRepairOnErrors bool) error {
	return cluster.Update(CseClusterUpdateInput{
		AutoRepairOnErrors: &autoRepairOnErrors,
	})
}

// Update executes a synchronous update on the receiver cluster to perform a update on any of the allowed parameters of the cluster. If the given
// timeout is 0, it waits forever for the cluster update to finish. Otherwise, if the timeout is reached and the cluster is not available,
// it will return an error (the cluster will be left in VCD in any state) and the latest status of the cluster will be available in the
// receiver CseClusterApiProviderCluster.
func (cluster *CseClusterApiProviderCluster) Update(input CseClusterUpdateInput) error {
	err := cluster.Refresh()
	if err != nil {
		return err
	}
	if cluster.Capvcd.Status.VcdKe.State == "" {
		return fmt.Errorf("can't update a Kubernetes cluster that does not have any state")
	}
	if cluster.Capvcd.Status.VcdKe.State != "provisioned" {
		return fmt.Errorf("can't update a Kubernetes cluster that is not in 'provisioned' state, as it is in '%s'", cluster.Capvcd.Status.VcdKe.State)
	}

	if input.AutoRepairOnErrors != nil {
		cluster.Capvcd.Spec.VcdKe.AutoRepairOnErrors = *input.AutoRepairOnErrors
	}
	updatedCapiYaml, err := cseUpdateCapiYaml(cluster.client, cluster.Capvcd.Spec.CapiYaml, input)
	if err != nil {
		return err
	}
	cluster.Capvcd.Spec.CapiYaml = updatedCapiYaml

	marshaledPayload, err := json.Marshal(cluster.Capvcd)
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
func (cluster *CseClusterApiProviderCluster) Delete(timeoutMinutes time.Duration) error {
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

// cseConvertToCapvcdCluster takes the receiver, which is a generic RDE that must represent an existing CSE Kubernetes cluster,
// and transforms it to a specific Container Service Extension CAPVCD object that represents the same cluster, but
// it is easy to explore and consume. If the receiver object does not contain a CAPVCD object, this method
// will obviously return an error.
func (rde *DefinedEntity) cseConvertToCapvcdCluster() (*CseClusterApiProviderCluster, error) {
	requiredType := "vmware:capvcdCluster"

	if !strings.Contains(rde.DefinedEntity.ID, requiredType) || !strings.Contains(rde.DefinedEntity.EntityType, requiredType) {
		return nil, fmt.Errorf("the receiver RDE is not a '%s' entity, it is '%s'", requiredType, rde.DefinedEntity.EntityType)
	}

	entityBytes, err := json.Marshal(rde.DefinedEntity.Entity)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the RDE contents to create a Capvcd instance: %s", err)
	}

	result := &CseClusterApiProviderCluster{
		Capvcd: &types.Capvcd{},
		ID:     rde.DefinedEntity.ID,
		Etag:   rde.Etag,
		client: rde.client,
	}
	if rde.DefinedEntity.Owner != nil {
		result.Owner = rde.DefinedEntity.Owner.Name
	}

	err = json.Unmarshal(entityBytes, result.Capvcd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the RDE contents to create a Capvcd instance: %s", err)
	}
	return result, nil
}

// waitUntilClusterIsProvisioned waits for the Kubernetes cluster to be in "provisioned" state, either indefinitely (if timeoutMinutes = 0)
// or until this timeout is reached. If the cluster is in "provisioned" state before the given timeout, it returns a CseClusterApiProviderCluster object
// representing the Kubernetes cluster with all the latest information.
// If one of the states of the cluster at a given point is "error", this function also checks whether the cluster has the "Auto Repair on Errors" flag enabled,
// so it keeps waiting if it's true.
// If timeout is reached before the cluster, it returns an error.
func waitUntilClusterIsProvisioned(client *Client, clusterId string, timeoutMinutes time.Duration) (*CseClusterApiProviderCluster, error) {
	var elapsed time.Duration
	logHttpResponse := util.LogHttpResponse
	sleepTime := 30

	// The following loop is constantly polling VCD to retrieve the RDE, which has a big JSON inside, so we avoid filling
	// the log with these big payloads. We use defer to be sure that we restore the initial logging state.
	defer func() {
		util.LogHttpResponse = logHttpResponse
	}()

	start := time.Now()
	var capvcdCluster *CseClusterApiProviderCluster
	for elapsed <= timeoutMinutes*time.Minute || timeoutMinutes == 0 { // If the user specifies timeoutMinutes=0, we wait forever
		util.LogHttpResponse = false
		rde, err := getRdeById(client, clusterId)
		util.LogHttpResponse = logHttpResponse
		if err != nil {
			return nil, err
		}

		capvcdCluster, err = rde.cseConvertToCapvcdCluster()
		if err != nil {
			return nil, err
		}

		switch capvcdCluster.Capvcd.Status.VcdKe.State {
		case "provisioned":
			return capvcdCluster, nil
		case "error":
			// We just finish if auto-recovery is disabled, otherwise we just let CSE fixing things in background
			if !capvcdCluster.Capvcd.Spec.VcdKe.AutoRepairOnErrors {
				// Try to give feedback about what went wrong, which is located in a set of events in the RDE payload
				return capvcdCluster, fmt.Errorf("got an error and 'auto repair on errors' is disabled, aborting. Errors: %s", capvcdCluster.Capvcd.Status.Capvcd.ErrorSet[len(capvcdCluster.Capvcd.Status.Capvcd.ErrorSet)-1].AdditionalDetails.DetailedError)
			}
		}

		util.Logger.Printf("[DEBUG] Cluster '%s' is in '%s' state, will check again in %d seconds", capvcdCluster.ID, capvcdCluster.Capvcd.Status.VcdKe.State, sleepTime)
		elapsed = time.Since(start)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return capvcdCluster, fmt.Errorf("timeout of %d minutes reached, latest cluster state obtained was '%s'", timeoutMinutes, capvcdCluster.Capvcd.Status.VcdKe.State)
}
