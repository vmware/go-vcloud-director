//go:build unit || ALL

package govcd

import (
	semver "github.com/hashicorp/go-version"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Test_cseUpdateKubernetesTemplateInYaml tests the update process of the Kubernetes template OVA in a CAPI YAML.
func Test_cseUpdateKubernetesTemplateInYaml(t *testing.T) {
	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read CAPI YAML test file: %s", err)
	}

	yamlDocs, err := unmarshalMultipleYamlDocuments(string(capiYaml))
	if err != nil {
		t.Fatalf("could not unmarshal CAPI YAML test file: %s", err)
	}

	// We explore the YAML documents to get the OVA template name that will be updated
	// with the new one.
	oldOvaName := ""
	for _, document := range yamlDocs {
		if document["kind"] != "VCDMachineTemplate" {
			continue
		}

		oldOvaName, err = traverseMapAndGet[string](document, "spec.template.spec.template")
		if err != nil {
			t.Fatalf("expected to find spec.template.spec.template in %v but got an error: %s", document, err)
		}
		break
	}
	if oldOvaName == "" {
		t.Fatalf("the OVA that needs to be changed is empty")
	}
	oldTkgBundle, err := getTkgVersionBundleFromVAppTemplateName(oldOvaName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// We call the function to update the old OVA with the new one
	newOvaName := "ubuntu-2004-kube-v1.19.16+vmware.1-tkg.2-fba68db15591c15fcd5f26b512663a42"
	newTkgBundle, err := getTkgVersionBundleFromVAppTemplateName(newOvaName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = cseUpdateKubernetesTemplateInYaml(yamlDocs, newOvaName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	updatedYaml, err := marshalMultipleYamlDocuments(yamlDocs)
	if err != nil {
		t.Fatalf("error marshaling %v: %s", yamlDocs, err)
	}

	// No document should have the old OVA
	if !strings.Contains(updatedYaml, newOvaName) || strings.Contains(updatedYaml, oldOvaName) {
		t.Fatalf("failed updating the Kubernetes OVA template:\n%s", updatedYaml)
	}
	if !strings.Contains(updatedYaml, newTkgBundle.KubernetesVersion) || strings.Contains(updatedYaml, oldTkgBundle.KubernetesVersion) {
		t.Fatalf("failed updating the Kubernetes version:\n%s", updatedYaml)
	}
	if !strings.Contains(updatedYaml, newTkgBundle.TkrVersion) || strings.Contains(updatedYaml, oldTkgBundle.TkrVersion) {
		t.Fatalf("failed updating the Tanzu release version:\n%s", updatedYaml)
	}
	if !strings.Contains(updatedYaml, newTkgBundle.TkgVersion) || strings.Contains(updatedYaml, oldTkgBundle.TkgVersion) {
		t.Fatalf("failed updating the Tanzu grid version:\n%s", updatedYaml)
	}
	if !strings.Contains(updatedYaml, newTkgBundle.CoreDnsVersion) || strings.Contains(updatedYaml, oldTkgBundle.CoreDnsVersion) {
		t.Fatalf("failed updating the CoreDNS version:\n%s", updatedYaml)
	}
	if !strings.Contains(updatedYaml, newTkgBundle.EtcdVersion) || strings.Contains(updatedYaml, oldTkgBundle.EtcdVersion) {
		t.Fatalf("failed updating the Etcd version:\n%s", updatedYaml)
	}
}

// Test_cseUpdateWorkerPoolsInYaml tests the update process of the Worker pools in a CAPI YAML.
func Test_cseUpdateWorkerPoolsInYaml(t *testing.T) {
	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read CAPI YAML test file: %s", err)
	}

	yamlDocs, err := unmarshalMultipleYamlDocuments(string(capiYaml))
	if err != nil {
		t.Fatalf("could not unmarshal CAPI YAML test file: %s", err)
	}
	// We explore the YAML documents to get the OVA template name that will be updated
	// with the new one.
	oldNodePools := map[string]CseWorkerPoolUpdateInput{}
	for _, document := range yamlDocs {
		if document["kind"] != "MachineDeployment" {
			continue
		}

		workerPoolName, err := traverseMapAndGet[string](document, "metadata.name")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}

		oldReplicas, err := traverseMapAndGet[float64](document, "spec.replicas")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}
		oldNodePools[workerPoolName] = CseWorkerPoolUpdateInput{
			MachineCount: int(oldReplicas),
		}
	}
	if len(oldNodePools) == -1 {
		t.Fatalf("didn't get any valid worker node pool")
	}

	// We call the function to update the old pools with the new ones
	newReplicas := 66
	newNodePools := map[string]CseWorkerPoolUpdateInput{}
	for name := range oldNodePools {
		newNodePools[name] = CseWorkerPoolUpdateInput{
			MachineCount: newReplicas,
		}
	}
	err = cseUpdateWorkerPoolsInYaml(yamlDocs, newNodePools)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// The worker pools should have now the new details updated
	for _, document := range yamlDocs {
		if document["kind"] != "MachineDeployment" {
			continue
		}

		retrievedReplicas, err := traverseMapAndGet[float64](document, "spec.replicas")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}
		if retrievedReplicas != float64(newReplicas) {
			t.Fatalf("expected %d replicas but got %0.f", newReplicas, retrievedReplicas)
		}
	}

	// Corner case: Wrong replicas
	newReplicas = -1
	newNodePools = map[string]CseWorkerPoolUpdateInput{}
	for name := range oldNodePools {
		newNodePools[name] = CseWorkerPoolUpdateInput{
			MachineCount: newReplicas,
		}
	}
	err = cseUpdateWorkerPoolsInYaml(yamlDocs, newNodePools)
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}

	// Corner case: No worker pool with that name exists
	newNodePools = map[string]CseWorkerPoolUpdateInput{
		"not-exist": {},
	}
	err = cseUpdateWorkerPoolsInYaml(yamlDocs, newNodePools)
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}
}

