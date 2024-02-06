//go:build unit || ALL

package govcd

import (
	semver "github.com/hashicorp/go-version"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Test_cseClusterSettingsInternal_generateCapiYamlAsJsonString tests cseClusterSettingsInternal.generateCapiYamlAsJsonString
func Test_cseClusterSettingsInternal_generateCapiYamlAsJsonString(t *testing.T) {
	v41, err := semver.NewVersion("4.1")
	if err != nil {
		t.Fatalf("%s", err)
	}

	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read YAML test file: %s", err)
	}
	expected, err := unmarshalMultipleYamlDocuments(string(capiYaml))
	if err != nil {
		t.Fatalf("could not unmarshal YAML test file: %s", err)
	}

	clusterSettings := cseClusterSettingsInternal{
		CseVersion:                *v41,
		Name:                      "test1",
		OrganizationName:          "tenant_org",
		VdcName:                   "tenant_vdc",
		NetworkName:               "tenant_net_routed",
		KubernetesTemplateOvaName: "ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc",
		TkgVersionBundle: tkgVersionBundle{
			EtcdVersion:       "v3.5.6_vmware.9",
			CoreDnsVersion:    "v1.9.3_vmware.8",
			TkgVersion:        "v2.2.0",
			TkrVersion:        "v1.25.7---vmware.2-tkg.1",
			KubernetesVersion: "v1.25.7+vmware.2",
		},
		CatalogName: "tkgm_catalog",
		ControlPlane: cseControlPlaneSettingsInternal{
			MachineCount:       1,
			DiskSizeGi:         20,
			SizingPolicyName:   "TKG small",
			StorageProfileName: "*",
		},
		WorkerPools: []cseWorkerPoolSettingsInternal{
			{
				Name:               "node-pool-1",
				MachineCount:       1,
				DiskSizeGi:         20,
				SizingPolicyName:   "TKG small",
				StorageProfileName: "*",
			},
		},
		VcdKeConfig: vcdKeConfig{
			MaxUnhealthyNodesPercentage: 100,
			NodeStartupTimeout:          "900",
			NodeNotReadyTimeout:         "300",
			NodeUnknownTimeout:          "200",
			ContainerRegistryUrl:        "projects.registry.vmware.com",
		},
		Owner:       "dummy",
		ApiToken:    "dummy",
		VcdUrl:      "https://www.my-vcd-instance.com",
		PodCidr:     "100.96.0.0/11",
		ServiceCidr: "100.64.0.0/13",
	}
	got, err := clusterSettings.generateCapiYamlAsJsonString()
	if err != nil {
		t.Fatalf("generateCapiYamlAsJsonString() failed: %s", err)
	}

	gotUnmarshaled, err := unmarshalMultipleYamlDocuments(strings.NewReplacer("\\n", "\n", "\\\"", "\"").Replace(got))
	if err != nil {
		t.Fatalf("could not unmarshal obtained YAML: %s", err)
	}

	if !reflect.DeepEqual(expected, gotUnmarshaled) {
		t.Errorf("generateCapiYamlAsJsonString() got = %v, want %v", gotUnmarshaled, expected)
	}
}
