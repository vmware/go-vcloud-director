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
	"reflect"
	"strings"
	"time"
)

func requireCseConfig(check *C, testConfig TestConfig) {
	skippedPrefix := fmt.Sprintf("skipped %s because:", check.TestName())
	if cse := os.Getenv("TEST_VCD_CSE"); cse == "" {
		check.Skip(fmt.Sprintf("%s the environment variable TEST_VCD_CSE is not set", skippedPrefix))
	}
	cseConfigValues := reflect.ValueOf(testConfig.Cse)
	cseConfigType := cseConfigValues.Type()
	for i := 0; i < cseConfigValues.NumField(); i++ {
		if cseConfigValues.Field(i).String() == "" {
			check.Skip(fmt.Sprintf("%s the config value '%s' inside 'cse' block of govcd_test_config.yaml is not set", skippedPrefix, cseConfigType.Field(i).Name))
		}
	}
}

// Test_Cse tests all possible combinations of the CSE CRUD operations.
func (vcd *TestVCD) Test_Cse(check *C) {
	requireCseConfig(check, vcd.config)

	// Prerequisites: We need to read several items before creating the cluster.
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

	sp, err := vdc.FindStorageProfileReference(vcd.config.Cse.StorageProfile)
	check.Assert(err, IsNil)

	policies, err := vcd.client.GetAllVdcComputePoliciesV2(url.Values{
		"filter": []string{"name==TKG small"},
	})
	check.Assert(err, IsNil)
	check.Assert(len(policies), Equals, 1)

	token, err := vcd.client.CreateToken(vcd.config.Provider.SysOrg, check.TestName())
	check.Assert(err, IsNil)
	defer func() {
		err = token.Delete()
		check.Assert(err, IsNil)
	}()
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTokens+token.Token.ID)

	apiToken, err := token.GetInitialApiToken()
	check.Assert(err, IsNil)

	cseVersion, err := semver.NewVersion(vcd.config.Cse.Version)
	check.Assert(err, IsNil)
	check.Assert(cseVersion, NotNil)

	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCrCI+QkLjgQVqR7c7dJfawJqCslVomo5I25JdolqlteX7RCUq0yncWyS+8MTYWCS03sm1jOroLOeuji8CDKCDCcKwQerJiOFoJS+VOK5xCjJ2u8RBGlIpXNcmIh2VriRJrV7TCKrFMSKLNF4/n83q4gWI/YPf6/dRhpPB72HYrdI4omvRlU4GG09jMmgiz+5Yb8wJEXYMsJni+MwPzFKe6TbMcqjBusDyeFGAhgyN7QJGpdNhAn1sqvqZrW2QjaE8P+4t8RzBo8B2ucyQazd6+lbYmOHq9366LjG160snzXrFzlARc4hhpjMzu9Bcm6i3ZZI70qhIbmi5IonbbVh8t"
	// Create the cluster
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
		SshPublicKey:       sshPublicKey,
		AutoRepairOnErrors: true,
	}
	cluster, err := org.CseCreateKubernetesCluster(clusterSettings, 150*time.Minute)

	// We assure that the cluster gets always deleted, even if the creation failed.
	// Deletion process only needs the cluster ID
	defer func() {
		check.Assert(cluster, NotNil)
		check.Assert(cluster.client, NotNil)
		check.Assert(cluster.ID, Not(Equals), "")
		err = cluster.Delete(0)
		check.Assert(err, IsNil)
	}()

	check.Assert(err, IsNil)
	assertCseClusterCreation(check, cluster, clusterSettings, tkgBundle)

	kubeconfig, err := cluster.GetKubeconfig(false)
	check.Assert(err, IsNil)
	check.Assert(true, Equals, strings.Contains(kubeconfig, cluster.Name))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-certificate-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "certificate-authority-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-key-data"))

	err = cluster.Refresh()
	check.Assert(err, IsNil)

	clusterGet, err := vcd.client.CseGetKubernetesClusterById(cluster.ID)
	check.Assert(err, IsNil)
	assertCseClusterEquals(check, clusterGet, cluster)
	check.Assert(clusterGet.Etag, Not(Equals), "")

	allClusters, err := org.CseGetKubernetesClustersByName(clusterGet.CseVersion, clusterGet.Name)
	check.Assert(err, IsNil)
	check.Assert(len(allClusters), Equals, 1)
	assertCseClusterEquals(check, allClusters[0], clusterGet)
	check.Assert(allClusters[0].Etag, Equals, "") // Can't recover ETag by name

	// Update worker pool with autoscaler
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{clusterSettings.WorkerPools[0].Name: {
		Autoscaler: &CseWorkerPoolAutoscaler{
			MaxSize: 2,
			MinSize: 1,
		}}}, true)
	check.Assert(err, IsNil)
	foundWorkerPool := false
	for _, nodePool := range cluster.WorkerPools {
		if nodePool.Name == clusterSettings.WorkerPools[0].Name {
			foundWorkerPool = true
			check.Assert(nodePool.MachineCount, Equals, 0) // The field is not used
			check.Assert(nodePool.Autoscaler, NotNil)
			check.Assert(nodePool.Autoscaler.MaxSize, Equals, 2)
			check.Assert(nodePool.Autoscaler.MinSize, Equals, 1)
		}
	}
	check.Assert(foundWorkerPool, Equals, true)

	// Update worker pool from Autoscaling to static 2 nodes
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{clusterSettings.WorkerPools[0].Name: {MachineCount: 2}}, true)
	check.Assert(err, IsNil)
	foundWorkerPool = false
	for _, nodePool := range cluster.WorkerPools {
		if nodePool.Name == clusterSettings.WorkerPools[0].Name {
			foundWorkerPool = true
			check.Assert(nodePool.MachineCount, Equals, 2)
			check.Assert(nodePool.Autoscaler, IsNil) // Autoscaler should be deactivated
		}
	}
	check.Assert(foundWorkerPool, Equals, true)

	// Add two new worker pools, one with autoscaler
	err = cluster.AddWorkerPools([]CseWorkerPoolSettings{{
		Name:         "new-pool-1",
		MachineCount: 1,
		DiskSizeGi:   20,
	}, {
		Name:       "new-pool-2",
		DiskSizeGi: 20,
		Autoscaler: &CseWorkerPoolAutoscaler{
			MaxSize: 2,
			MinSize: 1,
		},
	}}, true)
	check.Assert(err, IsNil)
	foundWorkerPool1, foundWorkerPool2 := false, false
	for _, nodePool := range cluster.WorkerPools {
		if nodePool.Name == "new-pool-1" {
			foundWorkerPool1 = true
			check.Assert(nodePool.MachineCount, Equals, 1)
			check.Assert(nodePool.DiskSizeGi, Equals, 20)
			check.Assert(nodePool.SizingPolicyId, Equals, "")
			check.Assert(nodePool.StorageProfileId, Equals, "")
			check.Assert(nodePool.Autoscaler, IsNil)
		}
		if nodePool.Name == "new-pool-2" {
			foundWorkerPool2 = true
			check.Assert(nodePool.MachineCount, Equals, 0) // Not used
			check.Assert(nodePool.DiskSizeGi, Equals, 20)
			check.Assert(nodePool.SizingPolicyId, Equals, "")
			check.Assert(nodePool.StorageProfileId, Equals, "")
			check.Assert(nodePool.Autoscaler, NotNil)
			check.Assert(nodePool.Autoscaler.MinSize, Equals, 1)
			check.Assert(nodePool.Autoscaler.MaxSize, Equals, 2)
		}
	}
	check.Assert(foundWorkerPool1, Equals, true)
	check.Assert(foundWorkerPool2, Equals, true)

	// Update control plane from 1 node to 3 (needs to be an odd number)
	err = cluster.UpdateControlPlane(CseControlPlaneUpdateInput{MachineCount: 3}, true)
	check.Assert(err, IsNil)
	check.Assert(cluster.ControlPlane.MachineCount, Equals, 3)

	// Turn off the node health check
	err = cluster.SetNodeHealthCheck(false, true)
	check.Assert(err, IsNil)
	check.Assert(cluster.NodeHealthCheck, Equals, false)

	// Update the auto repair flag
	check.Assert(err, IsNil)
	err = cluster.SetAutoRepairOnErrors(false, true)
	check.Assert(err, IsNil) // It won't fail in CSE >4.1.0 as the flag is already false, so we update nothing.
	check.Assert(cluster.AutoRepairOnErrors, Equals, false)

	// Upgrade the cluster if possible
	upgradeOvas, err := cluster.GetSupportedUpgrades(true)
	check.Assert(err, IsNil)
	if len(upgradeOvas) > 0 {
		err = cluster.UpgradeCluster(upgradeOvas[0].ID, true)
		check.Assert(err, IsNil)
		check.Assert(cluster.KubernetesVersion, Not(Equals), clusterGet.KubernetesVersion)
		check.Assert(cluster.TkgVersion, Not(Equals), clusterGet.TkgVersion)
		check.Assert(cluster.KubernetesTemplateOvaId, Not(Equals), clusterGet.KubernetesTemplateOvaId)
		upgradeOvas, err = cluster.GetSupportedUpgrades(true)
		check.Assert(err, IsNil)
		check.Assert(len(upgradeOvas), Equals, 0)
	} else {
		fmt.Println("WARNING: CseKubernetesCluster.UpgradeCluster method not tested. It was skipped as there's no OVA to upgrade the cluster")
	}

	// Helps to delete the cluster faster, also tests generic update method
	err = cluster.Update(CseClusterUpdateInput{
		ControlPlane: &CseControlPlaneUpdateInput{MachineCount: 1},
		WorkerPools: &map[string]CseWorkerPoolUpdateInput{
			clusterSettings.WorkerPools[0].Name: {
				MachineCount: 1,
			},
			"new-pool-1": {
				MachineCount: 0,
			},
			"new-pool-2": {
				MachineCount: 0, // Should remove autoscaler
			},
		},
	}, true)
	check.Assert(err, IsNil)
	check.Assert(cluster.ControlPlane.MachineCount, Equals, 1)
	for _, pool := range cluster.WorkerPools {
		if pool.Name == "new-pool-1" {
			check.Assert(pool.MachineCount, Equals, 0)
			check.Assert(pool.Autoscaler, IsNil)
		} else if pool.Name == "new-pool-2" {
			check.Assert(pool.MachineCount, Equals, 0)
			check.Assert(pool.Autoscaler, IsNil)
		} else {
			check.Assert(pool.MachineCount, Equals, 1)
		}
	}
}

