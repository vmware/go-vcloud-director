/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// GetOpenIdConnectSettings retrieves the current OpenID Connect settings for a given Organization
func (adminOrg *AdminOrg) GetOpenIdConnectSettings() (*types.OrgOAuthSettings, error) {
	if strings.TrimSpace(adminOrg.AdminOrg.HREF) == "" {
		return nil, fmt.Errorf("the HREF of the Organization is required to retrieve its OpenID Connect settings")
	}

	var settings types.OrgOAuthSettings

	_, err := adminOrg.client.ExecuteRequestWithApiVersion(adminOrg.AdminOrg.HREF+"/settings/oauth", http.MethodGet,
		types.MimeOAuthSettingsXml, "error getting Organization OpenID Connect settings: %s", nil, &settings,
		getHighestOidcApiVersion(adminOrg.client))
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// SetOpenIdConnectSettings sets the OpenID Connect configuration for a given Organization. If the well-known configuration
// endpoint is provided, the configuration is automatically retrieved from that URL.
// If other fields have been set in the input structure, the well-known configuration is overridden with these.
// If there are no fields informed, the configuration retrieved from the well-known configuration endpoint is applied as-is.
// ClientId, ClientSecret and Enabled properties are always mandatory, with and without well-known endpoint.
// This method returns an error if the settings can't be saved in VCD for any reason or if the provided settings are wrong.
func (adminOrg *AdminOrg) SetOpenIdConnectSettings(settings types.OrgOAuthSettings) (*types.OrgOAuthSettings, error) {
	if strings.TrimSpace(adminOrg.AdminOrg.HREF) == "" {
		return nil, fmt.Errorf("the HREF of the Organization is required to configure its OpenID Connect settings")
	}
	if settings.ClientId == nil || strings.TrimSpace(*settings.ClientId) == "" {
		return nil, fmt.Errorf("the Client ID is mandatory to configure OpenID Connect")
	}
	if settings.ClientSecret == nil || strings.TrimSpace(*settings.ClientSecret) == "" {
		return nil, fmt.Errorf("the Client Secret is mandatory to configure OpenID Connect")
	}
	if settings.Enabled == nil {
		return nil, fmt.Errorf("the OpenID Connect input settings must specify either enabled=true or enabled=false")
	}
	if settings.WellKnownEndpoint != nil {
		err := oidcValidateConnection(adminOrg.client, *settings.WellKnownEndpoint)
		if err != nil {
			return nil, err
		}
		wellKnownSettings, err := oidcConfigureWithEndpoint(adminOrg.client, adminOrg.AdminOrg.HREF, *settings.WellKnownEndpoint)
		if err != nil {
			return nil, err
		}
		// The following conditionals allow users to override the well-known automatic configuration values with their own,
		// mimicking what users can do in UI.
		// If an attribute was not set in the input settings, we pick the value that the well-known endpoint gave for that attribute,
		// but if it was explicitly set by the user, we take that one instead (overriding the well-known one).
		if settings.AccessTokenEndpoint == nil || *settings.AccessTokenEndpoint == "" {
			settings.AccessTokenEndpoint = wellKnownSettings.AccessTokenEndpoint
		}
		if settings.IssuerId == nil || *settings.IssuerId == "" {
			settings.IssuerId = wellKnownSettings.IssuerId
		}
		if settings.MaxClockSkew == nil {
			settings.MaxClockSkew = addrOf(60) // This is not returned, but a default value set in UI
		}
		if settings.JwksUri == nil || *settings.JwksUri == "" {
			settings.JwksUri = wellKnownSettings.JwksUri
		}
		if settings.UserInfoEndpoint == nil || *settings.UserInfoEndpoint == "" {
			settings.UserInfoEndpoint = wellKnownSettings.UserInfoEndpoint
		}
		if settings.UserAuthorizationEndpoint == nil || *settings.UserAuthorizationEndpoint == "" {
			settings.UserAuthorizationEndpoint = wellKnownSettings.UserAuthorizationEndpoint
		}
		if settings.ScimEndpoint == nil || *settings.ScimEndpoint == "" {
			settings.ScimEndpoint = wellKnownSettings.ScimEndpoint
		}
		if settings.Scope == nil || len(settings.Scope) == 0 {
			settings.Scope = wellKnownSettings.Scope
		}
		if settings.OIDCAttributeMapping == nil {
			settings.OIDCAttributeMapping = wellKnownSettings.OIDCAttributeMapping
		}
		if settings.OAuthKeyConfigurations == nil || len(settings.OAuthKeyConfigurations.OAuthKeyConfiguration) == 0 {
			settings.OAuthKeyConfigurations = wellKnownSettings.OAuthKeyConfigurations
		}
	}
	settings.Xmlns = types.XMLNamespaceVCloud
	if settings.OAuthKeyConfigurations != nil { // TODO: Can be nil? Check UI
		settings.OAuthKeyConfigurations.Xmlns = types.XMLNamespaceVCloud
		for i := range settings.OAuthKeyConfigurations.OAuthKeyConfiguration {
			settings.OAuthKeyConfigurations.OAuthKeyConfiguration[i].Xmlns = types.XMLNamespaceVCloud
		}
	}
	if settings.OIDCAttributeMapping != nil { // TODO: Can be nil? Check UI
		settings.OIDCAttributeMapping.Xmlns = types.XMLNamespaceVCloud
	}

	var createdSettings types.OrgOAuthSettings
	_, err := adminOrg.client.ExecuteRequestWithApiVersion(adminOrg.AdminOrg.HREF+"/settings/oauth", http.MethodPut,
		types.MimeOAuthSettingsXml, "error creating Organization OpenID Connect settings: %s", settings, &createdSettings,
		getHighestOidcApiVersion(adminOrg.client))
	if err != nil {
		return nil, err
	}

	return &createdSettings, nil
}

// DeleteOpenIdConnectSettings deletes the current OpenID Connect settings from a given Organization
func (adminOrg *AdminOrg) DeleteOpenIdConnectSettings() error {
	if strings.TrimSpace(adminOrg.AdminOrg.HREF) == "" {
		return fmt.Errorf("the HREF of the Organization is required to delete its OpenID Connect settings")
	}

	_, err := adminOrg.client.ExecuteRequest(adminOrg.AdminOrg.HREF+"/settings/oauth", http.MethodDelete,
		types.MimeOAuthSettingsXml, "error deleting Organization OpenID Connect settings: %s", nil, nil)
	if err != nil {
		return err
	}

	return nil
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
		return fmt.Errorf("could not establish a connection to %s", uri.String())
	}
	return nil
}

// oidcConfigureWithEndpoint uses the given endpoint to retrieve an OpenID Connect configuration
func oidcConfigureWithEndpoint(client *Client, orgHref, endpoint string) (*types.OrgOAuthSettings, error) {
	payload := types.OpenIdProviderInfo{
		Xmlns:                               types.XMLNamespaceVCloud,
		OpenIdProviderConfigurationEndpoint: endpoint,
	}
	var result types.OpenIdProviderConfiguration

	_, err := client.ExecuteRequestWithApiVersion(orgHref+"/settings/oauth/openIdProviderConfig", http.MethodPost,
		types.MimeOpenIdProviderInfoXml, "error getting OpenID Connect settings from endpoint: %s", payload, &result,
		getHighestOidcApiVersion(client))
	if err != nil {
		return nil, err
	}
	if result.OrgOAuthSettings == nil {
		return nil, fmt.Errorf("could not retrieve OpenID Connect configuration from %s, got a nil configuration", endpoint)
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
