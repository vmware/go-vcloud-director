* Added support for Shareable disks, i.e., independent disks that can be attached to multiple VMs which is available from
  API v35.0 onwards. Also added uuid to the Disk structure which is a new member that is returned from v36.0 onwards. This
  member holds a UUID that can be used to correlate the disk that is attached to a particular VM from the VCD side and the
  VM host side. [GH-383]
