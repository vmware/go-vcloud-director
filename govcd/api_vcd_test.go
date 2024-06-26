//go:build api || openapi || functional || catalog || vapp || gateway || network || org || query || extnetwork || task || vm || vdc || system || disk || lb || lbAppRule || lbAppProfile || lbServerPool || lbServiceMonitor || lbVirtualServer || user || search || nsxv || nsxt || auth || affinity || role || alb || certificate || vdcGroup || metadata || providervdc || rde || vsphere || uiPlugin || cse || slz || ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func init() {
	testingTags["api"] = "api_vcd_test.go"

	// To list the flags when we run "go test -tags functional -vcd-help", the flag name must start with "vcd"
	// They will all appear alongside the native flags when we use an invalid one
	setBoolFlag(&vcdHelp, "vcd-help", "VCD_HELP", "Show vcd flags")
	setBoolFlag(&enableDebug, "vcd-debug", "GOVCD_DEBUG", "enables debug output")
	setBoolFlag(&testVerbose, "vcd-verbose", "GOVCD_TEST_VERBOSE", "enables verbose output")
	setBoolFlag(&skipVappCreation, "vcd-skip-vapp-creation", "GOVCD_SKIP_VAPP_CREATION", "Skips vApp creation")
	setBoolFlag(&ignoreCleanupFile, "vcd-ignore-cleanup-file", "GOVCD_IGNORE_CLEANUP_FILE", "Does not process previous cleanup file")
	setBoolFlag(&debugShowRequestEnabled, "vcd-show-request", "GOVCD_SHOW_REQ", "Shows API request")
	setBoolFlag(&debugShowResponseEnabled, "vcd-show-response", "GOVCD_SHOW_RESP", "Shows API response")
	setBoolFlag(&connectAsOrgUser, "vcd-as-org-user", "VCD_TEST_ORG_USER", "Connect as Org user")
	setBoolFlag(&connectAsOrgUser, "vcd-test-org-user", "VCD_TEST_ORG_USER", "Connect as Org user")
	flag.IntVar(&connectTenantNum, "vcd-connect-tenant", connectTenantNum, "change index of tenant to use (0=first)")
}

const (
	// Names for entities created by the tests
	TestCreateOrg                 = "TestCreateOrg"
	TestDeleteOrg                 = "TestDeleteOrg"
	TestUpdateOrg                 = "TestUpdateOrg"
	TestCreateCatalog             = "TestCreateCatalog"
	TestCreateCatalogDesc         = "Catalog created by tests"
	TestRefreshOrgFullName        = "TestRefreshOrgFullName"
	TestUpdateCatalog             = "TestUpdateCatalog"
	TestDeleteCatalog             = "TestDeleteCatalog"
	TestRefreshOrg                = "TestRefreshOrg"
	TestComposeVapp               = "TestComposeVapp"
	TestComposeVappDesc           = "vApp created by tests"
	TestSetUpSuite                = "TestSetUpSuite"
	TestUploadOvf                 = "TestUploadOvf"
	TestDeleteCatalogItem         = "TestDeleteCatalogItem"
	TestCreateOrgVdc              = "TestCreateOrgVdc"
	TestRefreshOrgVdc             = "TestRefreshOrgVdc"
	TestCreateOrgVdcNetworkRouted = "TestCreateOrgVdcNetworkRouted"
	TestCreateOrgVdcNetworkDhcp   = "TestCreateOrgVdcNetworkDhcp"
	TestCreateOrgVdcNetworkIso    = "TestCreateOrgVdcNetworkIso"
	TestCreateOrgVdcNetworkDirect = "TestCreateOrgVdcNetworkDirect"
	TestUploadMedia               = "TestUploadMedia"
	TestCatalogUploadMedia        = "TestCatalogUploadMedia"
	TestCreateDisk                = "TestCreateDisk"
	TestUpdateDisk                = "TestUpdateDisk"
	TestDeleteDisk                = "TestDeleteDisk"
	TestRefreshDisk               = "TestRefreshDisk"
	TestAttachedVMDisk            = "TestAttachedVMDisk"
	TestVdcFindDiskByHREF         = "TestVdcFindDiskByHREF"
	TestFindDiskByHREF            = "TestFindDiskByHREF"
	TestDisk                      = "TestDisk"
	// #nosec G101 -- Not a credential
	TestVMAttachOrDetachDisk  = "TestVMAttachOrDetachDisk"
	TestVMAttachDisk          = "TestVMAttachDisk"
	TestVMDetachDisk          = "TestVMDetachDisk"
	TestCreateExternalNetwork = "TestCreateExternalNetwork"
	TestDeleteExternalNetwork = "TestDeleteExternalNetwork"
	TestLbServiceMonitor      = "TestLbServiceMonitor"
	TestLbServerPool          = "TestLbServerPool"
	TestLbAppProfile          = "TestLbAppProfile"
	TestLbAppRule             = "TestLbAppRule"
	TestLbVirtualServer       = "TestLbVirtualServer"
	TestLb                    = "TestLb"
	TestNsxvSnatRule          = "TestNsxvSnatRule"
	TestNsxvDnatRule          = "TestNsxvDnatRule"
)

const (
	TestRequiresSysAdminPrivileges = "Test %s requires system administrator privileges"
)

type Tenant struct {
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
	ApiToken string `yaml:"api_token,omitempty"`
	SysOrg   string `yaml:"sysOrg,omitempty"`
}

