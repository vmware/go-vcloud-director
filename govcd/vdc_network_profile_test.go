//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

package govcd

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_NsxtFirewall creates 20 firewall rules with randomized parameters
func (vcd *TestVCD) Test_VdcNetworkProfile(check *C) {
	skipNoNsxtConfiguration(vcd, check)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	existingVdcNetworkProfile, err := nsxtVdc.GetVdcNetworkProfile()
	check.Assert(err, IsNil)
	check.Assert(existingVdcNetworkProfile, NotNil)

	spew.Dump(existingVdcNetworkProfile)

	// Lookup Edge available Edge Cluster
	allEdgeClusters, err := nsxtVdc.GetAllNsxtEdgeClusters(nil)
	check.Assert(err, IsNil)
	check.Assert(allEdgeClusters, NotNil)

	networkProfileConfig := &types.VdcNetworkProfile{
		ServicesEdgeCluster: &types.VdcNetworkProfileServicesEdgeCluster{
			BackingID: allEdgeClusters[0].NsxtEdgeCluster.ID,
		},
	}

	newVdcNetworkProfile, err := nsxtVdc.UpdateVdcNetworkProfile(networkProfileConfig)
	check.Assert(err, IsNil)
	check.Assert(newVdcNetworkProfile, NotNil)
	check.Assert(newVdcNetworkProfile.ServicesEdgeCluster.BackingID, Equals, allEdgeClusters[0].NsxtEdgeCluster.ID)

	// Try to unset the value

	unsetNetworkProfileConfig := &types.VdcNetworkProfile{
		ServicesEdgeCluster:                 &types.VdcNetworkProfileServicesEdgeCluster{},
		VdcNetworkSegmentProfileTemplateRef: &types.OpenApiReferenceEE{ID: ""},
	}
	spew.Dump(unsetNetworkProfileConfig)

	unsetVdcNetworkProfile, err := nsxtVdc.UpdateVdcNetworkProfile(unsetNetworkProfileConfig)
	check.Assert(err, IsNil)
	check.Assert(unsetVdcNetworkProfile, NotNil)

	// Cleanup

	// err = nsxtVdc.DeleteVdcNetworkProfile()
	// check.Assert(err, IsNil)

}
