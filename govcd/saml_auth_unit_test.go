// +build unit ALL

package govcd

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"
)

// testVcdMockAuthToken is the expected vcdCli.Client.VCDToken value after `Authentication()`
// function passes mock SAML authentication process
const testVcdMockAuthToken = "e3b02b30b8ff4e87ac38db785b0172b5"

// TestSamlAdfsAuthenticate is a unit test using mock vCD and ADFS server endpoint to follow
// complete SAML auth flow. The `testVcdMockAuthToken` is expected as an outcome token because
// mock servers return static responses.
func TestSamlAdfsAuthenticate(t *testing.T) {
	// Spawn mock ADFS server
	adfsServer := testSpawnAdfsServer()
	adfsServerHost := adfsServer.URL
	defer adfsServer.Close()

	// Spawn mock vCD instance just enough to cover login details
	vcdServer := spawnVcdServer(adfsServerHost, "my-org")
	vcdServerHost := vcdServer.URL
	defer vcdServer.Close()

	// Setup vCD client pointing to mock API
	vcdUrl, err := url.Parse(vcdServerHost + "/api")
	if err != nil {
		t.Errorf("got errors: %s", err)
	}
	vcdCli := NewVCDClient(*vcdUrl, true, WithSamlAdfs(true, ""))
	err = vcdCli.Authenticate("fakeUser", "fakePass", "my-org")
	if err != nil {
		t.Errorf("got errors: %s", err)
	}

	// After authentication
	if vcdCli.Client.VCDToken != testVcdMockAuthToken {
		t.Errorf("received token does not match specified one")
	}
}

// spawnVcdServer establishes a mock vCD server with endpoints required to satisfy authentication
func spawnVcdServer(adfsServerHost, org string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/cloud/org/"+org+"/saml/metadata/alias/vcd", vCDSamlMetadataHandler)
	mux.HandleFunc("/login/"+org+"/saml/login/alias/vcd", getVcdAdfsRedirectHandler(adfsServerHost))
	mux.HandleFunc("/api/sessions", vCDLoginHandler)
	mux.HandleFunc("/api/versions", vCDApiVersionHandler)
	mux.HandleFunc("/api/org", vCDApiOrgHandler)

	server := httptest.NewTLSServer(mux)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("vCD mock server now listening on %s...\n", server.URL)
	}
	return server

}

