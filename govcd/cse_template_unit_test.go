//go:build unit || ALL

package govcd

import (
	semver "github.com/hashicorp/go-version"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_cseClusterSettingsInternal_generateCapiYamlAsJsonString(t *testing.T) {
	v41, err := semver.NewVersion("4.1")
	if err != nil {
		t.Fatalf("%s", err)
	}

	capiYaml, err := os.ReadFile("test-resources/capiYaml.yaml")
	if err != nil {
		t.Fatalf("could not read YAML test file: %s", err)
	}
	baseUnmarshaledYaml, err := unmarshalMultipleYamlDocuments(string(capiYaml))
	if err != nil {
		t.Fatalf("could not unmarshal YAML test file: %s", err)
	}

	tests := []struct {
		name         string
		input        cseClusterSettingsInternal
		expectedFunc func() []map[string]interface{}
		wantErr      string
	}{
		{
			name: "correct YAML without optionals",
			input: cseClusterSettingsInternal{
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
					ContainerRegistryUrl:        "projects.registry.vmware.com/tkg",
				},
				Owner:       "dummy",
				ApiToken:    "dummy",
				VcdUrl:      "https://www.my-vcd-instance.com",
				PodCidr:     "100.96.0.0/11",
				ServiceCidr: "100.64.0.0/13",
			},
			expectedFunc: func() []map[string]interface{} {
				return baseUnmarshaledYaml
			},
		},
		{
			name: "correct YAML without MachineHealthCheck",
			input: cseClusterSettingsInternal{
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
					ContainerRegistryUrl: "projects.registry.vmware.com/tkg",
				},
				Owner:       "dummy",
				ApiToken:    "dummy",
				VcdUrl:      "https://www.my-vcd-instance.com",
				PodCidr:     "100.96.0.0/11",
				ServiceCidr: "100.64.0.0/13",
			},
			expectedFunc: func() []map[string]interface{} {
				var result []map[string]interface{}
				for _, doc := range baseUnmarshaledYaml {
					if doc["kind"] == "MachineHealthCheck" {
						continue
					}
					result = append(result, doc)
				}
				return result
			},
		},
		{
			name: "correct YAML with every possible option",
			input: cseClusterSettingsInternal{
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
					Ip:                 "1.2.3.4",
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
					ContainerRegistryUrl:        "projects.registry.vmware.com/tkg",
				},
				VirtualIpSubnet: "6.7.8.9/24",
				Owner:           "dummy",
				ApiToken:        "dummy",
				VcdUrl:          "https://www.my-vcd-instance.com",
				PodCidr:         "100.96.0.0/11",
				ServiceCidr:     "100.64.0.0/13",
			},
			expectedFunc: func() []map[string]interface{} {
				var result []map[string]interface{}
				for _, doc := range baseUnmarshaledYaml {
					if doc["kind"] == "VCDCluster" {
						doc["spec"].(map[string]interface{})["controlPlaneEndpoint"] = map[string]interface{}{"host": "1.2.3.4"}
						doc["spec"].(map[string]interface{})["controlPlaneEndpoint"].(map[string]interface{})["port"] = 6443
						doc["spec"].(map[string]interface{})["loadBalancerConfigSpec"] = map[string]string{"vipSubnet": "6.7.8.9/24"}
					}
					result = append(result, doc)
				}
				return result
			},
		},
		{
			name: "wrong YAML with both Placement and vGPU policies in a Worker Pool",
			input: cseClusterSettingsInternal{
				CseVersion: *v41,
				WorkerPools: []cseWorkerPoolSettingsInternal{
					{
						Name:                "node-pool-1",
						MachineCount:        1,
						DiskSizeGi:          20,
						SizingPolicyName:    "TKG small",
						PlacementPolicyName: "policy",
						VGpuPolicyName:      "policy",
						StorageProfileName:  "*",
					},
				},
			},
			wantErr: "the worker pool 'node-pool-1' should have either a Placement Policy or a vGPU Policy, not both",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.generateCapiYamlAsJsonString()
			if err != nil {
				if err.Error() != tt.wantErr {
					t.Errorf("generateCapiYamlAsJsonString() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			gotUnmarshaled, err := unmarshalMultipleYamlDocuments(strings.NewReplacer("\\n", "\n", "\\\"", "\"").Replace(got))
			if err != nil {
				t.Fatalf("could not unmarshal obtained YAML: %s", err)
			}

			expected := tt.expectedFunc()
			if !reflect.DeepEqual(expected, gotUnmarshaled) {
				t.Errorf("generateCapiYamlAsJsonString() got =\n%v\nwant =\n%v\n", gotUnmarshaled, expected)
			}
		})
	}
}
