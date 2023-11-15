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

// Refresh fetches a fresh copy of the Admin Catalog
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
	adminCatalog := NewAdminCatalog(org.client)
	reqCatalog := &types.Catalog{
		Name: catalogName,
	}
	adminCatalog.AdminCatalog = &types.AdminCatalog{
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

	adminCatalog.AdminCatalog.ExternalCatalogSubscription.Password = password
	adminCatalog.AdminCatalog.ExternalCatalogSubscription.LocalCopy = localCopy
	_, err := org.client.ExecuteRequest(href, http.MethodPost, types.MimeAdminCatalog,
		"error subscribing to catalog: %s", adminCatalog.AdminCatalog, adminCatalog.AdminCatalog)
	if err != nil {
		return nil, err
	}
	// Before returning, check that there are no failing tasks
	err = adminCatalog.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing subscribed catalog %s: %s", catalogName, err)
	}
	if adminCatalog.AdminCatalog.Tasks != nil {
		msg := ""
		for _, task := range adminCatalog.AdminCatalog.Tasks.Task {
			if task.Status == "error" {
				if task.Error != nil {
					msg = task.Error.Error()
				}
				return nil, fmt.Errorf("error while subscribing catalog %s (task %s): %s", catalogName, task.Name, msg)
			}
			if task.Tasks != nil {
				for _, subTask := range task.Tasks.Task {
					if subTask.Status == "error" {
						if subTask.Error != nil {
							msg = subTask.Error.Error()
						}
						return nil, fmt.Errorf("error while subscribing catalog %s (subTask %s): %s", catalogName, subTask.Name, msg)
					}

				}
			}
		}
	}
	return adminCatalog, nil
}

// FullSubscriptionUrl returns the subscription URL from a publishing catalog
// adding the HOST if needed
func (cat *AdminCatalog) FullSubscriptionUrl() (string, error) {
	err := cat.Refresh()
	if err != nil {
		return "", err
	}
	if cat.AdminCatalog.PublishExternalCatalogParams == nil {
		return "", fmt.Errorf("AdminCatalog %s has no publishing parameters", cat.AdminCatalog.Name)
	}
	subscriptionUrl, err := buildFullUrl(cat.AdminCatalog.PublishExternalCatalogParams.CatalogPublishedUrl, cat.AdminCatalog.HREF)
	if err != nil {
		return "", err
	}
	return subscriptionUrl, nil
}

