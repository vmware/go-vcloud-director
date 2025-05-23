// Package types/v56 provider all types which are used by govcd package in order to perform API
// requests and parse responses

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/xml"
	"fmt"
	"sort"
)

// Maps status Attribute Values for VAppTemplate, VApp, VM, and Media Objects
var VAppStatuses = map[int]string{
	-1: "FAILED_CREATION",
	0:  "UNRESOLVED",
	1:  "RESOLVED",
	2:  "DEPLOYED",
	3:  "SUSPENDED",
	4:  "POWERED_ON",
	5:  "WAITING_FOR_INPUT",
	6:  "UNKNOWN",
	7:  "UNRECOGNIZED",
	8:  "POWERED_OFF",
	9:  "INCONSISTENT_STATE",
	10: "MIXED",
	11: "DESCRIPTOR_PENDING",
	12: "COPYING_CONTENTS",
	13: "DISK_CONTENTS_PENDING",
	14: "QUARANTINED",
	15: "QUARANTINE_EXPIRED",
	16: "REJECTED",
	17: "TRANSFER_TIMEOUT",
	18: "VAPP_UNDEPLOYED",
	19: "VAPP_PARTIALLY_DEPLOYED",
	20: "PARTIALLY_POWERED_OFF", // VCD 10.3+
	21: "PARTIALLY_SUSPENDED",
}

// Maps status Attribute Values for VDC Objects
var VDCStatuses = map[int]string{
	-1: "FAILED_CREATION",
	0:  "NOT_READY",
	1:  "READY",
	2:  "UNKNOWN",
	3:  "UNRECOGNIZED",
}

// VCD API

// DefaultStorageProfileSection is the name of the storage profile that will be specified for this virtual machine. The named storage profile must exist in the organization vDC that contains the virtual machine. If not specified, the default storage profile for the vDC is used.
// Type: DefaultStorageProfileSection_Type
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Name of the storage profile that will be specified for this virtual machine. The named storage profile must exist in the organization vDC that contains the virtual machine. If not specified, the default storage profile for the vDC is used.
// Since: 5.1
type DefaultStorageProfileSection struct {
	StorageProfile string `xml:"StorageProfile,omitempty"`
}

// CustomizationSection represents a vApp template customization settings.
// Type: CustomizationSectionType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a vApp template customization settings.
// Since: 1.0
type CustomizationSection struct {
	// FIXME: OVF Section needs to be laid down correctly
	Info string `xml:"ovf:Info"`
	//
	GoldMaster             bool     `xml:"goldMaster,attr,omitempty"`
	HREF                   string   `xml:"href,attr,omitempty"`
	Type                   string   `xml:"type,attr,omitempty"`
	CustomizeOnInstantiate bool     `xml:"CustomizeOnInstantiate"`
	Link                   LinkList `xml:"Link,omitempty"`
}

// LeaseSettingsSection represents vApp lease settings.
// Type: LeaseSettingsSectionType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents vApp lease settings.
// Since: 0.9
type LeaseSettingsSection struct {
	HREF                      string `xml:"href,attr,omitempty"`
	Type                      string `xml:"type,attr,omitempty"`
	DeploymentLeaseExpiration string `xml:"DeploymentLeaseExpiration,omitempty"`
	DeploymentLeaseInSeconds  int    `xml:"DeploymentLeaseInSeconds,omitempty"`
	Link                      *Link  `xml:"Link,omitempty"`
	StorageLeaseExpiration    string `xml:"StorageLeaseExpiration,omitempty"`
	StorageLeaseInSeconds     int    `xml:"StorageLeaseInSeconds,omitempty"`
}

// UpdateLeaseSettingsSection is an extended version of LeaseSettingsSection
// with additional fields for update
type UpdateLeaseSettingsSection struct {
	XMLName                   xml.Name `xml:"LeaseSettingsSection"`
	XmlnsOvf                  string   `xml:"xmlns:ovf,attr,omitempty"`
	Xmlns                     string   `xml:"xmlns,attr,omitempty"`
	OVFInfo                   string   `xml:"ovf:Info"`
	HREF                      string   `xml:"href,attr,omitempty"`
	Type                      string   `xml:"type,attr,omitempty"`
	DeploymentLeaseExpiration string   `xml:"DeploymentLeaseExpiration,omitempty"`
	DeploymentLeaseInSeconds  *int     `xml:"DeploymentLeaseInSeconds,omitempty"`
	Link                      *Link    `xml:"Link,omitempty"`
	StorageLeaseExpiration    string   `xml:"StorageLeaseExpiration,omitempty"`
	StorageLeaseInSeconds     *int     `xml:"StorageLeaseInSeconds,omitempty"`
}

// IPRange represents a range of IP addresses, start and end inclusive.
// Type: IpRangeType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a range of IP addresses, start and end inclusive.
// Since: 0.9
type IPRange struct {
	StartAddress string `xml:"StartAddress" json:"startAddress,omitempty"` // Start address of the IP range.
	EndAddress   string `xml:"EndAddress" json:"endAddress,omitempty"`     // End address of the IP range.
}

// DhcpService represents a DHCP network service.
// Type: DhcpServiceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a DHCP network service.
// Since:
type DhcpService struct {
	IsEnabled           bool     `xml:"IsEnabled"`                     // Enable or disable the service using this flag
	DefaultLeaseTime    int      `xml:"DefaultLeaseTime,omitempty"`    // Default lease in seconds for DHCP addresses.
	MaxLeaseTime        int      `xml:"MaxLeaseTime"`                  //	Max lease in seconds for DHCP addresses.
	IPRange             *IPRange `xml:"IpRange"`                       //	IP range for DHCP addresses.
	RouterIP            string   `xml:"RouterIp,omitempty"`            // Router IP.
	SubMask             string   `xml:"SubMask,omitempty"`             // The subnet mask.
	PrimaryNameServer   string   `xml:"PrimaryNameServer,omitempty"`   // The primary name server.
	SecondaryNameServer string   `xml:"SecondaryNameServer,omitempty"` // The secondary name server.
	DomainName          string   `xml:"DomainName,omitempty"`          //	The domain name.
}

// NetworkFeatures represents features of a network.
// Type: NetworkFeaturesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents features of a network.
// Since:
type NetworkFeatures struct {
	DhcpService          *DhcpService          `xml:"DhcpService,omitempty"`          // Substitute for NetworkService. DHCP service settings
	FirewallService      *FirewallService      `xml:"FirewallService,omitempty"`      // Substitute for NetworkService. Firewall service settings
	NatService           *NatService           `xml:"NatService,omitempty"`           // Substitute for NetworkService. NAT service settings
	StaticRoutingService *StaticRoutingService `xml:"StaticRoutingService,omitempty"` // Substitute for NetworkService. Static Routing service settings
	// TODO: Not Implemented
	// IpsecVpnService      IpsecVpnService      `xml:"IpsecVpnService,omitempty"`      // Substitute for NetworkService. Ipsec Vpn service settings
}

// IPAddresses a list of IP addresses
// Type: IpAddressesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A list of IP addresses.
// Since: 0.9
type IPAddresses struct {
	IPAddress []string `xml:"IpAddress,omitempty" json:"ipAddress,omitempty"` // A list of IP addresses.
}

// IPRanges represents a list of IP ranges.
// Type: IpRangesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of IP ranges.
// Since: 0.9
type IPRanges struct {
	IPRange []*IPRange `xml:"IpRange,omitempty" json:"ipRange,omitempty"` // IP range.
}

// IPScope specifies network settings like gateway, network mask, DNS servers IP ranges etc
// Type: IpScopeType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Specify network settings like gateway, network mask, DNS servers, IP ranges, etc.
// Since: 0.9
type IPScope struct {
	IsInherited           bool            `xml:"IsInherited" json:"isInherited,omitempty"`                             // True if the IP scope is inherit from parent network.
	Gateway               string          `xml:"Gateway,omitempty" json:"gateway,omitempty"`                           // Gateway of the network.
	Netmask               string          `xml:"Netmask,omitempty" json:"netmask,omitempty"`                           // Network mask.
	SubnetPrefixLength    string          `xml:"SubnetPrefixLength,omitempty"`                                         // Prefix length (as an string, used everywhere).
	SubnetPrefixLengthInt *int            `json:"subnetPrefixLength,omitempty"`                                        // Prefix length (as an int, used in VDC Templates).
	DNS1                  string          `xml:"Dns1,omitempty" json:"dns1,omitempty"`                                 // Primary DNS server.
	DNS2                  string          `xml:"Dns2,omitempty" json:"dns2,omitempty"`                                 // Secondary DNS server.
	DNSSuffix             string          `xml:"DnsSuffix,omitempty" json:"dnsSuffix,omitempty"`                       // DNS suffix.
	IsEnabled             bool            `xml:"IsEnabled,omitempty" json:"isEnabled,omitempty"`                       // Indicates if subnet is enabled or not. Default value is True.
	IPRanges              *IPRanges       `xml:"IpRanges,omitempty" json:"ipRanges,omitempty"`                         // IP ranges used for static pool allocation in the network.
	AllocatedIPAddresses  *IPAddresses    `xml:"AllocatedIpAddresses,omitempty" json:"allocatedIPAddresses,omitempty"` // Read-only list of allocated IP addresses in the network.
	SubAllocations        *SubAllocations `xml:"SubAllocations,omitempty" json:"subAllocations,omitempty"`             // Read-only list of IP addresses that are sub allocated to edge gateways.
}

// SubAllocations a list of IP addresses that are sub allocated to edge gateways.
// Type: SubAllocationsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A list of IP addresses that are sub allocated to edge gateways.
// Since: 5.1
type SubAllocations struct {
	// Attributes
	HREF string `xml:"href,attr,omitempty" json:"href,omitempty"` // The URI of the entity.
	Type string `xml:"type,attr,omitempty" json:"type,omitempty"` // The MIME type of the entity.
	// Elements
	Link          LinkList       `xml:"Link,omitempty" json:"link,omitempty"`                   // A reference to an entity or operation associated with this object.
	SubAllocation *SubAllocation `xml:"SubAllocation,omitempty" json:"subAllocation,omitempty"` // IP Range sub allocated to a edge gateway.
}

// SubAllocation IP range sub allocated to an edge gateway.
// Type: SubAllocationType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: IP range sub allocated to an edge gateway.
// Since: 5.1
type SubAllocation struct {
	EdgeGateway *Reference `xml:"EdgeGateway,omitempty" json:"edgeGateway,omitempty"` // Edge gateway that uses this sub allocation.
	IPRanges    *IPRanges  `xml:"IpRanges,omitempty" json:"ipRanges,omitempty"`       // IP range sub allocated to the edge gateway.
}

// IPScopes represents a list of IP scopes.
// Type: IpScopesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of IP scopes.
// Since: 5.1
type IPScopes struct {
	IPScope []*IPScope `xml:"IpScope" json:"ipScope,omitempty"` // IP scope.
}

// NetworkConfiguration is the configuration applied to a network. This is an abstract base type.
// The concrete types include those for vApp and Organization wide networks.
// Type: NetworkConfigurationType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: The configurations applied to a network. This is an abstract base type. The concrete types include those for vApp and Organization wide networks.
// Since: 0.9
type NetworkConfiguration struct {
	Xmlns                          string           `xml:"xmlns,attr,omitempty"`
	BackwardCompatibilityMode      bool             `xml:"BackwardCompatibilityMode" json:"backwardCompatibilityMode,omitempty"`
	IPScopes                       *IPScopes        `xml:"IpScopes,omitempty" json:"ipScopes,omitempty"`
	ParentNetwork                  *Reference       `xml:"ParentNetwork,omitempty" json:"parentNetwork,omitempty"`
	FenceMode                      string           `xml:"FenceMode" json:"fenceMode,omitempty"`
	RetainNetInfoAcrossDeployments *bool            `xml:"RetainNetInfoAcrossDeployments,omitempty" json:"retainNetInfoAcrossDeployments,omitempty"`
	Features                       *NetworkFeatures `xml:"Features,omitempty" json:"features,omitempty"`

	// SubInterface and DistributedInterface are mutually exclusive
	// When they are both nil, it means the "internal" interface (the default) will be used.
	// When one of them is set, the corresponding interface will be used.
	// They cannot be both set (we'll get an API error if we do).
	SubInterface         *bool `xml:"SubInterface,omitempty"`
	DistributedInterface *bool `xml:"DistributedInterface,omitempty"`
	GuestVlanAllowed     *bool `xml:"GuestVlanAllowed,omitempty"`
	// TODO: Not Implemented
	// RouterInfo                     RouterInfo           `xml:"RouterInfo,omitempty"`
	// SyslogServerSettings           SyslogServerSettings `xml:"SyslogServerSettings,omitempty"`
}

// VAppNetworkConfiguration represents a vApp network configuration
// Used in vApp network configuration actions as part of vApp type,
// VApp.NetworkConfigSection.NetworkConfig or directly as NetworkConfigSection.NetworkConfig for various API calls.
// Type: VAppNetworkConfigurationType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a vApp network configuration.
// Since: 0.9
type VAppNetworkConfiguration struct {
	HREF        string `xml:"href,attr,omitempty"`
	Type        string `xml:"type,attr,omitempty"`
	ID          string `xml:"id,attr,omitempty"`
	NetworkName string `xml:"networkName,attr"`

	Link          *Link                 `xml:"Link,omitempty"`
	Description   string                `xml:"Description,omitempty"`
	Configuration *NetworkConfiguration `xml:"Configuration"`
	IsDeployed    bool                  `xml:"IsDeployed"`
}

