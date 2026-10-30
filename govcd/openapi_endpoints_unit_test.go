//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"fmt"
	"testing"

	semver "github.com/hashicorp/go-version"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// Test_selectTmApiVersion exercises version policy directly, without HTTP clients or fixtures.
func Test_selectTmApiVersion(t *testing.T) {
	tests := []struct {
		name      string
		versions  []string
		minimum   string
		preferred string
		explicit  string
		want      string
		wantErr   bool
	}{
		{name: "legacy minimum", versions: []string{"39.0", "40.0", "40.1"}, minimum: "40.0", preferred: "37.2", want: "40.0"},
		{name: "product versions only", versions: []string{"9.0.0", "9.1.2"}, minimum: "40.0", want: "9.0.0"},
		{name: "product default newer than alias", versions: []string{"40.1", "9.1.2"}, minimum: "40.1", preferred: "9.1.2", want: "9.1.2"},
		{name: "alias default older than patch minimum", versions: []string{"40.1", "9.1.1", "9.1.2"}, minimum: "9.1.2", preferred: "40.1", want: "9.1.2"},
		{name: "alias does not satisfy patch minimum", versions: []string{"40.1"}, minimum: "9.1.1", wantErr: true},
		{name: "explicit older eligible version", versions: []string{"40.0", "9.1.2"}, minimum: "40.0", preferred: "9.1.2", explicit: "40.0", want: "40.0"},
		{name: "explicit version must be advertised", versions: []string{"40.0", "9.1.2"}, minimum: "40.0", explicit: "9.0.0", wantErr: true},
		{name: "explicit version below minimum", versions: []string{"40.0", "9.1.2"}, minimum: "40.1", explicit: "40.0", wantErr: true},
		{name: "earliest stable fallback", versions: []string{"9.1.2", "41.0", "9.1.1", "invalid"}, minimum: "40.1", want: "9.1.1"},
		{name: "preview requires opt in", versions: []string{"41.0", "41.0.0-alpha"}, minimum: "40.1", wantErr: true},
		{name: "explicit preview", versions: []string{"41.0"}, minimum: "40.1", explicit: "41.0", want: "41.0"},
		{name: "preview client default", versions: []string{"41.0"}, minimum: "40.1", preferred: "41.0", want: "41.0"},
		{name: "preview minimum is not opt in", versions: []string{"41.0"}, minimum: "41.0", wantErr: true},
		{name: "VCD versions are not TM versions", versions: []string{"31.0", "39.0", "39.0.0"}, minimum: "40.0", wantErr: true},
		{name: "invalid minimum", versions: []string{"40.0"}, minimum: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectTmApiVersion(renderSupportedVersions(tt.versions), tt.minimum, tt.preferred, tt.explicit)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("got version %q, error %v; want version %q, error %t", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// TestClient_getOpenApiHighestElevatedVersion aims to test out capabilities of getOpenApiHighestElevatedVersion
// It consists of:
// * A few manually defined tests for known endpoints
// * Automatically generated tests for each entry in endpointMinApiVersions to ensure it returns correct version
// * Automatically generated tests where available VCD version does not satisfy it
// * Automatically generated tests to check if each elevated API version is returned for endpoints that have it defined
func TestClient_getOpenApiHighestElevatedVersion(t *testing.T) {
	semverMinVcdApiVersion, err := semver.NewSemver(minApiVersion)
	if err != nil {
		t.Fatalf("error parsing 'minApiVersion': %s", err)
	}

	type testCase struct {
		name              string
		supportedVersions SupportedVersions
		endpoint          string
		wantVersion       string
		wantErr           bool
		// overrideClientMinApiVersion is an option to override default expected version that is
		// defined in global variable`minApiVersion`
		overrideClientMinApiVersion string
	}

	// Start with some statically defined tests for known endpoints
	tests := []testCase{
		{
			name:              "VCD_does_not_support_minimum_required_version",
			supportedVersions: renderSupportedVersions([]string{"27.0", "28.0", "29.0", "30.0", "31.0", "32.0", "33.0"}),
			endpoint:          types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules,
			wantVersion:       "",
			wantErr:           true, // NAT requires at least version 34.0
		},
		{
			name:                        "VCD_minimum_required_version_only",
			supportedVersions:           renderSupportedVersions([]string{"28.0", "29.0", "30.0", "31.0", "32.0", "33.0", "34.0"}),
			endpoint:                    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules,
			wantVersion:                 "34.0",
			wantErr:                     false, // NAT minimum requirement is version 34.0
			overrideClientMinApiVersion: "34.0",
		},
		{
			name: "VCD_minimum_required_version_only_unordered",
			// Explicitly pass in unordered VCD API supported versions to ensure that ordering and matching works well
			supportedVersions:           renderSupportedVersions([]string{"33.0", "34.0", "30.0", "31.0", "32.0", "28.0", "29.0"}),
			endpoint:                    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules,
			wantVersion:                 "34.0",
			wantErr:                     false, // NAT minimum requirement is version 34.0
			overrideClientMinApiVersion: "34.0",
		},
		{
			name:              "VCD_version_higher_than_elevated_version_entries",
			supportedVersions: renderSupportedVersions([]string{"37.0", "37.1", "37.2", "37.3", "38.0", "38.1", "39.0"}),
			endpoint:          types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups,
			wantVersion:       minApiVersion,
			wantErr:           false,
		},
	}

	// Generate unit tests for each defined endpoint in endpointMinApiVersions which does not have an elevated
	// version entry in endpointElevatedApiVersions.
	// Expect to get minimal supported version returned without error
	for endpointName, minimumRequiredVersion := range endpointMinApiVersions {
		// Do not create a test case for those endpoints which explicitly have elevated versions defined in
		// endpointElevatedApiVersions
		if _, hasEntry := endpointElevatedApiVersions[endpointName]; hasEntry {
			continue
		}

		wantVersion := minimumRequiredVersion
		semverWantVersion, err := semver.NewSemver(wantVersion)
		if err != nil {
			t.Fatalf("error parsing 'singleElevatedApiVersion': %s", err)
		}

		if semverWantVersion.LessThan(semverMinVcdApiVersion) {
			semverMinVcdApiVersionSegments := semverMinVcdApiVersion.Segments()
			wantVersion = fmt.Sprintf("%d.%d", semverMinVcdApiVersionSegments[0], semverMinVcdApiVersionSegments[1])
			if testVerbose {
				fmt.Printf("# Overriding wanted version to %s for endpoint %s\n", wantVersion, endpointName)
			}
		}

		tCase := testCase{
			name: fmt.Sprintf("%s_minimum_version_%s", minimumRequiredVersion, endpointName),
			// Put a list of versions which always satisfied minimum requirement
			supportedVersions: renderSupportedVersions([]string{
				"27.0", "28.0", "29.0", "30.0", "31.0", "32.0", "33.0", "34.0", "35.0", "35.1", "35.2", "36.0", "37.0", "38.0", "39.0", "40.0", "40.1", "41.0",
			}),
			endpoint:    endpointName,
			wantVersion: wantVersion,
			wantErr:     false,
		}
		tests = append(tests, tCase)
	}

	// Generate unit tests for each defined endpoint in endpointMinApiVersions which does not have supported version at all
	// Always expect an error and empty version
	for endpointName, minimumRequiredVersion := range endpointMinApiVersions {
		tCase := testCase{
			name: fmt.Sprintf("%s_unsatisfied_minimum_version_%s", minimumRequiredVersion, endpointName),
			supportedVersions: renderSupportedVersions([]string{
				"25.0",
			}),
			endpoint:    endpointName,
			wantVersion: "",
			wantErr:     true,
		}
		tests = append(tests, tCase)
	}

	// Generate tests for each elevated API version in endpoints which do have elevated rights defined
	// Expect to get either that version or minimum supported version
	for endpointName := range endpointMinApiVersions {
		// Do not create a test case for those endpoints which do not have endpointElevatedApiVersions specified
		if _, hasEntry := endpointElevatedApiVersions[endpointName]; !hasEntry {
			continue
		}

		// generate tests for all elevated rights and expect to get
		for _, singleElevatedApiVersion := range endpointElevatedApiVersions[endpointName] {
			wantVersion := singleElevatedApiVersion

			semverWantVersion, err := semver.NewSemver(wantVersion)
			if err != nil {
				t.Fatalf("error parsing 'singleElevatedApiVersion': %s", err)
			}

			if semverWantVersion.LessThan(semverMinVcdApiVersion) {
				semverMinVcdApiVersionSegments := semverMinVcdApiVersion.Segments()
				wantVersion = fmt.Sprintf("%d.%d", semverMinVcdApiVersionSegments[0], semverMinVcdApiVersionSegments[1])
				if testVerbose {
					fmt.Printf("# Overriding wanted version to %s for endpoint %s\n", wantVersion, endpointName)
				}
			}

			tCase := testCase{
				name: fmt.Sprintf("elevated_up_to_%s_for_%s", singleElevatedApiVersion, endpointName),
				// Insert some older versions and make it so that the highest elevated API version matches
				supportedVersions: renderSupportedVersions([]string{
					"27.0", singleElevatedApiVersion, "23.0", "30.0",
				}),
				endpoint:    endpointName,
				wantVersion: wantVersion,
				wantErr:     false,
			}
			tests = append(tests, tCase)
		}
	}

	// Run all defined tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				supportedVersions: tt.supportedVersions,
				APIVersion:        minApiVersion,
			}

			if tt.overrideClientMinApiVersion != "" {
				client.APIVersion = tt.overrideClientMinApiVersion
			}

			got, err := client.getOpenApiHighestElevatedVersion(tt.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOpenApiHighestElevatedVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantVersion {
				t.Errorf("getOpenApiHighestElevatedVersion() got = %v, wantVersion %v", got, tt.wantVersion)
			}
		})
	}
}

// renderSupportedVersions is a helper to create fake `SupportedVersions` type out of given VCD API version list
func renderSupportedVersions(versions []string) SupportedVersions {
	supportedVersions := SupportedVersions{}

	for _, version := range versions {
		supportedVersions.VersionInfos = append(supportedVersions.VersionInfos,
			VersionInfo{
				Version:    version,
				LoginUrl:   "https://fake-host/api/sessions",
				Deprecated: false,
			})
	}
	return supportedVersions
}
