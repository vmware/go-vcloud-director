/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package govcd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type ServiceAccount struct {
	ServiceAccount *types.ServiceAccount
	authParams     *types.ServiceAccountAuthParams
	org            *Org
}

// GetServiceAccountById gets a Service Account by its ID
func (org *Org) GetServiceAccountById(serviceAccountId string) (*ServiceAccount, error) {
	client := org.client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, serviceAccountId)
	if err != nil {
		return nil, err
	}

	newServiceAccount := &ServiceAccount{
		ServiceAccount: &types.ServiceAccount{},
		org:            org,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, newServiceAccount.ServiceAccount, nil)
	if err != nil {
		return nil, err
	}

	return newServiceAccount, nil
}

// GetAllServiceAccounts gets all service accounts with the specified query parameters
func (org *Org) GetAllServiceAccounts(queryParams url.Values) ([]*ServiceAccount, error) {
	client := org.client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	tenantContext, err := org.getTenantContext()
	if err != nil {
		return nil, err
	}

	// VCD has a pageSize limit on this specific endpoint
	queryParams.Add("pageSize", "32")
	typeResponses := []*types.ServiceAccount{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParams, &typeResponses, getTenantContextHeader(tenantContext))

	if err != nil {
		return nil, fmt.Errorf("failed to get service accounts: %s", err)
	}

	results := make([]*ServiceAccount, len(typeResponses))
	for sliceIndex := range typeResponses {
		results[sliceIndex] = &ServiceAccount{
			ServiceAccount: typeResponses[sliceIndex],
			org:            org,
		}
	}

	return results, nil
}

// GetServiceAccountByName gets a service account by its name
func (org *Org) GetServiceAccountByName(name string) (*ServiceAccount, error) {
	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("name==%s", name))

	serviceAccounts, err := org.GetAllServiceAccounts(queryParams)
	if err != nil {
		return nil, fmt.Errorf("error getting service account by name: %s", err)
	}

	serviceAccount, err := oneOrError("name", name, serviceAccounts)
	if err != nil {
		return nil, err
	}

	return serviceAccount, nil
}

// CreateServiceAccount creates a Service Account and sets it in `Created` status
func (vcdClient *VCDClient) CreateServiceAccount(orgName, name, scope, softwareId, softwareVersion, clientUri string) (*ServiceAccount, error) {
	saParams := &types.ApiTokenParams{
		ClientName:      name,
		Scope:           scope,
		SoftwareID:      softwareId,
		SoftwareVersion: softwareVersion,
		ClientURI:       clientUri,
	}

	newSaParams, err := vcdClient.RegisterToken(orgName, saParams)
	if err != nil {
		return nil, fmt.Errorf("failed to register Service account: %s", err)
	}

	org, err := vcdClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Org by name: %s", err)
	}

	serviceAccountID := "urn:vcloud:serviceAccount:" + newSaParams.ClientID
	serviceAccount, err := org.GetServiceAccountById(serviceAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Service account by ID: %s", err)
	}

	return serviceAccount, nil
}

// Update updates the modifiable fields of a Service Account
func (sa *ServiceAccount) Update(saConfig *types.ServiceAccount) (*ServiceAccount, error) {
	client := sa.org.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	saConfig.ID = sa.ServiceAccount.ID
	saConfig.Name = sa.ServiceAccount.Name
	saConfig.Status = sa.ServiceAccount.Status
	urlRef, err := client.OpenApiBuildEndpoint(endpoint, saConfig.ID)
	if err != nil {
		return nil, err
	}

	returnServiceAccount := &ServiceAccount{
		ServiceAccount: &types.ServiceAccount{},
		org:            sa.org,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, saConfig, returnServiceAccount.ServiceAccount, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating Service Account: %s", err)
	}

	return returnServiceAccount, nil
}

