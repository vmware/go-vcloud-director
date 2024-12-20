/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// TmLdapConfigure configures LDAP for the Tenant Manager "System" organization
func (vcdClient *VCDClient) TmLdapConfigure(settings *types.OrgLdapSettingsType) (*types.OrgLdapSettingsType, error) {
	if !vcdClient.Client.IsTm() {
		return nil, fmt.Errorf("this method is only supported in TM")
	}
	return nil, nil
}

// LdapConfigure configures LDAP for the given organization
func (org *TmOrg) LdapConfigure(settings *types.OrgLdapSettingsType) (*types.OrgLdapSettingsType, error) {
	return nil, nil
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
	return nil, nil
}

// GetLdapConfiguration retrieves LDAP configuration structure of the given organization
func (org *TmOrg) GetLdapConfiguration() (*types.OrgLdapSettingsType, error) {
	return nil, nil
}
