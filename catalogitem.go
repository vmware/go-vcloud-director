/*
* @Author: frapposelli
* @Date:   2014-10-22 13:45:20
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:11:01
 */

package govcloudair

import (
	"fmt"
	"net/url"
	"strings"
)

type CatalogItemRelLinks struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
	HREF string `xml:"href,attr"`
}

type Entities struct {
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
	HREF string `xml:"href,attr"`
}

type CatalogItem struct {
	Link          []CatalogItemRelLinks
	Description   string `xml:"Description"`
	Entity        Entities
	IsPublished   bool   `xml:"IsPublished"`
	DateCreated   string `xml:"DateCreated"`
	VersionNumber int    `xml:"VersionNumber"`
}

func (c *Client) RetrieveCatalogItem(catalogitemid string) (CatalogItem, error) {

	req, err := c.NewRequest(map[string]string{}, "GET", fmt.Sprintf("/catalogItem/%s", catalogitemid), "")
	if err != nil {
		return CatalogItem{}, err
	}

	resp, err := checkResp(c.Http.Do(req))
	if err != nil {
		return CatalogItem{}, fmt.Errorf("Error retreiving catalogitem: %s", err)
	}

	catalogitem := new(CatalogItem)

	err = decodeBody(resp, catalogitem)

	if err != nil {
		return CatalogItem{}, fmt.Errorf("Error decoding catalogitem response: %s", err)
	}

	// The request was successful
	return *catalogitem, nil
}

func (ci *CatalogItem) FindVappTemplateId() string {
	u, _ := url.Parse(ci.Entity.HREF)
	urlPath := strings.SplitAfter(u.Path, "/")
	return urlPath[len(urlPath)-1]
}
