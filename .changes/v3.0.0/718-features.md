* Added `VCDClient.GetNsxtManagerOpenApiByUrl` method to retrieve configured NSX-T Manager entry
  based on URL [GH-718]
* Added `VCDClient.GetVCenterByUrl` method to retrieve configured vCenter server entry based on URL
  [GH-718]
* Added `VCenter.RefreshStorageProfiles` to refresh storage profiles available in vCenter server
  [GH-718]
* Added `Region` and `types.Region` to structure for OpenAPI management of Regions with methods
  `VCDClient.CreateRegion`, `VCDClient.GetAllRegions`, `VCDClient.GetRegionByName`,
  `VCDClient.GetRegionById`, `Region.Update`, `Region.Delete` [GH-718]
* Added `Supervisor` and `types.Supervisor` structure for reading available Supervisors
  `VCDClient.GetAllSupervisors`, `VCDClient.GetSupervisorById`, `VCDClient.GetSupervisorByName`,
  `Vcenter.GetAllSupervisors` [GH-718]
* Added `SupervisorZone` and `types.SupervisorZone` structure for reading available Supervisor Zones
  `Supervisor.GetAllSupervisorZones`, `Supervisor.GetSupervisorZoneById`,
  `Supervisor.GetSupervisorZoneByName`, `Supervisor.`, `VCDClient.GetAllSupervisors`,
  `VCDClient.GetSupervisorById`, `VCDClient.GetSupervisorByName`, `Vcenter.GetAllSupervisors`
  [GH-718]
