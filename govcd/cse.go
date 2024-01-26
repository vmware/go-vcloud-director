package govcd

import (
	_ "embed"
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
	Etag   string
	client *Client
}

// CseClusterCreationInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to create a Kubernetes cluster.
type CseClusterCreationInput struct {
	Name                    string
	OrganizationId          string
	VdcId                   string
	NetworkId               string
	KubernetesTemplateOvaId string
	CseVersion              string
	ControlPlane            ControlPlaneInput
	WorkerPools             []WorkerPoolInput
	DefaultStorageClass     *DefaultStorageClassInput // Optional
	Owner                   string                    // Optional, if not set will pick the current user present in the VCDClient
	ApiToken                string
	NodeHealthCheck         bool
	PodCidr                 string
	ServiceCidr             string
	SshPublicKey            string
	VirtualIpSubnet         string
	AutoRepairOnErrors      bool
}

// ControlPlaneInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify the Control Plane inside a CseClusterCreationInput object.
type ControlPlaneInput struct {
	MachineCount      int
	DiskSizeGi        int
	SizingPolicyId    string // Optional
	PlacementPolicyId string // Optional
	StorageProfileId  string // Optional
	Ip                string // Optional
}

// WorkerPoolInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify one Worker Pool inside a CseClusterCreationInput object.
type WorkerPoolInput struct {
	Name              string
	MachineCount      int
	DiskSizeGi        int
	SizingPolicyId    string // Optional
	PlacementPolicyId string // Optional
	VGpuPolicyId      string // Optional
	StorageProfileId  string // Optional
}

// DefaultStorageClassInput defines the required elements that the consumer of these Container Service Extension (CSE) methods
// must set in order to specify a Default Storage Class inside a CseClusterCreationInput object.
type DefaultStorageClassInput struct {
	StorageProfileId string
	Name             string
	ReclaimPolicy    string
	Filesystem       string
}

//go:embed cse/tkg_versions.json
var cseTkgVersionsJson []byte

// CseCreateKubernetesCluster creates a Kubernetes cluster with the data given as input (CseClusterCreationInput). If the given
// timeout is 0, it waits forever for the cluster creation. Otherwise, if the timeout is reached and the cluster is not available,
// it will return an error (the cluster will be left in VCD in any state) and the latest status of the cluster in the returned CseClusterApiProviderCluster.
// If the cluster is created correctly, returns all the data in CseClusterApiProviderCluster.
func (vcdClient *VCDClient) CseCreateKubernetesCluster(clusterData CseClusterCreationInput, timeoutMinutes time.Duration) (*CseClusterApiProviderCluster, error) {
	goTemplateContents, err := clusterData.toCseClusterCreationGoTemplateContents(vcdClient)
	if err != nil {
		return nil, err
	}

	rdeContents, err := getCseKubernetesClusterCreationPayload(vcdClient, goTemplateContents)
	if err != nil {
		return nil, err
	}

	rde, err := vcdClient.CreateRde("vmware", "capvcdCluster", supportedCseVersions[clusterData.CseVersion][1], types.DefinedEntity{
		EntityType: goTemplateContents.RdeType.ID,
		Name:       goTemplateContents.Name,
		Entity:     rdeContents,
	}, &TenantContext{
		OrgId:   goTemplateContents.OrganizationId,
		OrgName: goTemplateContents.OrganizationName,
	})
	if err != nil {
		return nil, err
	}

	cluster, err := waitUntilClusterIsProvisioned(vcdClient, rde.DefinedEntity.ID, timeoutMinutes)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

// CseConvertToCapvcdCluster takes the receiver, which is a generic RDE that must represent an existing CSE Kubernetes cluster,
// and transforms it to a specific Container Service Extension CAPVCD object that represents the same cluster, but
// it is easy to explore and consume. If the receiver object does not contain a CAPVCD object, this method
// will obviously return an error.
func (rde *DefinedEntity) CseConvertToCapvcdCluster() (*CseClusterApiProviderCluster, error) {
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

	err = json.Unmarshal(entityBytes, result.Capvcd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the RDE contents to create a Capvcd instance: %s", err)
	}
	return result, nil
}

// Refresh gets the latest information about the receiver cluster and updates its properties.
func (cluster *CseClusterApiProviderCluster) Refresh() error {
	rde, err := getRdeById(cluster.client, cluster.ID)
	if err != nil {
		return err
	}
	refreshed, err := rde.CseConvertToCapvcdCluster()
	if err != nil {
		return err
	}
	cluster.Capvcd = refreshed.Capvcd
	cluster.Etag = refreshed.Etag
	return nil
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

// waitUntilClusterIsProvisioned waits for the Kubernetes cluster to be in "provisioned" state, either indefinitely (if timeoutMinutes = 0)
// or until this timeout is reached. If the cluster is in "provisioned" state before the given timeout, it returns a CseClusterApiProviderCluster object
// representing the Kubernetes cluster with all the latest information.
// If one of the states of the cluster at a given point is "error", this function also checks whether the cluster has the "Auto Repair on Errors" flag enabled,
// so it keeps waiting if it's true.
// If timeout is reached before the cluster, it returns an error.
func waitUntilClusterIsProvisioned(vcdClient *VCDClient, clusterId string, timeoutMinutes time.Duration) (*CseClusterApiProviderCluster, error) {
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
		rde, err := vcdClient.GetRdeById(clusterId)
		util.LogHttpResponse = logHttpResponse
		if err != nil {
			return nil, err
		}

		capvcdCluster, err = rde.CseConvertToCapvcdCluster()
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