// VAppNetwork represents a vApp network configuration
// Used as input PUT /network/{id}
// Type: VAppNetworkType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a vApp network configuration.
// Since: 0.9
type VAppNetwork struct {
	Xmlns    string `xml:"xmlns,attr,omitempty"`
	HREF     string `xml:"href,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	ID       string `xml:"id,attr,omitempty"`
	Name     string `xml:"name,attr"`
	Deployed *bool  `xml:"deployed,attr"` // True if the network is deployed.

	Link          *Link                 `xml:"Link,omitempty"`
	Description   string                `xml:"Description,omitempty"`
	Tasks         *TasksInProgress      `xml:"Tasks,omitempty"`
	Configuration *NetworkConfiguration `xml:"Configuration"`
}

// NetworkConfigSection is container for vApp networks.
// Type: NetworkConfigSectionType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for vApp networks.
// Since: 0.9
type NetworkConfigSection struct {
	// Extends OVF Section_Type
	// FIXME: Fix the OVF section
	XMLName xml.Name `xml:"NetworkConfigSection"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`
	Ovf     string   `xml:"xmlns:ovf,attr,omitempty"`

	Info string `xml:"ovf:Info"`
	//
	HREF          string                     `xml:"href,attr,omitempty"`
	Type          string                     `xml:"type,attr,omitempty"`
	Link          *Link                      `xml:"Link,omitempty"`
	NetworkConfig []VAppNetworkConfiguration `xml:"NetworkConfig,omitempty"`
}

// NetworkNames allows to extract network names
func (n NetworkConfigSection) NetworkNames() []string {
	var list []string
	for _, netConfig := range n.NetworkConfig {
		list = append(list, netConfig.NetworkName)
	}
	return list
}

// NetworkConnection represents a network connection in the virtual machine.
// Type: NetworkConnectionType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a network connection in the virtual machine.
// Since: 0.9
// Note. Order of fields does matter and shuffling it might cause API errors
type NetworkConnection struct {
	Network                          string `xml:"network,attr"`                      // Name of the network to which this NIC is connected.
	NeedsCustomization               bool   `xml:"needsCustomization,attr,omitempty"` // True if this NIC needs customization.
	NetworkConnectionIndex           int    `xml:"NetworkConnectionIndex"`            // Virtual slot number associated with this NIC. First slot number is 0.
	IPAddress                        string `xml:"IpAddress,omitempty"`               // IP address assigned to this NIC.
	IpType                           string `xml:"IpType,omitempty"`                  // IPv4
	SecondaryIpAddress               string `xml:"SecondaryIpAddress,omitempty"`
	SecondaryIpType                  string `xml:"SecondaryIpType,omitempty"` // IPv6
	IsConnected                      bool   `xml:"IsConnected"`               // If the virtual machine is undeployed, this value specifies whether the NIC should be connected upon deployment. If the virtual machine is deployed, this value reports the current status of this NIC's connection, and can be updated to change that connection status.
	MACAddress                       string `xml:"MACAddress,omitempty"`      // MAC address associated with the NIC.
	IPAddressAllocationMode          string `xml:"IpAddressAllocationMode"`   // IP address allocation mode for this connection. One of: POOL (A static IP address is allocated automatically from a pool of addresses.) DHCP (The IP address is obtained from a DHCP service.) MANUAL (The IP address is assigned manually in the IpAddress element.) NONE (No IP addressing mode specified.)
	SecondaryIpAddressAllocationMode string `xml:"SecondaryIpAddressAllocationMode,omitempty"`
	NetworkAdapterType               string `xml:"NetworkAdapterType,omitempty"`
	ExternalIPAddress                string `xml:"ExternalIpAddress,omitempty"` // If the network to which this NIC connects provides NAT services, the external address assigned to this NIC appears here.
}

// NetworkConnectionSection the container for the network connections of this virtual machine.
// Type: NetworkConnectionSectionType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for the network connections of this virtual machine.
// Since: 0.9
type NetworkConnectionSection struct {
	// Extends OVF Section_Type
	// FIXME: Fix the OVF section
	XMLName xml.Name `xml:"NetworkConnectionSection"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`
	Ovf     string   `xml:"xmlns:ovf,attr,omitempty"`

	Info string `xml:"ovf:Info"`
	//
	HREF                          string               `xml:"href,attr,omitempty"`
	Type                          string               `xml:"type,attr,omitempty"`
	PrimaryNetworkConnectionIndex int                  `xml:"PrimaryNetworkConnectionIndex"`
	NetworkConnection             []*NetworkConnection `xml:"NetworkConnection,omitempty"`
	Link                          *Link                `xml:"Link,omitempty"`
}

// InstantiationParams is a container for ovf:Section_Type elements that specify vApp configuration on instantiate, compose, or recompose.
// Type: InstantiationParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for ovf:Section_Type elements that specify vApp configuration on instantiate, compose, or recompose.
// Since: 0.9
type InstantiationParams struct {
	CustomizationSection         *CustomizationSection         `xml:"CustomizationSection,omitempty"`
	DefaultStorageProfileSection *DefaultStorageProfileSection `xml:"DefaultStorageProfileSection,omitempty"`
	GuestCustomizationSection    *GuestCustomizationSection    `xml:"GuestCustomizationSection,omitempty"`
	LeaseSettingsSection         *LeaseSettingsSection         `xml:"LeaseSettingsSection,omitempty"`
	NetworkConfigSection         *NetworkConfigSection         `xml:"NetworkConfigSection,omitempty"`
	NetworkConnectionSection     *NetworkConnectionSection     `xml:"NetworkConnectionSection,omitempty"`
	ProductSection               *ProductSection               `xml:"ProductSection,omitempty"`
	// TODO: Not Implemented
	// SnapshotSection              SnapshotSection              `xml:"SnapshotSection,omitempty"`
}

// OrgVDCNetwork represents an Org VDC network in the vCloud model.
// Type: OrgVdcNetworkType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents an Org VDC network in the vCloud model.
// Since: 5.1
type OrgVDCNetwork struct {
	XMLName         xml.Name              `xml:"OrgVdcNetwork"`
	Xmlns           string                `xml:"xmlns,attr,omitempty"`
	HREF            string                `xml:"href,attr,omitempty" json:"href,omitempty"`
	Type            string                `xml:"type,attr,omitempty" json:"type,omitempty"`
	ID              string                `xml:"id,attr,omitempty" json:"id,omitempty"`
	OperationKey    string                `xml:"operationKey,attr,omitempty" json:"operationKey,omitempty"`
	Name            string                `xml:"name,attr" json:"name"`
	Status          string                `xml:"status,attr,omitempty" json:"status,omitempty"`
	Link            []Link                `xml:"Link,omitempty" json:"link,omitempty"`
	Description     string                `xml:"Description,omitempty" json:"description"`
	Configuration   *NetworkConfiguration `xml:"Configuration,omitempty" json:"configuration,omitempty"`
	EdgeGateway     *Reference            `xml:"EdgeGateway,omitempty" json:"edgeGateway,omitempty"`
	ServiceConfig   *GatewayFeatures      `xml:"ServiceConfig,omitempty" json:"serviceConfig,omitempty"` // Specifies the service configuration for an isolated Org VDC networks
	IsShared        bool                  `xml:"IsShared" json:"isShared,omitempty"`
	VimPortGroupRef []*VimObjectRef       `xml:"VimPortGroupRef,omitempty" json:"vimPortGroupRef,omitempty"` // Needed to set up DHCP inside ServiceConfig
	Tasks           *TasksInProgress      `xml:"Tasks,omitempty" json:"tasks,omitempty"`
}

// SupportedHardwareVersions contains a list of VMware virtual hardware versions supported in this vDC.
// Type: SupportedHardwareVersionsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Contains a list of VMware virtual hardware versions supported in this vDC.
// Since: 1.5
type SupportedHardwareVersions struct {
	SupportedHardwareVersion []Reference `xml:"SupportedHardwareVersion,omitempty"` // A virtual hardware version supported in this vDC.
}

// Capabilities collection of supported hardware capabilities.
// Type: CapabilitiesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Collection of supported hardware capabilities.
// Since: 1.5
type Capabilities struct {
	SupportedHardwareVersions *SupportedHardwareVersions `xml:"SupportedHardwareVersions,omitempty" json:"supportedHardwareVersions,omitempty"` // Read-only list of virtual hardware versions supported by this vDC.
}

// Vdc represents the user view of an organization VDC.
// Type: VdcType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the user view of an organization VDC.
// Since: 0.9
type Vdc struct {
	HREF         string `xml:"href,attr,omitempty"`
	Type         string `xml:"type,attr,omitempty"`
	ID           string `xml:"id,attr,omitempty"`
	OperationKey string `xml:"operationKey,attr,omitempty"`
	Name         string `xml:"name,attr"`
	Status       int    `xml:"status,attr,omitempty"`

	Link                 LinkList             `xml:"Link,omitempty"`
	Description          string               `xml:"Description,omitempty"`
	Tasks                *TasksInProgress     `xml:"Tasks,omitempty"`
	AllocationModel      string               `xml:"AllocationModel"`
	ComputeCapacity      []*ComputeCapacity   `xml:"ComputeCapacity"`
	ResourceEntities     []*ResourceEntities  `xml:"ResourceEntities,omitempty"`
	AvailableNetworks    []*AvailableNetworks `xml:"AvailableNetworks,omitempty"`
	Capabilities         []*Capabilities      `xml:"Capabilities,omitempty"`
	NicQuota             int                  `xml:"NicQuota"`
	NetworkQuota         int                  `xml:"NetworkQuota"`
	UsedNetworkCount     int                  `xml:"UsedNetworkCount,omitempty"`
	VMQuota              int                  `xml:"VmQuota"`
	IsEnabled            bool                 `xml:"IsEnabled"`
	VdcStorageProfiles   *VdcStorageProfiles  `xml:"VdcStorageProfiles"`
	DefaultComputePolicy *Reference           `xml:"DefaultComputePolicy"`
}

// AdminVdc represents the admin view of an organization VDC.
// Type: AdminVdcType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the admin view of an organization VDC.
// Since: 0.9
type AdminVdc struct {
	Xmlns string `xml:"xmlns,attr"`
	Vdc

	VCpuInMhz2               *int64     `xml:"VCpuInMhz2,omitempty"`
	ResourceGuaranteedMemory *float64   `xml:"ResourceGuaranteedMemory,omitempty"`
	ResourceGuaranteedCpu    *float64   `xml:"ResourceGuaranteedCpu,omitempty"`
	VCpuInMhz                *int64     `xml:"VCpuInMhz,omitempty"`
	IsThinProvision          *bool      `xml:"IsThinProvision,omitempty"`
	NetworkPoolReference     *Reference `xml:"NetworkPoolReference,omitempty"`
	ProviderVdcReference     *Reference `xml:"ProviderVdcReference"`

	// ResourcePoolRefs is a read-only field and should be avoided in XML structure for write
	// operations because it breaks on Go marshalling bug https://github.com/golang/go/issues/9519
	ResourcePoolRefs              *VimObjectRefs `xml:"ResourcePoolRefs,omitempty"`
	UsesFastProvisioning          *bool          `xml:"UsesFastProvisioning,omitempty"`
	OverCommitAllowed             bool           `xml:"OverCommitAllowed,omitempty"`
	VmDiscoveryEnabled            bool           `xml:"VmDiscoveryEnabled,omitempty"`
	IsElastic                     *bool          `xml:"IsElastic,omitempty"`                     // Supported from 32.0 for the Flex model
	IncludeMemoryOverhead         *bool          `xml:"IncludeMemoryOverhead,omitempty"`         // Supported from 32.0 for the Flex model
	UniversalNetworkPoolReference *Reference     `xml:"UniversalNetworkPoolReference,omitempty"` // Reference to a universal network pool
}

// ProviderVdc represents a Provider VDC.
// Type: ProviderVdcType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a Provider VDC.
// Since: 0.9
type ProviderVdc struct {
	HREF         string `xml:"href,attr,omitempty" json:"href,omitempty"`
	Type         string `xml:"type,attr,omitempty" json:"type,omitempty"`
	ID           string `xml:"id,attr,omitempty" json:"id,omitempty"`
	OperationKey string `xml:"operationKey,attr,omitempty" json:"operationKey,omitempty"`
	Name         string `xml:"name,attr" json:"name"`
	Status       int    `xml:"status,attr,omitempty" json:"status,omitempty"` // -1 (creation failed), 0 (not ready), 1 (ready), 2 (unknown), 3 (unrecognized)

	AvailableNetworks     *AvailableNetworks       `xml:"AvailableNetworks,omitempty" json:"availableNetworks,omitempty"`         // Read-only list of available networks.
	Capabilities          *Capabilities            `xml:"Capabilities,omitempty" json:"capabilities,omitempty"`                   // Read-only list of virtual hardware versions supported by this Provider VDC.
	ComputeCapacity       *RootComputeCapacity     `xml:"ComputeCapacity,omitempty" json:"computeCapacity,omitempty"`             // Read-only indicator of CPU and memory capacity.
	Description           string                   `xml:"Description,omitempty" json:"description,omitempty"`                     // Optional description.
	IsEnabled             *bool                    `xml:"IsEnabled,omitempty" json:"isEnabled,omitempty"`                         // True if this Provider VDC is enabled and can provide resources to organization VDCs. A Provider VDC is always enabled on creation.
	Link                  *LinkList                `xml:"Link,omitempty" json:"link,omitempty"`                                   // A reference to an entity or operation associated with this object.
	NetworkPoolReferences *NetworkPoolReferences   `xml:"NetworkPoolReferences,omitempty" json:"networkPoolReferences,omitempty"` // Read-only list of network pools used by this Provider VDC.
	StorageProfiles       *ProviderStorageProfiles `xml:"StorageProfiles,omitempty" json:"storageProfiles,omitempty"`             // Container for references to vSphere storage profiles available to this Provider VDC.
	Tasks                 *TasksInProgress         `xml:"Tasks,omitempty" json:"tasks,omitempty"`                                 // A list of queued, running, or recently completed tasks associated with this entity.
}

// VMWProviderVdc represents an extension of ProviderVdc.
// Type: VMWProviderVdcType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents an extension of ProviderVdc.
// Since: 1.0
type VMWProviderVdc struct {
	ProviderVdc

	AvailableUniversalNetworkPool   *Reference         `xml:"AvailableUniversalNetworkPool,omitempty" json:"availableUniversalNetworkPool,omitempty"`     // Selectable universal network reference.
	ComputeProviderScope            string             `xml:"ComputeProviderScope,omitempty" json:"computeProviderScope,omitempty"`                       // The compute provider scope represents the compute fault domain for this provider VDC. This value is a tenant-facing tag that is shown to tenants when viewing fault domains of the child Organization VDCs (for ex. a VDC Group).
	DataStoreRefs                   VimObjectRefs      `xml:"DataStoreRefs" json:"dataStoreRefs"`                                                         // vSphere datastores backing this provider VDC.
	HighestSupportedHardwareVersion string             `xml:"HighestSupportedHardwareVersion,omitempty" json:"highestSupportedHardwareVersion,omitempty"` // The highest virtual hardware version supported by this Provider VDC. If empty or omitted on creation, the system sets it to the highest virtual hardware version supported by all hosts in the primary resource pool. You can modify it when you add more resource pools.
	HostReferences                  *VMWHostReferences `xml:"HostReferences,omitempty" json:"hostReferences,omitempty"`                                   // Shows all hosts which are connected to VC server.
	NsxTManagerReference            *Reference         `xml:"NsxTManagerReference,omitempty" json:"nsxTManagerReference,omitempty"`                       // An optional reference to a registered NSX-T Manager to back networking operations for this provider VDC.
	ResourcePoolRefs                *VimObjectRefs     `xml:"ResourcePoolRefs,omitempty" json:"resourcePoolRefs,omitempty"`                               // Resource pools backing this provider VDC. On create, you must specify a resource pool that is not used by (and is not the child of a resource pool used by) any other provider VDC. On modify, this element is required for schema validation, but its contents cannot be changed.
	VimServer                       []*Reference       `xml:"VimServer,omitempty" json:"vimServer,omitempty"`                                             // The vCenter server that provides the resource pools and datastores. A valid reference is required on create. On modify, this element is required for schema validation, but its contents cannot be changed.
}

// VMWHostReferences represents a list of available hosts.
// Type: VMWHostReferencesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of available hosts.
// Since: 1.0
type VMWHostReferences struct {
	HostReference []*Reference `xml:"HostReference,omitempty" json:"hostReference,omitempty"`
	Link          *Link        `xml:"Link,omitempty" json:"link,omitempty"`
}

// RootComputeCapacity represents compute capacity with units.
// Type: RootComputeCapacityType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents compute capacity with units.
// Since: 0.9
type RootComputeCapacity struct {
	Cpu       *ProviderVdcCapacity `xml:"Cpu" json:"cpu"`
	IsElastic bool                 `xml:"IsElastic,omitempty" json:"isElastic,omitempty"`
	IsHA      bool                 `xml:"IsHA,omitempty" json:"isHA,omitempty"`
	Memory    *ProviderVdcCapacity `xml:"Memory" json:"memory"`
}

// NetworkPoolReferences is a container for references to network pools in this vDC.
// Type: NetworkPoolReferencesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for references to network pools in this vDC.
// Since: 0.9
type NetworkPoolReferences struct {
	NetworkPoolReference []*Reference `xml:"NetworkPoolReference" json:"networkPoolReference"`
}

// ProviderStorageProfiles is a container for references to storage profiles associated with a Provider vDC.
// Type: ProviderVdcStorageProfilesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for references to storage profiles associated with a Provider vDC.
// Since: 0.9
type ProviderStorageProfiles struct {
	ProviderVdcStorageProfile []*Reference `xml:"ProviderVdcStorageProfile" json:"providerVdcStorageProfile,omitempty"`
}

// ProviderVdcCapacity represents resource capacity in a Provider vDC.
// Type: ProviderVdcCapacityType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents resource capacity in a Provider vDC.
// Since: 0.9
type ProviderVdcCapacity struct {
	Allocation int64  `xml:"Allocation,omitempty"`
	Overhead   int64  `xml:"Overhead,omitempty"`
	Reserved   int64  `xml:"Reserved,omitempty"`
	Total      int64  `xml:"Total,omitempty"`
	Units      string `xml:"Units"`
	Used       int64  `xml:"Used,omitempty"`
}

// VdcStorageProfileConfiguration represents the parameters to assign a storage profile in creation of organization vDC.
// Type: VdcStorageProfileParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the parameters to create a storage profile in an organization vDC.
// Since: 5.1
// https://code.vmware.com/apis/220/vcloud#/doc/doc/types/VdcStorageProfileParamsType.html
type VdcStorageProfileConfiguration struct {
	Enabled                   *bool      `xml:"Enabled,omitempty"`
	Units                     string     `xml:"Units"`
	Limit                     int64      `xml:"Limit"`
	Default                   bool       `xml:"Default"`
	ProviderVdcStorageProfile *Reference `xml:"ProviderVdcStorageProfile"`
}

// VdcStorageProfile represents the parameters for fetched storage profile
// Type: VdcStorageProfileParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/VdcStorageProfileType.html
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/71e12563-bc11-4d64-821d-92d30f8fcfa1/7424bf8e-aec2-44ad-be7d-b98feda7bae0/doc/doc/types/AdminVdcStorageProfileType.html
type VdcStorageProfile struct {
	ID                        string                         `xml:"id,attr,omitempty" json:"id,omitempty"`
	Xmlns                     string                         `xml:"xmlns,attr,omitempty"`
	Name                      string                         `xml:"name,attr" json:"name,omitempty"`
	Enabled                   *bool                          `xml:"Enabled,omitempty" json:"enabled,omitempty"`
	Units                     string                         `xml:"Units" json:"units,omitempty"`
	Limit                     int64                          `xml:"Limit" json:"limit,omitempty"`
	Default                   bool                           `xml:"Default" json:"default,omitempty"`
	IopsSettings              *VdcStorageProfileIopsSettings `xml:"IopsSettings" json:"iopsSettings,omitempty"`
	StorageUsedMB             int64                          `xml:"StorageUsedMB" json:"storageUsedMB,omitempty"`
	IopsAllocated             int64                          `xml:"IopsAllocated" json:"iopsAllocated,omitempty"`
	ProviderVdcStorageProfile *Reference                     `xml:"ProviderVdcStorageProfile" json:"providerVdcStorageProfile,omitempty"`
}

// AdminVdcStorageProfile represents the parameters for fetched storage profile
// Type: AdminVdcStorageProfileType
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/71e12563-bc11-4d64-821d-92d30f8fcfa1/7424bf8e-aec2-44ad-be7d-b98feda7bae0/doc/doc/types/AdminVdcStorageProfileType.html
type AdminVdcStorageProfile struct {
	Xmlns                     string                         `xml:"xmlns,attr"`
	Name                      string                         `xml:"name,attr"`
	Enabled                   *bool                          `xml:"Enabled,omitempty"`
	Units                     string                         `xml:"Units"`
	Limit                     int64                          `xml:"Limit"`
	Default                   bool                           `xml:"Default"`
	IopsSettings              *VdcStorageProfileIopsSettings `xml:"IopsSettings"`
	StorageUsedMB             int64                          `xml:"StorageUsedMB"`
	IopsAllocated             int64                          `xml:"IopsAllocated"`
	ProviderVdcStorageProfile *Reference                     `xml:"ProviderVdcStorageProfile"`
}

// VdcStorageProfileIopsSettings represents the parameters for VDC storage profiles Iops settings.
// Type: VdcStorageProfileIopsSettings
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/71e12563-bc11-4d64-821d-92d30f8fcfa1/7424bf8e-aec2-44ad-be7d-b98feda7bae0/doc/doc/types/VdcStorageProfileIopsSettingsType.html
type VdcStorageProfileIopsSettings struct {
	Xmlns                   string `xml:"xmlns,attr"`
	Enabled                 bool   `xml:"Enabled" json:"enabled,omitempty"`
	DiskIopsMax             int64  `xml:"DiskIopsMax" json:"diskIopsMax,omitempty"`
	DiskIopsDefault         int64  `xml:"DiskIopsDefault" json:"diskIopsDefault,omitempty"`
	StorageProfileIopsLimit int64  `xml:"StorageProfileIopsLimit,omitempty" json:"storageProfileIopsLimit,omitempty"`
	DiskIopsPerGbMax        int64  `xml:"DiskIopsPerGbMax" json:"diskIopsPerGbMax,omitempty"`
}

// VdcConfiguration models the payload for creating a VDC.
// Type: CreateVdcParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Parameters for creating an organization VDC
// Since: 5.1
// https://code.vmware.com/apis/220/vcloud#/doc/doc/types/CreateVdcParamsType.html
type VdcConfiguration struct {
	XMLName         xml.Name           `xml:"CreateVdcParams"`
	Xmlns           string             `xml:"xmlns,attr"`
	Name            string             `xml:"name,attr"`
	Description     string             `xml:"Description,omitempty"`
	AllocationModel string             `xml:"AllocationModel"` // Flex supported from 32.0
	ComputeCapacity []*ComputeCapacity `xml:"ComputeCapacity"`
	NicQuota        int                `xml:"NicQuota,omitempty"`
	NetworkQuota    int                `xml:"NetworkQuota,omitempty"`
	VmQuota         int                `xml:"VmQuota,omitempty"`
	IsEnabled       bool               `xml:"IsEnabled,omitempty"`
	// Create uses VdcStorageProfileConfiguration and fetch AdminVdcStorageProfile or VdcStorageProfile
	VdcStorageProfile        []*VdcStorageProfileConfiguration `xml:"VdcStorageProfile"`
	ResourceGuaranteedMemory *float64                          `xml:"ResourceGuaranteedMemory,omitempty"`
	ResourceGuaranteedCpu    *float64                          `xml:"ResourceGuaranteedCpu,omitempty"`
	VCpuInMhz                int64                             `xml:"VCpuInMhz,omitempty"`
	IsThinProvision          bool                              `xml:"IsThinProvision,omitempty"`
	NetworkPoolReference     *Reference                        `xml:"NetworkPoolReference,omitempty"`
	ProviderVdcReference     *Reference                        `xml:"ProviderVdcReference"`
	UsesFastProvisioning     bool                              `xml:"UsesFastProvisioning,omitempty"`
	OverCommitAllowed        bool                              `xml:"OverCommitAllowed,omitempty"`
	VmDiscoveryEnabled       bool                              `xml:"VmDiscoveryEnabled,omitempty"`
	IsElastic                *bool                             `xml:"IsElastic,omitempty"`             // Supported from 32.0 for the Flex model
	IncludeMemoryOverhead    *bool                             `xml:"IncludeMemoryOverhead,omitempty"` // Supported from 32.0 for the Flex model
}

// VMWVdcTemplate references a VDC Template.
// Type: VMWVdcTemplateType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// Description: A reference to a VDC template.
// Since: 5.7
type VMWVdcTemplate struct {
	HREF         string `json:"href,omitempty"`
	Type         string `json:"type,omitempty"`
	ID           string `json:"id,omitempty"`
	OperationKey string `json:"operationKey,omitempty"`
	Name         string `json:"name,omitempty"`

	Link                     LinkList                                  `json:"link,omitempty"`
	Description              string                                    `json:"description,omitempty"`
	Tasks                    *TasksInProgress                          `json:"tasks,omitempty"`
	TenantName               string                                    `json:"tenantName,omitempty"`
	TenantDescription        string                                    `json:"tenantDescription,omitempty"`
	NetworkBackingType       string                                    `json:"networkBackingType,omitempty"` // "NSX_V" or "NSX_T"
	ProviderVdcReference     []*VMWVdcTemplateProviderVdcSpecification `json:"providerVdcReference,omitempty"`
	VdcTemplateSpecification *VMWVdcTemplateSpecification              `json:"vdcTemplateSpecification,omitempty"`
}

// VMWVdcTemplateProviderVdcSpecification references a Provider VDC for a VDC Template.
// Type: VMWVdcTemplateProviderVdcSpecificationType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// Since: 5.7
type VMWVdcTemplateProviderVdcSpecification struct {
	HREF string `json:"href,omitempty"`
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	Binding []*VMWVdcTemplateBinding `json:"binding,omitempty"`
}

// VMWVdcTemplateBinding specifies a binding for a VDC Template
// Type: VMWVdcTemplateBindingType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// Description: A Binding pairs a Name element that contains a user-specified identifier in URN format with a Value element
// that contains a reference to an object. The Name can then be used anywhere in the request where a reference
// to that type of object is allowed. For example, when specifying multiple Provider VDCs in a VMWVdcTemplate,
// create a Binding where the Value is a reference to an external network in a candidate Provider VDC, then use
// the Name from that binding in place of the href attribute required by the Network element in the GatewayConfiguration
// of the VdcTemplateSpecification. When the template is instantiated, the Name is replaced by the network reference
// in the Value part of the Binding associated with the Provider VDC that the system selects during instantiation.
// Supported binding values are references to External networks and Edge clusters.
//
// Since: 5.10
type VMWVdcTemplateBinding struct {
	Name  string     `json:"name,omitempty"` // URI format
	Value *Reference `json:"value,omitempty"`
}

// VMWVdcTemplateSpecification references a VDC for a VDC Template.
// Type: VMWVdcTemplateSpecificationType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// Description: A reference to a Provider VDC.
// Since: 5.7
type VMWVdcTemplateSpecification struct {
	Type string `json:"_type,omitempty"`

	// Indicates that the Provider VDC's automatically-created VXLAN network pool should be used.
	// NetworkPoolReference must be empty if this element appears in the request.
	AutomaticNetworkPoolReference *AutomaticNetworkPoolReference `json:"automaticNetworkPoolReference,omitempty"`

	// Maximum number of virtual NICs allowed in this VDC. Defaults to 0, which specifies an unlimited number.
	NicQuota int `json:"nicQuota"`

	// The quota of VMs that can be created in this VDC. Includes VMs in both vApps and vApp templates, deployed, or otherwise.
	// Defaults to 0, which specifies an unlimited number.
	VmQuota int `json:"vmQuota"`

	// Maximum number of network objects that can be deployed in this VDC. Defaults to 0, which means no networks can be deployed.
	ProvisionedNetworkQuota int `json:"provisionedNetworkQuota"`

	// Defines a gateway and NAT Routed organization VDC network to be created.
	GatewayConfiguration *VdcTemplateSpecificationGatewayConfiguration `json:"gatewayConfiguration,omitempty"`

	// A set of name of Storage Profiles, with corresponding limit value, that all Provider VDCs must have, and that are selected at the time of VDC Template instantiation.
	StorageProfile []*VdcStorageProfile `json:"storageProfile,omitempty"`

	// Set to true to indicate if the FLEX VDC is to be elastic. This field can only be set on input for FLEX VDC templates
	// and Allocation VApp VDC templates. However, this field will be returned properly when read.
	IsElastic *bool `json:"isElastic,omitempty"`

	// Set to true to indicate if the FLEX VDC is to include memory overhead into its accounting for admission control.
	// This field can only be set on input for FLEX VDC templates and Allocation VApp VDC templates.
	// However, this field will be returned properly when read.
	IncludeMemoryOverhead *bool `json:"includeMemoryOverhead,omitempty"`

	// Boolean to request thin provisioning. Request will be honored only if the underlying datastore supports it.
	// Thin provisioning saves storage space by committing it on demand. This allows over-allocation of storage.
	ThinProvision bool `json:"thinProvision,omitempty"`

	// Boolean to request fast provisioning. Request will be honored only if the underlying datastore supports it.
	// Fast provisioning can reduce the time it takes to create virtual machines by using vSphere linked clones.
	// If you disable fast provisioning, all provisioning operations will result in full clones.
	FastProvisioningEnabled bool `json:"fastProvisioningEnabled,omitempty"`

	// Reference to a network pool in the Provider VDC. Must be empty if you specify AutomaticNetworkPoolReference.
	NetworkPoolReference        *Reference                 `json:"networkPoolReference,omitempty"`
	NetworkProfileConfiguration *VdcTemplateNetworkProfile `json:"networkProfileConfiguration,omitempty"`

	// Only in Flex VDCs
	CpuAllocationMhz           int `json:"cpuAllocationMhz"`
	CpuLimitMhzPerVcpu         int `json:"cpuLimitMhzPerVcpu"`
	VCpuInMhz                  int `json:"vCpuInMhz"`
	CpuLimitMhz                int `json:"cpuLimitMhz"`
	MemoryAllocationMB         int `json:"memoryAllocationMB"`
	MemoryLimitMb              int `json:"memoryLimitMb"`
	CpuGuaranteedPercentage    int `json:"cpuGuaranteedPercentage"`
	MemoryGuaranteedPercentage int `json:"memoryGuaranteedPercentage"`
}

// AutomaticNetworkPoolReference is an empty struct that states that the Network pool of the Edge Gateway in a VDC Template must
// be chosen automatically.
// Type: AutomaticNetworkPoolReferenceType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
type AutomaticNetworkPoolReference struct {
}

// VdcTemplateSpecificationGatewayConfiguration specifies the Edge Gateway configuration for a VDC Template.
// Type: VdcTemplateSpecificationGatewayConfigurationType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// Description: Defines a gateway and NAT Routed organization VDC network to be created.
// Since: 5.7
type VdcTemplateSpecificationGatewayConfiguration struct {
	// EdgeGateway configuration for the VDC created by this template. The following restrictions apply:
	// * You may not specify a BackwardCompatibilityMode or an EdgeGatewayServiceConfiguration.
	// * The GatewayInterfaces element must contain exactly one GatewayInterface. That GatewayInterface must have an InterfaceType
	//   of uplink and must not contain a SubnetParticipation element.
	Gateway *EdgeGateway `json:"gateway,omitempty"`

	// Org VDC network configuration created by this template. The following restrictions apply:
	// * You may not specify a BackwardCompatibilityMode, EdgeGatewayServiceConfiguration, or NetworkFeatures.
	// * The NetworkConfiguration must specify a FenceMode of natRouted.
	Network *OrgVDCNetwork `json:"network,omitempty"`
}

// VdcTemplateNetworkProfile specifies the network profile for a VDC Template.
// Type: VdcTemplateNetworkProfileType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// Description: Network Profile configuration that is applied to VDC instantiated from a template.
// In NSX_V VDCs Primary and Secondary Edge Clusters can be configured and used for Edge Gateway deployments.
// In NSX_T VDC only Services Edge Cluster can be configured and used for deploying DHCP/VApp services. Binding name from
// the binding names needs to specified as ReferenceType to PrimaryEdgeCluster SecondaryEdgeCluster and ServicesEdgeCluster
// properties. When VDC is instantiated, based on PVDC and binding name appropriate binding value is selected to configure network profiles.
// Since: 35.2
type VdcTemplateNetworkProfile struct {
	PrimaryEdgeCluster   *Reference `json:"primaryEdgeCluster,omitempty"`
	SecondaryEdgeCluster *Reference `json:"secondaryEdgeCluster,omitempty"`
	ServicesEdgeCluster  *Reference `json:"servicesEdgeCluster,omitempty"`
}

// InstantiateVdcTemplateParams specifies the network profile for a VDC Template.
// Type: InstantiateVdcTemplateParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A basic type used to pass arguments to the instantiate VDC template operation, this provides a name and
// optional description for a VDC instantiated from a template.
// Since: 5.7
type InstantiateVdcTemplateParams struct {
	XMLName     xml.Name   `xml:"InstantiateVdcTemplateParams"`
	Xmlns       string     `xml:"xmlns,attr"`
	Name        string     `xml:"name,attr"`
	Source      *Reference `xml:"Source,omitempty"`
	Description string     `xml:"Description,omitempty"`
}

// Task represents an asynchronous operation in VMware Cloud Director.
// Type: TaskType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents an asynchronous operation in VMware Cloud Director.u
// Since: 0.9
// Comments added from https://code.vmware.com/apis/912/vmware-cloud-director/doc/doc/types/TaskType.html
type Task struct {
	HREF             string           `xml:"href,attr,omitempty" json:"HREF,omitempty"`                         // The URI of the entity.
	Type             string           `xml:"type,attr,omitempty" json:"type,omitempty"`                         // The MIME type of the entity.
	ID               string           `xml:"id,attr,omitempty" json:"ID,omitempty"`                             // The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused.
	OperationKey     string           `xml:"operationKey,attr,omitempty" json:"operationKey,omitempty"`         // Optional unique identifier to support idempotent semantics for create and delete operations.
	Name             string           `xml:"name,attr" json:"name,omitempty"`                                   // The name of the entity.
	Status           string           `xml:"status,attr" json:"status,omitempty"`                               // The execution status of the task. One of queued, preRunning, running, success, error, aborted
	Operation        string           `xml:"operation,attr,omitempty" json:"operation,omitempty"`               // A message describing the operation that is tracked by this task.
	OperationName    string           `xml:"operationName,attr,omitempty" json:"operationName,omitempty"`       // The short name of the operation that is tracked by this task.
	ServiceNamespace string           `xml:"serviceNamespace,attr,omitempty" json:"serviceNamespace,omitempty"` // Identifier of the service that created the task. It must not start with com.vmware.vcloud and the length must be between 1 and 128 symbols.
	StartTime        string           `xml:"startTime,attr,omitempty" json:"startTime,omitempty"`               // The date and time the system started executing the task. May not be present if the task has not been executed yet.
	EndTime          string           `xml:"endTime,attr,omitempty" json:"endTime,omitempty"`                   // The date and time that processing of the task was completed. May not be present if the task is still being executed.
	ExpiryTime       string           `xml:"expiryTime,attr,omitempty" json:"expiryTime,omitempty"`             // The date and time at which the task resource will be destroyed and no longer available for retrieval. May not be present if the task has not been executed or is still being executed.
	CancelRequested  bool             `xml:"cancelRequested,attr,omitempty" json:"cancelRequested,omitempty"`   // Whether user has requested this processing to be canceled.
	Link             *LinkList        `xml:"Link,omitempty" json:"link,omitempty"`                              // A reference to an entity or operation associated with this object.
	Description      string           `xml:"Description,omitempty" json:"description,omitempty"`                // Optional description.
	Tasks            *TasksInProgress `xml:"Tasks,omitempty" json:"tasks,omitempty"`                            // A list of queued, running, or recently completed tasks associated with this entity.
	Owner            *Reference       `xml:"Owner,omitempty" json:"owner,omitempty"`                            // Reference to the owner of the task. This is typically the object that the task is creating or updating.
	Error            *Error           `xml:"Error,omitempty" json:"error,omitempty"`                            // Represents error information from a failed task.
	User             *Reference       `xml:"User,omitempty" json:"user,omitempty"`                              // The user who started the task.
	Organization     *Reference       `xml:"Organization,omitempty" json:"organization,omitempty"`              // The organization to which the User belongs.
	Progress         int              `xml:"Progress,omitempty" json:"progress,omitempty"`                      // Read-only indicator of task progress as an approximate percentage between 0 and 100. Not available for all tasks.
	Details          string           `xml:"Details,omitempty" json:"details,omitempty"`                        // Detailed message about the task. Also contained by the Owner entity when task status is preRunning.
	Result           *TaskResult      `xml:"Result,omitempty" json:"result,omitempty"`                          // Result contains additional details that the task may expose
	//
	// TODO: add the following fields
	// Params      anyType        The parameters with which this task was started.
	// VcTaskList  VcTaskListType List of Virtual Center tasks related to this vCD task.
}

// TaskResult contains additional details that the task may expose after finishing
type TaskResult struct {
	ResultContent struct {
		Text string `xml:",chardata"`
		Xsi  string `xml:"xsi,attr"`
		Ns11 string `xml:"ns11,attr"`
		Type string `xml:"type,attr"`
	} `xml:"ResultContent"`
}

// CapacityWithUsage represents a capacity and usage of a given resource.
// Type: CapacityWithUsageType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a capacity and usage of a given resource.
// Since: 0.9
type CapacityWithUsage struct {
	Units     string `xml:"Units"`
	Allocated int64  `xml:"Allocated"`
	Limit     int64  `xml:"Limit"`
	Reserved  int64  `xml:"Reserved,omitempty"`
	Used      int64  `xml:"Used,omitempty"`
}

// ComputeCapacity represents VDC compute capacity.
// Type: ComputeCapacityType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents VDC compute capacity.
// Since: 0.9
type ComputeCapacity struct {
	CPU    *CapacityWithUsage `xml:"Cpu"`
	Memory *CapacityWithUsage `xml:"Memory"`
}

// Reference is a reference to a resource. Contains an href attribute and optional name and type attributes.
// Type: ReferenceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A reference to a resource. Contains an href attribute and optional name and type attributes.
// Since: 0.9
type Reference struct {
	HREF string `xml:"href,attr,omitempty" json:"href,omitempty"`
	ID   string `xml:"id,attr,omitempty" json:"id,omitempty"`
	Type string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Name string `xml:"name,attr,omitempty" json:"name,omitempty"`
}

// ResourceReference represents a reference to a resource. Contains an href attribute, a resource status attribute, and optional name and type attributes.
// Type: ResourceReferenceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a reference to a resource. Contains an href attribute, a resource status attribute, and optional name and type attributes.
// Since: 0.9
type ResourceReference struct {
	HREF   string `xml:"href,attr"`
	ID     string `xml:"id,attr,omitempty"`
	Type   string `xml:"type,attr,omitempty"`
	Name   string `xml:"name,attr,omitempty"`
	Status string `xml:"status,attr,omitempty"`
}

// VdcStorageProfiles is a container for references to storage profiles associated with a vDC.
// Element: VdcStorageProfiles
// Type: VdcStorageProfilesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for references to storage profiles associated with a vDC.
// Since: 5.1
type VdcStorageProfiles struct {
	VdcStorageProfile []*Reference `xml:"VdcStorageProfile,omitempty"`
}

// ResourceEntities is a container for references to ResourceEntity objects in this vDC.
// Type: ResourceEntitiesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for references to ResourceEntity objects in this vDC.
// Since: 0.9
type ResourceEntities struct {
	ResourceEntity []*ResourceReference `xml:"ResourceEntity,omitempty"`
}

// AvailableNetworks is a container for references to available organization vDC networks.
// Type: AvailableNetworksType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for references to available organization vDC networks.
// Since: 0.9
type AvailableNetworks struct {
	Network []*Reference `xml:"Network,omitempty" json:"network,omitempty"`
}

// Link extends reference type by adding relation attribute. Defines a hyper-link with a relationship, hyper-link reference, and an optional MIME type.
// Type: LinkType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Extends reference type by adding relation attribute. Defines a hyper-link with a relationship, hyper-link reference, and an optional MIME type.
// Since: 0.9
type Link struct {
	HREF string `xml:"href,attr" json:"href,omitempty"`
	ID   string `xml:"id,attr,omitempty" json:"id,omitempty"`
	Type string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Name string `xml:"name,attr,omitempty" json:"name,omitempty"`
	Rel  string `xml:"rel,attr" json:"rel,omitempty"`
}

// OrgList represents a lists of Organizations
// Type: OrgType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of VMware Cloud Director organizations.
// Since: 0.9
type OrgList struct {
	Link LinkList `xml:"Link,omitempty"`
	Org  []*Org   `xml:"Org,omitempty"`
}

// Org represents the user view of a VMware Cloud Director organization.
// Type: OrgType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the user view of a VMware Cloud Director organization.
// Since: 0.9
type Org struct {
	HREF         string           `xml:"href,attr,omitempty"`
	Type         string           `xml:"type,attr,omitempty"`
	ID           string           `xml:"id,attr,omitempty"`
	OperationKey string           `xml:"operationKey,attr,omitempty"`
	Name         string           `xml:"name,attr"`
	Description  string           `xml:"Description,omitempty"`
	FullName     string           `xml:"FullName"`
	IsEnabled    bool             `xml:"IsEnabled,omitempty"`
	Link         LinkList         `xml:"Link,omitempty"`
	Tasks        *TasksInProgress `xml:"Tasks,omitempty"`
}

// List of the users within the organization
type OrgUserList struct {
	User []*Reference `xml:"UserReference,omitempty"`
}

type OrgGroupList struct {
	Group []*Reference `xml:"GroupReference,omitempty"`
}

// List of available roles in the organization
type OrgRoleType struct {
	RoleReference []*Reference `xml:"RoleReference,omitempty"`
}

// List of available rights in the organization
type RightsType struct {
	Links          LinkList     `xml:"Link,omitempty"`
	RightReference []*Reference `xml:"RightReference,omitempty"`
}

// AdminOrg represents the admin view of a VMware Cloud Director organization.
// Type: AdminOrgType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the admin view of a VMware Cloud Director organization.
// Since: 0.9
type AdminOrg struct {
	XMLName         xml.Name         `xml:"AdminOrg"`
	Xmlns           string           `xml:"xmlns,attr"`
	HREF            string           `xml:"href,attr,omitempty"`
	Type            string           `xml:"type,attr,omitempty"`
	ID              string           `xml:"id,attr,omitempty"`
	OperationKey    string           `xml:"operationKey,attr,omitempty"`
	Name            string           `xml:"name,attr"`
	Description     string           `xml:"Description,omitempty"`
	FullName        string           `xml:"FullName"`
	IsEnabled       bool             `xml:"IsEnabled,omitempty"`
	Link            LinkList         `xml:"Link,omitempty"`
	Tasks           *TasksInProgress `xml:"Tasks,omitempty"`
	Users           *OrgUserList     `xml:"Users,omitempty"`
	Groups          *OrgGroupList    `xml:"Groups,omitempty"`
	Catalogs        *CatalogsList    `xml:"Catalogs,omitempty"`
	OrgSettings     *OrgSettings     `xml:"Settings,omitempty"`
	Vdcs            *VDCList         `xml:"Vdcs,omitempty"`
	Networks        *NetworksList    `xml:"Networks,omitempty"`
	RightReferences *OrgRoleType     `xml:"RightReferences,omitempty"`
	RoleReferences  *OrgRoleType     `xml:"RoleReferences,omitempty"`
}

// OrgSettingsType represents the settings for a VMware Cloud Director organization.
// Type: OrgSettingsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the settings of a VMware Cloud Director organization.
// Since: 0.9
type OrgSettings struct {
	//attributes
	HREF string `xml:"href,attr,omitempty"` // The URI of the entity.
	Type string `xml:"type,attr,omitempty"` // The MIME type of the entity.
	//elements
	Link                      LinkList                   `xml:"Link,omitempty"`               // A reference to an entity or operation associated with this object.
	OrgGeneralSettings        *OrgGeneralSettings        `xml:"OrgGeneralSettings,omitempty"` // General Settings for the org, not-required
	OrgVAppLeaseSettings      *VAppLeaseSettings         `xml:"VAppLeaseSettings,omitempty"`
	OrgVAppTemplateSettings   *VAppTemplateLeaseSettings `xml:"VAppTemplateLeaseSettings,omitempty"` // Vapp template lease settings, not required
	OrgLdapSettings           *OrgLdapSettingsType       `xml:"OrgLdapSettings,omitempty"`           // LDAP settings, not-requried, defaults to none
	OrgPasswordPolicySettings *OrgPasswordPolicySettings `xml:"OrgPasswordPolicySettings,omitempty"` // Password policy settings for this organization.
}

// OrgGeneralSettingsType represents the general settings for a VMware Cloud Director organization.
// Type: OrgGeneralSettingsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the user view of a VMware Cloud Director organization.
// Since: 0.9
type OrgGeneralSettings struct {
	HREF string   `xml:"href,attr,omitempty"` // The URI of the entity.
	Type string   `xml:"type,attr,omitempty"` // The MIME type of the entity.
	Link LinkList `xml:"Link,omitempty"`      // A reference to an entity or operation associated with this object.

	CanPublishCatalogs       bool `xml:"CanPublishCatalogs,omitempty"`
	CanPublishExternally     bool `xml:"CanPublishExternally,omitempty"`
	CanSubscribe             bool `xml:"CanSubscribe,omitempty"`
	DeployedVMQuota          int  `xml:"DeployedVMQuota,omitempty"`
	StoredVMQuota            int  `xml:"StoredVmQuota,omitempty"`
	UseServerBootSequence    bool `xml:"UseServerBootSequence,omitempty"`
	DelayAfterPowerOnSeconds int  `xml:"DelayAfterPowerOnSeconds,omitempty"`
}

// VAppTemplateLeaseSettings represents the vapp template lease settings for a VMware Cloud Director organization.
// Type: VAppTemplateLeaseSettingsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the vapp template lease settings of a VMware Cloud Director organization.
// Since: 0.9
type VAppTemplateLeaseSettings struct {
	HREF string   `xml:"href,attr,omitempty"` // The URI of the entity.
	Type string   `xml:"type,attr,omitempty"` // The MIME type of the entity.
	Link LinkList `xml:"Link,omitempty"`      // A reference to an entity or operation associated with this object.

	DeleteOnStorageLeaseExpiration *bool `xml:"DeleteOnStorageLeaseExpiration,omitempty"`
	StorageLeaseSeconds            *int  `xml:"StorageLeaseSeconds,omitempty"`
}

type VAppLeaseSettings struct {
	HREF string   `xml:"href,attr,omitempty"` // The URI of the entity.
	Type string   `xml:"type,attr,omitempty"` // The MIME type of the entity.
	Link LinkList `xml:"Link,omitempty"`      // A reference to an entity or operation associated with this object.

	DeleteOnStorageLeaseExpiration   *bool `xml:"DeleteOnStorageLeaseExpiration,omitempty"`
	DeploymentLeaseSeconds           *int  `xml:"DeploymentLeaseSeconds,omitempty"`
	StorageLeaseSeconds              *int  `xml:"StorageLeaseSeconds,omitempty"`
	PowerOffOnRuntimeLeaseExpiration *bool `xml:"PowerOffOnRuntimeLeaseExpiration,omitempty"`
}

// OrgLdapSettingsType represents the ldap settings for a VMware Cloud Director organization.
// Type: OrgLdapSettingsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the ldap settings of a VMware Cloud Director organization.
// Since: 0.9
type OrgLdapSettingsType struct {
	XMLName xml.Name `xml:"OrgLdapSettings" json:"-"`                  // '-' always ignores the field in JSON: https://pkg.go.dev/encoding/json
	Xmlns   string   `xml:"xmlns,attr,omitempty" json:"-"`             // '-' always ignores the field in JSON: https://pkg.go.dev/encoding/json
	HREF    string   `xml:"href,attr,omitempty" json:"href,omitempty"` // The URI of the entity.
	Type    string   `xml:"type,attr,omitempty" json:"type,omitempty"` // The MIME type of the entity.
	Link    LinkList `xml:"Link,omitempty" json:"link,omitempty"`      // A reference to an entity or operation associated with this object.

	OrgLdapMode           string                 `xml:"OrgLdapMode,omitempty" json:"orgLdapMode,omitempty"`                     // LDAP mode you want
	CustomUsersOu         string                 `xml:"CustomUsersOu,omitempty" json:"customUsersOu,omitempty"`                 // If OrgLdapMode is SYSTEM, specifies an LDAP attribute=value pair to use for OU (organizational unit).
	CustomOrgLdapSettings *CustomOrgLdapSettings `xml:"CustomOrgLdapSettings,omitempty" json:"customOrgLdapSettings,omitempty"` // Needs to be set if user chooses custom mode
}

// CustomOrgLdapSettings represents the custom ldap settings for a VMware Cloud Director organization.
// Type: CustomOrgLdapSettingsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the custom ldap settings of a VMware Cloud Director organization.
// Since: 0.9
// Note. Order of these fields matter and API will error if it is changed
type CustomOrgLdapSettings struct {
	HREF string   `xml:"href,attr,omitempty" json:"-"` // The URI of the entity. '-' always ignores the field in JSON: https://pkg.go.dev/encoding/json
	Type string   `xml:"type,attr,omitempty" json:"-"` // The MIME type of the entity. '-' always ignores the field in JSON: https://pkg.go.dev/encoding/json
	Link LinkList `xml:"Link,omitempty" json:"-"`      // A reference to an entity or operation associated with this object. '-' always ignores the field in JSON: https://pkg.go.dev/encoding/json

	HostName                 string                  `xml:"HostName,omitempty" json:"hostName,omitempty"`
	Port                     int                     `xml:"Port" json:"port,omitempty"`
	IsSsl                    bool                    `xml:"IsSsl,omitempty" json:"isSsl,omitempty"`
	IsSslAcceptAll           bool                    `xml:"IsSslAcceptAll,omitempty" json:"isSslAcceptAll,omitempty"`
	SearchBase               string                  `xml:"SearchBase,omitempty" json:"searchBase,omitempty"`
	Username                 string                  `xml:"UserName,omitempty" json:"userName,omitempty"`
	Password                 string                  `xml:"Password,omitempty" json:"password,omitempty"`
	AuthenticationMechanism  string                  `xml:"AuthenticationMechanism" json:"authenticationMechanism,omitempty"`
	IsGroupSearchBaseEnabled bool                    `xml:"IsGroupSearchBaseEnabled" json:"isGroupSearchBaseEnabled,omitempty"`
	GroupSearchBase          string                  `xml:"GroupSearchBase,omitempty" json:"groupSearchBase,omitempty"`
	ConnectorType            string                  `xml:"ConnectorType" json:"connectorType,omitempty"`     // Defines LDAP service implementation type
	UserAttributes           *OrgLdapUserAttributes  `xml:"UserAttributes" json:"userAttributes,omitempty"`   // Defines how LDAP attributes are used when importing a user.
	GroupAttributes          *OrgLdapGroupAttributes `xml:"GroupAttributes" json:"groupAttributes,omitempty"` // Defines how LDAP attributes are used when importing a group.
	UseExternalKerberos      bool                    `xml:"UseExternalKerberos" json:"-"`                     // '-' always ignores the field in JSON: https://pkg.go.dev/encoding/json
	CustomUiButtonLabel      *string                 `xml:"CustomUiButtonLabel,omitempty" json:"customUiButtonLabel,omitempty"`
	Realm                    string                  `xml:"Realm,omitempty" json:"-"` // '-' always ignores the field in JSON: https://pkg.go.dev/encoding/json
}

// OrgLdapGroupAttributes	 represents the ldap group attribute settings for a VMware Cloud Director organization.
// Type: OrgLdapGroupAttributesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the ldap group attribute settings of a VMware Cloud Director organization.
// Since: 0.9
// Note. Order of these fields matter and API will error if it is changed
type OrgLdapGroupAttributes struct {
	ObjectClass          string `xml:"ObjectClass" json:"objectClass"`
	ObjectIdentifier     string `xml:"ObjectIdentifier" json:"objectIdentifier"`
	GroupName            string `xml:"GroupName" json:"groupName"`
	Membership           string `xml:"Membership" json:"membership"`
	MembershipIdentifier string `xml:"MembershipIdentifier" json:"membershipIdentifier"`
	BackLinkIdentifier   string `xml:"BackLinkIdentifier,omitempty" json:"backLinkIdentifier"`
}

// OrgLdapUserAttributesType represents the ldap user attribute settings for a VMware Cloud Director organization.
// Type: OrgLdapUserAttributesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the ldap user attribute settings of a VMware Cloud Director organization.
// Since: 0.9
// Note. Order of these fields matter and API will error if it is changed.
type OrgLdapUserAttributes struct {
	ObjectClass               string `xml:"ObjectClass" json:"objectClass"`
	ObjectIdentifier          string `xml:"ObjectIdentifier" json:"objectIdentifier"`
	Username                  string `xml:"UserName,omitempty" json:"userName"`
	Email                     string `xml:"Email" json:"email"`
	FullName                  string `xml:"FullName" json:"fullName"`
	GivenName                 string `xml:"GivenName" json:"givenName"`
	Surname                   string `xml:"Surname" json:"surname"`
	Telephone                 string `xml:"Telephone" json:"telephone"`
	GroupMembershipIdentifier string `xml:"GroupMembershipIdentifier" json:"groupMembershipIdentifier"`
	GroupBackLinkIdentifier   string `xml:"GroupBackLinkIdentifier,omitempty" json:"groupBackLinkIdentifier"`
}

// OrgPasswordPolicySettings represents password policy settings for this organization.
// Type: OrgPasswordPolicySettingsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents password policy settings for this organization.
// Since: 1.5
type OrgPasswordPolicySettings struct {
	Xmlns                         string    `xml:"xmlns,attr,omitempty"`
	HREF                          string    `xml:"href,attr,omitempty"`           // The URI of the entity.
	Type                          string    `xml:"type,attr,omitempty"`           // The MIME type of the entity.
	Link                          *LinkList `xml:"Link,omitempty"`                // A reference to an entity or operation associated with this object
	AccountLockoutEnabled         bool      `xml:"AccountLockoutEnabled"`         // Set to true to enable account lockout for logins to this organization
	InvalidLoginsBeforeLockout    int       `xml:"InvalidLoginsBeforeLockout"`    // Number of invalid login attempts that will trigger account lockout
	AccountLockoutIntervalMinutes int       `xml:"AccountLockoutIntervalMinutes"` // Number of minutes an account that is locked out will remain locked
}

// VDCList contains a list of references to Org VDCs
// Type: VdcListType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of organization vDCs.
// Since: 0.9
type VDCList struct {
	Vdcs []*Reference `xml:"Vdc,omitempty" json:"vdcs,omitempty"`
}

// NetworksListType contains a list of references to Org Networks
// Type: NetworksListType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of organization Networks.
// Since: 0.9
type NetworksList struct {
	Networks []*Reference `xml:"Network,omitempty"`
}

// CatalogsList contains a list of references to Org Catalogs
// Type: CatalogsListType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of organization Catalogs.
// Since: 0.9
type CatalogsList struct {
	Catalog []*Reference `xml:"CatalogReference,omitempty"`
}

// CatalogItem contains a reference to a VappTemplate or Media object and related metadata.
// Type: CatalogItemType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Contains a reference to a VappTemplate or Media object and related metadata.
// Since: 0.9
type CatalogItem struct {
	HREF          string           `xml:"href,attr,omitempty"`
	Type          string           `xml:"type,attr,omitempty"`
	ID            string           `xml:"id,attr,omitempty"`
	OperationKey  string           `xml:"operationKey,attr,omitempty"`
	Name          string           `xml:"name,attr"`
	Size          int64            `xml:"size,attr,omitempty"`
	DateCreated   string           `xml:"DateCreated,omitempty"`
	Description   string           `xml:"Description,omitempty"`
	Entity        *Entity          `xml:"Entity"`
	Link          LinkList         `xml:"Link,omitempty"`
	Tasks         *TasksInProgress `xml:"Tasks,omitempty"`
	VersionNumber int64            `xml:"VersionNumber,omitempty"`
}

// Entity is a basic entity type in the vCloud object model. Includes a name, an optional description, and an optional list of links.
// Type: EntityType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Basic entity type in the vCloud object model. Includes a name, an optional description, and an optional list of links.
// Since: 0.9
type Entity struct {
	HREF         string           `xml:"href,attr,omitempty"`
	Type         string           `xml:"type,attr,omitempty"`
	ID           string           `xml:"id,attr,omitempty"`
	OperationKey string           `xml:"operationKey,attr,omitempty"`
	Name         string           `xml:"name,attr"`
	Description  string           `xml:"Description,omitempty"`
	Link         LinkList         `xml:"Link,omitempty"`
	Tasks        *TasksInProgress `xml:"Tasks,omitempty"`
}

// CatalogItems is a container for references to catalog items.
// Type: CatalogItemsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for references to catalog items.
// Since: 0.9
type CatalogItems struct {
	CatalogItem []*Reference `xml:"CatalogItem"`
}

// Catalog represents the user view of a Catalog object.
// Type: CatalogType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the user view of a Catalog object.
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/CatalogType.html
// Since: 0.9
type Catalog struct {
	HREF                         string                        `xml:"href,attr,omitempty"`
	Type                         string                        `xml:"type,attr,omitempty"`
	ID                           string                        `xml:"id,attr,omitempty"`
	OperationKey                 string                        `xml:"operationKey,attr,omitempty"`
	Name                         string                        `xml:"name,attr"`
	CatalogItems                 []*CatalogItems               `xml:"CatalogItems,omitempty"`
	DateCreated                  string                        `xml:"DateCreated,omitempty"`
	Description                  string                        `xml:"Description,omitempty"`
	IsPublished                  bool                          `xml:"IsPublished,omitempty"`
	Link                         LinkList                      `xml:"Link,omitempty"`
	Owner                        *Owner                        `xml:"Owner,omitempty"`
	Tasks                        *TasksInProgress              `xml:"Tasks,omitempty"`
	VersionNumber                int64                         `xml:"VersionNumber,omitempty"`
	PublishExternalCatalogParams *PublishExternalCatalogParams `xml:"PublishExternalCatalogParams,omitempty"`
}

// AdminCatalog represents the Admin view of a Catalog object.
// Type: AdminCatalogType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the Admin view of a Catalog object.
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/AdminCatalogType.html
// Since: 0.9
type AdminCatalog struct {
	Catalog
	XMLName                      xml.Name                      `xml:"AdminCatalog"`
	Xmlns                        string                        `xml:"xmlns,attr"`
	PublishExternalCatalogParams *PublishExternalCatalogParams `xml:"PublishExternalCatalogParams,omitempty"`
	CatalogStorageProfiles       *CatalogStorageProfiles       `xml:"CatalogStorageProfiles,omitempty"`
	ExternalCatalogSubscription  *ExternalCatalogSubscription  `xml:"ExternalCatalogSubscriptionParams,omitempty"`
	IsPublished                  bool                          `xml:"IsPublished,omitempty"`
}

// PublishExternalCatalogParamsType represents the configuration parameters of a catalog published externally
// Type: PublishExternalCatalogParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the configuration parameters of a catalog published externally.
// Since: 5.5
type PublishExternalCatalogParams struct {
	Xmlns                    string `xml:"xmlns,attr,omitempty"`
	IsPublishedExternally    *bool  `xml:"IsPublishedExternally,omitempty"`    // True enables external publication as configured by these parameters.
	CatalogPublishedUrl      string `xml:"catalogPublishedUrl,omitempty"`      // Read-only endpoint URL supplied by the server. External subscribers can connect to the catalog at this URL.
	Password                 string `xml:"Password,omitempty"`                 // Password required when connecting to the endpoint.
	IsCachedEnabled          *bool  `xml:"IsCacheEnabled,omitempty"`           // True enables content caching for this catalog. All items in the catalog are created and stored in transfer storage. If false, items are not placed in transfer storage until they are requested by a subscriber. Note that access to this attribute is reserved to users with role that includes the right 'Catalog: VCSP Publish Subscribe Caching'.
	PreserveIdentityInfoFlag *bool  `xml:"PreserveIdentityInfoFlag,omitempty"` // True includes BIOS UUIDs and MAC addresses in the downloaded OVF package. If false, those information will be excluded.
}

// PublishCatalogParams represents the configuration parameters of a catalog published to other orgs.
// It is used in conjunction with the "IsPublished" state of the catalog itself
type PublishCatalogParams struct {
	XMLName     xml.Name `xml:"PublishCatalogParams"`
	Xmlns       string   `xml:"xmlns,attr,omitempty"`
	IsPublished *bool    `xml:"IsPublished,omitempty"` // True enables  publication (read-only access)
}

// ExternalCatalogSubscription represents the configuration parameters for a catalog that has an external subscription
// Type: ExternalCatalogSubscriptionParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the configuration parameters for a catalog that has an external subscription.
// Since: 5.5
type ExternalCatalogSubscription struct {
	XMLName                  xml.Name `xml:"ExternalCatalogSubscriptionParams"`
	Xmlns                    string   `xml:"xmlns,attr,omitempty"`
	ExpectedSslThumbprint    string   `xml:"ExpectedSslThumbprint,omitempty"`
	SubscribeToExternalFeeds bool     `xml:"SubscribeToExternalFeeds,omitempty"`
	Location                 string   `xml:"Location,omitempty"`
	Password                 string   `xml:"Password,omitempty"`
	LocalCopy                bool     `xml:"LocalCopy,omitempty"`
}

// CatalogStorageProfiles represents a container for storage profiles used by this catalog
// Type: CatalogStorageProfiles
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a container for storage profiles used by this catalog
// Since: 5.5
type CatalogStorageProfiles struct {
	VdcStorageProfile []*Reference `xml:"VdcStorageProfile,omitempty"`
}

// Owner represents the owner of this entity.
// Type: OwnerType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the owner of this entity.
// Since: 1.5
type Owner struct {
	HREF string     `xml:"href,attr,omitempty"`
	Type string     `xml:"type,attr,omitempty"`
	Link LinkList   `xml:"Link,omitempty"`
	User *Reference `xml:"User"`
}

// Error is the standard error message type used in the vCloud REST API.
// Type: ErrorType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: The standard error message type used in the vCloud REST API.
// Since: 0.9
type Error struct {
	Message                 string `xml:"message,attr" json:"message,omitempty"`
	MajorErrorCode          int    `xml:"majorErrorCode,attr" json:"majorErrorCode,omitempty"`
	MinorErrorCode          string `xml:"minorErrorCode,attr" json:"minorErrorCode,omitempty"`
	VendorSpecificErrorCode string `xml:"vendorSpecificErrorCode,attr,omitempty" json:"vendorSpecificErrorCode,omitempty"`
	StackTrace              string `xml:"stackTrace,attr,omitempty" json:"stackTrace,omitempty"`
}

func (err Error) Error() string {
	return fmt.Sprintf("API Error: %d: %s", err.MajorErrorCode, err.Message)
}

// NSXError is the standard error message type used in the NSX API which is proxied by vCD.
// It has attached method `Error() string` and implements Go's default `type error` interface.
type NSXError struct {
	XMLName    xml.Name `xml:"error"`
	ErrorCode  string   `xml:"errorCode"`
	Details    string   `xml:"details"`
	ModuleName string   `xml:"moduleName"`
}

// Error method implements Go's default `error` interface for NSXError and formats NSX error
// output for human readable output.
func (nsxErr NSXError) Error() string {
	return fmt.Sprintf("%s %s (API error: %s)", nsxErr.ModuleName, nsxErr.Details, nsxErr.ErrorCode)
}

// File represents a file to be transferred (uploaded or downloaded).
// Type: FileType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a file to be transferred (uploaded or downloaded).
// Since: 0.9
type File struct {
	HREF             string           `xml:"href,attr,omitempty"`
	Type             string           `xml:"type,attr,omitempty"`
	ID               string           `xml:"id,attr,omitempty"`
	OperationKey     string           `xml:"operationKey,attr,omitempty"`
	Name             string           `xml:"name,attr"`
	Size             int64            `xml:"size,attr,omitempty"`
	BytesTransferred int64            `xml:"bytesTransferred,attr,omitempty"`
	Checksum         string           `xml:"checksum,attr,omitempty"`
	Description      string           `xml:"Description,omitempty"`
	Link             LinkList         `xml:"Link,omitempty"`
	Tasks            *TasksInProgress `xml:"Tasks,omitempty"`
}

// FilesList represents a list of files to be transferred (uploaded or downloaded).
// Type: FilesListType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of files to be transferred (uploaded or downloaded).
// Since: 0.9
type FilesList struct {
	File []*File `xml:"File"`
}

// UndeployVAppParams parameters to an undeploy vApp request.
// Type: UndeployVAppParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Parameters to an undeploy vApp request.
// Since: 0.9
type UndeployVAppParams struct {
	Xmlns               string `xml:"xmlns,attr"`
	UndeployPowerAction string `xml:"UndeployPowerAction,omitempty"`
}

// VmCapabilities allows you to specify certain capabilities of this virtual machine.
// Type: VmCapabilitiesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Allows you to specify certain capabilities of this virtual machine.
// Since: 5.1
type VmCapabilities struct {
	Xmlns               string   `xml:"xmlns,attr"`
	HREF                string   `xml:"href,attr,omitempty"`
	Type                string   `xml:"type,attr,omitempty"`
	MemoryHotAddEnabled bool     `xml:"MemoryHotAddEnabled,omitempty"`
	CPUHotAddEnabled    bool     `xml:"CpuHotAddEnabled,omitempty"`
	Link                LinkList `xml:"Link,omitempty"`
}

// VMs represents a list of virtual machines.
// Type: VmsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a list of virtual machines.
// Since: 5.1
type VMs struct {
	HREF        string       `xml:"href,attr,omitempty"`
	Type        string       `xml:"type,attr,omitempty"`
	Link        LinkList     `xml:"Link,omitempty"`
	VMReference []*Reference `xml:"VmReference,omitempty"`
}

/*
 * Types that are completely valid (position, comment, coverage complete)
 */

// ComposeVAppParams represents vApp composition parameters
// Type: ComposeVAppParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents vApp composition parameters.
// Since: 0.9
type ComposeVAppParams struct {
	XMLName xml.Name `xml:"ComposeVAppParams"`
	Ovf     string   `xml:"xmlns:ovf,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr"`
	Xmlns   string   `xml:"xmlns,attr"`
	// Attributes
	Name        string `xml:"name,attr,omitempty"`        // Typically used to name or identify the subject of the request. For example, the name of the object being created or modified.
	Deploy      bool   `xml:"deploy,attr"`                // True if the vApp should be deployed at instantiation. Defaults to true.
	PowerOn     bool   `xml:"powerOn,attr"`               // True if the vApp should be powered-on at instantiation. Defaults to true.
	LinkedClone bool   `xml:"linkedClone,attr,omitempty"` // Reserved. Unimplemented.
	// Elements
	Description         string                       `xml:"Description,omitempty"`         // Optional description.
	VAppParent          *Reference                   `xml:"VAppParent,omitempty"`          // Reserved. Unimplemented.
	InstantiationParams *InstantiationParams         `xml:"InstantiationParams,omitempty"` // Instantiation parameters for the composed vApp.
	SourcedItem         *SourcedCompositionItemParam `xml:"SourcedItem,omitempty"`         // Composition item. One of: vApp vAppTemplate VM.
	AllEULAsAccepted    bool                         `xml:"AllEULAsAccepted,omitempty"`    // True confirms acceptance of all EULAs in a vApp template. Instantiation fails if this element is missing, empty, or set to false and one or more EulaSection elements are present.
}

type ReComposeVAppParams struct {
	XMLName xml.Name `xml:"RecomposeVAppParams"`
	Ovf     string   `xml:"xmlns:ovf,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr"`
	Xmlns   string   `xml:"xmlns,attr"`
	// Attributes
	Name        string `xml:"name,attr,omitempty"`        // Typically used to name or identify the subject of the request. For example, the name of the object being created or modified.
	Deploy      bool   `xml:"deploy,attr"`                // True if the vApp should be deployed at instantiation. Defaults to true.
	PowerOn     bool   `xml:"powerOn,attr"`               // True if the vApp should be powered-on at instantiation. Defaults to true.
	LinkedClone bool   `xml:"linkedClone,attr,omitempty"` // Reserved. Unimplemented.
	// Elements
	Description         string                       `xml:"Description,omitempty"`         // Optional description.
	VAppParent          *Reference                   `xml:"VAppParent,omitempty"`          // Reserved. Unimplemented.
	InstantiationParams *InstantiationParams         `xml:"InstantiationParams,omitempty"` // Instantiation parameters for the composed vApp.
	SourcedItem         *SourcedCompositionItemParam `xml:"SourcedItem,omitempty"`         // Composition item. One of: vApp vAppTemplate VM.
	AllEULAsAccepted    bool                         `xml:"AllEULAsAccepted,omitempty"`
	DeleteItem          *DeleteItem                  `xml:"DeleteItem,omitempty"`
}

// SmallRecomposeVappParams is used to update name and description of a vApp
// Using the full definition (ReComposeVAppParams), the description can be changed but not removed
type SmallRecomposeVappParams struct {
	XMLName     xml.Name `xml:"RecomposeVAppParams"`
	Ovf         string   `xml:"xmlns:ovf,attr"`
	Xsi         string   `xml:"xmlns:xsi,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Name        string   `xml:"name,attr"`
	Deploy      bool     `xml:"deploy,attr"`
	Description string   `xml:"Description"`
}

