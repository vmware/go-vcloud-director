//go:build unit || ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Tests that equalIds returns the expected result whether the target reference
// contains or not an ID
func Test_equalIds(t *testing.T) {

	type testData struct {
		wanted    string
		reference types.Reference
		expected  bool
	}
	var testItems = []testData{
		{
			// Regular case: all values are set in the reference
			wanted: "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0",
			reference: types.Reference{
				Name: "all_values",
				HREF: "https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
				ID:   "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0",
			},
			expected: true,
		},
		{
			// Catalog Item case: The ID is a simple UUID
			wanted: "urn:vcloud:catalogitem:97384890-180c-4563-b9b7-0dc50a2430b0",
			reference: types.Reference{
				Name: "all_values",
				HREF: "https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
				ID:   "97384890-180c-4563-b9b7-0dc50a2430b0",
			},
			expected: true,
		},
		{
			// Regular case: all values are set in the reference but wanted is empty
			wanted: "",
			reference: types.Reference{
				Name: "all_values",
				HREF: "https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
				ID:   "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0",
			},
			expected: false,
		},
		{
			// wanted and ID are different
			wanted: "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0",
			reference: types.Reference{
				Name: "not_matching",
				HREF: "https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
				ID:   "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b1",
			},
			expected: false,
		},
		{
			// Missing ID, the match happens with the HREF (as in VDC)
			wanted: "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0",
			reference: types.Reference{
				Name: "no_id",
				HREF: "https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
				ID:   "",
			},
			expected: true,
		},
		{
			// Missing ID, the UUID in the HREF is different from the one in wanted, will fail
			wanted: "urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0",
			reference: types.Reference{
				Name: "no_id_no_matching",
				HREF: "https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b1",
				ID:   "",
			},
			expected: false,
		},
		{
			// all bogus values. Matches by ID
			wanted: "urn:vcloud:catalog:deadbeef-0000-0000-0000-000000000000",
			reference: types.Reference{
				Name: "all_dummy_values",
				HREF: "https://vcd.somecompany.org/api/entity/deadbeef-0000-0000-0000-000000000000",
				ID:   "urn:vcloud:catalog:deadbeef-0000-0000-0000-000000000000",
			},
			expected: true,
		},
		{
			// Missing both ID and HREF, will fail
			wanted: "urn:vcloud:catalog:deadbeef-0000-0000-0000-000000000000",
			reference: types.Reference{
				Name: "missing_ids",
				HREF: "",
				ID:   "",
			},
			expected: false,
		},
		{
			// URL has ID also case
			wanted: "urn:vcloud:catalogitem:97384890-180c-4563-b9b7-0dc50a2430b0",
			reference: types.Reference{
				Name: "url_with_id",
				HREF: "https://vcd-a8bbe9be-13f2-4ce7-9187-d0d075c42531.cds.cloud.vmware.com/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
			},
			expected: true,
		},
	}
	for _, item := range testItems {
		result := equalIds(item.wanted, item.reference.ID, item.reference.HREF)
		if result == item.expected {
			if testVerbose {
				t.Logf("Test:     %s\nExpected: %v\nwanted:   '%s'\nID:       '%s'\nHREF:     '%s'",
					item.reference.Name, item.expected, item.wanted, item.reference.ID, item.reference.HREF)
			}
		} else {
			t.Logf("Test:     %s\nExpected: %v\nwanted:   '%s'\nID:       '%s'\nHREF:     '%s'",
				item.reference.Name, item.expected, item.wanted, item.reference.ID, item.reference.HREF)
			t.Fail()
		}
	}
}
