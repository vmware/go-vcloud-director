/*
* @Author: frapposelli
* @Date:   2014-10-13 11:10:56
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:10:53
 */

package govcloudair

import (
	"net/http"
	"os"
	"testing"

	"github.com/frapposelli/govcloud/testutil"
	. "gopkg.in/check.v1"
)

type S struct {
	client *Client
}

var _ = Suite(&S{})

var testServer = testutil.NewHTTPServer()

func (s *S) SetUpSuite(c *C) {
	testServer.Start()
	var err error
	s.client = &Client{Http: http.DefaultClient}
	authheader := map[string]string{"x-vchs-authorization": "012345678901234567890123456789"}
	testServer.Response(201, authheader, vaauthorization)
	err = s.client.vaauthorize("http://localhost:4444/api", "username", "password")
	testServer.Response(200, nil, vaservices)
	err = s.client.vaacquireservice("http://localhost:4444/api", "CI123456-789")
	testServer.Response(200, nil, vacompute)
	err = s.client.vaacquirecompute("VDC12345-6789")
	testServer.Response(201, nil, vabackend)
	err = s.client.vagetbackendauth("VDC12345-6789")
	if err != nil {
		panic(err)
	}
}

func (s *S) TearDownTest(c *C) {
	testServer.Flush()
}

func TestClient_vaauthorize(t *testing.T) {
	testServer.Start()
	var err error
	var client = Client{Http: http.DefaultClient}
	authheader := map[string]string{"x-vchs-authorization": "012345678901234567890123456789"}
	testServer.Response(201, authheader, vaauthorization)
	err = client.vaauthorize("http://localhost:4444/api", "username", "password")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.VAToken != "012345678901234567890123456789" {
		t.Fatalf("VAtoken not set on client: %s", client.VAToken)
	}

}

func TestClient_vaacquireservice(t *testing.T) {
	testServer.Start()
	var err error
	var client = Client{
		VAToken: "012345678901234567890123456789",
		Http:    http.DefaultClient,
	}
	testServer.Response(200, nil, vaservices)
	err = client.vaacquireservice("http://localhost:4444/api", "CI123456-789")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.ComputeHREF != "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000" {
		t.Fatalf("ComputeHREF not set on client: %s", client.ComputeHREF)
	}

	if client.Region != "US - Anywhere" {
		t.Fatalf("Region not set on client: %s", client.Region)
	}

}

func TestClient_vaacquirecompute(t *testing.T) {
	testServer.Start()
	var err error
	var client = Client{
		VAToken:     "012345678901234567890123456789",
		Region:      "US - Anywhere",
		ComputeHREF: "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000",
		Http:        http.DefaultClient,
	}
	testServer.Response(200, nil, vacompute)
	err = client.vaacquirecompute("VDC12345-6789")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.VDCHREF != "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" {
		t.Fatalf("VDCHREF not set on client: %s", client.VDCHREF)
	}

}

func TestClient_vagetbackendauth(t *testing.T) {
	testServer.Start()
	var err error
	var client = Client{
		VAToken:     "012345678901234567890123456789",
		Region:      "US - Anywhere",
		ComputeHREF: "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000",
		VDCHREF:     "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession",
		Http:        http.DefaultClient,
	}
	testServer.Response(201, nil, vabackend)
	err = client.vagetbackendauth("VDC12345-6789")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.VCDToken != "01234567890123456789012345678901" {
		t.Fatalf("VCDToken not set on client: %s", client.VCDToken)
	}
	if client.VCDAuthHeader != "x-vcloud-authorization" {
		t.Fatalf("VCDAuthHeader not set on client: %s", client.VCDAuthHeader)
	}
	if client.URL != "http://localhost:4444/api" {
		t.Fatalf("URL not set on client: %s", client.URL)
	}
	if client.VDC != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("VDC not set on client: %s", client.VDC)
	}

}

// Env variable tests

