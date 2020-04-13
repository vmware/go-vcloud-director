// +build unit

package govcd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

// Test filter engine using dependency injection

// Mocking object implementing QueryItem interface
type TestItem struct {
	name     string
	date     string
	ip       string
	ttype    string
	metadata stringMap
}

// mock QueryWithMetadataFields function that returns an empty result
var dummyQwithM = func(queryType string, params, notEncodedParams map[string]string,
	metadataFields []string, isSystem bool) (Results, error) {
	return Results{}, nil
}

// mock QueryByMetadataFields function that returns an empty result
var dummyQbyM = func(params, notEncodedParams map[string]string,
	metadataFilters map[string]MetadataFilter, isSystem bool) (Results, error) {
	return Results{}, nil
}

func (t TestItem) GetDate() string                    { return t.date }
func (t TestItem) GetName() string                    { return t.name }
func (t TestItem) GetType() string                    { return t.ttype }
func (t TestItem) GetIp() string                      { return t.ip }
func (t TestItem) GetMetadataValue(key string) string { return t.metadata[key] }
func (t TestItem) GetHref() string                    { return "" }

func makeNameCriteria(name string) *FilterDef {
	f := NewFilterDef()
	_ = f.AddFilter(FilterNameRegex, name)
	return f
}

func makeIpCriteria(ip string) *FilterDef {
	f := NewFilterDef()
	_ = f.AddFilter(FilterIp, ip)
	return f
}

func makeDateCriteria(date string, latest, earliest bool) *FilterDef {
	f := NewFilterDef()
	if date != "" {
		_ = f.AddFilter(FilterDate, date)
	}
	if earliest {
		_ = f.AddFilter(FilterEarliest, "true")
	}
	if latest {
		_ = f.AddFilter(FilterLatest, "true")
	}
	return f
}

func makeMDCriteria(useApi bool, wanted stringMap) *FilterDef {
	f := NewFilterDef()
	for key, value := range wanted {
		_ = f.AddMetadataFilter(key, value, "STRING", false, useApi)
	}
	return f
}

