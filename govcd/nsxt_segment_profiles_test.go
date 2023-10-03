//go:build network || nsxt || functional || openapi || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetNsxtSegmentProfiles(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	vcd.skipIfNotSysAdmin(check)

	nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)

	// Check filtering by NSX-T Manager ID
	filterByNsxtManager := copyOrNewUrlValues(nil)
	filterByNsxtManager = queryParameterFilterAnd(fmt.Sprintf("nsxTManagerRef.id==%s", nsxtManager.NsxtManager.ID), filterByNsxtManager)
	checkNsxtSegmentAllProfilesByFilter(vcd, check, filterByNsxtManager)

	// Check filtering by VDC ID
	filterByVdc := copyOrNewUrlValues(nil)
	filterByVdc = queryParameterFilterAnd(fmt.Sprintf("orgVdcId==%s", vcd.nsxtVdc.Vdc.ID), filterByVdc)
	checkNsxtSegmentAllProfilesByFilter(vcd, check, filterByVdc)

	// Check filtering by VDC Group ID
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	vdcGroup, err := adminOrg.GetVdcGroupByName(vcd.config.VCD.Nsxt.VdcGroup)
	check.Assert(err, IsNil)
	check.Assert(vdcGroup, NotNil)

	filterByVdcGroup := copyOrNewUrlValues(nil)
	filterByVdcGroup = queryParameterFilterAnd(fmt.Sprintf("vdcGroupId==%s", vdcGroup.VdcGroup.Id), filterByVdcGroup)
	checkNsxtSegmentAllProfilesByFilter(vcd, check, filterByVdcGroup)

	// IP Discovery profile by name
	ipDiscoveryProfileByNameInNsxtManager, err := vcd.client.GetIpDiscoveryProfileByName(vcd.config.VCD.Nsxt.IpDiscoveryProfile, filterByNsxtManager)
	check.Assert(err, IsNil)
	check.Assert(ipDiscoveryProfileByNameInNsxtManager.DisplayName, Equals, vcd.config.VCD.Nsxt.IpDiscoveryProfile)

	ipDiscoveryProfileByNameInVdc, err := vcd.client.GetIpDiscoveryProfileByName(vcd.config.VCD.Nsxt.IpDiscoveryProfile, filterByVdc)
	check.Assert(err, IsNil)
	check.Assert(ipDiscoveryProfileByNameInVdc.DisplayName, Equals, vcd.config.VCD.Nsxt.IpDiscoveryProfile)

	ipDiscoveryProfileByNameInVdcGroup, err := vcd.client.GetIpDiscoveryProfileByName(vcd.config.VCD.Nsxt.IpDiscoveryProfile, filterByVdcGroup)
	check.Assert(err, IsNil)
	check.Assert(ipDiscoveryProfileByNameInVdcGroup.DisplayName, Equals, vcd.config.VCD.Nsxt.IpDiscoveryProfile)

	// not found
	notFoundipDiscoveryProfileByNameInNsxtManager, err := vcd.client.GetIpDiscoveryProfileByName("invalid-name", filterByNsxtManager)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundipDiscoveryProfileByNameInNsxtManager, IsNil)

	// Mac Discovery Profile by name
	macDiscoveryProfileByNameInNsxtManager, err := vcd.client.GetMacDiscoveryProfileByName(vcd.config.VCD.Nsxt.MacDiscoveryProfile, filterByNsxtManager)
	check.Assert(err, IsNil)
	check.Assert(macDiscoveryProfileByNameInNsxtManager.DisplayName, Equals, vcd.config.VCD.Nsxt.MacDiscoveryProfile)

	macDiscoveryProfileByNameInVdc, err := vcd.client.GetMacDiscoveryProfileByName(vcd.config.VCD.Nsxt.MacDiscoveryProfile, filterByVdc)
	check.Assert(err, IsNil)
	check.Assert(macDiscoveryProfileByNameInVdc.DisplayName, Equals, vcd.config.VCD.Nsxt.MacDiscoveryProfile)

	macDiscoveryProfileByNameInVdcGroup, err := vcd.client.GetMacDiscoveryProfileByName(vcd.config.VCD.Nsxt.MacDiscoveryProfile, filterByVdcGroup)
	check.Assert(err, IsNil)
	check.Assert(macDiscoveryProfileByNameInVdcGroup.DisplayName, Equals, vcd.config.VCD.Nsxt.MacDiscoveryProfile)

	// not found
	notFoundmacDiscoveryProfileByNameInNsxtManager, err := vcd.client.GetMacDiscoveryProfileByName("invalid-name", filterByNsxtManager)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundmacDiscoveryProfileByNameInNsxtManager, IsNil)

	// Spoof Guard Profile by name
	spoofGuardProfileByNameInNsxtManager, err := vcd.client.GetSpoofGuardProfileByName(vcd.config.VCD.Nsxt.SpoofGuardProfile, filterByNsxtManager)
	check.Assert(err, IsNil)
	check.Assert(spoofGuardProfileByNameInNsxtManager.DisplayName, Equals, vcd.config.VCD.Nsxt.SpoofGuardProfile)

	spoofGuardProfileByNameInVdc, err := vcd.client.GetSpoofGuardProfileByName(vcd.config.VCD.Nsxt.SpoofGuardProfile, filterByVdc)
	check.Assert(err, IsNil)
	check.Assert(spoofGuardProfileByNameInVdc.DisplayName, Equals, vcd.config.VCD.Nsxt.SpoofGuardProfile)

	spoofGuardProfileByNameInVdcGroup, err := vcd.client.GetSpoofGuardProfileByName(vcd.config.VCD.Nsxt.SpoofGuardProfile, filterByVdcGroup)
	check.Assert(err, IsNil)
	check.Assert(spoofGuardProfileByNameInVdcGroup.DisplayName, Equals, vcd.config.VCD.Nsxt.SpoofGuardProfile)

	// not found
	notFoundspoofGuardProfileByNameInVdcGroup, err := vcd.client.GetSpoofGuardProfileByName("invalid-name", filterByNsxtManager)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundspoofGuardProfileByNameInVdcGroup, IsNil)

	// QoS Profile by name
	qosProfileByNameInNsxtManager, err := vcd.client.GetQoSProfileByName(vcd.config.VCD.Nsxt.QosProfile, filterByNsxtManager)
	check.Assert(err, IsNil)
	check.Assert(qosProfileByNameInNsxtManager.DisplayName, Equals, vcd.config.VCD.Nsxt.QosProfile)

	qosProfileByNameInVdc, err := vcd.client.GetQoSProfileByName(vcd.config.VCD.Nsxt.QosProfile, filterByVdc)
	check.Assert(err, IsNil)
	check.Assert(qosProfileByNameInVdc.DisplayName, Equals, vcd.config.VCD.Nsxt.QosProfile)

	qosProfileByNameInVdcGroup, err := vcd.client.GetQoSProfileByName(vcd.config.VCD.Nsxt.QosProfile, filterByVdcGroup)
	check.Assert(err, IsNil)
	check.Assert(qosProfileByNameInVdcGroup.DisplayName, Equals, vcd.config.VCD.Nsxt.QosProfile)

	// not found
	notFoundqosProfileByNameInNsxtManager, err := vcd.client.GetQoSProfileByName("invalid-name", filterByNsxtManager)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundqosProfileByNameInNsxtManager, IsNil)

	// Segment Security Profile by name
	segmentSecurityProfileByNameInNsxtManager, err := vcd.client.GetSegmentSecurityProfileByName(vcd.config.VCD.Nsxt.SegmentSecurityProfile, filterByNsxtManager)
	check.Assert(err, IsNil)
	check.Assert(segmentSecurityProfileByNameInNsxtManager.DisplayName, Equals, vcd.config.VCD.Nsxt.SegmentSecurityProfile)

	segmentSecurityProfileByNameInVdc, err := vcd.client.GetSegmentSecurityProfileByName(vcd.config.VCD.Nsxt.SegmentSecurityProfile, filterByVdc)
	check.Assert(err, IsNil)
	check.Assert(segmentSecurityProfileByNameInVdc.DisplayName, Equals, vcd.config.VCD.Nsxt.SegmentSecurityProfile)

	segmentSecurityProfileByNameInVdcGroup, err := vcd.client.GetSegmentSecurityProfileByName(vcd.config.VCD.Nsxt.SegmentSecurityProfile, filterByVdcGroup)
	check.Assert(err, IsNil)
	check.Assert(segmentSecurityProfileByNameInVdcGroup.DisplayName, Equals, vcd.config.VCD.Nsxt.SegmentSecurityProfile)

	// not found
	notFoundSegmentSecurityProfileByNameInVdcGroup, err := vcd.client.GetSegmentSecurityProfileByName("invalid-name", filterByNsxtManager)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundSegmentSecurityProfileByNameInVdcGroup, IsNil)
}

