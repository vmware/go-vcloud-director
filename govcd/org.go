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

// Interface for methods in common for Org and AdminOrg
type OrgOperations interface {
	GetCatalog(catalog string) (Catalog, error)
}

type Org struct {
	Org *types.Org
	c   *Client
}

func NewOrg(client *Client) *Org {
	return &Org{
		Org: new(types.Org),
		c:   client,
	}
}

// AdminOrg gives an admin representation of an org.
// Administrators can delete and update orgs with an admin org object.
// AdminOrg includes all members of the Org element, and adds several
// elements that can be viewed and modified only by system administrators.
type AdminOrg struct {
	AdminOrg *types.AdminOrg
	c        *Client
}

func NewAdminOrg(c *Client) *AdminOrg {
	return &AdminOrg{
		AdminOrg: new(types.AdminOrg),
		c:        c,
	}
}

// CreateCatalog creates a catalog with given name and description under the
// the given organization. Returns an AdminCatalog that contains a creation
// task.
func (adminOrg *AdminOrg) CreateCatalog(Name, Description string, isPublished bool) (AdminCatalog, error) {

	vcomp := &types.AdminCatalog{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        Name,
		Description: Description,
		IsPublished: isPublished,
	}

	s, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return AdminCatalog{}, fmt.Errorf("error parsing admin org's href: %v", err)
	}
	s.Path += "/catalogs"

	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	b := bytes.NewBufferString(xml.Header + string(output))

	req := adminOrg.c.NewRequest(map[string]string{}, "POST", *s, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.catalog+xml")

	resp, err := checkResp(adminOrg.c.Http.Do(req))
	if err != nil {
		return AdminCatalog{}, fmt.Errorf("error creating catalog: %s : %s", err, s.Path)
	}

	catalog := NewAdminCatalog(adminOrg.c)
	if err = decodeBody(resp, catalog.AdminCatalog); err != nil {
		return AdminCatalog{}, fmt.Errorf("error decoding task response: %s", err)
	}
	// Task is within the catalog
	return *catalog, nil

}

// If user specifies valid vdc name then this returns a vdc object.
// Otherwise it returns an empty vdc and an error.
func (org *Org) GetVdcByName(vdcname string) (Vdc, error) {
	HREF := ""
	for _, a := range org.Org.Link {
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
	req := org.c.NewRequest(map[string]string{}, "GET", *u, nil)
	resp, err := checkResp(org.c.Http.Do(req))
	if err != nil {
		return Vdc{}, fmt.Errorf("error getting vdc: %s", err)
	}

	vdc := NewVdc(org.c)
	if err = decodeBody(resp, vdc.Vdc); err != nil {
		return Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
	}
	// The request was successful
	return *vdc, nil
}

//   Deletes the org, returning an error if the vCD call fails.
func (adminOrg *AdminOrg) Delete(force bool, recursive bool) error {
	if force && recursive {
		//undeploys vapps
		err := adminOrg.undeployAllVApps()
		if err != nil {
			return fmt.Errorf("Could not undeploy with error %#v", err)
		}
		//removes vapps
		err = adminOrg.removeAllVApps()
		if err != nil {
			return fmt.Errorf("Could not remove vapp with error %#v", err)
		}
		//removes catalogs
		err = adminOrg.removeCatalogs()
		if err != nil {
			return fmt.Errorf("Could not remove all catalogs %#v", err)
		}
		//removes networks
		err = adminOrg.removeAllOrgNetworks()
		if err != nil {
			return fmt.Errorf("Could not remove all networks %#v", err)
		}
		//removes org vdcs
		err = adminOrg.removeAllOrgVDCs()
		if err != nil {
			return fmt.Errorf("Could not remove all vdcs %#v", err)
		}
	}
	// Disable org
	err := adminOrg.Disable()
	if err != nil {
		return fmt.Errorf("error disabling Org %s: %s", adminOrg.AdminOrg.ID, err)
	}
	// Get admin HREF
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %v", adminOrg.AdminOrg.HREF, err)
	}
	req := adminOrg.c.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, "DELETE", *orgHREF, nil)
	_, err = checkResp(adminOrg.c.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error deleting Org %s: %s", adminOrg.AdminOrg.ID, err)
	}
	return nil
}

