//go:build tm || functional || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_VCenter(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	if !vcd.config.Tm.CreateVcenter {
		check.Skip("Skipping vCenter creation")
	}

	cfg := &types.VSphereVirtualCenter{
		Name:      check.TestName(),
		Username:  vcd.config.Tm.VcenterUsername,
		Password:  vcd.config.Tm.VcenterPassword,
		Url:       vcd.config.Tm.VcenterUrl,
		IsEnabled: true,
	}

	v, err := vcd.client.CreateVcenter(cfg)
	check.Assert(err, IsNil)
	check.Assert(v, NotNil)

	// Add to cleanup list
	PrependToCleanupListOpenApi(v.VSphereVCenter.VcId, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVirtualCenters+v.VSphereVCenter.VcId)

	// Get By Name
	byName, err := vcd.client.GetVCenterByName(cfg.Name)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetVCenterById(v.VSphereVCenter.VcId)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	// Get All
	allTmOrgs, err := vcd.client.GetAllVCenters(nil)
	check.Assert(err, IsNil)
	check.Assert(allTmOrgs, NotNil)
	check.Assert(len(allTmOrgs) > 0, Equals, true)

	// Update
	v.VSphereVCenter.IsEnabled = false
	updated, err := v.Update(v.VSphereVCenter)
	check.Assert(err, IsNil)
	check.Assert(updated, NotNil)

	// Delete
	err = v.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetVCenterByName(cfg.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)
}
