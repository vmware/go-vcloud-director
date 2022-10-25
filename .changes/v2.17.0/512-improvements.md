* Added VCDClient.GetAllNsxtEdgeClusters for lookup of NSX-T Edge Clusters in wider scopes -
  Provider VDC, VDC Group or VDC [GH-512]
* Switch VDC.GetAllNsxtEdgeClusters to use 'orgVdcId' filter instead of '_context' (now deprecated)
  [GH-512]
* Removed a few log lines from SetLog() function which were being set before correct logging file
  was initialized. This resulted in a file `go-vcloud-director.log` even if other filename was used
  [GH-512]
