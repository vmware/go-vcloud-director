//go:build providervdc || functional || ALL

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"net/url"
)

func (vcd *TestVCD) Test_GetNetworkPools(check *C) {

	if vcd.skipAdminTests {
		check.Skip("this test requires system administrator privileges")
	}
	knownNetworkPoolName := vcd.config.VCD.NsxtProviderVdc.NetworkPool
	networkPools, err := vcd.client.GetNetworkPoolSummaries(nil)
	check.Assert(err, IsNil)
	check.Assert(len(networkPools) > 0, Equals, true)

	checkNetworkPoolName := false
	foundNetworkPool := false
	if knownNetworkPoolName != "" {
		checkNetworkPoolName = true
	}

	for i, nps := range networkPools {
		if nps.Name == knownNetworkPoolName {
			foundNetworkPool = true
		}
		networkPoolById, err := vcd.client.GetNetworkPoolById(nps.Id)
		check.Assert(err, IsNil)
		check.Assert(networkPoolById, NotNil)
		check.Assert(networkPoolById.NetworkPool.Id, Equals, nps.Id)
		check.Assert(networkPoolById.NetworkPool.Name, Equals, nps.Name)

		networkPoolByName, err := vcd.client.GetNetworkPoolByName(nps.Name)
		check.Assert(err, IsNil)
		check.Assert(networkPoolByName, NotNil)
		check.Assert(networkPoolByName.NetworkPool.Id, Equals, nps.Id)
		check.Assert(networkPoolByName.NetworkPool.Name, Equals, nps.Name)
		if testVerbose {
			fmt.Printf("%d, %# v\n", i, pretty.Formatter(networkPoolByName.NetworkPool))
		}
	}
	if checkNetworkPoolName {
		check.Assert(foundNetworkPool, Equals, true)
	}
}

func (vcd *TestVCD) Test_CreateNetworkPoolGeneve(check *C) {
	if vcd.skipAdminTests {
		check.Skip("this test requires system administrator privileges")
	}
	if vcd.config.VCD.Nsxt.Manager == "" {
		check.Skip("no manager name is available")
	}
	networkPoolName := check.TestName()

	managers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(managers), Equals, 1)

	manager := managers[0]
	managerId := "urn:vcloud:nsxtmanager:" + extractUuid(manager.HREF)

	transportZones, err := vcd.client.GetAllNsxtTransportZones(managerId, nil)
	check.Assert(err, IsNil)
	if len(transportZones) == 0 {
		check.Skip("no available transport zones found")
	}
	//var transportZone *types.TransportZone
	var importableTransportZones []*types.TransportZone

	for _, tz := range transportZones {
		if !tz.AlreadyImported {
			//transportZone = tz
			//break
			importableTransportZones = append(importableTransportZones, tz)
		}
	}
	if len(importableTransportZones) == 0 {
		check.Skip("no unimported transport zone found")
	}
	//check.Assert(transportZone.AlreadyImported, Equals, false)

	for _, transportZone := range importableTransportZones {
		config := types.NetworkPool{
			Name:        networkPoolName,
			Description: "test network pool",
			PoolType:    types.NetworkPoolGeneveType,
			ManagingOwnerRef: types.OpenApiReference{
				Name: manager.Name,
				ID:   managerId,
			},
			Backing: types.NetworkPoolBacking{
				TransportZoneRef: types.OpenApiReference{
					Name: transportZone.Name,
					ID:   transportZone.Id,
				},
				ProviderRef: types.OpenApiReference{
					Name: manager.Name,
					ID:   managerId,
				},
			},
		}
		runTestCreateNetworkPool("geneve-full-config-("+transportZone.Name+")", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPool(&config)
		}, func() {
			tzs, err := vcd.client.GetAllNsxtTransportZones(managerId, nil)
			check.Assert(err, IsNil)
			for _, tz := range tzs {
				if tz.Name == transportZone.Name {
					check.Assert(tz.AlreadyImported, Equals, true)
				}
			}
		},
			check)
		runTestCreateNetworkPool("geneve-names-("+transportZone.Name+")", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPoolGeneve(networkPoolName, "test network pool", manager.Name, transportZone.Name)
		}, nil, check)
	}
	runTestCreateNetworkPool("geneve-names-no-tz-name", func() (*NetworkPool, error) {
		return vcd.client.CreateNetworkPoolGeneve(networkPoolName, "test network pool", manager.Name, "")
	}, nil, check)
}

func (vcd *TestVCD) Test_CreateNetworkPoolPortgroup(check *C) {
	if vcd.skipAdminTests {
		check.Skip("this test requires system administrator privileges")
	}
	if vcd.config.VCD.VimServer == "" {
		check.Skip("no vCenter found in configuration")
	}

	vCenter, err := vcd.client.GetVCenterByName(vcd.config.VCD.VimServer)
	check.Assert(err, IsNil)

	networkPoolName := check.TestName()

	var params = make(url.Values)
	params.Set("virtualCenter.id", vCenter.VSphereVCenter.VcId)
	portgroups, err := vcd.client.GetAllVcenterImportableDvpgs(params)
	check.Assert(err, IsNil)
	check.Assert(len(portgroups) > 0, Equals, true)

	for _, pg := range portgroups {

		config := types.NetworkPool{
			Name:        networkPoolName,
			Description: "test network pool",
			PoolType:    types.NetworkPoolPortGroupType,
			ManagingOwnerRef: types.OpenApiReference{
				Name: vCenter.VSphereVCenter.Name,
				ID:   vCenter.VSphereVCenter.VcId,
			},
			Backing: types.NetworkPoolBacking{
				PortGroupRefs: []types.OpenApiReference{
					{
						ID:   pg.VcenterImportableDvpg.BackingRef.ID,
						Name: pg.VcenterImportableDvpg.BackingRef.Name,
					},
				},
				ProviderRef: types.OpenApiReference{
					Name: vCenter.VSphereVCenter.Name,
					ID:   vCenter.VSphereVCenter.VcId,
				},
			},
		}

		runTestCreateNetworkPool("portgroup-full-config-("+pg.VcenterImportableDvpg.BackingRef.Name+")", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPool(&config)
		}, nil, check)
		runTestCreateNetworkPool("portgroup-names-("+pg.VcenterImportableDvpg.BackingRef.Name+")", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPoolPortGroup(networkPoolName, "test network pool", vCenter.VSphereVCenter.Name, pg.VcenterImportableDvpg.BackingRef.Name)
		}, nil, check)
	}
	runTestCreateNetworkPool("portgroup-names-no-pg-name", func() (*NetworkPool, error) {
		return vcd.client.CreateNetworkPoolPortGroup(networkPoolName, "test network pool", vCenter.VSphereVCenter.Name, "")
	}, nil, check)
}

func runTestCreateNetworkPool(label string, creationFunc func() (*NetworkPool, error), postCreation func(), check *C) {
	fmt.Printf("[test create network pool] %s\n", label)

	networkPool, err := creationFunc()
	check.Assert(err, IsNil)
	defer func() {
		if networkPool != nil {
			_ = networkPool.Delete()
		}
	}()
	check.Assert(networkPool, NotNil)
	if postCreation != nil {
		postCreation()
	}
	err = networkPool.Delete()
	check.Assert(err, IsNil)
	networkPool = nil
}
