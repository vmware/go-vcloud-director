* Add trusted certificate management types `TrustedCertificate` and `types.TrustedCertificate`
  together with `VCDClient.CreateTrustedCertificate`, `VCDClient.GetAllTrustedCertificates`,
  `GetTrustedCertificateByName`, `VCDClient.GetTrustedCertificateById`, `TrustedCertificate.Update`,
  `TrustedCertificate.Delete` [GH-714]
* vCenter management types `VCenter` and `types.VSphereVirtualCenter` adds Create, Update and Delete
 methods: `VCDClient.CreateVcenter`, `VCDClient.GetAllVCenters`, `VCDClient.GetVCenterByName`,
 `VCDClient.GetVCenterById`, `VCenter.Update`, `VCenter.Delete`, `VCenter.Refresh` [GH-714]
* Add NSX-T Manager management types `NsxtManagerOpenApi`, `types.NsxtManagerOpenApi` and methods
  `VCDClient.CreateNsxtManagerOpenApi`, `VCDClient.GetAllNsxtManagersOpenApi`,
  `VCDClient.GetNsxtManagerOpenApiById`, `VCDClient.GetNsxtManagerOpenApiByName`,
  `TmNsxtManager.Update`, `TmNsxtManager.Delete` [GH-714]
