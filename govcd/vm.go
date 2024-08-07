/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type VM struct {
	VM     *types.Vm
	client *Client
}

type VMRecord struct {
	VM     *types.QueryResultVMRecordType
	client *Client
}

func NewVM(cli *Client) *VM {
	return &VM{
		VM:     new(types.Vm),
		client: cli,
	}
}

// NewVMRecord creates an instance with reference to types.QueryResultVMRecordType
func NewVMRecord(cli *Client) *VMRecord {
	return &VMRecord{
		VM:     new(types.QueryResultVMRecordType),
		client: cli,
	}
}

func (vm *VM) GetStatus() (string, error) {
	err := vm.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing VM: %s", err)
	}
	return types.VAppStatuses[vm.VM.Status], nil
}

// IsDeployed checks if the VM is deployed or not
func (vm *VM) IsDeployed() (bool, error) {
	err := vm.Refresh()
	if err != nil {
		return false, fmt.Errorf("error refreshing VM: %s", err)
	}
	return vm.VM.Deployed, nil
}

func (vm *VM) Refresh() error {

	if vm.VM.HREF == "" {
		return fmt.Errorf("cannot refresh VM, Object is empty")
	}

	refreshUrl := vm.VM.HREF

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	vm.VM = &types.Vm{}

	// 37.1 Introduced BootOptions and Firmware parameters of a VM
	_, err := vm.client.ExecuteRequestWithApiVersion(refreshUrl, http.MethodGet, "", "error refreshing VM: %s", nil, vm.VM, vm.client.GetSpecificApiVersionOnCondition(">=37.1", "37.1"))

	// The request was successful
	return err
}

// GetVirtualHardwareSection returns the virtual hardware items attached to a VM
func (vm *VM) GetVirtualHardwareSection() (*types.VirtualHardwareSection, error) {

	virtualHardwareSection := &types.VirtualHardwareSection{}

	if vm.VM.HREF == "" {
		return nil, fmt.Errorf("cannot refresh, invalid reference url")
	}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/virtualHardwareSection/", http.MethodGet,
		types.MimeVirtualHardwareSection, "error retrieving virtual hardware: %s", nil, virtualHardwareSection)

	// The request was successful
	return virtualHardwareSection, err
}

// GetNetworkConnectionSection returns current networks attached to VM
//
// The slice of NICs is not necessarily ordered by NIC index
func (vm *VM) GetNetworkConnectionSection() (*types.NetworkConnectionSection, error) {

	networkConnectionSection := &types.NetworkConnectionSection{}

	if vm.VM.HREF == "" {
		return networkConnectionSection, fmt.Errorf("cannot retrieve network when VM HREF is unset")
	}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/networkConnectionSection/", http.MethodGet,
		types.MimeNetworkConnectionSection, "error retrieving network connection: %s", nil, networkConnectionSection)

	// The request was successful
	return networkConnectionSection, err
}

// UpdateNetworkConnectionSection applies network configuration of types.NetworkConnectionSection for the VM
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) UpdateNetworkConnectionSection(networks *types.NetworkConnectionSection) error {
	if vm.VM.HREF == "" {
		return fmt.Errorf("cannot update network connection when VM HREF is unset")
	}

	// Retrieve current network configuration so that we are not altering any other internal fields
	updateNetwork, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return fmt.Errorf("cannot read network section for update: %s", err)
	}
	updateNetwork.PrimaryNetworkConnectionIndex = networks.PrimaryNetworkConnectionIndex
	updateNetwork.NetworkConnection = networks.NetworkConnection
	updateNetwork.Ovf = types.XMLNamespaceOVF
	updateNetwork.Xmlns = types.XMLNamespaceVCloud

	task, err := vm.client.ExecuteTaskRequest(vm.VM.HREF+"/networkConnectionSection/", http.MethodPut,
		types.MimeNetworkConnectionSection, "error updating network connection: %s", updateNetwork)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task completion after network update for vm %s: %s", vm.VM.Name, err)
	}

	return nil
}

// Deprecated: use client.GetVMByHref instead
func (client *Client) FindVMByHREF(vmHREF string) (VM, error) {

	newVm := NewVM(client)

	_, err := client.ExecuteRequest(vmHREF, http.MethodGet,
		"", "error retrieving VM: %s", nil, newVm.VM)

	return *newVm, err

}

func (vm *VM) PowerOn() (Task, error) {

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/power/action/powerOn"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering on VM: %s", nil)

}

// PowerOnAndForceCustomization is a synchronous function which is equivalent to the functionality
// one has in UI. It triggers customization which may be useful in some cases (like altering NICs)
//
// The VM _must_ be un-deployed for this action to actually work.
func (vm *VM) PowerOnAndForceCustomization() error {
	// PowerOnAndForceCustomization only works if the VM was previously un-deployed
	vmIsDeployed, err := vm.IsDeployed()
	if err != nil {
		return fmt.Errorf("unable to check if VM %s is un-deployed forcing customization: %s",
			vm.VM.Name, err)
	}

	if vmIsDeployed {
		return fmt.Errorf("VM %s must be undeployed before forcing customization", vm.VM.Name)
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/action/deploy"

	powerOnAndCustomize := &types.DeployVAppParams{
		Xmlns:              types.XMLNamespaceVCloud,
		PowerOn:            true,
		ForceCustomization: true,
	}

	task, err := vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering on VM with customization: %s", powerOnAndCustomize)

	if err != nil {
		return err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting for task completion after power on with customization %s: %s", vm.VM.Name, err)
	}

	return nil
}

func (vm *VM) PowerOff() (Task, error) {

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/power/action/powerOff"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", "error powering off VM: %s", nil)
}

// ChangeCPUCount sets number of available virtual logical processors
// (i.e. CPUs x cores per socket)
// Cpu cores count is inherited from template.
// https://communities.vmware.com/thread/576209
// Deprecated: use vm.ChangeCPU instead
func (vm *VM) ChangeCPUCount(virtualCpuCount int) (Task, error) {
	return vm.ChangeCPUCountWithCore(virtualCpuCount, nil)
}

// ChangeCPUCountWithCore sets number of available virtual logical processors
// (i.e. CPUs x cores per socket) and cores per socket.
// Socket count is a result of: virtual logical processors/cores per socket
// https://communities.vmware.com/thread/576209
// Deprecated: use vm.ChangeCPU instead
func (vm *VM) ChangeCPUCountWithCore(virtualCpuCount int, coresPerSocket *int) (Task, error) {

	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	newCpu := &types.OVFItem{
		XmlnsRasd:       types.XMLNamespaceRASD,
		XmlnsVCloud:     types.XMLNamespaceVCloud,
		XmlnsXsi:        types.XMLNamespaceXSI,
		XmlnsVmw:        types.XMLNamespaceVMW,
		VCloudHREF:      vm.VM.HREF + "/virtualHardwareSection/cpu",
		VCloudType:      types.MimeRasdItem,
		AllocationUnits: "hertz * 10^6",
		Description:     "Number of Virtual CPUs",
		ElementName:     strconv.Itoa(virtualCpuCount) + " virtual CPU(s)",
		InstanceID:      4,
		Reservation:     0,
		ResourceType:    types.ResourceTypeProcessor,
		VirtualQuantity: int64(virtualCpuCount),
		CoresPerSocket:  coresPerSocket,
		Link: &types.Link{
			HREF: vm.VM.HREF + "/virtualHardwareSection/cpu",
			Rel:  "edit",
			Type: types.MimeRasdItem,
		},
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/virtualHardwareSection/cpu"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeRasdItem, "error changing CPU count: %s", newCpu)

}

func (vm *VM) updateNicParameters(networks []map[string]interface{}, networkSection *types.NetworkConnectionSection) error {
	for tfNicSlot, network := range networks {
		for loopIndex := range networkSection.NetworkConnection {
			// Change network config only if we have the same virtual slot number as in .tf config
			if tfNicSlot == networkSection.NetworkConnection[loopIndex].NetworkConnectionIndex {

				// Determine what type of address is requested for the vApp
				var ipAllocationMode string
				ipAddress := "Any"

				var ipFieldString string
				ipField, ipIsSet := network["ip"]
				if ipIsSet {
					ipFieldString = ipField.(string)
				}

				switch {
				// TODO v3.0 remove from here when deprecated `ip` and `network_name` attributes are removed
				case ipIsSet && ipFieldString == "dhcp": // Deprecated ip="dhcp" mode
					ipAllocationMode = types.IPAllocationModeDHCP
				case ipIsSet && ipFieldString == "allocated": // Deprecated ip="allocated" mode
					ipAllocationMode = types.IPAllocationModePool
				case ipIsSet && ipFieldString == "none": // Deprecated ip="none" mode
					ipAllocationMode = types.IPAllocationModeNone

				// Deprecated ip="valid_ip" mode (currently it is hit by ip_allocation_mode=MANUAL as well)
				case ipIsSet && net.ParseIP(ipFieldString) != nil:
					ipAllocationMode = types.IPAllocationModeManual
					ipAddress = ipFieldString
				case ipIsSet && ipFieldString != "": // Deprecated ip="something_invalid" we default to DHCP. This is odd but backwards compatible.
					ipAllocationMode = types.IPAllocationModeDHCP
					// TODO v3.0 remove until here when deprecated `ip` and `network_name` attributes are removed

					// Removed for Coverity warning: dead code - We can reinstate after removing above code
				//case ipIsSet && net.ParseIP(ipFieldString) != nil && (network["ip_allocation_mode"].(string) == types.IPAllocationModeManual):
				//	ipAllocationMode = types.IPAllocationModeManual
				//	ipAddress = ipFieldString
				default: // New networks functionality. IP was not set and we're defaulting to provided ip_allocation_mode (only manual requires the IP)
					ipAllocationMode = network["ip_allocation_mode"].(string)
				}

				networkSection.NetworkConnection[loopIndex].NeedsCustomization = true
				networkSection.NetworkConnection[loopIndex].IsConnected = true
				networkSection.NetworkConnection[loopIndex].IPAddress = ipAddress
				networkSection.NetworkConnection[loopIndex].IPAddressAllocationMode = ipAllocationMode

				// for IPAllocationModeNone we hardcode special network name used by vcd 'none'
				if ipAllocationMode == types.IPAllocationModeNone {
					networkSection.NetworkConnection[loopIndex].Network = types.NoneNetwork
				} else {
					if _, ok := network["network_name"]; !ok {
						return fmt.Errorf("could not identify network name")
					}
					networkSection.NetworkConnection[loopIndex].Network = network["network_name"].(string)
				}

				// If we have one NIC only then it is primary by default, otherwise we check for "is_primary" key
				if (len(networks) == 1) || (network["is_primary"] != nil && network["is_primary"].(bool)) {
					networkSection.PrimaryNetworkConnectionIndex = tfNicSlot
				}
			}
		}
	}
	return nil
}

// ChangeNetworkConfig allows to update existing VM NIC configuration.f
func (vm *VM) ChangeNetworkConfig(networks []map[string]interface{}) (Task, error) {
	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	networkSection, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return Task{}, fmt.Errorf("could not retrieve network connection for VM: %s", err)
	}

	err = vm.updateNicParameters(networks, networkSection)
	if err != nil {
		return Task{}, fmt.Errorf("failed processing NIC parameters: %s", err)
	}

	networkSection.Xmlns = types.XMLNamespaceVCloud
	networkSection.Ovf = types.XMLNamespaceOVF
	networkSection.Info = "Specifies the available VM network connections"

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/networkConnectionSection/"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeNetworkConnectionSection, "error changing network config: %s", networkSection)
}

