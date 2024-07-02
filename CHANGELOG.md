## 2.25.0 (July 2, 2024)

### FEATURES
* Added types `SolutionLandingZone` and `types.SolutionLandingZone` for Solution Add-on Landing Zone configuration ([#665](https://github.com/vmware/go-vcloud-director/pull/665))
* Added method `DefinedEntity.Refresh` to reload RDE state ([#665](https://github.com/vmware/go-vcloud-director/pull/665))
* Added `VCDClient` methods `CreateSolutionLandingZone`, `GetAllSolutionLandingZones`,
  `GetExactlyOneSolutionLandingZone`, `GetSolutionLandingZoneById` for handling Solution Landing Zones ([#665](https://github.com/vmware/go-vcloud-director/pull/665))
* Added  `SolutionLandingZone` methods `Refresh`, `RdeId`, `Update`,
  `Delete` to help handling of Solution Landing Zones ([#665](https://github.com/vmware/go-vcloud-director/pull/665))
* Added `VM` methods `GetExtraConfig`, `UpdateExtraConfig`, `DeleteExtraConfig` to manage VM extra-configuration ([#666](https://github.com/vmware/go-vcloud-director/pull/666), [#691](https://github.com/vmware/go-vcloud-director/pull/691))
* Added `Client` method `GetSite` to retrieve generic data about the current site ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `Client` methods `GetSiteAssociationData` and `GetSiteRawAssociationData` to retrieve association about from current site ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `Client` methods `GetSiteAssociations` and `QueryAllSiteAssociations` to retrieve all site associations from current site ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `Client` method `GetSiteAssociationBySiteId` to retrieve a specific site association from current site ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `Client` method `CheckSiteAssociation` to check the status of a site association ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `Client` methods `SetSiteAssociationAsync` and `SetSiteAssociation` to set a site association with current site ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `Client` methods `RemoveSiteAssociationAsync` and `RemoveSiteAssociation` to delete a site association from current site ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `Client` method `QueryAllOrgAssociations` and `GetOrgAssociations` to retrieve all org associations visible to current user ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `AdminOrg` method `GetOrgAssociationByOrgId` to retrieve a specific organization association from current org ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `AdminOrg` methods `GetOrgAssociationData` and `GetOrgRawAssociationData` to retrieve association about from current org ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `AdminOrg` method `CheckOrgAssociation` to check the status of an org association ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `AdminOrg` methods `SetOrgAssociationAsync` and `SetOrgAssociation` to set an organization association with current org ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `AdminOrg` methods `RemoveOrgAssociationAsync` and `RemoveOrgAssociation` to delete an organization association from current org ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added function `RawDataToStructuredXml` and `ReadXmlDataFromFile` to extract specific data from string or file ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `AdminOrg` methods `QueryAllOrgs`, `QueryOrgByName`, `QueryOrgByID` to query organizations ([#612](https://github.com/vmware/go-vcloud-director/pull/612),[#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added `AdminOrg` methods `GetAllOrgs` and `GetAllOpenApiOrgVdcNetworks` to retrieve organizations and networks available to current user ([#669](https://github.com/vmware/go-vcloud-director/pull/669))
* Added types `SolutionAddOn`, `SolutionAddOnConfig` and `types.SolutionAddOn` for Solution Add-on
  Landing configuration ([#670](https://github.com/vmware/go-vcloud-director/pull/670))
* Added `VCDClient` methods `CreateSolutionAddOn`, `GetAllSolutionAddons`, `GetSolutionAddonById`,
  `GetSolutionAddonByName` for handling Solution Add-Ons ([#670](https://github.com/vmware/go-vcloud-director/pull/670))
* Added  `SolutionAddOn` methods `Update`, `RdeId`, `Delete` to help handling of Solution Landing
  Zones ([#670](https://github.com/vmware/go-vcloud-director/pull/670))
* Added `VCDClient` method `TrustAddOnImageCertificate` to trust certificate if it is not yet
  trusted ([#670](https://github.com/vmware/go-vcloud-director/pull/670))
* Added `AdminOrg` methods `GetOpenIdConnectSettings`, `SetOpenIdConnectSettings` and `DeleteOpenIdConnectSettings`
  to manage OpenID Connect settings ([#671](https://github.com/vmware/go-vcloud-director/pull/671))
* Added autoscaling capabilities when creating or updating CSE Kubernetes clusters, with `CseWorkerPoolSettings.Autoscaler` 
  and `CseWorkerPoolUpdateInput.Autoscaler`, that allows to configure this mechanism on specific worker pools ([#678](https://github.com/vmware/go-vcloud-director/pull/678))
* Added types `SolutionAddOnInstance` and `types.SolutionAddOnInstance` for Solution Add-on Instance
  management ([#679](https://github.com/vmware/go-vcloud-director/pull/679))
* Added `VCDClient` methods `GetAllSolutionAddonInstanceByName`, `GetAllSolutionAddonInstances`,
  `GetSolutionAddOnInstanceById` ([#679](https://github.com/vmware/go-vcloud-director/pull/679))
* Added `SolutionAddOn` methods `CreateSolutionAddOnInstance`, `GetAllInstances`,
  `GetInstanceByName`, `ValidateInputs`, `ConvertInputTypes` ([#679](https://github.com/vmware/go-vcloud-director/pull/679))
* Added `SolutionAddOnInstance` methods `GetParentSolutionAddOn`, `ReadCreationInputValues`
  `Delete`, `RdeId`, `Publishing` ([#679](https://github.com/vmware/go-vcloud-director/pull/679))
* Added methods to create, read, update and delete VDC Templates: `VCDClient.CreateVdcTemplate`, `VCDClient.GetVdcTemplateById`,
`VCDClient.GetVdcTemplateByName`, `VdcTemplate.Update` and `VdcTemplate.Delete` ([#686](https://github.com/vmware/go-vcloud-director/pull/686))
* Added methods to manage the access settings of VDC Templates: `VdcTemplate.SetAccessControl` and  `VdcTemplate.GetAccessControl` ([#686](https://github.com/vmware/go-vcloud-director/pull/686))
* Added the `VdcTemplate.InstantiateVdcAsync` and `VdcTemplate.InstantiateVdc` methods to instantiate VDC Templates ([#686](https://github.com/vmware/go-vcloud-director/pull/686))
* Added the `VCDClient.QueryAdminVdcTemplates` and `Org.QueryVdcTemplates` methods to get all VDC Template records ([#686](https://github.com/vmware/go-vcloud-director/pull/686))
* Added types `DataSolution` and `types.DataSolution` for Data Storage Extension (DSE) management
  ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added `DataSolution` methods `RdeId`, `Name`, `Update`, `Publish`, `Unpublish`,
  `PublishRightsBundle`, `UnpublishRightsBundle`, `PublishAccessControls`,
  `UnpublishAccessControls`, `GetAllAccessControls`, `GetAllAccessControlsForTenant`,
  `GetAllInstanceTemplates`, `PublishAllInstanceTemplates`, `UnPublishAllInstanceTemplates`,
  `GetAllDataSolutionOrgConfigs`, `GetDataSolutionOrgConfigForTenant` ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added `VCDClient` methods `GetAllDataSolutions`, `GetDataSolutionById`, `GetDataSolutionByName`,
  `GetAllInstanceTemplates` ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added types `DataSolutionInstanceTemplate` and `types.DataSolutionInstanceTemplate` for Data
  Storage Extension (DSE) Solution Instance Template management ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added `DataSolutionInstanceTemplate` methods `Name`, `GetAllAccessControls`,
  `GetAllAccessControlsForTenant`, `Publish`, `Unpublish`, `RdeId` ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added types `DataSolutionOrgConfig` and `types.DataSolutionOrgConfig` for Data Storage Extension
  (DSE) Solution Instance Org Configuration management ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added `DataSolutionOrgConfig` methods `CreateDataSolutionOrgConfig`,
  `GetAllDataSolutionOrgConfigs`, `Delete`, `RdeId` ([#689](https://github.com/vmware/go-vcloud-director/pull/689))

### IMPROVEMENTS
* Improved log traceability by sending `X-VMWARE-VCLOUD-CLIENT-REQUEST-ID` header in requests. The
  header will be formatted in such format `162-2024-04-11-08-41-34-171-` where the first number
  (`162`) is the API call sequence number in the life of that particular process followed by a
  hyphen separated date time with millisecond precision (`2024-04-11-08-41-34-171` for April 11th of
  year 2024 at time 08:41:34.171). The trailing hyphen `-` is here to separate response header
  `X-Vmware-Vcloud-Request-Id` suffix with double hyphen
  `162-2024-04-11-08-41-34-171--40d78874-27a3-4cad-bd43-2764f557226b` ([#656](https://github.com/vmware/go-vcloud-director/pull/656))
* Fix bug in `Client.GetSpecificApiVersionOnCondition` that could result in using unsupported API
  version ([#658](https://github.com/vmware/go-vcloud-director/pull/658))
* Added fields `NatAndFirewallServiceIntention` and `NetworkRouteAdvertisementIntention` to
  `types.ExternalNetworkV2`, which allow users to configure NAT, Firewall and Route Advertisement
  intentions for provider gateways in VCD 10.5.1+ ([#660](https://github.com/vmware/go-vcloud-director/pull/660))
* Added field `ActionValue` to `types.NsxtFirewallRule` instead of `Action` that is deprecated in
  VCD API. It allows users to use `REJECT` option ([#661](https://github.com/vmware/go-vcloud-director/pull/661))
* Added `DetectedGuestOS` to `QueryResultVMRecordType` ([#673](https://github.com/vmware/go-vcloud-director/pull/673))
* Added convenience method `DefinedEntity.State()` that will automatically check if path to `*State` has no nil pointers ([#679](https://github.com/vmware/go-vcloud-director/pull/679))
* Added method `NsxtEdgeGateway.GetUsedAndUnusedExternalIPAddressCountWithLimit` to count used
  and unused IPs assigned to Edge Gateway. It supports a `limitTo` argument that can prevent
  exhausting system resources when counting IPs in assigned subnets ([#682](https://github.com/vmware/go-vcloud-director/pull/682))
* Added `Message` field to `types.DefinedEntity` that can return a message for an RDE ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added convenience method `DefinedEntity.State()` that can return string value of state ([#689](https://github.com/vmware/go-vcloud-director/pull/689))
* Added `DefinedEntity` methods `SetAccessControl`, `GetAllAccessControls`, `GetAccessControlById`,
  `DeleteAccessControl` for managing RDE ACLs  ([#689](https://github.com/vmware/go-vcloud-director/pull/689))

### BUG FIXES
* Fixed a bug that caused CSE Kubernetes cluster creation to fail when the configured Organization VDC Network belongs to
  a VDC Group ([#674](https://github.com/vmware/go-vcloud-director/pull/674))
* Patched `vm.UpdateNetworkConnectionSection` method that could sometimes fail due to Go's XML
  library mishandling XML namespaces when VCD returns irregular payload ([#677](https://github.com/vmware/go-vcloud-director/pull/677))
* Patched `VdcGroup.CreateDistributedFirewallRule` method that returned incorrect single rule when
  `optionalAboveRuleId` is specified ([#680](https://github.com/vmware/go-vcloud-director/pull/680))
* Patched bug in core OpenAPI handling function `getOpenApiHighestElevatedVersion` that could
  sometimes choose unsupported API versions in future VCD versions ([#683](https://github.com/vmware/go-vcloud-director/pull/683))
* Fixed an error that occurred when updating an Edge Gateway configuration, with an Edge cluster configuration section
  (`OpenAPIEdgeGatewayEdgeClusterConfig`). If this section was added, the update operation failed in VCD 10.6+ ([#688](https://github.com/vmware/go-vcloud-director/pull/688))
* Patched `vm.updateExtraConfig` method that could sometimes fail due to random mishandling of XML namespaces in upstream libraries ([#690](https://github.com/vmware/go-vcloud-director/pull/690))

### NOTES
* Patched `Test_NsxtL2VpnTunnel` to match PresharedKey of VCD 10.5.1.1+ as it started returning
  `******` instead of PSK itself when performing GET ([#659](https://github.com/vmware/go-vcloud-director/pull/659))
* Amended the test `Test_RdeAndRdeType` to be compatible with VCD 10.6+ ([#681](https://github.com/vmware/go-vcloud-director/pull/681))
* Amended many tests to set `ResourceGuaranteedMemory` when spawning a `Flex` VDC ([#684](https://github.com/vmware/go-vcloud-director/pull/684), [#685](https://github.com/vmware/go-vcloud-director/pull/685))


## 2.24.0 (April 18, 2024)

### FEATURES
* Added method `Client.QueryVappNetworks` to retrieve all vApp networks ([#657](https://github.com/vmware/go-vcloud-director/pull/657))
* Added `VApp` methods `QueryAllVappNetworks`, `QueryVappNetworks`, `QueryVappOrgNetworks` to retrieve various types of vApp networks ([#657](https://github.com/vmware/go-vcloud-director/pull/657))

### BUG FIXES
* Fixed an issue that prevented CSE Kubernetes clusters from being upgraded to an OVA with higher Kubernetes version but same TKG version,
  and to an OVA with a higher patch version of Kubernetes ([#663](https://github.com/vmware/go-vcloud-director/pull/663))
* Fixed an issue that prevented CSE Kubernetes clusters from being upgraded to TKG v2.5.0 with Kubernetes v1.26.11 as it
  performed an invalid upgrade of CoreDNS ([#663](https://github.com/vmware/go-vcloud-director/pull/663))
* Fixed an issue that prevented reading the SSH Public Key from provisioned CSE Kubernetes clusters ([#663](https://github.com/vmware/go-vcloud-director/pull/663))

## 2.23.0 (March 22, 2024)

### FEATURES
* Added the type `CseKubernetesCluster` to manage Container Service Extension Kubernetes clusters for versions 4.1.0, 4.1.1,
  4.2.0 and 4.2.1 ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added methods `Org.CseCreateKubernetesCluster` and `Org.CseCreateKubernetesClusterAsync` to create Kubernetes clusters
  in a VCD appliance with Container Service Extension installed ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added methods `VCDClient.CseGetKubernetesClusterById` and `Org.CseGetKubernetesClustersByName` to retrieve a
  Container Service Extension Kubernetes cluster ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added the method `CseKubernetesCluster.GetKubeconfig` to retrieve the *kubeconfig* of a provisioned Container Service
  Extension Kubernetes cluster ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added the method `CseKubernetesCluster.Refresh` to refresh the information and properties of an existing Container
  Service Extension Kubernetes cluster ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added methods to update a Container Service Extension Kubernetes cluster: `CseKubernetesCluster.UpdateWorkerPools`,
  `CseKubernetesCluster.AddWorkerPools`, `CseKubernetesCluster.UpdateControlPlane`, `CseKubernetesCluster.UpgradeCluster`,
  `CseKubernetesCluster.SetNodeHealthCheck` and `CseKubernetesCluster.SetAutoRepairOnErrors` ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added the method  `CseKubernetesCluster.GetSupportedUpgrades` to retrieve all the valid TKGm OVAs that a given Container
  Service Extension Kubernetes cluster can use to be upgraded ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added the method `CseKubernetesCluster.Delete` to delete a cluster ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added types `CseClusterSettings`, `CseControlPlaneSettings`, `CseWorkerPoolSettings` and `CseDefaultStorageClassSettings`
  to configure the Container Service Extension Kubernetes clusters creation process ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added types `CseClusterUpdateInput`, `CseControlPlaneUpdateInput` and `CseWorkerPoolUpdateInput` to configure the
  Container Service Extension Kubernetes clusters update process ([#645](https://github.com/vmware/go-vcloud-director/pull/645), [#653](https://github.com/vmware/go-vcloud-director/pull/653), [#655](https://github.com/vmware/go-vcloud-director/pull/655))
* Added method `client.QueryVmList` to search VMs across VDCs ([#646](https://github.com/vmware/go-vcloud-director/pull/646))

### IMPROVEMENTS
* Added missing field `vdcName` to `types.QueryResultVMRecordType` ([#646](https://github.com/vmware/go-vcloud-director/pull/646))
* Added `VCDClient.GetAllIpSpaceFloatingIpSuggestions` and `types.IpSpaceFloatingIpSuggestion` to
  retrieve IP Space IP suggestions ([#648](https://github.com/vmware/go-vcloud-director/pull/648))
* Added support for VM disk consolidation using `vm.ConsolidateDisksAsync` and `vm.ConsolidateDisks`
  ([#650](https://github.com/vmware/go-vcloud-director/pull/650))
* Added public method `VApp.GetParentVDC` to retrieve parent VDC of vApp (previously it was private)
  ([#652](https://github.com/vmware/go-vcloud-director/pull/652))
* Added methods `Catalog.CaptureVappTemplate`, `Catalog.CaptureVappTemplateAsync` and type
  `types.CaptureVAppParams` that add support for creating catalog template from existing vApp
  ([#652](https://github.com/vmware/go-vcloud-director/pull/652))
* Added method `Org.GetVAppByHref` to retrieve a vApp by given HREF ([#652](https://github.com/vmware/go-vcloud-director/pull/652))
* Added methods `VAppTemplate.GetCatalogItemHref` and `VAppTemplate.GetCatalogItemId` that can return
  related catalog item ID and HREF ([#652](https://github.com/vmware/go-vcloud-director/pull/652))

### NOTES
* Removed the conditional API call with outdated API version from `Client.GetStorageProfileByHref` so it works
  with the newest VCD versions ([#639](https://github.com/vmware/go-vcloud-director/pull/639))
* Added a delay for all LDAP tests `Test_LDAP` after LDAP configuration, but before using them
  ([#643](https://github.com/vmware/go-vcloud-director/pull/643))
 
* Added internal generic functions to handle CRUD operations for inner and outer entities ([#644](https://github.com/vmware/go-vcloud-director/pull/644))
* Added section about OpenAPI CRUD functions to `CODING_GUIDELINES.md` [[#644](https://github.com/vmware/go-vcloud-director/pull/644)] 
* Converted `DefinedEntityType`, `DefinedEntity`, `DefinedInterface`, `IpSpace`, `IpSpaceUplink`,
  `DistributedFirewall`, `DistributedFirewallRule`, `NsxtSegmentProfileTemplate`,
  `GetAllIpDiscoveryProfiles`, `GetAllMacDiscoveryProfiles`, `GetAllSpoofGuardProfiles`,
  `GetAllQoSProfiles`, `GetAllSegmentSecurityProfiles` to use newly introduced generic CRUD
  functions ([#644](https://github.com/vmware/go-vcloud-director/pull/644))

## 2.22.0 (December 12, 2023)

### FEATURES
* Added support for VMware Cloud Director **10.5.1**
* Added metadata support to Runtime Defined Entities with methods `rde.GetMetadataByKey`, `rde.GetMetadataById` `rde.GetMetadata`,
  `rde.AddMetadata` and generic metadata methods `openApiMetadataEntry.Update` and `openApiMetadataEntry.Delete` ([#557](https://github.com/vmware/go-vcloud-director/pull/557), [#632](https://github.com/vmware/go-vcloud-director/pull/632))
* Added methods `SetReadOnlyAccessControl` and `IsSharedReadOnly` for `Catalog` and `AdminCatalog`, to handle read-only catalog sharing ([#559](https://github.com/vmware/go-vcloud-director/pull/559))
* Added `Firmware` field to `VmSpecSection` type and `BootOptions` to `Vm` type ([#607](https://github.com/vmware/go-vcloud-director/pull/607))
* Added `Vdc` methods `GetHardwareVersion`, `GetHighestHardwareVersion`,
  `FindOsFromId` ([#607](https://github.com/vmware/go-vcloud-director/pull/607))
* Added `VM` methods `UpdateBootOptions`, `UpdateBootOptionsAsync` ([#607](https://github.com/vmware/go-vcloud-director/pull/607))
* API calls for `AddRawVM`, `CreateStandaloneVmAsync`, `VM.Refresh`,
  `VM.UpdateVmSpecSectionAsync`, `addEmptyVmAsyncV10`, `getVMByHrefV10`
  and `UpdateBootOptionsAsync` get elevated to API version `37.1` if available, for `VmSpecSection.Firmware` and `BootOptions` support ([#607](https://github.com/vmware/go-vcloud-director/pull/607))
* Added `VCDClient` methods `CreateNetworkPool`, `CreateStandaloneVmAsync`, `CreateNetworkPoolGeneve`, `CreateNetworkPoolVlan`, `CreateNetworkPoolPortGroup` to create a network pool ([#613](https://github.com/vmware/go-vcloud-director/pull/613))
* Added method `VCDClient.GetAllNsxtTransportZones` to retrieve all NSX-T transport zones ([#613](https://github.com/vmware/go-vcloud-director/pull/613))
* Added method `VCDClient.GetAllVcenterDistributedSwitches` to retrieve all distributed switches ([#613](https://github.com/vmware/go-vcloud-director/pull/613))
* Added method `VCDClient.QueryNsxtManagers` to retrieve all NSX-T managers ([#613](https://github.com/vmware/go-vcloud-director/pull/613))
* Added `NetworkPool` methods `Update`, `Delete`, `GetOpenApiUrl` to manage a network pool ([#613](https://github.com/vmware/go-vcloud-director/pull/613))
* Added `NsxtManager` type and function `VCDClient.GetNsxtManagerByName` ([#618](https://github.com/vmware/go-vcloud-director/pull/618))
* Added support for Segment Profile Template management using new types `NsxtSegmentProfileTemplate` and `types.NsxtSegmentProfileTemplate` ([#618](https://github.com/vmware/go-vcloud-director/pull/618))
* Added support for reading Segment Profiles provided by NSX-T via functions
  `GetAllIpDiscoveryProfiles`, `GetIpDiscoveryProfileByName`, `GetAllMacDiscoveryProfiles`,
  `GetMacDiscoveryProfileByName`, `GetAllSpoofGuardProfiles`, `GetSpoofGuardProfileByName`,
  `GetAllQoSProfiles`, `GetQoSProfileByName`, `GetAllSegmentSecurityProfiles`,
  `GetSegmentSecurityProfileByName` ([#618](https://github.com/vmware/go-vcloud-director/pull/618))
* Added support for setting default Segment Profiles for NSX-T Org VDC Networks
  `OpenApiOrgVdcNetwork.GetSegmentProfile()`, `OpenApiOrgVdcNetwork.UpdateSegmentProfile()` ([#618](https://github.com/vmware/go-vcloud-director/pull/618))
* Added support for setting global default Segment Profiles
  `VCDClient.GetGlobalDefaultSegmentProfileTemplates()`,
  `VCDClient.UpdateGlobalDefaultSegmentProfileTemplates()` ([#618](https://github.com/vmware/go-vcloud-director/pull/618))
* Added new `types` for NSX-T L2 VPN Tunnel session management `NsxtL2VpnTunnel`, `EdgeL2VpnStretchedNetwork`, `types.EdgeL2VpnTunnelStatistics`, `types.EdgeL2VpnTunnelStatus` ([#619](https://github.com/vmware/go-vcloud-director/pull/619))
* Added new `NsxtEdgeGateway` methods `CreateL2VpnTunnel`, `GetAllL2VpnTunnels`, `GetL2VpnTunnelByName`, `GetL2VpnTunnelById` for creation and retrieval of NSX-T L2 VPN Tunnel sessions ([#619](https://github.com/vmware/go-vcloud-director/pull/619))
* Added `NsxtL2VpnTunnel` methods `Refresh`, `Update`, `Statistics`, `Status`, `Delete` ([#619](https://github.com/vmware/go-vcloud-director/pull/619))
* Added method `Catalog.UploadMediaFile` to upload any file as catalog Media ([#621](https://github.com/vmware/go-vcloud-director/pull/621),[#622](https://github.com/vmware/go-vcloud-director/pull/622))
* Added method `Media.Download` to download a Media item as a byte stream ([#622](https://github.com/vmware/go-vcloud-director/pull/622))
* Added `VAppTemplate` methods `GetLease` and `RenewLease` to retrieve and change storage lease  ([#623](https://github.com/vmware/go-vcloud-director/pull/623))
* Added `NsxtEdgeGateway` methods `GetDnsConfig` and `UpdateDnsConfig` ([#627](https://github.com/vmware/go-vcloud-director/pull/627))
* Added types `types.NsxtEdgeGatewayDns`, `types.NsxtDnsForwarderZoneConfig`
  for creation and management of DNS forwarder configuration ([#627](https://github.com/vmware/go-vcloud-director/pull/627))
* Added `NsxtEdgeGatewayDns` methods `Refresh`, `Update` and `Delete` ([#627](https://github.com/vmware/go-vcloud-director/pull/627))
* Add type `VgpuProfile` and its methods `GetAllVgpuProfiles`, `GetVgpuProfilesByProviderVdc`, `GetVgpuProfileById`, `GetVgpuProfileByName`, `GetVgpuProfileByTenantFacingName`, `Update` and `Refresh` for managing vGPU profiles ([#633](https://github.com/vmware/go-vcloud-director/pull/633))
* Update `ComputePolicyV2` type with new fields for managing vGPU policies ([#633](https://github.com/vmware/go-vcloud-director/pull/633))

### IMPROVEMENTS
* Added catalog parent retrieval to `client.GetCatalogByHref` and `client.GetAdminCatalogByHref` to facilitate tenant context handling ([#559](https://github.com/vmware/go-vcloud-director/pull/559))
* Add `VdcGroup.ForceDelete` function to optionally force VDC Group removal, which can be used for
  removing VDC Group with child elements ([#597](https://github.com/vmware/go-vcloud-director/pull/597))
* Bumped up minimal API version to 37.0 (drops support for VCD 10.3.x) ([#609](https://github.com/vmware/go-vcloud-director/pull/609))
* Add struct `IopsResource` to `types.DiskSettings` (in replacement of dropped field `iops`) initially supported in API 37.0 ([#609](https://github.com/vmware/go-vcloud-director/pull/609))
* Add field `SslEnabled` to struct `types.NsxtAlbPool` initially supported in API 37.0 ([#609](https://github.com/vmware/go-vcloud-director/pull/609))
* New method `NsxtEdgeGateway.GetAllocatedIpCountByUplinkType` complementing existing
  `NsxtEdgeGateway.GetAllocatedIpCount`. It will return allocated IP counts by uplink types (works
  with VCD 10.4.1+) ([#610](https://github.com/vmware/go-vcloud-director/pull/610))
* New method `NsxtEdgeGateway.GetPrimaryNetworkAllocatedIpCount` that will return total allocated IP
  count for primary uplink (T0 or T0 VRF) ([#610](https://github.com/vmware/go-vcloud-director/pull/610))
* New field `types.EdgeGatewayUplinks.BackingType` that defines backing type of NSX-T Edge Gateway
  Uplink ([#610](https://github.com/vmware/go-vcloud-director/pull/610))
* NSX-T Edge Gateway functions `GetNsxtEdgeGatewayById`, `GetNsxtEdgeGatewayByName`,
  `GetNsxtEdgeGatewayByNameAndOwnerId`, `GetAllNsxtEdgeGateways`, `CreateNsxtEdgeGateway`,
  `Refresh`, `Update` will additionally sort uplinks to ensure that element 0 contains primary
  network (T0 or T0 VRF) ([#610](https://github.com/vmware/go-vcloud-director/pull/610))
* Makes `DefinedEntityType` method `SetBehaviorAccessControls` more robust to avoid NullPointerException errors in VCD
  when the input is a nil slice ([#615](https://github.com/vmware/go-vcloud-director/pull/615))
* `types.IpSpace` support Firewall and NAT rule autocreation configuration using
  `types.DefaultGatewayServiceConfig` on VCD 10.5.0+ ([#628](https://github.com/vmware/go-vcloud-director/pull/628))
* Added metadata ignore support for Runtime Defined Entity metadata methods `rde.GetMetadataByKey`, `rde.GetMetadata`, `rde.AddMetadata`,
  `rde.UpdateMetadata` and `rde.DeleteMetadata` ([#632](https://github.com/vmware/go-vcloud-director/pull/632))

### BUG FIXES
* Added handling of catalog creation task, which was leaving the catalog not ready for action in some cases ([#590](https://github.com/vmware/go-vcloud-director/pull/590), [#602](https://github.com/vmware/go-vcloud-director/pull/602))
* Fix nil pointer dereference bug while fetching a NSX-V Backed Edge Gateway ([#594](https://github.com/vmware/go-vcloud-director/pull/594))
* Fix vApp network related tests ([#595](https://github.com/vmware/go-vcloud-director/pull/595))
* Fixed [Issue #1098](https://github.com/vmware/terraform-provider-vcd/issues/1098) crash in VDC creation ([#598](https://github.com/vmware/go-vcloud-director/pull/598))
* Addressed Issue [1134](https://github.com/vmware/terraform-provider-vcd/issues/1134): Can't use SYSTEM `ldap_mode` ([#625](https://github.com/vmware/go-vcloud-director/pull/625))

### DEPRECATIONS
* Deprecated `UpdateInternalDisksAsync` in favor of `UpdateVmSpecSectionAsync` ([#607](https://github.com/vmware/go-vcloud-director/pull/607))

### NOTES
* Improved the testing configuration to allow customizing the UI Plugin path ([#599](https://github.com/vmware/go-vcloud-director/pull/599))
* Added a configurable timeout to the testing options available in the Makefile ([#600](https://github.com/vmware/go-vcloud-director/pull/600))
* Improved test `Test_NsxtApplicationPortProfileTenant` ([#601](https://github.com/vmware/go-vcloud-director/pull/601))
* Added explicit removal for many resources in tests ([#605](https://github.com/vmware/go-vcloud-director/pull/605))
* Improved test `Test_InsertOrEjectMedia` ([#608](https://github.com/vmware/go-vcloud-director/pull/608))
* Addressed `gosec` 2.17.0 errors ([#608](https://github.com/vmware/go-vcloud-director/pull/608))

* Changed `Test_AddNewVMFromMultiVmTemplate` to use preloaded vApp template ([#609](https://github.com/vmware/go-vcloud-director/pull/609))
* Amended `testMetadataIgnore`, used by all metadata tests, to be compatible with VCD 10.5.1 ([#629](https://github.com/vmware/go-vcloud-director/pull/629))

### REMOVALS
* Removed field `iops` from `types.DiskSettings` (dropped in API version 37.0) ([#609](https://github.com/vmware/go-vcloud-director/pull/609))


## 2.21.0 (July 20, 2023)

### FEATURES
* Added NSX-T Edge Gateway DHCP forwarding configuration support `NsxtEdgeGateway.GetDhcpForwarder` and
  `NsxtEdgeGateway.UpdateDhcpForwarder` ([#573](https://github.com/vmware/go-vcloud-director/pull/573))
* Added methods to create, get, publish and delete UI Plugins `VCDClient.AddUIPlugin`, `VCDClient.GetAllUIPlugins`,
  `VCDClient.GetUIPluginById`, `VCDClient.GetUIPlugin`, `UIPlugin.Update`, `UIPlugin.GetPublishedTenants`,
  `UIPlugin.PublishAll`, `UIPlugin.UnpublishAll`, `UIPlugin.Publish`, `UIPlugin.Unpublish`, `UIPlugin.IsTheSameAs` and `UIPlugin.Delete` ([#575](https://github.com/vmware/go-vcloud-director/pull/575))
* Added AdminOrg methods `GetFederationSettings`, `SetFederationSettings`, `UnsetFederationSettings` to handle organization SAML settings ([#576](https://github.com/vmware/go-vcloud-director/pull/576))
* Added AdminOrg methods `GetServiceProviderSamlMetadata` and `RetrieveServiceProviderSamlMetadata` to retrieve service provider metadata for current organization ([#576](https://github.com/vmware/go-vcloud-director/pull/576))
* Added method `Client.RetrieveRemoteDocument` to download a document from a URL ([#576](https://github.com/vmware/go-vcloud-director/pull/576))
* Added function `ValidateSamlServiceProviderMetadata` to validate service oprovider metadata ([#576](https://github.com/vmware/go-vcloud-director/pull/576))
* Added function `GetErrorMessageFromErrorSlice` to return a single string from a slice of errors ([#576](https://github.com/vmware/go-vcloud-director/pull/576))
* Added Service Account CRUD support via `ServiceAccount` and `types.ServiceAccount`: `VCDClient.CreateServiceAccount`,
  `Org.GetServiceAccountById`, `Org.GetAllServiceAccounts`, `Org.GetServiceAccountByName`, 
  `ServiceAccount.Update`, `ServiceAccount.Authorize`, `ServiceAccount.Grant`, `ServiceAccount.Refresh`, 
  `ServiceAccount.Revoke`, `ServiceAccount.Delete`, `ServiceAccount.GetInitialApiToken` ([#577](https://github.com/vmware/go-vcloud-director/pull/577))
* Added API Token CRUD support via `Token` and `types.Token`: `VCDClient.CreateToken`,`VCDClient.GetTokenById`,
`VCDClient.GetAllTokens`,`VCDClient.GetTokenByNameAndUsername`, `VCDClient.RegisterToken` , `Token.GetInitialApiToken`, `Token.Delete`, `Client.GetApiToken` ([#577](https://github.com/vmware/go-vcloud-director/pull/577))
* Added IP Space CRUD support via `IpSpace` and `types.IpSpace` and `VCDClient.CreateIpSpace`,
  `VCDClient.GetAllIpSpaceSummaries`, `VCDClient.GetIpSpaceById`, `VCDClient.GetIpSpaceByName`,
  `VCDClient.GetIpSpaceByNameAndOrgId`, `IpSpace.Update`, `IpSpace.Delete` ([#578](https://github.com/vmware/go-vcloud-director/pull/578))
* Added IP Space Uplink CRUD support via `IpSpaceUplink` and `types.IpSpaceUplink` and
  `VCDClient.CreateIpSpaceUplink`, `VCDClient.GetAllIpSpaceUplinks`,
  `VCDClient.GetIpSpaceUplinkById`, `VCDClient.GetIpSpaceUplinkByName`, `IpSpaceUplink.Update`,
  `IpSpaceUplink.Delete` ([#579](https://github.com/vmware/go-vcloud-director/pull/579))
* Added IP Space Allocation CRUD support via `IpSpaceIpAllocation`, `types.IpSpaceIpAllocation`,
  `types.IpSpaceIpAllocationRequest`, `types.IpSpaceIpAllocationRequestResult`. Methods
  `IpSpace.AllocateIp`, `Org.IpSpaceAllocateIp`, `Org.GetIpSpaceAllocationByTypeAndValue`,
  `IpSpace.GetAllIpSpaceAllocations`, `Org.GetIpSpaceAllocationById`, `IpSpaceIpAllocation.Update`,
  `IpSpaceIpAllocation.Delete` ([#579](https://github.com/vmware/go-vcloud-director/pull/579))
* Added IP Space Org assignment to support Custom Quotas via `IpSpaceOrgAssignment`,
  `types.IpSpaceOrgAssignment`, `IpSpace.GetAllOrgAssignments`, `IpSpace.GetOrgAssignmentById`,
  `IpSpace.GetOrgAssignmentByOrgName`,  `IpSpace.GetOrgAssignmentByOrgId`,
  `IpSpaceOrgAssignment.Update` ([#579](https://github.com/vmware/go-vcloud-director/pull/579))
* Added method `VCDClient.QueryNsxtManagerByHref` to retrieve a NSX-T manager by its ID/HREF ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added method `VCDClient.CreateProviderVdc` to create a Provider VDC ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added method `VCDClient.ResourcePoolsFromIds` to convert list of IDs to resource pools ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `ProviderVdcExtended` methods `AddResourcePools`, `AddStorageProfiles`, `Delete`, `DeleteResourcePools`,`DeleteStorageProfiles`,`Disable`,`Enable`,`GetResourcePools`,`IsEnabled`,`Rename`,`Update` to fully manage a provider VDC ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added method `NetworkPool.GetOpenApiUrl` to generate the full URL of a network pool ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `ResourcePool` methods `GetAvailableHardwareVersions` and `GetDefaultHardwareVersion` to get hardware versions ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `VCDClient` method `GetAllResourcePools` to retrieve all resource pools regardless of vCenter affiliation ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `VCDClient` method `GetAllVcenters` to retrieve all vCenters ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `VCDClient` methods `GetNetworkPoolById`,`GetNetworkPoolByName`,`GetNetworkPoolSummaries` to retrieve network pools ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `VCDClient` methods `GetVcenterById`,`GetVcenterByName` to retrieve vCenters ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `VCenter` methods `GetAllResourcePools`,`VCenter.GetResourcePoolById`,`VCenter.GetResourcePoolByName` to retrieve resource pools ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added `VCenter` methods `GetAllStorageProfiles`,`GetStorageProfileById`,`GetStorageProfileByName` to retrieve storage profiles ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added method `VCenter.GetVimServerUrl` to retrieve the full URL of a vCenter within a VCD ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Added NSX-T Edge Gateway SLAAC Profile (DHCPv6) configuration support
  `NsxtEdgeGateway.GetSlaacProfile` and `NsxtEdgeGateway.UpdateSlaacProfile` ([#582](https://github.com/vmware/go-vcloud-director/pull/582))
* Added RDE Defined Interface Behaviors support with methods `DefinedInterface.AddBehavior`, `DefinedInterface.GetAllBehaviors`,
  `DefinedInterface.GetBehaviorById` `DefinedInterface.GetBehaviorByName`, `DefinedInterface.UpdateBehavior` and
  `DefinedInterface.DeleteBehavior` ([#584](https://github.com/vmware/go-vcloud-director/pull/584))
* Added RDE Defined Entity Type Behaviors support with methods `DefinedEntityType.GetAllBehaviors`,
  `DefinedEntityType.GetBehaviorById` `DefinedEntityType.GetBehaviorByName`, `DefinedEntityType.UpdateBehaviorOverride` and
  `DefinedEntityType.DeleteBehaviorOverride` ([#584](https://github.com/vmware/go-vcloud-director/pull/584))
* Added RDE Defined Entity Type Behavior Access Controls support with methods `DefinedEntityType.GetAllBehaviorsAccessControls` and 
  `DefinedEntityType.SetBehaviorAccessControls` ([#584](https://github.com/vmware/go-vcloud-director/pull/584))
* Added method to invoke Behaviors on Defined Entities `DefinedEntity.InvokeBehavior` and `DefinedEntity.InvokeBehaviorAndMarshal` ([#584](https://github.com/vmware/go-vcloud-director/pull/584))
* Added support for NSX-T Edge Gateway Static Route configuration via types
  `NsxtEdgeGatewayStaticRoute`, `types.NsxtEdgeGatewayStaticRoute` and methods
  `NsxtEdgeGateway.CreateStaticRoute`, `NsxtEdgeGateway.GetAllStaticRoutes`,
  `NsxtEdgeGateway.GetStaticRouteByNetworkCidr`, `NsxtEdgeGateway.GetStaticRouteByName`,
  `NsxtEdgeGateway.GetStaticRouteById`, `NsxtEdgeGatewayStaticRoute.Update`,
  `NsxtEdgeGatewayStaticRoute.Delete`  ([#586](https://github.com/vmware/go-vcloud-director/pull/586))
* Added types and methods `DistributedFirewallRule`, `VdcGroup.CreateDistributedFirewallRule`,
  `DistributedFirewallRule.Update`, `.DistributedFirewallRuleDelete` to manage NSX-T Distributed
  Firewall Rules one by one (opposed to managing all at once using `DistributedFirewall`) ([#587](https://github.com/vmware/go-vcloud-director/pull/587))
* Added method `Vdc.CreateVappFromTemplate` to create a vApp from a vApp template containing one or more VMs ([#588](https://github.com/vmware/go-vcloud-director/pull/588))
* Added method `Vdc.CloneVapp` to create a vApp from another vApp ([#588](https://github.com/vmware/go-vcloud-director/pull/588))
* Added method `VApp.DiscardSuspendedState` to take a vApp out of suspended state ([#588](https://github.com/vmware/go-vcloud-director/pull/588))

### IMPROVEMENTS
* `ExternalNetworkV2` now supports IP Spaces on VCD 10.4.1+ with new fields `UsingIpSpace` and
  `DedicatedOrg` ([#579](https://github.com/vmware/go-vcloud-director/pull/579))
* Added a new function `WithIgnoredMetadata` to configure the `Client` to ignore specific metadata entries
  in all non-deprecated metadata CRUD methods ([#581](https://github.com/vmware/go-vcloud-director/pull/581))
* NSX-T ALB Virtual Service supports IPv6 Virtual Service using field`IPv6VirtualIpAddress` in
  `types.NsxtAlbVirtualService` for VCD 10.4.0+ ([#582](https://github.com/vmware/go-vcloud-director/pull/582))
* Added field `EnableDualSubnetNetwork` to enable Dual-Stack mode for Org VDC networks in
  `types.OpenApiOrgVdcNetwork` ([#582](https://github.com/vmware/go-vcloud-director/pull/582))

### BUG FIXES
* Fixed [Issue #1066](https://github.com/vmware/terraform-provider-vcd/issues/1066) - Not possible to handle more than 128 storage profiles ([#580](https://github.com/vmware/go-vcloud-director/pull/580))
* Fixed a bug that caused `Client.GetCertificateFromLibraryByName` and `AdminOrg.GetCertificateFromLibraryByName` to fail
  retrieving certificates with `:` character in the name ([#589](https://github.com/vmware/go-vcloud-director/pull/589))

### DEPRECATIONS
* Deprecated method `Vdc.InstantiateVAppTemplate` (wrong implementation and result) in favor of `Vdc.CreateVappFromTemplate` ([#588](https://github.com/vmware/go-vcloud-director/pull/588))

### NOTES
* Internal - replaced 'takeStringPointer', 'takeIntAddress', 'takeBoolPointer' with generic 'addrOf'
  ([#571](https://github.com/vmware/go-vcloud-director/pull/571))
* Changed Org enablement status during tests for VCD 10.4.2, to circumvent a VCD bug that prevents creation of disabled Orgs ([#572](https://github.com/vmware/go-vcloud-director/pull/572))
* Skipped test `Test_VdcDuplicatedVmPlacementPolicyGetsACleanError` in 10.4.2 as the relevant bug we check for is fixed in that version ([#574](https://github.com/vmware/go-vcloud-director/pull/574))
* Added `unit` step to GitHub Actions ([#583](https://github.com/vmware/go-vcloud-director/pull/583))

## 2.20.0 (April 27, 2023)

### FEATURES
* Added method `AdminVdc.IsNsxv` to detect whether an Admin VDC is NSX-V ([#521](https://github.com/vmware/go-vcloud-director/pull/521))
* Added function `NewNsxvDistributedFirewall` to create a new NSX-V distributed firewall ([#521](https://github.com/vmware/go-vcloud-director/pull/521))
* Added `NsxvDistributedFirewall` methods `GetConfiguration`, `IsEnabled`, `Enable`, `Disable`, `UpdateConfiguration`, `Refresh` to handle CRUD operations with NSX-V distributed firewalls ([#521](https://github.com/vmware/go-vcloud-director/pull/521))
* Added `NsxvDistributedFirewall` methods `GetServices`, `GetServiceGroups`, `GetServiceById`, `GetServiceByName`, `GetServiceGroupById`, `GetServiceGroupByName` to retrieve specific services or service groups ([#521](https://github.com/vmware/go-vcloud-director/pull/521))
* Added `NsxvDistributedFirewall` methods `GetServicesByRegex` and `GetServiceGroupsByRegex` to search services or service groups by regular expression ([#521](https://github.com/vmware/go-vcloud-director/pull/521))
* Added support for Runtime Defined Entity Interfaces with client methods `VCDClient.CreateDefinedInterface`, `VCDClient.GetAllDefinedInterfaces`,
  `VCDClient.GetDefinedInterface`, `VCDClient.GetDefinedInterfaceById` and methods to manipulate them `DefinedInterface.Update`,
  `DefinedInterface.Delete` ([#527](https://github.com/vmware/go-vcloud-director/pull/527), [#566](https://github.com/vmware/go-vcloud-director/pull/566))
* Added method `VM.GetEnvironment` to retrieve OVF Environment ([#528](https://github.com/vmware/go-vcloud-director/pull/528))
* Added `NsxtEdgeGateway.Refresh` method to reload NSX-T Edge Gateway structure ([#532](https://github.com/vmware/go-vcloud-director/pull/532))
* Added `NsxtEdgeGateway.GetUsedIpAddresses` method to fetch used IP addresses in NSX-T Edge
  Gateway ([#532](https://github.com/vmware/go-vcloud-director/pull/532))
* Added `NsxtEdgeGateway.GetUsedIpAddressSlice` method to fetch used IP addresses in a slice
  ([#532](https://github.com/vmware/go-vcloud-director/pull/532))
* Added `NsxtEdgeGateway.GetUnusedExternalIPAddresses` method that can help to find an unused
  IP address in an Edge Gateway by given constraints ([#532](https://github.com/vmware/go-vcloud-director/pull/532),[#567](https://github.com/vmware/go-vcloud-director/pull/567))
* Added `NsxtEdgeGateway.GetAllUnusedExternalIPAddresses` method that can return all unused IP
  addresses in an Edge Gateway ([#532](https://github.com/vmware/go-vcloud-director/pull/532),[#567](https://github.com/vmware/go-vcloud-director/pull/567))
* Added `NsxtEdgeGateway.GetAllocatedIpCount` method that sums up `TotalIPCount` fields in all
  subnets ([#532](https://github.com/vmware/go-vcloud-director/pull/532))
* Added `NsxtEdgeGateway.QuickDeallocateIpCount` and `NsxtEdgeGateway.DeallocateIpCount`
  methods to manually alter Edge Gateway body for IP deallocation ([#532](https://github.com/vmware/go-vcloud-director/pull/532))
* Added support for Runtime Defined Entity instances with methods `DefinedEntityType.GetAllRdes`, `DefinedEntityType.GetRdeByName`,
  `DefinedEntityType.GetRdeById`, `DefinedEntityType.CreateRde` and methods to manipulate them `DefinedEntity.Resolve`,
  `DefinedEntity.Update`, `DefinedEntity.Delete` ([#544](https://github.com/vmware/go-vcloud-director/pull/544))
* Added generic `Client` methods `OpenApiPostItemAndGetHeaders` and `OpenApiGetItemAndHeaders` to be able to retrieve the
  response headers when performing a POST or GET operation to an OpenAPI endpoint ([#544](https://github.com/vmware/go-vcloud-director/pull/544))
* Added support for Runtime Defined Entity Types with client methods `VCDClient.CreateRdeType`, `VCDClient.GetAllRdeTypes`,
  `VCDClient.GetRdeType`, `VCDClient.GetRdeTypeById` and methods to manipulate them `DefinedEntityType.Update`,
  `DefinedEntityType.Delete` ([#545](https://github.com/vmware/go-vcloud-director/pull/545), [#566](https://github.com/vmware/go-vcloud-director/pull/566))
* Added support for NSX-T DHCP Bindings via `OpenApiOrgVdcNetworkDhcpBinding`,
  `types.OpenApiOrgVdcNetworkDhcpBinding` and functions
  `OpenApiOrgVdcNetwork.CreateOpenApiOrgVdcNetworkDhcpBinding`,
  `OpenApiOrgVdcNetwork.GetAllOpenApiOrgVdcNetworkDhcpBindings`,
  `OpenApiOrgVdcNetwork.GetOpenApiOrgVdcNetworkDhcpBindingById`,
  `OpenApiOrgVdcNetwork.GetOpenApiOrgVdcNetworkDhcpBindingByName`,
  `OpenApiOrgVdcNetworkDhcpBinding.Update`, `OpenApiOrgVdcNetworkDhcpBinding.Refresh`,
  `OpenApiOrgVdcNetworkDhcpBinding.Delete` ([#561](https://github.com/vmware/go-vcloud-director/pull/561))
* Added QoS Profile lookup functions `GetAllNsxtEdgeGatewayQosProfiles` and
  `GetNsxtEdgeGatewayQosProfileByDisplayName` ([#563](https://github.com/vmware/go-vcloud-director/pull/563))
* Added NSX-T Edge Gateway QoS (Rate Limiting) configuration support `NsxtEdgeGateway.GetQoS` and
  `NsxtEdgeGateway.UpdateQoS` ([#563](https://github.com/vmware/go-vcloud-director/pull/563))
* Added support for importable Distributed Virtual Port Group (DVPG) read via types
  `VcenterImportableDvpg` and `types.VcenterImportableDvpg` and methods
  `VCDClient.GetVcenterImportableDvpgByName`, `VCDClient.GetAllVcenterImportableDvpgs`,
  `Vdc.GetVcenterImportableDvpgByName`, `Vdc.GetAllVcenterImportableDvpgs` ([#564](https://github.com/vmware/go-vcloud-director/pull/564))

### IMPROVEMENTS
* NSX-T ALB settings for Edge Gateway gained support for IPv6 service network definition (VCD 10.4.0+)
  and Transparent mode (VCD 10.4.1+) by adding new fields to `types.NsxtAlbConfig` and automatically
  elevating API up to 37.1 ([#549](https://github.com/vmware/go-vcloud-director/pull/549))
* Added support for using subnet prefix length while creating vApp networks ([#550](https://github.com/vmware/go-vcloud-director/pull/550))
* Improve NSX-T IPSec VPN type `types.NsxtIpSecVpnTunnel` to support 'Certificate' Authentication
  mode ([#553](https://github.com/vmware/go-vcloud-director/pull/553))
* Added new field `TransparentModeEnabled` to `types.NsxtAlbVirtualService` which allows to preserve
  client IP for NSX-T ALB Virtual Service (VCD 10.4.1+) ([#560](https://github.com/vmware/go-vcloud-director/pull/560))
* Added new field `MemberGroupRef` to `types.NsxtAlbPool` which allows to define NSX-T ALB Pool
  membership by using Edge Firewall Group (`NsxtFirewallGroup`) instead of plain IPs (VCD 10.4.1+)
  ([#560](https://github.com/vmware/go-vcloud-director/pull/560))
* `types.OpenApiOrgVdcNetwork` got a new read only field `OrgVdcIsNsxTBacked` (available since API
  36.0) which indicates if an Org Network is backed by NSX-T and a function
  `OpenApiOrgVdcNetwork.IsNsxt()` ([#561](https://github.com/vmware/go-vcloud-director/pull/561))
* Added `SetServiceAccountApiToken` method of `VCDClient` that allows
  authenticating using a service account token file and handles the refresh token rotation ([#562](https://github.com/vmware/go-vcloud-director/pull/562))

### BUG FIXES
* Fixed a bug that prevented returning a specific error while authenticating client with invalid
  password ([#536](https://github.com/vmware/go-vcloud-director/pull/536))
* Fixed accessing uninitialized `Features` field while updating a vApp network ([#550](https://github.com/vmware/go-vcloud-director/pull/550))

### NOTES
* Created `Test_RenameCatalog` for making sure the contents of the Catalog don't change after rename ([#546](https://github.com/vmware/go-vcloud-director/pull/546))

## 2.19.0 (January 12, 2023)

### FEATURES
* Added client methods `GetCatalogByHref`, `GetCatalogById`, `GetCatalogByName` to retrieve Catalogs without an Org object ([#537](https://github.com/vmware/go-vcloud-director/pull/537))
* Added client methods `GetAdminCatalogByHref`, `GetAdminCatalogById`, `GetAdminCatalogByName` to retrieve AdminCatalogs without an AdminOrg object ([#537](https://github.com/vmware/go-vcloud-director/pull/537))
* Added method `VAppTemplate.GetVappTemplateRecord` to retrieve a VAppTemplate query record ([#537](https://github.com/vmware/go-vcloud-director/pull/537))

### BUG FIXES
* Removed URL checks from `CreateCatalogFromSubscriptionAsync` to allow catalog creation from non-VCD entities, such as vSphere shared library ([#537](https://github.com/vmware/go-vcloud-director/pull/537))
* Fixed flaky test `Test_CatalogAccessAsOrgUsers` that failed randomly for timing issues ([#540](https://github.com/vmware/go-vcloud-director/pull/540))

### NOTES
* Amended a quirky test `Test_CreateOrgVdcWithFlex` that failed randomly due to recovered VDC Storage Profiles being unordered ([#538](https://github.com/vmware/go-vcloud-director/pull/538))
* Amended a quirky test `Test_VMPowerOnPowerOff` that failed due to the testing VM not being powered off fast enough ([#538](https://github.com/vmware/go-vcloud-director/pull/538))

## 2.18.0 (December 14, 2022)

### FEATURES
* Added `VCDClient.GetAllAssignedVdcComputePoliciesV2` to retrieve Compute Policies without the need of an `AdminVdc` receiver ([#530](https://github.com/vmware/go-vcloud-director/pull/530))
* Added `client` methods `QueryCatalogRecords` and `GetCatalogByHref` ([#531](https://github.com/vmware/go-vcloud-director/pull/531))

### BUG FIXES
* Fixed issue that caused VM Group retrieval to fail if the Provider VDC had more than one Resource Pool ([#530](https://github.com/vmware/go-vcloud-director/pull/530))
* Fixed issue that prevented Org update because of wrong field position in LDAP settings ([#533](https://github.com/vmware/go-vcloud-director/pull/533))

## 2.17.0 (November 25, 2022)

### FEATURES
* Added new functions to get vApp Templates `Catalog.GetVAppTemplateByName`, `Catalog.GetVAppTemplateById`, `Catalog.GetVAppTemplateByNameOrId`,
  `Vdc.GetVAppTemplateByName`, `VCDClient.GetVAppTemplateByHref` and `VCDClient.GetVAppTemplateById` ([#495](https://github.com/vmware/go-vcloud-director/pull/495), [#520](https://github.com/vmware/go-vcloud-director/pull/520))
* Added new functions to query vApp Templates by name `Catalog.QueryVappTemplateWithName`, `Vdc.QueryVappTemplateWithName`, `AdminVdc.QueryVappTemplateWithName` ([#495](https://github.com/vmware/go-vcloud-director/pull/495))
* Added new functions to delete vApp Templates `VAppTemplate.DeleteAsync` and `VAppTemplate.Delete` ([#495](https://github.com/vmware/go-vcloud-director/pull/495))
* Added new functions to extract information from vApp Templates `VAppTemplate.GetCatalogName` and `VAppTemplate.GetVdcName` ([#495](https://github.com/vmware/go-vcloud-director/pull/495))
* Added `Client.TestConnection` method to check remote VCD endpoints ([#447](https://github.com/vmware/go-vcloud-director/pull/447), [#501](https://github.com/vmware/go-vcloud-director/pull/501))
* Added `Client.TestConnectionWithDefaults` method that uses `Client.TestConnection` with some default
  values  ([#447](https://github.com/vmware/go-vcloud-director/pull/447), [#501](https://github.com/vmware/go-vcloud-director/pull/501))
* Changed behavior of `Client.OpenApiPostItem` and `Client.OpenApiPostItemSync` so they accept
  response code 200 OK as valid. The reason is `TestConnection` endpoint requires a POST request and
  returns a 200OK when successful  ([#447](https://github.com/vmware/go-vcloud-director/pull/447), [#501](https://github.com/vmware/go-vcloud-director/pull/501))
* Added new methods `VCDClient.GetProviderVdcByHref`, `VCDClient.GetProviderVdcById`, `VCDClient.GetProviderVdcByName` and `ProviderVdc.Refresh` to retrieve Provider VDCs ([#502](https://github.com/vmware/go-vcloud-director/pull/502))
* Added new methods `VCDClient.GetProviderVdcExtendedByHref`, `VCDClient.GetProviderVdcExtendedById`, `VCDClient.GetProviderVdcExtendedByName` and `ProviderVdcExtended.Refresh` to retrieve the extended flavor of Provider VDCs ([#502](https://github.com/vmware/go-vcloud-director/pull/502))
* Added new methods `ProviderVdcExtended.ToProviderVdc`, to convert from an extended Provider VDC to a regular one ([#502](https://github.com/vmware/go-vcloud-director/pull/502))
* Added new methods `ProviderVdc.GetMetadata`, `ProviderVdc.AddMetadataEntry`, `ProviderVdc.AddMetadataEntryAsync`, `ProviderVdc.MergeMetadataAsync`, `ProviderVdc.MergeMetadata`, `ProviderVdc.DeleteMetadataEntry` and `ProviderVdc.DeleteMetadataEntryAsync` to manage Provider VDCs metadata ([#502](https://github.com/vmware/go-vcloud-director/pull/502))
* Added new methods `VCDClient.GetVmGroupById`, `VCDClient.GetVmGroupByNamedVmGroupIdAndProviderVdcUrn` and `VCDClient.GetVmGroupByNameAndProviderVdcUrn` to retrieve VM Groups. These are useful to create VM Placement Policies ([#504](https://github.com/vmware/go-vcloud-director/pull/504))
* Added new methods `VCDClient.GetLogicalVmGroupById`, `VCDClient.CreateLogicalVmGroup` and `LogicalVmGroup.Delete` to manage Logical VM Groups. These are useful to create VM Placement Policies ([#504](https://github.com/vmware/go-vcloud-director/pull/504))
* Added the function `VCDClient.GetMetadataByKeyAndHref` to get a specific metadata value using an entity reference ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added the functions `VCDClient.AddMetadataEntryWithVisibilityByHrefAsync`
  and `VCDClient.AddMetadataEntryWithVisibilityByHref` to add metadata with both visibility and domain
  to any entity by using its reference ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added the functions `VCDClient.MergeMetadataWithVisibilityByHrefAsync`
  and `VCDClient.MergeMetadataWithVisibilityByHref` to merge metadata data supporting also visibility and domain ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added the functions `VCDClient.DeleteMetadataEntryWithDomainByHrefAsync`
  and `VCDClient.DeleteMetadataEntryWithDomainByHref` to delete metadata from an entity using its reference ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added the function `GetMetadataByKey` to the following entities:
  `VM`, `Vdc`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `Catalog`, `AdminCatalog`,
  `Org`, `AdminOrg`, `Disk`, `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to get a specific metadata value ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added the functions `AddMetadataEntryWithVisibilityAsync` and `AddMetadataEntryWithVisibility` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to add metadata with both visibility and domain to them ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added the functions `MergeMetadataWithMetadataValuesAsync` and `MergeMetadataWithMetadataValues` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to merge metadata data supporting also visibility and domain ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added the functions `DeleteMetadataEntryWithDomainAsync` and `DeleteMetadataEntryWithDomain` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to delete metadata with the domain, that allows deleting metadata present in SYSTEM ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Added `AdminOrg` methods `CreateCatalogFromSubscriptionAsync` and `CreateCatalogFromSubscription` to create a
  subscribed catalog ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added method `AdminCatalog.FullSubscriptionUrl` to return the subscription URL of a published catalog ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added method `AdminCatalog.WaitForTasks` to wait for catalog tasks to complete ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added method `AdminCatalog.UpdateSubscriptionParams` to modify the terms of an existing subscription ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added methods `Catalog.QueryTaskList` and `AdminCatalog.QueryTaskList` to retrieve the tasks associated with a catalog ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added function `IsValidUrl` to determine if a URL is valid ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added `AdminCatalog` methods `Sync` and `LaunchSync` to synchronise a subscribed catalog ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added method `AdminCatalog.GetCatalogHref` to retrieve the HREF of a regular catalog ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added `AdminCatalog` methods `QueryCatalogItemList`, `QueryVappTemplateList`, and `QueryMediaList` to retrieve lists of
  dependent items ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added `AdminCatalog` methods `LaunchSynchronisationVappTemplates`, `LaunchSynchronisationAllVappTemplates`,
  `LaunchSynchronisationMediaItems`, and `LaunchSynchronisationAllMediaItems` to start synchronisation of dependent
  items ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added `AdminCatalog` methods `GetCatalogItemByHref` and `QueryCatalogItem` to retrieve a single Catalog Item ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added method `CatalogItem.LaunchSync` to start synchronisation of a catalog item ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added method `CatalogItem.Refresh` to get fresh contents for a catalog item ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added function `WaitResource` to wait for tasks associated to a gioven resource ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added function `MinimalShowTask` to display task progress with minimal info ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added functions `ResourceInProgress` and `ResourceComplete` to check on task activity for a given entity ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added functions `SkimTasksList`, `SkimTasksListMonitor`, `WaitTaskListCompletion`, `WaitTaskListCompletionMonitor` to
  process lists of tasks and lists of task IDs ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added `Client` methods `GetTaskByHREF` and `GetTaskById` to retrieve individual tasks ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Implemented `QueryItem` for `Task` and `AdminTask` (`GetHref`, `GetName`, `GetType`, `GetParentId`, `GetParentName`, `GetMetadataValue`, `GetDate`) ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Added `VCDClient.QueryMediaById` to query a media record using a media ID ([#520](https://github.com/vmware/go-vcloud-director/pull/520))
* Added `Vdc.QueryVappSynchronizedVmTemplate` to query a VM inside a vApp Template that must be synchronized in the catalog [[#520](https://github.com/vmware/go-vcloud-director/pull/520)] 
* Added `VCDClient.QueryVmInVAppTemplateByHref` and `VCDClient.QuerySynchronizedVmInVAppTemplateByHref` to query a VM
  inside a vApp Template by using the latter's hyper-reference ([#520](https://github.com/vmware/go-vcloud-director/pull/520))
* Added `VCDClient.QuerySynchronizedVAppTemplateById` to get a synchronized vApp Template query record from a vApp Template ID ([#520](https://github.com/vmware/go-vcloud-director/pull/520))

### IMPROVEMENTS
* Bumped Default API Version to V36.0 (VCD 10.3+) [#500](https://github.com/vmware/go-vcloud-director/pull/500)
* Added method `VM.Shutdown` to shut down guest OS ([#413](https://github.com/vmware/go-vcloud-director/pull/413), [#496](https://github.com/vmware/go-vcloud-director/pull/496))
* Added support for MoRef ID on VM record type. Using the MoRef ID, we can then correlate that back to vCenter Server and find the VM with matching MoRef ID  ([#491](https://github.com/vmware/go-vcloud-director/pull/491))
* Added support for querying VdcStorageProfile:  
  - functions `QueryAdminOrgVdcStorageProfileByID` and `QueryOrgVdcStorageProfileByID`  
  - query types `QtOrgVdcStorageProfile` and `QtAdminOrgVdcStorageProfile`  
  - data struct `QueryResultAdminOrgVdcStorageProfileRecordType` (non admin struct already was here)
  ([#499](https://github.com/vmware/go-vcloud-director/pull/499))
* Created new VDC Compute Policies CRUD methods using OpenAPI v2.0.0:
  `VCDClient.GetVdcComputePolicyV2ById`, `VCDClient.GetAllVdcComputePoliciesV2`, `VCDClient.CreateVdcComputePolicyV2`,
  `VdcComputePolicyV2.Update`, `VdcComputePolicyV2.Delete` and `AdminVdc.GetAllAssignedVdcComputePoliciesV2`.
  This version supports more filtering options like `isVgpuPolicy` ([#502](https://github.com/vmware/go-vcloud-director/pull/502), [#504](https://github.com/vmware/go-vcloud-director/pull/504), [#507](https://github.com/vmware/go-vcloud-director/pull/507))
* Simplified `Test_LDAP` by using a pre-configured LDAP server ([#505](https://github.com/vmware/go-vcloud-director/pull/505))
* Added VCDClient.GetAllNsxtEdgeClusters for lookup of NSX-T Edge Clusters in wider scopes -
  Provider VDC, VDC Group or VDC ([#512](https://github.com/vmware/go-vcloud-director/pull/512))
* Switched VDC.GetAllNsxtEdgeClusters to use 'orgVdcId' filter instead of '_context' (now deprecated)
  ([#512](https://github.com/vmware/go-vcloud-director/pull/512))
* Created `VM.UpdateComputePolicyV2` and `VM.UpdateComputePolicyV2Async` that uses v2.0.0 of VDC Compute Policy endpoint
  of OpenAPI and allows updating VM Sizing Policies and also VM Placement Policies for a given VM ([#513](https://github.com/vmware/go-vcloud-director/pull/513))
* Added `[]tenant` structure to simplify org user testing ([#515](https://github.com/vmware/go-vcloud-director/pull/515))
* Improved `Vdc.QueryVappVmTemplate` to avoid querying VMs in vApp Templates that are not synchronized in the catalog ([#520](https://github.com/vmware/go-vcloud-director/pull/520))

### BUG FIXES
* Changed `VdcComputePolicy.Description` to a non-omitempty pointer, to be able to send null values to VCD to set empty descriptions. ([#504](https://github.com/vmware/go-vcloud-director/pull/504))
* Fixed issue [#514](https://github.com/vmware/go-vcloud-director/issues/514) "ignoring pagination in network queries" ([#518](https://github.com/vmware/go-vcloud-director/pull/518))
* Fixed type `types.AdminVdc.ResourcePoolRefs` to make unmarshaling work (read-only) ([#494](https://github.com/vmware/go-vcloud-director/pull/494))
* Fixed Test_NsxtSecurityGroupGetAssociatedVms which had name clash ([#498](https://github.com/vmware/go-vcloud-director/pull/498))

### DEPRECATIONS
* Deprecated OpenAPI v1.0.0 VDC Compute Policies CRUD methods in favor of v2.0.0 ones:
  `Client.GetVdcComputePolicyById`, `Client.GetAllVdcComputePolicies`, `Client.CreateVdcComputePolicy`
  `VdcComputePolicy.Update`, `VdcComputePolicy.Delete` and `AdminVdc.GetAllAssignedVdcComputePolicies` ([#504](https://github.com/vmware/go-vcloud-director/pull/504))
* Deprecated the functions `VCDClient.AddMetadataEntryByHrefAsync`
  and `VCDClient.AddMetadataEntryByHref` in favor of `VCDClient.AddMetadataEntryWithVisibilityByHrefAsync`
  and `VCDClient.AddMetadataEntryWithVisibilityByHref` ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Deprecated the functions `VCDClient.MergeMetadataByHrefAsync`
  and `VCDClient.MergeMetadataByHref` in favor of `VCDClient.MergeMetadataWithVisibilityByHrefAsync`
  and `VCDClient.MergeMetadataWithVisibilityByHref` ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Deprecated the functions `AddMetadataEntryAsync` and `AddMetadataEntry` from the following entities:
  `VM`, `Vdc`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem` in favor of their `AddMetadataEntryWithVisibilityAsync` and `AddMetadataEntryWithVisibility`
  counterparts ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Deprecated the functions `MergeMetadataAsync` and `MergeMetadataAsync` from the following entities:
  `VM`, `Vdc`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem` in favor of their `MergeMetadataWithMetadataValuesAsync` and `MergeMetadataWithMetadataValues`
  counterparts ([#510](https://github.com/vmware/go-vcloud-director/pull/510))
* Deprecated the functions `DeleteMetadata` and `DeleteMetadataAsync` from the following entities:
  `VM`, `Vdc`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem` in favor of their `DeleteMetadataWithDomainAsync` and `DeleteMetadataWithDomain`
  counterparts ([#510](https://github.com/vmware/go-vcloud-director/pull/510))

### NOTES
* Switched `go.mod` to use Go 1.19 ([#511](https://github.com/vmware/go-vcloud-director/pull/511))
* Ran `make fmt` using Go 1.19 release (`fmt` automatically changes doc comment structure). This
  will prevent `make static` errors when running tests in pipeline using Go 1.19 ([#497](https://github.com/vmware/go-vcloud-director/pull/497))
* Updated branding `vCloud Director` -> `VMware Cloud Director` ([#497](https://github.com/vmware/go-vcloud-director/pull/497))
* package `io/ioutil` is deprecated as of Go 1.16. `staticcheck` started complaining about usage of
  deprecated packages. As a result packages `io` or `os` are used (still the same
  functions are used) ([#497](https://github.com/vmware/go-vcloud-director/pull/497))
* Adjusted `staticcheck` version naming to new format (from `2021.1.2` to `v0.3.3`) ([#497](https://github.com/vmware/go-vcloud-director/pull/497)]
* Added a new GitHub Action to run `gosec` on every push and pull request [[#516](https://github.com/vmware/go-vcloud-director/pull/516))
* Improved documentation for `types.OpenApiOrgVdcNetworkDhcp` ([#517](https://github.com/vmware/go-vcloud-director/pull/517))



## 2.16.0 (August 2, 2022)

### FEATURES
* Added support for `DnsServers` on `OpenApiOrgVdcNetworkDhcp` struct ([#465](https://github.com/vmware/go-vcloud-director/pull/465))
* Added new methods `Org.GetAllSecurityTaggedEntities`, `Org.GetAllSecurityTaggedEntitiesByName`,  `Org.GetAllSecurityTagValues`, `VM.GetVMSecurityTags`, `Org.UpdateSecurityTag` and `VM.UpdateVMSecurityTags` to deal with security tags ([#467](https://github.com/vmware/go-vcloud-director/pull/467))
* Added new structs `types.SecurityTag`, `types.SecurityTaggedEntity`, `types.SecurityTagValue` and `types.EntitySecurityTags` ([#467](https://github.com/vmware/go-vcloud-director/pull/467))
* Added `Vdc.GetControlAccess`, `Vdc.SetControlAccess` and `Vdc.DeleteControlAccess` to get, set and delete control access capabilities to VDCs ([#470](https://github.com/vmware/go-vcloud-director/pull/470))
* Added support to set, get and delete metadata to CatalogItem with the methods
  `CatalogItem.AddMetadataEntry`, `CatalogItem.AddMetadataEntryAsync`, `CatalogItem.GetMetadata`,
  `CatalogItem.DeleteMetadataEntry` and `CatalogItem.DeleteMetadataEntryAsync`. ([#471](https://github.com/vmware/go-vcloud-director/pull/471))
* Added `AdminCatalog.MergeMetadata`,`AdminCatalog.MergeMetadataAsync`, `AdminOrg.MergeMetadata`, `AdminOrg.MergeMetadataAsync`, 
`CatalogItem.MergeMetadata`, `CatalogItem.MergeMetadataAsync`, `Disk.MergeMetadata`, `Disk.MergeMetadataAsync`, `Media.MergeMetadata`, 
`Media.MergeMetadataAsync`, `MediaRecord.MergeMetadata`, `MediaRecord.MergeMetadataAsync`, `OpenAPIOrgVdcNetwork.MergeMetadata`, 
`OpenAPIOrgVdcNetwork.MergeMetadataAsync`, `OrgVDCNetwork.MergeMetadata`, `OrgVDCNetwork.MergeMetadataAsync`, 
`VApp.MergeMetadata`, `VApp.MergeMetadataAsync`, `VAppTemplate.MergeMetadata`, `VAppTemplate.MergeMetadataAsync`, 
`VM.MergeMetadata`, `VM.MergeMetadataAsync`, `Vdc.MergeMetadata`, `Vdc.MergeMetadataAsync` to merge metadata, 
which both updates existing metadata with same key and adds new entries for the non-existent ones ([#473](https://github.com/vmware/go-vcloud-director/pull/473))
* Added NSX-T Edge Gateway methods `NsxtEdgeGateway.GetNsxtRouteAdvertisement`,
  `NsxtEdgeGateway.GetNsxtRouteAdvertisementWithContext`,
  `NsxtEdgeGateway.UpdateNsxtRouteAdvertisement`,
  `NsxtEdgeGateway.UpdateNsxtRouteAdvertisementWithContext`,
  `NsxtEdgeGateway.DeleteNsxtRouteAdvertisement` and
  `NsxtEdgeGateway.DeleteNsxtRouteAdvertisementWithContext` that allow to manage NSX-T Route
  Advertisement ([#478](https://github.com/vmware/go-vcloud-director/pull/478), [#480](https://github.com/vmware/go-vcloud-director/pull/480))
* Added new methods `NsxtEdgeGateway.GetBgpConfiguration`, `NsxtEdgeGateway.UpdateBgpConfiguration`,
  `NsxtEdgeGateway.DisableBgpConfiguration` for BGP Configuration management on NSX-T Edge Gateway
  ([#480](https://github.com/vmware/go-vcloud-director/pull/480))
* Added new structs `types.EdgeBgpConfig`, `types.EdgeBgpGracefulRestartConfig`,
  `types.EdgeBgpConfigVersion` for BGP Configuration management on NSX-T Edge Gateway ([#480](https://github.com/vmware/go-vcloud-director/pull/480))
* Added support for Dynamic Security Groups in VCD 10.3 by expanding `types.NsxtFirewallGroup` to
  accommodate fields required for Dynamic Security Groups, implemented automatic API elevation to
  v36.0. Added New functions `VdcGroup.CreateNsxtFirewallGroup`,
  `NsxtFirewallGroup.IsDynamicSecurityGroup` ([#487](https://github.com/vmware/go-vcloud-director/pull/487))
* Added support for managing NSX-T Edge Gateway BGP IP Prefix Lists. It is done by adding types `EdgeBgpIpPrefixList` and
`types.EdgeBgpIpPrefixList` with functions `CreateBgpIpPrefixList`, `GetAllBgpIpPrefixLists`,
`GetBgpIpPrefixListByName`, `GetBgpIpPrefixListById`, `Update` and `Delete`  ([#488](https://github.com/vmware/go-vcloud-director/pull/488))
* Added support for managing NSX-T Edge Gateway BGP Neighbor. It is done by adding types `EdgeBgpNeighbor` and
  `types.EdgeBgpNeighbor` with functions `CreateBgpNeighbor`, `GetAllBgpNeighbors`,
  `GetBgpNeighborByIp`, `GetBgpNeighborById`, `Update` and `Delete`  ([#489](https://github.com/vmware/go-vcloud-director/pull/489))

### IMPROVEMENTS
* Added methods `client.CreateVdcComputePolicy`, `client.GetVdcComputePolicyById`, `client.GetAllVdcComputePolicies` ([#468](https://github.com/vmware/go-vcloud-director/pull/468))
* Added additional methods for convenience of NSX-T Org Network DHCP handling
  `OpenApiOrgVdcNetwork.GetOpenApiOrgVdcNetworkDhcp`, `OpenApiOrgVdcNetwork.DeletNetworkDhcp`
  `OpenApiOrgVdcNetwork.UpdateDhcp` ([#469](https://github.com/vmware/go-vcloud-director/pull/469))
* Added additional support for UDF type ISO files in `catalog.UploadMediaImage` ([#479](https://github.com/vmware/go-vcloud-director/pull/479))
* Added `SupportedFeatureSet` attribute to `NsxtAlbServiceEngineGroup` and `NsxtAlbConfig` to
support v37.0 license management for AVI Load Balancer and replace `LicenseType` from
`NsxtAlbController` ([#485](https://github.com/vmware/go-vcloud-director/pull/485))

### BUG FIXES
* Fixed method `adminOrg.FindCatalogRecords` to escape name in query URL ([#466](https://github.com/vmware/go-vcloud-director/pull/466))
* Fixed method `vm.WaitForDhcpIpByNicIndexes` to ignore not found Edge Gateway ([#481](https://github.com/vmware/go-vcloud-director/pull/481))

### DEPRECATIONS
* Deprecated `org.GetVdcComputePolicyById`, `adminOrg.GetVdcComputePolicyById` ([#468](https://github.com/vmware/go-vcloud-director/pull/468))
* Deprecated `org.GetAllVdcComputePolicies`, `adminOrg.GetAllVdcComputePolicies`, `org.CreateVdcComputePolicy` ([#468](https://github.com/vmware/go-vcloud-director/pull/468))


## 2.15.0 (April 14, 2022)

### FEATURES
* Added support for Shareable disks, i.e., independent disks that can be attached to multiple VMs which is available from
  API v35.0 onwards. Also added UUID to the Disk structure which is a new member that is returned from v36.0 onwards. This
  member holds a UUID that can be used to correlate the disk that is attached to a particular VM from the VCD side and the
  VM host side. ([#383](https://github.com/vmware/go-vcloud-director/pull/383))
* Added support for uploading OVF using URL `catalog.UploadOvfByLink` ([#422](https://github.com/vmware/go-vcloud-director/pull/422), [#426](https://github.com/vmware/go-vcloud-director/pull/426))
* Added support for updating vApp template `vAppTemplate.UpdateAsync` and `vAppTemplate.Update` ([#422](https://github.com/vmware/go-vcloud-director/pull/422))
* Added methods `catalog.PublishToExternalOrganizations` and `adminCatalog.PublishToExternalOrganizations` ([#424](https://github.com/vmware/go-vcloud-director/pull/424))
* Added types `types.MetadataStringValue`, `types.MetadataNumberValue`, `types.MetadataDateTimeValue` and `types.MetadataBooleanValue` 
  for adding different kind of metadata to entities ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Added support to set, get and delete metadata to AdminCatalog with the methods 
  `AdminCatalog.AddMetadataEntry`, `AdminCatalog.AddMetadataEntryAsync`, `AdminCatalog.GetMetadata`, 
  `AdminCatalog.DeleteMetadataEntry` and `AdminCatalog.DeleteMetadataEntryAsync`. ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Added support to get metadata from Catalog with method `Catalog.GetMetadata` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Added to `VM` and `VApp` the methods `DeleteMetadataEntry`, `DeleteMetadataEntryAsync`, `AddMetadataEntry` and `AddMetadataEntryAsync`
  so it follows the same convention as the rest of entities that uses metadata. ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Added methods `vm.ChangeCPU` and `vm.ChangeMemory` which uses the latest API structure instead of deprecated ones ([#432](https://github.com/vmware/go-vcloud-director/pull/432))
* Added environment variable `GOVCD_API_VERSION` so API version can be set manually ([#434](https://github.com/vmware/go-vcloud-director/pull/434))
* Added support to set, get and delete metadata to AdminOrg with the methods
  `AdminOrg.AddMetadataEntry`, `AdminOrg.AddMetadataEntryAsync`, `AdminOrg.GetMetadata`,
  `AdminOrg.DeleteMetadataEntry` and `AdminOrg.DeleteMetadataEntryAsync`. ([#438](https://github.com/vmware/go-vcloud-director/pull/438))
* Added support to get metadata to Org with the method
  `Org.GetMetadata`. ([#438](https://github.com/vmware/go-vcloud-director/pull/438))
* Added support to set, get and delete metadata to Disk with the methods
  `Disk.AddMetadataEntry`, `Disk.AddMetadataEntryAsync`, `Disk.GetMetadata`,
  `Disk.DeleteMetadataEntry` and `Disk.DeleteMetadataEntryAsync`. ([#438](https://github.com/vmware/go-vcloud-director/pull/438))
* Added new structure `AnyTypeEdgeGateway` which supports retreving both types of Edge Gateways
  (NSX-V and NSX-T) with methods `AdminOrg.GetAnyTypeEdgeGatewayById`,
  `Org.GetAnyTypeEdgeGatewayById`, `AnyTypeEdgeGateway.IsNsxt`, `AnyTypeEdgeGateway.IsNsxv`,
  `AnyTypeEdgeGateway.GetNsxtEdgeGateway` ([#443](https://github.com/vmware/go-vcloud-director/pull/443))
* Added functions `VdcGroup.GetCapabilities`, `VdcGroup.IsNsxt`,
  `VdcGroup.GetOpenApiOrgVdcNetworkByName`, `VdcGroup.GetAllOpenApiOrgVdcNetworks`,
  `Org.GetOpenApiOrgVdcNetworkByNameAndOwnerId` ([#443](https://github.com/vmware/go-vcloud-director/pull/443))
* Added method `AdminOrg.FindCatalogRecords` that allows to query `types.CatalogRecord` by their catalog name. ([#450](https://github.com/vmware/go-vcloud-director/pull/450))
* Added methods `Client.QueryWithNotEncodedParamsWithHeaders` and `Client.QueryWithNotEncodedParamsWithApiVersionWithHeaders` so HTTP headers can be added now when doing API queries. ([#450](https://github.com/vmware/go-vcloud-director/pull/450))
* Added functions `VdcGroup.GetNsxtFirewallGroupByName` and `VdcGroup.GetNsxtFirewallGroupById` ([#451](https://github.com/vmware/go-vcloud-director/pull/451))
* Added support for for Network Context Profile lookup using `GetAllNetworkContextProfiles` and
  `GetNetworkContextProfilesByNameScopeAndContext` functions ([#452](https://github.com/vmware/go-vcloud-director/pull/452))
* Added support for NSX-T Distributed Firewall rule management using type `DistributedFirewall` and
`VdcGroup.GetDistributedFirewall`, `VdcGroup.UpdateDistributedFirewall`,
`VdcGroup.DeleteAllDistributedFirewallRules`, `DistributedFirewall.DeleteAllRules` ([#452](https://github.com/vmware/go-vcloud-director/pull/452))
* Added support to set, get and delete metadata to the following resources via its HREF:
  `catalog`, `catalog item`, `edge gateway`, `independent disk`, `media`, `network`, `org`, `PVDC`, `PVDC storage profile`, `vApp`, `vApp template`,`VDC` and `VDC storage profile`;
  with the methods
  `VCDClient.GetMetadataByHref`, `VCDClient.AddMetadataEntryByHref`, `VCDClient.AddMetadataEntryByHrefAsync`,
  `VCDClient.DeleteMetadataEntryByHref` and `VCDClient.DeleteMetadataEntryByHrefAsync` ([#454](https://github.com/vmware/go-vcloud-director/pull/454))
* Added functions `VdcGroup.GetOpenApiOrgVdcNetworkById` and `VdcGroup.CreateOpenApiOrgVdcNetwork` ([#456](https://github.com/vmware/go-vcloud-director/pull/456))
* New method added `Disk.GetAttachedVmsHrefs` ([#436](https://github.com/vmware/go-vcloud-director/pull/436))

### IMPROVEMENTS
* Bumped Default API Version to V35.0 ([#434](https://github.com/vmware/go-vcloud-director/pull/434))
* Disk methods have now the ability to access new properties from API version 36.0. They are: `DiskRecordType.SharingType`, `DiskRecordType.UUID`, `DiskRecordType.Encrypted`, `Disk.SharingType`, `Disk.UUID` and `Disk.Encrypted` ([#436](https://github.com/vmware/go-vcloud-director/pull/436))
* Added support for `User` entities imported from LDAP, with `IsExternal` property ([#439](https://github.com/vmware/go-vcloud-director/pull/439))
* Added support for users list attribute for `Group` ([#439](https://github.com/vmware/go-vcloud-director/pull/439))
* Improved `group.Update()` to avoid sending the users list to VCD to avoid unwanted errors ([#449](https://github.com/vmware/go-vcloud-director/pull/449))
* NSX-T Edge Gateway now supports VDC Groups by switching from `OrgVdc` to `OwnerRef` field.
  Additional methods `NsxtEdgeGateway.MoveToVdcOrVdcGroup()`,
  `Org.GetNsxtEdgeGatewayByNameAndOwnerId()`, `VdcGroup.GetNsxtEdgeGatewayByName()`,
  `VdcGroup.GetAllNsxtEdgeGateways()`, `org.GetVdcGroupById` ([#440](https://github.com/vmware/go-vcloud-director/pull/440))
* Added additional helper functions `OwnerIsVdcGroup()`, `OwnerIsVdc()`, `VdcGroup.GetCapabilities()`,
  `VdcGroup.IsNsxt()` ([#440](https://github.com/vmware/go-vcloud-director/pull/440))
* Added support to set, get and delete metadata to VDC Networks with the methods
  `OrgVDCNetwork.AddMetadataEntry`, `OrgVDCNetwork.AddMetadataEntryAsync`, `OrgVDCNetwork.GetMetadata`,
  `OrgVDCNetwork.DeleteMetadataEntry` and `OrgVDCNetwork.DeleteMetadataEntryAsync` ([#442](https://github.com/vmware/go-vcloud-director/pull/442))
* Added `CanPublishExternally` and `CanSubscribe` attributes to `OrgGeneralSettings` struct. ([#444](https://github.com/vmware/go-vcloud-director/pull/444))
* Added workaround to tests for Org Catalog publishing bug when dealing with LDAP ([#458](https://github.com/vmware/go-vcloud-director/pull/458))
* Added clean-up actions to some tests that were uploading vAppTemplates/medias to catalogs ([#458](https://github.com/vmware/go-vcloud-director/pull/458))
* Added support to set, get and delete metadata to OpenAPI VDC Networks through XML with the methods
  `OpenApiOrgVdcNetwork.AddMetadataEntry`, `OpenApiOrgVdcNetwork.GetMetadata`,
  `OpenApiOrgVdcNetwork.DeleteMetadataEntry` ([#459](https://github.com/vmware/go-vcloud-director/pull/459))
* Added `Vdc.GetNsxtAppPortProfileByName` and `VdcGroup.GetNsxtAppPortProfileByName` for NSX-T
  Application Port Profile lookup ([#460](https://github.com/vmware/go-vcloud-director/pull/460))

### BUG FIXES
* Fixed Issue #431 "Wrong order in Task structure" ([#433](https://github.com/vmware/go-vcloud-director/pull/433))
* Fixed Issue where VDC creation with storage profile `enabled=false` wasn't working. `VdcStorageProfile.enabled` and `VdcStorageProfileConfiguration.enabled` changed to pointers ([#433](https://github.com/vmware/go-vcloud-director/pull/433))
* Fixed method `client.GetStorageProfileByHref` to return IOPS `IopsSettings` ([#435](https://github.com/vmware/go-vcloud-director/pull/435))
* `Vms.VmReference` changed to array to fix incorrect deserialization ([#436](https://github.com/vmware/go-vcloud-director/pull/436))
* `Catalog.QueryMediaList` method was not working because `fmt.Sprintf` was being misused ([#441](https://github.com/vmware/go-vcloud-director/pull/441))

### DEPRECATIONS
* Deprecated `vm.DeleteMetadata` in favor of `vm.DeleteMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vm.AddMetadata` in favor of `vm.AddMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vdc.DeleteMetadata` in favor of `vdc.DeleteMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vdc.AddMetadata` in favor of `vdc.AddMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vdc.AddMetadataAsync` in favor of `vdc.AddMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vdc.DeleteMetadataAsync` in favor of `vdc.DeleteMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vApp.DeleteMetadata` in favor of `vApp.DeleteMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vApp.AddMetadata` in favor of `vApp.AddMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vAppTemplate.AddMetadata` in favor of `vAppTemplate.AddMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vAppTemplate.AddMetadataAsync` in favor of `vAppTemplate.AddMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vAppTemplate.DeleteMetadata` in favor of `vAppTemplate.DeleteMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vAppTemplate.DeleteMetadataAsync` in favor of `vAppTemplate.DeleteMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `mediaRecord.AddMetadata` in favor of `mediaRecord.AddMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `mediaRecord.AddMetadataAsync` in favor of `mediaRecord.AddMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `mediaRecord.DeleteMetadata` in favor of `mediaRecord.DeleteMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `mediaRecord.DeleteMetadataAsync` in favor of `mediaRecord.DeleteMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `media.AddMetadata` in favor of `media.AddMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `media.AddMetadataAsync` in favor of `media.AddMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `media.DeleteMetadata` in favor of `media.DeleteMetadataEntry` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `media.DeleteMetadataAsync` in favor of `media.DeleteMetadataEntryAsync` ([#430](https://github.com/vmware/go-vcloud-director/pull/430))
* Deprecated `vm.ChangeMemorySize` in favor of `vm.ChangeMemory` ([#432](https://github.com/vmware/go-vcloud-director/pull/432))
* Deprecated `vm.ChangeCPUCount` and `vm.ChangeCPUCountWithCore` in favor of `vm.ChangeCPU` ([#432](https://github.com/vmware/go-vcloud-director/pull/432))

### NOTES
* Bumped `staticcheck` version to 2022.1 with Go 1.18 support ([#457](https://github.com/vmware/go-vcloud-director/pull/457))

## 2.14.0 (January 7, 2022)

### FEATURES
* Added type `NsxtAlbConfig` and functions `NsxtEdgeGateway.UpdateAlbSettings`, `NsxtEdgeGateway.GetAlbSettings`,
  `NsxtEdgeGateway.DisableAlb` ([#403](https://github.com/vmware/go-vcloud-director/pull/403))
* Added types `Certificate` and `types.CertificateLibraryItem` for handling Certificates in Certificate Library with corresponding
  methods `client.GetCertificateFromLibraryById`, `client.AddCertificateToLibrary`, `client.GetAllCertificatesFromLibrary`, `client.GetCertificateFromLibraryByName`, `adminOrg.GetCertificateFromLibraryById`, `adminOrg.AddCertificateToLibrary`, `adminOrg.GetAllCertificatesFromLibrary`, `adminOrg.GetCertificateFromLibraryByName`,
  `certificate.Update`, `certificate.Delete` ([#404](https://github.com/vmware/go-vcloud-director/pull/404))
* Added support for ALB Service Engine Group Assignment to NSX-T Edge Gateway via type
  `NsxtAlbServiceEngineGroupAssignment` and functions `GetAllAlbServiceEngineGroupAssignments`,
  `GetAlbServiceEngineGroupAssignmentById`, `GetAlbServiceEngineGroupAssignmentByName`,
  `CreateAlbServiceEngineGroupAssignment`, `Update`, `Delete`  ([#405](https://github.com/vmware/go-vcloud-director/pull/405))
* Added type `types.ApiTokenRefresh` to contain data from API token refresh ([#406](https://github.com/vmware/go-vcloud-director/pull/406))
* Added method `VCDClient.GetBearerTokenFromApiToken` to get a bearer token from an API token ([#406](https://github.com/vmware/go-vcloud-director/pull/406))
* Added method `VCDClient.SetApiToken` to set a token and get a bearer token using and API token and get token details in return ([#406](https://github.com/vmware/go-vcloud-director/pull/406))
* Added types `VdcGroup`, `types.VdcGroup`, `types.ParticipatingOrgVdcs`, `types.CandidateVdc`, `types.DfwPolicies` and `types.DefaultPolicy` for handling VDC groups with corresponding
  methods `adminOrg.CreateNsxtVdcGroup`, `adminOrg.CreateVdcGroup`, `adminOrg.GetAllNsxtVdcGroupCandidates`, `adminOrg.GetAllVdcGroupCandidates`, `adminOrg.GetAllVdcGroups`, `adminOrg.GetVdcGroupByName`, `adminOrg.GetVdcGroupById`, `vdcGroup.Update`, `vdcGroup.GenericUpdate`, `vdcGroup.Delete`, `vdcGroup.DisableDefaultPolicy`, `vdcGroup.EnableDefaultPolicy`, `vdcGroup.GetDfwPolicies`, `vdcGroup.DeActivateDfw`, `vdcGroup.ActivateDfw`, `vdcGroup.UpdateDefaultDfwPolicies`, `vdcGroup.UpdateDfwPolicies`  ([#410](https://github.com/vmware/go-vcloud-director/pull/410))
* Added support for ALB Pool to NSX-T Edge Gateway via type `NsxtAlbPool` and functions `GetAllAlbPools`,
  `GetAllAlbPoolSummaries`, `GetAlbPoolByName`, `GetAlbPoolById`, `CreateNsxtAlbPool`, `nsxtAlbPool.Update`,
  `nsxtAlbPool.Delete` ([#414](https://github.com/vmware/go-vcloud-director/pull/414))
* Added support for ALB Virtual Services to NSX-T Edge Gateway via type `NsxtAlbVirtualService` and functions `GetAllAlbVirtualServices`,
  `GetAllAlbGetAllAlbVirtualServiceSummaries`, `GetAlbVirtualServiceByName`, `GetAlbVirtualServiceById`,
  `CreateNsxtAlbVirtualService`, `NsxtAlbVirtualService.Update`, `NsxtAlbVirtualService.Delete` ([#417](https://github.com/vmware/go-vcloud-director/pull/417))

### IMPROVEMENTS
* `VCDClient.SetToken` has now the ability of transparently setting a bearer token when receiving an API token ([#406](https://github.com/vmware/go-vcloud-director/pull/406))
* Removed Coverity warnings from code ([#408](https://github.com/vmware/go-vcloud-director/pull/408), [#412](https://github.com/vmware/go-vcloud-director/pull/412))
* Added session info to go-vcloud-director logs ([#409](https://github.com/vmware/go-vcloud-director/pull/409))
* Added type `types.UpdateLeaseSettingsSection` to handle vApp lease settings. ([#420](https://github.com/vmware/go-vcloud-director/pull/420))
* Added methods `vApp.GetLease` and `vApp.RenewLease`, to query the state of the vApp lease and eventually modify it. ([#420](https://github.com/vmware/go-vcloud-director/pull/420))
* Added `LeaseSettingsSection` to `types.VApp` structure. ([#420](https://github.com/vmware/go-vcloud-director/pull/420))

### BUG FIXES
* Fixed Issue #728: `vm.UpdateInternalDisksAsync()` didn't send VM description and as a result would delete VM description ([#418](https://github.com/vmware/go-vcloud-director/pull/418))
* Removed hardcoded 0 value for Weight field in `ChangeCPUCountWithCore` function to avoid overriding shares ([#419](https://github.com/vmware/go-vcloud-director/pull/419))
* Fixed issue #421 "Wrong xml name in SourcedVmTemplateParams" ([#420](https://github.com/vmware/go-vcloud-director/pull/420))


## 2.13.0 (September 30, 2021)

### FEATURES
* Added method `AdminVdc.AddStorageProfile` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `AdminVdc.AddStorageProfileWait` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `AdminVdc.RemoveStorageProfile` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `AdminVdc.RemoveStorageProfileWait` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `AdminVdc.SetDefaultStorageProfile` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `AdminVdc.GetDefaultStorageProfileReference` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `VCDClient.GetStorageProfileByHref` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `Client.GetStorageProfileByHref` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `VCDClient.QueryProviderVdcStorageProfileByName` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `Client.QueryAllProviderVdcStorageProfiles` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added method `Client.QueryProviderVdcStorageProfiles` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Added types `NsxtAlbController` and `types.NsxtAlbController` for handling NSX-T ALB Controllers with corresponding
  functions `GetAllAlbControllers`, `GetAlbControllerByName`, `GetAlbControllerById`, `GetAlbControllerByUrl`,
  `CreateNsxtAlbController`, `Update`, `Delete` ([#398](https://github.com/vmware/go-vcloud-director/pull/398))
* Added types `NsxtAlbCloud` and `types.NsxtAlbCloud` for handling NSX-T ALB Clouds with corresponding functions
  `GetAllAlbClouds`, `GetAlbCloudByName`, `GetAlbCloudById`, `CreateAlbCloud`, `Delete` ([#398](https://github.com/vmware/go-vcloud-director/pull/398))
* Added type `NsxtAlbImportableCloud` and `types.NsxtAlbImportableCloud` for listing NSX-T ALB Importable Clouds with
  corresponding functions `GetAllAlbImportableClouds`, `GetAlbImportableCloudByName`, `GetAlbImportableCloudById`
  ([#398](https://github.com/vmware/go-vcloud-director/pull/398))
* Added types `NsxtAlbServiceEngineGroup` and `types.NsxtAlbServiceEngineGroup` for handling NSX-T ALB Service Engine
  Groups with corresponding functions `GetAllNsxtAlbServiceEngineGroups`, `GetAlbServiceEngineGroupByName`,
  `GetAlbServiceEngineGroupById`, `CreateNsxtAlbServiceEngineGroup`, `Update`, `Delete`, `Sync` ([#398](https://github.com/vmware/go-vcloud-director/pull/398))
* Added types `NsxtAlbImportableServiceEngineGroups` and `types.NsxtAlbImportableServiceEngineGroups` for listing NSX-T
  ALB Importable Service Engine Groups with corresponding functions `GetAllAlbImportableServiceEngineGroups`,
  `GetAlbImportableServiceEngineGroupByName`, `GetAlbImportableServiceEngineGroupById` ([#398](https://github.com/vmware/go-vcloud-director/pull/398))

### IMPROVEMENTS
* External network type ExternalNetworkV2 automatically elevates API version to maximum available out of 33.0, 35.0 and
  36.0, so that new functionality can be consumed. It uses a controlled version elevation mechanism to consume the newer
  features, but at the same time remain tested by not choosing the latest untested version blindly (more information in
  openapi_endpoints.go) ([#399](https://github.com/vmware/go-vcloud-director/pull/399))
* Added new field BackingTypeValue in favor of deprecated BackingType to types.ExternalNetworkV2Backing ([#399](https://github.com/vmware/go-vcloud-director/pull/399))
* Added new function `GetFilteredNsxtImportableSwitches` to query NSX-T Importable Switches (Segments) ([#399](https://github.com/vmware/go-vcloud-director/pull/399))
* Added `.changes` directory for changelog items ([#391](https://github.com/vmware/go-vcloud-director/pull/391))

* Aligned build tags to match go fmt with Go 1.17 ([#396](https://github.com/vmware/go-vcloud-director/pull/396))
* Improved `test-tags.sh` script to handle new build tag format ([#396](https://github.com/vmware/go-vcloud-director/pull/396))

### BUG FIXES
* Fixed handling of `staticcheck` in GitGub Actions ([#391](https://github.com/vmware/go-vcloud-director/pull/391))

* Fixed Issue #390: `catalog.Delete()` ignores returned task and responds immediately which could have caused failures ([#392](https://github.com/vmware/go-vcloud-director/pull/392))

* Fixed Issue #395 "BUG: can't update EGW - there is no ownerRef field" ([#397](https://github.com/vmware/go-vcloud-director/pull/397))

### DEPRECATIONS
* Deprecated `GetStorageProfileByHref`  in favor of either `client.GetStorageProfileByHref` or `vcdClient.GetStorageProfileByHref` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Deprecated `QueryProviderVdcStorageProfileByName` in favor of `VCDClient.QueryProviderVdcStorageProfileByName` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Deprecated `VCDClient.QueryProviderVdcStorageProfiles` in favor of either `client.QueryProviderVdcStorageProfiles` or `client.QueryAllProviderVdcStorageProfiles` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))
* Deprecated `Vdc.GetDefaultStorageProfileReference` in favor of `adminVdc.GetDefaultStorageProfileReference` ([#393](https://github.com/vmware/go-vcloud-director/pull/393))

## 2.12.1 (July 5, 2021)

BUGS FIXED:
* org.GetCatalogByName and org.GetCatalogById could not retrieve shared catalogs from different Orgs 
  [#389](https://github.com/vmware/go-vcloud-director/pull/389)


## 2.12.0 (June 30, 2021)

* Improved error handling and function receiver name in client
  [#379](https://github.com/vmware/go-vcloud-director/pull/379)
* Added method `vdc.QueryEdgeGateway` [#364](https://github.com/vmware/go-vcloud-director/pull/364)
* Deprecated `vdc.GetEdgeGatewayRecordsType` [#364](https://github.com/vmware/go-vcloud-director/pull/364)
* Dropped support for VCD 9.7 which is EOL now [#371](https://github.com/vmware/go-vcloud-director/pull/371)
* Bumped Default API Version to V33.0  [#371](https://github.com/vmware/go-vcloud-director/pull/371)
* Methods `GetVDCById` and `GetVDCByName` for `Org` now use queries behind the scenes because Org 
  structure does not list child VDCs anymore  [#371](https://github.com/vmware/go-vcloud-director/pull/371), 
  [#376](https://github.com/vmware/go-vcloud-director/pull/376), [#382](https://github.com/vmware/go-vcloud-director/pull/382)
* Methods `GetCatalogById` and `GetCatalogByName` for `Org`  now use queries behind the scenes because Org
  structure does not list child Catalogs anymore  [#371](https://github.com/vmware/go-vcloud-director/pull/371), 
  [#376](https://github.com/vmware/go-vcloud-director/pull/376)
* Drop legacy authentication mechanism (vcdAuthorize) and use only new Cloud API provided (vcdCloudApiAuthorize) as
  API V33.0 is sufficient for it [#371](https://github.com/vmware/go-vcloud-director/pull/371)
* Added NSX-T Firewall Group type (which represents a Security Group or an IP Set) support by using
  structures `NsxtFirewallGroup` and `NsxtFirewallGroupMemberVms`. The following methods are
  introduced for managing Security Groups and Ip Sets: `Vdc.CreateNsxtFirewallGroup`,
  `NsxtEdgeGateway.CreateNsxtFirewallGroup`, `Org.GetAllNsxtFirewallGroups`,
  `Vdc.GetAllNsxtFirewallGroups`, `Org.GetNsxtFirewallGroupByName`,
  `Vdc.GetNsxtFirewallGroupByName`, `NsxtEdgeGateway.GetNsxtFirewallGroupByName`,
  `Org.GetNsxtFirewallGroupById`, `Vdc.GetNsxtFirewallGroupById`,
  `NsxtEdgeGateway.GetNsxtFirewallGroupById`, `NsxtFirewallGroup.Update`,
  `NsxtFirewallGroup.Delete`, `NsxtFirewallGroup.GetAssociatedVms`,
  `NsxtFirewallGroup.IsSecurityGroup`, `NsxtFirewallGroup.IsIpSet`
  [#368](https://github.com/vmware/go-vcloud-director/pull/368)
* Added methods Org.QueryVmList and Org.QueryVmById to find VM by ID in an Org
  [#368](https://github.com/vmware/go-vcloud-director/pull/368)
* Added `NsxtAppPortProfile` and `types.NsxtAppPortProfile` for NSX-T Application Port Profile management 
  [#378](https://github.com/vmware/go-vcloud-director/pull/378)  
* Deprecated methods `vdc.ComposeRawVApp` and `vdc.ComposeVApp` [#387](https://github.com/vmware/go-vcloud-director/pull/387)
* Added method `vdc.CreateRawVApp` [#387](https://github.com/vmware/go-vcloud-director/pull/387)
* Removed deprecated method `adminOrg.GetRole`  
* Added Tenant Context management functions `Client.RemoveCustomHeader`, `Client.SetCustomHeader`, `WithHttpHeader`, 
  and many private methods to retrieve tenant context down the hierarchy. More details in `CODING_GUIDELINES.md`
  [#380](https://github.com/vmware/go-vcloud-director/pull/380)
* Added Rights management methods `AdminOrg.GetAllRights`, `AdminOrg.GetAllRightsCategories`, `AdminOrg.GetRightById`, 
  `AdminOrg.GetRightByName`, `Client.GetAllRights`, `Client.GetAllRightsCategories`, `Client.GetRightById`, 
  `Client.GetRightByName`, `client.GetRightsCategoryById`, `AdminOrg.GetRightsCategoryById` [#380](https://github.com/vmware/go-vcloud-director/pull/380)
* Added Global Role management methods `Client.GetAllGlobalRoles`, `Client.CreateGlobalRole`, `Client.GetGlobalRoleById`,
  `Client.GetGlobalRoleByName`, `GlobalRole.AddRights`, `GlobalRole.Delete`, `GlobalRole.GetRights`,
  `GlobalRole.GetTenants`, `GlobalRole.PublishAllTenants`, `GlobalRole.PublishTenants`, `GlobalRole.RemoveAllRights`,
  `GlobalRole.RemoveRights`, `GlobalRole.ReplacePublishedTenants`, `GlobalRole.UnpublishAllTenants`, 
  `GlobalRole.UnpublishTenants`, `GlobalRole.Update`, `GlobalRole.UpdateRights` [#380](https://github.com/vmware/go-vcloud-director/pull/380)
* Added Rights Bundle management methods `Client.CreateRightsBundle`, `Client.GetAllRightsBundles`, 
  `Client.GetRightsBundleById`, `Client.GetRightsBundleByName`, `RightsBundle.AddRights`, `RightsBundle.Delete`, 
  `RightsBundle.GetRights`, `RightsBundle.GetTenants`, `RightsBundle.PublishAllTenants`, `RightsBundle.PublishTenants`, 
  `RightsBundle.RemoveAllRights`, `RightsBundle.RemoveRights`, `RightsBundle.ReplacePublishedTenants`, 
  `RightsBundle.UnpublishAllTenants`, `RightsBundle.UnpublishTenants`, `RightsBundle.Update`, `RightsBundle.UpdateRights`
  [#380](https://github.com/vmware/go-vcloud-director/pull/380)
* Added Role managemnt methods `AdminOrg.GetAllRoles`, `AdminOrg.GetRoleById`, `AdminOrg.GetRoleByName`, 
  `Client.GetAllRoles`, `Role.AddRights`, `Role.GetRights`, `Role.RemoveAllRights`, `Role.RemoveRights`, `Role.UpdateRights`
  [#380](https://github.com/vmware/go-vcloud-director/pull/380)
* Added convenience function `FindMissingImpliedRights` [#380](https://github.com/vmware/go-vcloud-director/pull/380)
* Added methods `NsxtEdgeGateway.UpdateNsxtFirewall()`, `NsxtEdgeGateway.GetNsxtFirewall()`, `nsxtFirewall.DeleteAllRules()`,
  `nsxtFirewall.DeleteRuleById` [#381](https://github.com/vmware/go-vcloud-director/pull/381)
* Added NSX-T NAT support with types `NsxtNatRule` and `types.NsxtNatRule` as well as methods `edge.GetAllNsxtNatRules`,
  `edge.GetNsxtNatRuleByName`, `edge.GetNsxtNatRuleById`, `edge.CreateNatRule`, `nsxtNatRule.Update`, `nsxtNatRule.Delete`,
  `nsxtNatRule.IsEqualTo` [#382](https://github.com/vmware/go-vcloud-director/pull/382)
* Added `NsxtIpSecVpnTunnel` and `types.NsxtIpSecVpnTunnel` for NSX-T IPsec VPN Tunnel configuration
  [#385](https://github.com/vmware/go-vcloud-director/pull/385)

BREAKING CHANGES:
* Added parameter `description` to method `vdc.ComposeRawVapp` [#372](https://github.com/vmware/go-vcloud-director/pull/372)
* Added methods `vapp.Rename`, `vapp.UpdateDescription`, `vapp.UpdateNameDescription` [#372](https://github.com/vmware/go-vcloud-director/pull/372)
* Field `types.Disk.Size` is replaced with `types.Disk.SizeMb` as size in Kilobytes is not supported in V33.0
  [#371](https://github.com/vmware/go-vcloud-director/pull/371)
* Field `types.DiskRecordType.SizeB` is replaced with `types.DiskRecordType.SizeMb` as size in Kilobytes is not
  supported in V33.0 [#371](https://github.com/vmware/go-vcloud-director/pull/371)
* Added parameter `additionalHeader map[string]string` to functions `Client.OpenApiDeleteItem`, `Client.OpenApiGetAllItems`,
  `Client.OpenApiGetItem`, `Client.OpenApiPostItem`, `Client.OpenApiPutItem`, `Client.OpenApiPutItemAsync`,
  `Client.OpenApiPutItemSync` [#380](https://github.com/vmware/go-vcloud-director/pull/380)
* Renamed functions `GetOpenApiRoleById` -> `GetRoleById`, `GetOpenApiRoleByName` -> `GetRoleByName`, 
  `GetAllOpenApiRoles` -> `GetAllRoles` [#380](https://github.com/vmware/go-vcloud-director/pull/380)

IMPROVEMENTS:
* Only send xml.Header when payload is not empty (some WAFs block empty requests with XML header) 
  [#367](https://github.com/vmware/go-vcloud-director/pull/367)
* Improved test entity cleanup to find standalone VMs in any VDC (not only default NSX-V one)
  [#368](https://github.com/vmware/go-vcloud-director/pull/368)
* Improved test entity cleanup to allow specifying parent VDC for vApp removals
  [#368](https://github.com/vmware/go-vcloud-director/pull/368)
* Cleanup a few unnecessary type conversions detected by new staticcheck version 
  [#381](https://github.com/vmware/go-vcloud-director/pull/381)
* Improved `OpenApiGetAllItems` to still follow pages in VCD endpoints with BUG which don't return 'nextPage' link for
  pagination [#378](https://github.com/vmware/go-vcloud-director/pull/378)
* Improved LDAP container related tests to use correct port mapping for latest LDAP container version 
  [#378](https://github.com/vmware/go-vcloud-director/pull/378)


## 2.11.0 (March 10, 2021)

* Added structure and methods to handle Org VDC networks using OpenAPI - `OpenApiOrgVdcNetwork`. It supports VCD 9.7+
for all networks types for NSX-V and NSX-T backed VDCs [#354](https://github.com/vmware/go-vcloud-director/pull/354)
* Added `NsxtImportableSwitch` structure with `GetNsxtImportableSwitchByName` and `GetAllNsxtImportableSwitches` to 
lookup NSX-T segments for use in NSX-T Imported networks [#354](https://github.com/vmware/go-vcloud-director/pull/354)
* Added `vdc.IsNsxt` and `vdc.IsNsxv` methods to verify if VDC is backed by NSX-T or NSX-V [#354](https://github.com/vmware/go-vcloud-director/pull/354)
* Added types `types.CreateVmParams` and `types.InstantiateVmTemplateParams`  [#356](https://github.com/vmware/go-vcloud-director/pull/356)
* Added VDC methods `CreateStandaloneVMFromTemplate`, `CreateStandaloneVMFromTemplateAsync` `CreateStandaloneVm`, 
`CreateStandaloneVmAsync` [#356](https://github.com/vmware/go-vcloud-director/pull/356)
* Added VDC methods `QueryVmByName`, `QueryVmById`, `QueryVmList` [#356](https://github.com/vmware/go-vcloud-director/pull/356)
* Added VM methods `Delete`, `DeleteAsync` [#356](https://github.com/vmware/go-vcloud-director/pull/356)
* Added VDC methods `GetOpenApiOrgVdcNetworkDhcp`, `UpdateOpenApiOrgVdcNetworkDhcp` and `DeleteOpenApiOrgVdcNetworkDhcp`
for OpenAPI management of Org Network DHCP configurations [#357](https://github.com/vmware/go-vcloud-director/pull/357)

BREAKING CHANGES:
* Renamed `types.VM` to `types.Vm` to facilitate implementation of standalone VM 
[#356](https://github.com/vmware/go-vcloud-director/pull/356)

BUGS FIXED:
* Converted IPAddress field for IPAddresses struct to array [#350](https://github.com/vmware/go-vcloud-director/pull/350)

IMPROVEMENTS:
* Added generic OpenAPI entity cleanup for tests [348](https://github.com/vmware/go-vcloud-director/pull/348)

## 2.10.0 (December 18, 2020)

* Added functions to retrieve and use VCD version `client.GetVcdVersion`, `client.GetVcdShortVersion`, `client.GetVcdFullVersion`, `client.VersionEqualOrGreater` [#339](https://github.com/vmware/go-vcloud-director/pull/339)
* Added methods `VM.UpdateStorageProfile`, `VM.UpdateStorageProfileAsync` [#338](https://github.com/vmware/go-vcloud-director/pull/338)
* Added methods `adminVdc.UpdateStorageProfile` [#340](https://github.com/vmware/go-vcloud-director/pull/340)
* Added transparent support for bearer tokens [#341](https://github.com/vmware/go-vcloud-director/pull/341)
* Added transparent connection using `cloudapi/1.0.0/sessions` when access through `api/sessions` is disabled
* Added functions `edge.GetLbAppRules`, `edge.GetLbServerPools`, `edge.GetLbAppProfiles`, `edge.GetNsxvNatRules`, `client.GetOrgList`
* Exported private function `client.maxSupportedVersion` to `client.MaxSupportedVersion`
* Able to upload an OVF without ovf:size defined in File part. Some bug fix for uploading OVA/OVF. [#331](https://github.com/vmware/go-vcloud-director/pull/331)
* Add support for handling catalog storage profile (`adminOrg.CreateCatalogWithStorageProfile`,
`org.CreateCatalogWithStorageProfile`, `adminCatalog.Update`) [#345](https://github.com/vmware/go-vcloud-director/pull/345)
* Add convenience functions `AdminOrg.GetAllStorageProfileReferences`, `AdminOrg.GetStorageProfileReferenceById`, `AdminOrg.GetAllVDCs` [#345](https://github.com/vmware/go-vcloud-director/pull/345)
* Added VCD 10.1+ functions `(vdc *Vdc) GetNsxtEdgeClusterByName` and `(vdc *Vdc) GetAllNsxtEdgeClusters` for NSX-T Edge Cluster lookup [#344](https://github.com/vmware/go-vcloud-director/pull/344)
* Added VCD 10.1+ NSX-T Edge Gateway management functions `GetNsxtEdgeGatewayById`, `GetNsxtEdgeGatewayByName`, `GetAllNsxtEdgeGateways`, `CreateNsxtEdgeGateway`, `Update`, `Delete` [#344](https://github.com/vmware/go-vcloud-director/pull/344)

BREAKING CHANGES:

* type.VdcConfiguration (used for creation) changed the type for storage profile from `[]*VdcStorageProfile` to `[]*VdcStorageProfileConfiguration`

## 2.9.0 (October 15, 2020)

* Improved testing tags isolation [#320](https://github.com/vmware/go-vcloud-director/pull/320)
* Added command `make tagverify` to check tags isolation tests [#320](https://github.com/vmware/go-vcloud-director/pull/320)
* Added methods `Client.GetAccessControl`, `Client.SetAccessControl`[#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Added methods `VApp.GetAccessControl`, `VApp.SetAccessControl`, `VApp.RemoveAccessControl`, `VApp.IsShared` [#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Added methods `AdminCatalog.GetAccessControl`, `AdminCatalog.SetAccessControl`, `AdminCatalog.RemoveAccessControl`, `AdminCatalog.IsShared` [#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Added methods `Catalog.GetAccessControl`, `Catalog.SetAccessControl`, `Catalog.RemoveAccessControl`, `Catalog.IsShared` [#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Added methods `Vdc.GetVappAccessControl`, `AdminOrg.GetCatalogAccessControl`, `Org.GetCatalogAccessControl` [#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Added methods `Vdc.QueryVappList`, `Vdc.GetVappList`, `AdminVdc.GetVappList`, `client.GetQueryType` [#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Added VM and vApp to search query engine [#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Added tenant context for access control methods [#329](https://github.com/vmware/go-vcloud-director/pull/329)
* Loosen up `Test_LBAppRule` for invalid application script check to work with different error engine in VCD 10.2
[#326](https://github.com/vmware/go-vcloud-director/pull/326)
* Update VDC dynamic func to handle API version 35.0 [#327](https://github.com/vmware/go-vcloud-director/pull/327)
* Added methods `vm.UpdateVmCpuAndMemoryHotAdd` and `vm.UpdateVmCpuAndMemoryHotAddAsyc` [#324](https://github.com/vmware/go-vcloud-director/pull/324)
* Introduce low level OpenAPI client functions `OpenApiGetAllItems`,`OpenApiPostItemSync`,`OpenApiPostItemAsync`,
`OpenApiPostItem`, `OpenApiGetItem`, `OpenApiPutItem`, `OpenApiPutItemSync`, `OpenApiPutItemAsync`,
`OpenApiDeleteItem`, `OpenApiIsSupported`, `OpenApiBuildEndpoints`
[#325](https://github.com/vmware/go-vcloud-director/pull/325), [#333](https://github.com/vmware/go-vcloud-director/pull/333)
* Add OVF file upload support in UploadOvf function besides OVA. The input should be OVF file path inside the OVF folder. It will check if input file is XML content type, if yes, skip some OVA steps (like unpacking), if not, keep the old logic. [#323](https://github.com/vmware/go-vcloud-director/pull/323)
* Dropped support for VMware Cloud Director 9.5 [#330](https://github.com/vmware/go-vcloud-director/pull/330)
* Deprecated Vdc.UploadMediaImage because it no longer works with API V32.0+ [#330](https://github.com/vmware/go-vcloud-director/pull/330)
* Add methods `vapp.AddNewVMWithComputePolicy`, `org.GetVdcComputePolicyById`, `adminOrg.GetVdcComputePolicyById`, `org.GetAllVdcComputePolicies`, `adminOrg.GetAllVdcComputePolicies`, `adminOrg.CreateVdcComputePolicy`, `vdcComputePolicy.Update`, `vdcComputePolicy.Delete`, `adminVdc.GetAllAssignedVdcComputePolicies` and `adminVdc.SetAssignedComputePolicies` [#334] (https://github.com/vmware/go-vcloud-director/pull/334)
* Introduce NSX-T support for adminOrg.CreateOrgVdc() [#332](https://github.com/vmware/go-vcloud-director/pull/332)
* Introduce NSX-T support for external network using OpenAPI endpoint and `ExternalNetworkV2` type methods including `CreateExternalNetworkV2`, 
`GetExternalNetworkById`, `GetAllExternalNetworks`, `ExternalNetworkV2.Update`, and `ExternalNetworkV2.DELETE` [#335](https://github.com/vmware/go-vcloud-director/pull/335)
* Introduce NSX-T Query functions `client.QueryNsxtManagerByName` and `client.GetImportableNsxtTier0RouterByName` [#335](https://github.com/vmware/go-vcloud-director/pull/335)
* Add HTTP User-Agent header `go-vcloud-director` to all API calls and allow to customize it using
  `WithHttpUserAgent` configuration options function [#336](https://github.com/vmware/go-vcloud-director/pull/336)

## 2.8.0 (June 30, 2020)

* Changed signature for `FindAdminCatalogRecords`, which now returns normalized type `[]*types.CatalogRecord` [#298](https://github.com/vmware/go-vcloud-director/pull/298)
* Added methods `catalog.QueryVappTemplateList`, `catalog.QueryCatalogItemList`, `client.queryWithMetadataFields`, `client.queryByMetadataFilter` [#298](https://github.com/vmware/go-vcloud-director/pull/298)
* Added query engine based on `client.SearchByFilter`, type `FilterDef`, and interface `QueryItem` [#298](https://github.com/vmware/go-vcloud-director/pull/298)
* Added methods `adminOrg.QueryCatalogList` and `org.QueryCatalogList` [#298](https://github.com/vmware/go-vcloud-director/pull/298)
* Removed code that handled specific cases for API 29.0 and 30.0. This library now supports VCD versions from 9.5 to 10.1 included.
* Added `vdc.QueryVappVmTemplate` and changed `vapp.AddNewVMWithStorageProfile` to allow creating VM from VM template.
* Enhanced tests command line with flags that can be used instead of environment variables. [#305](https://github.com/vmware/go-vcloud-director/pull/305)
* Improve logging security of debug output for API requests and responses [#306](https://github.com/vmware/go-vcloud-director/pull/306)
* Append log files by default instead of overwriting. `GOVCD_LOG_OVERWRITE=true` environment
  variable can set to overwrite log file on every initialization
  [#307](https://github.com/vmware/go-vcloud-director/pull/307)
* Add configuration option `WithSamlAdfs` to `NewVCDClient()` to support SAML authentication using
  Active Directory Federations Services (ADFS) as IdP using WS-TRUST auth endpoint
  "/adfs/services/trust/13/usernamemixed"
  [#304](https://github.com/vmware/go-vcloud-director/pull/304)
* Implemented VM affinity rules CRUD: `vdc.CreateVmAffinityRuleAsync`, `vdc. CreateVmAffinityRule`, `vdc.GetAllVmAffinityRuleList`, `vdc.GetVmAffinityRuleList`, `vdc.GetVmAntiAffinityRuleList`
 `vdc.GetVmAffinityRuleByHref`, `vdc.GetVmAffinityRulesByName`, `vdc.GetVmAffinityRuleById`, `vdc.GetVmAffinityRuleByNameOrId`, `VmAffinityRule.Delete`, `VmAffinityRule.Update`,
 `VmAffinityRule.SetMandatory`, `VmAffinityRule.SetEnabled`, `VmAffinityRule.Refresh` [#313](https://github.com/vmware/go-vcloud-director/pull/313)
* Add method `client.QueryVmList` [#313](https://github.com/vmware/go-vcloud-director/pull/313)
* Add support for group management using `CreateGroup`, `GetGroupByHref`, `GetGroupById`,
  `GetGroupByName`, `GetGroupByNameOrId`, `Delete`, `Update`, `NewGroup` functions [#314](https://github.com/vmware/go-vcloud-director/pull/314)
* Add LDAP administration functions for Org `LdapConfigure`, `GetLdapConfiguration`, and `LdapDisable` [#314](https://github.com/vmware/go-vcloud-director/pull/314)
* Added methods `vapp.UpdateNetworkFirewallRules`, `vapp.UpdateNetworkFirewallRulesAsync`, `vapp.GetVappNetworkById`, `vapp.GetVappNetworkByName` and `vapp.GetVappNetworkByNameOrId` [#308](https://github.com/vmware/go-vcloud-director/pull/308)
* Added methods `vapp.UpdateNetworkNatRulesAsync`, `vapp.UpdateNetworkNatRulesAsync`, `vapp.RemoveAllNetworkFirewallRules` and `vapp.RemoveAllNetworkNatRules` [#316](https://github.com/vmware/go-vcloud-director/pull/316)
* Added methods `vapp.UpdateNetworkStaticRouting`, `vapp.UpdateNetworkStaticRoutingAsync` and `vapp.RemoveAllNetworkStaticRoutes` [#318](https://github.com/vmware/go-vcloud-director/pull/318)

## 2.7.0 (April 10,2020)

* Added methods `OrgVdcNetwork.Update`, `OrgVdcNetwork.UpdateAsync`, and `OrgVdcNetwork.Rename` [#292](https://github.com/vmware/go-vcloud-director/pull/292)
* Added methods `EdgeGateway.Update` and `EdgeGateway.UpdateAsync` [#292](https://github.com/vmware/go-vcloud-director/pull/292)
* Increment vCD API version used from 29.0 to 31.0
    * Add fields `AdminVdc.UniversalNetworkPoolReference and VM.Media`    
* Added methods `vapp.AddEmptyVm`, `vapp.AddEmptyVmAsync` and `vdc.QueryAllMedia` [#296](https://github.com/vmware/go-vcloud-director/pull/296)

NOTES:

* Improved test in function `deleteVapp()` to avoid deletion errors during test suite run
  [#297](https://github.com/vmware/go-vcloud-director/pull/297)

BUGS FIXED:
* Fix issue in Queries with vCD 10 version, which do not return network pool or provider VDC[#293](https://github.com/vmware/go-vcloud-director/pull/293)
* Session timeout for media, catalog item upload  [#294](https://github.com/vmware/go-vcloud-director/pull/294)
* Fix `vapp.RemoveNetwork`, `vapp.RemoveNetworkAsync` to use `DELETE` API call instead of update
  which can apply incorrect remaining vApp network configurations [#299](https://github.com/vmware/go-vcloud-director/pull/299)

## 2.6.0 (March 13, 2020)

* Moved `VCDClient.supportedVersions` to `VCDClient.Client.supportedVersions` [#274](https://github.com/vmware/go-vcloud-director/pull/274)    
* Added methods `VM.AddInternalDisk`, `VM.GetInternalDiskById`, `VM.DeleteInternalDisk`, `VM.UpdateInternalDisks` and `VM.UpdateInternalDisksAsync` [#272](https://github.com/vmware/go-vcloud-director/pull/272)
* Added methods `vdc.GetEdgeGatewayReferenceList` and `catalog.GetVappTemplateByHref` [#278](https://github.com/vmware/go-vcloud-director/pull/278)
* Improved functions to not expect XML namespaces provided in argument structure [#284](https://github.com/vmware/go-vcloud-director/pull/284)
* Change `int` and `bool` fields from types.VAppTemplateLeaseSettings and VAppLeaseSettings into pointers
* Added method `catalog.GetVappTemplateByHref`, and expose methods `vdc.GetEdgeGatewayByHref` and `vdc.GetEdgeGatewayRecordsType`
* Added methods `adminOrg.CreateOrgVdc`, `adminOrg.CreateOrgVdcAsync` and improved existing to support Flex VDC model. These new methods are dynamic as they change invocation behind the scenes based on vCD version [#285](https://github.com/vmware/go-vcloud-director/pull/285) 
* Deprecated functions `adminOrg.CreateVdc` and `adminOrg.CreateVdcWait` [#285](https://github.com/vmware/go-vcloud-director/pull/285)
* Added methods `EdgeGateway.GetAllNsxvDhcpLeases()`, `EdgeGateway.GetNsxvActiveDhcpLeaseByMac()`
  `VM.WaitForDhcpIpByNicIndexes()`, `VM.GetParentVApp()`, `VM.GetParentVdc()`
  [#283](https://github.com/vmware/go-vcloud-director/pull/283)
* `types.GetGuestCustomizationSection` now uses pointers for all bool values to distinguish between empty and false value [#291](https://github.com/vmware/go-vcloud-director/pull/291)
* Deprecated functions `Vapp.Customize()` and `VM.Customize()` in favor of `vm.SetGuestCustomizationSection` [#291](https://github.com/vmware/go-vcloud-director/pull/291)
* Added methods `vapp.AddNetwork`, `vapp.AddNetworkAsync`, `vapp.AddOrgNetwork`, `vapp.AddOrgNetworkAsync`, `vapp.UpdateNetwork`, `vapp.UpdateNetworkAsync`, `vapp.UpdateOrgNetwork`, `vapp.UpdateOrgNetworkAsync`, `vapp.RemoveNetwork`, `vapp.RemoveNetworkAsync` and `GetUuidFromHref` [#289](https://github.com/vmware/go-vcloud-director/pull/290)
* Deprecated functions `vapp.RemoveIsolatedNetwork`, `vapp.AddRAWNetworkConfig` and `vapp.AddIsolatedNetwork`  [#289](https://github.com/vmware/go-vcloud-director/pull/290)

BUGS FIXED:
* A data race in catalog/media item upload status reporting [#288](https://github.com/vmware/go-vcloud-director/pull/288)
* `Vapp.Customize()` and `VM.Customize()` ignores `changeSid` value and always set it to true [#291](https://github.com/vmware/go-vcloud-director/pull/291)

## 2.5.1 (December 12, 2019)

BUGS FIXED:
* Fix a bug where functions `GetAnyVnicIndexByNetworkName` and `GetVnicIndexByNetworkNameAndType`
  would not find vNic index when user is authenticated as org admin (not sysadmin)
  [#275](https://github.com/vmware/go-vcloud-director/pull/275)

## 2.5.0 (December 11, 2019)

* Change fields ResourceGuaranteedCpu, VCpuInMhz, IsThinProvision, NetworkPoolReference,
  ProviderVdcReference and UsesFastProvisioning in AdminVdc to pointers to allow understand if value
  was returned or not. 
* Added method VApp.AddNewVMWithStorageProfile that adds a VM with custom storage profile.
* Added command `make static` to run staticcheck on all packages
* Added `make static` to Travis regular checks
* Added ability to connect to the vCD using an authorization token
* Added method `VCDClient.SetToken`
* Added method `VCDClient.GetAuthResponse`
* Added script `scripts/get_token.sh`
* Incremented vCD API version used from 27.0 to 29.0
    * Remove fields `VdcEnabled`, `VAppParentHREF`, `VAppParentName`, `HighestSupportedVersion`, `VmToolsVersion`, `TaskHREF`, `TaskStatusName`, `TaskDetails`, `TaskStatus` from `QueryResultVMRecordType`
    * Added fields `ID, Type, ContainerName, ContainerID, OwnerName, Owner, NetworkHref, IpAddress, CatalogName, VmToolsStatus, GcStatus, AutoUndeployDate, AutoDeleteDate, AutoUndeployNotified, AutoDeleteNotified, Link, MetaData` to `QueryResultVMRecordType`, `DistributedInterface` to `NetworkConfiguration` and `RegenerateBiosUuid` to `VMGeneralParams`
    * Change to pointers `DistributedRoutingEnabled` in `GatewayConfiguration` and
    `DistributedInterface` in `NetworkConfiguration`
* Added new field to type `GatewayConfiguration`: `FipsModeEnabled` -
  [#267](https://github.com/vmware/go-vcloud-director/pull/267)
* Change bool to bool pointer for fields in type `GatewayConfiguration`: `HaEnabled`,
  `UseDefaultRouteForDNSRelay`, `AdvancedNetworkingEnabled` -
  [#267](https://github.com/vmware/go-vcloud-director/pull/267)
* Added method `EdgeGateway.GetLbVirtualServers` that gets all virtual servers configured on NSX load balancer. [#266](https://github.com/vmware/go-vcloud-director/pull/266)
* Added method `EdgeGateway.GetLbServerPools` that gets all pools configured on NSX load balancer. [#266](https://github.com/vmware/go-vcloud-director/pull/266)
* Added method `EdgeGateway.GetLbServiceMonitors` that gets all service monitors configured on NSX load balancer. [#266](https://github.com/vmware/go-vcloud-director/pull/266)
* Added field `SubInterface` to `NetworkConfiguration`. [#321](https://github.com/terraform-providers/terraform-provider-vcd/issues/321)
* Added methods `Vdc.FindEdgeGatewayNameByNetwork` and `Vdc.GetNetworkList`
* Added IP set handling functions `CreateNsxvIpSet`, `UpdateNsxvIpSet`, `GetNsxvIpSetByName`,
  `GetNsxvIpSetById`, `GetNsxvIpSetByNameOrId`, `GetAllNsxvIpSets`, `DeleteNsxvIpSetById`,
  `DeleteNsxvIpSetByName` [#269](https://github.com/vmware/go-vcloud-director/pull/269)
* Added `UpdateDhcpRelay`, `GetDhcpRelay` and `ResetDhcpRelay` methods for Edge Gatway DHCP relay
  management [#271](https://github.com/vmware/go-vcloud-director/pull/271)
* Added methods which allow override API versions `NewRequestWitNotEncodedParamsWithApiVersion`, 
   `ExecuteTaskRequestWithApiVersion`, `ExecuteRequestWithoutResponseWithApiVersion`,
   `ExecuteRequestWithApiVersion` [#274](https://github.com/vmware/go-vcloud-director/pull/274)

BUGS FIXED:
* Remove parentheses from filtering since they weren't treated correctly in some environment [#256]
  (https://github.com/vmware/go-vcloud-director/pull/256)
* Take into account all subnets (SubnetParticipation) on edge gateway interface instead of the first
  one [#260](https://github.com/vmware/go-vcloud-director/pull/260)
* Fix `OrgVdcNetwork` data structure to retrieve description. Previously, the description would not be retrieved because it was misplaced in the sequence.

## 2.4.0 (October 28, 2019)

* Deprecated functions `GetOrgByName` and `GetAdminOrgByName`
* Deprecated methods `AdminOrg.FetchUserByName`, `AdminOrg.FetchUserById`, `AdminOrg.FetchUserByNameOrId`, `AdminOrg.GetRole`.
* Added method `VCDClient.GetOrgByName`  and related `GetOrgById`, `GetOrgByNameOrId`
* Added method `VCDClient.GetAdminOrgByName` and related `GetAdminOrgById`, `GetAdminOrgByNameOrId`
* Added methods `AdminOrg.GetUserByName`, `GetUserById`, `GetUserByNameOrId`, `GetRoleReference`.
* Added method `VCDClient.QueryProviderVdcs` 
* Added method `VCDClient.QueryProviderVdcStorageProfiles` 
* Added method `VCDClient.QueryNetworkPools` 
* Added get/add/delete metadata functions for vApp template and media item [#225](https://github.com/vmware/go-vcloud-director/pull/225).
* Added `UpdateNetworkConnectionSection` for updating VM network configuration [#229](https://gifiltering which in some env wasn'tthub.com/vmware/go-vcloud-director/pull/229)
* Added `PowerOnAndForceCustomization`, `GetGuestCustomizationStatus`, `BlockWhileGuestCustomizationStatus` [#229](https://github.com/vmware/go-vcloud-director/pull/229)
* Deprecated methods `AdminOrg.GetAdminVdcByName`, `AdminOrg.GetVdcByName`, `AdminOrg.FindAdminCatalog`, `AdminOrg.FindCatalog`
* Deprecated methods `Catalog.FindCatalogItem`, `Org.FindCatalog`, `Org.GetVdcByName`
* Deprecated function `GetExternalNetwork`
* Added methods `Org.GetCatalogByName` and related `Org.GetCatalogById`, `GetCatalogItemByNameOrId`
* Added methods `VCDClient.GetExternalNetworkByName` and related `GetExternalNetworkById` and `GetExternalNetworkByNameOrId`
* Added methods `AdminOrg.GetCatalogByName` and related `Org.GetCatalogById`, `GetCatalogByNameOrId`
* Added methods `AdminOrg.GetAdminCatalogByName` and related `Org.GetAdminCatalogById`, `GetAdminCatalogByNameOrId`
* Added methods `Org.GetVDCByName` and related `GetVDCById`, `GetVDCByNameOrId`
* Added methods `AdminOrg.GetVDCByName` and related `GetVDCById`, `GetVDCByNameOrId`
* Added methods `AdminOrg.GetAdminVDCByName` and related `GetAdminVDCById`, `GetAdminVDCByNameOrId`
* Added methods `Catalog.Refresh` and `AdminCatalog.Refresh`
* Added method `vm.GetVirtualHardwareSection` to retrieve virtual hardware items [#200](https://github.com/vmware/go-vcloud-director/pull/200)
* Added methods `vm.SetProductSectionList` and `vm.GetProductSectionList` allowing to manipulate VM
guest properties [#235](https://github.com/vmware/go-vcloud-director/pull/235)
* Added methods `vapp.SetProductSectionList` and `vapp.GetProductSectionList` allowing to manipulate
vApp guest properties [#235](https://github.com/vmware/go-vcloud-director/pull/235)
* Added method GetStorageProfileByHref
* Added methods `CreateNsxvNatRule()`, `UpdateNsxvNatRule()`, `GetNsxvNatRuleById()`, `DeleteNsxvNatRuleById()`
which use the proxied NSX-V API of advanced edge gateway for handling NAT rules [#241](https://github.com/vmware/go-vcloud-director/pull/241)
* Added methods `GetVnicIndexByNetworkNameAndType()` and `GetNetworkNameAndTypeByVnicIndex()` [#241](https://github.com/vmware/go-vcloud-director/pull/241)
* Added methods `Vdc.GetVappByHref`, `Vdc.GetVAppByName` and related `GetVAppById`, `GetVAppByNameOrId`
* Added methods `Client.GetVMByHref` `Vapp.GetVAMByName` and related `GetVMById`, `GetVAMByNameOrId`
* Deprecated methods `Client.FindVMByHREF`, `Vdc.FindVMByName`, `Vdc.FindVAppByID`, and `Vdc.FindVAppByName`
* Added methods `Vm.GetGuestCustomizationSection` and `Vm.SetGuestCustomizationSection`  
* Added methods `CreateNsxvFirewallRule()`, `UpdateNsxvFirewallRule()`, `GetNsxvFirewallRuleById()`, `DeleteNsxvFirewallRuleById()`
which use the proxied NSX-V API of advanced edge gateway for handling firewall rules [#247](https://github.com/vmware/go-vcloud-director/pull/247)
* Added methods `GetFirewallParams()`, `UpdateFirewallParams()` for changing global firewall settings [#247](https://github.com/vmware/go-vcloud-director/pull/247)
* Added method `GetAnyVnicIndexByNetworkName()` to for easier interface (vNic) lookup in edge gateway [#247](https://github.com/vmware/go-vcloud-director/pull/247)
* Added method `ExecuteParamRequestWithCustomError()` which adds query parameter support on top of `ExecuteRequestWithCustomError()` [#247](https://github.com/vmware/go-vcloud-director/pull/247)
* Deprecated methods `VDC.FindDiskByHREF` and `FindDiskByHREF`
* Added methods `VDC.GetDiskByHref` `VDC.GetDisksByName` and related `GetDiskById`
* Added new methods `Catalog.QueryMedia`, `Catalog.GetMediaByName`, `Catalog.GetMediaById`, `Catalog.GetMediaByNameOrId`, `AdminCatalog.QueryMedia`, `AdminCatalog.GetMediaByName`, `AdminCatalog.GetMediaById`, `AdminCatalog.GetMediaByNameOrId`, `MediaRecord.Refresh`, `MediaRecord.Delete`, `MediaRecord.GetMetadata`, `MediaRecord.AddMetadata`, `MediaRecord.AddMetadataAsync`, `MediaRecord.DeleteMetadata`, `MediaRecord.DeleteMetadataAsync`, `Media.GetMetadata`, `Media.AddMetadata`, `Media.AddMetadataAsync`, `Media.DeleteMetadata`, `Media.DeleteMetadataAsync` [#245](https://github.com/vmware/go-vcloud-director/pull/245)
* Deprecated methods `Vdc.FindMediaImage`, `MediaItem`, `RemoveMediaImageIfExists`, `MediaItem.Delete`, `FindMediaAsCatalogItem`, `*MediaItem.Refresh`, `MediaItem.GetMetadata`, `MediaItem.AddMetadata`, `MediaItem.AddMetadataAsync`, `MediaItem.DeleteMetadata`, `MediaItem.DeleteMetadataAsync` [#245](https://github.com/vmware/go-vcloud-director/pull/245)
* Added method `VDC.QueryDisks` [#255](https://github.com/vmware/go-vcloud-director/pull/255)

IMPROVEMENTS:

* Move methods for `AdminOrg`, `AdminCatalog`, `AdminVdc` to new files `adminorg.go`,
 `admincatalog.go`, `adminvdc.go`.
* Added default value for HTTP timeout (600s) which is configurable

BUGS FIXED:

* Fix bug in AdminOrg.Update, where OrgGeneralSettings would not update correctly if it contained only one property
* Fix bug in External network creation and get when description wasn't populated.
* Fix bug in VDC creation when name with space caused an error
* Fix bug in Org Delete, which would remove catalogs shared from other organizations.
* Fix Vcd.StorageProfiles type from array to single.
* Fix AdminOrg.CreateUserSimple, where the Telephone field was ignored.

## 2.3.1 (July 29, 2019)

BUG FIXES:

* Remove `omitempty` struct tags from load balancer component boolean fields to allow sending `false` values to API [#222](https://github.com/vmware/go-vcloud-director/pull/222)

## 2.3.0 (July 26, 2019)

* Added edge gateway create/delete functions [#130](https://github.com/vmware/go-vcloud-director/issues/130).
* Added edge gateway global load balancer configuration support (e.g. enable/disable) [#219](https://github.com/vmware/go-vcloud-director/pull/219)
* Added load balancer service monitor [#196](https://github.com/vmware/go-vcloud-director/pull/196)
* Added load balancer server pool [#205](https://github.com/vmware/go-vcloud-director/pull/205)
* Added load balancer application profile [#208](https://github.com/vmware/go-vcloud-director/pull/208)
* Added load balancer application rule [#212](https://github.com/vmware/go-vcloud-director/pull/212)
* Added load balancer virtual server [#215](https://github.com/vmware/go-vcloud-director/pull/215)
* Added functions for refreshing, getting and update Org VDC [#206](https://github.com/vmware/go-vcloud-director/pull/206)
* Added VDC meta data create/get/delete functions [#203](https://github.com/vmware/go-vcloud-director/pull/203)
* Added org user create/delete/update functions [#18](https://github.com/vmware/go-vcloud-director/issues/18)
* Added load balancer application profile [#208](https://github.com/vmware/go-vcloud-director/pull/208)
* Added edge gateway SNAT/DNAT rule functions which support org VDC network and external network [#225](https://github.com/terraform-providers/terraform-provider-vcd/issues/225)
* Added edge gateway SNAT/DNAT rule functions which work with IDs [#244](https://github.com/terraform-providers/terraform-provider-vcd/issues/244)

## 2.2.0 (May 15, 2019)

FEATURES:

* Added external network get/create/delete functions
* Added metadata add/remove functions to VM.
* Added ability to do vCD version checks and comparison [#174](https://github.com/vmware/go-vcloud-director/pull/174)
using VCDClient.APIVCDMaxVersionIs(string) and VCDClient.APIClientVersionIs(string).
* Added ability to override currently used vCD API version WithAPIVersion(string) [#174](https://github.com/vmware/go-vcloud-director/pull/174).
* Added ability to enable nested hypervisor option for VM with VM.ToggleNestedHypervisor(bool) [#219](https://github.com/terraform-providers/terraform-provider-vcd/issues/219).


BREAKING CHANGES:

* vApp metadata now is attached to the vApp rather to first VM in vApp.
* vApp metadata is no longer added to first VM in vApp it will be added to vApp directly instead.

IMPROVEMENTS:
* Refactored code by introducing helper function to handle API calls. New functions ExecuteRequest, ExecuteTaskRequest, ExecuteRequestWithoutResponse
* Added authorization request header for media file and catalog item upload
* Tests files are now all tagged. Running them through Makefile works as before, but manual execution requires specific tags. Run `go test -v .` for tags list.

## 2.1.0 (March 21, 2019)

ARCHITECTURAL:

* Project switched to using Go modules. It is worth having a
look at [README.md](README.md) to understand how Go modules impact build and development.

FEATURES:

* New insert and eject media functions

IMPROVEMENTS:

* vApp vapp.PowerOn() implicitly waits for vApp to exit "UNRESOLVED" state which occurs shortly after creation and causes vapp.PowerOn() failure.
* VM has new functions which allows to configure cores for CPU. VM.ChangeCPUCountWithCore()

BREAKING CHANGES:

* Deprecate vApp.ChangeCPUCountWithCore() and vApp.ChangeCPUCount()
