package udf

import (
	"fmt"
	"time"

	"golang.org/x/exp/utf8string"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	sectorSize                         = 2048
	anchorVolumeDescriptorSectorNumber = 256
	tagSize                            = 16
	tagVersion2                        = 2 // ECMA-167/2
	entityIdSize                       = 32
)

const (
	dcharEncodingType8  = 8
	dcharEncodingType16 = 16
)

const (
	// Indicates the contents of the specified logical volume or file set
	// is complaint with domain defined by this document.
	entityIdentifierOSTACompliant = "*OSTA UDF Compliant"
	// Contains additional Logical Volume identification information.
	entityIdentifierLVInfo = "*UDF LV Info"
	// Contains free unused space within the implementation extended attribute space.
	entityIdentifierFreeEASpace = "*UDF FreeEASpace"
	entityIdentifierDvdCGMSInfo = "*UDF DVD CGMS Info"
)

const (
	osClassUndefined uint8 = 0
	osClassMac             = 3
	osClassUnix            = 4
	osClassWindows         = 6
)

const (
	tagPrimaryVolumeDescriptor           = 0x0001
	tagAnchorVolumeDescriptorPointer     = 0x0002
	tagVolumeDescriptorPointer           = 0x0003
	tagImplementationUseVolumeDescriptor = 0x0004
	tagPartitionDescriptor               = 0x0005
	tagLogicalVolumeDescriptor           = 0x0006
	tagUnallocatedSpaceDescriptor        = 0x0007
	tagTerminatingDescriptor             = 0x0008
	tagLogicalVolumeIntegrityDescriptor  = 0x0009
	tagFileSetDescriptor                 = 0x0100
	tagFileIdentifierDescriptor          = 0x0101
	tagAllocationExtentDescriptor        = 0x0102
	tagIndirectEntry                     = 0x0103
	tagTerminalEntry                     = 0x0104
	tagFileEntry                         = 0x0105
	tagExtendedAttributeHeaderDescriptor = 0x0106
	tagUnallocatedSpaceEntry             = 0x0107
	tagSpaceBitmapDescriptor             = 0x0108
	tagPartitionIntegrityEntry           = 0x0109
	tagExtendedFileEntry                 = 0x010a
)

const (
	// Shall mean that this is an Unallocated Space Entry
	fileTypeUnallocated uint8 = 1
	// Shall mean that this is a Partition Integrity Entry
	fileTypePartitionIntegrity = 2
	// Shall mean that this is an Indirect Entry
	fileTypeIndirect = 3
	// Shall mean that the file is a directory
	fileTypeDirectory = 4
	// Shall mean that the file shall be interpreted as a sequence of bytes, each of which may be randomly accessed
	fileTypeBytes = 5
	// Shall mean that the file is a block special device file as specified by ISO/IEC 9945-1
	fileTypeBlockDevice = 6
)

type FileCharacteristics uint8

const (
	// FileCharacteristicHidden If set to ZERO, shall mean that the existence of the file shall be made known
	// to the user; If set to ONE, shall mean that the existence of the file need not be made known to the user.
	FileCharacteristicHidden FileCharacteristics = 1 << (0 + iota)

	// FileCharacteristicDirectory If set to ZERO, shall mean that the file is not a directory (see 4/14.6.6);
	// If set to ONE, shall mean that the file is a directory.
	FileCharacteristicDirectory

	// FileCharacteristicDeleted If set to ONE, shall mean this File Identifier Descriptor identifies a file that
	// has been deleted; If set to ZERO, shall mean that this File Identifier Descriptor identifies a file that
	// has not been deleted.
	FileCharacteristicDeleted

	// FileCharacteristicParent If set to ONE, shall mean that the ICB field of this descriptor identifies the
	// ICB associated with the file in which is recorded the parent directory of the directory that this
	// descriptor is recorded in;
	// If set to ZERO, shall mean that the ICB field identifies the ICB associated with the file specified
	// by this descriptor
	FileCharacteristicParent

	// FileCharacteristicMetadata If this File Identifier Descriptor is not in a stream directory,
	// this bit shall be set to ZERO. If this File Identifier Descriptor is in a stream directory,
	// a value of ZERO shall indicate that this stream contains user data. A value of ONE shall indicate that the
	// stream contains implementation use data.
	FileCharacteristicMetadata
)

