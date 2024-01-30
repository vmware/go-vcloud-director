module github.com/vmware/go-vcloud-director/v2

go 1.19

require (
	github.com/araddon/dateparse v0.0.0-20190622164848-0fb0a474d195
	github.com/davecgh/go-spew v1.1.0
	github.com/hashicorp/go-version v1.2.0
	github.com/kr/pretty v0.2.1
	github.com/peterhellberg/link v1.1.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/howeyc/crc16 v0.0.0-20171223171357-2b2a61e366a6
	github.com/kr/text v0.1.0 // indirect
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
	golang.org/x/text v0.14.0
)

replace (
	gopkg.in/check.v1 => github.com/go-check/check v0.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v2 => github.com/go-yaml/yaml/v2 v2.2.2
)
