* Add `TmOperationOptions` structure with `APIVersion` and `TenantContext` properties for per-call
  Tenant Manager API-version and organization-context overrides [GH-829]
* `Client.IsTm` and `Client.OpenApiIsSupported` recognize servers that advertise only TM product versions such as
  `9.1.2`. TM operations compare the legacy aliases as release versions (`40.0` as `9.0.0`, `40.1` as `9.1.0`)
  and send the exact version string advertised by the server [GH-829]
* Run all OpenAPI requests (`OpenApiGetItem*`, `OpenApiPostItem*`, `OpenApiPutItem*`, and
  `OpenApiDeleteItem`) through a single implementation. Asynchronous tasks keep the API version and tenant
  context of the request that created them, and response bodies are closed on every path [GH-829]
* `ContentLibrary` methods send the tenant context used to retrieve or create the content library [GH-829]
* Add `VCDClient.CreateRegionStoragePolicy`, `VCDClient.CreateRegionStoragePolicyAsync`,
  `RegionStoragePolicy.Update`, and `RegionStoragePolicy.Delete` methods to create, update, and
  delete region storage policies [GH-829]
* Add `Region.DeleteWithForce`, `RegionQuota.DeleteWithForce`, and
  `ContentLibraryItem.DeleteWithForce` methods to request force deletion [GH-829]
* Add `ContentLibrary.Refresh`, `ContentLibrary.RefreshAsync`, `ContentLibrary.Sync`, and
  `ContentLibrary.SyncAsync` methods to refresh subscription metadata and synchronize subscribed
  content [GH-829]
* Add `ContentLibrary.GetProjectPermissions`, `ContentLibrary.UpdateProjectPermissions`, and
  `ContentLibrary.UpdateProjectPermissionsAsync` methods to read and update content-library project
  permissions [GH-829]
* Add `ContentLibraryItem.Promote` and `ContentLibraryItem.PromoteAsync` methods to promote a
  project content-library item to a provider library [GH-829]
* Add `VCDClient.GenerateContentLibraryDownloadBundle` and
  `VCDClient.GenerateContentLibraryDownloadBundleAsync` methods to generate OVA download bundles
  [GH-829]
* Add `Region.GetAllNamespaceCompatibleZoneGroups` and `Region.GetAllCapabilities` methods to
  retrieve namespace-compatible zone groups and region capabilities [GH-829]
* Add `VCDClient.GetAllTmComputeClusters` and `VCDClient.GetTmComputeClusterById` methods to
  retrieve compute-cluster inventory [GH-829]
* Add `SupervisorZone.GetAllVmClassReservations` method to retrieve supervisor-zone VM-class
  reservation capacity [GH-829]
* Add `VCDClient.GetAllTmReservedVmClasses` method to retrieve reserved VM-class inventory [GH-829]
* Add `VCDClient.AddVmClassesToRegionQuota`, `RegionQuota.AddVmClasses`, and
  `RegionQuota.GetAllVmClasses` methods to add and list region-quota VM-class assignments [GH-829]
* Add `RegionQuota.GetAllReservedVmClasses` and `RegionQuota.UpdateReservedVmClasses` methods to
  read and update region-quota VM-class reservations [GH-829]
* Add `RegionQuota.GetAllNamespaces` method to retrieve region-quota namespace summaries [GH-829]
* Add `RegionQuota.UpdateStoragePolicies` method to update all storage-policy assignments and limits
  for a region quota [GH-829]
* Add `RegionQuota.GetAllInfraPolicies` and `RegionQuota.UpdateInfraPolicies` methods to read and
  replace region-quota infrastructure-policy assignments [GH-829]
* Add `TmIpSpace.GetUtilizationSummary` method to retrieve IP-space, IP-block, and IP-range
  utilization [GH-829]
* Add `TmRegionalNetworkingSetting.GetAllProviderGatewayConnections`,
  `TmRegionalNetworkingSetting.GetProviderGatewayConnectionById`,
  `TmRegionalNetworkingSetting.CreateProviderGatewayConnection`,
  `TmRegionalNetworkingSetting.CreateProviderGatewayConnectionAsync`,
  `TmRegionalNetworkingSetting.DeleteProviderGatewayConnection`, and
  `TmRegionalNetworkingSetting.DeleteProviderGatewayConnectionAsync` methods to manage regional
  provider-gateway connections [GH-829]
