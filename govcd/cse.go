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
	_, err := getRemoteFile(&vcdClient.Client, fmt.Sprintf("%s/capiyaml_cluster.tmpl", clusterData.CseVersion))
	if err != nil {
		return nil, err
	}

	_, err = clusterData.toCseClusterCreationPayload(vcdClient)
	if err != nil {
		return nil, err
	}

	err = waitUntilClusterIsProvisioned(vcdClient, "", timeoutMinutes)
	if err != nil {
		return nil, err
	}

	result := &CseClusterApiProviderCluster{
		client: &vcdClient.Client,
		ID:     "",
	}
	return result, nil
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
		Etag:   rde.Etag,
		client: rde.client,
	}

	err = json.Unmarshal(entityBytes, result.Capvcd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the RDE contents to create a Capvcd instance: %s", err)
	}
	return result, nil
}

// waitUntilClusterIsProvisioned waits for the Kubernetes cluster to be in "provisioned" state, either indefinitely (if "operations_timeout_minutes=0")
// or until this timeout is reached. If one of the states is "error", this function also checks whether "auto_repair_on_errors=true" to keep
// waiting.
func waitUntilClusterIsProvisioned(vcdClient *VCDClient, clusterId string, timeoutMinutes time.Duration) error {
	var elapsed time.Duration
	logHttpResponse := util.LogHttpResponse
	sleepTime := 30

	// The following loop is constantly polling VCD to retrieve the RDE, which has a big JSON inside, so we avoid filling
	// the log with these big payloads. We use defer to be sure that we restore the initial logging state.
	defer func() {
		util.LogHttpResponse = logHttpResponse
	}()

	start := time.Now()
	latestState := ""
	for elapsed <= timeoutMinutes*time.Minute || timeoutMinutes == 0 { // If the user specifies operations_timeout_minutes=0, we wait forever
		util.LogHttpResponse = false
		rde, err := vcdClient.GetRdeById(clusterId)
		util.LogHttpResponse = logHttpResponse
		if err != nil {
			return err
		}

		capvcdCluster, err := rde.CseConvertToCapvcdCluster()
		if err != nil {
			return err
		}

		latestState = capvcdCluster.Capvcd.Status.VcdKe.State
		switch latestState {
		case "provisioned":
			return nil
		case "error":
			// We just finish if auto-recovery is disabled, otherwise we just let CSE fixing things in background
			if !capvcdCluster.Capvcd.Spec.VcdKe.AutoRepairOnErrors {
				// Try to give feedback about what went wrong, which is located in a set of events in the RDE payload
				return fmt.Errorf("got an error and 'auto_repair_on_errors=false', aborting. Errors: %s", capvcdCluster.Capvcd.Status.Capvcd.ErrorSet[len(capvcdCluster.Capvcd.Status.Capvcd.ErrorSet)-1].AdditionalDetails.DetailedError)
			}
		}

		util.Logger.Printf("[DEBUG] Cluster '%s' is in '%s' state, will check again in %d seconds", capvcdCluster.ID, latestState, sleepTime)
		elapsed = time.Since(start)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return fmt.Errorf("timeout of %d minutes reached, latest cluster state obtained was '%s'", timeoutMinutes, latestState)
}
