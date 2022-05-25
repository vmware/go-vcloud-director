/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import "github.com/vmware/go-vcloud-director/v2/types/v56"

func (egw *NsxtEdgeGateway) GetNsxtRouteAdvertisement() (*types.RouteAdvertisement, error) {

	return nil, nil
}

func (egw *NsxtEdgeGateway) UpdateNsxtRouteAdvertisement(enable bool, subnets []string) (*types.RouteAdvertisement, error) {
	return nil, nil
}

func (egw *NsxtEdgeGateway) DeleteNsxtRouteAdvertisement() error {
	return nil
}
