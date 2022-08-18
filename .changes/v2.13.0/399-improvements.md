* External network type ExternalNetworkV2 automatically elevates API version to maximum available out of 33.0, 35.0 and
  36.0, so that new functionality can be consumed. It uses a controlled version elevation mechanism to consume the newer
  features, but at the same time remain tested by not choosing the latest untested version blindly (more information in
  openapi_endpoints.go) [GH-399]
* Added new field BackingTypeValue in favor of deprecated BackingType to types.ExternalNetworkV2Backing [GH-399]
* Add new function `GetFilteredNsxtImportableSwitches` to query NSX-T Importable Switches (Segments) [GH-399] 
