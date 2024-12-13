* Added `TmProviderGateway` and `types.TmProviderGateway` structures and methods
  `VCDClient.CreateTmProviderGateway`, `VCDClient.GetAllTmProviderGateways`,
  `VCDClient.GetTmProviderGatewayByName`, `VCDClient.GetTmProviderGatewayById`,
  `VCDClient.GetTmProviderGatewayByNameAndRegionId`, `TmProviderGateway.Update`,
  `TmProviderGateway.Delete` to manage Provider Gateways [GH-725]
* Added `TmTier0Gateway` and `types.TmTier0Gateway` structures consumption and methods
  `VCDClient.GetAllTmTier0GatewaysWithContext`, `VCDClient.GetTmTier0GatewayWithContextByName` to
  read Tier 0 Gateways that are available for TM consumption [GH-725]
* Added `TmIpSpaceAssociation` and `types.TmIpSpaceAssociation` structures and methods
  `VCDClient.CreateTmIpSpaceAssociation`, `VCDClient.GetAllTmIpSpaceAssociations`,
  `VCDClient.GetTmIpSpaceAssociationById`,
  `VCDClient.GetAllTmIpSpaceAssociationsByProviderGatewayId`,
  `VCDClient.GetAllTmIpSpaceAssociationsByIpSpaceId`, `TmIpSpaceAssociation.Delete` to manage IP
  Space associations with Provider Gateways [GH-725]
