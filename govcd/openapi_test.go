// +build functional openapi ALL

package govcd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// Test_CloudAPIAudiTrail uses low level GET function to test out that pagination really works. It is an example how to
// fetch response from multiple pages in RAW json messages without having defined a clear struct.
func (vcd *TestVCD) Test_OpenAPIRawJsonAudiTrail(check *C) {
	minimumRequiredApiVersion := "33.0"
	skipOpenApiEndpoint(vcd, check, "1.0.0/auditTrail", "< "+minimumRequiredApiVersion)

	urlRef, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/auditTrail")
	check.Assert(err, IsNil)

	// Limit search of audi trails to the last 12 hours so that it doesn't take too long and set pageSize to be 1 result to
	// force following pages
	queryParams := url.Values{}
	filterTime := time.Now().Add(-12 * time.Hour).Format(types.FiqlQueryTimestampFormat)
	queryParams.Add("filter", "timestamp=gt="+filterTime)
	queryParams.Add("pageSize", "1")

	allResponses := []json.RawMessage{{}}
	err = vcd.vdc.client.OpenApiGetAllItems(minimumRequiredApiVersion, urlRef, queryParams, &allResponses)

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

// Test_OpenAPIInlineStructAudiTrail uses low level GET function to test out that get function can unmarshal directly
// to user defined inline type
func (vcd *TestVCD) Test_OpenAPIInlineStructAudiTrail(check *C) {
	minimumRequiredApiVersion := "33.0"
	skipOpenApiEndpoint(vcd, check, "1.0.0/auditTrail", "< "+minimumRequiredApiVersion)

	urlRef, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/auditTrail")
	check.Assert(err, IsNil)

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

	allResponses := []*AudiTrail{{}}

	// Define FIQL query to find events for the last 24 hours
	queryParams := url.Values{}
	filterTime := time.Now().Add(-24 * time.Hour).Format(types.FiqlQueryTimestampFormat)
	queryParams.Add("filter", "timestamp=gt="+filterTime)

	err = vcd.vdc.client.OpenApiGetAllItems(minimumRequiredApiVersion, urlRef, queryParams, &allResponses)

	check.Assert(err, IsNil)
	check.Assert(len(allResponses) > 1, Equals, true)

	// Check that all responses have IDs populated
	for _, v := range allResponses {
		check.Assert(v.EventID, NotNil)
	}
}

// Test_OpenAPIInlineStructCRUDRoles test aims to test out low level CloudAPI functions to check if all of them work as
// expected. It uses a very simple "Roles" endpoint which does not have bigger prerequisites and therefore is not
// dependent one more deployment specific features. It also supports all of the CloudAPI CRUD endpoints so is a good
// endpoint to test on
// Actions of the test:
// 1. Get all available roles using "Get all endpoint"
// 2.1 Use FIQL query filtering to retrieve specific item by ID on "Get All" endpoint
// 2.2 Use GET by ID endpoint to check that each of roles retrieved by get all can be found individually
// 2.3 Compare struct retrieved by using "Get all" endpoint and FIQL filter with struct retrieved by using "Get By ID"
// 3. Create a new role and verify it is created as specified by using deep equality
// 4. Delete created role
// 5. Test read for deleted item
func (vcd *TestVCD) Test_OpenAPIInlineStructCRUDRoles(check *C) {
	minimumRequiredApiVersion := "31.0"
	skipOpenApiEndpoint(vcd, check, "1.0.0/roles", "< "+minimumRequiredApiVersion)

	// Step 1 - Get all roles
	urlRef, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/roles")
	check.Assert(err, IsNil)

	type Roles struct {
		ID          string `json:"id,omitempty"`
		Name        string `json:"name"`
		Description string `json:"description"`
		BundleKey   string `json:"bundleKey"`
		ReadOnly    bool   `json:"readOnly"`
	}

	allExistingRoles := []*Roles{{}}
	err = vcd.vdc.client.OpenApiGetAllItems(minimumRequiredApiVersion, urlRef, nil, &allExistingRoles)
	check.Assert(err, IsNil)

	// Step 2 - Get all roles using query filters
	for _, oneRole := range allExistingRoles {
		// Step 2.1 - retrieve specific role by using FIQL filter
		urlRef2, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/roles")
		check.Assert(err, IsNil)

		queryParams := url.Values{}
		queryParams.Add("filter", "id=="+oneRole.ID)

		expectOneRoleResultById := []*Roles{{}}

		err = vcd.vdc.client.OpenApiGetAllItems(minimumRequiredApiVersion, urlRef2, queryParams, &expectOneRoleResultById)
		check.Assert(err, IsNil)
		check.Assert(len(expectOneRoleResultById) == 1, Equals, true)

		// Step 2.2 - retrieve specific role by using endpoint
		singleRef, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/roles/" + oneRole.ID)
		check.Assert(err, IsNil)

		oneRole := &Roles{}
		err = vcd.vdc.client.OpenApiGetItem(minimumRequiredApiVersion, singleRef, nil, oneRole)
		check.Assert(err, IsNil)
		check.Assert(oneRole, NotNil)

		// Step 2.3 - compare struct retrieve by using filter and the one retrieve by exact endpoint ID
		check.Assert(oneRole, DeepEquals, expectOneRoleResultById[0])

	}

	// Step 3 - Create a new role and ensure it is created as specified by doing deep comparison
	createUrl, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/roles")
	check.Assert(err, IsNil)

	newRole := &Roles{
		Name:        check.TestName(),
		Description: "Role created by test",
		BundleKey:   "com.vmware.vcloud.undefined.key",
		ReadOnly:    false,
	}
	newRoleResponse := &Roles{}
	err = vcd.client.Client.OpenApiPostItem(minimumRequiredApiVersion, createUrl, nil, newRole, newRoleResponse)
	check.Assert(err, IsNil)

	// Ensure supplied and created structs differ only by ID
	newRole.ID = newRoleResponse.ID
	check.Assert(newRoleResponse, DeepEquals, newRole)

	// Step 4 - delete created role
	deleteUrlRef, err := vcd.client.Client.BuildOpenApiEndpoint("1.0.0/roles/", newRoleResponse.ID)
	check.Assert(err, IsNil)

	err = vcd.client.Client.OpenApiDeleteItem(minimumRequiredApiVersion, deleteUrlRef, nil)
	check.Assert(err, IsNil)

	// Step 5 - try to read deleted role and expect error to contain 'ErrorEntityNotFound'
	// Read is tricky - it throws an error ACCESS_TO_RESOURCE_IS_FORBIDDEN when the resource with ID does not
	// exist therefore one cannot know what kind of error occurred.
	lostRole := &Roles{}
	err = vcd.client.Client.OpenApiGetItem(minimumRequiredApiVersion, deleteUrlRef, nil, lostRole)
	check.Assert(ContainsNotFound(err), Equals, true)
}

// skipOpenApiEndpoint is a helper to skip tests for particular unsupported OpenAPI endpoints
func skipOpenApiEndpoint(vcd *TestVCD, check *C, endpoint, skipWhenMaxVersion string) {
	if vcd.client.Client.APIVCDMaxVersionIs(skipWhenMaxVersion) {
		maxSupportedVersion, err := vcd.client.Client.maxSupportedVersion()
		if err != nil {
			panic(fmt.Sprintf("Could not get maximum supported version: %s", err))
		}
		skipText := fmt.Sprintf("Skipping test because OpenAPI endpoint '%s' must satisfy version constraint '%s'. Maximum supported version is %s",
			endpoint, skipWhenMaxVersion, maxSupportedVersion)
		check.Skip(skipText)
	}
}
