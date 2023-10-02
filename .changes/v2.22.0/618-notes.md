* Introduction of internal generic functions that are used to abstract away complexity of dealing
  with OpenAPI in the typed function. These functions are added: `genericCreateBareEntity`,
  `genericUpdateBareEntity`, `genericGetSingleBareEntity`, `genericGetAllBareFilteredEntities`,
  `genericLocalFilter`, `genericLocalFilterOneOrError`, `deleteById` [GH-618]
