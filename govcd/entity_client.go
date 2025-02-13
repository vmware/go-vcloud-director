package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

func (client *Client) GetEntityUrl(endpoint ...string) (*url.URL, error) {
	path := fmt.Sprintf(ccitypes.KubernetesSubpath, client.VCDHREF.Scheme, client.VCDHREF.Host)
	path = path + strings.Join(endpoint, "")

	return url.ParseRequestURI(path)
}

func (client *Client) PostEntity(urlRef *url.URL, params url.Values, payload, outType interface{}) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Posting %s item to CCI endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRefCopy.String(), reflect.TypeOf(outType))

	if !client.IsTm() {
		return fmt.Errorf("CCI Client is not supported on this version")
	}

	resp, err := client.performEntityPostPut(http.MethodPost, urlRefCopy, params, payload, nil)
	if err != nil {
		return err
	}

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding CCI JSON response after POST: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

func (client *Client) GetEntity(urlRef *url.URL, params url.Values, outType interface{}, additionalHeader map[string]string) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Getting item from CCI endpoint %s with expected response of type %s",
		urlRefCopy.String(), reflect.TypeOf(outType))

	if !client.IsTm() {
		return fmt.Errorf("CCI Client is not supported on this version")
	}

	req := client.newEntityRequest(params, http.MethodGet, urlRefCopy, nil, additionalHeader)
	resp, err := client.Http.Do(req)
	if err != nil {
		return fmt.Errorf("error performing GET request to CCI %s: %s", urlRefCopy.String(), err)
	}

	// HTTP 404: Not Found - include sentinel log message ErrorEntityNotFound so that errors can be
	// checked with ContainsNotFound() and IsNotFound()
	if resp.StatusCode == http.StatusNotFound {
		err := ParseErr(types.BodyTypeJSON, resp, &ccitypes.ApiError{})
		closeErr := resp.Body.Close()
		return fmt.Errorf("%s: %s [body close error: %s]", ErrorEntityNotFound, err, closeErr)
	}

	// resp is ignored below because it is the same as above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &ccitypes.ApiError{})

	// Any other error occurred
	if err != nil {
		return fmt.Errorf("error in HTTP GET request for CCI: %s", err)
	}

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding JSON response after GET: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

func (client *Client) DeleteEntity(urlRef *url.URL, params url.Values, additionalHeader map[string]string) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Deleting item at CCI endpoint %s", urlRefCopy.String())

	if !client.IsTm() {
		return fmt.Errorf("CCI Client is not supported on this version")
	}

	// Perform request
	req := client.newEntityRequest(params, http.MethodDelete, urlRefCopy, nil, additionalHeader)
	resp, err := client.Http.Do(req)
	if err != nil {
		return err
	}

	bodyBytes, err := rewrapRespBodyNoopCloser(resp)
	if err != nil {
		return err
	}
	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, string(bodyBytes))
	debugShowResponse(resp, bodyBytes)

	// resp is ignored below because it would be the same as above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &ccitypes.ApiError{})
	if err != nil {
		return fmt.Errorf("error in HTTP DELETE request for CCI: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

func (client *Client) performEntityPostPut(httpMethod string, urlRef *url.URL, params url.Values, payload interface{}, additionalHeader map[string]string) (*http.Response, error) {
	// Marshal payload if we have one
	body := new(bytes.Buffer)
	if payload != nil {
		marshaledJson, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling CCI JSON data for %s request %s", httpMethod, err)
		}
		body = bytes.NewBuffer(marshaledJson)
	}

	req := client.newEntityRequest(params, httpMethod, urlRef, body, additionalHeader)
	resp, err := client.Http.Do(req)
	if err != nil {
		return nil, err
	}

	// resp is ignored below because it is the same the one above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &ccitypes.ApiError{})
	if err != nil {
		return nil, fmt.Errorf("error in HTTP %s CCI request: %s", httpMethod, err)
	}
	return resp, nil
}

func (client *Client) newEntityRequest(params url.Values, method string, reqUrl *url.URL, body io.Reader, additionalHeader map[string]string) *http.Request {
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
			util.Logger.Printf("[DEBUG - newCciRequest] error reading body: %s", err)
		}
		body = bytes.NewReader(readBody)
	}

	req, err := http.NewRequest(method, reqUrlCopy.String(), body)
	if err != nil {
		util.Logger.Printf("[DEBUG - newCciRequest] error getting new request: %s", err)
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
	}

	for k, v := range client.customHeader {
		for _, v1 := range v {
			req.Header.Set(k, v1)
		}
	}
	for k, v := range additionalHeader {
		req.Header.Add(k, v)
	}

	// Inject JSON mime type if there are no overwrites
	if req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", types.JSONMime)
	}

	setHttpUserAgent(client.UserAgent, req)
	// TODO - can we send a traceable request ID somewhere so that responses contain it for better log tracing
	// setVcloudClientRequestId(client.RequestIdFunc, req)

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
