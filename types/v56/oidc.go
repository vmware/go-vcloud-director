/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

import (
	"net/url"
	"time"
)

// OrgOAuthSettingsType contains OAuth identity provider settings for an Organization.
type OrgOAuthSettingsType struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	Href  string `xml:"href,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	AccessTokenEndpoint        *url.URL `xml:"AccessTokenEndpoint,omitempty"`        // Identity Provider's OpenId Connect access token endpoint
	AutoRefreshKey             *bool    `xml:"AutoRefreshKey,omitempty"`             // Flag indicating whether VCD should auto-refresh the keys
	ClientId                   *string  `xml:"ClientId,omitempty"`                   // Client ID for vCD to use when talking to the Identity Provider
	ClientSecret               *string  `xml:"ClientSecret,omitempty"`               // Client Secret for vCD to use when talking to the Identity Provider
	Enabled                    *bool    `xml:"Enabled,omitempty"`                    // True if the OAuth Identity Provider for this organization is enabled. Unset or empty defaults to true
	IssuerId                   *string  `xml:"IssuerId,omitempty"`                   // Issuer Id for the OAuth Identity Provider
	JwksUri                    *url.URL `xml:"JwksUri,omitempty"`                    // Endpoint to fetch the keys from
	KeyExpireDurationInHours   *int     `xml:"KeyExpireDurationInHours,omitempty"`   // Duration in which the keys are set to expire
	KeyRefreshFrequencyInHours *int     `xml:"KeyRefreshFrequencyInHours,omitempty"` // Time interval, in hours, between subsequent key refresh attempts

	// Strategy to use when updated list of keys does not include keys known to VCD.
	// The values must be one of the below: ADD: Will add new keys to set of keys that VCD will use.
	// REPLACE: The retrieved list of keys will replace the existing list of keys and will become the definitive list of keys used by VCD going forward.
	// EXPIRE_AFTER: Keys known to VCD that are no longer returned by the OIDC server will be marked as expired, 'KeyExpireDurationInHours' specified hours after the key refresh is performed. After that later time, VCD will no longer use the keys.
	KeyRefreshStrategy *string `xml:"KeyRefreshStrategy,omitempty"`

	LastKeyRefreshAttempt     *time.Time                      `xml:"LastKeyRefreshAttempt,omitempty"`     // Last time refresh of the keys was attempted
	LastKeySuccessfulRefresh  *time.Time                      `xml:"LastKeySuccessfulRefresh,omitempty"`  // Last time refresh of the keys was successful
	Link                      *string                         `xml:"Link,omitempty"`                      // A reference to an entity or operation associated with this object
	MaxClockSkew              *int                            `xml:"MaxClockSkew,omitempty"`              // Allowed difference between token expiration and vCD system time in seconds
	OAuthKeyConfigurations    *OAuthKeyConfigurationsListType `xml:"OAuthKeyConfigurations,omitempty"`    // A list of OAuth Key configurations
	OIDCAttributeMapping      *OIDCAttributeMappingType       `xml:"OIDCAttributeMapping,omitempty"`      // Custom claim keys for the /userinfo endpoint
	OrgRedirectUri            string                          `xml:"OrgRedirectUri,omitempty"`            // OAuth redirect URI for this org. This value is read only
	ScimEndpoint              *url.URL                        `xml:"ScimEndpoint,omitempty"`              // Identity Provider's SCIM user information endpoint
	Scope                     *string                         `xml:"Scope,omitempty"`                     // Scope that VCD needs access to for authenticating the user
	UserAuthorizationEndpoint *url.URL                        `xml:"UserAuthorizationEndpoint,omitempty"` // Identity Provider's OpenID Connect user authorization endpoint
	UserInfoEndpoint          *url.URL                        `xml:"UserInfoEndpoint,omitempty"`          // Identity Provider's OpenID Connect user info endpoint
	WellKnownEndpoint         *url.URL                        `xml:"WellKnownEndpoint,omitempty"`         // Endpoint from the provider that serves OpenID Connect configuration values
}

// OAuthKeyConfigurationType describes the OAuth key configuration
type OAuthKeyConfigurationType struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	Href  string `xml:"href,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	Algorithm      string    `xml:"Algorithm"`       // Identifies the cryptographic algorithm family of the key. Supported values are RSA and EC for asymmetric keys
	ExpirationDate time.Time `xml:"ExpirationDate"`  // Expiration date for this key. If specified, tokens signed with this key should be considered invalid after this time
	Key            string    `xml:"Key"`             // PEM formatted key body. Key is used during validation of OAuth tokens for this Org
	KeyId          *string   `xml:"KeyId,omitempty"` // Identifier for the key used by the Identity Provider. This key id is expected to be present in the header portion of OAuth tokens issued by the Identity provider
	Link           *Link     `xml:"Link,omitempty"`  // A reference to an entity or operation associated with this object
}

// OAuthKeyConfigurationsListType contains a list of OAuth Key configurations
type OAuthKeyConfigurationsListType struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	Href  string `xml:"href,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	Link                  *Link                        `xml:"Link,omitempty"`                  // A reference to an entity or operation associated with this object
	OAuthKeyConfiguration []*OAuthKeyConfigurationType `xml:"OAuthKeyConfiguration,omitempty"` // OAuth key configuration
}

// OIDCAttributeMappingType contains custom claim keys for the /userinfo endpoint
type OIDCAttributeMappingType struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	Href  string `xml:"href,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	EmailAttributeName     *string `xml:"EmailAttributeName,omitempty"`     // The name of the OIDC attribute used to get the email from the IDP's userInfo
	FirstNameAttributeName *string `xml:"FirstNameAttributeName,omitempty"` // The name of the OIDC attribute used to get the first name from the IDP's userInfo. This is only used if the Full Name key is not specified
	FullNameAttributeName  *string `xml:"FullNameAttributeName,omitempty"`  // The name of the OIDC attribute used to get the full name from the IDP's userInfo. The full name attribute overrides the use of the firstName and lastName attributes
	GroupsAttributeName    *string `xml:"GroupsAttributeName,omitempty"`    // The name of the OIDC attribute used to get the full name from the IDP's userInfo. The full name attribute overrides the use of the firstName and lastName attributes
	LastNameAttributeName  *string `xml:"LastNameAttributeName,omitempty"`  // The name of the OIDC attribute used to get the last name from the IDP's userInfo. This is only used if the Full Name key is not specified
	Link                   *Link   `xml:"Link,omitempty"`                   // A reference to an entity or operation associated with this object
	RolesAttributeName     *string `xml:"RolesAttributeName,omitempty"`     // The name of the OIDC attribute used to get the user's roles from the IDP's userInfo
	SubjectAttributeName   *string `xml:"SubjectAttributeName,omitempty"`   // The name of the OIDC attribute used to get the username from the IDP's userInfo
}
