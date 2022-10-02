/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// AdminCatalog is a admin view of a VMware Cloud Director Catalog
// To be able to get an AdminCatalog representation, users must have
// admin credentials to the System org. AdminCatalog is used
// for creating, updating, and deleting a Catalog.
// Definition: https://code.vmware.com/apis/220/vcloud#/doc/doc/types/AdminCatalogType.html
type AdminCatalog struct {
	AdminCatalog *types.AdminCatalog
	client       *Client
	parent       organization
}

func NewAdminCatalog(client *Client) *AdminCatalog {
	return &AdminCatalog{
		AdminCatalog: new(types.AdminCatalog),
		client:       client,
	}
}

func NewAdminCatalogWithParent(client *Client, parent organization) *AdminCatalog {
	return &AdminCatalog{
		AdminCatalog: new(types.AdminCatalog),
		client:       client,
		parent:       parent,
	}
}

// Delete deletes the Catalog, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Catalog.html
func (adminCatalog *AdminCatalog) Delete(force, recursive bool) error {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.Delete(force, recursive)
}

// Update updates the Catalog definition from current Catalog struct contents.
// Any differences that may be legally applied will be updated.
// Returns an error if the call to vCD fails. Update automatically performs
// a refresh with the admin catalog it gets back from the rest api
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/PUT-Catalog.html
func (adminCatalog *AdminCatalog) Update() error {
	reqCatalog := &types.Catalog{
		Name:        adminCatalog.AdminCatalog.Catalog.Name,
		Description: adminCatalog.AdminCatalog.Description,
	}
	vcomp := &types.AdminCatalog{
		Xmlns:                  types.XMLNamespaceVCloud,
		Catalog:                *reqCatalog,
		CatalogStorageProfiles: adminCatalog.AdminCatalog.CatalogStorageProfiles,
		IsPublished:            adminCatalog.AdminCatalog.IsPublished,
	}
	catalog := &types.AdminCatalog{}
	_, err := adminCatalog.client.ExecuteRequest(adminCatalog.AdminCatalog.HREF, http.MethodPut,
		"application/vnd.vmware.admin.catalog+xml", "error updating catalog: %s", vcomp, catalog)
	adminCatalog.AdminCatalog = catalog
	return err
}

// UploadOvf uploads an ova file to a catalog. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process. On upload fail client may need to
// remove vCD catalog item which waits for files to be uploaded. Files from ova are extracted to system
// temp folder "govcd+random number" and left for inspection on error.
func (adminCatalog *AdminCatalog) UploadOvf(ovaFileName, itemName, description string, uploadPieceSize int64) (UploadTask, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	catalog.parent = adminCatalog.parent
	return catalog.UploadOvf(ovaFileName, itemName, description, uploadPieceSize)
}

func (adminCatalog *AdminCatalog) Refresh() error {
	if *adminCatalog == (AdminCatalog{}) || adminCatalog.AdminCatalog.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty or HREF is empty")
	}

	refreshedCatalog := &types.AdminCatalog{}

	_, err := adminCatalog.client.ExecuteRequest(adminCatalog.AdminCatalog.HREF, http.MethodGet,
		"", "error refreshing VDC: %s", nil, refreshedCatalog)
	if err != nil {
		return err
	}
	adminCatalog.AdminCatalog = refreshedCatalog

	return nil
}

// getOrgInfo finds the organization to which the admin catalog belongs, and returns its name and ID
func (adminCatalog *AdminCatalog) getOrgInfo() (*TenantContext, error) {
	return adminCatalog.getTenantContext()
}

// PublishToExternalOrganizations publishes a catalog to external organizations.
func (cat *AdminCatalog) PublishToExternalOrganizations(publishExternalCatalog types.PublishExternalCatalogParams) error {
	if cat.AdminCatalog == nil {
		return fmt.Errorf("cannot publish to external organization, Object is empty")
	}

	url := cat.AdminCatalog.HREF
	if url == "nil" || url == "" {
		return fmt.Errorf("cannot publish to external organization, HREF is empty")
	}

	adminOrg, err := cat.GetAdminParent()
	if err != nil {
		return fmt.Errorf("cannot get parent organization for catalog %s: %s", cat.AdminCatalog.Name, err)
	}
	if !adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs {
		return fmt.Errorf("parent organization %s of catalog %s can't publish catalogs", adminOrg.AdminOrg.Name, cat.AdminCatalog.Name)
	}
	if !adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally {
		return fmt.Errorf("parent organization %s of catalog %s can't publish to external orgs", adminOrg.AdminOrg.Name, cat.AdminCatalog.Name)
	}

	tenantContext, err := cat.getTenantContext()
	if err != nil {
		return fmt.Errorf("cannot publish to external organization, tenant context error: %s", err)
	}

	err = publishToExternalOrganizations(cat.client, url, tenantContext, publishExternalCatalog)
	if err != nil {
		return err
	}

	err = cat.Refresh()
	if err != nil {
		return err
	}

	return err
}

// GetAdminParent returns the OrAdming to which the catalog belongs
func (cat *AdminCatalog) GetAdminParent() (*AdminOrg, error) {
	adminOrg, _, err := getCatalogParent(cat.AdminCatalog.Name, cat.client, cat.AdminCatalog.Link, false)
	if err != nil {
		return nil, err
	}
	return adminOrg, nil
}

// GetParent returns the Org to which the catalog belongs
func (cat *AdminCatalog) GetParent() (*Org, error) {
	_, org, err := getCatalogParent(cat.AdminCatalog.Name, cat.client, cat.AdminCatalog.Link, false)
	if err != nil {
		return nil, err
	}
	return org, nil
}

