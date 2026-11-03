// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/peterhellberg/link"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// This file contains generalised low level methods to interact with VCD OpenAPI REST endpoints as documented in
// https://{VCD_HOST}/docs. In addition to this there are OpenAPI browser endpoints for tenant and provider
// respectively https://{VCD_HOST}/api-explorer/tenant/tenant-name and https://{VCD_HOST}/api-explorer/provider .
// OpenAPI has functions supporting below REST methods:
// GET /items (gets a slice of types like `[]types.OpenAPIEdgeGateway` or even `[]json.RawMessage` to process JSON as text.
// POST /items - creates an item
// PUT /items/URN - updates an item with specified URN
// GET /items/URN - retrieves an item with specified URN
// DELETE /items/URN - deletes an item with specified URN
//
// GET endpoints support FIQL for filtering in field `filter`. (FIQL IETF doc - https://tools.ietf.org/html/draft-nottingham-atompub-fiql-00)
// Not all API fields are supported for FIQL filtering and sometimes they return odd errors when filtering is
// unsupported. No exact documentation exists so far.
//
// Note. All functions accepting URL reference (*url.URL) will make a copy of URL because they may mutate URL reference.
// The parameter is kept as *url.URL for convenience because standard library provides pointer values.
//
// OpenAPI versioning.
// OpenAPI was introduced in VCD 9.5 (with API version 31.0). Endpoints are being added with each VCD iteration.
// Internally hosted documentation (https://HOSTNAME/docs/) can be used to check which endpoints where introduced in
// which VCD API version.
// Additionally each OpenAPI endpoint has a semantic version in its path (e.g.
// https://HOSTNAME/cloudapi/1.0.0/auditTrail). This versioned endpoint should ensure compatibility as VCD evolves.

// OpenApiIsSupported allows to check whether VCD supports OpenAPI. Each OpenAPI endpoint however is introduced with
// different VCD API versions so this is just a general check if OpenAPI is supported at all. Particular endpoint
// introduction version can be checked in self hosted docs (https://HOSTNAME/docs/)
func (client *Client) OpenApiIsSupported() bool {
	// TM also advertises product versions such as 9.1.2, which sort below VCD API 31.
	return client.IsTm() || client.APIVCDMaxVersionIs(">= 31")
}

// OpenApiBuildEndpoint helps to construct OpenAPI endpoint by using already configured VCD HREF while requiring only
// the last bit for endpoint. This is a variadic function and multiple pieces can be supplied for convenience. Leading
// '/' is added automatically.
// Sample URL construct: https://HOST/cloudapi/endpoint
func (client *Client) OpenApiBuildEndpoint(endpoint ...string) (*url.URL, error) {
	endpointString := client.rootVcdHref() + "/cloudapi/" + strings.Join(endpoint, "")
	urlRef, err := url.ParseRequestURI(endpointString)
	if err != nil {
		return nil, fmt.Errorf("error formatting OpenAPI endpoint: %s", err)
	}
	return urlRef, nil
}

// OpenApiGetAllItems retrieves and accumulates all pages then parsing them to a single 'outType' object. It works by at
// first crawling pages and accumulating all responses into []json.RawMessage (as strings). Because there is no
// intermediate unmarshalling to exact `outType` for every page it unmarshals into response struct in one go. 'outType'
// must be a slice of object (e.g. []*types.OpenAPIEdgeGateway) because this response contains slice of structs.
//
// Note. Query parameter 'pageSize' is defaulted to 128 (maximum supported) unless it is specified in queryParams
func (client *Client) OpenApiGetAllItems(apiVersion string, urlRef *url.URL, queryParams url.Values, outType interface{}, additionalHeader map[string]string) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Getting all items from endpoint %s for parsing into %s type\n",
		urlRefCopy.String(), reflect.TypeOf(outType))

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	// Page size is defaulted to 128 (maximum supported number) to reduce HTTP calls and improve performance unless caller
	// provides other value
	newQueryParams := defaultPageSize(queryParams, "128")
	util.Logger.Printf("[TRACE] Will use 'pageSize=%s'", newQueryParams.Get("pageSize"))

	// Perform API call to initial endpoint. The function call recursively follows pages using Link headers "nextPage"
	// until it crawls all results
	responses, err := client.openApiGetAllPages(apiVersion, urlRefCopy, newQueryParams, outType, nil, additionalHeader)
	if err != nil {
		return fmt.Errorf("error getting all pages for endpoint %s: %s", urlRefCopy.String(), err)
	}

	// Create a slice of raw JSON messages in text so that they can be unmarshalled to specified `outType` after multiple
	// calls are executed
	var rawJsonBodies []string
	for _, singleObject := range responses {
		rawJsonBodies = append(rawJsonBodies, string(singleObject))
	}

	// rawJsonBodies contains a slice of all response objects and they must be formatted as a JSON slice (wrapped
	// into `[]`, separated with semicolons) so that unmarshalling to specified `outType` works in one go
	allResponses := `[` + strings.Join(rawJsonBodies, ",") + `]`

	// Unmarshal all accumulated responses into `outType`
	if err = json.Unmarshal([]byte(allResponses), &outType); err != nil {
		return fmt.Errorf("error decoding values into type: %s", err)
	}

	return nil
}

