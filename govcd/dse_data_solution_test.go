//go:build slz || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"github.com/davecgh/go-spew/spew"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_Dse(check *C) {
	if vcd.client.Client.APIVCDMaxVersionIs("< 37.1") {
		check.Skip("Solution Landing Zones are supported in VCD 10.4.1+")
	}

	t, err := vcd.client.GetRdeType(dataSolutionOrgConfig[0], dataSolutionOrgConfig[1], dataSolutionOrgConfig[2])
	check.Assert(err, IsNil)

	spew.Dump(t.DefinedEntityType)

	// params := copyOrNewUrlValues(nil)
	// params = queryParameterFilterAnd(params,fmt.Sprintf(""))

	// allTypes, err := vcd.client.GetAllRdeTypes(params)
	// check.Assert(err, IsNil)
	// // spew.Dump(allTypes)

	// for i := range allTypes {
	// 	spew.Dump(allTypes[i].DefinedEntityType.ID)
	// }

	// if vcd.config.SolutionAddOn.Org == "" || vcd.config.SolutionAddOn.Catalog == "" {
	// 	check.Skip("Solution Add-On configuration is not present")
	// }

	// dseConfigs, err := vcd.client.GetAllDseConfigs(nil)
	// check.Assert(err, IsNil)
	// fmt.Println(len(dseConfigs))
	// // spew.Dump(dseConfigs[0].DefinedEntity.DefinedEntity)
	// // spew.Dump(dseConfigs[0].DseConfig)

	// for i := range dseConfigs {
	// 	fmt.Println(dseConfigs[i].DefinedEntity.DefinedEntity.Name)
	// 	// fmt.Println(dseConfigs[i].DseConfig.Spec.Artifacts[0].Name)
	// 	// spew.Dump(dseConfigs[i].DseConfig)
	// 	fmt.Println("-------")

	// 	// if dseConfigs[i].DefinedEntity.DefinedEntity.Name == "VCD Data Solutions" {
	// 	// 	spew.Dump(dseConfigs[i].DseConfig)
	// 	// }

	// }

	// singleItem, err := vcd.client.GetDseConfigByName("Confluent Platform")
	// check.Assert(err, IsNil)

	// spew.Dump(singleItem.DseConfig)

}
