/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"net/url"
	"strconv"
)

type CatalogOperations interface {
	FindCatalogItem(catalogitem string) (CatalogItem, error)
}

type AdminCatalog struct {
	AdminCatalog *types.AdminCatalog
	client       *Client
}

type Catalog struct {
	Catalog *types.Catalog
	client  *Client
}

func NewCatalog(client *Client) *Catalog {
	return &Catalog{
		Catalog: new(types.Catalog),
		client:  client,
	}
}

func NewAdminCatalog(client *Client) *AdminCatalog {
	return &AdminCatalog{
		AdminCatalog: new(types.AdminCatalog),
		client:       client,
	}
}

func (adminCatalog *AdminCatalog) Delete(force, recursive bool) error {
	adminCatalogHREF := adminCatalog.client.VCDHREF
	adminCatalogHREF.Path += "/admin/catalog/" + adminCatalog.AdminCatalog.ID[19:]

	req := adminCatalog.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, "DELETE", adminCatalogHREF, nil)

	_, err := checkResp(adminCatalog.client.Http.Do(req))

	if err != nil {
		return fmt.Errorf("error deleting Catalog %s: %s", adminCatalog.AdminCatalog.ID, err)
	}

	return nil
}

func (adminCatalog *AdminCatalog) Update() (Task, error) {

	vcomp := &types.AdminCatalog{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        adminCatalog.AdminCatalog.Name,
		Description: adminCatalog.AdminCatalog.Description,
		IsPublished: adminCatalog.AdminCatalog.IsPublished,
	}

	adminCatalogHREF, err := url.ParseRequestURI(adminCatalog.AdminCatalog.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parsing admin catalog's href: %v", err)
	}

	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	b := bytes.NewBufferString(xml.Header + string(output))

	req := adminCatalog.client.NewRequest(map[string]string{}, "PUT", *adminCatalogHREF, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.catalog+xml")

	resp, err := checkResp(adminCatalog.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error updating catalog: %s : %s", err, s.Path)
	}

	catalog := NewAdminCatalog(adminCatalog.client)
	if err = decodeBody(resp, catalog.AdminCatalog); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}

	adminCatalog.AdminCatalog = catalog.AdminCatalog

	return Task{
		Task: catalog.AdminCatalog.Tasks.Task[0],
		c:    adminCatalog.client,
	}, nil
}

func (catalog *Catalog) FindCatalogItem(catalogitem string) (CatalogItem, error) {

	for _, cis := range catalog.Catalog.CatalogItems {
		for _, ci := range cis.CatalogItem {
			if ci.Name == catalogitem && ci.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				u, err := url.ParseRequestURI(ci.HREF)

				if err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				req := catalog.client.NewRequest(map[string]string{}, "GET", *u, nil)

				resp, err := checkResp(catalog.client.Http.Do(req))
				if err != nil {
					return CatalogItem{}, fmt.Errorf("error retreiving catalog: %s", err)
				}

				cat := NewCatalogItem(catalog.client)

				if err = decodeBody(resp, cat.CatalogItem); err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				// The request was successful
				return *cat, nil
			}
		}
	}

	return CatalogItem{}, fmt.Errorf("can't find catalog item: %s", catalogitem)
}
