// +build api functional catalog vapp gateway network org query extnetwork task vm vdc system disk lb lbAppRule lbAppProfile lbServerPool lbServiceMonitor lbVirtualServer user ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

func init() {
	testingTags["api"] = "api_vcd_test.go"
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
	TestCreateOrgVdcNetworkEGW    = "TestCreateOrgVdcNetworkEGW"
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
	TestVMAttachOrDetachDisk      = "TestVMAttachOrDetachDisk"
	TestVMAttachDisk              = "TestVMAttachDisk"
	TestVMDetachDisk              = "TestVMDetachDisk"
	TestCreateExternalNetwork     = "TestCreateExternalNetwork"
	TestDeleteExternalNetwork     = "TestDeleteExternalNetwork"
	TestLbServiceMonitor          = "TestLbServiceMonitor"
	TestLbServerPool              = "TestLbServerPool"
	TestLbAppProfile              = "TestLbAppProfile"
	TestLbAppRule                 = "TestLbAppRule"
	TestLbVirtualServer           = "TestLbVirtualServer"
	TestLb                        = "TestLb"
	TestNsxvSnatRule              = "TestNsxvSnatRule"
	TestNsxvDnatRule              = "TestNsxvDnatRule"
)

const (
	TestRequiresSysAdminPrivileges = "Test %s requires system administrator privileges"
)

// Struct to get info from a config yaml file that the user
// specifies
type TestConfig struct {
	Provider struct {
		User            string `yaml:"user"`
		Password        string `yaml:"password"`
		Url             string `yaml:"url"`
		SysOrg          string `yaml:"sysOrg"`
		MaxRetryTimeout int    `yaml:"maxRetryTimeout,omitempty"`
		HttpTimeout     int64  `yaml:"httpTimeout,omitempty"`
	}
	VCD struct {
		Org         string `yaml:"org"`
		Vdc         string `yaml:"vdc"`
		ProviderVdc struct {
			Name           string `yaml:"name"`
			StorageProfile string `yaml:"storage_profile"`
			NetworkPool    string `yaml:"network_pool"`
		} `yaml:"provider_vdc"`
		Catalog struct {
			Name                   string `yaml:"name,omitempty"`
			Description            string `yaml:"description,omitempty"`
			CatalogItem            string `yaml:"catalogItem,omitempty"`
			CatalogItemDescription string `yaml:"catalogItemDescription,omitempty"`
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
		Disk                         struct {
			Size          int64 `yaml:"size,omitempty"`
			SizeForUpdate int64 `yaml:"sizeForUpdate,omitempty"`
		}
	} `yaml:"vcd"`
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
		OVAPath        string `yaml:"ovaPath,omitempty"`
		OVAChunkedPath string `yaml:"ovaChunkedPath,omitempty"`
	} `yaml:"ova"`
	Media struct {
		MediaPath       string `yaml:"mediaPath,omitempty"`
		Media           string `yaml:"mediaName,omitempty"`
		PhotonOsOvaPath string `yaml:"photonOsOvaPath,omitempty"`
	} `yaml:"media"`
}

// Test struct for vcloud-director.
// Test functions use the struct to get
// an org, vdc, vapp, and client to run
// tests on
type TestVCD struct {
	client         *VCDClient
	org            *Org
	vdc            *Vdc
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
	listText, err := ioutil.ReadFile(persistentCleanupListFile)
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
	file, err := os.Create(persistentCleanupListFile)
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
	err := writeCleanupList(cleanupEntityList)
	if err != nil {
		fmt.Printf("################ error writing cleanup list %s\n", err)
	}
}

// Prepend an entity to the cleanup list.
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

