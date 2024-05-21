package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Bucket represents an object storage bucket.
type Bucket struct {
	Name      string      `json:"name"`
	Tenant    string      `json:"tenant"`
	S3Href    string      `json:"s3Href"`
	S3AltHref string      `json:"s3AltHref"`
	Owner     BucketOwner `json:"owner"`
}

// BucketOwner represents the owner of a bucket.
type BucketOwner struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// S3Cors represents the configuration for Cross-Origin Resource Sharing (CORS) rules for an S3 bucket.
type S3Cors struct {
	CorsRules []S3CorsRule `json:"corsRules"`
}

// S3CorsRule represents a Cross-Origin Resource Sharing (CORS) rule for an S3 bucket.
type S3CorsRule struct {
	AllowedMethods []string `json:"allowedMethods"`
	MaxAgeSeconds  int32    `json:"maxAgeSeconds"`
	ExposeHeaders  []string `json:"exposeHeaders"`
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedHeaders []string `json:"allowedHeaders"`
}

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

// S3ApiGetBuckets retrieves a list of buckets from the S3 API in the specified region.
// It takes the region as a parameter and an optional additionalHeader map for additional headers.
// It returns a string containing the response body and an error if any.
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

// S3ApiGetBucket retrieves the contents of an S3 bucket with the specified name and region.
// It sends an HTTP GET request to the S3 API endpoint and returns the response body as a string.
// If the request is successful (HTTP status code 200), the function returns the response body.
// If the request fails or the bucket is not found, the function returns an error.
//
// Parameters:
// - name: The name of the S3 bucket.
// - region: The region where the S3 bucket is located.
// - additionalHeader: Additional headers to include in the request.
//
// Returns:
// - string: The response body as a string if the request is successful.
// - error: An error if the request fails or the bucket is not found.
//
// Example usage:
//
//	body, err := client.S3ApiGetBucket("my-bucket", "us-west-2", nil)
//	if err != nil {
//	  log.Printf("Error retrieving S3 bucket: %s", err)
//	} else {
//	  log.Printf("S3 bucket contents: %s", body)
//	}
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

// S3ApiCreateBucket creates a new S3 bucket with the specified name and region.
// It returns the HTTP response and an error, if any.
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

// S3ApiDeleteBucket deletes an S3 bucket with the specified name and region.
// It also accepts an optional additionalHeader parameter to include additional headers in the request.
// The function returns the HTTP response and an error, if any.
func (client *Client) S3ApiDeleteBucket(name, region string, additionalHeader map[string]string) (*http.Response, error) {

	client.S3ApiCleanBucket(name, region, additionalHeader)

	urlRef, _ := client.S3ApiBuildEndpoint(name)

	req := client.newS3ApiRequest(client.APIVersion, nil, http.MethodDelete, urlRef, nil, nil)

	resp, err := client.Http.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		util.Logger.Printf("[DEBUG - S3ApiDeleteBucket] error deleting bucket: status (%d - %s) %s", resp.StatusCode, resp.Status, err)
		return nil, err
	}

	return resp, nil
}

// S3ApiCleanBucket removes all objects and versions from an S3 bucket.
// It takes the name of the bucket, the region where the bucket is located, and additional headers as input.
// It returns the HTTP response and an error, if any.
// The function sends a POST request to the S3 API endpoint to delete all objects and versions in the bucket.
// The request body includes the following parameters:
// - quiet: Set to true to suppress the response body.
// - removeAll: Set to true to remove all objects and versions.
// - deleteVersion: Set to true to delete all versions of objects.
// - tryAsync: Set to false to perform the operation synchronously.
// If the request is successful and the response status code is less than 400, the function returns the response and nil error.
// Otherwise, it logs an error message and returns nil response and the error.
func (client *Client) S3ApiCleanBucket(name, region string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	body := `{
		"quiet": true,
		"removeAll": true,
		"deleteVersion": true,
		"tryAsync": false
	}`

	values, err := url.ParseQuery("delete")
	if err != nil {
		util.Logger.Printf("[DEBUG - ParseQuery] error ParseQuery bucket: %s", err)
		return nil, err
	}

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPost, urlRef, bytes.NewBuffer([]byte(body)), nil)

	resp, err := client.Http.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		util.Logger.Printf("[DEBUG - S3ApiDeleteBucket] error deleting bucket: status (%d - %s) %s", resp.StatusCode, resp.Status, err)
		return nil, err
	}

	return resp, nil
}

