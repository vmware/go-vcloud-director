//go:build unit || lb || lbAppProfile || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import "testing"

func Test_extractNSXObjectIDFromPath(t *testing.T) {
	type args struct {
		header string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Empty",
			args:    args{header: ""},
			want:    "",
			wantErr: true,
		},
		{
			name:    "No URL in Path",
			args:    args{header: "invalid_location_header"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "No Slash",
			args:    args{header: "applicationProfile-4"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Trailing Slash",
			args:    args{header: "/network/edges/edge-3/loadbalancer/config/applicationprofiles/applicationProfile-4/"},
			want:    "applicationProfile-4",
			wantErr: false,
		},
		{
			name:    "No Trailing Slash",
			args:    args{header: "/network/edges/edge-3/loadbalancer/config/applicationprofiles/applicationProfile-4"},
			want:    "applicationProfile-4",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractNsxObjectIdFromPath(tt.args.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractNsxObjectIdFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractNsxObjectIdFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
