//go:build functional || openapi || rde || ALL
// +build functional openapi rde ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

// Test_DefinedInterface tests the CRUD behavior of Defined Interfaces. First, it checks how many exist already,
// as VCD contains some pre-defined ones. Then we create a new one, so the number of existing ones should be increased by 1.
// We try to get this new created interface with the available getter methods and then perform an update.
// As a final step, we delete it and check that the deletion is correct (the receiver object is empty and doesn't exist in VCD).
func (vcd *TestVCD) Test_DefinedInterface(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointInterfaces
	skipOpenApiEndpointTest(vcd, check, endpoint)

	allDefinedInterfaces, err := vcd.client.GetAllDefinedInterfaces(nil)
	check.Assert(err, IsNil)
	alreadyPresentRDEs := len(allDefinedInterfaces)

	dummyRde := &types.DefinedInterface{
		Name:      check.TestName() + "_name",
		Namespace: check.TestName() + "_nss",
		Version:   "1.2.3",
		Vendor:    "vmware",
	}
	newDefinedInterface, err := vcd.client.CreateDefinedInterface(dummyRde)
	check.Assert(err, IsNil)
	check.Assert(newDefinedInterface, NotNil)
	check.Assert(newDefinedInterface.DefinedInterface.Name, Equals, dummyRde.Name)
	check.Assert(newDefinedInterface.DefinedInterface.Namespace, Equals, dummyRde.Namespace)
	check.Assert(newDefinedInterface.DefinedInterface.Version, Equals, dummyRde.Version)
	check.Assert(newDefinedInterface.DefinedInterface.Vendor, Equals, dummyRde.Vendor)
	check.Assert(newDefinedInterface.DefinedInterface.IsReadOnly, Equals, dummyRde.IsReadOnly)

	AddToCleanupListOpenApi(newDefinedInterface.DefinedInterface.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointInterfaces+newDefinedInterface.DefinedInterface.ID)

	allDefinedInterfaces, err = vcd.client.GetAllDefinedInterfaces(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allDefinedInterfaces), Equals, alreadyPresentRDEs+1)

	obtainedDefinedInterface, err := vcd.client.GetDefinedInterfaceById(newDefinedInterface.DefinedInterface.ID)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface.DefinedInterface, DeepEquals, *newDefinedInterface.DefinedInterface)

	obtainedDefinedInterface2, err := vcd.client.GetDefinedInterface(obtainedDefinedInterface.DefinedInterface.Vendor, obtainedDefinedInterface.DefinedInterface.Namespace, obtainedDefinedInterface.DefinedInterface.Version)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface2.DefinedInterface, DeepEquals, *obtainedDefinedInterface.DefinedInterface)

	err = newDefinedInterface.Update(types.DefinedInterface{
		Name: dummyRde.Name + "2", // Only name can be updated
	})
	check.Assert(err, IsNil)
	check.Assert(newDefinedInterface.DefinedInterface.Name, Equals, dummyRde.Name+"2")

	deletedId := newDefinedInterface.DefinedInterface.ID
	err = newDefinedInterface.Delete()
	check.Assert(err, IsNil)
	check.Assert(*newDefinedInterface.DefinedInterface, DeepEquals, types.DefinedInterface{})

	_, err = vcd.client.GetDefinedInterfaceById(deletedId)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)
}