// OpenApiGetItem is a low level OpenAPI client function to perform GET request for any item.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It responds with HTTP 403: Forbidden - If the user is not authorized or the entity does not exist. When HTTP 403 is
// returned this function returns "ErrorEntityNotFound: API_ERROR" so that one can use ContainsNotFound(err) to
// differentiate when an object was not found from any other error.
func (client *Client) OpenApiGetItem(apiVersion string, urlRef *url.URL, params url.Values, outType interface{}, additionalHeader map[string]string) error {
	_, err := client.OpenApiGetItemAndHeaders(apiVersion, urlRef, params, outType, additionalHeader)
	return err
}

// OpenApiGetItemAndHeaders is a low level OpenAPI client function to perform GET request for any item and return all the headers.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It responds with HTTP 403: Forbidden - If the user is not authorized or the entity does not exist. When HTTP 403 is
// returned this function returns "ErrorEntityNotFound: API_ERROR" so that one can use ContainsNotFound(err) to
// differentiate when an object was not found from any other error.
func (client *Client) OpenApiGetItemAndHeaders(apiVersion string, urlRef *url.URL, params url.Values, outType interface{}, additionalHeader map[string]string) (http.Header, error) {
	op := client.openApiOperation(apiVersion, urlRef, params, additionalHeader)
	op.forbiddenAsNotFound = true // Preserve the VCD single-item lookup contract.
	return op.read(outType)
}

// OpenApiPostItemSync is a low level OpenAPI client function to perform POST request for items that support synchronous
// requests. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways') that supports synchronous requests. It
// will return an error when endpoint does not support synchronous requests (HTTP response status code is not 200 or 201).
// Response will be unmarshalled into outType.
//
// Note. Even though it may return error if the item does not support synchronous request - the object may still be
// created. OpenApiPostItem would handle both cases and always return created item.
func (client *Client) OpenApiPostItemSync(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}) error {
	op := client.openApiOperation(apiVersion, urlRef, params, nil)
	op.expectedStatus = []int{http.StatusOK, http.StatusCreated}
	_, err := op.write(http.MethodPost, payload, outType)
	return err
}

// OpenApiPostItemAsync is a low level OpenAPI client function to perform POST request for items that support
// asynchronous requests. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways') that supports asynchronous
// requests. It will return an error if item does not support asynchronous request (does not respond with HTTP 202).
//
// Note. Even though it may return error if the item does not support asynchronous request - the object may still be
// created. OpenApiPostItem would handle both cases and always return created item.
func (client *Client) OpenApiPostItemAsync(apiVersion string, urlRef *url.URL, params url.Values, payload interface{}) (Task, error) {
	return client.OpenApiPostItemAsyncWithHeaders(apiVersion, urlRef, params, payload, nil)
}

