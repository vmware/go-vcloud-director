* Added new structure `AnyTypeEdgeGateway` which supports retreving both types of Edge Gateways
  (NSX-V and NSX-T) with methods `AdminOrg.GetAnyTypeEdgeGatewayById`,
  `Org.GetAnyTypeEdgeGatewayById`, `AnyTypeEdgeGateway.IsNsxt`, `AnyTypeEdgeGateway.IsNsxv`,
  `AnyTypeEdgeGateway.GetNsxtEdgeGateway` [GH-443]
* Added functions `VdcGroup.GetCapabilities`, `VdcGroup.IsNsxt`,
  `VdcGroup.GetOpenApiOrgVdcNetworkByName`, `VdcGroup.GetAllOpenApiOrgVdcNetworks`,
  `Org.GetOpenApiOrgVdcNetworkByNameAndOwnerId` [GH-443]

