* NSX-T Edge Gateway now supports VDC Groups by switching from `OrgVdc` to `OwnerRef` field.
  Additional methods `NsxtEdgeGateway.MoveToVdcOrVdcGroup()`,
  `Org.GetNsxtEdgeGatewayByNameAndOwnerId()`, `VdcGroup.GetNsxtEdgeGatewayByName()`,
  `VdcGroup.GetAllNsxtEdgeGateways()`, `org.GetVdcGroupById` [GH-440]
* Additional helper functions `OwnerIsVdcGroup()`, `OwnerIsVdc()`, `VdcGroup.GetCapabilities()`,
  `VdcGroup.IsNsxt()` [GH-440]