type DeleteItem struct {
	HREF string `xml:"href,attr,omitempty"`
}

// SourcedCompositionItemParam represents a vApp, vApp template or VM to include in a composed vApp.
// Type: SourcedCompositionItemParamType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a vApp, vApp template or VM to include in a composed vApp.
// Since: 0.9
type SourcedCompositionItemParam struct {
	// Attributes
	SourceDelete bool `xml:"sourceDelete,attr,omitempty"` // True if the source item should be deleted after composition is complete.
	// Elements
	Source              *Reference           `xml:"Source"`                        // Reference to a vApp, vApp template or virtual machine to include in the composition. Changing the name of the newly created VM by specifying name attribute is deprecated. Include VmGeneralParams element instead.
	VMGeneralParams     *VMGeneralParams     `xml:"VmGeneralParams,omitempty"`     // Specify name, description, and other properties of a VM during instantiation.
	VAppScopedLocalID   string               `xml:"VAppScopedLocalId,omitempty"`   // If Source references a VM, this value provides a unique identifier for the VM in the scope of the composed vApp.
	InstantiationParams *InstantiationParams `xml:"InstantiationParams,omitempty"` // If Source references a VM this can include any of the following OVF sections: VirtualHardwareSection OperatingSystemSection NetworkConnectionSection GuestCustomizationSection.
	NetworkAssignment   []*NetworkAssignment `xml:"NetworkAssignment,omitempty"`   // If Source references a VM, this element maps a network name specified in the VM to the network name of a vApp network defined in the composed vApp.
	StorageProfile      *Reference           `xml:"StorageProfile,omitempty"`      // If Source references a VM, this element contains a reference to a storage profile to be used for the VM. The specified storage profile must exist in the organization vDC that contains the composed vApp. If not specified, the default storage profile for the vDC is used.
	LocalityParams      *LocalityParams      `xml:"LocalityParams,omitempty"`      // Represents locality parameters. Locality parameters provide a hint that may help the placement engine optimize placement of a VM and an independent a Disk so that the VM can make efficient use of the disk.
	ComputePolicy       *ComputePolicy       `xml:"ComputePolicy,omitempty"`       // accessible only from version API 33.0
}