// Struct to get info from a config yaml file that the user
// specifies
type TestConfig struct {
	Provider struct {
		User       string `yaml:"user"`
		Password   string `yaml:"password"`
		Token      string `yaml:"token"`
		ApiToken   string `yaml:"api_token"`
		VcdVersion string `yaml:"vcdVersion,omitempty"`
		ApiVersion string `yaml:"apiVersion,omitempty"`

		// UseSamlAdfs specifies if SAML auth is used for authenticating vCD instead of local login.
		// The above `User` and `Password` will be used to authenticate against ADFS IdP when true.
		UseSamlAdfs bool `yaml:"useSamlAdfs"`

		// CustomAdfsRptId allows to set custom Relaying Party Trust identifier if needed. Only has
		// effect if `UseSamlAdfs` is true.
		CustomAdfsRptId string `yaml:"customAdfsRptId"`

		// The variables `SamlUser`, `SamlPassword` and `SamlCustomRptId` are optional and are
		// related to an additional test run specifically with SAML user/password. It is useful in
		// case local user is used for test run (defined by above 'User', 'Password' variables).
		// SamlUser takes ADFS friendly format ('contoso.com\username' or 'username@contoso.com')
		SamlUser        string `yaml:"samlUser,omitempty"`
		SamlPassword    string `yaml:"samlPassword,omitempty"`
		SamlCustomRptId string `yaml:"samlCustomRptId,omitempty"`

		Url             string `yaml:"url"`
		SysOrg          string `yaml:"sysOrg"`
		MaxRetryTimeout int    `yaml:"maxRetryTimeout,omitempty"`
		HttpTimeout     int64  `yaml:"httpTimeout,omitempty"`
	}
	Tenants []Tenant `yaml:"tenants,omitempty"`
	VCD     struct {
		Org         string `yaml:"org"`
		Vdc         string `yaml:"vdc"`
		ProviderVdc struct {
			Name           string `yaml:"name"`
			StorageProfile string `yaml:"storage_profile"`
			NetworkPool    string `yaml:"network_pool"`
		} `yaml:"provider_vdc"`
		NsxtProviderVdc struct {
			Name                   string `yaml:"name"`
			StorageProfile         string `yaml:"storage_profile"`
			StorageProfile2        string `yaml:"storage_profile_2"`
			NetworkPool            string `yaml:"network_pool"`
			PlacementPolicyVmGroup string `yaml:"placementPolicyVmGroup,omitempty"`
		} `yaml:"nsxt_provider_vdc"`
		Catalog struct {
			Name                      string `yaml:"name,omitempty"`
			NsxtBackedCatalogName     string `yaml:"nsxtBackedCatalogName,omitempty"`
			Description               string `yaml:"description,omitempty"`
			CatalogItem               string `yaml:"catalogItem,omitempty"`
			CatalogItemWithEfiSupport string `yaml:"catalogItemWithEfiSupport,omitempty"`
			NsxtCatalogItem           string `yaml:"nsxtCatalogItem,omitempty"`
			NsxtCatalogAddonDse       string `yaml:"nsxtCatalogAddonDse,omitempty"`
			CatalogItemDescription    string `yaml:"catalogItemDescription,omitempty"`
			CatalogItemWithMultiVms   string `yaml:"catalogItemWithMultiVms,omitempty"`
			VmNameInMultiVmItem       string `yaml:"vmNameInMultiVmItem,omitempty"`
		} `yaml:"catalog"`
		Network struct {
			Net1 string `yaml:"network1"`
			Net2 string `yaml:"network2,omitempty"`
		} `yaml:"network"`
		StorageProfile struct {
			SP1 string `yaml:"storageProfile1"`
			SP2 string `yaml:"storageProfile2,omitempty"`
		} `yaml:"storageProfile"`
		ExternalIp                   string `yaml:"externalIp,omitempty"`
		ExternalNetmask              string `yaml:"externalNetmask,omitempty"`
		InternalIp                   string `yaml:"internalIp,omitempty"`
		InternalNetmask              string `yaml:"internalNetmask,omitempty"`
		EdgeGateway                  string `yaml:"edgeGateway,omitempty"`
		ExternalNetwork              string `yaml:"externalNetwork,omitempty"`
		ExternalNetworkPortGroup     string `yaml:"externalNetworkPortGroup,omitempty"`
		ExternalNetworkPortGroupType string `yaml:"externalNetworkPortGroupType,omitempty"`
		VimServer                    string `yaml:"vimServer,omitempty"`
		LdapServer                   string `yaml:"ldapServer,omitempty"`
		OidcServer                   struct {
			Url               string `yaml:"url,omitempty"`
			WellKnownEndpoint string `yaml:"wellKnownEndpoint,omitempty"`
		} `yaml:"oidcServer,omitempty"`
		Nsxt struct {
			Manager                   string `yaml:"manager"`
			Tier0router               string `yaml:"tier0router"`
			Tier0routerVrf            string `yaml:"tier0routerVrf"`
			NsxtDvpg                  string `yaml:"nsxtDvpg"`
			GatewayQosProfile         string `yaml:"gatewayQosProfile"`
			Vdc                       string `yaml:"vdc"`
			ExternalNetwork           string `yaml:"externalNetwork"`
			EdgeGateway               string `yaml:"edgeGateway"`
			NsxtImportSegment         string `yaml:"nsxtImportSegment"`
			NsxtImportSegment2        string `yaml:"nsxtImportSegment2"`
			VdcGroup                  string `yaml:"vdcGroup"`
			VdcGroupEdgeGateway       string `yaml:"vdcGroupEdgeGateway"`
			NsxtEdgeCluster           string `yaml:"nsxtEdgeCluster"`
			RoutedNetwork             string `yaml:"routedNetwork"`
			IsolatedNetwork           string `yaml:"isolatedNetwork"`
			NsxtAlbControllerUrl      string `yaml:"nsxtAlbControllerUrl"`
			NsxtAlbControllerUser     string `yaml:"nsxtAlbControllerUser"`
			NsxtAlbControllerPassword string `yaml:"nsxtAlbControllerPassword"`
			NsxtAlbImportableCloud    string `yaml:"nsxtAlbImportableCloud"`
			NsxtAlbServiceEngineGroup string `yaml:"nsxtAlbServiceEngineGroup"`
			IpDiscoveryProfile        string `yaml:"ipDiscoveryProfile"`
			MacDiscoveryProfile       string `yaml:"macDiscoveryProfile"`
			SpoofGuardProfile         string `yaml:"spoofGuardProfile"`
			QosProfile                string `yaml:"qosProfile"`
			SegmentSecurityProfile    string `yaml:"segmentSecurityProfile"`
		} `yaml:"nsxt"`
	} `yaml:"vcd"`
	Vsphere struct {
		ResourcePoolForVcd1 string `yaml:"resourcePoolForVcd1,omitempty"`
		ResourcePoolForVcd2 string `yaml:"resourcePoolForVcd2,omitempty"`
	} `yaml:"vsphere,omitempty"`
	Logging struct {
		Enabled          bool   `yaml:"enabled,omitempty"`
		LogFileName      string `yaml:"logFileName,omitempty"`
		LogHttpRequest   bool   `yaml:"logHttpRequest,omitempty"`
		LogHttpResponse  bool   `yaml:"logHttpResponse,omitempty"`
		SkipResponseTags string `yaml:"skipResponseTags,omitempty"`
		ApiLogFunctions  string `yaml:"apiLogFunctions,omitempty"`
		VerboseCleanup   bool   `yaml:"verboseCleanup,omitempty"`
	} `yaml:"logging"`
	OVA struct {
		OvaPath            string `yaml:"ovaPath,omitempty"`
		OvaChunkedPath     string `yaml:"ovaChunkedPath,omitempty"`
		OvaMultiVmPath     string `yaml:"ovaMultiVmPath,omitempty"`
		OvaWithoutSizePath string `yaml:"ovaWithoutSizePath,omitempty"`
		OvfPath            string `yaml:"ovfPath,omitempty"`
		OvfUrl             string `yaml:"ovfUrl,omitempty"`
	} `yaml:"ova"`
	Media struct {
		MediaPath        string `yaml:"mediaPath,omitempty"`
		Media            string `yaml:"mediaName,omitempty"`
		NsxtMedia        string `yaml:"nsxtBackedMediaName,omitempty"`
		PhotonOsOvaPath  string `yaml:"photonOsOvaPath,omitempty"`
		MediaUdfTypePath string `yaml:"mediaUdfTypePath,omitempty"`
		UiPluginPath     string `yaml:"uiPluginPath,omitempty"`
	} `yaml:"media"`
	Cse struct {
		Version        string `yaml:"version,omitempty"`
		SolutionsOrg   string `yaml:"solutionsOrg,omitempty"`
		TenantOrg      string `yaml:"tenantOrg,omitempty"`
		TenantVdc      string `yaml:"tenantVdc,omitempty"`
		RoutedNetwork  string `yaml:"routedNetwork,omitempty"`
		EdgeGateway    string `yaml:"edgeGateway,omitempty"`
		StorageProfile string `yaml:"storageProfile,omitempty"`
		OvaCatalog     string `yaml:"ovaCatalog,omitempty"`
		OvaName        string `yaml:"ovaName,omitempty"`
	} `yaml:"cse,omitempty"`
	SolutionAddOn struct {
		Org           string `yaml:"org"`
		Vdc           string `yaml:"vdc"`
		RoutedNetwork string `yaml:"routedNetwork"`
		ComputePolicy string `yaml:"computePolicy"`
		StoragePolicy string `yaml:"storagePolicy"`
		Catalog       string `yaml:"catalog"`
		AddonImageDse string `yaml:"addonImageDse"`
		// DseSolutions contains a nested map of maps. This is done so that the structure is dynamic
		// enough to add new entries, yet maintain the flexibility to have different fields for each
		// of those entities
		DseSolutions map[string]map[string]string `yaml:"dseSolutions,omitempty"`
	} `yaml:"solutionAddOn,omitempty"`
}

// Test struct for vcloud-director.
// Test functions use the struct to get
// an org, vdc, vapp, and client to run
// tests on
type TestVCD struct {
	client         *VCDClient
	org            *Org
	vdc            *Vdc
	nsxtVdc        *Vdc
	vapp           *VApp
	config         TestConfig
	skipVappTests  bool
	skipAdminTests bool
}

// Cleanup entity structure used by the tear-down procedure
// at the end of the tests to remove leftover entities
type CleanupEntity struct {
	Name            string
	EntityType      string
	Parent          string
	CreatedBy       string
	OpenApiEndpoint string
}

// CleanupInfo is the data used to persist an entity list in a file
type CleanupInfo struct {
	VcdIp      string          // IP of the vCD where the test runs
	Info       string          // Information about this file
	EntityList []CleanupEntity // List of entities to remove
}

// Internally used by the test suite to run tests based on TestVCD structures
var _ = Suite(&TestVCD{})

// The list holding the entities to be examined and eventually removed
// at the end of the tests
var cleanupEntityList []CleanupEntity

// Lock for of cleanup entities persistent file
var persistentCleanupListLock sync.Mutex

// IP of the vCD being tested. It is initialized at the first client authentication
var persistentCleanupIp string

// Use this value to run a specific test that does not need a pre-created vApp.
var skipVappCreation bool = os.Getenv("GOVCD_SKIP_VAPP_CREATION") != ""

// vcdHelp shows command line options
var vcdHelp bool

// enableDebug enables debug output
var enableDebug bool

// ignoreCleanupFile prevents processing a previous cleanup file
var ignoreCleanupFile bool

// connectAsOrgUser connects as Org user instead of System administrator
var connectAsOrgUser bool
var connectTenantNum int

// Makes the name for the cleanup entities persistent file
// Using a name for each vCD allows us to run tests with different servers
// and persist the cleanup list for all.
func makePersistentCleanupFileName() string {
	var persistentCleanupListMask = "test_cleanup_list-%s.%s"
	if persistentCleanupIp == "" {
		fmt.Printf("######## Persistent Cleanup IP was not set ########\n")
		os.Exit(1)
	}
	reForbiddenChars := regexp.MustCompile(`[/]`)
	fileName := fmt.Sprintf(persistentCleanupListMask,
		reForbiddenChars.ReplaceAllString(persistentCleanupIp, ""), "json")
	return fileName

}

