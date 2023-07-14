/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// Package govcd provides a simple binding for VMware Cloud Director REST APIs.
package govcd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Client provides a client to VMware Cloud Director, values can be populated automatically using the Authenticate method.
type Client struct {
	APIVersion       string      // The API version required
	VCDToken         string      // Access Token (authorization header)
	VCDAuthHeader    string      // Authorization header
	VCDHREF          url.URL     // VCD API ENDPOINT
	Http             http.Client // HttpClient is the client to use. Default will be used if not provided.
	IsSysAdmin       bool        // flag if client is connected as system administrator
	UsingBearerToken bool        // flag if client is using a bearer token
	UsingAccessToken bool        // flag if client is using an API token

	// MaxRetryTimeout specifies a time limit (in seconds) for retrying requests made by the SDK
	// where VMware Cloud Director may take time to respond and retry mechanism is needed.
	// This must be >0 to avoid instant timeout errors.
	MaxRetryTimeout int

	// UseSamlAdfs specifies if SAML auth is used for authenticating vCD instead of local login.
	// The following conditions must be met so that authentication SAML authentication works:
	// * SAML IdP (Identity Provider) is Active Directory Federation Service (ADFS)
	// * Authentication endpoint "/adfs/services/trust/13/usernamemixed" must be enabled on ADFS
	// server
	UseSamlAdfs bool
	// CustomAdfsRptId allows to set custom Relaying Party Trust identifier. By default vCD Entity
	// ID is used as Relaying Party Trust identifier.
	CustomAdfsRptId string

	// UserAgent to send for API queries. Standard format is described as:
	// "User-Agent: <product> / <product-version> <comment>"
	UserAgent string

	// IgnoredMetadata allows to ignore metadata entries when using the methods defined in metadata_v2.go
	IgnoredMetadata []IgnoredMetadata

	supportedVersions SupportedVersions // Versions from /api/versions endpoint
	customHeader      http.Header
}

// AuthorizationHeader header key used by default to set the authorization token.
const AuthorizationHeader = "X-Vcloud-Authorization"

// BearerTokenHeader is the header key containing a bearer token
// #nosec G101 -- This is not a credential, it's just the header key
const BearerTokenHeader = "X-Vmware-Vcloud-Access-Token"

const ApiTokenHeader = "API-token"

// General purpose error to be used whenever an entity is not found from a "GET" request
// Allows a simpler checking of the call result
// such as
//
//	if err == ErrorEntityNotFound {
//	   // do what is needed in case of not found
//	}
var errorEntityNotFoundMessage = "[ENF] entity not found"
var ErrorEntityNotFound = fmt.Errorf(errorEntityNotFoundMessage)

// Triggers for debugging functions that show requests and responses
var debugShowRequestEnabled = os.Getenv("GOVCD_SHOW_REQ") != ""
var debugShowResponseEnabled = os.Getenv("GOVCD_SHOW_RESP") != ""

// Enables the debugging hook to show requests as they are processed.
//
//lint:ignore U1000 this function is used on request for debugging purposes
func enableDebugShowRequest() {
	debugShowRequestEnabled = true
}

// Disables the debugging hook to show requests as they are processed.
//
//lint:ignore U1000 this function is used on request for debugging purposes
func disableDebugShowRequest() {
	debugShowRequestEnabled = false
	err := os.Setenv("GOVCD_SHOW_REQ", "")
	if err != nil {
		util.Logger.Printf("[DEBUG - disableDebugShowRequest] error setting environment variable: %s", err)
	}
}

// Enables the debugging hook to show responses as they are processed.
//
//lint:ignore U1000 this function is used on request for debugging purposes
func enableDebugShowResponse() {
	debugShowResponseEnabled = true
}

// Disables the debugging hook to show responses as they are processed.
//
//lint:ignore U1000 this function is used on request for debugging purposes
func disableDebugShowResponse() {
	debugShowResponseEnabled = false
	err := os.Setenv("GOVCD_SHOW_RESP", "")
	if err != nil {
		util.Logger.Printf("[DEBUG - disableDebugShowResponse] error setting environment variable: %s", err)
	}
}

