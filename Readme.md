## vmware-govcd [![Build Status](https://travis-ci.org/vmware/govcloudair.svg?branch=master)](https://travis-ci.org/frapposelli/govcloudair) [![Coverage Status](https://img.shields.io/coveralls/vmware/govcloudair.svg)](https://coveralls.io/r/vmware/govcloudair) [![GoDoc](https://godoc.org/github.com/vmware/govcloudair?status.svg)](http://godoc.org/github.com/vmware/govcloudair)

This package was originally forked from [github.com/vmware/govcloudair](https://github.com/vmware/govcloudair) before pulling in [rickard-von-essen's](https://github.com/rickard-von-essen)
great changes to allow using a [vCloud Director API](https://github.com/rickard-von-essen/govcloudair/tree/vcd-5.5). On top of this I have added features as needed for a terraform provider for vCloud Director

### Example ###

```go
package main

import (
	"fmt"
	"net/url"

	"github.com/opencredo/vmware-govcd"
)

type Config struct {
	User     string
	Password string
	Org      string
	Href     string
	VDC      string
}

func (c *Config) Client() (*govcd.VCDClient, error) {
	u, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("Unable to pass url: %s", err)
	}

	vcdclient := govcd.NewVCDClient(*u)
	org, vcd, err := vcdclient.Authenticate(c.User, c.Password, c.Org, c.VDC)
	if err != nil {
		return nil, fmt.Errorf("Unable to authenticate: %s", err)
	}
	vcdclient.Org = org
	vcdclient.OrgVdc = vcd
	return vcdclient, nil
}

func main() {
  config := Config{
		User:     "Username",
		Password: "password",
		Org:      "vcd org",
		Href:     "vcd api url",
		VDC:      "vcd virtual datacenter name",
	}

  client, _ := config.Client() // We now have a client
  fmt.Printf("Session URL: %s", client.sessionHREF)
}
```
