/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/util"
)

func (adminOrg *AdminOrg) ConfigureLdapMode() error {
	util.Logger.Printf("[DEBUG] ConfigureLdap with org name %s", adminOrg.AdminOrg.Name)

	return nil
}

//
func (adminOrg *AdminOrg) ConfigureLdap() error {
	util.Logger.Printf("[DEBUG] ConfigureLdap with org name %s", adminOrg.AdminOrg.Name)

	return nil
}