// On-the-fly debug hook. If either debugShowRequestEnabled or the environment
// variable "GOVCD_SHOW_REQ" are enabled, this function will show the contents
// of the request as it is being processed.
func debugShowRequest(req *http.Request, payload string) {
	if debugShowRequestEnabled {
		header := "[\n"
		for key, value := range util.SanitizedHeader(req.Header) {
			header += fmt.Sprintf("\t%s => %s\n", key, value)
		}
		header += "]\n"
		fmt.Println("** REQUEST **")
		fmt.Printf("time:    %s\n", time.Now().Format("2006-01-02T15:04:05.000Z"))
		fmt.Printf("method:  %s\n", req.Method)
		fmt.Printf("host:    %s\n", req.Host)
		fmt.Printf("length:  %d\n", req.ContentLength)
		fmt.Printf("URL:     %s\n", req.URL.String())
		fmt.Printf("header:  %s\n", header)
		fmt.Printf("payload: %s\n", payload)
		fmt.Println()
	}
}

// On-the-fly debug hook. If either debugShowResponseEnabled or the environment
// variable "GOVCD_SHOW_RESP" are enabled, this function will show the contents
// of the response as it is being processed.
func debugShowResponse(resp *http.Response, body []byte) {
	if debugShowResponseEnabled {
		fmt.Println("## RESPONSE ##")
		fmt.Printf("time:   %s\n", time.Now().Format("2006-01-02T15:04:05.000Z"))
		fmt.Printf("status: %d - %s \n", resp.StatusCode, resp.Status)
		fmt.Printf("length: %d\n", resp.ContentLength)
		fmt.Printf("header: %v\n", util.SanitizedHeader(resp.Header))
		fmt.Printf("body:   %s\n", body)
		fmt.Println()
	}
}

// IsNotFound is a convenience function, similar to os.IsNotExist that checks whether a given error
// is a "Not found" error, such as
//
//	if isNotFound(err) {
//	   // do what is needed in case of not found
//	}
func IsNotFound(err error) bool {
	return err != nil && err == ErrorEntityNotFound
}

// ContainsNotFound is a convenience function, similar to os.IsNotExist that checks whether a given error
// contains a "Not found" error. It is almost the same as `IsNotFound` but checks if an error contains substring
// ErrorEntityNotFound
func ContainsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), ErrorEntityNotFound.Error())
}

// NewRequestWitNotEncodedParams allows passing complex values params that shouldn't be encoded like for queries. e.g. /query?filter=name=foo
func (client *Client) NewRequestWitNotEncodedParams(params map[string]string, notEncodedParams map[string]string, method string, reqUrl url.URL, body io.Reader) *http.Request {
	return client.NewRequestWitNotEncodedParamsWithApiVersion(params, notEncodedParams, method, reqUrl, body, client.APIVersion)
}

// NewRequestWitNotEncodedParamsWithApiVersion allows passing complex values params that shouldn't be encoded like for queries. e.g. /query?filter=name=foo
// * params - request parameters
// * notEncodedParams - request parameters which will be added not encoded
// * method - request type
// * reqUrl - request url
// * body - request body
// * apiVersion - provided Api version overrides default Api version value used in request parameter
func (client *Client) NewRequestWitNotEncodedParamsWithApiVersion(params map[string]string, notEncodedParams map[string]string, method string, reqUrl url.URL, body io.Reader, apiVersion string) *http.Request {
	return client.newRequest(params, notEncodedParams, method, reqUrl, body, apiVersion, nil)
}

