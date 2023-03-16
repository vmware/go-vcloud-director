//go:build unit || ALL

package govcd

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_defaultPageSize(t *testing.T) {
	type args struct {
		queryParams     url.Values
		defaultPageSize string
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{
			name: "NilQueryParams",
			args: args{nil, "128"},
			want: map[string][]string{"pageSize": []string{"128"}},
		},
		{
			name: "NotNilQueryParams",
			args: args{map[string][]string{"otherField": []string{"randomValue"}}, "128"},
			want: map[string][]string{"pageSize": []string{"128"}, "otherField": []string{"randomValue"}},
		},
		{
			name: "CustomPageSize",
			args: args{map[string][]string{"pageSize": []string{"1"}}, "128"},
			want: map[string][]string{"pageSize": []string{"1"}},
		},
		{
			name: "CustomPageSizeWithOtherFields",
			args: args{map[string][]string{"pageSize": []string{"1"}, "otherField": []string{"randomValue"}}, "128"},
			want: map[string][]string{"pageSize": []string{"1"}, "otherField": []string{"randomValue"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultPageSize(tt.args.queryParams, tt.args.defaultPageSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultPageSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
