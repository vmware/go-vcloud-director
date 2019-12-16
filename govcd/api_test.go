// +build api functional ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// IMPORTANT: DO NOT ADD build tags to this file

package govcd

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	. "gopkg.in/check.v1"
)

var testingTags = make(map[string]string)

var testVerbose bool = os.Getenv("GOVCD_TEST_VERBOSE") != ""

// longer than the 128 characters so nothing can be named this
var INVALID_NAME = `*******************************************INVALID
					****************************************************
					************************`

// This ID won't be found by lookup in any entity
var invalidEntityId = "one:two:three:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

func tagsHelp(t *testing.T) {

	var helpText string = `
# -----------------------------------------------------
# Tags are required to run the tests
# -----------------------------------------------------

At least one of the following tags should be defined:

   * ALL :       Runs all the tests (== functional + unit == all feature tests)

   * functional: Runs all the tests that use check.v1
   * unit:       Runs unit tests that do not use check.v1 and don't need a live vCD

   * catalog:    Runs catalog related tests (also catalog_item, media)
   * disk:       Runs disk related tests
   * extnetwork: Runs external network related tests
   * lb:       	 Runs load balancer related tests
   * network:    Runs network related tests
   * gateway:    Runs edge gateway related tests
   * org:        Runs org related tests
   * query:      Runs query related tests
   * system:     Runs system related tests
   * task:       Runs task related tests
   * user:       Runs user related tests
   * vapp:       Runs vapp related tests
   * vdc:        Runs vdc related tests
   * vm:         Runs vm related tests

Examples:

go test -tags functional -check.vv -timeout=45m .
go test -tags catalog -check.vv -timeout=15m .
go test -tags "query extension" -check.vv -timeout=5m .
go test -tags functional -check.vv -check.f Test_AddNewVM  -timeout=15m .
go test -v -tags unit .
`
	t.Logf(helpText)
}

// Tells indirectly if a tag has been set
// For every tag there is an `init` function that
// fills an item in `testingTags`
func isTagSet(tagName string) bool {
	_, ok := testingTags[tagName]
	return ok
}

// For troubleshooting:
// Shows which tags were set, and in which file.
func showTags() {
	if len(testingTags) > 0 {
		fmt.Println("# Defined tags:")
	}
	for k, v := range testingTags {
		fmt.Printf("# %s (%s)", k, v)
	}
}

// Checks whether any tags were defined, and raises an error if not
func TestTags(t *testing.T) {
	if len(testingTags) == 0 {
		t.Logf("# No tags were defined")
		tagsHelp(t)
		t.Fail()
		return
	}
	if os.Getenv("SHOW_TAGS") != "" {
		showTags()
	}
}

// Test_NewRequestWitNotEncodedParamsWithApiVersion verifies that api version override works
func (vcd *TestVCD) Test_NewRequestWitNotEncodedParamsWithApiVersion(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	queryUlr := vcd.client.Client.VCDHREF
	queryUlr.Path += "/query"

	apiVersion, err := vcd.client.Client.maxSupportedVersion()
	check.Assert(err, IsNil)

	req := vcd.client.Client.NewRequestWitNotEncodedParamsWithApiVersion(nil, map[string]string{"type": "media",
		"filter": "name==any"}, http.MethodGet, queryUlr, nil, apiVersion)

	resp, err := checkResp(vcd.client.Client.Http.Do(req))
	check.Assert(err, IsNil)

	check.Assert(resp.Header.Get("Content-Type"), Equals, "application/vnd.vmware.vcloud.query.records+xml;version="+apiVersion)

	// Repeats the call without API version change
	req = vcd.client.Client.NewRequestWitNotEncodedParams(nil, map[string]string{"type": "media",
		"filter": "name==any"}, http.MethodGet, queryUlr, nil)

	resp, err = checkResp(vcd.client.Client.Http.Do(req))
	check.Assert(err, IsNil)

	// Checks that the regularAPI version was not affected by the previous call
	check.Assert(resp.Header.Get("Content-Type"), Equals, "application/vnd.vmware.vcloud.query.records+xml;version="+vcd.client.Client.APIVersion)

	fmt.Printf("Test: %s run with api Version: %s\n", check.TestName(), apiVersion)
}
