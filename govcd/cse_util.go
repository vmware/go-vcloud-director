package govcd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	semver "github.com/hashicorp/go-version"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// getCseComponentsVersions gets the versions of the subcomponents that are part of Container Service Extension.
func getCseComponentsVersions(cseVersion semver.Version) (*cseComponentsVersions, error) {
	v41, _ := semver.NewVersion("4.1.0")
	v42, _ := semver.NewVersion("4.2.0")
	v43, _ := semver.NewVersion("4.3.0")

	if cseVersion.GreaterThanOrEqual(v41) && cseVersion.LessThan(v42) {
		return &cseComponentsVersions{
			VcdKeConfigRdeTypeVersion: "1.1.0",
			CapvcdRdeTypeVersion:      "1.2.0",
			CseInterfaceVersion:       "1.0.0",
		}, nil
	} else if cseVersion.GreaterThanOrEqual(v42) && cseVersion.LessThan(v43) {
		return &cseComponentsVersions{
			VcdKeConfigRdeTypeVersion: "1.1.0",
			CapvcdRdeTypeVersion:      "1.3.0",
			CseInterfaceVersion:       "1.0.0",
		}, nil
	}
	return nil, fmt.Errorf("the Container Service Extension version '%s' is not supported", cseVersion.String())
}

// cseConvertToCseKubernetesClusterType takes a generic RDE that must represent an existing CSE Kubernetes cluster,
// and transforms it to an equivalent CseKubernetesCluster object that represents the same cluster, but
// it is easy to explore and consume. If the input RDE is not a CSE Kubernetes cluster, this method
// will obviously return an error.
//
// WARNING: Don't use this method inside loops or avoid calling it multiple times in a row, as it performs many queries
// to VCD.
func cseConvertToCseKubernetesClusterType(rde *DefinedEntity) (*CseKubernetesCluster, error) {
	requiredType := "vmware:capvcdCluster"

	if !strings.Contains(rde.DefinedEntity.ID, requiredType) || !strings.Contains(rde.DefinedEntity.EntityType, requiredType) {
		return nil, fmt.Errorf("the receiver RDE is not a '%s' entity, it is '%s'", requiredType, rde.DefinedEntity.EntityType)
	}

	entityBytes, err := json.Marshal(rde.DefinedEntity.Entity)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the RDE contents to create a capvcdType instance: %s", err)
	}

	capvcd := &types.Capvcd{}
	err = json.Unmarshal(entityBytes, &capvcd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the RDE contents to create a Capvcd instance: %s", err)
	}

	result := &CseKubernetesCluster{
		CseClusterSettings: CseClusterSettings{
			Name:               rde.DefinedEntity.Name,
			ApiToken:           "******", // We can't return this one, we return the "standard" 6-asterisk value
			AutoRepairOnErrors: capvcd.Spec.VcdKe.AutoRepairOnErrors,
			ControlPlane:       CseControlPlaneSettings{},
		},
		ID:                         rde.DefinedEntity.ID,
		Etag:                       rde.Etag,
		ClusterResourceSetBindings: make([]string, len(capvcd.Status.Capvcd.ClusterResourceSetBindings)),
		Upgradeable:                capvcd.Status.Capvcd.Upgrade.Target.KubernetesVersion != "" && capvcd.Status.Capvcd.Upgrade.Target.TkgVersion != "",
		State:                      capvcd.Status.VcdKe.State,
		Events:                     make([]CseClusterEvent, 0),
		client:                     rde.client,
		capvcdType:                 capvcd,
	}
	version, err := semver.NewVersion(capvcd.Status.Capvcd.Upgrade.Current.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("could not read Kubernetes version: %s", err)
	}
	result.KubernetesVersion = *version

	version, err = semver.NewVersion(capvcd.Status.Capvcd.Upgrade.Current.TkgVersion)
	if err != nil {
		return nil, fmt.Errorf("could not read Tkg version: %s", err)
	}
	result.TkgVersion = *version

	version, err = semver.NewVersion(capvcd.Status.Capvcd.CapvcdVersion)
	if err != nil {
		return nil, fmt.Errorf("could not read Capvcd version: %s", err)
	}
	result.CapvcdVersion = *version

	version, err = semver.NewVersion(strings.TrimSpace(capvcd.Status.Cpi.Version)) // Note: We use trim as the version comes with spacing characters
	if err != nil {
		return nil, fmt.Errorf("could not read CPI version: %s", err)
	}
	result.CpiVersion = *version

	version, err = semver.NewVersion(capvcd.Status.Csi.Version)
	if err != nil {
		return nil, fmt.Errorf("could not read CSI version: %s", err)
	}
	result.CsiVersion = *version

	// Retrieve the Organization ID
	for i, binding := range capvcd.Status.Capvcd.ClusterResourceSetBindings {
		result.ClusterResourceSetBindings[i] = binding.Name
	}

	if len(capvcd.Status.Capvcd.ClusterApiStatus.ApiEndpoints) == 0 {
		return nil, fmt.Errorf("could not get Control Plane endpoint")
	}
	result.ControlPlane.Ip = capvcd.Status.Capvcd.ClusterApiStatus.ApiEndpoints[0].Host

	if len(result.capvcdType.Status.Capvcd.VcdProperties.Organizations) == 0 {
		return nil, fmt.Errorf("could not read Organizations from Capvcd type")
	}
	result.OrganizationId = result.capvcdType.Status.Capvcd.VcdProperties.Organizations[0].Id

	// Retrieve the VDC ID
	if len(result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs) == 0 {
		return nil, fmt.Errorf("could not read VDCs from Capvcd type")
	}
	result.VdcId = result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].Id

	// FIXME: This is a workaround, because for some reason the OrgVdcs[*].Id property contains the VDC name instead of the VDC ID.
	//        Once this is fixed, this conditional should not be needed anymore.
	if result.VdcId == result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].Name {
		vdcs, err := queryOrgVdcList(rde.client, map[string]string{})
		if err != nil {
			return nil, fmt.Errorf("could not get VDC IDs as no VDC was found: %s", err)
		}
		found := false
		for _, vdc := range vdcs {
			if vdc.Name == result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].Name {
				result.VdcId = fmt.Sprintf("urn:vcloud:vdc:%s", extractUuid(vdc.HREF))
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("could not get VDC IDs as no VDC with name '%s' was found", result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].Name)
		}
	}

	// Retrieve the Network ID
	params := url.Values{}
	params.Add("filter", fmt.Sprintf("name==%s", result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].OvdcNetworkName))
	params = queryParameterFilterAnd("ownerRef.id=="+result.VdcId, params)
	networks, err := getAllOpenApiOrgVdcNetworks(rde.client, params)
	if err != nil {
		return nil, fmt.Errorf("could not read Org VDC Network from Capvcd type: %s", err)
	}
	if len(networks) != 1 {
		return nil, fmt.Errorf("expected one Org VDC Network from Capvcd type, but got %d", len(networks))
	}
	result.NetworkId = networks[0].OpenApiOrgVdcNetwork.ID

	// Get the CSE version
	cseVersion, err := semver.NewVersion(capvcd.Status.VcdKe.VcdKeVersion)
	if err != nil {
		return nil, fmt.Errorf("could not read the CSE Version that the cluster uses: %s", err)
	}
	result.CseVersion = *cseVersion

	// Retrieve the Owner
	if rde.DefinedEntity.Owner == nil {
		return nil, fmt.Errorf("could not read Owner from RDE")
	}
	result.Owner = rde.DefinedEntity.Owner.Name

	// Here we retrieve several items that we need from now onwards, like Storage Profiles and Compute Policies
	storageProfiles := map[string]string{}
	if rde.client.IsSysAdmin {
		allSp, err := queryAdminOrgVdcStorageProfilesByVdcId(rde.client, result.VdcId)
		if err != nil {
			return nil, fmt.Errorf("could not get all the Storage Profiles: %s", err)
		}
		for _, recordType := range allSp {
			storageProfiles[recordType.Name] = fmt.Sprintf("urn:vcloud:vdcstorageProfile:%s", extractUuid(recordType.HREF))
		}
	} else {
		allSp, err := queryOrgVdcStorageProfilesByVdcId(rde.client, result.VdcId)
		if err != nil {
			return nil, fmt.Errorf("could not get all the Storage Profiles: %s", err)
		}
		for _, recordType := range allSp {
			storageProfiles[recordType.Name] = fmt.Sprintf("urn:vcloud:vdcstorageProfile:%s", extractUuid(recordType.HREF))
		}
	}

	computePolicies, err := getAllVdcComputePoliciesV2(rde.client, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get all the Compute Policies: %s", err)
	}

	if result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.K8SStorageClassName != "" { // This would mean there is a Default Storage Class defined
		result.DefaultStorageClass = &CseDefaultStorageClassSettings{
			Name:          result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.K8SStorageClassName,
			ReclaimPolicy: "retain",
			Filesystem:    result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.Filesystem,
		}
		for spName, spId := range storageProfiles {
			if spName == result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName {
				result.DefaultStorageClass.StorageProfileId = spId
			}
		}
		if result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.UseDeleteReclaimPolicy {
			result.DefaultStorageClass.ReclaimPolicy = "delete"
		}
	}

	// NOTE: We get the remaining elements from the CAPI YAML, despite they are also inside capvcdType.Status.
	// The reason is that any change on the cluster is immediately reflected in the CAPI YAML, but not in the capvcdType.Status
	// elements, which may take more than 10 minutes to be refreshed.
	yamlDocuments, err := unmarshalMultipleYamlDocuments(result.capvcdType.Spec.CapiYaml)
	if err != nil {
		return nil, err
	}

	// We need a map of worker pools and not a slice, because there are two types of YAML documents
	// that contain data about a specific worker pool (VCDMachineTemplate and MachineDeployment), and we can get them in no
	// particular order, so we store the worker pools with their name as key. This way we can easily (O(1)) fetch and update them.
	workerPools := map[string]CseWorkerPoolSettings{}
	for _, yamlDocument := range yamlDocuments {
		switch yamlDocument["kind"] {
		case "KubeadmControlPlane":
			result.ControlPlane.MachineCount = int(traverseMapAndGet[float64](yamlDocument, "spec.replicas"))
			users := traverseMapAndGet[[]interface{}](yamlDocument, "spec.kubeadmConfigSpec.users")
			if len(users) == 0 {
				return nil, fmt.Errorf("expected 'spec.kubeadmConfigSpec.users' slice to not to be empty")
			}
			keys := traverseMapAndGet[[]string](users[0], "sshAuthorizedKeys")
			if len(keys) > 0 {
				result.SshPublicKey = keys[0] // Optional field
			}
		case "VCDMachineTemplate":
			name := traverseMapAndGet[string](yamlDocument, "metadata.name")
			sizingPolicyName := traverseMapAndGet[string](yamlDocument, "spec.template.spec.sizingPolicy")
			placementPolicyName := traverseMapAndGet[string](yamlDocument, "spec.template.spec.placementPolicy")
			storageProfileName := traverseMapAndGet[string](yamlDocument, "spec.template.spec.storageProfile")
			diskSizeGi, err := strconv.Atoi(strings.ReplaceAll(traverseMapAndGet[string](yamlDocument, "spec.template.spec.diskSize"), "Gi", ""))
			if err != nil {
				return nil, err
			}

			if strings.Contains(name, "control-plane-node-pool") {
				// This is the single Control Plane
				for _, policy := range computePolicies {
					if sizingPolicyName == policy.VdcComputePolicyV2.Name && policy.VdcComputePolicyV2.IsSizingOnly {
						result.ControlPlane.SizingPolicyId = policy.VdcComputePolicyV2.ID
					} else if placementPolicyName == policy.VdcComputePolicyV2.Name && !policy.VdcComputePolicyV2.IsSizingOnly {
						result.ControlPlane.PlacementPolicyId = policy.VdcComputePolicyV2.ID
					}
				}
				for spName, spId := range storageProfiles {
					if storageProfileName == spName {
						result.ControlPlane.StorageProfileId = spId
					}
				}

				result.ControlPlane.DiskSizeGi = diskSizeGi

				// We retrieve the Kubernetes Template OVA just once for the Control Plane because all YAML blocks share the same
				vAppTemplateName := traverseMapAndGet[string](yamlDocument, "spec.template.spec.template")
				catalogName := traverseMapAndGet[string](yamlDocument, "spec.template.spec.catalog")
				vAppTemplates, err := queryVappTemplateListWithFilter(rde.client, map[string]string{
					"catalogName": catalogName,
					"name":        vAppTemplateName,
				})
				if err != nil {
					return nil, fmt.Errorf("could not find any vApp Template with name '%s' in Catalog '%s': %s", vAppTemplateName, catalogName, err)
				}
				if len(vAppTemplates) == 0 {
					return nil, fmt.Errorf("could not find any vApp Template with name '%s' in Catalog '%s'", vAppTemplateName, catalogName)
				}
				// The records don't have ID set, so we calculate it
				result.KubernetesTemplateOvaId = fmt.Sprintf("urn:vcloud:vapptemplate:%s", extractUuid(vAppTemplates[0].HREF))
			} else {
				// This is one Worker Pool. We need to check the map of worker pools, just in case we already saved the
				// machine count from MachineDeployment.
				if _, ok := workerPools[name]; !ok {
					workerPools[name] = CseWorkerPoolSettings{}
				}
				workerPool := workerPools[name]
				workerPool.Name = name
				for _, policy := range computePolicies {
					if sizingPolicyName == policy.VdcComputePolicyV2.Name && policy.VdcComputePolicyV2.IsSizingOnly {
						workerPool.SizingPolicyId = policy.VdcComputePolicyV2.ID
					} else if placementPolicyName == policy.VdcComputePolicyV2.Name && !policy.VdcComputePolicyV2.IsSizingOnly && !policy.VdcComputePolicyV2.IsVgpuPolicy {
						workerPool.PlacementPolicyId = policy.VdcComputePolicyV2.ID
					} else if placementPolicyName == policy.VdcComputePolicyV2.Name && !policy.VdcComputePolicyV2.IsSizingOnly && policy.VdcComputePolicyV2.IsVgpuPolicy {
						workerPool.VGpuPolicyId = policy.VdcComputePolicyV2.ID
					}
				}
				for spName, spId := range storageProfiles {
					if storageProfileName == spName {
						workerPool.StorageProfileId = spId
					}
				}
				workerPool.DiskSizeGi = diskSizeGi
				workerPools[name] = workerPool // Override the worker pool with the updated data
			}
		case "MachineDeployment":
			name := traverseMapAndGet[string](yamlDocument, "metadata.name")
			// This is one Worker Pool. We need to check the map of worker pools, just in case we already saved the
			// other information from VCDMachineTemplate.
			if _, ok := workerPools[name]; !ok {
				workerPools[name] = CseWorkerPoolSettings{}
			}
			workerPool := workerPools[name]
			workerPool.MachineCount = int(traverseMapAndGet[float64](yamlDocument, "spec.replicas"))
			workerPools[name] = workerPool // Override the worker pool with the updated data
		case "VCDCluster":
			result.VirtualIpSubnet = traverseMapAndGet[string](yamlDocument, "spec.loadBalancerConfigSpec.vipSubnet")
		case "Cluster":
			cidrBlocks := traverseMapAndGet[[]interface{}](yamlDocument, "spec.clusterNetwork.pods.cidrBlocks")
			if len(cidrBlocks) == 0 {
				return nil, fmt.Errorf("expected at least one 'spec.clusterNetwork.pods.cidrBlocks' item")
			}
			result.PodCidr = cidrBlocks[0].(string)

			cidrBlocks = traverseMapAndGet[[]interface{}](yamlDocument, "spec.clusterNetwork.services.cidrBlocks")
			if len(cidrBlocks) == 0 {
				return nil, fmt.Errorf("expected at least one 'spec.clusterNetwork.services.cidrBlocks' item")
			}
			result.ServiceCidr = cidrBlocks[0].(string)
		case "MachineHealthCheck":
			// This is quite simple, if we find this document, means that Machine Health Check is enabled
			result.NodeHealthCheck = true
		}
	}
	result.WorkerPools = make([]CseWorkerPoolSettings, len(workerPools))
	i := 0
	for _, workerPool := range workerPools {
		result.WorkerPools[i] = workerPool
		i++
	}

	// Add all events to the resulting cluster
	for _, s := range capvcd.Status.VcdKe.EventSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "event",
			OccurredAt: s.OccurredAt,
			Details:    s.AdditionalDetails.DetailedEvent,
		})
	}
	for _, s := range capvcd.Status.VcdKe.ErrorSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "error",
			OccurredAt: s.OccurredAt,
			Details:    s.AdditionalDetails.DetailedError,
		})
	}
	for _, s := range capvcd.Status.Capvcd.EventSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "event",
			OccurredAt: s.OccurredAt,
			Details:    s.Name,
		})
	}
	for _, s := range capvcd.Status.Capvcd.ErrorSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "error",
			OccurredAt: s.OccurredAt,
			Details:    s.AdditionalDetails.DetailedError,
		})
	}
	for _, s := range capvcd.Status.Cpi.EventSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "event",
			OccurredAt: s.OccurredAt,
			Details:    s.Name,
		})
	}
	for _, s := range capvcd.Status.Cpi.ErrorSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "error",
			OccurredAt: s.OccurredAt,
			Details:    s.AdditionalDetails.DetailedError,
		})
	}
	for _, s := range capvcd.Status.Csi.EventSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "event",
			OccurredAt: s.OccurredAt,
			Details:    s.Name,
		})
	}
	for _, s := range capvcd.Status.Csi.ErrorSet {
		result.Events = append(result.Events, CseClusterEvent{
			Name:       s.Name,
			Type:       "error",
			OccurredAt: s.OccurredAt,
			Details:    s.AdditionalDetails.DetailedError,
		})
	}

	sort.SliceStable(result.Events, func(i, j int) bool {
		return result.Events[i].OccurredAt.Before(result.Events[j].OccurredAt)
	})

	return result, nil
}

