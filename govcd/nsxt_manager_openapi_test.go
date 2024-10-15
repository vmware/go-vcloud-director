//go:build tm || functional || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmNsxtManager(check *C) {
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

	v, err := vcd.client.CreateNsxtManagerOpenApi(cfg)
	check.Assert(err, IsNil)
	check.Assert(v, NotNil)

	// Add to cleanup list
	PrependToCleanupListOpenApi(v.NsxtManagerOpenApi.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTmNsxManagers+v.NsxtManagerOpenApi.ID)

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
