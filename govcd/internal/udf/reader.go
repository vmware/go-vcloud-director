package udf

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func Open(reader io.ReaderAt) (r *ImageReader, err error) {
	defer func() {
		if fatalErr := recover(); fatalErr != nil {
			err = fmt.Errorf("%v", fatalErr)
		}
	}()
	r = &ImageReader{inner: reader}

	r.cdromVolumeDescriptors = r.readCdromVolumeDescriptorSequence(cdromVolumeDescriptorSectorNumber)

	// Validate image is UDF
	if !r.cdromVolumeDescriptors.hasAnyIdentifier(cdromVolumeIdentifierNSR02, cdromVolumeIdentifierNSR03) {
		err = fmt.Errorf("unsupported image format")
		return
	}

	r.anchorVolumeDescriptor = r.readDescriptorFromSectorWithTag(
		anchorVolumeDescriptorSectorNumber, tagAnchorVolumeDescriptorPointer).(*AnchorVolumeDescriptorPointer)
	r.mainVolumeDescriptors = r.readDescriptorSequence(&r.anchorVolumeDescriptor.MainVolumeDescriptorSequence)
	r.primaryVolume = r.mainVolumeDescriptors.get(tagPrimaryVolumeDescriptor).(*PrimaryVolumeDescriptor)
	r.partition = r.mainVolumeDescriptors.get(tagPartitionDescriptor).(*PartitionDescriptor)
	r.logicalVolume = r.mainVolumeDescriptors.get(tagLogicalVolumeDescriptor).(*LogicalVolumeDescriptor)
	r.fileSet = r.readDescriptorFromSectorWithTag(int64(r.partition.PartitionStartingLocation), tagFileSetDescriptor).(*FileSetDescriptor)
	return
}

type descriptorReader interface {
	Descriptor
	read(reader *bufferReader)
}

type ImageReader struct {
	inner                  io.ReaderAt
	cdromVolumeDescriptors CdromDescriptorList
	anchorVolumeDescriptor *AnchorVolumeDescriptorPointer
	mainVolumeDescriptors  DescriptorList
	primaryVolume          *PrimaryVolumeDescriptor
	partition              *PartitionDescriptor
	logicalVolume          *LogicalVolumeDescriptor
	fileSet                *FileSetDescriptor
}

func (r *ImageReader) RootDir() (fi *FileInfo, err error) {
	defer func() {
		if fatalErr := recover(); fatalErr != nil {
			err = fmt.Errorf("unable to read UDF root directory: %v", fatalErr)
		}
	}()
	partitionStart := int64(r.partition.PartitionStartingLocation)
	fileEntrySector := partitionStart + int64(r.fileSet.RootDirectoryICB.Location)
	fileEntry := r.readDescriptorFromSectorWithTag(fileEntrySector, tagFileEntry).(*FileEntryDescriptor)
	fi = &FileInfo{
		entry:         fileEntry,
		logicalVolume: r.logicalVolume,
	}
	return
}

func (r *ImageReader) ReadDir(parent *FileInfo) (children []FileInfo, err error) {
	defer func() {
		if fatalErr := recover(); fatalErr != nil {
			err = fmt.Errorf("unable to read UDF directory: %v", fatalErr)
		}
	}()

	if !parent.IsDir() {
		err = fmt.Errorf("entry %s is not a directory", parent.Name())
		return
	}

	if len(parent.entry.AllocationDescriptors) == 0 {
		children = []FileInfo{}
		return
	}

	children = make([]FileInfo, 0, 8)

	for idx := range parent.entry.AllocationDescriptors {
		allocDesc := &parent.entry.AllocationDescriptors[idx]
		bufSize := int(allocDesc.Length)
		partitionStart := int64(r.partition.PartitionStartingLocation)
		fileBufReader := r.readSectors(partitionStart+int64(allocDesc.Location), (bufSize+sectorSize-1)/sectorSize)

		for fileBufReader.off < bufSize {
			fid := r.readDescriptorWithTag(fileBufReader, tagFileIdentifierDescriptor).(*FileIdentifierDescriptor)
			if len(fid.FileIdentifier) > 0 {
				fileEntrySector := partitionStart + int64(fid.ICB.Location)
				fileEntry := r.readDescriptorFromSectorWithTag(fileEntrySector, tagFileEntry).(*FileEntryDescriptor)

				var filePath string
				if parent.IsRoot() {
					filePath = fid.FileIdentifier
				} else {
					filePath = filepath.Join(parent.path, fid.FileIdentifier)
				}

				children = append(children, FileInfo{
					entry:  fileEntry,
					fid:    fid,
					parent: parent,
					path:   filePath,
				})
			}
		}
	}

	return
}

