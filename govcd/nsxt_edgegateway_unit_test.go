//go:build unit || ALL

/*
* Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"net/netip"
	"reflect"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func Test_filterIpSlicesBySubnet(t *testing.T) {
	type args struct {
		ipRange []netip.Addr
		subnet  netip.Prefix
	}
	tests := []struct {
		name    string
		args    args
		want    []netip.Addr
		wantErr bool
	}{
		{name: "BothArgsEmpty", args: args{}, want: nil, wantErr: true},
		{name: "EmptyRange", args: args{subnet: netip.MustParsePrefix("10.10.10.1/24")}, want: nil, wantErr: true},
		{name: "EmptySubnet", args: args{ipRange: []netip.Addr{netip.MustParseAddr("10.1.1.1")}}, want: nil, wantErr: true},
		{
			name: "SingleIpMatchingSubnet",
			args: args{
				ipRange: []netip.Addr{netip.MustParseAddr("10.0.0.2")},
				subnet:  netip.MustParsePrefix("10.0.0.1/24"),
			},
			want:    []netip.Addr{netip.MustParseAddr("10.0.0.2")},
			wantErr: false,
		},
		{
			name: "SingleIpNotMatchingSubnet",
			args: args{
				ipRange: []netip.Addr{netip.MustParseAddr("10.0.0.2")},
				subnet:  netip.MustParsePrefix("20.0.0.1/24"),
			},
			want:    []netip.Addr{},
			wantErr: false,
		},
		{
			name: "ManyIPsSomeMatch",
			args: args{
				ipRange: []netip.Addr{
					netip.MustParseAddr("10.0.0.2"),
					netip.MustParseAddr("192.0.0.2"),
					netip.MustParseAddr("11.0.0.2"),
					netip.MustParseAddr("20.0.0.2"),
					netip.MustParseAddr("20.0.0.3"),
					netip.MustParseAddr("10.0.0.2"),
				},
				subnet: netip.MustParsePrefix("20.0.0.1/24"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("20.0.0.2"),
				netip.MustParseAddr("20.0.0.3"),
			},
			wantErr: false,
		},
		{
			name: "DuplicateIPsInRange",
			args: args{
				ipRange: []netip.Addr{
					netip.MustParseAddr("10.0.0.2"),
					netip.MustParseAddr("10.0.0.2"),
					netip.MustParseAddr("192.0.0.2"),
					netip.MustParseAddr("11.0.0.2"),
					netip.MustParseAddr("20.0.0.2"),
					netip.MustParseAddr("20.0.0.3"),
					netip.MustParseAddr("20.0.0.3"),
					netip.MustParseAddr("10.0.0.2"),
				},
				subnet: netip.MustParsePrefix("20.0.0.1/24"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("20.0.0.2"),
				netip.MustParseAddr("20.0.0.3"),
				netip.MustParseAddr("20.0.0.3"),
			},
			wantErr: false,
		},
		// IPv6
		{
			name: "IPv6SingleMatchingSubnet",
			args: args{
				ipRange: []netip.Addr{netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0001")},
				subnet:  netip.MustParsePrefix("2001:0DB8:0000:000b::/64"),
			},
			want:    []netip.Addr{netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0001")},
			wantErr: false,
		},
		{
			name: "IPv6SingleNotMatchingSubnet",
			args: args{
				ipRange: []netip.Addr{netip.MustParseAddr("2001:0DB6:0000:000b:0000:0000:0000:0001")},
				subnet:  netip.MustParsePrefix("2001:0DB8:0000:000b::/64"),
			},
			want:    []netip.Addr{},
			wantErr: false,
		},
		{
			name: "IPv6ManyIPsSomeMatch",
			args: args{
				ipRange: []netip.Addr{
					netip.MustParseAddr("2001:1111:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("2222:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0002"),
					netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0003"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
				},
				subnet: netip.MustParsePrefix("2001:0DB8:0000:000b::/64"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0001"),
				netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0002"),
				netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0003"),
			},
			wantErr: false,
		},
		{
			name: "IPv6ManyIPsSomeDuplicatesMatch",
			args: args{
				ipRange: []netip.Addr{
					netip.MustParseAddr("2001:1111:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("2222:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0002"),
					netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0002"),
					netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0003"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
				},
				subnet: netip.MustParsePrefix("2001:0DB8:0000:000b::/64"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0001"),
				netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0002"),
				netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0002"),
				netip.MustParseAddr("2001:0DB8:0000:000b:0000:0000:0000:0003"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterIpSlicesBySubnet(tt.args.ipRange, tt.args.subnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterIpRangesInSubnet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterIpRangesInSubnet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ipSliceDifference(t *testing.T) {
	type args struct {
		minuendSlice    []netip.Addr
		subtrahendSlice []netip.Addr
	}
	tests := []struct {
		name string
		args args
		want []netip.Addr
	}{
		{
			name: "BothParamsNil",
			args: args{
				minuendSlice:    nil,
				subtrahendSlice: nil,
			},
			want: nil,
		},
		{
			name: "MinuendNil",
			args: args{
				minuendSlice: nil,
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
				},
			},
			want: nil,
		},
		{
			name: "MinuendEmptySliceNilSubtrahend",
			args: args{
				minuendSlice:    make([]netip.Addr, 0),
				subtrahendSlice: nil,
			},
			want: make([]netip.Addr, 0),
		},
		{
			name: "MinuendEmptySlice",
			args: args{
				minuendSlice: []netip.Addr{{}},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
				},
			},
			want: []netip.Addr{{}},
		},
		{
			name: "SubtrahendNil",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
				},
				subtrahendSlice: nil,
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
			},
		},
		{
			name: "SubtractUnavailableIP",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
					netip.MustParseAddr("10.0.0.2"),
				},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("20.0.0.1"),
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
				netip.MustParseAddr("10.0.0.2"),
			},
		},
		{
			name: "SubtractIP",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
					netip.MustParseAddr("10.0.0.2"),
				},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.2"),
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
			},
		},
		{
			name: "RemoveAll",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
					netip.MustParseAddr("10.0.0.2"),
				},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
					netip.MustParseAddr("10.0.0.2"),
				},
			},
			want: nil,
		},
		{
			name: "SubtractIPWithDuplicates",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.1"),
					netip.MustParseAddr("10.0.0.2"),
					netip.MustParseAddr("10.0.0.2"),
				},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("10.0.0.2"),
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
			},
		},
		// IPv6
		{
			name: "IPv6MinuendNil",
			args: args{
				minuendSlice: nil,
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
				},
			},
			want: nil,
		},
		{
			name: "IPv6SubtrahendNil",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
				},
				subtrahendSlice: nil,
			},
			want: []netip.Addr{
				netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
			},
		},
		{
			name: "IPv6SubtractUnavailableIP",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0002"),
				},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("9001:0DB8:0000:000b:0000:0000:0000:0002"),
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
				netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0002"),
			},
		},
		{
			name: "IPv6SubtractIP",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0002"),
				},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0002"),
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
			},
		},
		{
			name: "IPv6SubtractIPWithDuplicates",
			args: args{
				minuendSlice: []netip.Addr{
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0002"),
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0002"),
				},
				subtrahendSlice: []netip.Addr{
					netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0002"),
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ipSliceDifference(tt.args.minuendSlice, tt.args.subtrahendSlice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ipRangeDifference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_flattenEdgeGatewayUplinkToIpSlice(t *testing.T) {
	type args struct {
		uplinks []types.EdgeGatewayUplinks
	}
	tests := []struct {
		name    string
		args    args
		want    []netip.Addr
		wantErr bool
	}{
		{
			name: "SingleStartAndEndAddresses",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.1"),
				netip.MustParseAddr("10.10.10.2"),
			},
			wantErr: false,
		},
		{
			name: "ReverseStartAndEnd",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.2",
												EndAddress:   "10.10.10.1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SameStartAndEndAddresses",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.1"),
			},
			wantErr: false,
		},
		{
			name: "StartAddressOnly",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.1"),
			},
			wantErr: false,
		},
		{
			name: "EmptyUplink",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{},
				},
			},
			want:    make([]netip.Addr, 0),
			wantErr: false,
		},
		{
			name: "EmptySubnets",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{},
					},
				},
			},
			want:    make([]netip.Addr, 0),
			wantErr: false,
		},
		{
			name: "EmptySubnetValues",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{},
						},
					},
				},
			},
			want:    make([]netip.Addr, 0),
			wantErr: false,
		},
		{
			name: "EmptySubnetValueIpRanges",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{},
								},
							},
						},
					},
				},
			},
			want:    make([]netip.Addr, 0),
			wantErr: false,
		},
		{
			name: "EmptySubnetValueIpRangeValues",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{},
									},
								},
							},
						},
					},
				},
			},
			want:    make([]netip.Addr, 0),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenEdgeGatewayUplinkToIpSlice(tt.args.uplinks)
			if (err != nil) != tt.wantErr {
				t.Errorf("ipSliceFromEdgeGatewayUplinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ipSliceFromEdgeGatewayUplinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

// buildSimpleUplinkStructure helps to avoid deep indentation in Test_getUnusedExternalIPAddress
// where the structure itself is simple enough that has only one subnet and one IP range. Other
// tests in this table test still contain the full structure as it would be less readable if it was
// wrapped into multiple function calls.
func buildSimpleUplinkStructure(ipRangeValues []types.OpenApiIPRangeValues) []types.EdgeGatewayUplinks {
	return []types.EdgeGatewayUplinks{
		{
			Subnets: types.OpenAPIEdgeGatewaySubnets{
				Values: []types.OpenAPIEdgeGatewaySubnetValue{
					{
						IPRanges: &types.OpenApiIPRanges{
							Values: ipRangeValues,
						},
					},
				},
			},
		},
	}
}

func Test_getUnusedExternalIPAddress(t *testing.T) {
	type args struct {
		uplinks         []types.EdgeGatewayUplinks
		usedIpAddresses []*types.GatewayUsedIpAddress
		requiredCount   int
		optionalSubnet  netip.Prefix
	}
	tests := []struct {
		name    string
		args    args
		want    []netip.Addr
		wantErr bool
	}{
		{
			name: "EmptyStructureError",
			args: args{
				uplinks:         []types.EdgeGatewayUplinks{},
				usedIpAddresses: []*types.GatewayUsedIpAddress{{}},
				requiredCount:   1,
				optionalSubnet:  netip.Prefix{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SingleIpAvailable",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.1",
						EndAddress:   "10.10.10.1",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   1,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.1"),
			},
			wantErr: false,
		},
		{
			name: "AvailableIPsFilteredOff",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.1",
						EndAddress:   "10.10.10.10",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   1,
				optionalSubnet:  netip.MustParsePrefix("20.10.10.0/24"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "AvailableIPsFilteredOffAndUsed",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.1",
						EndAddress:   "10.10.10.10",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{{IPAddress: "10.10.10.1"}},
				requiredCount:   1,
				optionalSubnet:  netip.MustParsePrefix("20.10.10.0/24"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SingleIpFromMany",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.15",
						EndAddress:   "10.10.10.200",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   1,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.15"),
			},
			wantErr: false,
		},
		{
			name: "CrossBoundary",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.255",
						EndAddress:   "10.10.11.1",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   3,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.255"),
				netip.MustParseAddr("10.10.11.0"),
				netip.MustParseAddr("10.10.11.1"),
			},
			wantErr: false,
		},
		{
			name: "CrossBoundaryPrefix",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.255",
						EndAddress:   "10.10.11.1",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   2,
				optionalSubnet:  netip.MustParsePrefix("10.10.11.0/24"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.11.0"),
				netip.MustParseAddr("10.10.11.1"),
			},
			wantErr: false,
		},
		{
			name: "CrossBoundaryPrefixAndUsed",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.255",
						EndAddress:   "10.10.11.1",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{{IPAddress: "10.10.11.0"}},
				requiredCount:   1,
				optionalSubnet:  netip.MustParsePrefix("10.10.11.0/24"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.11.1"),
			},
			wantErr: false,
		},
		{
			name: "IPv6SingleIpAvailable",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "4001:0DB8:0000:000b:0000:0000:0000:0001",
						EndAddress:   "4001:0DB8:0000:000b:0000:0000:0000:0001",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   1,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
			},
			wantErr: false,
		},
		{
			name: "SingleIpAvailableStartOnly",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.1",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   1,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.1"),
			},
			wantErr: false,
		},
		{
			name: "IPv6SingleIpAvailableStartOnly",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "4001:0DB8:0000:000b:0000:0000:0000:0001",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   1,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("4001:0DB8:0000:000b:0000:0000:0000:0001"),
			},
			wantErr: false,
		},
		{
			name: "InvalidIpRange",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						// Start Address is higher than end IP address
						StartAddress: "10.10.10.200",
						EndAddress:   "10.10.10.1",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   1,
				optionalSubnet:  netip.Prefix{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "InsufficientIPs",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.1",
						EndAddress:   "10.10.10.6",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   7,
				optionalSubnet:  netip.Prefix{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "InsufficientIPsWithUsed",
			args: args{
				uplinks: buildSimpleUplinkStructure([]types.OpenApiIPRangeValues{
					{
						StartAddress: "10.10.10.1",
						EndAddress:   "10.10.10.6",
					},
				}),
				usedIpAddresses: []*types.GatewayUsedIpAddress{
					{IPAddress: "10.10.10.1"},
				},
				requiredCount:  6,
				optionalSubnet: netip.Prefix{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "MultipleUplinks",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.6",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "20.10.10.1",
												EndAddress:   "20.10.10.6",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "30.10.10.1",
												EndAddress:   "30.10.10.6",
											},
										},
									},
								},
							},
						},
					},
				},
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   18,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.1"),
				netip.MustParseAddr("10.10.10.2"),
				netip.MustParseAddr("10.10.10.3"),
				netip.MustParseAddr("10.10.10.4"),
				netip.MustParseAddr("10.10.10.5"),
				netip.MustParseAddr("10.10.10.6"),
				netip.MustParseAddr("20.10.10.1"),
				netip.MustParseAddr("20.10.10.2"),
				netip.MustParseAddr("20.10.10.3"),
				netip.MustParseAddr("20.10.10.4"),
				netip.MustParseAddr("20.10.10.5"),
				netip.MustParseAddr("20.10.10.6"),
				netip.MustParseAddr("30.10.10.1"),
				netip.MustParseAddr("30.10.10.2"),
				netip.MustParseAddr("30.10.10.3"),
				netip.MustParseAddr("30.10.10.4"),
				netip.MustParseAddr("30.10.10.5"),
				netip.MustParseAddr("30.10.10.6"),
			},
			wantErr: false,
		},
		{
			name: "MultipleUplinksWithUsedIPsInsufficient",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.6",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "20.10.10.1",
												EndAddress:   "20.10.10.6",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "30.10.10.1",
												EndAddress:   "30.10.10.6",
											},
										},
									},
								},
							},
						},
					},
				},
				usedIpAddresses: []*types.GatewayUsedIpAddress{
					{IPAddress: "10.10.10.1"},
				},
				requiredCount:  18,
				optionalSubnet: netip.Prefix{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "MultipleUplinksAndRanges",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
											{
												StartAddress: "10.10.10.10",
												EndAddress:   "10.10.10.12",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "20.10.10.1",
												EndAddress:   "20.10.10.6",
											},
											{
												StartAddress: "20.10.10.200",
												EndAddress:   "20.10.10.201",
											},
											{
												StartAddress: "20.10.10.251",
												EndAddress:   "20.10.10.252",
											},
											{
												StartAddress: "20.10.10.255",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "30.10.10.1",
												EndAddress:   "30.10.10.2",
											},
										},
									},
								},
							},
						},
					},
				},
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   18,
				optionalSubnet:  netip.Prefix{},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.10.10.1"),
				netip.MustParseAddr("10.10.10.2"),
				netip.MustParseAddr("10.10.10.10"),
				netip.MustParseAddr("10.10.10.11"),
				netip.MustParseAddr("10.10.10.12"),
				netip.MustParseAddr("20.10.10.1"),
				netip.MustParseAddr("20.10.10.2"),
				netip.MustParseAddr("20.10.10.3"),
				netip.MustParseAddr("20.10.10.4"),
				netip.MustParseAddr("20.10.10.5"),
				netip.MustParseAddr("20.10.10.6"),
				netip.MustParseAddr("20.10.10.200"),
				netip.MustParseAddr("20.10.10.201"),
				netip.MustParseAddr("20.10.10.251"),
				netip.MustParseAddr("20.10.10.252"),
				netip.MustParseAddr("20.10.10.255"),
				netip.MustParseAddr("30.10.10.1"),
				netip.MustParseAddr("30.10.10.2"),
			},
			wantErr: false,
		},
		{
			name: "MultipleUplinksAndRangesInsufficientIPs",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
											{
												StartAddress: "10.10.10.10",
												EndAddress:   "10.10.10.12",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "20.10.10.1",
												EndAddress:   "20.10.10.6",
											},
											{
												StartAddress: "20.10.10.200",
												EndAddress:   "20.10.10.201",
											},
											{
												StartAddress: "20.10.10.251",
												EndAddress:   "20.10.10.252",
											},
											{
												StartAddress: "20.10.10.255",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "30.10.10.1",
												EndAddress:   "30.10.10.2",
											},
										},
									},
								},
							},
						},
					},
				},
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   25,
				optionalSubnet:  netip.Prefix{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "MultipleUplinksAndRangesInSubnet24",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
											{
												StartAddress: "10.10.10.10",
												EndAddress:   "10.10.10.12",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "20.10.10.1",
												EndAddress:   "20.10.10.6",
											},
											{
												StartAddress: "20.10.10.200",
												EndAddress:   "20.10.10.201",
											},
											{
												StartAddress: "20.10.10.251",
												EndAddress:   "20.10.10.252",
											},
											{
												StartAddress: "20.10.10.255",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "30.10.10.1",
												EndAddress:   "30.10.10.2",
											},
										},
									},
								},
							},
						},
					},
				},
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   2,
				optionalSubnet:  netip.MustParsePrefix("30.10.10.1/24"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("30.10.10.1"),
				netip.MustParseAddr("30.10.10.2"),
			},
			wantErr: false,
		},
		{
			name: "MultipleUplinksAndRangesInSubnet28",
			args: args{
				uplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
											{
												StartAddress: "10.10.10.10",
												EndAddress:   "10.10.10.12",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "20.10.10.1",
												EndAddress:   "20.10.10.6",
											},
											{
												StartAddress: "20.10.10.200",
												EndAddress:   "20.10.10.201",
											},
											{
												StartAddress: "20.10.10.251",
												EndAddress:   "20.10.10.252",
											},
											{
												StartAddress: "20.10.10.255",
											},
										},
									},
								},
							},
						},
					},
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "30.10.10.1",
												EndAddress:   "30.10.10.2",
											},
										},
									},
								},
							},
						},
					},
				},
				usedIpAddresses: []*types.GatewayUsedIpAddress{},
				requiredCount:   6,
				optionalSubnet:  netip.MustParsePrefix("20.10.10.1/28"),
			},
			want: []netip.Addr{
				netip.MustParseAddr("20.10.10.1"),
				netip.MustParseAddr("20.10.10.2"),
				netip.MustParseAddr("20.10.10.3"),
				netip.MustParseAddr("20.10.10.4"),
				netip.MustParseAddr("20.10.10.5"),
				netip.MustParseAddr("20.10.10.6"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUnusedExternalIPAddress(tt.args.uplinks, tt.args.usedIpAddresses, tt.args.requiredCount, tt.args.optionalSubnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUnusedExternalIPAddress() error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("getUnusedExternalIPAddress() = %v, want %v", got, tt.want)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUnusedExternalIPAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_flattenGatewayUsedIpAddressesToIpSlice(t *testing.T) {
	type args struct {
		usedIpAddresses []*types.GatewayUsedIpAddress
	}
	tests := []struct {
		name    string
		args    args
		want    []netip.Addr
		wantErr bool
	}{
		{
			name:    "SingleIP",
			args:    args{usedIpAddresses: []*types.GatewayUsedIpAddress{{IPAddress: "10.0.0.1"}}},
			want:    []netip.Addr{netip.MustParseAddr("10.0.0.1")},
			wantErr: false,
		},
		{
			name: "DuplicateIPs",
			args: args{
				usedIpAddresses: []*types.GatewayUsedIpAddress{
					{IPAddress: "10.0.0.1"},
					{IPAddress: "10.0.0.1"},
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
				netip.MustParseAddr("10.0.0.1"),
			},
			wantErr: false,
		},
		{
			name:    "NilSlice",
			args:    args{usedIpAddresses: nil},
			want:    []netip.Addr{},
			wantErr: false,
		},
		{
			name:    "InvalidIp",
			args:    args{usedIpAddresses: []*types.GatewayUsedIpAddress{{IPAddress: "ASD"}}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ManyIPs",
			args: args{
				usedIpAddresses: []*types.GatewayUsedIpAddress{
					{IPAddress: "10.0.0.1"},
					{IPAddress: "10.0.0.2"},
					{IPAddress: "10.0.0.3"},
					{IPAddress: "10.0.0.4"},
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
				netip.MustParseAddr("10.0.0.2"),
				netip.MustParseAddr("10.0.0.3"),
				netip.MustParseAddr("10.0.0.4"),
			},
			wantErr: false,
		},
		{
			name:    "IPv6SingleIP",
			args:    args{usedIpAddresses: []*types.GatewayUsedIpAddress{{IPAddress: "684D:1111:222:3333:4444:5555:6:77"}}},
			want:    []netip.Addr{netip.MustParseAddr("684D:1111:222:3333:4444:5555:6:77")},
			wantErr: false,
		},
		{
			name: "IPv6ManyIPs",
			args: args{
				usedIpAddresses: []*types.GatewayUsedIpAddress{
					{IPAddress: "2001:db8:3333:4444:5555:6666:7777:8888"},
					{IPAddress: "2002:db8:3333:4444:5555:6666:7777:8888"},
					{IPAddress: "2003:db8:3333:4444:5555:6666:7777:8888"},
					{IPAddress: "2004:db8:3333:4444:5555:6666:7777:8888"},
					{IPAddress: "2001:db8::68"},
				},
			},
			want: []netip.Addr{
				netip.MustParseAddr("2001:db8:3333:4444:5555:6666:7777:8888"),
				netip.MustParseAddr("2002:db8:3333:4444:5555:6666:7777:8888"),
				netip.MustParseAddr("2003:db8:3333:4444:5555:6666:7777:8888"),
				netip.MustParseAddr("2004:db8:3333:4444:5555:6666:7777:8888"),
				netip.MustParseAddr("2001:db8::68"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenGatewayUsedIpAddressesToIpSlice(tt.args.usedIpAddresses)
			if (err != nil) != tt.wantErr {
				t.Errorf("flattenGatewayUsedIpAddressesToIpSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenGatewayUsedIpAddressesToIpSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOpenAPIEdgeGateway_DeallocateIpCount tests that the function
// OpenAPIEdgeGateway.DeallocateIpCount is correctly processing the Edge Gateway uplink structure
func TestOpenAPIEdgeGateway_DeallocateIpCount(t *testing.T) {
	type fields struct {
		EdgeGatewayUplinks []types.EdgeGatewayUplinks
	}
	type args struct {
		deallocateIpCount int
		expectedCount     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "SingleStartAndEndAddresses",
			fields: fields{
				EdgeGatewayUplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
										},
									},
									TotalIPCount: takeIntAddress(2),
								},
							},
						},
					},
				},
			},
			args: args{
				deallocateIpCount: 1,
				expectedCount:     1,
			},
		},
		{
			// Here we check that the function is able to deallocate exactly one IP address (the
			// last). The API will return an error during such operation because Edge Gateway must
			// have at least one IP address allocated.
			name: "ExactlyOnIp",
			fields: fields{
				EdgeGatewayUplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.1",
											},
										},
									},
									TotalIPCount: takeIntAddress(1),
								},
							},
						},
					},
				},
			},
			args: args{
				deallocateIpCount: 1,
				expectedCount:     0,
			},
		},
		{
			name: "NegativeAllocationImpossible",
			fields: fields{
				EdgeGatewayUplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
										},
									},
									TotalIPCount: takeIntAddress(2),
								},
							},
						},
					},
				},
			},
			args: args{
				deallocateIpCount: -1,
			},
			wantErr: true,
		},
		{
			name: "MultipleSubnets",
			fields: fields{
				EdgeGatewayUplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
										},
									},
									TotalIPCount: takeIntAddress(2),
								},
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.20.10.1",
												EndAddress:   "10.20.10.2",
											},
										},
									},
									TotalIPCount: takeIntAddress(2),
								},
							},
						},
					},
				},
			},
			args: args{
				deallocateIpCount: 3,
				expectedCount:     1,
			},
		},
		{
			name: "RemoveMoreThanAvailable",
			fields: fields{
				EdgeGatewayUplinks: []types.EdgeGatewayUplinks{
					{
						Subnets: types.OpenAPIEdgeGatewaySubnets{
							Values: []types.OpenAPIEdgeGatewaySubnetValue{
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.10.10.1",
												EndAddress:   "10.10.10.2",
											},
										},
									},
									TotalIPCount: takeIntAddress(2),
								},
								{
									IPRanges: &types.OpenApiIPRanges{
										Values: []types.OpenApiIPRangeValues{
											{
												StartAddress: "10.20.10.1",
												EndAddress:   "10.20.10.2",
											},
										},
									},
									TotalIPCount: takeIntAddress(2),
								},
							},
						},
					},
				},
			},
			args: args{
				deallocateIpCount: 5, // only 4 IPs are available
				expectedCount:     1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			egw := &NsxtEdgeGateway{
				EdgeGateway: &types.OpenAPIEdgeGateway{
					EdgeGatewayUplinks: tt.fields.EdgeGatewayUplinks,
				},
			}
			var err error
			if err = egw.DeallocateIpCount(tt.args.deallocateIpCount); (err != nil) != tt.wantErr {
				t.Errorf("OpenAPIEdgeGateway.DeallocateIpCount() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Skip other validations if an error was expected
			if err != nil && tt.wantErr {
				return
			}

			allocatedIpCount, err := egw.GetAllocatedIpCount(false)
			if err != nil {
				t.Errorf("NsxtEdgeGateway.GetAllocatedIpCount() error = %v", err)
			}

			if allocatedIpCount != tt.args.expectedCount {
				t.Errorf("Allocated IP count %d != desired IP count %d", allocatedIpCount, tt.args.expectedCount)
			}

		})
	}
}