* Add `TmRegionalNetworkingSetting.GetAllSharedSubnets`,
  `TmRegionalNetworkingSetting.GetSharedSubnetById`,
  `TmRegionalNetworkingSetting.CreateSharedSubnet`,
  `TmRegionalNetworkingSetting.CreateSharedSubnetAsync`,
  `TmRegionalNetworkingSetting.DeleteSharedSubnet`, and
  `TmRegionalNetworkingSetting.DeleteSharedSubnetAsync` methods to manage regional shared-subnet
  associations [GH-829]
* Add `TmRegionalNetworkingSetting.GetAllDistributedVlanConnections`,
  `TmRegionalNetworkingSetting.GetDistributedVlanConnectionById`,
  `TmRegionalNetworkingSetting.CreateDistributedVlanConnection`,
  `TmRegionalNetworkingSetting.CreateDistributedVlanConnectionAsync`,
  `TmRegionalNetworkingSetting.DeleteDistributedVlanConnection`, and
  `TmRegionalNetworkingSetting.DeleteDistributedVlanConnectionAsync` methods to manage regional
  distributed VLAN connections [GH-829]
* Add `TmRegionalNetworkingSetting.GetAllDistributedVxlanConnections`,
  `TmRegionalNetworkingSetting.GetDistributedVxlanConnectionById`,
  `TmRegionalNetworkingSetting.CreateDistributedVxlanConnection`,
  `TmRegionalNetworkingSetting.CreateDistributedVxlanConnectionAsync`,
  `TmRegionalNetworkingSetting.DeleteDistributedVxlanConnection`, and
  `TmRegionalNetworkingSetting.DeleteDistributedVxlanConnectionAsync` methods to manage regional
  distributed VXLAN connections [GH-829]
* Add `TmRegionalNetworkingSetting.GetAviSetting`, `TmRegionalNetworkingSetting.UpdateAviSetting`,
  and `TmRegionalNetworkingSetting.UpdateAviSettingAsync` methods to read and update regional Avi
  settings [GH-829]
* Add `IsRestricted` property to `types.RegionStoragePolicy` [GH-829]
* Add `ResultTotal`, `PageCount`, `Page`, and `PageSize` properties to
  `types.RegionVirtualMachineClasses` [GH-829]
* Add `ResultTotal`, `PageCount`, `Page`, and `PageSize` properties to
  `types.VirtualDatacenterStoragePolicies` [GH-829]
* Add `types.ContentLibraryDownloadRequest` structure with `BundleName`, `MetadataFileContent`,
  `ImageIdentifiers`, and `ProjectID` properties for OVA bundle generation requests [GH-829]
* Add `types.ContentLibraryItemPromotion` structure with `TargetContentLibrary` property for
  content-library item promotion requests [GH-829]
* Add `types.ContentLibraryProjectPermissions` structure with `Values`, `ResultTotal`, `PageCount`,
  `Page`, and `PageSize` properties for explicit content-library project permissions and pagination
  [GH-829]
* Add `types.TmCompatibleZoneGroup` structure with `Zones` property for namespace-compatible zone
  groups [GH-829]
* Add `types.TmComputeCluster` structure with `ID`, `Address`, `VirtualCenter`, `Moref`,
  `Supervisor`, and `Zone` properties for compute-cluster inventory [GH-829]
* Add `types.TmIpSpaceUtilizationCounts` structure with `Name`, `TotalCount`, `UsedCount`,
  `AvailableCount`, and `UsedPercentage` properties for IP utilization counts [GH-829]
* Add `types.TmIpSpaceUtilizationSummary` structure with `TotalUtilization`, `IpBlockUtilizations`,
  and `IpRangeUtilizations` properties for IP-space utilization summaries [GH-829]
* Add `types.TmNamespaceSummary` structure with `ID`, `Name`, `Description`, `ProjectAssignment`,
  `Organization`, `Region`, `VirtualPrivateCloudName`, `WorkloadNetwork`, `ZonalResourceAllocation`,
  `ProxyURL`, `Status`, `VcReference`, `SegName`, and `LastModifiedTime` properties for region-quota
  namespace summaries [GH-829]
* Add `types.TmNamespaceSummaryResourceAllocation` structure with `MemoryLimitMiB`,
  `MemoryReservationMiB`, `CpuLimitMHz`, `CpuReservationMHz`, `VmClassReservations`,
  `RealizedMemoryLimitMiB`, `RealizedMemoryReservationMiB`, `RealizedCpuLimitMHz`,
  `RealizedCpuReservationMHz`, `MemoryUsedMiB`, `CpuUsedMHz`, and `MarkedForRemoval` properties for
  requested and realized namespace compute resources [GH-829]