func (c FileCharacteristics) HasFlags(mask FileCharacteristics) bool {
	return c&mask == mask
}

type Extent struct {
	Length   uint32
	Location uint32
}

type ExtentLong struct {
	Length   uint32
	Location uint64
}

type EntityID struct {
	Flags            uint8
	Identifier       string
	IdentifierSuffix string
}

type ImplementationUse struct {
	Entity EntityID
	// Additional vendor-specific data
	Implementation []byte
}

type Charspec struct {
	// CharacterSetType field shall have the value of 0 to indicate the CS0 coded character set.
	CharacterSetType uint8
	// CharacterSetInfo field shall contain the following byte values with the remainder of the field set to a value of 0:
	// #4F, #53, #54, #41, #20, #43, #6F, #6D, #70, #72, #65, #73, #73,
	// #65, #64, #20, #55, #6E, #69, #63, #6F, #64, #65
	CharacterSetInfo []byte
}

func NewDescriptor(tagId uint16) Descriptor {
	switch tagId {
	case tagPrimaryVolumeDescriptor:
		return &PrimaryVolumeDescriptor{}
	case tagAnchorVolumeDescriptorPointer:
		return &AnchorVolumeDescriptorPointer{}
	case tagVolumeDescriptorPointer:
		return &VolumeDescriptorPointer{}
	case tagImplementationUseVolumeDescriptor:
		return &ImplementationUseVolumeDescriptor{}
	case tagPartitionDescriptor:
		return &PartitionDescriptor{}
	case tagLogicalVolumeDescriptor:
		return &LogicalVolumeDescriptor{}
	case tagUnallocatedSpaceDescriptor:
		return &UnallocatedSpaceDescriptor{}
	case tagTerminatingDescriptor:
		return &TerminatingDescriptor{}
	case tagLogicalVolumeIntegrityDescriptor:
		return &LogicalVolumeIntegrityDescriptor{}
	case tagFileSetDescriptor:
		return &FileSetDescriptor{}
	case tagFileIdentifierDescriptor:
		return &FileIdentifierDescriptor{}
	case tagFileEntry:
		return &FileEntryDescriptor{}
	default:
		panic(fmt.Errorf("unexpected tag identifier %d", tagId))
	}
}

type Descriptor interface {
	GetIdentifier() int
	GetTag() *DescriptorTag
}

type DescriptorList []Descriptor

func (list DescriptorList) get(tagId int) Descriptor {
	if desc := list.find(tagId); desc != nil {
		return desc
	} else {
		panic(fmt.Sprintf("descriptor with ID %d does not exist in sequence", tagId))
	}
}

func (list DescriptorList) find(tagId int) Descriptor {
	for idx := 0; idx < len(list); idx++ {
		if list[idx].GetIdentifier() == tagId {
			return list[idx]
		}
	}
	return nil
}

// DescriptorTag Certain descriptors specified in Part 3 have a 16 byte structure, or tag,
// recorded at the start of the descriptor.
type DescriptorTag struct {
	TagIdentifier     uint16
	DescriptorVersion uint16
	// This field shall specify the sum modulo 256 of bytes 0-3 and 5-15 of the tag.
	TagChecksum     uint8
	TagSerialNumber uint16
	// This field shall specify the CRC of the bytes of the descriptor starting at the first byte after the descriptor tag.
	// The CRC shall be 16 bits long and be generated by the CRC-ITU-T polynomial
	DescriptorCRC uint16
	// The number of bytes shall be specified by the Descriptor CRC Length field.
	DescriptorCRCLength uint16
	TagLocation         uint32
}

