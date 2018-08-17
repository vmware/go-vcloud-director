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
	"strings"
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

// AdminOrg gives an admin representation of an org
// Users can delete, update orgs with an admin org object
// AdminOrg users have to get an org representation to use find Catalogs
type AdminOrg struct {
	AdminOrg *types.AdminOrg
	Org
}

func NewAdminOrg(c *Client) *AdminOrg {
	return &AdminOrg{
		AdminOrg: new(types.AdminOrg),
		Org: Org{
			c: c,
		},
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

//   Deletes the org, returning an error if the vCD call fails.
func (o *AdminOrg) Delete(force bool, recursive bool) error {
	if force && recursive {
		//undeploys vapps
		err := o.undeployAllVApps()
		if err != nil {
			return fmt.Errorf("Could not undeploy with error %#v", err)
		}
		//removes vapps
		err = o.removeAllVApps()
		if err != nil {
			return fmt.Errorf("Could not remove vapp with error %#v", err)
		}
		//removes catalogs
		err = o.removeCatalogs()
		if err != nil {
			return fmt.Errorf("Could not remove all catalogs %#v", err)
		}
		//removes networks
		err = o.removeAllOrgNetworks()
		if err != nil {
			return fmt.Errorf("Could not remove all networks %#v", err)
		}
		//removes org vdcs
		err = o.removeAllOrgVDCs()
		if err != nil {
			return fmt.Errorf("Could not remove all vdcs %#v", err)
		}
	}
	// Disable org
	err := o.Disable()
	if err != nil {
		return fmt.Errorf("error disabling Org %s: %s", o.AdminOrg.ID, err)
	}
	// Get admin HREF
	orgHREF, err := url.ParseRequestURI(o.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %v", o.AdminOrg.HREF, err)
	}
	req := o.c.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, "DELETE", *orgHREF, nil)
	_, err = checkResp(o.c.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error deleting Org %s: %s", o.AdminOrg.ID, err)
	}
	return nil
}

// Disables the org. Returns an error if the call to vCD fails.
func (o *AdminOrg) Disable() error {
	orgHREF, err := url.ParseRequestURI(o.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %v", o.AdminOrg.HREF, err)
	}
	orgHREF.Path += "/action/disable"
	req := o.c.NewRequest(map[string]string{}, "POST", *orgHREF, nil)
	_, err = checkResp(o.c.Http.Do(req))
	return err
}

//   Updates the Org definition from current org struct contents.
//   Any differences that may be legally applied will be updated.
//   Returns an error if the call to vCD fails.
func (o *AdminOrg) Update() (Task, error) {
	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        o.AdminOrg.Name,
		IsEnabled:   o.AdminOrg.IsEnabled,
		FullName:    o.AdminOrg.FullName,
		OrgSettings: o.AdminOrg.OrgSettings,
	}
	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	b := bytes.NewBufferString(xml.Header + string(output))
	// Update org
	orgHREF, err := url.ParseRequestURI(o.AdminOrg.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error getting AdminOrg HREF %s : %v", o.AdminOrg.HREF, err)
	}
	req := o.c.NewRequest(map[string]string{}, "PUT", *orgHREF, b)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")
	resp, err := checkResp(o.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error updating Org: %s", err)
	}
	// Create Return object
	task := NewTask(o.c)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, nil
}

// Undeploys every vapp within an organization
func (o *AdminOrg) undeployAllVApps() error {
	for _, a := range o.AdminOrg.Vdcs.Vdcs {
		adminVdcHREF, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}
		vdc, err := o.getVdcByAdminHREF(adminVdcHREF)
		if err != nil {
			return fmt.Errorf("Error retrieving vapp with url: %s and with error %s", adminVdcHREF.Path, err)
		}
		err = vdc.undeployAllVdcVApps()
		if err != nil {
			return fmt.Errorf("Error deleting vapp: %s", err)
		}
	}
	return nil
}