* Add `types.TmNamespaceSummaryZonalResourceAllocation` structure with `Zone`, `ResourceAllocation`,
  `ErrorMessage`, and `RealizedStatus` properties for namespace allocations within a zone [GH-829]
* Add `types.TmNamespaceVcReference` structure with `SupervisorUuid` and `VcUuid` properties for
  namespace vCenter and supervisor identifiers [GH-829]
* Add `types.TmNamespaceVmClassReservation` structure with `VirtualMachineClassName`, `Count`, and
  `UsedCount` properties for namespace VM-class reservations [GH-829]
* Add `types.TmRegionCapability` structure with `Name`, `Description`, `Value`, `Type`, `Category`,
  and `Settings` properties for region capabilities [GH-829]
* Add `types.TmRegionalAviSetting` structure with `Status`, `Active`, `ServiceEngineGroupMode`,
  `ServiceEngineQuota`, `ApplicationLimit`, and `ServiceEngineGroupRefs` properties for regional Avi
  settings [GH-829]
* Add `types.TmRegionalDistributedVlanConnection` structure with `Status`, `ID`,
  `DistributedVlanConnectionRef`, `IsDefault`, `VlanId`, `GatewayCidr`, and
  `IpSpaceRef` properties for regional distributed VLAN connections [GH-829]
* Add `types.TmRegionalDistributedVxlanConnection` structure with `Status`, `ID`,
  `DistributedVxlanConnectionRef`, `IsDefault`, `L3Vni`, and `RouteDistinguisher` properties for
  regional distributed VXLAN connections [GH-829]
* Add `types.TmRegionalProviderGatewayConnection` structure with `Status`, `ID`, `Name`,
  `ProviderGatewayRef`, `IsDefault`, and `AllowAdvertisingPrivateIpBlocks` properties for regional
  provider-gateway connections [GH-829]
* Add `types.TmRegionalSharedSubnet` structure with `Status`, `ID`, `SubnetRef`, `Name`, and
  `VlanId` properties for regional shared-subnet associations [GH-829]
* Add `types.TmReservedVmClass` structure with `Region`, `Zone`, `VirtualMachineClass`,
  `TotalCount`, `AvailableCount`, and `GroupNumbers` properties for reserved VM-class inventory
  [GH-829]
* Add `types.TmVdcInfraPolicies` structure with `ResultTotal`, `PageCount`, `Page`, `PageSize`, and
  `Values` properties for region-quota infrastructure-policy assignment pages [GH-829]
* Add `types.TmVdcInfraPolicy` structure with `InfraPolicy` and `Status` properties for region-quota
  infrastructure-policy assignments [GH-829]
* Add `types.TmVdcReservedVmClass` structure with `Region`, `Zone`, `Name`, `TotalCount`, and
  `AvailableCount` properties for region-quota VM-class reservations [GH-829]
* Add `types.TmVdcReservedVmClasses` structure with `ResultTotal`, `PageCount`, `Page`, `PageSize`,
  and `Values` properties for region-quota VM-class reservation pages [GH-829]
* Add `types.TmVmClassReservation` structure with `VirtualMachineClassName`, `MaximumCount`, and
  `AvailableCount` properties for supervisor-zone VM-class reservation capacity [GH-829]
* Add `IsRestricted` property to `types.StorageClass` [GH-829]
* Add `NetworkStack` property to `types.Zone` [GH-829]
* Add `Type` and `OrgRef` properties to `types.TmIpSpace` [GH-829]
* Add `SubProviderOrgRef` property to `types.TmProviderGateway` [GH-829]
* Add `IPSpaceType` and `DistributedVxlanConnectionRef` properties to `types.TmIpSpaceAssociation` [GH-829]
* Add `DistributedVxlanConnectionRef` property to `types.TmRegionalNetworkingSetting` [GH-829]
* Remove the `IsUnlimitedAllocation` property from `types.StorageClass`, `types.VirtualDatacenterStoragePolicy`,
  `types.TmVdcResourceAllocation`, and `types.Zone`, the `VcpuSpeedMHz` property from `types.TmVdcResourceAllocation`
  and `types.Zone`, and the `SupportsUnlimitedAllocation` property from `types.Region`, as they are not part of the
  server API schema [GH-829]
