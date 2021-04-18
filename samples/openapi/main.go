/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var (
	username    string
	password    string
	org         string
	apiEndpoint string
	mode        string
)

func init() {
	flag.StringVar(&username, "username", "", "Username")
	flag.StringVar(&password, "password", "", "Password")
	flag.StringVar(&org, "org", "System", "Org name. Default is 'System'")
	flag.StringVar(&apiEndpoint, "endpoint", "", "API endpoint (e.g. 'https://hostname/api')")
	flag.StringVar(&mode, "mode", "", "OpenAPI query mode: 1 - RAW json, 2 - inline type")
}

// Usage:
// # go build -o openapi
// # ./openapi --username my_user --password my_secret_password --org my-org --endpoint
// https://192.168.1.160/api  --mode 1
func main() {
	flag.Parse()

	if username == "" || password == "" || org == "" || apiEndpoint == "" || mode == "" {
		fmt.Printf("'username', 'password', 'org', 'endpoint' and 'mode' must be specified\n")
		os.Exit(1)
	}

	vcdURL, err := url.Parse(apiEndpoint)
	if err != nil {
		fmt.Printf("Error parsing supplied endpoint %s: %s", apiEndpoint, err)
		os.Exit(2)
	}

	ctx := context.Background()

	vcdCli := govcd.NewVCDClient(*vcdURL, true)
	err = vcdCli.Authenticate(ctx, username, password, org)
	if err != nil {

		fmt.Println(err)
		os.Exit(3)
	}

	if vcdCli.Client.APIVCDMaxVersionIs(ctx, "< 33.0") {
		fmt.Println("This example requires VCD API to support at least version 33.0 (VCD 10.0) to use '1.0.0/auditTrail' endpoint")
		os.Exit(4)
	}

	switch mode {
	case "1":
		openAPIGetRawJsonAuditTrail(ctx, vcdCli)
	case "2":
		openAPIGetStructAuditTrail(ctx, vcdCli)
	}

}

// openAPIGetRawJsonAuditTrail is an example function how to use low level function to interact
// with OpenAPI in VCD. This examples dumps to screen valid JSON which can then be processed using
// other tools (for example 'jq' in shell)
// It also uses FIQL query filter to retrieve auditTrail items only for the last 12 hours
func openAPIGetRawJsonAuditTrail(ctx context.Context, vcdClient *govcd.VCDClient) {
	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint("1.0.0/auditTrail")
	if err != nil {
		panic(err)
	}

	queryParams := url.Values{}
	filterTime := time.Now().Add(-12 * time.Hour).Format(types.FiqlQueryTimestampFormat)
	queryParams.Add("filter", "timestamp=gt="+filterTime)

	allResponses := []json.RawMessage{{}}
	err = vcdClient.Client.OpenApiGetAllItems(ctx, "33.0", urlRef, queryParams, &allResponses)
	if err != nil {
		panic(err)
	}

	// Wrap slice of response objects into JSON list so that it is correct JSON
	responseStrings := jsonRawMessagesToStrings(allResponses)
	allStringResponses := `[` + strings.Join(responseStrings, ",") + `]`
	fmt.Println(allStringResponses)
}

// openAPIGetStructAuditTrail is an example function how to use low level function to interact with
// OpenAPI in VCD and marshal responses into custom defined struct with tags.
// It also uses FIQL query filter to retrieve auditTrail items only for the last 12 hours
func openAPIGetStructAuditTrail(ctx context.Context, vcdClient *govcd.VCDClient) {
	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint("1.0.0/auditTrail")
	if err != nil {
		panic(err)
	}

	// Inline type
	type AudiTrail struct {
		EventID      string `json:"eventId"`
		Description  string `json:"description"`
		OperatingOrg struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"operatingOrg"`
		User struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"user"`
		EventEntity struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"eventEntity"`
		TaskID               interface{} `json:"taskId"`
		TaskCellID           string      `json:"taskCellId"`
		CellID               string      `json:"cellId"`
		EventType            string      `json:"eventType"`
		ServiceNamespace     string      `json:"serviceNamespace"`
		EventStatus          string      `json:"eventStatus"`
		Timestamp            string      `json:"timestamp"`
		External             bool        `json:"external"`
		AdditionalProperties struct {
			UserRoles                         string `json:"user.roles"`
			UserSessionID                     string `json:"user.session.id"`
			CurrentContextUserProxyAddress    string `json:"currentContext.user.proxyAddress"`
			CurrentContextUserClientIPAddress string `json:"currentContext.user.clientIpAddress"`
		} `json:"additionalProperties"`
	}

	response := []*AudiTrail{{}}

	queryParams := url.Values{}
	filterTime := time.Now().Add(-12 * time.Hour).Format(types.FiqlQueryTimestampFormat)
	queryParams.Add("filter", "timestamp=gt="+filterTime)

	err = vcdClient.Client.OpenApiGetAllItems(ctx, "33.0", urlRef, queryParams, &response)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Got %d results\n", len(response))

	for _, value := range response {
		fmt.Printf("%s - %s, -%s\n", value.Timestamp, value.User.Name, value.EventType)
	}
}

// jsonRawMessagesToStrings converts []*json.RawMessage to []string
func jsonRawMessagesToStrings(messages []json.RawMessage) []string {
	resultString := make([]string, len(messages))
	for index, message := range messages {
		resultString[index] = string(message)
	}
	return resultString
}
