//go:build api || openapi || functional || catalog || vapp || gateway || network || org || query || extnetwork || task || vm || vdc || system || disk || lb || lbAppRule || lbAppProfile || lbServerPool || lbServiceMonitor || lbVirtualServer || user || search || nsxv || nsxt || auth || affinity || role || alb || certificate || vdcGroup || metadata || providervdc || rde || vsphere || uiPlugin || cse || slz || tm || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"time"

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
	}
	if !vcd.config.Tm.CreateVcenter {
		check.Skip("vCenter is not configured and configuration is not allowed in config file")
	}
	printVerbose("# Will create vCenter %s\n", vcd.config.Tm.VcenterUrl)
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
		printVerbose("# Certificate for vCenter is trusted %s\n", trustedCert.TrustedCertificate.ID)
		AddToCleanupListOpenApi(trustedCert.TrustedCertificate.ID, check.TestName()+"trusted-cert", types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTrustedCertificates+trustedCert.TrustedCertificate.ID)
	}

	vc, err = vcd.client.CreateVcenter(vcCfg)
	check.Assert(err, IsNil)
	check.Assert(vc, NotNil)
	PrependToCleanupList(vcCfg.Name, "OpenApiEntityVcenter", check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVirtualCenters+vc.VSphereVCenter.VcId)

	printVerbose("# Sleeping after vCenter creation\n")
	time.Sleep(1 * time.Minute) // TODO: TM: Reevaluate need for sleep
	// Refresh connected vCenter to be sure that all artifacts are loaded
	printVerbose("# Refreshing vCenter %s\n", vc.VSphereVCenter.Url)
	err = vc.RefreshVcenter()
	check.Assert(err, IsNil)

	printVerbose("# Refreshing Storage Profiles in vCenter %s\n", vc.VSphereVCenter.Url)
	err = vc.RefreshStorageProfiles()
	check.Assert(err, IsNil)

	printVerbose("# Sleeping after vCenter refreshes\n")
	time.Sleep(1 * time.Minute) // TODO: TM: Reevaluate need for sleep
	vCenterCreated := true

	return vc, func() {
		if !vCenterCreated {
			return
		}
		printVerbose("# Disabling and deleting vCenter %s\n", vcd.config.Tm.VcenterUrl)
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
	}
	if !vcd.config.Tm.CreateNsxtManager {
		check.Skip("NSX-T Manager is not configured and configuration is not allowed in config file")
	}

	printVerbose("# Will create NSX-T Manager %s\n", vcd.config.Tm.NsxtManagerUrl)
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
		printVerbose("# Certificate for NSX-T Manager is trusted %s\n", trustedCert.TrustedCertificate.ID)
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
		printVerbose("# Deleting NSX-T Manager %s\n", nsxtManager.NsxtManagerOpenApi.Name)
		err = nsxtManager.Delete()
		check.Assert(err, IsNil)
	}
}

