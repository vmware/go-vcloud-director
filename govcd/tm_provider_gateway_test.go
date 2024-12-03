//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
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

}
