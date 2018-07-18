package govcd

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"testing"
)

type TestConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Url      string `yaml:"url"`
	Orgname  string `yaml:"org"`
	Vdcname  string `yaml:"vdc"`
}

// tests the ability to authenticate user with given username, password, org, and vdc
func TestAuthenticate(t *testing.T) {
	g, err := GetConfigStruct()
	vcdClient, err := GetTestVCDFromYaml(g)
	if err != nil {
		t.Errorf("Error retrieving vcd client: %v ", err)
	}

	err = vcdClient.Authenticate(g.User, g.Password, "System")
	if err != nil {
		t.Errorf("Could not authenticate with user %s password %s url %s: %v", g.User, g.Password, g.Url, err)
		t.Errorf("orgname : %s, vdcname : %s", g.Orgname, g.Vdcname)
	}

}

func GetConfigStruct() (TestConfig, error) {
	g := TestConfig{}
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not read config file: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &g)
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not unmarshal yaml file: %v", err)
	}
	return g, nil
}

func GetTestVCDFromYaml(g TestConfig) (*VCDClient, error) {
	u, err := url.ParseRequestURI(g.Url)
	if err != nil {
		return &VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	vcdClient := NewVCDClient(*u, true)
	return vcdClient, nil
}