// Deprecated: use vm.ChangeMemory instead
func (vm *VM) ChangeMemorySize(size int) (Task, error) {

	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	newMem := &types.OVFItem{
		XmlnsRasd:       types.XMLNamespaceRASD,
		XmlnsVCloud:     types.XMLNamespaceVCloud,
		XmlnsXsi:        types.XMLNamespaceXSI,
		VCloudHREF:      vm.VM.HREF + "/virtualHardwareSection/memory",
		VCloudType:      types.MimeRasdItem,
		AllocationUnits: "byte * 2^20",
		Description:     "Memory SizeMb",
		ElementName:     strconv.Itoa(size) + " MB of memory",
		InstanceID:      5,
		Reservation:     0,
		ResourceType:    types.ResourceTypeMemory,
		VirtualQuantity: int64(size),
		Weight:          0,
		Link: &types.Link{
			HREF: vm.VM.HREF + "/virtualHardwareSection/memory",
			Rel:  "edit",
			Type: types.MimeRasdItem,
		},
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/virtualHardwareSection/memory"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeRasdItem, "error changing memory size: %s", newMem)
}

func (vm *VM) RunCustomizationScript(computerName, script string) (Task, error) {
	return vm.Customize(computerName, script, false)
}

// GetGuestCustomizationStatus retrieves guest customization status.
// It can be one of "GC_PENDING", "REBOOT_PENDING", "GC_FAILED", "POST_GC_PENDING", "GC_COMPLETE"
func (vm *VM) GetGuestCustomizationStatus() (string, error) {
	guestCustomizationStatus := &types.GuestCustomizationStatusSection{}

	if vm.VM.HREF == "" {
		return "", fmt.Errorf("cannot retrieve guest customization, VM HREF is empty")
	}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/guestcustomizationstatus", http.MethodGet,
		types.MimeGuestCustomizationStatus, "error retrieving guest customization status: %s", nil, guestCustomizationStatus)

	// The request was successful
	return guestCustomizationStatus.GuestCustStatus, err
}

// BlockWhileGuestCustomizationStatus blocks until the customization status of VM exits unwantedStatus.
// It sleeps 3 seconds between iterations and times out after timeOutAfterSeconds of seconds.
//
// timeOutAfterSeconds must be more than 4 and less than 2 hours (60s*120)
func (vm *VM) BlockWhileGuestCustomizationStatus(unwantedStatus string, timeOutAfterSeconds int) error {
	if timeOutAfterSeconds < 5 || timeOutAfterSeconds > 60*120 {
		return fmt.Errorf("timeOutAfterSeconds must be in range 4<X<7200")
	}

	timeoutAfter := time.After(time.Duration(timeOutAfterSeconds) * time.Second)
	tick := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timed out waiting for VM guest customization status to exit state %s after %d seconds",
				unwantedStatus, timeOutAfterSeconds)
		case <-tick.C:
			currentStatus, err := vm.GetGuestCustomizationStatus()
			if err != nil {
				return fmt.Errorf("could not get VM customization status %s", err)
			}
			if currentStatus != unwantedStatus {
				return nil
			}
		}
	}
}

// Customize function allows to set ComputerName, apply customization script and enable or disable the changeSid option
//
// Deprecated: Use vm.SetGuestCustomizationSection()
func (vm *VM) Customize(computerName, script string, changeSid bool) (Task, error) {
	err := vm.Refresh()
	if err != nil {
		return Task{}, fmt.Errorf("error refreshing VM before running customization: %s", err)
	}

	vu := &types.GuestCustomizationSection{
		Ovf:   types.XMLNamespaceOVF,
		Xsi:   types.XMLNamespaceXSI,
		Xmlns: types.XMLNamespaceVCloud,

		HREF:                vm.VM.HREF,
		Type:                types.MimeGuestCustomizationSection,
		Info:                "Specifies Guest OS Customization Settings",
		Enabled:             addrOf(true),
		ComputerName:        computerName,
		CustomizationScript: script,
		ChangeSid:           &changeSid,
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/guestCustomizationSection/"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPut,
		types.MimeGuestCustomizationSection, "error customizing VM: %s", vu)
}

// Undeploy triggers a VM undeploy and power off action. "Power off" action in UI behaves this way.
func (vm *VM) Undeploy() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               types.XMLNamespaceVCloud,
		UndeployPowerAction: "powerOff",
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/action/undeploy"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeUndeployVappParams, "error undeploy VM: %s", vu)
}

// Shutdown triggers a VM undeploy and shutdown action. "Shut Down Guest OS" action in UI behaves
// this way.
//
// Note. Success of this operation depends on the VM having Guest Tools installed.
func (vm *VM) Shutdown() (Task, error) {

	vu := &types.UndeployVAppParams{
		Xmlns:               types.XMLNamespaceVCloud,
		UndeployPowerAction: "shutdown",
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/action/undeploy"

	// Return the task
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		types.MimeUndeployVappParams, "error undeploy VM: %s", vu)
}

// Attach or detach an independent disk
// Use the disk/action/attach or disk/action/detach links in a VM to attach or detach an independent disk.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 164 - 165,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vm *VM) attachOrDetachDisk(diskParams *types.DiskAttachOrDetachParams, rel string) (Task, error) {
	util.Logger.Printf("[TRACE] Attach or detach disk, href: %s, rel: %s \n", diskParams.Disk.HREF, rel)

	var attachOrDetachDiskLink *types.Link
	for _, link := range vm.VM.Link {
		if link.Rel == rel && link.Type == types.MimeDiskAttachOrDetachParams {
			util.Logger.Printf("[TRACE] Attach or detach disk - found the proper link for request, HREF: %s, name: %s, type: %s, id: %s, rel: %s \n",
				link.HREF,
				link.Name,
				link.Type,
				link.ID,
				link.Rel)
			attachOrDetachDiskLink = link
		}
	}

	if attachOrDetachDiskLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for attach or detach disk in disk Link")
	}

	diskParams.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return vm.client.ExecuteTaskRequest(attachOrDetachDiskLink.HREF, http.MethodPost,
		attachOrDetachDiskLink.Type, "error attach or detach disk: %s", diskParams)
}

// AttachDisk attaches an independent disk
// Call attachOrDetachDisk with disk and types.RelDiskAttach to attach an independent disk.
// Please verify the independent disk is not connected to any VM before calling this function.
// If the independent disk is connected to a VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 164 - 165,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vm *VM) AttachDisk(diskParams *types.DiskAttachOrDetachParams) (Task, error) {
	if diskParams == nil || diskParams.Disk == nil || diskParams.Disk.HREF == "" {
		return Task{}, fmt.Errorf("could not find disk info for attach")
	}
	util.Logger.Printf("[TRACE] Attach disk, HREF: %s\n", diskParams.Disk.HREF)

	return vm.attachOrDetachDisk(diskParams, types.RelDiskAttach)
}

// DetachDisk detaches an independent disk
// Call attachOrDetachDisk with disk and types.RelDiskDetach to detach an independent disk.
// Please verify the independent disk is connected the VM before calling this function.
// If the independent disk is not connected to the VM, the task will be failed.
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0 PDF Page 164 - 165,
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// 241956dd-e128-4fcc-8131-bf66e1edd895/vcloud_sp_api_guide_30_0.pdf
func (vm *VM) DetachDisk(diskParams *types.DiskAttachOrDetachParams) (Task, error) {

	if diskParams == nil || diskParams.Disk == nil || diskParams.Disk.HREF == "" {
		return Task{}, fmt.Errorf("could not find disk info for detach")
	}
	util.Logger.Printf("[TRACE] Detach disk, HREF: %s\n", diskParams.Disk.HREF)

	return vm.attachOrDetachDisk(diskParams, types.RelDiskDetach)
}

// HandleInsertMedia helper function finds media and calls InsertMedia
func (vm *VM) HandleInsertMedia(org *Org, catalogName, mediaName string) (Task, error) {

	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return Task{}, err
	}

	media, err := catalog.GetMediaByName(mediaName, false)
	if err != nil {
		return Task{}, err
	}

	return vm.InsertMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
			Name: media.Media.Name,
			ID:   media.Media.ID,
			Type: media.Media.Type,
		},
	})
}

// HandleEjectMediaAndAnswer helper function which finds media, calls EjectMedia, waits for task to complete and answer question.
// Also waits until VM status refreshes - this added as 9.7-10.0 vCD versions has lag in status update.
// answerYes - handles question risen when VM is running. True value enforces ejection.
func (vm *VM) HandleEjectMediaAndAnswer(org *Org, catalogName, mediaName string, answerYes bool) (*VM, error) {
	task, err := vm.HandleEjectMedia(org, catalogName, mediaName)
	if err != nil {
		return nil, fmt.Errorf("error: %s", err)
	}

	err = task.WaitTaskCompletion(answerYes)
	if err != nil {
		return nil, fmt.Errorf("error: %s", err)
	}

	for i := 0; i < 10; i++ {
		err = vm.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
		if !isMediaInjected(vm.VM.VirtualHardwareSection.Item) {
			return vm, nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil, fmt.Errorf("eject media executed but waiting for state update failed")
}

// check resource subtype for specific value which means media is injected
func isMediaInjected(items []*types.VirtualHardwareItem) bool {
	for _, hardwareItem := range items {
		if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
			return true
		}
	}
	return false
}

// HandleEjectMedia is a helper function which finds media and calls EjectMedia
func (vm *VM) HandleEjectMedia(org *Org, catalogName, mediaName string) (EjectTask, error) {
	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return EjectTask{}, err
	}

	media, err := catalog.GetMediaByName(mediaName, false)
	if err != nil {
		return EjectTask{}, err
	}

	task, err := vm.EjectMedia(&types.MediaInsertOrEjectParams{
		Media: &types.Reference{
			HREF: media.Media.HREF,
		},
	})

	return task, err
}

