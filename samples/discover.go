/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package main

/* This sample program shows how to list organizations, vDCs, vApps, and catalog items
using a vCD client

Usage:

BEFORE USING, copy the file sample_config.json to config.json and fill it with your vCD credentials
(or do the same with sample_config.yaml)
JSON is a subset of YAML. The YAML interpreter will take care of either configuration file
(http://ghodss.com/2014/the-right-way-to-handle-yaml-in-golang/)

(1) On the command line
  cd samples
	go run discover.go ./config_file.json
Or
	go run discover.go ./config_file.yaml

(2) In GoLand
	* In the menu "Run" / "edit configurations", add the full path to your JSON or YAML file into "Program arguments"
  * From the menu "Run", choose "Run 'go build discover.go'"

================
Troubleshooting.
================
If there are problems during the configuration load, you can use the SAMPLES_DEBUG environment variable to show
what was read from the file and how it was interpreted.

On the command line:

   SAMPLES_DEBUG=1 go run discover.go ./config_file.json

In Goland:

	* In the menu "Run" / "edit configurations", clock on "Environment", then add a variable with the "+" button: write
"SAMPLES_DEBUG" under "name" and "1" under "value."

*/

import (
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/go-vcloud-director/govcd"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Org      string `json:"org"`
	Href     string `json:"href"`
	VDC      string `json:"vdc"`
	Insecure bool   `json:"insecure"`
}

// Checks that a configuration structure is complete
func check_configuration(conf Config) {
	will_exit := false
	abort := func(s string) {
		fmt.Printf("configuration field '%s' empty or missing\n", s)
		will_exit = true
	}
	if conf.User == "" {
		abort("user")
	}
	if conf.Password == "" {
		abort("password")
	}
	if conf.Org == "" {
		abort("org")
	}
	if conf.Href == "" || conf.Href == "https://YOUR_VCD_IP/api" {
		abort("href")
	}
	if conf.VDC == "" {
		abort("vdc")
	}
	if will_exit {
		os.Exit(1)
	}
}

// Retrieves the configuration from a Json or Yaml file
func getConfig(config_file string) Config {
	var configuration Config
	buffer, err := ioutil.ReadFile(config_file)
	if err != nil {
		fmt.Printf("Configuration file %s not found\n%s\n", config_file, err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(buffer, &configuration)
	if err != nil {
		fmt.Printf("Error retrieving configuration from file %s\n%s\n", config_file, err)
		os.Exit(1)
	}
	check_configuration(configuration)

	// If something goes wrong, rerun the program after setting
	// the environment variable SAMPLES_DEBUG, and you can check how the
	// configuration was read
	if os.Getenv("SAMPLES_DEBUG") != "" {
		fmt.Printf("configuration text: %s\n", buffer)
		fmt.Printf("configuration rec: %#v\n", configuration)
		new_conf, _ := yaml.Marshal(configuration)
		fmt.Printf("YAML configuration: \n%s\n", new_conf)
	}
	return configuration
}

// Creates a vCD client
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
	// This program requires a configuration file, which must be passed
	// on the command line
	if len(os.Args) < 2 {
		fmt.Printf("Syntax: discover config_file.json\n")
		os.Exit(1)
	}

	// Reads the configuration file
	config := getConfig(os.Args[1])

	// Instantiates the client
	client, err := config.Client() // We now have a client
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	org, err := govcd.GetOrgByName(client, config.Org)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	vdc, err := org.GetVdcByName(config.VDC)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Organization items\n")
	fmt.Printf("Organization '%s' URL: %s\n", config.Org, org.Org.HREF)

	catalog_name := ""
	for N, item := range org.Org.Link {
		fmt.Printf("%3d %-40s %s\n", N, item.Name, item.Type)
		// Retrieve the first catalog name for further usage
		if item.Type == "application/vnd.vmware.vcloud.catalog+xml" && catalog_name == "" {
			catalog_name = item.Name
		}
	}
	fmt.Println("")

	fmt.Printf("\nvdc items\n")
	for _, res := range vdc.Vdc.ResourceEntities {
		for N, item := range res.ResourceEntity {
			fmt.Printf("%3d %-40s %s\n", N, item.Name, item.Type)
		}
	}
	fmt.Println("")

	if catalog_name != "" {
		fmt.Printf("\ncatalog items\n")
		cat, err := org.FindCatalog(catalog_name)
		if err != nil {
			fmt.Printf("Error retrieving catalog %s\n%s\n", catalog_name, err)
			os.Exit(1)
		}
		for _, item := range cat.Catalog.CatalogItems {
			for N, deepItem := range item.CatalogItem {
				fmt.Printf("%3d %-40s %s (ID: %s)\n", N, deepItem.Name, deepItem.Type, deepItem.ID)
			}
		}
	}
}
