* New method `NsxtEdgeGateway.ReorderUplinks()` that will ensure that NSX-T Tier0 Gateway backed
  uplink is at position 0 in the slice of uplinks. All NSX-T Edge Gateway retrieval functions will
  call it implicitly [GH-610]
* New method `NsxtEdgeGateway.GetAllocatedIpCountByUplinkType` complementing existing
  `NsxtEdgeGateway.GetAllocatedIpCount`. It will return allocated IP counts by uplink types [GH-610]