// LocalityParams represents locality parameters. Locality parameters provide a hint that may help the placement engine optimize placement of a VM with respect to another VM or an independent disk.
// Type: LocalityParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents locality parameters. Locality parameters provide a hint that may help the placement engine optimize placement of a VM with respect to another VM or an independent disk.
// Since: 5.1
type LocalityParams struct {
	// Elements
	ResourceEntity *Reference `xml:"ResourceEntity,omitempty"` // Reference to a Disk, or a VM.
}

// NetworkAssignment maps a network name specified in a VM to the network name of a vApp network defined in the VApp that contains the VM
// Type: NetworkAssignmentType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Maps a network name specified in a VM to the network name of a vApp network defined in the VApp that contains the VM
// Since: 0.9
type NetworkAssignment struct {
	// Attributes
	InnerNetwork     string `xml:"innerNetwork,attr"`     // Name of the network as specified in the VM.
	ContainerNetwork string `xml:"containerNetwork,attr"` // Name of the vApp network to map to.
}

// VMGeneralParams a set of overrides to source VM properties to apply to target VM during copying.
// Type: VmGeneralParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A set of overrides to source VM properties to apply to target VM during copying.
// Since: 5.6
type VMGeneralParams struct {
	// Elements
	Name               string `xml:"Name,omitempty"`               // Name of VM
	Description        string `xml:"Description,omitempty"`        // VM description
	NeedsCustomization bool   `xml:"NeedsCustomization,omitempty"` // True if this VM needs guest customization
	RegenerateBiosUuid bool   `xml:"RegenerateBiosUuid,omitempty"` // True if BIOS UUID of the virtual machine should be regenerated so that it is unique, and not the same as the source virtual machine's BIOS UUID.
}

// VApp represents a vApp
// Type: VAppType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a vApp.
// Since: 0.9
type VApp struct {
	// Attributes
	HREF                  string `xml:"href,attr,omitempty"`                  // The URI of the entity.
	Type                  string `xml:"type,attr,omitempty"`                  // The MIME type of the entity.
	ID                    string `xml:"id,attr,omitempty"`                    // The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused.
	OperationKey          string `xml:"operationKey,attr,omitempty"`          // Optional unique identifier to support idempotent semantics for create and delete operations.
	Name                  string `xml:"name,attr"`                            // The name of the entity.
	Status                int    `xml:"status,attr,omitempty"`                // Creation status of the resource entity.
	Deployed              bool   `xml:"deployed,attr,omitempty"`              // True if the virtual machine is deployed.
	OvfDescriptorUploaded bool   `xml:"ovfDescriptorUploaded,attr,omitempty"` // Read-only indicator that the OVF descriptor for this vApp has been uploaded.
	// Elements
	Link                 LinkList              `xml:"Link,omitempty"`                 // A reference to an entity or operation associated with this object.
	LeaseSettingsSection *LeaseSettingsSection `xml:"LeaseSettingsSection,omitempty"` // A reference to the lease section of the vApp
	NetworkConfigSection *NetworkConfigSection `xml:"NetworkConfigSection,omitempty"` // Represents vAPP network configuration
	Description          string                `xml:"Description,omitempty"`          // Optional description.
	Tasks                *TasksInProgress      `xml:"Tasks,omitempty"`                // A list of queued, running, or recently completed tasks associated with this entity.
	Files                *FilesList            `xml:"Files,omitempty"`                // Represents a list of files to be transferred (uploaded or downloaded). Each File in the list is part of the ResourceEntity.
	VAppParent           *Reference            `xml:"VAppParent,omitempty"`           // Reserved. Unimplemented.
	// TODO: OVF Sections to be implemented
	// Section OVF_Section `xml:"Section"`
	DateCreated       string          `xml:"DateCreated,omitempty"`       // Creation date/time of the vApp.
	Owner             *Owner          `xml:"Owner,omitempty"`             // vApp owner.
	IsAutoNature      bool            `xml:"autoNature,omitempty"`        // True if the vApp is auto generated with a standalone VM
	InMaintenanceMode bool            `xml:"InMaintenanceMode,omitempty"` // True if this vApp is in maintenance mode. Prevents users from changing vApp metadata.
	Children          *VAppChildren   `xml:"Children,omitempty"`          // Container for virtual machines included in this vApp.
	ProductSection    *ProductSection `xml:"ProductSection,omitempty"`
}

type ProductSectionList struct {
	XMLName        xml.Name        `xml:"ProductSectionList"`
	Ovf            string          `xml:"xmlns:ovf,attr,omitempty"`
	Xmlns          string          `xml:"xmlns,attr"`
	ProductSection *ProductSection `xml:"http://schemas.dmtf.org/ovf/envelope/1 ProductSection,omitempty"`
}

// SortByPropertyKeyName allows to sort ProductSectionList property slice by key name as the API is
// does not always return an ordered slice
func (p *ProductSectionList) SortByPropertyKeyName() {
	sort.SliceStable(p.ProductSection.Property, func(i, j int) bool {
		return p.ProductSection.Property[i].Key < p.ProductSection.Property[j].Key
	})
}

type ProductSection struct {
	Info     string      `xml:"Info,omitempty"`
	Property []*Property `xml:"http://schemas.dmtf.org/ovf/envelope/1 Property,omitempty"`
}

type Property struct {
	Key              string `xml:"http://schemas.dmtf.org/ovf/envelope/1 key,attr,omitempty"`
	Label            string `xml:"http://schemas.dmtf.org/ovf/envelope/1 Label,omitempty"`
	Description      string `xml:"http://schemas.dmtf.org/ovf/envelope/1 Description,omitempty"`
	DefaultValue     string `xml:"http://schemas.dmtf.org/ovf/envelope/1 value,attr"`
	Value            *Value `xml:"http://schemas.dmtf.org/ovf/envelope/1 Value,omitempty"`
	Type             string `xml:"http://schemas.dmtf.org/ovf/envelope/1 type,attr,omitempty"`
	UserConfigurable bool   `xml:"http://schemas.dmtf.org/ovf/envelope/1 userConfigurable,attr"`
}

type Value struct {
	Value string `xml:"http://schemas.dmtf.org/ovf/envelope/1 value,attr,omitempty"`
}

// MetadataValue is the type returned when querying a unique entry of metadata.
// Type: MetadataValueType
// Namespace: http://www.vmware.com/vcloud/v1.5
type MetadataValue struct {
	XMLName    xml.Name            `xml:"MetadataValue"`
	Xsi        string              `xml:"xmlns:xsi,attr"`
	Xmlns      string              `xml:"xmlns,attr"`
	Domain     *MetadataDomainTag  `xml:"Domain,omitempty"`
	TypedValue *MetadataTypedValue `xml:"TypedValue"`
}

