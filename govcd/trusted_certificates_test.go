//go:build tm || ALL

package govcd

import (
	_ "embed"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

// Test_TrustedCertificatesSystem tests CRUD of certificates in System
func (vcd *TestVCD) Test_TrustedCertificatesSystem(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	cfg := &types.TrustedCertificate{
		Alias:       check.TestName(),
		Certificate: certificate, // using embedded certificate
	}

	v, err := vcd.client.CreateTrustedCertificate(cfg, nil)
	check.Assert(err, IsNil)
	check.Assert(v, NotNil)

	// Add to cleanup list
	PrependToCleanupListOpenApi(v.TrustedCertificate.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTrustedCertificates+v.TrustedCertificate.ID)

	// Get By Name
	byName, err := vcd.client.GetTrustedCertificateByAlias(cfg.Alias, nil)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetTrustedCertificateById(v.TrustedCertificate.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	// Get All
	allTmOrgs, err := vcd.client.GetAllTrustedCertificates(nil, nil)
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

	notFoundByName, err := vcd.client.GetTrustedCertificateByAlias(updated.TrustedCertificate.Alias, nil)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)
}

// Test_TrustedCertificatesTenant tests CRUD of certificates in an Organization
func (vcd *TestVCD) Test_TrustedCertificatesTenant(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	// We are testing trusted certificates for a regular Organization
	orgSettings := &types.TmOrg{
		Name:        check.TestName(),
		DisplayName: check.TestName(),
		IsEnabled:   true,
	}
	org, err := vcd.client.CreateTmOrg(orgSettings)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	defer func() {
		err = org.Disable()
		check.Assert(err, IsNil)
		err = org.Delete()
		check.Assert(err, IsNil)
	}()

	// Add Organization to cleanup list
	PrependToCleanupListOpenApi(org.TmOrg.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointOrgs+org.TmOrg.ID)

	cfg := &types.TrustedCertificate{
		Alias:       check.TestName(),
		Certificate: certificate, // using embedded certificate
	}

	v, err := org.CreateTrustedCertificate(cfg)
	check.Assert(err, IsNil)
	check.Assert(v, NotNil)

	// Get By Name
	byName, err := org.GetTrustedCertificateByAlias(cfg.Alias)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := org.GetTrustedCertificateById(v.TrustedCertificate.ID)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	// Get All
	allTmOrgs, err := org.GetAllTrustedCertificates(nil)
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

	notFoundByName, err := org.GetTrustedCertificateByAlias(updated.TrustedCertificate.Alias)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)
}
