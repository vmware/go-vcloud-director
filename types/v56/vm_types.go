/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
package types

import "encoding/xml"

// Vm represents a virtual machine
// Type: VmType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a virtual machine.
// Since: 0.9
// This structure used to be called `VM`, and needed an XMLName to adjust the XML entity name upon marshalling.
// We have renamed it to `Vm` to remove the ambiguity and avoid XMLName conflicts when embedding this type
// into another structure.
// Now, there is no need for XMLName, as the name of the structure is the same as the XML entity
type Vm struct {
	// Attributes
	Ovf   string `xml:"xmlns:ovf,attr,omitempty"`
	Xsi   string `xml:"xmlns:xsi,attr,omitempty"`
	Xmlns string `xml:"xmlns,attr,omitempty"`

	HREF                    string `xml:"href,attr,omitempty"`                    // The URI of the entity.
	Type                    string `xml:"type,attr,omitempty"`                    // The MIME type of the entity.
	ID                      string `xml:"id,attr,omitempty"`                      // The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused
	OperationKey            string `xml:"operationKey,attr,omitempty"`            // Optional unique identifier to support idempotent semantics for create and delete operations.
	Name                    string `xml:"name,attr"`                              // The name of the entity.
	Status                  int    `xml:"status,attr,omitempty"`                  // Creation status of the resource entity.
	Deployed                bool   `xml:"deployed,attr,omitempty"`                // True if the virtual machine is deployed.
	NeedsCustomization      bool   `xml:"needsCustomization,attr,omitempty"`      // True if this virtual machine needs customization.
	NestedHypervisorEnabled bool   `xml:"nestedHypervisorEnabled,attr,omitempty"` // True if hardware-assisted CPU virtualization capabilities in the host should be exposed to the guest operating system.
	// Elements
	Link        LinkList         `xml:"Link,omitempty"`        // A reference to an entity or operation associated with this object.
	Description string           `xml:"Description,omitempty"` // Optional description.
	Tasks       *TasksInProgress `xml:"Tasks,omitempty"`       // A list of queued, running, or recently completed tasks associated with this entity.
	Files       *FilesList       `xml:"FilesList,omitempty"`   // Represents a list of files to be transferred (uploaded or downloaded). Each File in the list is part of the ResourceEntity.
	VAppParent  *Reference       `xml:"VAppParent,omitempty"`  // Reserved. Unimplemented.
	// TODO: OVF Sections to be implemented
	// Section OVF_Section `xml:"Section,omitempty"
	DateCreated string `xml:"DateCreated,omitempty"` // Creation date/time of the vApp.

	// Section ovf:VirtualHardwareSection
	VirtualHardwareSection *VirtualHardwareSection `xml:"VirtualHardwareSection,omitempty"`

	RuntimeInfoSection *RuntimeInfoSection `xml:"RuntimeInfoSection,omitempty"`

	// FIXME: Upstream bug? Missing NetworkConnectionSection
	NetworkConnectionSection *NetworkConnectionSection `xml:"NetworkConnectionSection,omitempty"`

	VAppScopedLocalID string `xml:"VAppScopedLocalId,omitempty"` // A unique identifier for the virtual machine in the scope of the vApp.

	Snapshots *SnapshotSection `xml:"SnapshotSection,omitempty"`

	// TODO: OVF Sections to be implemented
	// Environment OVF_Environment `xml:"Environment,omitempty"

	VmSpecSection *VmSpecSection `xml:"VmSpecSection,omitempty"`

	// GuestCustomizationSection contains settings for VM customization like admin password, SID
	// changes, domain join configuration, etc
	GuestCustomizationSection *GuestCustomizationSection `xml:"GuestCustomizationSection,omitempty"`

	VMCapabilities *VmCapabilities `xml:"VmCapabilities,omitempty"` // Allows you to specify certain capabilities of this virtual machine.
	StorageProfile *Reference      `xml:"StorageProfile,omitempty"` // A reference to a storage profile to be used for this object. The specified storage profile must exist in the organization vDC that contains the object. If not specified, the default storage profile for the vDC is used.
	ProductSection *ProductSection `xml:"ProductSection,omitempty"`
	ComputePolicy  *ComputePolicy  `xml:"ComputePolicy,omitempty"` // accessible only from version API 33.0
	Media          *Reference      `xml:"Media,omitempty"`         // Reference to the media object to insert in a new VM.
}