func (r *ImageReader) NewFileReader(file *FileInfo) (fileReader io.Reader, err error) {
	if file.IsDir() {
		err = fmt.Errorf("entry %s is not a file", file.Name())
		return
	}

	defer func() {
		if fatalErr := recover(); fatalErr != nil {
			err = fmt.Errorf("unable to open UDF file: %v", fatalErr)
		}
	}()

	partitionStart := int64(r.partition.PartitionStartingLocation)
	readers := make([]io.Reader, len(file.entry.AllocationDescriptors))

	for idx := range file.entry.AllocationDescriptors {
		allocDesc := &file.entry.AllocationDescriptors[idx]
		location := int64(allocDesc.Location)
		offset := sectorSize * (location + partitionStart)
		readers[idx] = io.NewSectionReader(r.inner, offset, int64(allocDesc.Length))
	}

	fileReader = io.MultiReader(readers...)
	return
}

func (r *ImageReader) readDescriptorSequence(ref *Extent) DescriptorList {
	descriptors := make(DescriptorList, 0, 8)
	for sector := int64(ref.Location); ; sector++ {
		descriptor := r.readDescriptorFromSector(sector)
		descriptors = append(descriptors, descriptor)
		if descriptor.GetIdentifier() == tagTerminatingDescriptor {
			break
		}
	}
	return descriptors
}

func (r *ImageReader) readDescriptor(bufferReader *bufferReader) Descriptor {
	tagId := bufferReader.peekUint16()
	descriptor := NewDescriptor(tagId)
	descriptor.(descriptorReader).read(bufferReader)
	return descriptor
}

func (r *ImageReader) readDescriptorWithTag(bufferReader *bufferReader, tagId int) Descriptor {
	descriptor := r.readDescriptor(bufferReader)
	if descriptor.GetIdentifier() != tagId {
		panic(fmt.Errorf("expected descriptor with tag ID %d but was %d", tagId, descriptor.GetIdentifier()))
	}
	return descriptor
}

func (r *ImageReader) readDescriptorFromSector(sector int64) Descriptor {
	return r.readDescriptor(r.readSector(sector))
}

func (r *ImageReader) readDescriptorFromSectorWithTag(sector int64, tagId int) Descriptor {
	descriptor := r.readDescriptorFromSector(sector)
	if descriptor.GetIdentifier() != tagId {
		panic(fmt.Errorf("expected descriptor with tag ID %d but was %d at sector %d",
			tagId, descriptor.GetIdentifier(), sector))
	}
	return descriptor
}

func (r *ImageReader) readCdromVolumeDescriptorSequence(sector int64) CdromDescriptorList {
	descriptors := make(CdromDescriptorList, 0, 8)
	for n := sector; ; n++ {
		descriptor := r.readCdromVolumeDescriptorFromSector(n)
		descriptors = append(descriptors, descriptor)
		if descriptor.GetHeader().Identifier == cdromVolumeIdentifierTEA01 {
			break
		}
	}
	return descriptors
}

func (r *ImageReader) readCdromVolumeDescriptor(bufferReader *bufferReader) CdromDescriptor {
	header := CdromVolumeDescriptorHeader{}
	header.read(bufferReader)

	switch header.Identifier {
	case cdromVolumeIdentifierBEA01:
		return &CdromExtendedAreaVolumeDescriptor{
			Header: header,
		}
	case cdromVolumeIdentifierBOOT2:
		return &CdromBootVolumeDescriptor{
			Header: header,
		}
	case cdromVolumeIdentifierCD001:
		return &CdromCdwVolumeDescriptor{
			Header: header,
		}
	case cdromVolumeIdentifierCDW02:
		return &CdromCdwVolumeDescriptor{
			Header: header,
		}
	case cdromVolumeIdentifierNSR02:
		return &CdromNsrVolumeDescriptor{
			Header: header,
		}
	case cdromVolumeIdentifierNSR03:
		return &CdromNsrVolumeDescriptor{
			Header: header,
		}
	case cdromVolumeIdentifierTEA01:
		return &CdromTerminalVolumeDescriptor{
			Header: header,
		}
	default:
		panic(fmt.Sprintf("unrecognized file system '%s'", header.Identifier))
	}
}

