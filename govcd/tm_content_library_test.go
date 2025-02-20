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

// Test_ContentLibraryProvider tests CRUD operations for a Content Library with the Provider user
func (vcd *TestVCD) Test_ContentLibraryProvider(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	cls, err := vcd.client.GetAllContentLibraries(nil, nil)
	check.Assert(err, IsNil)
	existingContentLibraryCount := len(cls)

	sc, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sc, NotNil)

	clDefinition := &types.ContentLibrary{
		Name:           check.TestName(),
		StorageClasses: []types.OpenApiReference{{ID: sc.StorageClass.ID}},
		AutoAttach:     true, // Always true for providers
		Description:    check.TestName(),
	}

	createdCl, err := vcd.client.CreateContentLibrary(clDefinition, nil)
	check.Assert(err, IsNil)
	check.Assert(createdCl, NotNil)
	AddToCleanupListOpenApi(createdCl.ContentLibrary.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraries+createdCl.ContentLibrary.ID)

	// Defer deletion for a correct cleanup
	defer func() {
		err = createdCl.Delete(true, true)
		check.Assert(err, IsNil)
	}()
	check.Assert(isUrn(createdCl.ContentLibrary.ID), Equals, true)
	check.Assert(createdCl.ContentLibrary.Name, Equals, clDefinition.Name)
	check.Assert(createdCl.ContentLibrary.Description, Equals, clDefinition.Description)
	check.Assert(len(createdCl.ContentLibrary.StorageClasses), Equals, 1)
	check.Assert(createdCl.ContentLibrary.StorageClasses[0].ID, Equals, sc.StorageClass.ID)
	check.Assert(createdCl.ContentLibrary.AutoAttach, Equals, clDefinition.AutoAttach)
	// "Computed" values
	check.Assert(createdCl.ContentLibrary.IsShared, Equals, true) // Always true for providers
	check.Assert(createdCl.ContentLibrary.IsSubscribed, Equals, false)
	check.Assert(createdCl.ContentLibrary.LibraryType, Equals, "PROVIDER")
	check.Assert(createdCl.ContentLibrary.VersionNumber, Equals, int64(1))
	check.Assert(createdCl.ContentLibrary.Org, NotNil)
	check.Assert(createdCl.ContentLibrary.Org.Name, Equals, "System")
	check.Assert(createdCl.ContentLibrary.SubscriptionConfig, IsNil)
	check.Assert(createdCl.ContentLibrary.CreationDate, Not(Equals), "")

	cls, err = vcd.client.GetAllContentLibraries(nil, nil)
	check.Assert(err, IsNil)
	check.Assert(len(cls), Equals, existingContentLibraryCount+1)
	for _, l := range cls {
		if l.ContentLibrary.ID == createdCl.ContentLibrary.ID {
			check.Assert(*l.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)
			break
		}
	}

	cl, err := vcd.client.GetContentLibraryByName(check.TestName(), nil)
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)

	cl, err = vcd.client.GetContentLibraryById(cl.ContentLibrary.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)

	// Test updating an existing Content Library
	clDefinition.Name = check.TestName() + "Updated"
	clDefinition.Description = check.TestName() + "Updated"
	updatedCl, err := createdCl.Update(clDefinition)
	check.Assert(err, IsNil)
	check.Assert(updatedCl, NotNil)
	check.Assert(updatedCl.ContentLibrary.ID, Equals, createdCl.ContentLibrary.ID)
	check.Assert(updatedCl.ContentLibrary.Name, Equals, clDefinition.Name)
	check.Assert(updatedCl.ContentLibrary.Description, Equals, clDefinition.Description)

	// Not found errors
	_, err = vcd.client.GetContentLibraryByName("notexist", nil)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetContentLibraryById("urn:vcloud:contentLibrary:aaaaaaaa-1111-0000-cccc-bbbb1111dddd", nil)
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = vcd.client.GetContentLibraryById("urn:vcloud:contentLibrary:aaaaaaaa-1111-0000-cccc-bbbb1111dddd", nil)
	check.Assert(ContainsNotFound(err), Equals, true)
}

