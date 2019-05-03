// +build api functional ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_APIVCDMaxVersionIs_Unauthenticated(check *C) {
	config, err := GetConfigStruct()
	check.Assert(err, IsNil)

	vcdClient, err := GetTestVCDFromYaml(config)
	check.Assert(err, IsNil)

	versionCheck := vcdClient.APIVCDMaxVersionIs(">= 27.0")
	check.Assert(versionCheck, Equals, true)
	check.Assert(vcdClient.supportedVersions.VersionInfos, Not(Equals), 0)
}

func (vcd *TestVCD) Test_APIClientVersionIs_Unauthenticated(check *C) {
	config, err := GetConfigStruct()
	check.Assert(err, IsNil)

	vcdClient, err := GetTestVCDFromYaml(config)
	check.Assert(err, IsNil)

	versionCheck := vcdClient.APIClientVersionIs(">= 27.0")
	check.Assert(versionCheck, Equals, true)
	check.Assert(vcdClient.supportedVersions.VersionInfos, Not(Equals), 0)
}

// Test_APIVCDMaxVersionIs uses already authenticated vcdClient (in SetupSuite)
func (vcd *TestVCD) Test_APIVCDMaxVersionIs(check *C) {

	// Minimum supported vCD 8.20 introduced API version 27.0
	versionCheck := vcd.client.APIVCDMaxVersionIs(">= 27.0")
	check.Assert(versionCheck, Equals, true)

	mockVcd := getMockVcdWithAPIVersion("27.0")

	var versionTests = []struct {
		version     string
		boolChecker Checker
		isSupported bool
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
		versionCheck := mockVcd.APIVCDMaxVersionIs(tt.version)
		check.Assert(versionCheck, tt.boolChecker, tt.isSupported)
	}
}

// Test_APIClientVersionIs uses already authenticated vcdClient (in SetupSuite)
func (vcd *TestVCD) Test_APIClientVersionIs(check *C) {

	// Check with currently set version
	versionCheck := vcd.client.APIClientVersionIs(fmt.Sprintf("= %s", vcd.client.Client.APIVersion))
	check.Assert(versionCheck, Equals, true)

	versionCheck = vcd.client.APIClientVersionIs(">= 27.0")
	check.Assert(versionCheck, Equals, true)

	mockVcd := getMockVcdWithAPIVersion("27.0")

	var versionTests = []struct {
		version     string
		boolChecker Checker
		isSupported bool
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
		versionCheck := mockVcd.APIClientVersionIs(tt.version)
		check.Assert(versionCheck, tt.boolChecker, tt.isSupported)
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

func getMockVcdWithAPIVersion(version string) *VCDClient {
	return &VCDClient{
		Client: Client{
			APIVersion: version,
		},
		supportedVersions: SupportedVersions{
			VersionInfos{
				VersionInfo{
					Version: version,
				},
			},
		},
	}
}
