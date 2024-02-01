//go:build rde || functional || ALL

/*
* Copyright 2024 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"
	"os"
)

func (vcd *TestVCD) Test_BuildLandingZoneRde(check *C) {

	isoFileName := os.Getenv("DSE_ISO") //"vmware-vcd-ds-1.3.0-22829404.iso"
	if isoFileName == "" {
		check.Skip("no .ISO defined")
	}
	// WIP
	rde, err := vcd.client.Client.CreateLandingZoneRde(isoFileName, "administrator", "TBA")
	check.Assert(err, IsNil)
	check.Assert(rde, NotNil)

	contents, err := getContentsFromIsoFiles(isoFileName, wantedFiles)
	check.Assert(err, IsNil)
	for k, v := range contents {
		fmt.Printf("%-15s: %-30s %d\n", k, v.foundFileName, len(v.contents))
	}

}
