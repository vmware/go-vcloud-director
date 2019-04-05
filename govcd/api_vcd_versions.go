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

type versionInfo struct {
	Version    string `xml:"Version"`
	LoginUrl   string `xml:"LoginUrl"`
	Deprecated bool   `xml:"deprecated,attr,omitempty"`
}

type versionInfos []versionInfo

type supportedVersions struct {
	versionInfos `xml:"VersionInfo"`
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

	suppVersions := new(supportedVersions)
	err = decodeBody(resp, suppVersions)
	if err != nil {
		return fmt.Errorf("error decoding versions response: %s", err)
	}

	vdcCli.supportedVersions = *suppVersions

	return nil
}

// maxSupportedVersion parses supported version list and returns the highest version in string format.
func (vdcCli *VCDClient) maxSupportedVersion() (string, error) {
	versions := make([]*semver.Version, len(vdcCli.supportedVersions.versionInfos))
	for i, raw := range vdcCli.supportedVersions.versionInfos {
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
	constraints, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		return false, fmt.Errorf("unable to parse version %s", versionConstraint)
	}
	for _, vi := range vdcCli.supportedVersions.versionInfos {
		v, _ := semver.NewVersion(vi.Version)
		if constraints.Check(v) {
			return true, nil
		}

	}
	return false, fmt.Errorf("version %s is not supported", constraints)
}

func (vdcCli *VCDClient) apiVerMatchesConstraint(version, versionConstraint string) (bool, error) {

	checkVer, err := semver.NewVersion(version)
	if err != nil {
		return false, fmt.Errorf("[ERROR] unable to parse max version %s : %s", checkVer, err)
	}
	// Create a provided constraint to check against current max version
	constraints, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		return false, fmt.Errorf("[ERROR] unable to parse given version constraint '%s' : %s", versionConstraint, err)
	}
	if constraints.Check(checkVer) {
		return true, fmt.Errorf("[TRACE] API version %s satisfies constraints '%s'", checkVer, constraints)
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
