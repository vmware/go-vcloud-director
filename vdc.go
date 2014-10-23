/*
* @Author: frapposelli
* @Date:   2014-10-20 15:18:51
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:11:05
 */

package govcloudair

import (
	"fmt"
	"net/url"
	"strings"
)

type VdcRelLinks struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	HREF string `xml:"href,attr"`
}

type ResourceEntity struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	HREF string `xml:"href,attr"`
}

type Network struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	HREF string `xml:"href,attr"`
}

type VdcStorageProfile struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	HREF string `xml:"href,attr"`
}

type AvailableNetworks struct {
	Network []Network
}

type ResourceEntities struct {
	ResourceEntity []ResourceEntity
}

type VdcStorageProfiles struct {
	VdcStorageProfile []VdcStorageProfile
}

type Capabilities struct {
	SupportedHardwareVersions []string `xml:"SupportedHardwareVersions>SupportedHardwareVersion"`
}

type Cpu struct {
	Units     string
	Allocated int
	Limit     int
	Reserved  int
	Used      int
	Overhead  int
}

type Memory struct {
	Units     string
	Allocated int
	Limit     int
	Reserved  int
	Used      int
	Overhead  int
}

type ComputeCapacity struct {
	Cpu    Cpu
	Memory Memory
}

type Vdc struct {
	Link               []VdcRelLinks
	AllocationModel    string `xml:"AllocationModel"`
	ComputeCapacity    []ComputeCapacity
	ResourceEntities   []ResourceEntities
	AvailableNetworks  []AvailableNetworks
	Capabilities       []Capabilities
	NicQuota           int  `xml:"NicQuota"`
	NetworkQuota       int  `xml:"NetworkQuota"`
	UsedNetworkCount   int  `xml:"UsedNetworkCount"`
	VmQuota            int  `xml:"VmQuota"`
	IsEnabled          bool `xml:"IsEnabled"`
	VdcStorageProfiles []VdcStorageProfiles
}

func (c *Client) RetrieveVDC() (Vdc, error) {

	req, err := c.NewRequest(map[string]string{}, "GET", fmt.Sprintf("/vdc/%s", c.VDC), "")
	if err != nil {
		return Vdc{}, err
	}

	resp, err := checkResp(c.Http.Do(req))
	if err != nil {
		return Vdc{}, fmt.Errorf("Error retreiving vdc: %s", err)
	}

	vdc := new(Vdc)

	err = decodeBody(resp, vdc)

	if err != nil {
		return Vdc{}, fmt.Errorf("Error decoding vdc response: %s", err)
	}

	// The request was successful
	return *vdc, nil
}

func (v *Vdc) FindVDCNetworkId(network string) string {

	for _, an := range v.AvailableNetworks {
		for _, n := range an.Network {
			if n.Name == network {
				// Build the backend url
				u, _ := url.Parse(n.HREF)
				// Get the orgguid from the url path
				urlPath := strings.SplitAfter(u.Path, "/")
				return urlPath[len(urlPath)-1]
			}
		}
	}

	return ""
}

func (v *Vdc) FindVDCStorageProfileId(storageprofile string) string {

	for _, sp := range v.VdcStorageProfiles {
		for _, s := range sp.VdcStorageProfile {
			if s.Name == storageprofile {
				// Build the backend url
				u, _ := url.Parse(s.HREF)
				// Get the orgguid from the url path
				urlPath := strings.SplitAfter(u.Path, "/")
				return urlPath[len(urlPath)-1]
			}
		}
	}

	return ""
}

func (v *Vdc) FindVDCOrgId() string {

	for _, av := range v.Link {
		if av.Rel == "up" && av.Type == "application/vnd.vmware.vcloud.org+xml" {
			// Build the backend url
			u, _ := url.Parse(av.HREF)
			// Get the orgguid from the url path
			urlPath := strings.SplitAfter(u.Path, "/")
			return urlPath[len(urlPath)-1]
		}
	}

	return ""
}