// MetadataTypedValue is the content of a metadata entry.
// Type: MetadataTypedValue
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: One of: MetadataStringValue, MetadataNumberValue, MetadataBooleanValue, MetadataDateTimeValue
// Since: 5.1
type MetadataTypedValue struct {
	XsiType string `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Value   string `xml:"Value"`
}

// Deprecated: Use MetadataTypedValue instead
type TypedValue = MetadataTypedValue

// Metadata is the user-defined metadata associated with an object.
// Type: MetadataType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: User-defined metadata associated with an object.
// Since: 1.5
type Metadata struct {
	XMLName       xml.Name         `xml:"Metadata"`
	Xmlns         string           `xml:"xmlns,attr"`
	HREF          string           `xml:"href,attr"`
	Type          string           `xml:"type,attr,omitempty"` // The MIME type of the entity.
	Xsi           string           `xml:"xmlns:xsi,attr"`
	Link          []*Link          `xml:"Link,omitempty"`
	MetadataEntry []*MetadataEntry `xml:"MetadataEntry,omitempty"`
}

// MetadataEntry is a single metadata entry.
// Type: MetadataEntryType
// Namespace: http://www.vmware.com/vcloud/v1.5
type MetadataEntry struct {
	Xmlns      string              `xml:"xmlns,attr"`
	HREF       string              `xml:"href,attr"`
	Type       string              `xml:"type,attr,omitempty"` // The MIME type of the entity
	Xsi        string              `xml:"xmlns:xsi,attr"`
	Domain     *MetadataDomainTag  `xml:"Domain,omitempty"`
	Key        string              `xml:"Key"`            // An arbitrary key name. Length cannot exceed 256 UTF-8 characters.
	Link       []*Link             `xml:"Link,omitempty"` // A reference to an entity or operation associated with this object.
	TypedValue *MetadataTypedValue `xml:"TypedValue"`
}

// MetadataDomainTag contains both the visibility and the domain of the metadata.
// Type: MetadataDomainTagType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A value of SYSTEM places this MetadataEntry in the SYSTEM domain. Omit or leave empty to place this MetadataEntry in the GENERAL domain.
// Since: 5.1
type MetadataDomainTag struct {
	Visibility string `xml:"visibility,attr"` // One of: PRIVATE (hidden), READONLY, READWRITE (read/write)
	Domain     string `xml:",chardata"`
}

// VAppChildren is a container for virtual machines included in this vApp.
// Type: VAppChildrenType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for virtual machines included in this vApp.
// Since: 0.9
type VAppChildren struct {
	VM []*Vm `xml:"Vm,omitempty"` // Represents a virtual machine.
}

// TasksInProgress is a list of queued, running, or recently completed tasks.
// Type: TasksInProgressType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A list of queued, running, or recently completed tasks.
// Since: 0.9
type TasksInProgress struct {
	// Elements
	Task []*Task `xml:"Task" json:"task"` // A task.
}

// VAppTemplateChildren is a container for virtual machines included in this vApp template.
// Type: VAppTemplateChildrenType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for virtual machines included in this vApp template.
// Since: 0.9
type VAppTemplateChildren struct {
	// Elements
	VM []*VAppTemplate `xml:"Vm"` // Represents a virtual machine in this vApp template.
}

// VAppTemplate represents a vApp template.
// Type: VAppTemplateType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a vApp template.
// Since: 0.9
type VAppTemplate struct {
	// Attributes
	HREF                  string `xml:"href,attr,omitempty"`                  // The URI of the entity.
	Type                  string `xml:"type,attr,omitempty"`                  // The MIME type of the entity.
	ID                    string `xml:"id,attr,omitempty"`                    // The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused.
	OperationKey          string `xml:"operationKey,attr,omitempty"`          // Optional unique identifier to support idempotent semantics for create and delete operations.
	Name                  string `xml:"name,attr"`                            // The name of the entity.
	Status                int    `xml:"status,attr,omitempty"`                // Creation status of the resource entity.
	OvfDescriptorUploaded string `xml:"ovfDescriptorUploaded,attr,omitempty"` // True if the OVF descriptor for this template has been uploaded.
	GoldMaster            bool   `xml:"goldMaster,attr,omitempty"`            // True if this template is a gold master.
	// Elements
	Link                  LinkList              `xml:"Link,omitempty"`                  // A reference to an entity or operation associated with this object.
	Description           string                `xml:"Description,omitempty"`           // Optional description.
	Tasks                 *TasksInProgress      `xml:"Tasks,omitempty"`                 // A list of queued, running, or recently completed tasks associated with this entity.
	Files                 *FilesList            `xml:"Files,omitempty"`                 // Represents a list of files to be transferred (uploaded or downloaded). Each File in the list is part of the ResourceEntity.
	Owner                 *Owner                `xml:"Owner,omitempty"`                 // vAppTemplate owner.
	Children              *VAppTemplateChildren `xml:"Children,omitempty"`              // Container for virtual machines included in this vApp template.
	VAppScopedLocalID     string                `xml:"VAppScopedLocalId"`               // A unique identifier for the VM in the scope of the vApp template.
	DefaultStorageProfile string                `xml:"DefaultStorageProfile,omitempty"` // The name of the storage profile to be used for this object. The named storage profile must exist in the organization vDC that contains the object. If not specified, the default storage profile for the vDC is used.
	DateCreated           string                `xml:"DateCreated,omitempty"`           // Creation date/time of the template.
	// FIXME: Upstream bug? Missing NetworkConfigSection, LeaseSettingSection and
	// CustomizationSection at least, NetworkConnectionSection is required when
	// using ComposeVApp action in the context of a Children VM (still
	// referenced by VAppTemplateType).
	NetworkConfigSection     *NetworkConfigSection     `xml:"NetworkConfigSection,omitempty"`
	NetworkConnectionSection *NetworkConnectionSection `xml:"NetworkConnectionSection,omitempty"`
	LeaseSettingsSection     *LeaseSettingsSection     `xml:"LeaseSettingsSection,omitempty"`
	CustomizationSection     *CustomizationSection     `xml:"CustomizationSection,omitempty"`
	ProductSection           *ProductSection           `xml:"ProductSection,omitempty"`
	// OVF Section needs to be added
	// Section               Section              `xml:"Section,omitempty"`
}

// VAppTemplateForUpdate represents a vApp template.
// It is shrunken version of VAppTemplateType used for update calls.
// Full VAppTemplateType isn't accepted by API
// Type: VAppTemplateType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a vApp template.
type VAppTemplateForUpdate struct {
	XMLName xml.Name `xml:"VAppTemplate"`
	// Attributes
	Xmlns      string `xml:"xmlns,attr,omitempty"`
	HREF       string `xml:"href,attr,omitempty"`       // The URI of the entity.
	ID         string `xml:"id,attr,omitempty"`         // The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused.
	Name       string `xml:"name,attr"`                 // The name of the entity.
	GoldMaster bool   `xml:"goldMaster,attr,omitempty"` // True if this template is a gold master.
	// Elements
	Link        LinkList `xml:"Link,omitempty"`        // A reference to an entity or operation associated with this object.
	Description string   `xml:"Description,omitempty"` // Optional description.
}

// CaptureVAppParams is a configuration that can be supplied for capturing a vApp template from
// existing vApp
type CaptureVAppParams struct {
	XMLName xml.Name `xml:"CaptureVAppParams"`

	Xmlns    string `xml:"xmlns,attr"`
	XmlnsNs0 string `xml:"xmlns:ns0,attr,omitempty"`

	// Name of vApp template
	Name string `xml:"name,attr"`
	// Description of vApp template
	Description string `xml:"Description,omitempty"`

	// Source vApp reference. At least HREF field must be set
	Source *Reference `xml:"Source"`

	// CustomizationSection section
	CustomizationSection CaptureVAppParamsCustomizationSection `xml:"CustomizationSection"`

	// TargetCatalogItem can be used to overwrite existing item. To overwrite an existing vApp
	// template with the one created by this capture, place a reference to the existing template
	// here. Otherwise, the operation creates a new vApp template.
	TargetCatalogItem *Reference `xml:"TargetCatalogItem,omitempty"`

	// CopyTpmOnInstantiate defines if TPM device is copied (`true`) to instantiated vApp from this
	// template or `false` if a new TPM device is created for instantiated vApp.
	// Note. Supported on VCD 10.4.2+
	CopyTpmOnInstantiate *bool `xml:"CopyTpmOnInstantiate"`
}

// CaptureVAppParamsCustomizationSection settings for CaptureVAppParams type
type CaptureVAppParamsCustomizationSection struct {
	// This field must contain value "CustomizeOnInstantiate Settings" so that API does not reject
	// the request
	Info string `xml:"ns0:Info,omitempty"`
	// CustomizeOnInstantiate marks if instantiating this template applies customization settings
	// (`true`). `false` creates an identical copy.
	CustomizeOnInstantiate bool `xml:"CustomizeOnInstantiate"`
}

// VMDiskChange represents a virtual machine only with Disk setting update part
type VMDiskChange struct {
	XMLName xml.Name `xml:"Vm"`
	Ovf     string   `xml:"xmlns:ovf,attr,omitempty"`
	Xsi     string   `xml:"xmlns:xsi,attr,omitempty"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`

	HREF        string `xml:"href,attr,omitempty"`   // The URI of the VM entity.
	Type        string `xml:"type,attr,omitempty"`   // The MIME type of the entity - application/vnd.vmware.vcloud.vm+xml
	Name        string `xml:"name,attr"`             // VM name
	Description string `xml:"Description,omitempty"` // Optional description.
	ID          string `xml:"id,attr,omitempty"`     // VM ID. The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused.

	VmSpecSection *VmSpecSection `xml:"VmSpecSection,omitempty"` // Container for the specification of this virtual machine. This is an alternative to using ovf:VirtualHardwareSection + ovf:OperatingSystemSection
}

// DiskSection from Vm/VmSpecSection struct
type DiskSection struct {
	DiskSettings []*DiskSettings `xml:"DiskSettings"`
}

// DiskSettings from Vm/VmSpecSection/DiskSection struct
type DiskSettings struct {
	DiskId              string        `xml:"DiskId,omitempty"`              // Specifies a unique identifier for this disk in the scope of the corresponding VM. This element is optional when creating a VM, but if it is provided it should be unique. This element is mandatory when updating an existing disk.
	SizeMb              int64         `xml:"SizeMb"`                        // The size of the disk in MB.
	UnitNumber          int           `xml:"UnitNumber"`                    // The device number on the SCSI or IDE controller of the disk.
	BusNumber           int           `xml:"BusNumber"`                     //	The number of the SCSI or IDE controller itself.
	AdapterType         string        `xml:"AdapterType"`                   // The type of disk controller, e.g. IDE vs SCSI and if SCSI bus-logic vs LSI logic.
	ThinProvisioned     *bool         `xml:"ThinProvisioned,omitempty"`     // Specifies whether the disk storage is pre-allocated or allocated on demand.
	Disk                *Reference    `xml:"Disk,omitempty"`                // Specifies reference to a named disk.
	StorageProfile      *Reference    `xml:"StorageProfile,omitempty"`      // Specifies reference to a storage profile to be associated with the disk.
	OverrideVmDefault   bool          `xml:"overrideVmDefault"`             // Specifies that the disk storage profile overrides the VM's default storage profile.
	IopsAllocation      *IopsResource `xml:"IopsAllocation"`                // IOPS definition for the disk - added in 10.4 in replacement of 'iops'
	VirtualQuantity     *int64        `xml:"VirtualQuantity,omitempty"`     // The actual size of the disk.
	VirtualQuantityUnit string        `xml:"VirtualQuantityUnit,omitempty"` // The units in which VirtualQuantity is measured.
}

type IopsResource struct {
	Reservation int64  `xml:"Reservation"` // The amount of reservation of IOPS on the underlying virtualization infrastructure. This is a read-only.
	Limit       int64  `xml:"Limit"`       // The limit for how much of IOPS can be consumed on the underlying virtualization infrastructure. This is only valid when the resource allocation is not unlimited.
	SharesLevel string `xml:"SharesLevel"` // LOW - NORMAL - HIGH - CUSTOM
	Shares      int64  `xml:"Shares"`      // Custom priority for IOPS. This is a read-only.
}

// MediaSection from Vm/VmSpecSection struct
type MediaSection struct {
	MediaSettings []*MediaSettings `xml:"MediaSettings"`
}

// MediaSettings from Vm/VmSpecSection/MediaSection struct
type MediaSettings struct {
	DeviceId    string     `xml:"DeviceId,omitempty"`    // Describes the media device whose media mount is being specified here. This deviceId must match the RASD.InstanceID attribute in the VirtualHardwareSection of the vApp's OVF description.
	MediaImage  *Reference `xml:"MediaImage,omitempty"`  // The media image that is mounted onto the device. This property can be 'null' which represents that no media is mounted on the device.
	MediaType   string     `xml:"MediaType,omitempty"`   // Specified the type of media that is mounted onto the device.
	MediaState  string     `xml:"MediaState,omitempty"`  // Specifies the state of the media device.
	UnitNumber  int        `xml:"UnitNumber"`            // Specified the type of media that is mounted onto the device.
	BusNumber   int        `xml:"BusNumber"`             //	The bus number of the media device controller.
	AdapterType string     `xml:"AdapterType,omitempty"` // The type of controller, e.g. IDE vs SCSI and if SCSI bus-logic vs LSI logic
}

// CpuResourceMhz from Vm/VmSpecSection struct
type CpuResourceMhz struct {
	Configured  int64  `xml:"Configured"`            // The amount of resource configured on the virtual machine.
	Reservation *int64 `xml:"Reservation,omitempty"` // The amount of reservation of this resource on the underlying virtualization infrastructure.
	Limit       *int64 `xml:"Limit,omitempty"`       // The limit for how much of this resource can be consumed on the underlying virtualization infrastructure. This is only valid when the resource allocation is not unlimited.
	SharesLevel string `xml:"SharesLevel,omitempty"` //	Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload.
	Shares      *int   `xml:"Shares,omitempty"`      // Custom priority for the resource. This field is read-only, unless the shares level is CUSTOM.
}

// MemoryResourceMb from Vm/VmSpecSection struct
type MemoryResourceMb struct {
	Configured  int64  `xml:"Configured"`            // The amount of resource configured on the virtual machine.
	Reservation *int64 `xml:"Reservation,omitempty"` // The amount of reservation of this resource on the underlying virtualization infrastructure.
	Limit       *int64 `xml:"Limit,omitempty"`       // The limit for how much of this resource can be consumed on the underlying virtualization infrastructure. This is only valid when the resource allocation is not unlimited.
	SharesLevel string `xml:"SharesLevel,omitempty"` //	Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload.
	Shares      *int   `xml:"Shares,omitempty"`      // Custom priority for the resource. This is a read-only, unless the share level is CUSTOM.
}

// HardwareVersion from Vm/VmSpecSection struct
type HardwareVersion struct {
	HREF  string `xml:"href,attr"`
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

// ovf:VirtualHardwareSection from Vm struct
type VirtualHardwareSection struct {
	// Extends OVF Section_Type
	XMLName xml.Name `xml:"VirtualHardwareSection"`
	Xmlns   string   `xml:"vcloud,attr,omitempty"`

	Info string                 `xml:"Info"`
	HREF string                 `xml:"href,attr,omitempty"`
	Type string                 `xml:"type,attr,omitempty"`
	Item []*VirtualHardwareItem `xml:"Item,omitempty"`

	ExtraConfig []*VmVirtualHardwareSectionExtraConfig `xml:"ExtraConfig,omitempty"`
	Link        []*Link                                `xml:"Link,omitempty"`
}

type VmVirtualHardwareSectionExtraConfig struct {
	Key      string `xml:"key,attr"`
	Value    string `xml:"value,attr"`
	Required bool   `xml:"required,attr"`
}

// Each ovf:Item parsed from the ovf:VirtualHardwareSection
type VirtualHardwareItem struct {
	XMLName             xml.Name                       `xml:"Item"`
	ResourceType        int                            `xml:"ResourceType,omitempty"`
	ResourceSubType     string                         `xml:"ResourceSubType,omitempty"`
	ElementName         string                         `xml:"ElementName,omitempty"`
	Description         string                         `xml:"Description,omitempty"`
	InstanceID          int                            `xml:"InstanceID,omitempty"`
	AutomaticAllocation bool                           `xml:"AutomaticAllocation,omitempty"`
	Address             string                         `xml:"Address,omitempty"`
	AddressOnParent     int                            `xml:"AddressOnParent,omitempty"`
	AllocationUnits     string                         `xml:"AllocationUnits,omitempty"`
	Reservation         int                            `xml:"Reservation,omitempty"`
	VirtualQuantity     int64                          `xml:"VirtualQuantity,omitempty"`
	Weight              int                            `xml:"Weight,omitempty"`
	CoresPerSocket      int                            `xml:"CoresPerSocket,omitempty"`
	Connection          []*VirtualHardwareConnection   `xml:"Connection,omitempty"`
	HostResource        []*VirtualHardwareHostResource `xml:"HostResource,omitempty"`
	Link                []*Link                        `xml:"Link,omitempty"`
	// Reference: https://code.vmware.com/apis/287/vcloud?h=Director#/doc/doc/operations/GET-DisksRasdItemsList-vApp.html
	Parent int `xml:"Parent,omitempty"`
}

// Connection info from ResourceType=10 (Network Interface)
type VirtualHardwareConnection struct {
	IPAddress         string `xml:"ipAddress,attr,omitempty"`
	PrimaryConnection bool   `xml:"primaryNetworkConnection,attr,omitempty"`
	IpAddressingMode  string `xml:"ipAddressingMode,attr,omitempty"`
	NetworkName       string `xml:",chardata"`
}

// HostResource info from ResourceType=17 (Hard Disk)
// Reference: vCloud API Programming Guide for Service Providers vCloud API 30.0, Page 188 - 189
// https://vdc-download.vmware.com/vmwb-repository/dcr-public/1b6cf07d-adb3-4dba-8c47-9c1c92b04857/
// def8435d-a54a-4923-b26a-e2d1915b09c3/vcloud_sp_api_guide_30_0.pdf
type VirtualHardwareHostResource struct {
	BusType           int    `xml:"busType,attr,omitempty"`
	BusSubType        string `xml:"busSubType,attr,omitempty"`
	Capacity          int    `xml:"capacity,attr,omitempty"`
	StorageProfile    string `xml:"storageProfileHref,attr,omitempty"`
	OverrideVmDefault bool   `xml:"storageProfileOverrideVmDefault,attr,omitempty"`
	Disk              string `xml:"disk,attr,omitempty"`
	//Iops              int    `xml:"iops,attr,omitempty"`
	//OsType            string `xml:"osType,attr,omitempty"`
}

// SnapshotSection from Vm struct
type SnapshotSection struct {
	// Extends OVF Section_Type
	XMLName  xml.Name        `xml:"SnapshotSection"`
	Info     string          `xml:"Info"`
	HREF     string          `xml:"href,attr,omitempty"`
	Type     string          `xml:"type,attr,omitempty"`
	Snapshot []*SnapshotItem `xml:"Snapshot,omitempty"`
}

// Each snapshot listed in the SnapshotSection
type SnapshotItem struct {
	Created   string `xml:"created,attr,omitempty"`
	PoweredOn bool   `xml:"poweredOn,attr,omitempty"`
	Size      int    `xml:"size,attr,omitempty"`
}

// OVFItem is a horrible kludge to process OVF, needs to be fixed with proper types.
type OVFItem struct {
	XMLName         xml.Name `xml:"vcloud:Item"`
	XmlnsRasd       string   `xml:"xmlns:rasd,attr"`
	XmlnsVCloud     string   `xml:"xmlns:vcloud,attr"`
	XmlnsXsi        string   `xml:"xmlns:xsi,attr"`
	XmlnsVmw        string   `xml:"xmlns:vmw,attr,omitempty"`
	VCloudHREF      string   `xml:"vcloud:href,attr"`
	VCloudType      string   `xml:"vcloud:type,attr"`
	AllocationUnits string   `xml:"rasd:AllocationUnits"`
	Description     string   `xml:"rasd:Description"`
	ElementName     string   `xml:"rasd:ElementName"`
	InstanceID      int      `xml:"rasd:InstanceID"`
	Reservation     int      `xml:"rasd:Reservation"`
	ResourceType    int      `xml:"rasd:ResourceType"`
	VirtualQuantity int64    `xml:"rasd:VirtualQuantity"`
	// Weight corresponds to Shares when used for CPU and/or memory settings
	Weight         int   `xml:"rasd:Weight,omitempty"`
	CoresPerSocket *int  `xml:"vmw:CoresPerSocket,omitempty"`
	Link           *Link `xml:"vcloud:Link"`
}

// DeployVAppParams are the parameters to a deploy vApp request
// Type: DeployVAppParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Parameters to a deploy vApp request.
// Since: 0.9
type DeployVAppParams struct {
	XMLName xml.Name `xml:"DeployVAppParams"`
	Xmlns   string   `xml:"xmlns,attr"`
	// Attributes
	PowerOn                bool `xml:"powerOn,attr"`                          // Used to specify whether to power on vapp on deployment, if not set default value is true.
	DeploymentLeaseSeconds int  `xml:"deploymentLeaseSeconds,attr,omitempty"` // Lease in seconds for deployment. A value of 0 is replaced by the organization default deploymentLeaseSeconds value.
	ForceCustomization     bool `xml:"forceCustomization,attr,omitempty"`     // Used to specify whether to force customization on deployment, if not set default value is false
}

// GuestCustomizationStatusSection holds information about guest customization status
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/76f491b4-679c-4e1e-8428-f813d668297a/a2555a1b-22f1-4cca-b481-2a98ab874022/doc/doc/operations/GET-GuestCustStatus.html
type GuestCustomizationStatusSection struct {
	XMLName xml.Name `xml:"GuestCustomizationStatusSection"`
	Xmlns   string   `xml:"xmlns,attr"`

	GuestCustStatus string `xml:"GuestCustStatus"`
}

// GuestCustomizationSection represents guest customization settings
// Type: GuestCustomizationSectionType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a guest customization settings.
// Since: 1.0
type GuestCustomizationSection struct {
	// Extends OVF Section_Type
	// Attributes
	Ovf   string `xml:"xmlns:ovf,attr,omitempty"`
	Xsi   string `xml:"xmlns:xsi,attr,omitempty"`
	Xmlns string `xml:"xmlns,attr,omitempty"`

	HREF string `xml:"href,attr,omitempty"` // A reference to the section in URL format.
	Type string `xml:"type,attr,omitempty"` // The MIME type of the section.
	// FIXME: Fix the OVF section
	Info string `xml:"ovf:Info"`
	// Elements
	Enabled               *bool    `xml:"Enabled,omitempty"`               // True if guest customization is enabled.
	ChangeSid             *bool    `xml:"ChangeSid,omitempty"`             // True if customization can change the Windows SID of this virtual machine.
	VirtualMachineID      string   `xml:"VirtualMachineId,omitempty"`      // Virtual machine ID to apply.
	JoinDomainEnabled     *bool    `xml:"JoinDomainEnabled,omitempty"`     // True if this virtual machine can join a Windows Domain.
	UseOrgSettings        *bool    `xml:"UseOrgSettings,omitempty"`        // True if customization should use organization settings (OrgGuestPersonalizationSettings) when joining a Windows Domain.
	DomainName            string   `xml:"DomainName,omitempty"`            // The name of the Windows Domain to join.
	DomainUserName        string   `xml:"DomainUserName,omitempty"`        // User name to specify when joining a Windows Domain.
	DomainUserPassword    string   `xml:"DomainUserPassword,omitempty"`    // Password to use with DomainUserName.
	MachineObjectOU       string   `xml:"MachineObjectOU,omitempty"`       // The name of the Windows Domain Organizational Unit (OU) in which the computer account for this virtual machine will be created.
	AdminPasswordEnabled  *bool    `xml:"AdminPasswordEnabled,omitempty"`  // True if guest customization can modify administrator password settings for this virtual machine.
	AdminPasswordAuto     *bool    `xml:"AdminPasswordAuto,omitempty"`     // True if the administrator password for this virtual machine should be automatically generated.
	AdminPassword         string   `xml:"AdminPassword,omitempty"`         // True if the administrator password for this virtual machine should be set to this string. (AdminPasswordAuto must be false.)
	AdminAutoLogonEnabled *bool    `xml:"AdminAutoLogonEnabled,omitempty"` // True if guest administrator should automatically log into this virtual machine.
	AdminAutoLogonCount   int      `xml:"AdminAutoLogonCount,omitempty"`   // Number of times administrator can automatically log into this virtual machine. In case AdminAutoLogon is set to True, this value should be between 1 and 100. Otherwise, it should be 0.
	ResetPasswordRequired *bool    `xml:"ResetPasswordRequired,omitempty"` // True if the administrator password for this virtual machine must be reset after first use.
	CustomizationScript   string   `xml:"CustomizationScript,omitempty"`   // Script to run on guest customization. The entire script must appear in this element. Use the XML entity &#13; to represent a newline. Unicode characters can be represented in the form &#xxxx; where xxxx is the character number.
	ComputerName          string   `xml:"ComputerName,omitempty"`          // Computer name to assign to this virtual machine.
	Link                  LinkList `xml:"Link,omitempty"`                  // A link to an operation on this section.
}

// InstantiateVAppTemplateParams represents vApp template instantiation parameters.
// Type: InstantiateVAppTemplateParamsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents vApp template instantiation parameters.
// Since: 0.9
type InstantiateVAppTemplateParams struct {
	XMLName xml.Name `xml:"InstantiateVAppTemplateParams"`
	Ovf     string   `xml:"xmlns:ovf,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr,omitempty"`
	Xmlns   string   `xml:"xmlns,attr"`
	// Attributes
	Name        string `xml:"name,attr,omitempty"`        // Typically used to name or identify the subject of the request. For example, the name of the object being created or modified.
	Deploy      bool   `xml:"deploy,attr"`                // True if the vApp should be deployed at instantiation. Defaults to true.
	PowerOn     bool   `xml:"powerOn,attr"`               // True if the vApp should be powered-on at instantiation. Defaults to true.
	LinkedClone bool   `xml:"linkedClone,attr,omitempty"` // Reserved. Unimplemented.
	// Elements
	Description         string                       `xml:"Description,omitempty"`         // Optional description.
	VAppParent          *Reference                   `xml:"VAppParent,omitempty"`          // Reserved. Unimplemented.
	InstantiationParams *InstantiationParams         `xml:"InstantiationParams,omitempty"` // Instantiation parameters for the composed vApp.
	Source              *Reference                   `xml:"Source"`                        // A reference to a source object such as a vApp or vApp template.
	IsSourceDelete      bool                         `xml:"IsSourceDelete,omitempty"`      // Set to true to delete the source object after the operation completes.
	SourcedItem         *SourcedCompositionItemParam `xml:"SourcedItem,omitempty"`         // Composition item. One of: vApp vAppTemplate VM.
	AllEULAsAccepted    bool                         `xml:"AllEULAsAccepted,omitempty"`    // True confirms acceptance of all EULAs in a vApp template. Instantiation fails if this element is missing, empty, or set to false and one or more EulaSection elements are present.
}

// CloneVAppParams is used to copy one vApp into another
type CloneVAppParams struct {
	XMLName xml.Name `xml:"CloneVAppParams"`
	Ovf     string   `xml:"xmlns:ovf,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr,omitempty"`
	Xmlns   string   `xml:"xmlns,attr"`
	// Attributes
	Name        string `xml:"name,attr,omitempty"`        // Typically used to name or identify the subject of the request. For example, the name of the object being created or modified.
	Deploy      bool   `xml:"deploy,attr"`                // True if the vApp should be deployed at instantiation. Defaults to true.
	PowerOn     bool   `xml:"powerOn,attr"`               // True if the vApp should be powered-on at instantiation. Defaults to true.
	LinkedClone bool   `xml:"linkedClone,attr,omitempty"` // Reserved. Unimplemented.
	// Elements
	Description         string                       `xml:"Description,omitempty"`         // Optional description.
	InstantiationParams *InstantiationParams         `xml:"InstantiationParams,omitempty"` // Instantiation parameters for the composed vApp.
	Source              *Reference                   `xml:"Source"`                        // A reference to a source object such as a vApp or vApp template.
	IsSourceDelete      *bool                        `xml:"IsSourceDelete"`                // Set to true to delete the source object after the operation completes.
	SourcedItem         *SourcedCompositionItemParam `xml:"SourcedItem,omitempty"`         // Composition item. One of: vApp vAppTemplate VM.
}

// EdgeGateway represents a gateway.
// Element: EdgeGateway
// Type: GatewayType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a gateway.
// Since: 5.1
type EdgeGateway struct {
	// Attributes
	Xmlns        string `xml:"xmlns,attr,omitempty"`
	HREF         string `xml:"href,attr,omitempty" json:"href,omitempty"`                 // The URI of the entity.
	Type         string `xml:"type,attr,omitempty" json:"type,omitempty"`                 // The MIME type of the entity.
	ID           string `xml:"id,attr,omitempty" json:"id,omitempty"`                     // The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused
	OperationKey string `xml:"operationKey,attr,omitempty" json:"operationKey,omitempty"` // Optional unique identifier to support idempotent semantics for create and delete operations.
	Name         string `xml:"name,attr" json:"name"`                                     // The name of the entity.
	Status       int    `xml:"status,attr,omitempty" json:"status,omitempty"`             // Creation status of the gateway. One of: 0 (The gateway is still being created) 1 (The gateway is ready) -1 (There was an error while creating the gateway).
	// Elements
	Link          LinkList              `xml:"Link,omitempty" json:"link,omitempty"`         // A link to an operation on this section.
	Description   string                `xml:"Description" json:"description"`               // Description.
	Tasks         *TasksInProgress      `xml:"Tasks,omitempty" json:"tasks,omitempty"`       //	A list of queued, running, or recently completed tasks associated with this entity.
	Configuration *GatewayConfiguration `xml:"Configuration" json:"configuration,omitempty"` // Gateway configuration.
}

// GatewayConfiguration is the gateway configuration
// Type: GatewayConfigurationType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Gateway Configuration.
// Since: 5.1
type GatewayConfiguration struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	// BackwardCompatibilityMode. Default is false. If set to true, will allow users to write firewall
	// rules in the old 1.5 format. The new format does not require to use direction in firewall
	// rules. Also, for firewall rules to allow NAT traffic the filter is applied on the original IP
	// addresses. Once set to true cannot be reverted back to false.
	BackwardCompatibilityMode bool `xml:"BackwardCompatibilityMode,omitempty" json:"backwardCompatibilityMode,omitempty"`
	// GatewayBackingConfig defines configuration of the vShield edge VM for this gateway. One of:
	// compact, full.
	GatewayBackingConfig string `xml:"GatewayBackingConfig" json:"gatewayBackingConfig,omitempty"`
	// GatewayInterfaces holds configuration for edge gateway interfaces, ip allocations, traffic
	// rate limits and ip sub-allocations
	GatewayInterfaces *GatewayInterfaces `xml:"GatewayInterfaces" json:"gatewayInterfaces,omitempty"`
	// EdgeGatewayServiceConfiguration represents Gateway Features.
	EdgeGatewayServiceConfiguration *GatewayFeatures `xml:"EdgeGatewayServiceConfiguration,omitempty" json:"edgeGatewayServiceConfiguration,omitempty"`
	// True if this gateway is highly available. (Requires two vShield edge VMs.)
	HaEnabled *bool `xml:"HaEnabled,omitempty" json:"haEnabled,omitempty"`
	// UseDefaultRouteForDNSRelay defines if the default gateway on the external network selected
	// for default route should be used as the DNS relay.
	UseDefaultRouteForDNSRelay *bool `xml:"UseDefaultRouteForDnsRelay,omitempty" json:"useDefaultRouteForDNSRelay,omitempty"`
	// AdvancedNetworkingEnabled allows to use NSX capabilities such dynamic routing (BGP, OSPF),
	// zero trust networking (DLR), enchanced VPN support (IPsec VPN, SSL VPN-Plus).
	AdvancedNetworkingEnabled *bool `xml:"AdvancedNetworkingEnabled,omitempty" json:"advancedNetworkingEnabled,omitempty"`
	// DistributedRoutingEnabled enables distributed routing on the gateway to allow creation of
	// many more organization VDC networks. Traffic in those networks is optimized for VM-to-VM
	// communication.
	DistributedRoutingEnabled *bool `xml:"DistributedRoutingEnabled,omitempty" json:"distributedRoutingEnabled,omitempty"`
	// FipsModeEnabled allows any secure communication to or from the NSX Edge uses cryptographic
	// algorithms or protocols that are allowed by United States Federal Information Processing
	// Standards (FIPS). FIPS mode turns on the cipher suites that comply with FIPS.
	FipsModeEnabled *bool `xml:"FipsModeEnabled,omitempty" json:"fipsModeEnabled,omitempty"`
	// EdgeClusterConfiguration represents the Edge Cluster Configuration for a given Edge Gateway.
	// Can be changed if a gateway needs to be placed on a specific set of Edge Clusters.
	// For NSX-V Edges, if nothing is specified on create or update, the Org VDC Default will be used.
	// For NSX-T Edges, Open API must be used and this field is read only.
	// If there is no value, the gateway uses the Edge Cluster of the connected External Network's backing Tier-0 router.
	EdgeClusterConfiguration *EdgeClusterConfiguration `xml:"EdgeClusterConfiguration,omitempty" json:"edgeClusterConfiguration,omitempty"`
}

// GatewayInterfaces is a list of Gateway Interfaces.
// Type: GatewayInterfacesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: A list of Gateway Interfaces.
// Since: 5.1
type GatewayInterfaces struct {
	GatewayInterface []*GatewayInterface `xml:"GatewayInterface" json:"gatewayInterface,omitempty"` // Gateway Interface.
}

// GatewayInterface is a gateway interface configuration.
// Type: GatewayInterfaceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Gateway Interface configuration.
// Since: 5.1
type GatewayInterface struct {
	Name                     string                 `xml:"Name,omitempty" json:"name,omitempty"`                                         // Internally generated name for the Gateway Interface.
	DisplayName              string                 `xml:"DisplayName,omitempty" json:"displayName,omitempty"`                           // Gateway Interface display name.
	Network                  *Reference             `xml:"Network" json:"network,omitempty"`                                             // A reference to the network connected to the gateway interface.
	InterfaceType            string                 `xml:"InterfaceType" json:"interfaceType,omitempty"`                                 // The type of interface: One of: Uplink, Internal
	SubnetParticipation      []*SubnetParticipation `xml:"SubnetParticipation,omitempty" json:"subnetParticipation,omitempty"`           // Slice of subnets for IP allocations.
	ApplyRateLimit           bool                   `xml:"ApplyRateLimit,omitempty" json:"applyRateLimit,omitempty"`                     // True if rate limiting is applied on this interface.
	InRateLimit              float64                `xml:"InRateLimit,omitempty" json:"inRateLimit,omitempty"`                           // Incoming rate limit expressed as Gbps.
	OutRateLimit             float64                `xml:"OutRateLimit,omitempty" json:"outRateLimit,omitempty"`                         // Outgoing rate limit expressed as Gbps.
	UseForDefaultRoute       bool                   `xml:"UseForDefaultRoute,omitempty" json:"useForDefaultRoute,omitempty"`             // True if this network is default route for the gateway.
	Connected                bool                   `xml:"Connected,omitempty" json:"connected,omitempty"`                               // True if interface is marked as connected in NSX
	QuickAddAllocatedIpCount int                    `xml:"QuickAddAllocatedIpCount,omitempty" json:"quickAddAllocatedIpCount,omitempty"` // If set on create or update api calls, the specified number of IP addresses will be additionally allocated for this uplink. IPs will be allocated from multiple subnets if needed
}

// SortBySubnetParticipationGateway allows to sort SubnetParticipation property slice by gateway
// address
func (g *GatewayInterface) SortBySubnetParticipationGateway() {
	sort.SliceStable(g.SubnetParticipation, func(i, j int) bool {
		return g.SubnetParticipation[i].Gateway < g.SubnetParticipation[j].Gateway
	})
}

// SubnetParticipation allows to chose which subnets a gateway can be a part of
// Type: SubnetParticipationType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Allows to chose which subnets a gateway can be part of
// Since: 5.1
//
// Note. Field order is important and should not be changed as API returns errors if IPRanges come
// before Gateway and Netmask
type SubnetParticipation struct {
	Gateway            string    `xml:"Gateway" json:"gateway,omitempty"`                                 // Gateway for subnet
	Netmask            string    `xml:"Netmask" json:"netmask,omitempty"`                                 // Netmask for the subnet.
	IPAddress          string    `xml:"IpAddress,omitempty" json:"ipAddress,omitempty"`                   // Ip Address to be assigned. Keep empty or omit element for auto assignment
	IPRanges           *IPRanges `xml:"IpRanges,omitempty" json:"ipRanges,omitempty"`                     // Range of IP addresses available for external interfaces.
	UseForDefaultRoute bool      `xml:"UseForDefaultRoute,omitempty" json:"useForDefaultRoute,omitempty"` // True if this network is default route for the gateway.
}

type EdgeGatewayServiceConfiguration struct {
	XMLName                xml.Name                `xml:"EdgeGatewayServiceConfiguration"`
	Xmlns                  string                  `xml:"xmlns,attr,omitempty"`
	GatewayDhcpService     *GatewayDhcpService     `xml:"GatewayDhcpService,omitempty"`
	FirewallService        *FirewallService        `xml:"FirewallService,omitempty"`
	NatService             *NatService             `xml:"NatService,omitempty"`
	GatewayIpsecVpnService *GatewayIpsecVpnService `xml:"GatewayIpsecVpnService,omitempty"` // Substitute for NetworkService. Gateway Ipsec VPN service settings
}

// GatewayFeatures represents edge gateway services.
// Element: EdgeGatewayServiceConfiguration
// Type: GatewayFeaturesType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents edge gateway services.
// Since: 5.1
type GatewayFeatures struct {
	XMLName                xml.Name
	Xmlns                  string                  `xml:"xmlns,attr,omitempty"`
	FirewallService        *FirewallService        `xml:"FirewallService,omitempty"`        // Substitute for NetworkService. Firewall service settings
	NatService             *NatService             `xml:"NatService,omitempty"`             // Substitute for NetworkService. NAT service settings
	GatewayDhcpService     *GatewayDhcpService     `xml:"GatewayDhcpService,omitempty"`     // Substitute for NetworkService. Gateway DHCP service settings
	GatewayIpsecVpnService *GatewayIpsecVpnService `xml:"GatewayIpsecVpnService,omitempty"` // Substitute for NetworkService. Gateway Ipsec VPN service settings
	StaticRoutingService   *StaticRoutingService   `xml:"StaticRoutingService,omitempty"`   // Substitute for NetworkService. Static Routing service settings
}

// StaticRoutingService represents Static Routing network service.
// Type: StaticRoutingServiceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents Static Routing network service.
// Since: 1.5
type StaticRoutingService struct {
	IsEnabled   bool           `xml:"IsEnabled"`             // Enable or disable the service using this flag
	StaticRoute []*StaticRoute `xml:"StaticRoute,omitempty"` // Details of each Static Route.
}

// StaticRoute represents a static route entry
// Type: StaticRouteType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description:
// Since:
type StaticRoute struct {
	Name             string     `xml:"Name"`                       // Name for the static route.
	Network          string     `xml:"Network"`                    // Network specification in CIDR.
	NextHopIP        string     `xml:"NextHopIp"`                  // IP Address of Next Hop router/gateway.
	Interface        string     `xml:"Interface,omitempty"`        // Interface to use for static routing. Internal and External are the supported values.
	GatewayInterface *Reference `xml:"GatewayInterface,omitempty"` // Gateway interface to which static route is bound.
}

// VendorTemplate is information about a vendor service template. This is optional.
// Type: VendorTemplateType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Information about a vendor service template. This is optional.
// Since: 5.1
type VendorTemplate struct {
	Name string `xml:"Name"` // Name of the vendor template. This is required.
	ID   string `xml:"Id"`   // ID of the vendor template. This is required.
}

// GatewayIpsecVpnService represents gateway IPsec VPN service.
// Type: GatewayIpsecVpnServiceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents gateway IPsec VPN service.
// Since: 5.1
type GatewayIpsecVpnService struct {
	IsEnabled bool                     `xml:"IsEnabled"`          // Enable or disable the service using this flag
	Endpoint  *GatewayIpsecVpnEndpoint `xml:"Endpoint,omitempty"` // List of IPsec VPN Service Endpoints.
	Tunnel    []*GatewayIpsecVpnTunnel `xml:"Tunnel"`             // List of IPsec VPN tunnels.
}

// GatewayIpsecVpnEndpoint represents an IPsec VPN endpoint.
// Type: GatewayIpsecVpnEndpointType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents an IPsec VPN endpoint.
// Since: 5.1
type GatewayIpsecVpnEndpoint struct {
	Network  *Reference `xml:"Network"`            // External network reference.
	PublicIP string     `xml:"PublicIp,omitempty"` // Public IP for IPsec endpoint.
}

// GatewayIpsecVpnTunnel represents an IPsec VPN tunnel.
// Type: GatewayIpsecVpnTunnelType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents an IPsec VPN tunnel.
// Since: 5.1
type GatewayIpsecVpnTunnel struct {
	Name        string `xml:"Name"`                  // The name of the tunnel.
	Description string `xml:"Description,omitempty"` // A description of the tunnel.
	// TODO: Fix this in a better way
	IpsecVpnThirdPartyPeer *IpsecVpnThirdPartyPeer `xml:"IpsecVpnThirdPartyPeer,omitempty"` // Details about the peer network.
	IpsecVpnLocalPeer      *IpsecVpnLocalPeer      `xml:"IpsecVpnLocalPeer"`                // Details about the local peer network.
	PeerIPAddress          string                  `xml:"PeerIpAddress"`                    // IP address of the peer endpoint.
	PeerID                 string                  `xml:"PeerId"`                           // Id for the peer end point
	LocalIPAddress         string                  `xml:"LocalIpAddress"`                   // Address of the local network.
	LocalID                string                  `xml:"LocalId"`                          // Id for local end point
	LocalSubnet            []*IpsecVpnSubnet       `xml:"LocalSubnet"`                      // List of local subnets in the tunnel.
	PeerSubnet             []*IpsecVpnSubnet       `xml:"PeerSubnet"`                       // List of peer subnets in the tunnel.
	SharedSecret           string                  `xml:"SharedSecret"`                     // Shared secret used for authentication.
	SharedSecretEncrypted  bool                    `xml:"SharedSecretEncrypted,omitempty"`  // True if shared secret is encrypted.
	EncryptionProtocol     string                  `xml:"EncryptionProtocol"`               // Encryption protocol to be used. One of: AES, AES256, TRIPLEDES
	Mtu                    int                     `xml:"Mtu"`                              // MTU for the tunnel.
	IsEnabled              bool                    `xml:"IsEnabled,omitempty"`              // True if the tunnel is enabled.
	IsOperational          bool                    `xml:"IsOperational,omitempty"`          // True if the tunnel is operational.
	ErrorDetails           string                  `xml:"ErrorDetails,omitempty"`           // Error details of the tunnel.
}

// IpsecVpnThirdPartyPeer represents details about a peer network
type IpsecVpnThirdPartyPeer struct {
	PeerID string `xml:"PeerId,omitempty"` // Id for the peer end point
}

// IpsecVpnThirdPartyPeer represents details about a peer network
type IpsecVpnLocalPeer struct {
	ID   string `xml:"Id"`   // Id for the peer end point
	Name string `xml:"Name"` // Name for the peer
}

// IpsecVpnSubnet represents subnet details.
// Type: IpsecVpnSubnetType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents subnet details.
// Since: 5.1
type IpsecVpnSubnet struct {
	Name    string `xml:"Name"`    // Gateway Name.
	Gateway string `xml:"Gateway"` // Subnet Gateway.
	Netmask string `xml:"Netmask"` // Subnet Netmask.
}

// GatewayDhcpService represents Gateway DHCP service.
// Type: GatewayDhcpServiceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents Gateway DHCP service.
// Since: 5.1
type GatewayDhcpService struct {
	IsEnabled bool               `xml:"IsEnabled,omitempty"` // Enable or disable the service using this flag
	Pool      []*DhcpPoolService `xml:"Pool,omitempty"`      // A DHCP pool.
}

// EdgeClusterConfiguration configures Edge clusters in an Edge Gateway.
// Type: EdgeClusterConfigurationType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Used for specifying specific Edge Cluster(s) for a given Edge Gateway. Specification is only applicable
// for NSX-V Edges, and if specified this takes precedence over the Edge Cluster configuration on an Org vDC. For NSX-T Edges,
// this is only read-only and edge management must be done via Cloud API.
// Since: 5.1
type EdgeClusterConfiguration struct {
	PrimaryEdgeCluster   *Reference `xml:"PrimaryEdgeCluster,omitempty" json:"primaryEdgeCluster,omitempty"`
	SecondaryEdgeCluster *Reference `xml:"SecondaryEdgeCluster,omitempty" json:"secondaryEdgeCluster,omitempty"`
}

// DhcpPoolService represents DHCP pool service.
// Type: DhcpPoolServiceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents DHCP pool service.
// Since: 5.1
type DhcpPoolService struct {
	IsEnabled        bool       `xml:"IsEnabled,omitempty"`        // True if this DHCP Pool is enabled.
	Network          *Reference `xml:"Network"`                    // Org vDC network to which the DHCP range is applicable.
	DefaultLeaseTime int        `xml:"DefaultLeaseTime,omitempty"` // Default lease period for DHCP range.
	MaxLeaseTime     int        `xml:"MaxLeaseTime"`               // Maximum lease period for DHCP range.
	LowIPAddress     string     `xml:"LowIpAddress"`               // Low IP address in DHCP range.
	HighIPAddress    string     `xml:"HighIpAddress"`              // High IP address in DHCP range.
}

// VMSelection represents details of an vm+nic+iptype selection.
// Type: VmSelectionType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents details of an vm+nic+iptype selection.
// Since: 5.1
type VMSelection struct {
	VAppScopedVMID string `xml:"VAppScopedVmId"` // VAppScopedVmId of VM to which this rule applies.
	VMNicID        int    `xml:"VmNicId"`        // VM NIC ID to which this rule applies.
	IPType         string `xml:"IpType"`         // The value can be one of:- assigned: assigned internal IP be automatically choosen. NAT: NATed external IP will be automatically choosen.
}

// FirewallRuleProtocols flags for a network protocol in a firewall rule
// Type: FirewallRuleType/Protocols
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description:
// Since:
type FirewallRuleProtocols struct {
	ICMP bool `xml:"Icmp,omitempty"` // True if the rule applies to the ICMP protocol.
	Any  bool `xml:"Any,omitempty"`  // True if the rule applies to any protocol.
	TCP  bool `xml:"Tcp,omitempty"`  // True if the rule applies to the TCP protocol.
	UDP  bool `xml:"Udp,omitempty"`  // True if the rule applies to the UDP protocol.
	// FIXME: this is supposed to extend protocol support to all the VSM supported protocols
	// Other string `xml:"Other,omitempty"` //	Any other protocol supported by vShield Manager
}

// FirewallRule represents a firewall rule
// Type: FirewallRuleType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a firewall rule.
// Since: 0.9
type FirewallRule struct {
	ID                   string                 `xml:"Id,omitempty"`                   // Firewall rule identifier.
	IsEnabled            bool                   `xml:"IsEnabled"`                      // Used to enable or disable the firewall rule. Default value is true.
	MatchOnTranslate     bool                   `xml:"MatchOnTranslate"`               // For DNATed traffic, match the firewall rules only after the destination IP is translated.
	Description          string                 `xml:"Description,omitempty"`          // A description of the rule.
	Policy               string                 `xml:"Policy,omitempty"`               // One of: drop (drop packets that match the rule), allow (allow packets that match the rule to pass through the firewall)
	Protocols            *FirewallRuleProtocols `xml:"Protocols,omitempty"`            // Specify the protocols to which the rule should be applied.
	IcmpSubType          string                 `xml:"IcmpSubType,omitempty"`          // ICMP subtype. One of: address-mask-request, address-mask-reply, destination-unreachable, echo-request, echo-reply, parameter-problem, redirect, router-advertisement, router-solicitation, source-quench, time-exceeded, timestamp-request, timestamp-reply, any.
	Port                 int                    `xml:"Port,omitempty"`                 // The port to which this rule applies. A value of -1 matches any port.
	DestinationPortRange string                 `xml:"DestinationPortRange,omitempty"` // Destination port range to which this rule applies.
	DestinationIP        string                 `xml:"DestinationIp,omitempty"`        // Destination IP address to which the rule applies. A value of Any matches any IP address.
	DestinationVM        *VMSelection           `xml:"DestinationVm,omitempty"`        // Details of the destination VM
	SourcePort           int                    `xml:"SourcePort,omitempty"`           // Destination port to which this rule applies. A value of -1 matches any port.
	SourcePortRange      string                 `xml:"SourcePortRange,omitempty"`      // Source port range to which this rule applies.
	SourceIP             string                 `xml:"SourceIp,omitempty"`             // Source IP address to which the rule applies. A value of Any matches any IP address.
	SourceVM             *VMSelection           `xml:"SourceVm,omitempty"`             // Details of the source VM
	Direction            string                 `xml:"Direction,omitempty"`            // Direction of traffic to which rule applies. One of: in (rule applies to incoming traffic. This is the default value), out (rule applies to outgoing traffic).
	EnableLogging        bool                   `xml:"EnableLogging"`                  // Used to enable or disable firewall rule logging. Default value is false.
}

// FirewallService represent a network firewall service.
// Type: FirewallServiceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a network firewall service.
// Since:
type FirewallService struct {
	IsEnabled        bool            `xml:"IsEnabled"`               // Enable or disable the service using this flag
	DefaultAction    string          `xml:"DefaultAction,omitempty"` // Default action of the firewall. One of: drop (Default. Drop packets that match the rule.), allow (Allow packets that match the rule to pass through the firewall)
	LogDefaultAction bool            `xml:"LogDefaultAction"`        // Flag to enable logging for default action. Default value is false.
	FirewallRule     []*FirewallRule `xml:"FirewallRule,omitempty"`  //	A firewall rule.
}

// NatService represents a NAT network service.
// Type: NatServiceType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a NAT network service.
// Since:
type NatService struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	// Elements

	IsEnabled  bool       `xml:"IsEnabled"`            // Enable or disable the service using this flag
	NatType    string     `xml:"NatType,omitempty"`    // One of: ipTranslation (use IP translation), portForwarding (use port forwarding)
	Policy     string     `xml:"Policy,omitempty"`     // One of: allowTraffic (Allow all traffic), allowTrafficIn (Allow inbound traffic only)
	NatRule    []*NatRule `xml:"NatRule,omitempty"`    // A NAT rule.
	ExternalIP string     `xml:"ExternalIp,omitempty"` // External IP address for rule.
}

// NatRule represents a NAT rule.
// Type: NatRuleType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a NAT rule.
// Since: 0.9
type NatRule struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	// Elements
	Description        string                 `xml:"Description,omitempty"`        // A description of the rule.
	RuleType           string                 `xml:"RuleType,omitempty"`           // Type of NAT rule. One of: SNAT (source NAT), DNAT (destination NAT)
	IsEnabled          *bool                  `xml:"IsEnabled"`                    // Used to enable or disable the firewall rule.
	ID                 string                 `xml:"Id,omitempty"`                 // Firewall rule identifier.
	GatewayNatRule     *GatewayNatRule        `xml:"GatewayNatRule,omitempty"`     // Defines SNAT and DNAT types.
	OneToOneBasicRule  *NatOneToOneBasicRule  `xml:"OneToOneBasicRule,omitempty"`  // Maps one internal IP address to one external IP address.
	OneToOneVMRule     *NatOneToOneVMRule     `xml:"OneToOneVmRule,omitempty"`     // Maps one VM NIC to one external IP addresses.
	PortForwardingRule *NatPortForwardingRule `xml:"PortForwardingRule,omitempty"` // Port forwarding internal to external IP addresses.
	VMRule             *NatVMRule             `xml:"VmRule,omitempty"`             // Port forwarding VM NIC to external IP addresses.
}

// GatewayNatRule represents the SNAT and DNAT rules.
// Type: GatewayNatRuleType represents the SNAT and DNAT rules.
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the SNAT and DNAT rules.
// Since: 5.1
type GatewayNatRule struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	// Elements
	Interface      *Reference `xml:"Interface,omitempty"`      // Interface to which rule is applied.
	OriginalIP     string     `xml:"OriginalIp"`               // Original IP for rule.
	OriginalPort   string     `xml:"OriginalPort,omitempty"`   // Original port for rule.
	TranslatedIP   string     `xml:"TranslatedIp"`             // Translated IP for rule.
	TranslatedPort string     `xml:"TranslatedPort,omitempty"` // Translated port for rule.
	Protocol       string     `xml:"Protocol,omitempty"`       // Protocol for rule.
	IcmpSubType    string     `xml:"IcmpSubType,omitempty"`    // ICMP subtype. One of: address-mask-request, address-mask-reply, destination-unreachable, echo-request, echo-reply, parameter-problem, redirect, router-advertisement, router-solicitation, source-quench, time-exceeded, timestamp-request, timestamp-reply, any.
}

// NatOneToOneBasicRule represents the NAT basic rule for one to one mapping of internal and external IP addresses from a network.
// Type: NatOneToOneBasicRuleType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the NAT basic rule for one to one mapping of internal and external IP addresses from a network.
// Since: 0.9
type NatOneToOneBasicRule struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	// Elements
	MappingMode       string `xml:"MappingMode"`       // One of: automatic (map IP addresses automatically), manual (map IP addresses manually using ExternalIpAddress and InternalIpAddress)
	ExternalIPAddress string `xml:"ExternalIpAddress"` // External IP address to map.
	InternalIPAddress string `xml:"InternalIpAddress"` // Internal IP address to map.
}

// NatOneToOneVMRule represents the NAT rule for one to one mapping of VM NIC and external IP addresses from a network.
// Type: NatOneToOneVmRuleType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the NAT rule for one to one mapping of VM NIC and external IP addresses from a network.
// Since: 0.9
type NatOneToOneVMRule struct {
	Xmlns string `xml:"xmlns,attr,omitempty"`
	// Elements
	MappingMode       string  `xml:"MappingMode"`       // Mapping mode.
	ExternalIPAddress *string `xml:"ExternalIpAddress"` // External IP address to map.
	VAppScopedVMID    string  `xml:"VAppScopedVmId"`    // VAppScopedVmId of VM to which this rule applies.
	VMNicID           int     `xml:"VmNicId"`           // VM NIC ID to which this rule applies.
}

// NatPortForwardingRule represents the NAT rule for port forwarding between internal IP/port and external IP/port.
// Type: NatPortForwardingRuleType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the NAT rule for port forwarding between internal IP/port and external IP/port.
// Since: 0.9
type NatPortForwardingRule struct {
	ExternalIPAddress string `xml:"ExternalIpAddress"`  // External IP address to map.
	ExternalPort      int    `xml:"ExternalPort"`       // External port to forward to.
	InternalIPAddress string `xml:"InternalIpAddress"`  // Internal IP address to map.
	InternalPort      int    `xml:"InternalPort"`       // Internal port to forward to.
	Protocol          string `xml:"Protocol,omitempty"` // Protocol to forward. One of: TCP (forward TCP packets), UDP (forward UDP packets), TCP_UDP (forward TCP and UDP packets).
}

// NatVMRule represents the NAT rule for port forwarding between VM NIC/port and external IP/port.
// Type: NatVmRuleType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents the NAT rule for port forwarding between VM NIC/port and external IP/port.
// Since: 0.9
type NatVMRule struct {
	ExternalIPAddress string `xml:"ExternalIpAddress,omitempty"` // External IP address to map.
	ExternalPort      int    `xml:"ExternalPort"`                // External port to forward to.
	VAppScopedVMID    string `xml:"VAppScopedVmId"`              // VAppScopedVmId of VM to which this rule applies.
	VMNicID           int    `xml:"VmNicId"`                     // VM NIC ID to which this rule applies.
	InternalPort      int    `xml:"InternalPort"`                // Internal port to forward to.
	Protocol          string `xml:"Protocol,omitempty"`          // Protocol to forward. One of: TCP (forward TCP packets), UDP (forward UDP packets), TCP_UDP (forward TCP and UDP packets).
}

// QueryResultEdgeGatewayRecordsType is a container for query results in records format.
// Type: QueryResultRecordsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for query results in records format.
// Since: 1.5
type QueryResultEdgeGatewayRecordsType struct {
	// Attributes
	HREF     string  `xml:"href,attr,omitempty"`     // The URI of the entity.
	Type     string  `xml:"type,attr,omitempty"`     // The MIME type of the entity.
	Name     string  `xml:"name,attr,omitempty"`     // The name of the entity.
	Page     int     `xml:"page,attr,omitempty"`     // Page of the result set that this container holds. The first page is page number 1.
	PageSize int     `xml:"pageSize,attr,omitempty"` // Page size, as a number of records or references.
	Total    float64 `xml:"total,attr,omitempty"`    // Total number of records or references in the container.
	// Elements
	Link              []*Link                             `xml:"Link,omitempty"`    // A reference to an entity or operation associated with this object.
	EdgeGatewayRecord []*QueryResultEdgeGatewayRecordType `xml:"EdgeGatewayRecord"` // A record representing a EdgeGateway result.
}

type QueryResultRecordsType struct {
	// Attributes
	HREF     string  `xml:"href,attr,omitempty"`     // The URI of the entity.
	Type     string  `xml:"type,attr,omitempty"`     // The MIME type of the entity.
	Name     string  `xml:"name,attr,omitempty"`     // The name of the entity.
	Page     int     `xml:"page,attr,omitempty"`     // Page of the result set that this container holds. The first page is page number 1.
	PageSize int     `xml:"pageSize,attr,omitempty"` // Page size, as a number of records or references.
	Total    float64 `xml:"total,attr,omitempty"`    // Total number of records or references in the container.
	// Elements
	Link                            []*Link                                           `xml:"Link,omitempty"`                  // A reference to an entity or operation associated with this object.
	EdgeGatewayRecord               []*QueryResultEdgeGatewayRecordType               `xml:"EdgeGatewayRecord"`               // A record representing a EdgeGateway result.
	VMRecord                        []*QueryResultVMRecordType                        `xml:"VMRecord"`                        // A record representing a VM result.
	AdminVMRecord                   []*QueryResultVMRecordType                        `xml:"AdminVMRecord"`                   // A record representing a Admin VM result.
	VAppRecord                      []*QueryResultVAppRecordType                      `xml:"VAppRecord"`                      // A record representing a VApp result.
	AdminVAppRecord                 []*QueryResultVAppRecordType                      `xml:"AdminVAppRecord"`                 // A record representing a VApp result as admin.
	OrgVdcStorageProfileRecord      []*QueryResultOrgVdcStorageProfileRecordType      `xml:"OrgVdcStorageProfileRecord"`      // A record representing storage profiles
	AdminOrgVdcStorageProfileRecord []*QueryResultAdminOrgVdcStorageProfileRecordType `xml:"AdminOrgVdcStorageProfileRecord"` // A record representing storage profiles as admin
	MediaRecord                     []*MediaRecordType                                `xml:"MediaRecord"`                     // A record representing media
	AdminMediaRecord                []*MediaRecordType                                `xml:"AdminMediaRecord"`                // A record representing Admin media
	VMWProviderVdcRecord            []*QueryResultVMWProviderVdcRecordType            `xml:"VMWProviderVdcRecord"`            // A record representing a Provider VDC result.
	ProviderVdcStorageProfileRecord []*QueryResultProviderVdcStorageProfileRecordType `xml:"ProviderVdcStorageProfileRecord"` // A record representing a Provider VDC storage profile result
	NetworkPoolRecord               []*QueryResultNetworkPoolRecordType               `xml:"NetworkPoolRecord"`               // A record representing a network pool
	DiskRecord                      []*DiskRecordType                                 `xml:"DiskRecord"`                      // A record representing a independent Disk.
	AdminDiskRecord                 []*DiskRecordType                                 `xml:"AdminDiskRecord"`                 // A record representing a independent Disk.
	VirtualCenterRecord             []*QueryResultVirtualCenterRecordType             `xml:"VirtualCenterRecord"`             // A record representing a vSphere server
	PortGroupRecord                 []*PortGroupRecordType                            `xml:"PortgroupRecord"`                 // A record representing a port group
	OrgVdcNetworkRecord             []*QueryResultOrgVdcNetworkRecordType             `xml:"OrgVdcNetworkRecord"`             // A record representing a org VDC network
	CatalogRecord                   []*CatalogRecord                                  `xml:"CatalogRecord"`                   // A record representing a catalog
	AdminCatalogRecord              []*CatalogRecord                                  `xml:"AdminCatalogRecord"`              // A record representing an admin catalog
	CatalogItemRecord               []*QueryResultCatalogItemType                     `xml:"CatalogItemRecord"`               // A record representing a catalog item
	AdminCatalogItemRecord          []*QueryResultCatalogItemType                     `xml:"AdminCatalogItemRecord"`          // A record representing an admin catalog item
	VappTemplateRecord              []*QueryResultVappTemplateType                    `xml:"VAppTemplateRecord"`              // A record representing a vApp template
	AdminVappTemplateRecord         []*QueryResultVappTemplateType                    `xml:"AdminVAppTemplateRecord"`         // A record representing an admin vApp template
	NsxtManagerRecord               []*QueryResultNsxtManagerRecordType               `xml:"NsxTManagerRecord"`               // A record representing NSX-T manager
	OrgVdcRecord                    []*QueryResultOrgVdcRecordType                    `xml:"OrgVdcRecord"`                    // A record representing Org VDC
	OrgVdcAdminRecord               []*QueryResultOrgVdcRecordType                    `xml:"AdminVdcRecord"`                  // A record representing Org VDC
	ResourcePoolRecord              []*QueryResultResourcePoolRecordType              `xml:"ResourcePoolRecord"`              // A record representing a Resource Pool
	VmGroupsRecord                  []*QueryResultVmGroupsRecordType                  `xml:"VmGroupsRecord"`                  // A record representing a VM Group
	TaskRecord                      []*QueryResultTaskRecordType                      `xml:"TaskRecord"`                      // A record representing a Task
	AdminTaskRecord                 []*QueryResultTaskRecordType                      `xml:"AdminTaskRecord"`                 // A record representing an Admin Task
	VappNetworkRecord               []*QueryResultVappNetworkRecordType               `xml:"VAppNetworkRecord"`               // A record representing a vApp network
	AdminVappNetworkRecord          []*QueryResultVappNetworkRecordType               `xml:"AdminVAppNetworkRecord"`          // A record representing an admin vApp network
	SiteAssociationRecord           []*QueryResultSiteAssociationRecord               `xml:"SiteAssociationRecord"`           // A record representing a site association
	OrgAssociationRecord            []*QueryResultOrgAssociationRecord                `xml:"OrgAssociationRecord"`            // A record representing an Org association
	OrgRecord                       []*QueryResultOrgRecordType                       `xml:"OrgRecord"`                       // A record representing an Organisation
	AdminOrgVdcTemplateRecord       []*QueryResultAdminOrgVdcTemplateRecordType       `xml:"AdminOrgVdcTemplateRecord"`       // A record representing an admin VDC Template
	OrgVdcTemplateRecord            []*QueryResultOrgVdcTemplateRecordType            `xml:"OrgVdcTemplateRecord"`            // A record representing an VDC Template
}

// QueryResultVmGroupsRecordType represent a VM Groups record
type QueryResultVmGroupsRecordType struct {
	HREF           string `xml:"href,attr,omitempty"`
	ID             string `xml:"vmGroupId,attr,omitempty"`
	Name           string `xml:"vmGroupName,attr,omitempty"`
	ClusterMoref   string `xml:"clusterMoref,attr,omitempty"`
	ClusterName    string `xml:"clusterName,attr,omitempty"`
	VcenterId      string `xml:"vcId,attr,omitempty"`
	NamedVmGroupId string `xml:"namedVmGroupId,attr,omitempty"`
}

// QueryResultAdminOrgVdcTemplateRecordType represents an admin VDC Template
type QueryResultAdminOrgVdcTemplateRecordType struct {
	HREF                     string `xml:"href,attr,omitempty"`
	ID                       string `xml:"id,attr,omitempty"`
	Name                     string `xml:"name,attr,omitempty"`
	Description              string `xml:"description,attr,omitempty"`
	TenantVisibleName        string `xml:"tenantVisibleName,attr,omitempty"`
	TenantVisibleDescription string `xml:"tenantVisibleDescription,attr,omitempty"`
	NetworkBackingType       string `xml:"networkBackingType,attr,omitempty"`
}

// QueryResultOrgVdcTemplateRecordType represents an admin VDC Template
type QueryResultOrgVdcTemplateRecordType struct {
	HREF        string `xml:"href,attr,omitempty"`
	ID          string `xml:"id,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Description string `xml:"description,attr,omitempty"`
	OrgHref     string `xml:"org,attr,omitempty"`
}

