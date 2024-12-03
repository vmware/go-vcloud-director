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

	ipSpace, ipSpaceCleanup := createTmIpSpace(vcd, region, check)
	defer ipSpaceCleanup()

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
		BackingType: "NSX_TIER0",
		BackingRef:  &types.OpenApiReference{ID: t0ByNameInRegion.TmTier0Gateway.ID},
		RegionRef:   &types.OpenApiReference{ID: region.Region.ID},
		IPSpaceRefs: []*types.OpenApiReference{{
			ID: ipSpace.TmIpSpace.ID,
		}},
	}

	_, err = vcd.client.CreateTmProviderGateway(t)
	panic(err)

}

// {
//     "name": "test-provider-gw",
//     "description": "",
//     "orgRef": null,
//     "backingRef": {
//         "id": "37273049-ecda-4974-baf6-5c107f30a969",
//         "name": "vcfcons-mgt-vc03-Tier0"
//     },
//     "backingType": "NSX_TIER0",
//     "regionRef": {
//         "id": "urn:vcloud:region:7544b246-84c6-40ad-8c3b-beed9fe145cd",
//         "name": "Terraform demo Region"
//     },
//     "ipSpaceRefs": [
//         {
//             "id": "urn:vcloud:ipSpace:feb3d26d-08e7-4e3f-90cc-f54f13a45697",
//             "name": "demo-ip-space"
//         }
//     ]
// }
