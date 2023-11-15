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

// Test_CreateNetworkPoolGeneve shows the creation of a "GENEVE" network pool
// using first the low-level method, then using the shortcut methods,
// and finally using the shortcut method without explicit transport zone
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
	var importableTransportZones []*types.TransportZone

	for _, tz := range transportZones {
		if !tz.AlreadyImported {
			importableTransportZones = append(importableTransportZones, tz)
		}
	}
	if len(importableTransportZones) == 0 {
		check.Skip("no unimported transport zone found")
	}

	for _, transportZone := range importableTransportZones {
		config := types.NetworkPool{
			Name:        networkPoolName,
			Description: "test network pool geneve",
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
		}, func(_ *NetworkPool) {
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
			return vcd.client.CreateNetworkPoolGeneve(networkPoolName, "test network pool geneve", manager.Name, transportZone.Name, types.BackingUseExplicit)
		}, nil, check)
	}
	if len(importableTransportZones) == 1 {
		// When no transport zone name is provided and there is only one TZ, we ask for that (unnamed) only one to be used
		runTestCreateNetworkPool("geneve-names-no-tz-name-only-element", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPoolGeneve(networkPoolName, "test network pool geneve", manager.Name, "", types.BackingUseWhenOnlyOne)
		}, nil, check)
	}
	// When no transport zone name is provided, the first one available will be used
	runTestCreateNetworkPool("geneve-names-no-tz-name-first-element", func() (*NetworkPool, error) {
		return vcd.client.CreateNetworkPoolGeneve(networkPoolName, "test network pool geneve", manager.Name, "", types.BackingUseFirstAvailable)
	}, nil, check)
}

// Test_CreateNetworkPoolPortgroup shows the creation of a "PORTGROUP_BACKED" network pool
// using first the low-level method, then using the shortcut methods,
// and finally using the shortcut method without explicit port group
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
	params.Set("filter", "virtualCenter.id=="+vCenter.VSphereVCenter.VcId)
	portgroups, err := vcd.client.GetAllVcenterImportableDvpgs(params)
	check.Assert(err, IsNil)
	check.Assert(len(portgroups) > 0, Equals, true)

	var sameHost []*VcenterImportableDvpg
	for _, pg := range portgroups {

		for _, other := range portgroups {
			if len(sameHost) > 0 {
				break
			}
			if other.VcenterImportableDvpg.BackingRef.ID == pg.VcenterImportableDvpg.BackingRef.ID {
				continue
			}
			if other.Parent().ID == pg.Parent().ID {
				sameHost = append(sameHost, pg)
				sameHost = append(sameHost, other)
				break
			}
		}
		config := types.NetworkPool{
			Name:        networkPoolName,
			Description: "test network pool port group",
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

		runTestCreateNetworkPool("port-group-full-config-("+pg.VcenterImportableDvpg.BackingRef.Name+")", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPool(&config)
		}, nil, check)
		runTestCreateNetworkPool("port-group-names-("+pg.VcenterImportableDvpg.BackingRef.Name+")", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPoolPortGroup(networkPoolName, "test network pool port group", vCenter.VSphereVCenter.Name,
				[]string{pg.VcenterImportableDvpg.BackingRef.Name}, types.BackingUseExplicit)
		}, nil, check)
	}
	if len(sameHost) == 2 {
		names := []string{
			sameHost[0].VcenterImportableDvpg.BackingRef.Name,
			sameHost[1].VcenterImportableDvpg.BackingRef.Name,
		}
		runTestCreateNetworkPool("port-group-multi-names", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPoolPortGroup(networkPoolName, "test network pool port group",
				vCenter.VSphereVCenter.Name, names, types.BackingUseExplicit)
		}, nil, check)
	}
	if len(portgroups) == 1 {
		// When no port group name is provided, and only one is available, we ask for that (unnamed) one to be used
		runTestCreateNetworkPool("port-group-names-no-pg-name-only-element", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPoolPortGroup(networkPoolName, "test network pool port group", vCenter.VSphereVCenter.Name, []string{""}, types.BackingUseWhenOnlyOne)
		}, nil, check)
	}
	// When no port group name is provided, the first one available will be used
	runTestCreateNetworkPool("port-group-names-no-pg-name-first-element", func() (*NetworkPool, error) {
		return vcd.client.CreateNetworkPoolPortGroup(networkPoolName, "test network pool port group", vCenter.VSphereVCenter.Name, []string{""}, types.BackingUseFirstAvailable)
	}, nil, check)
}

