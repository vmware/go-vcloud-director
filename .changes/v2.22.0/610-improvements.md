* New method `NsxtEdgeGateway.GetAllocatedIpCountByUplinkType` complementing existing
  `NsxtEdgeGateway.GetAllocatedIpCount`. It will return allocated IP counts by uplink types (works
  with VCD 10.4.1+) [GH-610]
* New method `NsxtEdgeGateway.GetPrimaryNetworkAllocatedIpCount` that will return total allocated IP
  count for primary uplink (T0 or T0 VRF) [GH-610]
* New field `types.EdgeGatewayUplinks.BackingType` that defines backing type of NSX-T Edge Gateway
  Uplink [GH-610]
* NSX-T Edge Gateway functions `GetNsxtEdgeGatewayById`, `GetNsxtEdgeGatewayByName`,
  `GetNsxtEdgeGatewayByNameAndOwnerId`, `GetAllNsxtEdgeGateways`, `CreateNsxtEdgeGateway`,
  `Refresh`, `Update` will additionally sort uplinks to ensure that element 0 contains primary
  network (T0 or T0 VRF) [GH-610]
