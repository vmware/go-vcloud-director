package govcd

import (
	"net/url"
	"testing"

	"github.com/vmware/go-vcloud-director/testutil"
	. "gopkg.in/check.v1"
)

// This is the test struct for vcloud-director.
// Test functions use the struct to get
// the test vdc and access to a test client
type TestVCD struct {
	client *VCDClient
	org    Org
	vdc    Vdc
	vapp   VApp
}

var testServer = testutil.NewHTTPServer()

var vcdu_api, _ = url.Parse("http://localhost:4444/api")
var vcdu_v, _ = url.Parse("http://localhost:4444/api/versions")
var vcdu_s, _ = url.Parse("http://localhost:4444/api/vchs/services")

var vcdauthheader = map[string]string{"x-vcloud-authorization": "012345678901234567890123456789"}

var _ = Suite(&TestVCD{})

func Test(t *testing.T) { TestingT(t) }

func (vcd *TestVCD) SetUpSuite(c *C) {
	testServer.Start()
	var err error
	vcd.client = NewVCDClient(*vcdu_api, false)
	if err != nil {
		panic(err)
	}

	testServer.ResponseMap(5, testutil.ResponseMap{
		"/api/versions":                                 testutil.Response{200, map[string]string{}, vcdversions},
		"/api/sessions":                                 testutil.Response{201, vcdauthheader, vcdsessions},
		"/api/org/00000000-0000-0000-0000-000000000000": testutil.Response{201, vcdauthheader, vcdorg},
		"/api/vdc/00000000-0000-0000-0000-000000000000": testutil.Response{201, vcdauthheader, vcdvdc},
	})

	vcd.org, vcd.vdc, err = vcd.client.Authenticate("username", "password", "organization", "organization vDC")
	vcd.vapp = *NewVApp(&vcd.client.Client)
	if err != nil {
		panic(err)
	}
}

func (vcd *TestVCD) TearDownTest(c *C) {
	testServer.Flush()
}