// Test_ContentLibraryProvider tests CRUD operations for a Content Library as a tenant (Organization)
func (vcd *TestVCD) Test_ContentLibraryTenant(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check) // As it creates testing resources first

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	org, orgCleanup := createOrg(vcd, check, false)
	defer orgCleanup()

	// A Region Quota is needed to have Storage classes available in the Organization
	_, regionQuotaCleanup := createRegionQuota(vcd, org, region, check)
	defer regionQuotaCleanup()

	cls, err := org.GetAllContentLibraries(nil)
	check.Assert(err, IsNil)
	existingContentLibraryCount := len(cls)

	sc, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sc, NotNil)

	clDefinition := &types.ContentLibrary{
		Name:           check.TestName(),
		StorageClasses: []types.OpenApiReference{{ID: sc.StorageClass.ID}},
		AutoAttach:     false,
		Description:    check.TestName(),
	}

	createdCl, err := org.CreateContentLibrary(clDefinition)
	check.Assert(err, IsNil)
	check.Assert(createdCl, NotNil)
	AddToCleanupListOpenApi(createdCl.ContentLibrary.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraries+createdCl.ContentLibrary.ID)

	// Defer deletion for a correct cleanup
	defer func() {
		err = createdCl.Delete(true, true)
		check.Assert(err, IsNil)
	}()
	check.Assert(isUrn(createdCl.ContentLibrary.ID), Equals, true)
	check.Assert(createdCl.ContentLibrary.Name, Equals, clDefinition.Name)
	check.Assert(createdCl.ContentLibrary.Description, Equals, clDefinition.Description)
	check.Assert(len(createdCl.ContentLibrary.StorageClasses), Equals, 1)
	check.Assert(createdCl.ContentLibrary.StorageClasses[0].ID, Equals, sc.StorageClass.ID)
	check.Assert(createdCl.ContentLibrary.AutoAttach, Equals, clDefinition.AutoAttach)
	// "Computed" values
	check.Assert(createdCl.ContentLibrary.IsShared, Equals, false) // False for tenants
	check.Assert(createdCl.ContentLibrary.IsSubscribed, Equals, false)
	check.Assert(createdCl.ContentLibrary.LibraryType, Equals, "TENANT")
	check.Assert(createdCl.ContentLibrary.VersionNumber, Equals, int64(1))
	check.Assert(createdCl.ContentLibrary.Org, NotNil)
	check.Assert(createdCl.ContentLibrary.Org.Name, Equals, org.TmOrg.Name)
	check.Assert(createdCl.ContentLibrary.SubscriptionConfig, IsNil)
	check.Assert(createdCl.ContentLibrary.CreationDate, Not(Equals), "")

	cls, err = org.GetAllContentLibraries(nil)
	check.Assert(err, IsNil)
	check.Assert(len(cls), Equals, existingContentLibraryCount+1)
	for _, l := range cls {
		if l.ContentLibrary.ID == createdCl.ContentLibrary.ID {
			check.Assert(*l.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)
			break
		}
	}

	cl, err := org.GetContentLibraryByName(check.TestName())
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)

	cl, err = org.GetContentLibraryById(cl.ContentLibrary.ID)
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)

	// Test updating an existing Content Library
	clDefinition.Name = check.TestName() + "Updated"
	clDefinition.Description = check.TestName() + "Updated"
	updatedCl, err := createdCl.Update(clDefinition)
	check.Assert(err, IsNil)
	check.Assert(updatedCl, NotNil)
	check.Assert(updatedCl.ContentLibrary.ID, Equals, createdCl.ContentLibrary.ID)
	check.Assert(updatedCl.ContentLibrary.Name, Equals, clDefinition.Name)
	check.Assert(updatedCl.ContentLibrary.Description, Equals, clDefinition.Description)

	// Not found errors
	_, err = org.GetContentLibraryByName("notexist")
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = org.GetContentLibraryById("urn:vcloud:contentLibrary:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(err, NotNil)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = org.GetContentLibraryById("urn:vcloud:contentLibrary:aaaaaaaa-1111-0000-cccc-bbbb1111dddd")
	check.Assert(ContainsNotFound(err), Equals, true)
}

