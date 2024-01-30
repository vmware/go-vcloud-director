package govcd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"regexp"
	"strings"
)

// cseClusterCreationGoTemplateArguments defines the required arguments that are required by the Go templates used internally to specify
// a Kubernetes cluster. These are not set by the user, but instead they are computed from a valid
// CseClusterCreateInput object in the CseClusterCreateInput.toCseClusterCreationGoTemplateContents method. These fields are then
// inserted in Go templates to render a final JSON that is valid to be used as the cluster Runtime Defined Entity (RDE) payload.
type cseClusterCreationGoTemplateArguments struct {
	CseVersion                string
	Name                      string
	OrganizationId            string
	OrganizationName          string
	VdcName                   string
	NetworkName               string
	KubernetesTemplateOvaName string
	TkgVersionBundle          tkgVersionBundle
	CatalogName               string
	RdeType                   *types.DefinedEntityType
	ControlPlane              controlPlane
	WorkerPools               []workerPool
	DefaultStorageClass       defaultStorageClass
	MachineHealthCheck        *machineHealthCheck
	Owner                     string
	ApiToken                  string
	VcdUrl                    string
	ContainerRegistryUrl      string
	VirtualIpSubnet           string
	SshPublicKey              string
	PodCidr                   string
	ServiceCidr               string
	AutoRepairOnErrors        bool
}

// controlPlane defines the Control Plane inside cseClusterCreationGoTemplateArguments
type controlPlane struct {
	MachineCount        int
	DiskSizeGi          int
	SizingPolicyName    string
	PlacementPolicyName string
	StorageProfileName  string
	Ip                  string
}

// workerPool defines a Worker pool inside cseClusterCreationGoTemplateArguments
type workerPool struct {
	Name                string
	MachineCount        int
	DiskSizeGi          int
	SizingPolicyName    string
	PlacementPolicyName string
	VGpuPolicyName      string
	StorageProfileName  string
}

// defaultStorageClass defines a Default Storage Class inside cseClusterCreationGoTemplateArguments
type defaultStorageClass struct {
	StorageProfileName     string
	Name                   string
	UseDeleteReclaimPolicy bool
	Filesystem             string
}

// machineHealthCheck is a type that contains only the required and relevant fields from the Container Service Extension (CSE) installation configuration,
// such as the Machine Health Check settings or the container registry URL.
type machineHealthCheck struct {
	MaxUnhealthyNodesPercentage float64
	NodeStartupTimeout          string
	NodeNotReadyTimeout         string
	NodeUnknownTimeout          string
}

// validate validates the CSE Kubernetes cluster creation input data. Returns an error if some of the fields is wrong.
func (ccd *CseClusterCreateInput) validate() error {
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

// toCseClusterCreationGoTemplateContents transforms user input data (receiver CseClusterCreateInput) into the final payload that
// will be used to render the Go templates that define a Kubernetes cluster creation payload (cseClusterCreationGoTemplateArguments).
func (input *CseClusterCreateInput) toCseClusterCreationGoTemplateContents(vcdClient *VCDClient) (*cseClusterCreationGoTemplateArguments, error) {
	err := input.validate()
	if err != nil {
		return nil, err
	}

	output := &cseClusterCreationGoTemplateArguments{}
	org, err := vcdClient.GetOrgById(input.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the Organization with ID '%s': %s", input.VdcId, err)
	}
	output.OrganizationId = org.Org.ID
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
	rdeType, err := vcdClient.GetRdeType("vmware", "capvcdCluster", currentCseVersion[1])
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
			storageProfile, err := vcdClient.GetStorageProfileById(id)
			if err != nil {
				return nil, fmt.Errorf("could not get Storage Profile with ID '%s': %s", id, err)
			}
			idToNameCache[id] = storageProfile.Name
		}
	}
	for _, id := range computePolicyIds {
		if _, alreadyPresent := idToNameCache[id]; !alreadyPresent {
			computePolicy, err := vcdClient.GetVdcComputePolicyV2ById(id)
			if err != nil {
				return nil, fmt.Errorf("could not get Compute Policy with ID '%s': %s", id, err)
			}
			idToNameCache[id] = computePolicy.VdcComputePolicyV2.Name
		}
	}

	// Now that everything is cached in memory, we can build the Node pools and Storage Class payloads
	output.WorkerPools = make([]workerPool, len(input.WorkerPools))
	for i, w := range input.WorkerPools {
		output.WorkerPools[i] = workerPool{
			Name:         w.Name,
			MachineCount: w.MachineCount,
			DiskSizeGi:   w.DiskSizeGi,
		}
		output.WorkerPools[i].SizingPolicyName = idToNameCache[w.SizingPolicyId]
		output.WorkerPools[i].PlacementPolicyName = idToNameCache[w.PlacementPolicyId]
		output.WorkerPools[i].VGpuPolicyName = idToNameCache[w.VGpuPolicyId]
		output.WorkerPools[i].StorageProfileName = idToNameCache[w.StorageProfileId]
	}
	output.ControlPlane = controlPlane{
		MachineCount:        input.ControlPlane.MachineCount,
		DiskSizeGi:          input.ControlPlane.DiskSizeGi,
		SizingPolicyName:    idToNameCache[input.ControlPlane.SizingPolicyId],
		PlacementPolicyName: idToNameCache[input.ControlPlane.PlacementPolicyId],
		StorageProfileName:  idToNameCache[input.ControlPlane.StorageProfileId],
		Ip:                  input.ControlPlane.Ip,
	}

	if input.DefaultStorageClass != nil {
		output.DefaultStorageClass = defaultStorageClass{
			StorageProfileName: idToNameCache[input.DefaultStorageClass.StorageProfileId],
			Name:               input.DefaultStorageClass.Name,
			Filesystem:         input.DefaultStorageClass.Filesystem,
		}
		output.DefaultStorageClass.UseDeleteReclaimPolicy = false
		if input.DefaultStorageClass.ReclaimPolicy == "delete" {
			output.DefaultStorageClass.UseDeleteReclaimPolicy = true
		}
	}

	mhc, err := getMachineHealthCheck(vcdClient, input.CseVersion, input.NodeHealthCheck)
	if err != nil {
		return nil, err
	}
	if mhc != nil {
		output.MachineHealthCheck = mhc
	}

	containerRegistryUrl, err := getContainerRegistryUrl(vcdClient, input.CseVersion)
	if err != nil {
		return nil, err
	}

	output.ContainerRegistryUrl = containerRegistryUrl

	output.Owner = input.Owner
	if input.Owner == "" {
		sessionInfo, err := vcdClient.Client.GetSessionInfo()
		if err != nil {
			return nil, fmt.Errorf("error getting the owner of the cluster: %s", err)
		}
		output.Owner = sessionInfo.User.Name
	}
	output.VcdUrl = strings.Replace(vcdClient.Client.VCDHREF.String(), "/api", "", 1)

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

	versionsMap := map[string]interface{}{}
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
	result.TkgVersion = versionMap.(map[string]interface{})["tkg"].(string)
	result.EtcdVersion = versionMap.(map[string]interface{})["etcd"].(string)
	result.CoreDnsVersion = versionMap.(map[string]interface{})["coreDns"].(string)
	return result, nil
}