// Test_CseWithAutoscaler tests the autoscaling capabilities in CSE clusters
func (vcd *TestVCD) Test_CseWithAutoscaler(check *C) {
	requireCseConfig(check, vcd.config)

	// Prerequisites: We need to read several items before creating the cluster.
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

	sp, err := vdc.FindStorageProfileReference(vcd.config.Cse.StorageProfile)
	check.Assert(err, IsNil)

	policies, err := vcd.client.GetAllVdcComputePoliciesV2(url.Values{
		"filter": []string{"name==TKG small"},
	})
	check.Assert(err, IsNil)
	check.Assert(len(policies), Equals, 1)

	token, err := vcd.client.CreateToken(vcd.config.Provider.SysOrg, check.TestName())
	check.Assert(err, IsNil)
	defer func() {
		err = token.Delete()
		check.Assert(err, IsNil)
	}()
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTokens+token.Token.ID)

	apiToken, err := token.GetInitialApiToken()
	check.Assert(err, IsNil)

	cseVersion, err := semver.NewVersion(vcd.config.Cse.Version)
	check.Assert(err, IsNil)
	check.Assert(cseVersion, NotNil)

	// Create the cluster
	clusterSettings := CseClusterSettings{
		Name:                    "test-cse-autoscaler",
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
			Name: "worker-pool-1",
			Autoscaler: &CseWorkerPoolAutoscaler{
				MaxSize: 2,
				MinSize: 1,
			},
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
	cluster, err := org.CseCreateKubernetesCluster(clusterSettings, 150*time.Minute)

	// We assure that the cluster gets always deleted, even if the creation failed.
	// Deletion process only needs the cluster ID
	defer func() {
		check.Assert(cluster, NotNil)
		check.Assert(cluster.client, NotNil)
		check.Assert(cluster.ID, Not(Equals), "")
		err = cluster.Delete(0)
		check.Assert(err, IsNil)
	}()

	check.Assert(err, IsNil)
	assertCseClusterCreation(check, cluster, clusterSettings, tkgBundle)

	kubeconfig, err := cluster.GetKubeconfig(false)
	check.Assert(err, IsNil)
	check.Assert(true, Equals, strings.Contains(kubeconfig, cluster.Name))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-certificate-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "certificate-authority-data"))
	check.Assert(true, Equals, strings.Contains(kubeconfig, "client-key-data"))

	err = cluster.Refresh()
	check.Assert(err, IsNil)

	clusterGet, err := vcd.client.CseGetKubernetesClusterById(cluster.ID)
	check.Assert(err, IsNil)
	assertCseClusterEquals(check, clusterGet, cluster)
	check.Assert(clusterGet.Etag, Not(Equals), "")

	allClusters, err := org.CseGetKubernetesClustersByName(clusterGet.CseVersion, clusterGet.Name)
	check.Assert(err, IsNil)
	check.Assert(len(allClusters), Equals, 1)
	assertCseClusterEquals(check, allClusters[0], clusterGet)
	check.Assert(allClusters[0].Etag, Equals, "") // Can't recover ETag by name

	// Update worker pool and deactivate autoscaler
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{clusterSettings.WorkerPools[0].Name: {
		Autoscaler: &CseWorkerPoolAutoscaler{
			MaxSize: 10,
			MinSize: 1,
		}}}, true)
	check.Assert(err, IsNil)
	foundWorkerPool := false
	for _, nodePool := range cluster.WorkerPools {
		if nodePool.Name == clusterSettings.WorkerPools[0].Name {
			foundWorkerPool = true
			check.Assert(nodePool.MachineCount, Equals, 0) // The field is not used
			check.Assert(nodePool.Autoscaler, NotNil)
			check.Assert(nodePool.Autoscaler.MinSize, Equals, 1)
			check.Assert(nodePool.Autoscaler.MaxSize, Equals, 10)
		}
	}
	check.Assert(foundWorkerPool, Equals, true)
}

