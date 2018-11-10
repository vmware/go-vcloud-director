/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/util"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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
	TestCreateOrgVdcNetworkEGW    = "TestCreateOrgVdcNetworkEGW"
	TestCreateOrgVdcNetworkIso    = "TestCreateOrgVdcNetworkIso"
	TestCreateOrgVdcNetworkDirect = "TestCreateOrgVdcNetworkDirect"
)

// Struct to get info from a config yaml file that the user
// specifies
type TestConfig struct {
	Provider struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Url      string `yaml:"url"`
		SysOrg   string `yaml:"sysOrg"`
	}
	VCD struct {
		Org         string `yaml:"org"`
		Vdc         string `yaml:"vdc"`
		ProviderVdc struct {
			Id               string `yaml:"id"`
			StorageProfileId string `yaml:"storage_profile_id"`
		} `yaml:"provider_vdc"`
		Catalog struct {
			Name                   string `yaml:"name,omitempty"`
			Description            string `yaml:"description,omitempty"`
			Catalogitem            string `yaml:"catalogItem,omitempty"`
			CatalogItemDescription string `yaml:"catalogItemDescription,omitempty"`
		} `yaml:"catalog"`
		Network        string `yaml:"network,omitempty"`
		StorageProfile struct {
			SP1 string `yaml:"storageProfile1"`
			SP2 string `yaml:"storageProfile2,omitempty"`
		} `yaml:"storageProfile"`
		ExternalIp      string `yaml:"externalIp,omitempty"`
		InternalIp      string `yaml:"internalIp,omitempty"`
		EdgeGateway     string `yaml:"edgeGateway,omitempty"`
		ExternalNetwork string `yaml:"externalNetwork,omitempty"`
	} `yaml:"vcd"`
	Logging struct {
		Enabled         bool   `yaml:"enabled,omitempty"`
		LogFileName     string `yaml:"logFileName,omitempty"`
		LogHttpRequest  bool   `yaml:"logHttpRequest,omitempty"`
		LogHttpResponse bool   `yaml:"logHttpResponse,omitempty"`
		VerboseCleanup  bool   `yaml:"verboseCleanup,omitempty"`
	} `yaml:"logging"`
	OVA struct {
		OVAPath        string `yaml:"ovaPath,omitempty"`
		OVAChunkedPath string `yaml:"ovaChunkedPath,omitempty"`
	} `yaml:"ova"`
}

// Test struct for vcloud-director.
// Test functions use the struct to get
// an org, vdc, vapp, and client to run
// tests on
type TestVCD struct {
	client         *VCDClient
	org            Org
	vdc            Vdc
	vapp           VApp
	config         TestConfig
	skipVappTests  bool
	skipAdminTests bool
}

// Cleanup entity structure used by the tear-down procedure
// at the end of the tests to remove leftover entities
type CleanupEntity struct {
	Name       string
	EntityType string
	Parent     string
	CreatedBy  string
}

// Internally used by the test suite to run tests based on TestVCD structures
var _ = Suite(&TestVCD{})

// The list holding the entities to be examined and eventually removed
// at the end of the tests
var cleanupEntityList []CleanupEntity

// Use this value to run a specific test that does not need a pre-created vApp.
var skipVappCreation bool = os.Getenv("GOVCD_SKIP_VAPP_CREATION") != ""

// Adds an entity to the cleanup list.
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
}

// Users use the environmental variable GOVCD_CONFIG as
// a config file for testing. Otherwise the default is govcd_test_config.yaml
// in the current directory. Throws an error if it cannot find your
// yaml file or if it cannot read it.
func GetConfigStruct() (TestConfig, error) {
	config := os.Getenv("GOVCD_CONFIG")
	config_struct := TestConfig{}
	if config == "" {
		// Finds the current directory, through the path of this running test
		_, current_filename, _, _ := runtime.Caller(0)
		current_directory := filepath.Dir(current_filename)
		config = current_directory + "/govcd_test_config.yaml"
	}
	// Looks if the configuration file exists before attempting to read it
	_, err := os.Stat(config)
	if os.IsNotExist(err) {
		return TestConfig{}, fmt.Errorf("Configuration file %s not found: %s", config, err)
	}
	yamlFile, err := ioutil.ReadFile(config)
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not read config file %s: %v", config, err)
	}
	err = yaml.Unmarshal(yamlFile, &config_struct)
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not unmarshal yaml file: %v", err)
	}
	return config_struct, nil
}