// OpenApiPostItemAsyncWithHeaders is a low level OpenAPI client function to perform POST request for items that support
// asynchronous requests. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways') that supports asynchronous
// requests. It will return an error if item does not support asynchronous request (does not respond with HTTP 202).
//
// Note. Even though it may return error if the item does not support asynchronous request - the object may still be
// created. OpenApiPostItem would handle both cases and always return created item.
func (client *Client) OpenApiPostItemAsyncWithHeaders(apiVersion string, urlRef *url.URL, params url.Values, payload interface{}, additionalHeader map[string]string) (Task, error) {
	task, err := client.openApiOperation(apiVersion, urlRef, params, additionalHeader).runAsync(http.MethodPost, payload)
	if err != nil {
		return Task{}, err
	}
	return *task, nil
}

// OpenApiPostItem is a low level OpenAPI client function to perform POST request for item supporting synchronous or
// asynchronous requests. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways'). When a task is
// synchronous - it will track task until it is finished and pick reference to marshal outType.
func (client *Client) OpenApiPostItem(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}, additionalHeader map[string]string) error {
	_, err := client.OpenApiPostItemAndGetHeaders(apiVersion, urlRef, params, payload, outType, additionalHeader)
	return err
}

// OpenApiPostItemAndGetHeaders is a low level OpenAPI client function to perform POST request for item supporting synchronous or
// asynchronous requests, that returns also the response headers. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways'). When a task is
// synchronous - it will track task until it is finished and pick reference to marshal outType.
func (client *Client) OpenApiPostItemAndGetHeaders(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}, additionalHeader map[string]string) (http.Header, error) {
	return client.openApiOperation(apiVersion, urlRef, params, additionalHeader).write(http.MethodPost, payload, outType)
}

// OpenApiPostUrlEncoded is a non-standard function used to send a POST request with `x-www-form-urlencoded` format.
// Accepts a map in format of key:value, marshals the response body in JSON format to outType.
// If additionalHeader contains a "Content-Type" header, it will be overwritten to "x-www-form-urlencoded"
func (client *Client) OpenApiPostUrlEncoded(apiVersion string, urlRef *url.URL, params url.Values, payloadMap map[string]string, outType interface{}, additionalHeaders map[string]string) error {
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Sending a POST request with 'Content-Type: x-www-form-urlencoded' header to endpoint %s with expected response of type %s", urlRefCopy.String(), reflect.TypeOf(outType))

	// Add all values of the payloadMap to the actual payload
	urlValues := url.Values{}
	for key, value := range payloadMap {
		urlValues.Add(key, value)
	}
	body := strings.NewReader(urlValues.Encode())

	// Create the header map if it's nil
	if additionalHeaders == nil {
		additionalHeaders = make(map[string]string)
	}
	// Overwrite the Content-Type header as this is a method only usable for x-www-form-urlencoded
	additionalHeaders["Content-Type"] = "application/x-www-form-urlencoded"

	req := client.newOpenApiRequest(apiVersion, params, http.MethodPost, urlRef, body, additionalHeaders)
	resp, err := client.openApiExecuteRequest(req, false, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.Logger.Printf("[TRACE] HTTP status code 200 expected. Got %d", resp.StatusCode)
	}

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding JSON response after POST: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

// OpenApiPutItemSync is a low level OpenAPI client function to perform PUT request for items that support synchronous
// requests. The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}') and support synchronous
// requests. As before, any successful HTTP status is accepted; 200 and 201 responses are decoded into outType.
// Response will be unmarshalled into outType.
//
// Note. Even though it may return error if the item does not support synchronous request - the object may still be
// updated. OpenApiPutItem would handle both cases and always return updated item.
func (client *Client) OpenApiPutItemSync(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}, additionalHeader map[string]string) error {
	_, err := client.openApiOperation(apiVersion, urlRef, params, additionalHeader).write(http.MethodPut, payload, outType)
	return err
}

// OpenApiPutItemAsync is a low level OpenAPI client function to perform PUT request for items that support asynchronous
// requests. The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}') that supports asynchronous
// requests. It will return an error if item does not support asynchronous request (does not respond with HTTP 202).
//
// Note. Even though it may return error if the item does not support asynchronous request - the object may still be
// created. OpenApiPutItem would handle both cases and always return created item.
func (client *Client) OpenApiPutItemAsync(apiVersion string, urlRef *url.URL, params url.Values, payload interface{}, additionalHeader map[string]string) (Task, error) {
	task, err := client.openApiOperation(apiVersion, urlRef, params, additionalHeader).runAsync(http.MethodPut, payload)
	if err != nil {
		return Task{}, err
	}
	return *task, nil
}

