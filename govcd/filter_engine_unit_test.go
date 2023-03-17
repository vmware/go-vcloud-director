//go:build unit

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"reflect"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Test filter engine using dependency injection

// Mocking object implementing QueryItem interface
type TestItem struct {
	name     string
	date     string
	ip       string
	ttype    string
	metadata StringMap
	parent   string
	parentId string
}

// mock queryWithMetadataFields function that returns an empty result
var dummyQwithM = func(queryType string, params, notEncodedParams map[string]string,
	metadataFields []string, isSystem bool) (Results, error) {
	return Results{}, nil
}

// mock QueryByMetadataFields function that returns an empty result
var dummyQbyM = func(queryType string, params, notEncodedParams map[string]string,
	metadataFilters map[string]MetadataFilter, isSystem bool) (Results, error) {
	return Results{}, nil
}

func (t TestItem) GetDate() string                    { return t.date }
func (t TestItem) GetName() string                    { return t.name }
func (t TestItem) GetType() string                    { return t.ttype }
func (t TestItem) GetIp() string                      { return t.ip }
func (t TestItem) GetParentName() string              { return t.parent }
func (t TestItem) GetParentId() string                { return t.parentId }
func (t TestItem) GetMetadataValue(key string) string { return t.metadata[key] }
func (t TestItem) GetHref() string                    { return "" }

func makeNameCriteria(name string) *FilterDef {
	f := NewFilterDef()
	_ = f.AddFilter(types.FilterNameRegex, name)
	return f
}

func makeIpCriteria(ip string) *FilterDef {
	f := NewFilterDef()
	_ = f.AddFilter(types.FilterIp, ip)
	return f
}

func makeDateCriteria(date string, latest, earliest bool) *FilterDef {
	f := NewFilterDef()
	if date != "" {
		_ = f.AddFilter(types.FilterDate, date)
	}
	if earliest {
		_ = f.AddFilter(types.FilterEarliest, "true")
	}
	if latest {
		_ = f.AddFilter(types.FilterLatest, "true")
	}
	return f
}

func makeMDCriteria(useApi bool, wanted StringMap) *FilterDef {
	f := NewFilterDef()
	for key, value := range wanted {
		_ = f.AddMetadataFilter(key, value, "STRING", false, useApi)
	}
	return f
}

func makeDateMDCriteria(date string, useApi bool, wanted StringMap) *FilterDef {
	f := makeMDCriteria(useApi, wanted)
	_ = f.AddFilter(types.FilterDate, date)
	return f
}