// S3ApiEditBucketTags edits the tags of an S3 bucket.
// It takes the name of the bucket, the region, a map of tags, and additional headers as input parameters.
// The function returns the HTTP response and an error, if any.
// The URL path for this operation is /tagging.
// The tags parameter is a map where the key represents the tag key and the value represents the tag value.
// The additionalHeader parameter is a map of additional headers to be included in the request.
// The function first builds the endpoint URL using the provided bucket name.
// It then constructs the tagSet and body JSON objects based on the tags parameter.
// The function marshals the body JSON object into a byte array.
// Finally, it sends a PUT request to the S3 API endpoint with the marshaled JSON data and returns the response.
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

// S3ApiEditBucketAcls edits the access control list (ACL) of an S3 bucket.
// It takes the name and region of the bucket, a list of ACLs, and additional headers as input.
// The function returns the HTTP response and an error, if any.
// The ACLs parameter is a list of maps, where each map represents an ACL.
// Each ACL map should contain the following keys:
//   - "user": The user type. Possible values are "TENANT", "AUTHENTICATED", "PUBLIC", and "SYSTEM-LOGGER".
//   - "permission": The permission level for the user. Possible values are "READ", "WRITE", "READ_ACP", "WRITE_ACP", and "FULL_CONTROL".
//
// The function updates the ACLs of the bucket and returns the updated HTTP response.
// If there is an error during the process, the function returns the error.
func (client *Client) S3ApiEditBucketAcls(name, region string, acls []map[string]interface{}, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	bucketJson, _ := client.S3ApiGetBucket(name, region, nil)

	var bucket *Bucket

	if err := json.Unmarshal([]byte(bucketJson), &bucket); err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketAcl] error unmarshal bucket: %s", err)
	}

	values, _ := url.ParseQuery("acl")
	var grants []map[string]interface{}
	for _, acl := range acls {
		grantee := make(map[string]interface{})
		grant := make(map[string]interface{})
		switch acl["user"] {
		case "TENANT":
			grantee["id"] = bucket.Tenant + "|"
		case "AUTHENTICATED":
			grantee["uri"] = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
		case "PUBLIC":
			grantee["uri"] = "http://acs.amazonaws.com/groups/global/AllUsers"
		case "SYSTEM-LOGGER":
			grantee["uri"] = "http://acs.amazonaws.com/groups/s3/LogDelivery"
		}
		grant["grantee"] = grantee
		grant["permission"] = acl["permission"]

		grants = append(grants, grant)
	}

	grants = append(grants, map[string]interface{}{"grantee": map[string]interface{}{"id": bucket.Owner.Id}, "permission": "FULL_CONTROL"})

	body := map[string]interface{}{}
	body["owner"] = bucket.Owner
	body["grants"] = grants
	data, err := json.Marshal(body)
	if err != nil {
		util.Logger.Printf("Error marshaling json %s", err)
	}

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPut, urlRef, bytes.NewBuffer(data), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketAcls] error editing bucket acls: %s", err)
		return nil, err
	}

	return resp, nil
}

// S3ApiEditBucketCors edits the CORS (Cross-Origin Resource Sharing) configuration for a bucket in the S3 API.
// It takes the name of the bucket, the region where the bucket is located, the CORS configuration as a JSON string,
// and additional headers as input parameters.
// The function returns the HTTP response and an error, if any.
func (client *Client) S3ApiEditBucketCors(name, region string, cors string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	var s3cors *S3Cors
	if err := json.Unmarshal([]byte(cors), &s3cors); err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketCors] error unmarshal cors: %s", err)
		return nil, err
	}

	values, _ := url.ParseQuery("cors")

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPut, urlRef, bytes.NewBuffer([]byte(cors)), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketCors] error editing bucket cors: %s", err)
		return nil, err
	}

	return resp, nil
}

