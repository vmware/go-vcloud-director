* Added support for enhanced vApp [GH-394]
    * Added `types.ComposeVAppParamsV2`, `types.ReComposeVAppParamsV2`
    * Added type `VAppV2`
    * Added methods `Vdc.ComposeVAppV2`, `VAppV2.RecomposeVAppV2`, `VAppV2.Refresh`, `VAppV2.Undeploy`, `VAppV2.Delete`
    * Added methods`VAppV2.RemoveAllNetworks`, `VAppV2.PowerOn`, `Vdc.GetVappV2ByName`, `Vdc.GetVappV2ById`, `Vdc.GetVappV2ByNameOrId`
    * Added methods `Vdc.GetVappV2ByHref`, `Client.GetVappV2ByHref`
* Added support for VM migration between vApps [GH-394]
    * Added methods `VM.MoveToVapp` `VApp.TransferAllVms`
    * Added functions `MoveVmsToVapp`
* Moved MutexKV from Terraform provider [GH-394]
    * Added method `MutexKV.Lock`,  `MutexKV.Unlock` (MutexKv, imported from Hashicorp)
    * Added functions `NewMutexKV`, `NewMutexKVSilent`
* Added support for parallel, distributed VM operations [GH-394]
    * Added function `util.RunWhenReady` (parallel scheduler)
    * Added function `CreateParallelVm`
