//go:build functional || openapi || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// Test_OpenApiRawJsonAuditTrail uses low level GET function to test out that pagination really works. It is an example
// how to fetch response from multiple pages in RAW json messages without having defined as struct.
func (vcd *TestVCD) Test_OpenApiRawJsonAuditTrail(check *C) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAuditTrail
	skipOpenApiEndpointTest(vcd, check, endpoint)
	apiVersion, err := vcd.client.Client.checkOpenApiEndpointCompatibility(endpoint)
	check.Assert(err, IsNil)

	urlRef, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint)
	check.Assert(err, IsNil)

	// Get a timestamp after which endpoint contains at least 10 elements
	filterTimeStamp := getAuditTrailTimestampWithElements(10, check, vcd, apiVersion, urlRef)

	// Limit search of audits trails to the last 12 hours so that it doesn't take too long and set pageSize to be 1 result
	// to force following pages
	queryParams := url.Values{}
	queryParams.Add("filter", "timestamp=gt="+filterTimeStamp)
	queryParams.Add("pageSize", "1") // pageSize=1 to enforce internal pagination
	queryParams.Add("sortDesc", "timestamp")

	allResponses := []json.RawMessage{{}}
	err = vcd.vdc.client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &allResponses, nil)

	check.Assert(err, IsNil)
	check.Assert(len(allResponses) > 1, Equals, true)

	// Build a regex ant match it internally so that we are sure auditTrail events are returned in RAW json message. There
	// should be the same amount of audit event IDs as total responses
	auditLogUrn := regexp.MustCompile("urn:vcloud:audit:")
	responseStrings := jsonRawMessagesToStrings(allResponses)
	allStringResponses := `[` + strings.Join(responseStrings, ",") + `]`
	matches := auditLogUrn.FindAllStringIndex(allStringResponses, -1)
	check.Assert(len(matches), Equals, len(allResponses))
}

// Test_OpenApiInlineStructAuditTrail uses low level GET function to test out that get function can unmarshal directly
// to user defined inline type
func (vcd *TestVCD) Test_OpenApiInlineStructAuditTrail(check *C) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAuditTrail
	skipOpenApiEndpointTest(vcd, check, endpoint)
	apiVersion, err := vcd.client.Client.checkOpenApiEndpointCompatibility(endpoint)
	check.Assert(err, IsNil)

	urlRef, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint)
	check.Assert(err, IsNil)

	// Inline type
	type AuditTrail struct {
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

	allResponses := []*AuditTrail{{}}

	// Define FIQL query to find events for the last 6 hours. At least login operations will already be here on test run
	queryParams := url.Values{}
	filterTime := time.Now().Add(-6 * time.Hour).Format(types.FiqlQueryTimestampFormat)
	queryParams.Add("filter", "timestamp=gt="+filterTime)

	err = vcd.vdc.client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &allResponses, nil)

	check.Assert(err, IsNil)
	check.Assert(len(allResponses) > 1, Equals, true)

	// Check that all responses have IDs populated
	for _, v := range allResponses {
		check.Assert(v.EventID, NotNil)
	}
}