type RuntimeInfoSection struct {
	Ns10        string `xml:"ns10,attr"`
	Type        string `xml:"type,attr"`
	Href        string `xml:"href,attr"`
	Info        string `xml:"Info"`
	VMWareTools struct {
		Version string `xml:"version,attr"`
	} `xml:"VMWareTools"`
}

// VmSpecSection from Vm struct
type VmSpecSection struct {
	Modified          *bool             `xml:"Modified,attr,omitempty"`
	Info              string            `xml:"ovf:Info"`
	OsType            string            `xml:"OsType,omitempty"`            // The type of the OS. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	NumCpus           *int              `xml:"NumCpus,omitempty"`           // Number of CPUs. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	NumCoresPerSocket *int              `xml:"NumCoresPerSocket,omitempty"` // Number of cores among which to distribute CPUs in this virtual machine. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	CpuResourceMhz    *CpuResourceMhz   `xml:"CpuResourceMhz,omitempty"`    // CPU compute resources. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	MemoryResourceMb  *MemoryResourceMb `xml:"MemoryResourceMb"`            // Memory compute resources. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	MediaSection      *MediaSection     `xml:"MediaSection,omitempty"`      // The media devices of this VM.
	DiskSection       *DiskSection      `xml:"DiskSection,omitempty"`       // virtual disks of this VM.
	HardwareVersion   *HardwareVersion  `xml:"HardwareVersion"`             // vSphere name of Virtual Hardware Version of this VM. Example: vmx-13 - This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	VmToolsVersion    string            `xml:"VmToolsVersion,omitempty"`    // VMware tools version of this VM.
	VirtualCpuType    string            `xml:"VirtualCpuType,omitempty"`    // The capabilities settings for this VM. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	TimeSyncWithHost  *bool             `xml:"TimeSyncWithHost,omitempty"`  // Synchronize the VM's time with the host.
}

// RecomposeVAppParamsForEmptyVm represents a vApp structure which allows to create VM.
type RecomposeVAppParamsForEmptyVm struct {
	XMLName          xml.Name    `xml:"RecomposeVAppParams"`
	XmlnsVcloud      string      `xml:"xmlns,attr"`
	XmlnsOvf         string      `xml:"xmlns:ovf,attr"`
	PowerOn          bool        `xml:"powerOn,attr,omitempty"` // True if the VM should be powered-on after creation. Defaults to false.
	CreateItem       *CreateItem `xml:"CreateItem,omitempty"`
	AllEULAsAccepted bool        `xml:"AllEULAsAccepted,omitempty"`
}

// CreateItem represents structure to create VM, part of RecomposeVAppParams structure.
type CreateItem struct {
	Name                      string                     `xml:"name,attr,omitempty"`
	Description               string                     `xml:"Description,omitempty"`
	GuestCustomizationSection *GuestCustomizationSection `xml:"GuestCustomizationSection,omitempty"`
	NetworkConnectionSection  *NetworkConnectionSection  `xml:"NetworkConnectionSection,omitempty"`
	VmSpecSection             *VmSpecSection             `xml:"VmSpecSection,omitempty"`
	StorageProfile            *Reference                 `xml:"StorageProfile,omitempty"`
	ComputePolicy             *ComputePolicy             `xml:"ComputePolicy,omitempty"` // accessible only from version API 33.0
	BootImage                 *Media                     `xml:"Media,omitempty"`         // boot image as vApp template. Href, Id and name needed.
}

