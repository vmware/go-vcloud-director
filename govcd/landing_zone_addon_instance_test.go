//go:build slz || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_SolutionAddOnInstance(check *C) {
	// if vcd.config.VCD.Catalog.NsxtCatalogAddonDse == "" {
	// 	check.Skip("missing 'VCD.Catalog.NsxtCatalogAddonDse' value")
	// }

	addOn, err := vcd.client.GetSolutionAddonByName("vmware.ds-1.4.0-23376809")
	check.Assert(err, IsNil)

	inputs := make(map[string]interface{})
	inputs["operation"] = "create instance"
	inputs["name"] = "ds-ASMyN3"
	inputs["input-delete-previous-uiplugin-versions"] = false

	addOnInstance, res, err := addOn.CreateSolutionAddOnInstance(inputs)
	check.Assert(err, IsNil)
	check.Assert(addOnInstance, NotNil)
	fmt.Println(res)

	// addOnInstance, err := addOn.GetInstanceByName("ds-ASMyN3")
	// check.Assert(err, IsNil)

	deleteInputs := make(map[string]interface{})
	deleteInputs["operation"] = "delete instance"
	deleteInputs["name"] = addOnInstance.SolutionAddOnInstance.AddonInstanceSolutionName
	deleteInputs["input-force-delete"] = true
	res2, err := addOnInstance.RemoveSolutionAddOnInstance(deleteInputs)
	check.Assert(err, IsNil)
	fmt.Println(res2)

}
