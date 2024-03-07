* Added public method `VApp.GetParentVDC` to retrieve parent VDC of vApp (previously it was private)
  [GH-652]
* Added methods `Catalog.CaptureVappTemplate`, `Catalog.CaptureVappTemplateAsync` and type
  `types.CaptureVAppParams` that add support for creating catalog template from existing vApp
  [GH-652]
* Added method `Org.GetVAppByHref` to retrieve a vApp by given HREF [GH-652]
* Added methods `VAppTemplate.GetCatalogItemHref` and `VAppTemplate.GetCatalogItemId` that can return
  related catalog item ID and HREF [GH-652]
