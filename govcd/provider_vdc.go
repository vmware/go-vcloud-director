package govcd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// ProviderVdc is the basic Provider VDC structure, contains the minimum set of attributes.
type ProviderVdc struct {
	ProviderVdc *types.ProviderVdc
	client      *Client
}

// ProviderVdcExtended is the extended Provider VDC structure, contains same attributes as ProviderVdc plus some more.
type ProviderVdcExtended struct {
	VMWProviderVdc *types.VMWProviderVdc
	client         *Client
}

func newProviderVdc(cli *Client) *ProviderVdc {
	return &ProviderVdc{
		ProviderVdc: new(types.ProviderVdc),
		client:      cli,
	}
}

func newProviderVdcExtended(cli *Client) *ProviderVdcExtended {
	return &ProviderVdcExtended{
		VMWProviderVdc: new(types.VMWProviderVdc),
		client:         cli,
	}
}

// GetProviderVdcByHref finds a Provider VDC by its HREF.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcByHref(providerVdcHref string) (*ProviderVdc, error) {
	providerVdc := newProviderVdc(&vcdClient.Client)

	_, err := vcdClient.Client.ExecuteRequest(providerVdcHref, http.MethodGet,
		"", "error retrieving Provider VDC: %s", nil, providerVdc.ProviderVdc)
	if err != nil {
		return nil, err
	}

	return providerVdc, nil
}

// GetProviderVdcExtendedByHref finds a Provider VDC with extended attributes by its HREF.
// On success, returns a pointer to the ProviderVdcExtended structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcExtendedByHref(providerVdcHref string) (*ProviderVdcExtended, error) {
	providerVdc := newProviderVdcExtended(&vcdClient.Client)

	_, err := vcdClient.Client.ExecuteRequest(getAdminExtensionURL(providerVdcHref), http.MethodGet,
		"", "error retrieving extended Provider VDC: %s", nil, providerVdc.VMWProviderVdc)
	if err != nil {
		return nil, err
	}

	return providerVdc, nil
}

// GetProviderVdcById finds a Provider VDC by URN.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcById(providerVdcId string) (*ProviderVdc, error) {
	providerVdcHref := vcdClient.Client.VCDHREF
	providerVdcHref.Path += "/admin/providervdc/" + extractUuid(providerVdcId)

	return vcdClient.GetProviderVdcByHref(providerVdcHref.String())
}

// GetProviderVdcExtendedById finds a Provider VDC with extended attributes by URN.
// On success, returns a pointer to the ProviderVdcExtended structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcExtendedById(providerVdcId string) (*ProviderVdcExtended, error) {
	providerVdcHref := vcdClient.Client.VCDHREF
	providerVdcHref.Path += "/admin/extension/providervdc/" + extractUuid(providerVdcId)

	return vcdClient.GetProviderVdcExtendedByHref(providerVdcHref.String())
}

// GetProviderVdcByName finds a Provider VDC by name.
// On success, returns a pointer to the ProviderVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcByName(providerVdcName string) (*ProviderVdc, error) {
	providerVdc, err := getProviderVdcByName(vcdClient, providerVdcName, false)
	if err != nil {
		return nil, err
	}
	return providerVdc.(*ProviderVdc), err
}

// GetProviderVdcExtendedByName finds a Provider VDC with extended attributes by name.
// On success, returns a pointer to the ProviderVdcExtended structure and a nil error
// On failure, returns a nil pointer and an error
func (vcdClient *VCDClient) GetProviderVdcExtendedByName(providerVdcName string) (*ProviderVdcExtended, error) {
	providerVdcExtended, err := getProviderVdcByName(vcdClient, providerVdcName, true)
	if err != nil {
		return nil, err
	}
	return providerVdcExtended.(*ProviderVdcExtended), err
}

// Refresh updates the contents of the Provider VDC associated to the receiver object.
func (providerVdc *ProviderVdc) Refresh() error {
	if providerVdc.ProviderVdc.HREF == "" {
		return fmt.Errorf("cannot refresh, receiver Provider VDC is empty")
	}

	unmarshalledVdc := &types.ProviderVdc{}

	_, err := providerVdc.client.ExecuteRequest(providerVdc.ProviderVdc.HREF, http.MethodGet,
		"", "error refreshing Provider VDC: %s", nil, unmarshalledVdc)
	if err != nil {
		return err
	}

	providerVdc.ProviderVdc = unmarshalledVdc

	return nil
}

