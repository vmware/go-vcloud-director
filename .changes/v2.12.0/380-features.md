* Added Tenant Context management functions `Client.RemoveCustomHeader`, `Client.SetCustomHeader`, `WithHttpHeader`, 
  and many private methods to retrieve tenant context down the hierarchy. More details in `CODING_GUIDELINES.md`
  [GH-380]
* Added Rights management methods `AdminOrg.GetAllRights`, `AdminOrg.GetAllRightsCategories`, `AdminOrg.GetRightById`, 
  `AdminOrg.GetRightByName`, `Client.GetAllRights`, `Client.GetAllRightsCategories`, `Client.GetRightById`, 
  `Client.GetRightByName`, `client.GetRightsCategoryById`, `AdminOrg.GetRightsCategoryById` [GH-380]
* Added Global Role management methods `Client.GetAllGlobalRoles`, `Client.CreateGlobalRole`, `Client.GetGlobalRoleById`,
  `Client.GetGlobalRoleByName`, `GlobalRole.AddRights`, `GlobalRole.Delete`, `GlobalRole.GetRights`,
  `GlobalRole.GetTenants`, `GlobalRole.PublishAllTenants`, `GlobalRole.PublishTenants`, `GlobalRole.RemoveAllRights`,
  `GlobalRole.RemoveRights`, `GlobalRole.ReplacePublishedTenants`, `GlobalRole.UnpublishAllTenants`, 
  `GlobalRole.UnpublishTenants`, `GlobalRole.Update`, `GlobalRole.UpdateRights` [GH-380]
* Added Rights Bundle management methods `Client.CreateRightsBundle`, `Client.GetAllRightsBundles`, 
  `Client.GetRightsBundleById`, `Client.GetRightsBundleByName`, `RightsBundle.AddRights`, `RightsBundle.Delete`, 
  `RightsBundle.GetRights`, `RightsBundle.GetTenants`, `RightsBundle.PublishAllTenants`, `RightsBundle.PublishTenants`, 
  `RightsBundle.RemoveAllRights`, `RightsBundle.RemoveRights`, `RightsBundle.ReplacePublishedTenants`, 
  `RightsBundle.UnpublishAllTenants`, `RightsBundle.UnpublishTenants`, `RightsBundle.Update`, `RightsBundle.UpdateRights`
  [GH-380]
* Added Role managemnt methods `AdminOrg.GetAllRoles`, `AdminOrg.GetRoleById`, `AdminOrg.GetRoleByName`, 
  `Client.GetAllRoles`, `Role.AddRights`, `Role.GetRights`, `Role.RemoveAllRights`, `Role.RemoveRights`, `Role.UpdateRights`
  [GH-380]
* Added convenience function `FindMissingImpliedRights` [GH-380]