// Test_cseUpdateControlPlaneInYaml tests the update process of the Control Plane in a CAPI YAML.
func Test_cseUpdateControlPlaneInYaml(t *testing.T) {
	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read CAPI YAML test file: %s", err)
	}

	yamlDocs, err := unmarshalMultipleYamlDocuments(string(capiYaml))
	if err != nil {
		t.Fatalf("could not unmarshal CAPI YAML test file: %s", err)
	}
	// We explore the YAML documents to get the OVA template name that will be updated
	// with the new one.
	oldControlPlane := CseControlPlaneUpdateInput{}
	for _, document := range yamlDocs {
		if document["kind"] != "KubeadmControlPlane" {
			continue
		}

		oldReplicas, err := traverseMapAndGet[float64](document, "spec.replicas")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}
		oldControlPlane = CseControlPlaneUpdateInput{
			MachineCount: int(oldReplicas),
		}
	}
	if reflect.DeepEqual(oldControlPlane, CseWorkerPoolUpdateInput{}) {
		t.Fatalf("didn't get any valid Control Plane")
	}

	// We call the function to update the old pools with the new ones
	newReplicas := 66
	newControlPlane := CseControlPlaneUpdateInput{
		MachineCount: newReplicas,
	}
	err = cseUpdateControlPlaneInYaml(yamlDocs, newControlPlane)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// The worker pools should have now the new details updated
	for _, document := range yamlDocs {
		if document["kind"] != "KubeadmControlPlane" {
			continue
		}

		retrievedReplicas, err := traverseMapAndGet[float64](document, "spec.replicas")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}
		if retrievedReplicas != float64(newReplicas) {
			t.Fatalf("expected %d replicas but got %0.f", newReplicas, retrievedReplicas)
		}
	}

	// Corner case: Wrong replicas
	newReplicas = -1
	newControlPlane = CseControlPlaneUpdateInput{
		MachineCount: newReplicas,
	}
	err = cseUpdateControlPlaneInYaml(yamlDocs, newControlPlane)
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}
}

