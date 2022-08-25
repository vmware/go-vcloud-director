/*
 * Copyright 2021 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtFirewallGroup uses OpenAPI endpoint to operate NSX-T Security Groups and IP Sets which use
// the same Firewall Group API endpoint
//
// IP sets are groups of objects to which the firewall rules apply. Combining multiple objects into
// IP sets helps reduce the total number of firewall rules to be created.
//
// Security groups are groups of Org Vdc networks to which distributed firewall rules apply.
// Grouping networks helps you to reduce the total number of distributed firewall rules to be
// created.
type NsxtFirewallGroup struct {
	NsxtFirewallGroup *types.NsxtFirewallGroup
	client            *Client
}

// CreateNsxtFirewallGroup allows users to create NSX-T Firewall Group
func (vdc *Vdc) CreateNsxtFirewallGroup(firewallGroupConfig *types.NsxtFirewallGroup) (*NsxtFirewallGroup, error) {
	return createNsxtFirewallGroup(vdc.client, firewallGroupConfig)
}

// CreateNsxtFirewallGroup allows users to create NSX-T Firewall Group
func (egw *NsxtEdgeGateway) CreateNsxtFirewallGroup(firewallGroupConfig *types.NsxtFirewallGroup) (*NsxtFirewallGroup, error) {
	return createNsxtFirewallGroup(egw.client, firewallGroupConfig)
}

// CreateNsxtFirewallGroup allows users to create NSX-T Firewall Group
func (vdcGroup *VdcGroup) CreateNsxtFirewallGroup(firewallGroupConfig *types.NsxtFirewallGroup) (*NsxtFirewallGroup, error) {
	return createNsxtFirewallGroup(vdcGroup.client, firewallGroupConfig)
}

// GetAllNsxtFirewallGroups allows users to retrieve all Firewall Groups for Org
// firewallGroupType can be one of the following:
// * types.FirewallGroupTypeSecurityGroup - for NSX-T Security Groups
// * types.FirewallGroupTypeIpSet - for NSX-T IP Sets
// * "" (empty) - search will not be limited and will get both - IP Sets and Security Groups
//
// It is possible to add additional filtering by using queryParameters of type 'url.Values'.
// One special filter is `_context==` filtering. Value can be one of the following:
//
// * Org Vdc Network ID (_context==networkId) - Returns all the firewall groups which the specified
// network is a member of.
//
// * Edge Gateway ID (_context==edgeGatewayId) - Returns all the firewall groups which are available
// to the specific edge gateway. Or use a shorthand NsxtEdgeGateway.GetAllNsxtFirewallGroups() which
// automatically injects this filter.
//
// * Network Provider ID (_context==networkProviderId) - Returns all the firewall groups which are
// available under a specific network provider. This context requires system admin privilege.
// 'networkProviderId' is NSX-T manager ID
func (org *Org) GetAllNsxtFirewallGroups(queryParameters url.Values, firewallGroupType string) ([]*NsxtFirewallGroup, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	if firewallGroupType != "" {
		queryParams = queryParameterFilterAnd(fmt.Sprintf("typeValue==%s", firewallGroupType), queryParameters)
	}

	return getAllNsxtFirewallGroups(org.client, queryParams)
}

// GetAllNsxtFirewallGroups allows users to retrieve all NSX-T Firewall Groups
func (vdc *Vdc) GetAllNsxtFirewallGroups(queryParameters url.Values, firewallGroupType string) ([]*NsxtFirewallGroup, error) {
	if vdc.IsNsxv() {
		return nil, errors.New("only NSX-T VDCs support Firewall Groups")
	}
	return getAllNsxtFirewallGroups(vdc.client, queryParameters)
}

// GetAllNsxtFirewallGroups allows users to retrieve all NSX-T Firewall Groups in a particular Edge Gateway
// firewallGroupType can be one of the following:
// * types.FirewallGroupTypeSecurityGroup - for NSX-T Security Groups
// * types.FirewallGroupTypeIpSet - for NSX-T IP Sets
// * "" (empty) - search will not be limited and will get both - IP Sets and Security Groups
func (egw *NsxtEdgeGateway) GetAllNsxtFirewallGroups(queryParameters url.Values, firewallGroupType string) ([]*NsxtFirewallGroup, error) {
	queryParams := copyOrNewUrlValues(queryParameters)

	if firewallGroupType != "" {
		queryParams = queryParameterFilterAnd(fmt.Sprintf("typeValue==%s", firewallGroupType), queryParameters)
	}

	// Automatically inject Edge Gateway filter because this is an Edge Gateway scoped query
	queryParams = queryParameterFilterAnd("_context=="+egw.EdgeGateway.ID, queryParams)

	return getAllNsxtFirewallGroups(egw.client, queryParams)
}

// GetNsxtFirewallGroupByName allows users to retrieve Firewall Group by Name
// firewallGroupType can be one of the following:
// * types.FirewallGroupTypeSecurityGroup - for NSX-T Security Groups
// * types.FirewallGroupTypeIpSet - for NSX-T IP Sets
// * "" (empty) - search will not be limited and will get both - IP Sets and Security Groups
//
// Note. One might get an error if IP Set and Security Group exist with the same name (two objects
// of the same type cannot exist) and firewallGroupType is left empty.
func (org *Org) GetNsxtFirewallGroupByName(name, firewallGroupType string) (*NsxtFirewallGroup, error) {
	queryParameters := url.Values{}
	if firewallGroupType != "" {
		queryParameters = queryParameterFilterAnd(fmt.Sprintf("typeValue==%s", firewallGroupType), queryParameters)
	}

	return getNsxtFirewallGroupByName(org.client, name, queryParameters)
}

// GetNsxtFirewallGroupByName allows users to retrieve Firewall Group by Name
// firewallGroupType can be one of the following:
// * types.FirewallGroupTypeSecurityGroup - for NSX-T Security Groups
// * types.FirewallGroupTypeIpSet - for NSX-T IP Sets
// * "" (empty) - search will not be limited and will get both - IP Sets and Security Groups
//
// Note. One might get an error if IP Set and Security Group exist with the same name (two objects
// of the same type cannot exist) and firewallGroupType is left empty.
func (vdc *Vdc) GetNsxtFirewallGroupByName(name, firewallGroupType string) (*NsxtFirewallGroup, error) {

	queryParameters := url.Values{}
	if firewallGroupType != "" {
		queryParameters = queryParameterFilterAnd(fmt.Sprintf("typeValue==%s", firewallGroupType), queryParameters)
	}
	return getNsxtFirewallGroupByName(vdc.client, name, queryParameters)
}

// GetNsxtFirewallGroupByName allows users to retrieve Firewall Group by Name in a particular VDC Group
// firewallGroupType can be one of the following:
// * types.FirewallGroupTypeSecurityGroup - for NSX-T Static Security Groups
// * types.FirewallGroupTypeVmCriteria - for NSX-T Dynamic Security Groups
// * types.FirewallGroupTypeIpSet - for NSX-T IP Sets
// * "" (empty) - search will not be limited and will get both - IP Sets and Security Groups
//
// Note. One might get an error if IP Set and Security Group exist with the same name (two objects
// of the same type cannot exist) and firewallGroupType is left empty.
func (vdcGroup *VdcGroup) GetNsxtFirewallGroupByName(name string, firewallGroupType string) (*NsxtFirewallGroup, error) {
	queryParameters := url.Values{}

	if firewallGroupType != "" {
		queryParameters = queryParameterFilterAnd(fmt.Sprintf("typeValue==%s", firewallGroupType), queryParameters)
	}

	// Automatically inject Edge Gateway filter because this is an Edge Gateway scoped query
	queryParameters = queryParameterFilterAnd("ownerRef.id=="+vdcGroup.VdcGroup.Id, queryParameters)

	return getNsxtFirewallGroupByName(vdcGroup.client, name, queryParameters)
}

// GetNsxtFirewallGroupByName allows users to retrieve Firewall Group by Name in a particular Edge Gateway
// firewallGroupType can be one of the following:
// * types.FirewallGroupTypeSecurityGroup - for NSX-T Security Groups
// * types.FirewallGroupTypeIpSet - for NSX-T IP Sets
// * "" (empty) - search will not be limited and will get both - IP Sets and Security Groups
//
// Note. One might get an error if IP Set and Security Group exist with the same name (two objects
// of the same type cannot exist) and firewallGroupType is left empty.
func (egw *NsxtEdgeGateway) GetNsxtFirewallGroupByName(name string, firewallGroupType string) (*NsxtFirewallGroup, error) {
	queryParameters := url.Values{}

	if firewallGroupType != "" {
		queryParameters = queryParameterFilterAnd(fmt.Sprintf("typeValue==%s", firewallGroupType), queryParameters)
	}

	// Automatically inject Edge Gateway filter because this is an Edge Gateway scoped query
	queryParameters = queryParameterFilterAnd("_context=="+egw.EdgeGateway.ID, queryParameters)

	return getNsxtFirewallGroupByName(egw.client, name, queryParameters)
}

// GetNsxtFirewallGroupById retrieves NSX-T Firewall Group by ID
func (org *Org) GetNsxtFirewallGroupById(id string) (*NsxtFirewallGroup, error) {
	return getNsxtFirewallGroupById(org.client, id)
}

// GetNsxtFirewallGroupById retrieves NSX-T Firewall Group by ID
func (vdc *Vdc) GetNsxtFirewallGroupById(id string) (*NsxtFirewallGroup, error) {
	return getNsxtFirewallGroupById(vdc.client, id)
}

// GetNsxtFirewallGroupById retrieves NSX-T Firewall Group by ID
func (egw *NsxtEdgeGateway) GetNsxtFirewallGroupById(id string) (*NsxtFirewallGroup, error) {
	return getNsxtFirewallGroupById(egw.client, id)
}

// GetNsxtFirewallGroupById retrieves NSX-T Firewall Group by ID
func (vdcGroup *VdcGroup) GetNsxtFirewallGroupById(id string) (*NsxtFirewallGroup, error) {
	return getNsxtFirewallGroupById(vdcGroup.client, id)
}

// Update allows users to update NSX-T Firewall Group
func (firewallGroup *NsxtFirewallGroup) Update(firewallGroupConfig *types.NsxtFirewallGroup) (*NsxtFirewallGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups
	apiVersion, err := firewallGroup.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if firewallGroupConfig.ID == "" {
		return nil, fmt.Errorf("cannot update NSX-T Firewall Group without ID")
	}

	urlRef, err := firewallGroup.client.OpenApiBuildEndpoint(endpoint, firewallGroupConfig.ID)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtFirewallGroup{
		NsxtFirewallGroup: &types.NsxtFirewallGroup{},
		client:            firewallGroup.client,
	}

	err = firewallGroup.client.OpenApiPutItem(apiVersion, urlRef, nil, firewallGroupConfig, returnObject.NsxtFirewallGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error updating NSX-T firewall group: %s", err)
	}

	return returnObject, nil
}

// Delete allows users to delete NSX-T Firewall Group
func (firewallGroup *NsxtFirewallGroup) Delete() error {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups
	apiVersion, err := firewallGroup.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	if firewallGroup.NsxtFirewallGroup.ID == "" {
		return fmt.Errorf("cannot delete NSX-T Firewall Group without ID")
	}

	urlRef, err := firewallGroup.client.OpenApiBuildEndpoint(endpoint, firewallGroup.NsxtFirewallGroup.ID)
	if err != nil {
		return err
	}

	err = firewallGroup.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting NSX-T Firewall Group: %s", err)
	}

	return nil
}

// GetAssociatedVms allows users to retrieve a list of references to child VMs (with vApps when they exist).
//
// Note. Only Security Groups have associated VMs. Executing it on an IP Set will return an error
// similar to: "only Security Groups have associated VMs. This Firewall Group has type 'IP_SET'"
func (firewallGroup *NsxtFirewallGroup) GetAssociatedVms() ([]*types.NsxtFirewallGroupMemberVms, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups
	apiVersion, err := firewallGroup.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if firewallGroup.NsxtFirewallGroup.ID == "" {
		return nil, fmt.Errorf("cannot retrieve associated VMs for NSX-T Firewall Group without ID")
	}

	if !firewallGroup.IsSecurityGroup() && !firewallGroup.IsDynamicSecurityGroup() {
		return nil, fmt.Errorf("only Security Groups have associated VMs. This Firewall Group has type '%s'",
			firewallGroup.NsxtFirewallGroup.Type)
	}

	urlRef, err := firewallGroup.client.OpenApiBuildEndpoint(endpoint, firewallGroup.NsxtFirewallGroup.ID, "/associatedVMs")
	if err != nil {
		return nil, err
	}

	associatedVms := []*types.NsxtFirewallGroupMemberVms{{}}

	err = firewallGroup.client.OpenApiGetAllItems(apiVersion, urlRef, nil, &associatedVms, nil)

	if err != nil {
		return nil, fmt.Errorf("error retrieving associated VMs: %s", err)
	}

	return associatedVms, nil
}

// IsSecurityGroup allows users to check if Firewall Group is a Static Security Group
func (firewallGroup *NsxtFirewallGroup) IsSecurityGroup() bool {
	return firewallGroup.NsxtFirewallGroup.Type == types.FirewallGroupTypeSecurityGroup
}

// IsDynamicSecurityGroup allows users to check if Firewall Group is a Dynamic Security Group
func (firewallGroup *NsxtFirewallGroup) IsDynamicSecurityGroup() bool {
	return firewallGroup.NsxtFirewallGroup.TypeValue == types.FirewallGroupTypeVmCriteria
}

// IsIpSet allows users to check if Firewall Group is an IP Set
func (firewallGroup *NsxtFirewallGroup) IsIpSet() bool {
	return firewallGroup.NsxtFirewallGroup.Type == types.FirewallGroupTypeIpSet
}

func getNsxtFirewallGroupByName(client *Client, name string, queryParameters url.Values) (*NsxtFirewallGroup, error) {
	queryParams := copyOrNewUrlValues(queryParameters)
	queryParams = queryParameterFilterAnd("name=="+name, queryParams)

	allGroups, err := getAllNsxtFirewallGroups(client, queryParams)
	if err != nil {
		return nil, fmt.Errorf("could not find NSX-T Firewall Group with name '%s': %s", name, err)
	}

	if len(allGroups) == 0 {
		return nil, fmt.Errorf("%s: expected exactly one NSX-T Firewall Group with name '%s'. Got %d", ErrorEntityNotFound, name, len(allGroups))
	}

	if len(allGroups) > 1 {
		return nil, fmt.Errorf("expected exactly one NSX-T Firewall Group with name '%s'. Got %d", name, len(allGroups))
	}

	// TODO API V36.0 - maybe it is fixed
	// There is a bug that not all data is present (e.g. missing IpAddresses field for IP_SET) when
	// using "getAll" endpoint therefore after finding the object by name we must retrieve it once
	// again using its direct endpoint.
	//
	// return allGroups[0], nil

	return getNsxtFirewallGroupById(client, allGroups[0].NsxtFirewallGroup.ID)
}

func getNsxtFirewallGroupById(client *Client, id string) (*NsxtFirewallGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if id == "" {
		return nil, fmt.Errorf("empty NSX-T Firewall Group ID specified")
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint, id)
	if err != nil {
		return nil, err
	}

	fwGroup := &NsxtFirewallGroup{
		NsxtFirewallGroup: &types.NsxtFirewallGroup{},
		client:            client,
	}

	err = client.OpenApiGetItem(apiVersion, urlRef, nil, fwGroup.NsxtFirewallGroup, nil)
	if err != nil {
		return nil, err
	}

	return fwGroup, nil
}

func getAllNsxtFirewallGroups(client *Client, queryParameters url.Values) ([]*NsxtFirewallGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	// This Object does not follow regular REST scheme and for get the endpoint must be
	// 1.0.0/firewallGroups/summaries therefore bellow "summaries" is appended to the path
	urlRef, err := client.OpenApiBuildEndpoint(endpoint, "summaries")
	if err != nil {
		return nil, err
	}

	typeResponses := []*types.NsxtFirewallGroup{{}}
	err = client.OpenApiGetAllItems(apiVersion, urlRef, queryParameters, &typeResponses, nil)
	if err != nil {
		return nil, err
	}

	// Wrap all typeResponses into NsxtEdgeGateway types with client
	wrappedResponses := make([]*NsxtFirewallGroup, len(typeResponses))
	for sliceIndex := range typeResponses {
		wrappedResponses[sliceIndex] = &NsxtFirewallGroup{
			NsxtFirewallGroup: typeResponses[sliceIndex],
			client:            client,
		}
	}

	return wrappedResponses, nil
}

func createNsxtFirewallGroup(client *Client, firewallGroupConfig *types.NsxtFirewallGroup) (*NsxtFirewallGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointFirewallGroups
	apiVersion, err := client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	returnObject := &NsxtFirewallGroup{
		NsxtFirewallGroup: &types.NsxtFirewallGroup{},
		client:            client,
	}

	err = client.OpenApiPostItem(apiVersion, urlRef, nil, firewallGroupConfig, returnObject.NsxtFirewallGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating NSX-T Firewall Group: %s", err)
	}

	return returnObject, nil
}
