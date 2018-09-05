package govcd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	neturl "net/url"
	"sync"
	"time"
)

type VCDClient struct {
	Client      Client     // Client for the underlying VCD instance
	sessionHREF neturl.URL // HREF for the session API
	QueryHREF   neturl.URL // HREF for the query API
	Mutex       sync.Mutex
}

type supportedVersions struct {
	VersionInfo struct {
		Version  string `xml:"Version"`
		LoginUrl string `xml:"LoginUrl"`
	} `xml:"VersionInfo"`
}

func (cli *VCDClient) vcdloginurl() error {
	apiEndpoint := cli.Client.VCDHREF
	apiEndpoint.Path += "/versions"
	// No point in checking for errors here
	req := cli.Client.NewRequest(map[string]string{}, "GET", apiEndpoint, nil)
	resp, err := checkResp(cli.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	supportedVersions := new(supportedVersions)
	err = decodeBody(resp, supportedVersions)
	if err != nil {
		return fmt.Errorf("error decoding versions response: %s", err)
	}
	url, err := neturl.Parse(supportedVersions.VersionInfo.LoginUrl)
	if err != nil {
		return fmt.Errorf("couldn't find a LoginUrl in versions")
	}
	cli.sessionHREF = *url
	return nil
}

func (cli *VCDClient) vcdauthorize(user, pass, org string) error {
	var missing_items []string
	if user == "" {
		missing_items = append(missing_items, "user")
	}
	if pass == "" {
		missing_items = append(missing_items, "password")
	}
	if org == "" {
		missing_items = append(missing_items, "org")
	}
	if len(missing_items) > 0 {
		return fmt.Errorf("Authorization is not possible because of these missing items: %v", missing_items)
	}
	// No point in checking for errors here
	req := cli.Client.NewRequest(map[string]string{}, "POST", cli.sessionHREF, nil)
	// Set Basic Authentication Header
	req.SetBasicAuth(user+"@"+org, pass)
	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/*+xml;version=5.5")
	resp, err := checkResp(cli.Client.Http.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Store the authentication header
	cli.Client.VCDToken = resp.Header.Get("x-vcloud-authorization")
	cli.Client.VCDAuthHeader = "x-vcloud-authorization"
	// Get query href
	cli.QueryHREF = cli.Client.VCDHREF
	cli.QueryHREF.Path += "/query"
	return nil
}

func NewVCDClient(vcdEndpoint neturl.URL, insecure bool) *VCDClient {

	return &VCDClient{
		Client: Client{
			APIVersion: "5.5",
			VCDHREF:    vcdEndpoint,
			Http: http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: insecure,
					},
					Proxy:               http.ProxyFromEnvironment,
					TLSHandshakeTimeout: 120 * time.Second,
				},
			},
		},
	}
}

// Authenticate is an helper function that performs a login in vCloud Director.
func (cli *VCDClient) Authenticate(username, password, org string) error {
	// LoginUrl
	err := cli.vcdloginurl()
	if err != nil {
		return fmt.Errorf("error finding LoginUrl: %s", err)
	}
	// Authorize
	err = cli.vcdauthorize(username, password, org)
	if err != nil {
		return fmt.Errorf("error authorizing: %s", err)
	}
	return nil
}

// Disconnect performs a disconnection from the vCloud Director API endpoint.
func (cli *VCDClient) Disconnect() error {
	if cli.Client.VCDToken == "" && cli.Client.VCDAuthHeader == "" {
		return fmt.Errorf("cannot disconnect, client is not authenticated")
	}
	req := cli.Client.NewRequest(map[string]string{}, "DELETE", cli.sessionHREF, nil)
	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version=5.5")
	// Set Authorization Header
	req.Header.Add(cli.Client.VCDAuthHeader, cli.Client.VCDToken)
	if _, err := checkResp(cli.Client.Http.Do(req)); err != nil {
		return fmt.Errorf("error processing session delete for vCloud Director: %s", err)
	}
	return nil
}
