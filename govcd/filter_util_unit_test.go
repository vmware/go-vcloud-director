//go:build unit || ALL

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"testing"

	"github.com/kr/pretty"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func Test_compareDate(t *testing.T) {
	type args struct {
		wanted string
		got    string
	}
	tests := []struct {
		args args
		want bool
	}{
		// Note: "YYYY-MM-DDThh:mm:ss.µµµZ" (time.RFC3339) is the format used for catalog item
		// and vApp templates creation dates

		{args{"=2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.500Z"}, true},
		{args{">2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.500Z"}, false},
		{args{"> 2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.500Z"}, false},
		{args{">2020-03-09T09:50:50.500Z", "2020-03-10T09:50:51.500Z"}, true},
		{args{">2020-03-09T09:50:51.500Z", "2020-03-09T09:51:51.500Z"}, true},
		{args{">=2020-03-09T09:50:51.500Z", "2020-03-09T09:51:51.500Z"}, true},
		{args{">=2020-03-09T09:50:51.500Z", "2020-03-09T09:51:51.500Z"}, true},
		{args{"<=2020-03-09T09:50:51.500Z", "2020-03-08T09:50:51.500Z"}, true},
		{args{">2020-03-09T09:50:51.500Z", "2020-04-08T00:00:01.0Z"}, true},
		{args{">2020-03-09", "2020-04-08T00:00:01.0Z"}, true},
		{args{"> January 10th, 2020", "2020-04-08T00:00:01.0Z"}, true},
		{args{"<= March 1st, 2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{"<= 01-mar-2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{"<= 02-feb-2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{">= 02-may-2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{">= 02-jan-2020", "2020-04-08T00:00:01.0Z"}, true},
		{args{">= 03-jan-2020", "2020-04-08T00:00:01.0Z"}, true},
		{args{">= 02-Apr-2020", "2020-04-08T00:00:01.0Z"}, true},
		{args{">2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.501Z"}, true},
		{args{"<2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.499Z"}, true},
		{args{"=2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.499Z"}, false},
		{args{"==2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.499Z"}, false},
		{args{"==2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.500Z"}, true},
		{args{">2020-04-18 10:00:00.123456", "2020-04-18 10:00:00.123457"}, true},
		{args{">2020-04-18 10:00:00.1234567", "2020-04-18 10:00:00.1234568"}, true},
		{args{">2020-04-18 10:00:00.12345678", "2020-04-18 10:00:00.12345679"}, true},
		{args{">2020-04-18 10:00:00.123456789", "2020-04-18 10:00:00.123456790"}, true},
		{args{"==2020-04-18 10:00:00.123456789", "2020-04-18 10:00:00.123456789"}, true},
		{args{"==2020-04-18 10:00:00.123456790", "2020-04-18 10:00:00.123456790"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.args.wanted, func(t *testing.T) {
			got, err := compareDate(tt.args.wanted, tt.args.got)
			if err != nil {
				t.Errorf("compareDate() = %v, error: %s", got, err)
			}
			if got != tt.want {
				t.Errorf("compareDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeDateFilters(t *testing.T) {
	input := []DateItem{
		{
			Name:   "one",
			Date:   "2020-01-01",
			Entity: nil,
		},
		{
			Name:   "two",
			Date:   "2020-02-01",
			Entity: nil,
		},
		{
			Name:   "three",
			Date:   "2020-03-01",
			Entity: nil,
		},
	}
	output, err := makeDateFilter(input)
	if err != nil {
		t.Errorf("error creating date filters :%s", err)
	}
	if len(output) != len(input)+2 {
		t.Errorf("len(output): want %d - got : %d", len(input)+2, len(output))
	}
	foundEqual := 0
	foundEarliest := false
	foundLatest := false
	wantLatest := ">2020-01-01"
	wantEarliest := "<2020-03-01"
	for _, iItem := range input {
		want := "==" + iItem.Date
		for _, oItem := range output {
			for name, filter := range oItem.Criteria.Filters {
				if name == types.FilterDate {
					if filter == want {
						foundEqual++
					}
					if filter == wantEarliest {
						foundEarliest = true
					}
					if filter == wantLatest {
						foundLatest = true
					}
				}
			}
		}
	}
	if foundEqual != len(input) {
		t.Errorf("foundEqual: want %d, found %d", len(input), foundEqual)
	}
	if !foundLatest {
		t.Errorf("foundLatest not detected")
	}
	if !foundEarliest {
		t.Errorf("foundEarliest not detected")
	}

	logVerbose(t, "%# v\n", pretty.Formatter(output))
}