type QueryResultVappNetworkRecordType struct {
	HREF                   string    `xml:"href,attr,omitempty"`
	ID                     string    `xml:"id,attr,omitempty"`
	Name                   string    `xml:"name,attr,omitempty"`
	Type                   string    `xml:"linkType,attr,omitempty"`
	IpScopeId              string    `xml:"ipScopeId,attr,omitempty"`
	IpScopeInherited       bool      `xml:"ipScopeInherited,attr,omitempty"`
	Gateway                string    `xml:"gateway,attr,omitempty"`
	Netmask                string    `xml:"netmask,attr,omitempty"`
	SubnetPrefixLength     int       `xml:"subnetPrefixLength,attr,omitempty"`
	Dns1                   string    `xml:"dns1,attr,omitempty"`
	Dns2                   string    `xml:"dns2,attr,omitempty"`
	DnsSuffix              string    `xml:"dnsSuffix,attr,omitempty"`
	Vapp                   string    `xml:"vApp,attr,omitempty"`            // the HREF of the parent vApp
	VappName               string    `xml:"vAppName,attr,omitempty"`        // the name of the parent vApp
	LinkNetworkName        string    `xml:"linkNetworkName,attr,omitempty"` // this field is filled when called in tenant context
	RealNetworkName        string    `xml:"realNetworkName,attr,omitempty"`
	RealNetworkPortgroupId string    `xml:"realNetworkPortgroupId,attr,omitempty"`
	VCenterName            string    `xml:"vcName,attr,omitempty"`
	VCenter                string    `xml:"vc,attr,omitempty"`
	IsBusy                 bool      `xml:"isBusy,attr,omitempty"`
	IsLinked               bool      `xml:"isLinked,attr,omitempty"`
	RetainNicResources     bool      `xml:"retainNicResources,attr,omitempty"`
	Metadata               *Metadata `xml:"Metadata,omitempty"`
}

// QueryResultResourcePoolRecordType represent a Resource Pool record
type QueryResultResourcePoolRecordType struct {
	HREF                string `xml:"href,attr,omitempty"`
	Name                string `xml:"name,attr,omitempty"`
	Moref               string `xml:"moref,attr,omitempty"`
	IsDeleted           bool   `xml:"isDeleted,attr,omitempty"`
	VcenterHREF         string `xml:"vc,attr,omitempty"`
	VcenterName         string `xml:"vcName,attr,omitempty"`
	ProviderVdcHREF     string `xml:"providerVdc,attr,omitempty"`
	ProviderName        string `xml:"providerName,attr,omitempty"`
	IsEnabled           bool   `xml:"isEnabled,attr,omitempty"`
	IsPrimary           bool   `xml:"isPrimary,attr,omitempty"`
	ClusterMoref        string `xml:"clusterMoref,attr,omitempty"`
	IsKubernetesEnabled bool   `xml:"isKubernetesEnabled,attr,omitempty"`
}

// QueryResultOrgVdcRecordType represents an Org VDC record
type QueryResultOrgVdcRecordType struct {
	HREF                           string    `xml:"href,attr,omitempty"`
	Name                           string    `xml:"name,attr,omitempty"`
	ComputeProviderScope           string    `xml:"computeProviderScope,attr,omitempty"`
	NetworkProviderScope           string    `xml:"networkProviderScope,attr,omitempty"`
	IsEnabled                      string    `xml:"isEnabled,attr,omitempty"`
	CpuAllocationMhz               *int      `xml:"cpuAllocationMhz,attr,omitempty"`
	CpuLimitMhz                    *int      `xml:"cpuLimitMhz,attr,omitempty"`
	CpuUsedMhz                     *int      `xml:"cpuUsedMhz,attr,omitempty"`
	MemoryAllocationMB             *int      `xml:"memoryAllocationMB,attr,omitempty"`
	MemoryLimitMB                  *int      `xml:"memoryLimitMB,attr,omitempty"`
	MemoryUsedMB                   *int      `xml:"memoryUsedMB,attr,omitempty"`
	StorageLimitMB                 *int      `xml:"storageLimitMB,attr,omitempty"`
	StorageUsedMB                  *int      `xml:"storageUsedMB,attr,omitempty"`
	StorageOverheadMB              *int      `xml:"storageOverheadMB,attr,omitempty"`
	MemoryOverheadMB               *int      `xml:"memoryOverheadMB,attr,omitempty"`
	NumberOfVApps                  *int      `xml:"numberOfVApps,attr,omitempty"`
	NumberOfUnmanagedVApps         *int      `xml:"numberOfUnmanagedVApps,attr,omitempty"`
	NumberOfMedia                  *int      `xml:"numberOfMedia,attr,omitempty"`
	NumberOfDisks                  *int      `xml:"numberOfDisks,attr,omitempty"`
	NumberOfVAppTemplates          *int      `xml:"numberOfVAppTemplates,attr,omitempty"`
	NumberOfStorageProfiles        *int      `xml:"numberOfStorageProfiles,attr,omitempty"`
	NumberOfVMs                    *int      `xml:"numberOfVMs,attr,omitempty"`
	NumberOfRunningVMs             *int      `xml:"numberOfRunningVMs,attr,omitempty"`
	NumberOfDeployedVApps          *int      `xml:"numberOfDeployedVApps,attr,omitempty"`
	NumberOfDeployedUnmanagedVApps *int      `xml:"numberOfDeployedUnmanagedVApps,attr,omitempty"`
	CpuOverheadMhz                 *int      `xml:"cpuOverheadMhz,attr,omitempty"`
	OrgName                        string    `xml:"orgName,attr,omitempty"`
	AllocationModel                string    `xml:"allocationModel,attr,omitempty"`
	VcName                         string    `xml:"vcName,attr,omitempty"`
	IsBusy                         string    `xml:"isBusy,attr,omitempty"`
	Status                         string    `xml:"status,attr,omitempty"`
	TaskStatusName                 string    `xml:"taskStatusName,attr,omitempty"`
	Task                           string    `xml:"task,attr,omitempty"`
	TaskStatus                     string    `xml:"taskStatus,attr,omitempty"`
	TaskDetails                    string    `xml:"taskDetails,attr,omitempty"`
	Metadata                       *Metadata `xml:"Metadata,omitempty"`

	// Admin Org VDC fields
	ProviderVdcName       string `xml:"providerVdcName,attr,omitempty"`
	ProviderVdc           string `xml:"providerVdc,attr,omitempty"`
	Org                   string `xml:"org,attr,omitempty"`
	NetworkPool           string `xml:"networkPool,attr,omitempty"`
	NumberOfResourcePools *int   `xml:"numberOfResourcePools,attr,omitempty"`
	UsedNetworksInVdc     string `xml:"usedNetworksInVdc,attr,omitempty"`
	IsThinProvisioned     string `xml:"isThinProvisioned,attr,omitempty"`
	IsFastProvisioned     string `xml:"isFastProvisioned,attr,omitempty"`
	NetworkProviderType   string `xml:"networkProviderType,attr,omitempty"`
	IsVCEnabled           string `xml:"isVCEnabled,attr,omitempty"`
	MemoryReservedMB      *int   `xml:"memoryReservedMB,attr,omitempty"`
	CpuReservedMhz        *int   `xml:"cpuReservedMhz,attr,omitempty"`
	Vc                    string `xml:"vc,attr,omitempty"`
}

// QueryResultCatalogItemType represents a catalog item as query result
type QueryResultCatalogItemType struct {
	HREF         string    `xml:"href,attr,omitempty"`         // The URI of the entity.
	ID           string    `xml:"id,attr,omitempty"`           // Catalog Item ID.
	Type         string    `xml:"type,attr,omitempty"`         // The MIME type of the entity.
	Entity       string    `xml:"entity,attr,omitempty"`       // Entity reference or ID
	EntityName   string    `xml:"entityName,attr,omitempty"`   // Entity name
	EntityType   string    `xml:"entityType,attr,omitempty"`   // Entity type
	Catalog      string    `xml:"catalog,attr,omitempty"`      // Catalog reference or ID
	CatalogName  string    `xml:"catalogName,attr,omitempty"`  // Catalog name
	OwnerName    string    `xml:"ownerName,attr,omitempty"`    // Owner name
	Owner        string    `xml:"owner,attr,omitempty"`        // Owner reference or ID
	IsPublished  bool      `xml:"isPublished,attr,omitempty"`  // True if this entity is in a published catalog
	Vdc          string    `xml:"vdc,attr,omitempty"`          // VDC reference or ID
	VdcName      string    `xml:"vdcName,attr,omitempty"`      // VDC name
	IsVdcEnabled bool      `xml:"isVdcEnabled,attr,omitempty"` // True if the containing VDC is enabled
	CreationDate string    `xml:"creationDate,attr,omitempty"` // Creation date
	IsExpired    bool      `xml:"isExpired,attr,omitempty"`    // True if this entity is expired
	Status       string    `xml:"status,attr,omitempty"`       // Status
	Name         string    `xml:"name,attr,omitempty"`         // Catalog Item name.
	Link         *Link     `xml:"Link,omitempty"`
	Metadata     *Metadata `xml:"Metadata,omitempty"`
}

// QueryResultVappTemplateType represents a vApp template as query result
type QueryResultVappTemplateType struct {
	HREF                    string    `xml:"href,attr,omitempty"`                    // The URI of the entity.
	ID                      string    `xml:"id,attr,omitempty"`                      // vApp template ID.
	Type                    string    `xml:"type,attr,omitempty"`                    // The MIME type of the entity.
	OwnerName               string    `xml:"ownerName,attr,omitempty"`               // Owner name
	Owner                   string    `xml:"owner,attr,omitempty"`                   // Owner reference or ID
	CatalogName             string    `xml:"catalogName,attr,omitempty"`             // Catalog name
	Catalog                 string    `xml:"catalog,attr,omitempty"`                 // Catalog reference or ID
	CatalogItem             string    `xml:"catalogItem,attr,omitempty"`             // CatalogItem reference or ID
	IsPublished             bool      `xml:"isPublished,attr,omitempty"`             // True if this entity is in a published catalog
	PublishSubscriptionType string    `xml:"publishSubscriptionType,attr,omitempty"` // PUBLISHED if parent catalog published externally, SUBSCRIBED if parent catalog subscribed to an external catalog, UNPUBLISHED otherwise.
	Name                    string    `xml:"name,attr,omitempty"`                    // vApp template name.
	Description             string    `xml:"description,attr,omitempty"`             // vApp template description.
	Vdc                     string    `xml:"vdc,attr,omitempty"`                     // VDC reference or ID
	VdcName                 string    `xml:"vdcName,attr,omitempty"`                 // VDC name
	IsVdcEnabled            bool      `xml:"isVdcEnabled,attr,omitempty"`            // true if the containing VDC is enabled
	Org                     string    `xml:"org,attr,omitempty"`                     // Organization reference or ID
	CreationDate            string    `xml:"creationDate,attr,omitempty"`            // Creation date
	IsBusy                  bool      `xml:"isBusy,attr,omitempty"`                  // True if the vApp template is busy
	IsGoldMaster            bool      `xml:"isGoldMaster,attr,omitempty"`            // True if the vApp template is a gold master
	IsEnabled               bool      `xml:"isEnabled,attr,omitempty"`               // True if the vApp template is enabled
	Status                  string    `xml:"status,attr,omitempty"`                  // Status
	IsDeployed              bool      `xml:"isDeployed,attr,omitempty"`              // True if this entity is deployed
	IsExpired               bool      `xml:"isExpired,attr,omitempty"`               // True if this entity is expired
	StorageProfileName      string    `xml:"storageProfileName,attr,omitempty"`      // Storage profile name
	Version                 string    `xml:"version,attr,omitempty"`                 // Storage profile name
	LastSuccessfulSync      string    `xml:"lastSuccessfulSync,attr,omitempty"`      // Date of last successful sync
	Link                    *Link     `xml:"Link,omitempty"`
	Metadata                *Metadata `xml:"Metadata,omitempty"`

	// Undocumented fields
	// Provisionally used in some catalog synchronisation tasks. To be removed and replaced by using a different algorithm
	// https://developer.vmware.com/apis/1260/vmware-cloud-director/doc/doc/types/QueryResultAdminVAppTemplateRecordType.html
	// https://developer.vmware.com/apis/1260/vmware-cloud-director/doc/doc/types/QueryResultVAppTemplateRecordType.html
	IsInCatalog              bool   `xml:"isInCatalog,attr,omitempty"`              // True if this vApp template is in a catalog
	HonorBootOrder           bool   `xml:"honorBootOrder,attr,omitempty"`           // ?
	NumberOfShadowVms        int    `xml:"numberOfShadowVMs,attr,omitempty"`        // number of shadow VMs
	NumberOfVms              int    `xml:"numberOfVMs,attr,omitempty"`              // number of VMs
	TaskStatusName           string `xml:"taskStatusName,attr,omitempty"`           // name of the associated task
	TaskStatus               string `xml:"taskStatus,attr,omitempty"`               // status of the associated task
	Task                     string `xml:"task,attr,omitempty"`                     // ID or reference of the associated task
	TaskDetails              string `xml:"taskDetails,attr,omitempty"`              // details of the associated task
	IsDeleteUndeployNotified bool   `xml:"isDeleteUndeployNotified,attr,omitempty"` // ?
	IsAutoUndeployNotified   bool   `xml:"isAutoUndeployNotified,attr,omitempty"`   // ?
	CpuAllocationInMhz       int    `xml:"cpuAllocationInMhz,attr,omitempty"`       // CPU allocation
	NumberOfCpus             int    `xml:"numberOfCpus,attr,omitempty"`             // Number of CPUs
	MemoryAllocationMb       int    `xml:"memoryAllocationMB,attr,omitempty"`       // Memory allocation in MB
	StorageKb                int    `xml:"storageKB,attr,omitempty"`                // Storage allocation in Kb
}

