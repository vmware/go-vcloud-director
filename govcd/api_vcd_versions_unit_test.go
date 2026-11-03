//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import "testing"

func Test_advertisesTmVersion(t *testing.T) {
	build := func(versions ...string) SupportedVersions {
		var s SupportedVersions
		for _, v := range versions {
			s.VersionInfos = append(s.VersionInfos, VersionInfo{Version: v})
		}
		return s
	}
	tests := []struct {
		name     string
		versions []string
		want     bool
	}{
		{"empty", nil, false},
		{"VCD 10.4", []string{"36.0", "36.1", "36.2", "36.3", "37.0"}, false},
		{"VCD 10.6", []string{"37.0", "38.0", "38.1", "39.0"}, false},
		{"legacy VCD", []string{"29.0", "30.0", "31.0"}, false},
		{"TM alias 40.0", []string{"39.0", "40.0"}, true},
		{"TM alias 40.1", []string{"40.1"}, true},
		{"TM preview alias", []string{"41.0.0-alpha"}, true},
		{"TM product version", []string{"9.0.0", "9.1.1", "9.1.2"}, true},
		{"too new three-part", []string{"31.0.0"}, false},
		{"garbage", []string{"abc", ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := advertisesTmVersion(build(tt.versions...)); got != tt.want {
				t.Errorf("advertisesTmVersion(%v) = %t, want %t", tt.versions, got, tt.want)
			}
		})
	}
}
