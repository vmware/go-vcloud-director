package govcloudair

import (
	"net/url"
	"testing"

	"github.com/vmware/govcloudair/testutil"
	. "gopkg.in/check.v1"
)

type K struct {
	client *VCDClient
	org    Org
}

var vcdu_api, _ = url.Parse("http://localhost:4444/api")
var vcdu_v, _ = url.Parse("http://localhost:4444/api/versions")
var vcdu_s, _ = url.Parse("http://localhost:4444/api/vchs/services")

var _ = Suite(&S{})

var vcdauthheader = map[string]string{"x-vcloud-authorization": "012345678901234567890123456789"}

func (s *K) SetUpSuite(c *C) {
	testServer.Start()
	var err error
	s.client = NewVCDClient(*vcdu_api)
	if err != nil {
		panic(err)
	}

	testServer.ResponseMap(5, testutil.ResponseMap{
		"/api/versions": testutil.Response{200, map[string]string{}, vcdversions},
	})

	s.org, err = s.client.Authenticate("username", "password", "organization")
	if err != nil {
		panic(err)
	}
}

func (s *K) TearDownTest(c *C) {
	testServer.Flush()
}

func TestClient_getloginurl(t *testing.T) {
	testServer.Start()
	var err error

	// Set up a working client
	client := NewVCDClient(*vcdu_api)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Set up a correct conversation
	testServer.ResponseMap(200, testutil.ResponseMap{
		"/api/versions": testutil.Response{200, nil, vcdversions},
	})

	u, err := client.vcdloginurl()
	testServer.Flush()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Test if token is correctly set on client.
	if u.Path != "/api/sessions" {
		t.Fatalf("Getting LoginUrl failed, url: %s", u.Path)
	}

}

func TestVCDClient_Authenticate(t *testing.T) {

	testServer.Start()
	var err error

	client := NewVCDClient(*vcdu_api)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// OK auth
	testServer.ResponseMap(5, testutil.ResponseMap{
		"/api/versions":                                 testutil.Response{200, nil, vcdversions},
		"/api/sessions":                                 testutil.Response{201, vcdauthheader, vcdsessions},
		"/api/org/00000000-0000-0000-0000-000000000000": testutil.Response{201, vcdauthheader, vcdorg},
	})

	org, err := client.Authenticate("username", "password", "organization")
	testServer.Flush()
	if err != nil {
		t.Fatalf("Error authenticating: %v", err)
	}

	if org.Org.FullName != "Organization (full)" {
		t.Fatalf("Orgname not parsed, got: %s", org.Org.FullName)
	}
}

// status: 200
var vcdversions = `
<?xml version="1.0" encoding="UTF-8"?>
<SupportedVersions xmlns="http://www.vmware.com/vcloud/versions" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/versions http://localhost:4444/api/versions/schema/versions.xsd">
    <VersionInfo>
        <Version>1.5</Version>
        <LoginUrl>http://localhost:4444/api/sessions</LoginUrl>
        <MediaTypeMapping>
            <MediaType>application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml</MediaType>
            <ComplexTypeName>InstantiateVAppTemplateParamsType</ComplexTypeName>
            <SchemaLocation>http://localhost:4444/api/v1.5/schema/master.xsd</SchemaLocation>
        </MediaTypeMapping>
        <MediaTypeMapping>
            <MediaType>application/vnd.vmware.admin.vmwProviderVdcReferences+xml</MediaType>
            <ComplexTypeName>VMWProviderVdcReferencesType</ComplexTypeName>
            <SchemaLocation>http://localhost:4444/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
        </MediaTypeMapping>
    </VersionInfo>
</SupportedVersions>
`

var vcdsessions = `
<?xml version="1.0" encoding="UTF-8"?>
<Session xmlns="http://www.vmware.com/vcloud/v1.5" userId="urn:vcloud:user:00000000-0000-0000-0000-000000000000" user="username" org="organization" type="application/vnd.vmware.vcloud.session+xml" href="http://localhost:4444/api/session/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://localhost:4444/vcloud/v1.5 http://localhost:4444/api/v1.5/schema/master.xsd">
    <Link rel="down" type="application/vnd.vmware.vcloud.orgList+xml" href="http://localhost:4444/api/org/"/>
    <Link rel="remove" href="http://localhost:4444/api/session/"/>
    <Link rel="down" type="application/vnd.vmware.vcloud.org+xml" name="username" href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000"/>
    <Link rel="down" type="application/vnd.vmware.vcloud.query.queryList+xml" href="http://localhost:4444/api/query"/>
    <Link rel="entityResolver" type="application/vnd.vmware.vcloud.entity+xml" href="http://localhost:4444/api/entity/"/>
    <Link rel="down:extensibility" type="application/vnd.vmware.vcloud.apiextensibility+xml" href="http://localhost:4444/api/extensibility"/>
</Session>
`

var vcdorg = `
<?xml version="1.0" encoding="UTF-8"?>
<Org xmlns="http://www.vmware.com/vcloud/v1.5" name="organization" id="urn:vcloud:org:00000000-0000-0000-0000-000000000000" type="application/vnd.vmware.vcloud.org+xml" href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://localhost:4444/api/v1.5/schema/master.xsd">
    <Link rel="down" type="application/vnd.vmware.vcloud.vdc+xml" name="organization vDC" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000"/>
    <Link rel="down" type="application/vnd.vmware.vcloud.catalog+xml" name="catalog-a" href="http://localhost:4444/api/catalog/00000000-0000-0000-0000-000000000000"/>
    <Link rel="down" type="application/vnd.vmware.vcloud.catalog+xml" name="catalog-b" href="http://localhost:4444/api/catalog/00000000-0000-0000-0000-000000000001"/>
    <Description/>
    <FullName>Organization (full)</FullName>
</Org>
`
