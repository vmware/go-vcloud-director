/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TmLdapConfigure configures LDAP for the Tenant Manager "System" organization
func (vcdClient *VCDClient) TmLdapConfigure(settings *types.TmLdapSettings) (*types.TmLdapSettings, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("this method is only supported in TM")
	}

	result, err := ldapExecuteRequest(vcdClient, "", http.MethodPut, settings)
	if err != nil {
		return nil, err
	}
	return result.(*types.TmLdapSettings), nil
}

// LdapConfigure configures LDAP for the given organization
func (org *TmOrg) LdapConfigure(settings *types.OrgLdapSettingsType) (*types.OrgLdapSettingsType, error) {
	result, err := ldapExecuteRequest(org.vcdClient, org.TmOrg.ID, http.MethodPut, settings)
	if err != nil {
		return nil, err
	}
	return result.(*types.OrgLdapSettingsType), nil
}

// TmLdapDisable wraps LdapConfigure to disable LDAP configuration for the "System" organization
func (vcdClient *VCDClient) TmLdapDisable() error {
	if !vcdClient.Client.IsTm() {
		return fmt.Errorf("this method is only supported in TM")
	}
	_, err := ldapExecuteRequest(vcdClient, "", http.MethodDelete, nil)
	return err
}

// LdapDisable wraps LdapConfigure to disable LDAP configuration for the given organization
func (org *TmOrg) LdapDisable() error {
	_, err := ldapExecuteRequest(org.vcdClient, org.TmOrg.ID, http.MethodDelete, nil)
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

	// If the call is a PUT, we prepare the body with the input settings
	var body io.Reader
	if method == http.MethodPut {
		text := bytes.Buffer{}
		encoder := json.NewEncoder(&text)
		err = encoder.Encode(payload)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(text.String())
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
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error getting LDAP settings: %s", err)
	}
	if method == http.MethodPut {
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
