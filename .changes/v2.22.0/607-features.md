* Added `firmware` field to `VmSpecSection` type, `BootOptions` to `Vm` type,
New functions `*Vdc.GetHardwareVersion`, `*Vdc.GetHighestHardwareVersion`, 
`*Vdc.FindOsFromId`, `*VM.UpdateBootOptions`, `*VM.UpdateBootOptionsAsync` [GH-607]
* API calls for `AddRawVM`, `CreateStandaloneVmAsync`, `*VM.Refresh()`, 
`*VM.UpdateVmSpecSectionAsync`, `addEmptyVmAsyncV10`, `getVMByHrefV10` 
and `UpdateBootOptionsAsync` get elevated to API version `37.1` if available, for `firmware` and `BootOptions` support [GH-607]
