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

// Test_DefinedInterface tests the CRUD behavior of Defined Interfaces as a System administrator and tenant user.
// This test can be run with GOVCD_SKIP_VAPP_CREATION option enabled.
func (vcd *TestVCD) Test_DefinedInterface(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointInterfaces)
	if len(vcd.config.Tenants) == 0 {
		check.Skip("skipping as there is no configured tenant users")
	}

	// Creates the clients for the System admin and the Tenant user
	systemAdministratorClient := vcd.client
	tenantUserClient := NewVCDClient(vcd.client.Client.VCDHREF, true)
	err := tenantUserClient.Authenticate(vcd.config.Tenants[0].User, vcd.config.Tenants[0].Password, vcd.config.Tenants[0].SysOrg)
	check.Assert(err, IsNil)

	// First, it checks how many exist already, as VCD contains some pre-defined ones.
	allDefinedInterfacesBySysAdmin, err := systemAdministratorClient.GetAllDefinedInterfaces(nil)
	check.Assert(err, IsNil)
	alreadyPresentRDEs := len(allDefinedInterfacesBySysAdmin)

	allDefinedInterfacesByTenant, err := tenantUserClient.GetAllDefinedInterfaces(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allDefinedInterfacesByTenant), Equals, len(allDefinedInterfacesBySysAdmin))

	// Then we create a new Defined Interface with System administrator
	dummyRde := &types.DefinedInterface{
		Name:      strings.ReplaceAll(check.TestName()+"name3", ".", ""),
		Namespace: strings.ReplaceAll(check.TestName()+"nss3", ".", ""),
		Version:   "1.2.3",
		Vendor:    "vmware",
	}
	newDefinedInterfaceFromSysAdmin, err := systemAdministratorClient.CreateDefinedInterface(dummyRde)
	check.Assert(err, IsNil)
	check.Assert(newDefinedInterfaceFromSysAdmin, NotNil)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Name, Equals, dummyRde.Name)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Namespace, Equals, dummyRde.Namespace)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Version, Equals, dummyRde.Version)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Vendor, Equals, dummyRde.Vendor)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.IsReadOnly, Equals, dummyRde.IsReadOnly)
	AddToCleanupListOpenApi(newDefinedInterfaceFromSysAdmin.DefinedInterface.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointInterfaces+newDefinedInterfaceFromSysAdmin.DefinedInterface.ID)

	// Tenants can't create Defined Interfaces
	nilDefinedInterface, err := tenantUserClient.CreateDefinedInterface(&types.DefinedInterface{
		Name:      strings.ReplaceAll(check.TestName()+"4", ".", ""),
		Namespace: strings.ReplaceAll(check.TestName()+"4", ".", ""),
		Version:   "4.5.6",
		Vendor:    "vmware",
	})
	check.Assert(err, NotNil)
	check.Assert(nilDefinedInterface, IsNil)
	check.Assert(strings.Contains(err.Error(), "ACCESS_TO_RESOURCE_IS_FORBIDDEN"), Equals, true)

	// As we created a new one, we check the new count is correct in both System admin and Tenant user
	allDefinedInterfacesBySysAdmin, err = systemAdministratorClient.GetAllDefinedInterfaces(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allDefinedInterfacesBySysAdmin), Equals, alreadyPresentRDEs+1)

	allDefinedInterfacesByTenant, err = tenantUserClient.GetAllDefinedInterfaces(nil)
	check.Assert(err, IsNil)
	check.Assert(len(allDefinedInterfacesByTenant), Equals, len(allDefinedInterfacesBySysAdmin))

	// Test the multiple ways of getting a Defined Interface in both users.
	obtainedDefinedInterface, err := systemAdministratorClient.GetDefinedInterfaceById(newDefinedInterfaceFromSysAdmin.DefinedInterface.ID)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface.DefinedInterface, DeepEquals, *newDefinedInterfaceFromSysAdmin.DefinedInterface)

	obtainedDefinedInterface, err = tenantUserClient.GetDefinedInterfaceById(newDefinedInterfaceFromSysAdmin.DefinedInterface.ID)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface.DefinedInterface, DeepEquals, *newDefinedInterfaceFromSysAdmin.DefinedInterface)

	obtainedDefinedInterface2, err := systemAdministratorClient.GetDefinedInterface(obtainedDefinedInterface.DefinedInterface.Vendor, obtainedDefinedInterface.DefinedInterface.Namespace, obtainedDefinedInterface.DefinedInterface.Version)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface2.DefinedInterface, DeepEquals, *obtainedDefinedInterface.DefinedInterface)

	obtainedDefinedInterface2, err = tenantUserClient.GetDefinedInterfaceById(newDefinedInterfaceFromSysAdmin.DefinedInterface.ID)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface2.DefinedInterface, DeepEquals, *obtainedDefinedInterface.DefinedInterface)

	// Update the Defined Interface as System administrator
	err = newDefinedInterfaceFromSysAdmin.Update(types.DefinedInterface{
		Name: dummyRde.Name + "2", // Only name can be updated
	})
	check.Assert(err, IsNil)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Name, Equals, dummyRde.Name+"2")

	// This one was obtained by the tenant, so it shouldn't be updatable
	err = obtainedDefinedInterface2.Update(types.DefinedInterface{
		Name: dummyRde.Name + "3",
	})
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "ACCESS_TO_RESOURCE_IS_FORBIDDEN"), Equals, true)

	// This one was obtained by the tenant, so it shouldn't be deletable
	err = obtainedDefinedInterface2.Delete()
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "ACCESS_TO_RESOURCE_IS_FORBIDDEN"), Equals, true)

	// We perform the actual removal with the System administrator
	deletedId := newDefinedInterfaceFromSysAdmin.DefinedInterface.ID
	err = newDefinedInterfaceFromSysAdmin.Delete()
	check.Assert(err, IsNil)
	check.Assert(*newDefinedInterfaceFromSysAdmin.DefinedInterface, DeepEquals, types.DefinedInterface{})

	_, err = systemAdministratorClient.GetDefinedInterfaceById(deletedId)
	check.Assert(err, NotNil)

	// API <= 36.0 returns a 400 BAD REQUEST instead of the usual 403 FORBIDDEN
	if vcd.client.Client.APIClientVersionIs("<= 36.0") {
		check.Assert(strings.Contains(err.Error(), "BAD_REQUEST"), Equals, true)
	} else {
		check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)
	}
}
