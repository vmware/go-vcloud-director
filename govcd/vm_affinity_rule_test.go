// +build vdc affinity functional ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"context"
	"fmt"
	"os"

	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// testGetVmAffinityRuleList tests that we can retrieve a list of VM affinity rules
func testGetVmAffinityRuleList(ctx context.Context, vdc *Vdc, check *C) {
	fmt.Printf("Running: test_GetVmAffinityRuleList\n")
	_, err := vdc.GetAllVmAffinityRuleList(ctx)
	check.Assert(err, IsNil)
}

// testGetVmAffinityRule tests VmAffinityRule retrieval by name, by ID, and by a combination of name and ID
func (vcd *TestVCD) testGetVmAffinityRule(ctx context.Context, vdc *Vdc, check *C) {
	fmt.Printf("Running: testGetVmAffinityRule \n")

	affinityRuleName := ""
	list, err := vdc.GetAllVmAffinityRuleList(ctx)
	check.Assert(err, IsNil)
	if len(list) == 0 {
		check.Skip("No affinity rules found")
		return
	}
	affinityRuleName = list[0].Name

	getByName := func(name string, refresh bool) (genericEntity, error) {
		rules, err := vdc.GetVmAffinityRulesByName(ctx, name, "")
		if err != nil {
			return genericEntity(nil), err
		}
		if len(rules) == 0 {
			return genericEntity(nil), ErrorEntityNotFound
		}
		if len(rules) > 1 {
			return genericEntity(nil), fmt.Errorf("more than one item found with this name")
		}
		gRule := VmAffinityRule(*rules[0])
		return &gRule, nil
	}
	getById := func(id string, refresh bool) (genericEntity, error) {
		rule, err := vdc.GetVmAffinityRuleById(ctx, id)
		if err != nil {
			return genericEntity(nil), err
		}
		gRule := VmAffinityRule(*rule)
		return &gRule, nil
	}
	getByNameOrId := func(id string, refresh bool) (genericEntity, error) {
		rule, err := vdc.GetVmAffinityRuleByNameOrId(ctx, id)
		if err != nil {
			return genericEntity(nil), err
		}
		gRule := VmAffinityRule(*rule)
		return &gRule, nil
	}

	var def = getterTestDefinition{
		parentType:    "Vdc",
		parentName:    vdc.Vdc.Name,
		entityType:    "VmAffinityRule",
		entityName:    affinityRuleName,
		getByName:     getByName,
		getById:       getById,
		getByNameOrId: getByNameOrId,
	}
	vcd.testFinderGetGenericEntity(def, check)
}

type affinityRuleData struct {
	name        string
	polarity    string
	creationVms []*types.Vm
	updateVms   []*types.Vm
}

