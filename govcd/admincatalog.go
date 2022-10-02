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

// CreateCatalogFromSubscription creates a new catalog by subscribing to a published catalog
// Parameter subscription has to be filled manually
func (org *AdminOrg) CreateCatalogFromSubscription(subscription types.ExternalCatalogSubscription,
	storageProfiles *types.CatalogStorageProfiles,
	catalogName, catalogDescription, password string, localCopy bool) (*AdminCatalog, error) {

	// If the receiving Org doesn't have any VDCs, it means that there is no storage that can be used
	// by a catalog
	if len(org.AdminOrg.Vdcs.Vdcs) == 0 {
		return nil, fmt.Errorf("org %s does not have any storage to support a catalog", org.AdminOrg.Name)
	}
	href := ""

	// The subscibed catalog creation is like a regular catalog creation, with the
	// difference that the subscription details are filled in
	for _, link := range org.AdminOrg.Link {
		if link.Rel == "add" && link.Type == types.MimeAdminCatalog {
			href = link.HREF
			break
		}
	}
	if href == "" {
		return nil, fmt.Errorf("catalog creation link not found for org %s", org.AdminOrg.Name)
	}
	reqCatalog := &types.Catalog{
		Name:        catalogName,
		Description: catalogDescription,
	}
	adminCatalog := types.AdminCatalog{
		Xmlns:                  types.XMLNamespaceVCloud,
		Catalog:                *reqCatalog,
		CatalogStorageProfiles: storageProfiles,
		ExternalCatalogSubscription: &types.ExternalCatalogSubscription{
			LocalCopy:                localCopy,
			Password:                 password,
			Location:                 subscription.Location,
			SubscribeToExternalFeeds: true,
		},
	}
	// The subscription URL returned by the API is in abbreviated form
	// such as "/vcsp/lib/65637586-c703-48ae-a7e2-82605d18db57/"
	// If the passed URL is so abbreviated, we need to add the host
	if !IsValidUrl(subscription.Location) {
		// Get the catalog base URL
		cutPosition := strings.Index(org.AdminOrg.HREF, "/api")
		catalogHost := org.AdminOrg.HREF[:cutPosition]
		subscriptionUrl, err := url.JoinPath(catalogHost, subscription.Location)
		if err != nil {
			return nil, fmt.Errorf("error composing subscription URL: %s", err)
		}
		adminCatalog.ExternalCatalogSubscription.Location = subscriptionUrl
	}

	adminCatalog.ExternalCatalogSubscription.Password = password
	adminCatalog.ExternalCatalogSubscription.LocalCopy = localCopy
	_, err := org.client.ExecuteRequest(href, http.MethodPost, types.MimeAdminCatalog,
		"error subscribing to catalog: %s", adminCatalog, nil)
	if err != nil {
		return nil, err
	}
	return org.GetAdminCatalogByName(catalogName, true)
}

// ImportFromCatalog creates a new catalog by subscribing to an existing catalog
// The subscription parameters are gathered, as much as possible, from the published catalog itself
func (org *AdminOrg) ImportFromCatalog(fromCatalog *AdminCatalog, profiles *types.CatalogStorageProfiles,
	catalogName, catalogDescription, password string, localCopy bool) (*AdminCatalog, error) {
	err := fromCatalog.Refresh()

	if err != nil {
		return nil, fmt.Errorf("error refreshing catalog %s: %s", fromCatalog.AdminCatalog.Name, err)
	}

	if fromCatalog.AdminCatalog.PublishExternalCatalogParams == nil {
		return nil, fmt.Errorf("catalog '%s' has not acivated its subscription", fromCatalog.AdminCatalog.Name)
	}

	params := types.ExternalCatalogSubscription{
		SubscribeToExternalFeeds: true,
		Location:                 fromCatalog.AdminCatalog.PublishExternalCatalogParams.CatalogPublishedUrl,
		Password:                 password,
		LocalCopy:                localCopy,
	}
	return org.CreateCatalogFromSubscription(params, profiles, catalogName, catalogDescription, password, localCopy)
}

// IsValidUrl returns true if the given URL is complete and usable
func IsValidUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