// Test_CseFailure tests cluster creation errors and their consequences
func (vcd *TestVCD) Test_CseFailure(check *C) {
	requireCseConfig(check, vcd.config)

	// Prerequisites: We need to read several items before creating the cluster.
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

	sp, err := vdc.FindStorageProfileReference(vcd.config.Cse.StorageProfile)
	check.Assert(err, IsNil)

	policies, err := vcd.client.GetAllVdcComputePoliciesV2(url.Values{
		"filter": []string{"name==TKG small"},
	})
	check.Assert(err, IsNil)
	check.Assert(len(policies), Equals, 1)

	token, err := vcd.client.CreateToken(vcd.config.Provider.SysOrg, check.TestName())
	check.Assert(err, IsNil)
	defer func() {
		err = token.Delete()
		check.Assert(err, IsNil)
	}()
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTokens+token.Token.ID)

	apiToken, err := token.GetInitialApiToken()
	check.Assert(err, IsNil)

	cseVersion, err := semver.NewVersion(vcd.config.Cse.Version)
	check.Assert(err, IsNil)
	check.Assert(cseVersion, NotNil)

	componentsVersions, err := getCseComponentsVersions(*cseVersion)
	check.Assert(err, IsNil)
	check.Assert(componentsVersions, NotNil)

	// Create the cluster
	clusterSettings := CseClusterSettings{
		Name:                    "test-cse-fail",
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
		Owner:              vcd.config.Provider.User,
		ApiToken:           apiToken.RefreshToken,
		NodeHealthCheck:    true,
		PodCidr:            "1.1.1.1/24", // This should make the cluster fail
		ServiceCidr:        "1.1.1.1/24", // This should make the cluster fail
		AutoRepairOnErrors: false,        // Must be false to avoid never-ending loops
	}
	cluster, err := org.CseCreateKubernetesCluster(clusterSettings, 150*time.Minute)

	// We assure that the cluster gets always deleted.
	// Deletion process only needs the cluster ID
	defer func() {
		check.Assert(cluster, NotNil)
		check.Assert(cluster.client, NotNil)
		check.Assert(cluster.ID, Not(Equals), "")
		err = cluster.Delete(0)
		check.Assert(err, IsNil)
	}()

	check.Assert(err, NotNil)
	check.Assert(cluster.client, NotNil)
	check.Assert(cluster.ID, Not(Equals), "")

	clusterGet, err := vcd.client.CseGetKubernetesClusterById(cluster.ID)
	check.Assert(err, IsNil)
	// We don't get an error when we retrieve a failed cluster, but some fields are missing
	check.Assert(clusterGet.ID, Equals, cluster.ID)
	check.Assert(clusterGet.Etag, Not(Equals), "")
	check.Assert(clusterGet.State, Equals, "error")
	check.Assert(len(clusterGet.Events), Not(Equals), 0)

	err = cluster.Refresh()
	check.Assert(err, IsNil)
	assertCseClusterEquals(check, cluster, clusterGet)

	allClusters, err := org.CseGetKubernetesClustersByName(clusterGet.CseVersion, clusterGet.Name)
	check.Assert(err, IsNil)
	check.Assert(len(allClusters), Equals, 1)
	assertCseClusterEquals(check, allClusters[0], clusterGet)
	check.Assert(allClusters[0].Etag, Equals, "") // Can't recover ETag by name

	_, err = cluster.GetKubeconfig(false)
	check.Assert(err, NotNil)

	// All updates should fail
	err = cluster.UpdateWorkerPools(map[string]CseWorkerPoolUpdateInput{clusterSettings.WorkerPools[0].Name: {MachineCount: 1}}, true)
	check.Assert(err, NotNil)
	err = cluster.AddWorkerPools([]CseWorkerPoolSettings{{
		Name:         "i-dont-care-i-will-fail",
		MachineCount: 1,
		DiskSizeGi:   20,
	}}, true)
	check.Assert(err, NotNil)
	err = cluster.UpdateControlPlane(CseControlPlaneUpdateInput{MachineCount: 1}, true)
	check.Assert(err, NotNil)
	err = cluster.SetNodeHealthCheck(false, true)
	check.Assert(err, NotNil)
	err = cluster.SetAutoRepairOnErrors(false, true)
	check.Assert(err, NotNil)

	upgradeOvas, err := cluster.GetSupportedUpgrades(true)
	check.Assert(err, IsNil)
	check.Assert(len(upgradeOvas), Equals, 0)

	err = cluster.UpgradeCluster(clusterSettings.KubernetesTemplateOvaId, true)
	check.Assert(err, NotNil)
}

