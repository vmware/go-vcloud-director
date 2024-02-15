package govcd

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

// This collection of files contains all the Go Templates and resources required for the Container Service Extension (CSE) methods
// to work.
//
//go:embed cse
var cseFiles embed.FS

// getUnmarshaledRdePayload gets the unmarshaled JSON payload to create the Runtime Defined Entity that represents
// a CSE Kubernetes cluster, by using the receiver information. This method uses all the Go Templates stored in cseFiles
func (clusterSettings *cseClusterSettingsInternal) getUnmarshaledRdePayload() (map[string]interface{}, error) {
	if clusterSettings == nil {
		return nil, fmt.Errorf("the receiver cluster settings is nil")
	}
	capiYaml, err := clusterSettings.generateCapiYamlAsJsonString()
	if err != nil {
		return nil, err
	}

	args := map[string]string{
		"Name":               clusterSettings.Name,
		"Org":                clusterSettings.OrganizationName,
		"VcdUrl":             clusterSettings.VcdUrl,
		"Vdc":                clusterSettings.VdcName,
		"Delete":             "false",
		"ForceDelete":        "false",
		"AutoRepairOnErrors": strconv.FormatBool(clusterSettings.AutoRepairOnErrors),
		"ApiToken":           clusterSettings.ApiToken,
		"CapiYaml":           capiYaml,
	}

	if clusterSettings.DefaultStorageClass.StorageProfileName != "" {
		args["DefaultStorageClassStorageProfile"] = clusterSettings.DefaultStorageClass.StorageProfileName
		args["DefaultStorageClassName"] = clusterSettings.DefaultStorageClass.Name
		args["DefaultStorageClassUseDeleteReclaimPolicy"] = strconv.FormatBool(clusterSettings.DefaultStorageClass.UseDeleteReclaimPolicy)
		args["DefaultStorageClassFileSystem"] = clusterSettings.DefaultStorageClass.Filesystem
	}

	rdeTmpl, err := getCseTemplate(clusterSettings.CseVersion, "rde")
	if err != nil {
		return nil, err
	}

	capvcdEmpty := template.Must(template.New(clusterSettings.Name).Parse(rdeTmpl))
	buf := &bytes.Buffer{}
	if err := capvcdEmpty.Execute(buf, args); err != nil {
		return nil, fmt.Errorf("could not render the Go template with the CAPVCD JSON: %s", err)
	}

	var result interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return nil, fmt.Errorf("could not generate a correct CAPVCD JSON: %s", err)
	}

	return result.(map[string]interface{}), nil
}

