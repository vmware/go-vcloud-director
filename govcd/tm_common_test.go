//go:build api || openapi || functional || catalog || vapp || gateway || network || org || query || extnetwork || task || vm || vdc || system || disk || lb || lbAppRule || lbAppProfile || lbServerPool || lbServiceMonitor || lbVirtualServer || user || search || nsxv || nsxt || auth || affinity || role || alb || certificate || vdcGroup || metadata || providervdc || rde || vsphere || uiPlugin || cse || slz || tm || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

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

	printVerbose("# Waiting for listener status to become 'CONNECTED'\n")
	err = waitForListenerStatusConnected(vc)
	check.Assert(err, IsNil)
	printVerbose("# Sleeping after vCenter is 'CONNECTED'\n")
	time.Sleep(4 * time.Second) // TODO: TM: Re-evaluate need for sleep
	// Refresh connected vCenter to be sure that all artifacts are loaded
	printVerbose("# Refreshing vCenter %s\n", vc.VSphereVCenter.Url)
	err = vc.RefreshVcenter()
	check.Assert(err, IsNil)

	printVerbose("# Refreshing Storage Profiles in vCenter %s\n", vc.VSphereVCenter.Url)
	err = vc.RefreshStorageProfiles()
	check.Assert(err, IsNil)

	printVerbose("# Sleeping after vCenter refreshes\n")
	time.Sleep(1 * time.Minute) // TODO: TM: Re-evaluate need for sleep
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