// Test_CseValidationErrors tests validation errors during cluster creation request
func (vcd *TestVCD) Test_CseValidationErrors(check *C) {
	requireCseConfig(check, vcd.config)

	org, err := vcd.client.GetOrgByName(vcd.config.Cse.TenantOrg)
	check.Assert(err, IsNil)

	settings := CseClusterSettings{}

	// Wrong CSE version
	cseVersion, err := semver.NewVersion("9.0.0")
	check.Assert(err, IsNil)
	check.Assert(cseVersion, NotNil)
	settings.CseVersion = *cseVersion
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("the Container Service Extension version '%s' is not supported", settings.CseVersion.String()), Equals, true)
	cseVersion, err = semver.NewVersion(vcd.config.Cse.Version)
	check.Assert(err, IsNil)
	check.Assert(cseVersion, NotNil)
	settings.CseVersion = *cseVersion

	// Wrong name
	settings.Name = "NotAValidName%%%1"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the name '%s' must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters", settings.Name), Equals, true)

	settings.Name = "valid"

	// Missing Organization ID
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the Organization ID is required", Equals, true)

	settings.OrganizationId = org.Org.ID

	// Missing VDC ID
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the VDC ID is required", Equals, true)

	vdc, err := org.GetVDCByName(vcd.config.Cse.TenantVdc, false)
	check.Assert(err, IsNil)
	settings.VdcId = vdc.Vdc.ID

	// Missing Network ID
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the Network ID is required", Equals, true)

	net, err := vdc.GetOrgVdcNetworkByName(vcd.config.Cse.RoutedNetwork, false)
	check.Assert(err, IsNil)
	settings.NetworkId = net.OrgVDCNetwork.ID

	// Missing Kubernetes OVA
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the Kubernetes Template OVA ID is required", Equals, true)

	catalog, err := org.GetCatalogByName(vcd.config.Cse.OvaCatalog, false)
	check.Assert(err, IsNil)
	ova, err := catalog.GetVAppTemplateByName(vcd.config.Cse.OvaName)
	check.Assert(err, IsNil)
	settings.KubernetesTemplateOvaId = ova.VAppTemplate.ID

	// No control plane nodes
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: number of Control Plane nodes must be odd and higher than 0, but it was '0'", Equals, true)

	settings.ControlPlane.MachineCount = 2

	// Wrong control plane nodes, it should not be even
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: number of Control Plane nodes must be odd and higher than 0, but it was '2'", Equals, true)

	settings.ControlPlane.MachineCount = 1

	// Wrong disk size for the control plane
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: disk size for the Control Plane in Gibibytes (Gi) must be at least 20, but it was '0'", Equals, true)

	settings.ControlPlane.DiskSizeGi = 20

	// No worker pool
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: there must be at least one Worker Pool", Equals, true)

	settings.WorkerPools = []CseWorkerPoolSettings{
		{},
	}

	// Wrong worker pool name
	settings.WorkerPools[0].Name = "NotAValidName%%%1"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the Worker Pool name '%s' must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters", settings.WorkerPools[0].Name), Equals, true)

	settings.WorkerPools[0].Name = "wp-1"

	// No worker pool replicas
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: number of Worker Pool '%s' nodes must higher than 0, but it was '0'", settings.WorkerPools[0].Name), Equals, true)

	settings.WorkerPools[0].MachineCount = 1

	// Try to set the autoscaler and the static machine count at same time
	settings.WorkerPools[0].Autoscaler = &CseWorkerPoolAutoscaler{MaxSize: 1, MinSize: 5}
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the Worker Pool '%s' is using Autoscaler (min=5,max=1), so can't set MachineCount to '1'", settings.WorkerPools[0].Name), Equals, true)

	// The autoscaler is configured wrong (min > max)
	settings.WorkerPools[0].MachineCount = 0
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the Autoscaler maximum size for Worker Pool '%s' cannot be less than the minimum", settings.WorkerPools[0].Name), Equals, true)

	// The autoscaler is configured wrong (max < 0)
	settings.WorkerPools[0].Autoscaler.MaxSize = -5
	settings.WorkerPools[0].Autoscaler.MinSize = -10
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the Autoscaler maximum size for Worker Pool '%s' must be a positive number", settings.WorkerPools[0].Name), Equals, true)

	// The autoscaler is configured wrong (min < 0)
	settings.WorkerPools[0].Autoscaler.MaxSize = 5
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the Autoscaler minimum size for Worker Pool '%s' must be a positive number", settings.WorkerPools[0].Name), Equals, true)

	// Wrong disk size for the worker pool
	settings.WorkerPools[0].Autoscaler.MinSize = 1
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: disk size for the Worker Pool '%s' in Gibibytes (Gi) must be at least 20, but it was '0'", settings.WorkerPools[0].Name), Equals, true)

	settings.WorkerPools[0].DiskSizeGi = 20
	settings.WorkerPools = append(settings.WorkerPools, CseWorkerPoolSettings{Name: "wp-1"})

	// Repeated worker pool name
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the names of the Worker Pools must be unique, but '%s' is repeated", settings.WorkerPools[0].Name), Equals, true)

	settings.WorkerPools[1] = CseWorkerPoolSettings{Name: "wp-2", MachineCount: 1, DiskSizeGi: 20}

	// Missing API token
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the API token is required", Equals, true)

	token, err := vcd.client.CreateToken(vcd.config.Provider.SysOrg, check.TestName())
	check.Assert(err, IsNil)
	defer func() {
		err = token.Delete()
		check.Assert(err, IsNil)
	}()
	AddToCleanupListOpenApi(token.Token.Name, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTokens+token.Token.ID)

	apiToken, err := token.GetInitialApiToken()
	check.Assert(err, IsNil)
	settings.ApiToken = apiToken.RefreshToken

	// Missing Pod CIDR
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the Pod CIDR is required", Equals, true)

	// Wrong Pod CIDR
	settings.PodCidr = "256.700.1.278/800"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "error creating the CSE Kubernetes cluster: the Pod CIDR is malformed"), Equals, true)

	// Missing Service CIDR
	settings.PodCidr = "192.168.1.0/20"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the Service CIDR is required", Equals, true)

	// Wrong Service CIDR
	settings.ServiceCidr = "256.700.1.278/800"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "error creating the CSE Kubernetes cluster: the Service CIDR is malformed"), Equals, true)

	// Wrong Virtual IP subnet
	settings.ServiceCidr = "192.168.1.0/20"
	settings.VirtualIpSubnet = "256.700.1.278/800"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "error creating the CSE Kubernetes cluster: the Virtual IP Subnet is malformed"), Equals, true)

	// Wrong Control Plane IP
	settings.VirtualIpSubnet = "192.154.1.0/20"
	settings.ControlPlane.Ip = "256.700.1.278"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "error creating the CSE Kubernetes cluster: the Control Plane IP is malformed"), Equals, true)

	// Wrong default storage class name
	settings.ControlPlane.Ip = "1.1.1.1"
	settings.DefaultStorageClass = &CseDefaultStorageClassSettings{Name: "NotAValidName%%%1"}
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == fmt.Sprintf("error creating the CSE Kubernetes cluster: the Default Storage Class name '%s' must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters", settings.DefaultStorageClass.Name), Equals, true)

	// Missing Storage profile ID
	settings.DefaultStorageClass.Name = "sp-1"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the Storage Profile ID for the Default Storage Class is required", Equals, true)

	sp, err := vdc.FindStorageProfileReference(vcd.config.Cse.StorageProfile)
	check.Assert(err, IsNil)

	policies, err := vcd.client.GetAllVdcComputePoliciesV2(url.Values{
		"filter": []string{"name==TKG small"},
	})
	check.Assert(err, IsNil)
	check.Assert(len(policies), Equals, 1)

	// Wrong retaining policy in the default storage class
	settings.DefaultStorageClass.StorageProfileId = sp.ID
	settings.DefaultStorageClass.ReclaimPolicy = "whatever"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the Reclaim Policy for the Default Storage Class must be either 'delete' or 'retain', but it was 'whatever'", Equals, true)

	// Wrong filesystem in the default storage class
	settings.DefaultStorageClass.ReclaimPolicy = "delete"
	settings.DefaultStorageClass.Filesystem = "oops"
	_, err = org.CseCreateKubernetesCluster(settings, 0)
	check.Assert(err, NotNil)
	check.Assert(err.Error() == "error creating the CSE Kubernetes cluster: the filesystem for the Default Storage Class must be either 'ext4' or 'xfs', but it was 'oops'", Equals, true)
}

