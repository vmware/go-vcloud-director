* Added `NsxtEdgeGateway.Refresh` function to reload NSX-T Edge Gateway structure [GH-532]
* Added `NsxtEdgeGateway.GetUsedIpAddresses` function to fetch used IP addresses in NSX-T Edge
  Gateway [GH-532]
* Added `NsxtEdgeGateway.GetUsedIpAddressSlice` function to fetch used IP addresses in a slice
  [GH-532]
* Added `NsxtEdgeGateway.GetUnusedExternalIPAddresses` function that can help to find an unused
  IP address in an Edge Gateway by given constraints [GH-532]
* Added `NsxtEdgeGateway.GetAllUnusedExternalIPAddresses` function that can return all unused IP
  addresses in an Edge Gateway [GH-532]
* Added `NsxtEdgeGateway.GetAllocatedIpCount` function that sums up `TotalIPCount` fields in all
  subnets [GH-532]