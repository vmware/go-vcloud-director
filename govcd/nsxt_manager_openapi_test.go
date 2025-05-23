//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtManagerOpenApi(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	if !vcd.config.Tm.CreateNsxtManager {
		check.Skip("Skipping NSX-T Manager creation")
	}

	cfg := &types.NsxtManagerOpenApi{
		Name:     check.TestName(),
		Username: vcd.config.Tm.NsxtManagerUsername,
		Password: vcd.config.Tm.NsxtManagerPassword,
		Url:      vcd.config.Tm.NsxtManagerUrl,
	}

	// Certificate must be trusted before adding NSX-T Manager
	url, err := url.Parse(cfg.Url)
	check.Assert(err, IsNil)
	_, err = vcd.client.AutoTrustHttpsCertificate(url, nil)
	check.Assert(err, IsNil)
	v, err := vcd.client.CreateNsxtManagerOpenApi(cfg)
	check.Assert(err, IsNil)
	check.Assert(v, NotNil)

	// Add to cleanup list
	PrependToCleanupListOpenApi(v.NsxtManagerOpenApi.ID, check.TestName(), types.OpenApiPathVcf+types.OpenApiEndpointNsxManagers+v.NsxtManagerOpenApi.ID)

	// Get By Name
	byName, err := vcd.client.GetNsxtManagerOpenApiByName(cfg.Name)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetNsxtManagerOpenApiById(v.NsxtManagerOpenApi.ID)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	// Get All
	allTmOrgs, err := vcd.client.GetAllNsxtManagersOpenApi(nil)
	check.Assert(err, IsNil)
	check.Assert(allTmOrgs, NotNil)
	check.Assert(len(allTmOrgs) > 0, Equals, true)

	// Update
	v.NsxtManagerOpenApi.Name = check.TestName() + "-updated"
	updated, err := v.Update(v.NsxtManagerOpenApi)
	check.Assert(err, IsNil)
	check.Assert(updated, NotNil)

	// Delete
	err = v.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetNsxtManagerOpenApiByName(cfg.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

}
