//go:build tm || functional || ALL

/*
 * Copyright 2025 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_FeatureFlag(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	allFFs, err := vcd.client.GetAllFeatureFlags()
	check.Assert(err, IsNil)
	check.Assert(len(allFFs) > 0, Equals, true)

	// Specific FF
	ffId := "urn:vcloud:featureflag:CLASSIC_TENANT_CREATION"
	ffById, err := vcd.client.GetFeatureFlagById(ffId)
	check.Assert(err, IsNil)

	originalValue := ffById.Enabled

	// flip value
	ffById.Enabled = !ffById.Enabled
	updatedFF, err := vcd.client.UpdateFeatureFlag(ffById)
	check.Assert(err, IsNil)
	check.Assert(updatedFF.Enabled, Equals, !originalValue)

	// restore state by flipping it once more
	updatedFF.Enabled = !updatedFF.Enabled
	restoredFF, err := vcd.client.UpdateFeatureFlag(updatedFF)
	check.Assert(err, IsNil)
	check.Assert(restoredFF.Enabled, Equals, originalValue)
}