func assertCseClusterCreation(check *C, createdCluster *CseKubernetesCluster, settings CseClusterSettings, expectedKubernetesData tkgVersionBundle) {
	check.Assert(createdCluster, NotNil)
	check.Assert(createdCluster.CseVersion.Original(), Equals, settings.CseVersion.Original())
	check.Assert(createdCluster.Name, Equals, settings.Name)
	check.Assert(createdCluster.OrganizationId, Equals, settings.OrganizationId)
	check.Assert(createdCluster.VdcId, Equals, settings.VdcId)
	check.Assert(createdCluster.NetworkId, Equals, settings.NetworkId)
	check.Assert(createdCluster.KubernetesTemplateOvaId, Equals, settings.KubernetesTemplateOvaId)
	check.Assert(createdCluster.ControlPlane.MachineCount, Equals, settings.ControlPlane.MachineCount)
	check.Assert(createdCluster.ControlPlane.SizingPolicyId, Equals, settings.ControlPlane.SizingPolicyId)
	check.Assert(createdCluster.ControlPlane.PlacementPolicyId, Equals, settings.ControlPlane.PlacementPolicyId)
	check.Assert(createdCluster.ControlPlane.StorageProfileId, Equals, settings.ControlPlane.StorageProfileId)
	check.Assert(createdCluster.ControlPlane.DiskSizeGi, Equals, settings.ControlPlane.DiskSizeGi)
	if settings.ControlPlane.Ip != "" {
		check.Assert(createdCluster.ControlPlane.Ip, Equals, settings.ControlPlane.Ip)
	} else {
		check.Assert(createdCluster.ControlPlane.Ip, Not(Equals), "")
	}
	check.Assert(createdCluster.WorkerPools, DeepEquals, settings.WorkerPools)
	if settings.DefaultStorageClass != nil {
		check.Assert(createdCluster.DefaultStorageClass, NotNil)
		check.Assert(*createdCluster.DefaultStorageClass, DeepEquals, *settings.DefaultStorageClass)
	}
	if settings.Owner != "" {
		check.Assert(createdCluster.Owner, Equals, settings.Owner)
	} else {
		check.Assert(createdCluster.Owner, Not(Equals), "")
	}
	check.Assert(createdCluster.ApiToken, Not(Equals), settings.ApiToken)
	check.Assert(createdCluster.ApiToken, Equals, "******") // This one can't be recovered
	check.Assert(createdCluster.NodeHealthCheck, Equals, settings.NodeHealthCheck)
	check.Assert(createdCluster.PodCidr, Equals, settings.PodCidr)
	check.Assert(createdCluster.ServiceCidr, Equals, settings.ServiceCidr)
	check.Assert(createdCluster.SshPublicKey, Equals, settings.SshPublicKey)
	check.Assert(createdCluster.VirtualIpSubnet, Equals, settings.VirtualIpSubnet)
	check.Assert(createdCluster.SshPublicKey, Equals, settings.SshPublicKey)

	v411, err := semver.NewVersion("4.1.1")
	check.Assert(err, IsNil)
	if settings.CseVersion.GreaterThanOrEqual(v411) {
		// Since CSE 4.1.1, the flag is automatically switched off when the cluster is created
		check.Assert(createdCluster.AutoRepairOnErrors, Equals, false)
	} else {
		check.Assert(createdCluster.AutoRepairOnErrors, Equals, settings.AutoRepairOnErrors)
	}
	check.Assert(createdCluster.VirtualIpSubnet, Equals, settings.VirtualIpSubnet)
	check.Assert(true, Equals, strings.Contains(createdCluster.ID, "urn:vcloud:entity:vmware:capvcdCluster:"))
	check.Assert(createdCluster.Etag, Not(Equals), "")
	check.Assert(createdCluster.KubernetesVersion.Original(), Equals, expectedKubernetesData.KubernetesVersion)
	check.Assert(createdCluster.TkgVersion.Original(), Equals, expectedKubernetesData.TkgVersion)
	check.Assert(createdCluster.CapvcdVersion.Original(), Not(Equals), "")
	check.Assert(createdCluster.CpiVersion.Original(), Not(Equals), "")
	check.Assert(createdCluster.CsiVersion.Original(), Not(Equals), "")
	check.Assert(len(createdCluster.ClusterResourceSetBindings), Not(Equals), 0)
	check.Assert(createdCluster.State, Equals, "provisioned")
	check.Assert(len(createdCluster.Events), Not(Equals), 0)
}

