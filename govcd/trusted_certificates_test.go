//go:build tm || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	_ "embed"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TrustedCertificates(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	cfg := &types.TrustedCertificate{
		Alias:       check.TestName(),
		Certificate: certificate, // using embedded certificate
	}

	v, err := vcd.client.CreateTrustedCertificate(cfg)
	check.Assert(err, IsNil)
	check.Assert(v, NotNil)

	// Add to cleanup list
	PrependToCleanupListOpenApi(v.TrustedCertificate.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTrustedCertificates+v.TrustedCertificate.ID)

	// Get By Name
	byName, err := vcd.client.GetTrustedCertificateByAlias(cfg.Alias)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetTrustedCertificateById(v.TrustedCertificate.ID)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	// Get All
	allTmOrgs, err := vcd.client.GetAllTrustedCertificates(nil)
	check.Assert(err, IsNil)
	check.Assert(allTmOrgs, NotNil)
	check.Assert(len(allTmOrgs) > 0, Equals, true)

	// Update
	v.TrustedCertificate.Alias = check.TestName() + "-rename"
	updated, err := v.Update(v.TrustedCertificate)
	check.Assert(err, IsNil)
	check.Assert(updated, NotNil)

	// Delete
	err = v.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetTrustedCertificateByAlias(updated.TrustedCertificate.Alias)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)
}
