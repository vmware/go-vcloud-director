* New types package `ccitypes` for CCI related types [GH-749]
* New method to get `CciClient` - `VCDClient.GetCciClient` [GH-749]
* `CciClient`  low level API interaction methods `IsSupported`, `GetCciUrl`, `PostItemAsync`, `PostItemSync`, `GetItem`, `DeleteItem` [GH-749]
* `CciClient.GetKubeConfig` method for retrieving KubeConfig [GH-749]
* Types `SupervisorNamespace` and `ccitypes.SupervisorNamespace` with methods
  `CciClient.CreateSupervisorNamespace`, `CciClient.GetSupervisorNamespaceByName` and
  `SupervisorNamespace.Delete` [GH-749]
