* Added `NsxtEdgeGateway.Refresh` method to reload NSX-T Edge Gateway structure [GH-532]
* Added `NsxtEdgeGateway.GetUsedIpAddresses` method to fetch used IP addresses in NSX-T Edge
  Gateway [GH-532]
* Added `NsxtEdgeGateway.GetUsedIpAddressSlice` method to fetch used IP addresses in a slice
  [GH-532]
* Added `NsxtEdgeGateway.GetUnusedExternalIPAddresses` method that can help to find an unused
  IP address in an Edge Gateway by given constraints [GH-532,GH-567]
* Added `NsxtEdgeGateway.GetAllUnusedExternalIPAddresses` method that can return all unused IP
  addresses in an Edge Gateway [GH-532,GH-567]
* Added `NsxtEdgeGateway.GetAllocatedIpCount` method that sums up `TotalIPCount` fields in all
  subnets [GH-532]
* Added `NsxtEdgeGateway.QuickDeallocateIpCount` and `NsxtEdgeGateway.DeallocateIpCount`
  methods to manually alter Edge Gateway body for IP deallocation [GH-532]