// InsertMedia insert media for a VM
// Call insertOrEjectMedia with media and types.RelMediaInsertMedia to insert media from VM.
func (vm *VM) InsertMedia(mediaParams *types.MediaInsertOrEjectParams) (Task, error) {
	util.Logger.Printf("[TRACE] Insert media, HREF: %s\n", mediaParams.Media.HREF)

	err := validateMediaParams(mediaParams)
	if err != nil {
		return Task{}, err
	}

	return vm.insertOrEjectMedia(mediaParams, types.RelMediaInsertMedia)
}

// EjectMedia ejects media from VM
// Call insertOrEjectMedia with media and types.RelMediaEjectMedia to eject media from VM.
// If media isn't inserted then task still will be successful.
func (vm *VM) EjectMedia(mediaParams *types.MediaInsertOrEjectParams) (EjectTask, error) {
	util.Logger.Printf("[TRACE] Detach disk, HREF: %s\n", mediaParams.Media.HREF)

	err := validateMediaParams(mediaParams)
	if err != nil {
		return EjectTask{}, err
	}

	task, err := vm.insertOrEjectMedia(mediaParams, types.RelMediaEjectMedia)
	if err != nil {
		return EjectTask{}, err
	}

	return *NewEjectTask(&task, vm), nil
}

// validates that media and media.href isn't empty
func validateMediaParams(mediaParams *types.MediaInsertOrEjectParams) error {
	if mediaParams.Media == nil {
		return fmt.Errorf("could not find media info for eject")
	}
	if mediaParams.Media.HREF == "" {
		return fmt.Errorf("could not find media HREF which is required for insert")
	}
	return nil
}

// Insert or eject a media for VM
// Use the vm/action/insert or vm/action/eject links in a VM to insert or eject media.
// Reference:
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/POST-InsertCdRom.html
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/POST-EjectCdRom.html
func (vm *VM) insertOrEjectMedia(mediaParams *types.MediaInsertOrEjectParams, linkRel string) (Task, error) {
	util.Logger.Printf("[TRACE] Insert or eject media, href: %s, name: %s, , linkRel: %s \n", mediaParams.Media.HREF, mediaParams.Media.Name, linkRel)

	var insertOrEjectMediaLink *types.Link
	for _, link := range vm.VM.Link {
		if link.Rel == linkRel && link.Type == types.MimeMediaInsertOrEjectParams {
			util.Logger.Printf("[TRACE] Insert or eject media - found the proper link for request, HREF: %s, "+
				"name: %s, type: %s, id: %s, rel: %s \n", link.HREF, link.Name, link.Type, link.ID, link.Rel)
			insertOrEjectMediaLink = link
		}
	}

	if insertOrEjectMediaLink == nil {
		return Task{}, fmt.Errorf("could not find request URL for insert or eject media")
	}

	mediaParams.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return vm.client.ExecuteTaskRequest(insertOrEjectMediaLink.HREF, http.MethodPost,
		insertOrEjectMediaLink.Type, "error insert or eject media: %s", mediaParams)
}

// GetQuestion uses the get existing VM question for operation which need additional response
// Reference:
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/GET-VmPendingQuestion.html
func (vm *VM) GetQuestion() (types.VmPendingQuestion, error) {

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/question"

	req := vm.client.NewRequest(map[string]string{}, http.MethodGet, *apiEndpoint, nil)

	resp, err := vm.client.Http.Do(req)
	if err != nil {
		return types.VmPendingQuestion{}, fmt.Errorf("error getting VM question: %s", err)
	}

	// vCD security feature - on no question return 403 access error
	if http.StatusForbidden == resp.StatusCode {
		util.Logger.Printf("No question found for VM: %s\n", vm.VM.ID)
		return types.VmPendingQuestion{}, nil
	}

	if http.StatusOK != resp.StatusCode {
		return types.VmPendingQuestion{}, fmt.Errorf("error getting question: %s", ParseErr(types.BodyTypeXML, resp, &types.Error{}))
	}

	question := &types.VmPendingQuestion{}

	if err = decodeBody(types.BodyTypeXML, resp, question); err != nil {
		return types.VmPendingQuestion{}, fmt.Errorf("error decoding question response: %s", err)
	}

	// The request was successful
	return *question, nil

}

// AnswerQuestion uses the provided answer to existing VM question for operation which need additional response
// Reference:
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/POST-AnswerVmPendingQuestion.html
func (vm *VM) AnswerQuestion(questionId string, choiceId int) error {

	//validate input
	if questionId == "" {
		return fmt.Errorf("questionId can not be empty")
	}

	answer := &types.VmQuestionAnswer{
		Xmlns:      types.XMLNamespaceVCloud,
		QuestionId: questionId,
		ChoiceId:   choiceId,
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	apiEndpoint.Path += "/question/action/answer"

	return vm.client.ExecuteRequestWithoutResponse(apiEndpoint.String(), http.MethodPost,
		"", "error asnwering question: %s", answer)
}

// ToggleHardwareVirtualization allows to either enable or disable hardware assisted
// CPU virtualization for the VM. It can only be performed on a powered off VM and
// will return an error otherwise. This is mainly useful for hypervisor nesting.
func (vm *VM) ToggleHardwareVirtualization(isEnabled bool) (Task, error) {
	vmStatus, err := vm.GetStatus()
	if err != nil {
		return Task{}, fmt.Errorf("unable to toggle hardware virtualization: %s", err)
	}
	if vmStatus != "POWERED_OFF" && vmStatus != "PARTIALLY_POWERED_OFF" {
		return Task{}, fmt.Errorf("hardware virtualization can be changed from powered off state, status: %s", vmStatus)
	}

	apiEndpoint := urlParseRequestURI(vm.VM.HREF)
	if isEnabled {
		apiEndpoint.Path += "/action/enableNestedHypervisor"
	} else {
		apiEndpoint.Path += "/action/disableNestedHypervisor"
	}
	errMessage := fmt.Sprintf("error toggling hypervisor nesting feature to %t for VM: %%s", isEnabled)
	return vm.client.ExecuteTaskRequest(apiEndpoint.String(), http.MethodPost,
		"", errMessage, nil)
}

// SetProductSectionList sets product section for a VM. It allows to change VM guest properties.
//
// The slice of properties "ProductSectionList.ProductSection.Property" is not necessarily ordered
// or returned as set before
func (vm *VM) SetProductSectionList(productSection *types.ProductSectionList) (*types.ProductSectionList, error) {
	err := setProductSectionList(vm.client, vm.VM.HREF, productSection)
	if err != nil {
		return nil, fmt.Errorf("unable to set VM product section: %s", err)
	}

	return vm.GetProductSectionList()
}

// GetProductSectionList retrieves product section for a VM. It allows to read VM guest properties.
//
// The slice of properties "ProductSectionList.ProductSection.Property" is not necessarily ordered
// or returned as set before
func (vm *VM) GetProductSectionList() (*types.ProductSectionList, error) {
	return getProductSectionList(vm.client, vm.VM.HREF)
}

// GetEnvironment returns the OVF Environment. It's only available for poweredOn VM
func (vm *VM) GetEnvironment() (*types.OvfEnvironment, error) {
	vmStatus, err := vm.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("unable to get OVF environment: %s", err)
	}

	if vmStatus != "POWERED_ON" {
		return nil, fmt.Errorf("OVF environment is only available when VM is powered on")
	}

	return vm.VM.Environment, nil
}

// GetGuestCustomizationSection retrieves guest customization section for a VM. It allows to read VM guest customization properties.
func (vm *VM) GetGuestCustomizationSection() (*types.GuestCustomizationSection, error) {
	if vm == nil || vm.VM.HREF == "" {
		return nil, fmt.Errorf("vm or href cannot be empty to get  guest customization section")
	}
	guestCustomizationSection := &types.GuestCustomizationSection{}

	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/guestCustomizationSection", http.MethodGet,
		types.MimeGuestCustomizationSection, "error retrieving guest customization section : %s", nil, guestCustomizationSection)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve guest customization section: %s", err)
	}

	return guestCustomizationSection, nil
}