// waitUntilClusterIsProvisioned waits for the Kubernetes cluster to be in "provisioned" state, either indefinitely (if timeoutMinutes = 0)
// or until the timeout is reached.
// If one of the states of the cluster at a given point is "error", this function also checks whether the cluster has the "AutoRepairOnErrors" flag enabled,
// so it keeps waiting if it's true.
// If timeout is reached before the cluster is in "provisioned" state, it returns an error.
func waitUntilClusterIsProvisioned(client *Client, clusterId string, timeoutMinutes time.Duration) error {
	var elapsed time.Duration
	sleepTime := 10

	start := time.Now()
	capvcd := &types.Capvcd{}
	for elapsed <= timeoutMinutes*time.Minute || timeoutMinutes == 0 { // If the user specifies timeoutMinutes=0, we wait forever
		rde, err := getRdeById(client, clusterId)
		if err != nil {
			return err
		}

		// Here we don't use cseConvertToCseKubernetesClusterType to avoid calling VCD. We only need the state.
		entityBytes, err := json.Marshal(rde.DefinedEntity.Entity)
		if err != nil {
			return fmt.Errorf("could not check the Kubernetes cluster state: %s", err)
		}
		err = json.Unmarshal(entityBytes, &capvcd)
		if err != nil {
			return fmt.Errorf("could not check the Kubernetes cluster state: %s", err)
		}

		switch capvcd.Status.VcdKe.State {
		case "provisioned":
			return nil
		case "error":
			// We just finish if auto-recovery is disabled, otherwise we just let CSE fixing things in background
			if !capvcd.Spec.VcdKe.AutoRepairOnErrors {
				// Give feedback about what went wrong
				errors := ""
				for _, event := range capvcd.Status.Capvcd.ErrorSet {
					errors += fmt.Sprintf("%s,\n", event.AdditionalDetails.DetailedError)
				}
				return fmt.Errorf("got an error and 'AutoRepairOnErrors' is disabled, aborting. Error events:\n%s", errors)
			}
		}

		util.Logger.Printf("[DEBUG] Cluster '%s' is in '%s' state, will check again in %d seconds", rde.DefinedEntity.ID, capvcd.Status.VcdKe.State, sleepTime)
		elapsed = time.Since(start)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return fmt.Errorf("timeout of %d minutes reached, latest cluster state obtained was '%s'", timeoutMinutes, capvcd.Status.VcdKe.State)
}

// validate validates the receiver CseClusterSettings. Returns an error if any of the fields is empty or wrong.
func (input *CseClusterSettings) validate() error {
	if input == nil {
		return fmt.Errorf("the receiver CseClusterSettings cannot be nil")
	}
	// This regular expression is used to validate the constraints placed by Container Service Extension on the names
	// of the components of the Kubernetes clusters:
	// Names must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters.
	cseNamesRegex, err := regexp.Compile(`^[a-z](?:[a-z0-9-]{0,29}[a-z0-9])?$`)
	if err != nil {
		return fmt.Errorf("could not compile regular expression '%s'", err)
	}

	_, err = getCseComponentsVersions(input.CseVersion)
	if err != nil {
		return err
	}
	if !cseNamesRegex.MatchString(input.Name) {
		return fmt.Errorf("the name '%s' must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters", input.Name)
	}
	if input.OrganizationId == "" {
		return fmt.Errorf("the Organization ID is required")
	}
	if input.VdcId == "" {
		return fmt.Errorf("the VDC ID is required")
	}
	if input.NetworkId == "" {
		return fmt.Errorf("the Network ID is required")
	}
	if input.KubernetesTemplateOvaId == "" {
		return fmt.Errorf("the Kubernetes Template OVA ID is required")
	}
	if input.ControlPlane.MachineCount < 1 || input.ControlPlane.MachineCount%2 == 0 {
		return fmt.Errorf("number of Control Plane nodes must be odd and higher than 0, but it was '%d'", input.ControlPlane.MachineCount)
	}
	if input.ControlPlane.DiskSizeGi < 20 {
		return fmt.Errorf("disk size for the Control Plane in Gibibytes (Gi) must be at least 20, but it was '%d'", input.ControlPlane.DiskSizeGi)
	}
	if len(input.WorkerPools) == 0 {
		return fmt.Errorf("there must be at least one Worker Pool")
	}
	for _, workerPool := range input.WorkerPools {
		if workerPool.MachineCount < 1 {
			return fmt.Errorf("number of Worker Pool '%s' nodes must higher than 0, but it was '%d'", workerPool.Name, workerPool.MachineCount)
		}
		if workerPool.DiskSizeGi < 20 {
			return fmt.Errorf("disk size for the Worker Pool '%s' in Gibibytes (Gi) must be at least 20, but it was '%d'", workerPool.Name, workerPool.DiskSizeGi)
		}
		if !cseNamesRegex.MatchString(workerPool.Name) {
			return fmt.Errorf("the Worker Pool name '%s' must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters", workerPool.Name)
		}
	}
	if input.DefaultStorageClass != nil { // This field is optional
		if !cseNamesRegex.MatchString(input.DefaultStorageClass.Name) {
			return fmt.Errorf("the Default Storage Class name '%s' must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters", input.DefaultStorageClass.Name)
		}
		if input.DefaultStorageClass.StorageProfileId == "" {
			return fmt.Errorf("the Storage Profile ID for the Default Storage Class is required")
		}
		if input.DefaultStorageClass.ReclaimPolicy != "delete" && input.DefaultStorageClass.ReclaimPolicy != "retain" {
			return fmt.Errorf("the Reclaim Policy for the Default Storage Class must be either 'delete' or 'retain', but it was '%s'", input.DefaultStorageClass.ReclaimPolicy)
		}
		if input.DefaultStorageClass.Filesystem != "ext4" && input.DefaultStorageClass.ReclaimPolicy != "xfs" {
			return fmt.Errorf("the filesystem for the Default Storage Class must be either 'ext4' or 'xfs', but it was '%s'", input.DefaultStorageClass.Filesystem)
		}
	}
	if input.ApiToken == "" {
		return fmt.Errorf("the API token is required")
	}
	if input.PodCidr == "" {
		return fmt.Errorf("the Pod CIDR is required")
	}
	if input.ServiceCidr == "" {
		return fmt.Errorf("the Service CIDR is required")
	}
	return nil
}

// toCseClusterSettingsInternal transforms user input data (CseClusterSettings) into the final payload that
// will be used to define a Container Service Extension Kubernetes cluster (cseClusterSettingsInternal).
//
// For example, the most relevant transformation is the change of the item IDs that are present in CseClusterSettings
// (such as CseClusterSettings.KubernetesTemplateOvaId) to their corresponding Names (e.g. cseClusterSettingsInternal.KubernetesTemplateOvaName),
// which are the identifiers that Container Service Extension uses internally.
func (input *CseClusterSettings) toCseClusterSettingsInternal(org Org) (*cseClusterSettingsInternal, error) {
	err := input.validate()
	if err != nil {
		return nil, err
	}

	output := &cseClusterSettingsInternal{}
	if org.Org == nil {
		return nil, fmt.Errorf("could not retrieve the Organization, it is nil")
	}
	output.OrganizationName = org.Org.Name

	vdc, err := org.GetVDCById(input.VdcId, true)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the VDC with ID '%s': %s", input.VdcId, err)
	}
	output.VdcName = vdc.Vdc.Name

	vAppTemplate, err := getVAppTemplateById(org.client, input.KubernetesTemplateOvaId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Kubernetes Template OVA with ID '%s': %s", input.KubernetesTemplateOvaId, err)
	}
	output.KubernetesTemplateOvaName = vAppTemplate.VAppTemplate.Name

	tkgVersions, err := getTkgVersionBundleFromVAppTemplate(vAppTemplate.VAppTemplate)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the required information from the Kubernetes Template OVA: %s", err)
	}
	output.TkgVersionBundle = tkgVersions

	catalogName, err := vAppTemplate.GetCatalogName()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Catalog name where the the Kubernetes Template OVA '%s' (%s) is hosted: %s", input.KubernetesTemplateOvaId, vAppTemplate.VAppTemplate.Name, err)
	}
	output.CatalogName = catalogName

	network, err := vdc.GetOrgVdcNetworkById(input.NetworkId, true)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Org VDC Network with ID '%s': %s", input.NetworkId, err)
	}
	output.NetworkName = network.OrgVDCNetwork.Name

	cseComponentsVersions, err := getCseComponentsVersions(input.CseVersion)
	if err != nil {
		return nil, err
	}
	rdeType, err := getRdeType(org.client, "vmware", "capvcdCluster", cseComponentsVersions.CapvcdRdeTypeVersion)
	if err != nil {
		return nil, err
	}
	output.RdeType = rdeType.DefinedEntityType

	// The input to create a cluster uses different entities IDs, but CSE cluster creation process uses Names.
	// For that reason, we need to transform IDs to Names by querying VCD. This process is optimized with a tiny cache map.
	idToNameCache := map[string]string{
		"": "", // Default empty value to map optional values that were not set, to avoid extra checks. For example, an empty vGPU Policy.
	}

	// Gather all the IDs of the Compute Policies and Storage Profiles, so we can transform them to Names in bulk.
	var computePolicyIds []string
	var storageProfileIds []string
	for _, w := range input.WorkerPools {
		computePolicyIds = append(computePolicyIds, w.SizingPolicyId, w.PlacementPolicyId, w.VGpuPolicyId)
		storageProfileIds = append(storageProfileIds, w.StorageProfileId)
	}
	computePolicyIds = append(computePolicyIds, input.ControlPlane.SizingPolicyId, input.ControlPlane.PlacementPolicyId)
	storageProfileIds = append(storageProfileIds, input.ControlPlane.StorageProfileId, input.DefaultStorageClass.StorageProfileId)

	// Retrieve the Compute Policies and Storage Profiles names and put them in the cache. The cache
	// reduces the calls to VCD. The URN format used by VCD guarantees that IDs are unique, so there is no possibility of clashes here.
	for _, id := range storageProfileIds {
		if _, alreadyPresent := idToNameCache[id]; !alreadyPresent {
			storageProfile, err := getStorageProfileById(org.client, id)
			if err != nil {
				return nil, fmt.Errorf("could not retrieve Storage Profile with ID '%s': %s", id, err)
			}
			idToNameCache[id] = storageProfile.Name
		}
	}
	for _, id := range computePolicyIds {
		if _, alreadyPresent := idToNameCache[id]; !alreadyPresent {
			computePolicy, err := getVdcComputePolicyV2ById(org.client, id)
			if err != nil {
				return nil, fmt.Errorf("could not retrieve Compute Policy with ID '%s': %s", id, err)
			}
			idToNameCache[id] = computePolicy.VdcComputePolicyV2.Name
		}
	}

	// Now that everything is cached in memory, we can build the Node pools and Storage Class payloads in a trivial way.
	output.WorkerPools = make([]cseWorkerPoolSettingsInternal, len(input.WorkerPools))
	for i, w := range input.WorkerPools {
		output.WorkerPools[i] = cseWorkerPoolSettingsInternal{
			Name:         w.Name,
			MachineCount: w.MachineCount,
			DiskSizeGi:   w.DiskSizeGi,
		}
		output.WorkerPools[i].SizingPolicyName = idToNameCache[w.SizingPolicyId]
		output.WorkerPools[i].PlacementPolicyName = idToNameCache[w.PlacementPolicyId]
		output.WorkerPools[i].VGpuPolicyName = idToNameCache[w.VGpuPolicyId]
		output.WorkerPools[i].StorageProfileName = idToNameCache[w.StorageProfileId]
	}
	output.ControlPlane = cseControlPlaneSettingsInternal{
		MachineCount:        input.ControlPlane.MachineCount,
		DiskSizeGi:          input.ControlPlane.DiskSizeGi,
		SizingPolicyName:    idToNameCache[input.ControlPlane.SizingPolicyId],
		PlacementPolicyName: idToNameCache[input.ControlPlane.PlacementPolicyId],
		StorageProfileName:  idToNameCache[input.ControlPlane.StorageProfileId],
		Ip:                  input.ControlPlane.Ip,
	}

	if input.DefaultStorageClass != nil {
		output.DefaultStorageClass = cseDefaultStorageClassInternal{
			StorageProfileName: idToNameCache[input.DefaultStorageClass.StorageProfileId],
			Name:               input.DefaultStorageClass.Name,
			Filesystem:         input.DefaultStorageClass.Filesystem,
		}
		output.DefaultStorageClass.UseDeleteReclaimPolicy = false
		if input.DefaultStorageClass.ReclaimPolicy == "delete" {
			output.DefaultStorageClass.UseDeleteReclaimPolicy = true
		}
	}

	vcdKeConfig, err := getVcdKeConfig(org.client, cseComponentsVersions.VcdKeConfigRdeTypeVersion, input.NodeHealthCheck)
	if err != nil {
		return nil, err
	}
	if vcdKeConfig != nil {
		output.VcdKeConfig = *vcdKeConfig
	}

	output.Owner = input.Owner
	if input.Owner == "" {
		sessionInfo, err := org.client.GetSessionInfo()
		if err != nil {
			return nil, fmt.Errorf("error getting the Owner: %s", err)
		}
		output.Owner = sessionInfo.User.Name
	}

	output.VcdUrl = strings.Replace(org.client.VCDHREF.String(), "/api", "", 1)

	// These don't change, don't need mapping
	output.ApiToken = input.ApiToken
	output.AutoRepairOnErrors = input.AutoRepairOnErrors
	output.CseVersion = input.CseVersion
	output.Name = input.Name
	output.PodCidr = input.PodCidr
	output.ServiceCidr = input.ServiceCidr
	output.SshPublicKey = input.SshPublicKey
	output.VirtualIpSubnet = input.VirtualIpSubnet

	return output, nil
}