func checkNsxtSegmentAllProfilesByFilter(vcd *TestVCD, check *C, filter url.Values) {
	ipDiscoverProfiles, err := vcd.client.GetAllIpDiscoveryProfiles(filter)
	check.Assert(err, IsNil)
	check.Assert(ipDiscoverProfiles, NotNil)
	check.Assert(len(ipDiscoverProfiles) > 1, Equals, true)

	macDiscoverProfiles, err := vcd.client.GetAllMacDiscoveryProfiles(filter)
	check.Assert(err, IsNil)
	check.Assert(macDiscoverProfiles, NotNil)
	check.Assert(len(macDiscoverProfiles) > 1, Equals, true)

	spoofGuardDiscoverProfiles, err := vcd.client.GetAllSpoofGuardProfiles(filter)
	check.Assert(err, IsNil)
	check.Assert(spoofGuardDiscoverProfiles, NotNil)
	check.Assert(len(spoofGuardDiscoverProfiles) > 1, Equals, true)

	qosDiscoverProfiles, err := vcd.client.GetAllQoSProfiles(filter)
	check.Assert(err, IsNil)
	check.Assert(qosDiscoverProfiles, NotNil)
	check.Assert(len(qosDiscoverProfiles) > 1, Equals, true)

	segmentSecurityDiscoverProfiles, err := vcd.client.GetAllSegmentSecurityProfiles(filter)
	check.Assert(err, IsNil)
	check.Assert(segmentSecurityDiscoverProfiles, NotNil)
	check.Assert(len(segmentSecurityDiscoverProfiles) > 1, Equals, true)
}
