//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

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
