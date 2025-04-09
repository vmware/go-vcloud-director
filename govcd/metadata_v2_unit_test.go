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

func Test_normaliseXmlMetadata(t *testing.T) {
	type args struct {
		key           string
		href          string
		objectName    string
		metadataEntry *types.MetadataValue
	}
	tests := []struct {
		name    string
		args    args
		want    *normalisedMetadata
		wantErr bool
	}{
		{
			name: "string normalised",
			args: args{
				key:        "key",
				objectName: "foo",
				href:       "/admin/catalog/67e119b7-083b-349e-8dfd-6cf0c19b83cf",
				metadataEntry: &types.MetadataValue{
					TypedValue: &types.MetadataTypedValue{
						XsiType: types.MetadataStringValue,
						Value:   "value",
					},
				},
			},
			want: &normalisedMetadata{
				ObjectType: "catalog",
				ObjectName: "foo",
				Key:        "key",
				Value:      "value",
			},
		},
		{
			name: "bool normalised",
			args: args{
				key:        "key",
				objectName: "foo",
				href:       "/admin/catalog/67e119b7-083b-349e-8dfd-6cf0c19b83cf",
				metadataEntry: &types.MetadataValue{
					TypedValue: &types.MetadataTypedValue{
						XsiType: types.MetadataBooleanValue,
						Value:   "true",
					},
				},
			},
			want: &normalisedMetadata{
				ObjectType: "catalog",
				ObjectName: "foo",
				Key:        "key",
				Value:      "true",
			},
		},
		{
			name: "number normalised",
			args: args{
				key:        "key",
				objectName: "foo",
				href:       "/admin/catalog/67e119b7-083b-349e-8dfd-6cf0c19b83cf",
				metadataEntry: &types.MetadataValue{
					TypedValue: &types.MetadataTypedValue{
						XsiType: types.MetadataNumberValue,
						Value:   "314159",
					},
				},
			},
			want: &normalisedMetadata{
				ObjectType: "catalog",
				ObjectName: "foo",
				Key:        "key",
				Value:      "314159",
			},
		},
		{
			name: "date normalised",
			args: args{
				key:        "key",
				objectName: "foo",
				href:       "/admin/catalog/67e119b7-083b-349e-8dfd-6cf0c19b83cf",
				metadataEntry: &types.MetadataValue{
					TypedValue: &types.MetadataTypedValue{
						XsiType: types.MetadataDateTimeValue,
						Value:   "2023-11-16T09:56:00.000Z",
					},
				},
			},
			want: &normalisedMetadata{
				ObjectType: "catalog",
				ObjectName: "foo",
				Key:        "key",
				Value:      "2023-11-16T09:56:00.000Z",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normaliseXmlMetadata(tt.args.key, tt.args.href, tt.args.objectName, tt.args.metadataEntry)
			if (err != nil) != tt.wantErr {
				t.Errorf("normaliseXmlMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normaliseXmlMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}
