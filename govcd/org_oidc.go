/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// GetOpenIdConnectSettings retrieves the current federation (SAML) settings for a given organization
func (adminOrg *AdminOrg) GetOpenIdConnectSettings() (*types.VcdOidcSettings, error) {
	return nil, nil
}
