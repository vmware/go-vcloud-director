//go:build unit || ALL

package govcd

import (
	"fmt"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TestClient_getOpenApiHighestElevatedVersion aims to test out capabilities of getOpenApiHighestElevatedVersion
// It consists of:
// * A few manually defined tests for known endpoints
// * Automatically generated tests for each entry in endpointMinApiVersions to ensure it returns correct version
// * Automatically generated tests where available VCD version does not satisfy it
// * Automatically generated tests to check if each elevated API version is returned for endpoints that have it defined
func TestClient_getOpenApiHighestElevatedVersion(t *testing.T) {
	type testCase struct {
		name              string
		supportedVersions SupportedVersions
		endpoint          string
		wantVersion       string
		wantErr           bool
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
			name:              "VCD_minimum_required_version_only",
			supportedVersions: renderSupportedVersions([]string{"28.0", "29.0", "30.0", "31.0", "32.0", "33.0", "34.0"}),
			endpoint:          types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules,
			wantVersion:       "34.0",
			wantErr:           false, // NAT minimum requirement is version 34.0
		},
		{
			name: "VCD_minimum_required_version_only_unordered",
			// Explicitly pass in unordered VCD API supported versions to ensure that ordering and matching works well
			supportedVersions: renderSupportedVersions([]string{"33.0", "34.0", "30.0", "31.0", "32.0", "28.0", "29.0"}),
			endpoint:          types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtNatRules,
			wantVersion:       "34.0",
			wantErr:           false, // NAT minimum requirement is version 34.0
		},
	}

	// Generate unit tests for each defined endpoint in endpointMinApiVersions which does not have an elevated
	// version entry in endpointElevatedApiVersions.
	// Always expect to get minimal version returned without error
	for endpointName, minimumRequiredVersion := range endpointMinApiVersions {
		// Do not create a test case for those endpoints which explicitly have elevated versions defined in
		// endpointElevatedApiVersions
		if _, hasEntry := endpointElevatedApiVersions[endpointName]; hasEntry {
			continue
		}

		tCase := testCase{
			name: fmt.Sprintf("%s_minimum_version_%s", minimumRequiredVersion, endpointName),
			// Put a list of versions which always satisfied minimum requirement
			supportedVersions: renderSupportedVersions([]string{
				"27.0", "28.0", "29.0", "30.0", "31.0", "32.0", "33.0", "34.0", "35.0", "35.1", "35.2", "36.0", "37.0", "38.0",
			}),
			endpoint:    endpointName,
			wantVersion: minimumRequiredVersion,
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
	for endpointName := range endpointMinApiVersions {
		// Do not create a test case for those endpoints which do not have endpointElevatedApiVersions specified
		if _, hasEntry := endpointElevatedApiVersions[endpointName]; !hasEntry {
			continue
		}

		// generate tests for all elevated rights and expect to get
		for _, singleElevatedApiVerion := range endpointElevatedApiVersions[endpointName] {

			tCase := testCase{
				name: fmt.Sprintf("elevated_up_to_%s_for_%s", singleElevatedApiVerion, endpointName),
				// Insert some older versions and make it so that the highest elevated API version matches
				supportedVersions: renderSupportedVersions([]string{
					"27.0", singleElevatedApiVerion, "23.0", "30.0",
				}),
				endpoint:    endpointName,
				wantVersion: singleElevatedApiVerion,
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
