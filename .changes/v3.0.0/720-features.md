* Added types `TmVdc` and `types.TmVdc` for managing Tenant Manager Org VDCs with methods
  `VCDClient.CreateTmVdc`, `VCDClient.GetAllTmVdcs`, `VCDClient.GetTmVdcByName`,
  `VCDClient.GetTmVdcById`, `VCDClient.GetTmVdcByNameAndOrgId`, `TmVdc.Update`, `TmVdc.Delete`,
  `TmVdc.AssignVmClasses`, `VCDClient.GetVmClassesFromRegionQuota`, `VCDClient.AssignVmClassesToRegionQuota`
  [GH-720, GH-738, GH-748]
* Added types `Zone` and `types.Zone` for reading Region Zones with methods `VCDClient.GetAllZones`,
  `VCDClient.GetZoneByName`, `VCDClient.GetZoneById`, `Region.GetAllZones`, `Region.GetZoneByName`
  [GH-720]
