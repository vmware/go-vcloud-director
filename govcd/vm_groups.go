package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
)

// LogicalVmGroup is used to create VM Placement Policies.
type LogicalVmGroup struct {
	LogicalVmGroup *types.LogicalVmGroup
	client         *Client
}

// VmGroup is used to create VM Placement Policies.
type VmGroup struct {
	VmGroup *types.QueryResultVmGroupsRecordType
	client  *Client
}

// GetVmGroupById finds a VM Group by its ID.
// On success, returns a pointer to the VmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetVmGroupById(id string) (*VmGroup, error) {
	return getVmGroupWithFilter(vcdClient, map[string]string{"vmGroupId": extractUuid(id)})
}

// GetVmGroupByNamedVmGroupIdAndProviderVdcUrn finds a VM Group by its Named VM Group ID and Provider VDC URN.
// On success, returns a pointer to the VmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetVmGroupByNamedVmGroupIdAndProviderVdcUrn(namedVmGroupId, pvdcUrn string) (*VmGroup, error) {
	resourcePool, err := getResourcePool(vcdClient, pvdcUrn)
	if err != nil {
		return nil, fmt.Errorf("could not get VM Group: %s", err)
	}
	return getVmGroupWithFilter(vcdClient, map[string]string{"namedVmGroupId": extractUuid(namedVmGroupId), "clusterMoref": resourcePool.ClusterMoref, "vcId": extractUuid(resourcePool.VcenterHREF)})
}

// GetVmGroupByNameAndProviderVdcUrn finds a VM Group by its name and associated Provider VDC URN.
// On success, returns a pointer to the VmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetVmGroupByNameAndProviderVdcUrn(name, pvdcUrn string) (*VmGroup, error) {
	resourcePool, err := getResourcePool(vcdClient, pvdcUrn)
	if err != nil {
		return nil, fmt.Errorf("could not get VM Group: %s", err)
	}
	return getVmGroupWithFilter(vcdClient, map[string]string{"vmGroupName": name, "clusterMoref": resourcePool.ClusterMoref, "vcId": extractUuid(resourcePool.VcenterHREF)})
}

// GetLogicalVmGroupById finds a Logical VM Group by its URN.
// On success, returns a pointer to the LogicalVmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetLogicalVmGroupById(logicalVmGroupId string) (*LogicalVmGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointLogicalVmGroups

	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	if logicalVmGroupId == "" {
		return nil, fmt.Errorf("empty Logical VM Group id")
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint, logicalVmGroupId)
	if err != nil {
		return nil, err
	}

	result := &LogicalVmGroup{
		LogicalVmGroup: &types.LogicalVmGroup{},
		client:         &vcdClient.Client,
	}

	err = vcdClient.Client.OpenApiGetItem(apiVersion, urlRef, nil, result.LogicalVmGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting Logical VM Group: %s", err)
	}

	return result, nil
}

// CreateLogicalVmGroup creates a new Logical VM Group in VCD
func (vcdClient *VCDClient) CreateLogicalVmGroup(logicalVmGroup types.LogicalVmGroup) (*LogicalVmGroup, error) {
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointLogicalVmGroups

	apiVersion, err := vcdClient.Client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdClient.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	result := &LogicalVmGroup{
		LogicalVmGroup: &types.LogicalVmGroup{},
		client:         &vcdClient.Client,
	}

	err = vcdClient.Client.OpenApiPostItem(apiVersion, urlRef, nil, logicalVmGroup, result.LogicalVmGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating the Logical VM Group: %s", err)
	}

	return result, nil
}

// Delete deletes the receiver Logical VM Group
func (logicalVmGroup *LogicalVmGroup) Delete() error {
	if logicalVmGroup.LogicalVmGroup.ID == "" {
		return fmt.Errorf("cannot delete Logical VM Group without id")
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointLogicalVmGroups

	apiVersion, err := logicalVmGroup.client.getOpenApiHighestElevatedVersion(endpoint)
	if err != nil {
		return err
	}

	urlRef, err := logicalVmGroup.client.OpenApiBuildEndpoint(endpoint, logicalVmGroup.LogicalVmGroup.ID)
	if err != nil {
		return err
	}

	err = logicalVmGroup.client.OpenApiDeleteItem(apiVersion, urlRef, nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting the Logical VM Group: %s", err)
	}
	return nil
}

// getVmGroupWithFilter finds a VM Group by specifying a filter=(filterKey==filterValue).
// On success, returns a pointer to the VmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func getVmGroupWithFilter(vcdClient *VCDClient, filter map[string]string) (*VmGroup, error) {
	filterEncoded := ""
	for k, v := range filter {
		filterEncoded += fmt.Sprintf("%s==%s;", url.QueryEscape(k), url.QueryEscape(v))
	}
	foundVmGroups, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "vmGroups",
		"filter":        filterEncoded[:len(filterEncoded)-1], // Removes the trailing ';'
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}
	if len(foundVmGroups.Results.VmGroupsRecord) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(foundVmGroups.Results.VmGroupsRecord) > 1 {
		return nil, fmt.Errorf("more than one VM Group found with the filter: %v", filter)
	}
	vmGroup := &VmGroup{
		VmGroup: foundVmGroups.Results.VmGroupsRecord[0],
		client:  &vcdClient.Client,
	}
	return vmGroup, nil
}

// getResourcePool returns the Resource Pool that can unequivocally identify a VM Group
func getResourcePool(vcdClient *VCDClient, pvdcUrn string) (*types.QueryResultResourcePoolRecordType, error) {
	foundResourcePools, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "resourcePool",
		"filter":        fmt.Sprintf("providerVdc==%s", url.QueryEscape(pvdcUrn)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, fmt.Errorf("could not get the Resource pool: %s", err)
	}
	if len(foundResourcePools.Results.ResourcePoolRecord) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(foundResourcePools.Results.ResourcePoolRecord) > 1 {
		return nil, fmt.Errorf("more than one Resource Pool found for the pVDC: %s", pvdcUrn)
	}
	return foundResourcePools.Results.ResourcePoolRecord[0], nil
}
