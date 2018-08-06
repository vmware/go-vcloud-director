/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	types "github.com/vmware/go-vcloud-director/types/v56"
)

type Org struct {
	Org *types.Org
	c   *Client
}

func NewOrg(c *Client) *Org {
	return &Org{
		Org: new(types.Org),
		c:   c,
	}
}

// If user specifies valid vdc name then this returns a vdc object.
// Otherwise it returns an empty vdc and an error.
func (o *Org) GetVdcByName(vdcname string) (Vdc, error) {
	HREF := ""
	for _, a := range o.Org.Link {
		if a.Type == "application/vnd.vmware.vcloud.vdc+xml" && a.Name == vdcname {
			HREF = a.HREF
			break
		}
	}
	if HREF == "" {
		return Vdc{}, fmt.Errorf("Error finding VDC from VDCName")
	}

	u, err := url.ParseRequestURI(HREF)
	if err != nil {
		return Vdc{}, fmt.Errorf("Error parsing url: %v", err)
	}
	req := o.c.NewRequest(map[string]string{}, "GET", *u, nil)
	resp, err := checkResp(o.c.Http.Do(req))
	if err != nil {
		return Vdc{}, fmt.Errorf("error getting vdc: %s", err)
	}

	vdc := NewVdc(o.c)
	if err = decodeBody(resp, vdc.Vdc); err != nil {
		return Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
	}
	// The request was successful
	return *vdc, nil
}

func (o *Org) FindCatalog(catalog string) (Catalog, error) {

	for _, av := range o.Org.Link {
		if av.Rel == "down" && av.Type == "application/vnd.vmware.vcloud.catalog+xml" && av.Name == catalog {
			u, err := url.ParseRequestURI(av.HREF)

			if err != nil {
				return Catalog{}, fmt.Errorf("error decoding org response: %s", err)
			}

			req := o.c.NewRequest(map[string]string{}, "GET", *u, nil)

			resp, err := checkResp(o.c.Http.Do(req))
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
