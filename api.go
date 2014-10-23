/*
* @Author: frapposelli
* @Date:   2014-10-10 17:29:37
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:10:59
 */

package govcloudair

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Client provides a client to vCloud Director API with vCloud Air Extensions
type Client struct {
	VAToken       string       // vCloud Air authorization token
	Region        string       // Region where the compute resource lives.
	ComputeHREF   string       // Compute HREF link.
	VDCHREF       string       // VDC HREF Link.
	VCDToken      string       // Access Token (authorization header)
	VCDAuthHeader string       // Authorization header
	URL           string       // URL to the backend vCloud to use
	VDC           string       // Name of the VDC you're using (also, Organization Name as Org == VDC in vCloud Air)
	Http          *http.Client // HttpClient is the client to use. Default will be used if not provided.
}

type Service struct {
	Region      string `xml:"region,attr"`
	ServiceID   string `xml:"serviceId,attr"`
	ServiceType string `xml:"serviceType,attr"`
	Type        string `xml:"type,attr"`
	HREF        string `xml:"href,attr"`
}

type Services struct {
	Service []Service
}

type VdcLinks struct {
	Rel  string `xml:"rel,attr"`
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	HREF string `xml:"href,attr"`
}

type VdcRef struct {
	Status string `xml:"status,attr"`
	Name   string `xml:"name,attr"`
	Type   string `xml:"type,attr"`
	HREF   string `xml:"href,attr"`
	Link   []VdcLinks
}

type ComputeResources struct {
	VdcRef []VdcRef
}

type VdcLink struct {
	AuthorizationToken  string `xml:"authorizationToken,attr"`
	AuthorizationHeader string `xml:"authorizationHeader,attr"`
	Name                string `xml:"name,attr"`
	HREF                string `xml:"href,attr"`
}

type VCloudSession struct {
	VdcLink []VdcLink
}

// VCDError is the error format that vCloud Director returns in case of an error
type VCDError struct {
	Message                 string `xml:"message,attr"`
	MajorErrorCode          int    `xml:"majorErrorCode,attr"`
	MinorErrorCode          string `xml:"minorErrorCode,attr"`
	VendorSpecificErrorCode string `xml:"vendorSpecificErrorCode,attr"`
	StackTrace              string `xml:"stackTrace,attr"`
}

func (c *Client) vaauthorize(vaendpoint, username, password string) error {

	if vaendpoint == "" {
		vaendpoint = os.Getenv("VCLOUDAIR_ENDPOINT")
	}

	if username == "" {
		username = os.Getenv("VCLOUDAIR_USERNAME")
	}

	if password == "" {
		password = os.Getenv("VCLOUDAIR_PASSWORD")
	}

	u, err := url.Parse(vaendpoint + "/vchs/sessions")
	if err != nil {
		return err
	}

	// Build the POST request for authentication
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return err
	}

	// Add the Accept header
	req.Header.Add("Accept", "application/xml;version=5.6")
	// Set Basic Authentication Header
	req.SetBasicAuth(username, password)

	client := &http.Client{}

	// TODO: wrap into checkresp to parse error.
	resp, err := checkResp(client.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Store the authentication header
	c.VAToken = resp.Header.Get("X-Vchs-Authorization")
	return nil

}