// AnchorVolumeDescriptorPointer structure shall only be recorded at 2 of the following 3 locations on the media:
// * Logical Sector 256.
// * Logical Sector (N - 256).
// * N
type AnchorVolumeDescriptorPointer struct {
	Tag                             DescriptorTag
	MainVolumeDescriptorSequence    Extent
	ReserveVolumeDescriptorSequence Extent
}

func (d *AnchorVolumeDescriptorPointer) GetIdentifier() int {
	return tagAnchorVolumeDescriptorPointer
}

func (d *AnchorVolumeDescriptorPointer) GetTag() *DescriptorTag {
	return &d.Tag
}

type VolumeDescriptorPointer struct {
	Tag DescriptorTag
	// This field shall specify the Volume Descriptor Sequence Number for this descriptor.
	VolumeDescriptorSequenceNumber uint32
	// This field shall specify the next extent in the Volume Descriptor Sequence. If the extent's length is 0, no such
	//extent is specified.
	NextVolumeDescriptorSequenceExtent Extent
}

func (d *VolumeDescriptorPointer) GetIdentifier() int {
	return tagVolumeDescriptorPointer
}

func (d *VolumeDescriptorPointer) GetTag() *DescriptorTag {
	return &d.Tag
}

// PrimaryVolumeDescriptor There shall be exactly one prevailing Primary Volume Descriptor recorded per volume.
type PrimaryVolumeDescriptor struct {
	Tag                                         DescriptorTag
	VolumeDescriptorSequenceNumber              uint32
	PrimaryVolumeDescriptorNumber               uint32
	VolumeIdentifier                            string
	VolumeSequenceNumber                        uint16
	MaximumVolumeSequenceNumber                 uint16
	InterchangeLevel                            uint16
	MaximumInterchangeLevel                     uint16
	CharacterSetList                            uint32
	MaximumCharacterSetList                     uint32
	VolumeSetIdentifier                         string
	DescriptorCharacterSet                      Charspec
	ExplanatoryCharacterSet                     Charspec
	VolumeAbstract                              Extent
	VolumeCopyrightNoticeExtent                 Extent
	ApplicationIdentifier                       EntityID
	RecordingDateTime                           time.Time
	ImplementationIdentifier                    EntityID
	ImplementationUse                           []byte
	PredecessorVolumeDescriptorSequenceLocation uint32
	Flags                                       uint16
}

func (d *PrimaryVolumeDescriptor) GetIdentifier() int {
	return tagPrimaryVolumeDescriptor
}

