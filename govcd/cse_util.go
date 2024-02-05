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
	"strconv"
	"strings"
	"time"
)

// getCseComponentsVersions gets the CSE components versions from its version.
// TODO: Is this really necessary? What happens in UI if I have a 1.1.0-1.2.0-1.0.0 (4.2) cluster and then CSE is updated to 4.3?
func getCseComponentsVersions(cseVersion semver.Version) (*cseComponentsVersions, error) {
	v42, _ := semver.NewVersion("4.1")

	if cseVersion.Equal(v42) {
		return &cseComponentsVersions{
			VcdKeConfigRdeTypeVersion: "1.1.0",
			CapvcdRdeTypeVersion:      "1.2.0",
			CseInterfaceVersion:       "1.0.0",
		}, nil
	}
	return nil, fmt.Errorf("not supported version %s", cseVersion.String())
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
		State:                      capvcd.Status.VcdKe.State,
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
	// FIXME: This is a workaround, because for some reason the ID contains the VDC name instead of the VDC ID.
	//        Once this is fixed, this conditional should not be needed anymore.
	result.VdcId = result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].Id
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

	yamlDocuments, err := unmarshalMultipleYamlDocuments(result.capvcdType.Spec.CapiYaml)
	if err != nil {
		return nil, err
	}

	// We need a map of worker pools and not a slice, because there are two types of YAML documents
	// that contain data about a specific worker pool (VCDMachineTemplate and MachineDeployment), and we can get them in no
	// particular order, so we store the worker pools with their name as key. This way we can easily fetch them and override them.
	workerPools := map[string]CseWorkerPoolSettings{}
	for _, yamlDocument := range yamlDocuments {
		switch yamlDocument["kind"] {
		case "KubeadmControlPlane":
			replicas, err := traverseMapAndGet[float64](yamlDocument, "spec.replicas")
			if err != nil {
				return nil, err
			}
			result.ControlPlane.MachineCount = int(replicas)

			users, err := traverseMapAndGet[[]any](yamlDocument, "spec.kubeadmConfigSpec.users")
			if err != nil {
				return nil, err
			}
			if len(users) == 0 {
				return nil, fmt.Errorf("expected 'spec.kubeadmConfigSpec.users' slice to not to be empty")
			}
			keys, err := traverseMapAndGet[[]string](users[0], "sshAuthorizedKeys")
			if err != nil && !strings.Contains(err.Error(), "key 'sshAuthorizedKeys' does not exist in input map") {
				return nil, err
			}
			if len(keys) > 0 {
				result.SshPublicKey = keys[0] // Optional field
			}
		case "VCDMachineTemplate":
			name, err := traverseMapAndGet[string](yamlDocument, "metadata.name")
			if err != nil {
				return nil, err
			}
			sizingPolicyName, err := traverseMapAndGet[string](yamlDocument, "spec.template.spec.sizingPolicy")
			if err != nil && !strings.Contains(err.Error(), "key 'sizingPolicy' does not exist in input map") {
				return nil, err
			}
			placementPolicyName, err := traverseMapAndGet[string](yamlDocument, "spec.template.spec.placementPolicy")
			if err != nil && !strings.Contains(err.Error(), "key 'placementPolicy' does not exist in input map") {
				return nil, err
			}
			storageProfileName, err := traverseMapAndGet[string](yamlDocument, "spec.template.spec.storageProfile")
			if err != nil && !strings.Contains(err.Error(), "key 'storageProfile' does not exist in input map") {
				return nil, err
			}
			diskSizeGiRaw, err := traverseMapAndGet[string](yamlDocument, "spec.template.spec.diskSize")
			if err != nil {
				return nil, err
			}
			diskSizeGi, err := strconv.Atoi(strings.ReplaceAll(diskSizeGiRaw, "Gi", ""))
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

				// We do it just once for the Control Plane because all VCDMachineTemplate blocks share the same OVA
				ovaName, err := traverseMapAndGet[string](yamlDocument, "spec.template.spec.template")
				if err != nil {
					return nil, err
				}
				// TODO: There is no method to retrieve vApp templates by name....
				result.KubernetesTemplateOvaId = ovaName
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
			name, err := traverseMapAndGet[string](yamlDocument, "metadata.name")
			if err != nil {
				return nil, err
			}
			// This is one Worker Pool. We need to check the map of worker pools, just in case we already saved the
			// other information from VCDMachineTemplate.
			if _, ok := workerPools[name]; !ok {
				workerPools[name] = CseWorkerPoolSettings{}
			}
			workerPool := workerPools[name]
			replicas, err := traverseMapAndGet[float64](yamlDocument, "spec.replicas")
			if err != nil {
				return nil, err
			}
			workerPool.MachineCount = int(replicas)
			workerPools[name] = workerPool // Override the worker pool with the updated data
		case "VCDCluster":
			subnet, err := traverseMapAndGet[string](yamlDocument, "spec.loadBalancerConfigSpec.vipSubnet")
			if err == nil {
				result.VirtualIpSubnet = subnet // This is optional
			}
		case "Cluster":
			cidrBlocks, err := traverseMapAndGet[[]any](yamlDocument, "spec.clusterNetwork.pods.cidrBlocks")
			if err != nil {
				return nil, err
			}
			if len(cidrBlocks) == 0 {
				return nil, fmt.Errorf("expected at least one 'spec.clusterNetwork.pods.cidrBlocks' item")
			}
			result.PodCidr = cidrBlocks[0].(string)

			cidrBlocks, err = traverseMapAndGet[[]any](yamlDocument, "spec.clusterNetwork.services.cidrBlocks")
			if err != nil {
				return nil, err
			}
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

	return result, nil
}

// waitUntilClusterIsProvisioned waits for the Kubernetes cluster to be in "provisioned" state, either indefinitely (if timeoutMinutes = 0)
// or until the timeout is reached. If the cluster is in "provisioned" state before the given timeout, it returns a CseKubernetesCluster object
// representing the Kubernetes cluster with all its latest details.
// If one of the states of the cluster at a given point is "error", this function also checks whether the cluster has the "Auto Repair on Errors" flag enabled,
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

		// Here we don't to use cseConvertToCseKubernetesClusterType to avoid calling VCD
		entityBytes, err := json.Marshal(rde.DefinedEntity.Entity)
		if err != nil {
			return fmt.Errorf("could not marshal the RDE contents to create a capvcdType instance: %s", err)
		}

		err = json.Unmarshal(entityBytes, &capvcd)
		if err != nil {
			return fmt.Errorf("could not unmarshal the RDE contents to create a Capvcd instance: %s", err)
		}

		switch capvcd.Status.VcdKe.State {
		case "provisioned":
			return nil
		case "error":
			// We just finish if auto-recovery is disabled, otherwise we just let CSE fixing things in background
			if !capvcd.Spec.VcdKe.AutoRepairOnErrors {
				// Give feedback about what went wrong
				errors := ""
				// TODO: Change to ErrorSet
				for _, event := range capvcd.Status.Capvcd.EventSet {
					errors += fmt.Sprintf("%s,\n", event.AdditionalDetails)
				}
				return fmt.Errorf("got an error and 'AutoRepairOnErrors' is disabled, aborting. Errors:\n%s", errors)
			}
		}

		util.Logger.Printf("[DEBUG] Cluster '%s' is in '%s' state, will check again in %d seconds", rde.DefinedEntity.ID, capvcd.Status.VcdKe.State, sleepTime)
		elapsed = time.Since(start)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return fmt.Errorf("timeout of %d minutes reached, latest cluster state obtained was '%s'", timeoutMinutes, capvcd.Status.VcdKe.State)
}