// Users use the environmental variable GOVCD_CONFIG as
// a config file for testing. Otherwise the default is govcd_test_config.yaml
// in the current directory. Throws an error if it cannot find your
// yaml file or if it cannot read it.
func GetConfigStruct() (TestConfig, error) {
	config := os.Getenv("GOVCD_CONFIG")
	configStruct := TestConfig{}
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
	yamlFile, err := ioutil.ReadFile(config)
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not read config file %s: %v", config, err)
	}
	err = yaml.Unmarshal(yamlFile, &configStruct)
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not unmarshal yaml file: %v", err)
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
	// org and vdc are the test org and vdc that is used in all other test cases
	err = vcd.client.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Running on vCD %s\nas user %s@%s\n", vcd.config.Provider.Url,
		vcd.config.Provider.User, vcd.config.Provider.SysOrg)
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

	// If neither the vApp or VM tags are set, we also skip the
	// creation of the default vApp
	if !isTagSet("vapp") && !isTagSet("vm") {
		// vcd.skipVappTests = true
		skipVappCreation = true
	}

	// Gets the persistent cleanup list from file, if exists.
	cleanupList, err := readCleanupList()
	if len(cleanupList) > 0 && err == nil {
		if os.Getenv("GOVCD_IGNORE_CLEANUP_FILE") == "" {
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
		vcd.vapp, err = vcd.createTestVapp(TestSetUpSuite)
		// If no vApp is created, we skip all vApp tests
		if err != nil {
			fmt.Printf("%v\n", err)
			panic("Creation failed - Bailing out")
		}
		if vcd.vapp == (VApp{}) {
			fmt.Printf("Creation of vApp %s failed unexpectedly. No error was reported, but vApp is empty\n", TestSetUpSuite)
			panic("initial vApp is empty - bailing out")
		}
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

// Gets the two or three components of a "parent" string, as passed to AddToCleanupList
// If the number of split strings is not 2 or 3 it return 3 empty strings
// Example input parent: my-org|my-vdc|my-edge-gw, separator: |
// Output output: first: my-org, second: my-vdc, third: my-edge-gw
func splitParent(parent string, separator string) (first, second, third string) {
	strList := strings.Split(parent, separator)
	if len(strList) < 2 || len(strList) > 3 {
		return "", "", ""
	}
	first = strList[0]
	second = strList[1]

	if len(strList) == 3 {
		third = strList[2]
	}

	return
}

func getOrgVdcEdgeByNames(vcd *TestVCD, orgName, vdcName, edgeName string) (*Org, *Vdc, *EdgeGateway, error) {
	if orgName == "" || vdcName == "" || edgeName == "" {
		return nil, nil, nil, fmt.Errorf("orgName, vdcName, edgeName cant be empty")
	}

	org, _ := vcd.client.GetOrgByName(orgName)
	if org == nil {
		vcd.infoCleanup("could not find org '%s'", orgName)
		return nil, nil, nil, fmt.Errorf("can't find org")
	}
	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		vcd.infoCleanup("could not find vdc '%s'", vdcName)
		return nil, nil, nil, fmt.Errorf("can't find org")
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
	case "vapp":
		vapp, err := vcd.vdc.GetVAppByName(entity.Name, true)
		if err != nil {
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
		for _, catalogItems := range catalog.Catalog.CatalogItems {
			for _, catalogItem := range catalogItems.CatalogItem {
				if catalogItem.Name == entity.Name {
					catalogItemApi, err := catalog.GetCatalogItemByName(catalogItem.Name, false)
					if catalogItemApi == nil || err != nil {
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
	case "vm":
		// nothing so far
		return
	case "disk":
		// Find disk by href rather than find disk by name, because disk name can be duplicated in VDC,
		// so the unique href is required for finding the disk.
		// [0] = disk's entity name, [1] = disk href
		disk, err := vcd.vdc.FindDiskByHREF(strings.Split(entity.Name, "|")[1])
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
		fmt.Printf("# %d ", i+1)
		vcd.removeLeftoverEntities(cleanupEntity)
		removePersistentCleanupList()
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

func (vcd *TestVCD) createTestVapp(name string) (VApp, error) {
	// Populate OrgVDCNetwork
	networks := []*types.OrgVDCNetwork{}
	net, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding network : %v", err)
	}
	networks = append(networks, net.OrgVDCNetwork)
	// Populate Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil || cat == nil {
		return VApp{}, fmt.Errorf("error finding catalog : %v", err)
	}
	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding catalog item : %v", err)
	}
	// Get VAppTemplate
	vAppTemplate, err := catitem.GetVAppTemplate()
	if err != nil {
		return VApp{}, fmt.Errorf("error finding vapptemplate : %v", err)
	}
	// Get StorageProfileReference
	storageProfileRef, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	if err != nil {
		return VApp{}, fmt.Errorf("error finding storage profile: %v", err)
	}
	// Compose VApp
	task, err := vcd.vdc.ComposeVApp(networks, vAppTemplate, storageProfileRef, name, "description", true)
	if err != nil {
		return VApp{}, fmt.Errorf("error composing vapp: %v", err)
	}
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(name, "vapp", "", "createTestVapp")
	err = task.WaitTaskCompletion()
	if err != nil {
		return VApp{}, fmt.Errorf("error composing vapp: %v", err)
	}
	// Get VApp
	vapp, err := vcd.vdc.GetVAppByName(name, true)
	if err != nil {
		return VApp{}, fmt.Errorf("error getting vapp: %v", err)
	}

	err = vapp.BlockWhileStatus("UNRESOLVED", vapp.client.MaxRetryTimeout)
	if err != nil {
		return VApp{}, fmt.Errorf("error waiting for created test vApp to have working state: %s", err)
	}

	return *vapp, err
}

func Test_splitParent(t *testing.T) {
	type args struct {
		parent    string
		separator string
	}
	tests := []struct {
		name       string
		args       args
		wantFirst  string
		wantSecond string
		wantThird  string
	}{
		{
			name:       "Empty",
			args:       args{parent: "", separator: "|"},
			wantFirst:  "",
			wantSecond: "",
			wantThird:  "",
		},
		{
			name:       "One",
			wantFirst:  "",
			wantSecond: "",
			wantThird:  "",
		},
		{
			name:       "Two",
			args:       args{parent: "first|second", separator: "|"},
			wantFirst:  "first",
			wantSecond: "second",
			wantThird:  "",
		},
		{
			name:       "Three",
			args:       args{parent: "first|second|third", separator: "|"},
			wantFirst:  "first",
			wantSecond: "second",
			wantThird:  "third",
		},
		{
			name:       "Four",
			args:       args{parent: "first|second|third|fourth", separator: "|"},
			wantFirst:  "",
			wantSecond: "",
			wantThird:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFirst, gotSecond, gotThird := splitParent(tt.args.parent, tt.args.separator)
			if gotFirst != tt.wantFirst {
				t.Errorf("splitParent() gotFirst = %v, want %v", gotFirst, tt.wantFirst)
			}
			if gotSecond != tt.wantSecond {
				t.Errorf("splitParent() gotSecond = %v, want %v", gotSecond, tt.wantSecond)
			}
			if gotThird != tt.wantThird {
				t.Errorf("splitParent() gotThird = %v, want %v", gotThird, tt.wantThird)
			}
		})
	}
}