// Test_OpenApiInlineStructCRUDRoles test aims to test out low level OpenAPI functions to check if all of them work as
// expected. It uses a very simple "InlineRoles" endpoint which does not have bigger prerequisites and therefore is not
// dependent one more deployment specific features. It also supports all of the OpenAPI CRUD endpoints so is a good
// endpoint to test on
// This test performs the following:
// 1. Gets all available roles using "Get all endpoint"
// 2.1 Uses FIQL query filtering to retrieve specific item by ID on "Get All" endpoint
// 2.2 Use GET by ID endpoint to check that each of roles retrieved by get all can be found individually
// 2.3 Compares retrieved struct by using "Get all" endpoint and FIQL filter with struct retrieved by using "Get By ID"
// endpoint
// 3. Creates a new role and verifies it is created as specified by using deep equality
// 4. Updates role description
// 5. Deletes created role
// 6. Tests read for deleted item
// 7. Create role once more using "Sync" version of POST function
// 7.1 Queries TestConnection endpoint using "Sync" version of POST function to see that it handles 200OK accordingly
// 8. Update role once more using "Sync" version of PUT function
// 9. Delete role once again
func (vcd *TestVCD) Test_OpenApiInlineStructCRUDRoles(check *C) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointRoles
	apiVersion, err := vcd.client.Client.checkOpenApiEndpointCompatibility(endpoint)
	check.Assert(err, IsNil)
	skipOpenApiEndpointTest(vcd, check, endpoint)

	// Step 1 - Get all roles
	urlRef, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint)
	check.Assert(err, IsNil)

	type InlineRoles struct {
		ID          string `json:"id,omitempty"`
		Name        string `json:"name"`
		Description string `json:"description"`
		BundleKey   string `json:"bundleKey"`
		ReadOnly    bool   `json:"readOnly"`
	}

	allExistingRoles := []*InlineRoles{{}}
	err = vcd.vdc.client.OpenApiGetAllItems(apiVersion, urlRef, nil, &allExistingRoles, nil)
	check.Assert(err, IsNil)

	// Step 2 - Get all roles using query filters
	for _, oneRole := range allExistingRoles {
		// Step 2.1 - retrieve specific role by using FIQL filter
		urlRef2, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint)
		check.Assert(err, IsNil)

		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+oneRole.ID)

		expectOneRoleResultById := []*InlineRoles{{}}

		err = vcd.vdc.client.OpenApiGetAllItems(apiVersion, urlRef2, queryParams, &expectOneRoleResultById, nil)
		check.Assert(err, IsNil)
		check.Assert(len(expectOneRoleResultById) == 1, Equals, true)

		// Step 2.2 - retrieve specific role by using endpoint
		singleRef, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint + oneRole.ID)
		check.Assert(err, IsNil)

		oneRole := &InlineRoles{}
		err = vcd.vdc.client.OpenApiGetItem(apiVersion, singleRef, nil, oneRole, nil)
		check.Assert(err, IsNil)
		check.Assert(oneRole, NotNil)

		// Step 2.3 - compare struct retrieved by using filter and the one retrieved by exact endpoint ID
		check.Assert(oneRole, DeepEquals, expectOneRoleResultById[0])

	}

	// Step 3 - Create a new role and ensure it is created as specified by doing deep comparison
	createUrl, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint)
	check.Assert(err, IsNil)

	newRole := &InlineRoles{
		Name:        check.TestName(),
		Description: "Role created by test",
		// This BundleKey is being set by VCD even if it is not sent
		BundleKey: types.VcloudUndefinedKey,
		ReadOnly:  false,
	}
	newRoleResponse := &InlineRoles{}
	err = vcd.client.Client.OpenApiPostItem(apiVersion, createUrl, nil, newRole, newRoleResponse, nil)
	check.Assert(err, IsNil)

	// Ensure supplied and created structs differ only by ID
	newRole.ID = newRoleResponse.ID
	check.Assert(newRoleResponse, DeepEquals, newRole)

	// Step 4 - update created role (change description)
	newRoleResponse.Description = "Updated description created by test"
	updateUrl, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint, newRoleResponse.ID)
	check.Assert(err, IsNil)

	updatedRoleResponse := &InlineRoles{}
	err = vcd.client.Client.OpenApiPutItem(apiVersion, updateUrl, nil, newRoleResponse, updatedRoleResponse, nil)
	check.Assert(err, IsNil)

	// Ensure supplied and response objects are identical (update worked)
	check.Assert(updatedRoleResponse, DeepEquals, newRoleResponse)

	// Step 5 - delete created role
	deleteUrlRef, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint, newRoleResponse.ID)
	check.Assert(err, IsNil)

	err = vcd.client.Client.OpenApiDeleteItem(apiVersion, deleteUrlRef, nil, nil)
	check.Assert(err, IsNil)

	// Step 6 - try to read deleted role and expect error to contain 'ErrorEntityNotFound'
	// Read is tricky - it throws an error ACCESS_TO_RESOURCE_IS_FORBIDDEN when the resource with ID does not
	// exist therefore one cannot know what kind of error occurred.
	lostRole := &InlineRoles{}
	err = vcd.client.Client.OpenApiGetItem(apiVersion, deleteUrlRef, nil, lostRole, nil)
	check.Assert(ContainsNotFound(err), Equals, true)

	// Step 7 - test synchronous POST and PUT functions (because Roles is a synchronous OpenAPI endpoint)
	newRole.ID = "" // unset ID as it cannot be set for creation
	err = vcd.client.Client.OpenApiPostItemSync(apiVersion, createUrl, nil, newRole, newRoleResponse)
	check.Assert(err, IsNil)

	// Ensure supplied and created structs differ only by ID
	newRole.ID = newRoleResponse.ID
	check.Assert(newRoleResponse, DeepEquals, newRole)

	// Step 7.1 test synchronous POST with return code 200 OK works accordingly - This is checked because OpenAPI endpoint TestConnection returns 200 instead of 201 when success
	var testConnectionResult types.TestConnectionResult
	testConnectionPayload := types.TestConnection{
		Host:    vcd.client.Client.VCDHREF.Host,
		Port:    443,
		Secure:  addrOf(true),
		Timeout: 10,
	}

	testConnectionEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTestConnection
	apiVersionTestConnection, err := vcd.client.Client.checkOpenApiEndpointCompatibility(testConnectionEndpoint)
	check.Assert(err, IsNil)

	urlRefTestConnection, err := vcd.client.Client.OpenApiBuildEndpoint(testConnectionEndpoint)
	check.Assert(err, IsNil)

	err = vcd.client.Client.OpenApiPostItemSync(apiVersionTestConnection, urlRefTestConnection, nil, testConnectionPayload, &testConnectionResult) // This call will get a 200 OK, which is what is being tested here
	check.Assert(err, IsNil)

	// Step 8 - update role using synchronous PUT function
	newRoleResponse.Description = "Updated description created by sync test"
	updateUrl2, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint, newRoleResponse.ID)
	check.Assert(err, IsNil)

	updatedRoleResponse2 := &InlineRoles{}
	err = vcd.client.Client.OpenApiPutItem(apiVersion, updateUrl2, nil, newRoleResponse, updatedRoleResponse2, nil)
	check.Assert(err, IsNil)

	// Ensure supplied and response objects are identical (update worked)
	check.Assert(updatedRoleResponse2, DeepEquals, newRoleResponse)

	// Step 9 - delete role once again
	deleteUrlRef2, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint, newRoleResponse.ID)
	check.Assert(err, IsNil)

	err = vcd.client.Client.OpenApiDeleteItem(apiVersion, deleteUrlRef2, nil, nil)
	check.Assert(err, IsNil)

}

