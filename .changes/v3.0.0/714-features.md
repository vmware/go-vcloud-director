* Add trusted certificate management types `TrustedCertificate` and `types.TrustedCertificate`
  together with `VCDClient.CreateTrustedCertificate`, `VCDClient.GetAllTrustedCertificates`,
  `GetTrustedCertificateByName`, `VCDClient.GetTrustedCertificateById`, `TrustedCertificate.Update`,
  `TrustedCertificate.Delete` [GH-714]
* vCenter management types `VCenter` and `types.VSphereVirtualCenter` adds Create, Update and Delete
 methods: `VCDClient.CreateVcenter`, `VCDClient.GetAllVCenters`, `VCDClient.GetVCenterByName`,
 `VCDClient.GetVCenterById`, `VCenter.Update`, `VCenter.Delete` [GH-714]
* Add NSX-T Manager management types `TmNsxtManager`, `types.TmNsxtManager` and methods
  `VCDClient.CreateTmNsxtManager`, `VCDClient.GetAllTmNsxtManagers`,
  `VCDClient.GetTmNsxtManagerById`, `VCDClient.GetTmNsxtManagerByName`, `TmNsxtManager.Update`,
  `TmNsxtManager.Delete` [GH-714]