// getTkgVersionBundleFromVAppTemplate returns a tkgVersionBundle with the details of
// all the Kubernetes cluster components versions given a valid Kubernetes Template OVA.
// If it is not a valid Kubernetes Template OVA, returns an error.
func getTkgVersionBundleFromVAppTemplate(template *types.VAppTemplate) (tkgVersionBundle, error) {
	result := tkgVersionBundle{}
	if template == nil {
		return result, fmt.Errorf("the Kubernetes Template OVA is nil")
	}
	if template.Children == nil || len(template.Children.VM) == 0 {
		return result, fmt.Errorf("the Kubernetes Template OVA '%s' doesn't have any child VM", template.Name)
	}
	if template.Children.VM[0].ProductSection == nil {
		return result, fmt.Errorf("the Product section of the Kubernetes Template OVA '%s' is empty, can't proceed", template.Name)
	}
	id := ""
	for _, prop := range template.Children.VM[0].ProductSection.Property {
		if prop != nil && prop.Key == "VERSION" {
			id = prop.DefaultValue // Use DefaultValue and not Value as the value we want is in the "value" attr
		}
	}
	if id == "" {
		return result, fmt.Errorf("could not find any VERSION property inside the Kubernetes Template OVA '%s' Product section", template.Name)
	}

	tkgVersionsMap := "cse/tkg_versions.json"
	cseTkgVersionsJson, err := cseFiles.ReadFile(tkgVersionsMap)
	if err != nil {
		return result, fmt.Errorf("failed reading %s: %s", tkgVersionsMap, err)
	}

	versionsMap := map[string]interface{}{}
	err = json.Unmarshal(cseTkgVersionsJson, &versionsMap)
	if err != nil {
		return result, fmt.Errorf("failed unmarshaling %s: %s", tkgVersionsMap, err)
	}
	versionMap, ok := versionsMap[id]
	if !ok {
		return result, fmt.Errorf("the Kubernetes Template OVA '%s' is not supported", template.Name)
	}

	// We don't need to check the Split result because the map checking above guarantees that the ID is well-formed.
	idParts := strings.Split(id, "-")
	result.KubernetesVersion = idParts[0]
	result.TkrVersion = strings.ReplaceAll(idParts[0], "+", "---") + "-" + idParts[1]
	result.TkgVersion = versionMap.(map[string]interface{})["tkg"].(string)
	result.EtcdVersion = versionMap.(map[string]interface{})["etcd"].(string)
	result.CoreDnsVersion = versionMap.(map[string]interface{})["coreDns"].(string)
	return result, nil
}

