* Ran `make fmt` using Go 1.19 release (`fmt` automatically changes doc comment structure). This
  will prevent `make static` errors when running tests in pipeline using Go 1.19 [GH-497]
* Updated branding `vCloud Director` -> `VMware Cloud Director` [GH-497]
* Go officially supports 2 last releases. With Go 1.19 being released it means that Go 1.18 is the
  minimum officially supported Go version and this set our hands free to use generics in this SDK
  (if there is a need for it). `go.mod` is updated to reflect Go minimum version 1.18 [GH-497]
* package `io/ioutil` is deprecated as of Go 1.16. `staticcheck` started complaining about usage of
  deprecated packages. As a result this PR switches packages to either `io` or `os` (still the same
  functions are used) [GH-497]
* Adjusted `staticcheck` version naming to new format (from `2021.1.2` to `v0.3.3`) [GH-497]
