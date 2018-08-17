package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"net/url"
	"strings"
)

// Creates an Organization based on settings, network, and org name.
// The Organization created will have these settings specified in the
// settings parameter. The settings variable is defined in types.go.
// Method will fail unless user has an admin token.
func CreateOrg(c *VCDClient, name string, fullName string, isEnabled bool, settings *types.OrgSettings) (Task, error) {
	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        name,
		IsEnabled:   isEnabled,
		FullName:    fullName,
		OrgSettings: settings,
	}
	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	b := bytes.NewBufferString(xml.Header + string(output))
	// Make Request
	u := c.Client.VCDHREF
	u.Path += "/admin/orgs"
	req := c.Client.NewRequest(map[string]string{}, "POST", u, b)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")
	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error instantiating a new Org: %s", err)
	}

	task := NewTask(&c.Client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, nil
}

// If user specifies a valid organization name, then this returns a
// organization object. Otherwise it returns an error and an empty
// Org object
func GetOrgByName(c *VCDClient, orgname string) (Org, error) {
	orgUrl, err := getOrgHREF(c, orgname)
	if err != nil {
		return Org{}, fmt.Errorf("Cannot find the url of the org: %s", err)
	}
	orgHREF, err := url.ParseRequestURI(orgUrl)
	if err != nil {
		return Org{}, fmt.Errorf("Error parsing org href: %v", err)
	}
	req := c.Client.NewRequest(map[string]string{}, "GET", *orgHREF, nil)
	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Org{}, fmt.Errorf("error retreiving org: %s", err)
	}

	org := NewOrg(&c.Client)
	if err = decodeBody(resp, org.Org); err != nil {
		return Org{}, fmt.Errorf("error decoding org response: %s", err)
	}
	return *org, nil
}

// If user specifies valid organization name, then this returns an admin organization object
// Otherwise returns an empty AdminOrg and an error.
func GetAdminOrgByName(c *VCDClient, orgname string) (AdminOrg, error) {
	orgUrl, err := getOrgHREF(c, orgname)
	if err != nil {
		return AdminOrg{}, fmt.Errorf("Cannot find OrgHREF: %s", err)
	}
	orgHREF := c.Client.VCDHREF
	orgHREF.Path += "/admin/org/" + strings.Split(orgUrl, "/org/")[1]
	req := c.Client.NewRequest(map[string]string{}, "GET", orgHREF, nil)
	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return AdminOrg{}, fmt.Errorf("error retreiving org: %s", err)
	}
	org := NewAdminOrg(&c.Client)
	if err = decodeBody(resp, org.AdminOrg); err != nil {
		return AdminOrg{}, fmt.Errorf("error decoding org response: %s", err)
	}
	return *org, nil
}

// Returns the HREF of the org with the name orgname
func getOrgHREF(c *VCDClient, orgname string) (string, error) {
	orgListHREF := c.Client.VCDHREF
	orgListHREF.Path += "/org"
	req := c.Client.NewRequest(map[string]string{}, "GET", orgListHREF, nil)
	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return "", fmt.Errorf("error retreiving org list: %s", err)
	}
	orgList := new(types.OrgList)
	if err = decodeBody(resp, orgList); err != nil {
		return "", fmt.Errorf("error decoding response: %s", err)
	}
	// Look for orgname within OrgList
	for _, a := range orgList.Org {
		if a.Name == orgname {
			return a.HREF, nil
		}
	}
	return "", fmt.Errorf("Couldn't find org with name: %s", orgname)
}