// newRequest is the parent of many "specific" "NewRequest" functions.
// Note. It is kept private to avoid breaking public API on every new field addition.
func (client *Client) newRequest(params map[string]string, notEncodedParams map[string]string, method string, reqUrl url.URL, body io.Reader, apiVersion string, additionalHeader http.Header) *http.Request {
	reqValues := url.Values{}

	// Build up our request parameters
	for key, value := range params {
		reqValues.Add(key, value)
	}

	// Add the params to our URL
	reqUrl.RawQuery = reqValues.Encode()

	for key, value := range notEncodedParams {
		if key != "" && value != "" {
			reqUrl.RawQuery += "&" + key + "=" + value
		}
	}

	// If the body contains data - try to read all contents for logging and re-create another
	// io.Reader with all contents to use it down the line
	var readBody []byte
	var err error
	if body != nil {
		readBody, err = io.ReadAll(body)
		if err != nil {
			util.Logger.Printf("[DEBUG - newRequest] error reading body: %s", err)
		}
		body = bytes.NewReader(readBody)
	}

	req, err := http.NewRequest(method, reqUrl.String(), body)
	if err != nil {
		util.Logger.Printf("[DEBUG - newRequest] error getting new request: %s", err)
	}

	if client.VCDAuthHeader != "" && client.VCDToken != "" {
		// Add the authorization header
		req.Header.Add(client.VCDAuthHeader, client.VCDToken)
	}
	if (client.VCDAuthHeader != "" && client.VCDToken != "") ||
		(additionalHeader != nil && additionalHeader.Get("Authorization") != "") {
		// Add the Accept header for VCD
		req.Header.Add("Accept", "application/*+xml;version="+apiVersion)
	}
	// The deprecated authorization token is 32 characters long
	// The bearer token is 612 characters long
	if len(client.VCDToken) > 32 {
		req.Header.Add("X-Vmware-Vcloud-Token-Type", "Bearer")
		req.Header.Add("Authorization", "bearer "+client.VCDToken)
	}

	// Merge in additional headers before logging if anywhere specified in additionalHeader
	// parameter
	if len(additionalHeader) > 0 {
		for headerName, headerValueSlice := range additionalHeader {
			for _, singleHeaderValue := range headerValueSlice {
				req.Header.Set(headerName, singleHeaderValue)
			}
		}
	}
	if client.customHeader != nil {
		for k, v := range client.customHeader {
			for _, v1 := range v {
				req.Header.Add(k, v1)
			}
		}
	}

	setHttpUserAgent(client.UserAgent, req)

	// Avoids passing data if the logging of requests is disabled
	if util.LogHttpRequest {
		payload := ""
		if req.ContentLength > 0 {
			payload = string(readBody)
		}
		util.ProcessRequestOutput(util.FuncNameCallStack(), method, reqUrl.String(), payload, req)
		debugShowRequest(req, payload)
	}

	return req

}

// NewRequest creates a new HTTP request and applies necessary auth headers if set.
func (client *Client) NewRequest(params map[string]string, method string, reqUrl url.URL, body io.Reader) *http.Request {
	return client.NewRequestWitNotEncodedParams(params, nil, method, reqUrl, body)
}

// NewRequestWithApiVersion creates a new HTTP request and applies necessary auth headers if set.
// Allows to override default request API Version
func (client *Client) NewRequestWithApiVersion(params map[string]string, method string, reqUrl url.URL, body io.Reader, apiVersion string) *http.Request {
	return client.NewRequestWitNotEncodedParamsWithApiVersion(params, nil, method, reqUrl, body, apiVersion)
}

// ParseErr takes an error XML resp, error interface for unmarshalling and returns a single string for
// use in error messages.
func ParseErr(bodyType types.BodyType, resp *http.Response, errType error) error {
	// if there was an error decoding the body, just return that
	if err := decodeBody(bodyType, resp, errType); err != nil {
		util.Logger.Printf("[ParseErr]: unhandled response <--\n%+v\n-->\n", resp)
		return fmt.Errorf("[ParseErr]: error parsing error body for non-200 request: %s (%+v)", err, resp)
	}

	// response body maybe empty for some error, such like 416, 400
	if errType.Error() == "API Error: 0: " {
		errType = fmt.Errorf(resp.Status)
	}

	return errType
}

