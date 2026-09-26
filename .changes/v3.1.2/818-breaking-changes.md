* Fields `ContextEntityID` and `NetworkProviderScope` of `types.NsxtNetworkContextProfile` changed
  from `interface{}` to `string`, and field `SubAttributes` of
  `types.NsxtNetworkContextProfileAttributes` changed from `interface{}` to
  `[]types.NsxtNetworkContextProfileSubAttribute` so that Network Context Profile create and
  update payloads can be built with typed values [GH-818]
