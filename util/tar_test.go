/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package util

import (
	"testing"
)

// Tests function sanitizedName providing bad paths and expects them be cleaned.
// Possible security issue https://github.com/vmware/pyvcloud/pull/268
func TestSanitizedName(t *testing.T) {
	tables := []struct {
		badPath  string
		goodPath string
	}{
		{"\\..\\1.txt", "1.txt"},
		{"///foo/bar", "foo/bar"},
		{"C:/loo/bar2", "loo/bar2"},
		{"C:\\loo\\bar2", "loo\\bar2"},
		{"../../foo../../ba..r", "foo../ba..r"},
		{"../my.file", "my.file"},
	}
	for _, table := range tables {
		fixedPath := sanitizedName(table.badPath)
		if fixedPath != table.goodPath {
			t.Errorf("expected and fixedPath didn't match - %s : %s", table.goodPath, fixedPath)
		}
	}
}