// Test_ContentLibrarySubscribed tests CRUD operations for a Content Library that is subscribed to another from vCenter
func (vcd *TestVCD) Test_ContentLibrarySubscribed(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)
	if vcd.config.Tm.SubscriptionContentLibraryUrl == "" {
		check.Skip("test configuration tm.subscriptionContentLibraryUrl is empty")
	}

	// Certificate must be trusted before adding subscribed content library
	url, err := url.Parse(vcd.config.Tm.SubscriptionContentLibraryUrl)
	check.Assert(err, IsNil)
	trustedCert, err := vcd.client.AutoTrustCertificate(url)
	check.Assert(err, IsNil)
	if trustedCert != nil {
		AddToCleanupListOpenApi(trustedCert.TrustedCertificate.ID, check.TestName()+"trusted-cert", types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTrustedCertificates+trustedCert.TrustedCertificate.ID)
	}

	nsxtManager, nsxtManagerCleanup := getOrCreateNsxtManager(vcd, check)
	defer nsxtManagerCleanup()
	vc, vcCleanup := getOrCreateVCenter(vcd, check)
	defer vcCleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)

	region, regionCleanup := getOrCreateRegion(vcd, nsxtManager, supervisor, check)
	defer regionCleanup()

	cls, err := vcd.client.GetAllContentLibraries(nil, nil)
	check.Assert(err, IsNil)
	existingContentLibraryCount := len(cls)

	sc, err := region.GetStorageClassByName(vcd.config.Tm.StorageClass)
	check.Assert(err, IsNil)
	check.Assert(sc, NotNil)

	clDefinition := &types.ContentLibrary{
		Name:           check.TestName(),
		StorageClasses: []types.OpenApiReference{{ID: sc.StorageClass.ID}},
		Description:    check.TestName(), // This should be ignored as it takes publisher description
		AutoAttach:     true,             // Always true for providers
		SubscriptionConfig: &types.ContentLibrarySubscriptionConfig{
			SubscriptionUrl: vcd.config.Tm.SubscriptionContentLibraryUrl,
			NeedLocalCopy:   true,
		},
	}

	createdCl, err := vcd.client.CreateContentLibrary(clDefinition, nil)
	check.Assert(err, IsNil)
	check.Assert(createdCl, NotNil)
	AddToCleanupListOpenApi(createdCl.ContentLibrary.Name, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointContentLibraries+createdCl.ContentLibrary.ID)

	// Defer deletion for a correct cleanup
	defer func() {
		err = createdCl.Delete(true, true)
		check.Assert(err, IsNil)
	}()
	check.Assert(isUrn(createdCl.ContentLibrary.ID), Equals, true)
	check.Assert(createdCl.ContentLibrary.Name, Equals, clDefinition.Name)
	check.Assert(createdCl.ContentLibrary.Description, Not(Equals), clDefinition.Description) // It takes publisher description
	check.Assert(len(createdCl.ContentLibrary.StorageClasses), Equals, 1)
	check.Assert(createdCl.ContentLibrary.StorageClasses[0].ID, Equals, sc.StorageClass.ID)
	check.Assert(createdCl.ContentLibrary.AutoAttach, Equals, clDefinition.AutoAttach)
	// "Computed" values
	check.Assert(createdCl.ContentLibrary.IsShared, Equals, true) // Always true for providers
	check.Assert(createdCl.ContentLibrary.IsSubscribed, Equals, true)
	check.Assert(createdCl.ContentLibrary.LibraryType, Equals, "PROVIDER")
	check.Assert(createdCl.ContentLibrary.VersionNumber >= int64(0), Equals, true) // Version can differ from 0
	check.Assert(createdCl.ContentLibrary.Org, NotNil)
	check.Assert(createdCl.ContentLibrary.Org.Name, Equals, "System")
	check.Assert(createdCl.ContentLibrary.SubscriptionConfig, NotNil)
	check.Assert(createdCl.ContentLibrary.SubscriptionConfig.SubscriptionUrl, Equals, vcd.config.Tm.SubscriptionContentLibraryUrl)
	check.Assert(createdCl.ContentLibrary.SubscriptionConfig.NeedLocalCopy, Equals, true)
	check.Assert(createdCl.ContentLibrary.SubscriptionConfig.Password, Equals, "******") // Password is returned sealed
	check.Assert(createdCl.ContentLibrary.CreationDate, Not(Equals), "")

	cls, err = vcd.client.GetAllContentLibraries(nil, nil)
	check.Assert(err, IsNil)
	check.Assert(len(cls), Equals, existingContentLibraryCount+1)
	for _, l := range cls {
		if l.ContentLibrary.ID == createdCl.ContentLibrary.ID {
			check.Assert(*l.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)
			break
		}
	}

	cl, err := vcd.client.GetContentLibraryByName(check.TestName(), nil)
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)

	cl, err = vcd.client.GetContentLibraryById(cl.ContentLibrary.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(cl, NotNil)
	check.Assert(*cl.ContentLibrary, DeepEquals, *createdCl.ContentLibrary)
}
