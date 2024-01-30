//go:build functional || openapi || cse || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
	"os"
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

	workerPoolName := "worker-pool-1"

	cluster, err := org.CseCreateKubernetesCluster(CseClusterCreateInput{
		Name:                    "test-cse",
		OrganizationId:          org.Org.ID,
		VdcId:                   vdc.Vdc.ID,
		NetworkId:               net.OrgVDCNetwork.ID,
		KubernetesTemplateOvaId: ova.VAppTemplate.ID,
		CseVersion:              "4.2",
		ControlPlane: ControlPlaneCreateInput{
			MachineCount:     1,
			DiskSizeGi:       20,
			SizingPolicyId:   policies[0].VdcComputePolicyV2.ID,
			StorageProfileId: sp.ID,
			Ip:               "",
		},
		WorkerPools: []WorkerPoolCreateInput{{
			Name:             workerPoolName,
			MachineCount:     1,
			DiskSizeGi:       20,
			SizingPolicyId:   policies[0].VdcComputePolicyV2.ID,
			StorageProfileId: sp.ID,
		}},
		DefaultStorageClass: &DefaultStorageClassCreateInput{
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
	}, 0)
	check.Assert(err, IsNil)
	check.Assert(cluster.ID, Not(Equals), "")
	check.Assert(cluster.Etag, Not(Equals), "")
	check.Assert(cluster.Capvcd.Status.VcdKe.State, Equals, "provisioned")

	err = cluster.Refresh()
	check.Assert(err, IsNil)

	clusterGet, err := org.CseGetKubernetesClusterById(cluster.ID)
	check.Assert(err, IsNil)
	check.Assert(cluster.ID, Equals, clusterGet.ID)
	check.Assert(cluster.Capvcd.Name, Equals, clusterGet.Capvcd.Name)
	check.Assert(cluster.Owner, Equals, clusterGet.Owner)
	check.Assert(cluster.Capvcd.Metadata, DeepEquals, clusterGet.Capvcd.Metadata)
	check.Assert(cluster.Capvcd.Spec.VcdKe, DeepEquals, clusterGet.Capvcd.Spec.VcdKe)

	// Update worker pool from 1 node to 2
	// Pre-check. This should be 1, as it was created with just 1 pool
	for _, nodePool := range cluster.Capvcd.Status.Capvcd.NodePool {
		if nodePool.Name == workerPoolName {
			check.Assert(nodePool.DesiredReplicas, Equals, 1)
		}
	}
	// Perform the update
	err = cluster.UpdateWorkerPools(map[string]WorkerPoolUpdateInput{workerPoolName: {MachineCount: 2}}, 0)
	check.Assert(err, IsNil)

	// Post-check. This should be 2, as it should have scaled up
	foundWorkerPool := false
	for _, nodePool := range cluster.Capvcd.Status.Capvcd.NodePool {
		if nodePool.Name == workerPoolName {
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