// Removes the list of cleanup entities
// To be called only after the list has been processed
func removePersistentCleanupList() {
	persistentCleanupListFile := makePersistentCleanupFileName()
	persistentCleanupListLock.Lock()
	defer persistentCleanupListLock.Unlock()
	_, err := os.Stat(persistentCleanupListFile)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(persistentCleanupListFile)
}

// Reads a cleanup list from file
func readCleanupList() ([]CleanupEntity, error) {
	persistentCleanupListFile := makePersistentCleanupFileName()
	persistentCleanupListLock.Lock()
	defer persistentCleanupListLock.Unlock()
	var cleanupInfo CleanupInfo
	_, err := os.Stat(persistentCleanupListFile)
	if os.IsNotExist(err) {
		return nil, err
	}
	listText, err := os.ReadFile(filepath.Clean(persistentCleanupListFile))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(listText, &cleanupInfo)
	return cleanupInfo.EntityList, err
}

// Writes a cleanup list to file.
// If the test suite terminates without having a chance to
// clean up properly, the next run of the test suite will try to
// remove the leftovers
func writeCleanupList(cleanupList []CleanupEntity) error {
	persistentCleanupListFile := makePersistentCleanupFileName()
	persistentCleanupListLock.Lock()
	defer persistentCleanupListLock.Unlock()
	cleanupInfo := CleanupInfo{
		VcdIp: persistentCleanupIp,
		Info: "Persistent list of entities to be destroyed. " +
			" If this file is found when starting the tests, its entities will be " +
			" processed for removal before any other operation.",
		EntityList: cleanupList,
	}
	listText, err := json.MarshalIndent(cleanupInfo, " ", " ")
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Clean(persistentCleanupListFile))
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	count, err := writer.Write(listText)
	if err != nil || count == 0 {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return file.Close()
}

// AddToCleanupList adds an entity to the cleanup list.
// To be called by all tests when a new entity has been created, before
// running any other operation.
// Items in the list will be deleted at the end of the tests if they still exist.
func AddToCleanupList(name, entityType, parent, createdBy string) {
	for _, item := range cleanupEntityList {
		// avoid adding the same item twice
		if item.Name == name && item.EntityType == entityType {
			return
		}
	}
	cleanupEntityList = append(cleanupEntityList, CleanupEntity{Name: name, EntityType: entityType, Parent: parent, CreatedBy: createdBy})
	err := writeCleanupList(cleanupEntityList)
	if err != nil {
		fmt.Printf("################ error writing cleanup list %s\n", err)
	}
}

// PrependToCleanupList prepends an entity to the cleanup list.
// To be called by all tests when a new entity has been created, before
// running any other operation.
// Items in the list will be deleted at the end of the tests if they still exist.
func PrependToCleanupList(name, entityType, parent, createdBy string) {
	for _, item := range cleanupEntityList {
		// avoid adding the same item twice
		if item.Name == name && item.EntityType == entityType {
			return
		}
	}
	cleanupEntityList = append([]CleanupEntity{{Name: name, EntityType: entityType, Parent: parent, CreatedBy: createdBy}}, cleanupEntityList...)
	err := writeCleanupList(cleanupEntityList)
	if err != nil {
		fmt.Printf("################ error writing cleanup list %s\n", err)
	}
}

// AddToCleanupListOpenApi adds an OpenAPI entity OpenApi objects `entityType=OpenApiEntity` and `openApiEndpoint`should
// be set in format "types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + ID"
func AddToCleanupListOpenApi(name, createdBy, openApiEndpoint string) {
	for _, item := range cleanupEntityList {
		// avoid adding the same item twice
		if item.OpenApiEndpoint == openApiEndpoint {
			return
		}
	}
	cleanupEntityList = append(cleanupEntityList, CleanupEntity{Name: name, EntityType: "OpenApiEntity", CreatedBy: createdBy, OpenApiEndpoint: openApiEndpoint})
	err := writeCleanupList(cleanupEntityList)
	if err != nil {
		fmt.Printf("################ error writing cleanup list %s\n", err)
	}
}

// PrependToCleanupListOpenApi prepends an OpenAPI entity OpenApi objects `entityType=OpenApiEntity` and
// `openApiEndpoint`should be set in format "types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + ID"
func PrependToCleanupListOpenApi(name, createdBy, openApiEndpoint string) {
	for _, item := range cleanupEntityList {
		// avoid adding the same item twice
		if item.OpenApiEndpoint == openApiEndpoint {
			return
		}
	}
	cleanupEntityList = append([]CleanupEntity{{Name: name, EntityType: "OpenApiEntity", CreatedBy: createdBy, OpenApiEndpoint: openApiEndpoint}}, cleanupEntityList...)
	err := writeCleanupList(cleanupEntityList)
	if err != nil {
		fmt.Printf("################ error writing cleanup list %s\n", err)
	}
}

// Users use the environmental variable GOVCD_CONFIG as
// a config file for testing. Otherwise the default is govcd_test_config.yaml
// in the current directory. Throws an error if it cannot find your
// yaml file or if it cannot read it.
func GetConfigStruct() (TestConfig, error) {
	config := os.Getenv("GOVCD_CONFIG")
	var configStruct TestConfig
	if config == "" {
		// Finds the current directory, through the path of this running test
		_, currentFilename, _, _ := runtime.Caller(0)
		currentDirectory := filepath.Dir(currentFilename)
		config = currentDirectory + "/govcd_test_config.yaml"
	}
	// Looks if the configuration file exists before attempting to read it
	_, err := os.Stat(config)
	if os.IsNotExist(err) {
		return TestConfig{}, fmt.Errorf("Configuration file %s not found: %s", config, err)
	}
	yamlFile, err := os.ReadFile(filepath.Clean(config))
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not read config file %s: %s", config, err)
	}
	err = yaml.Unmarshal(yamlFile, &configStruct)
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not unmarshal yaml file: %s", err)
	}
	if connectAsOrgUser {
		if len(configStruct.Tenants) == 0 {
			return TestConfig{}, fmt.Errorf("org user connection required, but 'tenants[%d]' is empty", connectTenantNum)
		}
		if connectTenantNum > len(configStruct.Tenants)-1 {
			return TestConfig{}, fmt.Errorf("org user connection required, but tenant number %d is higher than the number of tenants ", connectTenantNum)
		}
		// Change configStruct.Provider, to reuse the global fields, such as URL
		configStruct.Provider.User = configStruct.Tenants[connectTenantNum].User
		configStruct.Provider.Password = configStruct.Tenants[connectTenantNum].Password
		configStruct.Provider.SysOrg = configStruct.Tenants[connectTenantNum].SysOrg
		configStruct.Provider.Token = configStruct.Tenants[connectTenantNum].Token
		configStruct.Provider.ApiToken = configStruct.Tenants[connectTenantNum].ApiToken
	}
	return configStruct, nil
}

// Creates a VCDClient based on the endpoint given in the TestConfig argument.
// TestConfig struct can be obtained by calling GetConfigStruct. Throws an error
// if endpoint given is not a valid url.
func GetTestVCDFromYaml(testConfig TestConfig, options ...VCDClientOption) (*VCDClient, error) {
	configUrl, err := url.ParseRequestURI(testConfig.Provider.Url)
	if err != nil {
		return &VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}

	if testConfig.Provider.MaxRetryTimeout != 0 {
		options = append(options, WithMaxRetryTimeout(testConfig.Provider.MaxRetryTimeout))
	}

	if testConfig.Provider.HttpTimeout != 0 {
		options = append(options, WithHttpTimeout(testConfig.Provider.HttpTimeout))
	}

	if testConfig.Provider.UseSamlAdfs {
		options = append(options, WithSamlAdfs(true, testConfig.Provider.CustomAdfsRptId))
	}

	return NewVCDClient(*configUrl, true, options...), nil
}

// Necessary to enable the suite tests with TestVCD
func Test(t *testing.T) { TestingT(t) }

