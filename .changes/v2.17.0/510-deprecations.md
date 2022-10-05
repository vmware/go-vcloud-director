* Deprecated the functions `VCDClient.AddMetadataEntryByHrefAsync`
  and `VCDClient.AddMetadataEntryByHref` in favor of `VCDClient.AddMetadataEntryWithVisibilityByHrefAsync`
  and `VCDClient.AddMetadataEntryWithVisibilityByHref` [GH-510]
* Deprecated the functions `VCDClient.MergeMetadataByHrefAsync`
  and `VCDClient.MergeMetadataByHref` in favor of `VCDClient.MergeMetadataWithVisibilityByHrefAsync`
  and `VCDClient.MergeMetadataWithVisibilityByHref` [GH-510]
* Deprecated the functions `AddMetadataEntryAsync` and `AddMetadataEntry` from the following entities:
  `VM`, `Vdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem` in favor of their `AddMetadataEntryWithVisibilityAsync` and `AddMetadataEntryWithVisibility`
  counterparts [GH-510]
* Deprecated the functions `MergeMetadataAsync` and `MergeMetadataAsync` from the following entities:
  `VM`, `Vdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem` in favor of their `MergeMetadataWithMetadataValuesAsync` and `MergeMetadataWithMetadataValues`
  counterparts [GH-510]
* Deprecated the functions `Vdc.DeleteMetadataEntryAsync` and `Vdc.DeleteMetadataEntry` in favor of
  `AdminVdc.DeleteMetadataEntryAsync` and `AdminVdc.DeleteMetadataEntry` [GH-510]
