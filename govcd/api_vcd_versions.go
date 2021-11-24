/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	semver "github.com/hashicorp/go-version"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type VersionInfo struct {
	Version    string `xml:"Version"`
	LoginUrl   string `xml:"LoginUrl"`
	Deprecated bool   `xml:"deprecated,attr,omitempty"`
}

type VersionInfos []VersionInfo

type SupportedVersions struct {
	VersionInfos `xml:"VersionInfo"`
}

// VcdVersion contains the full information about a VCD version
type VcdVersion struct {
	Version *semver.Version
	Time    time.Time
}

// apiVersionToVcdVersion gets the vCD version from max supported API version
var apiVersionToVcdVersion = map[string]string{
	"29.0": "9.0",
	"30.0": "9.1",
	"31.0": "9.5",
	"32.0": "9.7",
	"33.0": "10.0",
	"34.0": "10.1",
	"35.0": "10.2",
	"36.0": "10.3", // Provisional version for non-GA release. It may change later
}

// vcdVersionToApiVersion gets the max supported API version from vCD version
var vcdVersionToApiVersion = map[string]string{
	"9.0":  "29.0",
	"9.1":  "30.0",
	"9.5":  "31.0",
	"9.7":  "32.0",
	"10.0": "33.0",
	"10.1": "34.0",
	"10.2": "35.0",
	"10.3": "36.0", // Provisional version for non-GA release. It may change later
}

// to make vcdVersionToApiVersion used
var _ = vcdVersionToApiVersion

// APIVCDMaxVersionIs compares against maximum vCD supported API version from /api/versions (not necessarily
// the currently used one). This allows to check what is the maximum API version that vCD instance
// supports and can be used to guess vCD product version. API 31.0 support was first introduced in
// vCD 9.5 (as per https://code.vmware.com/doc/preview?id=8072). Therefore APIMaxVerIs(">= 31.0")
// implies that you have vCD 9.5 or newer running inside.
// It does not require for the client to be authenticated.
//
// Format: ">= 27.0, < 32.0", ">= 30.0", "= 27.0"
//
// vCD version mapping to API version support https://code.vmware.com/doc/preview?id=8072
func (client *Client) APIVCDMaxVersionIs(versionConstraint string) bool {
	err := client.vcdFetchSupportedVersions()
	if err != nil {
		util.Logger.Printf("[ERROR] could not retrieve supported versions: %s", err)
		return false
	}

	util.Logger.Printf("[TRACE] checking max API version against constraints '%s'", versionConstraint)
	maxVersion, err := client.MaxSupportedVersion()
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find max supported version : %s", err)
		return false
	}

	isSupported, err := client.apiVersionMatchesConstraint(maxVersion, versionConstraint)
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find max supported version : %s", err)
		return false
	}

	return isSupported
}

// APIClientVersionIs allows to compare against currently used API version VCDClient.Client.APIVersion.
// Can be useful to validate if a certain feature can be used or not.
// It does not require for the client to be authenticated.
//
// Format: ">= 27.0, < 32.0", ">= 30.0", "= 27.0"
//
// vCD version mapping to API version support https://code.vmware.com/doc/preview?id=8072
func (client *Client) APIClientVersionIs(versionConstraint string) bool {

	util.Logger.Printf("[TRACE] checking current API version against constraints '%s'", versionConstraint)

	isSupported, err := client.apiVersionMatchesConstraint(client.APIVersion, versionConstraint)
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find supported version : %s", err)
		return false
	}

	return isSupported
}

