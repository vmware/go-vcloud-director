* Added `NsxtManager` type and function `VCDClient.GetNsxtManagerByName` [GH-618]
* Added support for Segment Profile Template management using new types `NsxtSegmentProfileTemplate` and `types.NsxtSegmentProfileTemplate` [GH-618]
* Added support for reading Segment Profiles provided by NSX-T via functions
  `GetAllIpDiscoveryProfiles`, `GetIpDiscoveryProfileByName`, `GetAllMacDiscoveryProfiles`,
  `GetMacDiscoveryProfileByName`, `GetAllSpoofGuardProfiles`, `GetSpoofGuardProfileByName`,
  `GetAllQoSProfiles`, `GetQoSProfileByName`, `GetAllSegmentSecurityProfiles`,
  `GetSegmentSecurityProfileByName` [GH-618]
* Added support for setting default Segment Profiles for NSX-T Org VDC Networks
  `OpenApiOrgVdcNetwork.GetSegmentProfile()`, `OpenApiOrgVdcNetwork.UpdateSegmentProfile()` [GH-618]
* Added support for setting global default Segment Profiles
  `VCDClient.GetGlobalDefaultSegmentProfileTemplates()`,
  `VCDClient.UpdateGlobalDefaultSegmentProfileTemplates()` [GH-618]
