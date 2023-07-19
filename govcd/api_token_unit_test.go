//go:build unit || ALL

/*
* Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"reflect"
	"testing"
)

func Test_readFileAndUnmarshalJSON(t *testing.T) {
	type args struct {
		filename string
		object   *testEntity
	}
	tests := []struct {
		name    string
		args    args
		want    *testEntity
		wantErr bool
	}{
		{
			name: "simpleCase",
			args: args{
				filename: "test-resources/test.json",
				object:   &testEntity{},
			},
			want:    &testEntity{Name: "test"},
			wantErr: false,
		},
		{
			name: "emptyFile",
			args: args{
				filename: "test-resources/test_empty.json",
				object:   &testEntity{},
			},
			want:    &testEntity{},
			wantErr: true,
		},
		{
			name: "emptyJSON",
			args: args{
				filename: "test-resources/test_emptyJSON.json",
				object:   &testEntity{},
			},
			want:    &testEntity{},
			wantErr: false,
		},
		{
			name: "nonexistentFile",
			args: args{
				filename: "thisfiledoesntexist.json",
				object:   &testEntity{},
			},
			want:    &testEntity{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := readFileAndUnmarshalJSON(tt.args.filename, tt.args.object)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFileAndUnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.args.object, tt.want) {
				t.Errorf("readFileAndUnmarshalJSON() = %v, want %v", tt.args.object, tt.want)
			}
		})
	}
}
