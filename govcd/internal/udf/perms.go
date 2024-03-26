package udf

import "io/fs"

// Permissions (BP 44)
// Bit 0: Other: If set to ZERO, shall mean that the user may not execute the file; If set to ONE, shall mean that
// the user may execute the file.
// Bit 1: Other: If set to ZERO, shall mean that the user may not write the file; If set to ONE, shall mean that
// the user may write the file.
// Bit 2: Other: If set to ZERO, shall mean that the user may not read the file; If set to ONE, shall mean that the
// user may read the file.
// Bit 3: Other: If set to ZERO, shall mean that the user may not change any attributes of the file; If set to
// ONE, shall mean that the user may change attributes of the file.
// Bit 4: Other: If set to ZERO, shall mean that the user may not delete the file; If set to ONE, shall mean that
// the user may delete the file.
// Bit 5: Group: If set to ZERO, shall mean that the user may not execute the file; If set to ONE, shall mean
// that the user may execute the file.
// Bit 6: Group: If set to ZERO, shall mean that the user may not write the file; If set to ONE, shall mean that
// the user may write the file.
// Bit 7: Group: If set to ZERO, shall mean that the user may not read the file; If set to ONE, shall mean that
// the user may read the file.
// Bit 8: Group: If set to ZERO, shall mean that the user may not change any attributes of the file; If set to
// ONE, shall mean that the user may change attributes of the file.
// Bit 9: Group: If set to ZERO, shall mean that the user may not delete the file; If set to ONE, shall mean that
// the user may delete the file.
// Bit 10: Owner: If set to ZERO, shall mean that the user may not execute the file; If set to ONE, shall mean
// that the user may execute the file.
// Bit 11: Owner: If set to ZERO, shall mean that the user may not write the file; If set to ONE, shall mean that
// the user may write the file.
// Bit 12: Owner: If set to ZERO, shall mean that the user may not read the file; If set to ONE, shall mean that
// the user may read the file.
// Bit 13: Owner: If set to ZERO, shall mean that the user may not change any attributes of the file; If set to
// ONE, shall mean that the user may change attributes of the file.
// Bit 14: Owner: If set to ZERO, shall mean that the user may not delete the file; If set to ONE, shall mean that
// the user may delete the file.
// 15-31 Reserved: Shall be set to ZERO.

type FilePerm uint32

const (
	FilePermExecute FilePerm = 1 << iota
	FilePermWrite
	FilePermRead
	FilePermChange
	FilePermDelete
)

const (
	FileModeOtherOffset = 5 * iota
	FileModeGroupOffset
	FileModeOwnerOffset
)

const (
	FileModeOtherMask = ((1 << 5) - 1) << (5 * iota)
	FileModeGroupMask
	FileModeOwnerMask
)

type FileMode uint32

func ToFileMode(mode fs.FileMode) FileMode {
	mode = mode.Perm()
	var r FileMode
	r |= FileMode(mode & 7)              // Other
	r |= FileMode((mode >> 3) & 7 << 5)  // Group
	r |= FileMode((mode >> 6) & 7 << 10) // Owner
	return r
}

func FromFileMode(mode FileMode) fs.FileMode {
	var r fs.FileMode
	r |= fs.FileMode(mode & 7)                // Other
	r |= fs.FileMode(((mode >> 5) & 7) << 3)  // Group
	r |= fs.FileMode(((mode >> 10) & 7) << 6) // Owner
	return r
}

func (m FileMode) Other() FileMode {
	return m & FileModeOtherMask
}

func (m FileMode) HasOther(perms FilePerm) bool {
	return (m>>FileModeOtherOffset)&FileMode(perms) == FileMode(perms)
}

func (m FileMode) SetOther(perms FilePerm) FileMode {
	return m | (FileMode(perms) << FileModeOtherOffset)
}

func (m FileMode) UnsetOther(perms FilePerm) FileMode {
	return m &^ (FileMode(perms) << FileModeOtherOffset)
}

func (m FileMode) Group() FileMode {
	return m & FileModeGroupMask
}

func (m FileMode) HasGroup(perms FilePerm) bool {
	return (m>>FileModeGroupOffset)&FileMode(perms) == FileMode(perms)
}

func (m FileMode) SetGroup(perms FilePerm) FileMode {
	return m | (FileMode(perms) << FileModeGroupOffset)
}

func (m FileMode) UnsetGroup(perms FilePerm) FileMode {
	return m &^ (FileMode(perms) << FileModeGroupOffset)
}
func (m FileMode) Owner() FileMode {
	return m & FileModeOwnerMask
}

func (m FileMode) HasOwner(perms FilePerm) bool {
	return (m>>FileModeOwnerOffset)&FileMode(perms) == FileMode(perms)
}

func (m FileMode) SetOwner(perms FilePerm) FileMode {
	return m | (FileMode(perms) << FileModeOwnerOffset)
}

func (m FileMode) UnsetOwner(perms FilePerm) FileMode {
	return m &^ (FileMode(perms) << FileModeOwnerOffset)
}
