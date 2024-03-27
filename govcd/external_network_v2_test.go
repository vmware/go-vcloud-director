//go:build extnetwork || network || nsxt || functional || openapi || ALL

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
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0router, types.ExternalNetworkBackingTypeNsxtTier0Router, false, "", "", "")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtVrf(check *C) {
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router, false, "", "", "")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtSegment(check *C) {
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.NsxtImportSegment, types.ExternalNetworkBackingTypeNsxtSegment, false, "", "", "")
}

func (vcd *TestVCD) testCreateExternalNetworkV2Nsxt(check *C, backingName, backingType string, useIpSpace bool, ownerOrgId, natAndFwIntention, raIntention string) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
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
	netNsxt := testExternalNetworkV2(check.TestName(), backingType, backingId, nsxtManagerId, useIpSpace, ownerOrgId, natAndFwIntention, raIntention)
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
	switch {
	case backingType == types.ExternalNetworkBackingTypeNsxtTier0Router || backingType == types.ExternalNetworkBackingTypeNsxtVrfTier0Router: // Lookup T0 or T0 VRF
		tier0RouterVrf, err := vcd.client.GetImportableNsxtTier0RouterByName(backingName, nsxtManagerId)
		check.Assert(err, IsNil)
		backingId = tier0RouterVrf.NsxtTier0Router.ID
	case backingType == types.ExternalNetworkBackingTypeNsxtSegment: // Lookup segment ID
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
	vcd.skipIfNotSysAdmin(check)
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

	net := testExternalNetworkV2(check.TestName(), vcd.config.VCD.ExternalNetworkPortGroupType, pgs[0].MoRef, vcUrn, false, "", "", "")

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

func testExternalNetworkV2(name, backingType, backingId, NetworkProviderId string, useIpSpace bool, ownerOrgId, natAndFwIntention, raIntention string) *types.ExternalNetworkV2 {
	net := &types.ExternalNetworkV2{
		ID:                                 "",
		Name:                               name,
		Description:                        "",
		NatAndFirewallServiceIntention:     natAndFwIntention,
		NetworkRouteAdvertisementIntention: raIntention,
		Subnets: types.ExternalNetworkV2Subnets{Values: []types.ExternalNetworkV2Subnet{
			{
				Gateway:      "1.1.1.1",
				PrefixLength: 24,
				DNSSuffix:    "",
				DNSServer1:   "",
				DNSServer2:   "",
				IPRanges: types.ExternalNetworkV2IPRanges{Values: []types.ExternalNetworkV2IPRange{
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
		NetworkBackings: types.ExternalNetworkV2Backings{Values: []types.ExternalNetworkV2Backing{
			{
				BackingID: backingId,
				NetworkProvider: types.NetworkProvider{
					ID: NetworkProviderId,
				},
				BackingTypeValue: backingType,
			},
		}},
	}

	if useIpSpace {
		// removing subnet definition when using IP Spaces
		net.Subnets = types.ExternalNetworkV2Subnets{}
		net.UsingIpSpace = &useIpSpace
	}

	if ownerOrgId != "" {
		net.DedicatedOrg = &types.OpenApiReference{ID: ownerOrgId}
	}

	return net
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtIpSpaceT0(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("IP Spaces are supported in VCD 10.4.1+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0router, types.ExternalNetworkBackingTypeNsxtTier0Router, true, "", "", "")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtIpSpaceVrf(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("IP Spaces are supported in VCD 10.4.1+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router, true, "", "", "")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtIpSpaceT0DedicatedOrg(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("IP Spaces are supported in VCD 10.4.1+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0router, types.ExternalNetworkBackingTypeNsxtTier0Router, true, vcd.org.Org.ID, "", "")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtIpSpaceVrfDedicatedOrg(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("IP Spaces are supported in VCD 10.4.1+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router, true, vcd.org.Org.ID, "", "")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtNatAndFwIntentionProviderGateway(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 38.1") {
		check.Skip("NAT and Firewall intentions are supported in VCD 10.5.1+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router, true, vcd.org.Org.ID, "PROVIDER_GATEWAY", "IP_SPACE_UPLINKS_ADVERTISED_FLEXIBLE")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtNatAndFwIntentionProviderAndEdgeGateway(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 38.1") {
		check.Skip("NAT and Firewall intentions are supported in VCD 10.5.1+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router, true, vcd.org.Org.ID, "PROVIDER_AND_EDGE_GATEWAY", "ALL_NETWORKS_ADVERTISED")
}

func (vcd *TestVCD) Test_CreateExternalNetworkV2NsxtNatAndFwIntentionEdgeGateway(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 38.1") {
		check.Skip("NAT and Firewall intentions are supported in VCD 10.5.1+")
	}
	vcd.testCreateExternalNetworkV2Nsxt(check, vcd.config.VCD.Nsxt.Tier0routerVrf, types.ExternalNetworkBackingTypeNsxtTier0Router, true, vcd.org.Org.ID, "EDGE_GATEWAY", "IP_SPACE_UPLINKS_ADVERTISED_STRICT")
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
