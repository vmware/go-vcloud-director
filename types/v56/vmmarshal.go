package types

import (
	"encoding/xml"
)

type ExtraConfigVirtualHardwareSectionMarshal struct {
	NS10 string `xml:"xmlns:ns10,attr,omitempty"`

	Info         string                        `xml:"ovf:Info"`
	Items        []*VirtualHardwareItemMarshal `xml:"ovf:Item,omitempty"`
	ExtraConfigs []*ExtraConfigMarshal         `xml:"vmw:ExtraConfig,omitempty"`
}
type ExtraConfigMarshal struct {
	Key      string `xml:"vmw:key,attr"`
	Value    string `xml:"vmw:value,attr"`
	Required bool   `xml:"ovf:required,attr"`
}

type VirtualHardwareItemMarshal struct {
	XMLName xml.Name `xml:"ovf:Item"`
	Type    string   `xml:"ns10:type,attr,omitempty"`
	Href    string   `xml:"ns10:href,attr,omitempty"`

	Address               *NillableElementMarshal               `xml:"rasd:Address"`
	AddressOnParent       *NillableElementMarshal               `xml:"rasd:AddressOnParent"`
	AllocationUnits       *NillableElementMarshal               `xml:"rasd:AllocationUnits"`
	AutomaticAllocation   *NillableElementMarshal               `xml:"rasd:AutomaticAllocation"`
	AutomaticDeallocation *NillableElementMarshal               `xml:"rasd:AutomaticDeallocation"`
	ConfigurationName     *NillableElementMarshal               `xml:"rasd:ConfigurationName"`
	Connection            []*VirtualHardwareConnectionMarshal   `xml:"rasd:Connection,omitempty"`
	ConsumerVisibility    *NillableElementMarshal               `xml:"rasd:ConsumerVisibility"`
	Description           *NillableElementMarshal               `xml:"rasd:Description"`
	ElementName           *NillableElementMarshal               `xml:"rasd:ElementName,omitempty"`
	Generation            *NillableElementMarshal               `xml:"rasd:Generation"`
	HostResource          []*VirtualHardwareHostResourceMarshal `xml:"rasd:HostResource,omitempty"`
	InstanceID            int                                   `xml:"rasd:InstanceID"`
	Limit                 *NillableElementMarshal               `xml:"rasd:Limit"`
	MappingBehavior       *NillableElementMarshal               `xml:"rasd:MappingBehavior"`
	OtherResourceType     *NillableElementMarshal               `xml:"rasd:OtherResourceType"`
	Parent                *NillableElementMarshal               `xml:"rasd:Parent"`
	PoolID                *NillableElementMarshal               `xml:"rasd:PoolID"`
	Reservation           *NillableElementMarshal               `xml:"rasd:Reservation"`
	ResourceSubType       *NillableElementMarshal               `xml:"rasd:ResourceSubType"`
	ResourceType          *NillableElementMarshal               `xml:"rasd:ResourceType"`
	VirtualQuantity       *NillableElementMarshal               `xml:"rasd:VirtualQuantity"`
	VirtualQuantityUnits  *NillableElementMarshal               `xml:"rasd:VirtualQuantityUnits"`
	Weight                *NillableElementMarshal               `xml:"rasd:Weight"`

	CoresPerSocket *CoresPerSocketMarshal `xml:"vmw:CoresPerSocket,omitempty"`
	Link           []*Link                `xml:"Link,omitempty"`
}

type NillableElementMarshal struct {
	XmlnsXsi string `xml:"xmlns:xsi,attr,omitempty"`
	XsiNil   string `xml:"xsi:nil,attr,omitempty"`
	Value    string `xml:",chardata"`
}

type CoresPerSocketMarshal struct {
	OvfRequired string `xml:"ovf:required,attr,omitempty"`
	Value       string `xml:",chardata"`
}

type VirtualHardwareConnectionMarshal struct {
	IpAddressingMode  string `xml:"ns10:ipAddressingMode,attr,omitempty"`
	IPAddress         string `xml:"ns10:ipAddress,attr,omitempty"`
	PrimaryConnection bool   `xml:"ns10:primaryNetworkConnection,attr,omitempty"`
	Value             string `xml:",chardata"`
}

type VirtualHardwareHostResourceMarshal struct {
	StorageProfile    string `xml:"ns10:storageProfileHref,attr,omitempty"`
	BusType           int    `xml:"ns10:busType,attr,omitempty"`
	BusSubType        string `xml:"ns10:busSubType,attr,omitempty"`
	Capacity          int    `xml:"ns10:capacity,attr,omitempty"`
	Iops              string `xml:"ns10:iops,attr,omitempty"`
	OverrideVmDefault string `xml:"ns10:storageProfileOverrideVmDefault,attr,omitempty"`
}