// vcdLoginHandler serves mock "/api/sessions"
func vCDLoginHandler(w http.ResponseWriter, r *http.Request) {
	// We expect POST method and not anything else
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	exectedHeader := `SIGN token="H4sIAAAAAAAA/5xTTY+bMBD9K4jeKiV20l20sRxLEdnDaru5bFX16jUTsIRtZJsS+usrPhdCD6jcPPZ7b95jhj5rYevCQ3JyDqyXRgc3lWt3DEurieFOOqK5Ake8IO+nt+9kv8WED49DRm+gBRlpztzz4EddwDHMvC8IQlVVbatvW2NTtMd4h/ADuqkctPjynIMC7cNOkTREK1B3ktLoN/CZSYJTnhorfaZWkHBw+8doIz5EiBh9hfpFX83g/J9ojPChQSdOpk0P8On5FerewhrXU+z/NW8d3xgOxUal110RMnqWKTi/kmlmxGV81ySAli2NsTCaOPLrER/af9sZTdbG1ENfnCvBvoOVPL8rXrgCFl+Ov+PclElwlhaENzYwNuVa/uHtTMbNuF2l4B4oWsIHxk7gUqoPsGyHowhHD0+HpxEyu78jGppDE7eMojEEILEsMrBdfTj95HkJ7Gv7NSlOq5/nnms+M1PydqKnbyeFucTiYlbqhZY7ySharjr7GwAA///dRZE6/wMAAA==",org="my-org"`
	if r.Header.Get("Authorization") != exectedHeader {
		w.WriteHeader(500)
		return
	}

	headers := w.Header()
	headers.Add("X-Vcloud-Authorization", testVcdMockAuthToken)

	resp := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<Session xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" xmlns:common="http://schemas.dmtf.org/wbem/wscim/1/common" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vmw="http://www.vmware.com/schema/ovf" xmlns:ovfenv="http://schemas.dmtf.org/ovf/environment/1" xmlns:vmext="http://www.vmware.com/vcloud/extension/v1.5" xmlns:ns9="http://www.vmware.com/vcloud/versions" locationId="c196c6f0-5c31-4929-a626-b29b2c9ff5ab@cb33a646-6652-4628-95b0-24bd981783b6" org="my-org" roles="Organization Administrator" user="test@test-forest.net" userId="urn:vcloud:user:8fd38079-b31c-4dfd-99f2-7073b3f7ec90" href="https://192.168.1.109/api/session" type="application/vnd.vmware.vcloud.session+xml">
		<Link rel="down" href="https://192.168.1.109/api/org/" type="application/vnd.vmware.vcloud.orgList+xml"/>
		<Link rel="remove" href="https://192.168.1.109/api/session"/>
		<Link rel="down" href="https://192.168.1.109/api/admin/" type="application/vnd.vmware.admin.vcloud+xml"/>
		<Link rel="down" href="https://192.168.1.109/api/org/c196c6f0-5c31-4929-a626-b29b2c9ff5ab" name="my-org" type="application/vnd.vmware.vcloud.org+xml"/>
		<Link rel="down" href="https://192.168.1.109/api/query" type="application/vnd.vmware.vcloud.query.queryList+xml"/>
		<Link rel="entityResolver" href="https://192.168.1.109/api/entity/" type="application/vnd.vmware.vcloud.entity+xml"/>
		<Link rel="down:extensibility" href="https://192.168.1.109/api/extensibility" type="application/vnd.vmware.vcloud.apiextensibility+xml"/>
		<Link rel="nsx" href="https://192.168.1.109/network" type="application/xml"/>
		<Link rel="openapi" href="https://192.168.1.109/cloudapi" type="application/json"/>
		<AuthorizedLocations>
			<Location>
				<LocationId>c196c6f0-5c31-4929-a626-b29b2c9ff5ab@cb33a646-6652-4628-95b0-24bd981783b6</LocationId>
				<SiteName>192.168.1.109</SiteName>
				<OrgName>my-org</OrgName>
				<RestApiEndpoint>https://192.168.1.109</RestApiEndpoint>
				<UIEndpoint>https://192.168.1.109</UIEndpoint>
				<AuthContext>my-org</AuthContext>
			</Location>
		</AuthorizedLocations>
	</Session>`)

	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}

// vCDApiVersionHandler server mock "/api/versions"
func vCDApiVersionHandler(w http.ResponseWriter, r *http.Request) {
	// We expect GET method and not anything else
	if r.Method != http.MethodGet {
		w.WriteHeader(500)
		return
	}

	resp := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<SupportedVersions xmlns="http://www.vmware.com/vcloud/versions" xmlns:ns2="http://www.vmware.com/vcloud/v1.5" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" xmlns:common="http://schemas.dmtf.org/wbem/wscim/1/common" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vmw="http://www.vmware.com/schema/ovf" xmlns:ovfenv="http://schemas.dmtf.org/ovf/environment/1" xmlns:vmext="http://www.vmware.com/vcloud/extension/v1.5">
		<VersionInfo deprecated="true">
			<Version>20.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="true">
			<Version>21.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="true">
			<Version>22.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="true">
			<Version>23.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="true">
			<Version>24.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="true">
			<Version>25.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="true">
			<Version>26.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="false">
			<Version>27.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="false">
			<Version>28.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="false">
			<Version>29.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="false">
			<Version>30.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="false">
			<Version>31.0</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
		</VersionInfo>
		<VersionInfo deprecated="true">
			<Version>5.5</Version>
			<LoginUrl>https://192.168.1.109/api/sessions</LoginUrl>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.error+xml</MediaType>
				<ComplexTypeName>ErrorType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/common.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.controlAccess+xml</MediaType>
				<ComplexTypeName>ControlAccessParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/common.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.owner+xml</MediaType>
				<ComplexTypeName>OwnerType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/common.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.query.references+xml</MediaType>
				<ComplexTypeName>ReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/common.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.fileUploadParams+xml</MediaType>
				<ComplexTypeName>FileUploadParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/common.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.fileUploadSocket+xml</MediaType>
				<ComplexTypeName>FileUploadSocketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/common.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.apiextensibility+xml</MediaType>
				<ComplexTypeName>ApiExtensibilityType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.service+xml</MediaType>
				<ComplexTypeName>ServiceType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.apidefinition+xml</MediaType>
				<ComplexTypeName>ApiDefinitionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.filedescriptor+xml</MediaType>
				<ComplexTypeName>FileDescriptorType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.media+xml</MediaType>
				<ComplexTypeName>MediaType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/media.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.cloneMediaParams+xml</MediaType>
				<ComplexTypeName>CloneMediaParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/media.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vms+xml</MediaType>
				<ComplexTypeName>VmsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vms.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.supportedSystemsInfo+xml</MediaType>
				<ComplexTypeName>SupportedOperatingSystemsInfoType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vms.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.catalog+xml</MediaType>
				<ComplexTypeName>CatalogType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/catalog.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.publishCatalogParams+xml</MediaType>
				<ComplexTypeName>PublishCatalogParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/catalog.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.task+xml</MediaType>
				<ComplexTypeName>TaskType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/task.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.taskOperationList+xml</MediaType>
				<ComplexTypeName>TaskOperationListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/task.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vAppTemplate+xml</MediaType>
				<ComplexTypeName>VAppTemplateType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vAppTemplate.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml</MediaType>
				<ComplexTypeName>UploadVAppTemplateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vAppTemplate.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.cloneVAppTemplateParams+xml</MediaType>
				<ComplexTypeName>CloneVAppTemplateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vAppTemplate.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.customizationSection+xml</MediaType>
				<ComplexTypeName>CustomizationSectionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vAppTemplate.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwNetworkPool.services+xml</MediaType>
				<ComplexTypeName>VendorServicesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vendorServices.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.entity+xml</MediaType>
				<ComplexTypeName>EntityType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/entity.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.entity.reference+xml</MediaType>
				<ComplexTypeName>EntityReferenceType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/entity.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.network+xml</MediaType>
				<ComplexTypeName>NetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.orgNetwork+xml</MediaType>
				<ComplexTypeName>OrgNetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vAppNetwork+xml</MediaType>
				<ComplexTypeName>VAppNetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.allocatedNetworkAddress+xml</MediaType>
				<ComplexTypeName>AllocatedIpAddressesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.subAllocations+xml</MediaType>
				<ComplexTypeName>SubAllocationsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.orgVdcNetwork+xml</MediaType>
				<ComplexTypeName>OrgVdcNetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.edgeGateway+xml</MediaType>
				<ComplexTypeName>GatewayType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.edgeGatewayServiceConfiguration+xml</MediaType>
				<ComplexTypeName>GatewayFeaturesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/network.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.session+xml</MediaType>
				<ComplexTypeName>SessionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/session.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.disk+xml</MediaType>
				<ComplexTypeName>DiskType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/disk.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.diskCreateParams+xml</MediaType>
				<ComplexTypeName>DiskCreateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/disk.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.diskAttachOrDetachParams+xml</MediaType>
				<ComplexTypeName>DiskAttachOrDetachParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/disk.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vdc+xml</MediaType>
				<ComplexTypeName>VdcType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.screenTicket+xml</MediaType>
				<ComplexTypeName>ScreenTicketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/screenTicket.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.productSections+xml</MediaType>
				<ComplexTypeName>ProductSectionListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/productSectionList.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.catalogItem+xml</MediaType>
				<ComplexTypeName>CatalogItemType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/catalogItem.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.tasksList+xml</MediaType>
				<ComplexTypeName>TasksListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/tasksList.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.orgList+xml</MediaType>
				<ComplexTypeName>OrgListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/organizationList.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.org+xml</MediaType>
				<ComplexTypeName>OrgType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/organization.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vm+xml</MediaType>
				<ComplexTypeName>VmType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vmCapabilitiesSection+xml</MediaType>
				<ComplexTypeName>VmCapabilitiesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vApp+xml</MediaType>
				<ComplexTypeName>VAppType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.rasdItemsList+xml</MediaType>
				<ComplexTypeName>RasdItemsListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.networkConfigSection+xml</MediaType>
				<ComplexTypeName>NetworkConfigSectionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.leaseSettingsSection+xml</MediaType>
				<ComplexTypeName>LeaseSettingsSectionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.networkConnectionSection+xml</MediaType>
				<ComplexTypeName>NetworkConnectionSectionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.runtimeInfoSection+xml</MediaType>
				<ComplexTypeName>RuntimeInfoSectionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.guestCustomizationSection+xml</MediaType>
				<ComplexTypeName>GuestCustomizationSectionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.snapshot+xml</MediaType>
				<ComplexTypeName>SnapshotType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.snapshotSection+xml</MediaType>
				<ComplexTypeName>SnapshotSectionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.composeVAppParams+xml</MediaType>
				<ComplexTypeName>ComposeVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.recomposeVAppParams+xml</MediaType>
				<ComplexTypeName>RecomposeVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.registerVAppParams+xml</MediaType>
				<ComplexTypeName>RegisterVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml</MediaType>
				<ComplexTypeName>InstantiateVAppTemplateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.instantiateOvfParams+xml</MediaType>
				<ComplexTypeName>InstantiateOvfParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.cloneVAppParams+xml</MediaType>
				<ComplexTypeName>CloneVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.deployVAppParams+xml</MediaType>
				<ComplexTypeName>DeployVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.undeployVAppParams+xml</MediaType>
				<ComplexTypeName>UndeployVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.mediaInsertOrEjectParams+xml</MediaType>
				<ComplexTypeName>MediaInsertOrEjectParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.captureVAppParams+xml</MediaType>
				<ComplexTypeName>CaptureVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vmPendingQuestion+xml</MediaType>
				<ComplexTypeName>VmPendingQuestionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vmPendingAnswer+xml</MediaType>
				<ComplexTypeName>VmQuestionAnswerType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.relocateVmParams+xml</MediaType>
				<ComplexTypeName>RelocateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.createSnapshotParams+xml</MediaType>
				<ComplexTypeName>CreateSnapshotParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vm.complianceResult+xml</MediaType>
				<ComplexTypeName>ComplianceResultType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vApp.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vdcStorageProfile+xml</MediaType>
				<ComplexTypeName>VdcStorageProfileType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vdcStorageProfile.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.certificateUpdateParams+xml</MediaType>
				<ComplexTypeName>CertificateUpdateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.certificateUploadSocketType+xml</MediaType>
				<ComplexTypeName>CertificateUploadSocketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.keystoreUpdateParams+xml</MediaType>
				<ComplexTypeName>KeystoreUpdateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.keystoreUploadSocketType+xml</MediaType>
				<ComplexTypeName>KeystoreUploadSocketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.sspiKeytabUpdateParams+xml</MediaType>
				<ComplexTypeName>SspiKeytabUpdateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.sspiKeytabUploadSocketType+xml</MediaType>
				<ComplexTypeName>SspiKeytabUploadSocketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.trustStoreUpdateParams+xml</MediaType>
				<ComplexTypeName>TrustStoreUpdateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.trustStoreUploadSocketType+xml</MediaType>
				<ComplexTypeName>TrustStoreUploadSocketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/upload.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.event+xml</MediaType>
				<ComplexTypeName>EventType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/event.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.providervdc+xml</MediaType>
				<ComplexTypeName>ProviderVdcType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.createVdcParams+xml</MediaType>
				<ComplexTypeName>CreateVdcParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vdc+xml</MediaType>
				<ComplexTypeName>AdminVdcType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vdcReferences+xml</MediaType>
				<ComplexTypeName>VdcReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.pvdcStorageProfile+xml</MediaType>
				<ComplexTypeName>ProviderVdcStorageProfileType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vdcStorageProfileParams+xml</MediaType>
				<ComplexTypeName>VdcStorageProfileParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vdcStorageProfile+xml</MediaType>
				<ComplexTypeName>AdminVdcStorageProfileType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.updateVdcStorageProfiles+xml</MediaType>
				<ComplexTypeName>UpdateVdcStorageProfilesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/providerVdc.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.user+xml</MediaType>
				<ComplexTypeName>UserType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/user.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.group+xml</MediaType>
				<ComplexTypeName>GroupType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/user.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.right+xml</MediaType>
				<ComplexTypeName>RightType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/user.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.role+xml</MediaType>
				<ComplexTypeName>RoleType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/user.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vcloud+xml</MediaType>
				<ComplexTypeName>VCloudType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.organization+xml</MediaType>
				<ComplexTypeName>AdminOrgType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vAppTemplateLeaseSettings+xml</MediaType>
				<ComplexTypeName>OrgVAppTemplateLeaseSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.orgSettings+xml</MediaType>
				<ComplexTypeName>OrgSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.organizationGeneralSettings+xml</MediaType>
				<ComplexTypeName>OrgGeneralSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vAppLeaseSettings+xml</MediaType>
				<ComplexTypeName>OrgLeaseSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.organizationFederationSettings+xml</MediaType>
				<ComplexTypeName>OrgFederationSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.organizationLdapSettings+xml</MediaType>
				<ComplexTypeName>OrgLdapSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.organizationEmailSettings+xml</MediaType>
				<ComplexTypeName>OrgEmailSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.organizationPasswordPolicySettings+xml</MediaType>
				<ComplexTypeName>OrgPasswordPolicySettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.catalog+xml</MediaType>
				<ComplexTypeName>AdminCatalogType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.guestPersonalizationSettings+xml</MediaType>
				<ComplexTypeName>OrgGuestPersonalizationSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.operationLimitsSettings+xml</MediaType>
				<ComplexTypeName>OrgOperationLimitsSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vCloudEntities.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.systemSettings+xml</MediaType>
				<ComplexTypeName>SystemSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.generalSettings+xml</MediaType>
				<ComplexTypeName>GeneralSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.amqpSettings+xml</MediaType>
				<ComplexTypeName>AmqpSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.amqpSettingsTest+xml</MediaType>
				<ComplexTypeName>AmqpSettingsTestType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.notificationsSettings+xml</MediaType>
				<ComplexTypeName>NotificationsSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.blockingTaskSettings+xml</MediaType>
				<ComplexTypeName>BlockingTaskSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.systemPasswordPolicySettings+xml</MediaType>
				<ComplexTypeName>SystemPasswordPolicySettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.ldapSettings+xml</MediaType>
				<ComplexTypeName>LdapSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.brandingSettings+xml</MediaType>
				<ComplexTypeName>BrandingSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.licenseSettings+xml</MediaType>
				<ComplexTypeName>LicenseType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.emailSettings+xml</MediaType>
				<ComplexTypeName>EmailSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.kerberosSettings+xml</MediaType>
				<ComplexTypeName>KerberosSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.lookupServiceSettings+xml</MediaType>
				<ComplexTypeName>LookupServiceSettingsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.lookupServiceParams+xml</MediaType>
				<ComplexTypeName>LookupServiceParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vcTrustStoreUpdateParams+xml</MediaType>
				<ComplexTypeName>VcTrustStoreUpdateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vcTrustStoreUploadSocket+xml</MediaType>
				<ComplexTypeName>VcTrustStoreUploadSocketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/settings.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.extensionServices+xml</MediaType>
				<ComplexTypeName>ExtensionServicesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.service+xml</MediaType>
				<ComplexTypeName>AdminServiceType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.apiFilter+xml</MediaType>
				<ComplexTypeName>ApiFilterType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.apiFilters+xml</MediaType>
				<ComplexTypeName>ApiFiltersType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.apiDefinition+xml</MediaType>
				<ComplexTypeName>AdminApiDefinitionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.apiDefinitions+xml</MediaType>
				<ComplexTypeName>AdminApiDefinitionsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.fileDescriptor+xml</MediaType>
				<ComplexTypeName>AdminFileDescriptorType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.serviceLink+xml</MediaType>
				<ComplexTypeName>AdminServiceLinkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.bundleUploadParams+xml</MediaType>
				<ComplexTypeName>BundleUploadParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.bundleUploadSocket+xml</MediaType>
				<ComplexTypeName>BundleUploadSocketType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.aclAccess+xml</MediaType>
				<ComplexTypeName>AclAccessType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.aclRule+xml</MediaType>
				<ComplexTypeName>AclRuleType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.resourceClassAction+xml</MediaType>
				<ComplexTypeName>ResourceClassActionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.resourceClass+xml</MediaType>
				<ComplexTypeName>ResourceClassType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.serviceResource+xml</MediaType>
				<ComplexTypeName>ServiceResourceType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.authorizationCheckParams+xml</MediaType>
				<ComplexTypeName>AuthorizationCheckParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.authorizationCheckResponse+xml</MediaType>
				<ComplexTypeName>AuthorizationCheckResponseType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/services.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwExtension+xml</MediaType>
				<ComplexTypeName>VMWExtensionType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.prepareHostParams+xml</MediaType>
				<ComplexTypeName>PrepareHostParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.registerVimServerParams+xml</MediaType>
				<ComplexTypeName>RegisterVimServerParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwvirtualcenter+xml</MediaType>
				<ComplexTypeName>VimServerType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwVimServerReferences+xml</MediaType>
				<ComplexTypeName>VMWVimServerReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vshieldmanager+xml</MediaType>
				<ComplexTypeName>ShieldManagerType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmsObjectRefsList+xml</MediaType>
				<ComplexTypeName>VmObjectRefsListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmObjectRef+xml</MediaType>
				<ComplexTypeName>VmObjectRefType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.importVmAsVAppParams+xml</MediaType>
				<ComplexTypeName>ImportVmAsVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.importVmIntoExistingVAppParams+xml</MediaType>
				<ComplexTypeName>ImportVmIntoExistingVAppParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.importVmAsVAppTemplateParams+xml</MediaType>
				<ComplexTypeName>ImportVmAsVAppTemplateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.importMediaParams+xml</MediaType>
				<ComplexTypeName>ImportMediaParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.host+xml</MediaType>
				<ComplexTypeName>HostType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vimObjectRef+xml</MediaType>
				<ComplexTypeName>VimObjectRefType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vimObjectRefs+xml</MediaType>
				<ComplexTypeName>VimObjectRefsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwprovidervdc+xml</MediaType>
				<ComplexTypeName>VMWProviderVdcType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.createProviderVdcParams+xml</MediaType>
				<ComplexTypeName>VMWProviderVdcParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwProviderVdcReferences+xml</MediaType>
				<ComplexTypeName>VMWProviderVdcReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwPvdcStorageProfile+xml</MediaType>
				<ComplexTypeName>VMWProviderVdcStorageProfileType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwexternalnet+xml</MediaType>
				<ComplexTypeName>VMWExternalNetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwExternalNetworkReferences+xml</MediaType>
				<ComplexTypeName>VMWExternalNetworkReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwNetworkPoolReferences+xml</MediaType>
				<ComplexTypeName>VMWNetworkPoolReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwNetworkPool+xml</MediaType>
				<ComplexTypeName>VMWNetworkPoolType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.portGroupPool+xml</MediaType>
				<ComplexTypeName>PortGroupPoolType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vlanPool+xml</MediaType>
				<ComplexTypeName>VlanPoolType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vxlanPool+xml</MediaType>
				<ComplexTypeName>VxlanPoolType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vxlanPool+xml</MediaType>
				<ComplexTypeName>VdsContextType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwHostReferences+xml</MediaType>
				<ComplexTypeName>VMWHostReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.resourcePoolList+xml</MediaType>
				<ComplexTypeName>ResourcePoolListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.licensingReport+xml</MediaType>
				<ComplexTypeName>LicensingReportType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.licensingReportList+xml</MediaType>
				<ComplexTypeName>LicensingReportListType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.datastore+xml</MediaType>
				<ComplexTypeName>DatastoreType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwStorageProfiles+xml</MediaType>
				<ComplexTypeName>VMWStorageProfilesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwProviderVdcResourcePoolSet+xml</MediaType>
				<ComplexTypeName>VMWProviderVdcResourcePoolSetType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwProviderVdcResourcePool+xml</MediaType>
				<ComplexTypeName>VMWProviderVdcResourcePoolType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.resourcePoolSetUpdateParams+xml</MediaType>
				<ComplexTypeName>UpdateResourcePoolSetParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.OrganizationVdcResourcePoolSet+xml</MediaType>
				<ComplexTypeName>OrganizationResourcePoolSetType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.strandedItemVimObjects+xml</MediaType>
				<ComplexTypeName>StrandedItemVimObjectType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.strandedItemVimObjects+xml</MediaType>
				<ComplexTypeName>StrandedItemVimObjectsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.strandedItem+xml</MediaType>
				<ComplexTypeName>StrandedItemType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.updateProviderVdcStorageProfiles+xml</MediaType>
				<ComplexTypeName>UpdateProviderVdcStorageProfilesParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.providerVdcMergeParams+xml</MediaType>
				<ComplexTypeName>ProviderVdcMergeParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vSphereWebClientUrl+xml</MediaType>
				<ComplexTypeName>VSphereWebClientUrlType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.updateRightsParams+xml</MediaType>
				<ComplexTypeName>UpdateRightsParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.rights+xml</MediaType>
				<ComplexTypeName>RightRefsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.entityReferences+xml</MediaType>
				<ComplexTypeName>EntityReferencesType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.userEntityRights+xml</MediaType>
				<ComplexTypeName>UserEntityRightsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.migrateVmParams+xml</MediaType>
				<ComplexTypeName>MigrateParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.blockingTask+xml</MediaType>
				<ComplexTypeName>BlockingTaskType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/taskExtensionRequest.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.blockingTaskOperationParams+xml</MediaType>
				<ComplexTypeName>BlockingTaskOperationParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/taskExtensionRequest.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.blockingTaskUpdateProgressOperationParams+xml</MediaType>
				<ComplexTypeName>BlockingTaskUpdateProgressParamsType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/taskExtensionRequest.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.rasdItem+xml</MediaType>
				<ComplexTypeName>RASD_Type</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/master.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.startupSection+xml</MediaType>
				<ComplexTypeName>StartupSection_Type</ComplexTypeName>
				<SchemaLocation>http://schemas.dmtf.org/ovf/envelope/1/dsp8023_1.1.0.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.virtualHardwareSection+xml</MediaType>
				<ComplexTypeName>VirtualHardwareSection_Type</ComplexTypeName>
				<SchemaLocation>http://schemas.dmtf.org/ovf/envelope/1/dsp8023_1.1.0.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.operatingSystemSection+xml</MediaType>
				<ComplexTypeName>OperatingSystemSection_Type</ComplexTypeName>
				<SchemaLocation>http://schemas.dmtf.org/ovf/envelope/1/dsp8023_1.1.0.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.networkSection+xml</MediaType>
				<ComplexTypeName>NetworkSection_Type</ComplexTypeName>
				<SchemaLocation>http://schemas.dmtf.org/ovf/envelope/1/dsp8023_1.1.0.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.vAppNetwork+xml</MediaType>
				<ComplexTypeName>VAppNetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/master.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.network+xml</MediaType>
				<ComplexTypeName>NetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/master.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.vcloud.orgNetwork+xml</MediaType>
				<ComplexTypeName>OrgNetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/master.xsd</SchemaLocation>
			</MediaTypeMapping>
			<MediaTypeMapping>
				<MediaType>application/vnd.vmware.admin.vmwexternalnet+xml</MediaType>
				<ComplexTypeName>VMWExternalNetworkType</ComplexTypeName>
				<SchemaLocation>http://192.168.1.109/api/v1.5/schema/vmwextensions.xsd</SchemaLocation>
			</MediaTypeMapping>
		</VersionInfo>
	</SupportedVersions>`)

	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}