// SetGuestCustomizationSection sets guest customization section for a VM. It allows to change VM guest customization properties.
func (vm *VM) SetGuestCustomizationSection(guestCustomizationSection *types.GuestCustomizationSection) (*types.GuestCustomizationSection, error) {
	if vm == nil || vm.VM.HREF == "" {
		return nil, fmt.Errorf("vm or href cannot be empty to get  guest customization section")
	}

	guestCustomizationSection.Xmlns = types.XMLNamespaceVCloud
	guestCustomizationSection.Ovf = types.XMLNamespaceOVF

	task, err := vm.client.ExecuteTaskRequest(vm.VM.HREF+"/guestCustomizationSection", http.MethodPut,
		types.MimeGuestCustomizationSection, "error setting product section: %s", guestCustomizationSection)

	if err != nil {
		return nil, fmt.Errorf("unable to set guest customization section: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("task for setting guest customization section failed: %s", err)
	}

	return vm.GetGuestCustomizationSection()
}

// UpgradeHardwareVersion triggers hardware upgrade and tracks the task to completion or returns an
// error
func (vm *VM) UpgradeHardwareVersion() error {
	if vm == nil || vm.VM.HREF == "" {
		return fmt.Errorf("vm or href cannot be empty to trigger Hardware Upgrade")
	}

	task, err := vm.client.ExecuteTaskRequest(vm.VM.HREF+"/action/upgradeHardwareVersion", http.MethodPost,
		types.AnyXMLMime, "error setting product section: %s", nil)

	if err != nil {
		return fmt.Errorf("unable to trigger VM Hardware Upgrade: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("task for VM Hardware Upgrade failed: %s", err)
	}

	return nil
}

// GetParentVApp find parent vApp for VM by checking its "up" "link".
//
// Note. The VM has a parent vApp defined even if it was created as a standalone
func (vm *VM) GetParentVApp() (*VApp, error) {
	if vm == nil || vm.VM == nil {
		return nil, fmt.Errorf("vm object cannot be nil to get parent vApp")
	}

	for _, link := range vm.VM.Link {
		if link.Type == types.MimeVApp && link.Rel == "up" {
			vapp := NewVApp(vm.client)
			vapp.VApp.HREF = link.HREF

			err := vapp.Refresh()

			if err != nil {
				return nil, fmt.Errorf("could not refresh parent vApp for VM %s: %s", vm.VM.Name, err)
			}

			return vapp, nil
		}
	}

	return nil, fmt.Errorf("could not find parent vApp link")
}

// GetParentVdc returns parent VDC for VM
func (vm *VM) GetParentVdc() (*Vdc, error) {
	if vm == nil || vm.VM == nil {
		return nil, fmt.Errorf("vm object cannot be nil to get parent vApp")
	}

	vapp, err := vm.GetParentVApp()
	if err != nil {
		return nil, fmt.Errorf("could not find parent vApp for VM %s: %s", vm.VM.Name, err)
	}

	vdc, err := vapp.GetParentVDC()
	if err != nil {
		return nil, fmt.Errorf("could not find parent vApp for VM %s: %s", vm.VM.Name, err)
	}

	return &vdc, nil
}

// getEdgeGatewaysForRoutedNics checks if any NICs are using routed networks and are attached to
// edge gateway
func (vm *VM) getEdgeGatewaysForRoutedNics(nicDhcpConfigs []nicDhcpConfig) ([]nicDhcpConfig, error) {
	// Lookup parent vDC for VM
	vdc, err := vm.GetParentVdc()
	if err != nil {
		return nil, fmt.Errorf("could not find parent vDC for VM %s: %s", vm.VM.Name, err)
	}

	for index, nic := range nicDhcpConfigs {
		edgeGatewayName, err := vm.getEdgeGatewayNameForNic(nic.vmNicIndex)
		if err != nil && !IsNotFound(err) {
			return nil, fmt.Errorf("could not validate if NIC %d uses routed network attached to edge gateway: %s",
				nic.vmNicIndex, err)
		}

		// This nicIndex is not attached to routed network, move further
		if IsNotFound(err) {
			util.Logger.Printf("[TRACE] [DHCP IP Lookup] VM '%s' NIC with index %d is not attached to edge gateway routed network\n",
				vm.VM.Name, nic.vmNicIndex)
		} else {
			// Lookup edge gateway
			edgeGateway, err := vdc.GetEdgeGatewayByName(edgeGatewayName, false)
			if ContainsNotFound(err) {
				util.Logger.Printf("[TRACE] [DHCP IP Lookup] edge gateway not found: %s. Ignoring.", edgeGatewayName)
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("could not lookup edge gateway for routed network on NIC %d: %s",
					nic.vmNicIndex, err)
			}

			util.Logger.Printf("[TRACE] [DHCP IP Lookup] VM '%s' NIC with index %d is attached to edge gateway routed network\n",
				vm.VM.Name, nic.vmNicIndex)
			nicDhcpConfigs[index].routedNetworkEdgeGateway = edgeGateway
		}
	}

	return nicDhcpConfigs, nil
}

// nicDhcpConfig is used to group data for carrying between multiple functions and optimizing on API
// calls
type nicDhcpConfig struct {
	vmNicIndex               int
	ip                       string
	mac                      string
	routedNetworkEdgeGateway *EdgeGateway
}

// nicDhcpConfigs is a slice of nicDhcpConfig
type nicDhcpConfigs []nicDhcpConfig

// getIpsFromNicDhcpConfigs extracts just IP addresses from nicDhcpConfigs
func getIpsFromNicDhcpConfigs(nicConfigs []nicDhcpConfig) []string {
	result := make([]string, len(nicConfigs))
	for index, nicConfig := range nicConfigs {
		result[index] = nicConfig.ip
	}
	return result
}

// allNicsHaveIps checks if all nicDhcpConfig in slice have not empty IP field
func allNicsHaveIps(nicConfigs []nicDhcpConfig) bool {
	allNicsHaveIps := true
	for _, nicConfig := range nicConfigs {
		if nicConfig.ip == "" {
			allNicsHaveIps = false
		}
	}
	return allNicsHaveIps
}

// WaitForDhcpIpByNicIndexes accepts a slice of NIC indexes in VM, tries to get these IPs up to
// maxWaitSeconds and then returns:
// * a list of IPs
// * whether the function hit timeout (some IP values may be available after timeout)
// * error
//
// This function checks a slice of nicIndexes and reuses all possible API calls. It may return a
// partial result for IP addresses when the timeout is hit.
//
// Getting a DHCP address is complicated because vCD (in UI and in types.NetworkConnectionSection)
// reports IP addresses only when guest tools are present on a VM. This function also attempts to
// check if VM NICs are attached to routed network on edge gateway - then there is a chance that
// built-in DHCP pools are used and active DHCP leases can be found.
//
// For this function to work - at least one the following must be true:
// * VM has guest tools (vCD UI shows IP address). (Takes longer time)
// * VM DHCP interface is connected to routed Org network and is using NSX-V Edge Gateway DHCP. (Takes
// less time, but is more constrained)
func (vm *VM) WaitForDhcpIpByNicIndexes(nicIndexes []int, maxWaitSeconds int, useNsxvDhcpLeaseCheck bool) ([]string, bool, error) {
	util.Logger.Printf("[TRACE] [DHCP IP Lookup] VM '%s' attempting to lookup IP addresses for DHCP NICs %v\n",
		vm.VM.Name, nicIndexes)
	// validate NIC indexes
	if len(nicIndexes) == 0 {
		return []string{}, false, fmt.Errorf("at least one NIC index must be specified")
	}
	for index, nicIndex := range nicIndexes {
		if nicIndex < 0 {
			return []string{}, false, fmt.Errorf("NIC index %d cannot be negative", index)
		}
	}

	// inject NIC indexes into structure []nicDhcpConfig as this allows to save API calls when
	// querying for multiple NICs
	// This slice is ordered the same as original slice of nicIndexes
	nicStates := make(nicDhcpConfigs, len(nicIndexes))
	for index, nicIndex := range nicIndexes {
		nicStates[index].vmNicIndex = nicIndex
	}
	var err error
	if useNsxvDhcpLeaseCheck { // Edge gateways have to be looked up when DHCP lease checks are enabled
		// Lookup edge gateways for routed networks and store them
		nicStates, err = vm.getEdgeGatewaysForRoutedNics(nicStates)
		if err != nil {
			return []string{}, false, fmt.Errorf("unable to validate if NICs are attached to edge gateway: %s", err)
		}
	}

	// Run a timer to wait for IPs being present until maxWaitSeconds
	timeoutAfter := time.After(time.Duration(maxWaitSeconds) * time.Second)
	tick := time.NewTicker(3 * time.Second)
	for {
		select {
		// If timeout occured - return as much as was found
		case <-timeoutAfter:
			ipSlice := getIpsFromNicDhcpConfigs(nicStates)
			util.Logger.Printf("[DEBUG] [DHCP IP Lookup] VM '%s' NICs with indexes %v did not all report IP "+
				"addresses after %d seconds. Indexes: %v ,IPs: '%s'\n", vm.VM.Name, nicIndexes,
				maxWaitSeconds, nicIndexes, strings.Join(ipSlice, ", "))
			return ipSlice, true, nil
		case <-tick.C:
			// Step 1 check if VMware tools reported IPs in NetworkConnectionSection (HTML5 UI reads it to show IPs as well).
			// Also populate MAC addresses into nicStates structure for later usage.
			nicStates, err = vm.getIpsMacsByNicIndexes(nicStates)
			if err != nil {
				return []string{}, false, fmt.Errorf("could not check IP addresses assigned to VM %s: %s",
					vm.VM.Name, err)
			}

			// All IP addresses found - return
			if allNicsHaveIps(nicStates) {
				util.Logger.Printf("[TRACE] [DHCP IP Lookup] VM '%s' NICs with indexes %v all reported their IPs using guest tools\n",
					vm.VM.Name, nicIndexes)
				return getIpsFromNicDhcpConfigs(nicStates), false, nil
			}

			util.Logger.Printf("[DEBUG] [DHCP IP Lookup] VM '%s' NICs with indexes %v did not all report their IPs using guest tools\n",
				vm.VM.Name, nicIndexes)

			// Step 2 If enabled - check if DHCP leases in edge gateways can hint IP addresses
			if useNsxvDhcpLeaseCheck {
				nicStates, err = vm.getIpsByDhcpLeaseMacs(nicStates)
				if err != nil {
					return []string{}, false, fmt.Errorf("could not check MAC leases for VM '%s': %s",
						vm.VM.Name, err)
				}

				// All IP addresses found - return
				if allNicsHaveIps(nicStates) {
					util.Logger.Printf("[DEBUG] [DHCP IP Lookup] VM '%s' NICs with indexes %v all reported their IPs after lease check\n",
						vm.VM.Name, nicIndexes)
					return getIpsFromNicDhcpConfigs(nicStates), false, nil
				}
				util.Logger.Printf("[DEBUG] [DHCP IP Lookup] VM '%s' NICs with indexes %v did not all report their IPs using DHCP leases\n",
					vm.VM.Name, nicIndexes)
			}
		}
	}
}

// getEdgeGatewayNameForNic checks if a network card with specified nicIndex uses routed network and
// is attached to particular edge gateway. Edge gateway name is returned if so.
func (vm *VM) getEdgeGatewayNameForNic(nicIndex int) (string, error) {
	if nicIndex < 0 {
		return "", fmt.Errorf("NIC index cannot be negative")
	}

	networkConnnectionSection, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return "", fmt.Errorf("could not get IP address for NIC %d: %s", nicIndex, err)
	}

	// Find NIC
	var networkConnection *types.NetworkConnection
	for _, nic := range networkConnnectionSection.NetworkConnection {
		if nic.NetworkConnectionIndex == nicIndex {
			networkConnection = nic
		}
	}

	if networkConnection == nil {
		return "", fmt.Errorf("could not find NIC with index %d", nicIndex)
	}

	// Validate if the VM is attached to routed org vdc network
	vdc, err := vm.GetParentVdc()
	if err != nil {
		return "", fmt.Errorf("could not find parent vDC for VM %s: %s", vm.VM.Name, err)
	}

	edgeGatewayName, err := vdc.FindEdgeGatewayNameByNetwork(networkConnection.Network)
	if err != nil && !IsNotFound(err) {
		return "", fmt.Errorf("could not find edge gateway name for network %s: %s",
			networkConnection.Network, err)
	}

	if edgeGatewayName == "" {
		return "", ErrorEntityNotFound
	}

	return edgeGatewayName, nil
}

// getIpsByDhcpLeaseMacs accepts a slice of nicDhcpConfig and tries to find an active DHCP lease for
// all defined MAC addresses
func (vm *VM) getIpsByDhcpLeaseMacs(nicConfigs []nicDhcpConfig) ([]nicDhcpConfig, error) {
	dhcpLeaseCache := make(map[string][]*types.EdgeDhcpLeaseInfo)

	var err error

	for index, nicConfig := range nicConfigs {
		// If the NIC does not have Edge Gateway defined - skip it
		if nicConfig.routedNetworkEdgeGateway == nil {
			util.Logger.Printf("[DEBUG] VM '%s' skipping check of DHCP lease for NIC index %d "+
				"because it is not attached to edge gateway\n", vm.VM.Name, index)
			continue
		}

		egw := nicConfig.routedNetworkEdgeGateway

		if dhcpLeaseCache[egw.EdgeGateway.Name] == nil {
			dhcpLeaseCache[egw.EdgeGateway.Name], err = egw.GetAllNsxvDhcpLeases()
			if err != nil && !IsNotFound(err) {
				return nicConfigs, fmt.Errorf("unable to get DHCP leases for edge gateway %s: %s",
					egw.EdgeGateway.Name, err)
			}
		}

		for _, lease := range dhcpLeaseCache[egw.EdgeGateway.Name] {
			util.Logger.Printf("[DEBUG] Checking DHCP lease: %#+v", lease)
			if lease.BindingState == "active" && lease.MacAddress == nicConfig.mac {
				nicConfigs[index].ip = lease.IpAddress
			}
		}

	}

	return nicConfigs, nil
}

// getIpsMacsByNicIndexes searches for NICs attached to VM by vmNicIndex and populated
// []nicDhcpConfig with IPs and MAC addresses
func (vm *VM) getIpsMacsByNicIndexes(nicConfigs []nicDhcpConfig) ([]nicDhcpConfig, error) {
	networkConnnectionSection, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return nil, fmt.Errorf("could not get IP configuration for VM %s : %s", vm.VM.Name, err)
	}

	// Find NICs and populate their IPs and MACs
	for sliceIndex, nicConfig := range nicConfigs {
		for _, nic := range networkConnnectionSection.NetworkConnection {
			if nic.NetworkConnectionIndex == nicConfig.vmNicIndex {
				nicConfigs[sliceIndex].ip = nic.IPAddress
				nicConfigs[sliceIndex].mac = nic.MACAddress
			}
		}
	}

	return nicConfigs, nil
}