// Refresh updates the contents of the extended Provider VDC associated to the receiver object.
func (providerVdcExtended *ProviderVdcExtended) Refresh() error {
	if providerVdcExtended.VMWProviderVdc.HREF == "" {
		return fmt.Errorf("cannot refresh, receiver extended Provider VDC is empty")
	}

	unmarshalledVdc := &types.VMWProviderVdc{}

	_, err := providerVdcExtended.client.ExecuteRequest(providerVdcExtended.VMWProviderVdc.HREF, http.MethodGet,
		"", "error refreshing extended Provider VDC: %s", nil, unmarshalledVdc)
	if err != nil {
		return err
	}

	providerVdcExtended.VMWProviderVdc = unmarshalledVdc

	return nil
}

// ToProviderVdc converts the receiver ProviderVdcExtended into the subset ProviderVdc
func (providerVdcExtended *ProviderVdcExtended) ToProviderVdc() (*ProviderVdc, error) {
	providerVdcHref := providerVdcExtended.client.VCDHREF
	providerVdcHref.Path += "/admin/providervdc/" + extractUuid(providerVdcExtended.VMWProviderVdc.ID)

	providerVdc := newProviderVdc(providerVdcExtended.client)

	_, err := providerVdcExtended.client.ExecuteRequest(providerVdcHref.String(), http.MethodGet,
		"", "error retrieving Provider VDC: %s", nil, providerVdc.ProviderVdc)
	if err != nil {
		return nil, err
	}

	return providerVdc, nil
}

// getProviderVdcByName finds a Provider VDC with extension (extended=true) or without extension (extended=false) by name
// On success, returns a pointer to the ProviderVdc (extended=false) or ProviderVdcExtended (extended=true) structure and a nil error
// On failure, returns a nil pointer and an error
func getProviderVdcByName(vcdClient *VCDClient, providerVdcName string, extended bool) (interface{}, error) {
	foundProviderVdcs, err := vcdClient.QueryWithNotEncodedParams(nil, map[string]string{
		"type":          "providerVdc",
		"filter":        fmt.Sprintf("name==%s", url.QueryEscape(providerVdcName)),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, err
	}
	if len(foundProviderVdcs.Results.VMWProviderVdcRecord) == 0 {
		return nil, ErrorEntityNotFound
	}
	if len(foundProviderVdcs.Results.VMWProviderVdcRecord) > 1 {
		return nil, fmt.Errorf("more than one Provider VDC found with name '%s'", providerVdcName)
	}
	if extended {
		return vcdClient.GetProviderVdcExtendedByHref(foundProviderVdcs.Results.VMWProviderVdcRecord[0].HREF)
	}
	return vcdClient.GetProviderVdcByHref(foundProviderVdcs.Results.VMWProviderVdcRecord[0].HREF)
}

// CreateProviderVdc creates a new provider VDC using the passed parameters
func (vcdClient *VCDClient) CreateProviderVdc(params *types.ProviderVdcCreation) (*ProviderVdcExtended, error) {
	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("functionality requires System Administrator privileges")
	}
	if params.Name == "" {
		return nil, fmt.Errorf("a non-empty name is needed to create a provider VDC")
	}
	if params.ResourcePoolRefs == nil || len(params.ResourcePoolRefs.VimObjectRef) == 0 {
		return nil, fmt.Errorf("resource pool is needed to create a provider VDC")
	}
	if len(params.StorageProfile) == 0 {
		return nil, fmt.Errorf("storage profile is needed to create a provider VDC")
	}
	if params.VimServer == nil {
		return nil, fmt.Errorf("vim server is needed to create a provider VDC")
	}
	pvdcCreateHREF := vcdClient.Client.VCDHREF
	pvdcCreateHREF.Path += "/admin/extension/providervdcsparams"

	resp, err := vcdClient.Client.executeJsonRequest(pvdcCreateHREF.String(), http.MethodPost, params, "error creating provider VDC: %s")
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	util.ProcessResponseOutput(util.CallFuncName(), resp, string(body))

	defer closeBody(resp)

	pvdc, err := vcdClient.GetProviderVdcExtendedByName(params.Name)
	if err != nil {
		return nil, err
	}

	// At this stage, the provider VDC is created, but the task may be still working.
	// Thus, we retrieve the associated tasks, and wait for their completion.
	if pvdc.VMWProviderVdc.Tasks == nil {
		err = pvdc.Refresh()
		if err != nil {
			return pvdc, fmt.Errorf("error refreshing provider VDC %s: %s", params.Name, err)
		}
		if pvdc.VMWProviderVdc.Tasks == nil {
			return pvdc, fmt.Errorf("provider VDC %s was created, but no completion task was found: %s", params.Name, err)
		}
	}
	for _, taskInProgress := range pvdc.VMWProviderVdc.Tasks.Task {
		task := Task{
			Task:   taskInProgress,
			client: pvdc.client,
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return pvdc, fmt.Errorf("provider VDC %s was created, but it is not ready: %s", params.Name, err)
		}
	}

	err = pvdc.Refresh()
	return pvdc, err
}

