//go:build api || openapi || functional || catalog || vapp || gateway || network || org || query || extnetwork || task || vm || vdc || system || disk || lb || lbAppRule || lbAppProfile || lbServerPool || lbServiceMonitor || lbVirtualServer || user || search || nsxv || nsxt || auth || affinity || role || alb || certificate || vdcGroup || metadata || providervdc || rde || vsphere || uiPlugin || cse || slz || tm || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

// getOrCreateVCenter will check configuration file and create vCenter if
// stated in the 'createVcenter' property and not present in TM.
// If created, it returns also a cleanup function to delete it afterward.
// Otherwise, it just retrieves the existing vCenter
func getOrCreateVCenter(vcd *TestVCD, check *C) (*VCenter, func()) {
	vc, err := vcd.client.GetVCenterByUrl(vcd.config.Tm.VcenterUrl)
	if err == nil {
		return vc, func() {}
	}
	if !ContainsNotFound(err) {
		check.Fatal(err)
		return nil, nil
	}
	if !vcd.config.Tm.CreateVcenter {
		check.Skip("vCenter is not configured and configuration is not allowed in config file")
		return nil, nil
	}

	vcCfg := &types.VSphereVirtualCenter{
		Name:      check.TestName() + "-vc",
		Username:  vcd.config.Tm.VcenterUsername,
		Password:  vcd.config.Tm.VcenterPassword,
		Url:       vcd.config.Tm.VcenterUrl,
		IsEnabled: true,
	}
	// Certificate must be trusted before adding vCenter
	url, err := url.Parse(vcCfg.Url)
	check.Assert(err, IsNil)
	trustedCert, err := vcd.client.AutoTrustCertificate(url)
	check.Assert(err, IsNil)
	if trustedCert != nil {
		AddToCleanupListOpenApi(trustedCert.TrustedCertificate.ID, check.TestName()+"trusted-cert", types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTrustedCertificates+trustedCert.TrustedCertificate.ID)
	}

	vc, err = vcd.client.CreateVcenter(vcCfg)
	check.Assert(err, IsNil)
	check.Assert(vc, NotNil)
	PrependToCleanupList(vcCfg.Name, "OpenApiEntityVcenter", check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVirtualCenters+vc.VSphereVCenter.VcId)

	// Refresh connected vCenter to be sure that all artifacts are loaded
	printVerbose("# Refreshing vCenter %s\n", vc.VSphereVCenter.Url)
	err = vc.Refresh()
	check.Assert(err, IsNil)

	printVerbose("# Refreshing Storage Profiles in vCenter %s\n", vc.VSphereVCenter.Url)
	err = vc.RefreshStorageProfiles()
	check.Assert(err, IsNil)

	vCenterCreated := true

	return vc, func() {
		if !vCenterCreated {
			return
		}
		err = vc.Disable()
		check.Assert(err, IsNil)
		err = vc.Delete()
		check.Assert(err, IsNil)
	}
}

// getOrCreateNsxtManager will check configuration file and create NSX-T Manager if
// stated in the 'createNsxtManager' property and not present in TM.
// If created, it returns also a cleanup function to delete it afterward.
// Otherwise, it just retrieves the existing NSX-T Manager
func getOrCreateNsxtManager(vcd *TestVCD, check *C) (*NsxtManagerOpenApi, func()) {
	nsxtManager, err := vcd.client.GetNsxtManagerOpenApiByUrl(vcd.config.Tm.NsxtManagerUrl)
	if err == nil {
		return nsxtManager, func() {}
	}
	if !ContainsNotFound(err) {
		check.Fatal(err)
		return nil, nil
	}
	if !vcd.config.Tm.CreateNsxtManager {
		check.Skip("NSX-T Manager is not configured and configuration is not allowed in config file")
		return nil, nil
	}

	nsxtCfg := &types.NsxtManagerOpenApi{
		Name:     check.TestName(),
		Username: vcd.config.Tm.NsxtManagerUsername,
		Password: vcd.config.Tm.NsxtManagerPassword,
		Url:      vcd.config.Tm.NsxtManagerUrl,
	}
	// Certificate must be trusted before adding NSX-T Manager
	url, err := url.Parse(nsxtCfg.Url)
	check.Assert(err, IsNil)
	trustedCert, err := vcd.client.AutoTrustCertificate(url)
	check.Assert(err, IsNil)
	if trustedCert != nil {
		AddToCleanupListOpenApi(trustedCert.TrustedCertificate.ID, check.TestName()+"trusted-cert", types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTrustedCertificates+trustedCert.TrustedCertificate.ID)
	}
	nsxtManager, err = vcd.client.CreateNsxtManagerOpenApi(nsxtCfg)
	check.Assert(err, IsNil)
	check.Assert(nsxtManager, NotNil)
	PrependToCleanupListOpenApi(nsxtManager.NsxtManagerOpenApi.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointNsxManagers+nsxtManager.NsxtManagerOpenApi.ID)
	nsxtManagerCreated := true

	return nsxtManager, func() {
		if !nsxtManagerCreated {
			return
		}
		err = nsxtManager.Delete()
		check.Assert(err, IsNil)
	}
}

// getOrCreateRegion will check configuration file and create a Region if
// stated in the 'createRegion' testing property not present in TM.
// Otherwise, it just retrieves it
func getOrCreateRegion(vcd *TestVCD, nsxtManager *NsxtManagerOpenApi, supervisor *Supervisor, check *C) (*Region, func()) {
	region, err := vcd.client.GetRegionByName(vcd.config.Tm.Region)
	if err == nil {
		return region, func() {}
	}
	if !ContainsNotFound(err) {
		check.Fatal(err)
		return nil, nil
	}
	if !vcd.config.Tm.CreateRegion {
		check.Skip("Region is not configured and configuration is not allowed in config file")
		return nil, nil
	}
	if nsxtManager == nil || supervisor == nil {
		check.Fatalf("getOrCreateRegion requires a not nil NSX-T Manager and Supervisor")
	}

	r := &types.Region{
		Name: check.TestName(),
		NsxManager: &types.OpenApiReference{
			ID: nsxtManager.NsxtManagerOpenApi.ID,
		},
		Supervisors: []types.OpenApiReference{
			{
				ID:   supervisor.Supervisor.SupervisorID,
				Name: supervisor.Supervisor.Name,
			},
		},
		StoragePolicies: []string{vcd.config.Tm.VcenterStorageProfile},
		IsEnabled:       true,
	}

	region, err = vcd.client.CreateRegion(r)
	check.Assert(err, IsNil)
	check.Assert(region, NotNil)
	regionCreated := true
	AddToCleanupListOpenApi(region.Region.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointRegions+region.Region.ID)

	return region, func() {
		if !regionCreated {
			return
		}
		err = region.Delete()
		check.Assert(err, IsNil)
	}
}

func createOrg(vcd *TestVCD, check *C, canManageOrgs bool) (*TmOrg, func()) {
	cfg := &types.TmOrg{
		Name:          check.TestName(),
		DisplayName:   check.TestName(),
		CanManageOrgs: canManageOrgs,
	}
	tmOrg, err := vcd.client.CreateTmOrg(cfg)
	check.Assert(err, IsNil)
	check.Assert(tmOrg, NotNil)

	PrependToCleanupListOpenApi(tmOrg.TmOrg.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgs+tmOrg.TmOrg.ID)

	return tmOrg, func() {
		err = tmOrg.Delete()
		check.Assert(err, IsNil)
	}
}
