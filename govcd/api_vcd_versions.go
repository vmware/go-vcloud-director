/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"sort"

	semver "github.com/hashicorp/go-version"

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

// APIMaxVerIs compares against maximum vCD supported API version from /api/versions (not necessarily
// the currently used one). This allows to check what is the maximum API version that vCD instance
// supports and can be used to guess vCD product version. API 31.0 support was first introduced in
// vCD 9.5 (as per https://code.vmware.com/doc/preview?id=8072). Therefore APIMaxVerIs(">= 31.0")
// implies that you have vCD 9.5 or newer running inside.
//
// Format: ">= 27.0, < 32.0", ">= 30.0", "= 27.0"
//
// vCD version mapping to API version support https://code.vmware.com/doc/preview?id=8072
func (vdcCli *VCDClient) APIMaxVerIs(versionConstraint string) bool {
	util.Logger.Printf("[TRACE] checking max API version against constraints '%s'", versionConstraint)
	maxVersion, err := vdcCli.maxSupportedVersion()
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find max supported version : %s", err)
		return false
	}

	isSupported, err := vdcCli.apiVerMatchesConstraint(maxVersion, versionConstraint)
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find max supported version : %s", err)
		return false
	}

	return isSupported
}

// APIClientVersionIs allows to compare against currently used API version VCDClient.Client.APIVersion.
// Can be useful to validate if a certain feature can be used or not.
//
// Format: ">= 27.0, < 32.0", ">= 30.0", "= 27.0"
//
// vCD version mapping to API version support https://code.vmware.com/doc/preview?id=8072
func (vdcCli *VCDClient) APIClientVersionIs(versionConstraint string) bool {
	util.Logger.Printf("[TRACE] checking current API version against constraints '%s'", versionConstraint)

	isSupported, err := vdcCli.apiVerMatchesConstraint(vdcCli.Client.APIVersion, versionConstraint)
	if err != nil {
		util.Logger.Printf("[ERROR] unable to find cur supported version : %s", err)
		return false
	}

	return isSupported
}

// vcdFetchSupportedVersions retrieves list of supported versions from
// /api/versions endpoint and stores them in VCDClient for future uses
func (vdcCli *VCDClient) vcdFetchSupportedVersions() error {
	apiEndpoint := vdcCli.Client.VCDHREF
	apiEndpoint.Path += "/versions"

	req := vdcCli.Client.NewRequest(map[string]string{}, "GET", apiEndpoint, nil)
	resp, err := checkResp(vdcCli.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	suppVersions := new(SupportedVersions)
	err = decodeBody(resp, suppVersions)
	if err != nil {
		return fmt.Errorf("error decoding versions response: %s", err)
	}

	vdcCli.supportedVersions = *suppVersions

	return nil
}

// maxSupportedVersion parses supported version list and returns the highest version in string format.
func (vdcCli *VCDClient) maxSupportedVersion() (string, error) {
	versions := make([]*semver.Version, len(vdcCli.supportedVersions.VersionInfos))
	for i, raw := range vdcCli.supportedVersions.VersionInfos {
		v, _ := semver.NewVersion(raw.Version)
		versions[i] = v
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
func (vdcCli *VCDClient) vcdCheckSupportedVersion(version string) (bool, error) {
	return vdcCli.checkSupportedVersionConstraint(fmt.Sprintf("= %s", version))
}

// Checks if there is at least one specified version matching the list returned by vCD.
// Constraint format can be in format ">= 27.0, < 32",">= 30" ,"= 27.0".
func (vdcCli *VCDClient) checkSupportedVersionConstraint(versionConstraint string) (bool, error) {
	for _, vi := range vdcCli.supportedVersions.VersionInfos {
		match, err := vdcCli.apiVerMatchesConstraint(vi.Version, versionConstraint)
		if err != nil {
			return false, fmt.Errorf("cannot match version: %s", err)
		}

		if match {
			return true, nil
		}
	}
	return false, fmt.Errorf("version %s is not supported", versionConstraint)
}

func (vdcCli *VCDClient) apiVerMatchesConstraint(version, versionConstraint string) (bool, error) {

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
		return true, nil
	}

	util.Logger.Printf("[TRACE] API version %s does not satisfy constraints '%s'", checkVer, constraints)
	return false, nil
}

// validateAPIVersion fetches API versions
func (vdcCli *VCDClient) validateAPIVersion() error {
	// vcdRetrieve supported versions
	err := vdcCli.vcdFetchSupportedVersions()
	if err != nil {
		return fmt.Errorf("could not retrieve versions: %s", err)
	}

	// Check if version is supported
	if ok, err := vdcCli.vcdCheckSupportedVersion(vdcCli.Client.APIVersion); !ok || err != nil {
		return fmt.Errorf("API version %s is not supported: %s", vdcCli.Client.APIVersion, err)
	}

	return nil
}
