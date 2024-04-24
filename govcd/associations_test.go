//go:build system || functional || ALL

/*
 * Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetSiteAssociations(check *C) {

	if !vcd.client.Client.IsSysAdmin {
		check.Skip(fmt.Sprintf("test %s requires system administrator privileges\n", check.TestName()))
	}
	siteQueryAssociations, err := vcd.client.Client.QueryAllSiteAssociations(nil, nil)
	check.Assert(err, IsNil)
	for i, s := range siteQueryAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(s))
	}
	fmt.Println()
	orgQueryAssociations, err := vcd.client.Client.QueryAllOrgAssociations(nil, nil)
	check.Assert(err, IsNil)
	for i, s := range orgQueryAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(s))
	}

	fmt.Println()
	associationData, err := vcd.client.Client.GetSiteAssociationData()
	check.Assert(err, IsNil)
	fmt.Printf("---- %# v\n", pretty.Formatter(associationData))

	fmt.Println()
	siteAssociations, err := vcd.client.Client.GetSiteAssociations()
	check.Assert(err, IsNil)
	for i, a := range siteAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(a))
	}

	fmt.Println()
	org, err := vcd.client.GetAdminOrgByName("gmaxia")
	check.Assert(err, IsNil)
	orgAssociations, err := org.GetOrgAssociations()
	check.Assert(err, IsNil)
	for i, s := range orgAssociations {
		fmt.Printf("%d %# v\n", i, pretty.Formatter(s))
	}

}
