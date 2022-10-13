/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type CatalogItem struct {
	CatalogItem *types.CatalogItem
	client      *Client
}

func NewCatalogItem(cli *Client) *CatalogItem {
	return &CatalogItem{
		CatalogItem: new(types.CatalogItem),
		client:      cli,
	}
}

func (catalogItem *CatalogItem) GetVAppTemplate() (VAppTemplate, error) {

	cat := NewVAppTemplate(catalogItem.client)

	_, err := catalogItem.client.ExecuteRequest(catalogItem.CatalogItem.Entity.HREF, http.MethodGet,
		"", "error retrieving vApp template: %s", nil, cat.VAppTemplate)

	// The request was successful
	return *cat, err

}

// Delete deletes the Catalog Item, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-CatalogItem.html
func (catalogItem *CatalogItem) Delete() error {
	util.Logger.Printf("[TRACE] Deleting catalog item: %#v", catalogItem.CatalogItem)
	catalogItemHREF := catalogItem.client.VCDHREF
	catalogItemHREF.Path += "/catalogItem/" + catalogItem.CatalogItem.ID[23:]

	util.Logger.Printf("[TRACE] Url for deleting catalog item: %#v and name: %s", catalogItemHREF, catalogItem.CatalogItem.Name)

	return catalogItem.client.ExecuteRequestWithoutResponse(catalogItemHREF.String(), http.MethodDelete,
		"", "error deleting Catalog item: %s", nil)
}

// queryCatalogItemList returns a list of Catalog Item for the given parent
func queryCatalogItemList(client *Client, parentField, parentValue string) ([]*types.QueryResultCatalogItemType, error) {
	return queryCatalogItemFilteredList(client, map[string]string{parentField: parentValue})
}

// queryCatalogItemFilteredList returns a list of Catalog Items with an optional filter
func queryCatalogItemFilteredList(client *Client, filter map[string]string) ([]*types.QueryResultCatalogItemType, error) {
	catalogItemType := types.QtCatalogItem
	if client.IsSysAdmin {
		catalogItemType = types.QtAdminCatalogItem
	}

	filterText := ""
	for k, v := range filter {
		if filterText != "" {
			filterText += ";"
		}
		filterText += fmt.Sprintf("%s==%s", k, url.QueryEscape(v))
	}

	notEncodedParams := map[string]string{
		"type": catalogItemType,
	}
	if filterText != "" {
		notEncodedParams["filter"] = filterText
	}
	results, err := client.cumulativeQuery(catalogItemType, nil, notEncodedParams)
	if err != nil {
		return nil, fmt.Errorf("error querying catalog items %s", err)
	}

	if client.IsSysAdmin {
		return results.Results.AdminCatalogItemRecord, nil
	} else {
		return results.Results.CatalogItemRecord, nil
	}
}

// QueryCatalogItemList returns a list of Catalog Item for the given catalog
func (catalog *Catalog) QueryCatalogItemList() ([]*types.QueryResultCatalogItemType, error) {
	return queryCatalogItemList(catalog.client, "catalog", catalog.Catalog.ID)
}

// QueryCatalogItemList returns a list of Catalog Item for the given admin catalog
func (catalog *AdminCatalog) QueryCatalogItemList() ([]*types.QueryResultCatalogItemType, error) {
	return queryCatalogItemList(catalog.client, "catalog", catalog.AdminCatalog.ID)
}

// QueryCatalogItem returns a named Catalog Item for the given catalog
func (catalog *Catalog) QueryCatalogItem(name string) (*types.QueryResultCatalogItemType, error) {
	return queryCatalogItem(catalog.client, "catalog", catalog.Catalog.ID, name)
}

// QueryCatalogItem returns a named Catalog Item for the given catalog
func (catalog *AdminCatalog) QueryCatalogItem(name string) (*types.QueryResultCatalogItemType, error) {
	return queryCatalogItem(catalog.client, "catalog", catalog.AdminCatalog.ID, name)
}

// queryCatalogItem returns a named Catalog Item for the given parent
func queryCatalogItem(client *Client, parentField, parentValue, name string) (*types.QueryResultCatalogItemType, error) {

	result, err := queryCatalogItemFilteredList(client, map[string]string{parentField: parentValue, "name": name})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(result) > 1 {
		return nil, fmt.Errorf("more than one item (%d) found with name %s", len(result), name)
	}
	return result[0], nil
}

func queryResultCatalogItemToCatalogItem(client *Client, qr *types.QueryResultCatalogItemType) *CatalogItem {
	var catalogItem = NewCatalogItem(client)
	catalogItem.CatalogItem = &types.CatalogItem{
		HREF:        qr.HREF,
		Type:        qr.Type,
		ID:          extractUuid(qr.HREF),
		Name:        qr.Name,
		DateCreated: qr.CreationDate,
		Entity: &types.Entity{
			HREF: qr.Entity,
			Type: qr.EntityType,
			Name: qr.EntityName,
		},
	}
	return catalogItem
}

// QueryCatalogItemList returns a list of Catalog Item for the given VDC
func (vdc *Vdc) QueryCatalogItemList() ([]*types.QueryResultCatalogItemType, error) {
	return queryCatalogItemList(vdc.client, "vdc", vdc.Vdc.ID)
}

// QueryCatalogItemList returns a list of Catalog Item for the given Admin VDC
func (vdc *AdminVdc) QueryCatalogItemList() ([]*types.QueryResultCatalogItemType, error) {
	return queryCatalogItemList(vdc.client, "vdc", vdc.AdminVdc.ID)
}

// queryVappTemplateList returns a list of vApp templates for the given parent
func queryVappTemplateList(client *Client, parentField, parentValue string) ([]*types.QueryResultVappTemplateType, error) {

	vappTemplateType := types.QtVappTemplate
	if client.IsSysAdmin {
		vappTemplateType = types.QtAdminVappTemplate
	}
	results, err := client.cumulativeQuery(vappTemplateType, nil, map[string]string{
		"type":   vappTemplateType,
		"filter": fmt.Sprintf("%s==%s", parentField, url.QueryEscape(parentValue)),
	})
	if err != nil {
		return nil, fmt.Errorf("error querying vApp templates %s", err)
	}

	if client.IsSysAdmin {
		return results.Results.AdminVappTemplateRecord, nil
	} else {
		return results.Results.VappTemplateRecord, nil
	}
}

// QueryVappTemplateList returns a list of vApp templates for the given VDC
func (vdc *Vdc) QueryVappTemplateList() ([]*types.QueryResultVappTemplateType, error) {
	return queryVappTemplateList(vdc.client, "vdcName", vdc.Vdc.Name)
}

// QueryVappTemplateList returns a list of vApp templates for the given VDC
func (vdc *AdminVdc) QueryVappTemplateList() ([]*types.QueryResultVappTemplateType, error) {
	return queryVappTemplateList(vdc.client, "vdcName", vdc.AdminVdc.Name)
}

// QueryVappTemplateList returns a list of vApp templates for the given catalog
func (catalog *Catalog) QueryVappTemplateList() ([]*types.QueryResultVappTemplateType, error) {
	return queryVappTemplateList(catalog.client, "catalogName", catalog.Catalog.Name)
}

// Sync synchronises a subscribed Catalog item
func (item *CatalogItem) Sync() error {
	return elementSync(item.client, item.CatalogItem.HREF, "catalog item")
}

// LaunchSync starts synchronisation of a subscribed Catalog item
func (item *CatalogItem) LaunchSync() (*Task, error) {
	return elementLaunchSync(item.client, item.CatalogItem.HREF, "catalog item")
}
