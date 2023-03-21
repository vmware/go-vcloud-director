//go:build api || functional || ALL

/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/json"
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetSessionInfo(check *C) {
	info, err := vcd.client.Client.GetSessionInfo()
	check.Assert(err, IsNil)
	check.Assert(info, NotNil)

	if testVerbose {
		var data []byte
		data, err = json.MarshalIndent(info, " ", " ")
		check.Assert(err, IsNil)
		fmt.Printf("%s\n", data)
		org, err := vcd.client.GetAdminOrgById(info.Org.ID)
		check.Assert(err, IsNil)
		for _, roleRef := range info.RoleRefs {
			role, err := org.GetRoleById(roleRef.ID)
			check.Assert(err, IsNil)
			fmt.Printf("%s\n", role.Role.Name)
			rights, err := role.GetRights(nil)
			check.Assert(err, IsNil)
			for i, right := range rights {
				fmt.Printf("\t%3d %s\n", i, right.Name)
			}
		}
	}
	check.Assert(info.Org, NotNil)
	check.Assert(info.Org.Name, Not(Equals), "")
	if vcd.client.Client.IsSysAdmin {
		check.Assert(info.Org.Name, Equals, "System")
	}
	check.Assert(info.User, Not(Equals), "")
	check.Assert(len(info.Roles), Not(Equals), 0)
}

func (vcd *TestVCD) Test_GetExtendedSessionInfo(check *C) {
	info, err := vcd.client.GetExtendedSessionInfo()
	check.Assert(err, IsNil)
	check.Assert(info, NotNil)
	if testVerbose {
		text, err := json.MarshalIndent(info, " ", " ")
		check.Assert(err, IsNil)
		fmt.Printf("%s\n", text)
	}
}