// OpenApiPutItem is a low level OpenAPI client function to perform PUT request for any item.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It handles synchronous and asynchronous tasks. When a task is synchronous - it will block until it is finished.
func (client *Client) OpenApiPutItem(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}, additionalHeader map[string]string) error {
	_, err := client.OpenApiPutItemAndGetHeaders(apiVersion, urlRef, params, payload, outType, additionalHeader)
	return err
}

// OpenApiPutItemAndGetHeaders is a low level OpenAPI client function to perform PUT request for any item and return the response headers.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It handles synchronous and asynchronous tasks. When a task is synchronous - it will block until it is finished.
func (client *Client) OpenApiPutItemAndGetHeaders(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}, additionalHeader map[string]string) (http.Header, error) {
	return client.openApiOperation(apiVersion, urlRef, params, additionalHeader).write(http.MethodPut, payload, outType)
}

// OpenApiDeleteItem is a low level OpenAPI client function to perform DELETE request for any item.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It handles synchronous and asynchronous tasks. When a task is synchronous - it will block until it is finished.
func (client *Client) OpenApiDeleteItem(apiVersion string, urlRef *url.URL, params url.Values, additionalHeader map[string]string) error {
	_, err := client.openApiOperation(apiVersion, urlRef, params, additionalHeader).write(http.MethodDelete, nil, nil)
	return err
}

// openApiOperation is the prepared request shared by generic CRUD helpers and the
// public OpenAPI methods. Version and tenant policies are applied before execution.
type openApiOperation struct {
	client              *Client
	apiVersion          string
	url                 *url.URL
	query               url.Values
	headers             map[string]string
	expectedStatus      []int
	forbiddenAsNotFound bool
}

func (client *Client) openApiOperation(apiVersion string, uri *url.URL, query url.Values, headers map[string]string) openApiOperation {
	return openApiOperation{client: client, apiVersion: apiVersion, url: uri, query: query, headers: headers, forbiddenAsNotFound: true}
}

// request owns JSON encoding; openApiExecuteRequest owns transport and error handling.
// The caller owns the successful response body. Error responses are closed here.
func (op openApiOperation) request(method string, payload interface{}) (*http.Response, error) {
	if !op.client.OpenApiIsSupported() {
		return nil, fmt.Errorf("OpenAPI is not supported on this server")
	}
	var body io.Reader
	if payload != nil {
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON data for %s request: %s", method, err)
		}
		body = bytes.NewReader(encoded)
	}
	req := op.client.newOpenApiRequest(op.apiVersion, op.query, method, op.url, body, op.headers)
	return op.client.openApiExecuteRequest(req, method == http.MethodGet && op.forbiddenAsNotFound, op.expectedStatus)
}

func (client *Client) openApiExecuteRequest(req *http.Request, forbiddenAsNotFound bool, expectedStatus []int) (*http.Response, error) {
	resp, err := client.Http.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}
	if resp.Request == nil {
		resp.Request = req
	}
	if forbiddenAsNotFound && resp.StatusCode == http.StatusForbidden {
		err := ParseErr(types.BodyTypeJSON, resp, &types.OpenApiError{})
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", ErrorEntityNotFound, err)
	}
	if _, err = checkRespWithErrType(types.BodyTypeJSON, resp, nil, &types.OpenApiError{}); err != nil {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("error in HTTP %s request: %s", req.Method, err)
	}
	if len(expectedStatus) > 0 && !slices.Contains(expectedStatus, resp.StatusCode) {
		util.Logger.Printf("[TRACE] %s %s expected HTTP status %v, got %d", req.Method, req.URL.Path, expectedStatus, resp.StatusCode)
		logOpenApiResponse(resp)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s %s: expected HTTP status %v, got %d", req.Method, req.URL.Path, expectedStatus, resp.StatusCode)
	}
	return resp, nil
}

func (op openApiOperation) read(outType interface{}) (http.Header, error) {
	util.Logger.Printf("[TRACE] Getting item from endpoint %s with expected response of type %T", op.url.String(), outType)
	resp, err := op.request(http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return nil, fmt.Errorf("error decoding JSON response after GET: %s", err)
	}
	return resp.Header, nil
}