func (d *PrimaryVolumeDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

// ImplementationUseVolumeDescriptor shall be recorded on every Volume of a Volume Set.
// The Volume may also contain additional Implementation Use Volume Descriptors which are implementation specific.
// The intended purpose of this descriptor is to aid in the identification of a Volume within a Volume Set
// that belongs to a specific Logical Volume.
type ImplementationUseVolumeDescriptor struct {
	Tag                            DescriptorTag
	VolumeDescriptorSequenceNumber uint32
	ImplementationIdentifier       EntityID
	ImplementationUse              LVInformation
}

func (d *ImplementationUseVolumeDescriptor) GetIdentifier() int {
	return tagImplementationUseVolumeDescriptor
}

func (d *ImplementationUseVolumeDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

type LVInformation struct {
	LVICharset              Charspec
	LogicalVolumeIdentifier string
	LVInfo1                 string
	LVInfo2                 string
	LVInfo3                 string
	ImplementationID        EntityID
	ImplementationUse       []byte
}

// PartitionDescriptor A Partition Access Type of Read-Only , Rewritable, Overwritable and WORM shall be supported.
// There shall be exactly one prevailing Partition Descriptor recorded per volume, with one exception.
// For Volume Sets that consist of single volume, the  volume may contain 2 Partitions with 2 prevailing
// Partition Descriptors only if one has an access type of read only and the other has an access type of Rewritable or Overwritable.
// The Logical Volume for this volume would consist of the contents of both partitions.
type PartitionDescriptor struct {
	Tag                            DescriptorTag
	VolumeDescriptorSequenceNumber uint32
	PartitionFlags                 uint16
	PartitionNumber                uint16
	PartitionContents              EntityID
	PartitionContentsUse           []byte
	AccessType                     uint32
	PartitionStartingLocation      uint32
	PartitionLength                uint32
	ImplementationIdentifier       EntityID
	ImplementationUse              []byte
}

func (d *PartitionDescriptor) GetIdentifier() int {
	return tagPartitionDescriptor
}

func (d *PartitionDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

type LogicalVolumeDescriptor struct {
	Tag                            DescriptorTag
	VolumeDescriptorSequenceNumber uint32
	DescriptorCharacterSet         Charspec
	LogicalVolumeIdentifier        string
	LogicalBlockSize               uint32
	DomainIdentifier               EntityID
	LogicalVolumeContentsUse       []byte
	MapTableLength                 uint32
	NumberOfPartitionMaps          uint32
	ImplementationIdentifier       EntityID
	ImplementationUse              []byte
	IntegritySequenceExtent        Extent
	PartitionMaps                  []PartitionMap
}

func (d *LogicalVolumeDescriptor) GetIdentifier() int {
	return tagLogicalVolumeDescriptor
}

func (d *LogicalVolumeDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

type LogicalVolumeHeaderDescriptor struct {
	UniqueID uint64
}

type PartitionMap struct {
	PartitionMapType     uint8
	PartitionMapLength   uint8
	VolumeSequenceNumber uint16
	PartitionNumber      uint16
}

// UnallocatedSpaceDescriptor shall be recorded, even if there is no free volume space.
// The first 32768 bytes of the Volume space shall not be used for the recording of ECMA 167 structures.
// This area shall not be referenced by the Unallocated Space Descriptor or any other ECMA 167 descriptor.
type UnallocatedSpaceDescriptor struct {
	Tag                            DescriptorTag
	VolumeDescriptorSequenceNumber uint32
	NumberOfAllocationDescriptors  uint32
	AllocationDescriptors          []Extent
}

func (d *UnallocatedSpaceDescriptor) GetIdentifier() int {
	return tagUnallocatedSpaceDescriptor
}

func (d *UnallocatedSpaceDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

// TerminatingDescriptor may be recorded to terminate a Volume Descriptor Sequence
type TerminatingDescriptor struct {
	Tag DescriptorTag
}

func (d *TerminatingDescriptor) GetIdentifier() int {
	return tagTerminatingDescriptor
}

func (d *TerminatingDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

type LogicalVolumeIntegrityDescriptor struct {
	Tag                       DescriptorTag
	RecordingDateTime         time.Time
	IntegrityType             uint32
	NextIntegrityExtent       Extent
	LogicalVolumeContentsUse  LogicalVolumeHeaderDescriptor
	NumberOfPartitions        uint32
	LengthOfImplementationUse uint32
	FreeSpaceTable            []uint32
	SizeTable                 []uint32
	ImplementationUse         ImplementationUse
}

func (d *LogicalVolumeIntegrityDescriptor) GetIdentifier() int {
	return tagLogicalVolumeIntegrityDescriptor
}

func (d *LogicalVolumeIntegrityDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

// FileSetDescriptor Only one File Set Descriptor shall be recorded. On WORM media, multiple File Sets may be recorded.
type FileSetDescriptor struct {
	Tag                                 DescriptorTag
	RecordingDateTime                   time.Time
	InterchangeLevel                    uint16
	MaximumInterchangeLevel             uint16
	CharacterSetList                    uint32
	MaximumCharacterSetList             uint32
	FileSetNumber                       uint32
	FileSetDescriptorNumber             uint32
	LogicalVolumeIdentifierCharacterSet Charspec
	LogicalVolumeIdentifier             string
	FileSetCharacterSet                 Charspec
	FileSetIdentifier                   string
	CopyrightFileIdentifier             string
	AbstractFileIdentifier              string
	RootDirectoryICB                    ExtentLong
	DomainIdentifier                    EntityID
	NextExtent                          ExtentLong
	SystemStreamDirectoryICB            ExtentLong
}

func (d *FileSetDescriptor) GetIdentifier() int {
	return tagFileSetDescriptor
}

func (d *FileSetDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

type FileEntryDescriptor struct {
	Tag                           DescriptorTag
	ICBTag                        ICBTag
	Uid                           uint32
	Gid                           uint32
	Permissions                   FileMode
	FileLinkCount                 uint16
	RecordFormat                  uint8
	RecordDisplayAttributes       uint8
	RecordLength                  uint32
	InformationLength             uint64
	LogicalBlocksRecorded         uint64
	AccessTime                    time.Time
	ModificationTime              time.Time
	AttributeTime                 time.Time
	Checkpoint                    uint32
	ExtendedAttributeICB          ExtentLong
	ImplementationIdentifier      EntityID
	UniqueID                      uint64
	LengthOfExtendedAttributes    uint32
	LengthOfAllocationDescriptors uint32
	ExtendedAttributes            []byte
	AllocationDescriptors         []Extent
}

func (d *FileEntryDescriptor) GetIdentifier() int {
	return tagFileEntry
}

func (d *FileEntryDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

type ICBTag struct {
	PriorRecordedNumberOfDirectEntries uint32
	StrategyType                       uint16
	StrategyParameter                  uint16
	MaximumNumberOfEntries             uint16
	FileType                           uint8
	ParentICBLocation                  LogicalBlockAddress
	Flags                              uint16
}

type LogicalBlockAddress struct {
	LogicalBlockNumber       uint32
	PartitionReferenceNumber uint16
}

// FileIdentifierDescriptor ECMA 167 4/14.4
type FileIdentifierDescriptor struct {
	Tag                       DescriptorTag
	FileVersionNumber         uint16
	FileCharacteristics       FileCharacteristics
	LengthOfFileIdentifier    uint8
	ICB                       ExtentLong
	LengthOfImplementationUse uint16
	ImplementationUse         []byte
	FileIdentifier            string
}

func (d *FileIdentifierDescriptor) GetIdentifier() int {
	return tagFileIdentifierDescriptor
}

func (d *FileIdentifierDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

func (d *FileIdentifierDescriptor) calculateSize() int {
	fileIdentifierLen := len(encodeDCharacters(d.FileIdentifier))
	size := 38 + int(d.LengthOfImplementationUse) + fileIdentifierLen
	paddingLen := 4*((size+3)/4) - size
	if paddingLen > 0 {
		size += paddingLen
	}
	return size
}

type ExtendedAttributeHeaderDescriptor struct {
	Tag                              DescriptorTag
	ImplementationAttributesLocation uint32
	ApplicationAttributesLocation    uint32
}

func (d *ExtendedAttributeHeaderDescriptor) GetIdentifier() int {
	return tagExtendedAttributeHeaderDescriptor
}

func (d *ExtendedAttributeHeaderDescriptor) GetTag() *DescriptorTag {
	return &d.Tag
}

type ImplementationUseExtendedAttribute struct {
	AttributeType            uint32
	AttributeSubtype         uint8
	AttributeLength          uint32
	ImplementationUseLength  uint32
	ImplementationIdentifier EntityID
	ImplementationData       []byte
}

func encodeDCharacters(value string) []byte {
	if len(value) == 0 {
		return []byte{}
	}
	var encodingType int
	var buf []byte
	var err error
	if utf8string.NewString(value).IsASCII() {
		encodingType = dcharEncodingType8
		win1252Enc := charmap.Windows1252.NewEncoder()
		if buf, _, err = transform.Bytes(win1252Enc, []byte(value)); err != nil {
			panic(err)
		}
	} else {
		encodingType = dcharEncodingType16
		utf16Enc := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()
		if buf, _, err = transform.Bytes(utf16Enc, []byte(value)); err != nil {
			panic(err)
		}
	}

	res := make([]byte, len(buf)+1)
	res[0] = byte(encodingType)
	copy(res[1:], buf)
	return res
}