func TestClient_vaauthorize_env(t *testing.T) {

	os.Setenv("VCLOUDAIR_ENDPOINT", "http://localhost:4444/api")
	os.Setenv("VCLOUDAIR_USERNAME", "username")
	os.Setenv("VCLOUDAIR_PASSWORD", "password")

	testServer.Start()
	var err error
	var client = Client{Http: http.DefaultClient}
	authheader := map[string]string{"x-vchs-authorization": "012345678901234567890123456789"}
	testServer.Response(201, authheader, vaauthorization)
	err = client.vaauthorize("", "", "")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.VAToken != "012345678901234567890123456789" {
		t.Fatalf("VAtoken not set on client: %s", client.VAToken)
	}

}

func TestClient_vaacquireservice_env(t *testing.T) {

	os.Setenv("VCLOUDAIR_ENDPOINT", "http://localhost:4444/api")
	os.Setenv("VCLOUDAIR_COMPUTEID", "CI123456-789")

	testServer.Start()
	var err error
	var client = Client{
		VAToken: "012345678901234567890123456789",
		Http:    http.DefaultClient,
	}
	testServer.Response(200, nil, vaservices)
	err = client.vaacquireservice("", "")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.ComputeHREF != "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000" {
		t.Fatalf("ComputeHREF not set on client: %s", client.ComputeHREF)
	}

	if client.Region != "US - Anywhere" {
		t.Fatalf("Region not set on client: %s", client.Region)
	}

}

func TestClient_vaacquirecompute_env(t *testing.T) {

	os.Setenv("VCLOUDAIR_VDCID", "VDC12345-6789")

	testServer.Start()
	var err error
	var client = Client{
		VAToken:     "012345678901234567890123456789",
		Region:      "US - Anywhere",
		ComputeHREF: "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000",
		Http:        http.DefaultClient,
	}
	testServer.Response(200, nil, vacompute)
	err = client.vaacquirecompute("")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.VDCHREF != "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" {
		t.Fatalf("VDCHREF not set on client: %s", client.VDCHREF)
	}

}

func TestClient_vagetbackendauth_env(t *testing.T) {

	os.Setenv("VCLOUDAIR_VDCID", "VDC12345-6789")

	testServer.Start()
	var err error
	var client = Client{
		VAToken:     "012345678901234567890123456789",
		Region:      "US - Anywhere",
		ComputeHREF: "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000",
		VDCHREF:     "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession",
		Http:        http.DefaultClient,
	}
	testServer.Response(201, nil, vabackend)
	err = client.vagetbackendauth("")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.VCDToken != "01234567890123456789012345678901" {
		t.Fatalf("VCDToken not set on client: %s", client.VCDToken)
	}
	if client.VCDAuthHeader != "x-vcloud-authorization" {
		t.Fatalf("VCDAuthHeader not set on client: %s", client.VCDAuthHeader)
	}
	if client.URL != "http://localhost:4444/api" {
		t.Fatalf("URL not set on client: %s", client.URL)
	}
	if client.VDC != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("VDC not set on client: %s", client.VDC)
	}

}

func makeClient(t *testing.T) Client {

	testServer.Start()
	var err error
	var client = Client{Http: http.DefaultClient}
	authheader := map[string]string{"x-vchs-authorization": "012345678901234567890123456789"}
	testServer.Response(201, authheader, vaauthorization)
	err = client.vaauthorize("http://localhost:4444/api", "username", "password")
	testServer.Response(200, nil, vaservices)
	err = client.vaacquireservice("http://localhost:4444/api", "CI123456-789")
	testServer.Response(200, nil, vacompute)
	err = client.vaacquirecompute("VDC12345-6789")
	testServer.Response(201, nil, vabackend)
	err = client.vagetbackendauth("VDC12345-6789")

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.VAToken != "012345678901234567890123456789" {
		t.Fatalf("VAtoken not set on client: %s", client.VAToken)
	}

	if client.ComputeHREF != "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000" {
		t.Fatalf("ComputeHREF not set on client: %s", client.ComputeHREF)
	}

	if client.Region != "US - Anywhere" {
		t.Fatalf("Region not set on client: %s", client.Region)
	}

	if client.VDCHREF != "http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" {
		t.Fatalf("VDCHREF not set on client: %s", client.VDCHREF)
	}

	if client.VCDToken != "01234567890123456789012345678901" {
		t.Fatalf("VCDToken not set on client: %s", client.VCDToken)
	}
	if client.VCDAuthHeader != "x-vcloud-authorization" {
		t.Fatalf("VCDAuthHeader not set on client: %s", client.VCDAuthHeader)
	}
	if client.URL != "http://localhost:4444/api" {
		t.Fatalf("URL not set on client: %s", client.URL)
	}
	if client.VDC != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("VDC not set on client: %s", client.VDC)
	}

	return client
}

