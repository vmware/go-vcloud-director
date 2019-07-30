// +build unit ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
* Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"testing"
)

// Tests reliability of getBareEntityUuid, which returns a bare UUID from an
// entity ID field
func Test_BareID(t *testing.T) {

	type idTest struct {
		rawId    string
		expected string
	}

	idTestList := []idTest{
		{
			"urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0",
			"97384890-180c-4563-b9b7-0dc50a2430b0",
		},
		{
			"urn:vcloud:org:deadbeef-0000-0000-0000-000000000000",
			"deadbeef-0000-0000-0000-000000000000",
		},
		{
			"urn:vcloud:task:11111111-0000-0000-0000-000000000000",
			"11111111-0000-0000-0000-000000000000",
		},
		{
			"urn:vcloud:task:aaaaaaaa-bbbb-ccc0-dddd-eeeeeeeeeeee",
			"aaaaaaaa-bbbb-ccc0-dddd-eeeeeeeeeeee",
		},
		{
			"urn:vcloud:vdc:72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"urn:composite-name:vdc:72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"urn:underscored_name:vdc:72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"urn:mixed_name-with-dashes:double-string:72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
	}
	idTestExpectedToFailList := []idTest{
		{
			"missing:one:digit:12345678-1234-1234-1234-12345678901",
			"",
		},
		{
			"missing:one:digit:1234567-1234-1234-1234-123456789012",
			"",
		},
		{
			"missing:one:digit:12345678-123-1234-1234-123456789012",
			"",
		},
		{
			"too:many:digits:123456789-1234-1234-1234-123456789012",
			"",
		},
		{
			"too:many:digits:12345678-12345-1234-1234-123456789012",
			"",
		},
		{
			"too:many:digits:12345678-1234-1234-1234-1234567890123",
			"",
		},
		{
			"unexpected:letters:in_ID:abcdefgh-1234-1234-1234-123456789012",
			"",
		},
		{
			"unexpected:letters:in_ID:abcdef00-x234-w234-1234-123456789012",
			"",
		},
		{
			"unexpected:letters:in,prefix:12345678-1234-1234-1234-123456789012",
			"",
		},
		{
			"unexpected:letters:in/prefix:12345678-1234-1234-1234-123456789012",
			"",
		},
	}

	for _, it := range idTestList {
		bareId, err := getBareEntityUuid(it.rawId)
		if err != nil {
			t.Logf("error extracting bare ID: %s", err)
			t.Fail()
		}
		if bareId == it.expected {
			if testVerbose {
				t.Logf("ID '%s': found '%s' as expected", it.rawId, it.expected)
			}
		} else {
			t.Logf("error getting bare ID: expected '%s' but found '%s'", it.expected, bareId)
			t.Fail()
		}
	}
	for _, it := range idTestExpectedToFailList {
		bareId, err := getBareEntityUuid(it.rawId)
		if err == nil {
			t.Logf("unexpected success with raw ID %s", it.rawId)
			t.Fail()
		}
		if bareId == it.expected {
			if testVerbose {
				t.Logf("ID '%s': found '%s' as expected", it.rawId, it.expected)
			}
		} else {
			t.Logf("error getting bare ID: expected '%s' but found '%s'", it.expected, bareId)
			t.Fail()
		}
	}
}
