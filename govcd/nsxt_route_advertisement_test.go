//go:build network || nsxt || functional || openapi || ALL
// +build network nsxt functional openapi ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtEdgeRouteAdvertisement(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointNsxtRouteAdvertisement)

	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	nsxtVdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	edge, err := nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Make sure we are using a dedicated Tier-0 gateway (otherwise route advertisement won't be available)
	edge.EdgeGateway.EdgeGatewayUplinks[0].Dedicated = true
	edge, err = edge.Update(edge.EdgeGateway)
	check.Assert(err, IsNil)
	check.Assert(edge, NotNil)

	network1 := "192.168.1.0/24"
	network2 := "192.168.2.0/24"
	networksToAdvertise := []string{network1, network2} // Sample networks to advertise

	// Test UpdateNsxtRouteAdvertisement
	nsxtEdgeRouteAdvertisement, err := edge.UpdateNsxtRouteAdvertisement(true, networksToAdvertise, true)
	check.Assert(err, IsNil)
	check.Assert(nsxtEdgeRouteAdvertisement, NotNil)
	check.Assert(nsxtEdgeRouteAdvertisement.Enable, Equals, true)
	check.Assert(len(nsxtEdgeRouteAdvertisement.Subnets), Equals, 2)
	check.Assert(checkNetworkInSubnetsSlice(network1, networksToAdvertise), IsNil)
	check.Assert(checkNetworkInSubnetsSlice(network2, networksToAdvertise), IsNil)

	// Test DeleteNsxtRouteAdvertisement
	err = edge.DeleteNsxtRouteAdvertisement(true)
	check.Assert(err, IsNil)
	nsxtEdgeRouteAdvertisement, err = edge.GetNsxtRouteAdvertisement(true)
	check.Assert(err, IsNil)
	check.Assert(nsxtEdgeRouteAdvertisement, NotNil)
	check.Assert(nsxtEdgeRouteAdvertisement.Enable, Equals, false)
	check.Assert(len(nsxtEdgeRouteAdvertisement.Subnets), Equals, 0)
}

func checkNetworkInSubnetsSlice(network string, subnets []string) error {
	for _, subnet := range subnets {
		if subnet == network {
			return nil
		}
	}
	return fmt.Errorf("network %s is not within the slice provided", network)
}