// getMachineHealthCheck gets the required information from the CSE Server configuration RDE
func getMachineHealthCheck(vcdClient *VCDClient, cseVersion string, isNodeHealthCheckActive bool) (*machineHealthCheck, error) {
	if !isNodeHealthCheckActive {
		return nil, nil
	}
	currentCseVersion := supportedCseVersions[cseVersion]

	rdes, err := vcdClient.GetRdesByName("vmware", "VCDKEConfig", currentCseVersion[0], "vcdKeConfig")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve VCDKEConfig RDE with version %s: %s", currentCseVersion[0], err)
	}
	if len(rdes) != 1 {
		return nil, fmt.Errorf("expected exactly one VCDKEConfig RDE but got %d", len(rdes))
	}
	// TODO: Get the struct Type for this one
	profiles, ok := rdes[0].DefinedEntity.Entity["profiles"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("wrong format of VCDKEConfig, expected a 'profiles' array")
	}
	if len(profiles) != 1 {
		return nil, fmt.Errorf("wrong format of VCDKEConfig, expected a single 'profiles' element, got %d", len(profiles))
	}

	// TODO: Get the struct Type for this one
	result := machineHealthCheck{}
	mhc := profiles[0].(map[string]interface{})["K8Config"].(map[string]interface{})["mhc"].(map[string]interface{})
	result.MaxUnhealthyNodesPercentage = mhc["maxUnhealthyNodes"].(float64)
	result.NodeStartupTimeout = mhc["nodeStartupTimeout"].(string)
	result.NodeNotReadyTimeout = mhc["nodeUnknownTimeout"].(string)
	result.NodeUnknownTimeout = mhc["nodeNotReadyTimeout"].(string)
	return &result, nil
}

// getContainerRegistryUrl gets the required information from the CSE Server configuration RDE
func getContainerRegistryUrl(vcdClient *VCDClient, cseVersion string) (string, error) {
	currentCseVersion := supportedCseVersions[cseVersion]

	rdes, err := vcdClient.GetRdesByName("vmware", "VCDKEConfig", currentCseVersion[0], "vcdKeConfig")
	if err != nil {
		return "", fmt.Errorf("could not retrieve VCDKEConfig RDE with version %s: %s", currentCseVersion[0], err)
	}
	if len(rdes) != 1 {
		return "", fmt.Errorf("expected exactly one VCDKEConfig RDE but got %d", len(rdes))
	}
	// TODO: Get the struct Type for this one
	profiles, ok := rdes[0].DefinedEntity.Entity["profiles"].([]interface{})
	if !ok {
		return "", fmt.Errorf("wrong format of VCDKEConfig, expected a 'profiles' array")
	}
	if len(profiles) != 1 {
		return "", fmt.Errorf("wrong format of VCDKEConfig, expected a single 'profiles' element, got %d", len(profiles))
	}
	// TODO: Check airgapped environments: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1.1a/VMware-Cloud-Director-Container-Service-Extension-Install-provider-4.1.1/GUID-F00BE796-B5F2-48F2-A012-546E2E694400.html
	return fmt.Sprintf("%s/tkg", profiles[0].(map[string]interface{})["containerRegistryUrl"].(string)), nil
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
