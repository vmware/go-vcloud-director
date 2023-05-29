/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// CreateApiToken is used for creating API tokens and works in two steps:
// 1. Register the token through the `register` endpoint
// 2. Fetch it using `token` endpoint
func (vcdClient *VCDClient) CreateApiToken(org, tokenName string) (*types.ApiTokenParams, error) {
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for API token is 10.3.1 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for API token is 36.1 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	apiTokenParams := &types.ApiTokenParams{
		ClientName: tokenName,
	}

	return vcdClient.registerToken(org, "36.1", apiTokenParams)
}

func (vcdClient *VCDClient) GetTokenById(tokenId string) (*types.Token, error) {
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for API token is 10.3.1 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for API token is 36.1 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	tokenUrn, err := BuildUrnWithUuid("urn:vcloud:token:", tokenId)
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTokens
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, tokenUrn)
	if err != nil {
		return nil, err
	}

	token := &types.Token{}
	err = vcdClient.Client.OpenApiGetItem(apiVersion, urlRef, nil, token, nil)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (vcdClient *VCDClient) DeleteToken(tokenId string) error {
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return fmt.Errorf("minimum version for API token is 10.3.1 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return fmt.Errorf("minimum API version for API token is 36.1 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	tokenUrn, err := BuildUrnWithUuid("urn:vcloud:token:", tokenId)
	if err != nil {
		return err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointTokens
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, tokenUrn)
	if err != nil {
		return err
	}

	err = vcdClient.Client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetInitialApiToken gets the newly created API token, usable only once per token.
func (vcdClient *VCDClient) GetInitialApiToken(org, apiToken, clientId string) (*types.ApiTokenRefresh, error) {
	data := bytes.NewBufferString(
		fmt.Sprintf("grant_type=%s&client_id=%s&assertion=%s",
			"urn:ietf:params:oauth:grant-type:jwt-bearer",
			clientId,
			vcdClient.Client.VCDToken,
		))

	token, err := vcdClient.getToken(org, "36.1", "CreateApiToken", data)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %s", err)
	}

	return token, nil
}

// SetApiToken behaves similarly to SetToken, with the difference that it will
// return full information about the bearer token, so that the caller can make decisions about token expiration
func (vcdClient *VCDClient) SetApiToken(org, apiToken string) (*types.ApiTokenRefresh, error) {
	tokenRefresh, err := vcdClient.GetBearerTokenFromApiToken(org, apiToken)
	if err != nil {
		return nil, err
	}
	err = vcdClient.SetToken(org, BearerTokenHeader, tokenRefresh.AccessToken)
	if err != nil {
		return nil, err
	}
	return tokenRefresh, nil
}

// GetBearerTokenFromApiToken uses an API token to retrieve a bearer token
// using the refresh token operation.
func (vcdClient *VCDClient) GetBearerTokenFromApiToken(org, token string) (*types.ApiTokenRefresh, error) {
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for API token is 10.3.1 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for API token is 36.1 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	data := bytes.NewBufferString(fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", token))
	tokenDef, err := vcdClient.getToken(org, "36.1", "GetBearerTokenFromApiToken", data)
	if err != nil {
		return nil, fmt.Errorf("error getting bearer token: %s", err)
	}

	return tokenDef, nil
}

func (vcdClient *VCDClient) registerToken(org, apiVersion string, token *types.ApiTokenParams) (*types.ApiTokenParams, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	resp, err := vcdClient.doTokenRequest(org, "register", apiVersion, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	newToken := &types.ApiTokenParams{}
	err = json.Unmarshal(body, newToken)
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func (vcdClient *VCDClient) getToken(org, apiVersion string, funcName string, data *bytes.Buffer) (*types.ApiTokenRefresh, error) {
	resp, err := vcdClient.doTokenRequest(org, "token", apiVersion, "application/x-www-form-urlencoded", data)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseData := "[" + strings.Repeat("*", 10) + "]"
	// If users request to see sensitive data, we pass the unchanged response body
	if util.LogPasswords {
		responseData = string(body)
	}
	util.ProcessResponseOutput(funcName, resp, responseData)

	newToken := &types.ApiTokenRefresh{}
	err = json.Unmarshal(body, newToken)
	if err != nil {
		return nil, err
	}

	if newToken.AccessToken == "" {
		// If the access token is empty, the body should contain a composite error message.
		// Attempting to decode it and return as much information as possible
		var errorBody map[string]string
		err2 := json.Unmarshal(body, &errorBody)
		if err2 == nil {
			errorMessage := ""
			for k, v := range errorBody {
				if v == "null" || v == "" {
					continue
				}
				errorMessage += fmt.Sprintf("%s: %s -  ", k, v)
			}
			return nil, fmt.Errorf("%s: %s", errorMessage, resp.Status)
		}

		// If decoding the error fails, we return the raw body (possibly an unencoded internal server error)
		return nil, fmt.Errorf("access token retrieved from API token was empty - %s %s", resp.Status, string(body))
	}

	return newToken, nil
}

func (vcdClient *VCDClient) doTokenRequest(org, endpoint, apiVersion, contentType string, data io.Reader) (*http.Response, error) {

	newUrl := url.URL{
		Scheme: vcdClient.Client.VCDHREF.Scheme,
		Host:   vcdClient.Client.VCDHREF.Host,
	}

	userDef := "tenant/" + org
	if strings.EqualFold(org, "system") {
		userDef = "provider"
	}

	reqHref, err := url.ParseRequestURI(fmt.Sprintf("%s/oauth/%s/%s", newUrl.String(), userDef, endpoint))
	if err != nil {
		return nil, fmt.Errorf("error getting request URL from %s : %s", reqHref.String(), err)
	}
	req := vcdClient.Client.NewRequest(nil, http.MethodPost, *reqHref, data)
	req.Header.Add("Accept", fmt.Sprintf("application/*;version=%s", apiVersion))
	req.Header.Add("Content-Type", contentType)

	return vcdClient.Client.Http.Do(req)
}

// SetApiTokenFile reads the API token file, sets the client's bearer
// token and fetches a new API token for next authentication request
// using SetApiToken
func (vcdClient *VCDClient) SetApiTokenFromFile(org, apiTokenFile string) (*types.ApiTokenRefresh, error) {
	apiToken, err := getApiTokenFromFile(org, apiTokenFile)
	if err != nil {
		return nil, err
	}

	return vcdClient.SetApiToken(org, apiToken.RefreshToken)
}

func getApiTokenFromFile(org, apiTokenFile string) (*types.ApiTokenRefresh, error) {
	apiToken := &types.ApiTokenRefresh{}
	// Read file contents and unmarshal them to apiToken
	err := readFileAndUnmarshalJSON(apiTokenFile, apiToken)
	if err != nil {
		return nil, err
	}

	return apiToken, nil
}

//func saveApiTokenToFile(filename, userAgent string, apiToken *types.ApiTokenRefresh) error {
//	return saveTokenToFile(filename, "API Token", userAgent, apiToken)
//}

func saveTokenToFile(filename, tokenType, userAgent string, token *types.ApiTokenRefresh) error {
	token = &types.ApiTokenRefresh{
		RefreshToken: token.RefreshToken,
		TokenType:    tokenType,
		UpdatedBy:    userAgent,
		UpdatedOn:    time.Now().Format(time.RFC3339),
	}
	err := marshalJSONAndWriteToFile(filename, token, 0600)
	if err != nil {
		return err
	}
	return nil
}

// readFileAndUnmarshalJSON reads a file and unmarshals it to the given variable
func readFileAndUnmarshalJSON(filename string, object any) error {
	data, err := os.ReadFile(path.Clean(filename))
	if err != nil {
		return fmt.Errorf("failed to read from file: %s", err)
	}

	err = json.Unmarshal(data, object)
	if err != nil {
		return fmt.Errorf("failed to unmarshal file contents to the object: %s", err)
	}

	return nil
}

// marshalJSONAndWriteToFile marshalls the given object into JSON and writes
// to a file with the given permissions in octal format (e.g 0600)
func marshalJSONAndWriteToFile(filename string, object any, permissions int) error {
	data, err := json.MarshalIndent(object, " ", " ")
	if err != nil {
		return fmt.Errorf("error marshalling object to JSON: %s", err)
	}

	err = os.WriteFile(filename, data, fs.FileMode(permissions))
	if err != nil {
		return fmt.Errorf("error writing to the file: %s", err)
	}

	return nil
}
