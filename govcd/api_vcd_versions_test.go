/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_APIMaxVerIs(check *C) {

	// Minimum support vCD 8.20 introduced API version 27.0
	r := vcd.client.APIMaxVerIs(">= 27.0")
	check.Assert(r, Equals, true)

	// Mocked tests
	mockVcd := &VCDClient{
		supportedVersions: supportedVersions{
			versionInfos{
				versionInfo{
					Version: "27.0",
				},
			},
		},
	}

	var versionTests = []struct {
		version      string
		boolChecker  Checker
		isSsupported bool
	}{
		{"= 27.0", Equals, true},
		{">= 27.0", Equals, true},
		{">= 25.0, <= 30", Equals, true},
		{"> 27.0", Equals, false},
		{"< 27.0", Equals, false},
		{"invalid", Equals, false},
		{"", Equals, false},
	}

	for _, tt := range versionTests {
		r := mockVcd.APIMaxVerIs(tt.version)
		check.Assert(r, tt.boolChecker, tt.isSsupported)
	}
}

func (vcd *TestVCD) Test_APICurVerIs(check *C) {

	// Check with currently set version
	r := vcd.client.APICurVerIs(fmt.Sprintf("= %s", vcd.client.Client.APIVersion))
	check.Assert(r, Equals, true)

	// Mocked tests
	mockVcd := &VCDClient{
		supportedVersions: supportedVersions{
			versionInfos{
				versionInfo{
					Version: "27.0",
				},
			},
		},
	}

	var versionTests = []struct {
		version      string
		boolChecker  Checker
		isSsupported bool
	}{
		{"= 27.0", Equals, true},
		{">= 27.0", Equals, true},
		{">= 25.0, <= 30", Equals, true},
		{"> 27.0", Equals, false},
		{"< 27.0", Equals, false},
		{"invalid", Equals, false},
		{"", Equals, false},
	}

	for _, tt := range versionTests {
		r := mockVcd.APICurVerIs(tt.version)
		check.Assert(r, tt.boolChecker, tt.isSsupported)
	}
}

func (vcd *TestVCD) Test_validateAPIVersion(check *C) {
	// valid version is checked automatically in SetUpSuite
	// we're checking only for a bad version here
	unsupportedVersion := "999.0"

	config, err := GetConfigStruct()
	check.Assert(err, IsNil)

	vcdClient, err := GetTestVCDFromYaml(config, WithAPIVersion(unsupportedVersion))
	check.Assert(err, IsNil)
	err = vcdClient.validateAPIVersion()
	check.Assert(err, ErrorMatches, "API version .* is not supported: version = .* is not supported")
}
