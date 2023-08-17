//go:build vapp || functional || ALL
// +build vapp functional ALL

/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"time"
)

// TestVappfromTemplateAndClone creates a vApp with multiple VMs at once, then clones such vApp into a new one
func (vcd *TestVCD) TestVappfromTemplateAndClone(check *C) {
	org, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(org, NotNil)

	vdc, err := org.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)
	check.Assert(vdc, NotNil)

	name := check.TestName()
	description := "test compose raw vApp with template"

	catalog, err := org.GetCatalogByName(vcd.config.VCD.Catalog.NsxtBackedCatalogName, false)
	check.Assert(err, IsNil)
	vappTemplateName := vcd.config.VCD.Catalog.CatalogItemWithMultiVms
	if vappTemplateName == "" {
		check.Skip(fmt.Sprintf("vApp template missing in configuration - Make sure there is such template in catalog %s -"+
			" Using test_resources/vapp_with_3_vms.ova",
			vcd.config.VCD.Catalog.NsxtBackedCatalogName))
	}
	vappTemplate, err := catalog.GetVAppTemplateByName(vappTemplateName)
	if err != nil {
		if ContainsNotFound(err) {
			check.Skip(fmt.Sprintf("vApp template %s not found - Make sure there is such template in catalog %s -"+
				" Using test_resources/vapp_with_3_vms.ova",
				vappTemplateName, vcd.config.VCD.Catalog.NsxtBackedCatalogName))
		}
	}
	check.Assert(err, IsNil)
	check.Assert(vappTemplate.VAppTemplate.Children, NotNil)
	check.Assert(vappTemplate.VAppTemplate.Children.VM, NotNil)

	var def = types.InstantiateVAppTemplateParams{
		Name:        name,
		Deploy:      true,
		PowerOn:     true,
		Description: description,
		Source: &types.Reference{
			HREF: vappTemplate.VAppTemplate.HREF,
			ID:   vappTemplate.VAppTemplate.ID,
		},
		IsSourceDelete:   false,
		AllEULAsAccepted: true,
	}

	start := time.Now()
	printVerbose("creating vapp '%s' from template '%s'\n", name, vappTemplateName)
	vapp, err := vdc.CreateVappFromTemplate(&def)
	check.Assert(err, IsNil)
	printVerbose("** created in %s\n", time.Since(start))

	AddToCleanupList(name, "vapp", vdc.Vdc.Name, name)

	check.Assert(vapp.VApp.Name, Equals, name)
	check.Assert(vapp.VApp.Description, Equals, description)

	check.Assert(vapp.VApp.Children, NotNil)
	check.Assert(vapp.VApp.Children.VM, NotNil)

	cloneName := name + "-clone"
	cloneDescription := description + " clone"
	var defClone = types.CloneVAppParams{
		Name:        cloneName,
		Deploy:      true,
		PowerOn:     true,
		Description: cloneDescription,
		Source: &types.Reference{
			HREF: vapp.VApp.HREF,
			Type: vapp.VApp.Type,
		},
		IsSourceDelete: addrOf(false),
	}

	start = time.Now()
	printVerbose("cloning vapp '%s' from vapp '%s'\n", cloneName, name)
	vapp2, err := vdc.CloneVapp(&defClone)
	check.Assert(err, IsNil)
	printVerbose("** cloned in %s\n", time.Since(start))

	AddToCleanupList(cloneName, "vapp", vdc.Vdc.Name, name)

	status, err := vapp2.GetStatus()
	check.Assert(err, IsNil)
	if status == "SUSPENDED" {
		printVerbose("\t discarding suspended state for vApp %s\n", vapp2.VApp.Name)
		err = vapp2.DiscardSuspendedState()
		check.Assert(err, IsNil)
		status, err = vapp2.GetStatus()
		check.Assert(err, IsNil)
		if status != "POWERED_ON" {
			printVerbose("\t powering on vApp %s\n", vapp2.VApp.Name)
			task, err := vapp2.PowerOn()
			check.Assert(err, IsNil)
			err = task.WaitTaskCompletion()
			check.Assert(err, IsNil)
		}
	}

	check.Assert(vapp2.VApp.Name, Equals, cloneName)
	check.Assert(vapp2.VApp.Description, Equals, cloneDescription)
	check.Assert(vapp.VApp.HREF, Not(Equals), vapp2.VApp.HREF)

	vappRemove(vapp, check)
	vappRemove(vapp2, check)
}

func vappRemove(vapp *VApp, check *C) {
	var task Task
	var err error
	status, err := vapp.GetStatus()
	check.Assert(err, IsNil)
	if status == "POWERED_ON" {
		printVerbose("powering off vApp '%s'\n", vapp.VApp.Name)
		task, err = vapp.Undeploy()
		check.Assert(err, IsNil)
		err = task.WaitTaskCompletion()
		check.Assert(err, IsNil)
	}

	printVerbose("removing networks from vApp '%s'\n", vapp.VApp.Name)
	task, err = vapp.RemoveAllNetworks()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	printVerbose("removing vApp '%s'\n", vapp.VApp.Name)
	task, err = vapp.Delete()
	check.Assert(err, IsNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)
}