func assertCseClusterEquals(check *C, obtainedCluster, expectedCluster *CseKubernetesCluster) {
	check.Assert(expectedCluster, NotNil)
	check.Assert(obtainedCluster, NotNil)
	check.Assert(obtainedCluster.CseVersion.Original(), Equals, expectedCluster.CseVersion.Original())
	check.Assert(obtainedCluster.Name, Equals, expectedCluster.Name)
	check.Assert(obtainedCluster.OrganizationId, Equals, expectedCluster.OrganizationId)
	check.Assert(obtainedCluster.VdcId, Equals, expectedCluster.VdcId)
	check.Assert(obtainedCluster.NetworkId, Equals, expectedCluster.NetworkId)
	check.Assert(obtainedCluster.KubernetesTemplateOvaId, Equals, expectedCluster.KubernetesTemplateOvaId)
	check.Assert(obtainedCluster.ControlPlane, DeepEquals, expectedCluster.ControlPlane)
	check.Assert(obtainedCluster.WorkerPools, DeepEquals, expectedCluster.WorkerPools)
	if expectedCluster.DefaultStorageClass != nil {
		check.Assert(obtainedCluster.DefaultStorageClass, NotNil)
		check.Assert(*obtainedCluster.DefaultStorageClass, DeepEquals, *expectedCluster.DefaultStorageClass)
	}
	check.Assert(obtainedCluster.Owner, Equals, expectedCluster.Owner)
	check.Assert(obtainedCluster.ApiToken, Equals, "******") // This one can't be recovered
	check.Assert(obtainedCluster.NodeHealthCheck, Equals, expectedCluster.NodeHealthCheck)
	check.Assert(obtainedCluster.PodCidr, Equals, expectedCluster.PodCidr)
	check.Assert(obtainedCluster.ServiceCidr, Equals, expectedCluster.ServiceCidr)
	check.Assert(obtainedCluster.SshPublicKey, Equals, expectedCluster.SshPublicKey)
	check.Assert(obtainedCluster.VirtualIpSubnet, Equals, expectedCluster.VirtualIpSubnet)
	check.Assert(obtainedCluster.AutoRepairOnErrors, Equals, expectedCluster.AutoRepairOnErrors)
	check.Assert(obtainedCluster.VirtualIpSubnet, Equals, expectedCluster.VirtualIpSubnet)
	check.Assert(obtainedCluster.ID, Equals, expectedCluster.ID)
	check.Assert(obtainedCluster.KubernetesVersion.Original(), Equals, expectedCluster.KubernetesVersion.Original())
	check.Assert(obtainedCluster.TkgVersion.Original(), Equals, expectedCluster.TkgVersion.Original())
	check.Assert(obtainedCluster.CapvcdVersion.Original(), Equals, expectedCluster.CapvcdVersion.Original())
	check.Assert(obtainedCluster.CpiVersion.Original(), Equals, expectedCluster.CpiVersion.Original())
	check.Assert(obtainedCluster.CsiVersion.Original(), Equals, expectedCluster.CsiVersion.Original())
	check.Assert(obtainedCluster.ClusterResourceSetBindings, DeepEquals, expectedCluster.ClusterResourceSetBindings)
	check.Assert(obtainedCluster.State, Equals, expectedCluster.State)
	check.Assert(len(obtainedCluster.Events) >= len(expectedCluster.Events), Equals, true)
}