// AddInternalDisk creates disk type *types.DiskSettings to the VM.
// Returns new disk ID and error.
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) AddInternalDisk(diskData *types.DiskSettings) (string, error) {
	err := vm.Refresh()
	if err != nil {
		return "", fmt.Errorf("error refreshing VM: %s", err)
	}

	err = vm.validateInternalDiskInput(diskData, vm.VM.Name, vm.VM.ID)
	if err != nil {
		return "", err
	}

	var diskSettings []*types.DiskSettings
	if vm.VM.VmSpecSection != nil && vm.VM.VmSpecSection.DiskSection != nil && vm.VM.VmSpecSection.DiskSection.DiskSettings != nil {
		diskSettings = vm.VM.VmSpecSection.DiskSection.DiskSettings
	}

	diskSettings = append(diskSettings, diskData)
	vmSpecSection := vm.VM.VmSpecSection
	vmSpecSection.DiskSection.DiskSettings = diskSettings

	vmSpecSection, err = vm.UpdateInternalDisks(vmSpecSection)
	if err != nil {
		return "", err
	}

	for _, diskSetting := range vmSpecSection.DiskSection.DiskSettings {
		if diskSetting.AdapterType == diskData.AdapterType &&
			diskSetting.BusNumber == diskData.BusNumber &&
			diskSetting.UnitNumber == diskData.UnitNumber {
			return diskSetting.DiskId, nil
		}
	}

	return "", fmt.Errorf("created disk wasn't in list of returned VM internal disks")
}

func (vm *VM) validateInternalDiskInput(diskData *types.DiskSettings, vmName, vmId string) error {
	if diskData.AdapterType == "" {
		return fmt.Errorf("[VM %s Id %s] disk settings missing required field: adapter type", vmName, vmId)
	}

	if diskData.BusNumber < 0 {
		return fmt.Errorf("[VM %s Id %s] disk settings bus number has to be 0 or higher", vmName, vmId)
	}

	if diskData.UnitNumber < 0 {
		return fmt.Errorf("[VM %s Id %s] disk settings unit number has to be 0 or higher", vmName, vmId)
	}

	if diskData.SizeMb < int64(0) {
		return fmt.Errorf("[VM %s Id %s] disk settings size MB has to be 0 or higher", vmName, vmId)
	}

	if diskData.IopsAllocation != nil && diskData.IopsAllocation.Reservation < int64(0) {
		return fmt.Errorf("[VM %s Id %s] disk settings iops reservation has to be 0 or higher", vmName, vmId)
	}

	if diskData.ThinProvisioned == nil {
		return fmt.Errorf("[VM %s Id %s] disk settings missing required field: thin provisioned", vmName, vmId)
	}

	if diskData.StorageProfile == nil {
		return fmt.Errorf("[VM %s Id %s]disk settings missing required field: storage profile", vmName, vmId)
	}

	return nil
}

// GetInternalDiskById returns a *types.DiskSettings if one exists.
// If it doesn't, returns nil and ErrorEntityNotFound or other err.
func (vm *VM) GetInternalDiskById(diskId string, refresh bool) (*types.DiskSettings, error) {
	if diskId == "" {
		return nil, fmt.Errorf("cannot get internal disk - provided disk Id is empty")
	}

	if refresh {
		err := vm.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing VM: %s", err)
		}
	}

	if vm.VM.VmSpecSection.DiskSection == nil || vm.VM.VmSpecSection.DiskSection.DiskSettings == nil ||
		len(vm.VM.VmSpecSection.DiskSection.DiskSettings) == 0 {
		return nil, fmt.Errorf("cannot get internal disk - VM doesn't have internal disks")
	}

	for _, diskSetting := range vm.VM.VmSpecSection.DiskSection.DiskSettings {
		if diskSetting.DiskId == diskId {
			return diskSetting, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// DeleteInternalDisk delete disk using provided disk ID.
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) DeleteInternalDisk(diskId string) error {
	err := vm.Refresh()
	if err != nil {
		return fmt.Errorf("error refreshing VM: %s", err)
	}

	diskSettings := vm.VM.VmSpecSection.DiskSection.DiskSettings
	if diskSettings == nil {
		diskSettings = []*types.DiskSettings{}
	}

	diskPlacement := -1
	for i, diskSetting := range vm.VM.VmSpecSection.DiskSection.DiskSettings {
		if diskSetting.DiskId == diskId {
			diskPlacement = i
		}
	}

	if diskPlacement == -1 {
		return ErrorEntityNotFound
	}

	// remove disk from slice
	diskSettings = append(diskSettings[:diskPlacement], diskSettings[diskPlacement+1:]...)

	vmSpecSection := vm.VM.VmSpecSection
	vmSpecSection.DiskSection.DiskSettings = diskSettings

	_, err = vm.UpdateInternalDisks(vmSpecSection)
	if err != nil {
		return fmt.Errorf("error deleting VM %s internal disk %s: %s", vm.VM.Name, diskId, err)
	}

	return nil
}

// UpdateInternalDisks applies disks configuration for the VM.
// types.VmSpecSection has to have all internal disk state. Disks which don't match provided ones in types.VmSpecSection
// will be deleted. Matched internal disk will be updated. New internal disk description found
// in types.VmSpecSection will be created. Returns updated types.VmSpecSection and error.
// Runs synchronously, VM is ready for another operation after this function returns.
func (vm *VM) UpdateInternalDisks(disksSettingToUpdate *types.VmSpecSection) (*types.VmSpecSection, error) {
	if vm.VM.HREF == "" {
		return nil, fmt.Errorf("cannot update internal disks - VM HREF is unset")
	}

	description := vm.VM.Description
	vm, err := vm.UpdateVmSpecSection(disksSettingToUpdate, description)
	if err != nil {
		return nil, err
	}

	return vm.VM.VmSpecSection, nil
}

// UpdateInternalDisksAsync applies disks configuration for the VM.
// types.VmSpecSection has to have all internal disk state. Disks which don't
// match provided ones in types.VmSpecSection will be deleted.
// Matched internal disk will be updated. New internal disk description found in types.VmSpecSection will be created.
// Returns Task and error.
//
// Deprecated: use UpdateInternalDisks or UpdateVmSpecSectionAsync instead
func (vm *VM) UpdateInternalDisksAsync(disksSettingToUpdate *types.VmSpecSection) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update disks, VM HREF is unset")
	}

	vmSpecSectionModified := true
	disksSettingToUpdate.Modified = &vmSpecSectionModified

	return vm.client.ExecuteTaskRequest(vm.VM.HREF+"/action/reconfigureVm", http.MethodPost,
		types.MimeVM, "error updating VM disks: %s", &types.VMDiskChange{
			Xmlns:         types.XMLNamespaceVCloud,
			Ovf:           types.XMLNamespaceOVF,
			Name:          vm.VM.Name,
			Description:   vm.VM.Description,
			VmSpecSection: disksSettingToUpdate,
		})
}

