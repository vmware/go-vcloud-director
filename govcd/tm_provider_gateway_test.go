//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmProviderGateway(check *C) {
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

	ipSpace1, ipSpaceCleanup1 := createTmIpSpace(vcd, region, check, "1", "0")
	defer ipSpaceCleanup1()

	ipSpace2, ipSpaceCleanup2 := createTmIpSpace(vcd, region, check, "2", "1")
	defer ipSpaceCleanup2()

	// Performing TM Tier 0 Gateway lookups to save time on prerequisites in separate tests
	allT0sWithContextFilter, err := vcd.client.GetAllTmTier0GatewaysWithContext(region.Region.ID, true)
	check.Assert(err, IsNil)
	check.Assert(len(allT0sWithContextFilter) > 0, Equals, true)

	t0ByNameInRegion, err := vcd.client.GetTmTier0GatewayWithContextByName(vcd.config.Tm.NsxtTier0Gateway, region.Region.ID, false)
	check.Assert(err, IsNil)
	check.Assert(t0ByNameInRegion, NotNil)
	check.Assert(t0ByNameInRegion.TmTier0Gateway.DisplayName, Equals, vcd.config.Tm.NsxtTier0Gateway)

	t0ByNameInNsxtManager, err := vcd.client.GetTmTier0GatewayWithContextByName(vcd.config.Tm.NsxtTier0Gateway, nsxtManager.NsxtManagerOpenApi.ID, false)
	check.Assert(err, IsNil)
	check.Assert(t0ByNameInNsxtManager, NotNil)
	check.Assert(t0ByNameInNsxtManager.TmTier0Gateway.DisplayName, Equals, vcd.config.Tm.NsxtTier0Gateway)
	// End of performing TM Tier 0 Gateway lookups to save time on prerequisites in separate tests

	// Provider Gateway
	t := &types.TmProviderGateway{
		Name:        check.TestName(),
		Description: check.TestName(),
		BackingType: "NSX_TIER0", // TODO TODO - does it support T0 VRF?
		BackingRef:  types.OpenApiReference{ID: t0ByNameInRegion.TmTier0Gateway.ID},
		RegionRef:   types.OpenApiReference{ID: region.Region.ID},
		IPSpaceRefs: []types.OpenApiReference{{
			ID: ipSpace1.TmIpSpace.ID,
		}},
	}

	createdTmProviderGateway, err := vcd.client.CreateTmProviderGateway(t)
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(createdTmProviderGateway.TmProviderGateway.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmProviderGateways+createdTmProviderGateway.TmProviderGateway.ID)
	defer func() {
		err = createdTmProviderGateway.Delete()
		check.Assert(err, IsNil)
	}()

	// Get TM VDC By Name
	byName, err := vcd.client.GetTmProviderGatewayByName(t.Name)
	check.Assert(err, IsNil)
	check.Assert(byName.TmProviderGateway, DeepEquals, createdTmProviderGateway.TmProviderGateway)

	// Get TM VDC By Id
	byId, err := vcd.client.GetTmProviderGatewayById(createdTmProviderGateway.TmProviderGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(byId.TmProviderGateway, DeepEquals, createdTmProviderGateway.TmProviderGateway)

	// Not Found tests
	byNameInvalid, err := vcd.client.GetTmProviderGatewayByName("fake-name")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byNameInvalid, IsNil)

	byIdInvalid, err := vcd.client.GetTmProviderGatewayById("urn:vcloud:providerGateway:5344b964-0000-0000-0000-d554913db643")
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(byIdInvalid, IsNil)

	// Update (IPSpaceRefs can only be managed using 'TmIpSpaceAssociation' after initial creation)
	createdTmProviderGateway.TmProviderGateway.Name = check.TestName() + "-update"
	updatedVdc, err := createdTmProviderGateway.Update(createdTmProviderGateway.TmProviderGateway)
	check.Assert(err, IsNil)
	check.Assert(updatedVdc.TmProviderGateway, DeepEquals, createdTmProviderGateway.TmProviderGateway)

	// IP Space Association management testing

	// Retrieve existing
	associationByProviderGateway, err := vcd.client.GetAllTmIpSpaceAssociationsByProviderGatewayId(createdTmProviderGateway.TmProviderGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(len(associationByProviderGateway) == 1, Equals, true)

	associationByIpSpace1, err := vcd.client.GetAllTmIpSpaceAssociationsByIpSpaceId(ipSpace1.TmIpSpace.ID)
	check.Assert(err, IsNil)
	check.Assert(len(associationByIpSpace1) == 1, Equals, true)

	// Attempt to find an association that does not exist
	associationByIpSpace2, err := vcd.client.GetAllTmIpSpaceAssociationsByIpSpaceId(ipSpace2.TmIpSpace.ID)
	check.Assert(err, IsNil)
	check.Assert(associationByIpSpace2, NotNil)

	// Create new IP Space Association
	ipSpaceAssociationCfg := &types.TmIpSpaceAssociation{
		Name:               "one",
		IPSpaceRef:         &types.OpenApiReference{ID: ipSpace2.TmIpSpace.ID},
		ProviderGatewayRef: &types.OpenApiReference{ID: createdTmProviderGateway.TmProviderGateway.ID},
	}
	newAssociation, err := vcd.client.CreateTmIpSpaceAssociation(ipSpaceAssociationCfg)
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(newAssociation.TmIpSpaceAssociation.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmIpSpaceAssociations+newAssociation.TmIpSpaceAssociation.ID)
	defer func() {
		err = newAssociation.Delete()
		check.Assert(err, IsNil)
	}()

	// Check new association numbers in Provider Gateway
	updatedAssociationByProviderGateway, err := vcd.client.GetAllTmIpSpaceAssociationsByProviderGatewayId(createdTmProviderGateway.TmProviderGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(len(updatedAssociationByProviderGateway) == 2, Equals, true)

	newAssociationByIpSpace2, err := vcd.client.GetAllTmIpSpaceAssociationsByIpSpaceId(ipSpace2.TmIpSpace.ID)
	check.Assert(err, IsNil)
	check.Assert(len(newAssociationByIpSpace2) == 1, Equals, true)

	// Get Association by ID
	associationById, err := vcd.client.GetTmIpSpaceAssociationById(newAssociationByIpSpace2[0].TmIpSpaceAssociation.ID)
	check.Assert(err, IsNil)
	check.Assert(associationById.TmIpSpaceAssociation, DeepEquals, newAssociationByIpSpace2[0].TmIpSpaceAssociation)

	// Delete Association
	err = newAssociationByIpSpace2[0].Delete()
	check.Assert(err, IsNil)

	// Double check Gateway Association count (should remain 1 again)
	postDeleteAssociationByProviderGateway, err := vcd.client.GetAllTmIpSpaceAssociationsByProviderGatewayId(createdTmProviderGateway.TmProviderGateway.ID)
	check.Assert(err, IsNil)
	check.Assert(len(postDeleteAssociationByProviderGateway) == 1, Equals, true)
}
