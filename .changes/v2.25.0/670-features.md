* Added types `SolutionAddOn`, `SolutionAddOnConfig` and `types.SolutionAddOn` for Solution Add-on
  Landing configuration [GH-670]
* Added `VDCClient` methods `CreateSolutionAddOn`, `GetAllSolutionAddons`, `GetSolutionAddonById`,
  `GetSolutionAddonByName` for handling Solution Add-Ons [GH-670]
* Added  `SolutionAddOn` methods `Update`, `RdeId`, `Delete` to help handling of Solution Landing
  Zones [GH-670]
* Added `VDCClient` method `TrustAddOnImageCertificate` to trust certificate if it is not yet
  trusted [GH-670]