// Disables the org. Returns an error if the call to vCD fails.
func (adminOrg *AdminOrg) Disable() error {
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return fmt.Errorf("error getting AdminOrg HREF %s : %v", adminOrg.AdminOrg.HREF, err)
	}
	orgHREF.Path += "/action/disable"
	req := adminOrg.c.NewRequest(map[string]string{}, "POST", *orgHREF, nil)
	_, err = checkResp(adminOrg.c.Http.Do(req))
	return err
}

//   Updates the Org definition from current org struct contents.
//   Any differences that may be legally applied will be updated.
//   Returns an error if the call to vCD fails.
func (adminOrg *AdminOrg) Update() (Task, error) {
	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        adminOrg.AdminOrg.Name,
		IsEnabled:   adminOrg.AdminOrg.IsEnabled,
		FullName:    adminOrg.AdminOrg.FullName,
		OrgSettings: adminOrg.AdminOrg.OrgSettings,
	}
	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	b := bytes.NewBufferString(xml.Header + string(output))
	// Update org
	orgHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error getting AdminOrg HREF %s : %v", adminOrg.AdminOrg.HREF, err)
	}
	req := adminOrg.c.NewRequest(map[string]string{}, "PUT", *orgHREF, b)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")
	resp, err := checkResp(adminOrg.c.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error updating Org: %s", err)
	}
	// Create Return object
	task := NewTask(adminOrg.c)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, nil
}