// QueryResultEdgeGatewayRecordType represents an edge gateway record as query result.
type QueryResultEdgeGatewayRecordType struct {
	// Attributes
	HREF                string `xml:"href,attr,omitempty"`                // The URI of the entity.
	Type                string `xml:"type,attr,omitempty"`                // The MIME type of the entity.
	Name                string `xml:"name,attr,omitempty"`                // EdgeGateway name.
	Vdc                 string `xml:"vdc,attr,omitempty"`                 // VDC Reference or ID
	OrgVdcName          string `xml:"orgVdcName,attr,omitempty"`          // VDC name
	NumberOfExtNetworks int    `xml:"numberOfExtNetworks,attr,omitempty"` // Number of external networks connected to the edgeGateway.	Yes	Yes
	NumberOfOrgNetworks int    `xml:"numberOfOrgNetworks,attr,omitempty"` // Number of org VDC networks connected to the edgeGateway	Yes	Yes
	IsBusy              bool   `xml:"isBusy,attr"`                        // True if this Edge Gateway is busy.	Yes	Yes
	GatewayStatus       string `xml:"gatewayStatus,attr,omitempty"`       //
	HaStatus            string `xml:"haStatus,attr,omitempty"`            // High Availability Status of the edgeGateway	Yes	Yes
}

// QueryResultVMRecordType represents a VM record as query result.
type QueryResultVMRecordType struct {
	// Attributes
	HREF                     string    `xml:"href,attr,omitempty"` // The URI of the entity.
	ID                       string    `xml:"id,attr,omitempty"`
	Moref                    string    `xml:"moref,attr,omitempty"`         // VM moref id.
	Name                     string    `xml:"name,attr,omitempty"`          // VM name.
	Type                     string    `xml:"type,attr,omitempty"`          // Contains the type of the resource.
	ContainerName            string    `xml:"containerName,attr,omitempty"` // The name of the vApp or vApp template that contains this VM.
	ContainerID              string    `xml:"container,attr,omitempty"`     // The ID of the vApp or vApp template that contains this VM.
	OwnerName                string    `xml:"ownerName,attr,omitempty"`
	Owner                    string    `xml:"owner,attr,omitempty"`
	VdcHREF                  string    `xml:"vdc,attr,omitempty"`
	VdcName                  string    `xml:"vdcName,attr,omitempty"`
	VAppTemplate             bool      `xml:"isVAppTemplate,attr,omitempty"`
	Deleted                  bool      `xml:"isDeleted,attr,omitempty"`
	GuestOS                  string    `xml:"guestOs,attr,omitempty"`
	DetectedGuestOS          string    `xml:"detectedGuestOs,attr,omitempty"`
	Cpus                     int       `xml:"numberOfCpus,attr,omitempty"`
	MemoryMB                 int       `xml:"memoryMB,attr,omitempty"`
	Status                   string    `xml:"status,attr,omitempty"`
	NetworkName              string    `xml:"networkName,attr,omitempty"`
	NetworkHref              string    `xml:"network,attr,omitempty"`
	IpAddress                string    `xml:"ipAddress,attr,omitempty"` // If configured, the IP Address of the VM on the primary network, otherwise empty.
	Busy                     bool      `xml:"isBusy,attr,omitempty"`
	Deployed                 bool      `xml:"isDeployed,attr,omitempty"` // True if the virtual machine is deployed.
	Published                bool      `xml:"isPublished,attr,omitempty"`
	CatalogName              string    `xml:"catalogName,attr,omitempty"`
	HardwareVersion          int       `xml:"hardwareVersion,attr,omitempty"`
	VmToolsStatus            string    `xml:"vmToolsStatus,attr,omitempty"`
	MaintenanceMode          bool      `xml:"isInMaintenanceMode,attr,omitempty"`
	AutoNature               bool      `xml:"isAutoNature,attr,omitempty"` //  	True if the parent vApp is a managed vApp
	StorageProfileName       string    `xml:"storageProfileName,attr,omitempty"`
	GcStatus                 string    `xml:"gcStatus,attr,omitempty"` // GC status of this VM.
	AutoUndeployDate         string    `xml:"autoUndeployDate,attr,omitempty"`
	AutoDeleteDate           string    `xml:"autoDeleteDate,attr,omitempty"`
	AutoUndeployNotified     bool      `xml:"isAutoUndeployNotified,attr,omitempty"`
	AutoDeleteNotified       bool      `xml:"isAutoDeleteNotified,attr,omitempty"`
	IsComputePolicyCompliant bool      `xml:"isComputePolicyCompliant,attr,omitempty"`
	VmSizingPolicyId         string    `xml:"vmSizingPolicyId,attr,omitempty"`
	VmPlacementPolicyId      string    `xml:"vmPlacementPolicyId,attr,omitempty"`
	Encrypted                bool      `xml:"encrypted,attr,omitempty"`
	DateCreated              string    `xml:"dateCreated,attr,omitempty"`
	TotalStorageAllocatedMb  string    `xml:"totalStorageAllocatedMb,attr,omitempty"`
	IsExpired                bool      `xml:"isExpired,attr,omitempty"`
	HostName                 string    `xml:"hostName,attr,omitempty"` // HostName=Hypervisor of virtual machine
	Link                     []*Link   `xml:"Link,omitempty"`
	MetaData                 *Metadata `xml:"Metadata,omitempty"`
	VmNameInVc               string    `xml:"vmNameInVc,attr,omitempty"`
}

// QueryResultVAppRecordType represents a VM record as query result.
type QueryResultVAppRecordType struct {
	// Attributes
	HREF                    string    `xml:"href,attr,omitempty"`         // The URI of the entity.
	Name                    string    `xml:"name,attr"`                   // The name of the entity.
	CreationDate            string    `xml:"creationDate,attr,omitempty"` // Creation date/time of the vApp.
	Busy                    bool      `xml:"isBusy,attr,omitempty"`
	Deployed                bool      `xml:"isDeployed,attr,omitempty"` // True if the vApp is deployed.
	Enabled                 bool      `xml:"isEnabled,attr,omitempty"`
	Expired                 bool      `xml:"isExpired,attr,omitempty"`
	MaintenanceMode         bool      `xml:"isInMaintenanceMode,attr,omitempty"`
	Public                  bool      `xml:"isPublic,attr,omitempty"`
	OwnerName               string    `xml:"ownerName,attr,omitempty"`
	Status                  string    `xml:"status,attr,omitempty"`
	VdcHREF                 string    `xml:"vdc,attr,omitempty"`
	VdcName                 string    `xml:"vdcName,attr,omitempty"`
	NumberOfVMs             int       `xml:"numberOfVMs,attr,omitempty"`
	NumberOfCPUs            int       `xml:"numberOfCpus,attr,omitempty"`
	CpuAllocationMhz        int       `xml:"cpuAllocationMhz,attr,omitempty"`
	CpuAllocationInMhz      int       `xml:"cpuAllocationInMhz,attr,omitempty"`
	StorageKB               int       `xml:"storageKB,attr,omitempty"`
	MemoryAllocationMB      int       `xml:"memoryAllocationMB,attr,omitempty"`
	AutoDeleteNotified      bool      `xml:"isAutoDeleteNotified,attr,omitempty"`
	AutoUndeployNotified    bool      `xml:"isAutoUndeployNotified,attr,omitempty"`
	VdcEnabled              bool      `xml:"isVdcEnabled,attr,omitempty"`
	HonorBootOrder          bool      `xml:"honorBookOrder,attr,omitempty"`
	HighestSupportedVersion int       `xml:"pvdcHighestSupportedHardwareVersion,attr,omitempty"`
	LowestHardwareVersion   int       `xml:"lowestHardwareVersionInVApp,attr,omitempty"`
	TaskHREF                string    `xml:"task,attr,omitempty"`
	TaskStatusName          string    `xml:"taskStatusName,attr,omitempty"`
	TaskStatus              string    `xml:"TaskStatus,attr,omitempty"`
	TaskDetails             string    `xml:"taskDetails,attr,omitempty"`
	MetaData                *Metadata `xml:"Metadata,omitempty"`
}

// QueryResultOrgVdcStorageProfileRecordType represents a storage
// profile as query result.
// https://code.vmware.com/apis/722/vmware-cloud-director/doc/doc/types/QueryResultOrgVdcStorageProfileRecordType.html
type QueryResultOrgVdcStorageProfileRecordType struct {
	// Attributes
	HREF                    string `xml:"href,attr,omitempty"`                    // The URI of the entity.
	ID                      string `xml:"id,attr,omitempty"`                      // The ID of the entity.
	Type                    string `xml:"type,attr,omitempty"`                    // Contains the type of the resource.
	Name                    string `xml:"name,attr,omitempty"`                    // Name of the storage profile.
	IsEnabled               bool   `xml:"isEnabled,attr,omitempty"`               // True if this entity is enabled.
	IsDefaultStorageProfile bool   `xml:"isDefaultStorageProfile,attr,omitempty"` // True if this is the default storage profile for a VDC.
	StorageUsedMB           uint64 `xml:"storageUsedMB,attr,omitempty"`           // Storage used in MB.
	StorageLimitMB          uint64 `xml:"storageLimitMB,attr,omitempty"`          // Storage limit in MB.
	IopsAllocated           uint64 `xml:"iopsAllocated,attr,omitempty"`           // Total currently allocated IOPS on the storage profile.
	IopsLimit               uint64 `xml:"iopsLimit,attr,omitempty"`               // IOPS limit for the storage profile.
	NumberOfConditions      int    `xml:"numberOfConditions,attr,omitempty"`      // Number of conditions on the storage profile.
	Vdc                     string `xml:"vdc,attr,omitempty"`                     // VDC reference or id.
	VdcName                 string `xml:"vdcName,attr,omitempty"`                 // VDC name.
	IsVdcBusy               bool   `xml:"isVdcBusy,attr,omitempty"`               // True if the associated VDC is busy.
	// Elements
	Link          []*Link          `xml:"Link,omitempty"`
	MetadataEntry []*MetadataEntry `xml:"MetadataEntry,omitempty"`
}

// QueryResultAdminOrgVdcStorageProfileRecordType represents a storage
// profile as query result.
// https://code.vmware.com/apis/722/vmware-cloud-director/doc/doc/types/QueryResultAdminOrgVdcStorageProfileRecordType.html
type QueryResultAdminOrgVdcStorageProfileRecordType struct {
	// Attributes
	HREF                    string `xml:"href,attr,omitempty"`                    // The URI of the entity.
	ID                      string `xml:"id,attr,omitempty"`                      // The ID of the entity.
	Type                    string `xml:"type,attr,omitempty"`                    // Contains the type of the resource.
	Name                    string `xml:"name,attr,omitempty"`                    // Name of the storage profile.
	IsEnabled               bool   `xml:"isEnabled,attr,omitempty"`               // True if this entity is enabled.
	IsDefaultStorageProfile bool   `xml:"isDefaultStorageProfile,attr,omitempty"` // True if this is the default storage profile for a VDC.
	StorageUsedMB           uint64 `xml:"storageUsedMB,attr,omitempty"`           // Storage used in MB.
	StorageLimitMB          uint64 `xml:"storageLimitMB,attr,omitempty"`          // Storage limit in MB.
	IopsAllocated           uint64 `xml:"iopsAllocated,attr,omitempty"`           // Total currently allocated IOPS on the storage profile.
	IopsLimit               uint64 `xml:"iopsLimit,attr,omitempty"`               // IOPS limit for the storage profile.
	NumberOfConditions      int    `xml:"numberOfConditions,attr,omitempty"`      // Number of conditions on the storage profile.
	Vdc                     string `xml:"vdc,attr,omitempty"`                     // VDC reference or id.
	VdcName                 string `xml:"vdcName,attr,omitempty"`                 // VDC name.
	Org                     string `xml:"org,attr,omitempty"`                     // Organization reference or id.
	VC                      string `xml:"vc,attr,omitempty"`                      // Virtual center reference or id.
	StorageProfileMoref     string `xml:"storageProfileMoref,omitempty"`
	// Elements
	Link          []*Link          `xml:"Link,omitempty"`
	MetadataEntry []*MetadataEntry `xml:"MetadataEntry,omitempty"`
}

// QueryResultVMWProviderVdcRecordType represents a Provider VDC as query result.
type QueryResultVMWProviderVdcRecordType struct {
	// Attributes
	HREF                    string `xml:"href,attr,omitempty"` // The URI of the entity.
	Name                    string `xml:"name,attr,omitempty"` // Provider VDC name.
	Status                  string `xml:"status,attr,omitempty"`
	IsBusy                  bool   `xml:"isBusy,attr,omitempty"`
	IsDeleted               bool   `xml:"isDeleted,attr,omitempty"`
	IsEnabled               bool   `xml:"isEnabled,attr,omitempty"`
	CpuAllocationMhz        int    `xml:"cpuAllocationMhz,attr,omitempty"`
	CpuLimitMhz             int    `xml:"cpuLimitMhz,attr,omitempty"`
	CpuUsedMhz              int    `xml:"cpuUsedMhz,attr,omitempty"`
	NumberOfDatastores      int    `xml:"numberOfDatastores,attr,omitempty"`
	NumberOfStorageProfiles int    `xml:"numberOfStorageProfiles,attr,omitempty"`
	NumberOfVdcs            int    `xml:"numberOfVdcs,attr,omitempty"`
	MemoryAllocationMB      int64  `xml:"memoryAllocationMB,attr,omitempty"`
	MemoryLimitMB           int64  `xml:"memoryLimitMB,attr,omitempty"`
	MemoryUsedMB            int64  `xml:"memoryUsedMB,attr,omitempty"`
	StorageAllocationMB     int64  `xml:"storageAllocationMB,attr,omitempty"`
	StorageLimitMB          int64  `xml:"storageLimitMB,attr,omitempty"`
	StorageUsedMB           int64  `xml:"storageUsedMB,attr,omitempty"`
	CpuOverheadMhz          int64  `xml:"cpuOverheadMhz,attr,omitempty"`
	StorageOverheadMB       int64  `xml:"storageOverheadMB,attr,omitempty"`
	MemoryOverheadMB        int64  `xml:"memoryOverheadMB,attr,omitempty"`
}

// QueryResultProviderVdcStorageProfileRecordType represents a Provider VDC storage profile as query result.
type QueryResultProviderVdcStorageProfileRecordType struct {
	// Attributes
	HREF                 string `xml:"href,attr,omitempty"` // The URI of the entity.
	Name                 string `xml:"name,attr,omitempty"` // Provider VDC Storage Profile name.
	ProviderVdcHREF      string `xml:"providerVdc,attr,omitempty"`
	VcHREF               string `xml:"vc,attr,omitempty"`
	StorageProfileMoref  string `xml:"storageProfileMoref,attr,omitempty"`
	IsEnabled            bool   `xml:"isEnabled,attr,omitempty"`
	StorageProvisionedMB int64  `xml:"storageProvisionedMB,attr,omitempty"`
	StorageRequestedMB   int64  `xml:"storageRequestedMB,attr,omitempty"`
	StorageTotalMB       int64  `xml:"storageTotalMB,attr,omitempty"`
	StorageUsedMB        int64  `xml:"storageUsedMB,attr,omitempty"`
	NumberOfConditions   int    `xml:"numberOfConditions,attr,omitempty"`
}

// QueryResultNetworkPoolRecordType represents a network pool as query result.
type QueryResultNetworkPoolRecordType struct {
	// Attributes
	HREF            string `xml:"href,attr,omitempty"` // The URI of the entity.
	Name            string `xml:"name,attr,omitempty"` // Network pool name.
	IsBusy          bool   `xml:"isBusy,attr,omitempty"`
	NetworkPoolType int    `xml:"networkPoolType,attr,omitempty"`
}

// Type: QueryResultVirtualCenterRecordType
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/QueryResultVirtualCenterRecordType.html
// Description: Type for a single virtualCenter query result in records format.
// Since: 1.5
type QueryResultVirtualCenterRecordType struct {
	HREF          string `xml:"href,attr,omitempty"`
	Name          string `xml:"name,attr,omitempty"`
	IsBusy        bool   `xml:"isBusy,attr,omitempty"`
	IsEnabled     bool   `xml:"isEnabled,attr,omitempty"`
	IsSupported   bool   `xml:"isSupported,attr,omitempty"`
	ListenerState string `xml:"listenerState,attr,omitempty"`
	Status        string `xml:"status,attr,omitempty"`
	Url           string `xml:"url,attr,omitempty"`
	UserName      string `xml:"userName,attr,omitempty"`
	VcVersion     string `xml:"vcVersion,attr,omitempty"`
	UUID          string `xml:"uuid,attr,omitempty"`
	VsmIP         string `xml:"vsmIP,attr,omitempty"`
}

// Namespace: http://www.vmware.com/vcloud/v1.5
// Retrieve a list of extension objects and operations.
// Since: 1.0
type Extension struct {
	Link LinkList `xml:"Link,omitempty"` // A reference to an entity or operation associated with this object.
}

// Namespace: http://www.vmware.com/vcloud/v1.5
// Retrieve a list of tasks
type TasksList struct {
	XMLName xml.Name `xml:"TasksList"`
	Task    []*Task  `xml:"Task,omitempty"`
}

type ExternalNetworkReferences struct {
	ExternalNetworkReference []*ExternalNetworkReference `xml:"ExternalNetworkReference,omitempty"` // A reference to an entity or operation associated with this object.
}

type ExternalNetworkReference struct {
	HREF string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
	Name string `xml:"name,attr,omitempty"`
}

// Type: VimObjectRefType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/VimObjectRefsType.html
// Description: Represents the Managed Object Reference (MoRef) and the type of a vSphere object.
// Since: 0.9
type VimObjectRef struct {
	VimServerRef  *Reference `xml:"VimServerRef" json:"vimServerRef"`
	MoRef         string     `xml:"MoRef" json:"moRef"`
	VimObjectType string     `xml:"VimObjectType" json:"vimObjectType"`
}

// Type: VimObjectRefsType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/VimObjectRefsType.html
// Description: List of VimObjectRef elements.
// Since: 0.9
type VimObjectRefs struct {
	VimObjectRef []*VimObjectRef `xml:"VimObjectRef" json:"vimObjectRef"`
}

// Type: VMWExternalNetworkType
// Namespace: http://www.vmware.com/vcloud/extension/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/VMWExternalNetworkType.html
// Description: External network type.
// Since: 1.0
type ExternalNetwork struct {
	XMLName          xml.Name              `xml:"VMWExternalNetwork"`
	HREF             string                `xml:"href,attr,omitempty"`
	Type             string                `xml:"type,attr,omitempty"`
	ID               string                `xml:"id,attr,omitempty"`
	OperationKey     string                `xml:"operationKey,attr,omitempty"`
	Name             string                `xml:"name,attr"`
	Link             []*Link               `xml:"Link,omitempty"`
	Description      string                `xml:"Description,omitempty"`
	Tasks            *TasksInProgress      `xml:"Tasks,omitempty"`
	Configuration    *NetworkConfiguration `xml:"Configuration,omitempty"`
	VimPortGroupRef  *VimObjectRef         `xml:"VimPortGroupRef,omitempty"`
	VimPortGroupRefs *VimObjectRefs        `xml:"VimPortGroupRefs,omitempty"`
	VCloudExtension  *VCloudExtension      `xml:"VCloudExtension,omitempty"`
}

// Type: MediaType
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/ca48e1bb-282b-4fdc-b827-649b819249ed/55142cf1-5bb8-4ab1-8d09-b84f717af5ec/doc/doc/types/MediaType.html
// Description: Represents Media image
// Since: 0.9
type Media struct {
	HREF         string           `xml:"href,attr,omitempty"`
	Type         string           `xml:"type,attr,omitempty"`
	ID           string           `xml:"id,attr,omitempty"`
	OperationKey string           `xml:"operationKey,attr,omitempty"`
	Name         string           `xml:"name,attr"`
	Status       int64            `xml:"status,attr,omitempty"`
	ImageType    string           `xml:"imageType,attr,omitempty"`
	Size         int64            `xml:"size,attr,omitempty"`
	Description  string           `xml:"Description,omitempty"`
	Files        *FilesList       `xml:"Files,omitempty"`
	Link         LinkList         `xml:"Link,omitempty"`
	Tasks        *TasksInProgress `xml:"Tasks,omitempty"`
	Owner        *Reference       `xml:"Owner,omitempty"`
	Entity       *Entity          `xml:"Entity"`
}

// Type: MediaRecord
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/GET-MediasFromQuery.html
// Issue that description partly matches with what is returned
// Description: Represents Media record
// Since: 1.5
type MediaRecordType struct {
	HREF               string    `xml:"href,attr,omitempty"`
	ID                 string    `xml:"id,attr,omitempty"`
	Type               string    `xml:"type,attr,omitempty"`
	OwnerName          string    `xml:"ownerName,attr,omitempty"`
	CatalogName        string    `xml:"catalogName,attr,omitempty"`
	IsPublished        bool      `xml:"isPublished,attr,omitempty"`
	Name               string    `xml:"name,attr"`
	Vdc                string    `xml:"vdc,attr,omitempty"`
	VdcName            string    `xml:"vdcName,attr,omitempty"`
	Org                string    `xml:"org,attr,omitempty"`
	CreationDate       string    `xml:"creationDate,attr,omitempty"`
	IsBusy             bool      `xml:"isBusy,attr,omitempty"`
	StorageB           int64     `xml:"storageB,attr,omitempty"`
	Owner              string    `xml:"owner,attr,omitempty"`
	Catalog            string    `xml:"catalog,attr,omitempty"`
	CatalogItem        string    `xml:"catalogItem,attr,omitempty"`
	Status             string    `xml:"status,attr,omitempty"`
	StorageProfileName string    `xml:"storageProfileName,attr,omitempty"`
	Version            int64     `xml:"version,attr,omitempty"`
	LastSuccessfulSync string    `xml:"lastSuccessfulSync,attr,omitempty"`
	TaskStatusName     string    `xml:"taskStatusName,attr,omitempty"`
	IsInCatalog        bool      `xml:"isInCatalog,attr,omitempty"`
	Task               string    `xml:"task,attr,omitempty"`
	IsIso              bool      `xml:"isIso,attr,omitempty"`
	IsVdcEnabled       bool      `xml:"isVdcEnabled,attr,omitempty"`
	TaskStatus         string    `xml:"taskStatus,attr,omitempty"`
	TaskDetails        string    `xml:"taskDetails,attr,omitempty"`
	Link               *Link     `xml:"Link,omitempty"`
	Metadata           *Metadata `xml:"Metadata,omitempty"`
}

// DiskCreateParams element for create independent disk
// Reference: vCloud API 30.0 - DiskCreateParamsType
// https://code.vmware.com/apis/287/vcloud?h=Director#/doc/doc/types/DiskCreateParamsType.html
type DiskCreateParams struct {
	XMLName         xml.Name         `xml:"DiskCreateParams"`
	Xmlns           string           `xml:"xmlns,attr,omitempty"`
	Disk            *Disk            `xml:"Disk"`
	Locality        *Reference       `xml:"Locality,omitempty"`
	VCloudExtension *VCloudExtension `xml:"VCloudExtension,omitempty"`
}

// Represents an independent disk
// Reference: vCloud API 30.0 - DiskType
// https://code.vmware.com/apis/287/vcloud?h=Director#/doc/doc/types/DiskType.html
type Disk struct {
	XMLName      xml.Name `xml:"Disk"`
	Xmlns        string   `xml:"xmlns,attr,omitempty"`
	HREF         string   `xml:"href,attr,omitempty"`
	Type         string   `xml:"type,attr,omitempty"`
	Id           string   `xml:"id,attr,omitempty"`
	OperationKey string   `xml:"operationKey,attr,omitempty"`
	Name         string   `xml:"name,attr"`
	Status       int      `xml:"status,attr,omitempty"`
	// Size of the disk in bytes. No longer supported in API V33.0+.
	// Size            int64            `xml:"size,attr"`
	// SizeMb is the size of disk in MB. It has replaced Size (in bytes) field as of API V33.0
	SizeMb          int64            `xml:"sizeMb,attr,omitempty"`
	Iops            *int             `xml:"iops,attr,omitempty"`
	BusType         string           `xml:"busType,attr,omitempty"`
	BusSubType      string           `xml:"busSubType,attr,omitempty"`
	Encrypted       bool             `xml:"encrypted,attr,omitempty"`
	Shareable       bool             `xml:"shareable,attr,omitempty"`
	SharingType     string           `xml:"sharingType,attr,omitempty"`
	UUID            string           `xml:"uuid,attr,omitempty"`
	Description     string           `xml:"Description,omitempty"`
	Files           *FilesList       `xml:"Files,omitempty"`
	Link            []*Link          `xml:"Link,omitempty"`
	StorageProfile  *Reference       `xml:"StorageProfile,omitempty"`
	Owner           *Owner           `xml:"Owner,omitempty"`
	Tasks           *TasksInProgress `xml:"Tasks,omitempty"`
	VCloudExtension *VCloudExtension `xml:"VCloudExtension,omitempty"`
}

// General purpose extension element
// Not related to extension services
// Reference: vCloud API 30.0 - DiskAttachOrDetachParamsType
// https://code.vmware.com/apis/287/vcloud?h=Director#/doc/doc/types/VCloudExtensionType.html
type VCloudExtension struct {
	Required bool `xml:"required,attr,omitempty"`
}

// Parameters for attaching or detaching an independent disk
// Reference: vCloud API 30.0 - DiskAttachOrDetachParamsType
// https://code.vmware.com/apis/287/vcloud?h=Director#/doc/doc/types/DiskAttachOrDetachParamsType.html
type DiskAttachOrDetachParams struct {
	XMLName         xml.Name         `xml:"DiskAttachOrDetachParams"`
	Xmlns           string           `xml:"xmlns,attr,omitempty"`
	Disk            *Reference       `xml:"Disk"`
	BusNumber       *int             `xml:"BusNumber,omitempty"`
	UnitNumber      *int             `xml:"UnitNumber,omitempty"`
	VCloudExtension *VCloudExtension `xml:"VCloudExtension,omitempty"`
}

// Represents a list of virtual machines
// Reference: vCloud API 30.0 - VmsType
// https://code.vmware.com/apis/287/vcloud?h=Director#/doc/doc/types/FilesListType.html
type Vms struct {
	XMLName     xml.Name     `xml:"Vms"`
	Xmlns       string       `xml:"xmlns,attr,omitempty"`
	Type        string       `xml:"type,attr"`
	HREF        string       `xml:"href,attr"`
	VmReference []*Reference `xml:"VmReference,omitempty"`
}

// Parameters for inserting and ejecting virtual media for VM as CD/DVD
// Reference: vCloud API 30.0 - MediaInsertOrEjectParamsType
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/MediaInsertOrEjectParamsType.html
type MediaInsertOrEjectParams struct {
	XMLName         xml.Name         `xml:"MediaInsertOrEjectParams"`
	Xmlns           string           `xml:"xmlns,attr,omitempty"`
	Media           *Reference       `xml:"Media"`
	VCloudExtension *VCloudExtension `xml:"VCloudExtension,omitempty"`
}

// Parameters for VM pending questions
// Reference: vCloud API 27.0 - VmPendingQuestionType
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/VmPendingQuestionType.html
type VmPendingQuestion struct {
	XMLName    xml.Name                      `xml:"VmPendingQuestion"`
	Xmlns      string                        `xml:"xmlns,attr,omitempty"`
	Type       string                        `xml:"type,attr"`
	HREF       string                        `xml:"href,attr"`
	Link       LinkList                      `xml:"Link,omitempty"`
	Question   string                        `xml:"Question"`
	QuestionId string                        `xml:"QuestionId"`
	Choices    []*VmQuestionAnswerChoiceType `xml:"Choices"`
}

