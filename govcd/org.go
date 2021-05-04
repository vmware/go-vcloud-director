/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type Org struct {
	Org    *types.Org
	client *Client
}

func NewOrg(client *Client) *Org {
	return &Org{
		Org:    new(types.Org),
		client: client,
	}
}

// Given an org with a valid HREF, the function refetches the org
// and updates the user's org data. Otherwise if the function fails,
// it returns an error. Users should use refresh whenever they have
// a stale org due to the creation/update/deletion of a resource
// within the org or the org itself.
func (org *Org) Refresh() error {
	if *org == (Org{}) {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledOrg := &types.Org{}

	_, err := org.client.ExecuteRequest(org.Org.HREF, http.MethodGet,
		"", "error refreshing organization: %s", nil, unmarshalledOrg)
	if err != nil {
		return err
	}
	org.Org = unmarshalledOrg

	// The request was successful
	return nil
}

// Given a valid catalog name, FindCatalog returns a Catalog object.
// If no catalog is found, then returns an empty catalog and no error.
// Otherwise it returns an error.
// Deprecated: use org.GetCatalogByName instead
func (org *Org) FindCatalog(catalogName string) (Catalog, error) {

	for _, link := range org.Org.Link {
		if link.Rel == "down" && link.Type == "application/vnd.vmware.vcloud.catalog+xml" && link.Name == catalogName {

			cat := NewCatalog(org.client)

			_, err := org.client.ExecuteRequest(link.HREF, http.MethodGet,
				"", "error retrieving catalog: %s", nil, cat.Catalog)

			return *cat, err
		}
	}

	return Catalog{}, nil
}

// GetVdcByName if user specifies valid vdc name then this returns a vdc object.
// If no vdc is found, then it returns an empty vdc and no error.
// Otherwise it returns an empty vdc and an error.
// Deprecated: use org.GetVDCByName instead
func (org *Org) GetVdcByName(vdcname string) (Vdc, error) {
	for _, link := range org.Org.Link {
		if link.Name == vdcname {
			vdc := NewVdc(org.client)

			_, err := org.client.ExecuteRequest(link.HREF, http.MethodGet,
				"", "error retrieving vdc: %s", nil, vdc.Vdc)

			return *vdc, err
		}
	}
	return Vdc{}, nil
}

// CreateCatalog creates a catalog with specified name and description
func CreateCatalog(client *Client, links types.LinkList, Name, Description string) (AdminCatalog, error) {
	adminCatalog, err := CreateCatalogWithStorageProfile(client, links, Name, Description, nil)
	if err != nil {
		return AdminCatalog{}, nil
	}
	return *adminCatalog, nil
}

// CreateCatalogWithStorageProfile is like CreateCatalog, but allows to specify storage profile
func CreateCatalogWithStorageProfile(client *Client, links types.LinkList, Name, Description string, storageProfiles *types.CatalogStorageProfiles) (*AdminCatalog, error) {
	reqCatalog := &types.Catalog{
		Name:        Name,
		Description: Description,
	}
	vcomp := &types.AdminCatalog{
		Xmlns:                  types.XMLNamespaceVCloud,
		Catalog:                *reqCatalog,
		CatalogStorageProfiles: storageProfiles,
	}

	var createOrgLink *types.Link
	for _, link := range links {
		if link.Rel == "add" && link.Type == types.MimeAdminCatalog {
			util.Logger.Printf("[TRACE] Create org - found the proper link for request, HREF: %s, "+
				"name: %s, type: %s, id: %s, rel: %s \n", link.HREF, link.Name, link.Type, link.ID, link.Rel)
			createOrgLink = link
		}
	}

	if createOrgLink == nil {
		return nil, fmt.Errorf("creating catalog failed to find url")
	}

	catalog := NewAdminCatalog(client)
	_, err := client.ExecuteRequest(createOrgLink.HREF, http.MethodPost,
		"application/vnd.vmware.admin.catalog+xml", "error creating catalog: %s", vcomp, catalog.AdminCatalog)

	return catalog, err
}

// CreateCatalog creates a catalog with given name and description under
// the given organization. Returns an Catalog that contains a creation
// task.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-CreateCatalog.html
func (org *Org) CreateCatalog(name, description string) (Catalog, error) {
	catalog, err := org.CreateCatalogWithStorageProfile(name, description, nil)
	if err != nil {
		return Catalog{}, err
	}
	return *catalog, nil
}

// CreateCatalogWithStorageProfile is like CreateCatalog but additionally allows to specify storage profiles
func (org *Org) CreateCatalogWithStorageProfile(name, description string, storageProfiles *types.CatalogStorageProfiles) (*Catalog, error) {
	catalog := NewCatalog(org.client)
	adminCatalog, err := CreateCatalogWithStorageProfile(org.client, org.Org.Link, name, description, storageProfiles)
	if err != nil {
		return nil, err
	}
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog, nil
}

func validateVdcConfiguration(vdcDefinition *types.VdcConfiguration) error {
	if vdcDefinition.Name == "" {
		return errors.New("VdcConfiguration missing required field: Name")
	}
	if vdcDefinition.AllocationModel == "" {
		return errors.New("VdcConfiguration missing required field: AllocationModel")
	}
	if vdcDefinition.ComputeCapacity == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity")
	}
	if len(vdcDefinition.ComputeCapacity) != 1 {
		return errors.New("VdcConfiguration invalid field: ComputeCapacity must only have one element")
	}
	if vdcDefinition.ComputeCapacity[0] == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0]")
	}
	if vdcDefinition.ComputeCapacity[0].CPU == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].CPU")
	}
	if vdcDefinition.ComputeCapacity[0].CPU.Units == "" {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].CPU.Units")
	}
	if vdcDefinition.ComputeCapacity[0].Memory == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].Memory")
	}
	if vdcDefinition.ComputeCapacity[0].Memory.Units == "" {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")
	}
	if vdcDefinition.VdcStorageProfile == nil || len(vdcDefinition.VdcStorageProfile) == 0 {
		return errors.New("VdcConfiguration missing required field: VdcStorageProfile")
	}
	if vdcDefinition.VdcStorageProfile[0].Units == "" {
		return errors.New("VdcConfiguration missing required field: VdcStorageProfile.Units")
	}
	if vdcDefinition.ProviderVdcReference == nil {
		return errors.New("VdcConfiguration missing required field: ProviderVdcReference")
	}
	if vdcDefinition.ProviderVdcReference.HREF == "" {
		return errors.New("VdcConfiguration missing required field: ProviderVdcReference.HREF")
	}
	return nil
}

