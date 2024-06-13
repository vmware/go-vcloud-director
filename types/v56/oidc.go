/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

// OrgOAuthSettings contains OAuth identity provider settings for an Organization.
type OrgOAuthSettings struct {
	Xmlns string `xml:"xmlns,attr"`
	Href  string `xml:"href,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`

	Link                      *LinkList                   `xml:"Link,omitempty"`                      // A reference to an entity or operation associated with this object
	OrgRedirectUri            string                      `xml:"OrgRedirectUri,omitempty"`            // OAuth redirect URI for this org. This value is read only
	IssuerId                  string                      `xml:"IssuerId,omitempty"`                  // Issuer Id for the OAuth Identity Provider
	OAuthKeyConfigurations    *OAuthKeyConfigurationsList `xml:"OAuthKeyConfigurations,omitempty"`    // A list of OAuth Key configurations
	Enabled                   bool                        `xml:"Enabled"`                             // True if the OAuth Identity Provider for this organization is enabled. Unset or empty defaults to true
	ClientId                  string                      `xml:"ClientId,omitempty"`                  // Client ID for VCD to use when talking to the Identity Provider
	ClientSecret              string                      `xml:"ClientSecret,omitempty"`              // Client Secret for vCD to use when talking to the Identity Provider
	UserAuthorizationEndpoint string                      `xml:"UserAuthorizationEndpoint,omitempty"` // Identity Provider's OpenID Connect user authorization endpoint
	AccessTokenEndpoint       string                      `xml:"AccessTokenEndpoint,omitempty"`       // Identity Provider's OpenId Connect access token endpoint
	UserInfoEndpoint          string                      `xml:"UserInfoEndpoint,omitempty"`          // Identity Provider's OpenID Connect user info endpoint
	ScimEndpoint              string                      `xml:"ScimEndpoint,omitempty"`              // Identity Provider's SCIM user information endpoint
	Scope                     []string                    `xml:"Scope,omitempty"`                     // Scope that VCD needs access to for authenticating the user
	OIDCAttributeMapping      *OIDCAttributeMapping       `xml:"OIDCAttributeMapping,omitempty"`      // Custom claim keys for the /userinfo endpoint
	MaxClockSkew              int                         `xml:"MaxClockSkew,omitempty"`              // Allowed difference between token expiration and vCD system time in seconds
	JwksUri                   string                      `xml:"JwksUri,omitempty"`                   // Endpoint to fetch the keys from
	AutoRefreshKey            bool                        `xml:"AutoRefreshKey"`                      // Flag indicating whether VCD should auto-refresh the keys

	// Strategy to use when updated list of keys does not include keys known to VCD.
	// The values must be one of the below: ADD: Will add new keys to set of keys that VCD will use.
	// REPLACE: The retrieved list of keys will replace the existing list of keys and will become the definitive list of keys used by VCD going forward.
	// EXPIRE_AFTER: Keys known to VCD that are no longer returned by the OIDC server will be marked as expired, 'KeyExpireDurationInHours' specified hours after the key refresh is performed. After that later time, VCD will no longer use the keys.
	KeyRefreshStrategy string `xml:"KeyRefreshStrategy,omitempty"`

	KeyRefreshFrequencyInHours int    `xml:"KeyRefreshFrequencyInHours,omitempty"` // Time interval, in hours, between subsequent key refresh attempts
	KeyExpireDurationInHours   int    `xml:"KeyExpireDurationInHours,omitempty"`   // Duration in which the keys are set to expire
	WellKnownEndpoint          string `xml:"WellKnownEndpoint,omitempty"`          // Endpoint from the Identity Provider that serves OpenID Connect configuration value
	LastKeyRefreshAttempt      string `xml:"LastKeyRefreshAttempt,omitempty"`      // Last time refresh of the keys was attempted
	LastKeySuccessfulRefresh   string `xml:"LastKeySuccessfulRefresh,omitempty"`   // Last time refresh of the keys was successful

	// Added in v37.1
	EnableIdTokenClaims *bool `xml:"EnableIdTokenClaims"` // Flag indicating whether Id-Token Claims should be used when establishing user details
	// Added in v38.0
	UsePKCE                                    *bool `xml:"UsePKCE"`                                    // Flag indicating whether client must use PKCE (Proof Key for Code Exchange), which provides additional verification against potential authorization code interception. Default is false
	SendClientCredentialsAsAuthorizationHeader *bool `xml:"SendClientCredentialsAsAuthorizationHeader"` // Flag indicating whether client credentials should be sent as an Authorization header when fetching the token. Default is false, which means client credentials will be sent within the body of the request
	// Added in v38.1
	CustomUiButtonLabel *string `xml:"CustomUiButtonLabel,omitempty"` // Custom label to use when displaying this OpenID Connect configuration on the VCD login pane. If null, a default label will be used
}

