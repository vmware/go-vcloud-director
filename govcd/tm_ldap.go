// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TmLdapConfigure configures LDAP for the Tenant Manager "System" organization. If trustSslCertificate=true,
// it will automatically trust the certificate of LDAP server, only if settings.IsSsl=true.
// If settings.IsSsl=true and trustSslCertificate=false this method returns an error
func (vcdClient *VCDClient) TmLdapConfigure(settings *types.TmLdapSettings, trustSslCertificate bool) (*types.TmLdapSettings, error) {
	var trustedCert *TrustedCertificate

	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("this method is only supported in TM")
	}
	if settings == nil {
		return nil, fmt.Errorf("LDAP settings cannot be nil")
	}

	// Always test connection, because this method always changes LDAP. There's no NONE mode like in
	// the TmOrg receiver methods (that consume types.OrgLdapSettingsType)
	connResult, err := ldapValidateConnection(&vcdClient.Client, settings.HostName, settings.Port, settings.IsSsl)
	if err != nil {
		return nil, err
	}

	if !trustSslCertificate && connResult.TargetProbe.SSLResult == types.UntrustedCertificate {
		return nil, fmt.Errorf("cannot configure LDAP '%s:%d' over SSL without trusting the SSL certificate", settings.HostName, settings.Port)
	}

	if trustSslCertificate && connResult.TargetProbe.SSLResult == types.UntrustedCertificate {
		// We retrieve existing LDAP configuration to check if the given input settings are for a fresh connection, or to override
		// an existing one
		existing, err := vcdClient.TmGetLdapConfiguration()
		if err != nil {
			return nil, fmt.Errorf("error retrieving existing LDAP configuration: %s", err)
		}

		// If SSL is configured and retrieved settings are empty, or the hostname changed, we need to trust the certificate
		needsCheckCert := settings.IsSsl && (existing.HostName == "" || existing.HostName != settings.HostName)
		if needsCheckCert {
			trustedCert, err = trustCertificate(vcdClient, nil, settings.HostName, connResult.TargetProbe.CertificateChain)
			if err != nil {
				return nil, fmt.Errorf("could not trust certificate for %s, Connection result: '%s', SSL result: '%s': %s",
					settings.HostName, connResult.TargetProbe.ConnectionResult, connResult.TargetProbe.SSLResult, err)
			}
		}
	}

	result, err := ldapExecuteRequest(vcdClient, "", http.MethodPut, settings)
	if err != nil {
		// An error happened, so we must clean the trusted certificate
		if trustedCert != nil {
			innerErr := trustedCert.Delete()
			if innerErr != nil {
				return nil, fmt.Errorf("%s - Also, cleanup of SSL certificate '%s' failed: %s", err, trustedCert.TrustedCertificate.ID, innerErr)
			}
		}
		return nil, err
	}
	return result.(*types.TmLdapSettings), nil
}

// LdapConfigure configures LDAP for the receiver Organization. If trustSslCertificate=true,
// it will automatically trust the certificate of LDAP server, only if settings.IsSsl=true.
// If settings.IsSsl=true and trustSslCertificate=false this method returns an error
func (org *TmOrg) LdapConfigure(settings *types.OrgLdapSettingsType, trustSslCertificate bool) (*types.OrgLdapSettingsType, error) {
	var trustedCert *TrustedCertificate

	if settings == nil {
		return nil, fmt.Errorf("LDAP settings cannot be nil")
	}

	// Validate the connection with LDAP only when types.LdapModeCustom mode is selected (as it's the only one that configures an LDAP server)
	if settings.CustomOrgLdapSettings != nil && settings.OrgLdapMode == types.LdapModeCustom {
		connResult, err := ldapValidateConnection(&org.vcdClient.Client, settings.CustomOrgLdapSettings.HostName, settings.CustomOrgLdapSettings.Port, settings.CustomOrgLdapSettings.IsSsl)
		if err != nil {
			return nil, err
		}

		if !trustSslCertificate && connResult.TargetProbe.SSLResult == types.UntrustedCertificate {
			return nil, fmt.Errorf("cannot configure LDAP '%s:%d' over SSL without trusting the SSL certificate", settings.CustomOrgLdapSettings.HostName, settings.CustomOrgLdapSettings.Port)
		}

		if trustSslCertificate && connResult.TargetProbe.SSLResult == types.UntrustedCertificate {
			// We retrieve existing LDAP configuration to check if the given input settings are for a fresh connection, or to override
			// an existing one
			existing, err := org.GetLdapConfiguration()
			if err != nil {
				return nil, fmt.Errorf("error retrieving existing LDAP configuration: %s", err)
			}

			// Pre-condition here: Settings are always LDAP=CUSTOM and CustomOrgLdapSettings is not nil.
			// Just check if existing LDAP config is new (different from CUSTOM), or if the host changed, as we need to trust certificate
			// in that case
			needsCheckCert := settings.CustomOrgLdapSettings.IsSsl && (existing.OrgLdapMode != types.LdapModeCustom ||
				(existing.CustomOrgLdapSettings != nil && existing.CustomOrgLdapSettings.HostName != settings.CustomOrgLdapSettings.HostName))
			if needsCheckCert {
				trustedCert, err = trustCertificate(org.vcdClient, &TenantContext{
					OrgId:   org.TmOrg.ID,
					OrgName: org.TmOrg.Name,
				}, settings.CustomOrgLdapSettings.HostName, connResult.TargetProbe.CertificateChain)
				if err != nil {
					return nil, fmt.Errorf("could not trust certificate for %s, Connection result: '%s', SSL result: '%s': %s",
						settings.CustomOrgLdapSettings.HostName, connResult.TargetProbe.ConnectionResult, connResult.TargetProbe.SSLResult, err)
				}
			}
		}
	}

	result, err := ldapExecuteRequest(org.vcdClient, org.TmOrg.ID, http.MethodPut, settings)
	if err == nil {
		return result.(*types.OrgLdapSettingsType), nil
	}

	// An error happened, so we must clean the trusted certificate
	if trustedCert != nil {
		innerErr := trustedCert.Delete()
		if innerErr != nil {
			return nil, fmt.Errorf("%s - Also, cleanup of SSL certificate '%s' failed: %s", err, trustedCert.TrustedCertificate.ID, innerErr)
		}
	}
	return nil, err
}