// GetCatalogByHref  finds a Catalog by HREF
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (org *Org) GetCatalogByHref(catalogHref string) (*Catalog, error) {
	cat := NewCatalog(org.client)

	_, err := org.client.ExecuteRequest(catalogHref, http.MethodGet,
		"", "error retrieving catalog: %s", nil, cat.Catalog)
	if err != nil {
		return nil, err
	}
	// The request was successful
	return cat, nil
}

// GetCatalogByName  finds a Catalog by Name
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
//
// refresh has no effect here, but is kept to preserve signature
func (org *Org) GetCatalogByName(catalogName string, refresh bool) (*Catalog, error) {

	orgCatalogs, err := org.QueryCatalogList()
	if err != nil {
		return nil, fmt.Errorf("error retrieving Catalog list for Org '%s': %s", org.Org.Name, err)
	}

	for _, catalog := range orgCatalogs {
		// Get Catalog HREF
		if catalog.Name == catalogName {
			return org.GetCatalogByHref(catalog.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetCatalogById finds a Catalog by ID
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (org *Org) GetCatalogById(catalogId string, refresh bool) (*Catalog, error) {
	catalogs, err := org.QueryCatalogList()
	if err != nil {
		return nil, fmt.Errorf("error retrieving Catalog List for Org '%s': %s", org.Org.Name, err)
	}

	for _, catalog := range catalogs {
		if equalIds(catalogId, catalog.ID, catalog.HREF) {
			return org.GetCatalogByHref(catalog.HREF)
		}
	}

	return nil, ErrorEntityNotFound
}

// GetCatalogByNameOrId finds a Catalog by name or ID
// On success, returns a pointer to the Catalog structure and a nil error
// On failure, returns a nil pointer and an error
func (org *Org) GetCatalogByNameOrId(identifier string, refresh bool) (*Catalog, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return org.GetCatalogByName(name, refresh) }
	getById := func(id string, refresh bool) (interface{}, error) { return org.GetCatalogById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*Catalog), err
}

// GetVDCByHref finds a VDC by HREF
// On success, returns a pointer to the VDC structure and a nil error
// On failure, returns a nil pointer and an error
func (org *Org) GetVDCByHref(vdcHref string) (*Vdc, error) {
	vdc := NewVdc(org.client)
	_, err := org.client.ExecuteRequest(vdcHref, http.MethodGet,
		"", "error retrieving VDC: %s", nil, vdc.Vdc)
	if err != nil {
		return nil, err
	}
	// The request was successful
	return vdc, nil
}

// GetVDCByName finds a VDC by Name
// On success, returns a pointer to the VDC structure and a nil error
// On failure, returns a nil pointer and an error
//
// refresh has no effect and is kept to preserve signature
func (org *Org) GetVDCByName(vdcName string, refresh bool) (*Vdc, error) {
	vdcQuery, err := org.queryOrgVdcByName(vdcName)
	if ContainsNotFound(err) {
		return nil, ErrorEntityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying VDC: %s", err)
	}
	// This is not an AdminOrg and admin HREF must be removed if it exists
	href := strings.Replace(vdcQuery.HREF, "/api/admin", "/api", 1)
	return org.GetVDCByHref(href)
}

// GetVDCById finds a VDC by ID
// On success, returns a pointer to the VDC structure and a nil error
// On failure, returns a nil pointer and an error
//
// refresh has no effect and is kept to preserve signature
func (org *Org) GetVDCById(vdcId string, refresh bool) (*Vdc, error) {
	vdcQuery, err := org.queryOrgVdcById(vdcId)
	if ContainsNotFound(err) {
		return nil, ErrorEntityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying VDC: %s", err)
	}

	// This is not an AdminOrg and admin HREF must be removed if it exists
	href := strings.Replace(vdcQuery.HREF, "/api/admin", "/api", 1)
	return org.GetVDCByHref(href)
}

// GetVDCByNameOrId finds a VDC by name or ID
// On success, returns a pointer to the VDC structure and a nil error
// On failure, returns a nil pointer and an error
func (org *Org) GetVDCByNameOrId(identifier string, refresh bool) (*Vdc, error) {
	// This function cannot directly use shorthand getEntityByNameOrId because org.GetVDCById API throws an error when the
	// supplied identified is not UUID or URN. Therefore this function does not even try to to find by ID when supplied
	// identifier does not look like an ID
	if isUrn(identifier) || IsUuid(identifier) {
		vdcById, byIdErr := org.GetVDCById(identifier, refresh)
		if byIdErr == nil {
			// Found by ID = return it
			return vdcById, nil
		}

		// Return real error if it is not 'ErrorEntityNotFound'
		if byIdErr != nil && !ContainsNotFound(byIdErr) {
			return nil, byIdErr
		}
	}

	// Not found by ID, try by name
	vdcByName, byNameErr := org.GetVDCByName(identifier, false)
	return vdcByName, byNameErr
}

// QueryCatalogList returns a list of catalogs for this organization
func (org *Org) QueryCatalogList() ([]*types.CatalogRecord, error) {
	util.Logger.Printf("[DEBUG] QueryCatalogList with Org name %s", org.Org.Name)
	queryType := org.client.GetQueryType(types.QtCatalog)
	results, err := org.client.cumulativeQuery(queryType, nil, map[string]string{
		"type":          queryType,
		"filter":        fmt.Sprintf("orgName==%s", url.QueryEscape(org.Org.Name)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}

	var catalogs []*types.CatalogRecord

	if org.client.IsSysAdmin {
		catalogs = results.Results.AdminCatalogRecord
	} else {
		catalogs = results.Results.CatalogRecord
	}
	util.Logger.Printf("[DEBUG] QueryCatalogList returned with : %#v and error: %s", catalogs, err)
	return catalogs, nil
}

// queryOrgVdcByName returns a single QueryResultOrgVdcRecordType
func (org *Org) queryOrgVdcByName(vdcName string) (*types.QueryResultOrgVdcRecordType, error) {
	allVdcs, err := queryOrgVdcList(org.client, "name", vdcName, org.Org.ID)
	if err != nil {
		return nil, err
	}

	if allVdcs == nil || len(allVdcs) < 1 {
		return nil, ErrorEntityNotFound
	}

	if len(allVdcs) > 1 {
		return nil, fmt.Errorf("found more than 1 VDC with Name '%s'", vdcName)
	}

	return allVdcs[0], nil
}

// queryOrgVdcById returns a single Org VDC query result
func (org *Org) queryOrgVdcById(vdcId string) (*types.QueryResultOrgVdcRecordType, error) {
	allVdcs, err := queryOrgVdcList(org.client, "id", vdcId, org.Org.ID)

	if err != nil {
		return nil, err
	}

	if len(allVdcs) < 1 {
		return nil, ErrorEntityNotFound
	}

	return allVdcs[0], nil
}

// queryOrgVdcList returns all Org VDCs using query endpoint
func (org *Org) queryOrgVdcList() ([]*types.QueryResultOrgVdcRecordType, error) {
	return queryOrgVdcList(org.client, "orgName", org.Org.Name, org.Org.ID)
}

func queryOrgVdcList(client *Client, filterFieldName, filterFieldValue, orgId string) ([]*types.QueryResultOrgVdcRecordType, error) {
	queryType := client.GetQueryType(types.QtOrgVdc)
	filterText := fmt.Sprintf("%s==%s", filterFieldName, url.QueryEscape(filterFieldValue))

	results, err := client.cumulativeQuery(queryType, nil, map[string]string{
		"type":          queryType,
		"filter":        filterText,
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, fmt.Errorf("error querying Org VDCs %s", err)
	}

	if client.IsSysAdmin {
		return results.Results.OrgVdcAdminRecord, nil
	} else {
		return results.Results.OrgVdcRecord, nil
	}
}

// GetTaskList returns Tasks for Organization and error.
func (org *Org) GetTaskList() (*types.TasksList, error) {

	for _, link := range org.Org.Link {
		if link.Rel == "down" && link.Type == "application/vnd.vmware.vcloud.tasksList+xml" {

			tasksList := &types.TasksList{}

			_, err := org.client.ExecuteRequest(link.HREF, http.MethodGet, "",
				"error getting taskList: %s", nil, tasksList)
			if err != nil {
				return nil, err
			}

			return tasksList, nil
		}
	}

	return nil, fmt.Errorf("link not found")
}