// decodeBody is used to decode a response body of types.BodyType
func decodeBody(bodyType types.BodyType, resp *http.Response, out interface{}) error {
	body, err := io.ReadAll(resp.Body)

	// In case of JSON, body does not have indents in response therefore it must be indented
	if bodyType == types.BodyTypeJSON {
		body, err = indentJsonBody(body)
		if err != nil {
			return err
		}
	}

	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(body))
	if err != nil {
		return err
	}

	debugShowResponse(resp, body)

	// only attempt to unmarshal if body is not empty
	if len(body) > 0 {
		switch bodyType {
		case types.BodyTypeXML:
			if err = xml.Unmarshal(body, &out); err != nil {
				return err
			}
		case types.BodyTypeJSON:
			if err = json.Unmarshal(body, &out); err != nil {
				return err
			}

		default:
			panic(fmt.Sprintf("unknown body type: %d", bodyType))
		}
	}

	return nil
}

// indentJsonBody indents raw JSON body for easier readability
func indentJsonBody(body []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error indenting response JSON: %s", err)
	}
	body = prettyJSON.Bytes()
	return body, nil
}

// checkResp wraps http.Client.Do() and verifies the request, if status code
// is 2XX it passes back the response, if it's a known invalid status code it
// parses the resultant XML error and returns a descriptive error, if the
// status code is not handled it returns a generic error with the status code.
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	return checkRespWithErrType(types.BodyTypeXML, resp, err, &types.Error{})
}

// checkRespWithErrType allows to specify custom error errType for checkResp unmarshaling
// the error.
func checkRespWithErrType(bodyType types.BodyType, resp *http.Response, err, errType error) (*http.Response, error) {
	if err != nil {
		return resp, err
	}

	switch resp.StatusCode {
	// Valid request, return the response.
	case
		http.StatusOK,        // 200
		http.StatusCreated,   // 201
		http.StatusAccepted,  // 202
		http.StatusNoContent, // 204
		http.StatusFound:     // 302
		return resp, nil
	// Invalid request, parse the XML error returned and return it.
	case
		http.StatusBadRequest,                   // 400
		http.StatusUnauthorized,                 // 401
		http.StatusForbidden,                    // 403
		http.StatusNotFound,                     // 404
		http.StatusMethodNotAllowed,             // 405
		http.StatusNotAcceptable,                // 406
		http.StatusProxyAuthRequired,            // 407
		http.StatusRequestTimeout,               // 408
		http.StatusConflict,                     // 409
		http.StatusGone,                         // 410
		http.StatusLengthRequired,               // 411
		http.StatusPreconditionFailed,           // 412
		http.StatusRequestEntityTooLarge,        // 413
		http.StatusRequestURITooLong,            // 414
		http.StatusUnsupportedMediaType,         // 415
		http.StatusRequestedRangeNotSatisfiable, // 416
		http.StatusLocked,                       // 423
		http.StatusFailedDependency,             // 424
		http.StatusUpgradeRequired,              // 426
		http.StatusPreconditionRequired,         // 428
		http.StatusTooManyRequests,              // 429
		http.StatusRequestHeaderFieldsTooLarge,  // 431
		http.StatusUnavailableForLegalReasons,   // 451
		http.StatusInternalServerError,          // 500
		http.StatusServiceUnavailable,           // 503
		http.StatusGatewayTimeout:               // 504
		return nil, ParseErr(bodyType, resp, errType)
	// Unhandled response.
	default:
		return nil, fmt.Errorf("unhandled API response, please report this issue, status code: %s", resp.Status)
	}
}

// ExecuteTaskRequest helper function creates request, runs it, checks response and parses task from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// E.g. client.ExecuteTaskRequest(updateDiskLink.HREF, http.MethodPut, updateDiskLink.Type, "error updating disk: %s", xmlPayload)
func (client *Client) ExecuteTaskRequest(pathURL, requestType, contentType, errorMessage string, payload interface{}) (Task, error) {
	return client.executeTaskRequest(pathURL, requestType, contentType, errorMessage, payload, client.APIVersion)
}

// ExecuteTaskRequestWithApiVersion helper function creates request, runs it, checks response and parses task from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// apiVersion - api version which will be used in request
// E.g. client.ExecuteTaskRequest(updateDiskLink.HREF, http.MethodPut, updateDiskLink.Type, "error updating disk: %s", xmlPayload)
func (client *Client) ExecuteTaskRequestWithApiVersion(pathURL, requestType, contentType, errorMessage string, payload interface{}, apiVersion string) (Task, error) {
	return client.executeTaskRequest(pathURL, requestType, contentType, errorMessage, payload, apiVersion)
}