// task retains the effective request context without modifying the shared client.
// Only tenant headers are copied from the JSON request: task polling uses XML.
func (op openApiOperation) task(resp *http.Response) (*Task, error) {
	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("asynchronous response has no task Location")
	}
	ref, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	ref = resp.Request.URL.ResolveReference(ref)
	util.Logger.Printf("[TRACE] Asynchronous task detected, tracking task with HREF: %s", ref.String())
	taskClient := *op.client
	taskClient.APIVersion = op.apiVersion
	taskClient.customHeader = op.client.customHeader.Clone()
	if taskClient.customHeader == nil {
		taskClient.customHeader = make(http.Header)
	}
	for _, name := range []string{types.HeaderTenantContext, types.HeaderAuthContext} {
		taskClient.customHeader.Del(name)
		for _, value := range resp.Request.Header.Values(name) {
			taskClient.customHeader.Add(name, value)
		}
	}
	task := NewTask(&taskClient)
	task.Task.HREF = ref.String()
	return task, nil
}

func (op openApiOperation) runAsync(method string, payload interface{}) (*Task, error) {
	util.Logger.Printf("[TRACE] %s async request to endpoint %s with payload type %T and expected task response", method, op.url.String(), payload)
	op.expectedStatus = []int{http.StatusAccepted}
	resp, err := op.request(method, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logOpenApiResponse(resp)
	return op.task(resp)
}

// logOpenApiResponse logs a response whose body will not be decoded. The caller
// owns closing the body; a logging read error must not change the operation result.
func logOpenApiResponse(resp *http.Response) {
	if !util.LogHttpResponse && !debugShowResponseEnabled {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		util.Logger.Printf("[DEBUG] Error reading OpenAPI response body for logging: %s", err)
	}
	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(body))
	debugShowResponse(resp, body)
}

// write handles immediate resource responses and asynchronous creation, update or
// deletion. A nil output means wait for completion without retrieving a resource.
// Without an explicit expectedStatus, request accepts the legacy success statuses,
// including 204 No Content. Only resource responses contain JSON to decode.
func (op openApiOperation) write(method string, payload, outType interface{}) (http.Header, error) {
	if method == http.MethodDelete {
		util.Logger.Printf("[TRACE] Deleting item at endpoint %s", op.url.String())
	} else {
		util.Logger.Printf("[TRACE] %s request to endpoint %s with payload type %T and expected response of type %T", method, op.url.String(), payload, outType)
	}
	resp, err := op.request(method, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusAccepted {
		logOpenApiResponse(resp)
		task, err := op.task(resp)
		if err != nil {
			return nil, err
		}
		if err = task.WaitTaskCompletion(); err != nil {
			return nil, fmt.Errorf("error waiting completion of task (%s): %s", task.Task.HREF, err)
		}
		if outType != nil {
			if method == http.MethodPost {
				if task.Task.Owner == nil || task.Task.Owner.ID == "" {
					return nil, fmt.Errorf("could not find the created item UUID from the task '%s' so it can't be retrieved", task.Task.ID)
				}
				op.url, err = url.Parse(strings.TrimSuffix(op.url.String(), "/") + "/" + url.PathEscape(task.Task.Owner.ID))
				if err != nil {
					return nil, err
				}
			}
			op.query = nil
			op.expectedStatus = []int{http.StatusOK}
			if _, err = op.read(outType); err != nil {
				return nil, fmt.Errorf("error retrieving item after %s: %s", method, err)
			}
		}
	} else if outType != nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		util.Logger.Printf("[TRACE] Synchronous response detected (HTTP Status %d), decoding response into %T", resp.StatusCode, outType)
		if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
			return nil, fmt.Errorf("error decoding JSON response after %s: %s", method, err)
		}
	} else {
		util.Logger.Printf("[TRACE] Synchronous response detected (HTTP Status %d), no response object to decode", resp.StatusCode)
		logOpenApiResponse(resp)
	}
	return resp.Header, nil
}

