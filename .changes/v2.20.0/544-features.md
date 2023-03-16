* Added support for Runtime Defined Entity instances with methods `DefinedEntityType.GetAllRdes`, `DefinedEntityType.GetRdeByName`,
  `DefinedEntityType.GetRdeById`, `DefinedEntityType.CreateRde` and methods to manipulate them `DefinedEntity.Resolve`,
  `DefinedEntity.Update`, `DefinedEntity.Delete` [GH-544]
* Add generic `Client` methods `OpenApiPostItemAndGetHeaders` and `OpenApiGetItemAndHeaders` to be able to retrieve the
  response headers when performing a POST or GET operation to an OpenAPI endpoint [GH-544]
