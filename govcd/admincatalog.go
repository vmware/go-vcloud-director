/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
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

// CreateCatalogFromSubscriptionAsync creates a new catalog by subscribing to a published catalog
// Parameter subscription needs to be filled manually
func (org *AdminOrg) CreateCatalogFromSubscriptionAsync(subscription types.ExternalCatalogSubscription,
	storageProfiles *types.CatalogStorageProfiles,
	catalogName, password string, localCopy bool) (*AdminCatalog, error) {

	// If the receiving Org doesn't have any VDCs, it means that there is no storage that can be used
	// by a catalog
	if len(org.AdminOrg.Vdcs.Vdcs) == 0 {
		return nil, fmt.Errorf("org %s does not have any storage to support a catalog", org.AdminOrg.Name)
	}
	href := ""

	// The subscribed catalog creation is like a regular catalog creation, with the
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
		Name: catalogName,
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

// FullSubscriptionUrl returns the subscription URL from a publisher catalog
// adding the HOST if needed
func (cat *AdminCatalog) FullSubscriptionUrl() string {
	if cat.AdminCatalog.PublishExternalCatalogParams == nil {
		return ""
	}
	subscriptionUrl := cat.AdminCatalog.PublishExternalCatalogParams.CatalogPublishedUrl
	var err error
	if !IsValidUrl(subscriptionUrl) {
		// Get the catalog base URL
		cutPosition := strings.Index(cat.AdminCatalog.HREF, "/api")
		catalogHost := cat.AdminCatalog.HREF[:cutPosition]
		subscriptionUrl, err = url.JoinPath(catalogHost, subscriptionUrl)
		if err != nil {
			return ""
		}
	}
	return subscriptionUrl
}

// IsValidUrl returns true if the given URL is complete and usable
func IsValidUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// CreateCatalogFromSubscription is a wrapper around CreateCatalogFromSubscriptionAsync
// After catalog creation, it waits for the import tasks to complete within a given timeout
func (org *AdminOrg) CreateCatalogFromSubscription(subscription types.ExternalCatalogSubscription,
	storageProfiles *types.CatalogStorageProfiles,
	catalogName, password string, localCopy bool, timeout time.Duration) (*AdminCatalog, error) {
	noTimeout := timeout == 0
	adminCatalog, err := org.CreateCatalogFromSubscriptionAsync(subscription, storageProfiles, catalogName, password, localCopy)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	for noTimeout || time.Since(start) < timeout {
		if noTimeout {
			util.Logger.Printf("[TRACE] [CreateCatalogFromSubscription] no timeout given - Elapsed %s", time.Since(start))
		}
		err = adminCatalog.Refresh()
		if err != nil {
			return nil, err
		}
		if ResourceComplete(adminCatalog.AdminCatalog.Tasks) {
			return adminCatalog, nil
		}
	}
	return nil, fmt.Errorf("adminCatalog %s still not complete after %s", adminCatalog.AdminCatalog.Name, timeout)
}

// Sync synchronises a subscribed AdminCatalog
func (cat *AdminCatalog) Sync() error {
	// if the catalog was not subscribed, return
	if cat.AdminCatalog.ExternalCatalogSubscription == nil || cat.AdminCatalog.ExternalCatalogSubscription.Location == "" {
		return nil
	}
	// The sync operation is only available for Catalog, not AdminCatalog.
	// We use the embedded Catalog object for this purpose
	catalogHref, err := cat.GetCatalogHref()
	if err != nil || catalogHref == "" {
		return fmt.Errorf("empty catalog HREF for admin catalog %s", cat.AdminCatalog.Name)
	}
	return elementSync(cat.client, catalogHref, "admin catalog")
}

// LaunchSync starts synchronisation of a subscribed AdminCatalog
func (cat *AdminCatalog) LaunchSync() (*Task, error) {
	// if the catalog was not subscribed, return
	if cat.AdminCatalog.ExternalCatalogSubscription == nil || cat.AdminCatalog.ExternalCatalogSubscription.Location == "" {
		return nil, nil
	}
	// The sync operation is only available for Catalog, not AdminCatalog.
	// We use the embedded Catalog object for this purpose
	catalogHref, err := cat.GetCatalogHref()
	if err != nil || catalogHref == "" {
		return nil, fmt.Errorf("empty catalog HREF for admin catalog %s", cat.AdminCatalog.Name)
	}
	return elementLaunchSync(cat.client, catalogHref, "admin catalog")
}

// GetCatalogHref retrieves the regular catalog HREF from an admin catalog
func (cat *AdminCatalog) GetCatalogHref() (string, error) {
	href := ""
	for _, link := range cat.AdminCatalog.Link {
		if link.Rel == "alternate" && link.Type == types.MimeCatalog {
			href = link.HREF
			break
		}
	}
	if href == "" {
		return "", fmt.Errorf("no regular Catalog HREF found for admin Catalog %s", cat.AdminCatalog.Name)
	}
	return href, nil
}

// QueryVappTemplateList returns a list of vApp templates for the given catalog
func (catalog *AdminCatalog) QueryVappTemplateList() ([]*types.QueryResultVappTemplateType, error) {
	return queryVappTemplateListWithFilter(catalog.client, map[string]string{"catalogName": catalog.AdminCatalog.Name})
}

// QueryMediaList retrieves a list of media items for the Admin Catalog
func (catalog *AdminCatalog) QueryMediaList() ([]*types.MediaRecordType, error) {
	return queryMediaList(catalog.client, catalog.AdminCatalog.HREF)
}

// LaunchSynchronisationVappTemplates starts synchronisation of a list of vApp templates
func (cat *AdminCatalog) LaunchSynchronisationVappTemplates(nameList []string) ([]*Task, error) {
	var taskList []*Task
	for _, element := range nameList {
		catalogItem, err := cat.QueryCatalogItem(element)
		if err != nil {
			return nil, err
		}
		task, err := queryResultCatalogItemToCatalogItem(cat.client, catalogItem).LaunchSync()
		if err != nil {
			return nil, err
		}
		taskList = append(taskList, task)
	}
	return taskList, nil
}

// LaunchSynchronisationAllVappTemplates starts synchronisation of all vApp templates for a given catalog
func (cat *AdminCatalog) LaunchSynchronisationAllVappTemplates() ([]*Task, error) {
	vappTemplatesList, err := cat.QueryVappTemplateList()
	if err != nil {
		return nil, err
	}
	var nameList []string
	for _, element := range vappTemplatesList {
		nameList = append(nameList, element.Name)
	}
	return cat.LaunchSynchronisationVappTemplates(nameList)
}

// LaunchSynchronisationMediaItems starts synchronisation of a list of media items
func (cat *AdminCatalog) LaunchSynchronisationMediaItems(nameList []string) ([]*Task, error) {
	var taskList []*Task
	mediaList, err := cat.QueryMediaList()
	if err != nil {
		return nil, err
	}
	var actionList []string

	var found = make(map[string]string)
	for _, element := range mediaList {
		if contains(element.Name, nameList) {
			util.Logger.Printf("scheduling for synchronisation Media item %s with catalog item HREF %s\n", element.Name, element.CatalogItem)
			actionList = append(actionList, element.CatalogItem)
			found[element.Name] = element.CatalogItem
		}
	}
	if len(actionList) < len(nameList) {
		var foundList []string
		for k := range found {
			foundList = append(foundList, k)
		}
		return nil, fmt.Errorf("%d names provided [%v] but %d actions scheduled [%v]", len(nameList), nameList, len(actionList), foundList)
	}
	for _, element := range actionList {
		util.Logger.Printf("synchronising Media catalog item HREF %s\n", element)
		catalogItem, err := cat.GetCatalogItemByHref(element)
		if err != nil {
			return nil, err
		}
		task, err := catalogItem.LaunchSync()
		if err != nil {
			return nil, err
		}
		taskList = append(taskList, task)
	}
	return taskList, nil
}

// LaunchSynchronisationAllMediaItems starts synchronisation of all media items for a given catalog
func (cat *AdminCatalog) LaunchSynchronisationAllMediaItems() ([]*Task, error) {
	var taskList []*Task
	mediaList, err := cat.QueryMediaList()
	if err != nil {
		return nil, err
	}
	for _, element := range mediaList {
		catalogItem, err := cat.GetCatalogItemByHref(element.CatalogItem)
		if err != nil {
			return nil, err
		}
		task, err := catalogItem.LaunchSync()
		if err != nil {
			return nil, err
		}
		taskList = append(taskList, task)
	}
	return taskList, nil
}

// GetCatalogItemByHref finds a CatalogItem by HREF
// On success, returns a pointer to the CatalogItem structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *AdminCatalog) GetCatalogItemByHref(catalogItemHref string) (*CatalogItem, error) {
	catItem := NewCatalogItem(cat.client)

	_, err := cat.client.ExecuteRequest(catalogItemHref, http.MethodGet,
		"", "error retrieving catalog item: %s", nil, catItem.CatalogItem)
	if err != nil {
		return nil, err
	}
	return catItem, nil
}