// Helper function creates request, runs it, checks response and parses task from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// apiVersion - api version which will be used in request
// E.g. client.ExecuteTaskRequest(updateDiskLink.HREF, http.MethodPut, updateDiskLink.Type, "error updating disk: %s", xmlPayload)
func (client *Client) executeTaskRequest(pathURL, requestType, contentType, errorMessage string, payload interface{}, apiVersion string) (Task, error) {

	if !isMessageWithPlaceHolder(errorMessage) {
		return Task{}, fmt.Errorf("error message has to include place holder for error")
	}

	resp, err := executeRequestWithApiVersion(pathURL, requestType, contentType, payload, client, apiVersion)
	if err != nil {
		return Task{}, fmt.Errorf(errorMessage, err)
	}

	task := NewTask(client)

	if err = decodeBody(types.BodyTypeXML, resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return Task{}, fmt.Errorf(errorMessage, err)
	}

	// The request was successful
	return *task, nil
}

// ExecuteRequestWithoutResponse helper function creates request, runs it, checks response and do not expect any values from it.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// E.g. client.ExecuteRequestWithoutResponse(catalogItemHREF.String(), http.MethodDelete, "", "error deleting Catalog item: %s", nil)
func (client *Client) ExecuteRequestWithoutResponse(pathURL, requestType, contentType, errorMessage string, payload interface{}) error {
	return client.executeRequestWithoutResponse(pathURL, requestType, contentType, errorMessage, payload, client.APIVersion)
}

// ExecuteRequestWithoutResponseWithApiVersion helper function creates request, runs it, checks response and do not expect any values from it.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// apiVersion - api version which will be used in request
// E.g. client.ExecuteRequestWithoutResponse(catalogItemHREF.String(), http.MethodDelete, "", "error deleting Catalog item: %s", nil)
func (client *Client) ExecuteRequestWithoutResponseWithApiVersion(pathURL, requestType, contentType, errorMessage string, payload interface{}, apiVersion string) error {
	return client.executeRequestWithoutResponse(pathURL, requestType, contentType, errorMessage, payload, apiVersion)
}

// Helper function creates request, runs it, checks response and do not expect any values from it.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// apiVersion - api version which will be used in request
// E.g. client.ExecuteRequestWithoutResponse(catalogItemHREF.String(), http.MethodDelete, "", "error deleting Catalog item: %s", nil)
func (client *Client) executeRequestWithoutResponse(pathURL, requestType, contentType, errorMessage string, payload interface{}, apiVersion string) error {

	if !isMessageWithPlaceHolder(errorMessage) {
		return fmt.Errorf("error message has to include place holder for error")
	}

	resp, err := executeRequestWithApiVersion(pathURL, requestType, contentType, payload, client, apiVersion)
	if err != nil {
		return fmt.Errorf(errorMessage, err)
	}

	// log response explicitly because decodeBody() was not triggered
	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, fmt.Sprintf("%s", resp.Body))

	debugShowResponse(resp, []byte("SKIPPED RESPONSE"))
	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	// The request was successful
	return nil
}

// ExecuteRequest helper function creates request, runs it, check responses and parses out interface from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// out - structure to be used for unmarshalling xml
// E.g. 	unmarshalledAdminOrg := &types.AdminOrg{}
// client.ExecuteRequest(adminOrg.AdminOrg.HREF, http.MethodGet, "", "error refreshing organization: %s", nil, unmarshalledAdminOrg)
func (client *Client) ExecuteRequest(pathURL, requestType, contentType, errorMessage string, payload, out interface{}) (*http.Response, error) {
	return client.executeRequest(pathURL, requestType, contentType, errorMessage, payload, out, client.APIVersion)
}