// generateCapiYamlAsJsonString generates the "capiYaml" property of the RDE that represents a Kubernetes cluster. This
// "capiYaml" property is a YAML encoded as a JSON string. This method uses the Go Templates stored in cseFiles.
func (clusterSettings *cseClusterSettingsInternal) generateCapiYamlAsJsonString() (string, error) {
	if clusterSettings == nil {
		return "", fmt.Errorf("the receiver cluster settings is nil")
	}

	clusterTmpl, err := getCseTemplate(clusterSettings.CseVersion, "capiyaml_cluster")
	if err != nil {
		return "", err
	}

	// This YAML snippet contains special strings, such as "%,", that render wrong using the Go template engine
	sanitizedTemplate := strings.NewReplacer("%", "%%").Replace(clusterTmpl)
	capiYamlEmpty := template.Must(template.New(clusterSettings.Name + "-cluster").Parse(sanitizedTemplate))

	nodePoolYaml, err := clusterSettings.generateWorkerPoolsYaml()
	if err != nil {
		return "", err
	}

	memoryHealthCheckYaml, err := clusterSettings.generateMachineHealthCheckYaml()
	if err != nil {
		return "", err
	}

	args := map[string]string{
		"ClusterName":                 clusterSettings.Name,
		"TargetNamespace":             clusterSettings.Name + "-ns",
		"TkrVersion":                  clusterSettings.TkgVersionBundle.TkrVersion,
		"TkgVersion":                  clusterSettings.TkgVersionBundle.TkgVersion,
		"UsernameB64":                 base64.StdEncoding.EncodeToString([]byte(clusterSettings.Owner)),
		"ApiTokenB64":                 base64.StdEncoding.EncodeToString([]byte(clusterSettings.ApiToken)),
		"PodCidr":                     clusterSettings.PodCidr,
		"ServiceCidr":                 clusterSettings.ServiceCidr,
		"VcdSite":                     clusterSettings.VcdUrl,
		"Org":                         clusterSettings.OrganizationName,
		"OrgVdc":                      clusterSettings.VdcName,
		"OrgVdcNetwork":               clusterSettings.NetworkName,
		"Catalog":                     clusterSettings.CatalogName,
		"VAppTemplate":                clusterSettings.KubernetesTemplateOvaName,
		"ControlPlaneSizingPolicy":    clusterSettings.ControlPlane.SizingPolicyName,
		"ControlPlanePlacementPolicy": clusterSettings.ControlPlane.PlacementPolicyName,
		"ControlPlaneStorageProfile":  clusterSettings.ControlPlane.StorageProfileName,
		"ControlPlaneDiskSize":        fmt.Sprintf("%dGi", clusterSettings.ControlPlane.DiskSizeGi),
		"ControlPlaneMachineCount":    strconv.Itoa(clusterSettings.ControlPlane.MachineCount),
		"ControlPlaneEndpoint":        clusterSettings.ControlPlane.Ip,
		"DnsVersion":                  clusterSettings.TkgVersionBundle.CoreDnsVersion,
		"EtcdVersion":                 clusterSettings.TkgVersionBundle.EtcdVersion,
		"ContainerRegistryUrl":        clusterSettings.VcdKeConfig.ContainerRegistryUrl,
		"KubernetesVersion":           clusterSettings.TkgVersionBundle.KubernetesVersion,
		"SshPublicKey":                clusterSettings.SshPublicKey,
		"VirtualIpSubnet":             clusterSettings.VirtualIpSubnet,
	}

	buf := &bytes.Buffer{}
	if err := capiYamlEmpty.Execute(buf, args); err != nil {
		return "", fmt.Errorf("could not generate a correct CAPI YAML: %s", err)
	}

	prettyYaml := ""
	if memoryHealthCheckYaml != "" {
		prettyYaml += fmt.Sprintf("%s\n---\n", memoryHealthCheckYaml)
	}
	// The final "pretty" YAML. To embed it in the final payload it must be marshaled into a one-line JSON string
	prettyYaml += fmt.Sprintf("%s\n---\n%s", nodePoolYaml, buf.String())

	// We don't use a standard json.Marshal() as the YAML contains special characters that are not encoded properly, such as '<'.
	buf.Reset()
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err = enc.Encode(prettyYaml)
	if err != nil {
		return "", fmt.Errorf("could not encode the CAPI YAML into JSON: %s", err)
	}

	// Removes trailing quotes from the final JSON string
	return strings.Trim(strings.TrimSpace(buf.String()), "\""), nil
}

