* Added `TmOrg` and `types.TmOrg` to structure for OpenAPI management of Organizations with methods
  `VCDClient.CreateTmOrg`, `VCDClient.GetAllTmOrgs`, `VCDClient.GetTmOrgByName`,
  `VCDClient.GetTmOrgById`, `TmOrg.Update`, `TmOrg.Delete`, `.TmOrg.Disable` [GH-710]
* Added `TmVdc` and `types.TmVdc` to structure for OpenAPI management of Organizations VDCs with
  methods `VCDClient.CreateTmVdc`, `VCDClient.GetAllTmVdcs`, `VCDClient.GetTmVdcByName`,
  `VCDClient.GetTmVdcById`, `TmVdc.Update`, `TmVdc.Delete`, `.TmVdc.Disable` [GH-710]
* Added types `Region` and `types.Region` for managing regions with methods
  `VCDClient.CreateRegion`, `VCDClient.GetAllRegions`, `VCDClient.GetRegionByName`,
  `VCDClient.GetRegionById`, `Region.Update`, `Region.Delete` [GH-710]
* Added types `Supervisor` and `types.Supervisor` for reading Supervisors
  `VCDClient.GetAllSupervisors`, `VCDClient.GetSupervisorById`, `VCDClient.GetSupervisorByName`
  [GH-710]
* Added types `SupervisorZone` and `types.SupervisorZone` for reading Supervisors
  `VCDClient.GetAllSupervisorZones`, `VCDClient.GetSupervisorZoneById`,
  `VCDClient.GetSupervisorZoneByName` [GH-710]