func (r *ImageReader) readCdromVolumeDescriptorFromSector(sector int64) CdromDescriptor {
	return r.readCdromVolumeDescriptor(r.readSector(sector))
}

func (r *ImageReader) readSector(sector int64) *bufferReader {
	return r.read(sector*sectorSize, sectorSize)
}

func (r *ImageReader) readSectors(sector int64, count int) *bufferReader {
	return r.read(sector*sectorSize, sectorSize*count)
}

func (r *ImageReader) read(offset int64, size int) *bufferReader {
	data := make([]byte, size)
	if n, err := r.inner.ReadAt(data, offset); err != nil {
		panic(err)
	} else if n != size {
		panic(io.ErrUnexpectedEOF)
	}
	return &bufferReader{buf: data}
}

func (e *Extent) read(reader *bufferReader) {
	e.Length = reader.readUint32()
	e.Location = reader.readUint32()
}

func (e *ExtentLong) read(reader *bufferReader) {
	e.Length = reader.readUint32()
	e.Location = reader.readUint48()
	reader.skipBytes(6) // Reserved
}

func (e *EntityID) read(reader *bufferReader) {
	e.Flags = reader.readUint8()
	e.Identifier = reader.readString(23)
	e.IdentifierSuffix = reader.readString(8)
}

func (iu *ImplementationUse) read(reader *bufferReader, size int) {
	iu.Entity.read(reader)
	size = size - entityIdSize // ImplementationUse - EntityID
	if size > 0 {
		iu.Implementation = reader.readBytes(size)
	}
}

func (c *Charspec) read(reader *bufferReader) {
	c.CharacterSetType = reader.readUint8()
	c.CharacterSetInfo = reader.readBytes(63)
}

func (tag *DescriptorTag) read(reader *bufferReader) {
	tag.TagIdentifier = reader.readUint16()
	tag.DescriptorVersion = reader.readUint16()
	tag.TagChecksum = reader.readUint8()
	reader.skipBytes(1) // Reserved
	tag.TagSerialNumber = reader.readUint16()
	tag.DescriptorCRC = reader.readUint16()
	tag.DescriptorCRCLength = reader.readUint16()
	tag.TagLocation = reader.readUint32()
}

func (d *AnchorVolumeDescriptorPointer) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.MainVolumeDescriptorSequence.read(reader)
	d.ReserveVolumeDescriptorSequence.read(reader)
	reader.skipBytes(480) // Reserved
}

func (d *VolumeDescriptorPointer) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.VolumeDescriptorSequenceNumber = reader.readUint32()
	d.NextVolumeDescriptorSequenceExtent.read(reader)
	// This field shall be reserved for future standardisation and all bytes shall be set to #00
	reader.skipBytes(484)
}

func (d *PrimaryVolumeDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.VolumeDescriptorSequenceNumber = reader.readUint32()
	d.PrimaryVolumeDescriptorNumber = reader.readUint32()
	d.VolumeIdentifier = reader.readDString(32)
	d.VolumeSequenceNumber = reader.readUint16()
	d.MaximumVolumeSequenceNumber = reader.readUint16()
	d.InterchangeLevel = reader.readUint16()
	d.MaximumInterchangeLevel = reader.readUint16()
	d.CharacterSetList = reader.readUint32()
	d.MaximumCharacterSetList = reader.readUint32()
	d.VolumeSetIdentifier = reader.readDString(128)
	d.DescriptorCharacterSet.read(reader)
	d.ExplanatoryCharacterSet.read(reader)
	d.VolumeAbstract.read(reader)
	d.VolumeCopyrightNoticeExtent.read(reader)
	d.ApplicationIdentifier.read(reader)
	d.RecordingDateTime = reader.readTimestamp()
	d.ImplementationIdentifier.read(reader)
	d.ImplementationUse = reader.readBytes(64)
	d.PredecessorVolumeDescriptorSequenceLocation = reader.readUint32()
	d.Flags = reader.readUint16()
	reader.skipBytes(22) // Reserved
}

func (d *ImplementationUseVolumeDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.VolumeDescriptorSequenceNumber = reader.readUint32()
	d.ImplementationIdentifier.read(reader)
	d.ImplementationUse.read(reader)
}