func (c *Client) vaacquireservice(vaendpoint, computeid string) error {

	if vaendpoint == "" {
		vaendpoint = os.Getenv("VCLOUDAIR_ENDPOINT")
	}

	if computeid == "" {
		computeid = os.Getenv("VCLOUDAIR_COMPUTEID")
	}

	u, err := url.Parse(vaendpoint + "/vchs/services")
	if err != nil {
		return fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Build the GET request
	req, err := http.NewRequest("GET", u.String(), nil)

	if err != nil {
		return fmt.Errorf("Error creating request: %s", err)
	}

	// Add the Accept header
	req.Header.Add("Accept", "application/xml;version=5.6")
	// Set Authorization Header
	req.Header.Add("x-vchs-authorization", c.VAToken)

	client := &http.Client{}

	// TODO: wrap into checkresp to parse error.
	resp, err := checkResp(client.Do(req))
	if err != nil {
		return fmt.Errorf("Error processing compute action: %s", err)
	}

	// Read response Body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error parsing response Body: %s", err)
	}
	defer resp.Body.Close()

	// Initialize a new Services struct
	services := Services{}

	// Unmarshal XML response in the Services struct
	err = xml.Unmarshal([]byte(data), &services)
	if err != nil {
		return fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Loop in the Services struct to find right service and compute resource.
	for _, s := range services.Service {
		if s.ServiceID == computeid {
			c.ComputeHREF = s.HREF
			c.Region = s.Region
		}
	}

	// If the right compute resource cannot be found, exit gracefully
	if c.ComputeHREF == "" {
		return fmt.Errorf("Error finding the right compute resource")
	}

	return nil

}

func (c *Client) vaacquirecompute(vdcid string) error {

	if vdcid == "" {
		vdcid = os.Getenv("VCLOUDAIR_VDCID")
	}

	u, err := url.Parse(c.ComputeHREF)
	if err != nil {
		return fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Build the GET request for backend server identification
	req, err := http.NewRequest("GET", u.String(), nil)

	if err != nil {
		return fmt.Errorf("Error creating request: %s", err)
	}

	// Add the Accept header
	req.Header.Add("Accept", "application/xml;version=5.6")
	// Set Authorization Header
	req.Header.Add("x-vchs-authorization", c.VAToken)

	client := &http.Client{}

	// TODO: wrap into checkresp to parse error
	resp, err := checkResp(client.Do(req))
	if err != nil {
		return fmt.Errorf("Error processing compute action: %s", err)
	}

	// Read response Body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error parsing response Body: %s", err)
	}
	defer resp.Body.Close()

	// Initialize a new ComputeResources struct
	computeresources := ComputeResources{}

	// Unmarshal XML response in the ComputeResources struct
	err = xml.Unmarshal([]byte(data), &computeresources)
	if err != nil {
		return fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Iterate through the ComputeResources struct searching for the right
	// backend server
	for _, s := range computeresources.VdcRef {
		if s.Name == vdcid {
			for _, t := range s.Link {
				if t.Name == vdcid {
					c.VDCHREF = t.HREF
				}
			}
		}
	}

	// If the right backend resource cannot be found, exit gracefully
	if c.VDCHREF == "" {
		return fmt.Errorf("Error finding the right backend resource")
	}

	return nil
}

func (c *Client) vagetbackendauth(vdcid string) error {

	if vdcid == "" {
		vdcid = os.Getenv("VCLOUDAIR_VDCID")
	}

	u, err := url.Parse(c.VDCHREF)
	if err != nil {
		return fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Build the POST request for authentication
	req, err := http.NewRequest("POST", u.String(), nil)

	if err != nil {
		return fmt.Errorf("Error creating request: %s", err)
	}

	// Add the Accept header
	req.Header.Add("Accept", "application/xml;version=5.6")
	// Set Authorization Header
	req.Header.Add("x-vchs-authorization", c.VAToken)

	client := &http.Client{}

	// TODO: wrap into checkresp to parse error
	resp, err := checkResp(client.Do(req))
	if err != nil {
		return fmt.Errorf("Error processing backend url action: %s", err)
	}
	defer resp.Body.Close()

	// Read response Body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error parsing response Body: %s", err)
	}

	// Initialize a new VCloudSession struct
	vcloudsession := VCloudSession{}

	// Unmarshal XML response in the VCloudSession struct
	err = xml.Unmarshal([]byte(data), &vcloudsession)
	if err != nil {
		return fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Get the backend session information
	for _, s := range vcloudsession.VdcLink {
		if s.Name == vdcid {
			// Fetch the authorization token
			c.VCDToken = s.AuthorizationToken

			// Fetch the authorization header
			c.VCDAuthHeader = s.AuthorizationHeader

			// Build the backend url
			u, _ := url.Parse(s.HREF)
			c.URL = u.Scheme + "://" + u.Host + "/api"

			// Get the vdcguid from the url path
			urlPath := strings.SplitAfter(u.Path, "/")
			c.VDC = urlPath[len(urlPath)-1]
		}
	}

	if c.VCDToken == "" || c.VCDAuthHeader == "" || c.URL == "" || c.VDC == "" {
		return fmt.Errorf("Error finding the right backend resource")
	}

	return nil
}

// Helper function that performs a complete login in vCloud Air and in the backend vCloud Director instance and returns a complete new client.
func NewClient(vaendpoint, username, password, computeid, vdcid string) (*Client, error) {

	Client := Client{Http: http.DefaultClient}

	// Authorize
	err := Client.vaauthorize(vaendpoint, username, password)
	if err != nil {
		return nil, fmt.Errorf("Error Authorizing: %s", err)
	}

	// Get Service
	err = Client.vaacquireservice(vaendpoint, computeid)
	if err != nil {
		return nil, fmt.Errorf("Error Acquiring Service: %s", err)
	}

	// Get Compute
	err = Client.vaacquirecompute(vdcid)
	if err != nil {
		return nil, fmt.Errorf("Error Acquiring Compute: %s", err)
	}

	// Get Backend Authorization
	err = Client.vagetbackendauth(vdcid)
	if err != nil {
		return nil, fmt.Errorf("Error Acquiring Backend Authorization: %s", err)
	}

	return &Client, nil

}

// Creates a new request with the params
func (c *Client) NewRequest(params map[string]string, method string, endpoint string, contenttype string) (*http.Request, error) {
	p := url.Values{}
	u, err := url.Parse(c.URL + endpoint)

	if err != nil {
		return nil, fmt.Errorf("Error parsing base URL: %s", err)
	}

	// Build up our request parameters
	for k, v := range params {
		p.Add(k, v)
	}

	// Add the params to our URL
	u.RawQuery = p.Encode()

	// Build the request
	req, err := http.NewRequest(method, u.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("Error creating request: %s", err)
	}

	// Add the authorization header
	req.Header.Add(c.VCDAuthHeader, c.VCDToken)
	// Add the Accept header
	req.Header.Add("Accept", "application/*+xml;version=5.6")
	// If not empty, add Content-Type header
	if contenttype != "" {
		req.Header.Add("Content-Type", contenttype)
	}

	return req, nil

}

// parseErr is used to take an error json resp
// and return a single string for use in error messages
func parseErr(resp *http.Response) error {
	errBody := VCDError{}

	err := decodeBody(resp, &errBody)

	// if there was an error decoding the body, just return that
	if err != nil {
		return fmt.Errorf("Error parsing error body for non-200 request: %s", err)
	}

	return fmt.Errorf("API Error: %d: %s", errBody.MajorErrorCode, errBody.Message)
}

// decodeBody is used to XML decode a body
func decodeBody(resp *http.Response, out interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err = xml.Unmarshal(body, &out); err != nil {
		return err
	}

	return nil
}

// checkResp wraps http.Client.Do() and verifies that the
// request was successful. A non-200 request returns an error
// formatted to included any validation problems or otherwise
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	// If the err is already there, there was an error higher
	// up the chain, so just return that
	if err != nil {
		return resp, err
	}

	switch i := resp.StatusCode; {
	case i == 200:
		return resp, nil
	case i == 201:
		return resp, nil
	case i == 202:
		return resp, nil
	case i == 204:
		return resp, nil
	case i == 404:
		return nil, fmt.Errorf("API Error: %s", resp.Status)
	default:
		return nil, parseErr(resp)
	}
}
