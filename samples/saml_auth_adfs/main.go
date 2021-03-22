/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var (
	username        string
	password        string
	org             string
	apiEndpoint     string
	customAdfsRptId string
)

func init() {
	flag.StringVar(&username, "username", "", "Username")
	flag.StringVar(&password, "password", "", "Password")
	flag.StringVar(&org, "org", "System", "Org name. Default is 'System'")
	flag.StringVar(&apiEndpoint, "endpoint", "", "API endpoint (e.g. 'https://hostname/api')")
	flag.StringVar(&customAdfsRptId, "rpt", "", "Custom Relaying party trust ID. Default is vCD SAML Entity ID")
}

// Usage:
// # go build -o auth
// # ./auth --username test@test-forest.net --password asdasd --org my-org --endpoint https://192.168.1.160/api
func main() {
	flag.Parse()

	if username == "" || password == "" || org == "" || apiEndpoint == "" {
		fmt.Printf("At least 'username', 'password', 'org' and 'endpoint' must be specified\n")
		os.Exit(1)
	}

	vcdURL, err := url.Parse(apiEndpoint)
	if err != nil {
		fmt.Printf("Error parsing supplied endpoint %s: %s", apiEndpoint, err)
		os.Exit(2)
	}

	ctx := context.Background()

	// Create VCD client allowing insecure TLS connection and using SAML auth.
	// WithSamlAdfs() allows SAML authentication when vCD uses Microsoft Active Directory
	// Federation Services (ADFS) as SAML IdP. The code below allows to authenticate ADFS using
	// WS-TRUST endpoint "/adfs/services/trust/13/usernamemixed"
	// Input parameters:
	// user - username for authentication against ADFS server (e.g. 'test@test-forest.net' or 'test-forest.net\test')
	// password - password for authentication against ADFS server
	// org  - Org to authenticate to. Can be 'System'.
	// customAdfsRptId - override relaying party trust ID. If it is empty - vCD Entity ID will be used
	// as Relaying Party Trust ID.
	vcdCli := govcd.NewVCDClient(*vcdURL, true, govcd.WithSamlAdfs(true, customAdfsRptId))
	err = vcdCli.Authenticate(ctx, username, password, org)
	if err != nil {

		fmt.Println(err)
		os.Exit(3)
	}

	// To prove authentication worked - just fetch all edge gateways and dump them on the screen
	edgeGatewayResults, err := vcdCli.Query(ctx, map[string]string{"type": "edgeGateway"})
	if err != nil {
		fmt.Printf("Error retrieving Edge Gateways: %s\n", err)
		os.Exit(4)
	}

	fmt.Printf("Found %d Edge Gateways\n", len(edgeGatewayResults.Results.EdgeGatewayRecord))
	for _, v := range edgeGatewayResults.Results.EdgeGatewayRecord {
		fmt.Println(v.Name)
	}
}
