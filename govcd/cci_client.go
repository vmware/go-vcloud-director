package govcd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// constants that are used for tracking entity state after it is being created
const (
	stateWaitTimeout    = 1 * time.Hour
	stateWaitDelay      = 5 * time.Second
	stateWaitMinTimeout = 5 * time.Second
)

// CciClient
type CciClient struct {
	VCDClient *VCDClient
}

func (cciClient *CciClient) IsSupported() bool {
	return cciClient.VCDClient.Client.APIVCDMaxVersionIs(">= 40")
}

func (tpClient *CciClient) GetCciUrl(endpoint ...string) (*url.URL, error) {
	path := fmt.Sprintf(ccitypes.CciKubernetesSubpath, tpClient.VCDClient.Client.VCDHREF.Scheme, tpClient.VCDClient.Client.VCDHREF.Host)
	path = path + strings.Join(endpoint, "")

	return url.ParseRequestURI(path)
}

type urlRetriever func(interface{}) (*url.URL, error)

func (cciClient *CciClient) PostItemAsync(urlRef *url.URL, responseUrlFunc urlRetriever, params url.Values, payload, outType interface{}) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Posting %s item to CCI endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRefCopy.String(), reflect.TypeOf(outType))

	if !cciClient.IsSupported() {
		return fmt.Errorf("CCI Client is not supported on this version")
	}

	resp, err := cciClient.cciPerformPostPut(http.MethodPost, urlRefCopy, params, payload, nil)
	if err != nil {
		return err
	}

	// if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
	// 	util.Logger.Printf("[TRACE] Synchronous task expected (HTTP status code 200 or 201). Got %d", resp.StatusCode)
	// 	return fmt.Errorf("POST request expected sync task (HTTP response 200 or 201), got %d", resp.StatusCode)
	// }

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding CCI JSON response after POST: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	responseUrl, err := responseUrlFunc(outType)
	if err != nil {
		return fmt.Errorf("error getting CCI entity URL after creation: %s", err)
	}

	// WAIT for entity to transition from "CREATING" or "WAITING" to "CREATED"
	_, err = cciClient.waitForState(context.TODO(), "label", responseUrl, []string{"CREATING", "WAITING"}, []string{"CREATED"})
	if err != nil {
		return fmt.Errorf("error waiting for CCI entity state: %s", err)
	}

	// Retrieve the final state of original entity
	err = cciClient.GetItem(responseUrl, nil, outType, nil)
	if err != nil {
		return fmt.Errorf("error retrieving final CCI entity after creation: %s", err)
	}

	return nil
}

func (cciClient *CciClient) PostItemSync(urlRef *url.URL, params url.Values, payload, outType interface{}) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Posting %s item to CCI endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRefCopy.String(), reflect.TypeOf(outType))

	if !cciClient.IsSupported() {
		return fmt.Errorf("CCI Client is not supported on this version")
	}

	resp, err := cciClient.cciPerformPostPut(http.MethodPost, urlRefCopy, params, payload, nil)
	if err != nil {
		return err
	}

	// if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
	// 	util.Logger.Printf("[TRACE] Synchronous task expected (HTTP status code 200 or 201). Got %d", resp.StatusCode)
	// 	return fmt.Errorf("POST request expected sync task (HTTP response 200 or 201), got %d", resp.StatusCode)
	// }

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding CCI JSON response after POST: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

func (cciClient *CciClient) GetItem(urlRef *url.URL, params url.Values, outType interface{}, additionalHeader map[string]string) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Getting item from CCI endpoint %s with expected response of type %s",
		urlRefCopy.String(), reflect.TypeOf(outType))

	if !cciClient.IsSupported() {
		return fmt.Errorf("CCI Client is not supported on this version")
	}

	req := cciClient.newCciRequest(params, http.MethodGet, urlRefCopy, nil, additionalHeader)
	resp, err := cciClient.VCDClient.Client.Http.Do(req)
	if err != nil {
		return fmt.Errorf("error performing GET request to CCI %s: %s", urlRefCopy.String(), err)
	}

	// HTTP 404: Not Found - include sentinel log message ErrorEntityNotFound so that errors can be
	// checked with ContainsNotFound() and IsNotFound()

	if resp.StatusCode == http.StatusNotFound {
		err := ParseErr(types.BodyTypeJSON, resp, &ccitypes.CciApiError{})
		closeErr := resp.Body.Close()
		return fmt.Errorf("%s: %s [body close error: %s]", ErrorEntityNotFound, err, closeErr)
	}

	// resp is ignored below because it is the same as above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &ccitypes.CciApiError{})

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

