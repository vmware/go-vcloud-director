* Added `RegionStoragePolicy` and `types.RegionStoragePolicy` structures to read Region Storage Policies
  with methods `Region.GetAllStoragePolicies`, `Region.GetStoragePolicyByName`, `Region.GetStoragePolicyById`
  `VCDClient.GetRegionStoragePolicyById` [GH-711, GH-721]
* Added `ContentLibrary` and `types.ContentLibrary` structures to manage Content Libraries with
  methods `VCDClient.CreateContentLibrary`, `VCDClient.GetAllContentLibraries`,
  `VCDClient.GetContentLibraryByName`, `VCDClient.GetContentLibraryById`, `ContentLibrary.Update`,
  `ContentLibrary.Delete` [GH-711, GH-721, GH-735]
