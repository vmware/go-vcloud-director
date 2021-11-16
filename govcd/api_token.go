/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// SetApiToken behaves similarly to SetToken, with the difference that it will
// return full information about the bearer token, so that the caller can make decisions about token expiration
func (vcdCli *VCDClient) SetApiToken(org, apiToken string) (*types.ApiTokenRefresh, error) {
	usableApiToken, err := vcdCli.GetBearerTokenFromApiToken(org, apiToken)
	if err != nil {
		return nil, err
	}
	err = vcdCli.SetToken(org, BearerTokenHeader, usableApiToken.AccessToken)
	if err != nil {
		return nil, err
	}
	return usableApiToken, nil
}

// GetBearerTokenFromApiToken receives an API token and retrieves a bearer token
// using the refresh token operation.
func (vcdCli *VCDClient) GetBearerTokenFromApiToken(org, token string) (*types.ApiTokenRefresh, error) {
	if vcdCli.Client.APIVCDMaxVersionIs("< 36.1") {
		version, err := vcdCli.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for API token is 10.3.1 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for API token is 36.1 - Version detected: %s", vcdCli.Client.APIVersion)
	}
	var userDef string
	urlStr := strings.Replace(vcdCli.Client.VCDHREF.String(), "/api", "", 1)
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

	options := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": token,
	}
	req := vcdCli.Client.NewRequest(options, http.MethodPost, *reqHref, nil)
	req.Header.Add("Accept", "application/*;version=36.1")

	resp, err := vcdCli.Client.Http.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte
	var tokenDef types.ApiTokenRefresh
	if resp.Body != nil {
		body, err = ioutil.ReadAll(resp.Body)
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