// AddEmptyVm adds an empty VM (without template) to vApp and returns the new created VM or an error.
func (vapp *VApp) AddEmptyVm(reComposeVAppParams *types.RecomposeVAppParamsForEmptyVm) (*VM, error) {
	apiVersion, err := vapp.client.MaxSupportedVersion()
	if err != nil {
		return nil, err
	}
	vmFunctions := getVmVersionedFuncsByVdcVersion("vm" + apiVersionToVcdVersion[apiVersion])
	if vmFunctions.AddEmptyVm == nil {
		return nil, fmt.Errorf("function AddEmptyVm is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] AddEmptyVm call function for version %s", vmFunctions.SupportedVersion)
	return vmFunctions.AddEmptyVm(vapp, reComposeVAppParams)
}

// AddEmptyVmAsync adds an empty VM (without template) to the vApp and returns a Task and an error.
func (vapp *VApp) AddEmptyVmAsync(reComposeVAppParams *types.RecomposeVAppParamsForEmptyVm) (Task, error) {
	apiVersion, err := vapp.client.MaxSupportedVersion()
	if err != nil {
		return Task{}, err
	}
	vmFunctions := getVmVersionedFuncsByVdcVersion("vm" + apiVersionToVcdVersion[apiVersion])
	if vmFunctions.AddEmptyVmAsync == nil {
		return Task{}, fmt.Errorf("function AddEmptyVmAsync is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] AddEmptyVmAsync call function for version %s", vmFunctions.SupportedVersion)

	return vmFunctions.AddEmptyVmAsync(vapp, reComposeVAppParams)
}

// validateEmptyVmParams checks if all required parameters are provided
func validateEmptyVmParams(reComposeVAppParams *types.RecomposeVAppParamsForEmptyVm) error {
	if reComposeVAppParams.CreateItem == nil {
		return fmt.Errorf("[AddEmptyVmAsync] CreateItem can't be empty")
	}

	if reComposeVAppParams.CreateItem.Name == "" {
		return fmt.Errorf("[AddEmptyVmAsync] CreateItem.Name can't be empty")
	}

	if reComposeVAppParams.CreateItem.VmSpecSection == nil {
		return fmt.Errorf("[AddEmptyVmAsync] CreateItem.VmSpecSection can't be empty")
	}

	if reComposeVAppParams.CreateItem.VmSpecSection.HardwareVersion == nil {
		return fmt.Errorf("[AddEmptyVmAsync] CreateItem.VmSpecSection.HardwareVersion can't be empty")
	}

	if reComposeVAppParams.CreateItem.VmSpecSection.HardwareVersion.Value == "" {
		return fmt.Errorf("[AddEmptyVmAsync] CreateItem.VmSpecSection.HardwareVersion.Value can't be empty")
	}

	if reComposeVAppParams.CreateItem.VmSpecSection.MemoryResourceMb == nil {
		return fmt.Errorf("[AddEmptyVmAsync] CreateItem.VmSpecSection.MemoryResourceMb can't be empty")
	}

	if reComposeVAppParams.CreateItem.VmSpecSection.MemoryResourceMb.Configured <= int64(0) {
		return fmt.Errorf("[AddEmptyVmAsync] CreateItem.VmSpecSection.MemoryResourceMb.Configured can't be empty")
	}

	return nil
}

// UpdateVmSpecSection updates VM Spec section and returns refreshed VM or error.
func (vm *VM) UpdateVmSpecSection(vmSettingsToUpdate *types.VmSpecSection, description string) (*VM, error) {
	task, err := vm.UpdateVmSpecSectionAsync(vmSettingsToUpdate, description)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = vm.Refresh()
	if err != nil {
		return nil, err
	}

	return vm, nil

}

// UpdateVmSpecSectionAsync updates VM Spec section and returns Task and error.
func (vm *VM) UpdateVmSpecSectionAsync(vmSettingsToUpdate *types.VmSpecSection, description string) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update VM spec section, VM HREF is unset")
	}

	// Firmware field is unavailable on <37.1 API Versions
	if vmSettingsToUpdate.Firmware != "" && vm.client.APIVCDMaxVersionIs("<37.1") {
		return Task{}, fmt.Errorf("VM Firmware can only be set on VCD 10.4.1+ (API 37.1+)")
	}

	vmSpecSectionModified := true
	vmSettingsToUpdate.Modified = &vmSpecSectionModified

	// `reconfigureVm` updates VM name, Description, and any or all of the following sections.
	//    VirtualHardwareSection
	//    OperatingSystemSection
	//    NetworkConnectionSection
	//    GuestCustomizationSection
	// Sections not included in the request body will not be updated.

	vmPayload := &types.Vm{
		Xmlns:         types.XMLNamespaceVCloud,
		Ovf:           types.XMLNamespaceOVF,
		Name:          vm.VM.Name,
		Description:   description,
		VmSpecSection: vmSettingsToUpdate,
	}

	// Since 37.1 there is a Firmware field in VmSpecSection
	return vm.client.ExecuteTaskRequestWithApiVersion(vm.VM.HREF+"/action/reconfigureVm",
		http.MethodPost, types.MimeVM, "error updating VM spec section: %s", vmPayload,
		vm.client.GetSpecificApiVersionOnCondition(">=37.1", "37.1"))
}

// UpdateComputePolicyV2 updates VM Compute policy with the given compute policies using v2.0.0 OpenAPI endpoint,
// and returns an error if something went wrong, or the refreshed VM if all went OK.
// Updating with an empty compute policy ID will remove it from the VM. Both policies can't be empty as the VM requires
// at least one policy.
func (vm *VM) UpdateComputePolicyV2(sizingPolicyId, placementPolicyId, vGpuPolicyId string) (*VM, error) {
	task, err := vm.UpdateComputePolicyV2Async(sizingPolicyId, placementPolicyId, vGpuPolicyId)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = vm.Refresh()
	if err != nil {
		return nil, err
	}

	return vm, nil

}

// UpdateComputePolicy updates VM compute policy and returns refreshed VM or error.
// Deprecated: Use VM.UpdateComputePolicyV2 instead
func (vm *VM) UpdateComputePolicy(computePolicy *types.VdcComputePolicy) (*VM, error) {
	task, err := vm.UpdateComputePolicyAsync(computePolicy)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = vm.Refresh()
	if err != nil {
		return nil, err
	}

	return vm, nil

}

// UpdateComputePolicyV2Async updates VM Compute policy with the given compute policies using v2.0.0 OpenAPI endpoint,
// and returns a Task and an error. Updating with an empty compute policy ID will remove it from the VM. Both
// policies can't be empty as the VM requires at least one policy.
// WARNING: At the moment, vGPU Policies are not supported. Using one will return an error.
func (vm *VM) UpdateComputePolicyV2Async(sizingPolicyId, placementPolicyId, vGpuPolicyId string) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update VM compute policy, VM HREF is unset")
	}

	sizingIsEmpty := strings.TrimSpace(sizingPolicyId) == ""
	placementIsEmpty := strings.TrimSpace(placementPolicyId) == ""
	vGpuPolicyIsEmpty := strings.TrimSpace(vGpuPolicyId) == ""

	if !vGpuPolicyIsEmpty {
		return Task{}, fmt.Errorf("vGPU policies are not supported, hence %s should be empty", vGpuPolicyId)
	}

	if sizingIsEmpty && placementIsEmpty {
		return Task{}, fmt.Errorf("either sizing policy ID or placement policy ID is needed")
	}

	// `reconfigureVm` updates VM name, Description, and any or all of the following sections.
	//    VirtualHardwareSection
	//    OperatingSystemSection
	//    NetworkConnectionSection
	//    GuestCustomizationSection
	// Sections not included in the request body will not be updated.

	computePolicy := &types.ComputePolicy{}

	if !sizingIsEmpty {
		vdcSizingPolicyHref, err := vm.client.OpenApiBuildEndpoint(types.OpenApiPathVersion2_0_0, types.OpenApiEndpointVdcComputePolicies, sizingPolicyId)
		if err != nil {
			return Task{}, fmt.Errorf("error constructing HREF for sizing policy")
		}
		computePolicy.VmSizingPolicy = &types.Reference{HREF: vdcSizingPolicyHref.String()}
	}

	if !placementIsEmpty {
		vdcPlacementPolicyHref, err := vm.client.OpenApiBuildEndpoint(types.OpenApiPathVersion2_0_0, types.OpenApiEndpointVdcComputePolicies, placementPolicyId)
		if err != nil {
			return Task{}, fmt.Errorf("error constructing HREF for placement policy")
		}
		computePolicy.VmPlacementPolicy = &types.Reference{HREF: vdcPlacementPolicyHref.String()}
	}

	return vm.client.ExecuteTaskRequest(vm.VM.HREF+"/action/reconfigureVm", http.MethodPost,
		types.MimeVM, "error updating VM spec section: %s", &types.Vm{
			Xmlns:         types.XMLNamespaceVCloud,
			Ovf:           types.XMLNamespaceOVF,
			Name:          vm.VM.Name,
			Description:   vm.VM.Description,
			ComputePolicy: computePolicy,
		})
}

// UpdateComputePolicyAsync updates VM Compute policy and returns Task and error.
// Deprecated: Use VM.UpdateComputePolicyV2Async instead
func (vm *VM) UpdateComputePolicyAsync(computePolicy *types.VdcComputePolicy) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update VM compute policy, VM HREF is unset")
	}

	// `reconfigureVm` updates VM name, Description, and any or all of the following sections.
	//    VirtualHardwareSection
	//    OperatingSystemSection
	//    NetworkConnectionSection
	//    GuestCustomizationSection
	// Sections not included in the request body will not be updated.

	vcdComputePolicyHref, err := vm.client.OpenApiBuildEndpoint(types.OpenApiPathVersion1_0_0, types.OpenApiEndpointVdcComputePolicies, computePolicy.ID)
	if err != nil {
		return Task{}, fmt.Errorf("error constructing HREF for compute policy")
	}

	return vm.client.ExecuteTaskRequest(vm.VM.HREF+"/action/reconfigureVm", http.MethodPost,
		types.MimeVM, "error updating VM spec section: %s", &types.Vm{
			Xmlns:         types.XMLNamespaceVCloud,
			Ovf:           types.XMLNamespaceOVF,
			Name:          vm.VM.Name,
			Description:   vm.VM.Description,
			ComputePolicy: &types.ComputePolicy{VmSizingPolicy: &types.Reference{HREF: vcdComputePolicyHref.String()}},
		})
}

// QueryVmList returns a list of all VMs in all the organizations available to the caller
func (client *Client) QueryVmList(filter types.VmQueryFilter) ([]*types.QueryResultVMRecordType, error) {
	var vmList []*types.QueryResultVMRecordType
	queryType := client.GetQueryType(types.QtVm)
	params := map[string]string{
		"type":          queryType,
		"filterEncoded": "true",
	}
	if filter.String() != "" {
		params["filter"] = filter.String()
	}
	vmResult, err := client.cumulativeQuery(queryType, nil, params)
	if err != nil {
		return nil, fmt.Errorf("error getting VM list : %s", err)
	}
	vmList = vmResult.Results.VMRecord
	if client.IsSysAdmin {
		vmList = vmResult.Results.AdminVMRecord
	}
	return vmList, nil
}