// TmLdapDisable wraps LdapConfigure to disable LDAP configuration for the "System" organization.
// Disabling LDAP does not remove any trusted certificate.
func (vcdClient *VCDClient) TmLdapDisable() error {
	if !vcdClient.Client.IsTm() {
		return fmt.Errorf("this method is only supported in TM")
	}
	// It is a different endpoint, so we don't need to check certificates.
	_, err := ldapExecuteRequest(vcdClient, "", http.MethodDelete, nil)
	return err
}

// LdapDisable wraps LdapConfigure to disable LDAP configuration for the given organization
// Disabling LDAP does not remove any trusted certificate.
func (org *TmOrg) LdapDisable() error {
	// For Orgs, deletion is PUT call with empty payload. We don't need to check certificates.
	_, err := ldapExecuteRequest(org.vcdClient, org.TmOrg.ID, http.MethodPut, &types.OrgLdapSettingsType{OrgLdapMode: types.LdapModeNone})
	return err
}

// TmGetLdapConfiguration retrieves LDAP configuration structure for the "System" organization in Tenant Manager
func (vcdClient *VCDClient) TmGetLdapConfiguration() (*types.TmLdapSettings, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("this method is only supported in TM")
	}
	result, err := ldapExecuteRequest(vcdClient, "", http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	return result.(*types.TmLdapSettings), nil
}

// GetLdapConfiguration retrieves LDAP configuration structure of the given organization
func (org *TmOrg) GetLdapConfiguration() (*types.OrgLdapSettingsType, error) {
	result, err := ldapExecuteRequest(org.vcdClient, org.TmOrg.ID, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	return result.(*types.OrgLdapSettingsType), nil
}

// ldapValidateConnection executes a test probe against the given endpoint to validate that the client
// can establish a connection.
func ldapValidateConnection(client *Client, endpoint string, port int, isSecure bool) (*types.TestConnectionResult, error) {
	uri, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	result, err := client.TestConnection(types.TestConnection{
		Host:                          endpoint,
		HostnameVerificationAlgorithm: "LDAPS",
		Port:                          port,
		Secure:                        &isSecure,
	})
	if err != nil {
		return nil, err
	}

	if result.TargetProbe == nil || !result.TargetProbe.CanConnect || result.TargetProbe.ConnectionResult != "SUCCESS" {
		return nil, fmt.Errorf("could not establish a connection to %s. Result: %s", uri.String(), result.TargetProbe.Result)
	}
	return result, nil
}

// ldapExecuteRequest executes a request to the LDAP endpoint with the given payload and HTTP method
func ldapExecuteRequest(vcdClient *VCDClient, orgId, method string, payload interface{}) (interface{}, error) {
	if method == http.MethodPut && payload == nil {
		return nil, fmt.Errorf("the LDAP settings cannot be nil when performing a PUT call")
	}

	var endpoint *url.URL
	var err error
	if orgId != "" {
		endpoint, err = url.ParseRequestURI(fmt.Sprintf("%s/admin/org/%s/settings/ldap", vcdClient.Client.VCDHREF.String(), extractUuid(orgId)))
	} else {
		endpoint, err = url.ParseRequestURI(fmt.Sprintf("%s/admin/extension/settings/ldapSettings", vcdClient.Client.VCDHREF.String()))
	}
	if err != nil {
		return nil, err
	}

	// If the call is PUT, we prepare the body with the input settings
	var body io.Reader
	if method == http.MethodPut {
		text, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(text))
	}
	// Minimum version is 40.0 for TM LDAP
	apiVersion := vcdClient.Client.APIVersion
	if vcdClient.Client.APIClientVersionIs("< 40.0") {
		apiVersion = "40.0"
	}

	// Set headers with content type + version
	headers := http.Header{}
	headers.Set("Accept", fmt.Sprintf("%s;version=%s", types.JSONAllMime, apiVersion))
	headers.Set("Content-Type", types.JSONAllMime)

	// Perform the HTTP call with the obtained endpoint and method
	req := vcdClient.Client.newRequest(nil, nil, method, *endpoint, body, "", headers)
	resp, err := checkJsonResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error performing %s call to LDAP settings: %s", method, err)
	}

	// Other than DELETE with get a body response
	if method != http.MethodDelete {
		// Organization result type is different from System type
		if orgId != "" {
			var result types.OrgLdapSettingsType
			err = decodeBody(types.BodyTypeJSON, resp, &result)
			if err != nil {
				return nil, fmt.Errorf("error decoding Organization LDAP settings: %s", err)
			}
			return &result, nil
		} else {
			var result types.TmLdapSettings
			err = decodeBody(types.BodyTypeJSON, resp, &result)
			if err != nil {
				return nil, fmt.Errorf("error decoding LDAP settings: %s", err)
			}
			return &result, nil
		}
	}
	return nil, nil
}
