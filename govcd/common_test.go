//go:build api || auth || functional || catalog || vapp || gateway || network || org || query || extnetwork || task || vm || vdc || system || disk || lb || lbAppRule || lbAppProfile || lbServerPool || lbServiceMonitor || lbVirtualServer || user || role || nsxv || nsxt || openapi || affinity || search || alb || certificate || vdcGroup || metadata || providervdc || rde || uiPlugin || vsphere || cse || slz || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/vmware/go-vcloud-director/v2/util"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

// createAndGetResourcesForVmCreation creates vAPP and two VM for the testing
func (vcd *TestVCD) createAndGetResourcesForVmCreation(check *C, vmName string) (*Vdc, *EdgeGateway, VAppTemplate, *VApp, types.NetworkConnectionSection, error) {
	if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("No Catalog name given for VDC tests")
	}

	if vcd.config.VCD.Catalog.CatalogItem == "" {
		check.Skip("No Catalog item given for VDC tests")
	}
	// Get org and vdc
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	vdc, err := org.GetVDCByName(vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)
	edge, err := vcd.vdc.GetEdgeGatewayByName(vcd.config.VCD.EdgeGateway, false)
	check.Assert(err, IsNil)
	// Find catalog and catalog item
	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	check.Assert(catalog, NotNil)
	catalogItem, err := catalog.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	check.Assert(err, IsNil)
	// Skip the test if catalog item is not Photon OS
	if !isItemPhotonOs(*catalogItem) {
		check.Skip(fmt.Sprintf("Skipping test because catalog item %s is not Photon OS",
			vcd.config.VCD.Catalog.CatalogItem))
	}
	fmt.Printf("# Creating RAW vApp '%s'", vmName)
	vappTemplate, err := catalogItem.GetVAppTemplate()
	check.Assert(err, IsNil)
	// Compose Raw vApp
	vapp, err := vdc.CreateRawVApp(vmName, "")
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)
	// vApp was created - let's add it to cleanup list
	AddToCleanupList(vmName, "vapp", "", "createTestVapp")
	// Wait until vApp becomes configurable
	initialVappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	if initialVappStatus != "RESOLVED" { // RESOLVED vApp is ready to accept operations
		err = vapp.BlockWhileStatus(initialVappStatus, vapp.client.MaxRetryTimeout)
		check.Assert(err, IsNil)
	}
	fmt.Printf(". Done\n")
	fmt.Printf("# Attaching VDC network '%s' to vApp '%s'", vcd.config.VCD.Network.Net1, vmName)
	// Attach VDC network to vApp so that VMs can use it
	net, err := vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	check.Assert(err, IsNil)
	task, err := vapp.AddRAWNetworkConfig([]*types.OrgVDCNetwork{net.OrgVDCNetwork})
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
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
func spawnVM(name string, memorySize int, vdc Vdc, vapp VApp, net types.NetworkConnectionSection, vAppTemplate VAppTemplate, check *C, customizationScript string, powerOn bool) (VM, error) {
	fmt.Printf("# Spawning VM '%s'", name)
	task, err := vapp.AddNewVM(name, vAppTemplate, &net, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	vm, err := vapp.GetVMByName(name, true)
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	fmt.Printf("# Applying 2 vCPU and "+strconv.Itoa(memorySize)+"MB configuration for VM '%s'", name)
	err = vm.ChangeCPU(2, 1)
	check.Assert(err, IsNil)

	err = vm.ChangeMemory(int64(memorySize))
	check.Assert(err, IsNil)
	fmt.Printf(". Done\n")

	if customizationScript != "" {
		fmt.Printf("# Applying customization script for VM '%s'", name)
		task, err = vm.RunCustomizationScript(name, customizationScript)
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		fmt.Printf(". Done\n")
	}

	if powerOn {
		fmt.Printf("# Powering on VM '%s'", name)
		task, err = vm.PowerOn()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
		fmt.Printf(". Done\n")
	}

	return *vm, nil
}

// isItemPhotonOs checks if a catalog item is Photon OS
func isItemPhotonOs(item CatalogItem) bool {
	vappTemplate, err := item.GetVAppTemplate()
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

// cacheLoadBalancer is meant to store load balancer settings before any operations so that all
// configuration can be checked after manipulation
func testCacheLoadBalancer(edge EdgeGateway, check *C) (*types.LbGeneralParamsWithXml, string) {
	beforeLb, err := edge.GetLBGeneralParams()
	check.Assert(err, IsNil)
	beforeLbXml := testGetEdgeEndpointXML(types.LbConfigPath, edge, check)
	return beforeLb, beforeLbXml
}

// testGetEdgeEndpointXML is used for additional validation that modifying edge gateway endpoint
// does not change any single field. It returns an XML string of whole configuration
func testGetEdgeEndpointXML(endpoint string, edge EdgeGateway, check *C) string {

	httpPath, err := edge.buildProxiedEdgeEndpointURL(endpoint)
	check.Assert(err, IsNil)

	resp, err := edge.client.ExecuteRequestWithCustomError(httpPath, http.MethodGet, types.AnyXMLMime,
		fmt.Sprintf("unable to get XML from endpoint %s: %%s", endpoint), nil, &types.NSXError{})
	check.Assert(err, IsNil)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Logger.Printf("error closing response Body [testGetEdgeEndpointXML]: %s", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	check.Assert(err, IsNil)

	return string(body)
}

// testCheckLoadBalancerConfig validates if both raw XML string and load balancer struct remain
// identical after settings manipulation.
func testCheckLoadBalancerConfig(beforeLb *types.LbGeneralParamsWithXml, beforeLbXml string, edge EdgeGateway, check *C) {
	afterLb, err := edge.GetLBGeneralParams()
	check.Assert(err, IsNil)

	afterLbXml := testGetEdgeEndpointXML(types.LbConfigPath, edge, check)

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

// deployVappForTest aims to replace createVappForTest
func deployVappForTest(vcd *TestVCD, vappName string) (*VApp, error) {
	// Populate OrgVDCNetwork
	net, err := vcd.vdc.GetOrgVdcNetworkByName(vcd.config.VCD.Network.Net1, false)
	if err != nil {
		return nil, fmt.Errorf("error finding network : %s", err)
	}

	// Populate Catalog
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	if err != nil || cat == nil {
		return nil, fmt.Errorf("error finding catalog : %s", err)
	}
	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.CatalogItem, false)
	if err != nil {
		return nil, fmt.Errorf("error finding catalog item : %s", err)
	}
	// Get VAppTemplate
	vAppTemplate, err := catitem.GetVAppTemplate()
	if err != nil {
		return nil, fmt.Errorf("error finding vapptemplate : %s", err)
	}
	// Get StorageProfileReference
	storageProfileRef, err := vcd.vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	if err != nil {
		return nil, fmt.Errorf("error finding storage profile: %s", err)
	}

	// Create empty vApp
	vapp, err := vcd.vdc.CreateRawVApp(vappName, "description")
	if err != nil {
		return nil, fmt.Errorf("error creating vapp: %s", err)
	}

	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
	AddToCleanupList(vappName, "vapp", "", "createTestVapp")

	// Create vApp networking
	vAppNetworkConfig, err := vapp.AddOrgNetwork(&VappNetworkSettings{}, net.OrgVDCNetwork, false)
	if err != nil {
		return nil, fmt.Errorf("error creating vApp network. %s", err)
	}

	// Create VM with only one NIC connected to vapp_net
	networkConnectionSection := &types.NetworkConnectionSection{
		PrimaryNetworkConnectionIndex: 0,
	}

	netConn := &types.NetworkConnection{
		Network:                 vAppNetworkConfig.NetworkConfig[0].NetworkName,
		IsConnected:             true,
		NetworkConnectionIndex:  0,
		IPAddressAllocationMode: types.IPAllocationModePool,
	}

	networkConnectionSection.NetworkConnection = append(networkConnectionSection.NetworkConnection, netConn)

	task, err := vapp.AddNewVMWithStorageProfile("test_vm", vAppTemplate, networkConnectionSection, &storageProfileRef, true)
	if err != nil {
		return nil, fmt.Errorf("error creating the VM: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error while waiting for the VM to be created %s", err)
	}

	err = vapp.BlockWhileStatus("UNRESOLVED", vapp.client.MaxRetryTimeout)
	if err != nil {
		return nil, fmt.Errorf("error waiting for created test vApp to have working state: %s", err)
	}

	return vapp, nil
}

// Checks whether an independent disk is attached to a VM, and detaches it
// moved from disk_test.go
func (vcd *TestVCD) detachIndependentDisk(disk Disk) error {

	// See if the disk is attached to the VM
	vmRef, err := disk.AttachedVM()
	if err != nil {
		return err
	}
	// If the disk is attached to the VM, detach disk from the VM
	if vmRef != nil {

		vm, err := vcd.client.Client.GetVMByHref(vmRef.HREF)
		if err != nil {
			return err
		}

		// Detach the disk from VM
		task, err := vm.DetachDisk(&types.DiskAttachOrDetachParams{
			Disk: &types.Reference{
				HREF: disk.Disk.HREF,
			},
		})
		if err != nil {
			return err
		}
		err = task.WaitTaskCompletion()
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
func (vcd *TestVCD) ensureVappIsSuitableForVMTest(vapp VApp) error {
	status, err := vapp.GetStatus()

	if err != nil {
		return err
	}

	// If vApp is not powered on (status = 4), power on vApp
	if status != types.VAppStatuses[4] {
		task, err := vapp.PowerOn()
		if err != nil {
			return err
		}
		err = task.WaitTaskCompletion()
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
func (vcd *TestVCD) ensureVMIsSuitableForVMTest(vm *VM) error {
	// if the VM is not powered on (status = 4) or not powered off, wait for the VM power on
	// wait for around 1 min
	valid := false
	for i := 0; i < 6; i++ {
		status, err := vm.GetStatus()
		if err != nil {
			return err
		}

		// If the VM is powered on (status = 4)
		if status == types.VAppStatuses[4] {
			// Prevent affect Test_ChangeMemorySize
			// because TestVCD.Test_AttachedVMDisk is run before Test_ChangeMemorySize and Test_ChangeMemorySize will fail the test if the VM is powered on,
			task, err := vm.Undeploy()
			if err != nil {
				return err
			}
			err = task.WaitTaskCompletion()
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
func doesOrgExist(check *C, vcd *TestVCD) {
	var org *AdminOrg
	for i := 0; i < 30; i++ {
		org, _ = vcd.client.GetAdminOrgByName(TestDeleteOrg)
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
func (vcd *TestVCD) testCreateExternalNetwork(testName, networkName, dnsSuffix string) (skippingReason string, externalNetwork *types.ExternalNetwork, task Task, err error) {

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

	virtualCenters, err := QueryVirtualCenters(vcd.client, fmt.Sprintf("name==%s", vcd.config.VCD.VimServer))
	if err != nil {
		return "", externalNetwork, Task{}, err
	}
	if len(virtualCenters) == 0 {
		return fmt.Sprintf("No vSphere server found with name '%s'", vcd.config.VCD.VimServer), externalNetwork, Task{}, nil
	}
	vimServerHref := virtualCenters[0].HREF

	// Resolve port group info
	portGroups, err := QueryPortGroups(vcd.client, fmt.Sprintf("name==%s;portgroupType==%s", url.QueryEscape(vcd.config.VCD.ExternalNetworkPortGroup), vcd.config.VCD.ExternalNetworkPortGroupType))
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
	task, err = CreateExternalNetwork(vcd.client, externalNetwork)
	return skippingReason, externalNetwork, task, err
}

// deleteLbServerPoolIfExists is used to cleanup before creation of component. It returns error only if there was
// other error than govcd.ErrorEntityNotFound
// moved from lbserverpool_test.go
func deleteLbServerPoolIfExists(edge EdgeGateway, name string) error {
	err := edge.DeleteLbServerPoolByName(name)
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
func deleteLbServiceMonitorIfExists(edge EdgeGateway, name string) error {
	err := edge.DeleteLbServiceMonitorByName(name)
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
func deleteLbAppProfileIfExists(edge EdgeGateway, name string) error {
	err := edge.DeleteLbAppProfileByName(name)
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
func deleteLbAppRuleIfExists(edge EdgeGateway, name string) error {
	err := edge.DeleteLbAppRuleByName(name)
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
func deleteVapp(vcd *TestVCD, name string) error {
	vapp, err := vcd.vdc.GetVAppByName(name, true)
	if err != nil {
		return fmt.Errorf("error getting vApp: %s", err)
	}
	task, _ := vapp.Undeploy()
	_ = task.WaitTaskCompletion()

	// Detach all Org networks during vApp removal because network removal errors if it happens
	// very quickly (as the next task) after vApp removal
	task, _ = vapp.RemoveAllNetworks()
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error removing networks from vApp: %s", err)
	}

	task, err = vapp.Delete()
	if err != nil {
		return fmt.Errorf("error deleting vApp: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for vApp deletion task: %s", err)
	}
	return nil
}

func deleteNsxtVapp(vcd *TestVCD, name string) error {
	vapp, err := vcd.nsxtVdc.GetVAppByName(name, true)
	if err != nil {
		return fmt.Errorf("error getting vApp: %s", err)
	}
	task, _ := vapp.Undeploy()
	_ = task.WaitTaskCompletion()

	// Detach all Org networks during vApp removal because network removal errors if it happens
	// very quickly (as the next task) after vApp removal
	task, _ = vapp.RemoveAllNetworks()
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error removing networks from vApp: %s", err)
	}

	task, err = vapp.Delete()
	if err != nil {
		return fmt.Errorf("error deleting vApp: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for vApp deletion task: %s", err)
	}
	return nil
}

// makeEmptyVapp creates a given vApp without any VM
func makeEmptyVapp(vdc *Vdc, name string, description string) (*VApp, error) {

	vapp, err := vdc.CreateRawVApp(name, description)
	if err != nil {
		return nil, err
	}
	if vapp == nil {
		return nil, fmt.Errorf("[makeEmptyVapp] unexpected nil vApp returned")
	}
	initialVappStatus, err := vapp.GetStatus()
	if err != nil {
		return nil, err
	}
	if initialVappStatus != "RESOLVED" {
		err = vapp.BlockWhileStatus(initialVappStatus, vapp.client.MaxRetryTimeout)
		if err != nil {
			return nil, err
		}
	}
	return vapp, nil
}

// makeEmptyVm creates an empty VM inside a given vApp
func makeEmptyVm(vapp *VApp, name string) (*VM, error) {
	newDisk := types.DiskSettings{
		AdapterType:     "5",
		SizeMb:          int64(100),
		BusNumber:       0,
		UnitNumber:      0,
		ThinProvisioned: addrOf(true),
	}
	requestDetails := &types.RecomposeVAppParamsForEmptyVm{
		CreateItem: &types.CreateItem{
			Name:                      name,
			NetworkConnectionSection:  &types.NetworkConnectionSection{},
			Description:               "created by makeEmptyVm",
			GuestCustomizationSection: nil,
			VmSpecSection: &types.VmSpecSection{
				Modified:          addrOf(true),
				Info:              "Virtual Machine specification",
				OsType:            "debian10Guest",
				NumCpus:           addrOf(1),
				NumCoresPerSocket: addrOf(1),
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

	vm, err := vapp.AddEmptyVm(requestDetails)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// spawnTestVdc spawns a VDC in a given adminOrgName to be used in tests
func spawnTestVdc(vcd *TestVCD, check *C, adminOrgName string) *Vdc {
	adminOrg, err := vcd.client.GetAdminOrgByName(adminOrgName)
	check.Assert(err, IsNil)

	providerVdcHref := getVdcProviderVdcHref(vcd, check)
	storageProfile, err := vcd.client.QueryProviderVdcStorageProfileByName(vcd.config.VCD.ProviderVdc.StorageProfile, providerVdcHref)
	check.Assert(err, IsNil)
	networkPoolHref := getVdcNetworkPoolHref(vcd, check)

	vdcConfiguration := &types.VdcConfiguration{
		Name:            check.TestName() + "-VDC",
		Xmlns:           types.XMLNamespaceVCloud,
		AllocationModel: "Flex",
		ComputeCapacity: []*types.ComputeCapacity{
			&types.ComputeCapacity{
				CPU: &types.CapacityWithUsage{
					Units:     "MHz",
					Allocated: 1024,
					Limit:     1024,
				},
				Memory: &types.CapacityWithUsage{
					Allocated: 1024,
					Limit:     1024,
					Units:     "MB",
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
			Enabled: addrOf(true),
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: storageProfile.HREF,
			},
		},
		},
		NetworkPoolReference: &types.Reference{
			HREF: networkPoolHref,
		},
		ProviderVdcReference: &types.Reference{
			HREF: providerVdcHref,
		},
		IsEnabled:                true,
		IsThinProvision:          true,
		UsesFastProvisioning:     true,
		IsElastic:                addrOf(true),
		IncludeMemoryOverhead:    addrOf(true),
		ResourceGuaranteedMemory: addrOf(1.00),
	}

	vdc, err := adminOrg.CreateOrgVdc(vdcConfiguration)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	AddToCleanupList(vdcConfiguration.Name, "vdc", vcd.org.Org.Name, check.TestName())

	return vdc
}

// spawnTestOrg spawns an Org to be used in tests
func spawnTestOrg(vcd *TestVCD, check *C, nameSuffix string) string {
	newOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	newOrgName := check.TestName() + "-" + nameSuffix
	task, err := CreateOrg(vcd.client, newOrgName, newOrgName, newOrgName, newOrg.AdminOrg.OrgSettings, true)
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
	AddToCleanupList(newOrgName, "org", "", check.TestName())

	return newOrgName
}

func getVdcProviderVdcHref(vcd *TestVCD, check *C) string {
	results, err := vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "providerVdc",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.Name),
	})
	check.Assert(err, IsNil)
	if len(results.Results.VMWProviderVdcRecord) == 0 {
		check.Skip(fmt.Sprintf("No Provider VDC found with name '%s'", vcd.config.VCD.ProviderVdc.Name))
	}
	providerVdcHref := results.Results.VMWProviderVdcRecord[0].HREF

	return providerVdcHref
}

func getVdcNetworkPoolHref(vcd *TestVCD, check *C) string {
	results, err := vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "networkPool",
		"filter": fmt.Sprintf("name==%s", vcd.config.VCD.ProviderVdc.NetworkPool),
	})
	check.Assert(err, IsNil)
	if len(results.Results.NetworkPoolRecord) == 0 {
		check.Skip(fmt.Sprintf("No network pool found with name '%s'", vcd.config.VCD.ProviderVdc.NetworkPool))
	}
	networkPoolHref := results.Results.NetworkPoolRecord[0].HREF

	return networkPoolHref
}

// extractIdsFromOpenApiReferences extracts []string with IDs from []types.OpenApiReference which contains ID and Names
func extractIdsFromOpenApiReferences(refs []types.OpenApiReference) []string {
	resultStrings := make([]string, len(refs))
	for index := range refs {
		resultStrings[index] = refs[index].ID
	}

	return resultStrings
}

// checkSkipWhenApiToken skips the test if the connection was established using an API token
func (vcd *TestVCD) checkSkipWhenApiToken(check *C) {
	if vcd.client.Client.UsingAccessToken {
		check.Skip("This test can't run on API token")
	}
}

func createNsxtVAppAndVm(vcd *TestVCD, check *C) (*VApp, *VM) {
	cat, err := vcd.org.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	check.Assert(err, IsNil)
	check.Assert(cat, NotNil)
	// Populate Catalog Item
	catitem, err := cat.GetCatalogItemByName(vcd.config.VCD.Catalog.NsxtCatalogItem, false)
	check.Assert(err, IsNil)
	check.Assert(catitem, NotNil)
	// Get VAppTemplate
	vapptemplate, err := catitem.GetVAppTemplate()
	check.Assert(err, IsNil)
	check.Assert(vapptemplate.VAppTemplate.Children.VM[0].HREF, NotNil)

	return createNsxtVAppAndVmFromCustomTemplate(vcd, check, &vapptemplate)
}

func createNsxtVAppAndVmFromCustomTemplate(vcd *TestVCD, check *C, vapptemplate *VAppTemplate) (*VApp, *VM) {
	vapp, err := vcd.nsxtVdc.CreateRawVApp(check.TestName(), check.TestName())
	check.Assert(err, IsNil)
	check.Assert(vapp, NotNil)
	// After a successful creation, the entity is added to the cleanup list.
	AddToCleanupList(vapp.VApp.Name, "vapp", vcd.nsxtVdc.Vdc.Name, check.TestName())

	// Check that vApp is powered-off
	vappStatus, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vappStatus, Equals, "RESOLVED")

	task, err := vapp.PowerOn()
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	vappStatus, err = vapp.GetStatus()
	check.Assert(err, IsNil)
	check.Assert(vappStatus, Equals, "POWERED_ON")

	// Once the operation is successful, we won't trigger a failure
	// until after the vApp deletion
	check.Check(vapp.VApp.Name, Equals, check.TestName())
	check.Check(vapp.VApp.Description, Equals, check.TestName())

	// Construct VM
	vmDef := &types.ReComposeVAppParams{
		Ovf:              types.XMLNamespaceOVF,
		Xsi:              types.XMLNamespaceXSI,
		Xmlns:            types.XMLNamespaceVCloud,
		AllEULAsAccepted: true,
		// Deploy:           false,
		Name: vapp.VApp.Name,
		// PowerOn: false, // Not touching power state at this phase
		SourcedItem: &types.SourcedCompositionItemParam{
			Source: &types.Reference{
				HREF: vapptemplate.VAppTemplate.Children.VM[0].HREF,
				Name: check.TestName() + "-vm-tmpl",
			},
			VMGeneralParams: &types.VMGeneralParams{
				Description: "test-vm-description",
			},
			InstantiationParams: &types.InstantiationParams{
				NetworkConnectionSection: &types.NetworkConnectionSection{},
			},
		},
	}
	vm, err := vapp.AddRawVM(vmDef)
	check.Assert(err, IsNil)
	check.Assert(vm, NotNil)
	check.Assert(vm.VM.Name, Equals, vmDef.SourcedItem.Source.Name)

	// Refresh vApp to have latest state
	err = vapp.Refresh()
	check.Assert(err, IsNil)

	return vapp, vm
}

// makeVappGroup creates multiple vApps, each with several VMs,
// as defined in `groupDefinition`.
// Returns a list of vApps
func makeVappGroup(label string, vdc *Vdc, groupDefinition map[string][]string) ([]*VApp, error) {
	var vappList []*VApp
	for vappName, vmNames := range groupDefinition {
		existingVapp, err := vdc.GetVAppByName(vappName, false)
		if err == nil {

			if existingVapp.VApp.Children == nil || len(existingVapp.VApp.Children.VM) == 0 {
				return nil, fmt.Errorf("found vApp %s but without VMs", vappName)
			}
			foundVms := 0
			for _, vmName := range vmNames {
				for _, existingVM := range existingVapp.VApp.Children.VM {
					if existingVM.Name == vmName {
						foundVms++
					}
				}
			}
			if foundVms < 2 {
				return nil, fmt.Errorf("found vApp %s but with %d VMs instead of 2 ", vappName, foundVms)
			}

			vappList = append(vappList, existingVapp)
			if testVerbose {
				fmt.Printf("Using existing vApp %s\n", vappName)
			}
			continue
		}

		if testVerbose {
			fmt.Printf("Creating vApp %s\n", vappName)
		}
		vapp, err := makeEmptyVapp(vdc, vappName, "")
		if err != nil {
			return nil, err
		}
		if os.Getenv("GOVCD_KEEP_TEST_OBJECTS") == "" {
			AddToCleanupList(vappName, "vapp", vdc.Vdc.Name, label)
		}
		for _, vmName := range vmNames {
			if testVerbose {
				fmt.Printf("\tCreating VM %s/%s\n", vappName, vmName)
			}
			_, err := makeEmptyVm(vapp, vmName)
			if err != nil {
				return nil, err
			}
		}
		vappList = append(vappList, vapp)
	}
	return vappList, nil
}
