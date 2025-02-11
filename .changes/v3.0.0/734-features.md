* Added types `RegionQuotaStoragePolicy`, `types.VirtualDatacenterStoragePolicies` and `types.VirtualDatacenterStoragePolicy`
  to manage VDC Storage Policies, with methods `VCDClient.CreateRegionQuotaStoragePolicies`, `RegionQuota.CreateStoragePolicies`,
  `VCDClient.GetAllTmVdcStoragePolicies`, `VCDClient.GetAllRegionQuotaStoragePolicies`, `RegionQuota.GetAllStoragePolicies`,
  `VCDClient.GetTmVdcStoragePolicyById`, `RegionQuota.GetStoragePolicyByName`, `VCDClient.GetRegionQuotaStoragePolicyById`,
  `RegionQuota.GetStoragePolicyById`, `VCDClient.UpdateRegionQuotaStoragePolicy`, `RegionQuotaStoragePolicy.Update`,
  `VCDClient.DeleteRegionQuotaStoragePolicy`, `RegionQuotaStoragePolicy.Delete` [GH-734, GH-748]
