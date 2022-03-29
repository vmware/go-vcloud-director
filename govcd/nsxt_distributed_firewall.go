package govcd

import (
	"errors"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// DistributedFirewall contains a types.DfwFirewallRule
type DistributedFirewall struct {
	DistributedFirewallRuleContainer *types.DistributedFirewallRules
	client                           *Client
	VdcGroup                         *VdcGroup
}

func (vdcGroup *VdcGroup) GetDistributedFirewall() (*DistributedFirewall, error) {
	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, "default"))
	if err != nil {
		return nil, err
	}

	returnObject := &DistributedFirewall{
		DistributedFirewallRuleContainer: &types.DistributedFirewallRules{},
		client:                           client,
		VdcGroup:                         vdcGroup,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, returnObject.DistributedFirewallRuleContainer, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Distributed Firewall rules: %s", err)
	}

	return returnObject, nil
}

func (vdcGroup *VdcGroup) UpdateDistributedFirewall(dfwRules *types.DistributedFirewallRules) (*DistributedFirewall, error) {
	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, "default"))
	if err != nil {
		return nil, err
	}

	returnObject := &DistributedFirewall{
		DistributedFirewallRuleContainer: &types.DistributedFirewallRules{},
		client:                           client,
		VdcGroup:                         vdcGroup,
	}

	err = client.OpenApiPutItem(apiVersion, urlRef, nil, dfwRules, returnObject.DistributedFirewallRuleContainer, nil)
	if err != nil {
		return nil, fmt.Errorf("error setting Distributed Firewall rules: %s", err)
	}

	return returnObject, nil
}

func (vdcGroup *VdcGroup) DeleteAllDistributedFirewallRules() error {
	_, err := vdcGroup.UpdateDistributedFirewall(&types.DistributedFirewallRules{})
	return err
}

func (firewall *DistributedFirewall) DeleteAllRules() error {
	if firewall.VdcGroup != nil && firewall.VdcGroup.VdcGroup != nil && firewall.VdcGroup.VdcGroup.Id == "" {
		return errors.New("empty VDC Group ID for parent VDC Group")
	}

	return firewall.VdcGroup.DeleteAllDistributedFirewallRules()
}
