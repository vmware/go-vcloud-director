package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/util"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	neturl "net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
		Org     string `yaml:"org"`
		Vdc     string `yaml:"vdc"`
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
		ExternalIp  string `yaml:"externalIp,omitempty"`
		InternalIp  string `yaml:"internalIp,omitempty"`
		EdgeGateway string `yaml:"edgeGateway,omitempty"`
	} `yaml:"vcd"`
	Logging struct {
		Enabled         bool   `yaml:"enabled,omitempty"`
		LogFileName     string `yaml:"logFileName,omitempty"`
		LogHttpRequest  bool   `yaml:"logHttpRequest,omitempty"`
		LogHttpResponse bool   `yaml:"logHttpResponse,omitempty"`
	} `yaml:"logging"`
}

// Test struct for vcloud-director.
// Test functions use the struct to get
// an org, vdc, vapp, and client to run
// tests on
type TestVCD struct {
	client        *VCDClient
	org           Org
	vdc           Vdc
	vapp          VApp
	config        TestConfig
	skipVappTests bool
}

var _ = Suite(&TestVCD{})

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
func GetTestVCDFromYaml(g TestConfig) (*VCDClient, error) {
	url, err := neturl.ParseRequestURI(g.Provider.Url)
	if err != nil {
		return &VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	vcdClient := NewVCDClient(*url, true)
	return vcdClient, nil
}

// Necessary to enable the suite tests with TestVCD
func Test(t *testing.T) { TestingT(t) }

// Sets the org, vdc, vapp, and vcdClient for a
// TestVCD struct. An error is thrown during if something goes wrong
// getting config file, creating vcd, during authentication, or
// when creating a new vapp. If this method panics, no test
// case that uses the TestVCD struct is run.
func (vcd *TestVCD) SetUpSuite(check *C) {
	config, err := GetConfigStruct()
	if err != nil {
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
	if err != nil {
		panic(err)
	}
	vcd.client = vcdClient
	// org and vdc are the test org and vdc that is used in all other test cases
	err = vcd.client.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
	if err != nil {
		panic(err)
	}
	// set org
	vcd.org, err = GetOrgByName(vcd.client, config.VCD.Org)
	if err != nil {
		panic(err)
	}
	// set vdc
	vcd.vdc, err = vcd.org.GetVdcByName(config.VCD.Vdc)
	if err != nil {
		panic(err)
	}
	// creates a new VApp for vapp tests
	if config.VCD.Network != "" && config.VCD.StorageProfile.SP1 != "" &&
		config.VCD.Catalog.Name != "" && config.VCD.Catalog.Catalogitem != "" {
		vcd.vapp, err = vcd.createTestVapp("go-vapp-tests")
		if err != nil {
			fmt.Printf("%v", err)
			vcd.skipVappTests = true
		}
	} else {
		vcd.skipVappTests = true
		fmt.Printf("Skipping all vapp tests because one of the following wasn't given: Network, StorageProfile, Catalog, Catalogitem")
	}

}

func (vcd *TestVCD) TearDownSuite(check *C) {
	if vcd.skipVappTests {
		check.Skip("Vapp tests skipped, no vapp to be deleted")
	}
	err := vcd.vapp.Refresh()
	if err != nil {
		panic(err)
	}
	task, _ := vcd.vapp.Undeploy()
	_ = task.WaitTaskCompletion()
	task, err = vcd.vapp.Delete()
	if err != nil {
		panic(err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		panic(err)
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
