// +build extnetwork network functional openapi ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_CreateExternalNetworkV2Nsxt(check *C) {
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0router, types.ExternalNetworkBackingTypeNsxtTier0Router)
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtVrf(check *C) {
	// The documented backing type of NSX-T VRF router is "NSXT_VRF_TIER0" (types.ExternalNetworkBackingTypeNsxtVrfTier0Router)
	// but although it is documented - it fails and requires the same "NSXT_TIER0" (types.ExternalNetworkBackingTypeNsxtTier0Router)
	// backing type to be specified
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router)
}

func (vcd *TestVCD) testCreateExternalNetworkV2Nsxt(check *C, nsxtTier0Router, backingType string) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	skipOpenApiEndpointTest(vcd, check, endpoint)
	skipNoNsxtConfiguration(vcd, check)

	fmt.Printf("Running: %s\n", check.TestName())

	// NSX-T details
	man, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	nsxtManagerId, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", extractUuid(man[0].HREF))
	check.Assert(err, IsNil)

	tier0RouterVrf, err := vcd.client.GetImportableNsxtTier0RouterByName(nsxtTier0Router, nsxtManagerId)
	check.Assert(err, IsNil)

	// Create network and test CRUD capabilities
	netNsxt := testExternalNetworkV2(backingType, tier0RouterVrf.NsxtTier0Router.ID, nsxtManagerId)
	createdNet, err := CreateExternalNetworkV2(vcd.client, netNsxt)
	check.Assert(err, IsNil)

	createdNet.ExternalNetwork.Name = "changed_name"
	updatedNet, err := createdNet.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedNet.ExternalNetwork.Name, Equals, createdNet.ExternalNetwork.Name)

	read1, err := GetExternalNetworkV2ById(vcd.client, createdNet.ExternalNetwork.ID)
	check.Assert(err, IsNil)
	check.Assert(createdNet.ExternalNetwork.ID, Equals, read1.ExternalNetwork.ID)

	byName, err := GetExternalNetworkV2ByName(vcd.client, read1.ExternalNetwork.Name)
	check.Assert(err, IsNil)
	check.Assert(createdNet.ExternalNetwork.ID, Equals, byName.ExternalNetwork.ID)

	readAllNetworks, err := GetAllExternalNetworksV2(vcd.client, nil)
	check.Assert(err, IsNil)
	var foundNetwork bool
	for i := range readAllNetworks {
		if readAllNetworks[i].ExternalNetwork.ID == createdNet.ExternalNetwork.ID {
			foundNetwork = true
			break
		}
	}
	check.Assert(foundNetwork, Equals, true)

	err = createdNet.Delete()
	check.Assert(err, IsNil)

	_, err = GetExternalNetworkV2ById(vcd.client, createdNet.ExternalNetwork.ID)
	check.Assert(ContainsNotFound(err), Equals, true)
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2Nsxv(check *C) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	skipOpenApiEndpointTest(vcd, check, endpoint)

	fmt.Printf("Running: %s\n", check.TestName())

	var err error
	var pgs []*types.PortGroupRecordType

	switch vcd.config.VCD.ExternalNetworkPortGroupType {
	case types.ExternalNetworkBackingDvPortgroup:
		pgs, err = QueryDistributedPortGroup(vcd.client, vcd.config.VCD.ExternalNetworkPortGroup)
	case types.ExternalNetworkBackingTypeNetwork:
		pgs, err = QueryNetworkPortGroup(vcd.client, vcd.config.VCD.ExternalNetworkPortGroup)
	default:
		check.Errorf("unrecognized external network portgroup type: %s", vcd.config.VCD.ExternalNetworkPortGroupType)
	}
	check.Assert(err, IsNil)
	check.Assert(len(pgs), Equals, 1)

	// Query
	vcHref, err := getVcenterHref(vcd.client, vcd.config.VCD.VimServer)
	check.Assert(err, IsNil)
	vcuuid := extractUuid(vcHref)

	vcUrn, err := BuildUrnWithUuid("urn:vcloud:vimserver:", vcuuid)
	check.Assert(err, IsNil)

	neT := testExternalNetworkV2(vcd.config.VCD.ExternalNetworkPortGroupType, pgs[0].MoRef, vcUrn)

	r, err := CreateExternalNetworkV2(vcd.client, neT)
	check.Assert(err, IsNil)

	r.ExternalNetwork.Name = "changed_name"
	updatedNet, err := r.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedNet.ExternalNetwork.Name, Equals, r.ExternalNetwork.Name)

	err = r.Delete()
	check.Assert(err, IsNil)
}

func testExternalNetworkV2(backingType, backingId, NetworkProviderId string) *types.ExternalNetworkV2 {
	neT := &types.ExternalNetworkV2{
		ID:          "",
		Name:        "testNet",
		Description: "",
		Subnets: types.ExternalNetworkV2Subnets{[]types.ExternalNetworkV2Subnet{
			{
				Gateway:      "1.1.1.1",
				PrefixLength: 24,
				DNSSuffix:    "",
				DNSServer1:   "",
				DNSServer2:   "",
				IPRanges: types.ExternalNetworkV2IPRanges{[]types.ExternalNetworkV2IPRange{
					{
						StartAddress: "1.1.1.3",
						EndAddress:   "1.1.1.50",
					},
				}},
				Enabled:      true,
				UsedIPCount:  0,
				TotalIPCount: 0,
			},
		}},
		NetworkBackings: types.ExternalNetworkV2Backings{[]types.ExternalNetworkV2Backing{
			{
				BackingID: backingId,
				// Name:        tier0Router.NsxtTier0Router.DisplayName,
				BackingType: backingType,
				NetworkProvider: types.NetworkProvider{
					// Name: vcd.config.Nsxt.Manager,
					ID: NetworkProviderId,
				},
			},
		}},
	}

	return neT
}

func getVcenterHref(vcdClient *VCDClient, name string) (string, error) {
	virtualCenters, err := QueryVirtualCenters(vcdClient, fmt.Sprintf("(name==%s)", name))
	if err != nil {
		return "", err
	}
	if len(virtualCenters) == 0 || len(virtualCenters) > 1 {
		return "", fmt.Errorf("vSphere server found %d instances with name '%s' while expected one", len(virtualCenters), name)
	}
	return virtualCenters[0].HREF, nil
}