// Deletes every vapp within an organization
func (o *AdminOrg) removeAllVApps() error {
	for _, a := range o.AdminOrg.Vdcs.Vdcs {
		adminVdcHREF, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}
		vdc, err := o.getVdcByAdminHREF(adminVdcHREF)
		if err != nil {
			return fmt.Errorf("Error retrieving vapp with url: %s and with error %s", adminVdcHREF.Path, err)
		}
		err = vdc.removeAllVdcVApps()
		if err != nil {
			return fmt.Errorf("Error deleting vapp: %s", err)
		}
	}
	return nil
}

// Gets a vdc within org associated with an admin vdc url u
func (o *AdminOrg) getVdcByAdminHREF(url *url.URL) (*Vdc, error) {
	// get non admin vdc path
	non_admin := strings.Split(url.Path, "/admin")
	url.Path = non_admin[0] + non_admin[1]
	req := o.c.NewRequest(map[string]string{}, "GET", *url, nil)
	resp, err := checkResp(o.c.Http.Do(req))
	if err != nil {
		return &Vdc{}, fmt.Errorf("error retreiving vdc: %s", err)
	}

	vdc := NewVdc(o.c)
	if err = decodeBody(resp, vdc.Vdc); err != nil {
		return &Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
	}
	return vdc, nil
}

// Removes all vdcs in a org
func (o *AdminOrg) removeAllOrgVDCs() error {
	for _, a := range o.AdminOrg.Vdcs.Vdcs {
		// Get admin Vdc HREF
		adminVdcUrl := o.c.VCDHREF
		adminVdcUrl.Path += "/admin/vdc/" + strings.Split(a.HREF, "/vdc/")[1] + "/action/disable"
		req := o.c.NewRequest(map[string]string{}, "POST", adminVdcUrl, nil)
		_, err := checkResp(o.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error disabling vdc: %s", err)
		}
		// Get admin vdc HREF for normal deletion
		adminVdcUrl.Path = strings.Split(adminVdcUrl.Path, "/action/disable")[0]
		req = o.c.NewRequest(map[string]string{
			"recursive": "true",
			"force":     "true",
		}, "DELETE", adminVdcUrl, nil)
		resp, err := checkResp(o.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting vdc: %s", err)
		}
		task := NewTask(o.c)
		if err = decodeBody(resp, task.Task); err != nil {
			return fmt.Errorf("error decoding task response: %s", err)
		}
		if task.Task.Status == "error" {
			return fmt.Errorf("vdc not properly destroyed")
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Couldn't finish removing vdc %#v", err)
		}

	}

	return nil
}

// Removes All networks in the org
func (o *AdminOrg) removeAllOrgNetworks() error {
	for _, a := range o.AdminOrg.Networks.Networks {
		// Get Network HREF
		networkHREF := o.c.VCDHREF
		networkHREF.Path += "/admin/network/" + strings.Split(a.HREF, "/network/")[1] //gets id
		req := o.c.NewRequest(map[string]string{}, "DELETE", networkHREF, nil)
		resp, err := checkResp(o.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting newtork: %s, %s", err, networkHREF.Path)
		}

		task := NewTask(o.c)
		if err = decodeBody(resp, task.Task); err != nil {
			return fmt.Errorf("error decoding task response: %s", err)
		}
		if task.Task.Status == "error" {
			return fmt.Errorf("network not properly destroyed")
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Couldn't finish removing network %#v", err)
		}
	}
	return nil
}

// Forced removal of all organization catalogs
func (o *AdminOrg) removeCatalogs() error {
	for _, a := range o.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		catalogHREF := o.c.VCDHREF
		catalogHREF.Path += "/admin/catalog/" + strings.Split(a.HREF, "/catalog/")[1] //gets id
		req := o.c.NewRequest(map[string]string{
			"force":     "true",
			"recursive": "true",
		}, "DELETE", catalogHREF, nil)
		_, err := checkResp(o.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting catalog: %s, %s", err, catalogHREF.Path)
		}
	}
	return nil

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