// vCDApiOrgHandler serves mock "/api/org"
func vCDApiOrgHandler(w http.ResponseWriter, r *http.Request) {
	// We expect GET method and not anything else
	if r.Method != http.MethodGet {
		w.WriteHeader(500)
		return
	}

	resp := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<OrgList xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" xmlns:common="http://schemas.dmtf.org/wbem/wscim/1/common" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vmw="http://www.vmware.com/schema/ovf" xmlns:ovfenv="http://schemas.dmtf.org/ovf/environment/1" xmlns:vmext="http://www.vmware.com/vcloud/extension/v1.5" xmlns:ns9="http://www.vmware.com/vcloud/versions" href="https://192.168.1.109/api/org/" type="application/vnd.vmware.vcloud.orgList+xml">
		<Org href="https://192.168.1.109/api/org/c196c6f0-5c31-4929-a626-b29b2c9ff5ab" name="my-org" type="application/vnd.vmware.vcloud.org+xml"/>
	</OrgList>`)

	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}

// vCDSamlMetadataHandler serves mock "/cloud/org/" + org + "/saml/metadata/alias/vcd"
func vCDSamlMetadataHandler(w http.ResponseWriter, r *http.Request) {
	re := []byte(`<?xml version="1.0" encoding="UTF-8"?>
	<md:EntityDescriptor ID="https___192.168.1.109_cloud_org_my-org_saml_metadata_alias_vcd" entityID="https://192.168.1.109/cloud/org/my-org/saml/metadata/alias/vcd" xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"><md:SPSSODescriptor AuthnRequestsSigned="true" WantAssertionsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"><md:KeyDescriptor use="signing"><ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:X509Data><ds:X509Certificate>MIIC4jCCAcqgAwIBAgIEP4rcAjANBgkqhkiG9w0BAQUFADAzMTEwLwYDVQQDEyh2Q2xvdWQgRGly
	ZWN0b3Igb3JnYW5pemF0aW9uIENlcnRpZmljYXRlMB4XDTIwMDMxOTA3MjkwOFoXDTIxMDMxOTA3
	MjkwOFowMzExMC8GA1UEAxModkNsb3VkIERpcmVjdG9yIG9yZ2FuaXphdGlvbiBDZXJ0aWZpY2F0
	ZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKvB9rOzZOUW5AK3TAH9h3p3oFzVOljB
	XSNvOz/OKEL7kVafnPUdxfqJvoZhtTxPOQ9VC9m9t+2sumyXiWCHaOgB/xNWGjzCJci1xFk6YD7j
	y3J1XoQ+JHnL93QJZQK9didH1sjJ7XvtjFA5+1DJyHdTb5CuBH3/Qekyrok3a5ZnwujbwtwGL2NN
	GLjQhEIkioJ67ge/jQWvF5BtthsKh3Jy9SZvMK/cR/s5LfrHHvVu7/ftELlRmfTcBBV2HaZ0lu1H
	QSFvop1pgkD/UIkiqiuI/CdpJwVHoh5AILwOKXHnj1iMqMM+zgRUSFT3LitUM0nsYMypr5ubXbl5
	kpfxlGsCAwEAATANBgkqhkiG9w0BAQUFAAOCAQEAcQRO4lzuS6ec3SX3Vt0EzdKOw7pcsRpHXxbE
	+TlgBlGge0JDDoliaf3Y5QGVjdvMYPn7iwBNHN+DkhRB/5CvgszzhKbyV/FEx+ulnII0Qw03aWVK
	h8L5iPS/1qfBOc67tSKuEuQfXoSSDmJbb3bNmXz1FDh9URAUhoI8wJxYa8SQxiTpaof+WlZ7pRVW
	z9peoenDOMVGcW41gpGA/uXE3PbH66Z5nJTxJvrpFkMtXyu+RBfWHkhQFi9FMWYS9viW+wg+JCqH
	0febOWgCGPqmZ2uUDSMcoSnlYnNdpcv1QXr0NtrKIZt4aXePRmoS7Lxjh671TcznlB7jNCqz+Koh
	5Q==</ds:X509Certificate></ds:X509Data></ds:KeyInfo></md:KeyDescriptor><md:KeyDescriptor use="encryption"><ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:X509Data><ds:X509Certificate>MIIC4jCCAcqgAwIBAgIEP4rcAjANBgkqhkiG9w0BAQUFADAzMTEwLwYDVQQDEyh2Q2xvdWQgRGly
	ZWN0b3Igb3JnYW5pemF0aW9uIENlcnRpZmljYXRlMB4XDTIwMDMxOTA3MjkwOFoXDTIxMDMxOTA3
	MjkwOFowMzExMC8GA1UEAxModkNsb3VkIERpcmVjdG9yIG9yZ2FuaXphdGlvbiBDZXJ0aWZpY2F0
	ZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKvB9rOzZOUW5AK3TAH9h3p3oFzVOljB
	XSNvOz/OKEL7kVafnPUdxfqJvoZhtTxPOQ9VC9m9t+2sumyXiWCHaOgB/xNWGjzCJci1xFk6YD7j
	y3J1XoQ+JHnL93QJZQK9didH1sjJ7XvtjFA5+1DJyHdTb5CuBH3/Qekyrok3a5ZnwujbwtwGL2NN
	GLjQhEIkioJ67ge/jQWvF5BtthsKh3Jy9SZvMK/cR/s5LfrHHvVu7/ftELlRmfTcBBV2HaZ0lu1H
	QSFvop1pgkD/UIkiqiuI/CdpJwVHoh5AILwOKXHnj1iMqMM+zgRUSFT3LitUM0nsYMypr5ubXbl5
	kpfxlGsCAwEAATANBgkqhkiG9w0BAQUFAAOCAQEAcQRO4lzuS6ec3SX3Vt0EzdKOw7pcsRpHXxbE
	+TlgBlGge0JDDoliaf3Y5QGVjdvMYPn7iwBNHN+DkhRB/5CvgszzhKbyV/FEx+ulnII0Qw03aWVK
	h8L5iPS/1qfBOc67tSKuEuQfXoSSDmJbb3bNmXz1FDh9URAUhoI8wJxYa8SQxiTpaof+WlZ7pRVW
	z9peoenDOMVGcW41gpGA/uXE3PbH66Z5nJTxJvrpFkMtXyu+RBfWHkhQFi9FMWYS9viW+wg+JCqH
	0febOWgCGPqmZ2uUDSMcoSnlYnNdpcv1QXr0NtrKIZt4aXePRmoS7Lxjh671TcznlB7jNCqz+Koh
	5Q==</ds:X509Certificate></ds:X509Data></ds:KeyInfo></md:KeyDescriptor><md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://192.168.1.109/cloud/org/my-org/saml/SingleLogout/alias/vcd"/><md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://192.168.1.109/cloud/org/my-org/saml/SingleLogout/alias/vcd"/><md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat><md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</md:NameIDFormat><md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</md:NameIDFormat><md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat><md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName</md:NameIDFormat><md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://192.168.1.109/cloud/org/my-org/saml/SSO/alias/vcd" index="0" isDefault="true"/><md:AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:profiles:holder-of-key:SSO:browser" Location="https://192.168.1.109/cloud/org/my-org/saml/HoKSSO/alias/vcd" hoksso:ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" index="1" xmlns:hoksso="urn:oasis:names:tc:SAML:2.0:profiles:holder-of-key:SSO:browser"/></md:SPSSODescriptor></md:EntityDescriptor>`)
	_, _ = w.Write(re)
}
func getVcdAdfsRedirectHandler(adfsServerHost string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(500)
			return
		}
		headers := w.Header()
		headers.Add("Location", adfsServerHost+"/adfs/ls/?SAMLRequest=lZJBT8MwDIXv%2FIoq9zZpt3UjWjsNJsQkEBMtHLhlqdsFtcmI0wL%2Fnm5lYlyQOFmW7M9P73m%2B%2BGhqrwOLyuiEhAEjHmhpCqWrhDzlN%2F6MLNKLOYqmjvZ82bqdfoS3FtB5S0Swrt%2B7NhrbBmwGtlMSnh7vErJzbo%2Bc0vAyCsJ4FoRByC6prE1bUGMr2nz6h3Lg0ix7oKJWAmknC%2BKterjSwh0VnTjvSvsxq2IWaybKKnD9kF8a25dAg6OiKJHWSIl3Y6yEo9CElKJGIN56lRBRFmxaxFs2epXAKgC5VbvxbBqqSdUfXeNGIKoOfpYQW1hrdEK7hEQsYj4b%2B9E0ZzFnEz4ZBaMJeyHexhpnpKmvlB5ca63mRqBCrkUDyJ3k2fL%2BjkcB49thCPltnm%2F8zUOWE%2B%2F55H50cL%2FPQyMf%2FP6btf8%2BTNIhHn5UbM8JfwPEKUCS%2FieuBpwohBM%2Fmc3puYD0u%2F39LukX&RelayState=aHR0cHM6Ly8xOTIuMTY4LjEuMTA5L3RlbmFudC9teS1vcmc%3D&SigAlg=http%3A%2F%2Fwww.w3.org%2F2000%2F09%2Fxmldsig%23rsa-sha1&Signature=EXL0%2BO1aLhXKAMCKTaqduTW5tWsg94ANZ8hC60MtT4kwitvFUQ7VsQT3qtPj8MFbz0tvN9lX79R0yRwMPilP0zb50uuaVpaJy7qUpHiPyBa5HHA2xG2beyNjlUmC%2BOJSBjfx3k6YMkEzRqfKY6KD%2BKxSMsnSJuazBrWdzihoe4dMgWDS5Dpl2YOC0Ychc1huqedCD2WlE4QRfmtXq0oXlydPVSIYCtHXF1pwYq1j9%2B2q0oK9%2BEEoha0mCMWD74t5hei0kVJldFTcSXx0kgqPi6Rih7aP8%2BlKxnUFu4%2Bo7u9n9Oh8SLV3Tz%2Ba9A9cq4OxdCzyQCOwPRYs3GCb8iIB8g%3D%3D")

		w.WriteHeader(http.StatusFound)
	}
}

