## 2.2.0 (Unreleased)

FEATURES:

* Added metadata add/remove functions to VM.
* Added ability to do vCD version checks and comparison [#174](https://github.com/vmware/go-vcloud-director/pull/174)
using VCDClient.APIVCDMaxVersionIs(string) and VCDClient.APIClientVersionIs(string).
* Added ability to override currently used vCD API version WithAPIVersion(string) [#174](https://github.com/vmware/go-vcloud-director/pull/174).
* Added ability to enable nested hypervisor option for VM with VM.ToggleNestedHypervisor(bool) [#219](https://github.com/terraform-providers/terraform-provider-vcd/issues/219).


BREAKING CHANGES:

* vApp metadata now is attached to the vApp rather to first VM in vApp.
* vApp metadata is no longer added to first VM in vApp it will be added to vApp directly instead.

IMPROVEMENTS:
* Refactored code by introducing helper function to handle API calls. New functions ExecuteRequest, ExecuteTaskRequest, ExecuteRequestWithoutResponse
* Tests files are now all tagged. Running them through Makefile works as before, but manual execution requires specific tags. Run `go test -v .` for tags list.

## 2.1.0 (March 21, 2019)

ARCHITECTURAL:

* Project switched to using Go modules. It is worth having a
look at [README.md](README.md) to understand how Go modules impact build and development.

FEATURES:

* New insert and eject media functions

IMPROVEMENTS:

* vApp vapp.PowerOn() implicitly waits for vApp to exit "UNRESOLVED" state which occurs shortly after creation and causes vapp.PowerOn() failure.
* VM has new functions which allows to configure cores for CPU. VM.ChangeCPUCountWithCore()

BREAKING CHANGES:

* Deprecate vApp.ChangeCPUCountWithCore() and vApp.ChangeCPUCount()
