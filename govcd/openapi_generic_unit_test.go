//go:build unit || ALL

package govcd

import (
	"testing"
)

func Test_urlFromEndpoint(t *testing.T) {
	type args struct {
		endpoint       string
		endpointParams []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "PlaceholderAndSuffix",
			args:    args{endpoint: "1.0.0/edgeGateways/%s/ipsec/tunnels/", endpointParams: []string{"edgeGatewayId", "suffix"}},
			want:    "1.0.0/edgeGateways/edgeGatewayId/ipsec/tunnels/suffix",
			wantErr: false,
		},
		{
			name:    "PlaceholderUnsatisfiedEmptySlice",
			args:    args{endpoint: "1.0.0/edgeGateways/%s/ipsec/tunnels/", endpointParams: []string{}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "PlaceholderUnsatisfiedNilSlice",
			args:    args{endpoint: "1.0.0/edgeGateways/%s/ipsec/tunnels/", endpointParams: nil},
			want:    "",
			wantErr: true,
		},
		{
			name:    "NoPlaceholderNilSlice",
			args:    args{endpoint: "1.0.0/edgeGateways/ipsec/tunnels/", endpointParams: nil},
			want:    "1.0.0/edgeGateways/ipsec/tunnels/",
			wantErr: false,
		},
		{
			name:    "NoPlaceholderEmptySlice",
			args:    args{endpoint: "1.0.0/edgeGateways/ipsec/tunnels/", endpointParams: []string{}},
			want:    "1.0.0/edgeGateways/ipsec/tunnels/",
			wantErr: false,
		},
		{
			name:    "InsufficientPlaceholderReplacements",
			args:    args{endpoint: "1.0.0/edgeGateways/%s/ipsec/%s/tunnels/", endpointParams: []string{"replacement-one"}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "MultipleSuffixes",
			args:    args{endpoint: "1.0.0/edgeGateways/%s/ipsec/%s/tunnels/", endpointParams: []string{"replacement-one", "replacement-two", "suffix-one", "/", "suffix-two"}},
			want:    "1.0.0/edgeGateways/replacement-one/ipsec/replacement-two/tunnels/suffix-one/suffix-two",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := urlFromEndpoint(tt.args.endpoint, tt.args.endpointParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("urlFromEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("urlFromEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
