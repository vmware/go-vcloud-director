/*
* @Author: frapposelli
* @Date:   2014-10-21 17:47:49
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:11:01
 */

package govcloudair

import (
	"fmt"
	"net/url"
	"strings"
)

type OrgRelLinks struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
	HREF string `xml:"href,attr"`
}

type Org struct {
	Link        []OrgRelLinks
	Description string `xml:"Description"`
	FullName    string `xml:"FullName"`
}

func (c *Client) RetrieveOrg(orgid string) (Org, error) {

	req, err := c.NewRequest(map[string]string{}, "GET", fmt.Sprintf("/org/%s", orgid), "")
	if err != nil {
		return Org{}, err
	}

	resp, err := checkResp(c.Http.Do(req))
	if err != nil {
		return Org{}, fmt.Errorf("Error retreiving org: %s", err)
	}

	org := new(Org)

	err = decodeBody(resp, org)

	if err != nil {
		return Org{}, fmt.Errorf("Error decoding org response: %s", err)
	}

	// The request was successful
	return *org, nil
}

func (o *Org) FindOrgCatalogId(catalog string) string {

	for _, av := range o.Link {
		if av.Rel == "down" && av.Type == "application/vnd.vmware.vcloud.catalog+xml" && av.Name == catalog {
			// Build the backend url
			u, _ := url.Parse(av.HREF)
			// Get the orgguid from the url path
			urlPath := strings.SplitAfter(u.Path, "/")
			return urlPath[len(urlPath)-1]
		}
	}

	return ""
}

func (o *Org) FindOrgVDCId(vdc string) string {

	for _, av := range o.Link {
		if av.Rel == "down" && av.Type == "application/vnd.vmware.vcloud.vdc+xml" && av.Name == vdc {
			// Build the backend url
			u, _ := url.Parse(av.HREF)
			// Get the orgguid from the url path
			urlPath := strings.SplitAfter(u.Path, "/")
			return urlPath[len(urlPath)-1]
		}
	}

	return ""
}

func (o *Org) FindOrgNetworkId(network string) string {

	for _, av := range o.Link {
		if av.Rel == "down" && av.Type == "application/vnd.vmware.vcloud.orgNetwork+xml" && av.Name == network {
			// Build the backend url
			u, _ := url.Parse(av.HREF)
			// Get the orgguid from the url path
			urlPath := strings.SplitAfter(u.Path, "/")
			return urlPath[len(urlPath)-1]
		}
	}

	return ""
}