func makeDateMDCriteria(date string, useApi bool, wanted stringMap) *FilterDef {
	f := makeMDCriteria(useApi, wanted)
	_ = f.AddFilter(FilterDate, date)
	fmt.Printf("%# v\n", pretty.Formatter(f))
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
			metadata: stringMap{"xyz": "xxx"},
		},
		TestItem{
			name:     "two",
			ip:       "10.10.8.9",
			date:     "2020-02-01",
			ttype:    "test-name",
			metadata: stringMap{"abc": "yyy"},
		},
		TestItem{
			name:     "three",
			ip:       "10.150.20.1",
			date:     "2020-02-03 10:00:00",
			ttype:    "test-date",
			metadata: stringMap{"abc": "-+!", "other": "aaa"},
		},
	}

	// Injection function that replaces the resultConverterFunc and returns the test data
	var getData = func(string, Results) ([]QueryItem, error) {
		return data, nil
	}

	// Test cases that should succeed
	testsSuccess := []testDef{
		// Gets a name by regular expression (finds 1)
		{"name1", args{dummyQbyM, dummyQwithM, getData, "", makeNameCriteria(`o\w+`)}, []QueryItem{data[0]}, false},
		// Gets a name by regular expression (finds 2)
		{"name2", args{dummyQbyM, dummyQwithM, getData, "", makeNameCriteria(`t\w+`)}, []QueryItem{data[1], data[2]}, false},
		// Gets a name by full value (finds 1)
		{"name3", args{dummyQbyM, dummyQwithM, getData, "", makeNameCriteria(`two`)}, []QueryItem{data[1]}, false},
		// Use date comparison to get an item that's a few hours older than the date referenced (finds 1)
		{"dateComparison", args{dummyQbyM, dummyQwithM, getData, "", makeDateCriteria("> 2020-02-03", false, false)}, []QueryItem{data[2]}, false},
		// Gets the newest element (finds 1)
		{"dateLatest", args{dummyQbyM, dummyQwithM, getData, "", makeDateCriteria("", true, false)}, []QueryItem{data[2]}, false},
		// Gets the oldest element (finds 1)
		{"dateEarliest", args{dummyQbyM, dummyQwithM, getData, "", makeDateCriteria("", false, true)}, []QueryItem{data[0]}, false},
		// Gets the items with IP containing "10" (finds 3)
		{"ip10", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`10`)}, []QueryItem{data[0], data[1], data[2]}, false},
		// Gets the items with IP starting by 10 (finds 2)
		{"ip^10", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`^10`)}, []QueryItem{data[1], data[2]}, false},
		// Gets the items with IP ending with "10" (finds 1)
		{"ip10$", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`10$`)}, []QueryItem{data[0]}, false},
		// Gets the items with IP starting by 192 (finds 1)
		{"ip192", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`^192`)}, []QueryItem{data[0]}, false},
		// Gets the items with IP starting by 10.150 (finds 1)
		{"ip10.150", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`10\.150`)}, []QueryItem{data[2]}, false},
		// Gets the items with metadata "abc" and any value (finds 2)
		{"metaAbc1", args{dummyQbyM, dummyQwithM, getData, "", makeMDCriteria(false, stringMap{"abc": `\S+`})}, []QueryItem{data[1], data[2]}, false},
		// Gets the items with metadata "abc" and alphanumeric value (finds 1)
		{"metaAbc2", args{dummyQbyM, dummyQwithM, getData, "", makeMDCriteria(false, stringMap{"abc": `\w+`})}, []QueryItem{data[1]}, false},
		// Gets the items with metadata "xyz" and alphanumeric value (finds 1)
		{"metaXyz", args{dummyQbyM, dummyQwithM, getData, "", makeMDCriteria(false, stringMap{"xyz": `\w+`})}, []QueryItem{data[0]}, false},
		// Gets the items with two metadata values (finds 1)
		{"metaOther", args{dummyQbyM, dummyQwithM, getData, "", makeMDCriteria(false, stringMap{"abc": `\S+`, "other": `\w+`})}, []QueryItem{data[2]}, false},
		// Gets the items with metadata "abc" and any value combined with date search (finds 1)
		{"metaCombined", args{dummyQbyM, dummyQwithM, getData, "", makeDateMDCriteria("> 2020-02-02", false, stringMap{"abc": `\S+`})}, []QueryItem{data[2]}, false},
	}

	// Test cases that should fail, i.e. will return an empty list
	testsFailure := []testDef{
		// Searches for a date that is one second higher than the highest date in the test set
		{"failDateHigher", args{dummyQbyM, dummyQwithM, getData, "", makeDateCriteria(">= 2020-02-03 10:00:01", false, false)}, []QueryItem{}, false},
		// Searches for a date that is one second lower than the lowest date in the test set
		{"failDateLower", args{dummyQbyM, dummyQwithM, getData, "", makeDateCriteria("< 2019-12-31 23:59:59", false, false)}, []QueryItem{}, false},
		// Searches for an empty name
		{"failEmptyName", args{dummyQbyM, dummyQwithM, getData, "", makeNameCriteria(`^\s+$`)}, []QueryItem{}, false},
		// Searches for a non existing name
		{"failWrongName", args{dummyQbyM, dummyQwithM, getData, "", makeNameCriteria(`BilboBaggins`)}, []QueryItem{}, false},
		// Searches for non-existing IPs
		{"failWrongIp1", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`127.0.0.1`)}, []QueryItem{}, false},
		{"failWrongIp2", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`0.0.0.0`)}, []QueryItem{}, false},
		{"failWrongIp3", args{dummyQbyM, dummyQwithM, getData, "", makeIpCriteria(`^255`)}, []QueryItem{}, false},
		// Searches for a non-existing metadata key
		{"failWrongMeta", args{dummyQbyM, dummyQwithM, getData, "", makeMDCriteria(false, stringMap{"OneRing": `ToRuleThemAll`})}, []QueryItem{}, false},
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
			t.Logf("%s\n", explanation)
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
			t.Logf("%s\n", explanation)
		})
	}
}
