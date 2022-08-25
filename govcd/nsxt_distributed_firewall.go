package govcd

import (
	"errors"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// DistributedFirewall contains a types.DistributedFirewallRules which handles Distributed Firewall
// rules in a VDC Group
type DistributedFirewall struct {
	DistributedFirewallRuleContainer *types.DistributedFirewallRules
	client                           *Client
	VdcGroup                         *VdcGroup
}

// GetDistributedFirewall retrieves Distributed Firewall in a VDC Group which contains all rules
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) GetDistributedFirewall() (*DistributedFirewall, error) {
	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault))
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

// UpdateDistributedFirewall updates Distributed Firewall in a VDC Group
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) UpdateDistributedFirewall(dfwRules *types.DistributedFirewallRules) (*DistributedFirewall, error) {
	client := vdcGroup.client
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVdcGroupsDfwRules
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// "default" policy is hardcoded because there is no other policy supported
	urlRef, err := client.OpenApiBuildEndpoint(fmt.Sprintf(endpoint, vdcGroup.VdcGroup.Id, types.DistributedFirewallPolicyDefault))
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
		return nil, fmt.Errorf("error updating Distributed Firewall rules: %s", err)
	}

	return returnObject, nil
}

// DeleteAllDistributedFirewallRules removes all Distributed Firewall rules
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (vdcGroup *VdcGroup) DeleteAllDistributedFirewallRules() error {
	_, err := vdcGroup.UpdateDistributedFirewall(&types.DistributedFirewallRules{})
	return err
}

// DeleteAllRules removes all Distributed Firewall rules
//
// Note. This function works only with `default` policy as this was the only supported when this
// functions was created
func (firewall *DistributedFirewall) DeleteAllRules() error {
	if firewall.VdcGroup != nil && firewall.VdcGroup.VdcGroup != nil && firewall.VdcGroup.VdcGroup.Id == "" {
		return errors.New("empty VDC Group ID for parent VDC Group")
	}

	return firewall.VdcGroup.DeleteAllDistributedFirewallRules()
}
