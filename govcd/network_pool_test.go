//go:build providervdc || functional || ALL

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
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

func (vcd *TestVCD) Test_CreateNetworkPool(check *C) {
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
	var transportZone *types.TransportZone
	for _, tz := range transportZones {
		if !tz.AlreadyImported {
			transportZone = tz
			break
		}
	}
	if transportZone == nil {
		check.Skip("no unimported transport zone found")
	}
	check.Assert(transportZone.AlreadyImported, Equals, false)

	config := types.NetworkPool{
		Name:        networkPoolName,
		Description: "test network pool",
		PoolType:    "GENEVE",
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
	runTestCreateNetworkPool("geneve-full-config",
		func() (*NetworkPool, error) {
			return vcd.client.CreateNetworkPool(&config)
		}, check,
	)
	runTestCreateNetworkPool("geneve-names", func() (*NetworkPool, error) {
		return vcd.client.CreateNetworkPoolGeneve(networkPoolName, "test network pool", manager.Name, transportZone.Name)
	}, check,
	)
	runTestCreateNetworkPool("geneve-names-no-tz-name", func() (*NetworkPool, error) {
		return vcd.client.CreateNetworkPoolGeneve(networkPoolName, "test network pool", manager.Name, "")
	}, check,
	)
}

func runTestCreateNetworkPool(label string, creationFunc func() (*NetworkPool, error), check *C) {
	fmt.Printf("[test create network pool] %s\n", label)

	networkPool, err := creationFunc()
	check.Assert(err, IsNil)
	defer func() {
		if networkPool != nil {
			_ = networkPool.Delete()
		}
	}()
	check.Assert(networkPool, NotNil)
	err = networkPool.Delete()
	check.Assert(err, IsNil)
	networkPool = nil
}