// vcdFetchSupportedVersions retrieves list of supported versions from
// /api/versions endpoint and stores them in VCDClient for future uses.
// It only does it once.
func (client *Client) vcdFetchSupportedVersions() error {
	// Only fetch /versions if it is not stored already
	numVersions := len(client.supportedVersions.VersionInfos)
	if numVersions > 0 {
		util.Logger.Printf("[TRACE] skipping fetch of versions because %d are stored", numVersions)
		return nil
	}

	apiEndpoint := client.VCDHREF
	apiEndpoint.Path += "/versions"

	suppVersions := new(SupportedVersions)
	_, err := client.ExecuteRequest(apiEndpoint.String(), http.MethodGet,
		"", "error fetching versions: %s", nil, suppVersions)

	client.supportedVersions = *suppVersions

	// Log all supported API versions in one line to help identify vCD version from logs
	allApiVersions := make([]string, len(client.supportedVersions.VersionInfos))
	for versionIndex, version := range client.supportedVersions.VersionInfos {
		allApiVersions[versionIndex] = version.Version
	}
	util.Logger.Printf("[DEBUG] supported API versions : %s", strings.Join(allApiVersions, ","))

	return err
}

// MaxSupportedVersion parses supported version list and returns the highest version in string format.
func (client *Client) MaxSupportedVersion() (string, error) {
	versions := make([]*semver.Version, len(client.supportedVersions.VersionInfos))
	for index, versionInfo := range client.supportedVersions.VersionInfos {
		version, err := semver.NewVersion(versionInfo.Version)
		if err != nil {
			return "", fmt.Errorf("error parsing version %s: %s", versionInfo.Version, err)
		}
		versions[index] = version
	}
	// Sort supported versions in order lowest-highest
	sort.Sort(semver.Collection(versions))

	switch {
	case len(versions) > 1:
		return versions[len(versions)-1].Original(), nil
	case len(versions) == 1:
		return versions[0].Original(), nil
	default:
		return "", fmt.Errorf("could not identify supported versions")
	}
}

// vcdCheckSupportedVersion checks if there is at least one specified version exactly matching listed ones.
// Format example "27.0"
func (client *Client) vcdCheckSupportedVersion(version string) error {
	return client.checkSupportedVersionConstraint(fmt.Sprintf("= %s", version))
}

// Checks if there is at least one specified version matching the list returned by vCD.
// Constraint format can be in format ">= 27.0, < 32",">= 30" ,"= 27.0".
func (client *Client) checkSupportedVersionConstraint(versionConstraint string) error {
	for _, versionInfo := range client.supportedVersions.VersionInfos {
		versionMatch, err := client.apiVersionMatchesConstraint(versionInfo.Version, versionConstraint)
		if err != nil {
			return fmt.Errorf("cannot match version: %s", err)
		}

		if versionMatch {
			return nil
		}
	}
	return fmt.Errorf("version %s is not supported", versionConstraint)
}

func (client *Client) apiVersionMatchesConstraint(version, versionConstraint string) (bool, error) {

	checkVer, err := semver.NewVersion(version)
	if err != nil {
		return false, fmt.Errorf("[ERROR] unable to parse version %s : %s", version, err)
	}
	// Create a provided constraint to check against current max version
	constraints, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		return false, fmt.Errorf("[ERROR] unable to parse given version constraint '%s' : %s", versionConstraint, err)
	}
	if constraints.Check(checkVer) {
		util.Logger.Printf("[INFO] API version %s satisfies constraints '%s'", checkVer, constraints)
		return true, nil
	}

	util.Logger.Printf("[TRACE] API version %s does not satisfy constraints '%s'", checkVer, constraints)
	return false, nil
}

// validateAPIVersion fetches API versions
func (client *Client) validateAPIVersion() error {
	err := client.vcdFetchSupportedVersions()
	if err != nil {
		return fmt.Errorf("could not retrieve supported versions: %s", err)
	}

	// Check if version is supported
	err = client.vcdCheckSupportedVersion(client.APIVersion)
	if err != nil {
		return fmt.Errorf("API version %s is not supported: %s", client.APIVersion, err)
	}

	return nil
}

// GetSpecificApiVersionOnCondition returns default version or wantedApiVersion if it is connected to version
// described in vcdApiVersionCondition
// f.e. values ">= 32.0", "32.0" returns 32.0 if vCD version is above or 9.7
func (client *Client) GetSpecificApiVersionOnCondition(vcdApiVersionCondition, wantedApiVersion string) string {
	apiVersion := client.APIVersion
	if client.APIVCDMaxVersionIs(vcdApiVersionCondition) {
		apiVersion = wantedApiVersion
	}
	return apiVersion
}

