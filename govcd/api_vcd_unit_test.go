// +build api unit ALL

/*
* Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
* Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import "testing"

func init() {
	testingTags["api_unit"] = "api_vcd_unit_test.go"
}

func Test_splitParent(t *testing.T) {
	type args struct {
		parent    string
		separator string
	}
	tests := []struct {
		name       string
		args       args
		wantFirst  string
		wantSecond string
		wantThird  string
	}{
		{
			name:       "Empty",
			args:       args{parent: "", separator: "|"},
			wantFirst:  "",
			wantSecond: "",
			wantThird:  "",
		},
		{
			name:       "One",
			wantFirst:  "",
			wantSecond: "",
			wantThird:  "",
		},
		{
			name:       "Two",
			args:       args{parent: "first|second", separator: "|"},
			wantFirst:  "first",
			wantSecond: "second",
			wantThird:  "",
		},
		{
			name:       "Three",
			args:       args{parent: "first|second|third", separator: "|"},
			wantFirst:  "first",
			wantSecond: "second",
			wantThird:  "third",
		},
		{
			name:       "Four",
			args:       args{parent: "first|second|third|fourth", separator: "|"},
			wantFirst:  "",
			wantSecond: "",
			wantThird:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFirst, gotSecond, gotThird := splitParent(tt.args.parent, tt.args.separator)
			if gotFirst != tt.wantFirst {
				t.Errorf("splitParent() gotFirst = %v, want %v", gotFirst, tt.wantFirst)
			}
			if gotSecond != tt.wantSecond {
				t.Errorf("splitParent() gotSecond = %v, want %v", gotSecond, tt.wantSecond)
			}
			if gotThird != tt.wantThird {
				t.Errorf("splitParent() gotThird = %v, want %v", gotThird, tt.wantThird)
			}
		})
	}
}
