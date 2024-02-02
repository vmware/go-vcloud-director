package govcd

import (
	"fmt"
	"sigs.k8s.io/yaml"
	"strings"
)

// traverseMapAndGet traverses the input interface{}, which should be a map of maps, by following the path specified as
// "keyA.keyB.keyC.keyD", doing something similar to, visually speaking, map["keyA"]["keyB"]["keyC"]["keyD"], or in other words,
// it goes inside every inner map, which are inside the initial map, until the given path is finished.
// The final value, "keyD" in the same example, should be of type ResultType, which is a generic type requested during the call
// to this function.
func traverseMapAndGet[ResultType any](input interface{}, path string) (ResultType, error) {
	var nothing ResultType
	if input == nil {
		return nothing, fmt.Errorf("the input is nil")
	}
	inputMap, ok := input.(map[string]any)
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
			traversedMap, ok := traversed.(map[string]any)
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
	resultTyped, ok := result.(ResultType)
	if !ok {
		return nothing, fmt.Errorf("could not convert obtained type %T to requested %T", result, nothing)
	}
	return resultTyped, nil
}

// cseUpdateKubernetesTemplateInYaml updates the Kubernetes template OVA used by all the VCDMachineTemplate blocks
func cseUpdateKubernetesTemplateInYaml(yamlDocuments []map[string]any, kubernetesTemplateOvaName string) error {
	tkgBundle, err := getTkgVersionBundleFromVAppTemplateName(kubernetesTemplateOvaName)
	if err != nil {
		return err
	}
	for _, d := range yamlDocuments {
		switch d["kind"] {
		case "VCDMachineTemplate":
			_, err := traverseMapAndGet[string](d, "spec.template.spec.template")
			if err != nil {
				return fmt.Errorf("incorrect CAPI YAML: %s", err)
			}
			d["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["template"] = kubernetesTemplateOvaName
		case "MachineDeployment":
			_, err := traverseMapAndGet[string](d, "spec.template.spec.version")
			if err != nil {
				return fmt.Errorf("incorrect CAPI YAML: %s", err)
			}
			d["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["version"] = tkgBundle.KubernetesVersion
		case "Cluster":
			_, err := traverseMapAndGet[string](d, "metadata.annotations.TKGVERSION")
			if err != nil {
				return fmt.Errorf("incorrect CAPI YAML: %s", err)
			}
			d["metadata"].(map[string]any)["annotations"].(map[string]any)["TKGVERSION"] = tkgBundle.TkgVersion

			_, err = traverseMapAndGet[string](d, "metadata.labels.tanzuKubernetesRelease")
			if err != nil {
				return fmt.Errorf("incorrect CAPI YAML: %s", err)
			}
			d["metadata"].(map[string]any)["labels"].(map[string]any)["tanzuKubernetesRelease"] = tkgBundle.TkrVersion
		case "KubeadmControlPlane":
			_, err := traverseMapAndGet[string](d, "spec.version")
			if err != nil {
				return fmt.Errorf("incorrect CAPI YAML: %s", err)
			}
			d["spec"].(map[string]any)["version"] = tkgBundle.KubernetesVersion

			_, err = traverseMapAndGet[string](d, "spec.kubeadmConfigSpec.clusterConfiguration.dns.imageTag")
			if err != nil {
				return fmt.Errorf("incorrect CAPI YAML: %s", err)
			}
			d["spec"].(map[string]any)["kubeadmConfigSpec"].(map[string]any)["clusterConfiguration"].(map[string]any)["dns"].(map[string]any)["imageTag"] = tkgBundle.CoreDnsVersion

			_, err = traverseMapAndGet[string](d, "spec.kubeadmConfigSpec.clusterConfiguration.etcd.local.imageTag")
			if err != nil {
				return fmt.Errorf("incorrect CAPI YAML: %s", err)
			}
			d["spec"].(map[string]any)["kubeadmConfigSpec"].(map[string]any)["clusterConfiguration"].(map[string]any)["etcd"].(map[string]any)["local"].(map[string]any)["imageTag"] = tkgBundle.EtcdVersion
		}
	}
	return nil
}

func cseUpdateControlPlaneInYaml(yamlDocuments []map[string]any, input CseControlPlaneUpdateInput) error {
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
			return fmt.Errorf("incorrect CAPI YAML: %s", err)
		}
		d["spec"].(map[string]any)["replicas"] = float64(input.MachineCount) // As it was originally unmarshalled as a float64
		updated = true
	}
	if !updated {
		return fmt.Errorf("could not update the KubeadmControlPlane block in the CAPI YAML")
	}
	return nil
}

