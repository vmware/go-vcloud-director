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

// SetApiTokenFile reads the API token file, sets the client's bearer
// token and fetches a new API token for next authentication request
// using SetApiToken
func (vcdClient *VCDClient) SetApiTokenFromFile(org, apiTokenFile string) (*types.ApiTokenRefresh, error) {
	apiToken := &types.ApiTokenRefresh{}
	// Read file contents and unmarshal them to apiToken
	err := readFileAndUnmarshalJSON(apiTokenFile, apiToken)
	if err != nil {
		return nil, err
	}

	return vcdClient.SetApiToken(org, apiToken.RefreshToken)
}

// SetServiceAccountApiToken gets the refresh token from a provided file, fetches
// the bearer token and updates the provided file with the new refresh token for
// next usage as service account API tokens are one-time use
func (vcdClient *VCDClient) SetServiceAccountApiToken(org, apiTokenFile string) error {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return fmt.Errorf("minimum version for Service Account authentication is 10.4 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return fmt.Errorf("minimum API version for Service Account authentication is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	apiToken, err := vcdClient.SetApiTokenFromFile(org, apiTokenFile)
	if err != nil {
		return err
	}

	// leave only the refresh token to not leave any sensitive information
	apiToken = &types.ApiTokenRefresh{
		RefreshToken: apiToken.RefreshToken,
		TokenType:    "Service Account",
		UpdatedBy:    vcdClient.Client.UserAgent,
		UpdatedOn:    time.Now().Format(time.RFC3339),
	}
	err = marshalJSONAndWriteToFile(apiTokenFile, apiToken, 0600)
	if err != nil {
		return err
	}

	return nil
}

// CreateServiceAccount goes through the process of activating a Service Account
// Service account activation takes 4 steps, which are done here:
// Created -> Requested -> Granted -> Active
// 1. Create a service account
// 2. Authorize it
// 3. Grant the created service account access
// 4. Fetch the initial API token as it requires different payload compared to normal API token fetch
//
// Example usage:
// saToken, err := vcdClient.CreateServiceAccount("org1", "account1", "urn:vcloud:vapp:administrator", "123e4567-e89b-12d3-a456-426614174000", "1.0.0", "http://client.example.com")
//
// softwareVersion and clientUri are optional
func (vcdClient *VCDClient) CreateServiceAccount(org, name, scope, softwareId, softwareVersion, clientUri string) (*types.ApiTokenRefresh, error) {
	saParams, err := vcdClient.RegisterServiceAccount(org, name, scope, softwareId, softwareVersion, clientUri)
	if err != nil {
		return nil, fmt.Errorf("failed to register service account: %s", err)
	}

	saAuthParams, err := vcdClient.AuthorizeServiceAccount(org, saParams.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize service account: %s", err)
	}

	err = vcdClient.GrantServiceAccount(org, saAuthParams.UserCode)
	if err != nil {
		return nil, fmt.Errorf("failed to grant service account: %s", err)
	}

	data := bytes.NewBufferString(
		fmt.Sprintf("grant_type=%s&client_id=%s&device_code=%s",
			"urn:ietf:params:oauth:grant-type:device_code",
			saParams.ClientID,
			saAuthParams.DeviceCode,
		))

	token, err := vcdClient.getToken(org, "37.0", "CreateServiceAccount", data)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial service account token: %s", err)
	}

	return token, nil
}

// RegisterServiceAccount registers(creates) a Service Account and sets it in `Created` status
func (vcdClient *VCDClient) RegisterServiceAccount(org, name, scope, softwareId, softwareVersion, clientUri string) (*types.ApiTokenParams, error) {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for Service Accounts is 10.4.0 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for Service Accounts is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	serviceAccountParams := &types.ApiTokenParams{
		ClientName:      name,
		Scope:           scope,
		SoftwareID:      softwareId,
		SoftwareVersion: softwareVersion,
		ClientURI:       clientUri,
	}

	return vcdClient.registerToken(org, "37.0", serviceAccountParams)
}

// AuthorizeServiceAccount authorizes a service account and returns a DeviceID and UserCode which will be used while granting
// the request, and sets the Service Account in `Requested` status
func (vcdClient *VCDClient) AuthorizeServiceAccount(org, clientID string) (*types.ServiceAccountAuthParams, error) {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for Service Accounts is 10.4.0 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for Service Accounts is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	serviceAccountParams := &types.ApiTokenParams{
		ClientID: clientID,
	}

	data := bytes.NewBufferString(
		fmt.Sprintf("client_id=%s",
			serviceAccountParams.ClientID,
		))

	resp, err := vcdClient.doTokenRequest(org, "device_authorization", "37.0", "application/x-www-form-urlencoded", data)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	saAuthParams := &types.ServiceAccountAuthParams{}
	err = json.Unmarshal(body, saAuthParams)
	if err != nil {
		return nil, err
	}

	return saAuthParams, nil
}

// GrantServiceAccount Grants access to the Service Account and sets it in `Granted` status
func (vcdClient *VCDClient) GrantServiceAccount(org, userCode string) error {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return fmt.Errorf("minimum version for Service Accounts is 10.4.0 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return fmt.Errorf("minimum API version for Service Accounts is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	// This is the only place that this field is used, so a local struct is created
	type serviceAccountGrant struct {
		UserCode string `json:"userCode"`
	}

	serviceAccount := &serviceAccountGrant{
		UserCode: userCode,
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccountGrant
	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return err
	}

	err = vcdClient.Client.OpenApiPostItem("37.0", urlRef, nil, serviceAccount, nil, nil)
	if err != nil {
		return err
	}

	return err
}

// GetServiceAccountToken gets the initial API token for the Service Account and sets it in `Active` status
func (vcdClient *VCDClient) GetServiceAccountToken(org string, saAuth *types.ServiceAccountAuthParams) error {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return fmt.Errorf("minimum version for Service Accounts is 10.4.0 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return fmt.Errorf("minimum API version for Service Accounts is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	return nil
}

// CreateApiToken is used for creating API tokens and works in two steps:
// 1. Register the token through the `register` endpoint
// 2. Fetch it using `token` endpoint
func (vcdClient *VCDClient) CreateApiToken(org, tokenName string) (*types.ApiTokenRefresh, error) {
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

	createdApiTokenParams, err := vcdClient.registerToken(org, "36.1", apiTokenParams)
	if err != nil {
		return nil, fmt.Errorf("error registering the token: %s", err)
	}

	data := bytes.NewBufferString(
		fmt.Sprintf("grant_type=%s&client_id=%s&assertion=%s",
			"urn:ietf:params:oauth:grant-type:jwt-bearer",
			createdApiTokenParams.ClientID,
			vcdClient.Client.VCDToken,
		))

	token, err := vcdClient.getToken(org, "36.1", "CreateApiToken", data)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %s", err)
	}

	return token, nil
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
