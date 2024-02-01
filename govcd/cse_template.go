package govcd

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

// getCseKubernetesClusterCreationPayload gets the payload for the RDE that will trigger a Kubernetes cluster creation.
// It generates a valid YAML that is embedded inside the RDE JSON, then it is returned as an unmarshaled
// generic map, that allows to be sent to VCD as it is.
func getCseKubernetesClusterCreationPayload(goTemplateContents *cseClusterSettingsInternal) (map[string]interface{}, error) {
	capiYaml, err := generateCapiYaml(goTemplateContents)
	if err != nil {
		return nil, err
	}

	args := map[string]string{
		"Name":               goTemplateContents.Name,
		"Org":                goTemplateContents.OrganizationName,
		"VcdUrl":             goTemplateContents.VcdUrl,
		"Vdc":                goTemplateContents.VdcName,
		"Delete":             "false",
		"ForceDelete":        "false",
		"AutoRepairOnErrors": strconv.FormatBool(goTemplateContents.AutoRepairOnErrors),
		"ApiToken":           goTemplateContents.ApiToken,
		"CapiYaml":           capiYaml,
	}

	if goTemplateContents.DefaultStorageClass.StorageProfileName != "" {
		args["DefaultStorageClassStorageProfile"] = goTemplateContents.DefaultStorageClass.StorageProfileName
		args["DefaultStorageClassName"] = goTemplateContents.DefaultStorageClass.Name
		args["DefaultStorageClassUseDeleteReclaimPolicy"] = strconv.FormatBool(goTemplateContents.DefaultStorageClass.UseDeleteReclaimPolicy)
		args["DefaultStorageClassFileSystem"] = goTemplateContents.DefaultStorageClass.Filesystem
	}

	rdeTmpl, err := getCseTemplate(goTemplateContents.CseVersion, "rde")
	if err != nil {
		return nil, err
	}

	capvcdEmpty := template.Must(template.New(goTemplateContents.Name).Parse(rdeTmpl))
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

// generateNodePoolYaml generates YAML blocks corresponding to the Kubernetes node pools.
func generateNodePoolYaml(clusterDetails *cseClusterSettingsInternal) (string, error) {
	workerPoolTmpl, err := getCseTemplate(clusterDetails.CseVersion, "capiyaml_workerpool")
	if err != nil {
		return "", err
	}

	nodePoolEmptyTmpl := template.Must(template.New(clusterDetails.Name + "-worker-pool").Parse(workerPoolTmpl))
	resultYaml := ""
	buf := &bytes.Buffer{}

	// We can have many worker pools, we build a YAML object for each one of them.
	for _, workerPool := range clusterDetails.WorkerPools {

		// Check the correctness of the compute policies in the node pool block
		if workerPool.PlacementPolicyName != "" && workerPool.VGpuPolicyName != "" {
			return "", fmt.Errorf("the worker pool '%s' should have either a Placement Policy or a vGPU Policy, not both", workerPool.Name)
		}
		placementPolicy := workerPool.PlacementPolicyName
		if workerPool.VGpuPolicyName != "" {
			placementPolicy = workerPool.VGpuPolicyName // For convenience, we just use one of the variables as both cannot be set at same time
		}

		if err := nodePoolEmptyTmpl.Execute(buf, map[string]string{
			"ClusterName":             clusterDetails.Name,
			"NodePoolName":            workerPool.Name,
			"TargetNamespace":         clusterDetails.Name + "-ns",
			"Catalog":                 clusterDetails.CatalogName,
			"VAppTemplate":            clusterDetails.KubernetesTemplateOvaName,
			"NodePoolSizingPolicy":    workerPool.SizingPolicyName,
			"NodePoolPlacementPolicy": placementPolicy, // Can be either Placement or vGPU policy
			"NodePoolStorageProfile":  workerPool.StorageProfileName,
			"NodePoolDiskSize":        fmt.Sprintf("%dGi", workerPool.DiskSizeGi),
			"NodePoolEnableGpu":       strconv.FormatBool(workerPool.VGpuPolicyName != ""),
			"NodePoolMachineCount":    strconv.Itoa(workerPool.MachineCount),
			"KubernetesVersion":       clusterDetails.TkgVersionBundle.KubernetesVersion,
		}); err != nil {
			return "", fmt.Errorf("could not generate a correct Node Pool YAML: %s", err)
		}
		resultYaml += fmt.Sprintf("%s\n---\n", buf.String())
		buf.Reset()
	}
	return resultYaml, nil
}

// generateMemoryHealthCheckYaml generates a YAML block corresponding to the Kubernetes memory health check.
func generateMemoryHealthCheckYaml(mhcSettings *cseMachineHealthCheckInternal, cseVersion, clusterName string) (string, error) {
	if mhcSettings == nil {
		return "", nil
	}

	mhcTmpl, err := getCseTemplate(cseVersion, "capiyaml_mhc")
	if err != nil {
		return "", err
	}

	mhcEmptyTmpl := template.Must(template.New(clusterName + "-mhc").Parse(mhcTmpl))
	buf := &bytes.Buffer{}

	if err := mhcEmptyTmpl.Execute(buf, map[string]string{
		"ClusterName":                clusterName,
		"TargetNamespace":            clusterName + "-ns",
		"MaxUnhealthyNodePercentage": fmt.Sprintf("%.0f%%", mhcSettings.MaxUnhealthyNodesPercentage), // With the 'percentage' suffix
		"NodeStartupTimeout":         fmt.Sprintf("%ss", mhcSettings.NodeStartupTimeout),             // With the 'second' suffix
		"NodeUnknownTimeout":         fmt.Sprintf("%ss", mhcSettings.NodeUnknownTimeout),             // With the 'second' suffix
		"NodeNotReadyTimeout":        fmt.Sprintf("%ss", mhcSettings.NodeNotReadyTimeout),            // With the 'second' suffix
	}); err != nil {
		return "", fmt.Errorf("could not generate a correct Memory Health Check YAML: %s", err)
	}
	return fmt.Sprintf("%s\n---\n", buf.String()), nil

}

// generateCapiYaml generates the YAML string that is required during Kubernetes cluster creation, to be embedded
// in the CAPVCD cluster JSON payload. This function picks data from the Terraform schema and the createClusterDto to
// populate several Go templates and build a final YAML.
func generateCapiYaml(clusterDetails *cseClusterSettingsInternal) (string, error) {
	clusterTmpl, err := getCseTemplate(clusterDetails.CseVersion, "capiyaml_cluster")
	if err != nil {
		return "", err
	}

	// This YAML snippet contains special strings, such as "%,", that render wrong using the Go template engine
	sanitizedTemplate := strings.NewReplacer("%", "%%").Replace(clusterTmpl)
	capiYamlEmpty := template.Must(template.New(clusterDetails.Name + "-cluster").Parse(sanitizedTemplate))

	nodePoolYaml, err := generateNodePoolYaml(clusterDetails)
	if err != nil {
		return "", err
	}

	memoryHealthCheckYaml, err := generateMemoryHealthCheckYaml(clusterDetails.MachineHealthCheck, clusterDetails.CseVersion, clusterDetails.Name)
	if err != nil {
		return "", err
	}

	args := map[string]string{
		"ClusterName":                 clusterDetails.Name,
		"TargetNamespace":             clusterDetails.Name + "-ns",
		"TkrVersion":                  clusterDetails.TkgVersionBundle.TkrVersion,
		"TkgVersion":                  clusterDetails.TkgVersionBundle.TkgVersion,
		"UsernameB64":                 base64.StdEncoding.EncodeToString([]byte(clusterDetails.Owner)),
		"ApiTokenB64":                 base64.StdEncoding.EncodeToString([]byte(clusterDetails.ApiToken)),
		"PodCidr":                     clusterDetails.PodCidr,
		"ServiceCidr":                 clusterDetails.ServiceCidr,
		"VcdSite":                     clusterDetails.VcdUrl,
		"Org":                         clusterDetails.OrganizationName,
		"OrgVdc":                      clusterDetails.VdcName,
		"OrgVdcNetwork":               clusterDetails.NetworkName,
		"Catalog":                     clusterDetails.CatalogName,
		"VAppTemplate":                clusterDetails.KubernetesTemplateOvaName,
		"ControlPlaneSizingPolicy":    clusterDetails.ControlPlane.SizingPolicyName,
		"ControlPlanePlacementPolicy": clusterDetails.ControlPlane.PlacementPolicyName,
		"ControlPlaneStorageProfile":  clusterDetails.ControlPlane.StorageProfileName,
		"ControlPlaneDiskSize":        fmt.Sprintf("%dGi", clusterDetails.ControlPlane.DiskSizeGi),
		"ControlPlaneMachineCount":    strconv.Itoa(clusterDetails.ControlPlane.MachineCount),
		"ControlPlaneEndpoint":        clusterDetails.ControlPlane.Ip,
		"DnsVersion":                  clusterDetails.TkgVersionBundle.CoreDnsVersion,
		"EtcdVersion":                 clusterDetails.TkgVersionBundle.EtcdVersion,
		"ContainerRegistryUrl":        clusterDetails.ContainerRegistryUrl,
		"KubernetesVersion":           clusterDetails.TkgVersionBundle.KubernetesVersion,
		"SshPublicKey":                clusterDetails.SshPublicKey,
		"VirtualIpSubnet":             clusterDetails.VirtualIpSubnet,
	}

	buf := &bytes.Buffer{}
	if err := capiYamlEmpty.Execute(buf, args); err != nil {
		return "", fmt.Errorf("could not generate a correct CAPI YAML: %s", err)
	}
	// The final "pretty" YAML. To embed it in the final payload it must be marshaled into a one-line JSON string
	prettyYaml := fmt.Sprintf("%s\n%s\n%s", memoryHealthCheckYaml, nodePoolYaml, buf.String())

	// We don't use a standard json.Marshal() as the YAML contains special
	// characters that are not encoded properly, such as '<'.
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
