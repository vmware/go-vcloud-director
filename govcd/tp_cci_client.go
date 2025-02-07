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
	"github.com/vmware/go-vcloud-director/v3/tptypes"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// constants that are used for tracking entity state after it is being created
const (
	stateWaitTimeout    = 1 * time.Hour
	stateWaitDelay      = 5 * time.Second
	stateWaitMinTimeout = 5 * time.Second
)

type TpClient struct {
	VCDClient *VCDClient
	// Client      Client  // Client for the underlying VCD instance
	// sessionHREF url.URL // HREF for the session API
	// QueryHREF   url.URL // HREF for the query API
}

func (tpClient *TpClient) IsSupported() bool {
	return tpClient.VCDClient.Client.APIVCDMaxVersionIs(">= 40")
}

func (tpClient *TpClient) GetCciUrl(endpoint ...string) (*url.URL, error) {
	path := fmt.Sprintf(tptypes.CciKubernetesSubpath, tpClient.VCDClient.Client.VCDHREF.Scheme, tpClient.VCDClient.Client.VCDHREF.Host)
	path = path + strings.Join(endpoint, "")

	return url.ParseRequestURI(path)
}

func (tpClient *TpClient) PostItem(urlRef *url.URL, responseUrlRef *url.URL, params url.Values, payload, outType interface{}) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Posting %s item to endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRefCopy.String(), reflect.TypeOf(outType))

	if !tpClient.IsSupported() {
		return fmt.Errorf("TP Client is not supported on this version")
	}

	resp, err := tpClient.cciPerformPostPut(http.MethodPost, urlRefCopy, params, payload, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		util.Logger.Printf("[TRACE] Synchronous task expected (HTTP status code 200 or 201). Got %d", resp.StatusCode)
		return fmt.Errorf("POST request expected sync task (HTTP response 200 or 201), got %d", resp.StatusCode)

	}

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding JSON response after POST: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	// WAIT for entity to transition from "CREATING" or "WAITING" to "CREATED"
	_, err = tpClient.waitForState(context.TODO(), "label", responseUrlRef, []string{"CREATING", "WAITING"}, []string{"CREATED"})
	if err != nil {
		return fmt.Errorf("error waiting for CCI entity state: %s", err)
	}

	// Retrieve the final state of original entity
	err = tpClient.GetItem(responseUrlRef, nil, outType, nil)
	if err != nil {
		return fmt.Errorf("error retrieving final entity after creation: %s", err)
	}

	return nil
}

func (tpClient *TpClient) GetItem(urlRef *url.URL, params url.Values, outType interface{}, additionalHeader map[string]string) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Getting item from endpoint %s with expected response of type %s",
		urlRefCopy.String(), reflect.TypeOf(outType))

	if !tpClient.IsSupported() {
		return fmt.Errorf("TP Client is not supported on this version")
	}

	req := tpClient.newCciRequest(params, http.MethodGet, urlRefCopy, nil, additionalHeader)
	resp, err := tpClient.VCDClient.Client.Http.Do(req)
	if err != nil {
		return fmt.Errorf("error performing GET request to %s: %s", urlRefCopy.String(), err)
	}

	// Bypassing the regular path using function checkRespWithErrType and returning parsed error directly
	// HTTP 403: Forbidden - is returned if the user is not authorized or the entity does not exist.
	if resp.StatusCode == http.StatusForbidden {
		err := ParseErr(types.BodyTypeJSON, resp, &tptypes.CciApiError{})
		closeErr := resp.Body.Close()
		return fmt.Errorf("%s: %s [body close error: %s]", ErrorEntityNotFound, err, closeErr)
	}

	// resp is ignored below because it is the same as above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &tptypes.CciApiError{})

	// Any other error occurred
	if err != nil {
		return fmt.Errorf("error in HTTP GET request: %s", err)
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

func (tpClient *TpClient) DeleteItem(urlRef *url.URL, params url.Values, additionalHeader map[string]string) error {
	// copy passed in URL ref so that it is not mutated
	urlRefCopy := copyUrlRef(urlRef)

	util.Logger.Printf("[TRACE] Deleting item at endpoint %s", urlRefCopy.String())

	if !tpClient.IsSupported() {
		return fmt.Errorf("TP Client is not supported on this version")
	}

	// Perform request
	req := tpClient.newCciRequest(params, http.MethodDelete, urlRefCopy, nil, additionalHeader)

	resp, err := tpClient.VCDClient.Client.Http.Do(req)
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
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &tptypes.CciApiError{})
	if err != nil {
		return fmt.Errorf("error in HTTP DELETE request: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	/// TODO Track deletion ?????

	return nil
}

// TODO - can we do this quite cheap without needing to pull in Hashicorp package at SDK level
func (tpClient *TpClient) waitForState(ctx context.Context, entityLabel string, addr *url.URL, pendingStates, targetStates []string) (any, error) {
	stateChangeFunc := retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (any, string, error) {
			cciEntity, err := tpClient.getAnyCciState(addr)
			if err != nil {
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

func (tpClient *TpClient) getAnyCciState(urlRef *url.URL) (*tptypes.CciEntityStatus, error) {
	entityStatus := tptypes.CciEntityStatus{}

	err := tpClient.GetItem(urlRef, nil, &entityStatus, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting entity status from URL '%s': %s", urlRef.String(), err)
	}
	return &entityStatus, nil
}

func (tpClient *TpClient) cciPerformPostPut(httpMethod string, urlRef *url.URL, params url.Values, payload interface{}, additionalHeader map[string]string) (*http.Response, error) {
	// Marshal payload if we have one
	body := new(bytes.Buffer)
	if payload != nil {
		marshaledJson, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON data for %s request %s", httpMethod, err)
		}
		body = bytes.NewBuffer(marshaledJson)
	}

	req := tpClient.newCciRequest(params, httpMethod, urlRef, body, additionalHeader)
	resp, err := tpClient.VCDClient.Client.Http.Do(req)
	if err != nil {
		return nil, err
	}

	// resp is ignored below because it is the same the one above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &tptypes.CciApiError{})
	if err != nil {
		return nil, fmt.Errorf("error in HTTP %s request: %s", httpMethod, err)
	}
	return resp, nil
}

func (tpClient *TpClient) newCciRequest(params url.Values, method string, reqUrl *url.URL, body io.Reader, additionalHeader map[string]string) *http.Request {
	client := tpClient.VCDClient.Client

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