// Authorize authorizes a service account and returns a DeviceID and UserCode which will be used while granting
// the request, and sets the Service Account in `Requested` status
func (sa *ServiceAccount) Authorize() error {
	client := sa.org.client

	uuid := extractUuid(sa.ServiceAccount.ID)
	data := map[string]string{
		"client_id": uuid,
	}

	userDef := "tenant/" + sa.org.orgName()
	if strings.EqualFold(sa.org.orgName(), "system") {
		userDef = "provider"
	}

	endpoint := fmt.Sprintf("%s://%s/oauth/%s/device_authorization", client.VCDHREF.Scheme, client.VCDHREF.Host, userDef)
	urlRef, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return fmt.Errorf("error getting request url from %s: %s", urlRef.String(), err)
	}

	// Not an OpenAPI endpoint so hardcoding the Service Account minimal version
	err = client.OpenApiPostUrlEncoded("37.0", urlRef, nil, data, &sa.authParams, nil)
	if err != nil {
		return fmt.Errorf("error authorizing service account: %s", err)
	}

	return nil
}

// Grant Grants access to the Service Account and sets it in `Granted` status
func (sa *ServiceAccount) Grant() error {
	if sa.authParams == nil {
		return fmt.Errorf("error: userCode is unset, service account needs to be authorized")
	}

	client := sa.org.client
	// This is the only place where this field is used, so a local struct is created
	type serviceAccountGrant struct {
		UserCode string `json:"userCode"`
	}

	userCode := &serviceAccountGrant{
		UserCode: sa.authParams.UserCode,
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccountGrant
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return fmt.Errorf("error granting service account: %s", err)
	}

	tenantContext, err := sa.org.getTenantContext()
	if err != nil {
		return fmt.Errorf("error granting service account: %s", err)
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, userCode, nil, getTenantContextHeader(tenantContext))
	if err != nil {
		return fmt.Errorf("error granting service account: %s", err)
	}

	return nil
}

// GetInitialApiToken gets the initial API token for the Service Account and sets it in `Active` status
func (sa *ServiceAccount) GetInitialApiToken() (*types.ApiTokenRefresh, error) {
	if sa.authParams == nil {
		return nil, fmt.Errorf("error: service account must be authorized and granted")
	}
	client := sa.org.client
	uuid := extractUuid(sa.ServiceAccount.ID)
	data := map[string]string{
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		"client_id":   uuid,
		"device_code": sa.authParams.DeviceCode,
	}
	token, err := client.getAccessToken(sa.ServiceAccount.Org.Name, "CreateServiceAccount", data)
	if err != nil {
		return nil, fmt.Errorf("error getting initial api token: %s", err)
	}
	return token, nil
}

// Refresh updates the Service Account object
func (sa *ServiceAccount) Refresh() error {
	if sa.ServiceAccount == nil || sa.org.client == nil || sa.ServiceAccount.ID == "" {
		return fmt.Errorf("cannot refresh Service Account without ID")
	}

	updatedServiceAccount, err := sa.org.GetServiceAccountById(sa.ServiceAccount.ID)
	if err != nil {
		return err
	}
	sa.ServiceAccount = updatedServiceAccount.ServiceAccount

	return nil
}

// Revoke revokes the service account and its' API token and puts it back in 'Created' stage
func (sa *ServiceAccount) Revoke() error {
	client := sa.org.client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, sa.ServiceAccount.ID, "/revoke")
	if err != nil {
		return err
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a Service Account
func (sa *ServiceAccount) Delete() error {
	client := sa.org.client

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, sa.ServiceAccount.ID)
	if err != nil {
		return err
	}

	err = client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return err
	}

	return nil
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

	err = SaveServiceAccountToFile(apiTokenFile, vcdClient.Client.UserAgent, apiToken)
	if err != nil {
		return fmt.Errorf("failed to save service account token to %s: %s", apiTokenFile, err)
	}
	return nil
}

// SaveServiceAccountToFile saves the API token of the Service Account to a file
func SaveServiceAccountToFile(filename, useragent string, saToken *types.ApiTokenRefresh) error {
	return saveTokenToFile(filename, "Service Account", useragent, saToken)
}
