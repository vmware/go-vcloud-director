# go-vcloud-director [![Build Status](https://travis-ci.org/vmware/go-vcloud-director.svg?branch=master)](https://travis-ci.org/vmware/go-vcloud-director) [![Coverage Status](https://coveralls.io/repos/vmware/go-vcloud-director/badge.svg?branch=master&service=github)](https://coveralls.io/github/vmware/go-vcloud-director?branch=master) [![GoDoc](https://godoc.org/github.com/vmware/go-vcloud-director?status.svg)](http://godoc.org/github.com/vmware/go-vcloud-director) [![Chat](https://img.shields.io/badge/chat-on%20slack-brightgreen.svg)](https://vmwarecode.slack.com/messages/CBBBXVB16)

This repo contains the `go-vcloud-director` package which implements
an SDK for vCloud Director. The project serves the needs of Golang
developers who need to integrate with vCloud Director. It is also the
basis of the [vCD Terraform
Provider](https://github.com/terraform-providers/terraform-provider-vcd).

## Contributions ##

Contributions to `go-vcloud-director` are gladly welcome and range
from participating in community discussions to submitting pull
requests.  Please see the [contributing guide](CONTRIBUTING.md) for
details on joining the community.

### Install and Build ###

Create a standard Golang development tree with bin, pkg, and src directories. 
Set GOPATH to the root directory. Then:
```
go get github.com/vmware/go-vcloud-director
cd $GOPATH/src/github.com/vmware/go-vcloud-director/govcd
go build
```
This command only builds a library. There is no executable.

### Example ###

To show the SDK in action run the example shown below.  
```
mkdir $GOPATH/src/example
cd $GOPATH/src/example
./example user_name "password" org_name vcd_IP vdc_name 
```
Here's the code:
```go
package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/go-vcloud-director/govcd"
)

type Config struct {
	User     string
	Password string
	Org      string
	Href     string
	VDC      string
	Insecure bool
}

func (c *Config) Client() (*govcd.VCDClient, error) {
	u, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("unable to pass url: %s", err)
	}

	vcdclient := govcd.NewVCDClient(*u, c.Insecure)
	err = vcdclient.Authenticate(c.User, c.Password, c.Org)
	if err != nil {
		return nil, fmt.Errorf("unable to authenticate: %s", err)
	}
	return vcdclient, nil
}

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Syntax: example user password org VCD_IP VDC ")
		os.Exit(1)
	}
	config := Config{
		User:     os.Args[1],
		Password: os.Args[2],
		Org:      os.Args[3],
		Href:     fmt.Sprintf("https://%s/api", os.Args[4]),
		VDC:      os.Args[5],
		Insecure: true,
	}

	client, err := config.Client() // We now have a client
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	org, err := govcd.GetOrgByName(client, config.Org)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	vdc, err := org.GetVdcByName(config.VDC)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Org URL: %s\n", org.Org.HREF)
	fmt.Printf("VDC URL: %s\n", vdc.Vdc.HREF)
}

```
