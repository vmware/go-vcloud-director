//go:build functional || openapi || cse || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	semver "github.com/hashicorp/go-version"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
	"os"
	"strings"
)

const (
	TestRequiresCseConfiguration = "Test %s requires CSE configuration details"
)

func skipCseTests(testConfig TestConfig) bool {
	if cse := os.Getenv("TEST_VCD_CSE"); cse == "" {
		return true
	}
	return testConfig.Cse.SolutionsOrg == "" || testConfig.Cse.TenantOrg == "" || testConfig.Cse.OvaName == "" ||
		testConfig.Cse.RoutedNetwork == "" || testConfig.Cse.EdgeGateway == "" || testConfig.Cse.OvaCatalog == "" || testConfig.Cse.TenantVdc == "" ||
		testConfig.VCD.StorageProfile.SP1 == ""
}

// Test_Cse
func (vcd *TestVCD) Test_Cse(check *C) {
	if skipCseTests(vcd.config) {
		check.Skip(fmt.Sprintf(TestRequiresCseConfiguration, check.TestName()))
	}

	org, err := vcd.client.GetOrgByName(vcd.config.Cse.TenantOrg)
	check.Assert(err, IsNil)

	catalog, err := org.GetCatalogByName(vcd.config.Cse.OvaCatalog, false)
	check.Assert(err, IsNil)

	ova, err := catalog.GetVAppTemplateByName(vcd.config.Cse.OvaName)
	check.Assert(err, IsNil)
	tkgBundle, err := getTkgVersionBundleFromVAppTemplateName(ova.VAppTemplate.Name)
	check.Assert(err, IsNil)

	vdc, err := org.GetVDCByName(vcd.config.Cse.TenantVdc, false)
	check.Assert(err, IsNil)

	net, err := vdc.GetOrgVdcNetworkByName(vcd.config.Cse.RoutedNetwork, false)
	check.Assert(err, IsNil)

	sp, err := vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)

	policies, err := vcd.client.GetAllVdcComputePoliciesV2(url.Values{
		"filter": []string{"name==TKG small"},
	})
	check.Assert(err, IsNil)
	check.Assert(len(policies), Equals, 1)

	token, err := vcd.client.CreateToken(vcd.config.Provider.SysOrg, check.TestName()+"124") // TODO: Remove number suffix
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTokens+token.Token.ID)

	apiToken, err := token.GetInitialApiToken()
	check.Assert(err, IsNil)

	cseVersion, err := semver.NewVersion("4.1")
	check.Assert(err, IsNil)
	check.Assert(cseVersion, NotNil)

	clusterSettings := CseClusterSettings{
		Name:                    "test-cse",
		OrganizationId:          org.Org.ID,
		VdcId:                   vdc.Vdc.ID,
		NetworkId:               net.OrgVDCNetwork.ID,
		KubernetesTemplateOvaId: ova.VAppTemplate.ID,
		CseVersion:              *cseVersion,
		ControlPlane: CseControlPlaneSettings{
			MachineCount:     1,
			DiskSizeGi:       20,
			SizingPolicyId:   policies[0].VdcComputePolicyV2.ID,
			StorageProfileId: sp.ID,
			Ip:               "",
		},
		WorkerPools: []CseWorkerPoolSettings{{
			Name:             "worker-pool-1",
			MachineCount:     1,
			DiskSizeGi:       20,
			SizingPolicyId:   policies[0].VdcComputePolicyV2.ID,
			StorageProfileId: sp.ID,
		}},
		DefaultStorageClass: &CseDefaultStorageClassSettings{
			StorageProfileId: sp.ID,
			Name:             "storage-class-1",
			ReclaimPolicy:    "delete",
			Filesystem:       "ext4",
		},
		Owner:              vcd.config.Provider.User,
		ApiToken:           apiToken.RefreshToken,
		NodeHealthCheck:    true,
		PodCidr:            "100.96.0.0/11",
		ServiceCidr:        "100.64.0.0/13",
		AutoRepairOnErrors: true,
	}

	cluster, err := org.CseCreateKubernetesCluster(clusterSettings, 0)
	check.Assert(err, IsNil)
	check.Assert(true, Equals, strings.Contains(cluster.ID, "urn:vcloud:entity:vmware:capvcdCluster:"))
	check.Assert(cluster.Etag, Not(Equals), "")
	check.Assert(cluster.CseClusterSettings, DeepEquals, clusterSettings)
	check.Assert(cluster.KubernetesVersion, Equals, tkgBundle.KubernetesVersion)
	check.Assert(cluster.TkgVersion, Equals, tkgBundle.TkgVersion)
	check.Assert(cluster.CapvcdVersion, Not(Equals), "")
	check.Assert(cluster.CpiVersion, Not(Equals), "")
	check.Assert(cluster.CsiVersion, Not(Equals), "")
	check.Assert(len(cluster.ClusterResourceSetBindings), Not(Equals), 0)
	check.Assert(cluster.State, Equals, "provisioned")
	check.Assert(len(cluster.Events), Not(Equals), 0)

	kubeconfig, err := cluster.GetKubeconfig()
	check.Assert(err, IsNil)
	check.Assert(true, Equals, strings.Contains(kubeconfig, cluster.Name))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-certificate-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "certificate-authority-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-key-data"))

	err = cluster.Refresh()
	check.Assert(err, IsNil)

	clusterGet, err := org.CseGetKubernetesClusterById(cluster.ID)
	check.Assert(err, IsNil)
	check.Assert(cluster.ID, Equals, clusterGet.ID)
	check.Assert(cluster.Name, Equals, clusterGet.Name)
	check.Assert(cluster.Owner, Equals, clusterGet.Owner)
	check.Assert(cluster.capvcdType.Metadata, DeepEquals, clusterGet.capvcdType.Metadata)
	check.Assert(cluster.capvcdType.Spec.VcdKe, DeepEquals, clusterGet.capvcdType.Spec.VcdKe)

	// Update worker pool from 1 node to 2
	// Pre-check. This should be 1, as it was created with just 1 pool
	for _, nodePool := range cluster.capvcdType.Status.Capvcd.NodePool {
		if nodePool.Name == clusterSettings.WorkerPools[0].Name {
			check.Assert(nodePool.DesiredReplicas, Equals, 1)
		}
	}
	// Perform the update
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{clusterSettings.WorkerPools[0].Name: {MachineCount: 2}}, true)
	check.Assert(err, IsNil)

	// Post-check. This should be 2, as it should have scaled up
	foundWorkerPool := false
	for _, nodePool := range cluster.capvcdType.Status.Capvcd.NodePool {
		if nodePool.Name == clusterSettings.WorkerPools[0].Name {
			foundWorkerPool = true
			check.Assert(nodePool.DesiredReplicas, Equals, 2)
		}
	}
	check.Assert(foundWorkerPool, Equals, true)

	err = cluster.Delete(0)
	check.Assert(err, IsNil)

	err = token.Delete()
	check.Assert(err, IsNil)

}

func (vcd *TestVCD) Test_Deleteme(check *C) {
	org, err := vcd.client.GetOrgByName(vcd.config.Cse.TenantOrg)
	check.Assert(err, IsNil)

	cluster, err := org.CseGetKubernetesClusterById("urn:vcloud:entity:vmware:capvcdCluster:e8e82bcc-50a1-484f-9dd0-20965ab3e865")
	check.Assert(err, IsNil)

	workerPoolName := "cse-test1-worker-node-pool-1"

	kubeconfig, err := cluster.GetKubeconfig()
	check.Assert(err, IsNil)
	check.Assert(true, Equals, strings.Contains(kubeconfig, cluster.Name))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-certificate-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "certificate-authority-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-key-data"))

	// Perform the update
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{workerPoolName: {MachineCount: 1}}, true)
	check.Assert(err, IsNil)

	// Post-check. This should be 2, as it should have scaled up
	foundWorkerPool := false
	for _, nodePool := range cluster.WorkerPools {
		if nodePool.Name == workerPoolName {
			foundWorkerPool = true
			check.Assert(nodePool.MachineCount, Equals, 1)
		}
	}
	check.Assert(foundWorkerPool, Equals, true)

}