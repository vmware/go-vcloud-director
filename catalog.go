/*
* @Author: frapposelli
* @Date:   2014-10-22 12:18:21
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:11:00
 */

package govcloudair

import (
	"fmt"
)

type CatalogRelLinks struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
	HREF string `xml:"href,attr"`
}

type CatalogItemLinks struct {
	Id   string `xml:"id,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
	HREF string `xml:"href,attr"`
}

type CatalogItems struct {
	CatalogItem []CatalogItemLinks
}

type Catalog struct {
	Link          []CatalogRelLinks
	Description   string `xml:"Description"`
	CatalogItems  []CatalogItems
	IsPublished   bool   `xml:"IsPublished"`
	DateCreated   string `xml:"DateCreated"`
	VersionNumber int    `xml:"VersionNumber"`
}

func (c *Client) RetrieveCatalog(catalogid string) (Catalog, error) {

	req, err := c.NewRequest(map[string]string{}, "GET", fmt.Sprintf("/catalog/%s", catalogid), "")
	if err != nil {
		return Catalog{}, err
	}

	resp, err := checkResp(c.Http.Do(req))
	if err != nil {
		return Catalog{}, fmt.Errorf("Error retreiving catalog: %s", err)
	}

	catalog := new(Catalog)

	err = decodeBody(resp, catalog)

	if err != nil {
		return Catalog{}, fmt.Errorf("Error decoding catalog response: %s", err)
	}

	// The request was successful
	return *catalog, nil
}

func (c *Catalog) FindCatalogItemId(catalogitem string) string {

	for _, cis := range c.CatalogItems {
		for _, ci := range cis.CatalogItem {
			if ci.Name == catalogitem && ci.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				return ci.Id
			}
		}
	}

	return ""
}