// Parameters for VM question answer choice
// Reference: vCloud API 27.0 - VmQuestionAnswerChoiceType
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/VmQuestionAnswerChoiceType.html
type VmQuestionAnswerChoiceType struct {
	Id   int    `xml:"Id"`
	Text string `xml:"Text,omitempty"`
}

// Parameters for VM question answer
// Reference: vCloud API 27.0 - VmQuestionAnswerType
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/VmQuestionAnswerType.html
type VmQuestionAnswer struct {
	XMLName    xml.Name `xml:"VmQuestionAnswer"`
	Xmlns      string   `xml:"xmlns,attr,omitempty"`
	ChoiceId   int      `xml:"ChoiceId"`
	QuestionId string   `xml:"QuestionId"`
}

// Represents an independent disk record
// Reference: vCloud API 27.0 - DiskType
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/QueryResultDiskRecordType.html
type DiskRecordType struct {
	Xmlns              string    `xml:"xmlns,attr,omitempty"`
	HREF               string    `xml:"href,attr,omitempty"`
	Id                 string    `xml:"id,attr,omitempty"`
	Type               string    `xml:"type,attr,omitempty"`
	Name               string    `xml:"name,attr,omitempty"`
	Vdc                string    `xml:"vdc,attr,omitempty"`
	SizeMb             int64     `xml:"sizeMb,attr,omitempty"`
	Iops               int64     `xml:"iops,attr,omitempty"`
	Encrypted          bool      `xml:"encrypted,attr,omitempty"`
	UUID               string    `xml:"uuid,attr,omitempty"`
	DataStore          string    `xml:"dataStore,attr,omitempty"`
	DataStoreName      string    `xml:"datastoreName,attr,omitempty"`
	OwnerName          string    `xml:"ownerName,attr,omitempty"`
	VdcName            string    `xml:"vdcName,attr,omitempty"`
	Task               string    `xml:"task,attr,omitempty"`
	StorageProfile     string    `xml:"storageProfile,attr,omitempty"`
	StorageProfileName string    `xml:"storageProfileName,attr,omitempty"`
	Status             string    `xml:"status,attr,omitempty"`
	BusType            string    `xml:"busType,attr,omitempty"`
	BusSubType         string    `xml:"busSubType,attr,omitempty"`
	BusTypeDesc        string    `xml:"busTypeDesc,attr,omitempty"`
	AttachedVmCount    int32     `xml:"attachedVmCount,attr,omitempty"`
	SharingType        string    `xml:"sharingType,attr,omitempty"`
	IsAttached         bool      `xml:"isAttached,attr,omitempty"`
	IsShareable        bool      `xml:"isShareable,attr,omitempty"`
	Description        string    `xml:"description,attr,omitempty"`
	Link               []*Link   `xml:"Link,omitempty"`
	Metadata           *Metadata `xml:"Metadata,omitempty"`
}

// Represents port group
// Reference: vCloud API 27.0 - Port group type
// https://code.vmware.com/apis/72/doc/doc/types/QueryResultPortgroupRecordType.html
type PortGroupRecordType struct {
	Xmlns         string  `xml:"xmlns,attr,omitempty"`
	HREF          string  `xml:"href,attr,omitempty"`
	Id            string  `xml:"id,attr,omitempty"`
	Type          string  `xml:"type,attr,omitempty"`
	MoRef         string  `xml:"moref,attr,omitempty"`
	Name          string  `xml:"name,attr,omitempty"`
	PortgroupType string  `xml:"portgroupType,attr,omitempty"`
	Vc            string  `xml:"vc,attr,omitempty"`
	VcName        string  `xml:"vcName,attr,omitempty"`
	IsVCEnabled   bool    `xml:"isVCEnabled,attr,omitempty"`
	Network       string  `xml:"network,attr,omitempty"`
	NetworkName   string  `xml:"networkName,attr,omitempty"`
	ScopeType     int     `xml:"scopeType,attr,omitempty"` // Scope of network using the portgroup(1=Global, 2=Organization, 3=vApp)
	Link          []*Link `xml:"Link,omitempty"`
}

// Represents org VDC Network
// Reference: vCloud API 27.0 - Org VDC Network
// https://code.vmware.com/apis/72/doc/doc/types/QueryResultOrgVdcNetworkRecordType.html
type QueryResultOrgVdcNetworkRecordType struct {
	Xmlns              string    `xml:"xmlns,attr,omitempty"`
	HREF               string    `xml:"href,attr,omitempty"`
	Id                 string    `xml:"id,attr,omitempty"`
	Type               string    `xml:"type,attr,omitempty"`
	Name               string    `xml:"name,attr,omitempty"`
	DefaultGateway     string    `xml:"defaultGateway,attr,omitempty"`
	Netmask            string    `xml:"netmask,attr,omitempty"`
	Dns1               string    `xml:"dns1,attr,omitempty"`
	Dns2               string    `xml:"dns2,attr,omitempty"`
	DnsSuffix          string    `xml:"dnsSuffix,attr,omitempty"`
	LinkType           int       `xml:"linkType,attr,omitempty"` // 0 = direct, 1 = routed, 2 = isolated
	ConnectedTo        string    `xml:"connectedTo,attr,omitempty"`
	Vdc                string    `xml:"vdc,attr,omitempty"`
	IsBusy             bool      `xml:"isBusy,attr,omitempty"`
	IsShared           bool      `xml:"isShared,attr,omitempty"`
	VdcName            string    `xml:"vdcName,attr,omitempty"`
	IsIpScopeInherited bool      `xml:"isIpScopeInherited,attr,omitempty"`
	Link               []*Link   `xml:"Link,omitempty"`
	Metadata           *Metadata `xml:"Metadata,omitempty"`
}

type QueryResultNsxtManagerRecordType struct {
	Xmlns      string  `xml:"xmlns,attr,omitempty"`
	Name       string  `xml:"name,attr"`
	URL        string  `xml:"url,attr"`
	HREF       string  `xml:"href,attr"`
	Site       string  `xml:"site,attr"`
	LocationId string  `xml:"locationId,attr"`
	SiteName   string  `xml:"siteName,attr"`
	Link       []*Link `xml:"Link,omitempty"`
}

// Represents org VDC Network
// Reference: vCloud API 27.0 - Network Pool
// https://code.vmware.com/apis/72/vcloud-director#/doc/doc/types/VMWNetworkPoolType.html
type VMWNetworkPool struct {
	HREF        string           `xml:"href,attr,omitempty"`
	Id          string           `xml:"id,attr,omitempty"`
	Type        string           `xml:"type,attr,omitempty"`
	Name        string           `xml:"name,attr"`
	Status      int              `xml:"status,attr,omitempty"`
	Description string           `xml:"netmask,omitempty"`
	Tasks       *TasksInProgress `xml:"Tasks,omitempty"`
}

type GroupReference struct {
	GroupReference []*Reference `xml:"GroupReference,omitempty"`
}

// Represents an org user
// Reference: vCloud API 27.0 - UserType
// https://code.vmware.com/apis/442/vcloud-director#/doc/doc/types/UserType.html
// Note that the order of fields is important. If this structure needs to change,
// the field order must be preserved.
type User struct {
	XMLName         xml.Name         `xml:"User"`
	Xmlns           string           `xml:"xmlns,attr"`
	Href            string           `xml:"href,attr"`
	Type            string           `xml:"type,attr"`
	ID              string           `xml:"id,attr"`
	OperationKey    string           `xml:"operationKey,attr"`
	Name            string           `xml:"name,attr"`
	Links           LinkList         `xml:"Link,omitempty"`
	Description     string           `xml:"Description,omitempty"`
	FullName        string           `xml:"FullName,omitempty"`
	EmailAddress    string           `xml:"EmailAddress,omitempty"`
	Telephone       string           `xml:"Telephone,omitempty"`
	IsEnabled       bool             `xml:"IsEnabled,omitempty"`
	IsLocked        bool             `xml:"IsLocked,omitempty"`
	IM              string           `xml:"IM,omitempty"`
	NameInSource    string           `xml:"NameInSource,omitempty"`
	IsExternal      bool             `xml:"IsExternal,omitempty"`
	ProviderType    string           `xml:"ProviderType,omitempty"`
	IsGroupRole     bool             `xml:"IsGroupRole,omitempty"`
	StoredVmQuota   int              `xml:"StoredVmQuota"`
	DeployedVmQuota int              `xml:"DeployedVmQuota"`
	Role            *Reference       `xml:"Role,omitempty"`
	GroupReferences *GroupReference  `xml:"GroupReferences,omitempty"`
	Password        string           `xml:"Password,omitempty"`
	Tasks           *TasksInProgress `xml:"Tasks"`
}

// Group represents Org group definition
type Group struct {
	XMLName xml.Name `xml:"Group"`
	Xmlns   string   `xml:"xmlns,attr"`
	// Id holds ID in format urn:vcloud:group:252fe08e-ae1b-409c-9dda-a531bb1ed69a
	ID string `xml:"id,attr,omitempty"`
	// Href holds reference to group object
	Href string `xml:"href,attr,omitempty"`
	// Type holds mime type for group
	Type string `xml:"type,attr"`
	// Description sets description for group
	Description string `xml:"Description"`
	// Name of the group. Cannot be updated.
	Name string `xml:"name,attr"`
	// ProviderType - 'SAML', 'INTEGRATED'
	ProviderType string `xml:"ProviderType"`
	// Role - reference to existing role
	Role *Reference `xml:"Role,omitempty"`
	// UsersList - references to existing users of type User
	UsersList *UsersList `xml:"UsersList,omitempty"`
}

// UsersList is a tagged list of User Reference's
type UsersList struct {
	UserReference []*Reference `xml:"UserReference,omitempty"`
}

// Type: AdminCatalogRecord
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/QueryResultCatalogRecordType.html
// Issue that description partly matches with what is returned
// Description: Represents Catalog record
// Since: 1.5
type CatalogRecord struct {
	HREF                    string    `xml:"href,attr,omitempty"`
	ID                      string    `xml:"id,attr,omitempty"`
	Type                    string    `xml:"type,attr,omitempty"`
	Name                    string    `xml:"name,attr,omitempty"`
	Description             string    `xml:"description,attr,omitempty"`
	IsPublished             bool      `xml:"isPublished,attr,omitempty"`
	IsShared                bool      `xml:"isShared,attr,omitempty"`
	IsLocal                 bool      `xml:"isLocal,attr,omitempty"`
	CreationDate            string    `xml:"creationDate,attr,omitempty"`
	OrgName                 string    `xml:"orgName,attr,omitempty"`
	OwnerName               string    `xml:"ownerName,attr,omitempty"`
	NumberOfVAppTemplates   int64     `xml:"numberOfVAppTemplates,attr,omitempty"`
	NumberOfMedia           int64     `xml:"numberOfMedia,attr,omitempty"`
	Owner                   string    `xml:"owner,attr,omitempty"`
	PublishSubscriptionType string    `xml:"publishSubscriptionType,attr,omitempty"`
	Version                 int64     `xml:"version,attr,omitempty"`
	Status                  string    `xml:"status,attr,omitempty"`
	Link                    *Link     `xml:"Link,omitempty"`
	Metadata                *Metadata `xml:"Metadata,omitempty"`
}

type AdminCatalogRecord CatalogRecord

// VmAffinityRule defines an affinity (or anti-affinity) rule for a group of VmReferences`
// https://code.vmware.com/apis/722/doc/doc/types/VmAffinityRuleType.html
type VmAffinityRule struct {
	XMLName         xml.Name         `xml:"VmAffinityRule"`
	Xmlns           string           `xml:"xmlns,attr"`
	HREF            string           `xml:"href,attr,omitempty"`
	ID              string           `xml:"id,attr,omitempty"`
	Name            string           `xml:"Name"`
	OperationKey    string           `xml:"OperationKey,attr,omitempty"` // Optional unique identifier to support idempotent semantics for create and delete operations
	IsEnabled       *bool            `xml:"IsEnabled"`                   // True if the affinity rule is enabled
	IsMandatory     *bool            `xml:"IsMandatory"`                 // True if this affinity rule is mandatory. When a rule is mandatory, a host failover will not power on the VM if doing so would violate the rule
	Polarity        string           `xml:"Polarity"`                    // The polarity of this rule. One of: Affinity, Anti-Affinity
	VmReferences    []*VMs           `xml:"VmReferences"`                // A list of VmReferences under a specific VM affinity rule.
	Link            []*Link          `xml:"Link,omitempty"`              //
	VCloudExtension *VCloudExtension `xml:"VCloudExtension,omitempty"`   // An optional extension element that can contain an arbitrary number of elements and attributes
}

// VmAffinityRules defines a list of VmAffinityRule
type VmAffinityRules struct {
	HREF           string            `xml:"href,attr,omitempty"`
	Type           string            `xml:"type,attr,omitempty"`
	Link           *Link             `xml:"Link,omitempty"` //
	VmAffinityRule []*VmAffinityRule `xml:"VmAffinityRule,omitempty"`
}

// ControlAccessParams specifies access controls for a resource.
type ControlAccessParams struct {
	XMLName             xml.Name           `xml:"ControlAccessParams"`
	Xmlns               string             `xml:"xmlns,attr"`
	IsSharedToEveryone  bool               `xml:"IsSharedToEveryone"`            // If true, the resource is shared with everyone in the organization. Defaults to false.
	EveryoneAccessLevel *string            `xml:"EveryoneAccessLevel,omitempty"` // If IsSharedToEveryone is true, this element must be present to specify the access level. for all members of the organization. One of: FullControl Change ReadOnly
	AccessSettings      *AccessSettingList `xml:"AccessSettings,omitempty"`      // The access settings to be applied if IsSharedToEveryone is false. Required on create and modify if IsSharedToEveryone is false.
}

// AccessSettingList is a tagged list of AccessSetting
type AccessSettingList struct {
	AccessSetting []*AccessSetting `xml:"AccessSetting"`
}

// LocalSubject is the user, group, or organization to which control access settings apply.
type LocalSubject struct {
	HREF string `xml:"href,attr"`           // Required - The URL with the full identification of the subject
	Name string `xml:"name,attr,omitempty"` // The name of the subject. Not needed in input, but it is returned on reading
	Type string `xml:"type,attr,omitempty"` // The MIME type of the subject. So far, we are using users, groups, and organizations
}

// AccessSetting controls access to the resource.
type AccessSetting struct {
	XMLName         xml.Name         `xml:"AccessSetting"`
	Subject         *LocalSubject    `xml:"Subject,omitempty"`         // The user or group to which these settings apply.
	ExternalSubject *ExternalSubject `xml:"ExternalSubject,omitempty"` // Subject existing external of VCD, to which these settings apply.
	AccessLevel     string           `xml:"AccessLevel"`               // The access level for the subject. One of: FullControl Change ReadOnly Deny (only for a VDC resource)
}

// ExternalSubjectType is a reference to a user or group managed by an identity provider configured for use in this organization.
type ExternalSubject struct {
	IdpType   string `xml:"IdpType"`   // The type of identity provider for example: OAUTH, SAML, LDAP etc for this SubjectID.
	IsUser    bool   `xml:"IsUser"`    // If true, SubjectID is a reference to a user defined by this organization's identity provider. If false or empty, SubjectID is a reference to a group defined by this organization's identity provider.
	SubjectId string `xml:"SubjectId"` // The primary key that your identity provider uses to uniquely identify the user or group referenced in SubjectId.
}

type VdcComputePolicyReferences struct {
	Xmlns                     string       `xml:"xmlns,attr"`
	HREF                      string       `xml:"href,attr,omitempty"`
	Type                      string       `xml:"type,attr,omitempty"`
	Link                      *Link        `xml:"Link,omitempty"`
	VdcComputePolicyReference []*Reference `xml:"VdcComputePolicyReference,omitempty"`
}

// Structure returned by /api/admin call
type VCloud struct {
	XMLName     xml.Name `xml:"VCloud"`
	Xmlns       string   `xml:"xmlns,attr,omitempty"`
	Name        string   `xml:"name,attr"`
	HREF        string   `xml:"href,attr"`
	Type        string   `xml:"type,attr,omitempty"`
	Description string   `xml:"Description"` // Contains VCD version, build number and build timestamp
	Link        *Link    `xml:"Link,omitempty"`
	// TODO: Add other fields if needed
	// OrganizationReferences
	// ProviderVdcReferences
	// RightReferences
	// RoleReferences
	// Networks
}

// UpdateVdcStorageProfiles is used to add a storage profile to an Org VDC or to remove one
type UpdateVdcStorageProfiles struct {
	XMLName              xml.Name                        `xml:"UpdateVdcStorageProfiles"`
	Xmlns                string                          `xml:"xmlns,attr,omitempty"`
	Name                 string                          `xml:"name,attr"`
	Description          string                          `xml:"Description,omitempty"`
	AddStorageProfile    *VdcStorageProfileConfiguration `xml:"AddStorageProfile,omitempty"`
	RemoveStorageProfile *Reference                      `xml:"RemoveStorageProfile,omitempty"`
}

// Token is used for managing VCD API Tokens for a User in an Org
type Token struct {
	ID    string            `json:"id,omitempty"`
	Name  string            `json:"name,omitempty"`
	Owner *OpenApiReference `json:"owner,omitempty"`
	Org   *OpenApiReference `json:"org,omitempty"`
	Type  string            `json:"type,omitempty"`
}

// ServiceAccount is used for managing a Service Account that belongs to a specific Org
type ServiceAccount struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	SoftwareID      string            `json:"softwareId,omitempty"`
	SoftwareVersion string            `json:"softwareVersion,omitempty"`
	Role            *OpenApiReference `json:"role,omitempty"`
	URI             string            `json:"uri,omitempty"`
	Org             *OpenApiReference `json:"org,omitempty"`
	Status          string            `json:"status,omitempty"`
}

// ApiTokenRefresh contains the access token resulting from a refresh_token operation
type ApiTokenRefresh struct {
	AccessToken  string `json:"access_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UpdatedBy    string `json:"updated_by,omitempty"`
	UpdatedOn    string `json:"updated_on,omitempty"`
}

// ApiTokenParams contains the parameters required and returned by oauth/register operation
type ApiTokenParams struct {
	ClientName              string   `json:"client_name"`
	ClientID                string   `json:"client_id,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	SoftwareID              string   `json:"software_id,omitempty"`
	SoftwareVersion         string   `json:"software_version,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
}

// ServiceAccountAuthParams is used to store the generated user code and device code that
// are needed for granting and activating a Service Account
type ServiceAccountAuthParams struct {
	DeviceCode      string `json:"device_code,omitempty"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri,omitempty"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	Interval        int    `json:"interval,omitempty"`
}

/**/
type QueryResultTaskRecordType struct {
	HREF             string    `xml:"href,attr,omitempty"`             // Contains the URI to the resource.
	ID               string    `xml:"id,attr,omitempty"`               //	The resource identifier, expressed in URN format. The value of this attribute uniquely identifies the resource, persists for the life of the resource, and is never reused. 	Yes 	Yes
	Type             string    `xml:"type,attr,omitempty"`             //	Contains the type of the resource.
	Org              string    `xml:"org,attr,omitempty"`              //	Organization reference or id
	OrgName          string    `xml:"orgName,attr,omitempty"`          //	Organization name
	Name             string    `xml:"name,attr,omitempty"`             //	The name of this task.
	OperationFull    string    `xml:"operationFull,attr,omitempty"`    //	The full human-readable name of this task.
	Message          string    `xml:"message,attr,omitempty"`          //	message
	StartDate        string    `xml:"startDate,attr,omitempty"`        //	Start date
	EndDate          string    `xml:"endDate,attr,omitempty"`          //	End date
	Status           string    `xml:"status,attr,omitempty"`           //	Status
	Progress         int       `xml:"progress,attr,omitempty"`         //	Progress of the task, expressed as a percentage.
	OwnerName        string    `xml:"ownerName,attr,omitempty"`        //	Owner name
	Object           string    `xml:"object,attr,omitempty"`           //	Object
	ObjectType       string    `xml:"objectType,attr,omitempty"`       //	Object
	ObjectName       string    `xml:"objectName,attr,omitempty"`       //	Object name
	ServiceNamespace string    `xml:"serviceNamespace,attr,omitempty"` //	Service name space
	Link             *Link     `xml:"Link,omitempty"`
	Metadata         *Metadata `xml:"Metadata,omitempty"`
}

// QueryResultOrgVdcRecordType represents an Organisation record
type QueryResultOrgRecordType struct {
	HREF               string    `xml:"href,attr,omitempty"`
	Type               string    `xml:"type,attr,omitempty"`
	ID                 string    `xml:"id,attr,omitempty"`
	Name               string    `xml:"name,attr"`
	DisplayName        string    `xml:"displayName,attr,omitempty"`
	IsEnabled          bool      `xml:"isEnabled,attr,omitempty"`
	IsReadOnly         bool      `xml:"isReadOnly,attr,omitempty"`
	CanPublishCatalogs bool      `xml:"canPublishCatalogs,attr,omitempty"`
	DeployedVMQuota    *int      `xml:"deployedVMQuota,attr,omitempty"`
	StoredVMQuota      *int      `xml:"storedVMQuota,attr,omitempty"`
	NumberOfCatalogs   *int      `xml:"numberOfCatalogs,attr,omitempty"`
	NumberOfVdcs       *int      `xml:"numberOfVdcs,attr,omitempty"`
	NumberOfVApps      *int      `xml:"numberOfVApps,attr,omitempty"`
	NumberOfGroups     *int      `xml:"numberOfGroups,attr,omitempty"`
	NumberOfDisks      *int      `xml:"numberOfDisks,attr,omitempty"`
	Link               *LinkList `xml:"Link,omitempty"`
	Metadata           *Metadata `xml:"Metadata,omitempty"`
}

// ProviderVdcCreation contains the data needed to create a provider VDC.
// Note that this is a subset of the full structure of a provider VDC.
type ProviderVdcCreation struct {
	Name                            string         `json:"name"`
	Description                     string         `json:"description"`
	ResourcePoolRefs                *VimObjectRefs `json:"resourcePoolRefs"`
	HighestSupportedHardwareVersion string         `json:"highestSupportedHardwareVersion"`
	IsEnabled                       bool           `json:"isEnabled"`
	VimServer                       []*Reference   `json:"vimServer"`
	StorageProfile                  []string       `json:"storageProfile"`
	NsxTManagerReference            *Reference     `json:"nsxTManagerReference"`
	NetworkPool                     *Reference     `json:"networkPool"`
	AutoCreateNetworkPool           bool           `json:"autoCreateNetworkPool"`
}

// AddResourcePool is used to add one or more resource pools to a provider VDC
type AddResourcePool struct {
	VimObjectRef []*VimObjectRef `xml:"AddItem" json:"addItem"`
}

// DeleteResourcePool is used to remove one or more resource pools from a provider VDC
type DeleteResourcePool struct {
	ResourcePoolRefs []*Reference `xml:"DeleteItem" json:"deleteItem"`
}

// AddStorageProfiles is used to add storage profiles to an existing provider VDC
type AddStorageProfiles struct {
	AddStorageProfile []string `json:"addStorageProfile"`
}

type EnableStorageProfile struct {
	Enabled bool `json:"enabled"`
}

type RemoveStorageProfile struct {
	RemoveStorageProfile []*Reference `json:"removeStorageProfile"`
}

// VirtualHardwareVersion describes supported hardware by the VMs created on the VDC
type VirtualHardwareVersion struct {
	HardDiskAdapter           []*HardDiskAdapter                 `xml:"HardDiskAdapter"`
	Link                      Link                               `xml:"Link"`
	MaxCPUs                   int                                `xml:"maxCPUs"`
	MaxCoresPerSocket         int                                `xml:"maxCoresPerSocket"`
	MaxMemorySizeMb           int                                `xml:"maxMemorySizeMb"`
	MaxNICs                   int                                `xml:"maxNICs"`
	Name                      string                             `xml:"name"`
	SupportedMemorySizeGb     []int                              `xml:"supportedMemorySizeGb"`
	SupportedCoresPerSocket   []int                              `xml:"supportedCoresPerSocket"`
	SupportedOperatingSystems *SupportedOperatingSystemsInfoType `xml:"supportedOperatingSystems"`

	SupportsHotAdd     *bool `xml:"supportsHotAdd"`
	SupportsHotPlugPCI *bool `xml:"supportsHotPlugPCI"`
	SupportsNestedHV   *bool `xml:"supportsNestedHV"`
}

// HardDiskAdapter describes a hard disk controller type
type HardDiskAdapter struct {
	Id                string `xml:"id,attr"`
	LegacyId          int    `xml:"legacyId,attr"`
	Name              string `xml:"name,attr"`
	MaximumDiskSizeGb int    `xml:"maximumDiskSizeGb,attr"`

	BusNumberRanges struct {
		Begin int `xml:"begin,attr"`
		End   int `xml:"end,attr"`
	} `xml:"BusNumberRanges>Range"`
	UnitNumberRanges struct {
		Begin int `xml:"begin,attr"`
		End   int `xml:"end,attr"`
	} `xml:"UnitNumberRanges>Range"`

	ReservedBusUnitNumber struct {
		BusNumber  int `xml:"busNumber,attr"`
		UnitNumber int `xml:"unitNumber,attr"`
	} `xml:"ReservedBusUnitNumber"`
}

// SupportedOperatingSystemsInfoType describes what operating system families a hardware version supports
type SupportedOperatingSystemsInfoType struct {
	Link                      *Link
	OperatingSystemFamilyInfo []*OperatingSystemFamilyInfoType `xml:"OperatingSystemFamilyInfo"`
}

// OperatingSystemFamilyInfoType describes operating systems of a given OS family
type OperatingSystemFamilyInfoType struct {
	Name                    string                     `xml:"Name"`
	OperatingSystemFamilyId *int                       `xml:"OperatingSystemFamilyId"`
	OperatingSystems        []*OperatingSystemInfoType `xml:"OperatingSystem"`
}

// OperatingSystemInfoType describes a operating system
type OperatingSystemInfoType struct {
	OperatingSystemId          *int   `xml:"OperatingSystemId,omitempty"`
	DefaultHardDiskAdapterType string `xml:"DefaultHardDiskAdapterType"`
	SupportedHardDiskAdapter   []struct {
		Ref string `xml:"ref,attr"`
	} `xml:"SupportedHardDiskAdapter,omitempty"`
	MinimumHardDiskSizeGigabytes *int   `xml:"MinimumHardDiskSizeGigabytes"`
	MinimumMemoryMegabytes       *int   `xml:"MinimumMemoryMegabytes"`
	Name                         string `xml:"Name"`
	InternalName                 string `xml:"InternalName"`
	Supported                    *bool  `xml:"Supported"`
	SupportLevel                 string `xml:"SupportLevel"`
	X64                          *bool  `xml:"x64"`
	MaximumCpuCount              *int   `xml:"MaximumCpuCount"`
	MaximumCoresPerSocket        *int   `xml:"MaximumCoresPerSocket"`
	MaximumSocketCount           *int   `xml:"MaximumSocketCount"`
	MinimumHardwareVersion       *int   `xml:"MinimumHardwareVersion"`
	PersonalizationEnabled       *bool  `xml:"PersonalizationEnabled"`
	PersonalizationAuto          *bool  `xml:"PersonalizationAuto"`
	SysprepPackagingSupported    *bool  `xml:"SysprepPackagingSupported"`
	SupportsMemHotAdd            *bool  `xml:"SupportsMemHotAdd"`
	CimOsId                      *int   `xml:"cimOsId"`
	CimVersion                   *int   `xml:"CimVersion"`
	SupportedForCreate           *bool  `xml:"SupportedForCreate"`

	RecommendedNIC []struct {
		Name string `xml:"name,attr"`
		Id   *int   `xml:"id,attr,omitempty"`
	} `xml:"RecommendedNIC"`

	SupportedNICType []struct {
		Name string `xml:"name,attr"`
		Id   *int   `xml:"id,attr,omitempty"`
	} `xml:"SupportedNICType"`

	RecommendedFirmware string   `xml:"RecommendedFirmware"`
	SupportedFirmware   []string `xml:"SupportedFirmware"`
	SupportsTPM         *bool    `xml:"SupportsTPM"`
}