// Creates a VCDClient based on the endpoint given in the TestConfig argument.
// TestConfig struct can be obtained by calling GetConfigStruct. Throws an error
// if endpoint given is not a valid url.
func GetTestVCDFromYaml(testConfig TestConfig) (*VCDClient, error) {
	configUrl, err := url.ParseRequestURI(testConfig.Provider.Url)
	if err != nil {
		return &VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	vcdClient := NewVCDClient(*configUrl, true)
	return vcdClient, nil
}

// Necessary to enable the suite tests with TestVCD
func Test(t *testing.T) { TestingT(t) }

// Sets the org, vdc, vapp, and vcdClient for a
// TestVCD struct. An error is thrown if something goes wrong
// getting config file, creating vcd, during authentication, or
// when creating a new vapp. If this method panics, no test
// case that uses the TestVCD struct is run.
func (vcd *TestVCD) SetUpSuite(check *C) {
	config, err := GetConfigStruct()
	if config == (TestConfig{}) || err != nil {
		panic(err)
	}
	vcd.config = config

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
	} else {
		util.EnableLogging = false
	}
	util.SetLog()
	vcdClient, err := GetTestVCDFromYaml(config)
	if vcdClient == nil || err != nil {
		panic(err)
	}
	vcd.client = vcdClient
	if config.Provider.SysOrg != "System" {
		fmt.Printf("Skipping OrgAdmin tests\n")
		vcd.skipAdminTests = true
	}
	// org and vdc are the test org and vdc that is used in all other test cases
	err = vcd.client.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
	if err != nil {
		panic(err)
	}
	// set org
	vcd.org, err = GetOrgByName(vcd.client, config.VCD.Org)
	if err != nil || vcd.org == (Org{}) {
		panic(err)
	}
	// set vdc
	vcd.vdc, err = vcd.org.GetVdcByName(config.VCD.Vdc)
	if err != nil || vcd.vdc == (Vdc{}) {
		panic(err)
	}
	// creates a new VApp for vapp tests
	if !skipVappCreation && config.VCD.Network != "" && config.VCD.StorageProfile.SP1 != "" &&
		config.VCD.Catalog.Name != "" && config.VCD.Catalog.Catalogitem != "" {
		vcd.vapp, err = vcd.createTestVapp(TestSetUpSuite)
		// If no vApp is created, we skip all vApp tests
		if vcd.vapp == (VApp{}) || err != nil {
			fmt.Printf("%v", err)
			vcd.skipVappTests = true
		}
		// After a successful creation, the vAPp is added to the cleanup list
		AddToCleanupList(TestSetUpSuite, "vapp", "", "SetUpSuite")
	} else {
		vcd.skipVappTests = true
		fmt.Printf("Skipping all vapp tests because one of the following wasn't given: Network, StorageProfile, Catalog, Catalogitem")
	}
}

// Shows the detail of cleanup operations only if the relevant verbosity
// has been enabled
func (vcd *TestVCD) infoCleanup(format string, args ...interface{}) {
	if vcd.config.Logging.VerboseCleanup {
		fmt.Printf(format, args...)
	}
}

// Gets the two components of a "parent" string, as passed to AddToCleanupList
func splitParent(parent string, separator string) (first, second string) {
	strList := strings.Split(parent, separator)
	if len(strList) != 2 {
		return "", ""
	}
	first = strList[0]
	second = strList[1]
	return
}

