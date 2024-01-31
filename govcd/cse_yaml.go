package govcd

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
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
	inputMap, ok := input.(map[any]any)
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
			traversedMap, ok := traversed.(map[any]any)
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
func cseUpdateKubernetesTemplateInYaml(yamlDocuments []map[any]any, kubernetesTemplateOvaName string) error {
	updated := false
	for _, d := range yamlDocuments {
		if d["kind"] != "VCDMachineTemplate" {
			continue
		}

		_, err := traverseMapAndGet[string](d, "spec.template.spec.template")
		if err != nil {
			return fmt.Errorf("incorrect CAPI YAML: %s", err)
		}
		d["spec"].(map[any]any)["template"].(map[any]any)["spec"].(map[any]any)["template"] = kubernetesTemplateOvaName
		updated = true
	}
	if !updated {
		return fmt.Errorf("could not find any template inside the VCDMachineTemplate blocks in the CAPI YAML")
	}
	return nil
}

func updateControlPlaneYaml(docs []map[any]any, input CseControlPlaneUpdateInput) error {
	return nil
}

func updateWorkerPoolsYaml(docs []map[any]any, m map[string]CseWorkerPoolUpdateInput) error {
	return nil
}

func updateNodeHealthCheckYaml(docs []map[any]any, b bool) error {
	return nil
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
		err := updateControlPlaneYaml(yamlDocs, *input.ControlPlane)
		if err != nil {
			return "", err
		}
	}

	if input.WorkerPools != nil {
		err := updateWorkerPoolsYaml(yamlDocs, *input.WorkerPools)
		if err != nil {
			return "", err
		}
	}

	if input.NodeHealthCheck != nil {
		err := updateNodeHealthCheckYaml(yamlDocs, *input.NodeHealthCheck)
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
func marshalMultipleYamlDocuments(yamlDocuments []map[any]any) (string, error) {
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
func unmarshalMultipleYamlDocuments(yamlDocuments string) ([]map[any]any, error) {
	if len(strings.TrimSpace(yamlDocuments)) == 0 {
		return []map[any]any{}, nil
	}

	dec := yaml.NewDecoder(bytes.NewReader([]byte(yamlDocuments)))
	documentCount := strings.Count(yamlDocuments, "---")
	if documentCount == 0 {
		// If it doesn't have any separator, we can assume it's just a single document.
		// Otherwise, it will fail afterward
		documentCount = 1
	}
	yamlDocs := make([]map[any]any, documentCount)
	i := 0
	for i < documentCount {
		err := dec.Decode(&yamlDocs[i])
		if err != nil {
			return nil, err
		}
		i++
	}
	return yamlDocs, nil
}
