/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// ConfigureLdapMode allows to configure LDAP mode in use by the Org
func (adminOrg *AdminOrg) ConfigureLdapMode(settings *types.OrgLdapSettingsType) error {
	util.Logger.Printf("[DEBUG] Configuring LDAP mode for Org name %s", adminOrg.AdminOrg.Name)

	settings.Xmlns = types.XMLNamespaceVCloud
	// adminOrg.client.supportedVersions
	href := adminOrg.AdminOrg.HREF + "/settings/ldap"
	_, err := adminOrg.client.ExecuteRequest(href, http.MethodPut, types.MimeOrgLdapSettings,
		"error updating Ldap settings: %s", settings, nil)
	if err != nil {
		return fmt.Errorf("error updating LDAP mode for Org name '%s': %s", adminOrg.AdminOrg.Name, err)
	}

	return nil
}
