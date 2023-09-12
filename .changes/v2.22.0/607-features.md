* Added `Firmware` field to `VmSpecSection` type and `BootOptions` to `Vm` type [GH-607]
* Added `Vdc` methods `GetHardwareVersion`, `GetHighestHardwareVersion`, 
`FindOsFromId` [GH-607] 
* Added `VM` methods `UpdateBootOptions`, `UpdateBootOptionsAsync` [GH-607]
* API calls for `AddRawVM`, `CreateStandaloneVmAsync`, `VM.Refresh`, 
`VM.UpdateVmSpecSectionAsync`, `addEmptyVmAsyncV10`, `getVMByHrefV10` 
and `UpdateBootOptionsAsync` get elevated to API version `37.1` if available, for `VmSpecSection.Firmware` and `BootOptions` support [GH-607]