func Test_searchByFilter(t *testing.T) {
	type args struct {
		qByM      queryByMetadataFunc
		qWithM    queryWithMetadataFunc
		converter resultsConverterFunc
		queryType string
		criteria  *FilterDef
	}

	type testDef struct {
		name    string
		args    args
		want    []QueryItem
		wantErr bool
	}

	// This is the fake data being tested
	data := []QueryItem{
		TestItem{
			name:     "one",
			ip:       "192.168.1.10",
			date:     "2020-01-01",
			ttype:    "test-name",
			metadata: StringMap{"xyz": "xxx"},
		},
		TestItem{
			name:     "two",
			ip:       "10.10.8.9",
			date:     "2020-02-01",
			ttype:    "test-name",
			metadata: StringMap{"abc": "yyy"},
		},
		TestItem{
			name:     "three",
			ip:       "10.150.20.1",
			date:     "2020-02-03 10:00:00",
			ttype:    "test-date",
			metadata: StringMap{"abc": "-+!", "other": "aaa"},
		},
	}

	// Injection function that replaces the resultConverterFunc and returns the test data
	var getData = func(string, Results) ([]QueryItem, error) {
		return data, nil
	}

	// Test cases that should succeed
	testsSuccess := []testDef{
		// Empty filter: gets all items
		{
			name: "empty-filter1",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria: &FilterDef{
					Filters:              map[string]string{},
					Metadata:             []MetadataDef{},
					UseMetadataApiFilter: false,
				},
			},
			want: data,
		},
		// Empty filter: gets all items
		{
			name: "empty-filter2",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  &FilterDef{},
			},
			want: data,
		},
		// Null filter: gets all items
		{
			name: "null-filter",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  nil,
			},
			want: data,
		},
		// Gets a name by regular expression (finds 1)
		{
			name: "name1",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeNameCriteria(`o\w+`),
			},
			want: []QueryItem{data[0]},
		},
		// Gets a name by regular expression (finds 2)
		{
			name: "name2",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeNameCriteria(`t\w+`),
			},
			want: []QueryItem{data[1], data[2]},
		},
		// Gets a name by full value (finds 1)
		{
			name: "name3",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeNameCriteria(`two`),
			},
			want: []QueryItem{data[1]},
		},
		// Use date comparison to get an item that's a few hours older than the date referenced (finds 1)
		{
			name: "dateComparison",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeDateCriteria("> 2020-02-03", false, false),
			},
			want: []QueryItem{data[2]},
		},
		// Gets the newest element (finds 1)
		{
			name: "dateLatest",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeDateCriteria("", true, false),
			},
			want: []QueryItem{data[2]},
		},
		// Gets the oldest element (finds 1)
		{
			name: "dateEarliest",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeDateCriteria("", false, true),
			},
			want: []QueryItem{data[0]},
		},
		// Gets the items with IP containing "10" (finds 3)
		{
			name: "ip10",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`10`),
			},
			want: []QueryItem{data[0], data[1], data[2]},
		},
		// Gets the items with IP starting by 10 (finds 2)
		{
			name: "ip^10",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`^10`),
			},
			want: []QueryItem{data[1], data[2]},
		},
		// Gets the items with IP ending with "10" (finds 1)
		{
			name: "ip10$",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`10$`),
			},
			want: []QueryItem{data[0]},
		},
		// Gets the items with IP starting by 192 (finds 1)
		{
			name: "ip192",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`^192`),
			},
			want: []QueryItem{data[0]},
		},
		// Gets the items with IP starting by 10.150 (finds 1)
		{
			name: "ip10.150",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`10\.150`),
			},
			want: []QueryItem{data[2]},
		},
		// Gets the items with metadata "abc" and any value (finds 2)
		{
			name: "metaAbc1",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeMDCriteria(false, StringMap{"abc": `\S+`}),
			},
			want: []QueryItem{data[1], data[2]},
		},
		// Gets the items with metadata "abc" and alphanumeric value (finds 1)
		{
			name: "metaAbc2",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeMDCriteria(false, StringMap{"abc": `\w+`}),
			},
			want: []QueryItem{data[1]},
		},
		// Gets the items with metadata "xyz" and alphanumeric value (finds 1)
		{
			name: "metaXyz",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeMDCriteria(false, StringMap{"xyz": `\w+`}),
			},
			want: []QueryItem{data[0]},
		},
		// Gets the items with two metadata values (finds 1)
		{
			name: "metaOther",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeMDCriteria(false, StringMap{"abc": `\S+`, "other": `\w+`}),
			},
			want: []QueryItem{data[2]},
		},
		// Gets the items with metadata "abc" and any value combined with date search (finds 1)
		{
			name: "metaCombined",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeDateMDCriteria("> 2020-02-02", false, StringMap{"abc": `\S+`}),
			},
			want: []QueryItem{data[2]},
		},
	}

	// Test cases that should fail, i.e. will return an empty list
	testsFailure := []testDef{
		// Searches for a date that is one second higher than the highest date in the test set
		{
			name: "failDateHigher",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeDateCriteria(">= 2020-02-03 10:00:01", false, false),
			},
			want: nil,
		},
		// Searches for a date that is one second lower than the lowest date in the test set
		{
			name: "failDateLower",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeDateCriteria("< 2019-12-31 23:59:59", false, false),
			},
			want: nil,
		},
		// Searches for an empty name
		{
			name: "failEmptyName",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeNameCriteria(`^\s+$`),
			},
			want: nil,
		},
		// Searches for a non existing name
		{
			name: "failWrongName",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeNameCriteria(`BilboBaggins`),
			},
			want: nil,
		},
		// Searches for non-existing IPs
		{
			name: "failWrongIp1",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`127.0.0.1`),
			},
			want: nil,
		},
		{
			name: "failWrongIp2",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`0.0.0.0`),
			},
			want: nil,
		},
		{
			name: "failWrongIp3",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeIpCriteria(`^255`),
			},
			want: nil,
		},
		// Searches for a non-existing metadata key
		{
			name: "failWrongMeta",
			args: args{
				qByM:      dummyQbyM,
				qWithM:    dummyQwithM,
				converter: getData,
				criteria:  makeMDCriteria(false, StringMap{"OneRing": `ToRuleThemAll`}),
			},
			want: nil,
		},
	}
	for _, tt := range testsSuccess {
		t.Run(tt.name, func(t *testing.T) {
			got, explanation, err := searchByFilter(tt.args.qByM, tt.args.qWithM, tt.args.converter, tt.args.queryType, tt.args.criteria)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchByFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("searchByFilter() got = %v, want %v", got, tt.want)
			}
			logVerbose(t, "%s\n", explanation)
		})
	}
	for _, tt := range testsFailure {
		t.Run(tt.name, func(t *testing.T) {
			got, explanation, err := searchByFilter(tt.args.qByM, tt.args.qWithM, tt.args.converter, tt.args.queryType, tt.args.criteria)
			if (err != nil) != tt.wantErr {
				t.Errorf("searchByFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			success := false
			if len(got) == 0 && len(tt.want) == 0 {
				success = true
			}
			if !success {
				t.Errorf("searchByFilter() got = %v, want %v", got, tt.want)
			}
			logVerbose(t, "%s\n", explanation)
		})
	}
}
