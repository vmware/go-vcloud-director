//go:build unit || ALL

package govcd

import (
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

	// We call the function to update the old OVA with the new one
	newOvaName := "my-super-ova-name"
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
		t.Fatalf("failed updating the Kubernetes OVA template in the Control Plane:\n%s", updatedYaml)
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

		oldReplicas, err := traverseMapAndGet[int](document, "spec.replicas")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}
		oldNodePools[workerPoolName] = CseWorkerPoolUpdateInput{
			MachineCount: oldReplicas,
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

		retrievedReplicas, err := traverseMapAndGet[int](document, "spec.replicas")
		if err != nil {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}
		if retrievedReplicas != newReplicas {
			t.Fatalf("expected %d replicas but got %d", newReplicas, retrievedReplicas)
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
		yamlDocuments []map[any]any
		want          []map[any]any
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
			yamlDocuments: []map[any]any{},
			want:          []map[any]any{},
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
			wantErr: "the input is a string, not a map[any]any",
		},
		{
			name: "map is empty",
			args: args{
				input: map[any]any{},
			},
			wantErr: "the map is empty",
		},
		{
			name: "map does not have key",
			args: args{
				input: map[any]any{
					"keyA": "value",
				},
				path: "keyB",
			},
			wantErr: "key 'keyB' does not exist in input map",
		},
		{
			name: "map has a single simple key",
			args: args{
				input: map[any]any{
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
				input: map[any]any{
					"keyA": map[any]any{
						"keyB": "value",
					},
				},
				path: "keyA",
			},
			wantType: "map",
			want: map[any]any{
				"keyB": "value",
			},
		},
		{
			name: "map has a complex structure",
			args: args{
				input: map[any]any{
					"keyA": map[any]any{
						"keyB": map[any]any{
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
				input: map[any]any{
					"keyA": map[any]any{
						"keyB": map[any]any{
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
				input: map[any]any{
					"keyA": map[any]any{
						"keyB": map[any]any{
							"keyC": map[any]any{},
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
				got, err = traverseMapAndGet[map[any]any](tt.args.input, tt.args.path)
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
