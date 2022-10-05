* Added the functions `VCDClient.AddMetadataEntryWithVisibilityByHrefAsync`
  and `VCDClient.AddMetadataEntryWithVisibilityByHref` to add metadata with both visibility and domain
  to any entity by using its reference [GH-510]
* Added the functions `VCDClient.MergeMetadataWithVisibilityByHrefAsync`
  and `VCDClient.MergeMetadataWithVisibilityByHref` to merge metadata data supporting also visibility and domain [GH-510]
* Added the functions `AddMetadataEntryWithVisibilityAsync` and `AddMetadataEntryWithVisibility` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem` to add metadata with both visibility and domain to them [GH-510]
* Added the functions `MergeMetadataWithMetadataValuesAsync` and `MergeMetadataWithMetadataValues` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem` to merge metadata data supporting also visibility and domain [GH-510]
* Added the functions `AdminVdc.DeleteMetadataEntryAsync` and `AdminVdc.DeleteMetadataEntryAsync` [GH-510]
