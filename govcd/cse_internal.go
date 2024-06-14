package govcd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	semver "github.com/hashicorp/go-version"
	"strconv"
	"strings"
	"text/template"
)

// This collection of files contains all the Go Templates and resources required for the Container Service Extension (CSE) methods
// to work.
//
//go:embed cse
var cseFiles embed.FS

// getUnmarshalledRdePayload gets the unmarshalled JSON payload to create the Runtime Defined Entity that represents
// a CSE Kubernetes cluster, by using the receiver information. This method uses all the Go Templates stored in cseFiles
func (clusterSettings *cseClusterSettingsInternal) getUnmarshalledRdePayload() (map[string]interface{}, error) {
	if clusterSettings == nil {
		return nil, fmt.Errorf("the receiver CSE Kubernetes cluster settings object is nil")
	}
	capiYaml, err := clusterSettings.generateCapiYamlAsJsonString()
	if err != nil {
		return nil, err
	}

	templateArgs := map[string]string{
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
		templateArgs["DefaultStorageClassStorageProfile"] = clusterSettings.DefaultStorageClass.StorageProfileName
		templateArgs["DefaultStorageClassName"] = clusterSettings.DefaultStorageClass.Name
		templateArgs["DefaultStorageClassUseDeleteReclaimPolicy"] = strconv.FormatBool(clusterSettings.DefaultStorageClass.UseDeleteReclaimPolicy)
		templateArgs["DefaultStorageClassFileSystem"] = clusterSettings.DefaultStorageClass.Filesystem
	}

	rdeTemplate, err := getCseTemplate(clusterSettings.CseVersion, "rde")
	if err != nil {
		return nil, err
	}

	rdePayload := template.Must(template.New(clusterSettings.Name).Parse(rdeTemplate))
	buf := &bytes.Buffer{}
	if err := rdePayload.Execute(buf, templateArgs); err != nil {
		return nil, fmt.Errorf("could not render the Go template with the RDE JSON: %s", err)
	}

	var result interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return nil, fmt.Errorf("could not generate a correct RDE payload: %s", err)
	}

	return result.(map[string]interface{}), nil
}

// generateCapiYamlAsJsonString generates the "capiYaml" property of the RDE that represents a Kubernetes cluster. This
// "capiYaml" property is a YAML encoded as a JSON string. This method uses the Go Templates stored in cseFiles.
func (clusterSettings *cseClusterSettingsInternal) generateCapiYamlAsJsonString() (string, error) {
	if clusterSettings == nil {
		return "", fmt.Errorf("the receiver cluster settings is nil")
	}

	capiYamlTemplate, err := getCseTemplate(clusterSettings.CseVersion, "capiyaml_cluster")
	if err != nil {
		return "", err
	}

	// This YAML snippet contains special strings, such as "%,", that render wrong using the Go template engine
	sanitizedCapiYamlTemplate := strings.NewReplacer("%", "%%").Replace(capiYamlTemplate)
	capiYaml := template.Must(template.New(clusterSettings.Name + "-cluster").Parse(sanitizedCapiYamlTemplate))

	nodePoolYaml, err := clusterSettings.generateWorkerPoolsYaml()
	if err != nil {
		return "", err
	}

	memoryHealthCheckYaml, err := clusterSettings.generateMachineHealthCheckYaml()
	if err != nil {
		return "", err
	}

	autoscalerNeeded := false
	// We'll need the Autoscaler YAML document if at least one Worker Pool uses it
	for _, wp := range clusterSettings.WorkerPools {
		if wp.Autoscaler != nil {
			autoscalerNeeded = true
			break
		}
	}
	autoscalerYaml := ""
	if autoscalerNeeded {
		autoscalerYaml, err = clusterSettings.generateAutoscalerYaml()
		if err != nil {
			return "", err
		}
	}

	templateArgs := map[string]interface{}{
		"ClusterName":                 clusterSettings.Name,
		"TargetNamespace":             clusterSettings.Name + "-ns",
		"TkrVersion":                  clusterSettings.TkgVersionBundle.TkrVersion,
		"TkgVersion":                  clusterSettings.TkgVersionBundle.TkgVersion,
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
		"Base64Certificates":          clusterSettings.VcdKeConfig.Base64Certificates,
	}

	buf := &bytes.Buffer{}
	if err := capiYaml.Execute(buf, templateArgs); err != nil {
		return "", fmt.Errorf("could not generate a correct CAPI YAML: %s", err)
	}

	// The final "pretty" YAML. To embed it in the final payload it must be marshaled into a one-line JSON string
	prettyYaml := ""
	if memoryHealthCheckYaml != "" {
		prettyYaml += fmt.Sprintf("%s\n---\n", memoryHealthCheckYaml)
	}
	if autoscalerYaml != "" {
		prettyYaml += fmt.Sprintf("%s\n---\n", autoscalerYaml)
	}
	prettyYaml += fmt.Sprintf("%s\n---\n%s", nodePoolYaml, buf.String())

	// We don't use a standard json.Marshal() as the YAML contains special characters that are not encoded properly, such as '<'.
	buf.Reset()
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err = enc.Encode(prettyYaml)
	if err != nil {
		return "", fmt.Errorf("could not encode the CAPI YAML into a JSON string: %s", err)
	}

	// Removes trailing quotes from the final JSON string
	return strings.Trim(strings.TrimSpace(buf.String()), "\""), nil
}