// testCRUDVmAffinityRule tests creation, update, deletion, and read for VM affinity rules
func (vcd *TestVCD) testCRUDVmAffinityRule(ctx context.Context, orgName string, vdc *Vdc, data affinityRuleData, check *C) {
	fmt.Printf("Running: testCRUDVmAffinityRule (%s-%s-%d)\n", data.name, data.polarity, len(data.creationVms))

	var vmReferences []*types.Reference

	for _, vm := range data.creationVms {
		vmReferences = append(vmReferences, &types.Reference{HREF: vm.HREF})
		if testVerbose {
			fmt.Printf("rule %s %s: %s\n", data.polarity, data.name, vm.Name)
		}
	}
	affinityRuleDef := &types.VmAffinityRule{
		Name:        data.name,
		IsEnabled:   takeBoolPointer(true),
		IsMandatory: takeBoolPointer(true),
		Polarity:    data.polarity,
		VmReferences: []*types.VMs{
			&types.VMs{
				VMReference: vmReferences,
			},
		},
	}
	vmAffinityRule, err := vdc.CreateVmAffinityRule(ctx, affinityRuleDef)
	check.Assert(err, IsNil)
	AddToCleanupList(vmAffinityRule.VmAffinityRule.ID, "affinity_rule", orgName+"|"+vdc.Vdc.Name, "testCRUDVmAffinityRule")

	// Update with VM replacement
	for i, vm := range data.updateVms {
		if i >= len(data.creationVms) {
			break
		}
		vmAffinityRule.VmAffinityRule.VmReferences[0].VMReference[i].HREF = vm.HREF
		if testVerbose {
			fmt.Printf("rule %s: update %s\n", data.polarity, vm.Name)
		}
	}

	vmAffinityRule.VmAffinityRule.Name = data.name + "-with-change"
	err = vmAffinityRule.Update(ctx)
	check.Assert(err, IsNil)
	vmAffinityRule, err = vdc.GetVmAffinityRuleByHref(ctx, vmAffinityRule.VmAffinityRule.HREF)
	check.Assert(err, IsNil)
	check.Assert(vmAffinityRule.VmAffinityRule.Name, Equals, data.name+"-with-change")

	// Update with VM removal
	if len(data.creationVms) > 2 {
		if testVerbose {
			fmt.Printf("removing %s\n", vmAffinityRule.VmAffinityRule.VmReferences[0].VMReference[0].HREF)
		}
		vmAffinityRule.VmAffinityRule.Name = data.name + "-with-removal"
		vmAffinityRule.VmAffinityRule.VmReferences[0].VMReference[0] = nil
		err = vmAffinityRule.Update(ctx)
		check.Assert(err, IsNil)
		vmAffinityRule, err = vdc.GetVmAffinityRuleByHref(ctx, vmAffinityRule.VmAffinityRule.HREF)
		check.Assert(err, IsNil)
		check.Assert(vmAffinityRule.VmAffinityRule.Name, Equals, data.name+"-with-removal")
	}

	err = vmAffinityRule.SetEnabled(ctx, false)
	check.Assert(err, IsNil)
	vmAffinityRule, err = vdc.GetVmAffinityRuleByHref(ctx, vmAffinityRule.VmAffinityRule.HREF)
	check.Assert(err, IsNil)
	check.Assert(*vmAffinityRule.VmAffinityRule.IsEnabled, Equals, false)
	check.Assert(*vmAffinityRule.VmAffinityRule.IsMandatory, Equals, true)

	err = vmAffinityRule.SetMandatory(ctx, false)
	check.Assert(err, IsNil)
	vmAffinityRule, err = vdc.GetVmAffinityRuleByHref(ctx, vmAffinityRule.VmAffinityRule.HREF)
	check.Assert(err, IsNil)
	check.Assert(*vmAffinityRule.VmAffinityRule.IsMandatory, Equals, false)

	err = vmAffinityRule.SetEnabled(ctx, true)
	check.Assert(err, IsNil)
	vmAffinityRule, err = vdc.GetVmAffinityRuleByHref(ctx, vmAffinityRule.VmAffinityRule.HREF)
	check.Assert(err, IsNil)
	check.Assert(*vmAffinityRule.VmAffinityRule.IsEnabled, Equals, true)
	check.Assert(*vmAffinityRule.VmAffinityRule.IsMandatory, Equals, false)

	err = vmAffinityRule.SetMandatory(ctx, true)
	check.Assert(err, IsNil)
	vmAffinityRule, err = vdc.GetVmAffinityRuleByHref(ctx, vmAffinityRule.VmAffinityRule.HREF)
	check.Assert(err, IsNil)
	check.Assert(*vmAffinityRule.VmAffinityRule.IsMandatory, Equals, true)
	check.Assert(*vmAffinityRule.VmAffinityRule.IsEnabled, Equals, true)

	// Read
	testGetVmAffinityRuleList(ctx, vdc, check)
	vcd.testGetVmAffinityRule(ctx, vdc, check)

	// Delete
	err = vmAffinityRule.Delete(ctx)
	check.Assert(err, IsNil)
	if testVerbose {
		fmt.Println()
	}
}