// Removes leftover entities that may still exist after failed tests
// or the ones that were explicitly created for several tests and
// were relying on this procedure to clean up at the end.
func (vcd *TestVCD) removeLeftoverEntities(entity CleanupEntity) {
	var introMsg string = "removeLeftoverEntries: [INFO] Attempting cleanup of %s '%s' instantiated by %s\n"
	var notFoundMsg string = "removeLeftoverEntries: [INFO] No action for %s '%s'\n"
	var removedMsg string = "removeLeftoverEntries: [INFO] Removed %s '%s' created by %s\n"
	var notDeletedMsg string = "removeLeftoverEntries: [ERROR] Error deleting %s '%s': %s\n"
	var splitParentNotFound string = "removeLeftopverEntries: [ERROR] missing parent info (%s). The parent fields must be defined with a separator '|'\n"
	// NOTE: this is a cleanup function that should continue even if errors are found.
	// For this reason, the [ERROR] messages won't be followed by a program termination
	vcd.infoCleanup(introMsg, entity.EntityType, entity.Name, entity.CreatedBy)
	switch entity.EntityType {
	case "vapp":
		vapp, err := vcd.vdc.FindVAppByName(entity.Name)
		if vapp == (VApp{}) || err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		task, _ := vapp.Undeploy()
		_ = task.WaitTaskCompletion()
		task, err = vapp.Delete()
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
		org, err := GetAdminOrgByName(vcd.client, entity.Parent)
		if org == (AdminOrg{}) || err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [INFO] organization '%s' not found\n", entity.Parent)
			return
		}
		catalog, err := org.FindAdminCatalog(entity.Name)
		if catalog == (AdminCatalog{}) || err != nil {
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
		org, err := GetAdminOrgByName(vcd.client, entity.Name)
		if org == (AdminOrg{}) || err != nil {
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
	case "catalogItem":
		if entity.Parent == "" {
			vcd.infoCleanup("removeLeftoverEntries: [ERROR] No Org provided for catalogItem '%s'\n", strings.Split(entity.Parent, "|")[0])
			return
		}
		org, err := GetAdminOrgByName(vcd.client, strings.Split(entity.Parent, "|")[0])
		if org == (AdminOrg{}) || err != nil {
			vcd.infoCleanup("removeLeftoverEntries: [INFO] organization '%s' not found\n", entity.Parent)
			return
		}
		catalog, err := org.FindCatalog(strings.Split(entity.Parent, "|")[1])
		if catalog == (Catalog{}) || err != nil {
			vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
			return
		}
		for _, catalogItems := range catalog.Catalog.CatalogItems {
			for _, catalogItem := range catalogItems.CatalogItem {
				if catalogItem.Name == entity.Name {
					catalogItemApi, err := catalog.FindCatalogItem(catalogItem.Name)
					if catalogItemApi == (CatalogItem{}) || err != nil {
						vcd.infoCleanup(notFoundMsg, entity.EntityType, entity.Name)
						return
					}
					err = catalogItemApi.Delete()
					vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
					if err != nil {
						vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
					}
				}
			}
		}
		return
	case "edgegateway":
		//TODO: find an easy way of undoing edge GW customization
		return
	case "network":
		orgName, vdcName := splitParent(entity.Parent, "|")
		if orgName == "" || vdcName == "" {
			vcd.infoCleanup(splitParentNotFound, entity.Parent)
			return
		}
		org, err := GetAdminOrgByName(vcd.client, orgName)
		if org == (AdminOrg{}) || err != nil {
			vcd.infoCleanup(notFoundMsg, "org", orgName)
			return
		}
		vdc, err := org.GetVdcByName(vdcName)
		if vdc == (Vdc{}) || err != nil {
			vcd.infoCleanup(notFoundMsg, "vdc", vdcName)
			return
		}
		err = RemoveOrgVdcNetworkIfExists(vdc, entity.Name)
		if err == nil {
			vcd.infoCleanup(removedMsg, entity.EntityType, entity.Name, entity.CreatedBy)
		} else {
			vcd.infoCleanup(notDeletedMsg, entity.EntityType, entity.Name, err)
		}
		return
	case "vdc":
		// nothing so far
		return
	case "vm":
		// nothing so far
		return
	default:
		// If we reach this point, we are trying to clean up an entity that
		// we aren't prepared for yet.
		fmt.Printf("removeLeftoverEntries: [ERROR] Unrecognized type %s for entity '%s'\n", entity.EntityType, entity.Name)
	}
}

func (vcd *TestVCD) TearDownSuite(check *C) {
	// We will try to remove every entity that has been registered into
	// CleanupEntityList. Entities that have already been cleaned up by their
	// functions will be ignored.
	for _, cleanupEntity := range cleanupEntityList {
		vcd.removeLeftoverEntities(cleanupEntity)
	}
}

// Tests getloginurl with the endpoint given
// in the config file.
func TestClient_getloginurl(t *testing.T) {
	config, err := GetConfigStruct()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	client, err := GetTestVCDFromYaml(config)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	err = client.vcdloginurl()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if client.sessionHREF.Path != "/api/sessions" {
		t.Fatalf("Getting LoginUrl failed, url: %s", client.sessionHREF.Path)
	}
}

// Tests Authenticate with the vcd credentials given in the config file
func TestVCDClient_Authenticate(t *testing.T) {
	config, err := GetConfigStruct()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	client, err := GetTestVCDFromYaml(config)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	err = client.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
	if err != nil {
		t.Fatalf("Error authenticating: %v", err)
	}
}
