/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"cmp"
	"encoding/xml"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// GetOpenIdConnectSettings retrieves the current OpenID Connect settings for a given Organization
func (adminOrg *AdminOrg) GetOpenIdConnectSettings() (*types.OrgOAuthSettings, error) {
	return oidcExecuteRequest(adminOrg, http.MethodGet, nil)
}

// SetOpenIdConnectSettings sets the OpenID Connect configuration for a given Organization. If the well-known configuration
// endpoint is provided, the configuration is automatically retrieved from that URL.
// If other fields have been set in the input structure, the corresponding values retrieved from the well-known endpoint are overridden.
// If there are no fields set, the configuration retrieved from the well-known configuration endpoint is applied as-is.
// ClientId and ClientSecret properties are always mandatory, with and without well-known endpoint.
// This method returns an error if the settings can't be saved in VCD for any reason or if the provided settings are wrong.
func (adminOrg *AdminOrg) SetOpenIdConnectSettings(settings types.OrgOAuthSettings) (*types.OrgOAuthSettings, error) {
	if settings.ClientId == "" {
		return nil, fmt.Errorf("the Client ID is mandatory to configure OpenID Connect")
	}
	if settings.ClientSecret == "" {
		return nil, fmt.Errorf("the Client Secret is mandatory to configure OpenID Connect")
	}
	if settings.WellKnownEndpoint != "" {
		err := oidcValidateConnection(adminOrg.client, settings.WellKnownEndpoint)
		if err != nil {
			return nil, err
		}
		wellKnownSettings, err := oidcConfigureWithEndpoint(adminOrg.client, adminOrg.AdminOrg.HREF, settings.WellKnownEndpoint)
		if err != nil {
			return nil, err
		}

		// The following statements allow users to override the well-known automatic configuration values with their own,
		// mimicking what users can do in UI.
		// If an attribute was not set in the input settings, the well-known endpoint value will be chosen.
		settings.AccessTokenEndpoint = cmp.Or(settings.AccessTokenEndpoint, wellKnownSettings.AccessTokenEndpoint)
		settings.IssuerId = cmp.Or(settings.IssuerId, wellKnownSettings.IssuerId)
		settings.JwksUri = cmp.Or(settings.JwksUri, wellKnownSettings.JwksUri)
		settings.UserInfoEndpoint = cmp.Or(settings.UserInfoEndpoint, wellKnownSettings.UserInfoEndpoint)
		settings.UserAuthorizationEndpoint = cmp.Or(settings.UserAuthorizationEndpoint, wellKnownSettings.UserAuthorizationEndpoint)
		settings.ScimEndpoint = cmp.Or(settings.ScimEndpoint, wellKnownSettings.ScimEndpoint)

		if settings.Scope == nil || len(settings.Scope) == 0 {
			settings.Scope = wellKnownSettings.Scope
		}

		if settings.OIDCAttributeMapping == nil {
			// The whole mapping is missing, we take the whole struct from well-known endpoint
			settings.OIDCAttributeMapping = wellKnownSettings.OIDCAttributeMapping
		} else {
			// Some mappings are present, others are missing. We take the missing ones from well-known endpoint
			settings.OIDCAttributeMapping.EmailAttributeName = cmp.Or(settings.OIDCAttributeMapping.EmailAttributeName, wellKnownSettings.OIDCAttributeMapping.EmailAttributeName)
			settings.OIDCAttributeMapping.SubjectAttributeName = cmp.Or(settings.OIDCAttributeMapping.SubjectAttributeName, wellKnownSettings.OIDCAttributeMapping.SubjectAttributeName)
			settings.OIDCAttributeMapping.LastNameAttributeName = cmp.Or(settings.OIDCAttributeMapping.LastNameAttributeName, wellKnownSettings.OIDCAttributeMapping.LastNameAttributeName)
			settings.OIDCAttributeMapping.RolesAttributeName = cmp.Or(settings.OIDCAttributeMapping.RolesAttributeName, wellKnownSettings.OIDCAttributeMapping.RolesAttributeName)
			settings.OIDCAttributeMapping.FullNameAttributeName = cmp.Or(settings.OIDCAttributeMapping.FullNameAttributeName, wellKnownSettings.OIDCAttributeMapping.FullNameAttributeName)
			settings.OIDCAttributeMapping.GroupsAttributeName = cmp.Or(settings.OIDCAttributeMapping.GroupsAttributeName, wellKnownSettings.OIDCAttributeMapping.GroupsAttributeName)
			settings.OIDCAttributeMapping.FirstNameAttributeName = cmp.Or(settings.OIDCAttributeMapping.FirstNameAttributeName, wellKnownSettings.OIDCAttributeMapping.FirstNameAttributeName)
		}

		if settings.OAuthKeyConfigurations == nil {
			settings.OAuthKeyConfigurations = wellKnownSettings.OAuthKeyConfigurations
		}
	}
	// Perform early validations. These are required in UI before sending the payload.
	if settings.UserAuthorizationEndpoint == "" {
		return nil, fmt.Errorf("the User Authorization Endpoint is mandatory to configure OpenID Connect")
	}
	if settings.AccessTokenEndpoint == "" {
		return nil, fmt.Errorf("the Access Token Endpoint is mandatory to configure OpenID Connect")
	}
	if settings.UserInfoEndpoint == "" {
		return nil, fmt.Errorf("the User Info Endpoint is mandatory to configure OpenID Connect")
	}
	if settings.MaxClockSkew < 0 {
		return nil, fmt.Errorf("the Max Clock Skew must be positive to correctly configure OpenID Connect")
	}
	if settings.OIDCAttributeMapping == nil || settings.OIDCAttributeMapping.SubjectAttributeName == "" ||
		settings.OIDCAttributeMapping.EmailAttributeName == "" || settings.OIDCAttributeMapping.FullNameAttributeName == "" ||
		settings.OIDCAttributeMapping.FirstNameAttributeName == "" || settings.OIDCAttributeMapping.LastNameAttributeName == "" {
		return nil, fmt.Errorf("the Subject, Email, Full name, First Name and Last name are mandatory OIDC Attribute (Claims) Mappings, to configure OpenID Connect")
	}
	if settings.OAuthKeyConfigurations == nil || len(settings.OAuthKeyConfigurations.OAuthKeyConfiguration) == 0 {
		return nil, fmt.Errorf("the OIDC Key Configuration is mandatory to configure OpenID Connect")
	}

	// Perform connectivity validations
	err := oidcValidateConnection(adminOrg.client, settings.UserAuthorizationEndpoint)
	if err != nil {
		return nil, err
	}
	err = oidcValidateConnection(adminOrg.client, settings.AccessTokenEndpoint)
	if err != nil {
		return nil, err
	}
	err = oidcValidateConnection(adminOrg.client, settings.UserInfoEndpoint)
	if err != nil {
		return nil, err
	}

	// The namespace must be set for all structures, otherwise the API call fails
	settings.Xmlns = types.XMLNamespaceVCloud
	settings.OAuthKeyConfigurations.Xmlns = types.XMLNamespaceVCloud
	for i := range settings.OAuthKeyConfigurations.OAuthKeyConfiguration {
		settings.OAuthKeyConfigurations.OAuthKeyConfiguration[i].Xmlns = types.XMLNamespaceVCloud
	}
	settings.OIDCAttributeMapping.Xmlns = types.XMLNamespaceVCloud

	result, err := oidcExecuteRequest(adminOrg, http.MethodPut, &settings)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteOpenIdConnectSettings deletes the current OpenID Connect settings from a given Organization
func (adminOrg *AdminOrg) DeleteOpenIdConnectSettings() error {
	_, err := oidcExecuteRequest(adminOrg, http.MethodDelete, nil)
	if err != nil {
		return err
	}
	return nil
}

// oidcExecuteRequest executes a request to the OIDC endpoint with the given payload and HTTP method
func oidcExecuteRequest(adminOrg *AdminOrg, method string, payload *types.OrgOAuthSettings) (*types.OrgOAuthSettings, error) {
	if adminOrg.AdminOrg.HREF == "" {
		return nil, fmt.Errorf("the HREF of the Organization is required to use OpenID Connect")
	}
	endpoint, err := url.Parse(adminOrg.AdminOrg.HREF + "/settings/oauth")
	if err != nil {
		return nil, fmt.Errorf("error parsing Organization '%s' OpenID Connect URL: %s", adminOrg.AdminOrg.Name, err)
	}
	if endpoint == nil {
		return nil, fmt.Errorf("error parsing Organization '%s' OpenID Connect URL: it is nil", adminOrg.AdminOrg.Name)
	}
	if method == http.MethodPut && payload == nil {
		return nil, fmt.Errorf("the OIDC settings cannot be nil when performing a PUT call")
	}

	// Set Organization "tenant context" headers
	headers := make(http.Header)
	headers.Set("Content-Type", types.MimeOAuthSettingsXml)
	for k, v := range getTenantContextHeader(&TenantContext{
		OrgId:   adminOrg.AdminOrg.ID,
		OrgName: adminOrg.AdminOrg.Name,
	}) {
		headers.Add(k, v)
	}

	// If the call is a PUT, we prepare the body with the input settings
	var body io.Reader
	if method == http.MethodPut {
		text := bytes.Buffer{}
		encoder := xml.NewEncoder(&text)
		err = encoder.Encode(*payload)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(text.String())
	}

	// Perform the HTTP call with the custom headers and obtained API version
	req := adminOrg.client.newRequest(nil, nil, method, *endpoint, body, getHighestOidcApiVersion(adminOrg.client), headers)
	resp, err := checkResp(adminOrg.client.Http.Do(req))

	// Check the errors and get the response
	switch method {
	case http.MethodDelete:
		if err != nil {
			return nil, fmt.Errorf("error deleting Organization OpenID Connect settings: %s", err)
		}
		if resp != nil && resp.StatusCode != http.StatusNoContent {
			return nil, fmt.Errorf("error deleting Organization OpenID Connect settings, expected status code %d - received %d", http.StatusNoContent, resp.StatusCode)
		}
		return nil, nil
	case http.MethodGet:
		if err != nil {
			return nil, fmt.Errorf("error getting Organization OpenID Connect settings: %s", err)
		}
		var result types.OrgOAuthSettings
		err = decodeBody(types.BodyTypeXML, resp, &result)
		if err != nil {
			return nil, fmt.Errorf("error decoding Organization OpenID Connect settings: %s", err)
		}
		return &result, nil
	case http.MethodPut:
		if err != nil {
			return nil, fmt.Errorf("error setting Organization OpenID Connect settings: %s", err)
		}
		// Note: This branch of the switch should be exactly the same as the GET operation, however there is a bug found in VCD 10.5.1.1:
		// the PUT call returns a wrong redirect URL.
		// For that reason, we ignore the response body and call GetOpenIdConnectSettings() to return the correct response body to the caller.
		if resp != nil && resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error saving Organization OpenID Connect settings, expected status code %d - received %d", http.StatusOK, resp.StatusCode)
		}
		return adminOrg.GetOpenIdConnectSettings()
	default:
		return nil, fmt.Errorf("not supported HTTP method %s", method)
	}
}

