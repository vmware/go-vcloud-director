//go:build unit || ALL

package govcd

import (
	semver "github.com/hashicorp/go-version"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
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

	oldOvaVersion := "v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc" // Matches the version in capiYaml.yaml
	if strings.Count(string(capiYaml), oldOvaVersion) < 2 {
		t.Fatalf("the testing YAML doesn't contain the OVA to change")
	}

	oldTkgBundle, err := getTkgVersionBundleFromVAppTemplate(&types.VAppTemplate{
		Name: "dummy",
		Children: &types.VAppTemplateChildren{VM: []*types.VAppTemplate{
			{
				ProductSection: &types.ProductSection{
					Property: []*types.Property{
						{
							Key:          "VERSION",
							DefaultValue: oldOvaVersion,
						},
					},
				},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	// We call the function to update the old OVA with the new one
	newOva := &types.VAppTemplate{
		ID:   "urn:vcloud:vapptemplate:e23b8a5c-e676-4d82-9050-c906a3ac2fea",
		Name: "dummy",
		Children: &types.VAppTemplateChildren{VM: []*types.VAppTemplate{
			{
				ProductSection: &types.ProductSection{
					Property: []*types.Property{
						{
							Key:          "VERSION",
							DefaultValue: "v1.19.16+vmware.1-tkg.2-fba68db15591c15fcd5f26b512663a42",
						},
					},
				},
			},
		}},
	}
	newTkgBundle, err := getTkgVersionBundleFromVAppTemplate(newOva)
	if err != nil {
		t.Fatal(err)
	}

	err = cseUpdateKubernetesTemplateInYaml(yamlDocs, newOva)
	if err != nil {
		t.Fatal(err)
	}

	updatedYaml, err := marshalMultipleYamlDocuments(yamlDocs)
	if err != nil {
		t.Fatalf("error marshaling %v: %s", yamlDocs, err)
	}

	// No document should have the old OVA
	if strings.Count(updatedYaml, newOva.Name) < 2 || strings.Contains(updatedYaml, oldOvaVersion) {
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
	// We explore the YAML documents to get the current Worker pool
	oldNodePools := map[string]CseWorkerPoolUpdateInput{}
	for _, document := range yamlDocs {
		if document["kind"] != "MachineDeployment" {
			continue
		}

		workerPoolName := traverseMapAndGet[string](document, "metadata.name", ".")
		if workerPoolName == "" {
			t.Fatalf("incorrect CAPI YAML: %s", err)
		}

		oldNodePools[workerPoolName] = CseWorkerPoolUpdateInput{
			MachineCount: int(traverseMapAndGet[float64](document, "spec.replicas", ".")),
		}
	}
	if len(oldNodePools) == 0 {
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
		t.Fatal(err)
	}

	// The worker pools should have now the new details updated
	for _, document := range yamlDocs {
		if document["kind"] != "MachineDeployment" {
			continue
		}

		retrievedReplicas := traverseMapAndGet[float64](document, "spec.replicas", ".")
		if traverseMapAndGet[float64](document, "spec.replicas", ".") != float64(newReplicas) {
			t.Fatalf("expected %d replicas but got %0.f", newReplicas, retrievedReplicas)
		}
		autoscalerMinSize := traverseMapAndGet[string](document, "metadata|annotations|cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size", "|")
		if autoscalerMinSize != "" {
			t.Fatalf("didn't expect autoscaler min size but got '%s'", autoscalerMinSize)
		}
		autoscalerMaxSize := traverseMapAndGet[string](document, "metadata|annotations|cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size", "|")
		if autoscalerMaxSize != "" {
			t.Fatalf("didn't expect autoscaler max size but got '%s'", autoscalerMaxSize)
		}
	}

	// We call the function to update the pools with autoscaling
	newNodePools = map[string]CseWorkerPoolUpdateInput{}
	for name := range oldNodePools {
		newNodePools[name] = CseWorkerPoolUpdateInput{
			MachineCount: -1, // We don't care about replicas as we use autoscaling
			Autoscaler: &CseWorkerPoolAutoscaler{
				MaxSize: 50,
				MinSize: 10,
			},
		}
	}
	err = cseUpdateWorkerPoolsInYaml(yamlDocs, newNodePools)
	if err != nil {
		t.Fatal(err)
	}

	// The worker pools should have now the new details updated
	for _, document := range yamlDocs {
		if document["kind"] != "MachineDeployment" {
			continue
		}

		retrievedReplicas := traverseMapAndGet[float64](document, "spec.replicas", ".")
		if retrievedReplicas != 0 {
			t.Fatalf("didn't expect replicas but got '%0.f'", retrievedReplicas)
		}
		autoscalerMinSize := traverseMapAndGet[string](document, "metadata|annotations|cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size", "|")
		if autoscalerMinSize != "10" {
			t.Fatalf("expected autoscaler min size '10' but got '%s'", autoscalerMinSize)
		}
		autoscalerMaxSize := traverseMapAndGet[string](document, "metadata|annotations|cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size", "|")
		if autoscalerMaxSize != "50" {
			t.Fatalf("expected autoscaler min size '50' but got '%s'", autoscalerMaxSize)
		}
	}

	// Corner case: Wrong replicas
	newReplicas = -1
	newNodePools = map[string]CseWorkerPoolUpdateInput{}
	for name := range oldNodePools {
		newNodePools[name] = CseWorkerPoolUpdateInput{
			MachineCount: newReplicas,
			Autoscaler:   nil,
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
	// We explore the YAML documents to get the current Control plane
	oldControlPlane := CseControlPlaneUpdateInput{}
	for _, document := range yamlDocs {
		if document["kind"] != "KubeadmControlPlane" {
			continue
		}

		oldControlPlane = CseControlPlaneUpdateInput{
			MachineCount: int(traverseMapAndGet[float64](document, "spec.replicas", ".")),
		}
	}
	if reflect.DeepEqual(oldControlPlane, CseWorkerPoolUpdateInput{}) {
		t.Fatalf("didn't get any valid Control Plane")
	}

	// We call the function to update the control plane
	newReplicas := 67
	newControlPlane := CseControlPlaneUpdateInput{
		MachineCount: newReplicas,
	}
	err = cseUpdateControlPlaneInYaml(yamlDocs, newControlPlane)
	if err != nil {
		t.Fatal(err)
	}

	// The control plane should have now the new details updated
	for _, document := range yamlDocs {
		if document["kind"] != "KubeadmControlPlane" {
			continue
		}

		retrievedReplicas := traverseMapAndGet[float64](document, "spec.replicas", ".")
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

	newReplicas = 2 // Should be odd, hence fails
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
		clusterName = traverseMapAndGet[string](doc, "metadata.name", ".")
	}
	if clusterName == "" {
		t.Fatal("could not find the cluster name in the CAPI YAML test file")
	}

	v, err := semver.NewVersion("4.1")
	if err != nil {
		t.Fatalf("incorrect version: %s", err)
	}

	// Deactivates Machine Health Check
	yamlDocs, err = cseUpdateNodeHealthCheckInYaml(yamlDocs, clusterName, *v, vcdKeConfig{})
	if err != nil {
		t.Fatal(err)
	}

	// The resulting documents should not have that document
	for _, document := range yamlDocs {
		if document["kind"] == "MachineHealthCheck" {
			t.Fatal("Expected the MachineHealthCheck to be deleted, but it is there")
		}
	}

	// Enables Machine Health Check
	yamlDocs, err = cseUpdateNodeHealthCheckInYaml(yamlDocs, clusterName, *v, vcdKeConfig{
		MaxUnhealthyNodesPercentage: 12,
		NodeStartupTimeout:          "34",
		NodeNotReadyTimeout:         "56",
		NodeUnknownTimeout:          "78",
	})
	if err != nil {
		t.Fatal(err)
	}

	// The resulting documents should have a MachineHealthCheck
	found := false
	for _, document := range yamlDocs {
		if document["kind"] != "MachineHealthCheck" {
			continue
		}
		maxUnhealthy := traverseMapAndGet[string](document, "spec.maxUnhealthy", ".")
		if maxUnhealthy != "12%" {
			t.Fatalf("expected a 'spec.maxUnhealthy' = 12%%, but got %s", maxUnhealthy)
		}
		nodeStartupTimeout := traverseMapAndGet[string](document, "spec.nodeStartupTimeout", ".")
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
			want:          8,
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
		input     interface{}
		path      string
		separator string
	}
	tests := []struct {
		name     string
		args     args
		wantType string
		want     interface{}
	}{
		{
			name: "input is nil",
			args: args{
				input:     nil,
				separator: ".",
			},
			wantType: "string",
			want:     "",
		},
		{
			name: "input is not a map",
			args: args{
				input:     "error",
				separator: ".",
			},
			wantType: "string",
			want:     "",
		},
		{
			name: "map is empty",
			args: args{
				input:     map[string]interface{}{},
				separator: ".",
			},
			wantType: "float64",
			want:     float64(0),
		},
		{
			name: "map does not have key",
			args: args{
				input: map[string]interface{}{
					"keyA": "value",
				},
				path:      "keyB",
				separator: ".",
			},
			wantType: "string",
			want:     "",
		},
		{
			name: "map has a single simple key",
			args: args{
				input: map[string]interface{}{
					"keyA": "value",
				},
				path:      "keyA",
				separator: ".",
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
				path:      "keyA",
				separator: ".",
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
				path:      "keyA.keyB.keyC",
				separator: ".",
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
				path:      "keyA.keyB.keyC.keyD",
				separator: ".",
			},
			wantType: "string",
			want:     "",
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
				path:      "keyA.keyB.keyC",
				separator: ".",
			},
			wantType: "string",
			want:     "",
		},
		{
			name: "requested path has special characters but separator is different",
			args: args{
				input: map[string]interface{}{
					"keyA.foo.bar": "result",
				},
				path:      "keyA.foo.bar",
				separator: "#",
			},
			wantType: "string",
			want:     "result",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			if tt.wantType == "string" {
				got = traverseMapAndGet[string](tt.args.input, tt.args.path, tt.args.separator)
			} else if tt.wantType == "map" {
				got = traverseMapAndGet[map[string]interface{}](tt.args.input, tt.args.path, tt.args.separator)
			} else if tt.wantType == "float64" {
				got = traverseMapAndGet[float64](tt.args.input, tt.args.path, tt.args.separator)
			} else {
				t.Fatalf("wantType type not used in this test")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("traverseMapAndGet() got = %v, want %v", got, tt.want)
			}
		})
	}
}