// Sets the org, vdc, vapp, and vcdClient for a
// TestVCD struct. An error is thrown if something goes wrong
// getting config file, creating vcd, during authentication, or
// when creating a new vapp. If this method panics, no test
// case that uses the TestVCD struct is run.
func (vcd *TestVCD) SetUpSuite(check *C) {
	flag.Parse()
	setTestEnv()
	if vcdHelp {
		fmt.Println("vcd flags:")
		fmt.Println()
		// Prints only the flags defined in this package
		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, "vcd-") {
				fmt.Printf("  -%-40s %s (%v)\n", f.Name, f.Usage, f.Value)
			}
		})
		fmt.Println()
		// This will skip the whole suite.
		// Instead, running os.Exit(0) will panic
		check.Skip("--- showing help ---")
	}
	config, err := GetConfigStruct()
	if config.Provider.Url == "" || err != nil {
		panic(err)
	}
	vcd.config = config

	// This library sets HTTP User-Agent to be `go-vcloud-director` by default and all HTTP calls
	// expected to contain this header. An explicit test cannot capture future HTTP requests, but
	// of them should use logging so this should be a good 'gate' to ensure ALL HTTP calls going out
	// of this library do include HTTP User-Agent.
	util.TogglePanicEmptyUserAgent(true)

	if vcd.config.Logging.Enabled {
		util.EnableLogging = true
		if vcd.config.Logging.LogFileName != "" {
			util.ApiLogFileName = vcd.config.Logging.LogFileName
		}
		if vcd.config.Logging.LogHttpRequest {
			util.LogHttpRequest = true
		}
		if vcd.config.Logging.LogHttpResponse {
			util.LogHttpResponse = true
		}
		if vcd.config.Logging.SkipResponseTags != "" {
			util.SetSkipTags(vcd.config.Logging.SkipResponseTags)
		}
		if vcd.config.Logging.ApiLogFunctions != "" {
			util.SetApiLogFunctions(vcd.config.Logging.ApiLogFunctions)
		}
	} else {
		util.EnableLogging = false
	}
	util.SetLog()
	vcdClient, err := GetTestVCDFromYaml(config)
	if vcdClient == nil || err != nil {
		panic(err)
	}
	vcd.client = vcdClient
	token := os.Getenv("VCD_TOKEN")
	if token == "" {
		token = config.Provider.Token
	}

	apiToken := os.Getenv("VCD_API_TOKEN")
	if apiToken == "" {
		apiToken = config.Provider.ApiToken
	}

	authenticationMode := "password"
	if apiToken != "" {
		authenticationMode = "API-token"
		err = vcd.client.SetToken(config.Provider.SysOrg, ApiTokenHeader, apiToken)
	} else {
		if token != "" {
			authenticationMode = "token"
			err = vcd.client.SetToken(config.Provider.SysOrg, AuthorizationHeader, token)
		} else {
			err = vcd.client.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
		}
	}
	if config.Provider.UseSamlAdfs {
		authenticationMode = "SAML password"
	}
	if err != nil {
		panic(err)
	}
	versionInfo := ""
	version, versionTime, err := vcd.client.Client.GetVcdVersion()
	if err == nil {
		versionInfo = fmt.Sprintf("version %s built at %s", version, versionTime)
	}
	fmt.Printf("Running on VCD %s (%s)\nas user %s@%s (using %s)\n", vcd.config.Provider.Url, versionInfo,
		vcd.config.Provider.User, vcd.config.Provider.SysOrg, authenticationMode)
	if !vcd.client.Client.IsSysAdmin {
		vcd.skipAdminTests = true
	}

	// Sets the vCD IP value, removing the elements that would
	// not be appropriate in a file name
	reHttp := regexp.MustCompile(`^https?://`)
	reApi := regexp.MustCompile(`/api/?`)
	persistentCleanupIp = vcd.config.Provider.Url
	persistentCleanupIp = reHttp.ReplaceAllString(persistentCleanupIp, "")
	persistentCleanupIp = reApi.ReplaceAllString(persistentCleanupIp, "")
	// set org
	vcd.org, err = vcd.client.GetOrgByName(config.VCD.Org)
	if err != nil {
		fmt.Printf("error retrieving org %s: %s\n", config.VCD.Org, err)
		os.Exit(1)
	}
	// set vdc
	vcd.vdc, err = vcd.org.GetVDCByName(config.VCD.Vdc, false)
	if err != nil || vcd.vdc == nil {
		panic(err)
	}

	// configure NSX-T VDC for convenience if it is specified in configuration
	if config.VCD.Nsxt.Vdc != "" {
		vcd.nsxtVdc, err = vcd.org.GetVDCByName(config.VCD.Nsxt.Vdc, false)
		if err != nil {
			panic(fmt.Errorf("error geting NSX-T VDC '%s': %s", config.VCD.Nsxt.Vdc, err))
		}
	}

	// If neither the vApp or VM tags are set, we also skip the
	// creation of the default vApp
	if !isTagSet("vapp") && !isTagSet("vm") {
		// vcd.skipVappTests = true
		skipVappCreation = true
	}

	// Gets the persistent cleanup list from file, if exists.
	cleanupList, err := readCleanupList()
	if len(cleanupList) > 0 && err == nil {
		if !ignoreCleanupFile {
			// If we found a cleanup file and we want to process it (default)
			// We proceed to cleanup the leftovers before any other operation
			fmt.Printf("*** Found cleanup file %s\n", makePersistentCleanupFileName())
			for i, cleanupEntity := range cleanupList {
				fmt.Printf("# %d ", i+1)
				vcd.removeLeftoverEntities(cleanupEntity)
			}
		}
		removePersistentCleanupList()
	}

	// creates a new VApp for vapp tests
	if !skipVappCreation && config.VCD.Network.Net1 != "" && config.VCD.StorageProfile.SP1 != "" &&
		config.VCD.Catalog.Name != "" && config.VCD.Catalog.CatalogItem != "" {
		// deployVappForTest replaces the old createTestVapp() because it was using bad implemented method vdc.ComposeVApp
		vcd.vapp, err = deployVappForTest(vcd, TestSetUpSuite)
		// If no vApp is created, we skip all vApp tests
		if err != nil {
			fmt.Printf("%s\n", err)
			panic("Creation failed - Bailing out")
		}
		if vcd.vapp == nil {
			fmt.Printf("Creation of vApp %s failed unexpectedly. No error was reported, but vApp is empty\n", TestSetUpSuite)
			panic("initial vApp is empty - bailing out")
		}
	} else {
		vcd.skipVappTests = true
		fmt.Println("Skipping all vapp tests because one of the following wasn't given: Network, StorageProfile, Catalog, Catalogitem")
	}
}

// Shows the detail of cleanup operations only if the relevant verbosity
// has been enabled
func (vcd *TestVCD) infoCleanup(format string, args ...interface{}) {
	if vcd.config.Logging.VerboseCleanup {
		fmt.Printf(format, args...)
	}
}

func getOrgVdcByNames(vcd *TestVCD, orgName, vdcName string) (*Org, *Vdc, error) {
	if orgName == "" || vdcName == "" {
		return nil, nil, fmt.Errorf("orgName, vdcName cant be empty")
	}

	org, _ := vcd.client.GetOrgByName(orgName)
	if org == nil {
		vcd.infoCleanup("could not find org '%s'", orgName)
		return nil, nil, fmt.Errorf("can't find org")
	}
	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		vcd.infoCleanup("could not find vdc '%s'", vdcName)
		return nil, nil, fmt.Errorf("can't find vdc")
	}

	return org, vdc, nil
}

func getOrgVdcEdgeByNames(vcd *TestVCD, orgName, vdcName, edgeName string) (*Org, *Vdc, *EdgeGateway, error) {
	if orgName == "" || vdcName == "" || edgeName == "" {
		return nil, nil, nil, fmt.Errorf("orgName, vdcName, edgeName cant be empty")
	}

	org, vdc, err := getOrgVdcByNames(vcd, orgName, vdcName)
	if err != nil {
		return nil, nil, nil, err
	}

	edge, err := vdc.GetEdgeGatewayByName(edgeName, false)

	if err != nil {
		vcd.infoCleanup("could not find edge '%s': %s", edgeName, err)
	}
	return org, vdc, edge, nil
}

var splitParentNotFound string = "removeLeftoverEntries: [ERROR] missing parent info (%s). The parent fields must be defined with a separator '|'\n"
var notFoundMsg string = "removeLeftoverEntries: [INFO] No action for %s '%s'\n"

func (vcd *TestVCD) getAdminOrgAndVdcFromCleanupEntity(entity CleanupEntity) (org *AdminOrg, vdc *Vdc, err error) {
	orgName, vdcName, _ := splitParent(entity.Parent, "|")
	if orgName == "" || vdcName == "" {
		vcd.infoCleanup(splitParentNotFound, entity.Parent)
		return nil, nil, fmt.Errorf("can't find parents names")
	}
	org, err = vcd.client.GetAdminOrgByName(orgName)
	if err != nil {
		vcd.infoCleanup(notFoundMsg, "org", orgName)
		return nil, nil, fmt.Errorf("can't find org")
	}
	vdc, err = org.GetVDCByName(vdcName, false)
	if vdc == nil || err != nil {
		vcd.infoCleanup(notFoundMsg, "vdc", vdcName)
		return nil, nil, fmt.Errorf("can't find vdc")
	}
	return org, vdc, nil
}

