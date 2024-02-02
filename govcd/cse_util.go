package govcd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// cseConvertToCseKubernetesClusterType takes a generic RDE that must represent an existing CSE Kubernetes cluster,
// and transforms it to an equivalent CseKubernetesCluster object that represents the same cluster, but
// it is easy to explore and consume. If the input RDE is not a CSE Kubernetes cluster, this method
// will obviously return an error.
// The nameToIdCache maps names with their IDs. This is used to reduce calls to VCD to retrieve this information.
func cseConvertToCseKubernetesClusterType(rde *DefinedEntity, nameToIdCache map[string]string) (*CseKubernetesCluster, error) {
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
		},
		ID:                         rde.DefinedEntity.ID,
		Etag:                       rde.Etag,
		KubernetesVersion:          capvcd.Status.Capvcd.Upgrade.Current.KubernetesVersion,
		TkgVersion:                 capvcd.Status.Capvcd.Upgrade.Current.TkgVersion,
		CapvcdVersion:              capvcd.Status.Capvcd.CapvcdVersion,
		ClusterResourceSetBindings: make([]string, len(capvcd.Status.Capvcd.ClusterResourceSetBindings)),
		CpiVersion:                 capvcd.Status.Cpi.Version,
		CsiVersion:                 capvcd.Status.Csi.Version,
		State:                      capvcd.Status.VcdKe.State,
		client:                     rde.client,
		capvcdType:                 capvcd,
		nameToIdCache:              nameToIdCache,
	}
	if result.nameToIdCache == nil {
		result.nameToIdCache = map[string]string{}
	}

	for i, binding := range capvcd.Status.Capvcd.ClusterResourceSetBindings {
		result.ClusterResourceSetBindings[i] = binding.ClusterResourceSetName
	}

	if len(result.capvcdType.Status.Capvcd.VcdProperties.Organizations) == 0 {
		return nil, fmt.Errorf("could not read Organizations from Capvcd type")
	}
	result.OrganizationId = result.capvcdType.Status.Capvcd.VcdProperties.Organizations[0].Id
	if len(result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs) == 0 {
		return nil, fmt.Errorf("could not read VDCs from Capvcd type")
	}
	result.VdcId = result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].Id

	// To retrieve the Network ID, we check that it is not already cached. If it's not, we retrieve it with
	// the VDC ID and name filters
	if _, ok := result.nameToIdCache[result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].OvdcNetworkName]; !ok {
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
		result.nameToIdCache[result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].OvdcNetworkName] = networks[0].OpenApiOrgVdcNetwork.ID
	}
	result.NetworkId = result.nameToIdCache[result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].OvdcNetworkName]

	if rde.DefinedEntity.Owner == nil {
		return nil, fmt.Errorf("could not read Owner from RDE")
	}
	result.Owner = rde.DefinedEntity.Owner.Name

	if result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.K8SStorageClassName != "" { // This would mean there is a Default Storage Class defined
		result.DefaultStorageClass = &CseDefaultStorageClassSettings{
			Name:          result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.K8SStorageClassName,
			ReclaimPolicy: result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName,
			Filesystem:    result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName,
		}

		// To retrieve the Storage Profile ID, we check that it is not already cached. If it's not, we retrieve it
		if _, ok := result.nameToIdCache[result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName]; !ok {
			// TODO: There is no method to retrieve Storage profiles by name....
			result.nameToIdCache[result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName] = ""
		}
		result.DefaultStorageClass.StorageProfileId = result.nameToIdCache[result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName]
	}

	yamlDocuments, err := unmarshalMultipleYamlDocuments(result.capvcdType.Spec.CapiYaml)
	if err != nil {
		return nil, err
	}

	result.CseVersion = "" // TODO: Get opposite from supportedVersionsMap

	//var workerPools []CseWorkerPoolSettings
	for _, yamlDocument := range yamlDocuments {
		switch yamlDocument["kind"] {
		case "KubeadmControlPlane":
		// result.ControlPlane.MachineCount
		// result.SshPublicKey
		case "VCDMachineTemplate":
			// Obtain all name->ID
			if strings.Contains("name", "control-plane-node-pool") {
				// TODO: There is no method to retrieve vApp templates by name....
				// result.KubernetesTemplateOvaId
				// TODO: There is no method to retrieve vApp templates by name....
				// getAllVdcComputePoliciesV2()
				// result.ControlPlane.SizingPolicyId
				// result.ControlPlane.PlacementPolicyId
				// result.ControlPlane.StorageProfileId
				// result.ControlPlane.DiskSizeGi
				fmt.Print("b")
			} else {
				fmt.Print("a")
				//workerPool := CseWorkerPoolSettings{}
				//workerPools = append(workerPools, workerPool)
			}
		case "VCDCluster":
		// result.ControlPlane.Ip
		// result.VirtualIpSubnet
		case "Cluster":
		// result.PodCidr
		// result.ServicesCidr
		case "MachineHealthCheck":
			// This is quite simple, if we find this document, means that Machine Health Check is enabled
			result.NodeHealthCheck = true
		}
	}

	// // TODO: This needs a refactoring
	//		if nodePool.PlacementPolicy != "" {
	//			policies, err := vcdClient.GetAllVdcComputePoliciesV2(url.Values{
	//				"filter": []string{fmt.Sprintf("name==%s", nodePool.PlacementPolicy)},
	//			})
	//			if err != nil {
	//				return nil, err // TODO
	//			}
	//			nameToIds[nodePool.PlacementPolicy] = policies[0].VdcComputePolicyV2.ID
	//		}
	//		if nodePool.SizingPolicy != "" {
	//			policies, err := vcdClient.GetAllVdcComputePoliciesV2(url.Values{
	//				"filter": []string{fmt.Sprintf("name==%s", nodePool.SizingPolicy)},
	//			})
	//			if err != nil {
	//				return nil, err // TODO
	//			}
	//			nameToIds[nodePool.SizingPolicy] = policies[0].VdcComputePolicyV2.ID
	//		}
	//		if nodePool.StorageProfile != "" {
	//			ref, err := vdc.FindStorageProfileReference(nodePool.StorageProfile)
	//			if err != nil {
	//				return nil, fmt.Errorf("could not get Default Storage Class options from 'spec.vcdKe.defaultStorageClassOptions': %s", err) // TODO
	//			}
	//			nameToIds[nodePool.StorageProfile] = ref.ID
	//		}
	//		block["sizing_policy_id"] = nameToIds[nodePool.SizingPolicy]
	//		if nodePool.NvidiaGpuEnabled { // TODO: Be sure this is a worker node pool and not control plane (doesnt have this attr)
	//			block["vgpu_policy_id"] = nameToIds[nodePool.PlacementPolicy] // It's a placement policy here
	//		} else {
	//			block["placement_policy_id"] = nameToIds[nodePool.PlacementPolicy]
	//		}
	//		block["storage_profile_id"] = nameToIds[nodePool.StorageProfile]
	//		block["disk_size_gi"] = nodePool.DiskSizeMb / 1024
	//
	//		if strings.HasSuffix(nodePool.Name, "-control-plane-node-pool") {
	//			// Control Plane
	//			if len(cluster.Capvcd.Status.Capvcd.ClusterApiStatus.ApiEndpoints) == 0 {
	//				return nil, fmt.Errorf("could not retrieve Cluster IP")
	//			}
	//			block["ip"] = cluster.Capvcd.Status.Capvcd.ClusterApiStatus.ApiEndpoints[0].Host
	//			controlPlaneBlocks[0] = block
	//		} else {
	//			// Worker node
	//			block["name"] = nodePool.Name
	//
	//			nodePoolBlocks[i] = block
	//		}

	return result, nil
}