func waitForListenerStatusConnected(v *VCenter) error {
	startTime := time.Now()
	tryCount := 20
	for c := 0; c < tryCount; c++ {
		err := v.Refresh()
		if err != nil {
			return fmt.Errorf("error refreshing vCenter: %s", err)
		}

		if v.VSphereVCenter.ListenerState == "CONNECTED" {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("waiting for listener state to become 'CONNECTED' expired after %d tries (%d seconds), got '%s'",
		tryCount, int(time.Since(startTime)/time.Second), v.VSphereVCenter.ListenerState)
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
		Name:        check.TestName(),
		Username:    vcd.config.Tm.NsxtManagerUsername,
		Description: check.TestName(), // TODO: TM: Latest build throws SQL error if not populated
		Password:    vcd.config.Tm.NsxtManagerPassword,
		Url:         vcd.config.Tm.NsxtManagerUrl,
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
func getOrCreateContentLibrary(vcd *TestVCD, storageClass *StorageClass, check *C) (*ContentLibrary, func()) {
	if vcd.config.Tm.ContentLibrary == "" {
		check.Fatal("testing configuration property 'tm.contentLibrary' is required")
	}
	cl, err := vcd.client.GetContentLibraryByName(vcd.config.Tm.ContentLibrary, nil)
	if err == nil {
		return cl, func() {}
	}
	if !ContainsNotFound(err) {
		check.Fatal(err)
	}

	payload := types.ContentLibrary{
		Name: vcd.config.Tm.ContentLibrary,
		StorageClasses: types.OpenApiReferences{{
			Name: storageClass.StorageClass.Name,
			ID:   storageClass.StorageClass.ID,
		}},
		Description: check.TestName(),
	}

	contentLibrary, err := vcd.client.CreateContentLibrary(&payload, nil)
	check.Assert(err, IsNil)
	check.Assert(contentLibrary, NotNil)
	contentLibraryCreated := true
	AddToCleanupListOpenApi(contentLibrary.ContentLibrary.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraries+contentLibrary.ContentLibrary.ID)

	return contentLibrary, func() {
		if !contentLibraryCreated {
			return
		}
		err = contentLibrary.Delete(true, true)
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

// Creates a VDC (Region Quota) for testing in Tenant Manager and configures it with
// the first found VM class and the configured Storage Class.
func createVdc(vcd *TestVCD, org *TmOrg, region *Region, check *C) (*RegionQuota, func()) {
	if vcd.config.Tm.StorageClass == "" {
		check.Fatal("testing configuration property 'tm.storageClass' is required")
	}
	if org == nil || org.TmOrg == nil {
		check.Fatal("an Organization is required to create the Region Quota")
	}
	if region == nil || region.Region == nil {
		check.Fatal("a Region is required to create the Region Quota")
	}
	regionZones, err := region.GetAllZones(nil)
	check.Assert(err, IsNil)
	check.Assert(len(regionZones) > 0, Equals, true)

	vmClasses, err := region.GetAllVmClasses(nil)
	check.Assert(err, IsNil)
	check.Assert(len(vmClasses) > 0, Equals, true)

	sp, err := region.GetStoragePolicyByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sp, NotNil)

	cfg := &types.TmVdc{
		Name: fmt.Sprintf("%s_%s", org.TmOrg.Name, region.Region.Name),
		Org: &types.OpenApiReference{
			Name: org.TmOrg.Name,
			ID:   org.TmOrg.ID,
		},
		Region: &types.OpenApiReference{
			Name: region.Region.Name,
			ID:   region.Region.ID,
		},
		Supervisors: region.Region.Supervisors,
		ZoneResourceAllocation: []*types.TmVdcZoneResourceAllocation{{
			Zone: &types.OpenApiReference{ID: regionZones[0].Zone.ID},
			ResourceAllocation: types.TmVdcResourceAllocation{
				CPUReservationMHz:    100,
				CPULimitMHz:          500,
				MemoryReservationMiB: 256,
				MemoryLimitMiB:       512,
			},
		}},
	}
	vdc, err := vcd.client.CreateRegionQuota(cfg)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	PrependToCleanupListOpenApi(vdc.TmVdc.ID, cfg.Name, types.OpenApiPathVcf+types.OpenApiEndpointTmVdcs+vdc.TmVdc.ID)

	err = vdc.AssignVmClasses(&types.RegionVirtualMachineClasses{
		Values: types.OpenApiReferences{{Name: vmClasses[0].Name, ID: vmClasses[0].ID}},
	})
	check.Assert(err, IsNil)
	_, err = vdc.CreateStoragePolicies(&types.VirtualDatacenterStoragePolicies{
		Values: []types.VirtualDatacenterStoragePolicy{
			{
				RegionStoragePolicy: types.OpenApiReference{
					ID: sp.RegionStoragePolicy.ID,
				},
				StorageLimitMiB: 100,
				VirtualDatacenter: types.OpenApiReference{
					ID: vdc.TmVdc.ID,
				},
			},
		},
	})
	check.Assert(err, IsNil)

	return vdc, func() {
		err = vdc.Delete()
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

func createTmProviderGateway(vcd *TestVCD, region *Region, check *C) (*TmProviderGateway, func()) {
	ipSpace, ipSpaceCleanup1 := createTmIpSpace(vcd, region, check, "1", "0")

	t0ByNameInRegion, err := vcd.client.GetTmTier0GatewayWithContextByName(vcd.config.Tm.NsxtTier0Gateway, region.Region.ID, false)
	check.Assert(err, IsNil)
	check.Assert(t0ByNameInRegion, NotNil)

	t := &types.TmProviderGateway{
		Name:        check.TestName(),
		Description: check.TestName(),
		BackingType: "NSX_TIER0",
		BackingRef:  types.OpenApiReference{ID: t0ByNameInRegion.TmTier0Gateway.ID},
		RegionRef:   types.OpenApiReference{ID: region.Region.ID},
		IPSpaceRefs: []types.OpenApiReference{{
			ID: ipSpace.TmIpSpace.ID,
		}},
	}

	pg, err := vcd.client.CreateTmProviderGateway(t)
	check.Assert(err, IsNil)
	check.Assert(pg, NotNil)
	AddToCleanupListOpenApi(pg.TmProviderGateway.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointTmProviderGateways+pg.TmProviderGateway.ID)

	return pg, func() {
		printVerbose("# Deleting Provider Gateway %s\n", pg.TmProviderGateway.Name)
		err = pg.Delete()
		check.Assert(err, IsNil)

		ipSpaceCleanup1()
	}
}

func setOrgShortLogname(vcd *TestVCD, org *TmOrg, check *C) func() {
	t := &types.TmOrgNetworkingSettings{OrgNameForLogs: "test"}
	_, err := org.UpdateOrgNetworkingSettings(t)
	check.Assert(err, IsNil)

	return func() {
		t := &types.TmOrgNetworkingSettings{OrgNameForLogs: ""}
		_, err := org.UpdateOrgNetworkingSettings(t)
		check.Assert(err, IsNil)
	}
}
