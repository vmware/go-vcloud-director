//go:build query || functional || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_CheckCumulativeQuery(check *C) {
	vcd.skipIfNotSysAdmin(check)
	pvdcs, err := vcd.client.QueryProviderVdcs()
	check.Assert(err, IsNil)
	var storageProfileMap = make(map[string]bool)

	for _, pvdcRec := range pvdcs {
		pvdc, err := vcd.client.GetProviderVdcByHref(pvdcRec.HREF)
		check.Assert(err, IsNil)
		for _, sp := range pvdc.ProviderVdc.StorageProfiles.ProviderVdcStorageProfile {
			storageProfileMap[sp.Name] = true
		}
	}
	if len(storageProfileMap) < 2 {
		check.Skip("not enough storage profiles found for this test")
	}

	checkQuery := func(pageSize string) {
		var foundStorageProfileMap = make(map[string]bool)
		results, err := vcd.client.Client.cumulativeQuery(types.QtProviderVdcStorageProfile, nil, map[string]string{
			"type":     types.QtProviderVdcStorageProfile,
			"pageSize": pageSize,
		})

		check.Assert(err, IsNil)
		check.Assert(results, NotNil)
		check.Assert(results.Results, NotNil)
		check.Assert(results.Results.ProviderVdcStorageProfileRecord, NotNil)

		// Removing duplicates from results
		for _, sp := range results.Results.ProviderVdcStorageProfileRecord {
			foundStorageProfileMap[sp.Name] = true
		}
		check.Assert(len(foundStorageProfileMap), Equals, len(storageProfileMap))
	}
	checkQuery("1")
	checkQuery("2")
	checkQuery("25")
}
