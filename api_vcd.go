package govcloudair

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type VCDClient struct {
	OrgRef url.URL // vCloud Director OrgRef
	Client Client  // Client for the underlying VCD instance
}

type supportedVersions struct {
	VersionInfo struct {
		Version  string `xml:"Version"`
		LoginUrl string `xml:"LoginUrl"`
	} `xml:"VersionInfo"`
}

type vCloudOrg struct {
	Link []struct {
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
		Name string `xml:"name,attr"`
		HREF string `xml:"href,attr"`
	} `xml:"Link"`
}

func (c *VCDClient) vcdloginurl() (u *url.URL, err error) {

	s := c.Client.VCDVDCHREF
	s.Path += "/versions"

	// No point in checking for errors here
	req := c.Client.NewRequest(map[string]string{}, "GET", s, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return &url.URL{}, err
	}
	defer resp.Body.Close()

	supportedVersions := new(supportedVersions)

	err = decodeBody(resp, supportedVersions)

	if err != nil {
		return u, fmt.Errorf("error decoding versions response: %s", err)
	}

	u, err = url.Parse(supportedVersions.VersionInfo.LoginUrl)
	if err != nil {
		return u, fmt.Errorf("couldn't find a LoginUrl in versions")
	}
	return u, nil
}

func (c *VCDClient) vcdauthorize(user, pass, org string, sessionRef *url.URL) error {

	if user == "" {
		user = os.Getenv("VCLOUD_USERNAME")
	}

	if pass == "" {
		pass = os.Getenv("VCLOUD_PASSWORD")
	}

	if org == "" {
		org = os.Getenv("VCLOUD_ORG")
	}

	// No point in checking for errors here
	req := c.Client.NewRequest(map[string]string{}, "POST", *sessionRef, nil)

	// Set Basic Authentication Header
	req.SetBasicAuth(user+"@"+org, pass)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/*+xml;version=5.5")

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Store the authentication header
	c.Client.VCDToken = resp.Header.Get("x-vcloud-authorization")
	c.Client.VCDAuthHeader = "x-vcloud-authorization"

	session := new(session)
	err = decodeBody(resp, session)

	if err != nil {
		fmt.Errorf("error decoding session response: %s", err)
	}

	// Loop in the session struct to find the organization.
	for _, s := range session.Link {
		if s.Type == "application/vnd.vmware.vcloud.org+xml" && s.Rel == "down" {
			u, err := url.Parse(s.HREF)
			c.OrgRef = *u
			return err
		}
	}

	return fmt.Errorf("couldn't find a Organization in current session")
}

func (c *VCDClient) retrieveVDCHREF() error {

	req := c.Client.NewRequest(map[string]string{}, "GET", c.OrgRef, nil)

	// Add the Accept header for vCD
	//req.Header.Add("Accept", "application/*+xml;version=5.5")

	// TODO: wrap into checkresp to parse error
	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error processing org: %s", err)
	}
	defer resp.Body.Close()

	vcloudorg := new(vCloudOrg)

	err = decodeBody(resp, vcloudorg)

	if err != nil {
		return fmt.Errorf("error decoding vcloudorg response: %s", err)
	}

	// Get the VDC ref from the Org
	for _, s := range vcloudorg.Link {
		if s.Type == "application/vnd.vmware.vcloud.vdc+xml" && s.Rel == "down" {
			u, err := url.Parse(s.HREF)
			if err != nil {
				return err
			}
			c.Client.VCDVDCHREF = *u
			return nil
		}
	}

	return fmt.Errorf("error finding the organization VDC")
}

func NewVCDClient(vcdEndpoint url.URL) *VCDClient {

	return &VCDClient{
		Client: Client{
			APIVersion: "5.5",
			VCDVDCHREF: vcdEndpoint,
			Http:       http.Client{Transport: &http.Transport{TLSHandshakeTimeout: 120 * time.Second}},
		},
	}
}

// Authenticate is an helper function that performs a login in vCloud Director.
func (c *VCDClient) Authenticate(username, password, org string) (Vdc, error) {

	// LoginUrl
	sessionRef, err := c.vcdloginurl()
	if err != nil {
		return Vdc{}, fmt.Errorf("error finding LoginUrl: %s", err)
	}
	// Authorize
	err = c.vcdauthorize(username, password, org, sessionRef)
	if err != nil {
		return Vdc{}, fmt.Errorf("error Authorizing: %s", err)
	}

	// Get Org VDC
	err = c.retrieveVDCHREF()
	if err != nil {
		return Vdc{}, fmt.Errorf("error Acquiring VDC url: %s", err)
	}

	v, err := c.Client.retrieveVDC()
	if err != nil {
		return Vdc{}, fmt.Errorf("error Acquiring VDC: %s", err)
	}

	return v, nil
}