// Undeploys every vapp within an organization
func (adminOrg *AdminOrg) undeployAllVApps() error {
	for _, a := range adminOrg.AdminOrg.Vdcs.Vdcs {
		adminVdcHREF, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}
		vdc, err := adminOrg.getVdcByAdminHREF(adminVdcHREF)
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
func (adminOrg *AdminOrg) removeAllVApps() error {
	for _, a := range adminOrg.AdminOrg.Vdcs.Vdcs {
		adminVdcHREF, err := url.Parse(a.HREF)
		if err != nil {
			return err
		}
		vdc, err := adminOrg.getVdcByAdminHREF(adminVdcHREF)
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
func (adminOrg *AdminOrg) getVdcByAdminHREF(url *url.URL) (*Vdc, error) {
	// get non admin vdc path
	non_admin := strings.Split(url.Path, "/admin")
	url.Path = non_admin[0] + non_admin[1]
	req := adminOrg.c.NewRequest(map[string]string{}, "GET", *url, nil)
	resp, err := checkResp(adminOrg.c.Http.Do(req))
	if err != nil {
		return &Vdc{}, fmt.Errorf("error retreiving vdc: %s", err)
	}

	vdc := NewVdc(adminOrg.c)
	if err = decodeBody(resp, vdc.Vdc); err != nil {
		return &Vdc{}, fmt.Errorf("error decoding vdc response: %s", err)
	}
	return vdc, nil
}

// Removes all vdcs in a org
func (adminOrg *AdminOrg) removeAllOrgVDCs() error {
	for _, a := range adminOrg.AdminOrg.Vdcs.Vdcs {
		// Get admin Vdc HREF
		adminVdcUrl := adminOrg.c.VCDHREF
		adminVdcUrl.Path += "/admin/vdc/" + strings.Split(a.HREF, "/vdc/")[1] + "/action/disable"
		req := adminOrg.c.NewRequest(map[string]string{}, "POST", adminVdcUrl, nil)
		_, err := checkResp(adminOrg.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error disabling vdc: %s", err)
		}
		// Get admin vdc HREF for normal deletion
		adminVdcUrl.Path = strings.Split(adminVdcUrl.Path, "/action/disable")[0]
		req = adminOrg.c.NewRequest(map[string]string{
			"recursive": "true",
			"force":     "true",
		}, "DELETE", adminVdcUrl, nil)
		resp, err := checkResp(adminOrg.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting vdc: %s", err)
		}
		task := NewTask(adminOrg.c)
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
func (adminOrg *AdminOrg) removeAllOrgNetworks() error {
	for _, a := range adminOrg.AdminOrg.Networks.Networks {
		// Get Network HREF
		networkHREF := adminOrg.c.VCDHREF
		networkHREF.Path += "/admin/network/" + strings.Split(a.HREF, "/network/")[1] //gets id
		req := adminOrg.c.NewRequest(map[string]string{}, "DELETE", networkHREF, nil)
		resp, err := checkResp(adminOrg.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting newtork: %s, %s", err, networkHREF.Path)
		}

		task := NewTask(adminOrg.c)
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
func (adminOrg *AdminOrg) removeCatalogs() error {
	for _, a := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		catalogHREF := adminOrg.c.VCDHREF
		catalogHREF.Path += "/admin/catalog/" + strings.Split(a.HREF, "/catalog/")[1] //gets id
		req := adminOrg.c.NewRequest(map[string]string{
			"force":     "true",
			"recursive": "true",
		}, "DELETE", catalogHREF, nil)
		_, err := checkResp(adminOrg.c.Http.Do(req))
		if err != nil {
			return fmt.Errorf("error deleting catalog: %s, %s", err, catalogHREF.Path)
		}
	}
	return nil

}

func (adminOrg *AdminOrg) GetAdminCatalog(catalog string) (AdminCatalog, error) {
	for _, a := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if a.Name == catalog {
			catalogURL, err := url.ParseRequestURI(a.HREF)
			if err != nil {
				return AdminCatalog{}, fmt.Errorf("error decoding catalog url: %s", err)
			}
			req := adminOrg.c.NewRequest(map[string]string{}, "GET", *catalogURL, nil)
			resp, err := checkResp(adminOrg.c.Http.Do(req))
			if err != nil {
				return AdminCatalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}
			adminCatalog := NewAdminCatalog(adminOrg.c)

			if err = decodeBody(resp, adminCatalog.AdminCatalog); err != nil {
				return AdminCatalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}

			// The request was successful
			return *adminCatalog, nil
		}
	}
	return AdminCatalog{}, fmt.Errorf("can't find catalog: %s", catalog)
}

func (adminOrg *AdminOrg) GetCatalog(catalog string) (Catalog, error) {
	for _, a := range adminOrg.AdminOrg.Catalogs.Catalog {
		// Get Catalog HREF
		if a.Name == catalog {
			splitbyAdminHREF := strings.Split(a.HREF, "/admin")
			catalogHREF := splitbyAdminHREF[0] + splitbyAdminHREF[1]
			catalogURL, err := url.ParseRequestURI(catalogHREF)
			if err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog url: %s", err)
			}
			req := adminOrg.c.NewRequest(map[string]string{}, "GET", *catalogURL, nil)
			resp, err := checkResp(adminOrg.c.Http.Do(req))
			if err != nil {
				return Catalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}
			cat := NewCatalog(adminOrg.c)

			if err = decodeBody(resp, cat.Catalog); err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}

			// The request was successful
			return *cat, nil
		}
	}
	return Catalog{}, fmt.Errorf("can't find catalog: %s", catalog)
}

func (org *Org) GetCatalog(catalog string) (Catalog, error) {

	for _, av := range org.Org.Link {
		if av.Rel == "down" && av.Type == "application/vnd.vmware.vcloud.catalog+xml" && av.Name == catalog {
			u, err := url.ParseRequestURI(av.HREF)

			if err != nil {
				return Catalog{}, fmt.Errorf("error decoding org response: %s", err)
			}

			req := org.c.NewRequest(map[string]string{}, "GET", *u, nil)

			resp, err := checkResp(org.c.Http.Do(req))
			if err != nil {
				return Catalog{}, fmt.Errorf("error retreiving catalog: %s", err)
			}

			cat := NewCatalog(org.c)

			if err = decodeBody(resp, cat.Catalog); err != nil {
				return Catalog{}, fmt.Errorf("error decoding catalog response: %s", err)
			}

			// The request was successful
			return *cat, nil

		}
	}

	return Catalog{}, fmt.Errorf("can't find catalog: %s", catalog)
}
