## 2.12.0 (unreleased)

* Added method `vdc.QueryEdgeGateway` [#364](https://github.com/vmware/go-vcloud-director/pull/364)
* Deprecated `vdc.GetEdgeGatewayRecordsType` [#364](https://github.com/vmware/go-vcloud-director/pull/364)

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
* Increment vCD API version used from 27.0 to 29.0
    * Remove fields `VdcEnabled`, `VAppParentHREF`, `VAppParentName`, `HighestSupportedVersion`, `VmToolsVersion`, `TaskHREF`, `TaskStatusName`, `TaskDetails`, `TaskStatus` from `QueryResultVMRecordType`
    * Add fields `ID, Type, ContainerName, ContainerID, OwnerName, Owner, NetworkHref, IpAddress, CatalogName, VmToolsStatus, GcStatus, AutoUndeployDate, AutoDeleteDate, AutoUndeployNotified, AutoDeleteNotified, Link, MetaData` to `QueryResultVMRecordType`, `DistributedInterface` to `NetworkConfiguration` and `RegenerateBiosUuid` to `VMGeneralParams`
    * Change to pointers `DistributedRoutingEnabled` in `GatewayConfiguration` and
    `DistributedInterface` in `NetworkConfiguration`
* Add new field to type `GatewayConfiguration`: `FipsModeEnabled` -
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
* Add authorization request header for media file and catalog item upload
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