// generateWorkerPoolsYaml generates YAML blocks corresponding to the cluster Worker Pools. The blocks are separated by
// the standard YAML separator (---), but does not add one at the end.
func (clusterSettings *cseClusterSettingsInternal) generateWorkerPoolsYaml() (string, error) {
	if clusterSettings == nil {
		return "", fmt.Errorf("the receiver CSE Kubernetes cluster settings object is nil")
	}

	workerPoolsTemplate, err := getCseTemplate(clusterSettings.CseVersion, "capiyaml_workerpool")
	if err != nil {
		return "", err
	}

	workerPools := template.Must(template.New(clusterSettings.Name + "-worker-pool").Parse(workerPoolsTemplate))
	resultYaml := ""
	buf := &bytes.Buffer{}

	// We can have many Worker Pools, we build a YAML object for each one of them.
	for i, wp := range clusterSettings.WorkerPools {

		// Check the correctness of the Compute Policies in the node pool block
		if wp.PlacementPolicyName != "" && wp.VGpuPolicyName != "" {
			return "", fmt.Errorf("the Worker Pool '%s' should have either a Placement Policy or a vGPU Policy, not both", wp.Name)
		}
		placementPolicy := wp.PlacementPolicyName
		if wp.VGpuPolicyName != "" {
			// For convenience, we just use one of the variables as both cannot be set at same time
			placementPolicy = wp.VGpuPolicyName
		}

		args := map[string]string{
			"ClusterName":             clusterSettings.Name,
			"NodePoolName":            wp.Name,
			"TargetNamespace":         clusterSettings.Name + "-ns",
			"Catalog":                 clusterSettings.CatalogName,
			"VAppTemplate":            clusterSettings.KubernetesTemplateOvaName,
			"NodePoolSizingPolicy":    wp.SizingPolicyName,
			"NodePoolPlacementPolicy": placementPolicy, // Can be either Placement or vGPU policy
			"NodePoolStorageProfile":  wp.StorageProfileName,
			"NodePoolDiskSize":        fmt.Sprintf("%dGi", wp.DiskSizeGi),
			"NodePoolEnableGpu":       strconv.FormatBool(wp.VGpuPolicyName != ""),
			"KubernetesVersion":       clusterSettings.TkgVersionBundle.KubernetesVersion,
		}

		if wp.Autoscaler != nil {
			args["AutoscalerMaxSize"] = strconv.Itoa(wp.Autoscaler.MaxSize)
			args["AutoscalerMinSize"] = strconv.Itoa(wp.Autoscaler.MinSize)
		} else {
			args["NodePoolMachineCount"] = strconv.Itoa(wp.MachineCount)
		}

		if err := workerPools.Execute(buf, args); err != nil {
			return "", fmt.Errorf("could not generate a correct Worker Pool '%s' YAML block: %s", wp.Name, err)
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
		return "", fmt.Errorf("the receiver CSE Kubernetes cluster settings object is nil")
	}

	if clusterSettings.VcdKeConfig.NodeStartupTimeout == "" &&
		clusterSettings.VcdKeConfig.NodeUnknownTimeout == "" &&
		clusterSettings.VcdKeConfig.NodeNotReadyTimeout == "" &&
		clusterSettings.VcdKeConfig.MaxUnhealthyNodesPercentage == 0 {
		return "", nil
	}

	mhcTemplate, err := getCseTemplate(clusterSettings.CseVersion, "capiyaml_mhc")
	if err != nil {
		return "", err
	}

	machineHealthCheck := template.Must(template.New(clusterSettings.Name + "-mhc").Parse(mhcTemplate))
	buf := &bytes.Buffer{}

	if err := machineHealthCheck.Execute(buf, map[string]string{
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

// generateAutoscalerYaml generates YAML documents corresponding to the cluster Autoscaler
func (clusterSettings *cseClusterSettingsInternal) generateAutoscalerYaml() (string, error) {
	if clusterSettings == nil {
		return "", fmt.Errorf("the receiver CSE Kubernetes cluster settings object is nil")
	}

	k8sVersion, err := semver.NewVersion(clusterSettings.TkgVersionBundle.KubernetesVersion)
	if err != nil {
		return "", err
	}
	k8sVersionSegments := k8sVersion.Segments()

	autoscalerTemplate, err := getCseTemplate(clusterSettings.CseVersion, "autoscaler")
	if err != nil {
		return "", err
	}

	autoscaler := template.Must(template.New(clusterSettings.Name + "-autoscaler").Parse(autoscalerTemplate))
	resultYaml := ""
	buf := &bytes.Buffer{}

	if err := autoscaler.Execute(buf, map[string]string{
		"TargetNamespace":    clusterSettings.Name + "-ns",
		"AutoscalerReplicas": "1",
		"AutoscalerVersion":  fmt.Sprintf("v%d.%d.0", k8sVersionSegments[0], k8sVersionSegments[1]), // Autoscaler version matches the Kubernetes minor
	}); err != nil {
		return "", fmt.Errorf("could not generate a correct Autoscaler YAML block: %s", err)
	}
	resultYaml += fmt.Sprintf("%s\n", buf.String())

	buf.Reset()

	return resultYaml, nil
}
