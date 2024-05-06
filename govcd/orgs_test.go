//go:build org || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetOrgs(check *C) {

	orgs, err := vcd.client.GetAllOrgs(nil, true)
	check.Assert(err, IsNil)
	for i, org := range orgs {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(org.Org))
	}
	fmt.Println()
	orgs2, err := vcd.client.GetAllOrgs(nil, false)
	check.Assert(err, IsNil)
	for i, org := range orgs2 {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(org.Org))
	}
}