// Test_cseUpdateNodeHealthCheckInYaml tests the update process of the Machine Health Check capabilities in a CAPI YAML.
func Test_cseUpdateNodeHealthCheckInYaml(t *testing.T) {
	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read CAPI YAML test file: %s", err)
	}

	yamlDocs, err := unmarshalMultipleYamlDocuments(string(capiYaml))
	if err != nil {
		t.Fatalf("could not unmarshal CAPI YAML test file: %s", err)
	}

	clusterName := ""
	for _, doc := range yamlDocs {
		if doc["kind"] != "Cluster" {
			continue
		}
		clusterName, err = traverseMapAndGet[string](doc, "metadata.name")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}
	}
	if clusterName == "" {
		t.Fatal("could not find the cluster name in the CAPI YAML test file")
	}

	v, err := semver.NewVersion("4.1")
	if err != nil {
		t.Fatalf("incorrect version: %s", err)
	}

	// Deactivates Machine Health Check
	yamlDocs, err = cseUpdateNodeHealthCheckInYaml(yamlDocs, clusterName, *v, nil)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// The resulting documents should not have that document
	for _, document := range yamlDocs {
		if document["kind"] == "MachineHealthCheck" {
			t.Fatal("Expected the MachineHealthCheck to be deleted, but it is there")
		}
	}

	// Enables Machine Health Check
	yamlDocs, err = cseUpdateNodeHealthCheckInYaml(yamlDocs, clusterName, *v, &vcdKeConfig{
		MaxUnhealthyNodesPercentage: 12,
		NodeStartupTimeout:          "34",
		NodeNotReadyTimeout:         "56",
		NodeUnknownTimeout:          "78",
	})
	if err != nil {
		t.Fatalf("%s", err)
	}

	// The resulting documents should have a MachineHealthCheck
	found := false
	for _, document := range yamlDocs {
		if document["kind"] != "MachineHealthCheck" {
			continue
		}
		maxUnhealthy, err := traverseMapAndGet[string](document, "spec.maxUnhealthy")
		if err != nil {
			t.Fatalf("%s", err)
		}
		if maxUnhealthy != "12%" {
			t.Fatalf("expected a 'spec.maxUnhealthy' = 12%%, but got %s", maxUnhealthy)
		}
		nodeStartupTimeout, err := traverseMapAndGet[string](document, "spec.nodeStartupTimeout")
		if err != nil {
			t.Fatalf("%s", err)
		}
		if nodeStartupTimeout != "34s" {
			t.Fatalf("expected a 'spec.nodeStartupTimeout' = 34s, but got %s", nodeStartupTimeout)
		}
		found = true
	}
	if !found {
		t.Fatalf("expected a MachineHealthCheck block but got nothing")
	}
}

// Test_unmarshalMultplieYamlDocuments tests the unmarshalling of multiple YAML documents with unmarshalMultplieYamlDocuments
func Test_unmarshalMultplieYamlDocuments(t *testing.T) {
	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read YAML test file: %s", err)
	}

	tests := []struct {
		name          string
		yamlDocuments string
		want          int
		wantErr       bool
	}{
		{
			name:          "unmarshal correct amount of documents",
			yamlDocuments: string(capiYaml),
			want:          9,
			wantErr:       false,
		},
		{
			name:          "unmarshal single yaml document",
			yamlDocuments: "test: foo",
			want:          1,
			wantErr:       false,
		},
		{
			name:          "unmarshal empty yaml document",
			yamlDocuments: "",
			want:          0,
		},
		{
			name:          "unmarshal wrong yaml document",
			yamlDocuments: "thisIsNotAYaml",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalMultipleYamlDocuments(tt.yamlDocuments)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalMultplieYamlDocuments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("unmarshalMultplieYamlDocuments() got %d documents, want %d", len(got), tt.want)
			}
		})
	}
}