// Test_CreateNetworkPoolVlan shows the creation of a "VLAN" network pool
// using first the low-level method, then using the shortcut methods,
// and finally using the shortcut method without explicit distributed switch
func (vcd *TestVCD) Test_CreateNetworkPoolVlan(check *C) {
	if vcd.skipAdminTests {
		check.Skip("this test requires system administrator privileges")
	}
	if vcd.config.VCD.VimServer == "" {
		check.Skip("no vCenter found in configuration")
	}

	vCenter, err := vcd.client.GetVCenterByName(vcd.config.VCD.VimServer)
	check.Assert(err, IsNil)

	networkPoolName := check.TestName()

	switches, err := vcd.client.GetAllVcenterDistributedSwitches(vCenter.VSphereVCenter.VcId, nil)
	check.Assert(err, IsNil)
	if len(switches) == 0 {
		check.Skip("no available distributed found in vCenter")
	}
	// range ID for network pools
	ranges := []types.VlanIdRange{
		{StartId: 1, EndId: 100},
		{StartId: 201, EndId: 300},
	}
	// updateWithRanges updates the network pool
	updateWithRanges := func(pool *NetworkPool) {
		check.Assert(len(pool.NetworkPool.Backing.VlanIdRanges.Values), Equals, 2)
		pool.NetworkPool.Backing.VlanIdRanges.Values = []types.VlanIdRange{{StartId: 1001, EndId: 2000}}

		updatedName := pool.NetworkPool.Name + "-changed"
		updatedDescription := pool.NetworkPool.Description + " - changed"
		pool.NetworkPool.Name = updatedName
		pool.NetworkPool.Description = updatedDescription
		err = pool.Update()
		check.Assert(err, IsNil)
		retrievedNetworkPool, err := pool.vcdClient.GetNetworkPoolById(pool.NetworkPool.Id)
		check.Assert(err, IsNil)
		check.Assert(retrievedNetworkPool, NotNil)
		check.Assert(retrievedNetworkPool.NetworkPool.Id, Equals, pool.NetworkPool.Id)
		check.Assert(retrievedNetworkPool.NetworkPool.Name, Equals, updatedName)
		check.Assert(retrievedNetworkPool.NetworkPool.Description, Equals, updatedDescription)

		err = pool.Update()
		check.Assert(err, IsNil)
		newPool, err := vcd.client.GetNetworkPoolById(pool.NetworkPool.Id)
		check.Assert(err, IsNil)
		check.Assert(len(newPool.NetworkPool.Backing.VlanIdRanges.Values), Equals, 1)
	}
	for _, sw := range switches {
		config := types.NetworkPool{
			Name:        networkPoolName,
			Description: "test network pool VLAN",
			PoolType:    types.NetworkPoolVlanType,
			ManagingOwnerRef: types.OpenApiReference{
				Name: vCenter.VSphereVCenter.Name,
				ID:   vCenter.VSphereVCenter.VcId,
			},
			Backing: types.NetworkPoolBacking{
				VlanIdRanges: types.VlanIdRanges{
					Values: ranges,
				},
				VdsRefs: []types.OpenApiReference{
					{
						Name: sw.BackingRef.Name,
						ID:   sw.BackingRef.ID,
					},
				},
				ProviderRef: types.OpenApiReference{
					Name: vCenter.VSphereVCenter.Name,
					ID:   vCenter.VSphereVCenter.VcId,
				},
			},
		}

		runTestCreateNetworkPool("vlan-full-config-("+sw.BackingRef.Name+")",
			func() (*NetworkPool, error) {
				return vcd.client.CreateNetworkPool(&config)
			},
			updateWithRanges,
			check)

		runTestCreateNetworkPool("vlan-names-("+sw.BackingRef.Name+")",
			func() (*NetworkPool, error) {
				return vcd.client.CreateNetworkPoolVlan(networkPoolName, "test network pool VLAN", vCenter.VSphereVCenter.Name, sw.BackingRef.Name, ranges, types.BackingUseExplicit)
			},
			updateWithRanges,
			check)
	}
	if len(switches) == 1 {
		// When no switch name is provided, and only one is available, we ask to use that (unnamed) one
		runTestCreateNetworkPool("vlan-names-no-sw-name-only-element", func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPoolVlan(networkPoolName, "test network pool VLAN", vCenter.VSphereVCenter.Name, "", ranges, types.BackingUseWhenOnlyOne)
		},
			updateWithRanges,
			check)
	}
	// When no switch name is provided, the first one available will be used
	runTestCreateNetworkPool("vlan-names-no-sw-name-first-element", func() (*NetworkPool, error) {
		return vcd.client.CreateNetworkPoolVlan(networkPoolName, "test network pool VLAN", vCenter.VSphereVCenter.Name, "", ranges, types.BackingUseFirstAvailable)
	},
		updateWithRanges,
		check)
}

// runTestCreateNetworkPool runs a generic test for network pool creation, using `creationFunc` for creating the object
// and `postCreation` to run updates or other management actions
func runTestCreateNetworkPool(label string, creationFunc func() (*NetworkPool, error), postCreation func(pool *NetworkPool), check *C) {
	fmt.Printf("[test create network pool] %s\n", label)

	networkPool, err := creationFunc()
	check.Assert(err, IsNil)
	defer func() {
		if networkPool != nil {
			_ = networkPool.Delete()
		}
	}()
	check.Assert(networkPool, NotNil)
	networkPoolName := networkPool.NetworkPool.Name
	if postCreation != nil {
		postCreation(networkPool)
		// Refresh the network pool
		networkPool, err = networkPool.vcdClient.GetNetworkPoolById(networkPool.NetworkPool.Id)
		check.Assert(err, IsNil)
	}

	// if no update was run through the postCreation
	if networkPool.NetworkPool.Name == networkPoolName {
		updatedName := networkPool.NetworkPool.Name + "-update"
		updatedDescription := networkPool.NetworkPool.Description + " - update"
		networkPool.NetworkPool.Name = updatedName
		networkPool.NetworkPool.Description = updatedDescription
		err = networkPool.Update()
		check.Assert(err, IsNil)
		retrievedNetworkPool, err := networkPool.vcdClient.GetNetworkPoolById(networkPool.NetworkPool.Id)
		check.Assert(err, IsNil)
		check.Assert(retrievedNetworkPool, NotNil)
		check.Assert(retrievedNetworkPool.NetworkPool.Id, Equals, networkPool.NetworkPool.Id)
		check.Assert(retrievedNetworkPool.NetworkPool.Name, Equals, updatedName)
		check.Assert(retrievedNetworkPool.NetworkPool.Description, Equals, updatedDescription)
	}

	err = networkPool.Delete()
	check.Assert(err, IsNil)
	networkPool = nil
}