// ExecuteRequestWithApiVersion helper function creates request, runs it, check responses and parses out interface from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// out - structure to be used for unmarshalling xml
// apiVersion - api version which will be used in request
// E.g. 	unmarshalledAdminOrg := &types.AdminOrg{}
// client.ExecuteRequest(adminOrg.AdminOrg.HREF, http.MethodGet, "", "error refreshing organization: %s", nil, unmarshalledAdminOrg)
func (client *Client) ExecuteRequestWithApiVersion(pathURL, requestType, contentType, errorMessage string, payload, out interface{}, apiVersion string) (*http.Response, error) {
	return client.executeRequest(pathURL, requestType, contentType, errorMessage, payload, out, apiVersion)
}

// Helper function creates request, runs it, check responses and parses out interface from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// out - structure to be used for unmarshalling xml
// apiVersion - api version which will be used in request
// E.g. 	unmarshalledAdminOrg := &types.AdminOrg{}
// client.ExecuteRequest(adminOrg.AdminOrg.HREF, http.MethodGet, "", "error refreshing organization: %s", nil, unmarshalledAdminOrg)
func (client *Client) executeRequest(pathURL, requestType, contentType, errorMessage string, payload, out interface{}, apiVersion string) (*http.Response, error) {

	if !isMessageWithPlaceHolder(errorMessage) {
		return &http.Response{}, fmt.Errorf("error message has to include place holder for error")
	}

	resp, err := executeRequestWithApiVersion(pathURL, requestType, contentType, payload, client, apiVersion)
	if err != nil {
		return resp, fmt.Errorf(errorMessage, err)
	}

	if err = decodeBody(types.BodyTypeXML, resp, out); err != nil {
		return resp, fmt.Errorf("error decoding response: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return resp, fmt.Errorf("error closing response body: %s", err)
	}

	// The request was successful
	return resp, nil
}

// ExecuteRequestWithCustomError sends the request and checks for 2xx response. If the returned status code
// was not as expected - the returned error will be unmarshalled to `errType` which implements Go's standard `error`
// interface.
func (client *Client) ExecuteRequestWithCustomError(pathURL, requestType, contentType, errorMessage string,
	payload interface{}, errType error) (*http.Response, error) {
	return client.ExecuteParamRequestWithCustomError(pathURL, map[string]string{}, requestType, contentType,
		errorMessage, payload, errType)
}

// ExecuteParamRequestWithCustomError behaves exactly like ExecuteRequestWithCustomError but accepts
// query parameter specification
func (client *Client) ExecuteParamRequestWithCustomError(pathURL string, params map[string]string,
	requestType, contentType, errorMessage string, payload interface{}, errType error) (*http.Response, error) {
	if !isMessageWithPlaceHolder(errorMessage) {
		return &http.Response{}, fmt.Errorf("error message has to include place holder for error")
	}

	resp, err := executeRequestCustomErr(pathURL, params, requestType, contentType, payload, client, errType, client.APIVersion)
	if err != nil {
		return &http.Response{}, fmt.Errorf(errorMessage, err)
	}

	// read from resp.Body io.Reader for debug output if it has body
	var bodyBytes []byte
	if resp.Body != nil {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return &http.Response{}, fmt.Errorf("could not read response body: %s", err)
		}
		// Restore the io.ReadCloser to its original state with no-op closer
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(bodyBytes))
	debugShowResponse(resp, bodyBytes)

	return resp, nil
}

// executeRequest does executeRequestCustomErr and checks for vCD errors in API response
func executeRequestWithApiVersion(pathURL, requestType, contentType string, payload interface{}, client *Client, apiVersion string) (*http.Response, error) {
	return executeRequestCustomErr(pathURL, map[string]string{}, requestType, contentType, payload, client, &types.Error{}, apiVersion)
}