// Removes leftover entities that may still exist after failed tests
// or the ones that were explicitly created for several tests and
// were relying on this procedure to clean up at the end.
func (vcd *TestVCD) removeLeftoverEntities(entity CleanupEntity) {
	var introMsg string = "removeLeftoverEntries: [INFO] Attempting cleanup of %s '%s' instantiated by %s\n"
	var removedMsg string = "removeLeftoverEntries: [INFO] Removed %s '%s' created by %s\n"
	var notDeletedMsg string = "removeLeftoverEntries: [ERROR] Error deleting %s '%s': %s\n"
	// NOTE: this is a cleanup function that should continue even if errors are found.
	// For this reason, the [ERROR] messages won't be followed by a program termination
	vcd.infoCleanup(introMsg, entity.EntityType, entity.Name, entity.CreatedBy)
	switch entity.EntityType {

	// openApiEntity can be used to delete any OpenAPI entity due to the API being uniform and allowing the same
	// low level OpenApiDeleteItem()
	case "OpenApiEntity":
		// entity.OpenApiEndpoint contains "endpoint/{ID}"
		// (in format types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointOrgVdcNetworks + ID) but
		// to lookup used API version this ID must not be present therefore below we remove suffix ID.
		// This is done by splitting whole path by "/" and rebuilding path again without last element in slice (which is
		// expected to be the ID)
		// Sometimes API endpoint path might contain URNs in the middle (e.g. OpenApiEndpointNsxtNatRules). They are
		// replaced back to string placeholder %s to match original definitions
		endpointSlice := strings.Split(entity.OpenApiEndpoint, "/")
		endpointWithUuid := strings.Join(endpointSlice[:len(endpointSlice)-1], "/") + "/"
		// replace any "urns" (e.g. 'urn:vcloud:gateway:64966c36-e805-44e2-980b-c1077ab54956') with '%s' to match API definitions
		re := regexp.MustCompile(`urn[^\/]+`) // Regexp matches from 'urn' up to next '/' in the path
		endpointRemovedUuids := re.ReplaceAllString(endpointWithUuid, "%s")
		apiVersion, _ := vcd.client.Client.checkOpenApiEndpointCompatibility(endpointRemovedUuids)

		// Build UP complete endpoint address
		urlRef, err := vcd.client.Client.OpenApiBuildEndpoint(entity.OpenApiEndpoint)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		// Validate if the resource still exists
		err = vcd.client.Client.OpenApiGetItem(apiVersion, urlRef, nil, nil, nil)

		// RDE Framework has a bug in VCD 10.3.0 that causes "not found" errors to return as "400 bad request",
		// so we need to amend them
		if strings.Contains(entity.OpenApiEndpoint, types.OpenApiEndpointRdeInterfaces) {
			err = amendRdeApiError(&vcd.client.Client, err)
		}
		// UI Plugin has a bug in VCD 10.4.x that causes "not found" errors to return a NullPointerException,
		// so we need to amend them
		if strings.Contains(entity.OpenApiEndpoint, types.OpenApiEndpointExtensionsUi) {
			err = amendUIPluginGetByIdError(entity.Name, err)
		}

		if ContainsNotFound(err) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		// Attempt to use supplied path in entity.Parent for element deletion
		err = vcd.client.Client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
	// 	OpenApiEntityFirewall has different API structure therefore generic `OpenApiEntity` case does not fit cleanup
	case "OpenApiEntityFirewall":
		apiVersion, err := vcd.client.Client.checkOpenApiEndpointCompatibility(types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointNsxtFirewallRules)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		urlRef, err := vcd.client.Client.OpenApiBuildEndpoint(entity.Name)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		// Attempt to use supplied path in entity.Parent for element deletion
		err = vcd.client.Client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
	case "OpenApiEntityGlobalDefaultSegmentProfileTemplate":
		// Check if any default settings are applied
		gdSpt, err := vcd.client.GetGlobalDefaultSegmentProfileTemplates()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		if gdSpt.VappNetworkSegmentProfileTemplateRef == nil && gdSpt.VdcNetworkSegmentProfileTemplateRef == nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		_, err = vcd.client.UpdateGlobalDefaultSegmentProfileTemplates(&types.NsxtGlobalDefaultSegmentProfileTemplate{})
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
	// 	OpenApiEntityAlbSettingsDisable has different API structure therefore generic `OpenApiEntity` case does not fit cleanup
	case "OpenApiEntityAlbSettingsDisable":
		edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(entity.Parent)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}

		edgeAlbSettingsConfig, err := edge.GetAlbSettings()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		if edgeAlbSettingsConfig.Enabled == false {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		err = edge.DisableAlb()
		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
	case "vapp":
		vdc := vcd.vdc
		var err error

		// Check if parent VDC was specified. If not - use the default NSX-V VDC
		if entity.Parent != "" {
			vdc, err = vcd.org.GetVDCByName(entity.Parent, true)
			if err != nil {
				vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
				return
			}
		}

		vapp, err := vdc.GetVAppByName(entity.Name, true)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		task, _ := vapp.Undeploy()
		_ = task.WaitTaskCompletion()
		// Detach all Org networks during vApp removal because network removal errors if it happens
		// very quickly (as the next task) after vApp removal
		task, _ = vapp.RemoveAllNetworks()
		_ = task.WaitTaskCompletion()
		task, err = vapp.Delete()
		_ = task.WaitTaskCompletion()

		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	case "catalog":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No Org provided for catalog '%s'\n", entity.Name)
			return
		}
		org, err := vcd.client.GetAdminOrgByName(entity.Parent)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [INFO] organization '%s' not found\n", entity.Parent)
			return
		}
		catalog, err := org.GetAdminCatalogByName(entity.Name, false)
		if catalog == nil || err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		err = catalog.Delete(true, true)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	case "org":
		org, err := vcd.client.GetAdminOrgByName(entity.Name)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		err = org.Delete(true, true)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "provider_vdc":
		pvdc, err := vcd.client.GetProviderVdcExtendedByName(entity.Name)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		err = pvdc.Disable()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		task, err := pvdc.Delete()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "catalogItem":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No Org provided for catalogItem '%s'\n", strings.Split(entity.Parent, "|")[0])
			return
		}
		org, err := vcd.client.GetAdminOrgByName(strings.Split(entity.Parent, "|")[0])
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [INFO] organization '%s' not found\n", entity.Parent)
			return
		}
		catalog, err := org.GetCatalogByName(strings.Split(entity.Parent, "|")[1], false)
		if catalog == nil || err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		catalogItem, err := catalog.GetCatalogItemByName(entity.Name, false)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		err = catalogItem.Delete()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "edgegateway":
		_, vdc, err := vcd.getAdminOrgAndVdcFromCleanupEntity(entity)
		if err != nil {
			return
		}
		edge, err := vdc.GetEdgeGatewayByName(entity.Name, false)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [INFO] edge gateway '%s' not found\n", entity.Name)
			return
		}
		err = edge.Delete(true, true)
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
			return
		}
		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "network":
		_, vdc, err := vcd.getAdminOrgAndVdcFromCleanupEntity(entity)
		if err != nil {
			vcd.infoCleanup("%s", err)
		}
		_, errExists := vdc.GetOrgVdcNetworkByName(entity.Name, false)
		networkExists := errExists == nil

		err = RemoveOrgVdcNetworkIfExists(*vdc, entity.Name)
		if err == nil {
			if networkExists {
				vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
			} else {
				vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			}
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "affinity_rule":
		_, vdc, err := vcd.getAdminOrgAndVdcFromCleanupEntity(entity)
		if err != nil {
			vcd.infoCleanup("adminOrg + VDC: %s", err)
			return
		}
		affinityRule, err := vdc.GetVmAffinityRuleById(entity.Name)
		if err != nil {
			if ContainsNotFound(err) {
				vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
				return
			} else {
				vcd.infoCleanup("retrieving affinity rule %s", err)
				return
			}
		}
		err = affinityRule.Delete()
		if err != nil {
			vcd.infoCleanup("affinity rule deletion: %s", err)
		}
		return

	case "externalNetwork":
		externalNetwork, err := vcd.client.GetExternalNetworkByName(entity.Name)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "externalNetwork", entity.Name)
			return
		}
		err = externalNetwork.DeleteWait()
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "mediaImage":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No VDC and ORG provided for media '%s'\n", entity.Name)
			return
		}
		_, vdc, err := vcd.getAdminOrgAndVdcFromCleanupEntity(entity)
		if err != nil {
			vcd.infoCleanup("%s", err)
		}
		err = RemoveMediaImageIfExists(*vdc, entity.Name)
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "mediaCatalogImage":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No Catalog provided for media '%s'\n", entity.Name)
			return
		}
		orgName, catalogName, _ := splitParent(entity.Parent, "|")
		if orgName == "" || catalogName == "" {
			vcd.infoCleanup(splitParentNotFound, entity.Parent)
		}

		org, err := vcd.client.GetAdminOrgByName(orgName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [INFO] organization '%s' not found\n", entity.Parent)
			return
		}
		adminCatalog, err := org.GetAdminCatalogByName(catalogName, false)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		_, err = adminCatalog.GetMediaByName(entity.Name, true)
		if ContainsNotFound(err) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		err = adminCatalog.RemoveMediaIfExists(entity.Name)
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "user":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No ORG provided for user '%s'\n", entity.Name)
			return
		}
		org, err := vcd.client.GetAdminOrgByName(entity.Parent)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "org", entity.Parent)
			return
		}
		user, err := org.GetUserByName(entity.Name, true)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "user", entity.Name)
			return
		}
		err = user.Delete(true)
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "group":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No ORG provided for group '%s'\n", entity.Name)
			return
		}
		org, err := vcd.client.GetAdminOrgByName(entity.Parent)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "org", entity.Parent)
			return
		}
		group, err := org.GetGroupByName(entity.Name, true)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "group", entity.Name)
			return
		}
		err = group.Delete()
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "vdc":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No ORG provided for VDC '%s'\n", entity.Name)
			return
		}
		org, err := vcd.client.GetAdminOrgByName(entity.Parent)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "org", entity.Parent)
			return
		}
		vdc, err := org.GetVDCByName(entity.Name, false)
		if vdc == nil || err != nil {
			vcd.infoCleanup(notFoundMsg, "vdc", entity.Name)
			return
		}
		err = vdc.DeleteWait(true, true)
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "nsxv_dfw":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No ORG provided for VDC '%s'\n", entity.Name)
			return
		}
		org, err := vcd.client.GetAdminOrgByName(entity.Parent)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "org", entity.Parent)
			return
		}
		vdc, err := org.GetVDCByName(entity.Name, false)
		if vdc == nil || err != nil {
			vcd.infoCleanup(notFoundMsg, "vdc", entity.Name)
			return
		}
		dfw := NewNsxvDistributedFirewall(vdc.client, vdc.Vdc.ID)
		enabled, err := dfw.IsEnabled()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] checking distributed firewall from VCD '%s': %s", entity.Name, err)
			return
		}
		if !enabled {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		err = dfw.Disable()
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] removing distributed firewall from VCD '%s': %s", entity.Name, err)
			return
		}
	case "standaloneVm":
		vm, err := vcd.org.QueryVmById(entity.Name) // The VM ID must be passed as Name
		if IsNotFound(err) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] retrieving standalone VM '%s'. %s\n",
				entity.Name, err)
			return
		}
		err = vm.Delete()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] deleting VM '%s' : %s\n",
				entity.Name, err)
			return
		}
		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
	case "vm":
		vapp, err := vcd.vdc.GetVAppByName(entity.Parent, true)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		vm, err := vapp.GetVMByName(entity.Name, false)

		if err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		// Try to undeploy and ignore errors if it doesn't work (VM may already be powered off)
		task, _ := vm.Undeploy()
		_ = task.WaitTaskCompletion()

		err = vapp.RemoveVM(*vm)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] Deleting VM '%s' in vApp '%s': %s\n",
				entity.Name, vapp.VApp.Name, err)
			return
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	case "disk":
		// Find disk by href rather than find disk by name, because disk name can be duplicated in VDC,
		// so the unique href is required for finding the disk.
		disk, err := vcd.vdc.GetDiskByHref(entity.Name)
		if err != nil {
			// If the disk is not found, we just need to show that it was not found, as
			// it was likely deleted during the regular tests
			vcd.infoCleanup(notFoundMsg, entity.Name, err)
			return
		}

		// See if the disk is attached to the VM
		vmRef, err := disk.AttachedVM()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] Deleting %s '%s', cannot find attached VM: %s\n",
				entity.EntityType, entity.Name, err)
			return
		}
		// If the disk is attached to the VM, detach disk from the VM
		if vmRef != nil {
			vcd.infoCleanup("removeLeftoverEntries: [INFO] Deleting %s '%s', VM: '%s|%s', disk is attached, detaching disk\n",
				entity.EntityType, entity.Name, vmRef.Name, vmRef.HREF)

			vm, err := vcd.client.Client.GetVMByHref(vmRef.HREF)
			if err != nil {
				vcd.infoCleanup(
					"removeLeftoverEntries: [ERROR] Deleting %s '%s', VM: '%s|%s', cannot find the VM details: %s\n",
					entity.EntityType, entity.Name, vmRef.Name, vmRef.HREF, err)
				return
			}

			// Detach the disk from VM
			task, err := vm.DetachDisk(&types.DiskAttachOrDetachParams{
				Disk: &types.Reference{
					HREF: disk.Disk.HREF,
				},
			})
			if err != nil {
				vcd.infoCleanup(
					"removeLeftoverEntries: [ERROR] Detaching %s '%s', VM: '%s|%s': %s\n",
					entity.EntityType, entity.Name, vmRef.Name, vmRef.HREF, err)
				return
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				vcd.infoCleanup(
					"removeLeftoverEntries: [ERROR] Deleting %s '%s', VM: '%s|%s', waitTaskCompletion of detach disk is failed: %s\n",
					entity.EntityType, entity.Name, vmRef.Name, vmRef.HREF, err)
				return
			}

			// We need to refresh the disk info to obtain remove href for delete disk
			// because attached disk is not showing remove disk href in disk.Disk.Link
			err = disk.Refresh()
			if err != nil {
				vcd.infoCleanup(
					"removeLeftoverEntries: [ERROR] Deleting %s '%s', cannot refresh disk: %s\n",
					entity.EntityType, entity.Name, err)
				return
			}
		}

		// Delete disk
		deleteDiskTask, err := disk.Delete()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] Deleting %s '%s', cannot delete disk: %s\n",
				entity.EntityType, entity.Name, err)
			return
		}
		err = deleteDiskTask.WaitTaskCompletion()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] Deleting %s '%s', waitTaskCompletion of delete disk is failed: %s\n",
				entity.EntityType, entity.Name, err)
			return
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "lbServiceMonitor":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.DeleteLbServiceMonitorByName(entity.Name)
		if err != nil && strings.Contains(err.Error(), ErrorEntityNotFound.Error()) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	case "lbServerPool":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.DeleteLbServerPoolByName(entity.Name)
		if err != nil && strings.Contains(err.Error(), ErrorEntityNotFound.Error()) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "lbAppProfile":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.DeleteLbAppProfileByName(entity.Name)
		if err != nil && strings.Contains(err.Error(), ErrorEntityNotFound.Error()) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	case "lbVirtualServer":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.DeleteLbVirtualServerByName(entity.Name)
		if err != nil && strings.Contains(err.Error(), ErrorEntityNotFound.Error()) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "lbAppRule":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.DeleteLbAppRuleByName(entity.Name)
		if err != nil && strings.Contains(err.Error(), ErrorEntityNotFound.Error()) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	// Edge gateway DHCP relay settings cannot actually be "deleted". They can only be unset and this is
	// what this cleanup case does.
	case "dhcpRelayConfig":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.ResetDhcpRelay()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	case "vdcMetaData":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No VDC and ORG provided for vDC meta data '%s'\n", entity.Name)
			return
		}
		_, vdc, err := vcd.getAdminOrgAndVdcFromCleanupEntity(entity)
		if err != nil {
			vcd.infoCleanup("%s", err)
		}
		_, err = vdc.DeleteMetadata(entity.Name)
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return

	case "nsxvNatRule":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.DeleteNsxvNatRuleById(entity.Name)
		if IsNotFound(err) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	case "nsxvFirewallRule":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, edgeName := splitParent(entity.Parent, "|")

		_, _, edge, err := getOrgVdcEdgeByNames(vcd, orgName, vdcName, edgeName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = edge.DeleteNsxvFirewallRuleById(entity.Name)
		if IsNotFound(err) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "ipSet":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No parent specified '%s'\n", entity.Name)
			return
		}

		orgName, vdcName, _ := splitParent(entity.Parent, "|")

		_, vdc, err := getOrgVdcByNames(vcd, orgName, vdcName)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		err = vdc.DeleteNsxvIpSetByName(entity.Name)
		if IsNotFound(err) {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
	case "fastProvisioning":
		orgName, vdcName, _ := splitParent(entity.Parent, "|")
		if orgName == "" || vdcName == "" {
			vcd.infoCleanup(splitParentNotFound, entity.Parent)
		}
		org, err := vcd.client.GetAdminOrgByName(orgName)
		if err != nil {
			vcd.infoCleanup(notFoundMsg, "org", orgName)
		}
		adminVdc, err := org.GetAdminVDCByName(vdcName, false)
		if adminVdc == nil || err != nil {
			vcd.infoCleanup(notFoundMsg, "vdc", vdcName)
		}
		fastProvisioningValue := false
		if entity.Name == "enable" {
			fastProvisioningValue = true
		}

		if adminVdc != nil && *adminVdc.AdminVdc.UsesFastProvisioning != fastProvisioningValue {
			adminVdc.AdminVdc.UsesFastProvisioning = &fastProvisioningValue
			_, err = adminVdc.Update()
			if err != nil {
				vcd.infoCleanup("updateLeftoverEntries: [INFO] revert back VDC fast provisioning value % s failed\n", entity.Name)
				return
			}
			vcd.infoCleanup("updateLeftoverEntries: [INFO] reverted back VDC fast provisioning value %s \n", entity.Name)
		} else {
			vcd.infoCleanup("updateLeftoverEntries: [INFO] VDC fast provisioning left as it is %s \n", entity.Name)
		}
		return
	case "orgLdapSettings":
		org, err := vcd.client.GetAdminOrgByName(entity.Parent)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] Clearing LDAP settings for Org '%s': %s",
				entity.Parent, err)
			return
		}

		ldapConfig, err := org.GetLdapConfiguration()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] Couldn't get LDAP settings for Org '%s': %s",
				entity.Parent, err)
			return
		}

		// This is done to avoid calling LdapDisable() if it has been unconfigured, due to bug with Org catalog publish settings
		if ldapConfig.OrgLdapMode != types.LdapModeNone {
			err = org.LdapDisable()
			if err != nil {
				vcd.infoCleanup("removeLeftoverEntries: [ERROR] Could not clear LDAP settings for Org '%s': %s",
					entity.Parent, err)
				vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
				return
			}
		}

		vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
		return

	case "vdcComputePolicy":
		policy, err := vcd.client.GetVdcComputePolicyV2ById(entity.Name)
		if policy == nil || err != nil {
			vcd.infoCleanup(notFoundMsg, "vdcComputePolicy", entity.Name)
			return
		}
		err = policy.Delete()
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return

	case "logicalVmGroup":
		logicalVmGroup, err := vcd.client.GetLogicalVmGroupById(entity.Name)
		if logicalVmGroup == nil || err != nil {
			vcd.infoCleanup(notFoundMsg, "logicalVmGroup", entity.Name)
			return
		}
		err = logicalVmGroup.Delete()
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return

	case "nsxtDhcpForwarder":
		edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(entity.Name)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		dhcpForwarder, err := edge.GetDhcpForwarder()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		if dhcpForwarder.Enabled == false && len(dhcpForwarder.DhcpServers) == 0 {
			vcd.infoCleanup(notFoundMsg, "dhcpForwarder", entity.Name)
			return
		}

		_, err = edge.UpdateDhcpForwarder(&types.NsxtEdgeGatewayDhcpForwarder{})
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "nsxtEdgeGatewayDns":
		edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(entity.Name)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		dns, err := edge.GetDnsConfig()
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		if dns.NsxtEdgeGatewayDns.Enabled == false && dns.NsxtEdgeGatewayDns.DefaultForwarderZone == nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}

		err = dns.Delete()
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return
	case "slaacProfile":
		edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(entity.Name)
		if err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] %s \n", err)
		}

		_, err = edge.UpdateSlaacProfile(&types.NsxtEdgeGatewaySlaacProfile{Enabled: false, Mode: "SLAAC"})
		if err != nil {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}

		vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		return

	default:
		// If we reach this point, we are trying to clean up an entity that
		// we aren't prepared for yet.
		fmt.Printf("removeLeftoverEntries: [ERROR] Unrecognized type %s for entity '%s'\n",
			entity.EntityType, entity.Name)
	}
}

