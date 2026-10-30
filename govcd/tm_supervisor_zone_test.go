//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmSupervisorZoneVmClassReservations(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	vc, cleanup := getOrCreateVCenter(vcd, check)
	defer cleanup()
	supervisor, err := vc.GetSupervisorByName(vcd.config.Tm.VcenterSupervisor)
	check.Assert(err, IsNil)
	zones, err := supervisor.GetAllSupervisorZones(nil)
	check.Assert(err, IsNil)
	check.Assert(len(zones) > 0, Equals, true)
	for _, zone := range zones {
		reservations, err := zone.GetAllVmClassReservations(url.Values{"pageSize": {"1"}})
		check.Assert(err, IsNil)
		for _, reservation := range reservations {
			check.Assert(reservation.VirtualMachineClassName, Not(Equals), "")
			check.Assert(reservation.MaximumCount >= reservation.AvailableCount, Equals, true)
		}
	}
}
