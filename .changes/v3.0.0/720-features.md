* Added types `RegionQuota` and `types.TmVdc` for managing Tenant Manager Org VDCs with methods
  `VCDClient.CreateRegionQuota`, `VCDClient.GetAllRegionQuotas`, `VCDClient.GetRegionQuotaByName`,
  `VCDClient.GetRegionQuotaByNameAndOrgId`, `VCDClient.GetRegionQuotaById`, `RegionQuota.Update`, `RegionQuota.Delete`,
  `VCDClient.AssignVmClassesToRegionQuota`, `RegionQuota.AssignVmClasses`, 
  `VCDClient.GetVmClassesFromRegionQuota` [GH-720, GH-738, GH-748, GH-782]
* Added types `Zone` and `types.Zone` for reading Region Zones with methods `VCDClient.GetAllZones`,
  `VCDClient.GetZoneByName`, `VCDClient.GetZoneById`, `Region.GetAllZones`, `Region.GetZoneByName`
  [GH-720]