func (vcd *TestVCD) Test_OpenApiTestConnection(check *C) {
	// TestConnection is going to be used against the same VCD instance as the client is connected
	urlTest1 := vcd.client.Client.VCDHREF
	urlTest1.Path = "vcsp/lib/a0c959b4-a6dd-4a68-8042-5025f42d845e"
	urlTest2 := vcd.client.Client.VCDHREF
	urlTest2.Scheme = "http"
	urlTest3 := vcd.client.Client.VCDHREF
	urlTest3.Host = "imadethisup.io"
	urlTest4 := vcd.client.Client.VCDHREF
	urlTest4.Host = fmt.Sprintf("%s:666", urlTest4.Hostname()) // For testing custom port feature
	tests := []struct {
		SubscriptionURL  string
		WantedConnection bool
		WantedError      bool
	}{
		{
			SubscriptionURL:  urlTest1.String(),
			WantedConnection: true,  // it connects and it does SSL connection
			WantedError:      false, //
		},
		{
			SubscriptionURL:  urlTest2.String(),
			WantedConnection: true, // it connects but it does not do SSL connection
			WantedError:      true,
		},
		{
			SubscriptionURL:  urlTest3.String(), // it doesn't do neither connection nor SSL
			WantedConnection: false,
			WantedError:      true,
		},
		{
			SubscriptionURL:  urlTest4.String(), // it doesn't do neither connection nor SSL but tests custom port
			WantedConnection: false,
			WantedError:      true,
		},
	}

	for _, test := range tests {
		result, err := vcd.client.Client.TestConnectionWithDefaults(test.SubscriptionURL)
		check.Assert(err == nil, Equals, !test.WantedError)
		check.Assert(result, Equals, test.WantedConnection)
	}
}

// getAuditTrailTimestampWithElements helps to pick good timestamp filter so that it doesn't take long time to retrieve
// too many items
func getAuditTrailTimestampWithElements(elementCount int, check *C, vcd *TestVCD, apiVersion string, urlRef *url.URL) string {
	client := vcd.client.Client
	qp := url.Values{}
	qp.Add("pageSize", "128")
	qp.Add("sortDesc", "timestamp") // Need to get the newest
	req := client.newOpenApiRequest(apiVersion, qp, http.MethodGet, urlRef, nil, nil)

	resp, err := client.Http.Do(req)
	check.Assert(err, IsNil)

	type AuditTrailTimestamp struct {
		Timestamp string `json:"timestamp"`
	}

	onePageAuditTrail := make([]AuditTrailTimestamp, 1)
	onePageResponse := &types.OpenApiPages{}
	err = decodeBody(types.BodyTypeJSON, resp, &onePageResponse)
	check.Assert(err, IsNil)

	err = resp.Body.Close()
	check.Assert(err, IsNil)

	err = json.Unmarshal(onePageResponse.Values, &onePageAuditTrail)
	check.Assert(err, IsNil)

	var singleElement AuditTrailTimestamp

	// Find newest element limited by provided elementCount
	if len(onePageAuditTrail) < elementCount {
		singleElement = onePageAuditTrail[(len(onePageAuditTrail) - 1)]
	} else {
		singleElement = onePageAuditTrail[(elementCount - 1)]
	}

	timeFormat := dateparse.MustParse(singleElement.Timestamp)

	return timeFormat.Format(types.FiqlQueryTimestampFormat)
}
