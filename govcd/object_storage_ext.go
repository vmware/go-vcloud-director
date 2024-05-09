package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// ObjectStorageApiBuildEndpoint helps to construct ObjectStorageApI endpoint by using already configured VCD HREF while requiring only
// the last bit for s3 API.
// Sample URL construct: https://s3.HOST//api/v1/s3
func (client *Client) S3ApiBuildEndpoint(endpoint ...string) (*url.URL, error) {
	endpointString := client.VCDHREF.Scheme + "://s3." + client.VCDHREF.Host + "/api/v1/s3"
	if endpoint != nil {
		endpointString = endpointString + "/" + strings.Join(endpoint, "/")
	}

	urlRef, err := url.ParseRequestURI(endpointString)
	if err != nil {
		return nil, fmt.Errorf("error formatting S3API endpoint: %s", err)
	}
	return urlRef, nil
}

// newS3ApiRequest is a low level function used in upstream S3API functions which handles logging and
// authentication for each API request
func (client *Client) newS3ApiRequest(apiVersion string, params url.Values, method string, reqUrl *url.URL, body io.Reader, additionalHeader map[string]string) *http.Request {
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
			util.Logger.Printf("[DEBUG - newS3ApiRequest] error reading body: %s", err)
		}
		body = bytes.NewReader(readBody)
	}

	req, err := http.NewRequest(method, reqUrlCopy.String(), body)
	if err != nil {
		util.Logger.Printf("[DEBUG - newS3ApiRequest] error getting new request: %s", err)
	}

	if client.VCDAuthHeader != "" && client.VCDToken != "" {
		// Add the authorization header
		req.Header.Add(client.VCDAuthHeader, client.VCDToken)
		// The deprecated authorization token is 32 characters long
		// The bearer token is 612 characters long
		if len(client.VCDToken) > 32 {
			req.Header.Add("Authorization", "Bearer "+client.VCDToken)
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
		req.Header.Add(k, v)
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

func (client *Client) S3ApiGetBuckets(region string, additionalHeader map[string]string) (string, error) {
	urlRef, _ := client.S3ApiBuildEndpoint()

	values, _ := url.ParseQuery("offset=0&limit=1000")

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodGet, urlRef, nil, nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiCreateBucket] error creating: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[DEBUG - S3ApiGetBuckets] Error reading HTTP response body: %s", err)
			return "", err
		}
		return string(body), nil
	} else {
		return "", ErrorEntityNotFound
	}
}

func (client *Client) S3ApiGetBucket(name, region string, additionalHeader map[string]string) (string, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	req := client.newS3ApiRequest(client.APIVersion, nil, http.MethodGet, urlRef, nil, nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiCreateBucket] error creating: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[DEBUG - S3ApiGetBuckets] Error reading HTTP response body: %s", err)
			return "", err
		}
		return string(body), nil
	} else {
		return "", ErrorEntityNotFound
	}
}

func (client *Client) S3ApiCreateBucket(name, region string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	body := fmt.Sprintf(`{"name":"%s", "locationConstraint":"%s"}`, name, region)

	req := client.newS3ApiRequest(client.APIVersion, nil, http.MethodPut, urlRef, bytes.NewBuffer([]byte(body)), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiCreateBucket] error creating: %s", err)
		return nil, err
	}

	return resp, nil
}

func (client *Client) S3ApiDeleteBucket(name, region string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	body := `{
		"quiet": true,
		"removeAll": true,
		"deleteVersion": true,
		"tryAsync": true
	}`

	values, err := url.ParseQuery("delete")
	if err != nil {
		util.Logger.Printf("[DEBUG - ParseQuery] error ParseQuery bucket: %s", err)
		return nil, err
	}

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodDelete, urlRef, bytes.NewBuffer([]byte(body)), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiDeleteBucket] error deleting bucket: %s", err)
		return nil, err
	}

	return resp, nil
}

func (client *Client) S3ApiEditBucketTags(name, region string, tags map[string]string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	values, _ := url.ParseQuery("tagging")

	tagSet := make(map[string][]interface{})
	for k, v := range tags {
		tag := make(map[string]interface{})
		tag["key"] = k
		tag["value"] = v
		tagSet["tags"] = append(tagSet["tags"], tag)
	}

	body := make(map[string][]interface{})
	body["tagSets"] = append(body["tagSet"], tagSet)

	data, err := json.Marshal(body)
	if err != nil {
		util.Logger.Printf("Error marshaling json %s", err)
	}

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPut, urlRef, bytes.NewBuffer(data), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketTags] error editing bucket tags: %s", err)
		return nil, err
	}

	return resp, nil
}