func (vcd *TestVCD) TearDownSuite(check *C) {
	// We will try to remove every entity that has been registered into
	// CleanupEntityList. Entities that have already been cleaned up by their
	// functions will be ignored.
	for i, cleanupEntity := range cleanupEntityList {
		fmt.Printf("# %d of %d - ", i+1, len(cleanupEntityList))
		vcd.removeLeftoverEntities(cleanupEntity)
		removePersistentCleanupList()
	}
}

// Tests getloginurl with the endpoint given
// in the config file.
func (vcd *TestVCD) TestClient_getloginurl(check *C) {
	if os.Getenv("GOVCD_API_VERSION") != "" {
		check.Skip("custom API version is being used")
	}
	config, err := GetConfigStruct()
	if err != nil {
		check.Fatalf("err: %s", err)
	}
	client, err := GetTestVCDFromYaml(config)
	if err != nil {
		check.Fatalf("err: %s", err)
	}

	err = client.vcdloginurl()
	if err != nil {
		check.Fatalf("err: %s", err)
	}

	if client.sessionHREF.Path != "/cloudapi/1.0.0/sessions" {
		check.Fatalf("Getting LoginUrl failed, url: %s", client.sessionHREF.Path)
	}
}

// Tests Authenticate with the vcd credentials (or token) given in the config file
func (vcd *TestVCD) TestVCDClient_Authenticate(check *C) {
	config, err := GetConfigStruct()
	if err != nil {
		check.Fatalf("err: %s", err)
	}
	client, err := GetTestVCDFromYaml(config)
	if err != nil {
		check.Fatalf("err: %s", err)
	}
	apiToken := os.Getenv("VCD_API_TOKEN")
	if apiToken == "" {
		apiToken = config.Provider.ApiToken
	}
	if apiToken != "" {
		err = client.SetToken(config.Provider.SysOrg, ApiTokenHeader, apiToken)
	} else {
		token := os.Getenv("VCD_TOKEN")
		if token == "" {
			token = config.Provider.Token
		}
		if token != "" {
			err = client.SetToken(config.Provider.SysOrg, AuthorizationHeader, token)
		} else {
			err = client.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
		}
	}

	if err != nil {
		check.Fatalf("Error authenticating: %s", err)
	}
}

