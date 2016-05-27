/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"

	types "github.com/vmware/govcloudair/types/v56"
)

// FetchOrgList fetches the org list from a set of links that hopefully contain a link to an org list
func FetchOrgList(links types.LinkList, client Client) (*OrgList, error) {
	lnk := links.ForType(types.MimeOrgList, types.RelDown)
	if lnk == nil {
		return nil, errors.New("no link for orgList")
	}

	var orgList OrgList
	u, err := url.ParseRequestURI(lnk.HREF)
	if err != nil {
		return nil, err
	}

	resp, err := client.DoHTTP(client.NewRequest(nil, "GET", u, nil))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("could not complete request with vca, because (status %d) %s", resp.StatusCode, resp.Status)
	}

	dec := xml.NewDecoder(resp.Body)
	if err := dec.Decode(&orgList); err != nil {
		return nil, err
	}

	return &orgList, nil
}

// OrgList represents a list of organizations.
// Type: OrgListType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Since: 0.9
type OrgList struct {
	HREF string `xml:"href,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`

	// ResourceType
	Links types.LinkList `xml:"Link,omitempty"`

	// OrgListType
	Orgs []types.Reference `xml:"Org,omitempty"`
}

// FirstOrg retrieves the first organization from the org list
func (o *OrgList) FirstOrg(client Client) (*types.Org, error) {
	if len(o.Orgs) == 0 {
		return nil, errors.New("orgList has no orgs, can't get the first")
	}

	ref := o.Orgs[0]
	u, err := url.ParseRequestURI(ref.HREF)
	if err != nil {
		return nil, err
	}
	var org types.Org
	resp, err := client.DoHTTP(client.NewRequest(nil, "GET", u, nil))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("could not complete request with vca, because (status %d) %s", resp.StatusCode, resp.Status)
	}

	dec := xml.NewDecoder(resp.Body)
	if err := dec.Decode(&org); err != nil {
		return nil, err
	}
	// if err := client.XMLRequest(HTTPGet, ref.HREF, ref.Type, nil, &org); err != nil {
	// 	return nil, err
	// }

	return &org, nil
}

// Org a vcloud org client
type Org struct {
	Org *types.Org
	c   Client
}

// NewOrg creates a new org client
func NewOrg(c Client) *Org {
	return &Org{
		Org: new(types.Org),
		c:   c,
	}
}

// FindCatalog finds a catalog in the org
func (o *Org) FindCatalog(catalog string) (Catalog, error) {

	for _, av := range o.Org.Link {
		if av.Rel == "down" && av.Type == "application/vnd.vmware.vcloud.catalog+xml" && av.Name == catalog {
			u, err := url.ParseRequestURI(av.HREF)

			if err != nil {
				return Catalog{}, fmt.Errorf("error decoding org response: %s", err)
			}

			req := o.c.NewRequest(map[string]string{}, "GET", u, nil)

			resp, err := checkResp(o.c.DoHTTP(req))
			if err != nil {
				return Catalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}

			cat := NewCatalog(o.c)

			if err = decodeBody(resp, cat.Catalog); err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}

			// The request was successful
			return *cat, nil

		}
	}

	return Catalog{}, fmt.Errorf("can't find catalog: %s", catalog)
}
