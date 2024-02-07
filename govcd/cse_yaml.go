package govcd

import (
	"fmt"
	semver "github.com/hashicorp/go-version"
	"sigs.k8s.io/yaml"
	"strings"
)

// updateCapiYaml takes a YAML and modifies its Kubernetes Template OVA, its Control plane, its Worker pools
// and its Node Health Check capabilities, by using the new values provided as input.
// If some of the values of the input is not provided, it doesn't change them.
// If none of the values is provided, it just returns the same untouched YAML.
func (cluster *CseKubernetesCluster) updateCapiYaml(input CseClusterUpdateInput) (string, error) {
	if cluster == nil {
		return cluster.capvcdType.Spec.CapiYaml, fmt.Errorf("receiver cluster is nil")
	}
	if input.ControlPlane == nil && input.WorkerPools == nil && input.NodeHealthCheck == nil && input.KubernetesTemplateOvaId == nil {
		return cluster.capvcdType.Spec.CapiYaml, nil
	}

	// The YAML contains multiple documents, so we cannot use a simple yaml.Unmarshal() as this one just gets the first
	// document it finds.
	yamlDocs, err := unmarshalMultipleYamlDocuments(cluster.capvcdType.Spec.CapiYaml)
	if err != nil {
		return cluster.capvcdType.Spec.CapiYaml, fmt.Errorf("error unmarshaling YAML: %s", err)
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
		err = cseUpdateKubernetesTemplateInYaml(yamlDocs, vAppTemplate.VAppTemplate.Name)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	if input.ControlPlane != nil {
		err := cseUpdateControlPlaneInYaml(yamlDocs, *input.ControlPlane)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	if input.WorkerPools != nil {
		err := cseUpdateWorkerPoolsInYaml(yamlDocs, *input.WorkerPools)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	if input.NewWorkerPools != nil {
		yamlDocs, err = cseAddWorkerPoolsInYaml(yamlDocs, *cluster, *input.NewWorkerPools)
		if err != nil {
			return cluster.capvcdType.Spec.CapiYaml, err
		}
	}

	if input.NodeHealthCheck != nil {
		vcdKeConfig, err := getVcdKeConfig(cluster.client, input.vcdKeConfigVersion, *input.NodeHealthCheck)
		if err != nil {
			return "", err
		}
		yamlDocs, err = cseUpdateNodeHealthCheckInYaml(yamlDocs, input.clusterName, input.cseVersion, vcdKeConfig)
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
func cseUpdateKubernetesTemplateInYaml(yamlDocuments []map[string]interface{}, kubernetesTemplateOvaName string) error {
	tkgBundle, err := getTkgVersionBundleFromVAppTemplateName(kubernetesTemplateOvaName)
	if err != nil {
		return err
	}
	for _, d := range yamlDocuments {
		switch d["kind"] {
		case "VCDMachineTemplate":
			_, err := traverseMapAndGet[string](d, "spec.template.spec.template")
			if err != nil {
				return fmt.Errorf("incorrect YAML: %s", err)
			}
			d["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["template"] = kubernetesTemplateOvaName
		case "MachineDeployment":
			_, err := traverseMapAndGet[string](d, "spec.template.spec.version")
			if err != nil {
				return fmt.Errorf("incorrect YAML: %s", err)
			}
			d["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["version"] = tkgBundle.KubernetesVersion
		case "Cluster":
			_, err := traverseMapAndGet[string](d, "metadata.annotations.TKGVERSION")
			if err != nil {
				return fmt.Errorf("incorrect YAML: %s", err)
			}
			d["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["TKGVERSION"] = tkgBundle.TkgVersion
			_, err = traverseMapAndGet[string](d, "metadata.labels.tanzuKubernetesRelease")
			if err != nil {
				return fmt.Errorf("incorrect YAML: %s", err)
			}
			d["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["tanzuKubernetesRelease"] = tkgBundle.TkrVersion
		case "KubeadmControlPlane":
			_, err := traverseMapAndGet[string](d, "spec.version")
			if err != nil {
				return fmt.Errorf("incorrect YAML: %s", err)
			}
			d["spec"].(map[string]interface{})["version"] = tkgBundle.KubernetesVersion
			_, err = traverseMapAndGet[string](d, "spec.kubeadmConfigSpec.clusterConfiguration.dns.imageTag")
			if err != nil {
				return fmt.Errorf("incorrect YAML: %s", err)
			}
			d["spec"].(map[string]interface{})["kubeadmConfigSpec"].(map[string]interface{})["clusterConfiguration"].(map[string]interface{})["dns"].(map[string]interface{})["imageTag"] = tkgBundle.CoreDnsVersion
			_, err = traverseMapAndGet[string](d, "spec.kubeadmConfigSpec.clusterConfiguration.etcd.local.imageTag")
			if err != nil {
				return fmt.Errorf("incorrect YAML: %s", err)
			}
			d["spec"].(map[string]interface{})["kubeadmConfigSpec"].(map[string]interface{})["clusterConfiguration"].(map[string]interface{})["etcd"].(map[string]interface{})["local"].(map[string]interface{})["imageTag"] = tkgBundle.EtcdVersion
		}
	}
	return nil
}

// cseUpdateControlPlaneInYaml modifies the given Kubernetes cluster YAML contents by changing the Control Plane with the input parameters.
func cseUpdateControlPlaneInYaml(yamlDocuments []map[string]interface{}, input CseControlPlaneUpdateInput) error {
	if input.MachineCount < 0 {
		return fmt.Errorf("incorrect machine count for Control Plane: %d. Should be at least 0", input.MachineCount)
	}

	updated := false
	for _, d := range yamlDocuments {
		if d["kind"] != "KubeadmControlPlane" {
			continue
		}
		_, err := traverseMapAndGet[float64](d, "spec.replicas")
		if err != nil {
			return fmt.Errorf("incorrect YAML: %s", err)
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

		workerPoolName, err := traverseMapAndGet[string](d, "metadata.name")
		if err != nil {
			return fmt.Errorf("incorrect YAML: %s", err)
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

		if workerPools[workerPoolToUpdate].MachineCount < 0 {
			return fmt.Errorf("incorrect machine count for worker pool %s: %d. Should be at least 0", workerPoolToUpdate, workerPools[workerPoolToUpdate].MachineCount)
		}

		_, err = traverseMapAndGet[float64](d, "spec.replicas")
		if err != nil {
			return fmt.Errorf("incorrect YAML: %s", err)
		}

		d["spec"].(map[string]interface{})["replicas"] = float64(workerPools[workerPoolToUpdate].MachineCount) // As it was originally unmarshalled as a float64
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
	internalSettings := cseClusterSettingsInternal{WorkerPools: make([]cseWorkerPoolSettingsInternal, len(newWorkerPools))}
	for i, workerPool := range newWorkerPools {
		internalSettings.WorkerPools[i] = cseWorkerPoolSettingsInternal{
			Name:                workerPool.Name,
			MachineCount:        workerPool.MachineCount,
			DiskSizeGi:          workerPool.DiskSizeGi,
			SizingPolicyName:    workerPool.SizingPolicyId,
			PlacementPolicyName: workerPool.PlacementPolicyId,
			VGpuPolicyName:      workerPool.VGpuPolicyId,
			StorageProfileName:  workerPool.StorageProfileId,
		}
	}
	internalSettings.Name = cluster.Name
	internalSettings.CseVersion = cluster.CseVersion
	nodePoolsYaml, err := internalSettings.generateNodePoolYaml()
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
func cseUpdateNodeHealthCheckInYaml(yamlDocuments []map[string]interface{}, clusterName string, cseVersion semver.Version, vcdKeConfig *vcdKeConfig) ([]map[string]interface{}, error) {
	mhcPosition := -1
	result := make([]map[string]interface{}, len(yamlDocuments))
	for i, d := range yamlDocuments {
		if d["kind"] == "MachineHealthCheck" {
			mhcPosition = i
		}
		result[i] = d
	}

	if mhcPosition < 0 {
		// There is no MachineHealthCheck block
		if vcdKeConfig == nil {
			// We don't want it neither, so nothing to do
			return result, nil
		}

		// We need to add the block to the slice of YAML documents
		settings := &cseClusterSettingsInternal{CseVersion: cseVersion, Name: clusterName, VcdKeConfig: *vcdKeConfig}
		mhcYaml, err := settings.generateMemoryHealthCheckYaml()
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
		if vcdKeConfig != nil {
			// We want it, but it is already there, so nothing to do
			// TODO: What happens in UI if the VCDKEConfig MHC values are changed, does it get reflected in the cluster?
			//       If that's the case, we might need to update this value always
			return result, nil
		}

		// We don't want Machine Health Checks, we delete the YAML document
		result[mhcPosition] = result[len(result)-1] // We override the MachineHealthCheck block with the last document
		result = result[:len(result)-1]             // We remove the last document (now duplicated)
	}
	return result, nil
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
// unmarshals all of them into a slice of generic maps with the corresponding content.
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
// "keyA.keyB.keyC.keyD", doing something similar to, visually speaking, map["keyA"]["keyB"]["keyC"]["keyD"], or in other words,
// it goes inside every inner map iteratively, until the given path is finished.
// The final value, "keyD" in the same example, should be of any type T.
func traverseMapAndGet[T any](input interface{}, path string) (T, error) {
	var nothing T
	if input == nil {
		return nothing, fmt.Errorf("the input is nil")
	}
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nothing, fmt.Errorf("the input is a %T, not a map[string]interface{}", input)
	}
	if len(inputMap) == 0 {
		return nothing, fmt.Errorf("the map is empty")
	}
	pathUnits := strings.Split(path, ".")
	completed := false
	i := 0
	var result interface{}
	for !completed {
		subPath := pathUnits[i]
		traversed, ok := inputMap[subPath]
		if !ok {
			return nothing, fmt.Errorf("key '%s' does not exist in input map", subPath)
		}
		if i < len(pathUnits)-1 {
			traversedMap, ok := traversed.(map[string]interface{})
			if !ok {
				return nothing, fmt.Errorf("key '%s' is a %T, not a map[string]interface{}, but there are still %d paths to explore", subPath, traversed, len(pathUnits)-(i+1))
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
		return nothing, fmt.Errorf("could not convert obtained type %T to requested %T", result, nothing)
	}
	return resultTyped, nil
}