func (vcd *TestVCD) TestVCDClient_AuthenticateInvalidPassword(check *C) {
	config, err := GetConfigStruct()
	if err != nil {
		check.Fatalf("err: %s", err)
	}
	client, err := GetTestVCDFromYaml(config)
	if err != nil {
		check.Fatalf("error getting client structure: %s", err)
	}

	err = client.Authenticate(config.Provider.User, "INVALID-PASSWORD", config.Provider.SysOrg)
	if err == nil || !strings.Contains(err.Error(), "401") {
		check.Fatalf("expected error for invalid credentials")
	}
}

func (vcd *TestVCD) TestVCDClient_AuthenticateInvalidToken(check *C) {
	config, err := GetConfigStruct()
	if err != nil {
		check.Fatalf("err: %s", err)
	}
	client, err := GetTestVCDFromYaml(config)
	if err != nil {
		check.Fatalf("error getting client structure: %s", err)
	}

	err = client.SetToken(config.Provider.SysOrg, AuthorizationHeader, "invalid-token")
	if err == nil || !strings.Contains(err.Error(), "401") {
		check.Fatalf("expected error for invalid credentials")
	}
}

func (vcd *TestVCD) findFirstVm(vapp VApp) (types.Vm, string) {
	for _, vm := range vapp.VApp.Children.VM {
		if vm.Name != "" {
			return *vm, vm.Name
		}
	}
	return types.Vm{}, ""
}

func (vcd *TestVCD) findFirstVapp() VApp {
	client := vcd.client
	config := vcd.config
	org, err := client.GetOrgByName(config.VCD.Org)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	vdc, err := org.GetVDCByName(config.VCD.Vdc, false)
	if err != nil {
		fmt.Println(err)
		return VApp{}
	}
	wantedVapp := vcd.vapp.VApp.Name
	if wantedVapp == "" {
		// As no vApp is defined in config, we search for one randomly
		for _, res := range vdc.Vdc.ResourceEntities {
			for _, item := range res.ResourceEntity {
				if item.Type == "application/vnd.vmware.vcloud.vApp+xml" {
					wantedVapp = item.Name
					break
				}
			}
		}
	}
	vapp, err := vdc.GetVAppByName(wantedVapp, false)
	if err != nil {
		return VApp{}
	}
	return *vapp
}

// Test_NewRequestWitNotEncodedParamsWithApiVersion verifies that api version override works
func (vcd *TestVCD) Test_NewRequestWitNotEncodedParamsWithApiVersion(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())
	queryUlr := vcd.client.Client.VCDHREF
	queryUlr.Path += "/query"

	apiVersion, err := vcd.client.Client.MaxSupportedVersion()
	check.Assert(err, IsNil)

	req := vcd.client.Client.NewRequestWitNotEncodedParamsWithApiVersion(nil, map[string]string{"type": "media",
		"filter": "name==any"}, http.MethodGet, queryUlr, nil, apiVersion)

	check.Assert(req.Header.Get("User-Agent"), Equals, vcd.client.Client.UserAgent)

	resp, err := checkResp(vcd.client.Client.Http.Do(req))
	check.Assert(err, IsNil)

	check.Assert(resp.Header.Get("Content-Type"), Equals, types.MimeQueryRecords+";version="+apiVersion)

	bodyBytes, err := rewrapRespBodyNoopCloser(resp)
	check.Assert(err, IsNil)

	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(bodyBytes))
	debugShowResponse(resp, bodyBytes)

	// Repeats the call without API version change
	req = vcd.client.Client.NewRequestWitNotEncodedParams(nil, map[string]string{"type": "media",
		"filter": "name==any"}, http.MethodGet, queryUlr, nil)

	resp, err = checkResp(vcd.client.Client.Http.Do(req))
	check.Assert(err, IsNil)

	// Checks that the regularAPI version was not affected by the previous call
	check.Assert(resp.Header.Get("Content-Type"), Equals, types.MimeQueryRecords+";version="+vcd.client.Client.APIVersion)

	bodyBytes, err = rewrapRespBodyNoopCloser(resp)
	check.Assert(err, IsNil)
	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(bodyBytes))
	debugShowResponse(resp, bodyBytes)

	fmt.Printf("Test: %s run with api Version: %s\n", check.TestName(), apiVersion)
}

