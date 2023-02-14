package types

import "testing"

func TestOpenAPIEdgeGateway_DeallocateIpCount(t *testing.T) {
	type fields struct {
		Status                    string
		ID                        string
		Name                      string
		Description               string
		OwnerRef                  *OpenApiReference
		OrgVdc                    *OpenApiReference
		Org                       *OpenApiReference
		EdgeGatewayUplinks        []EdgeGatewayUplinks
		DistributedRoutingEnabled *bool
		EdgeClusterConfig         *OpenAPIEdgeGatewayEdgeClusterConfig
		OrgVdcNetworkCount        *int
		GatewayBacking            *OpenAPIEdgeGatewayBacking
		ServiceNetworkDefinition  string
		UsingIpSpace              *bool
	}
	type args struct {
		ipCount int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			egw := &OpenAPIEdgeGateway{
				Status:                    tt.fields.Status,
				ID:                        tt.fields.ID,
				Name:                      tt.fields.Name,
				Description:               tt.fields.Description,
				OwnerRef:                  tt.fields.OwnerRef,
				OrgVdc:                    tt.fields.OrgVdc,
				Org:                       tt.fields.Org,
				EdgeGatewayUplinks:        tt.fields.EdgeGatewayUplinks,
				DistributedRoutingEnabled: tt.fields.DistributedRoutingEnabled,
				EdgeClusterConfig:         tt.fields.EdgeClusterConfig,
				OrgVdcNetworkCount:        tt.fields.OrgVdcNetworkCount,
				GatewayBacking:            tt.fields.GatewayBacking,
				ServiceNetworkDefinition:  tt.fields.ServiceNetworkDefinition,
				UsingIpSpace:              tt.fields.UsingIpSpace,
			}
			if err := egw.DeallocateIpCount(tt.args.ipCount); (err != nil) != tt.wantErr {
				t.Errorf("OpenAPIEdgeGateway.DeallocateIpCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