/*
// GetCatalogByHref retrieves a catalog without the parent Org
func (client *Client) GetCatalogByHref(catalogHref string) (*Catalog, error) {
	if strings.Contains(catalogHref, "/admin/") {
		return nil, fmt.Errorf("catalog HREF '%s' is for an Admin catalog", catalogHref)
	}
	cat := NewCatalog(client)

	_, err := client.ExecuteRequest(catalogHref, http.MethodGet,
		"", "error retrieving catalog: %s", nil, cat.Catalog)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

// GetCatalogById retrieves a catalog by ID without the parent Org
func (client *Client) GetCatalogById(catalogId string) (*Catalog, error) {
	catalogHREF := client.VCDHREF
	catalogHREF.Path += "/catalog/" + extractUuid(catalogId)

	util.Logger.Printf("[TRACE] Url for retrieving catalog : %s", catalogHREF.String())

	return client.GetCatalogByHref(catalogHREF.String())
}

// GetAdminCatalogByHref retrieves an admin catalog without the parent Org
func (client *Client) GetAdminCatalogByHref(catalogHref string) (*AdminCatalog, error) {
	if !strings.Contains(catalogHref, "/admin/") {
		return nil, fmt.Errorf("catalog HREF '%s' is NOT for an Admin Catalog", catalogHref)
	}
	cat := NewAdminCatalog(client)

	_, err := client.ExecuteRequest(catalogHref, http.MethodGet,
		"", "error retrieving catalog: %s", nil, cat.AdminCatalog)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

// GetAdminCatalogById retrieves a catalog by ID without the parent Org
func (client *Client) GetAdminCatalogById(catalogId string) (*AdminCatalog, error) {
	catalogHREF := client.VCDHREF
	catalogHREF.Path += "/admin/catalog/" + extractUuid(catalogId)

	util.Logger.Printf("[TRACE] Url for retrieving Admin Catalog : %s", catalogHREF.String())
	return client.GetAdminCatalogByHref(catalogHREF.String())
}
*/