// Disable changes the Provider VDC state from enabled to disabled
func (pvdc *ProviderVdcExtended) Disable() error {
	util.Logger.Printf("[TRACE] ProviderVdc.Disable")

	href, err := url.JoinPath(pvdc.VMWProviderVdc.HREF, "action", "disable")

	if err != nil {
		return err
	}

	err = pvdc.client.ExecuteRequestWithoutResponse(href, http.MethodPost, "", "error disabling provider VDC: %s", nil)
	if err != nil {
		return err
	}
	err = pvdc.Refresh()
	if err != nil {
		return err
	}
	if pvdc.IsEnabled() {
		return fmt.Errorf("provider VDC was disabled, but its status is still shown as 'enabled'")
	}
	return nil
}

// IsEnabled shows whether the Provider VDC is enabled
func (pvdc *ProviderVdcExtended) IsEnabled() bool {
	if pvdc.VMWProviderVdc.IsEnabled == nil {
		return false
	}
	return *pvdc.VMWProviderVdc.IsEnabled
}

// Enable changes the Provider VDC state from disabled to enabled
func (pvdc *ProviderVdcExtended) Enable() error {
	util.Logger.Printf("[TRACE] ProviderVdc.Enable")

	href, err := url.JoinPath(pvdc.VMWProviderVdc.HREF, "action", "enable")

	if err != nil {
		return err
	}

	err = pvdc.client.ExecuteRequestWithoutResponse(href, http.MethodPost, "",
		"error enabling provider VDC: %s", nil)
	if err != nil {
		return err
	}
	err = pvdc.Refresh()
	if err != nil {
		return err
	}
	if !pvdc.IsEnabled() {
		return fmt.Errorf("provider VDC was enabled, but its status is still shown as 'disabled'")
	}
	return nil
}

// Delete removes a Provider VDC
// The provider VDC must be disabled for deletion to succeed
// Deletion will also fail if the Provider VDC is backing other resources, such as organization VDCs
func (pvdc *ProviderVdcExtended) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] ProviderVdc.Delete")

	if pvdc.IsEnabled() {
		return Task{}, fmt.Errorf("provider VDC %s is enabled - can't delete", pvdc.VMWProviderVdc.Name)
	}
	// Return the task
	return pvdc.client.ExecuteTaskRequest(pvdc.VMWProviderVdc.HREF, http.MethodDelete,
		"", "error deleting provider VDC: %s", nil)
}

// Update can change some of the provider VDC internals
// In practical terms, only name and description are guaranteed to be changed through this method.
// The other admitted changes need to go through separate API calls
func (pvdc *ProviderVdcExtended) Update() error {

	resp, err := pvdc.client.executeJsonRequest(pvdc.VMWProviderVdc.HREF, http.MethodPut, pvdc.VMWProviderVdc,
		"error updating provider VDC: %s")

	if err != nil {
		return err
	}
	defer closeBody(resp)

	return pvdc.checkProgress("updating")
}

// Rename changes name and/or description from a provider VDC
func (pvdc *ProviderVdcExtended) Rename(name, description string) error {
	if name == "" {
		return fmt.Errorf("provider VDC name cannot be empty")
	}
	pvdc.VMWProviderVdc.Name = name
	pvdc.VMWProviderVdc.Description = description
	return pvdc.Update()
}