func cseUpdateWorkerPoolsInYaml(yamlDocuments []map[string]any, workerPools map[string]CseWorkerPoolUpdateInput) error {
	updated := 0
	for _, d := range yamlDocuments {
		if d["kind"] != "MachineDeployment" {
			continue
		}

		workerPoolName, err := traverseMapAndGet[string](d, "metadata.name")
		if err != nil {
			return fmt.Errorf("incorrect CAPI YAML: %s", err)
		}

		workerPoolToUpdate := ""
		for wpName := range workerPools {
			if wpName == workerPoolName {
				workerPoolToUpdate = wpName
			}
		}
		// This worker pool is not going to be updated, continue searching for another one
		if workerPoolToUpdate == "" {
			continue
		}

		if workerPools[workerPoolToUpdate].MachineCount < 0 {
			return fmt.Errorf("incorrect machine count for worker pool %s: %d. Should be at least 0", workerPoolToUpdate, workerPools[workerPoolToUpdate].MachineCount)
		}

		_, err = traverseMapAndGet[float64](d, "spec.replicas")
		if err != nil {
			return fmt.Errorf("incorrect CAPI YAML: %s", err)
		}

		d["spec"].(map[string]any)["replicas"] = float64(workerPools[workerPoolToUpdate].MachineCount) // As it was originally unmarshalled as a float64
		updated++
	}
	if updated != len(workerPools) {
		return fmt.Errorf("could not update all the Node pools. Updated %d, expected %d", updated, len(workerPools))
	}
	return nil
}

func cseAddWorkerPoolsInYaml(docs []map[string]any, inputs []CseWorkerPoolSettings) ([]map[string]any, error) {
	return nil, nil
}