// generateWorkerPoolsYaml generates YAML blocks corresponding to the cluster Worker Pools. The blocks are separated by
// the standard YAML separator (---), but does not add one at the end.
func (clusterSettings *cseClusterSettingsInternal) generateWorkerPoolsYaml() (string, error) {
	if clusterSettings == nil {
		return "", fmt.Errorf("the receiver cluster settings is nil")
	}

	workerPoolTmpl, err := getCseTemplate(clusterSettings.CseVersion, "capiyaml_workerpool")
	if err != nil {
		return "", err
	}

	nodePoolEmptyTmpl := template.Must(template.New(clusterSettings.Name + "-worker-pool").Parse(workerPoolTmpl))
	resultYaml := ""
	buf := &bytes.Buffer{}

	// We can have many Worker Pools, we build a YAML object for each one of them.
	for i, workerPool := range clusterSettings.WorkerPools {

		// Check the correctness of the Compute Policies in the node pool block
		if workerPool.PlacementPolicyName != "" && workerPool.VGpuPolicyName != "" {
			return "", fmt.Errorf("the worker pool '%s' should have either a Placement Policy or a vGPU Policy, not both", workerPool.Name)
		}
		placementPolicy := workerPool.PlacementPolicyName
		if workerPool.VGpuPolicyName != "" {
			// For convenience, we just use one of the variables as both cannot be set at same time
			placementPolicy = workerPool.VGpuPolicyName
		}

		if err := nodePoolEmptyTmpl.Execute(buf, map[string]string{
			"ClusterName":             clusterSettings.Name,
			"NodePoolName":            workerPool.Name,
			"TargetNamespace":         clusterSettings.Name + "-ns",
			"Catalog":                 clusterSettings.CatalogName,
			"VAppTemplate":            clusterSettings.KubernetesTemplateOvaName,
			"NodePoolSizingPolicy":    workerPool.SizingPolicyName,
			"NodePoolPlacementPolicy": placementPolicy, // Can be either Placement or vGPU policy
			"NodePoolStorageProfile":  workerPool.StorageProfileName,
			"NodePoolDiskSize":        fmt.Sprintf("%dGi", workerPool.DiskSizeGi),
			"NodePoolEnableGpu":       strconv.FormatBool(workerPool.VGpuPolicyName != ""),
			"NodePoolMachineCount":    strconv.Itoa(workerPool.MachineCount),
			"KubernetesVersion":       clusterSettings.TkgVersionBundle.KubernetesVersion,
		}); err != nil {
			return "", fmt.Errorf("could not generate a correct Node Pool YAML: %s", err)
		}
		resultYaml += fmt.Sprintf("%s\n", buf.String())
		if i < len(clusterSettings.WorkerPools)-1 {
			resultYaml += "---\n"
		}
		buf.Reset()
	}
	return resultYaml, nil
}

// generateMachineHealthCheckYaml generates a YAML block corresponding to the cluster Machine Health Check.
// The generated YAML does not contain a separator (---) at the end.
func (clusterSettings *cseClusterSettingsInternal) generateMachineHealthCheckYaml() (string, error) {
	if clusterSettings == nil {
		return "", fmt.Errorf("the receiver cluster settings is nil")
	}

	if clusterSettings.VcdKeConfig.NodeStartupTimeout == "" &&
		clusterSettings.VcdKeConfig.NodeUnknownTimeout == "" &&
		clusterSettings.VcdKeConfig.NodeNotReadyTimeout == "" &&
		clusterSettings.VcdKeConfig.MaxUnhealthyNodesPercentage == 0 {
		return "", nil
	}

	mhcTmpl, err := getCseTemplate(clusterSettings.CseVersion, "capiyaml_mhc")
	if err != nil {
		return "", err
	}

	mhcEmptyTmpl := template.Must(template.New(clusterSettings.Name + "-mhc").Parse(mhcTmpl))
	buf := &bytes.Buffer{}

	if err := mhcEmptyTmpl.Execute(buf, map[string]string{
		"ClusterName":     clusterSettings.Name,
		"TargetNamespace": clusterSettings.Name + "-ns",
		// With the 'percentage' suffix
		"MaxUnhealthyNodePercentage": fmt.Sprintf("%.0f%%", clusterSettings.VcdKeConfig.MaxUnhealthyNodesPercentage),
		// These values coming from VCDKEConfig (CSE Server settings) may have an "s" suffix. We make sure we don't duplicate it
		"NodeStartupTimeout":  fmt.Sprintf("%ss", strings.ReplaceAll(clusterSettings.VcdKeConfig.NodeStartupTimeout, "s", "")),
		"NodeUnknownTimeout":  fmt.Sprintf("%ss", strings.ReplaceAll(clusterSettings.VcdKeConfig.NodeUnknownTimeout, "s", "")),
		"NodeNotReadyTimeout": fmt.Sprintf("%ss", strings.ReplaceAll(clusterSettings.VcdKeConfig.NodeNotReadyTimeout, "s", "")),
	}); err != nil {
		return "", fmt.Errorf("could not generate a correct Machine Health Check YAML: %s", err)
	}
	return fmt.Sprintf("%s\n", buf.String()), nil

}
