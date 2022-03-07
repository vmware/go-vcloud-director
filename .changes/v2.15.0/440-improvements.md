* NSX-T Edge Gateway now supports VDC Groups by switching from `OrgVdc` to `OwnerRef` field.
  Additional methods `NsxtEdgeGateway.MoveToVdc()`, `Org.GetNsxtEdgeGatewayByNameAndOwnerId()`,
  `VdcGroup.GetNsxtEdgeGatewayByName()`, `VdcGroup.GetAllNsxtEdgeGateways()`, `org.GetVdcGroupById`
  [GH-440]
* Additional helper functions `OwnerIsVdcGroup()`, `OwnerIsVdc()`, `VdcGroup.GetCapabilities()`,
  `VdcGroup.IsNsxt()` [GH-440]