// UpdateSubscriptionParams modifies the subscription parameters of an already subscribed catalog
func (catalog *AdminCatalog) UpdateSubscriptionParams(params types.ExternalCatalogSubscription) error {
	var href string
	for _, link := range catalog.AdminCatalog.Link {
		if link.Rel == "externalCatalogSubscriptionParams" && link.Type == types.MimeSubscribeToExternalCatalog {
			href = link.HREF
			break
		}
	}
	if href == "" {
		return fmt.Errorf("catalog subscription link not found for catalog %s", catalog.AdminCatalog.Name)
	}
	_, err := catalog.client.ExecuteRequest(href, http.MethodPost, types.MimeAdminCatalog,
		"error subscribing to catalog: %s", params, nil)
	if err != nil {
		return err
	}
	return catalog.Refresh()
}

/*
// QueryTaskList retrieves a list of tasks associated to the Admin Catalog
func (catalog *AdminCatalog) QueryTaskList(filter map[string]string) ([]*types.QueryResultTaskRecordType, error) {
	catalogHref, err := catalog.GetCatalogHref()
	if err != nil {
		return nil, err
	}
	var newFilter = map[string]string{
		"object": catalogHref,
	}
	for k, v := range filter {
		newFilter[k] = v
	}
	return catalog.client.QueryTaskList(newFilter)
}
*/