// ComputePolicy represents structure to manage VM compute polices, part of RecomposeVAppParams structure.
type ComputePolicy struct {
	HREF                   string     `xml:"href,attr,omitempty"`
	Type                   string     `xml:"type,attr,omitempty"`
	Link                   *Link      `xml:"Link,omitempty"`                   // A reference to an entity or operation associated with this object.
	VmPlacementPolicy      *Reference `xml:"VmPlacementPolicy,omitempty"`      // VdcComputePolicy that defines VM's placement on a host through various affinity constraints.
	VmPlacementPolicyFinal *bool      `xml:"VmPlacementPolicyFinal,omitempty"` // True indicates that the placement policy cannot be removed from a VM that is instantiated with it. This value defaults to false.
	VmSizingPolicy         *Reference `xml:"VmSizingPolicy,omitempty"`         // VdcComputePolicy that defines VM's sizing and resource allocation.
	VmSizingPolicyFinal    *bool      `xml:"VmSizingPolicyFinal,omitempty"`    // True indicates that the sizing policy cannot be removed from a VM that is instantiated with it. This value defaults to false.
}

// CreateVmParams is used to create a standalone VM without a template
type CreateVmParams struct {
	XMLName     xml.Name   `xml:"CreateVmParams"`
	XmlnsOvf    string     `xml:"xmlns:ovf,attr"`
	Xmlns       string     `xml:"xmlns,attr,omitempty"`
	Name        string     `xml:"name,attr,omitempty"`    // Typically used to name or identify the subject of the request. For example, the name of the object being created or modified.
	PowerOn     bool       `xml:"powerOn,attr,omitempty"` // True if the VM should be powered-on after creation. Defaults to false.
	Description string     `xml:"Description,omitempty"`  // Optional description
	CreateVm    *Vm        `xml:"CreateVm"`               // Read-only information about the VM to create. This information appears in the Task returned by a createVm request.
	Media       *Reference `xml:"Media,omitempty"`        // Reference to the media object to insert in the new VM.
}

// InstantiateVmTemplateParams is used to create a standalone VM with a template
type InstantiateVmTemplateParams struct {
	XMLName               xml.Name                 `xml:"InstantiateVmTemplateParams"`
	XmlnsOvf              string                   `xml:"xmlns:ovf,attr"`
	Xmlns                 string                   `xml:"xmlns,attr,omitempty"`
	Name                  string                   `xml:"name,attr,omitempty"`             // Typically used to name or identify the subject of the request. For example, the name of the object being created or modified.
	PowerOn               bool                     `xml:"powerOn,attr,omitempty"`          // True if the VM should be powered-on after creation. Defaults to false.
	Description           string                   `xml:"Description,omitempty"`           // Optional description
	SourcedVmTemplateItem *SourcedVmTemplateParams `xml:"SourcedVmTemplateItem,omitempty"` // Represents virtual machine instantiation parameters.
	AllEULAsAccepted      bool                     `xml:"AllEULAsAccepted,omitempty"`      // True confirms acceptance of all EULAs in a vApp template. Instantiation fails if this element is missing, empty, or set to false and one or more EulaSection elements are present.
	ComputePolicy         *ComputePolicy           `xml:"ComputePolicy,omitempty"`         // A reference to a vdc compute policy. This contains VM's actual vdc compute policy reference and also optionally an add-on policy which always defines VM's sizing.
}

// SourcedVmTemplateParams represents the standalone VM instantiation parameters
type SourcedVmTemplateParams struct {
	LocalityParams                *LocalityParams      `xml:"LocalityParams,omitempty"`                // Locality parameters provide a hint that may help optimize placement of a VM and an independent a Disk so that the VM can make efficient use of the disk.
	Source                        *Reference           `xml:"Source"`                                  // A reference to an existing VM template
	VmCapabilities                *VmCapabilities      `xml:"VmCapabilities,omitempty"`                // Describes the capabilities (hot swap, etc.) the instantiated VM should have.
	VmGeneralParams               *VMGeneralParams     `xml:"VmGeneralParams,omitempty"`               // Specify name, description, and other properties of a VM during instantiation.
	VmTemplateInstantiationParams *InstantiationParams `xml:"VmTemplateInstantiationParams,omitempty"` // Same as InstantiationParams used for VMs within a vApp
	StorageProfile                *Reference           `xml:"StorageProfile,omitempty"`                // A reference to a storage profile to be used for the VM. The specified storage profile must exist in the organization vDC that contains the composed vApp. If not specified, the default storage profile for the vDC is used.
}