func (lvi *LVInformation) read(reader *bufferReader) {
	lvi.LVICharset.read(reader)
	lvi.LogicalVolumeIdentifier = reader.readDString(128)
	lvi.LVInfo1 = reader.readDString(36)
	lvi.LVInfo2 = reader.readDString(36)
	lvi.LVInfo3 = reader.readDString(36)
	lvi.ImplementationID.read(reader)
	lvi.ImplementationUse = reader.readBytes(128)
}

func (d *PartitionDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.VolumeDescriptorSequenceNumber = reader.readUint32()
	d.PartitionFlags = reader.readUint16()
	d.PartitionNumber = reader.readUint16()
	d.PartitionContents.read(reader)
	d.PartitionContentsUse = reader.readBytes(128)
	d.AccessType = reader.readUint32()
	d.PartitionStartingLocation = reader.readUint32()
	d.PartitionLength = reader.readUint32()
	d.ImplementationIdentifier.read(reader)
	d.ImplementationUse = reader.readBytes(128)
	reader.skipBytes(156)
}

func (d *LogicalVolumeDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.VolumeDescriptorSequenceNumber = reader.readUint32()
	d.DescriptorCharacterSet.read(reader)
	d.LogicalVolumeIdentifier = reader.readDString(128)
	d.LogicalBlockSize = reader.readUint32()
	d.DomainIdentifier.read(reader)
	d.LogicalVolumeContentsUse = reader.readBytes(16)
	d.MapTableLength = reader.readUint32()
	d.NumberOfPartitionMaps = reader.readUint32()
	d.ImplementationIdentifier.read(reader)
	d.ImplementationUse = reader.readBytes(128)
	d.IntegritySequenceExtent.read(reader)
	d.PartitionMaps = make([]PartitionMap, d.NumberOfPartitionMaps)
	for idx := 0; idx < int(d.NumberOfPartitionMaps); idx++ {
		d.PartitionMaps[idx].read(reader)
	}
}

func (d *LogicalVolumeIntegrityDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.RecordingDateTime = reader.readTimestamp()
	d.IntegrityType = reader.readUint32()
	d.NextIntegrityExtent.read(reader)
	d.LogicalVolumeContentsUse.read(reader)
	d.NumberOfPartitions = reader.readUint32()
	d.LengthOfImplementationUse = reader.readUint32()
	d.FreeSpaceTable = make([]uint32, d.NumberOfPartitions)
	for idx := range d.FreeSpaceTable {
		d.FreeSpaceTable[idx] = reader.readUint32()
	}
	d.SizeTable = make([]uint32, d.NumberOfPartitions)
	for idx := range d.SizeTable {
		d.SizeTable[idx] = reader.readUint32()
	}
	d.ImplementationUse.read(reader, int(d.LengthOfImplementationUse))
}

func (h *LogicalVolumeHeaderDescriptor) read(reader *bufferReader) {
	h.UniqueID = reader.readUint64()
	reader.skipBytes(24) // Reserved
}

func (pm *PartitionMap) read(reader *bufferReader) {
	pm.PartitionMapType = reader.readUint8()
	if pm.PartitionMapType != 1 {
		panic(fmt.Sprintf("expected partition map 1 but got %d", pm.PartitionMapType))
	}
	pm.PartitionMapLength = reader.readUint8()
	if pm.PartitionMapLength != 6 {
		panic(fmt.Sprintf("expected partition map 1 to be 6 bytes long but was %d", pm.PartitionMapLength))
	}
	pm.VolumeSequenceNumber = reader.readUint16()
	pm.PartitionNumber = reader.readUint16()
}

func (d *UnallocatedSpaceDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.VolumeDescriptorSequenceNumber = reader.readUint32()
	d.NumberOfAllocationDescriptors = reader.readUint32()
	d.AllocationDescriptors = make([]Extent, d.NumberOfAllocationDescriptors)
	for idx := 0; idx < int(d.NumberOfAllocationDescriptors); idx++ {
		d.AllocationDescriptors[idx].read(reader)
	}
}

func (d *TerminatingDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	reader.skipBytes(496) // Reserved
}