// QueryVmList returns a list of all VMs in a given Org
func (org *Org) QueryVmList(filter types.VmQueryFilter) ([]*types.QueryResultVMRecordType, error) {
	if org.client.IsSysAdmin {
		return queryVmList(filter, org.client, "org", org.Org.HREF)
	}
	return queryVmList(filter, org.client, "", "")
}

// QueryVmList returns a list of all VMs in a given VDC
func (vdc *Vdc) QueryVmList(filter types.VmQueryFilter) ([]*types.QueryResultVMRecordType, error) {
	return queryVmList(filter, vdc.client, "vdc", vdc.Vdc.HREF)
}

// queryVmList is extracted and used by org.QueryVmList and vdc.QueryVmList to adjust filtering scope
func queryVmList(filter types.VmQueryFilter, client *Client, filterParent, filterParentHref string) ([]*types.QueryResultVMRecordType, error) {
	var vmList []*types.QueryResultVMRecordType
	queryType := client.GetQueryType(types.QtVm)
	params := map[string]string{
		"type":          queryType,
		"filterEncoded": "true",
	}
	filterText := ""
	if filter.String() != "" {
		filterText = filter.String()
	}
	if filterParent != "" {
		if filterText == "" {
			filterText = fmt.Sprintf("%s==%s", filterParent, filterParentHref)
		} else {
			filterText = fmt.Sprintf("%s;%s==%s", filterText, filterParent, filterParentHref)
		}
		params["filter"] = filterText
	}
	vmResult, err := client.cumulativeQuery(queryType, nil, params)
	if err != nil {
		return nil, fmt.Errorf("error getting VM list : %s", err)
	}
	vmList = vmResult.Results.VMRecord
	if client.IsSysAdmin {
		vmList = vmResult.Results.AdminVMRecord
	}
	return vmList, nil
}

// QueryVmList retrieves a list of VMs across all VDC, using parameters defined in searchParams
func QueryVmList(vmType types.VmQueryFilter, client *Client, searchParams map[string]string) ([]*types.QueryResultVMRecordType, error) {
	var vmList []*types.QueryResultVMRecordType
	queryType := client.GetQueryType(types.QtVm)
	params := map[string]string{
		"type":          queryType,
		"filterEncoded": "true",
	}
	filterText := ""
	if vmType.String() != "" {
		// The first filter will be the type of VM, i.e. deployed (inside a vApp) or not (inside a vApp template)
		filterText = vmType.String()
	}
	for k, v := range searchParams {
		filterText = fmt.Sprintf("%s;%s==%s", filterText, k, v)
	}

	params["filter"] = filterText
	vmResult, err := client.cumulativeQuery(queryType, nil, params)
	if err != nil {
		return nil, fmt.Errorf("error getting VM list : %s", err)
	}
	vmList = vmResult.Results.VMRecord
	if client.IsSysAdmin {
		vmList = vmResult.Results.AdminVMRecord
	}
	return vmList, nil
}

// UpdateVmCpuAndMemoryHotAdd updates VM Capabilities and returns refreshed VM or error.
func (vm *VM) UpdateVmCpuAndMemoryHotAdd(cpuAdd, memoryAdd bool) (*VM, error) {
	task, err := vm.UpdateVmCpuAndMemoryHotAddAsync(cpuAdd, memoryAdd)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = vm.Refresh()
	if err != nil {
		return nil, err
	}

	return vm, nil

}

// UpdateVmCpuAndMemoryHotAddAsync updates VM Capabilities and returns Task and error.
func (vm *VM) UpdateVmCpuAndMemoryHotAddAsync(cpuHot, memoryAdd bool) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update VM capabilities, VM HREF is unset")
	}

	return vm.client.ExecuteTaskRequest(vm.VM.HREF+"/vmCapabilities", http.MethodPut,
		types.MimeVmCapabilities, "error updating VM capabilities section: %s", &types.VmCapabilities{
			Xmlns:               types.XMLNamespaceVCloud,
			CPUHotAddEnabled:    cpuHot,
			MemoryHotAddEnabled: memoryAdd,
		})
}

// vmVersionedFuncs is a generic representation of VM CRUD operations across multiple versions
type vmVersionedFuncs struct {
	SupportedVersion string
	GetVMByHref      func(client *Client, vmHref string) (*VM, error)
	AddEmptyVm       func(vapp *VApp, reComposeVAppParams *types.RecomposeVAppParamsForEmptyVm) (*VM, error)
	AddEmptyVmAsync  func(vapp *VApp, reComposeVAppParams *types.RecomposeVAppParamsForEmptyVm) (Task, error)
}

// VM function mapping for API version 33.0 (from vCD 10.0)
var vmVersionedFuncsV10 = vmVersionedFuncs{
	SupportedVersion: "33.0",
	GetVMByHref:      getVMByHrefV10,
	AddEmptyVm:       addEmptyVmV10,
	AddEmptyVmAsync:  addEmptyVmAsyncV10,
}

// vmVersionedFuncsByVcdVersion is a map of VDC functions by vCD version
var vmVersionedFuncsByVcdVersion = map[string]vmVersionedFuncs{
	"vm10.2": vmVersionedFuncsV10,
	"vm10.1": vmVersionedFuncsV10,
	"vm10.0": vmVersionedFuncsV10,
	// If we add a new function to this list, we also need to update the "default" entry
	// The "default" entry will hold the highest currently available function
	"default": vmVersionedFuncsV10,
}

// getVmVersionedFuncsByVdcVersion is a wrapper function that retrieves the requested versioned VDC function
// When the wanted version does  not exist in the map, it returns the highest available one.
func getVmVersionedFuncsByVdcVersion(version string) vmVersionedFuncs {
	f, ok := vmVersionedFuncsByVcdVersion[version]
	if ok {
		return f
	} else {
		return vmVersionedFuncsByVcdVersion["default"]
	}
}

// addEmptyVmAsyncV10 adds an empty VM (without template) to the vApp and returns a Task and an error.
func addEmptyVmAsyncV10(vapp *VApp, reComposeVAppParams *types.RecomposeVAppParamsForEmptyVm) (Task, error) {
	err := validateEmptyVmParams(reComposeVAppParams)
	if err != nil {
		return Task{}, err
	}
	apiEndpoint := urlParseRequestURI(vapp.VApp.HREF)
	apiEndpoint.Path += "/action/recomposeVApp"

	reComposeVAppParams.XmlnsVcloud = types.XMLNamespaceVCloud
	reComposeVAppParams.XmlnsOvf = types.XMLNamespaceOVF

	// Return the task
	return vapp.client.ExecuteTaskRequestWithApiVersion(apiEndpoint.String(), http.MethodPost,
		types.MimeRecomposeVappParams, "error instantiating a new VM: %s", reComposeVAppParams,
		vapp.client.GetSpecificApiVersionOnCondition(">=37.1", "37.1"))
}

// addEmptyVmV10 adds an empty VM (without template) to vApp and returns the new created VM or an error.
func addEmptyVmV10(vapp *VApp, reComposeVAppParams *types.RecomposeVAppParamsForEmptyVm) (*VM, error) {
	task, err := addEmptyVmAsyncV10(vapp, reComposeVAppParams)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = vapp.Refresh()
	if err != nil {
		return nil, fmt.Errorf("error refreshing vApp: %s", err)
	}

	//vApp Might Not Have Any VMs
	if vapp.VApp.Children == nil {
		return nil, ErrorEntityNotFound
	}

	util.Logger.Printf("[TRACE] Looking for VM: %s", reComposeVAppParams.CreateItem.Name)
	for _, child := range vapp.VApp.Children.VM {

		util.Logger.Printf("[TRACE] Looking at: %s", child.Name)
		if child.Name == reComposeVAppParams.CreateItem.Name {
			return getVMByHrefV10(vapp.client, child.HREF)
		}

	}
	util.Logger.Printf("[TRACE] Couldn't find VM: %s", reComposeVAppParams.CreateItem.Name)
	return nil, ErrorEntityNotFound
}

// getVMByHrefV10 returns a VM reference by running a vCD API call
// If no valid VM is found, it returns a nil VM reference and an error
// Note that the pointer receiver here is a Client instead of a VApp, because
// there are cases where we know the VM HREF but not which VApp it belongs to.
// V10 of function overrides API version to allow to access compute policy in VM.
func getVMByHrefV10(client *Client, vmHref string) (*VM, error) {

	newVm := NewVM(client)

	_, err := client.ExecuteRequestWithApiVersion(vmHref, http.MethodGet,
		"", "error retrieving vm: %s", nil, newVm.VM, client.GetSpecificApiVersionOnCondition(">=37.1", "37.1"))

	if err != nil {

		return nil, err
	}

	return newVm, nil
}

// GetVMByHref returns a VM reference by running a vCD API call
// If no valid VM is found, it returns a nil VM reference and an error
// Note that the pointer receiver here is a Client instead of a VApp, because
// there are cases where we know the VM HREF but not which VApp it belongs to.
func (client *Client) GetVMByHref(vmHref string) (*VM, error) {

	apiVersion, err := client.MaxSupportedVersion()
	if err != nil {
		return nil, err
	}
	vmFunctions := getVmVersionedFuncsByVdcVersion("vm" + apiVersionToVcdVersion[apiVersion])
	if vmFunctions.GetVMByHref == nil {
		return nil, fmt.Errorf("function GetVMByHref is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] GetVMByHref call function for version %s", vmFunctions.SupportedVersion)

	return vmFunctions.GetVMByHref(client, vmHref)
}

// UpdateStorageProfile updates VM storage profile and returns refreshed VM or error.
func (vm *VM) UpdateStorageProfile(storageProfileHref string) (*VM, error) {
	task, err := vm.UpdateStorageProfileAsync(storageProfileHref)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = vm.Refresh()
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// UpdateStorageProfileAsync updates VM storage profile and returns Task and error.
func (vm *VM) UpdateStorageProfileAsync(storageProfileHref string) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update VM storage profile, VM HREF is unset")
	}
	if storageProfileHref == "" {
		return Task{}, fmt.Errorf("cannot update VM storage profile, storage profile HREF is unset")
	}

	// `reconfigureVm` updates VM name, Description, and any or all of the following sections.
	//    VirtualHardwareSection
	//    OperatingSystemSection
	//    NetworkConnectionSection
	//    GuestCustomizationSection
	// Sections not included in the request body will not be updated.
	return vm.client.ExecuteTaskRequest(vm.VM.HREF+"/action/reconfigureVm", http.MethodPost,
		types.MimeVM, "error updating VM storage profile: %s", &types.Vm{
			Xmlns:          types.XMLNamespaceVCloud,
			Ovf:            types.XMLNamespaceOVF,
			Name:           vm.VM.Name,
			Description:    vm.VM.Description,
			StorageProfile: &types.Reference{HREF: storageProfileHref},
		})
}