// OAuthKeyConfigurationsList contains a list of OAuth Key configurations
type OAuthKeyConfigurationsList struct {
	Xmlns string `xml:"xmlns,attr"`

	Link                  *LinkList               `xml:"Link,omitempty"`                  // A reference to an entity or operation associated with this object
	OAuthKeyConfiguration []OAuthKeyConfiguration `xml:"OAuthKeyConfiguration,omitempty"` // OAuth key configuration
}

// OAuthKeyConfiguration describes the OAuth key configuration
type OAuthKeyConfiguration struct {
	Xmlns string `xml:"xmlns,attr"`

	Link           *LinkList `xml:"Link,omitempty"`           // A reference to an entity or operation associated with this object
	KeyId          string    `xml:"KeyId,omitempty"`          // Identifier for the key used by the Identity Provider. This key id is expected to be present in the header portion of OAuth tokens issued by the Identity provider
	Algorithm      string    `xml:"Algorithm,omitempty"`      // Identifies the cryptographic algorithm family of the key. Supported values are RSA and EC for asymmetric keys
	Key            string    `xml:"Key,omitempty"`            // PEM formatted key body. Key is used during validation of OAuth tokens for this Org
	ExpirationDate string    `xml:"ExpirationDate,omitempty"` // Expiration date for this key. If specified, tokens signed with this key should be considered invalid after this time
}

// OIDCAttributeMapping contains custom claim keys for the /userinfo endpoint
type OIDCAttributeMapping struct {
	Xmlns string `xml:"xmlns,attr"`

	Link                   *LinkList `xml:"Link,omitempty"`                   // A reference to an entity or operation associated with this object
	SubjectAttributeName   string    `xml:"SubjectAttributeName,omitempty"`   // The name of the OIDC attribute used to get the username from the IDP's userInfo
	EmailAttributeName     string    `xml:"EmailAttributeName,omitempty"`     // The name of the OIDC attribute used to get the email from the IDP's userInfo
	FullNameAttributeName  string    `xml:"FullNameAttributeName,omitempty"`  // The name of the OIDC attribute used to get the full name from the IDP's userInfo. The full name attribute overrides the use of the firstName and lastName attributes
	FirstNameAttributeName string    `xml:"FirstNameAttributeName,omitempty"` // The name of the OIDC attribute used to get the first name from the IDP's userInfo. This is only used if the Full Name key is not specified
	LastNameAttributeName  string    `xml:"LastNameAttributeName,omitempty"`  // The name of the OIDC attribute used to get the last name from the IDP's userInfo. This is only used if the Full Name key is not specified
	GroupsAttributeName    string    `xml:"GroupsAttributeName,omitempty"`    // The name of the OIDC attribute used to get the full name from the IDP's userInfo. The full name attribute overrides the use of the firstName and lastName attributes
	RolesAttributeName     string    `xml:"RolesAttributeName,omitempty"`     // The name of the OIDC attribute used to get the user's roles from the IDP's userInfo
}

// OpenIdProviderInfo contains the information about the OpenID Connect provider for creating initial org oauth settings
type OpenIdProviderInfo struct {
	Xmlns string `xml:"xmlns,attr"`

	OpenIdProviderConfigurationEndpoint string    `xml:"OpenIdProviderConfigurationEndpoint,omitempty"` // URL for the OAuth IDP well known openId connect configuration endpoint
	Link                                *LinkList `xml:"Link,omitempty"`                                // A reference to an entity or operation associated with this object
}

// OpenIdProviderConfiguration is result from reading the IDP OpenID Provider config endpoint
type OpenIdProviderConfiguration struct {
	Xmlns string `xml:"xmlns,attr"`

	Link                   *LinkList        `xml:"Link,omitempty"`                   // A reference to an entity or operation associated with this object
	OrgOAuthSettings       OrgOAuthSettings `xml:"OrgOAuthSettings,omitempty"`       // OrgOauthSettings object configured using information from the IDP
	ProviderConfigResponse string           `xml:"ProviderConfigResponse,omitempty"` // Raw response from the IDP config endpoint
}
