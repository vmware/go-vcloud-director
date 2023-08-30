* New method `NsxtEdgeGateway.GetAllocatedIpCountByUplinkType` complementing existing
  `NsxtEdgeGateway.GetAllocatedIpCount`. It will return allocated IP counts by uplink types [GH-610]
* New field `types.EdgeGatewayUplinks.BackingType` that defines backing type of NSX-T Edge Gateway
  Uplink [GH-610]
* NSX-T Edge Gateway functions `GetNsxtEdgeGatewayById`, `GetNsxtEdgeGatewayByName`,
  `GetNsxtEdgeGatewayByNameAndOwnerId`, `GetAllNsxtEdgeGateways`, `CreateNsxtEdgeGateway`,
  `Refresh`, `Update` will additionally sort uplinks to ensure that element 0 contains primary
  network (T0 or T0 VRF) [GH-610]
