* Added methods to create, read, update and delete VDC Templates: `VCDClient.CreateVdcTemplate`, `VCDClient.GetVdcTemplateById`,
`VCDClient.GetVdcTemplateByName`, `VdcTemplate.Update` and `VdcTemplate.Delete` [GH-686]
* Added methods to manage the access settings of VDC Templates: `VdcTemplate.SetAccessControl` and  `VdcTemplate.GetAccessControl` [GH-686]
* Added the `VdcTemplate.InstantiateVdcAsync` and `VdcTemplate.InstantiateVdc` methods to instantiate VDC Templates [GH-686]
* Added the `VCDClient.QueryAdminVdcTemplates` and `Org.QueryVdcTemplates` methods to get all VDC Template records [GH-686]
