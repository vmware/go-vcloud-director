/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
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

// SetServiceAccountApiToken reads the current Service Account API token,
// sets the client's bearer token and fetches a new API token for next
// authentication request using SetApiToken and overwrites the old file.
func (vcdClient *VCDClient) SetServiceAccountApiToken(org, apiTokenFile string) error {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return fmt.Errorf("minimum version for Service Account authentication is 10.4 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return fmt.Errorf("minimum API version for Service Account authentication is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	saApiToken := &types.ApiTokenRefresh{}
	// Read file contents and unmarshal them to saApiToken
	err := readFileAndUnmarshalJSON(apiTokenFile, saApiToken)
	if err != nil {
		return err
	}

	// Get bearer token and update the refresh token for the next authentication request
	saApiToken, err = vcdClient.SetApiToken(org, saApiToken.RefreshToken)
	if err != nil {
		return err
	}

	// leave only the refresh token to not leave any sensitive information
	saApiToken = &types.ApiTokenRefresh{
		RefreshToken: saApiToken.RefreshToken,
		TokenType:    "Service Account",
		UpdatedBy:    vcdClient.Client.UserAgent,
		UpdatedOn:    time.Now().Format(time.RFC3339),
	}
	err = marshalJSONAndWriteToFile(apiTokenFile, saApiToken, 0600)
	if err != nil {
		return err
	}

	return nil
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
	var userDef string
	newUrl := new(url.URL)
	newUrl.Scheme = vcdClient.Client.VCDHREF.Scheme
	newUrl.Host = vcdClient.Client.VCDHREF.Host
	urlStr := newUrl.String()
	if strings.EqualFold(org, "system") {
		userDef = "provider"
	} else {
		userDef = fmt.Sprintf("tenant/%s", org)
	}
	reqUrl := fmt.Sprintf("%s/oauth/%s/token", urlStr, userDef)
	reqHref, err := url.ParseRequestURI(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting request URL from %s : %s", reqUrl, err)
	}

	data := bytes.NewBufferString(fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", token))
	req := vcdClient.Client.NewRequest(nil, http.MethodPost, *reqHref, data)
	req.Header.Add("Accept", "application/*;version=36.1")

	resp, err := vcdClient.Client.Http.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte
	var tokenDef types.ApiTokenRefresh
	if resp.Body != nil {
		body, err = io.ReadAll(resp.Body)
	}

	// The default response data to show in the logs is a string of asterisks
	responseData := "[" + strings.Repeat("*", 10) + "]"
	// If users request to see sensitive data, we pass the unchanged response body
	if util.LogPasswords {
		responseData = string(body)
	}
	util.ProcessResponseOutput("GetBearerTokenFromApiToken", resp, responseData)
	if len(body) == 0 {
		return nil, fmt.Errorf("refresh token was empty: %s", resp.Status)
	}
	if err != nil {
		return nil, fmt.Errorf("error extracting refresh token: %s", err)
	}

	err = json.Unmarshal(body, &tokenDef)
	if err != nil {
		return nil, fmt.Errorf("error decoding token text: %s", err)
	}
	if tokenDef.AccessToken == "" {
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
	return &tokenDef, nil
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