// waitUntilClusterIsProvisioned waits for the Kubernetes cluster to be in "provisioned" state, either indefinitely (if timeoutMinutes = 0)
// or until the timeout is reached. If the cluster is in "provisioned" state before the given timeout, it returns a CseKubernetesCluster object
// representing the Kubernetes cluster with all its latest details.
// If one of the states of the cluster at a given point is "error", this function also checks whether the cluster has the "Auto Repair on Errors" flag enabled,
// so it keeps waiting if it's true.
// If timeout is reached before the cluster is in "provisioned" state, it returns an error.
func waitUntilClusterIsProvisioned(client *Client, clusterId string, timeoutMinutes time.Duration) (*CseKubernetesCluster, error) {
	var elapsed time.Duration
	sleepTime := 10

	start := time.Now()
	cluster := &CseKubernetesCluster{}
	for elapsed <= timeoutMinutes*time.Minute || timeoutMinutes == 0 { // If the user specifies timeoutMinutes=0, we wait forever
		rde, err := getRdeById(client, clusterId)
		if err != nil {
			return nil, err
		}

		cluster, err = cseConvertToCseKubernetesClusterType(rde, cluster.nameToIdCache)
		if err != nil {
			return nil, err
		}

		switch cluster.State {
		case "provisioned":
			return cluster, nil
		case "error":
			// We just finish if auto-recovery is disabled, otherwise we just let CSE fixing things in background
			if !cluster.AutoRepairOnErrors {
				// Give feedback about what went wrong
				errors := ""
				for _, event := range cluster.Events {
					if event.Type == "error" {
						errors += fmt.Sprintf("%s\n", event.Details)
					}
				}
				return cluster, fmt.Errorf("got an error and 'auto repair on errors' is disabled, aborting. Errors:\n%s", errors)
			}
		}

		util.Logger.Printf("[DEBUG] Cluster '%s' is in '%s' state, will check again in %d seconds", cluster.ID, cluster.State, sleepTime)
		elapsed = time.Since(start)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return cluster, fmt.Errorf("timeout of %d minutes reached, latest cluster state obtained was '%s'", timeoutMinutes, cluster.State)
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
	if _, ok := supportedCseVersions[input.CseVersion]; !ok {
		return fmt.Errorf("the CSE version '%s' is not supported. Must be one of %v", input.CseVersion, getKeys(supportedCseVersions))
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

	currentCseVersion, ok := supportedCseVersions[input.CseVersion]
	if !ok {
		return output, fmt.Errorf("the CSE version '%s' is not supported. List of supported versions: %v", input.CseVersion, getKeys(supportedCseVersions))
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

	vcdKeConfig, err := getVcdKeConfig(org.client, supportedCseVersions[input.CseVersion].VcdKeConfigRdeTypeVersion, input.NodeHealthCheck)
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

func getCseTemplate(cseVersion, templateName string) (string, error) {
	result, err := cseFiles.ReadFile(fmt.Sprintf("cse/%s/%s.tmpl", cseVersion, templateName))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// getKeys retrieves all the keys from the given map and returns them as a slice
func getKeys[K comparable, V any](input map[K]V) []K {
	result := make([]K, len(input))
	i := 0
	for k := range input {
		result[i] = k
		i++
	}
	return result
}
