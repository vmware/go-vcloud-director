/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

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

	token, err := vcdClient.GetServiceAccountApiToken(org, saParams.ClientID, saAuthParams.DeviceCode)
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
func (vcdClient *VCDClient) GetServiceAccountApiToken(org, clientId, deviceCode string) (*types.ApiTokenRefresh, error) {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for Service Accounts is 10.4.0 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for Service Accounts is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}
	data := bytes.NewBufferString(
		fmt.Sprintf("grant_type=%s&client_id=%s&device_code=%s",
			"urn:ietf:params:oauth:grant-type:device_code",
			clientId,
			deviceCode,
		))

	token, err := vcdClient.getToken(org, "37.0", "CreateServiceAccount", data)
	if err != nil {
		return nil, err
	}
	return token, nil
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

	err = vcdClient.SaveServiceAccountToFile(apiTokenFile, "Service Account", apiToken)
	if err != nil {
		return fmt.Errorf("failed to save service account token to %s: %s", apiTokenFile, err)
	}
	return nil
}

func (vcdClient *VCDClient) SaveServiceAccountToFile(filename, tokentype string, saToken *types.ApiTokenRefresh) error {
	return saveTokenToFile(filename, "Service Account", vcdClient.Client.UserAgent, saToken)
}
