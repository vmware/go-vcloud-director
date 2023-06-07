//go:build api || functional || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_ServiceAccount(check *C) {
	isApiTokenEnabled, err := vcd.client.Client.VersionEqualOrGreater("10.4.0", 3)
	check.Assert(err, IsNil)
	if !isApiTokenEnabled {
		check.Skip("This test requires VCD 10.4.0 or greater")
	}

	org, err := vcd.client.GetOrgByName(vcd.config.Provider.SysOrg)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	serviceAccount, err := org.CreateServiceAccount(
		check.TestName(),
		"urn:vcloud:role:System%20Administrator",
		"12345678-1234-1234-1234-1234567890ab",
		"",
		"",
	)
	check.Assert(err, IsNil)
	check.Assert(serviceAccount, NotNil)
	check.Assert(serviceAccount.ServiceAccount.Status, Equals, "CREATED")

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	check.Assert(err, IsNil)

	AddToCleanupListOpenApi(check.TestName(), check.TestName(), endpoint+serviceAccount.ServiceAccount.ID)

	err = serviceAccount.Authorize()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.ServiceAccount.Status, Equals, "REQUESTED")

	err = serviceAccount.Grant()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.ServiceAccount.Status, Equals, "GRANTED")

	_, err = serviceAccount.GetInitialApiToken()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.ServiceAccount.Status, Equals, "ACTIVE")

	err = serviceAccount.Revoke()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.ServiceAccount.Status, Equals, "CREATED")

	err = serviceAccount.Delete()
	check.Assert(err, IsNil)

	notFound, err := org.GetServiceAccountById(serviceAccount.ServiceAccount.ID)
	check.Assert(err, NotNil)
	check.Assert(notFound, IsNil)
}
