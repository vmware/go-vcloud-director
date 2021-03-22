// +build api functional catalog vapp gateway network org query extnetwork task vm vdc system disk lb lbAppRule lbAppProfile lbServerPool lbServiceMonitor lbVirtualServer user nsxv affinity ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// createAndGetResourcesForVmCreation creates vAPP and two VM for the testing
func (vcd *TestVCD) createAndGetResourcesForVmCreation(ctx context.Context, check *C, vmName string) (*Vdc, *EdgeGateway, VAppTemplate, *VApp, types.NetworkConnectionSection, error) {
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("No Catalog name given for VDC tests")
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("No Catalog item given for VDC tests")
	}
	// Get org and vdc
	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(ctx, vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	edge, err := vcd.vdc.GetEdgeGatewayByName(ctx, vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	// Find catalog and catalog item
	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	catalogItem, err := catalog.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	// Skip the test if catalog item is not Photon OS
	if !isItemPhotonOs(ctx, *catalogItem) {
		check.Skip(fmt.Sprintf("Skipping test because catalog item %s is not Photon OS",
			vcd.config.VCD.Catalog.CatalogItem))
	}
	fmt.Printf("# Creating RAW vApp '%s'", vmName)
	vappTemplate, err := catalogItem.GetVAppTemplate(ctx)
	check.Assert(err, IsNil)
	// Compose Raw vApp
	err = vdc.ComposeRawVApp(ctx, vmName)
	check.Assert(err, IsNil)
	vapp, err := vdc.GetVAppByName(ctx, vmName, true)
	check.Assert(err, IsNil)
	// vApp was created - let's add it to cleanup list
	AddToCleanupList(vmName, "vapp", "", "createTestVapp")
	// Wait until vApp becomes configurable
	initialVappStatus, err := vapp.GetStatus(ctx)
	check.Assert(err, IsNil)
	if initialVappStatus != "RESOLVED" { // RESOLVED vApp is ready to accept operations
		err = vapp.BlockWhileStatus(ctx, initialVappStatus, vapp.client.MaxRetryTimeout)
		check.Assert(err, IsNil)
	}
	fmt.Printf(". Done\n")
	fmt.Printf("# Attaching VDC network '%s' to vApp '%s'", vcd.config.VCD.Network.Net1, vmName)
	// Attach VDC network to vApp so that VMs can use it
	net, err := vdc.GetOrgVdcNetworkByName(ctx, vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	task, err := vapp.AddRAWNetworkConfig(ctx, []*types.OrgVDCNetwork{net.OrgVDCNetwork})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")
	// Spawn 2 VMs with python servers in the newly created vApp
	desiredNetConfig := types.NetworkConnectionSection{}
	desiredNetConfig.PrimaryNetworkConnectionIndex = 0
	desiredNetConfig.NetworkConnection = append(desiredNetConfig.NetworkConnection,
		&types.NetworkConnection{
			IsConnected:             true,
			IPAddressAllocationMode: types.IPAllocationModePool,
			Network:                 vcd.config.VCD.Network.Net1,
			NetworkConnectionIndex:  0,
		})
	return vdc, edge, vappTemplate, vapp, desiredNetConfig, err
}

// spawnVM spawns VMs in provided vApp from template and also applies customization script to
// spawn a Python 3 HTTP server
func spawnVM(ctx context.Context, name string, memorySize int, vdc Vdc, vapp VApp, net types.NetworkConnectionSection, vAppTemplate VAppTemplate, check *C, customizationScript string, powerOn bool) (VM, error) {
	fmt.Printf("# Spawning VM '%s'", name)
	task, err := vapp.AddNewVM(ctx, name, vAppTemplate, &net, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	vm, err := vapp.GetVMByName(ctx, name, true)
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	fmt.Printf("# Applying 2 vCPU and "+strconv.Itoa(memorySize)+"MB configuration for VM '%s'", name)
	task, err = vm.ChangeCPUCount(ctx, 2)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)

	task, err = vm.ChangeMemorySize(ctx, memorySize)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion(ctx)
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	if customizationScript != "" {
		fmt.Printf("# Applying customization script for VM '%s'", name)
		task, err = vm.RunCustomizationScript(ctx, name, customizationScript)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion(ctx)
		check.Assert(err, IsNil)
		fmt.Printf(". Done\n")
	}

	if powerOn {
		fmt.Printf("# Powering on VM '%s'", name)
		task, err = vm.PowerOn(ctx)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion(ctx)
		check.Assert(err, IsNil)
		fmt.Printf(". Done\n")
	}

	return *vm, nil
}

// isItemPhotonOs checks if a catalog item is Photon OS
func isItemPhotonOs(ctx context.Context, item CatalogItem) bool {
	vappTemplate, err := item.GetVAppTemplate(ctx)
	// Unable to get template - can validate it's Photon OS
	if err != nil {
		return false
	}
	// Photon OS template has exactly 1 child
	if len(vappTemplate.VAppTemplate.Children.VM) != 1 {
		return false
	}

	// If child name is not "Photon OS" it's not Photon OS
	if vappTemplate.VAppTemplate.Children.VM[0].Name != "Photon OS" {
		return false
	}

	return true
}

// catalogItemIsPhotonOs returns true if test config  catalog item is Photon OS image
func catalogItemIsPhotonOs(ctx context.Context, vcd *TestVCD) bool {
	// Get Org, Vdc
	org, err := vcd.client.GetAdminOrgByName(ctx, vcd.config.VCD.Org)
	if err != nil {
		return false
	}
	// Find catalog and catalog item
	catalog, err := org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	if err != nil {
		return false
	}
	catalogItem, err := catalog.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	if err != nil {
		return false
	}
	if !isItemPhotonOs(ctx, *catalogItem) {
		return false
	}

	return true
}

// cacheLoadBalancer is meant to store load balancer settings before any operations so that all
// configuration can be checked after manipulation
func testCacheLoadBalancer(ctx context.Context, edge EdgeGateway, check *C) (*types.LbGeneralParamsWithXml, string) {
	beforeLb, err := edge.GetLBGeneralParams(ctx)
	check.Assert(err, IsNil)
	beforeLbXml := testGetEdgeEndpointXML(ctx, types.LbConfigPath, edge, check)
	return beforeLb, beforeLbXml
}

// testGetEdgeEndpointXML is used for additional validation that modifying edge gateway endpoint
// does not change any single field. It returns an XML string of whole configuration
func testGetEdgeEndpointXML(ctx context.Context, endpoint string, edge EdgeGateway, check *C) string {

	httpPath, err := edge.buildProxiedEdgeEndpointURL(endpoint)
	check.Assert(err, IsNil)

	resp, err := edge.client.ExecuteRequestWithCustomError(ctx, httpPath, http.MethodGet, types.AnyXMLMime,
		fmt.Sprintf("unable to get XML from endpoint %s: %%s", endpoint), nil, &types.NSXError{})
	check.Assert(err, IsNil)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check.Assert(err, IsNil)

	return string(body)
}

// testCheckLoadBalancerConfig validates if both raw XML string and load balancer struct remain
// identical after settings manipulation.
func testCheckLoadBalancerConfig(ctx context.Context, beforeLb *types.LbGeneralParamsWithXml, beforeLbXml string, edge EdgeGateway, check *C) {
	afterLb, err := edge.GetLBGeneralParams(ctx)
	check.Assert(err, IsNil)

	afterLbXml := testGetEdgeEndpointXML(ctx, types.LbConfigPath, edge, check)

	// remove `<version></version>` tag from both XML represntation and struct for deep comparison
	// because this version changes with each update and will never be the same after a few
	// operations

	reVersion := regexp.MustCompile(`<version>\w*<\/version>`)
	beforeLbXml = reVersion.ReplaceAllLiteralString(beforeLbXml, "")
	afterLbXml = reVersion.ReplaceAllLiteralString(afterLbXml, "")

	beforeLb.Version = ""
	afterLb.Version = ""

	check.Assert(beforeLb, DeepEquals, afterLb)
	check.Assert(beforeLbXml, DeepEquals, afterLbXml)
}

// isTcpPortOpen checks if remote TCP port is open or closed every 8 seconds until timeout is
// reached
func isTcpPortOpen(host, port string, timeout int) bool {
	retryTimeout := timeout
	// due to the VMs taking long time to boot it needs to be at least 5 minutes
	// may be even more in slower environments
	if timeout < 5*60 { // 5 minutes
		retryTimeout = 5 * 60 // 5 minutes
	}
	timeOutAfterInterval := time.Duration(retryTimeout) * time.Second
	timeoutAfter := time.After(timeOutAfterInterval)
	tick := time.NewTicker(time.Duration(8) * time.Second)

	for {
		select {
		case <-timeoutAfter:
			fmt.Printf(" Failed\n")
			return false
		case <-tick.C:
			timeout := time.Second * 3
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
			if err != nil {
				fmt.Printf(".")
			}
			// Connection established - the port is open
			if conn != nil {
				defer conn.Close()
				fmt.Printf(" Done\n")
				return true
			}
		}
	}

}

// moved from vapp_test.go
func createVappForTest(ctx context.Context, vcd *TestVCD, vappName string) (*VApp, error) {
	// Populate OrgVDCNetwork
	var networks []*types.OrgVDCNetwork
	net, err := vcd.vdc.GetOrgVdcNetworkByName(ctx, vcd.config.VCD.Network.Net1, false)
	if err != nil {
		return nil, fmt.Errorf("error finding network : %s", err)
	}
	networks = append(networks, net.OrgVDCNetwork)
	// Populate Catalog
	cat, err := vcd.org.GetCatalogByName(ctx, vcd.config.VCD.Catalog.Name, false)
	if err != nil || cat == nil {
		return nil, fmt.Errorf("error finding catalog : %s", err)
	}
	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(ctx, vcd.config.VCD.Catalog.CatalogItem, false)
	if err != nil {
		return nil, fmt.Errorf("error finding catalog item : %s", err)
	}
	// Get VAppTemplate
	vAppTemplate, err := catitem.GetVAppTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("error finding vapptemplate : %s", err)
	}
	// Get StorageProfileReference
	storageProfileRef, err := vcd.vdc.FindStorageProfileReference(ctx, vcd.config.VCD.StorageProfile.SP1)
	if err != nil {
		return nil, fmt.Errorf("error finding storage profile: %s", err)
	}
	// Compose VApp
	task, err := vcd.vdc.ComposeVApp(ctx, networks, vAppTemplate, storageProfileRef, vappName, "description", true)
	if err != nil {
		return nil, fmt.Errorf("error composing vapp: %s", err)
	}
	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(vappName, "vapp", "", "createTestVapp")
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return nil, fmt.Errorf("error composing vapp: %s", err)
	}
	// Get VApp
	vapp, err := vcd.vdc.GetVAppByName(ctx, vappName, true)
	if err != nil {
		return nil, fmt.Errorf("error getting vapp: %s", err)
	}

	err = vapp.BlockWhileStatus(ctx, "UNRESOLVED", vapp.client.MaxRetryTimeout)
	if err != nil {
		return nil, fmt.Errorf("error waiting for created test vApp to have working state: %s", err)
	}

	return vapp, nil
}