func cseUpdateNodeHealthCheckInYaml(yamlDocuments []map[string]any, clusterName string, vcdKeConfig *vcdKeConfig) ([]map[string]any, error) {
	mhcPosition := -1
	result := make([]map[string]any, len(yamlDocuments))
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
		mhcYaml, err := generateMemoryHealthCheckYaml(*vcdKeConfig, "4.2", clusterName)
		if err != nil {
			return nil, err
		}
		var mhc map[string]any
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

// cseUpdateCapiYaml takes a CAPI YAML and modifies its Kubernetes template, its Control plane, its Worker pools
// and its Node Health Check capabilities, by using the new values provided as input.
// If some of the values of the input is not provided, it doesn't change them.
// If none of the values is provided, it just returns the same untouched YAML.
func cseUpdateCapiYaml(client *Client, capiYaml string, input CseClusterUpdateInput) (string, error) {
	if input.ControlPlane == nil && input.WorkerPools == nil && input.NodeHealthCheck == nil && input.KubernetesTemplateOvaId == nil {
		return capiYaml, nil
	}

	// The CAPI YAML contains multiple documents, so we cannot use a simple yaml.Unmarshal() as this one just gets the first
	// document it finds.
	yamlDocs, err := unmarshalMultipleYamlDocuments(capiYaml)
	if err != nil {
		return capiYaml, fmt.Errorf("error unmarshaling CAPI YAML: %s", err)
	}

	// As a side note, we can't optimize this one with "if <current value> equals <new value> do nothing" because
	// in order to retrieve the current value we would need to explore the YAML anyway, which is what we also need to do to update it.
	// So in this special case this "optimization" would optimize nothing. The same happens with other YAML values.
	if input.KubernetesTemplateOvaId != nil {
		vAppTemplate, err := getVAppTemplateById(client, *input.KubernetesTemplateOvaId)
		if err != nil {
			return capiYaml, fmt.Errorf("could not retrieve the Kubernetes OVA with ID '%s': %s", *input.KubernetesTemplateOvaId, err)
		}

		err = cseUpdateKubernetesTemplateInYaml(yamlDocs, vAppTemplate.VAppTemplate.Name)
		if err != nil {
			return capiYaml, err
		}
	}

	if input.ControlPlane != nil {
		err := cseUpdateControlPlaneInYaml(yamlDocs, *input.ControlPlane)
		if err != nil {
			return capiYaml, err
		}
	}

	if input.WorkerPools != nil {
		err := cseUpdateWorkerPoolsInYaml(yamlDocs, *input.WorkerPools)
		if err != nil {
			return capiYaml, err
		}
	}

	if input.NewWorkerPools != nil {
		yamlDocs, err = cseAddWorkerPoolsInYaml(yamlDocs, *input.NewWorkerPools)
		if err != nil {
			return capiYaml, err
		}
	}

	if input.NodeHealthCheck != nil {
		vcdKeConfig, err := getVcdKeConfig(client, input.vcdKeConfigVersion, *input.NodeHealthCheck)
		if err != nil {
			return "", err
		}
		yamlDocs, err = cseUpdateNodeHealthCheckInYaml(yamlDocs, input.clusterName, &vcdKeConfig)
		if err != nil {
			return "", err
		}
	}

	return marshalMultipleYamlDocuments(yamlDocs)
	/*
		if d.HasChange("control_plane.0.machine_count") {
			for _, yamlDoc := range yamlDocs {
				if yamlDoc["kind"] == "KubeadmControlPlane" {
					yamlDoc["spec"].(map[string]interface{})["replicas"] = d.Get("control_plane.0.machine_count")
				}
			}
		}
		// The node pools can only be created and resized
		var newNodePools []map[string]interface{}
		if d.HasChange("node_pool") {
			for _, nodePoolRaw := range d.Get("node_pool").(*schema.Set).List() {
				nodePool := nodePoolRaw.(map[string]interface{})
				for _, yamlDoc := range yamlDocs {
					if yamlDoc["kind"] == "MachineDeployment" {
						if yamlDoc["metadata"].(map[string]interface{})["name"] == nodePool["name"].(string) {
							yamlDoc["spec"].(map[string]interface{})["replicas"] = nodePool["machine_count"].(int)
						} else {
							// TODO: Create node pool
							newNodePools = append(newNodePools, map[string]interface{}{})
						}
					}
				}
			}
		}
		if len(newNodePools) > 0 {
			yamlDocs = append(yamlDocs, newNodePools...)
		}

		if d.HasChange("node_health_check") {
			oldNhc, newNhc := d.GetChange("node_health_check")
			if oldNhc.(bool) && !newNhc.(bool) {
				toDelete := 0
				for i, yamlDoc := range yamlDocs {
					if yamlDoc["kind"] == "MachineHealthCheck" {
						toDelete = i
					}
				}
				yamlDocs[toDelete] = yamlDocs[len(yamlDocs)-1] // We delete the MachineHealthCheck block by putting the last doc in its place
				yamlDocs = yamlDocs[:len(yamlDocs)-1]          // Then we remove the last doc
			} else {
				// Add the YAML block
				vcdKeConfig, err := getVcdKeConfiguration(d, vcdClient)
				if err != nil {
					return diag.FromErr(err)
				}
				rawYaml, err := generateMemoryHealthCheckYaml(d, vcdClient, *vcdKeConfig, d.Get("name").(string))
				if err != nil {
					return diag.FromErr(err)
				}
				yamlBlock := map[string]interface{}{}
				err = yaml.Unmarshal([]byte(rawYaml), &yamlBlock)
				if err != nil {
					return diag.Errorf("error updating Memory Health Check: %s", err)
				}
				yamlDocs = append(yamlDocs, yamlBlock)
			}
			util.Logger.Printf("not done but make static complains :)")
		}

		updatedYaml, err := yaml.Marshal(yamlDocs)
		if err != nil {
			return diag.Errorf("error updating cluster: %s", err)
		}

		// This must be done with retries due to the possible clash on ETags
		_, err = runWithRetry(
			"update cluster",
			"could not update cluster",
			1*time.Minute,
			nil,
			func() (any, error) {
				rde, err := vcdClient.GetRdeById(d.Id())
				if err != nil {
					return nil, fmt.Errorf("could not update Kubernetes cluster with ID '%s': %s", d.Id(), err)
				}

				rde.DefinedEntity.Entity["spec"].(map[string]interface{})["capiYaml"] = updatedYaml
				rde.DefinedEntity.Entity["spec"].(map[string]interface{})["vcdKe"].(map[string]interface{})["autoRepairOnErrors"] = d.Get("auto_repair_on_errors").(bool)

				//	err = rde.Update(*rde.DefinedEntity)
				util.Logger.Printf("ADAM: PERFORM UPDATE: %v", rde.DefinedEntity.Entity)
				if err != nil {
					return nil, err
				}
				return nil, nil
			},
		)
		if err != nil {
			return diag.FromErr(err)
		}

		state, err = waitUntilClusterIsProvisioned(vcdClient, d, rde.DefinedEntity.ID)
		if err != nil {
			return diag.Errorf("Kubernetes cluster update failed: %s", err)
		}
		if state != "provisioned" {
			return diag.Errorf("Kubernetes cluster update failed, cluster is not in 'provisioned' state, but '%s'", state)
		}*/
}

// marshalMultipleYamlDocuments takes a slice of maps representing multiple YAML documents (one per item in the slice) and
// marshals all of them into a single string with the corresponding separators "---".
func marshalMultipleYamlDocuments(yamlDocuments []map[string]any) (string, error) {
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
func unmarshalMultipleYamlDocuments(yamlDocuments string) ([]map[string]any, error) {
	if len(strings.TrimSpace(yamlDocuments)) == 0 {
		return []map[string]any{}, nil
	}

	splitYamlDocs := strings.Split(yamlDocuments, "---\n")
	result := make([]map[string]any, len(splitYamlDocs))
	for i, yamlDoc := range splitYamlDocs {
		err := yaml.Unmarshal([]byte(yamlDoc), &result[i])
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal document %s: %s", yamlDoc, err)
		}
	}

	return result, nil
}
