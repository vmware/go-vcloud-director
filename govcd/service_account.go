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

type serviceAccount struct {
	*types.ServiceAccount
	authParams *types.ServiceAccountAuthParams
	vcdClient  *VCDClient
}

func (vcdClient *VCDClient) GetServiceAccountById(serviceAccountId string) (*serviceAccount, error) {
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return nil, fmt.Errorf("minimum version for API token is 10.3.1 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return nil, fmt.Errorf("minimum API version for API token is 36.1 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	tokenUrn, err := BuildUrnWithUuid("urn:vcloud:serviceAccount:", serviceAccountId)
	if err != nil {
		return nil, err
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, tokenUrn)
	if err != nil {
		return nil, err
	}

	result := &serviceAccount{
		vcdClient: vcdClient,
	}

	err = vcdClient.Client.OpenApiGetItem(apiVersion, urlRef, nil, result, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteServiceAccountByID deletes a Service Account
func (vcdClient *VCDClient) DeleteServiceAccountByID(serviceAccountID string) error {
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		version, err := vcdClient.Client.GetVcdFullVersion()
		if err == nil {
			return fmt.Errorf("minimum version for Service Accounts is 10.4.0 - Version detected: %s", version.Version)
		}
		// If we can't get the VCD version, we return API version info
		return fmt.Errorf("minimum API version for Service Accounts is 37.0 - Version detected: %s", vcdClient.Client.APIVersion)
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, serviceAccountID)
	if err != nil {
		return err
	}

	err = vcdClient.Client.OpenApiDeleteItem("37.0", urlRef, nil, nil)
	if err != nil {
		return err
	}

	return err
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

	saParams := &types.ApiTokenParams{
		ClientName:      name,
		Scope:           scope,
		SoftwareID:      softwareId,
		SoftwareVersion: softwareVersion,
		ClientURI:       clientUri,
	}

	saParams, err := vcdClient.RegisterToken(org, "37.0", saParams)
	if err != nil {
		return nil, err
	}

	return saParams, nil
}

// AuthorizeServiceAccount authorizes a service account and returns a DeviceID and UserCode which will be used while granting
// the request, and sets the Service Account in `Requested` status
func (sa *serviceAccount) Authorize() error {
	client := sa.vcdClient.Client

	uuid := extractUuid(sa.ID)
	data := bytes.NewBufferString(
		fmt.Sprintf("client_id=%s",
			uuid,
		))

	resp, err := client.doTokenRequest(sa.Org.Name, "device_authorization", "37.0", "application/x-www-form-urlencoded", data)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &sa.authParams)
	if err != nil {
		return err
	}

	return nil
}

// GrantServiceAccount Grants access to the Service Account and sets it in `Granted` status
func (sa *serviceAccount) Grant() error {
	client := sa.vcdClient.Client
	// This is the only place where this field is used, so a local struct is created
	type serviceAccountGrant struct {
		UserCode string `json:"userCode"`
	}

	userCode := &serviceAccountGrant{
		UserCode: sa.authParams.UserCode,
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccountGrant
	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return err
	}

	err = client.OpenApiPostItem("37.0", urlRef, nil, userCode, nil, nil)
	if err != nil {
		return err
	}

	return err
}

// Refresh updates the Service Account object
func (sa *serviceAccount) Refresh() error {
	uuid := extractUuid(sa.ID)
	updatedServiceAccount, err := sa.vcdClient.GetServiceAccountById(uuid)
	if err != nil {
		return err
	}
	sa.ServiceAccount = updatedServiceAccount.ServiceAccount

	return nil
}

// Revoke revokes the service account and its' API token and puts it back in 'Created' stage
func (sa *serviceAccount) Revoke() error {
	client := sa.vcdClient.Client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, sa.ID, "/revoke")
	if err != nil {
		return err
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetServiceAccountToken gets the initial API token for the Service Account and sets it in `Active` status
func (sa *serviceAccount) GetApiToken() (*types.ApiTokenRefresh, error) {
	client := sa.vcdClient.Client
	uuid := extractUuid(sa.ID)
	data := bytes.NewBufferString(
		fmt.Sprintf("grant_type=%s&client_id=%s&device_code=%s",
			"urn:ietf:params:oauth:grant-type:device_code",
			uuid,
			sa.authParams.DeviceCode,
		))

	token, err := client.GetApiToken(sa.Org.Name, "37.0", "CreateServiceAccount", data)
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
