package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/url"
	"strings"
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

// This constant is useful when managing Logical VM Groups by referencing VM Groups, as these are
// XML based and don't deal with IDs with full URNs, while Logical VM Groups are OpenAPI based and they do.
const vmGroupUrnPrefix = "urn:vcloud:namedVmGroup"

// GetVmGroupById finds a VM Group by its ID.
// On success, returns a pointer to the VmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetVmGroupById(id string) (*VmGroup, error) {
	return getVmGroupWithFilter(vcdClient, "vmGroupId=="+url.QueryEscape(extractUuid(id)))
}

// GetVmGroupByNamedVmGroupIdAndProviderVdcUrn finds a VM Group by its Named VM Group ID and Provider VDC URN.
// On success, returns a pointer to the VmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetVmGroupByNamedVmGroupIdAndProviderVdcUrn(namedVmGroupId, pvdcUrn string) (*VmGroup, error) {
	id := extractUuid(namedVmGroupId)
	resourcePools, err := getResourcePools(vcdClient, pvdcUrn)
	if err != nil {
		return nil, fmt.Errorf("could not get VM Group with namedVmGroupId=%s: %s", id, err)
	}
	filter, err := buildFilterForVmGroups(resourcePools, "namedVmGroupId", id)
	if err != nil {
		return nil, fmt.Errorf("could not get VM Group with namedVmGroupId=%s: %s", id, err)
	}
	return getVmGroupWithFilter(vcdClient, filter)
}

// GetVmGroupByNameAndProviderVdcUrn finds a VM Group by its name and associated Provider VDC URN.
// On success, returns a pointer to the VmGroup structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetVmGroupByNameAndProviderVdcUrn(name, pvdcUrn string) (*VmGroup, error) {
	resourcePools, err := getResourcePools(vcdClient, pvdcUrn)
	if err != nil {
		return nil, fmt.Errorf("could not get VM Group with vmGroupName=%s: %s", name, err)
	}
	filter, err := buildFilterForVmGroups(resourcePools, "vmGroupName", name)
	if err != nil {
		return nil, fmt.Errorf("could not get VM Group with vmGroupName=%s: %s", name, err)
	}
	return getVmGroupWithFilter(vcdClient, filter)
}

// buildFilterForVmGroups builds a filter to search for VM Groups based on the given resource pools and the desired
// identifier key and value.
func buildFilterForVmGroups(resourcePools []*types.QueryResultResourcePoolRecordType, idKey, idValue string) (string, error) {
	if strings.TrimSpace(idKey) == "" || strings.TrimSpace(idValue) == "" {
		return "", fmt.Errorf("identifier must have a key and value to be able to search")
	}
	clusterMorefs := ""
	vCenters := ""
	for _, resourcePool := range resourcePools {
		if resourcePool.ClusterMoref != "" {
			clusterMorefs += fmt.Sprintf("clusterMoref==%s,", url.QueryEscape(resourcePool.ClusterMoref))
		}
		if resourcePool.VcenterHREF != "" {
			vCenters += fmt.Sprintf("vcId==%s,", url.QueryEscape(extractUuid(resourcePool.VcenterHREF)))
		}
	}

	if len(clusterMorefs) == 0 || len(vCenters) == 0 {
		return "", fmt.Errorf("could not retrieve Resource pools information to retrieve VM Group with %s=%s", idKey, idValue)
	}
	// Removes trailing ","
	clusterMorefs = clusterMorefs[:len(clusterMorefs)-1]
	vCenters = vCenters[:len(vCenters)-1]

	return fmt.Sprintf("(%s==%s;(%s);(%s))", url.QueryEscape(idKey), url.QueryEscape(idValue), clusterMorefs, vCenters), nil
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
func getVmGroupWithFilter(vcdClient *VCDClient, filter string) (*VmGroup, error) {
	foundVmGroups, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "vmGroups",
		"filter":        filter,
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

// getResourcePools returns the Resource Pool that can unequivocally identify a VM Group
func getResourcePools(vcdClient *VCDClient, pvdcUrn string) ([]*types.QueryResultResourcePoolRecordType, error) {
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
	return foundResourcePools.Results.ResourcePoolRecord, nil
}
