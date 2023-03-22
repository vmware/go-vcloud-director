//go:build unit || ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"testing"
)

// Tests reliability of getBareEntityUuid, which returns a bare UUID from an
// entity ID field
func Test_BareEntityID(t *testing.T) {

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

// Tests reliability of GetUuidFromHref, which returns a bare UUID from an
// entity HREF field when id is at the end of URL
func Test_GetUuidFromHrefIdAtEnd(t *testing.T) {

	type idTest struct {
		rawHref  string
		expected string
	}

	idTestList := []idTest{
		{
			"https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
			"97384890-180c-4563-b9b7-0dc50a2430b0",
		},
		{
			"https://vcd.somecompany.org/api/entity/deadbeef-0000-0000-0000-000000000000",
			"deadbeef-0000-0000-0000-000000000000",
		},
		{
			"https://vcd.somecompany.org/api/entity/11111111-0000-0000-0000-000000000000",
			"11111111-0000-0000-0000-000000000000",
		},
		{
			"https://vcd.somecompany.org/api/entity/aaaaaaaa-bbbb-ccc0-dddd-eeeeeeeeeeee",
			"aaaaaaaa-bbbb-ccc0-dddd-eeeeeeeeeeee",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
	}
	idTestExpectedToFailList := []idTest{
		{
			"https://vcd.somecompany.org/api/entity/12345678-1234-1234-1234-12345678901",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/1234567-1234-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/12345678-123-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/123456789-1234-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/12345678-12345-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/12345678-1234-1234-1234-1234567890123",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/abcdefgh-1234-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/abcdef00-x234-w234-1234-123456789012",
			"",
		},
	}

	for _, it := range idTestList {
		bareId, err := GetUuidFromHref(it.rawHref, true)
		if err != nil {
			t.Logf("error extracting UUID from HREF: %s", err)
			t.Fail()
		}
		if bareId == it.expected {
			if testVerbose {
				t.Logf("ID '%s': found '%s' as expected", it.rawHref, it.expected)
			}
		} else {
			t.Logf("error getting UUID from HREF: expected '%s' but found '%s'", it.expected, bareId)
			t.Fail()
		}
	}
	for _, it := range idTestExpectedToFailList {
		bareId, err := GetUuidFromHref(it.rawHref, true)
		if err == nil {
			t.Logf("unexpected success with raw HREF %s", it.rawHref)
			t.Fail()
		}
		if bareId == it.expected {
			if testVerbose {
				t.Logf("ID '%s': found '%s' as expected", it.rawHref, it.expected)
			}
		} else {
			t.Logf("error getting UUID from HREF: expected '%s' but found '%s'", it.expected, bareId)
			t.Fail()
		}
	}
}

// Tests reliability of GetUuidFromHref, which returns a bare UUID from an
// entity HREF field when Id is not the end of URL
func Test_GetUuidFromHref(t *testing.T) {

	type idTest struct {
		rawHref  string
		expected string
	}

	idTestList := []idTest{
		{
			"https://vcd.somecompany.org/api/entity/97384890-180c-4563-b9b7-0dc50a2430b0",
			"97384890-180c-4563-b9b7-0dc50a2430b0",
		},
		{
			"https://vcd.somecompany.org/api/entity/deadbeef-0000-0000-0000-000000000000/action/reset",
			"deadbeef-0000-0000-0000-000000000000",
		},
		{
			"https://vcd.somecompany.org/api/entity/11111111-0000-0000-0000-000000000000",
			"11111111-0000-0000-0000-000000000000",
		},
		{
			"https://vcd.somecompany.org/api/entity/aaaaaaaa-bbbb-ccc0-dddd-eeeeeeeeeeee/network",
			"aaaaaaaa-bbbb-ccc0-dddd-eeeeeeeeeeee",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325/test1/test2/test3",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325/",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
		{
			"https://vcd.somecompany.org/api/entity/72fefde7-4fed-45b8-a774-79b72c870325/test1/",
			"72fefde7-4fed-45b8-a774-79b72c870325",
		},
	}
	idTestExpectedToFailList := []idTest{
		{
			"https://vcd.somecompany.org/api/entity/12345678-1234-1234-1234-12345678901",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/1234567-1234-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/12345678-123-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/123456789-1234-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/12345678-12345-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/abcdefgh-1234-1234-1234-123456789012",
			"",
		},
		{
			"https://vcd.somecompany.org/api/entity/abcdef00-x234-w234-1234-123456789012",
			"",
		},
	}

	for _, it := range idTestList {
		bareId, err := GetUuidFromHref(it.rawHref, false)
		if err != nil {
			t.Logf("error extracting UUID from HREF: %s", err)
			t.Fail()
		}
		if bareId == it.expected {
			if testVerbose {
				t.Logf("ID '%s': found '%s' as expected", it.rawHref, it.expected)
			}
		} else {
			t.Logf("error getting UUID from HREF: expected '%s' but found '%s'", it.expected, bareId)
			t.Fail()
		}
	}
	for _, it := range idTestExpectedToFailList {
		bareId, err := GetUuidFromHref(it.rawHref, false)
		if err == nil {
			t.Logf("unexpected success with raw HREF %s", it.rawHref)
			t.Fail()
		}
		if bareId == it.expected {
			if testVerbose {
				t.Logf("ID '%s': found '%s' as expected", it.rawHref, it.expected)
			}
		} else {
			t.Logf("error getting UUID from HREF: expected '%s' but found '%s'", it.expected, bareId)
			t.Fail()
		}
	}
}