// openApiGetAllPages is a recursive function that helps to accumulate responses from multiple pages for GET query. It
// works by at first crawling pages and accumulating all responses into []json.RawMessage (as strings). Because there is
// no intermediate unmarshalling to exact `outType` for every page it can unmarshal into direct `outType` supplied.
// outType must be a slice of object (e.g. []*types.OpenApiRole) because accumulated responses are in JSON list
//
// It follows pages in two ways:
// * Finds a 'nextPage' link and uses it to recursively crawl all pages (default for all, except for API bug)
// * Uses fields 'resultTotal', 'page', and 'pageSize' to calculate if it should crawl further on. It is only done
// because there is a BUG in API and in some endpoints it does not return 'nextPage' link as well as null 'pageCount'
//
// In general 'nextPage' header is preferred because some endpoints
// (like cloudapi/1.0.0/nsxTResources/importableTier0Routers) do not contain pagination details and nextPage header
// contains a base64 encoded data chunk via a supplied `cursor` field
// (e.g. ...importableTier0Routers?filter=_context==urn:vcloud:nsxtmanager:85aa2514-6a6f-4a32-8904-9695dc0f0298&
// cursor=eyJORVRXT1JLSU5HX0NVUlNPUl9PRkZTRVQiOiIwIiwicGFnZVNpemUiOjEsIk5FVFdPUktJTkdfQ1VSU09SIjoiMDAwMTMifQ==)
// The 'cursor' in example contains such values {"NETWORKING_CURSOR_OFFSET":"0","pageSize":1,"NETWORKING_CURSOR":"00013"}
func (client *Client) openApiGetAllPages(apiVersion string, urlRef *url.URL, queryParams url.Values, outType interface{}, responses []json.RawMessage, additionalHeader map[string]string) ([]json.RawMessage, error) {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	if responses == nil {
		responses = []json.RawMessage{}
	}

	op := client.openApiOperation(apiVersion, urlRefCopy, queryParams, additionalHeader)
	op.forbiddenAsNotFound = false
	pages := &types.OpenApiPages{}
	headers, err := op.read(pages)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON page: %s", err)
	}

	// Accumulate all responses in a single page as JSON text using json.RawMessage
	// After pages are unwrapped one can marshal response into specified type
	// singleQueryResponses := &json.RawMessage{}
	var singleQueryResponses []json.RawMessage
	if err = json.Unmarshal(pages.Values, &singleQueryResponses); err != nil {
		return nil, fmt.Errorf("error decoding values into accumulation type: %s", err)
	}
	responses = append(responses, singleQueryResponses...)

	// Check if there is still 'nextPage' linked and continue accumulating responses if so
	nextPageUrlRef, err := findRelLink("nextPage", headers)
	if err != nil && !IsNotFound(err) {
		return nil, fmt.Errorf("error looking for 'nextPage' in 'Link' header: %s", err)
	}

	if nextPageUrlRef != nil {
		responses, err = client.openApiGetAllPages(apiVersion, nextPageUrlRef, url.Values{}, outType, responses, additionalHeader)
		if err != nil {
			return nil, fmt.Errorf("got error on page %d: %s", pages.Page, err)
		}
	}

	// If nextPage header was not found, but we are not at the last page - the query URL should be forged manually to
	// overcome OpenAPI BUG when it does not return 'nextPage' header
	// Some API calls do not return `OpenApiPages` results at all (just values)
	// In some endpoints the page field is returned as `null` and this code block cannot handle it.
	if nextPageUrlRef == nil && pages.PageSize != 0 && pages.Page != 0 {
		// Next URL page ref was not found therefore one must double-check if it is not an API BUG. There are endpoints which
		// return only Total results and pageSize (not 'pageCount' and not 'nextPage' header)
		pageCount := pages.ResultTotal / pages.PageSize // This division returns number of "full pages" (containing 'pageSize' amount of results)
		if pages.ResultTotal%pages.PageSize > 0 {       // Check if is an incomplete page (containing less than 'pageSize' results)
			pageCount++ // Total pageCount is "number of complete pages + 1 incomplete" if it exists)
		}
		if pages.Page < pageCount {
			// Clone all originally supplied query parameters to avoid overwriting them
			urlQueryString := queryParams.Encode()
			urlQuery, err := url.ParseQuery(urlQueryString)
			if err != nil {
				return nil, fmt.Errorf("error cloning queryParams: %s", err)
			}

			// Increase page query by one to fetch "next" page
			urlQuery.Set("page", strconv.Itoa(pages.Page+1))

			responses, err = client.openApiGetAllPages(apiVersion, urlRefCopy, urlQuery, outType, responses, additionalHeader)
			if err != nil {
				return nil, fmt.Errorf("got error on page %d: %s", pages.Page, err)
			}
		}

	}

	return responses, nil
}

