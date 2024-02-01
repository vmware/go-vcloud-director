package govcd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"regexp"
	"strings"
	"time"
)

// cseConvertToCseClusterApiProviderClusterType takes a generic RDE that must represent an existing CSE Kubernetes cluster,
// and transforms it to a specific Container Service Extension CseKubernetesCluster object that represents the same cluster, but
// it is easy to explore and consume. If the receiver object does not contain a CAPVCD object, this method
// will obviously return an error.
func cseConvertToCseClusterApiProviderClusterType(rde *DefinedEntity) (*CseKubernetesCluster, error) {
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
	}
	for i, binding := range capvcd.Status.Capvcd.ClusterResourceSetBindings {
		result.ClusterResourceSetBindings[i] = binding.ClusterResourceSetName
	}

	if len(result.capvcdType.Status.Capvcd.VcdProperties.Organizations) == 0 {
		return nil, fmt.Errorf("could not read Organizations from Capvcd type")
	}
	result.OrganizationId = result.capvcdType.Status.Capvcd.VcdProperties.Organizations[0].Id
	if len(result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs) == 0 {
		return nil, fmt.Errorf("could not read Org VDC Network from Capvcd type")
	}
	result.VdcId = result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].Id
	result.NetworkId = result.capvcdType.Status.Capvcd.VcdProperties.OrgVdcs[0].OvdcNetworkName // TODO: ID

	if rde.DefinedEntity.Owner == nil {
		return nil, fmt.Errorf("could not read Owner from RDE")
	}
	result.Owner = rde.DefinedEntity.Owner.Name

	if result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.K8SStorageClassName != "" {
		result.DefaultStorageClass = &CseDefaultStorageClassSettings{
			StorageProfileId: "", // TODO: ID
			Name:             result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.K8SStorageClassName,
			ReclaimPolicy:    result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName,
			Filesystem:       result.capvcdType.Spec.VcdKe.DefaultStorageClassOptions.VcdStorageProfileName,
		}
	}

	yamlDocuments, err := unmarshalMultipleYamlDocuments(result.capvcdType.Spec.CapiYaml)
	if err != nil {
		return nil, err
	}

	result.KubernetesTemplateOvaId = "" // TODO: YAML > ID
	result.CseVersion = ""              // TODO: Get opposite from supportedVersionsMap
	// TODO: YAML > Control Plane
	// TODO: YAML > Worker pools
	// TODO: YAML > Health check
	// 			YAML PodCidr:            "",
	//			YAML ServiceCidr:        "",
	//			YAML SshPublicKey:       "",
	// 			YAML VirtualIpSubnet:    "",

	for _, yamlDocument := range yamlDocuments {
		switch yamlDocument["kind"] {

		}
	}
	if err != nil {
		return nil, fmt.Errorf("could not get the cluster state from the RDE contents: %s", err)
	}

	return result, nil
}