func (d *FileSetDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.RecordingDateTime = reader.readTimestamp()
	d.InterchangeLevel = reader.readUint16()
	d.MaximumInterchangeLevel = reader.readUint16()
	d.CharacterSetList = reader.readUint32()
	d.MaximumCharacterSetList = reader.readUint32()
	d.FileSetNumber = reader.readUint32()
	d.FileSetDescriptorNumber = reader.readUint32()
	d.LogicalVolumeIdentifierCharacterSet.read(reader)
	d.LogicalVolumeIdentifier = reader.readDString(128)
	d.FileSetCharacterSet.read(reader)
	d.FileSetIdentifier = reader.readDString(32)
	d.CopyrightFileIdentifier = reader.readDString(32)
	d.AbstractFileIdentifier = reader.readDString(32)
	d.RootDirectoryICB.read(reader)
	d.DomainIdentifier.read(reader)
	d.NextExtent.read(reader)
	d.SystemStreamDirectoryICB.read(reader)
	reader.skipBytes(32) // Reserved
}

func (d *FileEntryDescriptor) read(reader *bufferReader) {
	d.Tag.read(reader)
	d.ICBTag.read(reader)
	d.Uid = reader.readUint32()
	d.Gid = reader.readUint32()
	d.Permissions = FileMode(reader.readUint32())
	d.FileLinkCount = reader.readUint16()
	d.RecordFormat = reader.readUint8()
	d.RecordDisplayAttributes = reader.readUint8()
	d.RecordLength = reader.readUint32()
	d.InformationLength = reader.readUint64()
	d.LogicalBlocksRecorded = reader.readUint64()
	d.AccessTime = reader.readTimestamp()
	d.ModificationTime = reader.readTimestamp()
	d.AttributeTime = reader.readTimestamp()
	d.Checkpoint = reader.readUint32()
	d.ExtendedAttributeICB.read(reader)
	d.ImplementationIdentifier.read(reader)
	d.UniqueID = reader.readUint64()
	d.LengthOfExtendedAttributes = reader.readUint32()
	d.LengthOfAllocationDescriptors = reader.readUint32()
	d.ExtendedAttributes = reader.readBytes(int(d.LengthOfExtendedAttributes))
	d.AllocationDescriptors = make([]Extent, d.LengthOfAllocationDescriptors/8)
	for i := range d.AllocationDescriptors {
		d.AllocationDescriptors[i].read(reader)
	}
}

func (icb *ICBTag) read(reader *bufferReader) {
	icb.PriorRecordedNumberOfDirectEntries = reader.readUint32()
	icb.StrategyType = reader.readUint16()
	icb.StrategyParameter = reader.readUint16()
	icb.MaximumNumberOfEntries = reader.readUint16()
	reader.skipBytes(1)
	icb.FileType = reader.readUint8()
	icb.ParentICBLocation.read(reader)
	icb.Flags = reader.readUint16()
}

func (lba *LogicalBlockAddress) read(reader *bufferReader) {
	lba.LogicalBlockNumber = reader.readUint32()
	lba.PartitionReferenceNumber = reader.readUint16()
}

func (d *FileIdentifierDescriptor) read(reader *bufferReader) {
	start := reader.off
	d.Tag.read(reader)
	d.FileVersionNumber = reader.readUint16()
	d.FileCharacteristics = FileCharacteristics(reader.readUint8())
	d.LengthOfFileIdentifier = reader.readUint8()
	d.ICB.read(reader)
	d.LengthOfImplementationUse = reader.readUint16()
	d.ImplementationUse = reader.readBytes(int(d.LengthOfImplementationUse))
	d.FileIdentifier = reader.readDCharacters(int(d.LengthOfFileIdentifier))

	// Padding: 4 x ip((L_FI+L_IU+38+3)/4) - (L_FI+L_IU+38) bytes long and shall contain all #00 bytes.
	// L_FI: Length of File Identifier
	// L_IU: Length of Implementation Use
	currentSize := reader.off - start
	paddingLen := 4*((currentSize+3)/4) - currentSize
	reader.skipBytes(paddingLen)
}

func (d *CdromVolumeDescriptorHeader) read(reader *bufferReader) {
	d.Type = reader.readUint8()
	d.Identifier = string(reader.readBytes(5))
	d.Version = reader.readUint8()
}

type bufferReader struct {
	off int
	buf []byte
}

func (r *bufferReader) eof() bool {
	return r.off >= len(r.buf)
}

func (r *bufferReader) skipBytes(len int) {
	r.off += len
}

func (r *bufferReader) readBytes(len int) []byte {
	start := r.off
	r.off += len
	return r.buf[start:r.off]
}

func (r *bufferReader) readInt8() int8 {
	return int8(r.readUint8())
}

func (r *bufferReader) readUint8() uint8 {
	start := r.off
	r.off += 1
	return r.buf[start]
}