// Test_marshalMultplieYamlDocuments tests the marshalling of multiple YAML documents with marshalMultplieYamlDocuments
func Test_marshalMultplieYamlDocuments(t *testing.T) {
	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read YAML test file: %s", err)
	}

	unmarshaledCapiYaml, err := unmarshalMultipleYamlDocuments(string(capiYaml))
	if err != nil {
		t.Fatalf("could not unmarshal the YAML test file: %s", err)
	}

	tests := []struct {
		name          string
		yamlDocuments []map[string]interface{}
		want          []map[string]interface{}
		wantErr       bool
	}{
		{
			name:          "marshal correct amount of documents",
			yamlDocuments: unmarshaledCapiYaml,
			want:          unmarshaledCapiYaml,
			wantErr:       false,
		},
		{
			name:          "marshal empty slice",
			yamlDocuments: []map[string]interface{}{},
			want:          []map[string]interface{}{},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalMultipleYamlDocuments(tt.yamlDocuments)
			if (err != nil) != tt.wantErr {
				t.Errorf("marshalMultipleYamlDocuments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotUnmarshaled, err := unmarshalMultipleYamlDocuments(got) // We unmarshal the result to compare it exactly with DeepEqual
			if err != nil {
				t.Errorf("unmarshalMultipleYamlDocuments() failed %s", err)
				return
			}
			if !reflect.DeepEqual(gotUnmarshaled, tt.want) {
				t.Errorf("marshalMultipleYamlDocuments() got =\n%v, want =\n%v", gotUnmarshaled, tt.want)
			}
		})
	}
}

// Test_traverseMapAndGet tests traverseMapAndGet function
func Test_traverseMapAndGet(t *testing.T) {
	type args struct {
		input any
		path  string
	}
	tests := []struct {
		name     string
		args     args
		wantType string
		want     any
		wantErr  string
	}{
		{
			name: "input is nil",
			args: args{
				input: nil,
			},
			wantErr: "the input is nil",
		},
		{
			name: "input is not a map",
			args: args{
				input: "error",
			},
			wantErr: "the input is a string, not a map[string]interface{}",
		},
		{
			name: "map is empty",
			args: args{
				input: map[string]interface{}{},
			},
			wantErr: "the map is empty",
		},
		{
			name: "map does not have key",
			args: args{
				input: map[string]interface{}{
					"keyA": "value",
				},
				path: "keyB",
			},
			wantErr: "key 'keyB' does not exist in input map",
		},
		{
			name: "map has a single simple key",
			args: args{
				input: map[string]interface{}{
					"keyA": "value",
				},
				path: "keyA",
			},
			wantType: "string",
			want:     "value",
		},
		{
			name: "map has a single complex key",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": "value",
					},
				},
				path: "keyA",
			},
			wantType: "map",
			want: map[string]interface{}{
				"keyB": "value",
			},
		},
		{
			name: "map has a complex structure",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": map[string]interface{}{
							"keyC": "value",
						},
					},
				},
				path: "keyA.keyB.keyC",
			},
			wantType: "string",
			want:     "value",
		},
		{
			name: "requested path is deeper than the map structure",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": map[string]interface{}{
							"keyC": "value",
						},
					},
				},
				path: "keyA.keyB.keyC.keyD",
			},
			wantErr: "key 'keyC' is a string, not a map, but there are still 1 paths to explore",
		},
		{
			name: "obtained value does not correspond to the desired type",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": map[string]interface{}{
							"keyC": map[string]interface{}{},
						},
					},
				},
				path: "keyA.keyB.keyC",
			},
			wantType: "string",
			wantErr:  "could not convert obtained type map[string]interface {} to requested string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			var err error
			if tt.wantType == "string" {
				got, err = traverseMapAndGet[string](tt.args.input, tt.args.path)
			} else if tt.wantType == "map" {
				got, err = traverseMapAndGet[map[string]interface{}](tt.args.input, tt.args.path)
			} else {
				t.Fatalf("wantType type not used in this test")
			}

			if err != nil {
				if tt.wantErr != err.Error() {
					t.Errorf("traverseMapAndGet() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("traverseMapAndGet() got = %v, want %v", got, tt.want)
			}
		})
	}
}