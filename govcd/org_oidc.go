/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"strings"
)

// GetOpenIdConnectSettings retrieves the current federation (SAML) settings for a given organization
func (adminOrg *AdminOrg) GetOpenIdConnectSettings() (*types.OrgOAuthSettingsType, error) {
	if strings.TrimSpace(adminOrg.AdminOrg.ID) == "" {
		return nil, fmt.Errorf("the ID of the Organization is required to retrieve its OpenID Connect settings")
	}
	if strings.TrimSpace(adminOrg.AdminOrg.HREF) == "" {
		return nil, fmt.Errorf("the HREF of the Organization is required to retrieve its OpenID Connect settings")
	}

	var settings types.OrgOAuthSettingsType

	_, err := adminOrg.client.ExecuteRequest(adminOrg.AdminOrg.HREF+"/settings/oauth", http.MethodGet,
		types.MimeOAuthSettingsXml, "error getting Organization OpenID Connect settings: %s", nil, &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}
