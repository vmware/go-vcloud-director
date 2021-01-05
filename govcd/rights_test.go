package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_Rights(check *C) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	// Step 1 - Get all rights
	allExistingRights, err := adminOrg.GetAllOpenApiRights(nil)
	check.Assert(err, IsNil)
	check.Assert(allExistingRights, NotNil)

	fmt.Printf("how many rights: %d\n",len(allExistingRights))

	for _, oneRight := range allExistingRights {
		fmt.Printf("%-20s %-53s %s\n", oneRight.Name, oneRight.ID, oneRight.Category)
	}
}