// newOpenApiRequest is a low level function used in upstream OpenAPI functions which handles logging and
// authentication for each API request
func (client *Client) newOpenApiRequest(apiVersion string, params url.Values, method string, reqUrl *url.URL, body io.Reader, additionalHeader map[string]string) *http.Request {
	// copy passed in URL ref so that it is not mutated
	reqUrlCopy := copyUrlRef(reqUrl)

	// Add the params to our URL
	reqUrlCopy.RawQuery += params.Encode()

	// If the body contains data - try to read all contents for logging and re-create another
	// io.Reader with all contents to use it down the line
	var readBody []byte
	var err error
	if body != nil {
		readBody, err = io.ReadAll(body)
		if err != nil {
			util.Logger.Printf("[DEBUG - newOpenApiRequest] error reading body: %s", err)
		}
		body = bytes.NewReader(readBody)
	}

	req, err := http.NewRequest(method, reqUrlCopy.String(), body)
	if err != nil {
		util.Logger.Printf("[DEBUG - newOpenApiRequest] error getting new request: %s", err)
	}

	if client.VCDAuthHeader != "" && client.VCDToken != "" {
		// Add the authorization header
		req.Header.Add(client.VCDAuthHeader, client.VCDToken)
		// The deprecated authorization token is 32 characters long
		// The bearer token is 612 characters long
		if len(client.VCDToken) > 32 {
			req.Header.Add("Authorization", "bearer "+client.VCDToken)
			req.Header.Add("X-Vmware-Vcloud-Token-Type", "Bearer")
		}
		// Add the Accept header for VCD
		acceptMime := types.JSONMime + ";version=" + apiVersion
		req.Header.Add("Accept", acceptMime)
	}

	for k, v := range client.customHeader {
		for _, v1 := range v {
			req.Header.Set(k, v1)
		}
	}
	for k, v := range additionalHeader {
		if strings.Contains(v, "{{MEDIA_TYPE}}") || strings.Contains(v, "{{API_VERSION}}") {
			v = strings.Replace(v, "{{MEDIA_TYPE}}", types.JSONMime, 1)
			v = strings.Replace(v, "{{API_VERSION}}", apiVersion, 1)
			req.Header.Set(k, v)
		} else {
			req.Header.Add(k, v)
		}
	}

	// Inject JSON mime type if there are no overwrites
	if req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", types.JSONMime)
	}

	setHttpUserAgent(client.UserAgent, req)
	setVcloudClientRequestId(client.RequestIdFunc, req)

	// Avoids passing data if the logging of requests is disabled
	if util.LogHttpRequest {
		payload := ""
		if req.ContentLength > 0 {
			payload = string(readBody)
		}
		util.ProcessRequestOutput(util.FuncNameCallStack(), method, reqUrlCopy.String(), payload, req)
		debugShowRequest(req, payload)
	}

	return req
}

// findRelLink looks for link to "nextPage" in "Link" header. It will return when first occurrence is found.
// Sample Link header:
// Link: [<https://HOSTNAME/cloudapi/1.0.0/auditTrail?sortAsc=&pageSize=25&sortDesc=&page=7>;rel="lastPage";
// type="application/json";model="AuditTrailEvents" <https://HOSTNAME/cloudapi/1.0.0/auditTrail?sortAsc=&pageSize=25&sortDesc=&page=2>;
// rel="nextPage";type="application/json";model="AuditTrailEvents"]
// Returns *url.Url or ErrorEntityNotFound
func findRelLink(relFieldName string, header http.Header) (*url.URL, error) {
	headerLinks := link.ParseHeader(header)

	for relKeyName, linkAddress := range headerLinks {
		switch {
		// When map key has more than one name (separated by space). In such cases it can have map key as
		// "lastPage nextPage" when nextPage==lastPage or similar and one specific field needs to be matched.
		case strings.Contains(relKeyName, " "):
			relNameSlice := strings.Split(relKeyName, " ")
			for _, oneRelName := range relNameSlice {
				if oneRelName == relFieldName {
					return url.Parse(linkAddress.String())
				}
			}
		case relKeyName == relFieldName:
			return url.Parse(linkAddress.String())
		}
	}

	return nil, ErrorEntityNotFound
}