// validate validates the CSE Kubernetes cluster creation input data. Returns an error if some of the fields is wrong.
func (input *CseClusterSettings) validate() error {
	cseNamesRegex, err := regexp.Compile(`^[a-z](?:[a-z0-9-]{0,29}[a-z0-9])?$`)
	if err != nil {
		return fmt.Errorf("could not compile regular expression '%s'", err)
	}

	if !cseNamesRegex.MatchString(input.Name) {
		return fmt.Errorf("the cluster name is required and must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters, but it was: '%s'", input.Name)
	}

	if input.OrganizationId == "" {
		return fmt.Errorf("the Organization ID is required")
	}
	if input.VdcId == "" {
		return fmt.Errorf("the VDC ID is required")
	}
	if input.KubernetesTemplateOvaId == "" {
		return fmt.Errorf("the Kubernetes template OVA ID is required")
	}
	if input.NetworkId == "" {
		return fmt.Errorf("the Network ID is required")
	}
	_, err = getCseComponentsVersions(input.CseVersion)
	if err != nil {
		return fmt.Errorf("the CSE version '%s' is not supported", input.CseVersion.String())
	}
	if input.ControlPlane.MachineCount < 1 || input.ControlPlane.MachineCount%2 == 0 {
		return fmt.Errorf("number of control plane nodes must be odd and higher than 0, but it was '%d'", input.ControlPlane.MachineCount)
	}
	if input.ControlPlane.DiskSizeGi < 20 {
		return fmt.Errorf("disk size for the Control Plane in Gibibytes (Gi) must be at least 20, but it was '%d'", input.ControlPlane.DiskSizeGi)
	}
	if len(input.WorkerPools) == 0 {
		return fmt.Errorf("there must be at least one Worker pool")
	}
	for _, workerPool := range input.WorkerPools {
		if !cseNamesRegex.MatchString(workerPool.Name) {
			return fmt.Errorf("the Worker pool name is required and must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters, but it was: '%s'", workerPool.Name)
		}
		if workerPool.DiskSizeGi < 20 {
			return fmt.Errorf("disk size for the Worker pool '%s' in Gibibytes (Gi) must be at least 20, but it was '%d'", workerPool.Name, workerPool.DiskSizeGi)
		}
		if workerPool.MachineCount < 1 {
			return fmt.Errorf("number of Worker pool '%s' nodes must higher than 0, but it was '%d'", workerPool.Name, workerPool.MachineCount)
		}
	}
	if input.DefaultStorageClass != nil {
		if !cseNamesRegex.MatchString(input.DefaultStorageClass.Name) {
			return fmt.Errorf("the Default Storage Class name is required and must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters, but it was: '%s'", input.DefaultStorageClass.Name)
		}
		if input.DefaultStorageClass.StorageProfileId == "" {
			return fmt.Errorf("the Storage Profile ID for the Default Storage Class is required")
		}
		if input.DefaultStorageClass.ReclaimPolicy != "delete" && input.DefaultStorageClass.ReclaimPolicy != "retain" {
			return fmt.Errorf("the reclaim policy for the Default Storage Class must be either 'delete' or 'retain', but it was '%s'", input.DefaultStorageClass.ReclaimPolicy)
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

// cseClusterSettingsToInternal transforms user input data (CseClusterSettings) into the final payload that
// will be used to render the Go templates that define a Kubernetes cluster creation payload (cseClusterSettingsInternal).
func cseClusterSettingsToInternal(input CseClusterSettings, org Org) (cseClusterSettingsInternal, error) {
	output := cseClusterSettingsInternal{}
	err := input.validate()
	if err != nil {
		return output, err
	}

	if org.Org == nil {
		return output, fmt.Errorf("the Organization is nil")
	}

	output.OrganizationName = org.Org.Name

	vdc, err := org.GetVDCById(input.VdcId, true)
	if err != nil {
		return output, fmt.Errorf("could not retrieve the VDC with ID '%s': %s", input.VdcId, err)
	}
	output.VdcName = vdc.Vdc.Name

	vAppTemplate, err := getVAppTemplateById(org.client, input.KubernetesTemplateOvaId)
	if err != nil {
		return output, fmt.Errorf("could not retrieve the Kubernetes Template OVA with ID '%s': %s", input.KubernetesTemplateOvaId, err)
	}
	output.KubernetesTemplateOvaName = vAppTemplate.VAppTemplate.Name

	tkgVersions, err := getTkgVersionBundleFromVAppTemplateName(vAppTemplate.VAppTemplate.Name)
	if err != nil {
		return output, fmt.Errorf("could not retrieve the required information from the Kubernetes Template OVA: %s", err)
	}
	output.TkgVersionBundle = tkgVersions

	catalogName, err := vAppTemplate.GetCatalogName()
	if err != nil {
		return output, fmt.Errorf("could not retrieve the Catalog name where the the Kubernetes Template OVA '%s' is hosted: %s", input.KubernetesTemplateOvaId, err)
	}
	output.CatalogName = catalogName

	network, err := vdc.GetOrgVdcNetworkById(input.NetworkId, true)
	if err != nil {
		return output, fmt.Errorf("could not retrieve the Org VDC Network with ID '%s': %s", input.NetworkId, err)
	}
	output.NetworkName = network.OrgVDCNetwork.Name

	currentCseVersion, err := getCseComponentsVersions(input.CseVersion)
	if err != nil {
		return output, fmt.Errorf("the CSE version '%s' is not supported: %s", input.CseVersion.String(), err)
	}
	rdeType, err := getRdeType(org.client, "vmware", "capvcdCluster", currentCseVersion.CapvcdRdeTypeVersion)
	if err != nil {
		return output, err
	}
	output.RdeType = rdeType.DefinedEntityType

	// The input to create a cluster uses different entities IDs, but CSE cluster creation process uses names.
	// For that reason, we need to transform IDs to Names by querying VCD. This process is optimized with a tiny "nameToIdCache" map.
	idToNameCache := map[string]string{
		"": "", // Default empty value to map optional values that were not set
	}
	var computePolicyIds []string
	var storageProfileIds []string
	for _, w := range input.WorkerPools {
		computePolicyIds = append(computePolicyIds, w.SizingPolicyId, w.PlacementPolicyId, w.VGpuPolicyId)
		storageProfileIds = append(storageProfileIds, w.StorageProfileId)
	}
	computePolicyIds = append(computePolicyIds, input.ControlPlane.SizingPolicyId, input.ControlPlane.PlacementPolicyId)
	storageProfileIds = append(storageProfileIds, input.ControlPlane.StorageProfileId, input.DefaultStorageClass.StorageProfileId)

	for _, id := range storageProfileIds {
		if _, alreadyPresent := idToNameCache[id]; !alreadyPresent {
			storageProfile, err := getStorageProfileById(org.client, id)
			if err != nil {
				return output, fmt.Errorf("could not retrieve Storage Profile with ID '%s': %s", id, err)
			}
			idToNameCache[id] = storageProfile.Name
		}
	}
	for _, id := range computePolicyIds {
		if _, alreadyPresent := idToNameCache[id]; !alreadyPresent {
			computePolicy, err := getVdcComputePolicyV2ById(org.client, id)
			if err != nil {
				return output, fmt.Errorf("could not retrieve Compute Policy with ID '%s': %s", id, err)
			}
			idToNameCache[id] = computePolicy.VdcComputePolicyV2.Name
		}
	}

	// Now that everything is cached in memory, we can build the Node pools and Storage Class payloads
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

	cseVersions, err := getCseComponentsVersions(input.CseVersion)
	if err != nil {
		return output, err
	}

	vcdKeConfig, err := getVcdKeConfig(org.client, cseVersions.VcdKeConfigRdeTypeVersion, input.NodeHealthCheck)
	if err != nil {
		return output, err
	}
	output.VcdKeConfig = vcdKeConfig

	output.Owner = input.Owner
	if input.Owner == "" {
		sessionInfo, err := org.client.GetSessionInfo()
		if err != nil {
			return output, fmt.Errorf("error getting the owner of the cluster: %s", err)
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

// tkgVersionBundle is a type that contains all the versions of the components of
// a Kubernetes cluster that can be obtained with the vApp Template name, downloaded
// from VMware Customer connect:
// https://customerconnect.vmware.com/downloads/details?downloadGroup=TKG-240&productId=1400
type tkgVersionBundle struct {
	EtcdVersion       string
	CoreDnsVersion    string
	TkgVersion        string
	TkrVersion        string
	KubernetesVersion string
}

// getTkgVersionBundleFromVAppTemplateName returns a tkgVersionBundle with the details of
// all the Kubernetes cluster components versions given a valid vApp Template name, that should
// correspond to a Kubernetes template. If it is not a valid vApp Template, returns an error.
func getTkgVersionBundleFromVAppTemplateName(ovaName string) (tkgVersionBundle, error) {
	result := tkgVersionBundle{}

	if strings.Contains(ovaName, "photon") {
		return result, fmt.Errorf("the OVA '%s' uses Photon, and it is not supported", ovaName)
	}

	cutPosition := strings.LastIndex(ovaName, "kube-")
	if cutPosition < 0 {
		return result, fmt.Errorf("the OVA '%s' is not a Kubernetes template OVA", ovaName)
	}
	parsedOvaName := strings.ReplaceAll(ovaName, ".ova", "")[cutPosition+len("kube-"):]

	cseTkgVersionsJson, err := cseFiles.ReadFile("cse/tkg_versions.json")
	if err != nil {
		return result, err
	}

	versionsMap := map[string]any{}
	err = json.Unmarshal(cseTkgVersionsJson, &versionsMap)
	if err != nil {
		return result, fmt.Errorf("failed unmarshaling cse/tkg_versions.json: %s", err)
	}
	versionMap, ok := versionsMap[parsedOvaName]
	if !ok {
		return result, fmt.Errorf("the Kubernetes OVA '%s' is not supported", parsedOvaName)
	}

	ovaParts := strings.Split(parsedOvaName, "-")
	if len(ovaParts) < 2 {
		return result, fmt.Errorf("unexpected error parsing the OVA name '%s', it doesn't follow the original naming convention", parsedOvaName)
	}

	result.KubernetesVersion = ovaParts[0]
	result.TkrVersion = strings.ReplaceAll(ovaParts[0], "+", "---") + "-" + ovaParts[1]
	result.TkgVersion = versionMap.(map[string]any)["tkg"].(string)
	result.EtcdVersion = versionMap.(map[string]any)["etcd"].(string)
	result.CoreDnsVersion = versionMap.(map[string]any)["coreDns"].(string)
	return result, nil
}

// getVcdKeConfig gets the required information from the CSE Server configuration RDE
func getVcdKeConfig(client *Client, vcdKeConfigVersion string, isNodeHealthCheckActive bool) (vcdKeConfig, error) {
	result := vcdKeConfig{}
	rdes, err := getRdesByName(client, "vmware", "VCDKEConfig", vcdKeConfigVersion, "vcdKeConfig")
	if err != nil {
		return result, fmt.Errorf("could not retrieve VCDKEConfig RDE with version %s: %s", vcdKeConfigVersion, err)
	}
	if len(rdes) != 1 {
		return result, fmt.Errorf("expected exactly one VCDKEConfig RDE but got %d", len(rdes))
	}

	profiles, ok := rdes[0].DefinedEntity.Entity["profiles"].([]any)
	if !ok {
		return result, fmt.Errorf("wrong format of VCDKEConfig, expected a 'profiles' array")
	}
	if len(profiles) != 1 {
		return result, fmt.Errorf("wrong format of VCDKEConfig, expected a single 'profiles' element, got %d", len(profiles))
	}
	// TODO: Check airgapped environments: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1.1a/VMware-Cloud-Director-Container-Service-Extension-Install-provider-4.1.1/GUID-F00BE796-B5F2-48F2-A012-546E2E694400.html
	result.ContainerRegistryUrl = fmt.Sprintf("%s/tkg", profiles[0].(map[string]any)["containerRegistryUrl"])

	if isNodeHealthCheckActive {
		mhc, ok := profiles[0].(map[string]any)["K8Config"].(map[string]any)["mhc"].(map[string]any)
		if !ok {
			return result, nil
		}
		result.MaxUnhealthyNodesPercentage = mhc["maxUnhealthyNodes"].(float64)
		result.NodeStartupTimeout = mhc["nodeStartupTimeout"].(string)
		result.NodeNotReadyTimeout = mhc["nodeUnknownTimeout"].(string)
		result.NodeUnknownTimeout = mhc["nodeNotReadyTimeout"].(string)
	}

	return result, nil
}

func getCseTemplate(cseVersion semver.Version, templateName string) (string, error) {
	cseVersionSegments := cseVersion.Segments()
	result, err := cseFiles.ReadFile(fmt.Sprintf("cse/%d.%d/%s.tmpl", cseVersionSegments[0], cseVersionSegments[1], templateName))
	if err != nil {
		return "", err
	}
	return string(result), nil
}
