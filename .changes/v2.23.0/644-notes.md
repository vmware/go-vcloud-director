* Added internal generic functions to handle CRUD operations for inner and outer entities [GH-644]
* Added section about OpenAPI CRUD functions to `CODING_GUIDELINES.md` [GH-644] 
* Converted `DefinedEntityType`, `DefinedEntity`, `DefinedInterface`, `IpSpace`, `IpSpaceUplink`,
  `DistributedFirewall`, `DistributedFirewallRule`, `NsxtSegmentProfileTemplate`,
  `GetAllIpDiscoveryProfiles`, `GetAllMacDiscoveryProfiles`, `GetAllSpoofGuardProfiles`,
  `GetAllQoSProfiles`, `GetAllSegmentSecurityProfiles` to use newly introduced generic CRUD
  functions [GH-644]
