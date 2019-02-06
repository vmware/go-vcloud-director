package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/go-vcloud-director/govcd"
)

type Config struct {
	User     string
	Password string
	Org      string
	Href     string
	VDC      string
	Insecure bool
}

func (c *Config) Client() (*govcd.VCDClient, error) {
	u, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("Unable to pass url: %s", err)
	}

	vcdclient := govcd.NewVCDClient(*u, c.Insecure)
	err = vcdclient.Authenticate(c.User, c.Password, c.Org)
	if err != nil {
		return nil, fmt.Errorf("Unable to authenticate: %s", err)
	}
	return vcdclient, nil
}

func main() {
	config := Config{
		User:     "myuser",
		Password: "password",
		Org:      "MyOrg",
		Href:     "https://vcd-host/api",
		VDC:      "My-VDC",
	}

	client, err := config.Client() // We now have a client
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	org, errOrg := govcd.GetOrgByName(client, config.Org)
	if errOrg != nil {
		fmt.Println("GetOrgByName: ", errOrg)
		os.Exit(1)
	}
	fmt.Printf("Org URL: %s\n", org.Org.HREF)

	vdc, errVdc := org.GetVdcByName(config.VDC)
	if errVdc != nil {
		fmt.Println("GetOrgByName: ", errVdc)
		os.Exit(1)
	}

	fmt.Println("VDC: ", vdc)
}
