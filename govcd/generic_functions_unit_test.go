//go:build unit || ALL

/*
* Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"reflect"
	"testing"
)

func Test_oneOrError(t *testing.T) {
	type args struct {
		key         string
		name        string
		entitySlice []*testEntity
	}
	tests := []struct {
		name                  string
		args                  args
		want                  *testEntity
		wantErr               bool
		wantErrEntityNotFound bool
	}{
		{
			name: "SingleEntity",
			args: args{
				key:         "name",
				name:        "test",
				entitySlice: []*testEntity{{Name: "test"}},
			},
			want:    &testEntity{Name: "test"},
			wantErr: false,
		},
		{
			name: "NoEntities",
			args: args{
				key:         "name",
				name:        "test",
				entitySlice: []*testEntity{},
			},
			want:                  nil,
			wantErr:               true,
			wantErrEntityNotFound: true,
		},
		{
			name: "TwoEntities",
			args: args{
				key:         "name",
				name:        "test",
				entitySlice: []*testEntity{{Name: "test"}, {Name: "best"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ThreeEntities",
			args: args{
				key:         "name",
				name:        "test",
				entitySlice: []*testEntity{{Name: "test"}, {Name: "best"}, {Name: "rest"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NilEntities",
			args: args{
				key:         "name",
				name:        "test",
				entitySlice: nil,
			},
			want:                  nil,
			wantErr:               true,
			wantErrEntityNotFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := oneOrError(tt.args.key, tt.args.name, tt.args.entitySlice)
			if (err != nil) != tt.wantErr {
				t.Errorf("oneOrError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrEntityNotFound && !ContainsNotFound(err) {
				t.Errorf("oneOrError() error = %v, wantErrEntityNotFound %v", err, tt.wantErrEntityNotFound)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("oneOrError() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testEntity struct {
	Name string `json:"name"`
}

// TODO Write tests for marshalJSONAndWriteFile
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
				filename: "testdata/test.json",
				object:   &testEntity{},
			},
			want:    &testEntity{Name: "test"},
			wantErr: false,
		},
		{
			name: "emptyFile",
			args: args{
				filename: "testdata/test_empty.json",
				object:   &testEntity{},
			},
			want:    &testEntity{},
			wantErr: true,
		},
		{
			name: "emptyJSON",
			args: args{
				filename: "testdata/test_emptyjson.json",
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