// executeRequestCustomErr performs request and unmarshals API error to errType if not 2xx status was returned
func executeRequestCustomErr(pathURL string, params map[string]string, requestType, contentType string, payload interface{}, client *Client, errType error, apiVersion string) (*http.Response, error) {
	requestURI, err := url.ParseRequestURI(pathURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse path request URI '%s': %s", pathURL, err)
	}

	var req *http.Request
	switch {
	// Only send data (and xml.Header) if the payload is actually provided to avoid sending empty body with XML header
	// (some Web Application Firewalls block requests when empty XML header is set but not body provided)
	case payload != nil:
		marshaledXml, err := xml.MarshalIndent(payload, "  ", "    ")
		if err != nil {
			return &http.Response{}, fmt.Errorf("error marshalling xml data %s", err)
		}
		body := bytes.NewBufferString(xml.Header + string(marshaledXml))

		req = client.NewRequestWithApiVersion(params, requestType, *requestURI, body, apiVersion)

	default:
		req = client.NewRequestWithApiVersion(params, requestType, *requestURI, nil, apiVersion)
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	setHttpUserAgent(client.UserAgent, req)

	resp, err := client.Http.Do(req)
	if err != nil {
		return resp, err
	}

	return checkRespWithErrType(types.BodyTypeXML, resp, err, errType)
}

// setHttpUserAgent adds User-Agent string to HTTP request. When supplied string is empty - header will not be set
func setHttpUserAgent(userAgent string, req *http.Request) {
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
}

func isMessageWithPlaceHolder(message string) bool {
	err := fmt.Errorf(message, "test error")
	return !strings.Contains(err.Error(), "%!(EXTRA")
}

// combinedTaskErrorMessage is a general purpose function
// that returns the contents of the operation error and, if found, the error
// returned by the associated task
func combinedTaskErrorMessage(task *types.Task, err error) string {
	extendedError := err.Error()
	if task.Error != nil {
		extendedError = fmt.Sprintf("operation error: %s - task error: [%d - %s] %s",
			err, task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
	}
	return extendedError
}

// addrOf is a generic function to return the address of a variable
// Note. It is mainly meant for converting literal values to pointers (e.g. `addrOf(true)`)
// and not getting the address of a variable (e.g. `addrOf(variable)`)
func addrOf[T any](variable T) *T {
	return &variable
}

// IsUuid returns true if the identifier is a bare UUID
func IsUuid(identifier string) bool {
	reUuid := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	return reUuid.MatchString(identifier)
}

// isUrn validates if supplied identifier is of URN format (e.g. urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc)
// it checks for the following criteria:
// 1. idenfifier is not empty
// 2. identifier has 4 elements separated by ':'
// 3. element 1 is 'urn' and element 4 is valid UUID
func isUrn(identifier string) bool {
	if identifier == "" {
		return false
	}

	ss := strings.Split(identifier, ":")
	if len(ss) != 4 {
		return false
	}

	if ss[0] != "urn" && !IsUuid(ss[3]) {
		return false
	}

	return true
}

// BuildUrnWithUuid helps to build valid URNs where APIs require URN format, but other API responds with UUID (or
// extracted from HREF)
func BuildUrnWithUuid(urnPrefix, uuid string) (string, error) {
	if !IsUuid(uuid) {
		return "", fmt.Errorf("supplied uuid '%s' is not valid UUID", uuid)
	}

	urn := urnPrefix + uuid
	if !isUrn(urn) {
		return "", fmt.Errorf("failed building valid URN '%s'", urn)
	}

	return urn, nil
}

// takeFloatAddress is a helper that returns the address of an `float64`
func takeFloatAddress(x float64) *float64 {
	return &x
}

// SetCustomHeader adds custom HTTP header values to a client
func (client *Client) SetCustomHeader(values map[string]string) {
	if len(client.customHeader) == 0 {
		client.customHeader = make(http.Header)
	}
	for k, v := range values {
		client.customHeader.Add(k, v)
	}
}

// RemoveCustomHeader remove custom header values from the client
func (client *Client) RemoveCustomHeader() {
	if client.customHeader != nil {
		client.customHeader = nil
	}
}

// RemoveProvidedCustomHeaders removes custom header values from the client
func (client *Client) RemoveProvidedCustomHeaders(values map[string]string) {
	if client.customHeader != nil {
		for k := range values {
			client.customHeader.Del(k)
		}
	}
}

// Retrieves the administrator URL of a given HREF
func getAdminURL(href string) string {
	adminApi := "/api/admin/"
	if strings.Contains(href, adminApi) {
		return href
	}
	return strings.ReplaceAll(href, "/api/", adminApi)
}

// Retrieves the admin extension URL of a given HREF
func getAdminExtensionURL(href string) string {
	adminExtensionApi := "/api/admin/extension/"
	if strings.Contains(href, adminExtensionApi) {
		return href
	}
	return strings.ReplaceAll(getAdminURL(href), "/api/admin/", adminExtensionApi)
}

// TestConnection calls API to test a connection against a VCD, including SSL handshake and hostname verification.
func (client *Client) TestConnection(testConnection types.TestConnection) (*types.TestConnectionResult, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTestConnection

	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnTestConnectionResult := &types.TestConnectionResult{
		TargetProbe: &types.ProbeResult{},
		ProxyProbe:  &types.ProbeResult{},
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, testConnection, returnTestConnectionResult, nil)
	if err != nil {
		return nil, fmt.Errorf("error performing test connection: %s", err)
	}

	return returnTestConnectionResult, nil
}

// TestConnectionWithDefaults calls TestConnection given a subscriptionURL. The rest of parameters are set as default.
// It returns whether it could reach the server and establish SSL connection or not.
func (client *Client) TestConnectionWithDefaults(subscriptionURL string) (bool, error) {
	if subscriptionURL == "" {
		return false, fmt.Errorf("TestConnectionWithDefaults needs to be passed a host. i.e. my-host.vmware.com")
	}

	url, err := url.Parse(subscriptionURL)
	if err != nil {
		return false, fmt.Errorf("unable to parse URL - %s", err)
	}

	// Get port
	var port int
	if v := url.Port(); v != "" {
		port, err = strconv.Atoi(v)
		if err != nil {
			return false, fmt.Errorf("couldn't parse port provided - %s", err)
		}
	} else {
		switch url.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		}
	}

	testConnectionConfig := types.TestConnection{
		Host:    url.Hostname(),
		Port:    port,
		Secure:  addrOf(true), // Default value used by VCD UI
		Timeout: 30,           // Default value used by VCD UI
	}

	testConnectionResult, err := client.TestConnection(testConnectionConfig)
	if err != nil {
		return false, err
	}

	if !testConnectionResult.TargetProbe.CanConnect {
		return false, fmt.Errorf("the remote host is not reachable")
	}

	if !testConnectionResult.TargetProbe.SSLHandshake {
		return true, fmt.Errorf("unsupported or unrecognized SSL message")
	}

	return true, nil
}

// buildUrl uses the Client base URL to create a customised URL
func (client *Client) buildUrl(elements ...string) (string, error) {
	baseUrl := client.VCDHREF.String()
	if !IsValidUrl(baseUrl) {
		return "", fmt.Errorf("incorrect URL %s", client.VCDHREF.String())
	}
	if strings.HasSuffix(baseUrl, "/") {
		baseUrl = strings.TrimRight(baseUrl, "/")
	}
	if strings.HasSuffix(baseUrl, "/api") {
		baseUrl = strings.TrimRight(baseUrl, "/api")
	}
	return url.JoinPath(baseUrl, elements...)
}

// ---------------------------------------------------------------------
// The following functions are needed to avoid strict Coverity warnings
// ---------------------------------------------------------------------

// urlParseRequestURI returns a URL, discarding the error
func urlParseRequestURI(href string) *url.URL {
	apiEndpoint, err := url.ParseRequestURI(href)
	if err != nil {
		util.Logger.Printf("[DEBUG - urlParseRequestURI] error parsing request URI: %s", err)
	}
	return apiEndpoint
}

// safeClose closes a file and logs the error, if any. This can be used instead of file.Close()
func safeClose(file *os.File) {
	if err := file.Close(); err != nil {
		util.Logger.Printf("Error closing file: %s\n", err)
	}
}

// isSuccessStatus returns true if the given status code is between 200 and 300
func isSuccessStatus(statusCode int) bool {
	if statusCode >= http.StatusOK && // 200
		statusCode < http.StatusMultipleChoices { // 300
		return true
	}
	return false
}
