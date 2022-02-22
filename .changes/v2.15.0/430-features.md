* Added types `types.MetadataStringValue`, `types.MetadataNumberValue`, `types.MetadataDateTimeValue` and `types.MetadataBooleanValue` 
  for adding different kind of metadata to entities [GH-430]
* Added support to set, get and delete metadata to AdminCatalog with the methods 
  `AdminCatalog.AddMetadataEntry`, `AdminCatalog.AddMetadataEntryAsync`, `AdminCatalog.GetMetadata`, 
  `AdminCatalog.DeleteMetadataEntry` and `AdminCatalog.DeleteMetadataEntryAsync`. [GH-430]
* Added support to get metadata from Catalog with method `Catalog.GetMetadata` [GH-430]
* Added to *VM* and *VApp* the methods `DeleteMetadataEntry`, `DeleteMetadataEntryAsync`, `AddMetadataEntry` and `AddMetadataEntryAsync`
  so it follows the same convention as the rest of entities that uses metadata. [GH-430]
