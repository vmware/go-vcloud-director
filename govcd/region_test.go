//go:build tm || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmRegion(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	// TODO - remove from here
	// rCleanup, err := vcd.client.GetRegionByName(check.TestName())
	// check.Assert(err, IsNil)
	// check.Assert(rCleanup, NotNil)
	// err = rCleanup.Delete()
	// check.Assert(err, IsNil)
	// check.Assert(true, Equals, false)
	// TODO - remove until here

	vc, nsxtManager := getOrCreateVcAndNsxtManager(vcd, check)
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

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

	createdRegion, err := vcd.client.CreateRegion(r)
	check.Assert(err, IsNil)
	check.Assert(createdRegion.Region, NotNil)
	AddToCleanupListOpenApi(createdRegion.Region.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointRegions+createdRegion.Region.ID)

	check.Assert(createdRegion.Region.Status, Equals, "READY") // Region is operational

	// Get By Name
	byName, err := vcd.client.GetRegionByName(r.Name)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetRegionById(createdRegion.Region.ID)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	check.Assert(byName.Region, DeepEquals, byId.Region)

	// Get All
	allRegions, err := vcd.client.GetAllRegions(nil)
	check.Assert(err, IsNil)
	check.Assert(allRegions, NotNil)
	check.Assert(len(allRegions) > 0, Equals, true)

	// TODO: TM: No Update so far
	// Update
	// createdRegion.Region.IsEnabled = false
	// updated, err := createdRegion.Update(createdRegion.Region)
	// check.Assert(err, IsNil)
	// check.Assert(updated, NotNil)

	// Delete
	err = createdRegion.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetRegionByName(createdRegion.Region.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

}

// getOrCreateVcAndNsxtManager will check configuration file
func getOrCreateVcAndNsxtManager(vcd *TestVCD, check *C) (*VCenter, *NsxtManagerOpenApi) {
	vc, err := vcd.client.GetVCenterByUrl(vcd.config.Tm.VcenterUrl)
	if ContainsNotFound(err) && !vcd.config.Tm.CreateVcenter {
		check.Skip("vCenter is not configured and configuration is not allowed in config file")
	}
	if ContainsNotFound(err) {
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

		printVerbose("# Refreshing Storage Profiles %s\n", vc.VSphereVCenter.Url)
		err = vc.RefreshStorageProfiles()
		check.Assert(err, IsNil)
	}

	nsxtManager, err := vcd.client.GetNsxtManagerOpenApiByUrl(vcd.config.Tm.NsxtManagerUrl)
	if ContainsNotFound(err) && !vcd.config.Tm.CreateNsxtManager {
		check.Skip("NSX-T Manager is not configured and configuration is not allowed in config file")
	}
	if ContainsNotFound(err) {
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
	}

	return vc, nsxtManager
}
