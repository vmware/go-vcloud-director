package govcd

import (
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"net/url"
)

// If user specifies a valid organization name, then this returns a
// organization object. Otherwise it returns an error and an empty
// Org object
func GetOrgByName(c *VCDClient, orgname string) (Org, error) {
	OrgHREF, err := getOrgHREF(c, orgname)
	if err != nil {
		return Org{}, fmt.Errorf("Cannot find the url of the org: %s", err)
	}
	u, err := url.ParseRequestURI(OrgHREF)
	if err != nil {
		return Org{}, fmt.Errorf("Error parsing org href: %v", err)
	}
	req := c.Client.NewRequest(map[string]string{}, "GET", *u, nil)
	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Org{}, fmt.Errorf("error retreiving org: %s", err)
	}

	o := NewOrg(&c.Client)
	if err = decodeBody(resp, o.Org); err != nil {
		return Org{}, fmt.Errorf("error decoding org response: %s", err)
	}
	return *o, nil
}

// Returns the HREF of the org with the name orgname
func getOrgHREF(c *VCDClient, orgname string) (string, error) {
	s := c.Client.VCDHREF
	s.Path += "/org"
	req := c.Client.NewRequest(map[string]string{}, "GET", s, nil)
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
