/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// IMPORTANT: DO NOT ADD build tags to this file

package govcd

import (
	"fmt"
	"os"
	"testing"
)

var testingTags = make(map[string]string)

var testVerbose bool = os.Getenv("GOVCD_TEST_VERBOSE") != ""

// longer than the 128 characters so nothing can be named this
var INVALID_NAME = `*******************************************INVALID
					****************************************************
					************************`

// This ID won't be found by lookup in any entity
var invalidEntityId = "urn:vcloud:three:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

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
   * vdcGroup:   Runs vdc group related tests
   * certificate Runs certificate related tests
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

func printVerbose(format string, args ...interface{}) {
	if testVerbose {
		fmt.Printf(format, args...)
	}
}

func logVerbose(t *testing.T, format string, args ...interface{}) {
	if testVerbose {
		t.Logf(format, args...)
	}
}
