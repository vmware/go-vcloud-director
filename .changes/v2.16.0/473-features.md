* Added metadata interface `MetadataCompatible` to manage uniformly all metadata-compatible resources [GH-473]
* Added `AdminCatalog.MergeMetadata`,`AdminCatalog.MergeMetadataAsync`, `AdminOrg.MergeMetadata`, `AdminOrg.MergeMetadataAsync`, 
`CatalogItem.MergeMetadata`, `CatalogItem.MergeMetadataAsync`, `Disk.MergeMetadata`, `Disk.MergeMetadataAsync`, `Media.MergeMetadata`, 
`Media.MergeMetadataAsync`, `MediaRecord.MergeMetadata`, `MediaRecord.MergeMetadataAsync`, `OpenAPIOrgVdcNetwork.MergeMetadata`, 
`OpenAPIOrgVdcNetwork.MergeMetadataAsync`, `OrgVDCNetwork.MergeMetadata`, `OrgVDCNetwork.MergeMetadataAsync`, 
`VApp.MergeMetadata`, `VApp.MergeMetadataAsync`, `VAppTemplate.MergeMetadata`, `VAppTemplate.MergeMetadataAsync`, 
`VM.MergeMetadata`, `VM.MergeMetadataAsync`, `Vdc.MergeMetadata`, `Vdc.MergeMetadataAsync` to merge metadata, 
which both updates existing metadata with same key and adds new entries for the non-existent ones [GH-473]