// testSpawnAdfsServer spawns mock HTTPS server to server ADFS auth endpoint
// "/adfs/services/trust/13/usernamemixed"
func testSpawnAdfsServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/adfs/services/trust/13/usernamemixed", adfsSamlAuthHandler)
	server := httptest.NewTLSServer(mux)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("ADFS mock server now listening on %s...\n", server.URL)
	}
	return server
}

func adfsSamlAuthHandler(w http.ResponseWriter, r *http.Request) {

	// it must be POST method and not anything else
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	// This is the expected body with dynamic strings replaced to 'REPLACED' word
	expectedBody := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" 
	xmlns:a="http://www.w3.org/2005/08/addressing" 
	xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
	<s:Header>
		<a:Action s:mustUnderstand="1">http://docs.oasis-open.org/ws-sx/ws-trust/200512/RST/Issue</a:Action>
		<a:ReplyTo>
			<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
		</a:ReplyTo>
		<a:To s:mustUnderstand="1">REPLACED</a:To>
		<o:Security s:mustUnderstand="1" 
			xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
			<u:Timestamp u:Id="_0">
				<u:Created>REPLACED</u:Created>
				<u:Expires>REPLACED</u:Expires>
			</u:Timestamp>
			<o:UsernameToken>
				<o:Username>fakeUser</o:Username>
				<o:Password o:Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText">fakePass</o:Password>
			</o:UsernameToken>
		</o:Security>
	</s:Header>
	<s:Body>
		<trust:RequestSecurityToken xmlns:trust="http://docs.oasis-open.org/ws-sx/ws-trust/200512">
			<wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy">
				<a:EndpointReference>
					<a:Address>https://192.168.1.109/cloud/org/my-org/saml/metadata/alias/vcd</a:Address>
				</a:EndpointReference>
			</wsp:AppliesTo>
			<trust:KeySize>0</trust:KeySize>
			<trust:KeyType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Bearer</trust:KeyType>
			<i:RequestDisplayToken xml:lang="en" 
				xmlns:i="http://schemas.xmlsoap.org/ws/2005/05/identity" />
			<trust:RequestType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Issue</trust:RequestType>
			<trust:TokenType>http://docs.oasis-open.org/wss/oasis-wss-saml-token-profile-1.1#SAMLV2.0</trust:TokenType>
		</trust:RequestSecurityToken>
	</s:Body>
</s:Envelope>`

	// Replace known dynamic strings to 'REPLACED' string
	gotBody, _ := ioutil.ReadAll(r.Body)
	gotBodyString := string(gotBody)
	re := regexp.MustCompile(`(<a:To s:mustUnderstand="1">).*(</a:To>)`)
	gotBodyString = re.ReplaceAllString(gotBodyString, `${1}REPLACED${2}`)

	re2 := regexp.MustCompile(`(<u:Created>).*(</u:Created>)`)
	gotBodyString = re2.ReplaceAllString(gotBodyString, `${1}REPLACED${2}`)

	re3 := regexp.MustCompile(`(<u:Expires>).*(</u:Expires>)`)
	gotBodyString = re3.ReplaceAllString(gotBodyString, `${1}REPLACED${2}`)

	if gotBodyString != expectedBody {
		w.WriteHeader(500)
		return
	}

	resp := []byte(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://www.w3.org/2005/08/addressing" xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"><s:Header><a:Action s:mustUnderstand="1">http://docs.oasis-open.org/ws-sx/ws-trust/200512/RSTRC/IssueFinal</a:Action><o:Security s:mustUnderstand="1" xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"><u:Timestamp u:Id="_0"><u:Created>2020-04-27T06:05:53.281Z</u:Created><u:Expires>2020-04-27T06:10:53.281Z</u:Expires></u:Timestamp></o:Security></s:Header><s:Body><trust:RequestSecurityTokenResponseCollection xmlns:trust="http://docs.oasis-open.org/ws-sx/ws-trust/200512"><trust:RequestSecurityTokenResponse><trust:Lifetime><wsu:Created xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">2020-04-27T06:05:53.281Z</wsu:Created><wsu:Expires xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">2020-04-27T07:05:53.281Z</wsu:Expires></trust:Lifetime><wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy"><wsa:EndpointReference xmlns:wsa="http://www.w3.org/2005/08/addressing"><wsa:Address>https://192.168.1.109/cloud/org/my-org/saml/metadata/alias/vcd</wsa:Address></wsa:EndpointReference></wsp:AppliesTo><trust:RequestedSecurityToken><EncryptedAssertion xmlns="urn:oasis:names:tc:SAML:2.0:assertion"><xenc:EncryptedData Type="http://www.w3.org/2001/04/xmlenc#Element" xmlns:xenc="http://www.w3.org/2001/04/xmlenc#"><xenc:EncryptionMethod Algorithm="http://www.w3.org/2001/04/xmlenc#aes256-cbc"/><KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#"><e:EncryptedKey xmlns:e="http://www.w3.org/2001/04/xmlenc#"><e:EncryptionMethod Algorithm="http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p"><DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/></e:EncryptionMethod><KeyInfo><ds:X509Data xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:X509IssuerSerial><ds:X509IssuerName>CN=vCloud Director organization Certificate</ds:X509IssuerName><ds:X509SerialNumber>1066064898</ds:X509SerialNumber></ds:X509IssuerSerial></ds:X509Data></KeyInfo><e:CipherData><e:CipherValue>******</e:CipherValue></e:CipherData></e:EncryptedKey></KeyInfo><xenc:CipherData><xenc:CipherValue>******</xenc:CipherValue></xenc:CipherData></xenc:EncryptedData></EncryptedAssertion></trust:RequestedSecurityToken><i:RequestedDisplayToken xmlns:i="http://schemas.xmlsoap.org/ws/2005/05/identity"><i:DisplayToken xml:lang="en"><i:DisplayClaim Uri="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"><i:DisplayTag>E-Mail Address</i:DisplayTag><i:Description>The e-mail address of the user</i:Description><i:DisplayValue>test@test-forest.net</i:DisplayValue></i:DisplayClaim><i:DisplayClaim Uri="groups"><i:DisplayValue>Domain Users</i:DisplayValue></i:DisplayClaim><i:DisplayClaim Uri="groups"><i:DisplayValue>VCDUsers</i:DisplayValue></i:DisplayClaim></i:DisplayToken></i:RequestedDisplayToken><trust:RequestedAttachedReference><SecurityTokenReference b:TokenType="http://docs.oasis-open.org/wss/oasis-wss-saml-token-profile-1.1#SAMLV2.0" xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" xmlns:b="http://docs.oasis-open.org/wss/oasis-wss-wssecurity-secext-1.1.xsd"><KeyIdentifier ValueType="http://docs.oasis-open.org/wss/oasis-wss-saml-token-profile-1.1#SAMLID">_83958fe2-1b12-4706-a6e5-af2143616c23</KeyIdentifier></SecurityTokenReference></trust:RequestedAttachedReference><trust:RequestedUnattachedReference><SecurityTokenReference b:TokenType="http://docs.oasis-open.org/wss/oasis-wss-saml-token-profile-1.1#SAMLV2.0" xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" xmlns:b="http://docs.oasis-open.org/wss/oasis-wss-wssecurity-secext-1.1.xsd"><KeyIdentifier ValueType="http://docs.oasis-open.org/wss/oasis-wss-saml-token-profile-1.1#SAMLID">_83958fe2-1b12-4706-a6e5-af2143616c23</KeyIdentifier></SecurityTokenReference></trust:RequestedUnattachedReference><trust:TokenType>http://docs.oasis-open.org/wss/oasis-wss-saml-token-profile-1.1#SAMLV2.0</trust:TokenType><trust:RequestType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Issue</trust:RequestType><trust:KeyType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Bearer</trust:KeyType></trust:RequestSecurityTokenResponse></trust:RequestSecurityTokenResponseCollection></s:Body></s:Envelope>`)
	_, err := w.Write(resp)
	if err != nil {
		panic(err)
	}
}
