/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TmLdapConfigure configures LDAP for the Tenant Manager "System" organization
func (vcdClient *VCDClient) TmLdapConfigure(settings *types.OrgLdapSettingsType) (*types.OrgLdapSettingsType, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("this method is only supported in TM")
	}
	return ldapExecuteRequest(vcdClient, "", http.MethodPut, settings)
}

// LdapConfigure configures LDAP for the given organization
func (org *TmOrg) LdapConfigure(settings *types.OrgLdapSettingsType) (*types.OrgLdapSettingsType, error) {
	return ldapExecuteRequest(org.vcdClient, org.TmOrg.ID, http.MethodPut, settings)
}

// LdapDisable wraps LdapConfigure to disable LDAP configuration for the "System" organization
func (vcdClient *VCDClient) LdapDisable() error {
	if !vcdClient.Client.IsTm() {
		return fmt.Errorf("this method is only supported in TM")
	}
	_, err := vcdClient.TmLdapConfigure(&types.OrgLdapSettingsType{OrgLdapMode: types.LdapModeNone})
	return err
}

// LdapDisable wraps LdapConfigure to disable LDAP configuration for the given organization
func (org *TmOrg) LdapDisable() error {
	_, err := org.LdapConfigure(&types.OrgLdapSettingsType{OrgLdapMode: types.LdapModeNone})
	return err
}

// GetLdapConfiguration retrieves LDAP configuration structure for the "System" organization
func (vcdClient *VCDClient) GetLdapConfiguration() (*types.OrgLdapSettingsType, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("this method is only supported in TM")
	}
	return ldapExecuteRequest(vcdClient, "", http.MethodGet, nil)
}

// GetLdapConfiguration retrieves LDAP configuration structure of the given organization
func (org *TmOrg) GetLdapConfiguration() (*types.OrgLdapSettingsType, error) {
	return ldapExecuteRequest(org.vcdClient, org.TmOrg.ID, http.MethodGet, nil)
}

// ldapExecuteRequest executes a request to the LDAP endpoint with the given payload and HTTP method
func ldapExecuteRequest(vcdClient *VCDClient, orgId, method string, payload *types.OrgLdapSettingsType) (*types.OrgLdapSettingsType, error) {
	if method == http.MethodPut && payload == nil {
		return nil, fmt.Errorf("the LDAP settings cannot be nil when performing a PUT call")
	}

	var endpoint *url.URL
	var err error
	if orgId == "" {
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
		encoder := xml.NewEncoder(&text)
		err = encoder.Encode(*payload)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(text.String())
	}

	// Perform the HTTP call with the obtained endpoint and method
	req := vcdClient.Client.newRequest(nil, nil, method, *endpoint, body, vcdClient.Client.APIVersion, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return nil, fmt.Errorf("error getting LDAP settings: %s", err)
	}
	var result types.OrgLdapSettingsType
	err = decodeBody(types.BodyTypeJSON, resp, &result)
	if err != nil {
		return nil, fmt.Errorf("error decoding LDAP settings: %s", err)
	}
	return &result, nil
}
