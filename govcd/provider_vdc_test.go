//go:build pvdc || functional || ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_ProviderVdc(check *C) {
	providerVdc, err := vcd.client.GetProviderVdcById("0f32bce0-62c0-481e-bb71-254b1fd40434")
	check.Assert(err, IsNil)
	check.Assert(providerVdc, NotNil)
}