// oidcValidateConnection executes a test probe against the given endpoint to validate that the client
// can establish a connection.
func oidcValidateConnection(client *Client, endpoint string) error {
	uri, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	isSecure := strings.ToLower(uri.Scheme) == "https"

	rawPort := uri.Port()
	if rawPort == "" {
		rawPort = "80"
		if isSecure {
			rawPort = "443"
		}
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		return err
	}

	result, err := client.TestConnection(types.TestConnection{
		Host:   uri.Hostname(),
		Port:   port,
		Secure: &isSecure,
	})
	if err != nil {
		return err
	}

	if result.TargetProbe == nil || !result.TargetProbe.CanConnect || (isSecure && !result.TargetProbe.SSLHandshake) {
		return fmt.Errorf("could not establish a connection to %s://%s", uri.Scheme, uri.Host)
	}
	return nil
}

// oidcConfigureWithEndpoint uses the given endpoint to retrieve an OpenID Connect configuration
func oidcConfigureWithEndpoint(client *Client, orgHref, endpoint string) (types.OrgOAuthSettings, error) {
	payload := types.OpenIdProviderInfo{
		Xmlns:                               types.XMLNamespaceVCloud,
		OpenIdProviderConfigurationEndpoint: endpoint,
	}
	var result types.OpenIdProviderConfiguration

	_, err := client.ExecuteRequestWithApiVersion(orgHref+"/settings/oauth/openIdProviderConfig", http.MethodPost,
		types.MimeOpenIdProviderInfoXml, "error getting OpenID Connect settings from endpoint: %s", payload, &result,
		getHighestOidcApiVersion(client))
	if err != nil {
		return types.OrgOAuthSettings{}, err
	}

	return result.OrgOAuthSettings, nil
}

// getHighestOidcApiVersion tries to get the highest possible version for the OpenID Connect endpoint
func getHighestOidcApiVersion(client *Client) string {
	// v38.1 adds CustomUiButtonLabel
	targetVersion := client.GetSpecificApiVersionOnCondition(">= 38.1", "38.1")
	if targetVersion != "38.1" {
		// v38.0 adds SendClientCredentialsAsAuthorizationHeader, UsePKCE,
		targetVersion = client.GetSpecificApiVersionOnCondition(">= 38.0", "38.0")
		if targetVersion != "38.0" {
			// v37.1 adds EnableIdTokenClaims
			targetVersion = client.GetSpecificApiVersionOnCondition(">= 37.1", "37.1")
		}
	} // Otherwise we get the default API version
	return targetVersion
}
