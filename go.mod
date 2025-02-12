module github.com/vmware/go-vcloud-director/v3

go 1.23.0

toolchain go1.23.4

require (
	github.com/araddon/dateparse v0.0.0-20190622164848-0fb0a474d195
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.36.0
	github.com/kr/pretty v0.3.1
	github.com/peterhellberg/link v1.1.0
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
	golang.org/x/text v0.21.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.32.1
	k8s.io/client-go v0.32.1
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.32.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
)

replace (
	gopkg.in/check.v1 => github.com/go-check/check v0.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v2 => github.com/go-yaml/yaml/v2 v2.4.0
)