// waitUntilClusterIsProvisioned waits for the Kubernetes cluster to be in "provisioned" state, either indefinitely (if timeoutMinutes = 0)
// or until this timeout is reached. If the cluster is in "provisioned" state before the given timeout, it returns a CseKubernetesCluster object
// representing the Kubernetes cluster with all the latest information.
// If one of the states of the cluster at a given point is "error", this function also checks whether the cluster has the "Auto Repair on Errors" flag enabled,
// so it keeps waiting if it's true.
// If timeout is reached before the cluster, it returns an error.
func waitUntilClusterIsProvisioned(client *Client, clusterId string, timeoutMinutes time.Duration) (*CseKubernetesCluster, error) {
	var elapsed time.Duration
	logHttpResponse := util.LogHttpResponse
	sleepTime := 30

	// The following loop is constantly polling VCD to retrieve the RDE, which has a big JSON inside, so we avoid filling
	// the log with these big payloads. We use defer to be sure that we restore the initial logging state.
	defer func() {
		util.LogHttpResponse = logHttpResponse
	}()

	start := time.Now()
	var capvcdCluster *CseKubernetesCluster
	for elapsed <= timeoutMinutes*time.Minute || timeoutMinutes == 0 { // If the user specifies timeoutMinutes=0, we wait forever
		util.LogHttpResponse = false
		rde, err := getRdeById(client, clusterId)
		util.LogHttpResponse = logHttpResponse
		if err != nil {
			return nil, err
		}

		capvcdCluster, err = cseConvertToCseClusterApiProviderClusterType(rde)
		if err != nil {
			return nil, err
		}

		switch capvcdCluster.capvcdType.Status.VcdKe.State {
		case "provisioned":
			return capvcdCluster, nil
		case "error":
			// We just finish if auto-recovery is disabled, otherwise we just let CSE fixing things in background
			if !capvcdCluster.capvcdType.Spec.VcdKe.AutoRepairOnErrors {
				// Try to give feedback about what went wrong, which is located in a set of events in the RDE payload
				return capvcdCluster, fmt.Errorf("got an error and 'auto repair on errors' is disabled, aborting")
				// TODO				return capvcdCluster, fmt.Errorf("got an error and 'auto repair on errors' is disabled, aborting. Errors: %s", capvcdCluster.capvcdType.Status.Capvcd.ErrorSet[len(capvcdCluster.capvcdType.Status.Capvcd.ErrorSet)-1].AdditionalDetails.DetailedError)
			}
		}

		util.Logger.Printf("[DEBUG] Cluster '%s' is in '%s' state, will check again in %d seconds", capvcdCluster.ID, capvcdCluster.capvcdType.Status.VcdKe.State, sleepTime)
		elapsed = time.Since(start)
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	return capvcdCluster, fmt.Errorf("timeout of %d minutes reached, latest cluster state obtained was '%s'", timeoutMinutes, capvcdCluster.capvcdType.Status.VcdKe.State)
}

// validate validates the CSE Kubernetes cluster creation input data. Returns an error if some of the fields is wrong.
func (ccd *CseClusterSettings) validate() error {
	cseNamesRegex, err := regexp.Compile(`^[a-z](?:[a-z0-9-]{0,29}[a-z0-9])?$`)
	if err != nil {
		return fmt.Errorf("could not compile regular expression '%s'", err)
	}

	if !cseNamesRegex.MatchString(ccd.Name) {
		return fmt.Errorf("the cluster name is required and must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters, but it was: '%s'", ccd.Name)
	}

	if ccd.OrganizationId == "" {
		return fmt.Errorf("the Organization ID is required")
	}
	if ccd.VdcId == "" {
		return fmt.Errorf("the VDC ID is required")
	}
	if ccd.KubernetesTemplateOvaId == "" {
		return fmt.Errorf("the Kubernetes template OVA ID is required")
	}
	if ccd.NetworkId == "" {
		return fmt.Errorf("the Network ID is required")
	}
	if _, ok := supportedCseVersions[ccd.CseVersion]; !ok {
		return fmt.Errorf("the CSE version '%s' is not supported. Must be one of %v", ccd.CseVersion, getKeys(supportedCseVersions))
	}
	if ccd.ControlPlane.MachineCount < 1 || ccd.ControlPlane.MachineCount%2 == 0 {
		return fmt.Errorf("number of control plane nodes must be odd and higher than 0, but it was '%d'", ccd.ControlPlane.MachineCount)
	}
	if ccd.ControlPlane.DiskSizeGi < 20 {
		return fmt.Errorf("disk size for the Control Plane in Gibibytes (Gi) must be at least 20, but it was '%d'", ccd.ControlPlane.DiskSizeGi)
	}
	if len(ccd.WorkerPools) == 0 {
		return fmt.Errorf("there must be at least one Worker pool")
	}
	for _, workerPool := range ccd.WorkerPools {
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
	if ccd.DefaultStorageClass != nil {
		if !cseNamesRegex.MatchString(ccd.DefaultStorageClass.Name) {
			return fmt.Errorf("the Default Storage Class name is required and must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters, but it was: '%s'", ccd.DefaultStorageClass.Name)
		}
		if ccd.DefaultStorageClass.StorageProfileId == "" {
			return fmt.Errorf("the Storage Profile ID for the Default Storage Class is required")
		}
		if ccd.DefaultStorageClass.ReclaimPolicy != "delete" && ccd.DefaultStorageClass.ReclaimPolicy != "retain" {
			return fmt.Errorf("the reclaim policy for the Default Storage Class must be either 'delete' or 'retain', but it was '%s'", ccd.DefaultStorageClass.ReclaimPolicy)
		}
		if ccd.DefaultStorageClass.Filesystem != "ext4" && ccd.DefaultStorageClass.ReclaimPolicy != "xfs" {
			return fmt.Errorf("the filesystem for the Default Storage Class must be either 'ext4' or 'xfs', but it was '%s'", ccd.DefaultStorageClass.Filesystem)
		}
	}
	if ccd.ApiToken == "" {
		return fmt.Errorf("the API token is required")
	}
	if ccd.PodCidr == "" {
		return fmt.Errorf("the Pod CIDR is required")
	}
	if ccd.ServiceCidr == "" {
		return fmt.Errorf("the Service CIDR is required")
	}

	return nil
}

// cseClusterSettingsToInternal transforms user input data (CseClusterSettings) into the final payload that
// will be used to render the Go templates that define a Kubernetes cluster creation payload (cseClusterSettingsInternal).
func cseClusterSettingsToInternal(input CseClusterSettings, org *Org) (*cseClusterSettingsInternal, error) {
	err := input.validate()
	if err != nil {
		return nil, err
	}

	if org == nil || org.Org == nil {
		return nil, fmt.Errorf("cannot manipulate the CSE Kubernetes cluster creation input, the Organization is nil")
	}

	output := &cseClusterSettingsInternal{}
	output.OrganizationName = org.Org.Name

	vdc, err := org.GetVDCById(input.VdcId, true)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the VDC with ID '%s': %s", input.VdcId, err)
	}
	output.VdcName = vdc.Vdc.Name

	vAppTemplate, err := getVAppTemplateById(org.client, input.KubernetesTemplateOvaId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Kubernetes OVA with ID '%s': %s", input.KubernetesTemplateOvaId, err)
	}
	output.KubernetesTemplateOvaName = vAppTemplate.VAppTemplate.Name

	tkgVersions, err := getTkgVersionBundleFromVAppTemplateName(vAppTemplate.VAppTemplate.Name)
	if err != nil {
		return nil, err
	}
	output.TkgVersionBundle = tkgVersions

	catalogName, err := vAppTemplate.GetCatalogName()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Catalog name of the OVA '%s': %s", input.KubernetesTemplateOvaId, err)
	}
	output.CatalogName = catalogName

	network, err := vdc.GetOrgVdcNetworkById(input.NetworkId, true)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Org VDC Network with ID '%s': %s", input.NetworkId, err)
	}
	output.NetworkName = network.OrgVDCNetwork.Name

	currentCseVersion := supportedCseVersions[input.CseVersion]
	rdeType, err := getRdeType(org.client, "vmware", "capvcdCluster", currentCseVersion[1])
	if err != nil {
		return nil, fmt.Errorf("could not retrieve RDE Type vmware:capvcdCluster:'%s': %s", currentCseVersion[1], err)
	}
	output.RdeType = rdeType.DefinedEntityType

	// The input to create a cluster uses different entities IDs, but CSE cluster creation process uses names.
	// For that reason, we need to transform IDs to Names by querying VCD. This process is optimized with a tiny "cache" map.
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
				return nil, fmt.Errorf("could not get Storage Profile with ID '%s': %s", id, err)
			}
			idToNameCache[id] = storageProfile.Name
		}
	}
	for _, id := range computePolicyIds {
		if _, alreadyPresent := idToNameCache[id]; !alreadyPresent {
			computePolicy, err := getVdcComputePolicyV2ById(org.client, id)
			if err != nil {
				return nil, fmt.Errorf("could not get Compute Policy with ID '%s': %s", id, err)
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

	mhc, err := getMachineHealthCheck(org.client, supportedCseVersions[input.CseVersion][0], input.NodeHealthCheck)
	if err != nil {
		return nil, err
	}
	if mhc != nil {
		output.MachineHealthCheck = mhc
	}

	containerRegistryUrl, err := getContainerRegistryUrl(org.client, input.CseVersion)
	if err != nil {
		return nil, err
	}

	output.ContainerRegistryUrl = containerRegistryUrl

	output.Owner = input.Owner
	if input.Owner == "" {
		sessionInfo, err := org.client.GetSessionInfo()
		if err != nil {
			return nil, fmt.Errorf("error getting the owner of the cluster: %s", err)
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
		return result, err
	}
	versionMap, ok := versionsMap[parsedOvaName]
	if !ok {
		return result, fmt.Errorf("the Kubernetes OVA '%s' is not supported", parsedOvaName)
	}

	// The map checking above guarantees that all splits and replaces will work
	result.KubernetesVersion = strings.Split(parsedOvaName, "-")[0]
	result.TkrVersion = strings.ReplaceAll(strings.Split(parsedOvaName, "-")[0], "+", "---") + "-" + strings.Split(parsedOvaName, "-")[1]
	result.TkgVersion = versionMap.(map[string]any)["tkg"].(string)
	result.EtcdVersion = versionMap.(map[string]any)["etcd"].(string)
	result.CoreDnsVersion = versionMap.(map[string]any)["coreDns"].(string)
	return result, nil
}

// getMachineHealthCheck gets the required information from the CSE Server configuration RDE
func getMachineHealthCheck(client *Client, vcdKeConfigVersion string, isNodeHealthCheckActive bool) (*cseMachineHealthCheckInternal, error) {
	if !isNodeHealthCheckActive {
		return nil, nil
	}

	rdes, err := getRdesByName(client, "vmware", "VCDKEConfig", vcdKeConfigVersion, "vcdKeConfig")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve VCDKEConfig RDE with version %s: %s", vcdKeConfigVersion, err)
	}
	if len(rdes) != 1 {
		return nil, fmt.Errorf("expected exactly one VCDKEConfig RDE but got %d", len(rdes))
	}
	// TODO: Get the struct Type for this one
	profiles, ok := rdes[0].DefinedEntity.Entity["profiles"].([]any)
	if !ok {
		return nil, fmt.Errorf("wrong format of VCDKEConfig, expected a 'profiles' array")
	}
	if len(profiles) != 1 {
		return nil, fmt.Errorf("wrong format of VCDKEConfig, expected a single 'profiles' element, got %d", len(profiles))
	}

	mhc, ok := profiles[0].(map[string]any)["K8Config"].(map[string]any)["mhc"].(map[string]any)
	if !ok {
		return nil, nil
	}
	result := cseMachineHealthCheckInternal{}
	result.MaxUnhealthyNodesPercentage = mhc["maxUnhealthyNodes"].(float64)
	result.NodeStartupTimeout = mhc["nodeStartupTimeout"].(string)
	result.NodeNotReadyTimeout = mhc["nodeUnknownTimeout"].(string)
	result.NodeUnknownTimeout = mhc["nodeNotReadyTimeout"].(string)
	return &result, nil
}

// getContainerRegistryUrl gets the required information from the CSE Server configuration RDE
func getContainerRegistryUrl(client *Client, cseVersion string) (string, error) {
	currentCseVersion := supportedCseVersions[cseVersion]

	rdes, err := getRdesByName(client, "vmware", "VCDKEConfig", currentCseVersion[0], "vcdKeConfig")
	if err != nil {
		return "", fmt.Errorf("could not retrieve VCDKEConfig RDE with version %s: %s", currentCseVersion[0], err)
	}
	if len(rdes) != 1 {
		return "", fmt.Errorf("expected exactly one VCDKEConfig RDE but got %d", len(rdes))
	}
	// TODO: Get the struct Type for this one
	profiles, ok := rdes[0].DefinedEntity.Entity["profiles"].([]any)
	if !ok {
		return "", fmt.Errorf("wrong format of VCDKEConfig, expected a 'profiles' array")
	}
	if len(profiles) != 1 {
		return "", fmt.Errorf("wrong format of VCDKEConfig, expected a single 'profiles' element, got %d", len(profiles))
	}
	// TODO: Check airgapped environments: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1.1a/VMware-Cloud-Director-Container-Service-Extension-Install-provider-4.1.1/GUID-F00BE796-B5F2-48F2-A012-546E2E694400.html
	return fmt.Sprintf("%s/tkg", profiles[0].(map[string]any)["containerRegistryUrl"].(string)), nil
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
