//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VdcNetworkProfile(check *C) {
	vcd.skipIfNotSysAdmin(check)
	skipNoNsxtConfiguration(vcd, check)
	if vcd.config.VCD.Nsxt.NsxtEdgeCluster == "" {
		check.Skip("missing value for vcd.config.VCD.Nsxt.NsxtEdgeCluster")
	}

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	existingVdcNetworkProfile, err := nsxtVdc.GetVdcNetworkProfile()
	check.Assert(err, IsNil)
	check.Assert(existingVdcNetworkProfile, NotNil)

	// Lookup Edge available Edge Cluster
	edgeCluster, err := nsxtVdc.GetNsxtEdgeClusterByName(vcd.config.VCD.Nsxt.NsxtEdgeCluster)
	check.Assert(err, IsNil)
	check.Assert(edgeCluster, NotNil)

	networkProfileConfig := &types.VdcNetworkProfile{
		ServicesEdgeCluster: &types.VdcNetworkProfileServicesEdgeCluster{
			BackingID: edgeCluster.NsxtEdgeCluster.ID,
		},
	}

	newVdcNetworkProfile, err := nsxtVdc.UpdateVdcNetworkProfile(networkProfileConfig)
	check.Assert(err, IsNil)
	check.Assert(newVdcNetworkProfile, NotNil)
	check.Assert(newVdcNetworkProfile.ServicesEdgeCluster.BackingID, Equals, edgeCluster.NsxtEdgeCluster.ID)

	// Unset Edge Cluster (and other values) by sending empty structure
	unsetNetworkProfileConfig := &types.VdcNetworkProfile{}
	unsetVdcNetworkProfile, err := nsxtVdc.UpdateVdcNetworkProfile(unsetNetworkProfileConfig)
	check.Assert(err, IsNil)
	check.Assert(unsetVdcNetworkProfile, NotNil)

	networkProfileAfterCleanup, err := nsxtVdc.GetVdcNetworkProfile()
	check.Assert(err, IsNil)
	check.Assert(networkProfileAfterCleanup.ServicesEdgeCluster, IsNil)
	// Cleanup

	err = nsxtVdc.DeleteVdcNetworkProfile()
	check.Assert(err, IsNil)
}