func (tkgVersions tkgVersionBundle) compareTkgVersion(tkgVersion string) int {
	receiverVersion, err := semver.NewVersion(tkgVersions.TkgVersion)
	if err != nil {
		return -2
	}
	inputVersion, err := semver.NewVersion(tkgVersion)
	if err != nil {
		return -2
	}
	return receiverVersion.Compare(inputVersion)
}

func (tkgVersions tkgVersionBundle) kubernetesVersionIsOneMinorHigher(kubernetesVersion string) bool {
	receiverVersionTokens := strings.Split(tkgVersions.KubernetesVersion, ".")
	if len(receiverVersionTokens) < 3 {
		return false
	}
	vTokens := strings.Split(kubernetesVersion, ".")
	if len(vTokens) < 3 {
		return false
	}
	minor, err := strconv.Atoi(receiverVersionTokens[1])
	if err != nil {
		return false
	}

	return receiverVersionTokens[0] == vTokens[0] && fmt.Sprintf("%d", minor+1) == vTokens[1] && receiverVersionTokens[2] == vTokens[2]
}

// getVcdKeConfig gets the required information from the CSE Server configuration RDE (VCDKEConfig), such as the
// Machine Health Check settings and the Container Registry URL.
func getVcdKeConfig(client *Client, vcdKeConfigVersion string, isNodeHealthCheckActive bool) (*vcdKeConfig, error) {
	rdes, err := getRdesByName(client, "vmware", "VCDKEConfig", vcdKeConfigVersion, "vcdKeConfig")
	if err != nil {
		return nil, err
	}
	if len(rdes) != 1 {
		return nil, fmt.Errorf("expected exactly one VCDKEConfig RDE with version '%s', but got %d", vcdKeConfigVersion, len(rdes))
	}

	profiles, ok := rdes[0].DefinedEntity.Entity["profiles"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("wrong format of VCDKEConfig RDE contents, expected a 'profiles' array")
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("wrong format of VCDKEConfig RDE contents, expected a non-empty 'profiles' element")
	}

	result := &vcdKeConfig{}
	// We append /tkg as required, even in air-gapped environments:
	// https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2/VMware-Cloud-Director-Container-Service-Extension-Install-provider-4.2/GUID-B5C19221-2ECA-4DCD-8EA1-8E391F6217C1.html
	result.ContainerRegistryUrl = fmt.Sprintf("%s/tkg", profiles[0].(map[string]interface{})["containerRegistryUrl"])

	if isNodeHealthCheckActive {
		mhc, ok := profiles[0].(map[string]interface{})["K8Config"].(map[string]interface{})["mhc"]
		if !ok {
			// If there is no "mhc" entry in the VCDKEConfig JSON, we skip setting this part of the Kubernetes cluster configuration
			return result, nil
		}
		result.MaxUnhealthyNodesPercentage = mhc.(map[string]interface{})["maxUnhealthyNodes"].(float64)
		result.NodeStartupTimeout = mhc.(map[string]interface{})["nodeStartupTimeout"].(string)
		result.NodeNotReadyTimeout = mhc.(map[string]interface{})["nodeUnknownTimeout"].(string)
		result.NodeUnknownTimeout = mhc.(map[string]interface{})["nodeNotReadyTimeout"].(string)
	}

	return result, nil
}

// getCseTemplate reads the Go template present in the embedded cseFiles filesystem.
func getCseTemplate(cseVersion semver.Version, templateName string) (string, error) {
	minimumVersion, err := semver.NewVersion("4.1")
	if err != nil {
		return "", err
	}
	if cseVersion.LessThan(minimumVersion) {
		return "", fmt.Errorf("the Container Service minimum version is '%s'", minimumVersion.String())
	}
	versionSegments := cseVersion.Segments()
	// We try with major.minor.patch
	fullTemplatePath := fmt.Sprintf("cse/%d.%d.%d/%s.tmpl", versionSegments[0], versionSegments[1], versionSegments[2], templateName)
	result, err := cseFiles.ReadFile(fullTemplatePath)
	if err != nil {
		// We try now just with major.minor
		fullTemplatePath = fmt.Sprintf("cse/%d.%d/%s.tmpl", versionSegments[0], versionSegments[1], templateName)
		result, err = cseFiles.ReadFile(fullTemplatePath)
		if err != nil {
			return "", fmt.Errorf("could not read Go template '%s.tmpl' for CSE version %s", templateName, cseVersion.String())
		}
	}
	return string(result), nil
}
