package types

import "encoding/xml"

// NOTE: The types in this file were created using goxmlstruct
// (https://github.com/twpayne/go-xmlstruct)

// -----------------------------------
// type FirewallConfiguration
// -----------------------------------

type FirewallConfiguration struct {
	ContextID      string          `xml:"contextId"`
	Layer3Sections *Layer3Sections `xml:"layer3Sections"`
	Layer2Sections *Layer2Sections `xml:"layer2Sections"`
}

type Layer2Sections struct {
	Section *FirewallSection `xml:"section"`
}

type Layer3Sections struct {
	Section *FirewallSection `xml:"section"`
}

type FirewallSection struct {
	XMLName          xml.Name                      `xml:"section"`
	GenerationNumber string                        `xml:"generationNumber,attr"`
	ID               *int                          `xml:"id,attr"`
	Name             string                        `xml:"name,attr"`
	Stateless        bool                          `xml:"stateless,attr"`
	TcpStrict        bool                          `xml:"tcpStrict,attr"`
	Timestamp        int                           `xml:"timestamp,attr"`
	Type             string                        `xml:"type,attr"`
	UseSid           bool                          `xml:"useSid,attr"`
	Rule             []NsxvDistributedFirewallRule `xml:"rule"`
}

type NsxvDistributedFirewallRule struct {
	Disabled      bool           `xml:"disabled,attr"`
	ID            *int           `xml:"id,attr"`
	Logged        bool           `xml:"logged,attr"`
	Name          string         `xml:"name"`
	Action        string         `xml:"action"` // allow, deny
	AppliedToList *AppliedToList `xml:"appliedToList"`
	SectionID     *int           `xml:"sectionId"`
	Sources       *Sources       `xml:"sources"`
	Destinations  *Destinations  `xml:"destinations"`
	Services      *Services      `xml:"services"`
	Direction     string         `xml:"direction"` // in, out, inout
	PacketType    string         `xml:"packetType"`
	Tag           string         `xml:"tag"`
}

type AppliedToList struct {
	AppliedTo []AppliedTo `xml:"appliedTo"`
}

type AppliedTo struct {
	Name    string `xml:"name"`
	Value   string `xml:"value"`
	Type    string `xml:"type"`
	IsValid bool   `xml:"isValid"`
}

type Sources struct {
	Excluded bool     `xml:"excluded,attr"`
	Source   []Source `xml:"source"`
}

type Source struct {
	Name    string `xml:"name"`
	Value   string `xml:"value"`
	Type    string `xml:"type"`
	IsValid bool   `xml:"isValid"`
}

type Destinations struct {
	Excluded    bool          `xml:"excluded,attr"`
	Destination []Destination `xml:"destination"`
}

type Destination struct {
	Name    string `xml:"name"`
	Value   string `xml:"value"`
	Type    string `xml:"type"`
	IsValid bool   `xml:"isValid"`
}

type Services struct {
	Service []Service `xml:"service"`
}

type Service struct {
	IsValid         bool    `xml:"isValid"`
	SourcePort      *string `xml:"sourcePort"`
	DestinationPort *string `xml:"destinationPort"`
	Protocol        *int    `xml:"protocol"`
	ProtocolName    *string `xml:"protocolName"`
	Name            string  `xml:"name,omitempty"`
	Value           string  `xml:"value,omitempty"`
	Type            string  `xml:"type,omitempty"`
}

// -----------------------------------
// type Service
// -----------------------------------

type ApplicationList struct {
	Application []Application `xml:"application"`
}

type Application struct {
	ObjectID           string          `xml:"objectId"`
	ObjectTypeName     string          `xml:"objectTypeName"`
	VsmUuid            string          `xml:"vsmUuid"`
	NodeID             string          `xml:"nodeId"`
	Revision           string          `xml:"revision"`
	Type               ApplicationType `xml:"type"`
	Name               string          `xml:"name"`
	Scope              Scope           `xml:"scope"`
	ClientHandle       struct{}        `xml:"clientHandle"`
	ExtendedAttributes struct{}        `xml:"extendedAttributes"`
	IsUniversal        bool            `xml:"isUniversal"`
	UniversalRevision  string          `xml:"universalRevision"`
	IsTemporal         bool            `xml:"isTemporal"`
	InheritanceAllowed bool            `xml:"inheritanceAllowed"`
	Element            Element         `xml:"element"`
	Layer              string          `xml:"layer"`
	IsReadOnly         bool            `xml:"isReadOnly"`
	Description        *string         `xml:"description"`
}

type ApplicationType struct {
	TypeName string `xml:"typeName"`
}

type Scope struct {
	ID             string `xml:"id"`
	ObjectTypeName string `xml:"objectTypeName"`
	Name           string `xml:"name"`
}

type Element struct {
	ApplicationProtocol *string `xml:"applicationProtocol"`
	Value               *string `xml:"value"`
	SourcePort          *string `xml:"sourcePort"`
	AppGuidName         *string `xml:"appGuidName"`
}

// -----------------------------------
// type ServiceGroup
// -----------------------------------

type ApplicationGroupList struct {
	ApplicationGroup []ApplicationGroup `xml:"applicationGroup"`
}

type ApplicationGroup struct {
	ObjectID           string          `xml:"objectId"`
	ObjectTypeName     string          `xml:"objectTypeName"`
	VsmUuid            string          `xml:"vsmUuid"`
	NodeID             string          `xml:"nodeId"`
	Revision           string          `xml:"revision"`
	Type               ApplicationType `xml:"type"`
	Name               string          `xml:"name"`
	Scope              Scope           `xml:"scope"`
	ClientHandle       struct{}        `xml:"clientHandle"`
	ExtendedAttributes struct{}        `xml:"extendedAttributes"`
	IsUniversal        bool            `xml:"isUniversal"`
	UniversalRevision  string          `xml:"universalRevision"`
	IsTemporal         bool            `xml:"isTemporal"`
	InheritanceAllowed bool            `xml:"inheritanceAllowed"`
	IsReadOnly         bool            `xml:"isReadOnly"`
	Member             []Member        `xml:"member"`
}

type Member struct {
	ObjectID           string          `xml:"objectId"`
	ObjectTypeName     string          `xml:"objectTypeName"`
	VsmUuid            string          `xml:"vsmUuid"`
	NodeID             string          `xml:"nodeId"`
	Revision           string          `xml:"revision"`
	Type               ApplicationType `xml:"type"`
	Name               string          `xml:"name"`
	Scope              Scope           `xml:"scope"`
	ClientHandle       struct{}        `xml:"clientHandle"`
	ExtendedAttributes struct{}        `xml:"extendedAttributes"`
	IsUniversal        bool            `xml:"isUniversal"`
	UniversalRevision  bool            `xml:"universalRevision"`
	IsTemporal         bool            `xml:"isTemporal"`
}
