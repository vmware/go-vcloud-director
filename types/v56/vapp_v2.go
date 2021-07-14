package types

import "encoding/xml"

type VAppV2 VApp

type ComposeVAppParamsV2 struct {
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
	Description         string                         `xml:"Description,omitempty"`         // Optional description.
	VAppParent          *Reference                     `xml:"VAppParent,omitempty"`          // Reserved. Unimplemented.
	InstantiationParams *InstantiationParams           `xml:"InstantiationParams,omitempty"` // Instantiation parameters for the composed vApp.
	SourcedItem         []*SourcedCompositionItemParam `xml:"SourcedItem,omitempty"`         // Composition items. One of: vApp vAppTemplate VM.
	CreateItem          []*CreateItem                  `xml:"CreateItem,omitempty"`          // Composition items. One of: vApp vAppTemplate VM.
	AllEULAsAccepted    bool                           `xml:"AllEULAsAccepted,omitempty"`    // True confirms acceptance of all EULAs in a vApp template. Instantiation fails if this element is missing, empty, or set to false and one or more EulaSection elements are present.
}
