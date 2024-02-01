//go:build rde || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_BuildLandingZoneRde(check *C) {
	contents, err := getContentsFromIsoFiles("/Users/gmaxia/workdir/git/dataclouder/data-solutions/vmware-vcd-ds-1.3.0-22829404.iso", wantedFiles)
	check.Assert(err, IsNil)
	for k, v := range contents {
		fmt.Printf("%-15s: %-30s %d\n", k, v.foundFileName, len(v.contents))
	}

}