// AddResourcePools adds resource pools to the Provider VDC
func (pvdc *ProviderVdcExtended) AddResourcePools(resourcePools []*ResourcePool) error {
	util.Logger.Printf("[TRACE] ProviderVdc.AddResourcePools")

	href, err := url.JoinPath(pvdc.VMWProviderVdc.HREF, "action", "updateResourcePools")
	if err != nil {
		return err
	}

	var items []*types.VimObjectRef

	for _, rp := range resourcePools {
		vcenterUrl, err := rp.vcenter.GetVimServerUrl()
		if err != nil {
			return err
		}
		item := types.VimObjectRef{
			MoRef:         rp.ResourcePool.Moref,
			VimObjectType: "RESOURCE_POOL",
			VimServerRef: &types.Reference{
				HREF: vcenterUrl,
				ID:   extractUuid(rp.vcenter.VSphereVCenter.VcId),
				Name: rp.vcenter.VSphereVCenter.Name,
				Type: "application/vnd.vmware.admin.vmwvirtualcenter+xml",
			},
		}
		items = append(items, &item)
	}

	input := types.AddResourcePool{VimObjectRef: items}

	resp, err := pvdc.client.executeJsonRequest(href, http.MethodPost, input, "error updating provider VDC resource pools: %s")
	if err != nil {
		return err
	}
	task := NewTask(pvdc.client)
	err = decodeBody(types.BodyTypeJSON, resp, task.Task)
	if err != nil {
		return err
	}

	defer closeBody(resp)
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}
	return pvdc.Refresh()
}

// DeleteResourcePools removes resource pools from the Provider VDC
func (pvdc *ProviderVdcExtended) DeleteResourcePools(resourcePools []*ResourcePool) error {
	util.Logger.Printf("[TRACE] ProviderVdc.DeleteResourcePools")

	href, err := url.JoinPath(pvdc.VMWProviderVdc.HREF, "action", "updateResourcePools")
	if err != nil {
		return err
	}

	usedResourcePools, err := pvdc.GetResourcePools()
	if err != nil {
		return fmt.Errorf("error retrieving used resource pools: %s", err)
	}

	var items []*types.Reference

	for _, rp := range resourcePools {

		var foundUsed *types.QueryResultResourcePoolRecordType
		for _, urp := range usedResourcePools {
			if rp.ResourcePool.Moref == urp.Moref {
				foundUsed = urp
				break
			}
		}
		if foundUsed == nil {
			return fmt.Errorf("resource pool %s not found in provider VDC %s", rp.ResourcePool.Name, pvdc.VMWProviderVdc.Name)
		}
		if foundUsed.IsPrimary {
			return fmt.Errorf("resource pool %s (%s) can not be removed, because it is the primary one for provider VDC %s",
				rp.ResourcePool.Name, rp.ResourcePool.Moref, pvdc.VMWProviderVdc.Name)
		}
		if foundUsed.IsEnabled {
			err = disableResourcePool(pvdc.client, foundUsed.HREF)
			if err != nil {
				return fmt.Errorf("error disabling resource pool %s: %s", foundUsed.Name, err)
			}
		}

		item := types.Reference{
			HREF: foundUsed.HREF,
			ID:   extractUuid(foundUsed.HREF),
			Name: foundUsed.Name,
			Type: "application/vnd.vmware.admin.vmwProviderVdcResourcePool+xml",
		}
		items = append(items, &item)
	}

	input := types.DeleteResourcePool{ResourcePoolRefs: items}

	resp, err := pvdc.client.executeJsonRequest(href, http.MethodPost, input, "error removing resource pools from provider VDC: %s")
	if err != nil {
		return err
	}
	defer closeBody(resp)
	task := NewTask(pvdc.client)
	err = decodeBody(types.BodyTypeJSON, resp, task.Task)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}
	return pvdc.Refresh()
}

// GetResourcePools returns the Resource Pools belonging to this provider VDC
func (pvdc *ProviderVdcExtended) GetResourcePools() ([]*types.QueryResultResourcePoolRecordType, error) {
	resourcePools, err := pvdc.client.cumulativeQuery(types.QtResourcePool, nil, map[string]string{
		"type":          types.QtResourcePool,
		"filter":        fmt.Sprintf("providerVdc==%s", url.QueryEscape(extractUuid(pvdc.VMWProviderVdc.HREF))),
		"filterEncoded": "true",
	})
	if err != nil {
		return nil, fmt.Errorf("could not get the Resource pool: %s", err)
	}
	return resourcePools.Results.ResourcePoolRecord, nil
}

