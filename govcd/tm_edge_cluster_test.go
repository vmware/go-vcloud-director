//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmEdgeCluster(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	// Sync
	err = vcd.client.TmSyncEdgeClusters()
	check.Assert(err, IsNil)

	// Lookup
	allClusters, err := vcd.client.GetAllTmEdgeClusters(nil)
	check.Assert(err, IsNil)
	check.Assert(allClusters, NotNil)
	check.Assert(len(allClusters) > 0, Equals, true)

	ecByName, err := vcd.client.GetTmEdgeClusterByName(vcd.config.Tm.NsxtEdgeCluster)
	check.Assert(err, IsNil)
	check.Assert(ecByName, NotNil)
	check.Assert(ecByName.TmEdgeCluster.Name, Equals, vcd.config.Tm.NsxtEdgeCluster)

	ecById, err := vcd.client.GetTmEdgeClusterById(ecByName.TmEdgeCluster.ID)
	check.Assert(err, IsNil)
	check.Assert(ecById, NotNil)

	// resetting both values to 0 so that no different values are returned
	ecByName.TmEdgeCluster.AvgCPUUsagePercentage = 0
	ecByName.TmEdgeCluster.AvgMemoryUsagePercentage = 0
	ecById.TmEdgeCluster.AvgCPUUsagePercentage = 0
	ecById.TmEdgeCluster.AvgMemoryUsagePercentage = 0

	check.Assert(ecById.TmEdgeCluster, DeepEquals, ecByName.TmEdgeCluster)

	ecByNameAndRegionId, err := vcd.client.GetTmEdgeClusterByNameAndRegionId(vcd.config.Tm.NsxtEdgeCluster, region.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(ecByNameAndRegionId, NotNil)
	// resetting both values to 0 so that no different values are returned
	ecByNameAndRegionId.TmEdgeCluster.AvgCPUUsagePercentage = 0
	ecByNameAndRegionId.TmEdgeCluster.AvgMemoryUsagePercentage = 0
	check.Assert(ecByNameAndRegionId.TmEdgeCluster, DeepEquals, ecById.TmEdgeCluster)

	ecByNameAndWrongRegionId, err := vcd.client.GetTmEdgeClusterByNameAndRegionId(vcd.config.Tm.NsxtEdgeCluster, "urn:vcloud:region:167d34b3-0000-0000-0000-a388505e6102")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(ecByNameAndWrongRegionId, IsNil)

	// Qos modifications
	qosUpdate := &types.TmEdgeCluster{
		ID:        ecByName.TmEdgeCluster.ID,
		Name:      ecByName.TmEdgeCluster.Name,
		RegionRef: ecByName.TmEdgeCluster.RegionRef,
		DefaultQosConfig: types.TmEdgeClusterDefaultQosConfig{
			IngressProfile: &types.TmEdgeClusterQosProfile{
				CommittedBandwidthMbps: 100,
				BurstSizeBytes:         5,
				Type:                   "DEFAULT",
			},
			EgressProfile: &types.TmEdgeClusterQosProfile{
				CommittedBandwidthMbps: 200,
				BurstSizeBytes:         6,
				Type:                   "DEFAULT",
			},
		},
	}

	updatedQos, err := ecById.Update(qosUpdate)
	check.Assert(err, IsNil)
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps, DeepEquals, qosUpdate.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps)
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes, DeepEquals, qosUpdate.DefaultQosConfig.EgressProfile.BurstSizeBytes)
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.EgressProfile.Type, DeepEquals, qosUpdate.DefaultQosConfig.EgressProfile.Type)
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.EgressProfile.ID, Not(Equals), "")
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.EgressProfile.Name, Not(Equals), "")

	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps, DeepEquals, qosUpdate.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps)
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes, DeepEquals, qosUpdate.DefaultQosConfig.IngressProfile.BurstSizeBytes)
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.IngressProfile.Type, DeepEquals, qosUpdate.DefaultQosConfig.IngressProfile.Type)
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.IngressProfile.ID, Not(Equals), "")
	check.Assert(updatedQos.TmEdgeCluster.DefaultQosConfig.IngressProfile.Name, Not(Equals), "")

	// Remove QoS configuration
	err = updatedQos.Delete()
	check.Assert(err, IsNil)

	afterQosRemoval, err := vcd.client.GetTmEdgeClusterByName(vcd.config.Tm.NsxtEdgeCluster)
	check.Assert(err, IsNil)
	check.Assert(afterQosRemoval, NotNil)
	check.Assert(afterQosRemoval.TmEdgeCluster.DefaultQosConfig.EgressProfile, NotNil)
	check.Assert(afterQosRemoval.TmEdgeCluster.DefaultQosConfig.EgressProfile.BurstSizeBytes, Equals, -1)
	check.Assert(afterQosRemoval.TmEdgeCluster.DefaultQosConfig.EgressProfile.CommittedBandwidthMbps, Equals, -1)
	check.Assert(afterQosRemoval.TmEdgeCluster.DefaultQosConfig.IngressProfile, NotNil)
	check.Assert(afterQosRemoval.TmEdgeCluster.DefaultQosConfig.IngressProfile.BurstSizeBytes, Equals, -1)
	check.Assert(afterQosRemoval.TmEdgeCluster.DefaultQosConfig.IngressProfile.CommittedBandwidthMbps, Equals, -1)

	// Check Transport Node endpoint
	tn, err := afterQosRemoval.GetTransportNodeStatus()
	check.Assert(err, IsNil)
	check.Assert(tn, NotNil)
	check.Assert(len(tn) > 0, Equals, true)
	for i := range tn {
		check.Assert(tn[i].NodeName != "", Equals, true)
	}
}
