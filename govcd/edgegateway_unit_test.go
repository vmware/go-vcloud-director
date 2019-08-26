// +build unit ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"regexp"
	"testing"
)

func TestGetPseudoUUID(t *testing.T) {

	var seen = make(map[string]int)

	reUuid := regexp.MustCompile(`^[A-F0-9]{8}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{12}$`)
	for N := 0; N < 1000; N++ {
		uuid, _ := getPseudoUuid()
		if !reUuid.MatchString(uuid) {
			t.Logf("string %s doesn't look like a UUID", uuid)
			t.Fail()
		}
		previous, found := seen[uuid]
		if found {
			t.Logf("uuid %s already in the generated list at position %d", uuid, previous)
			t.Fail()
		}
		seen[uuid] = N
	}
}