// Checks whether an independent disk is attached to a VM, and detaches it
// moved from disk_test.go
func (vcd *TestVCD) detachIndependentDisk(ctx context.Context, disk Disk) error {

	// See if the disk is attached to the VM
	vmRef, err := disk.AttachedVM(ctx)
	if err != nil {
		return err
	}
	// If the disk is attached to the VM, detach disk from the VM
	if vmRef != nil {

		vm, err := vcd.client.Client.GetVMByHref(ctx, vmRef.HREF)
		if err != nil {
			return err
		}

		// Detach the disk from VM
		task, err := vm.DetachDisk(ctx, &types.DiskAttachOrDetachParams{
			Disk: &types.Reference{
				HREF: disk.Disk.HREF,
			},
		})
		if err != nil {
			return err
		}
		err = task.WaitTaskCompletion(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// moved from vapp_test.go
func verifyNetworkConnectionSection(check *C, actual, desired *types.NetworkConnectionSection) {

	check.Assert(len(actual.NetworkConnection), Equals, len(desired.NetworkConnection))
	check.Assert(actual.PrimaryNetworkConnectionIndex, Equals, desired.PrimaryNetworkConnectionIndex)

	sort.SliceStable(actual.NetworkConnection, func(i, j int) bool {
		return actual.NetworkConnection[i].NetworkConnectionIndex <
			actual.NetworkConnection[j].NetworkConnectionIndex
	})

	for _, nic := range actual.NetworkConnection {
		actualNic := actual.NetworkConnection[nic.NetworkConnectionIndex]
		desiredNic := desired.NetworkConnection[nic.NetworkConnectionIndex]

		check.Assert(actualNic.MACAddress, Not(Equals), "")
		check.Assert(actualNic.NetworkAdapterType, Not(Equals), "")
		check.Assert(actualNic.IPAddressAllocationMode, Equals, desiredNic.IPAddressAllocationMode)
		check.Assert(actualNic.Network, Equals, desiredNic.Network)
		check.Assert(actualNic.NetworkConnectionIndex, Equals, desiredNic.NetworkConnectionIndex)

		if actualNic.IPAddressAllocationMode != types.IPAllocationModeNone {
			check.Assert(actualNic.IPAddress, Not(Equals), "")
		}
	}
}

// Ensure vApp is suitable for VM test
// Some VM tests may fail if vApp is not powered on, so VM tests can call this function to ensure the vApp is suitable for VM tests
// moved from vm_test.go
func (vcd *TestVCD) ensureVappIsSuitableForVMTest(ctx context.Context, vapp VApp) error {
	status, err := vapp.GetStatus(ctx)

	if err != nil {
		return err
	}

	// If vApp is not powered on (status = 4), power on vApp
	if status != types.VAppStatuses[4] {
		task, err := vapp.PowerOn(ctx)
		if err != nil {
			return err
		}
		err = task.WaitTaskCompletion(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Ensure VM is suitable for VM test
// Please call ensureVappAvailableForVMTest first to power on the vApp because this function cannot handle VM in suspension state due to lack of VM APIs (e.g. discard VM suspend API)
// Some VM tests may fail if VM is not powered on or powered off, so VM tests can call this function to ensure the VM is suitable for VM tests
// moved from vm_test.go
func (vcd *TestVCD) ensureVMIsSuitableForVMTest(ctx context.Context, vm *VM) error {
	// if the VM is not powered on (status = 4) or not powered off, wait for the VM power on
	// wait for around 1 min
	valid := false
	for i := 0; i < 6; i++ {
		status, err := vm.GetStatus(ctx)
		if err != nil {
			return err
		}

		// If the VM is powered on (status = 4)
		if status == types.VAppStatuses[4] {
			// Prevent affect Test_ChangeMemorySize
			// because TestVCD.Test_AttachedVMDisk is run before Test_ChangeMemorySize and Test_ChangeMemorySize will fail the test if the VM is powered on,
			task, err := vm.PowerOff(ctx)
			if err != nil {
				return err
			}
			err = task.WaitTaskCompletion(ctx)
			if err != nil {
				return err
			}
		}

		// If the VM is powered on (status = 4) or powered off (status = 8)
		if status == types.VAppStatuses[4] || status == types.VAppStatuses[8] {
			valid = true
		}

		// If 1st to 5th attempt is completed, sleep 10 seconds and try again
		// The last attempt will exit this for loop immediately, so no need to sleep
		if i < 5 {
			time.Sleep(time.Second * 10)
		}
	}

	if !valid {
		return errors.New("the VM is not powered on or powered off")
	}

	return nil
}

// moved from org_test.go
func doesOrgExist(ctx context.Context, check *C, vcd *TestVCD) {
	var org *AdminOrg
	for i := 0; i < 30; i++ {
		org, _ = vcd.client.GetAdminOrgByName(ctx, TestDeleteOrg)
		if org == nil {
			break
		} else {
			time.Sleep(time.Second)
		}
	}
	check.Assert(org, IsNil)
}

// Helper function that creates an external network to be used in other tests
// moved from externalnetwork_test.go
func (vcd *TestVCD) testCreateExternalNetwork(ctx context.Context, testName, networkName, dnsSuffix string) (skippingReason string, externalNetwork *types.ExternalNetwork, task Task, err error) {

	if vcd.skipAdminTests {
		return fmt.Sprintf(TestRequiresSysAdminPrivileges, testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.ExternalNetwork == "" {
		return fmt.Sprintf("%s: External network isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.VimServer == "" {
		return fmt.Sprintf("%s: Vim server isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.ExternalNetworkPortGroup == "" {
		return fmt.Sprintf("%s: Port group isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	if vcd.config.VCD.ExternalNetworkPortGroupType == "" {
		return fmt.Sprintf("%s: Port group type isn't configured. Test can't proceed", testName), externalNetwork, Task{}, nil
	}

	virtualCenters, err := QueryVirtualCenters(ctx, vcd.client, fmt.Sprintf("name==%s", vcd.config.VCD.VimServer))
	if err != nil {
		return "", externalNetwork, Task{}, err
	}
	if len(virtualCenters) == 0 {
		return fmt.Sprintf("No vSphere server found with name '%s'", vcd.config.VCD.VimServer), externalNetwork, Task{}, nil
	}
	vimServerHref := virtualCenters[0].HREF

	// Resolve port group info
	portGroups, err := QueryPortGroups(ctx, vcd.client, fmt.Sprintf("name==%s;portgroupType==%s", url.QueryEscape(vcd.config.VCD.ExternalNetworkPortGroup), vcd.config.VCD.ExternalNetworkPortGroupType))
	if err != nil {
		return "", externalNetwork, Task{}, err
	}
	if len(portGroups) == 0 {
		return fmt.Sprintf("No port group found with name '%s'", vcd.config.VCD.ExternalNetworkPortGroup), externalNetwork, Task{}, nil
	}
	if len(portGroups) > 1 {
		return fmt.Sprintf("More than one port group found with name '%s'", vcd.config.VCD.ExternalNetworkPortGroup), externalNetwork, Task{}, nil
	}

	externalNetwork = &types.ExternalNetwork{
		Name:        networkName,
		Description: "Test Create External Network",
		Configuration: &types.NetworkConfiguration{
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{&types.IPScope{
					Gateway:   "192.168.201.1",
					Netmask:   "255.255.255.0",
					DNS1:      "192.168.202.253",
					DNS2:      "192.168.202.254",
					DNSSuffix: dnsSuffix,
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "192.168.201.3",
								EndAddress:   "192.168.201.100",
							},
							&types.IPRange{
								StartAddress: "192.168.201.105",
								EndAddress:   "192.168.201.140",
							},
						},
					},
				}, &types.IPScope{
					Gateway:   "192.168.231.1",
					Netmask:   "255.255.255.0",
					DNS1:      "192.168.232.253",
					DNS2:      "192.168.232.254",
					DNSSuffix: dnsSuffix,
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "192.168.231.3",
								EndAddress:   "192.168.231.100",
							},
							&types.IPRange{
								StartAddress: "192.168.231.105",
								EndAddress:   "192.168.231.140",
							},
							&types.IPRange{
								StartAddress: "192.168.231.145",
								EndAddress:   "192.168.231.150",
							},
						},
					},
				},
				}},
			FenceMode: "isolated",
		},
		VimPortGroupRefs: &types.VimObjectRefs{
			VimObjectRef: []*types.VimObjectRef{
				&types.VimObjectRef{
					VimServerRef: &types.Reference{
						HREF: vimServerHref,
					},
					MoRef:         portGroups[0].MoRef,
					VimObjectType: vcd.config.VCD.ExternalNetworkPortGroupType,
				},
			},
		},
	}
	task, err = CreateExternalNetwork(ctx, vcd.client, externalNetwork)
	return skippingReason, externalNetwork, task, err
}

// deleteLbServerPoolIfExists is used to cleanup before creation of component. It returns error only if there was
// other error than govcd.ErrorEntityNotFound
// moved from lbserverpool_test.go
func deleteLbServerPoolIfExists(ctx context.Context, edge EdgeGateway, name string) error {
	err := edge.DeleteLbServerPoolByName(ctx, name)
	if err != nil && !ContainsNotFound(err) {
		return err
	}
	if err != nil && ContainsNotFound(err) {
		return nil
	}

	fmt.Printf("# Removed leftover LB server pool'%s'\n", name)
	return nil
}

// deleteLbServiceMonitorIfExists is used to cleanup before creation of component. It returns error only if there was
// other error than govcd.ErrorEntityNotFound
// moved from lbservicemonitor_test.go
func deleteLbServiceMonitorIfExists(ctx context.Context, edge EdgeGateway, name string) error {
	err := edge.DeleteLbServiceMonitorByName(ctx, name)
	if err != nil && !ContainsNotFound(err) {
		return err
	}
	if err != nil && ContainsNotFound(err) {
		return nil
	}

	fmt.Printf("# Removed leftover LB service monitor'%s'\n", name)
	return nil
}

// deleteLbAppProfileIfExists is used to cleanup before creation of component. It returns error only if there was
// other error than govcd.ErrorEntityNotFound
// moved from lbappprofile_test.go
func deleteLbAppProfileIfExists(ctx context.Context, edge EdgeGateway, name string) error {
	err := edge.DeleteLbAppProfileByName(ctx, name)
	if err != nil && !ContainsNotFound(err) {
		return err
	}
	if err != nil && ContainsNotFound(err) {
		return nil
	}

	fmt.Printf("# Removed leftover LB app profile '%s'\n", name)
	return nil
}

// deleteLbAppRuleIfExists is used to cleanup before creation of component. It returns error only if there was
// other error than govcd.ErrorEntityNotFound
// moved from lbapprule_test.go
func deleteLbAppRuleIfExists(ctx context.Context, edge EdgeGateway, name string) error {
	err := edge.DeleteLbAppRuleByName(ctx, name)
	if err != nil && !ContainsNotFound(err) {
		return err
	}
	if err != nil && ContainsNotFound(err) {
		return nil
	}

	fmt.Printf("# Removed leftover LB app rule '%s'\n", name)
	return nil
}

// moved from vm_test.go
func deleteVapp(ctx context.Context, vcd *TestVCD, name string) error {
	vapp, err := vcd.vdc.GetVAppByName(ctx, name, true)
	if err != nil {
		return fmt.Errorf("error getting vApp: %s", err)
	}
	task, _ := vapp.Undeploy(ctx)
	_ = task.WaitTaskCompletion(ctx)

	// Detach all Org networks during vApp removal because network removal errors if it happens
	// very quickly (as the next task) after vApp removal
	task, _ = vapp.RemoveAllNetworks(ctx)
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return fmt.Errorf("error removing networks from vApp: %s", err)
	}

	task, err = vapp.Delete(ctx)
	if err != nil {
		return fmt.Errorf("error deleting vApp: %s", err)
	}
	err = task.WaitTaskCompletion(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for vApp deletion task: %s", err)
	}
	return nil
}

// makeEmptyVapp creates a given vApp without any VM
func makeEmptyVapp(ctx context.Context, vdc *Vdc, name string) (*VApp, error) {

	err := vdc.ComposeRawVApp(ctx, name)
	if err != nil {
		return nil, err
	}
	vapp, err := vdc.GetVAppByName(ctx, name, true)
	if err != nil {
		return nil, err
	}
	initialVappStatus, err := vapp.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	if initialVappStatus != "RESOLVED" {
		err = vapp.BlockWhileStatus(ctx, initialVappStatus, vapp.client.MaxRetryTimeout)
		if err != nil {
			return nil, err
		}
	}
	return vapp, nil
}

// makeEmptyVm creates an empty VM inside a given vApp
func makeEmptyVm(ctx context.Context, vapp *VApp, name string) (*VM, error) {
	newDisk := types.DiskSettings{
		AdapterType:     "5",
		SizeMb:          int64(100),
		BusNumber:       0,
		UnitNumber:      0,
		ThinProvisioned: takeBoolPointer(true),
	}
	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      name,
			NetworkConnectionSection:  &types.NetworkConnectionSection{},
			Description:               "created by makeEmptyVm",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          takeBoolPointer(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           takeIntAddress(1),
				NumCoresPerSocket: takeIntAddress(1),
				CpuResourceMhz:    &types.CpuResourceMhz{Configured: 1},
				MemoryResourceMb:  &types.MemoryResourceMb{Configured: 512},
				MediaSection:      nil,
				DiskSection:       &types.DiskSection{DiskSettings: []*types.DiskSettings{&newDisk}},
				HardwareVersion:   &types.HardwareVersion{Value: "vmx-13"},
				VmToolsVersion:    "",
				VirtualCpuType:    "VM32",
				TimeSyncWithHost:  nil,
			},
			BootImage: nil,
		},
		AllEULAsAccepted: true,
	}

	vm, err := vapp.AddEmptyVm(ctx, requestDetails)
	if err != nil {
		return nil, err
	}

	return vm, nil
}