// Test_VmAffinityRule prepares the environment for testing VM affinity rules
// After creating appropriate VMs, it runs the CRUD test for several combination of affinity rules
func (vcd *TestVCD) Test_VmAffinityRule(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	if vcd.config.VCD.Org == "" {
		check.Skip("Test_VmAffinityRule: Org name not given")
		return
	}
	if vcd.config.VCD.Vdc == "" {
		check.Skip("Test_VmAffinityRule: VDC name not given")
		return
	}
	ctx := context.Background()
	org, err := vcd.client.GetOrgByName(ctx, vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(ctx, vcd.config.VCD.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	vappDefinition := map[string][]string{
		"Test_EmptyVmVapp1": []string{"Test_EmptyVm1a", "Test_EmptyVm1b"},
		"Test_EmptyVmVapp2": []string{"Test_EmptyVm2a", "Test_EmptyVm2b"},
		"Test_EmptyVmVapp3": []string{"Test_EmptyVm3a", "Test_EmptyVm3b"},
	}
	vappList, err := makeVappGroup(ctx, "TestVdc_CreateVmAffinityRule", vcd.vdc, vappDefinition)
	check.Assert(err, IsNil)

	defer func() {
		if os.Getenv("GOVCD_KEEP_TEST_OBJECTS") != "" {
			if testVerbose {
				fmt.Printf("Skipping vApp removal: GOVCD_KEEP_TEST_OBJECTS was set \n")
			}
			return
		}
		for _, vapp := range vappList {
			if testVerbose {
				fmt.Printf("Removing vApp %s\n", vapp.VApp.Name)
			}
			task, err := vapp.Delete(ctx)
			if err == nil {
				_ = task.WaitTaskCompletion(ctx)
			}
		}
	}()
	check.Assert(len(vappList), Equals, len(vappDefinition))

	vcd.testCRUDVmAffinityRule(ctx, org.Org.Name, vdc, affinityRuleData{
		name:     "Test_VmAffinityRule1",
		polarity: types.PolarityAffinity,
		creationVms: []*types.Vm{
			vappList[0].VApp.Children.VM[0],
			vappList[0].VApp.Children.VM[1],
		},
		updateVms: []*types.Vm{
			vappList[1].VApp.Children.VM[0],
			vappList[1].VApp.Children.VM[1],
		},
	}, check)

	vcd.testCRUDVmAffinityRule(ctx, org.Org.Name, vdc, affinityRuleData{
		name:     "Test_VmAffinityRule2",
		polarity: types.PolarityAffinity,
		creationVms: []*types.Vm{
			vappList[0].VApp.Children.VM[0],
			vappList[0].VApp.Children.VM[1],
			vappList[1].VApp.Children.VM[0],
		},
		updateVms: []*types.Vm{
			vappList[2].VApp.Children.VM[0],
			vappList[2].VApp.Children.VM[1],
		},
	}, check)

	vcd.testCRUDVmAffinityRule(ctx, org.Org.Name, vdc, affinityRuleData{
		name:     "Test_VmAffinityRule3",
		polarity: types.PolarityAntiAffinity,
		creationVms: []*types.Vm{
			vappList[0].VApp.Children.VM[0],
			vappList[1].VApp.Children.VM[0],
		},
		updateVms: []*types.Vm{
			vappList[2].VApp.Children.VM[0],
		},
	}, check)

}

// makeVappGroup creates multiple vApps, each with several VMs,
// as defined in `groupDefinition`.
// Returns a list of vApps
func makeVappGroup(ctx context.Context, label string, vdc *Vdc, groupDefinition map[string][]string) ([]*VApp, error) {
	var vappList []*VApp
	for vappName, vmNames := range groupDefinition {
		existingVapp, err := vdc.GetVAppByName(ctx, vappName, false)
		if err == nil {

			if existingVapp.VApp.Children == nil || len(existingVapp.VApp.Children.VM) == 0 {
				return nil, fmt.Errorf("found vApp %s but without VMs", vappName)
			}
			foundVms := 0
			for _, vmName := range vmNames {
				for _, existingVM := range existingVapp.VApp.Children.VM {
					if existingVM.Name == vmName {
						foundVms++
					}
				}
			}
			if foundVms < 2 {
				return nil, fmt.Errorf("found vApp %s but with %d VMs instead of 2 ", vappName, foundVms)
			}

			vappList = append(vappList, existingVapp)
			if testVerbose {
				fmt.Printf("Using existing vApp %s\n", vappName)
			}
			continue
		}

		if testVerbose {
			fmt.Printf("Creating vApp %s\n", vappName)
		}
		vapp, err := makeEmptyVapp(ctx, vdc, vappName)
		if err != nil {
			return nil, err
		}
		if os.Getenv("GOVCD_KEEP_TEST_OBJECTS") == "" {
			AddToCleanupList(vappName, "vapp", vdc.Vdc.Name, label)
		}
		for _, vmName := range vmNames {
			if testVerbose {
				fmt.Printf("\tCreating VM %s/%s\n", vappName, vmName)
			}
			_, err := makeEmptyVm(ctx, vapp, vmName)
			if err != nil {
				return nil, err
			}
		}
		vappList = append(vappList, vapp)
	}
	return vappList, nil
}