func (r *bufferReader) readInt16() int16 {
	return int16(r.readUint16())
}

func (r *bufferReader) readUint16() uint16 {
	start := r.off
	r.off += 2
	return binary.LittleEndian.Uint16(r.buf[start:r.off])
}

func (r *bufferReader) peekUint16() uint16 {
	return binary.LittleEndian.Uint16(r.buf[r.off : r.off+2])
}

func (r *bufferReader) readInt32() int32 {
	return int32(r.readUint32())
}

func (r *bufferReader) readUint32() uint32 {
	start := r.off
	r.off += 4
	return binary.LittleEndian.Uint32(r.buf[start:r.off])
}

func (r *bufferReader) readUint48() uint64 {
	start := r.off
	r.off += 6
	data := make([]byte, 8)
	copy(data, r.buf[start:r.off])
	return binary.LittleEndian.Uint64(data)
}

func (r *bufferReader) readInt64() int64 {
	return int64(r.readUint64())
}

func (r *bufferReader) readUint64() uint64 {
	start := r.off
	r.off += 8
	return binary.LittleEndian.Uint64(r.buf[start:r.off])
}

func (r *bufferReader) readString(size int) string {
	if size == 0 {
		return ""
	}
	start := r.off
	r.off += size
	return string(r.buf[start:r.off])
}

func (r *bufferReader) readDString(size int) string {
	if size == 0 {
		return ""
	}
	start := r.off
	r.off += size
	length := int(r.buf[r.off-1])
	if length == 0 {
		return ""
	}

	// The CompressionID shall identify the compression algorithm used to compress the CompressedBitStream field.
	// 8: Value indicates there are 8 bits per character in the CompressedBitStream.
	// 16: Value indicates there are 16 bits per character in the CompressedBitStream.
	compressionId := r.buf[start]
	if compressionId != 8 {
		panic("expecting character length to be 8 bit long")
	}

	return string(r.buf[start+1 : start+length])
}

func (r *bufferReader) readDCharacters(length int) string {
	if length == 0 {
		return ""
	}
	encodingType := r.readUint8()
	length--
	switch encodingType {
	case dcharEncodingType8:
		win1252Dec := charmap.Windows1252.NewDecoder()
		s, _, err := transform.Bytes(win1252Dec, r.readBytes(length))
		if err != nil {
			panic(err)
		}
		return string(s)
	case dcharEncodingType16:
		utf16Dec := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
		s, _, err := transform.Bytes(utf16Dec, r.readBytes(length))
		if err != nil {
			panic(err)
		}
		return string(s)
	default:
		panic(fmt.Sprintf("unsupported string encoding type %d", encodingType))
	}
}

func (r *bufferReader) readTimestamp() time.Time {
	r.skipBytes(2) // TypeAndTimezone
	year := r.readUint16()
	month := r.readUint8()
	day := r.readUint8()
	hour := r.readUint8()
	minute := r.readUint8()
	second := r.readUint8()
	r.skipBytes(3) // Centiseconds+HundredsofMicroseconds+Microseconds
	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)
}

type FileInfo struct {
	fid           *FileIdentifierDescriptor
	logicalVolume *LogicalVolumeDescriptor
	entry         *FileEntryDescriptor
	parent        *FileInfo
	path          string
}

func (f *FileInfo) Name() string {
	if f.IsRoot() {
		return f.logicalVolume.LogicalVolumeIdentifier
	} else {
		return f.fid.FileIdentifier
	}
}

func (f *FileInfo) Path() string {
	return f.path
}

func (f *FileInfo) Size() int64 {
	return int64(f.entry.InformationLength)
}

func (f *FileInfo) Mode() FileMode {
	return f.entry.Permissions
}

func (f *FileInfo) FileMode() fs.FileMode {
	mode := FromFileMode(f.entry.Permissions)
	if f.IsDir() {
		mode |= fs.ModeDir
	}
	return mode
}

func (f *FileInfo) ModTime() time.Time {
	return f.entry.ModificationTime
}

func (f *FileInfo) IsRoot() bool {
	return f.fid == nil
}

func (f *FileInfo) IsDir() bool {
	return f.entry.ICBTag.FileType == fileTypeDirectory
}

func (f *FileInfo) Uid() uint32 {
	return f.entry.Uid
}

func (f *FileInfo) Gid() uint32 {
	return f.entry.Gid
}

func (f *FileInfo) Sys() any {
	return f
}
