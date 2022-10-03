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

	catalogItemType := types.QtCatalogItem
	if client.IsSysAdmin {
		catalogItemType = types.QtAdminCatalogItem
	}

	filterText := fmt.Sprintf("%s==%s", parentField, url.QueryEscape(parentValue))

	results, err := client.cumulativeQuery(catalogItemType, nil, map[string]string{
		"type":   catalogItemType,
		"filter": filterText,
	})
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

// QueryCatalogItemList returns a list of Catalog Item for the given VDC
func (vdc *Vdc) QueryCatalogItemList() ([]*types.QueryResultCatalogItemType, error) {
	return queryCatalogItemList(vdc.client, "vdc", vdc.Vdc.ID)
}

// QueryCatalogItemList returns a list of Catalog Item for the given Admin VDC
func (vdc *AdminVdc) QueryCatalogItemList() ([]*types.QueryResultCatalogItemType, error) {
	return queryCatalogItemList(vdc.client, "vdc", vdc.AdminVdc.ID)
}

// queryVappTemplateListWithParentField returns a list of vApp templates for the given parent
func queryVappTemplateListWithParentField(client *Client, parentField, parentValue string) ([]*types.QueryResultVappTemplateType, error) {
	return queryVappTemplateListWithFilter(client, map[string]string{
		parentField: parentValue,
	})
}

// queryVappTemplateListWithFilter returns a list of vApp templates filtered by the given filter map.
// The filter map will build a filter like filterKey==filterValue;filterKey2==filterValue2;...
func queryVappTemplateListWithFilter(client *Client, filter map[string]string) ([]*types.QueryResultVappTemplateType, error) {
	vappTemplateType := types.QtVappTemplate
	if client.IsSysAdmin {
		vappTemplateType = types.QtAdminVappTemplate
	}
	filterEncoded := ""
	for k, v := range filter {
		filterEncoded += fmt.Sprintf("%s==%s;", url.QueryEscape(k), url.QueryEscape(v))
	}
	results, err := client.cumulativeQuery(vappTemplateType, nil, map[string]string{
		"type":   vappTemplateType,
		"filter": filterEncoded[:len(filterEncoded)-1], // Removes the trailing ';'
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
	return queryVappTemplateListWithParentField(vdc.client, "vdcName", vdc.Vdc.Name)
}

// QueryVappTemplateWithName returns one vApp template for the given VDC with the given name.
// Returns an error if it finds more than one.
func (vdc *Vdc) QueryVappTemplateWithName(vAppTemplateName string) (*types.QueryResultVappTemplateType, error) {
	vAppTemplates, err := queryVappTemplateListWithFilter(vdc.client, map[string]string{
		"vdcName": vdc.Vdc.Name,
		"name":    vAppTemplateName,
	})
	if err != nil {
		return nil, err
	}
	if len(vAppTemplates) != 1 {
		if len(vAppTemplates) == 0 {
			return nil, ErrorEntityNotFound
		}
		return nil, fmt.Errorf("found %d vApp Templates with name %s in VDC %s", len(vAppTemplates), vAppTemplateName, vdc.Vdc.Name)
	}
	return vAppTemplates[0], nil
}

// QueryVappTemplateList returns a list of vApp templates for the given VDC
func (vdc *AdminVdc) QueryVappTemplateList() ([]*types.QueryResultVappTemplateType, error) {
	return queryVappTemplateListWithParentField(vdc.client, "vdcName", vdc.AdminVdc.Name)
}

// QueryVappTemplateWithName returns one vApp template for the given VDC with the given name.
// Returns an error if it finds more than one.
func (vdc *AdminVdc) QueryVappTemplateWithName(vAppTemplateName string) (*types.QueryResultVappTemplateType, error) {
	vAppTemplates, err := queryVappTemplateListWithFilter(vdc.client, map[string]string{
		"vdcName": vdc.AdminVdc.Name,
		"name":    vAppTemplateName,
	})
	if err != nil {
		return nil, err
	}
	if len(vAppTemplates) != 1 {
		if len(vAppTemplates) == 0 {
			return nil, ErrorEntityNotFound
		}
		return nil, fmt.Errorf("found %d vApp Templates with name %s in VDC %s", len(vAppTemplates), vAppTemplateName, vdc.AdminVdc.Name)
	}
	return vAppTemplates[0], nil
}

// QueryVappTemplateList returns a list of vApp templates for the given catalog
func (catalog *Catalog) QueryVappTemplateList() ([]*types.QueryResultVappTemplateType, error) {
	return queryVappTemplateListWithParentField(catalog.client, "catalogName", catalog.Catalog.Name)
}

// QueryVappTemplateWithName returns one vApp template for the given Catalog with the given name.
// Returns an error if it finds more than one.
func (catalog *Catalog) QueryVappTemplateWithName(vAppTemplateName string) (*types.QueryResultVappTemplateType, error) {
	vAppTemplates, err := queryVappTemplateListWithFilter(catalog.client, map[string]string{
		"catalogName": catalog.Catalog.Name,
		"name":        vAppTemplateName,
	})
	if err != nil {
		return nil, err
	}
	if len(vAppTemplates) != 1 {
		if len(vAppTemplates) == 0 {
			return nil, ErrorEntityNotFound
		}
		return nil, fmt.Errorf("found %d vApp Templates with name %s in Catalog %s", len(vAppTemplates), vAppTemplateName, catalog.Catalog.Name)
	}
	return vAppTemplates[0], nil
}