// buildFullUrl gets a (possibly incomplete) URL and returns it completed, using the provided HREF as basis
func buildFullUrl(subscriptionUrl, href string) (string, error) {
	var err error
	if !IsValidUrl(subscriptionUrl) {
		// Get the entity base URL
		cutPosition := strings.Index(href, "/api")
		host := href[:cutPosition]
		subscriptionUrl, err = url.JoinPath(host, subscriptionUrl)
		if err != nil {
			return "", err
		}
	}
	return subscriptionUrl, nil
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

// WaitForTasks waits for the catalog's tasks to complete
func (cat *AdminCatalog) WaitForTasks() error {
	if ResourceInProgress(cat.AdminCatalog.Tasks) {
		err := WaitResource(func() (*types.TasksInProgress, error) {
			err := cat.Refresh()
			if err != nil {
				return nil, err
			}
			return cat.AdminCatalog.Tasks, nil
		})
		return err
	}
	return nil
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
	err = cat.WaitForTasks()
	if err != nil {
		return err
	}
	return elementSync(cat.client, catalogHref, "admin catalog")
}

// LaunchSync starts synchronisation of a subscribed AdminCatalog
func (cat *AdminCatalog) LaunchSync() (*Task, error) {
	err := checkIfSubscribedCatalog(cat)
	if err != nil {
		return nil, err
	}
	// The sync operation is only available for Catalog, not AdminCatalog.
	// We use the embedded Catalog object for this purpose
	catalogHref, err := cat.GetCatalogHref()
	if err != nil || catalogHref == "" {
		return nil, fmt.Errorf("empty catalog HREF for admin catalog %s", cat.AdminCatalog.Name)
	}
	err = cat.WaitForTasks()
	if err != nil {
		return nil, err
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
	return launchSynchronisationVappTemplates(cat, nameList, true)
}

// launchSynchronisationVappTemplates waits for existing tasks to complete and then starts synchronisation for a list of vApp templates
// optionally checking for running tasks
// TODO: re-implement without the undocumented task-related fields
func launchSynchronisationVappTemplates(cat *AdminCatalog, nameList []string, checkForRunningTasks bool) ([]*Task, error) {
	err := checkIfSubscribedCatalog(cat)
	if err != nil {
		return nil, err
	}
	util.Logger.Printf("[TRACE] launchSynchronisationVappTemplates - AdminCatalog '%s' - 'make_local_copy=%v]\n", cat.AdminCatalog.Name, cat.AdminCatalog.ExternalCatalogSubscription.LocalCopy)
	var taskList []*Task

	for _, element := range nameList {
		var queryResultCatalogItem *types.QueryResultCatalogItemType

		if checkForRunningTasks {
			queryResultVappTemplate, err := cat.QueryVappTemplateWithName(element)
			if err != nil {
				return nil, err
			}
			err = checkIfTaskComplete(cat.client, queryResultVappTemplate.Task, queryResultVappTemplate.TaskStatus)
			if err != nil {
				return nil, err
			}
			queryResultCatalogItem = &types.QueryResultCatalogItemType{
				HREF:        queryResultVappTemplate.CatalogItem,
				ID:          extractUuid(queryResultVappTemplate.CatalogItem),
				Type:        types.MimeCatalogItem,
				Entity:      queryResultVappTemplate.HREF,
				EntityName:  queryResultVappTemplate.Name,
				EntityType:  "vapptemplate",
				Catalog:     cat.AdminCatalog.HREF,
				CatalogName: cat.AdminCatalog.Name,
				Status:      queryResultVappTemplate.Status,
				Name:        queryResultVappTemplate.Name,
			}
		} else {
			queryResultCatalogItem, err = cat.QueryCatalogItem(element)
			if err != nil {
				return nil, fmt.Errorf("error retrieving catalog item %s: %s", element, err)
			}
		}
		task, err := queryResultCatalogItemToCatalogItem(cat.client, queryResultCatalogItem).LaunchSync()
		if err != nil {
			return nil, err
		}
		if task != nil {
			taskList = append(taskList, task)
		}
	}
	return taskList, nil
}

// LaunchSynchronisationAllVappTemplates waits for existing tasks to complete and then starts synchronisation of all vApp templates for a given catalog
// TODO: re-implement without the undocumented task-related fields
func (cat *AdminCatalog) LaunchSynchronisationAllVappTemplates() ([]*Task, error) {
	err := checkIfSubscribedCatalog(cat)
	if err != nil {
		return nil, err
	}
	util.Logger.Printf("[TRACE] AdminCatalog '%s' LaunchSynchronisationAllVappTemplates - 'make_local_copy=%v]\n", cat.AdminCatalog.Name, cat.AdminCatalog.ExternalCatalogSubscription.LocalCopy)
	vappTemplatesList, err := cat.QueryVappTemplateList()
	if err != nil {
		return nil, err
	}
	var nameList []string
	for _, element := range vappTemplatesList {
		err = checkIfTaskComplete(cat.client, element.Task, element.TaskStatus)
		if err != nil {
			return nil, err
		}
		nameList = append(nameList, element.Name)
	}
	// Launch synchronisation for each item, without checking for running tasks, as it was already done in this function
	return launchSynchronisationVappTemplates(cat, nameList, false)
}

func checkIfTaskComplete(client *Client, taskHref, taskStatus string) error {
	complete := taskStatus == "" || isTaskCompleteOrError(taskStatus)
	if !complete {
		task, err := client.GetTaskById(taskHref)
		if err != nil {
			return err
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return err
		}
	}
	return nil
}

func checkIfSubscribedCatalog(catalog *AdminCatalog) error {
	err := catalog.Refresh()
	if err != nil {
		return err
	}
	if catalog.AdminCatalog.ExternalCatalogSubscription == nil || catalog.AdminCatalog.ExternalCatalogSubscription.Location == "" {
		return fmt.Errorf("catalog '%s' is not subscribed", catalog.AdminCatalog.Name)
	}
	return nil
}

// LaunchSynchronisationMediaItems waits for existing tasks to complete and then starts synchronisation of a list of media items
// TODO: re-implement without the undocumented task-related fields
func (cat *AdminCatalog) LaunchSynchronisationMediaItems(nameList []string) ([]*Task, error) {
	err := checkIfSubscribedCatalog(cat)
	if err != nil {
		return nil, err
	}
	util.Logger.Printf("[TRACE] AdminCatalog '%s' LaunchSynchronisationMediaItems\n", cat.AdminCatalog.Name)
	var taskList []*Task
	mediaList, err := cat.QueryMediaList()
	if err != nil {
		return nil, err
	}
	var actionList []string

	var found = make(map[string]string)
	for _, element := range mediaList {
		if contains(element.Name, nameList) {
			complete := element.TaskStatus == "" || isTaskCompleteOrError(element.TaskStatus)
			if !complete {
				if element.Task != "" {
					task, err := cat.client.GetTaskById(element.Task)
					if err != nil {
						return nil, err
					}
					err = task.WaitTaskCompletion()
					if err != nil {
						return nil, err
					}
				}
			}
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
		if task != nil {
			taskList = append(taskList, task)
		}
	}
	return taskList, nil
}

// LaunchSynchronisationAllMediaItems waits for existing tasks to complete and then starts synchronisation of all media items for a given catalog
// TODO re-implement without the non-documented task-related fields
func (cat *AdminCatalog) LaunchSynchronisationAllMediaItems() ([]*Task, error) {
	err := checkIfSubscribedCatalog(cat)
	if err != nil {
		return nil, err
	}
	util.Logger.Printf("[TRACE] AdminCatalog '%s' LaunchSynchronisationAllMediaItems\n", cat.AdminCatalog.Name)
	var taskList []*Task
	mediaList, err := cat.QueryMediaList()
	if err != nil {
		return nil, err
	}
	for _, element := range mediaList {
		if isTaskRunning(element.TaskStatus) {
			task, err := cat.client.GetTaskByHREF(element.Task)
			if err != nil {
				return nil, err
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return nil, err
			}
		}
		catalogItem, err := cat.GetCatalogItemByHref(element.CatalogItem)
		if err != nil {
			return nil, err
		}
		task, err := catalogItem.LaunchSync()
		if err != nil {
			return nil, err
		}
		if task != nil {
			taskList = append(taskList, task)
		}
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

// UpdateSubscriptionParams modifies the subscription parameters of an already subscribed catalog
func (catalog *AdminCatalog) UpdateSubscriptionParams(params types.ExternalCatalogSubscription) error {
	err := checkIfSubscribedCatalog(catalog)
	if err != nil {
		return err
	}
	var href string
	for _, link := range catalog.AdminCatalog.Link {
		if link.Rel == "subscribeToExternalCatalog" && link.Type == types.MimeSubscribeToExternalCatalog {
			href = link.HREF
			break
		}
	}
	if href == "" {
		return fmt.Errorf("catalog subscription link not found for catalog %s", catalog.AdminCatalog.Name)
	}
	_, err = catalog.client.ExecuteRequest(href, http.MethodPost, types.MimeAdminCatalog,
		"error subscribing to catalog: %s", params, nil)
	if err != nil {
		return err
	}
	return catalog.Refresh()
}

// QueryTaskList retrieves a list of tasks associated to the Admin Catalog
func (catalog *AdminCatalog) QueryTaskList(filter map[string]string) ([]*types.QueryResultTaskRecordType, error) {
	catalogHref, err := catalog.GetCatalogHref()
	if err != nil {
		return nil, err
	}
	if filter == nil {
		filter = make(map[string]string)
	}
	filter["object"] = catalogHref
	return catalog.client.QueryTaskList(filter)
}

// GetAdminCatalogByHref allows retrieving a catalog from HREF, without a fully qualified AdminOrg object
func (client *Client) GetAdminCatalogByHref(catalogHref string) (*AdminCatalog, error) {
	catalogHref = strings.Replace(catalogHref, "/api/catalog", "/api/admin/catalog", 1)

	cat := NewAdminCatalog(client)

	_, err := client.ExecuteRequest(catalogHref, http.MethodGet,
		"", "error retrieving catalog: %s", nil, cat.AdminCatalog)

	if err != nil {
		return nil, err
	}

	// Setting the catalog parent, necessary to handle the tenant context
	org := NewAdminOrg(client)
	for _, link := range cat.AdminCatalog.Link {
		if link.Rel == "up" && link.Type == types.MimeAdminOrg {
			_, err = client.ExecuteRequest(link.HREF, http.MethodGet,
				"", "error retrieving parent Org: %s", nil, org.AdminOrg)
			if err != nil {
				return nil, fmt.Errorf("error retrieving catalog parent: %s", err)
			}
			break
		}
	}

	cat.parent = org
	return cat, nil
}

// QueryCatalogRecords given a catalog name, retrieves the catalogRecords that match its name
// Returns a list of catalog records for such name, empty list if none was found
func (client *Client) QueryCatalogRecords(name string, ctx TenantContext) ([]*types.CatalogRecord, error) {
	util.Logger.Printf("[DEBUG] QueryCatalogRecords")

	var filter string
	if name != "" {
		filter = fmt.Sprintf("name==%s", url.QueryEscape(name))
	}

	var tenantHeaders map[string]string

	if client.IsSysAdmin && ctx.OrgId != "" && ctx.OrgName != "" {
		// Set tenant context headers just for the query
		tenantHeaders = map[string]string{
			types.HeaderAuthContext:   ctx.OrgName,
			types.HeaderTenantContext: ctx.OrgId,
		}
	}

	queryType := types.QtCatalog

	results, err := client.cumulativeQueryWithHeaders(queryType, nil, map[string]string{
		"type":          queryType,
		"filter":        filter,
		"filterEncoded": "true",
	}, tenantHeaders)
	if err != nil {
		return nil, err
	}

	catalogs := results.Results.CatalogRecord

	util.Logger.Printf("[DEBUG] QueryCatalogRecords returned with : %#v (%d) and error: %v", catalogs, len(catalogs), err)
	return catalogs, nil
}

// GetAdminCatalogById allows retrieving a catalog from ID, without a fully qualified AdminOrg object
func (client *Client) GetAdminCatalogById(catalogId string) (*AdminCatalog, error) {
	href, err := url.JoinPath(client.VCDHREF.String(), "admin", "catalog", extractUuid(catalogId))
	if err != nil {
		return nil, err
	}
	return client.GetAdminCatalogByHref(href)
}

// GetAdminCatalogByName allows retrieving a catalog from name, without a fully qualified AdminOrg object
func (client *Client) GetAdminCatalogByName(parentOrg, catalogName string) (*AdminCatalog, error) {
	catalogs, err := queryCatalogList(client, nil)
	if err != nil {
		return nil, err
	}
	var parentOrgs []string
	for _, cat := range catalogs {
		if cat.Name == catalogName && cat.OrgName == parentOrg {
			return client.GetAdminCatalogByHref(cat.HREF)
		}
		if cat.Name == catalogName {
			parentOrgs = append(parentOrgs, cat.OrgName)
		}
	}
	parents := ""
	if len(parentOrgs) > 0 {
		parents = fmt.Sprintf(" - Found catalog %s in Orgs %v", catalogName, parentOrgs)
	}
	return nil, fmt.Errorf("no catalog '%s' found in Org %s%s", catalogName, parentOrg, parents)
}
