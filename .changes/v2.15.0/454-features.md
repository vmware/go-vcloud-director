* Added support to set, get and delete metadata to the following resources via its HREF:
  `catalog`, `catalog item`, `edge gateway`, `independent disk`, `media`, `network`, `org`, `PVDC`, `PVDC storage profile`, `vApp`, `vApp template`,`VDC` and `VDC storage profile`;
  with the methods
  `VCDClient.GetMetadataByHref`, `VCDClient.AddMetadataEntryByHref`, `VCDClient.AddMetadataEntryByHrefAsync`,
  `VCDClient.DeleteMetadataEntryByHref` and `VCDClient.DeleteMetadataEntryByHrefAsync` [GH-454]
