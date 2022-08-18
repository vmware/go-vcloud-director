//go:build task || functional || ALL
// +build task functional ALL

/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

/*
TO BE REMOVED (or reintroduced with different scope) : task completion is tested as part of vdc_test.go
import (
	"fmt"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_WaitTaskCompletion(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	check.Skip("Disabled: need a reliable way of triggering a task")
	fmt.Printf("%#v\n", vcd.vapp.VApp)
	task, err := vcd.vapp.Deploy()

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

}
*/

func init() {
	testingTags["task"] = "task_test.go"
}
