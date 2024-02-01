package udf

import (
	"fmt"
	"slices"
)

const (
	cdromVolumeDescriptorSectorNumber = 16

	cdromVolumeIdentifierBEA01 = "BEA01"
	cdromVolumeIdentifierBOOT2 = "BOOT2"
	cdromVolumeIdentifierCD001 = "CD001"
	cdromVolumeIdentifierCDW02 = "CDW02"
	cdromVolumeIdentifierNSR02 = "NSR02"
	cdromVolumeIdentifierNSR03 = "NSR03"
	cdromVolumeIdentifierTEA01 = "TEA01"
)

type CdromDescriptor interface {
	GetHeader() *CdromVolumeDescriptorHeader
}

type CdromDescriptorList []CdromDescriptor

func (list CdromDescriptorList) hasAnyIdentifier(identifiers ...string) bool {
	for idx := 0; idx < len(list); idx++ {
		identifier := list[idx].GetHeader().Identifier
		if slices.Contains(identifiers, identifier) {
			return true
		}
	}
	return false
}

func (list CdromDescriptorList) getByType(descType uint8) CdromDescriptor {
	if desc := list.findByType(descType); desc != nil {
		return desc
	} else {
		panic(fmt.Sprintf("CDROM descriptor with type %d does not exist in sequence", descType))
	}
}

func (list CdromDescriptorList) findByType(descType uint8) CdromDescriptor {
	for idx := 0; idx < len(list); idx++ {
		if list[idx].GetHeader().Type == descType {
			return list[idx]
		}
	}
	return nil
}

func (list CdromDescriptorList) getByIdentifier(identifier string) CdromDescriptor {
	if desc := list.findByIdentifier(identifier); desc != nil {
		return desc
	} else {
		panic(fmt.Sprintf("CDROM descriptor with identifier %s does not exist in sequence", identifier))
	}
}

func (list CdromDescriptorList) findByIdentifier(identifier string) CdromDescriptor {
	for idx := 0; idx < len(list); idx++ {
		if list[idx].GetHeader().Identifier == identifier {
			return list[idx]
		}
	}
	return nil
}

type CdromVolumeDescriptorHeader struct {
	Type       uint8
	Identifier string
	Version    uint8
}

type CdromExtendedAreaVolumeDescriptor struct {
	Header CdromVolumeDescriptorHeader
}

func (d *CdromExtendedAreaVolumeDescriptor) GetHeader() *CdromVolumeDescriptorHeader {
	return &d.Header
}

type CdromBootVolumeDescriptor struct {
	Header CdromVolumeDescriptorHeader
}

func (d *CdromBootVolumeDescriptor) GetHeader() *CdromVolumeDescriptorHeader {
	return &d.Header
}

type CdromCdwVolumeDescriptor struct {
	Header CdromVolumeDescriptorHeader
}

func (d *CdromCdwVolumeDescriptor) GetHeader() *CdromVolumeDescriptorHeader {
	return &d.Header
}

type CdromNsrVolumeDescriptor struct {
	Header CdromVolumeDescriptorHeader
}

func (d *CdromNsrVolumeDescriptor) GetHeader() *CdromVolumeDescriptorHeader {
	return &d.Header
}

type CdromTerminalVolumeDescriptor struct {
	Header CdromVolumeDescriptorHeader
}

func (d *CdromTerminalVolumeDescriptor) GetHeader() *CdromVolumeDescriptorHeader {
	return &d.Header
}
