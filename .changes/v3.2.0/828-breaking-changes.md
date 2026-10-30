* Rename `StorageCapacityMB` and `StorageConsumedMB` properties at `RegionStoragePolicy` to
  `StorageCapacityMiB` and `StorageConsumedMiB` [GH-828]
* `StorageCapacityMiB` property at `StorageClass` now uses `*int64` [GH-828]
* `StorageLimitMiB` property at `VirtualDatacenterStoragePolicy` now uses `*int64` [GH-828]
* `StorageClasses` property at `ContentLibrary` now uses `[]ContentLibraryStorageClass`
  instead of `OpenApiReferences`. Each element retains `ID` and `Name` and adds
  `KubernetesCompliantName` and `Region` properties [GH-828]
* `NeedLocalCopy` property at `ContentLibrarySubscriptionConfig` now uses `*bool` [GH-828]
* `CPULimitMHz`, `CPUReservationMHz`, `MemoryLimitMiB`, and `MemoryReservationMiB` properties
  at `TmVdcResourceAllocation` now uses `*int64` [GH-828]
* `MaxCidrCount` and `MaxIPCount` properties at `TmIpSpaceQuota` now uses `*int` [GH-828]
* `ProviderGatewayRef` property at `TmRegionalNetworkingSetting` now uses `*OpenApiReference` [GH-828]