// getOrCreateRegion will check configuration file and create a Region if
// stated in the 'createRegion' testing property not present in TM.
// Otherwise, it just retrieves it
func getOrCreateRegion(vcd *TestVCD, nsxtManager *NsxtManagerOpenApi, supervisor *Supervisor, check *C) (*Region, func()) {
	if vcd.config.Tm.Region == "" {
		check.Fatal("testing configuration property 'tm.region' is required")
	}
	region, err := vcd.client.GetRegionByName(vcd.config.Tm.Region)
	if err == nil {
		return region, func() {}
	}
	if !ContainsNotFound(err) {
		check.Fatal(err)
	}
	if !vcd.config.Tm.CreateRegion {
		check.Skip("Region is not configured and configuration is not allowed in config file")
	}
	if nsxtManager == nil || supervisor == nil {
		check.Fatalf("getOrCreateRegion requires a not nil NSX-T Manager and Supervisor")
	}

	r := &types.Region{
		Name: vcd.config.Tm.Region,
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
	check.Assert(region.Region.Status, Equals, "READY") // Region must be READY to be operational

	return region, func() {
		if !regionCreated {
			return
		}
		printVerbose("# Deleting Region %s\n", region.Region.Name)
		err = region.Delete()
		check.Assert(err, IsNil)
	}
}

// getOrCreateContentLibrary will check configuration file and create a Content Library if
// not present in TM. Otherwise, it just retrieves it
func getOrCreateContentLibrary(vcd *TestVCD, storagePolicy *RegionStoragePolicy, check *C) (*ContentLibrary, func()) {
	if vcd.config.Tm.ContentLibrary == "" {
		check.Fatal("testing configuration property 'tm.contentLibrary' is required")
	}
	cl, err := vcd.client.GetContentLibraryByName(vcd.config.Tm.ContentLibrary)
	if err == nil {
		return cl, func() {}
	}
	if !ContainsNotFound(err) {
		check.Fatal(err)
	}

	payload := types.ContentLibrary{
		Name: vcd.config.Tm.ContentLibrary,
		StorageClasses: types.OpenApiReferences{{
			Name: storagePolicy.RegionStoragePolicy.Name,
			ID:   storagePolicy.RegionStoragePolicy.ID,
		}},
		Description: check.TestName(),
	}

	contentLibrary, err := vcd.client.CreateContentLibrary(&payload)
	check.Assert(err, IsNil)
	check.Assert(contentLibrary, NotNil)
	contentLibraryCreated := true
	AddToCleanupListOpenApi(contentLibrary.ContentLibrary.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraries+contentLibrary.ContentLibrary.ID)

	return contentLibrary, func() {
		if !contentLibraryCreated {
			return
		}
		err = contentLibrary.Delete()
		check.Assert(err, IsNil)
	}
}

func createOrg(vcd *TestVCD, check *C, canManageOrgs bool) (*TmOrg, func()) {
	cfg := &types.TmOrg{
		Name:          check.TestName(),
		DisplayName:   check.TestName(),
		CanManageOrgs: canManageOrgs,
		IsEnabled:     true,
	}
	tmOrg, err := vcd.client.CreateTmOrg(cfg)
	check.Assert(err, IsNil)
	check.Assert(tmOrg, NotNil)

	PrependToCleanupListOpenApi(tmOrg.TmOrg.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgs+tmOrg.TmOrg.ID)

	return tmOrg, func() {
		if tmOrg.TmOrg.IsEnabled {
			err = tmOrg.Disable()
			check.Assert(err, IsNil)
		}
		err = tmOrg.Delete()
		check.Assert(err, IsNil)
	}
}

func createTmIpSpace(vcd *TestVCD, region *Region, check *C, nameSuffix, octet3 string) (*TmIpSpace, func()) {
	ipSpaceType := &types.TmIpSpace{
		Name:        check.TestName() + "-" + nameSuffix,
		RegionRef:   types.OpenApiReference{ID: region.Region.ID},
		Description: check.TestName(),
		DefaultQuota: types.TmIpSpaceDefaultQuota{
			MaxCidrCount:  3,
			MaxIPCount:    -1,
			MaxSubnetSize: 24,
		},
		ExternalScopeCidr: fmt.Sprintf("12.12.%s.0/30", octet3),
		InternalScopeCidrBlocks: []types.TmIpSpaceInternalScopeCidrBlocks{
			{
				Cidr: fmt.Sprintf("10.0.%s.0/24", octet3),
			},
		},
	}

	ipSpace, err := vcd.client.CreateTmIpSpace(ipSpaceType)
	check.Assert(err, IsNil)
	check.Assert(ipSpace, NotNil)
	AddToCleanupListOpenApi(ipSpace.TmIpSpace.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmIpSpaces+ipSpace.TmIpSpace.ID)

	return ipSpace, func() {
		printVerbose("# Deleting IP Space %s\n", ipSpace.TmIpSpace.Name)
		err = ipSpace.Delete()
		check.Assert(err, IsNil)
	}
}