func TestClient_getloginurl(t *testing.T) {
	testServer.Start()
	var err error

	// Set up a working client
	client := NewVCDClient(*vcdu_api, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Set up a correct conversation
	testServer.ResponseMap(200, testutil.ResponseMap{
		"/api/versions": testutil.Response{200, nil, vcdversions},
	})

	err = client.vcdloginurl()
	testServer.Flush()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Test if token is correctly set on client.
	if client.sessionHREF.Path != "/api/sessions" {
		t.Fatalf("Getting LoginUrl failed, url: %s", client.sessionHREF.Path)
	}
}

func TestVCDClient_Authenticate(t *testing.T) {

	testServer.Start()
	var err error

	client := NewVCDClient(*vcdu_api, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// OK auth
	testServer.ResponseMap(5, testutil.ResponseMap{
		"/api/versions":                                 testutil.Response{200, nil, vcdversions},
		"/api/sessions":                                 testutil.Response{201, vcdauthheader, vcdsessions},
		"/api/org/00000000-0000-0000-0000-000000000000": testutil.Response{201, vcdauthheader, vcdorg},
		"/api/vdc/00000000-0000-0000-0000-000000000000": testutil.Response{201, vcdauthheader, vcdorg},
	})

	org, _, err := client.Authenticate("username", "password", "organization", "organization vDC")
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
    <Link href="http://localhost:4444/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854" name="Public Catalog" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link href="http://localhost:4444/api/catalog/5cb6451b-8091-4c89-930d-1ff9653cb12d" name="PSE" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/catalog/5cb6451b-8091-4c89-930d-1ff9653cb12d/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/catalog/5cb6451b-8091-4c89-930d-1ff9653cb12d/action/controlAccess" rel="controlAccess" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link href="http://localhost:4444/api/catalog/8715ed97-348a-4bdd-aea3-e13e84213e6f" name="GCoE" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/catalog/8715ed97-348a-4bdd-aea3-e13e84213e6f/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/catalog/8715ed97-348a-4bdd-aea3-e13e84213e6f/action/controlAccess" rel="controlAccess" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link href="http://localhost:4444/api/catalog/92d4ad5a-217e-4bd1-8bcf-be9cc70dcfa6" name="Vagrant" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/catalog/92d4ad5a-217e-4bd1-8bcf-be9cc70dcfa6/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/catalog/92d4ad5a-217e-4bd1-8bcf-be9cc70dcfa6/action/controlAccess" rel="controlAccess" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link href="http://localhost:4444/api/admin/org/00000000-0000-0000-0000-000000000000/catalogs" rel="add" type="application/vnd.vmware.admin.catalog+xml"/>
    <Link href="http://localhost:4444/api/network/8d0cbfe2-25b3-4a1f-b608-5ffeabc7a53d" name="M916272752-5793-default-isolated" rel="down" type="application/vnd.vmware.vcloud.orgNetwork+xml"/>
    <Link href="http://localhost:4444/api/network/cb0f4c9e-1a46-49d4-9fcb-d228000a6bc1" name="networkName" rel="down" type="application/vnd.vmware.vcloud.orgNetwork+xml"/>
    <Link href="http://localhost:4444/api/supportedSystemsInfo/" rel="down" type="application/vnd.vmware.vcloud.supportedSystemsInfo+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/metadata" rel="down" type="application/vnd.vmware.vcloud.metadata+xml"/>
    <Link href="http://localhost:4444/api/org/00000000-0000-0000-0000-000000000000/hybrid" rel="down" type="application/vnd.vmware.vcloud.hybridOrg+xml"/>
    <Description/>
    <FullName>Organization (full)</FullName>
</Org>
`

var vcdvdc = `
<?xml version="1.0" encoding="UTF-8"?>
<Vdc xmlns="http://www.vmware.com/vcloud/v1.5" status="1" name="organization vDC" id="urn:vcloud:vdc:00000000-0000-0000-0000-000000000001" type="application/vnd.vmware.vcloud.vdc+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://localhost:4444/api/v1.5/schema/master.xsd">
    <Link rel="down" type="application/vnd.vmware.vcloud.metadata+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/metadata"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/action/uploadVAppTemplate"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.media+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/media"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/action/instantiateVAppTemplate"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.cloneVAppParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/action/cloneVApp"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.cloneVAppTemplateParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/action/cloneVAppTemplate"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.cloneMediaParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/action/cloneMedia"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.captureVAppParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/action/captureVApp"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.composeVAppParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/action/composeVApp"/>
    <Link rel="add" type="application/vnd.vmware.vcloud.diskCreateParams+xml" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000001/disk"/>
    <Description>organization vDC</Description>
    <AllocationModel>AllocationVApp</AllocationModel>
    <ComputeCapacity>
        <Cpu>
            <Units>MHz</Units>
            <Allocated>0</Allocated>
            <Limit>0</Limit>
            <Reserved>0</Reserved>
            <Used>0</Used>
            <Overhead>0</Overhead>
        </Cpu>
        <Memory>
            <Units>MB</Units>
            <Allocated>0</Allocated>
            <Limit>0</Limit>
            <Reserved>0</Reserved>
            <Used>0</Used>
            <Overhead>0</Overhead>
        </Memory>
    </ComputeCapacity>
    <ResourceEntities>
        <ResourceEntity type="application/vnd.vmware.vcloud.vAppTemplate+xml" name="vApp_CentOS-6.6_1" href="http://localhost:4444/api/vAppTemplate/vappTemplate-00000000-0000-0000-0000-000000000001"/>
        <ResourceEntity type="application/vnd.vmware.vcloud.media+xml" name="CentOS-7.0-1406-x86_64-DVD.iso" href="http://localhost:4444/api/media/00000000-0000-0000-0000-000000000001"/>
        <ResourceEntity type="application/vnd.vmware.vcloud.media+xml" name="CentOS-6.6-x86_64-minimal.iso" href="http://localhost:4444/api/media/00000000-0000-0000-0000-000000000001"/>
    </ResourceEntities>
    <AvailableNetworks>
        <Network type="application/vnd.vmware.vcloud.network+xml" name="Internal Network" href="http://localhost:4444/api/network/00000000-0000-0000-0000-000000000001"/>
    </AvailableNetworks>
    <Capabilities>
        <SupportedHardwareVersions>
            <SupportedHardwareVersion>vmx-04</SupportedHardwareVersion>
            <SupportedHardwareVersion>vmx-07</SupportedHardwareVersion>
            <SupportedHardwareVersion>vmx-08</SupportedHardwareVersion>
            <SupportedHardwareVersion>vmx-09</SupportedHardwareVersion>
            <SupportedHardwareVersion>vmx-10</SupportedHardwareVersion>
        </SupportedHardwareVersions>
    </Capabilities>
    <NicQuota>0</NicQuota>
    <NetworkQuota>100</NetworkQuota>
    <UsedNetworkCount>0</UsedNetworkCount>
    <VmQuota>0</VmQuota>
    <IsEnabled>true</IsEnabled>
    <VdcStorageProfiles>
        <VdcStorageProfile type="application/vnd.vmware.vcloud.vdcStorageProfile+xml" name="Gold-Datastore" href="http://localhost:4444/api/vdcStorageProfile/00000000-0000-0000-0000-000000000001"/>
        <VdcStorageProfile type="application/vnd.vmware.vcloud.vdcStorageProfile+xml" name="Bronze-Datastore" href="http://localhost:4444/api/vdcStorageProfile/00000000-0000-0000-0000-000000000001"/>
        <VdcStorageProfile type="application/vnd.vmware.vcloud.vdcStorageProfile+xml" name="Silver-Datastore" href="http://localhost:4444/api/vdcStorageProfile/00000000-0000-0000-0000-000000000001"/>
    </VdcStorageProfiles>
</Vdc>
`
