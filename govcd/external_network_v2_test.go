//go:build extnetwork || network || functional || openapi || ALL
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
	// As of 10.1.2 release it is not officially supported (support only introduced in 10.2.0) therefore skipping this test for
	// 10.1.X. 10.1.1 allowed to create it, but 10.1.2 introduced a validator and throws error.
	if vcd.client.Client.APIVCDMaxVersionIs("< 35") {
		check.Skip("NSX-T VRF-Lite backed external networks are officially supported only in 10.2.0+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router)
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtSegment(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 36") {
		check.Skip("NSX-T segment backed external networks are supported only in 10.3.0+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.NsxtImportSegment, types.ExternalNetworkBackingTypeNsxtSegment)
}

func (vcd *TestVCD) testCreateExternalNetworkV2Nsxt(check *C, backingName, backingType string) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointExternalNetworks
	skipOpenApiEndpointTest(vcd, check, endpoint)
	skipNoNsxtConfiguration(vcd, check)

	fmt.Printf("Running: %s\n", check.TestName())

	// NSX-T details
	man, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	nsxtManagerId, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", extractUuid(man[0].HREF))
	check.Assert(err, IsNil)

	backingId := getBackingIdByNameAndType(check, backingName, backingType, vcd, nsxtManagerId)

	// Create network and test CRUD capabilities
	netNsxt := testExternalNetworkV2(vcd, check.TestName(), backingType, backingId, nsxtManagerId)
	createdNet, err := CreateExternalNetworkV2(vcd.client, netNsxt)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := endpoint + createdNet.ExternalNetwork.ID
	AddToCleanupListOpenApi(createdNet.ExternalNetwork.Name, check.TestName(), openApiEndpoint)

	createdNet.ExternalNetwork.Name = check.TestName() + "changed_name"
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

// getBackingIdByNameAndType looks up Backing ID by name and type
func getBackingIdByNameAndType(check *C, backingName string, backingType string, vcd *TestVCD, nsxtManagerId string) string {
	var backingId string
	switch backingType {
	case types.ExternalNetworkBackingTypeNsxtTier0Router: // Lookup T0 router ID
		tier0RouterVrf, err := vcd.client.GetImportableNsxtTier0RouterByName(backingName, nsxtManagerId)
		check.Assert(err, IsNil)
		backingId = tier0RouterVrf.NsxtTier0Router.ID
	case types.ExternalNetworkBackingTypeNsxtSegment: // Lookup segment ID
		bareNsxtManagerId, err := getBareEntityUuid(nsxtManagerId)
		check.Assert(err, IsNil)
		filter := map[string]string{"nsxTManager": bareNsxtManagerId}

		nsxtSegment, err := vcd.client.GetFilteredNsxtImportableSwitches(filter)
		check.Assert(err, IsNil)
		backingId = nsxtSegment[0].NsxtImportableSwitch.ID
	}
	return backingId
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
	vcUuid := extractUuid(vcHref)

	vcUrn, err := BuildUrnWithUuid("urn:vcloud:vimserver:", vcUuid)
	check.Assert(err, IsNil)

	net := testExternalNetworkV2(vcd, check.TestName(), vcd.config.VCD.ExternalNetworkPortGroupType, pgs[0].MoRef, vcUrn)

	r, err := CreateExternalNetworkV2(vcd.client, net)
	check.Assert(err, IsNil)

	// Use generic "OpenApiEntity" resource cleanup type
	openApiEndpoint := endpoint + r.ExternalNetwork.ID
	AddToCleanupListOpenApi(r.ExternalNetwork.Name, check.TestName(), openApiEndpoint)

	r.ExternalNetwork.Name = check.TestName() + "changed_name"
	updatedNet, err := r.Update()
	check.Assert(err, IsNil)
	check.Assert(updatedNet.ExternalNetwork.Name, Equals, r.ExternalNetwork.Name)

	err = r.Delete()
	check.Assert(err, IsNil)
}

func testExternalNetworkV2(vcd *TestVCD, name, backingType, backingId, NetworkProviderId string) *types.ExternalNetworkV2 {
	net := &types.ExternalNetworkV2{
		ID:          "",
		Name:        name,
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
				NetworkProvider: types.NetworkProvider{
					ID: NetworkProviderId,
				},
			},
		}},
	}

	// Starting with VCD 10.2 field BackingType is deprecated in favor of BackingTypeValue, and it only accepts new values
	if vcd.client.Client.APIVCDMaxVersionIs(">= 35.0") {
		net.NetworkBackings.Values[0].BackingTypeValue = backingType
	} else {
		net.NetworkBackings.Values[0].BackingType = backingType
	}

	return net
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
