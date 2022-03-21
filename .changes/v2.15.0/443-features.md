* Added new structure `AnyEdgeGateway` which supports retreving both types of Edge Gateways (NSX-V
  and NSX-T) with methods `AdminOrg.GetAnyEdgeGatewayById`, `Org.GetAnyEdgeGatewayById`,
  `AnyEdgeGateway.IsNsxt`, `AnyEdgeGateway.IsNsxv`, `AnyEdgeGateway.GetNsxtEdgeGateway` [GH-443]
* Added functions `VdcGroup.GetCapabilities`, `VdcGroup.IsNsxt`,
  `VdcGroup.GetOpenApiOrgVdcNetworkByName`, `VdcGroup.GetAllOpenApiOrgVdcNetworks`,
  `Org.GetOpenApiOrgVdcNetworkByNameAndOwnerId` [GH-443]