// UpdateBootOptions updates the Boot Options of a VM and returns the updated instance of the VM
func (vm *VM) UpdateBootOptions(bootOptions *types.BootOptions) (*VM, error) {
	task, err := vm.UpdateBootOptionsAsync(bootOptions)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}

	err = vm.Refresh()
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// UpdateBootOptionsAsync updates the boot options of a VM
func (vm *VM) UpdateBootOptionsAsync(bootOptions *types.BootOptions) (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot update VM boot options, VM HREF is unset")
	}

	if vm.client.APIVCDMaxVersionIs("<37.1") {

		if bootOptions.BootRetryEnabled != nil || bootOptions.BootRetryDelay != nil ||
			bootOptions.EfiSecureBootEnabled != nil || bootOptions.NetworkBootProtocol != "" {
			return Task{}, fmt.Errorf("error: Boot retry, EFI Secure Boot and Boot Network Protocol options were introduced in VCD 10.4.1")
		}
	}

	if bootOptions == nil {
		return Task{}, fmt.Errorf("cannot update VM boot options, none given")
	}

	return vm.client.ExecuteTaskRequestWithApiVersion(vm.VM.HREF+"/action/reconfigureVm", http.MethodPost,
		types.MimeVM, "error updating VM boot options: %s", &types.Vm{
			Xmlns:       types.XMLNamespaceVCloud,
			Ovf:         types.XMLNamespaceOVF,
			Name:        vm.VM.Name,
			Description: vm.VM.Description,
			// We need to add ComputePolicy in the Request Body or settings will
			// be set to default sizing policy set in the VDC if the VM is Not
			// compliant with the current sizing policy
			ComputePolicy: vm.VM.ComputePolicy,
			BootOptions:   bootOptions,
		}, vm.client.GetSpecificApiVersionOnCondition(">=37.1", "37.1"))
}

// DeleteAsync starts a standalone VM deletion, returning a task
func (vm *VM) DeleteAsync() (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("no HREF found for this VM")
	}

	task, err := vm.Undeploy()
	if err == nil {
		err = task.WaitTaskCompletion()
		if err != nil {
			return Task{}, err
		}
	}
	return vm.client.ExecuteTaskRequest(vm.VM.HREF, http.MethodDelete,
		"", "error deleting VM: %s", nil)
}

// Delete deletes a standalone VM
func (vm *VM) Delete() error {
	task, err := vm.DeleteAsync()
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

func (vm *VM) getTenantContext() (*TenantContext, error) {
	parentVdc, err := vm.GetParentVdc()
	if err != nil {
		return nil, err
	}
	return parentVdc.getTenantContext()
}

// ChangeMemory sets memory value. Size is MB
func (vm *VM) ChangeMemory(sizeInMb int64) error {
	vmSpecSection := vm.VM.VmSpecSection
	description := vm.VM.Description
	// update treats same values as changes and fails, with no values provided - no changes are made for that section
	vmSpecSection.DiskSection = nil

	vmSpecSection.MemoryResourceMb.Configured = sizeInMb

	_, err := vm.UpdateVmSpecSection(vmSpecSection, description)
	if err != nil {
		return fmt.Errorf("error changing memory size: %s", err)
	}
	return nil
}

// ChangeCPUCount sets number of available virtual logical processors
// (i.e. CPUs x cores per socket)
// Cpu cores count is inherited from template.
// https://communities.vmware.com/thread/576209
// Deprecated: use ChangeCPUAndCoreCount
func (vm *VM) ChangeCPU(cpus, cpuCores int) error {
	vmSpecSection := vm.VM.VmSpecSection
	description := vm.VM.Description
	// update treats same values as changes and fails, with no values provided - no changes are made for that section
	vmSpecSection.DiskSection = nil

	vmSpecSection.NumCpus = &cpus
	// has to come together
	vmSpecSection.NumCoresPerSocket = &cpuCores

	_, err := vm.UpdateVmSpecSection(vmSpecSection, description)
	if err != nil {
		return fmt.Errorf("error changing cpu size: %s", err)
	}
	return nil
}

// ChangeCPUAndCoreCount sets CPU and CPU core counts
// Accepts values or `nil` for both parameters.
func (vm *VM) ChangeCPUAndCoreCount(cpus, cpuCores *int) error {
	vmSpecSection := vm.VM.VmSpecSection
	description := vm.VM.Description
	// update treats same values as changes and fails, with no values provided - no changes are made for that section
	vmSpecSection.DiskSection = nil

	vmSpecSection.NumCpus = cpus
	// has to come together
	vmSpecSection.NumCoresPerSocket = cpuCores

	_, err := vm.UpdateVmSpecSection(vmSpecSection, description)
	if err != nil {
		return fmt.Errorf("error changing CPU size: %s", err)
	}
	return nil
}

// ConsolidateDisksAsync triggers VM disk consolidation task
func (vm *VM) ConsolidateDisksAsync() (Task, error) {
	if vm.VM.HREF == "" {
		return Task{}, fmt.Errorf("cannot consolidate disks, VM HREF is unset")
	}

	return vm.client.ExecuteTaskRequest(vm.VM.HREF+"/action/consolidate", http.MethodPost,
		types.AnyXMLMime, "error consolidating VM disks: %s", nil)
}

// ConsolidateDisks triggers VM disk consolidation task and waits until it is completed
func (vm *VM) ConsolidateDisks() error {
	task, err := vm.ConsolidateDisksAsync()
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// GetExtraConfig retrieves the extra configuration items from a VM
func (vm *VM) GetExtraConfig() ([]*types.ExtraConfigMarshal, error) {
	if vm.VM.HREF == "" {
		return nil, fmt.Errorf("cannot update VM spec section, VM HREF is unset")
	}

	virtualHardwareSection := &types.ResponseVirtualHardwareSection{}
	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/virtualHardwareSection/", http.MethodGet, types.MimeVirtualHardwareSection, "error retrieving virtual hardware: %s", nil, virtualHardwareSection)
	if err != nil {
		return nil, err
	}

	convertedExtraConfig := convertExtraConfig(virtualHardwareSection.ExtraConfigs)

	return convertedExtraConfig, nil
}

// UpdateExtraConfig adds or changes items in the VM Extra Configuration set
// Returns the modified set
// Note: an item with an empty `Value` will be deleted.
func (vm *VM) UpdateExtraConfig(update []*types.ExtraConfigMarshal) ([]*types.ExtraConfigMarshal, error) {
	return vm.updateExtraConfig(update, false)
}

// DeleteExtraConfig removes items from the VM Extra Configuration set
// Returns the modified set
func (vm *VM) DeleteExtraConfig(deleteItems []*types.ExtraConfigMarshal) ([]*types.ExtraConfigMarshal, error) {
	return vm.updateExtraConfig(deleteItems, true)
}

// updateExtraConfig adds, changes, or delete items in the VM Extra Configuration set
func (vm *VM) updateExtraConfig(update []*types.ExtraConfigMarshal, wantDelete bool) ([]*types.ExtraConfigMarshal, error) {
	if vm.VM.HREF == "" {
		return nil, fmt.Errorf("cannot update VM spec section, VM HREF is unset")
	}

	virtualHardwareSection := &types.ResponseVirtualHardwareSection{}
	_, err := vm.client.ExecuteRequest(vm.VM.HREF+"/virtualHardwareSection/", http.MethodGet, types.MimeVirtualHardwareSection, "error retrieving virtual hardware: %s", nil, virtualHardwareSection)
	if err != nil {
		return nil, err
	}

	var newExtraConfig []*types.ExtraConfigMarshal

	var invalidKeys []string

	if wantDelete {
		for _, ec := range update {
			newExtraConfig = append(newExtraConfig, &types.ExtraConfigMarshal{Key: ec.Key, Value: ""})
		}

	} else {
		for _, ec := range update {
			if strings.Contains(ec.Key, " ") {
				invalidKeys = append(invalidKeys, ec.Key)
				continue
			}
			newExtraConfig = append(newExtraConfig, ec)
		}
		if len(invalidKeys) > 0 {
			return nil, fmt.Errorf("[vm.UpdateExtraConfig] invalid keys provided: [%s]", strings.Join(invalidKeys, ","))
		}
	}

	requestVirtualHardwareSection := &types.RequestVirtualHardwareSection{
		Info:  "Virtual hardware requirements",
		Ovf:   types.XMLNamespaceOVF,
		Rasd:  types.XMLNamespaceRASD,
		Vssd:  types.XMLNamespaceVSSD,
		Ns2:   types.XMLNamespaceVCloud,
		Ns3:   types.XMLNamespaceVCloud,
		Ns4:   types.XMLNamespaceVCloud,
		Ns5:   types.XMLNamespaceVCloud,
		Vmw:   types.XMLNamespaceVMW,
		Xmlns: types.XMLNamespaceVCloud,

		Type:   virtualHardwareSection.Type,
		System: virtualHardwareSection.System,
		Item:   virtualHardwareSection.Item,

		ExtraConfigs: newExtraConfig,
	}

	task, err := vm.client.ExecuteTaskRequest(vm.VM.HREF+"/virtualHardwareSection/", http.MethodPut,
		types.MimeVirtualHardwareSection, "error updating VM spec section: %s", requestVirtualHardwareSection)
	if err != nil {
		return nil, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("error waiting task: %s", err)
	}

	xtraCfg, err := vm.GetExtraConfig()
	if err != nil {
		return nil, fmt.Errorf("got error while retrieving extra config: %s", err)
	}

	return xtraCfg, nil
}

func convertExtraConfig(source []*types.ExtraConfig) []*types.ExtraConfigMarshal {
	resp := make([]*types.ExtraConfigMarshal, len(source))
	for index, field := range source {
		resp[index] = &types.ExtraConfigMarshal{
			Key:      field.Key,
			Value:    field.Value,
			Required: field.Required,
		}
	}

	return resp
}
