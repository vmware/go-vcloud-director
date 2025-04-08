//go:build query || functional || ALL

package govcd

import (
	. "gopkg.in/check.v1"
)

// TODO: Need to add a check to check the contents of the query
func (vcd *TestVCD) Test_Query(check *C) {
	// Get the Org populated
	_, err := vcd.client.Query(map[string]string{"type": "vm"})
	check.Assert(err, IsNil)
}