func TestClient_NewRequest(t *testing.T) {
	c := makeClient(t)

	params := map[string]string{
		"foo": "bar",
		"baz": "bar",
	}
	req, err := c.NewRequest(params, "POST", "/bar", "application/xml")
	if err != nil {
		t.Fatalf("bad: %v", err)
	}

	encoded := req.URL.Query()
	if encoded.Get("foo") != "bar" {
		t.Fatalf("bad: %v", encoded)
	}

	if encoded.Get("baz") != "bar" {
		t.Fatalf("bad: %v", encoded)
	}

	if req.URL.String() != "http://localhost:4444/api/bar?baz=bar&foo=bar" {
		t.Fatalf("bad base url: %v", req.URL.String())
	}

	if req.Header.Get("x-vcloud-authorization") != "01234567890123456789012345678901" {
		t.Fatalf("bad auth header: %v", req.Header)
	}

	if req.Header.Get("Content-Type") != "application/xml" {
		t.Fatalf("bad content-type header: %v", req.Header)
	}

	if req.Method != "POST" {
		t.Fatalf("bad method: %v", req.Method)
	}
}

// status: 201
var vaauthorization = `
	<?xml version="1.0" ?>
	<Session href="http://localhost:4444/api/vchs/session" type="application/xml;class=vnd.vmware.vchs.session" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	    <Link href="http://localhost:4444/api/vchs/services" rel="down" type="application/xml;class=vnd.vmware.vchs.servicelist"/>
	    <Link href="http://localhost:4444/api/vchs/session" rel="remove"/>
	</Session>
	`

// status: 200
var vaservices = `
	<?xml version="1.0" ?>
	<Services href="http://localhost:4444/api/vchs/services" type="application/xml;class=vnd.vmware.vchs.servicelist" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	    <Service href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000" region="US - Anywhere" serviceId="CI123456-789" serviceType="compute:vpc" type="application/xml;class=vnd.vmware.vchs.compute"/>
	</Services>
	`

// status: 200
var vacompute = `
	<?xml version="1.0" ?>
	<Compute href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000" serviceId="CI123456-789" serviceType="compute:vpc" type="application/xml;class=vnd.vmware.vchs.compute" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	    <Link href="http://localhost:4444/api/vchs/services" name="services" rel="up" type="application/xml;class=vnd.vmware.vchs.servicelist"/>
	    <VdcRef href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000" name="VDC12345-6789" status="Active" type="application/xml;class=vnd.vmware.vchs.vdcref">
	        <Link href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" name="VDC12345-6789" rel="down" type="application/xml;class=vnd.vmware.vchs.vcloudsession"/>
	    </VdcRef>
	</Compute>
	`

// status: 201
var vabackend = `
<?xml version="1.0" ?>
<VCloudSession href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" name="VDC12345-6789-session" type="application/xml;class=vnd.vmware.vchs.vcloudsession" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <Link href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000" name="vdc" rel="up" type="application/xml;class=vnd.vmware.vchs.vdcref"/>
    <VdcLink authorizationHeader="x-vcloud-authorization" authorizationToken="01234567890123456789012345678901" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000" name="VDC12345-6789" rel="vcloudvdc" type="application/vnd.vmware.vcloud.vdc+xml"/>
</VCloudSession>
	`
