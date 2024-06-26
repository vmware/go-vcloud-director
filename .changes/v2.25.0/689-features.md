* Added types `DataSolution` and `types.DataSolution` for Data Storage Extension (DSE) management
  [GH-689]
* Added `DataSolution` methods `RdeId`, `Name`, `Update`, `Publish`, `Unpublish`,
  `PublishRightsBundle`, `UnpublishRightsBundle`, `PublishAccessControls`,
  `UnpublishAccessControls`, `GetAllAccessControls`, `GetAllAccessControlsForTenant`,
  `GetAllInstanceTemplates`, `PublishAllInstanceTemplates`, `UnPublishAllInstanceTemplates`,
  `GetAllDataSolutionOrgConfigs`, `GetDataSolutionOrgConfigForTenant` [GH-689]
* Added `VCDClient` methods `GetAllDataSolutions`, `GetDataSolutionById`, `GetDataSolutionByName`,
  `GetAllInstanceTemplates` [GH-689]
* Added types `DataSolutionInstanceTemplate` and `types.DataSolutionInstanceTemplate` for Data
  Storage Extension (DSE) Solution Instance Template management [GH-689]
* Added `DataSolutionInstanceTemplate` methods `Name`, `GetAllAccessControls`,
  `GetAllAccessControlsForTenant`, `Publish`, `Unpublish`, `RdeId` [GH-689]
* Added types `DataSolutionOrgConfig` and `types.DataSolutionOrgConfig` for Data Storage Extension
  (DSE) Solution Instance Org Configuration management [GH-689]
* Added `DataSolutionOrgConfig` methods `CreateDataSolutionOrgConfig`,
  `GetAllDataSolutionOrgConfigs`, `Delete`, `RdeId` [GH-689]
