//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"reflect"
	"testing"
)

func Test_normaliseOpenApiMetadata(t *testing.T) {
	type args struct {
		objectType    string
		name          string
		metadataEntry *types.OpenApiMetadataEntry
	}
	tests := []struct {
		name    string
		args    args
		want    *normalisedMetadata
		wantErr bool
	}{
		{
			name: "number normalised",
			args: args{
				objectType: "entity",
				name:       "foo",
				metadataEntry: &types.OpenApiMetadataEntry{
					KeyValue: types.OpenApiMetadataKeyValue{
						Domain: "TENANT",
						Key:    "key",
						Value: types.OpenApiMetadataTypedValue{
							Value: 314159,
							Type:  types.OpenApiMetadataNumberEntry,
						},
					},
				},
			},
			want: &normalisedMetadata{
				ObjectType: "entity",
				ObjectName: "foo",
				Key:        "key",
				Value:      "314159",
			},
		},
		{
			name: "string normalised",
			args: args{
				objectType: "entity",
				name:       "foo",
				metadataEntry: &types.OpenApiMetadataEntry{
					KeyValue: types.OpenApiMetadataKeyValue{
						Domain: "TENANT",
						Key:    "key",
						Value: types.OpenApiMetadataTypedValue{
							Value: "value",
							Type:  types.OpenApiMetadataStringEntry,
						},
					},
				},
			},
			want: &normalisedMetadata{
				ObjectType: "entity",
				ObjectName: "foo",
				Key:        "key",
				Value:      "value",
			},
		},
		{
			name: "bool normalised",
			args: args{
				objectType: "entity",
				name:       "foo",
				metadataEntry: &types.OpenApiMetadataEntry{
					KeyValue: types.OpenApiMetadataKeyValue{
						Domain: "TENANT",
						Key:    "key",
						Value: types.OpenApiMetadataTypedValue{
							Value: true,
							Type:  types.OpenApiMetadataBooleanEntry,
						},
					},
				},
			},
			want: &normalisedMetadata{
				ObjectType: "entity",
				ObjectName: "foo",
				Key:        "key",
				Value:      "true",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normaliseOpenApiMetadata(tt.args.objectType, tt.args.name, tt.args.metadataEntry)
			if (err != nil) != tt.wantErr {
				t.Errorf("normaliseOpenApiMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normaliseOpenApiMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}
