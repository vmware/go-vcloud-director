* Added the function `VCDClient.GetMetadataByKeyAndHref` to get a specific metadata value using an entity reference [GH-510]
* Added the functions `VCDClient.AddMetadataEntryWithVisibilityByHrefAsync`
  and `VCDClient.AddMetadataEntryWithVisibilityByHref` to add metadata with both visibility and domain
  to any entity by using its reference [GH-510]
* Added the functions `VCDClient.MergeMetadataWithVisibilityByHrefAsync`
  and `VCDClient.MergeMetadataWithVisibilityByHref` to merge metadata data supporting also visibility and domain [GH-510]
* Added the functions `VCDClient.DeleteMetadataEntryWithDomainByHrefAsync`
  and `VCDClient.DeleteMetadataEntryWithDomainByHref` to delete metadata from an entity using its reference [GH-510]
* Added the function `GetMetadataByKey` to the following entities:
  `VM`, `Vdc`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `Catalog`, `AdminCatalog`,
  `Org`, `AdminOrg`, `Disk`, `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to get a specific metadata value [GH-510]
* Added the functions `AddMetadataEntryWithVisibilityAsync` and `AddMetadataEntryWithVisibility` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to add metadata with both visibility and domain to them [GH-510]
* Added the functions `MergeMetadataWithMetadataValuesAsync` and `MergeMetadataWithMetadataValues` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to merge metadata data supporting also visibility and domain [GH-510]
* Added the functions `DeleteMetadataEntryWithDomainAsync` and `DeleteMetadataEntryWithDomain` to the following entities:
  `VM`, `AdminVdc`, `ProviderVdc`, `VApp`, `VAppTemplate`, `MediaRecord`, `Media`, `AdminCatalog`, `AdminOrg`, `Disk`,
  `OrgVDCNetwork`, `CatalogItem`, `OpenApiOrgVdcNetwork` to delete metadata with the domain, that allows deleting metadata present in SYSTEM [GH-510]
