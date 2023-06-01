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

	saParams, err := vcd.client.RegisterServiceAccount(
		vcd.config.Provider.SysOrg,
		check.TestName(),
		"urn:vcloud:role:System%20Administrator",
		"12345678-1234-1234-1234-1234567890ab",
		"",
		"",
	)
	check.Assert(err, IsNil)
	check.Assert(saParams, NotNil)
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointServiceAccounts
	urn, err := BuildUrnWithUuid("urn:vcloud:serviceAccount:", saParams.ClientID)
	check.Assert(err, IsNil)

	AddToCleanupListOpenApi(check.TestName(), check.TestName(), endpoint+urn)

	serviceAccount, err := vcd.client.GetServiceAccountById(saParams.ClientID)
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.Status, Equals, "CREATED")

	err = serviceAccount.Authorize()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.Status, Equals, "REQUESTED")

	err = serviceAccount.Grant()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.Status, Equals, "GRANTED")

	_, err = serviceAccount.GetInitialApiToken()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.Status, Equals, "ACTIVE")

	err = serviceAccount.Revoke()
	check.Assert(err, IsNil)

	err = serviceAccount.Refresh()
	check.Assert(err, IsNil)
	check.Assert(serviceAccount.Status, Equals, "CREATED")

	err = vcd.client.DeleteServiceAccountByID(serviceAccount.ID)
	check.Assert(err, IsNil)
}