func (cciClient *CciClient) DeleteItem(urlRef *url.URL, params url.Values, additionalHeader map[string]string) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Deleting item at CCI endpoint %s", urlRefCopy.String())

	if !cciClient.IsSupported() {
		return fmt.Errorf("CCI Client is not supported on this version")
	}

	// Perform request
	req := cciClient.newCciRequest(params, http.MethodDelete, urlRefCopy, nil, additionalHeader)
	resp, err := cciClient.VCDClient.Client.Http.Do(req)
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
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &ccitypes.CciApiError{})
	if err != nil {
		return fmt.Errorf("error in HTTP DELETE request for CCI: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	_, err = cciClient.waitForState(context.TODO(), "label", urlRefCopy, []string{"DELETING", "WAITING"}, []string{"DELETED"})
	if err != nil {
		return fmt.Errorf("error waiting for CCI entity state: %s", err)
	}

	return nil
}

// TODO - can we do this quite cheap without needing to pull in Hashicorp package at SDK level
func (cciClient *CciClient) waitForState(ctx context.Context, entityLabel string, addr *url.URL, pendingStates, targetStates []string) (any, error) {
	stateChangeFunc := retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (any, string, error) {
			cciEntity, err := cciClient.getAnyCciState(addr)
			if err != nil {
				if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
					return "", "DELETED", nil
				}
				return nil, "", err
			}

			log.Printf("[DEBUG] %s %s current phase is %s", entityLabel, cciEntity.GetName(), cciEntity.Status.Phase)
			if strings.EqualFold(cciEntity.Status.Phase, "ERROR") {
				return nil, "", fmt.Errorf("%s %s is in an ERROR state", entityLabel, cciEntity.GetName())
			}

			return cciEntity, strings.ToUpper(cciEntity.Status.Phase), nil
		},
		Timeout:    stateWaitTimeout,
		Delay:      stateWaitDelay,
		MinTimeout: stateWaitMinTimeout,
	}

	result, err := stateChangeFunc.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error waiting entity %s state to transition from '%s' to '%s': %s",
			entityLabel, strings.Join(pendingStates, ","), strings.Join(targetStates, ","), err)
	}

	return result, nil
}

func (cciClient *CciClient) getAnyCciState(urlRef *url.URL) (*ccitypes.CciEntityStatus, error) {
	entityStatus := ccitypes.CciEntityStatus{}

	err := cciClient.GetItem(urlRef, nil, &entityStatus, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting CCI entity status from URL '%s': %s", urlRef.String(), err)
	}
	return &entityStatus, nil
}

func (cciClient *CciClient) cciPerformPostPut(httpMethod string, urlRef *url.URL, params url.Values, payload interface{}, additionalHeader map[string]string) (*http.Response, error) {
	// Marshal payload if we have one
	body := new(bytes.Buffer)
	if payload != nil {
		marshaledJson, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling CCI JSON data for %s request %s", httpMethod, err)
		}
		body = bytes.NewBuffer(marshaledJson)
	}

	req := cciClient.newCciRequest(params, httpMethod, urlRef, body, additionalHeader)
	resp, err := cciClient.VCDClient.Client.Http.Do(req)
	if err != nil {
		return nil, err
	}

	// resp is ignored below because it is the same the one above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &ccitypes.CciApiError{})
	if err != nil {
		return nil, fmt.Errorf("error in HTTP %s CCI request: %s", httpMethod, err)
	}
	return resp, nil
}

func (cciClient *CciClient) newCciRequest(params url.Values, method string, reqUrl *url.URL, body io.Reader, additionalHeader map[string]string) *http.Request {
	client := cciClient.VCDClient.Client

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
		// Add the Accept header if apiVersion is specified

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
	setVcloudClientRequestId(client.RequestIdFunc, req) // TODO - can we send a traceable request ID somewhere so that responses contain it?

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
