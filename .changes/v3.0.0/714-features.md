* Add trusted certificate management types `TrustedCertificate` and `types.TrustedCertificate`
  together with `VCDClient.AutoTrustCertificate`, `VCDClient.CreateTrustedCertificate`,
  `VCDClient.GetAllTrustedCertificates`, `GetTrustedCertificateByName`,
  `VCDClient.GetTrustedCertificateById`, `TrustedCertificate.Update`, `TrustedCertificate.Delete`
  [GH-714]
* vCenter management types `VCenter` and `types.VSphereVirtualCenter` adds Create, Update and Delete
 methods: `VCDClient.CreateVcenter`, `VCDClient.GetAllVCenters`, `VCDClient.GetVCenterByName`,
 `VCDClient.GetVCenterById`, `VCenter.Update`, `VCenter.Delete`, `VCenter.RefreshVcenter`,
 `VCenter.Refresh` [GH-714, GH-724, GH-747, GH-753, GH-756, GH-759]
* Add NSX-T Manager management types `NsxtManagerOpenApi`, `types.NsxtManagerOpenApi` and methods
  `VCDClient.CreateNsxtManagerOpenApi`, `VCDClient.GetAllNsxtManagersOpenApi`,
  `VCDClient.GetNsxtManagerOpenApiById`, `VCDClient.GetNsxtManagerOpenApiByName`,
  `TmNsxtManager.Update`, `TmNsxtManager.Delete` [GH-714, GH-722, GH-747]
* Added async vCenter creation function `VCDClient.CreateVcenterAsync` that exposes the creation
  task of as it is needed in some cases to retrieve ID of incomplete vCenter creation [GH-736]