// jsonRawMessagesToStrings converts []*json.RawMessage to []string
func jsonRawMessagesToStrings(messages []json.RawMessage) []string {
	resultString := make([]string, len(messages))
	for index, message := range messages {
		resultString[index] = string(message)
	}

	return resultString
}

// copyOrNewUrlValues either creates a copy of parameters or instantiates a new url.Values if nil parameters are
// supplied. It helps to avoid mutating supplied parameter when additional values must be injected internally.
func copyOrNewUrlValues(parameters url.Values) url.Values {
	parameterCopy := make(map[string][]string)

	// if supplied parameters are nil - we just return new initialized
	if parameters == nil {
		return parameterCopy
	}

	// Copy URL values
	for key, value := range parameters {
		parameterCopy[key] = value
	}

	return parameterCopy
}

// queryParameterFilterAnd is a helper to append "AND" clause to FIQL filter by using ';' (semicolon) if any values are
// already set in 'filter' value of parameters. If none existed before then 'filter' value will be set.
//
// Note. It does a copy of supplied 'parameters' value and does not mutate supplied original parameters.
func queryParameterFilterAnd(filter string, parameters url.Values) url.Values {
	newParameters := copyOrNewUrlValues(parameters)

	existingFilter := newParameters.Get("filter")
	if existingFilter == "" {
		newParameters.Set("filter", filter)
		return newParameters
	}

	newParameters.Set("filter", existingFilter+";"+filter)
	return newParameters
}

// defaultPageSize allows to set 'pageSize' query parameter to defaultPageSize if one is not already specified in
// url.Values while preserving all other supplied url.Values
func defaultPageSize(queryParams url.Values, defaultPageSize string) url.Values {
	newQueryParams := url.Values{}
	if queryParams != nil {
		newQueryParams = queryParams
	}

	if _, ok := newQueryParams["pageSize"]; !ok {
		newQueryParams.Set("pageSize", defaultPageSize)
	}

	return newQueryParams
}

// copyUrlRef creates a copy of URL reference by re-parsing it
func copyUrlRef(in *url.URL) *url.URL {
	// error is ignored because we expect to have correct URL supplied and this greatly simplifies code inside.
	newUrlRef, err := url.Parse(in.String())
	if err != nil {
		util.Logger.Printf("[DEBUG - copyUrlRef] error parsing URL: %s", err)
	}
	return newUrlRef
}

// shouldDoSlowSearch returns true and nil url.Values if the filter value contains commas, semicolons or asterisks,
// as the encoding is rejected by VCD with an error: QueryParseException: Cannot parse the supplied filter, so
// the caller knows that it needs to run a brute force search and NOT use filtering in any case.
// Also, url.QueryEscape as well as url.Values.Encode() both encode the space as a + character, so in this case
// it returns true and nil to specify a brute force search too. Reference to issue:
// https://github.com/golang/go/issues/4013
// https://github.com/czos/goamz/pull/11/files
// When this function returns false, it returns the url.Values that are not encoded, so make sure that the
// client encodes them before sending them.
func shouldDoSlowSearch(filterKey, filterValue string) (bool, url.Values) {
	if strings.Contains(filterValue, ",") || strings.Contains(filterValue, ";") ||
		strings.Contains(filterValue, " ") || strings.Contains(filterValue, "+") || strings.Contains(filterValue, "*") {
		return true, nil
	} else {
		params := url.Values{}
		params.Set("filter", fmt.Sprintf(filterKey+"==%s", filterValue))
		params.Set("filterEncoded", "true")
		return false, params
	}
}
