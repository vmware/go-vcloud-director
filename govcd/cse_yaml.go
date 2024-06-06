package govcd

import (
	"fmt"
	semver "github.com/hashicorp/go-version"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

// updateCapiYaml takes a YAML and modifies its Kubernetes Template OVA, its Control plane, its Worker pools
// and its Node Health Check capabilities, by using the new values provided as input.
// If some of the values of the input is not provided, it doesn't change them.
// If none of the values is provided, it just returns the same untouched YAML.
func (cluster *CseKubernetesCluster) updateCapiYaml(input CseClusterUpdateInput) (string, error) {
	if cluster == nil || cluster.capvcdType == nil {
		return "", fmt.Errorf("receiver cluster is nil")
	}

	if input.ControlPlane == nil && input.WorkerPools == nil && input.NodeHealthCheck == nil && input.KubernetesTemplateOvaId == nil && input.NewWorkerPools == nil {
		return cluster.capvcdType.Spec.CapiYaml, nil
	}

	// The YAML contains multiple documents, so we cannot use a simple yaml.Unmarshal() as this one just gets the first
	// document it finds.
	yamlDocs, err := unmarshalMultipleYamlDocuments(cluster.capvcdType.Spec.CapiYaml)
	if err != nil {
		return cluster.capvcdType.Spec.CapiYaml, fmt.Errorf("error unmarshalling YAML: %s", err)
	}

	if input.ControlPlane != nil {
		err := cseUpdateControlPlaneInYaml(yamlDocs, *input.ControlPlane)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	// Modify or add the autoscaler capabilities
	yamlDocs, err = cseUpdateAutoscalerInYaml(yamlDocs, cluster.Name, cluster.CseVersion, cluster.KubernetesVersion, input.WorkerPools, input.NewWorkerPools)
	if err != nil {
		return cluster.capvcdType.Spec.CapiYaml, err
	}

	if input.WorkerPools != nil {
		err := cseUpdateWorkerPoolsInYaml(yamlDocs, *input.WorkerPools)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	// Order matters. We need to add the new pools before updating the Kubernetes template.
	if input.NewWorkerPools != nil {
		// Worker pool names must be unique
		for _, existingPool := range cluster.WorkerPools {
			for _, newPool := range *input.NewWorkerPools {
				if newPool.Name == existingPool.Name {
					return cluster.capvcdType.Spec.CapiYaml, fmt.Errorf("there is an existing Worker Pool with name '%s'", existingPool.Name)
				}
			}
		}

		yamlDocs, err = cseAddWorkerPoolsInYaml(yamlDocs, *cluster, *input.NewWorkerPools)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	// As a side note, we can't optimize this one with "if <current value> equals <new value> do nothing" because
	// in order to retrieve the current value we would need to explore the YAML anyway, which is what we also need to do to update it.
	// Also, even if we did it, the current value obtained from YAML would be a Name, but the new value is an ID, so we would need to query VCD anyway
	// as well.
	// So in this special case this "optimization" would optimize nothing. The same happens with other YAML values.
	if input.KubernetesTemplateOvaId != nil {
		vAppTemplate, err := getVAppTemplateById(cluster.client, *input.KubernetesTemplateOvaId)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, fmt.Errorf("could not retrieve the Kubernetes Template OVA with ID '%s': %s", *input.KubernetesTemplateOvaId, err)
		}
		// Check the versions of the selected OVA before upgrading
		versions, err := getTkgVersionBundleFromVAppTemplate(vAppTemplate.VAppTemplate)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, fmt.Errorf("could not retrieve the TKG versions of OVA '%s': %s", *input.KubernetesTemplateOvaId, err)
		}
		if versions.compareTkgVersion(cluster.capvcdType.Status.Capvcd.Upgrade.Current.TkgVersion) < 0 || !versions.kubernetesVersionIsUpgradeableFrom(cluster.capvcdType.Status.Capvcd.Upgrade.Current.KubernetesVersion) {
			return cluster.capvcdType.Spec.CapiYaml, fmt.Errorf("cannot perform an OVA change as the new one '%s' has an older TKG/Kubernetes version (%s/%s)", vAppTemplate.VAppTemplate.Name, versions.TkgVersion, versions.KubernetesVersion)
		}
		err = cseUpdateKubernetesTemplateInYaml(yamlDocs, vAppTemplate.VAppTemplate)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	if input.NodeHealthCheck != nil {
		cseComponentsVersions, err := getCseComponentsVersions(cluster.CseVersion)
		if err != nil {
			return "", err
		}
		vcdKeConfig, err := getVcdKeConfig(cluster.client, cseComponentsVersions.VcdKeConfigRdeTypeVersion, *input.NodeHealthCheck)
		if err != nil {
			return "", err
		}
		yamlDocs, err = cseUpdateNodeHealthCheckInYaml(yamlDocs, cluster.Name, cluster.CseVersion, vcdKeConfig)
		if err != nil {
			return "", err
		}
	}

	return marshalMultipleYamlDocuments(yamlDocs)
}

// cseUpdateKubernetesTemplateInYaml modifies the given Kubernetes cluster YAML by modifying the Kubernetes Template OVA
// used by all the cluster elements.
// The caveat here is that not only VCDMachineTemplate needs to be changed with the new OVA name, but also
// other fields that reference the related Kubernetes version, TKG version and other derived information.
func cseUpdateKubernetesTemplateInYaml(yamlDocuments []map[string]interface{}, kubernetesTemplateOva *types.VAppTemplate) error {
	tkgBundle, err := getTkgVersionBundleFromVAppTemplate(kubernetesTemplateOva)
	if err != nil {
		return err
	}
	for _, d := range yamlDocuments {
		switch d["kind"] {
		case "VCDMachineTemplate":
			ok := traverseMapAndGet[string](d, "spec.template.spec.template", ".") != ""
			if !ok {
				return fmt.Errorf("the VCDMachineTemplate 'spec.template.spec.template' field is missing")
			}
			d["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["template"] = kubernetesTemplateOva.Name
		case "MachineDeployment":
			ok := traverseMapAndGet[string](d, "spec.template.spec.version", ".") != ""
			if !ok {
				return fmt.Errorf("the MachineDeployment 'spec.template.spec.version' field is missing")
			}
			d["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["version"] = tkgBundle.KubernetesVersion
		case "Cluster":
			ok := traverseMapAndGet[string](d, "metadata.annotations.TKGVERSION", ".") != ""
			if !ok {
				return fmt.Errorf("the Cluster 'metadata.annotations.TKGVERSION' field is missing")
			}
			d["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["TKGVERSION"] = tkgBundle.TkgVersion
			ok = traverseMapAndGet[string](d, "metadata.labels.tanzuKubernetesRelease", ".") != ""
			if !ok {
				return fmt.Errorf("the Cluster 'metadata.labels.tanzuKubernetesRelease' field is missing")
			}
			d["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["tanzuKubernetesRelease"] = tkgBundle.TkrVersion
		case "KubeadmControlPlane":
			ok := traverseMapAndGet[string](d, "spec.version", ".") != ""
			if !ok {
				return fmt.Errorf("the KubeadmControlPlane 'spec.version' field is missing")
			}
			d["spec"].(map[string]interface{})["version"] = tkgBundle.KubernetesVersion
			ok = traverseMapAndGet[string](d, "spec.kubeadmConfigSpec.clusterConfiguration.dns.imageTag", ".") != ""
			if !ok {
				return fmt.Errorf("the KubeadmControlPlane 'spec.kubeadmConfigSpec.clusterConfiguration.dns.imageTag' field is missing")
			}
			d["spec"].(map[string]interface{})["kubeadmConfigSpec"].(map[string]interface{})["clusterConfiguration"].(map[string]interface{})["dns"].(map[string]interface{})["imageTag"] = tkgBundle.CoreDnsVersion
			ok = traverseMapAndGet[string](d, "spec.kubeadmConfigSpec.clusterConfiguration.etcd.local.imageTag", ".") != ""
			if !ok {
				return fmt.Errorf("the KubeadmControlPlane 'spec.kubeadmConfigSpec.clusterConfiguration.etcd.local.imageTag' field is missing")
			}
			d["spec"].(map[string]interface{})["kubeadmConfigSpec"].(map[string]interface{})["clusterConfiguration"].(map[string]interface{})["etcd"].(map[string]interface{})["local"].(map[string]interface{})["imageTag"] = tkgBundle.EtcdVersion
		case "Deployment":
			// Update also the autoscaler version
			deploymentName := traverseMapAndGet[string](d, "metadata.name", ".")
			if deploymentName == "cluster-autoscaler" {
				k8sVersion, err := semver.NewVersion(tkgBundle.KubernetesVersion)
				if err != nil {
					return err
				}
				k8sVersionSegments := k8sVersion.Segments()
				d["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["image"] = fmt.Sprintf("k8s.gcr.io/autoscaling/cluster-autoscaler:v%d.%d.0", k8sVersionSegments[0], k8sVersionSegments[1])
			}
		}
	}
	return nil
}

// cseUpdateControlPlaneInYaml modifies the given Kubernetes cluster YAML contents by changing the Control Plane with the input parameters.
func cseUpdateControlPlaneInYaml(yamlDocuments []map[string]interface{}, input CseControlPlaneUpdateInput) error {
	if input.MachineCount < 1 || input.MachineCount%2 == 0 {
		return fmt.Errorf("incorrect machine count for Control Plane: %d. Should be at least 1 and an odd number", input.MachineCount)
	}

	updated := false
	for _, d := range yamlDocuments {
		if d["kind"] != "KubeadmControlPlane" {
			continue
		}
		d["spec"].(map[string]interface{})["replicas"] = float64(input.MachineCount) // As it was originally unmarshalled as a float64
		updated = true
	}
	if !updated {
		return fmt.Errorf("could not find the KubeadmControlPlane object in the YAML")
	}
	return nil
}

// cseUpdateControlPlaneInYaml modifies the given Kubernetes cluster YAML contents by changing
// the existing Worker Pools with the input parameters.
func cseUpdateWorkerPoolsInYaml(yamlDocuments []map[string]interface{}, workerPools map[string]CseWorkerPoolUpdateInput) error {
	updated := 0

	for _, d := range yamlDocuments {
		if d["kind"] != "MachineDeployment" {
			continue
		}

		workerPoolName := traverseMapAndGet[string](d, "metadata.name", ".")
		if workerPoolName == "" {
			return fmt.Errorf("the MachineDeployment 'metadata.name' field is empty")
		}

		workerPoolToUpdate := ""
		for wpName := range workerPools {
			if wpName == workerPoolName {
				workerPoolToUpdate = wpName
			}
		}
		// This worker pool must not be updated as it is not present in the input, continue searching for the ones we want
		if workerPoolToUpdate == "" {
			continue
		}

		if workerPools[workerPoolToUpdate].Autoscaler != nil {
			if workerPools[workerPoolToUpdate].Autoscaler.MinSize > workerPools[workerPoolToUpdate].Autoscaler.MaxSize {
				return fmt.Errorf("incorrect MinSize for worker pool %s: %d should be less than the maximum %d", workerPoolToUpdate, workerPools[workerPoolToUpdate].Autoscaler.MinSize, workerPools[workerPoolToUpdate].Autoscaler.MaxSize)
			}
			if d["metadata"].(map[string]interface{})["annotations"] == nil {
				d["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{}
			}
			d["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"] = strconv.Itoa(workerPools[workerPoolToUpdate].Autoscaler.MaxSize)
			d["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"] = strconv.Itoa(workerPools[workerPoolToUpdate].Autoscaler.MinSize)
			delete(d["spec"].(map[string]interface{}), "replicas") // This is required to avoid conflicts with Autoscaler
		} else {
			if workerPools[workerPoolToUpdate].MachineCount < 0 {
				return fmt.Errorf("incorrect machine count for worker pool %s: %d. Should be at least 0", workerPoolToUpdate, workerPools[workerPoolToUpdate].MachineCount)
			}
			d["spec"].(map[string]interface{})["replicas"] = float64(workerPools[workerPoolToUpdate].MachineCount) // As it was originally unmarshalled as a float64

			// Removes the autoscaler information, as we used static replicas
			if d["metadata"].(map[string]interface{})["annotations"] != nil {
				delete(d["metadata"].(map[string]interface{})["annotations"].(map[string]interface{}), "cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size")
				delete(d["metadata"].(map[string]interface{})["annotations"].(map[string]interface{}), "cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size")
			}
		}
		updated++
	}
	if updated != len(workerPools) {
		return fmt.Errorf("could not update all the Node pools. Updated %d, expected %d", updated, len(workerPools))
	}
	return nil
}

// cseAddWorkerPoolsInYaml modifies the given Kubernetes cluster YAML contents by adding new Worker Pools
// described by the input parameters.
// NOTE: This function doesn't modify the input, but returns a copy of the YAML with the added unmarshalled documents.
func cseAddWorkerPoolsInYaml(docs []map[string]interface{}, cluster CseKubernetesCluster, newWorkerPools []CseWorkerPoolSettings) ([]map[string]interface{}, error) {
	if len(newWorkerPools) == 0 {
		return docs, nil
	}

	var computePolicyIds []string
	var storageProfileIds []string
	for _, w := range newWorkerPools {
		computePolicyIds = append(computePolicyIds, w.SizingPolicyId, w.PlacementPolicyId, w.VGpuPolicyId)
		storageProfileIds = append(storageProfileIds, w.StorageProfileId)
	}

	idToNameCache, err := idToNames(cluster.client, computePolicyIds, storageProfileIds)
	if err != nil {
		return nil, err
	}

	internalSettings := cseClusterSettingsInternal{WorkerPools: make([]cseWorkerPoolSettingsInternal, len(newWorkerPools))}
	for i, workerPool := range newWorkerPools {
		internalSettings.WorkerPools[i] = cseWorkerPoolSettingsInternal{
			Name:                workerPool.Name,
			MachineCount:        workerPool.MachineCount,
			DiskSizeGi:          workerPool.DiskSizeGi,
			StorageProfileName:  idToNameCache[workerPool.StorageProfileId],
			SizingPolicyName:    idToNameCache[workerPool.SizingPolicyId],
			VGpuPolicyName:      idToNameCache[workerPool.VGpuPolicyId],
			PlacementPolicyName: idToNameCache[workerPool.PlacementPolicyId],
		}
		if workerPool.Autoscaler != nil {
			internalSettings.WorkerPools[i].Autoscaler = &CseWorkerPoolAutoscaler{
				MaxSize: workerPool.Autoscaler.MaxSize,
				MinSize: workerPool.Autoscaler.MinSize,
			}
		}
	}

	// Extra information needed to render the YAML. As all the worker pools share the same
	// Kubernetes OVA name, version and Catalog, we pick this info from any of the available ones.
	for _, doc := range docs {
		if internalSettings.CatalogName == "" && doc["kind"] == "VCDMachineTemplate" {
			internalSettings.CatalogName = traverseMapAndGet[string](doc, "spec.template.spec.catalog", ".")
		}
		if internalSettings.KubernetesTemplateOvaName == "" && doc["kind"] == "VCDMachineTemplate" {
			internalSettings.KubernetesTemplateOvaName = traverseMapAndGet[string](doc, "spec.template.spec.template", ".")
		}
		if internalSettings.TkgVersionBundle.KubernetesVersion == "" && doc["kind"] == "MachineDeployment" {
			internalSettings.TkgVersionBundle.KubernetesVersion = traverseMapAndGet[string](doc, "spec.template.spec.version", ".")
		}
		if internalSettings.CatalogName != "" && internalSettings.KubernetesTemplateOvaName != "" && internalSettings.TkgVersionBundle.KubernetesVersion != "" {
			break
		}
	}
	internalSettings.Name = cluster.Name
	internalSettings.CseVersion = cluster.CseVersion
	nodePoolsYaml, err := internalSettings.generateWorkerPoolsYaml()
	if err != nil {
		return nil, err
	}

	newWorkerPoolsYamlDocs, err := unmarshalMultipleYamlDocuments(nodePoolsYaml)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(docs))
	copy(result, docs)
	return append(result, newWorkerPoolsYamlDocs...), nil
}

// cseUpdateNodeHealthCheckInYaml updates the Kubernetes cluster described in the given YAML documents by adding or removing
// the MachineHealthCheck object.
// NOTE: This function doesn't modify the input, but returns a copy of the YAML with the modifications.
func cseUpdateNodeHealthCheckInYaml(yamlDocuments []map[string]interface{}, clusterName string, cseVersion semver.Version, vcdKeConfig vcdKeConfig) ([]map[string]interface{}, error) {
	mhcPosition := -1
	result := make([]map[string]interface{}, len(yamlDocuments))
	for i, d := range yamlDocuments {
		if d["kind"] == "MachineHealthCheck" {
			mhcPosition = i
		}
		result[i] = d
	}

	machineHealthCheckEnabled := vcdKeConfig.NodeUnknownTimeout != "" && vcdKeConfig.NodeStartupTimeout != "" && vcdKeConfig.NodeNotReadyTimeout != "" &&
		vcdKeConfig.MaxUnhealthyNodesPercentage != 0

	if mhcPosition < 0 {
		// There is no MachineHealthCheck block
		if !machineHealthCheckEnabled {
			// We don't want it neither, so nothing to do
			return result, nil
		}

		// We need to add the block to the slice of YAML documents
		settings := &cseClusterSettingsInternal{CseVersion: cseVersion, Name: clusterName, VcdKeConfig: vcdKeConfig}
		mhcYaml, err := settings.generateMachineHealthCheckYaml()
		if err != nil {
			return nil, err
		}
		var mhc map[string]interface{}
		err = yaml.Unmarshal([]byte(mhcYaml), &mhc)
		if err != nil {
			return nil, err
		}
		result = append(result, mhc)
	} else {
		// There is a MachineHealthCheck block
		if machineHealthCheckEnabled {
			// We want it, but it is already there, so nothing to do
			return result, nil
		}

		// We don't want Machine Health Checks, we delete the YAML document
		result[mhcPosition] = result[len(result)-1] // We override the MachineHealthCheck block with the last document
		result = result[:len(result)-1]             // We remove the last document (now duplicated)
	}
	return result, nil
}

// cseUpdateAutoscalerInYaml adds a new YAML document (Autoscaler) to the output if the input worker pools require it and it's not present.
// If it's present, modifies the YAML documents by scaling the Autoscaler replicas to 1.
// If none of the input worker pools requires autoscaling, the YAML documents are modified to reduce the Autoscaler replicas to 0.
func cseUpdateAutoscalerInYaml(yamlDocuments []map[string]interface{}, clusterName string, cseVersion, kubernetesVersion semver.Version,
	existingWorkerPools *map[string]CseWorkerPoolUpdateInput, newWorkerPools *[]CseWorkerPoolSettings) ([]map[string]interface{}, error) {
	autoscalerNeeded := false
	// We'll need the Autoscaler YAML document if at least one Worker Pool uses it
	if existingWorkerPools != nil {
		for _, wp := range *existingWorkerPools {
			if wp.Autoscaler != nil {
				autoscalerNeeded = true
				break
			}
		}
	}

	// We'll need the Autoscaler YAML document if at least one of the new Worker Pools uses it
	if !autoscalerNeeded && newWorkerPools != nil {
		for _, wp := range *newWorkerPools {
			if wp.Autoscaler != nil {
				autoscalerNeeded = true
				break
			}
		}
	}

	// Search for Autoscaler YAML document
	for _, d := range yamlDocuments {
		if d["kind"] != "Deployment" {
			continue
		}
		if traverseMapAndGet[string](d, "metadata.name", ".") != "cluster-autoscaler" {
			continue
		}
		if traverseMapAndGet[string](d, "metadata.namespace", ".") != "kube-system" {
			continue
		}
		// Reaching here means that an Autoscaler was found. We need to modify its configuration.
		if autoscalerNeeded {
			d["spec"].(map[string]interface{})["replicas"] = float64(1) // As it was originally unmarshalled as a float64
		} else {
			d["spec"].(map[string]interface{})["replicas"] = float64(0) // As it was originally unmarshalled as a float64
		}
		// We also keep the image up-to-date with the Kubernetes version
		k8sVersionSegments := kubernetesVersion.Segments()
		d["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})[0].(map[string]interface{})["image"] = fmt.Sprintf("k8s.gcr.io/autoscaling/cluster-autoscaler:v%d.%d.0", k8sVersionSegments[0], k8sVersionSegments[1])
		return yamlDocuments, nil
	}

	// This part is only reached if we didn't find any Autoscaler document, so we add it new if it's needed.
	if autoscalerNeeded {
		settings := &cseClusterSettingsInternal{Name: clusterName, CseVersion: cseVersion, TkgVersionBundle: tkgVersionBundle{KubernetesVersion: kubernetesVersion.String()}}
		autoscalerYaml, err := settings.generateAutoscalerYaml()
		if err != nil {
			return nil, err
		}
		autoscaler, err := unmarshalMultipleYamlDocuments(autoscalerYaml)
		if err != nil {
			return nil, err
		}
		return append(yamlDocuments, autoscaler...), nil
	}

	// Otherwise the documents are returned without change
	return yamlDocuments, nil
}

// marshalMultipleYamlDocuments takes a slice of maps representing multiple YAML documents (one per item in the slice) and
// marshals all of them into a single string with the corresponding separators "---".
func marshalMultipleYamlDocuments(yamlDocuments []map[string]interface{}) (string, error) {
	result := ""
	for i, yamlDoc := range yamlDocuments {
		updatedSingleDoc, err := yaml.Marshal(yamlDoc)
		if err != nil {
			return "", fmt.Errorf("error marshaling the updated CAPVCD YAML '%v': %s", yamlDoc, err)
		}
		result += fmt.Sprintf("%s\n", updatedSingleDoc)
		if i < len(yamlDocuments)-1 { // The last document doesn't need the YAML separator
			result += "---\n"
		}
	}
	return result, nil
}

// unmarshalMultipleYamlDocuments takes a multi-document YAML (multiple YAML documents are separated by "---") and
// unmarshalls all of them into a slice of generic maps with the corresponding content.
func unmarshalMultipleYamlDocuments(yamlDocuments string) ([]map[string]interface{}, error) {
	if len(strings.TrimSpace(yamlDocuments)) == 0 {
		return []map[string]interface{}{}, nil
	}

	splitYamlDocs := strings.Split(yamlDocuments, "---\n")
	result := make([]map[string]interface{}, len(splitYamlDocs))
	for i, yamlDoc := range splitYamlDocs {
		err := yaml.Unmarshal([]byte(yamlDoc), &result[i])
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal document %s: %s", yamlDoc, err)
		}
	}

	return result, nil
}

// traverseMapAndGet traverses the input interface{}, which should be a map of maps, by following the path specified as
// "keyA%keyB%keyC%keyD" (if keySeparator="%"), or "keyA.keyB.keyC.keyD" (if keySeparator="."), etc. doing something similar to,
// visually speaking, map["keyA"]["keyB"]["keyC"]["keyD"], or in other words, it goes inside every inner map iteratively,
// until the given path is finished.
// If the path doesn't lead to any value, or if the value is nil, or there is any other issue, returns the "zero" value of T.
func traverseMapAndGet[T any](input interface{}, path string, keySeparator string) T {
	var nothing T
	if input == nil {
		return nothing
	}
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nothing
	}
	if len(inputMap) == 0 {
		return nothing
	}
	pathUnits := strings.Split(path, keySeparator)
	completed := false
	i := 0
	var result interface{}
	for !completed {
		subPath := pathUnits[i]
		traversed, ok := inputMap[subPath]
		if !ok {
			return nothing
		}
		if i < len(pathUnits)-1 {
			traversedMap, ok := traversed.(map[string]interface{})
			if !ok {
				return nothing
			}
			inputMap = traversedMap
		} else {
			completed = true
			result = traversed
		}
		i++
	}
	resultTyped, ok := result.(T)
	if !ok {
		return nothing
	}
	return resultTyped
}
