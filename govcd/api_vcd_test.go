package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/testutil"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

// Struct to get info from a config yaml file that the user
// specifies
type TestConfig struct {
	Provider struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Url      string `yaml:"url"`
	}
	VCD struct {
		Org     string `yaml:"org"`
		Vdc     string `yaml:"vdc"`
		Catalog struct {
			Name                   string `yaml:"name,omitempty"`
			Description            string `yaml:"description,omitempty"`
			Catalogitem            string `yaml:"catalogitem,omitempty"`
			CatalogItemDescription string `yaml:"catalogitemdescription,omitempty"`
		}
		Network        string `yaml:"network,omitempty"`
		StorageProfile struct {
			SP1 string `yaml:"storageprofile1,omitempty"`
			SP2 string `yaml:"storageprofile2,omitempty"`
		}
		VApp string `yaml:"vapp,omitempty"`
	}
}

// Test struct for vcloud-director.
// Test functions use the struct to get
// an org, vdc, vapp, and client to run
// tests on
type TestVCD struct {
	client *VCDClient
	org    Org
	vdc    Vdc
	vapp   VApp
	config TestConfig
}

var testServer = testutil.NewHTTPServer()

var vcdu_api, _ = url.Parse("http://localhost:4444/api")
var vcdu_v, _ = url.Parse("http://localhost:4444/api/versions")
var vcdu_s, _ = url.Parse("http://localhost:4444/api/vchs/services")

var vcdauthheader = map[string]string{"x-vcloud-authorization": "012345678901234567890123456789"}

var _ = Suite(&TestVCD{})

// Users use the environmental variable VCLOUD_CONFIG as
// a config file for testing. Otherwise the default is config.yaml
// in the home directory. Throws an error if it cannot find your
// yaml file or if it cannot read it.
func GetConfigStruct() (TestConfig, error) {
	config := os.Getenv("VCLOUD_CONFIG")
	config_struct := TestConfig{}
	if config == "" {
		config = os.Getenv("HOME") + "/config.yaml"
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
	u, err := url.ParseRequestURI(g.Provider.Url)
	if err != nil {
		return &VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	vcdClient := NewVCDClient(*u, true)
	return vcdClient, nil
}

// Neccessary to enable the suite tests with TestVCD
func Test(t *testing.T) { TestingT(t) }

// Sets the org, vdc, vapp, and vcdClient for a
// TestVCD struct. An error is thrown during if something goes wrong
// getting config file, creating vcd, during authentication, or
// when creating a new vapp. If this method panics, no test
// case that uses the TestVCD struct is run.
func (vcd *TestVCD) SetUpSuite(test *C) {
	// this will be removed once all tests are converted to
	// a real vcd
	testServer.Start()

	config, err := GetConfigStruct()
	if err != nil {
		panic(err)
	}
	vcd.config = config

	vcdClient, err := GetTestVCDFromYaml(config)
	if err != nil {
		panic(err)
	}
	vcd.client = vcdClient
	// org and vdc are the test org and vdc that is used in all other test cases
	err = vcd.client.Authenticate(config.Provider.User, config.Provider.Password, config.VCD.Org)
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
	vcd.vapp = *NewVApp(&vcd.client.Client)
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
	err = client.Authenticate(config.Provider.User, config.Provider.Password, config.VCD.Org)
	if err != nil {
		t.Fatalf("Error authenticating: %v", err)
	}
}