// S3ApiEditBucketVersioning edits the versioning status of an S3 bucket.
// It takes the bucket name, region, versioning status, and additional headers as input parameters.
// The function returns the HTTP response and an error, if any.
// The versioning status can be either true (enabled) or false (suspended).
// The function constructs the request URL, sets the versioning status in the request body,
// and sends a PUT request to the S3 API endpoint to update the bucket versioning status.
// If the request is successful, the function returns the HTTP response.
// If there is an error, the function logs the error and returns nil and the error.
func (client *Client) S3ApiEditBucketVersioning(name, region string, versioning bool, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	values, _ := url.ParseQuery("versioning")

	versioningStatus := "Enabled"
	if !versioning {
		versioningStatus = "Suspended"
	}

	body := `{"status": "` + versioningStatus + `"}`

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPut, urlRef, bytes.NewBuffer([]byte(body)), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketVersioning] error versioning bucket: %s", err)
		return nil, err
	}

	return resp, nil
}

// S3ApiEditBucketEncryption edits the encryption settings for a bucket in the S3 API.
// It takes the bucket name, region, encryption key, and additional headers as parameters.
// The function returns the HTTP response and an error, if any.
// The encryption settings are updated by sending a PUT request to the S3 API endpoint.
// The encryption algorithm used is AES256, and the encryption key is optional.
// If a key is provided, it is used to encrypt the bucket's contents.
// If no key is provided, the bucket's contents are encrypted using a randomly generated key.
// The function returns the HTTP response and an error, if any.
func (client *Client) S3ApiEditBucketEncryption(name, region string, key string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	values, _ := url.ParseQuery("encryption")

	sseRules := make(map[string][]interface{})
	rule := make(map[string]interface{})
	sseByDefault := make(map[string]interface{})
	sseByDefault["sseAlgorithm"] = "AES256"

	if key != "" && len(key) > 0 {
		// key := make([]byte, 32)
		// _, err := rand.Read(key)
		// if err != nil {
		// 	return nil, fmt.Errorf("S3ApiEditBucketEncryption - Error generating AES key : %s", encryption)
		// }

		// sseByDefault["sseCKey"] = base64.StdEncoding.EncodeToString(key)
		sseByDefault["sseCKey"] = key
	}

	rule["sseByDefault"] = sseByDefault
	sseRules["sseRules"] = append(sseRules["sseRules"], rule)

	data, err := json.Marshal(sseRules)
	if err != nil {
		util.Logger.Printf("Error marshaling json %s", err)
	}

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPut, urlRef, bytes.NewBuffer(data), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketVersioning] error versioning bucket: %s", err)
		return nil, err
	}

	return resp, nil
}

func (client *Client) S3ApiEditBucketReplication(name, region, id, targetBucket, prefix string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name)

	values, _ := url.ParseQuery("replication")

	body := `{
		"rule": [
			{
				"id": "` + id + `",
				"status": "Enabled",
				"destination": {
					"bucket": "arn:aws:s3:::` + targetBucket + `"
				},
				"prefix": "` + prefix + `"
			}
		]
	}`

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPut, urlRef, bytes.NewBuffer([]byte(body)), nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditBucketReplication] error replication bucket: %s", err)
		return nil, err
	}

	return resp, nil
}

// S3ApiUploadObject uploads an object to an S3-compatible storage service.
// It takes the name of the storage service, the region, the object key, the source file path,
// and additional headers as input parameters.
// It returns the HTTP response and an error, if any.
func (client *Client) S3ApiUploadObject(name, region, objectKey, source string, additionalHeader map[string]string) (*http.Response, error) {
	urlRef, _ := client.S3ApiBuildEndpoint(name, objectKey)

	_, err := os.Stat(source)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditObjectUpload] error getting object stats: %s", err)
		return nil, err
	}

	file, err := os.ReadFile(source)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditObjectUpload] error reading object: %s", err)
		return nil, err
	}

	contentType := http.DetectContentType(file)

	values, _ := url.ParseQuery("cors")
	headers := make(map[string]string)
	headers["Content-Type"] = contentType

	req := client.newS3ApiRequest(client.APIVersion, values, http.MethodPut, urlRef, bytes.NewBuffer(file), headers)

	resp, err := client.Http.Do(req)
	if err != nil {
		util.Logger.Printf("[DEBUG - S3ApiEditObjectUpload] error uploading object: %s", err)
		return nil, err
	}

	return resp, nil
}
