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
	tkgBundle, err := getTkgVersionBundleFromVAppTemplate(ova.VAppTemplate)
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
	check.Assert(cluster.CseVersion.String(), Equals, cseVersion.String())
	check.Assert(cluster.Name, Equals, clusterSettings.Name)
	check.Assert(cluster.OrganizationId, Equals, clusterSettings.OrganizationId)
	check.Assert(cluster.VdcId, Equals, clusterSettings.VdcId)
	check.Assert(cluster.NetworkId, Equals, clusterSettings.NetworkId)
	check.Assert(cluster.KubernetesTemplateOvaId, Equals, clusterSettings.KubernetesTemplateOvaId)
	check.Assert(cluster.ControlPlane, DeepEquals, clusterSettings.ControlPlane)
	check.Assert(cluster.WorkerPools, DeepEquals, clusterSettings.WorkerPools)
	check.Assert(cluster.DefaultStorageClass, NotNil)
	check.Assert(*cluster.DefaultStorageClass, DeepEquals, *clusterSettings.DefaultStorageClass)
	check.Assert(cluster.Owner, Equals, clusterSettings.Owner)
	check.Assert(cluster.ApiToken, Not(Equals), clusterSettings.ApiToken)
	check.Assert(cluster.ApiToken, Equals, "******") // This one can't be recovered
	check.Assert(cluster.NodeHealthCheck, Equals, clusterSettings.NodeHealthCheck)
	check.Assert(cluster.PodCidr, Equals, clusterSettings.PodCidr)
	check.Assert(cluster.ServiceCidr, Equals, clusterSettings.ServiceCidr)
	check.Assert(cluster.SshPublicKey, Equals, clusterSettings.SshPublicKey)
	check.Assert(cluster.VirtualIpSubnet, Equals, clusterSettings.VirtualIpSubnet)
	check.Assert(cluster.AutoRepairOnErrors, Equals, clusterSettings.AutoRepairOnErrors)
	check.Assert(cluster.VirtualIpSubnet, Equals, clusterSettings.VirtualIpSubnet)
	check.Assert(true, Equals, strings.Contains(cluster.ID, "urn:vcloud:entity:vmware:capvcdCluster:"))
	check.Assert(cluster.Etag, Not(Equals), "")
	check.Assert(cluster.KubernetesVersion, Equals, tkgBundle.KubernetesVersion)
	check.Assert(cluster.TkgVersion, Equals, tkgBundle.TkgVersion)
	check.Assert(cluster.CapvcdVersion, Not(Equals), "")
	check.Assert(cluster.CpiVersion, Not(Equals), "")
	check.Assert(cluster.CsiVersion, Not(Equals), "")
	check.Assert(cluster.Upgradeable, Equals, true)
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
	check.Assert(cluster.CseVersion.String(), Equals, clusterGet.CseVersion.String())
	check.Assert(cluster.Name, Equals, clusterGet.Name)
	check.Assert(cluster.OrganizationId, Equals, clusterGet.OrganizationId)
	check.Assert(cluster.VdcId, Equals, clusterGet.VdcId)
	check.Assert(cluster.NetworkId, Equals, clusterGet.NetworkId)
	check.Assert(cluster.KubernetesTemplateOvaId, Equals, clusterGet.KubernetesTemplateOvaId)
	check.Assert(cluster.ControlPlane, DeepEquals, clusterGet.ControlPlane)
	check.Assert(cluster.WorkerPools, DeepEquals, clusterGet.WorkerPools)
	check.Assert(cluster.DefaultStorageClass, NotNil)
	check.Assert(*cluster.DefaultStorageClass, DeepEquals, *clusterGet.DefaultStorageClass)
	check.Assert(cluster.Owner, Equals, clusterGet.Owner)
	check.Assert(cluster.ApiToken, Not(Equals), clusterGet.ApiToken)
	check.Assert(clusterGet.ApiToken, Equals, "******") // This one can't be recovered
	check.Assert(cluster.NodeHealthCheck, Equals, clusterGet.NodeHealthCheck)
	check.Assert(cluster.PodCidr, Equals, clusterGet.PodCidr)
	check.Assert(cluster.ServiceCidr, Equals, clusterGet.ServiceCidr)
	check.Assert(cluster.SshPublicKey, Equals, clusterGet.SshPublicKey)
	check.Assert(cluster.VirtualIpSubnet, Equals, clusterGet.VirtualIpSubnet)
	check.Assert(cluster.AutoRepairOnErrors, Equals, clusterGet.AutoRepairOnErrors)
	check.Assert(cluster.VirtualIpSubnet, Equals, clusterGet.VirtualIpSubnet)
	check.Assert(cluster.ID, Equals, clusterGet.ID)
	check.Assert(clusterGet.Etag, Not(Equals), "")
	check.Assert(cluster.KubernetesVersion, Equals, clusterGet.KubernetesVersion)
	check.Assert(cluster.TkgVersion.String(), Equals, clusterGet.TkgVersion.String())
	check.Assert(cluster.CapvcdVersion.String(), Equals, clusterGet.CapvcdVersion.String())
	check.Assert(cluster.ClusterResourceSetBindings, DeepEquals, clusterGet.ClusterResourceSetBindings)
	check.Assert(cluster.CpiVersion.String(), Equals, clusterGet.CpiVersion.String())
	check.Assert(cluster.CsiVersion.String(), Equals, clusterGet.CsiVersion.String())
	check.Assert(cluster.Upgradeable, Equals, clusterGet.Upgradeable)
	check.Assert(cluster.State, Equals, clusterGet.State)

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

	cluster, err := org.CseGetKubernetesClusterById("urn:vcloud:entity:vmware:capvcdCluster:60e287b2-db49-4316-84c0-e0d3d58e8f52")
	check.Assert(err, IsNil)

	upgrades, err := cluster.GetSupportedUpgrades(true)
	check.Assert(err, IsNil)
	check.Assert(len(upgrades) > 0, Equals, true)

	workerPoolName := "cse-test1-worker-node-pool-1"

	kubeconfig, err := cluster.GetKubeconfig()
	check.Assert(err, IsNil)
	check.Assert(true, Equals, strings.Contains(kubeconfig, cluster.Name))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-certificate-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "certificate-authority-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-key-data"))

	// Perform the update
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{workerPoolName: {MachineCount: 2}}, true)
	check.Assert(err, IsNil)

	// Post-check. This should be 2, as it should have scaled up
	foundWorkerPool := false
	for _, nodePool := range cluster.WorkerPools {
		if nodePool.Name == workerPoolName {
			foundWorkerPool = true
			check.Assert(nodePool.MachineCount, Equals, 2)
		}
	}
	check.Assert(foundWorkerPool, Equals, true)

	// Revert back (resources can be limited)
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{workerPoolName: {MachineCount: 1}}, true)
	check.Assert(err, IsNil)

	// Perform the update
	err = cluster.UpdateControlPlane(CseControlPlaneUpdateInput{MachineCount: 2}, true)
	check.Assert(err, IsNil)

	// Post-check. This should be 2, as it should have scaled up
	check.Assert(cluster.ControlPlane.MachineCount, Equals, 2)

	// Revert back (resources can be limited)
	err = cluster.UpdateControlPlane(CseControlPlaneUpdateInput{MachineCount: 1}, true)
	check.Assert(err, IsNil)

	err = cluster.UpdateControlPlane(CseControlPlaneUpdateInput{MachineCount: 1}, true)
	check.Assert(err, IsNil)
}
