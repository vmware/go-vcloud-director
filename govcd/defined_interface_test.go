//go:build functional || openapi || rde || ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

// Test_DefinedInterface tests the CRUD behavior of Defined Interfaces as a System administrator and tenant user.
// This test can be run with GOVCD_SKIP_VAPP_CREATION option enabled.
func (vcd *TestVCD) Test_DefinedInterface(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeInterfaces)
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

	// Then we create a new Defined Interface with System administrator. We replace the dots in both
	// nss and name as API is quirky at versions of VCD <= 10.3.0
	dummyRde := &types.DefinedInterface{
		Name:    strings.ReplaceAll(check.TestName()+"name", ".", ""),
		Nss:     strings.ReplaceAll(check.TestName()+"nss", ".", ""),
		Version: "1.2.3",
		Vendor:  "vmware",
	}
	newDefinedInterfaceFromSysAdmin, err := systemAdministratorClient.CreateDefinedInterface(dummyRde)
	check.Assert(err, IsNil)
	check.Assert(newDefinedInterfaceFromSysAdmin, NotNil)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Name, Equals, dummyRde.Name)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Nss, Equals, dummyRde.Nss)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Version, Equals, dummyRde.Version)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Vendor, Equals, dummyRde.Vendor)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.IsReadOnly, Equals, dummyRde.IsReadOnly)
	AddToCleanupListOpenApi(newDefinedInterfaceFromSysAdmin.DefinedInterface.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeInterfaces+newDefinedInterfaceFromSysAdmin.DefinedInterface.ID)

	// Tenants can't create Defined Interfaces. We replace the dots in both
	// nss and name as API is quirky at versions of VCD <= 10.3.0
	nilDefinedInterface, err := tenantUserClient.CreateDefinedInterface(&types.DefinedInterface{
		Name:    strings.ReplaceAll(check.TestName()+"name2", ".", ""),
		Nss:     strings.ReplaceAll(check.TestName()+"name2", ".", ""),
		Version: "4.5.6",
		Vendor:  "vmware",
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

	obtainedDefinedInterface2, err := systemAdministratorClient.GetDefinedInterface(obtainedDefinedInterface.DefinedInterface.Vendor, obtainedDefinedInterface.DefinedInterface.Nss, obtainedDefinedInterface.DefinedInterface.Version)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface2.DefinedInterface, DeepEquals, *obtainedDefinedInterface.DefinedInterface)

	obtainedDefinedInterface2, err = tenantUserClient.GetDefinedInterface(obtainedDefinedInterface.DefinedInterface.Vendor, obtainedDefinedInterface.DefinedInterface.Nss, obtainedDefinedInterface.DefinedInterface.Version)
	check.Assert(err, IsNil)
	check.Assert(*obtainedDefinedInterface2.DefinedInterface, DeepEquals, *obtainedDefinedInterface.DefinedInterface)

	// Update the Defined Interface as System administrator
	err = newDefinedInterfaceFromSysAdmin.Update(types.DefinedInterface{
		Name: dummyRde.Name + "3", // Only name can be updated
	})
	check.Assert(err, IsNil)
	check.Assert(newDefinedInterfaceFromSysAdmin.DefinedInterface.Name, Equals, dummyRde.Name+"3")

	// This one was obtained by the tenant, so it shouldn't be updatable
	err = obtainedDefinedInterface2.Update(types.DefinedInterface{
		Name: dummyRde.Name + "4",
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
	check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)
}

// Test_DefinedInterfaceBehavior tests the CRUD methods of Defined Interfaces to create Behaviors.
// This test can be run with GOVCD_SKIP_VAPP_CREATION option enabled.
func (vcd *TestVCD) Test_DefinedInterfaceBehavior(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeInterfaceBehaviors)

	// Create a new Defined Interface with dummy values, so we can test behaviors on it
	sanizitedTestName := strings.NewReplacer("_", "", ".", "").Replace(check.TestName())
	di, err := vcd.client.CreateDefinedInterface(&types.DefinedInterface{
		Name:    sanizitedTestName,
		Nss:     "nss",
		Version: "1.0.0",
		Vendor:  "vmware",
	})
	check.Assert(err, IsNil)
	AddToCleanupListOpenApi(di.DefinedInterface.ID, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointRdeInterfaces+di.DefinedInterface.ID)
	defer func() {
		err := di.Delete()
		check.Assert(err, IsNil)
	}()

	// Create a new Behavior payload with an Activity type.
	behaviorPayload := types.Behavior{
		Name:        sanizitedTestName,
		Description: "Generated by " + check.TestName(),
		Execution: map[string]interface{}{
			"id":   "TestActivity",
			"type": "Activity",
		},
	}
	behavior, err := di.AddBehavior(behaviorPayload)
	check.Assert(err, IsNil)
	check.Assert(behavior.Name, Equals, behaviorPayload.Name)
	check.Assert(behavior.Description, Equals, behaviorPayload.Description)
	check.Assert(behavior.Ref, Equals, fmt.Sprintf("urn:vcloud:behavior-interface:%s:%s:%s:%s", behaviorPayload.Name, di.DefinedInterface.Vendor, di.DefinedInterface.Nss, di.DefinedInterface.Version))
	check.Assert(behavior.ID, Equals, behavior.Ref)

	// Try to add the same behavior again.
	_, err = di.AddBehavior(behaviorPayload)
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), "RDE_BEHAVIOR_ALREADY_EXISTS"), Equals, true)

	// We check that the Behaviors can be retrieved
	allBehaviors, err := di.GetAllBehaviors(nil)
	check.Assert(err, IsNil)
	check.Assert(1, Equals, len(allBehaviors))
	check.Assert(allBehaviors[0], DeepEquals, behavior)

	// Error getting non-existing Behaviors
	_, err = di.GetBehaviorById("urn:vcloud:behavior-interface:notexist:notexist:notexist:9.9.9")
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)

	_, err = di.GetBehaviorByName("DoesNotExist")
	check.Assert(err, NotNil)
	check.Assert(strings.Contains(err.Error(), ErrorEntityNotFound.Error()), Equals, true)

	// Getting behaviors correctly
	retrievedBehavior, err := di.GetBehaviorById(behavior.ID)
	check.Assert(err, IsNil)
	check.Assert(retrievedBehavior, NotNil)
	check.Assert(retrievedBehavior.Name, Equals, behavior.Name)
	check.Assert(retrievedBehavior.Description, Equals, behavior.Description)
	check.Assert(retrievedBehavior.Execution, DeepEquals, behavior.Execution)

	retrievedBehavior2, err := di.GetBehaviorByName(behavior.Name)
	check.Assert(err, IsNil)
	check.Assert(retrievedBehavior, NotNil)
	check.Assert(retrievedBehavior, DeepEquals, retrievedBehavior2)

	updatePayload := types.Behavior{
		Description: "Updated description",
		Execution: map[string]interface{}{
			"id":   "TestActivityUpdated",
			"type": "Activity",
		},
		Ref:  "notGoingToUpdate1",
		Name: "notGoingToUpdate2",
	}
	_, err = di.UpdateBehavior(updatePayload)
	check.Assert(err, NotNil)
	check.Assert(err.Error(), Equals, "ID of the Behavior to update is empty")

	updatePayload.ID = retrievedBehavior.ID
	updatedBehavior, err := di.UpdateBehavior(updatePayload)
	check.Assert(err, IsNil)
	check.Assert(updatedBehavior.ID, Equals, retrievedBehavior.ID)
	check.Assert(updatedBehavior.Ref, Equals, retrievedBehavior.Ref)   // This cannot be updated
	check.Assert(updatedBehavior.Name, Equals, retrievedBehavior.Name) // This cannot be updated
	check.Assert(updatedBehavior.Execution, DeepEquals, updatePayload.Execution)
	check.Assert(updatedBehavior.Description, Equals, updatePayload.Description)

	err = di.DeleteBehavior(behavior.ID)
	check.Assert(err, IsNil)
}
