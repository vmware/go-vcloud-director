//go:build unit || ALL
// +build unit ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"
)

func TestVmGroupFilterWithResourcePools(t *testing.T) {

	getDummyResourcePools := func(howMany int, generateErrors bool) []*types.QueryResultResourcePoolRecordType {
		var resourcePools []*types.QueryResultResourcePoolRecordType
		clusterMoref := ""
		vCenterHREF := ""
		for i := 0; i < howMany; i++ {
			if !generateErrors {
				clusterMoref = fmt.Sprintf("domain-%d", i%10)
				vCenterHREF = fmt.Sprintf("https://my-company-vcd.com/api/admin/extension/vimServer/f583b76e-9e34-48e7-b90d-930653ee161%d", i%10)
			}
			resourcePools = append(resourcePools, &types.QueryResultResourcePoolRecordType{
				ClusterMoref: clusterMoref,
				VcenterHREF:  vCenterHREF,
			})
		}
		return resourcePools
	}

	type testData struct {
		resourcePools  []*types.QueryResultResourcePoolRecordType
		idKey          string
		idValue        string
		expectedFilter string
		expectedError  string
	}
	var testItems = []testData{
		{
			resourcePools:  getDummyResourcePools(2, false),
			idKey:          "namedVmGroupId",
			idValue:        "12345678-9012-3456-7890-123456789012",
			expectedFilter: "(namedVmGroupId==12345678-9012-3456-7890-123456789012;(clusterMoref==domain-0,clusterMoref==domain-1);(vcId==f583b76e-9e34-48e7-b90d-930653ee1610,vcId==f583b76e-9e34-48e7-b90d-930653ee1611))",
			expectedError:  "",
		},
		{
			resourcePools:  getDummyResourcePools(1, false),
			idKey:          "namedVmGroupId",
			idValue:        "12345678-9012-3456-7890-123456789012",
			expectedFilter: "(namedVmGroupId==12345678-9012-3456-7890-123456789012;(clusterMoref==domain-0);(vcId==f583b76e-9e34-48e7-b90d-930653ee1610))",
			expectedError:  "",
		},
		{
			resourcePools:  getDummyResourcePools(2, false),
			idKey:          "vmGroupName",
			idValue:        "testVmGroup",
			expectedFilter: "(vmGroupName==testVmGroup;(clusterMoref==domain-0,clusterMoref==domain-1);(vcId==f583b76e-9e34-48e7-b90d-930653ee1610,vcId==f583b76e-9e34-48e7-b90d-930653ee1611))",
			expectedError:  "",
		},
		{
			resourcePools:  getDummyResourcePools(1, false),
			idKey:          "vmGroupName",
			idValue:        "testVmGroup",
			expectedFilter: "(vmGroupName==testVmGroup;(clusterMoref==domain-0);(vcId==f583b76e-9e34-48e7-b90d-930653ee1610))",
			expectedError:  "",
		},
		{
			resourcePools:  getDummyResourcePools(1, true),
			idKey:          "someKey",
			idValue:        "someValue",
			expectedFilter: "",
			expectedError:  "could not retrieve Resource pools information to retrieve VM Group with someKey=someValue",
		},
		{
			resourcePools:  getDummyResourcePools(1, true),
			idKey:          "",
			idValue:        "",
			expectedFilter: "someFilter",
			expectedError:  "identifier must have a key and value to be able to search",
		},
		{
			resourcePools:  getDummyResourcePools(0, true),
			idKey:          "someKey",
			idValue:        "someValue",
			expectedFilter: "",
			expectedError:  "could not retrieve Resource pools information to retrieve VM Group with someKey=someValue",
		},
	}
	for _, test := range testItems {
		filter, err := buildFilterForVmGroups(test.resourcePools, test.idKey, test.idValue)
		if test.expectedError == "" {
			// Successful path
			if filter != test.expectedFilter {
				t.Errorf("Expected this filter: '%s' for %s=%s but got: %s", test.expectedFilter, test.idKey, test.idValue, filter)
			}
		} else {
			// Error path
			if err == nil || err.Error() != test.expectedError {
				t.Errorf("Expected error for %s=%s but got: %s", test.idKey, test.idValue, err)
			}
		}
	}

}