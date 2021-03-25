// +build query functional ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

// TODO: Need to add a check to check the contents of the query
func (vcd *TestVCD) Test_Query(check *C) {
	// Get the Org populated
	_, err := vcd.client.Query(ctx, map[string]string{"type": "vm"})
	check.Assert(err, IsNil)
}