// setBoolFlag binds a flag to a boolean variable (passed as pointer)
// it also uses an optional environment variable that, if set, will
// update the variable before binding it to the flag.
func setBoolFlag(varPointer *bool, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		*varPointer = true
	}
	flag.BoolVar(varPointer, name, *varPointer, help)
}

// setTestEnv enables environment variables that are also used in non-test code
func setTestEnv() {
	if enableDebug {
		_ = os.Setenv("GOVCD_DEBUG", "1")
	}
	if debugShowRequestEnabled {
		_ = os.Setenv("GOVCD_SHOW_REQ", "1")
	}
	if debugShowResponseEnabled {
		_ = os.Setenv("GOVCD_SHOW_RESP", "1")
	}
}

func skipWhenMediaPathMissing(vcd *TestVCD, check *C) {
	if vcd.config.Media.MediaPath == "" {
		check.Skip("Skipping test because no iso path given")
	}
}

func skipNoNsxtConfiguration(vcd *TestVCD, check *C) {
	generalMessage := "Missing NSX-T config: "
	if vcd.config.VCD.NsxtProviderVdc.Name == "" {
		check.Skip(generalMessage + "No provider vdc specified")
	}
	if vcd.config.VCD.NsxtProviderVdc.NetworkPool == "" {
		check.Skip(generalMessage + "No network pool specified")
	}

	if vcd.config.VCD.Nsxt.Vdc == "" {
		check.Skip(generalMessage + "No NSX-T VDC specified")
	}

	if vcd.config.VCD.Nsxt.NsxtImportSegment == "" {
		check.Skip(generalMessage + "No NSX-T Unused segment (for imported Org VDC network) specified")
	}

	if vcd.config.VCD.NsxtProviderVdc.StorageProfile == "" {
		check.Skip(generalMessage + "No storage profile specified")
	}

	if vcd.config.VCD.Nsxt.Manager == "" {
		check.Skip(generalMessage + "No NSX-T manager specified")
	}

	if vcd.config.VCD.Nsxt.Tier0router == "" {
		check.Skip(generalMessage + "No NSX-T Tier-0 router specified")
	}

	if vcd.config.VCD.Nsxt.Tier0routerVrf == "" {
		check.Skip(generalMessage + "No VRF NSX-T Tier-0 router specified")
	}

	if vcd.config.VCD.Nsxt.EdgeGateway == "" {
		check.Skip(generalMessage + "No NSX-T Edge Gateway specified in configuration")
	}

	if vcd.config.VCD.Nsxt.IpDiscoveryProfile == "" ||
		vcd.config.VCD.Nsxt.MacDiscoveryProfile == "" ||
		vcd.config.VCD.Nsxt.SpoofGuardProfile == "" ||
		vcd.config.VCD.Nsxt.QosProfile == "" ||
		vcd.config.VCD.Nsxt.SegmentSecurityProfile == "" {
		check.Skip(generalMessage + "NSX-T Segment Profiles are not specified in configuration")
	}
}

func skipNoNsxtAlbConfiguration(vcd *TestVCD, check *C) {
	skipNoNsxtConfiguration(vcd, check)
	generalMessage := "Missing NSX-T ALB config: "

	if vcd.config.VCD.Nsxt.NsxtAlbControllerUrl == "" {
		check.Skip(generalMessage + "No NSX-T ALB Controller URL specified in configuration")
	}

	if vcd.config.VCD.Nsxt.NsxtAlbControllerUser == "" {
		check.Skip(generalMessage + "No NSX-T ALB Controller Name specified in configuration")
	}

	if vcd.config.VCD.Nsxt.NsxtAlbControllerPassword == "" {
		check.Skip(generalMessage + "No NSX-T ALB Controller Password specified in configuration")
	}

	if vcd.config.VCD.Nsxt.NsxtAlbImportableCloud == "" {
		check.Skip(generalMessage + "No NSX-T ALB Controller Importable Cloud Name")
	}
	if vcd.config.VCD.Nsxt.NsxtAlbServiceEngineGroup == "" {
		check.Skip(generalMessage + "No NSX-T ALB Service Engine Group name specified in configuration")
	}
}

// skipOpenApiEndpointTest is a helper to skip tests for particular unsupported OpenAPI endpoints
func skipOpenApiEndpointTest(vcd *TestVCD, check *C, endpoint string) {
	minimumRequiredApiVersion := endpointMinApiVersions[endpoint]

	constraint := ">= " + minimumRequiredApiVersion
	if !vcd.client.Client.APIVCDMaxVersionIs(constraint) {
		maxSupportedVersion, err := vcd.client.Client.MaxSupportedVersion()
		if err != nil {
			panic(fmt.Sprintf("Could not get maximum supported version: %s", err))
		}
		skipText := fmt.Sprintf("Skipping test because OpenAPI endpoint '%s' must satisfy API version constraint '%s'. Maximum supported version is %s",
			endpoint, constraint, maxSupportedVersion)
		check.Skip(skipText)
	}
}

// newUserConnection returns a connection for a given user
func newUserConnection(href, userName, password, orgName string, insecure bool) (*VCDClient, error) {
	u, err := url.ParseRequestURI(href)
	if err != nil {
		return nil, fmt.Errorf("[newUserConnection] unable to pass url: %s", err)
	}
	vcdClient := NewVCDClient(*u, insecure)
	err = vcdClient.Authenticate(userName, password, orgName)
	if err != nil {
		return nil, fmt.Errorf("[newUserConnection] unable to authenticate: %s", err)
	}
	return vcdClient, nil
}

// newOrgUserConnection creates a new Org User and returns a connection to it.
// Attention: Set the user to use only lowercase letters. If you put upper case letters the function fails on waiting
// because VCD creates the user with lowercase letters.
func newOrgUserConnection(adminOrg *AdminOrg, userName, password, href string, insecure bool) (*VCDClient, *OrgUser, error) {
	_, err := adminOrg.GetUserByName(userName, false)
	if err == nil {
		// user exists
		return nil, nil, fmt.Errorf("user %s already exists", userName)
	}
	_, err = adminOrg.CreateUserSimple(OrgUserConfiguration{
		Name:            userName,
		Password:        password,
		RoleName:        OrgUserRoleOrganizationAdministrator,
		ProviderType:    OrgUserProviderIntegrated,
		IsEnabled:       true,
		DeployedVmQuota: 0,
		StoredVmQuota:   0,
		FullName:        userName,
		Description:     "Test user created by newOrgUserConnection",
	})
	if err != nil {
		return nil, nil, err
	}

	AddToCleanupList(userName, "user", adminOrg.AdminOrg.Name, "newOrgUserConnection")

	_ = adminOrg.Refresh()
	newUser, err := adminOrg.GetUserByName(userName, false)
	if err != nil {
		return nil, nil, fmt.Errorf("[newOrgUserConnection] unable to retrieve newly created user: %s", err)
	}

	vcdClient, err := newUserConnection(href, userName, password, adminOrg.AdminOrg.Name, insecure)
	if err != nil {
		return nil, nil, fmt.Errorf("[newOrgUserConnection] error connecting new user: %s", err)
	}

	return vcdClient, newUser, nil
}

func (vcd *TestVCD) skipIfNotSysAdmin(check *C) {
	if !vcd.client.Client.IsSysAdmin {
		check.Skip(fmt.Sprintf("Skipping %s: requires system administrator privileges", check.TestName()))
	}
}

// retryOnError is a function that will attempt to execute function with signature `func() error`
// multiple times (until maxRetries) and waiting given retryInterval between tries. It will return
// original deletion error for troubleshooting.
func retryOnError(operation func() error, maxRetries int, retryInterval time.Duration) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		fmt.Printf("# retrying after %v (Attempt %d/%d)\n", retryInterval, attempt+1, maxRetries)
		fmt.Printf("# error was: %s", err)
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("exceeded maximum retries, final error: %s", err)
}