// GetVcdVersion finds the VCD version and the time of build
func (client *Client) GetVcdVersion() (string, time.Time, error) {

	path := client.VCDHREF
	path.Path += "/admin"
	var admin types.VCloud
	_, err := client.ExecuteRequest(path.String(), http.MethodGet,
		"", "error retrieving admin info: %s", nil, &admin)
	if err != nil {
		return "", time.Time{}, err
	}
	description := admin.Description

	if description == "" {
		return "", time.Time{}, fmt.Errorf("no version information found")
	}
	reVersion := regexp.MustCompile(`^\s*(\S+)\s+(.*)`)

	versionList := reVersion.FindAllStringSubmatch(description, -1)

	if len(versionList) == 0 || len(versionList[0]) < 2 {
		return "", time.Time{}, fmt.Errorf("error getting version information from description %s", description)
	}
	version := versionList[0][1]
	versionDate := versionList[0][2]
	versionTime, err := dateparse.ParseStrict(versionDate)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("[version %s] could not convert date %s to formal date: %s", version, versionDate, err)
	}

	return version, versionTime, nil
}

// GetVcdShortVersion returns the VCD version (three digits, no build info)
func (client *Client) GetVcdShortVersion() (string, error) {

	vcdVersion, err := client.GetVcdFullVersion()
	if err != nil {
		return "", fmt.Errorf("error getting version digits: %s", err)
	}
	digits := vcdVersion.Version.Segments()
	return fmt.Sprintf("%d.%d.%d", digits[0], digits[1], digits[2]), nil
}

// GetVcdFullVersion returns the full VCD version information as a structure
func (client *Client) GetVcdFullVersion() (VcdVersion, error) {
	var vcdVersion VcdVersion
	version, versionTime, err := client.GetVcdVersion()
	if err != nil {
		return VcdVersion{}, err
	}

	vcdVersion.Version, err = semver.NewVersion(version)
	if err != nil {
		return VcdVersion{}, err
	}
	if len(vcdVersion.Version.Segments()) < 4 {
		return VcdVersion{}, fmt.Errorf("error getting version digits from version %s", version)
	}
	vcdVersion.Time = versionTime
	return vcdVersion, nil
}

// intListToVersion converts a list of integers into a dot-separated string
func intListToVersion(digits []int, atMost int) string {
	result := ""
	for i, digit := range digits {
		if result != "" {
			result += "."
		}
		if i >= atMost {
			result += "0"
		} else {
			result += fmt.Sprintf("%d", digit)
		}
	}
	return result
}

// VersionEqualOrGreater return true if the current version is the same or greater than the one being compared.
// If howManyDigits is > 3, the comparison includes the build.
// Examples:
//  client version is 1.2.3.1234
//  compare version is 1.2.3.2000
// function return true if howManyDigits is <= 3, but false if howManyDigits is > 3
//
//  client version is 1.2.3.1234
//  compare version is 1.1.1.0
// function returns true regardless of value of howManyDigits
func (client *Client) VersionEqualOrGreater(compareTo string, howManyDigits int) (bool, error) {

	fullVersion, err := client.GetVcdFullVersion()
	if err != nil {
		return false, err
	}
	compareToVersion, err := semver.NewVersion(compareTo)
	if err != nil {
		return false, err
	}
	if howManyDigits < 4 {
		currentString := intListToVersion(fullVersion.Version.Segments(), howManyDigits)
		compareToString := intListToVersion(compareToVersion.Segments(), howManyDigits)
		fullVersion.Version, err = semver.NewVersion(currentString)
		if err != nil {
			return false, err
		}
		compareToVersion, err = semver.NewVersion(compareToString)
		if err != nil {
			return false, err
		}
	}

	return fullVersion.Version.GreaterThanOrEqual(compareToVersion), nil
}