// disableResourcePool disables a resource pool while it is assigned to a provider VDC
// Calling this function is a prerequisite to removing a resource pool from a provider VDC
func disableResourcePool(client *Client, resourcePoolHref string) error {
	href, err := url.JoinPath(resourcePoolHref, "action", "disable")
	if err != nil {
		return err
	}
	return client.ExecuteRequestWithoutResponse(href, http.MethodPost, "", "error disabling resource pool: %s", nil)
}

// AddStorageProfiles adds the given storage profiles in this provider VDC
func (pvdc *ProviderVdcExtended) AddStorageProfiles(storageProfileNames []string) error {
	href, err := url.JoinPath(pvdc.VMWProviderVdc.HREF, "storageProfiles")
	if err != nil {
		return err
	}

	addStorageProfiles := &types.AddStorageProfiles{AddStorageProfile: storageProfileNames}

	resp, err := pvdc.client.executeJsonRequest(href, http.MethodPost, addStorageProfiles,
		"error adding storage profiles to provider VDC: %s")
	if err != nil {
		return err
	}

	defer closeBody(resp)

	return pvdc.checkProgress("adding storage profiles")
}

func (pvdc *ProviderVdcExtended) checkProgress(label string) error {
	// Let's keep this timeout as a precaution against an infinite wait
	timeout := 2 * time.Minute
	start := time.Now()
	err := pvdc.Refresh()
	if err != nil {
		return err
	}

	var elapsed time.Duration
	for ResourceInProgress(pvdc.VMWProviderVdc.Tasks) {
		err = pvdc.Refresh()
		if err != nil {
			return fmt.Errorf("error %s: %s", label, err)
		}
		time.Sleep(200 * time.Millisecond)
		elapsed = time.Since(start)
		if elapsed > timeout {
			return fmt.Errorf("error %s within %s", label, timeout)
		}
	}
	util.Logger.Printf("[ProviderVdcExtended.checkProgress] called by %s - running %s - elapsed: %s\n",
		util.CallFuncName(), label, elapsed)
	return nil
}

// disableStorageProfile disables a storage profile while it is assigned to a provider VDC
// Calling this function is a prerequisite to removing a storage profile from a provider VDC
func disableStorageProfile(client *Client, storageProfileHref string) error {
	disablePayload := &types.EnableStorageProfile{Enabled: false}
	resp, err := client.executeJsonRequest(storageProfileHref, http.MethodPut, disablePayload,
		"error disabling storage profile in provider VDC: %s")

	defer closeBody(resp)
	return err
}

// DeleteStorageProfiles removes storage profiles from the Provider VDC
func (pvdc *ProviderVdcExtended) DeleteStorageProfiles(storageProfiles []string) error {
	util.Logger.Printf("[TRACE] ProviderVdc.DeleteStorageProfiles")

	href, err := url.JoinPath(pvdc.VMWProviderVdc.HREF, "storageProfiles")
	if err != nil {
		return err
	}

	usedStorageProfileRefs := pvdc.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile

	var toBeDeleted []*types.Reference

	for _, sp := range storageProfiles {
		var foundUsed bool
		for _, usp := range usedStorageProfileRefs {
			if sp == usp.Name {
				foundUsed = true
				toBeDeleted = append(toBeDeleted, &types.Reference{HREF: usp.HREF})
				break
			}
		}
		if !foundUsed {
			return fmt.Errorf("storage profile %s not found in provider VDC %s", sp, pvdc.VMWProviderVdc.Name)
		}
	}

	for _, sp := range toBeDeleted {
		err = disableStorageProfile(pvdc.client, sp.HREF)
		if err != nil {
			return fmt.Errorf("error disabling storage profile %s from provider VDC %s: %s", sp.Name, pvdc.VMWProviderVdc.Name, err)
		}
	}
	input := &types.RemoveStorageProfile{RemoveStorageProfile: toBeDeleted}

	resp, err := pvdc.client.executeJsonRequest(href, http.MethodPost, input, "error removing storage profiles from provider VDC: %s")
	if err != nil {
		return err
	}
	defer closeBody(resp)
	task := NewTask(pvdc.client)
	err = decodeBody(types.BodyTypeJSON, resp, task.Task)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}
	return pvdc.Refresh()
}
