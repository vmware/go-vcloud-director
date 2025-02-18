/*
 * Copyright 2023 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const labelVirtualCenter = "vCenter Server"

// vCenter task is sometimes unreliable and trying to refresh it immediately after it becomes
// connected causes a "BUSY_ENTITY" error (which has a few different messages)
var maximumVcenterRetryTime = 120 * time.Second                                         // The maximum time a single operation will be retried before giving up
var vCenterEntityBusyRegexp = regexp.MustCompile(`(is currently busy|400|BUSY_ENTITY)`) // Regexp to match entity busy error

type VCenter struct {
	VSphereVCenter *types.VSphereVirtualCenter
	client         *VCDClient
}

// wrap is a hidden helper that facilitates the usage of a generic CRUD function
//
//lint:ignore U1000 this method is used in generic functions, but annoys staticcheck
func (v VCenter) wrap(inner *types.VSphereVirtualCenter) *VCenter {
	v.VSphereVCenter = inner
	return &v
}

// CreateVcenter adds new vCenter connection
func (vcdClient *VCDClient) CreateVcenter(config *types.VSphereVirtualCenter) (*VCenter, error) {
	c := crudConfig{
		entityLabel: labelVirtualCenter,
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
	}
	outerType := VCenter{client: vcdClient}
	return createOuterEntity(&vcdClient.Client, outerType, c, config)
}

// CreateVcenterAsync adds new vCenter and returns its task for tracking
func (vcdClient *VCDClient) CreateVcenterAsync(config *types.VSphereVirtualCenter) (*Task, error) {
	c := crudConfig{
		entityLabel: labelVirtualCenter,
		endpoint:    types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
	}
	return createInnerEntityAsync(&vcdClient.Client, c, config)
}

// GetAllVCenters retrieves all vCenter servers based on optional query filtering
func (vcdClient *VCDClient) GetAllVCenters(queryParams url.Values) ([]*VCenter, error) {
	c := crudConfig{
		entityLabel:     labelVirtualCenter,
		endpoint:        types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		queryParameters: queryParams,
	}

	outerType := VCenter{client: vcdClient}
	return getAllOuterEntities(&vcdClient.Client, outerType, c)
}

// GetVCenterByName retrieves vCenter server by name
func (vcdClient *VCDClient) GetVCenterByName(name string) (*VCenter, error) {
	if name == "" {
		return nil, fmt.Errorf("%s lookup requires name", labelVirtualCenter)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", "name=="+name)

	vCenters, err := vcdClient.GetAllVCenters(queryParams)
	if err != nil {
		return nil, err
	}

	singleEntity, err := oneOrError("name", name, vCenters)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}

// GetVCenterByUrl looks up if there is an existing vCenter added with a given URL
func (vcdClient *VCDClient) GetVCenterByUrl(vcUrl string) (*VCenter, error) {
	if vcUrl == "" {
		return nil, fmt.Errorf("%s lookup requires URL", labelVirtualCenter)
	}

	// API filtering by URL is not supported so relying on local filtering
	vCenters, err := vcdClient.GetAllVCenters(nil)
	if err != nil {
		return nil, err
	}

	filteredEntities := make([]*VCenter, 0)
	for _, vc := range vCenters {
		if vc.VSphereVCenter.Url == vcUrl {
			filteredEntities = append(filteredEntities, vc)
		}

	}

	singleEntity, err := oneOrError("Url", vcUrl, filteredEntities)
	if err != nil {
		return nil, err
	}

	return singleEntity, nil
}

// GetVCenterById retrieves vCenter server by ID
func (vcdClient *VCDClient) GetVCenterById(id string) (*VCenter, error) {
	c := crudConfig{
		entityLabel:    labelVirtualCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{id},
	}

	outerType := VCenter{client: vcdClient}
	return getOuterEntity(&vcdClient.Client, outerType, c)
}

// Update given vCenter configuration
func (v *VCenter) Update(TmNsxtManagerConfig *types.VSphereVirtualCenter) (*VCenter, error) {
	c := crudConfig{
		entityLabel:    labelVirtualCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{v.VSphereVCenter.VcId},
	}
	outerType := VCenter{client: v.client}
	return updateOuterEntity(&v.client.Client, outerType, c, TmNsxtManagerConfig)
}

// Delete vCenter configuration
func (v *VCenter) Delete() error {
	c := crudConfig{
		entityLabel:    labelVirtualCenter,
		endpoint:       types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointVirtualCenters,
		endpointParams: []string{v.VSphereVCenter.VcId},
	}
	return deleteEntityById(&v.client.Client, c)
}

// Disable is an update shortcut for disabling vCenter
func (v *VCenter) Disable() error {
	v.VSphereVCenter.IsEnabled = false
	_, err := v.Update(v.VSphereVCenter)
	return err
}

func (v VCenter) GetVimServerUrl() (string, error) {
	return url.JoinPath(v.client.Client.rootVcdHref(), "api", "admin", "extension", "vimServer", extractUuid(v.VSphereVCenter.VcId))
}

// Refresh vCenter structure
func (v *VCenter) Refresh() error {
	// Retrieval endpoints by Name and by ID return differently formated url (the by Id one returns
	// URL with port http://host:443, while the one by name - doesn't). Using the same getByName to
	// match format everywhere

	// newVcenter, err := v.client.GetVCenterById(v.VSphereVCenter.VcId)
	newVcenter, err := v.client.GetVCenterByName(v.VSphereVCenter.Name) // TODO: TM: use above retrieval by ID
	if err != nil {
		return fmt.Errorf("error refreshing vCenter: %s", err)
	}

	v.VSphereVCenter = newVcenter.VSphereVCenter
	return nil
}

// RefreshVcenter triggers a refresh operation on vCenter that syncs up vCenter components such as
// supervisors
// It uses legacy endpoint as there is no OpenAPI endpoint for this operation
func (v *VCenter) RefreshVcenter() error {
	return runWithRetry(v.refreshVcenter, vCenterEntityBusyRegexp, maximumVcenterRetryTime)
}

func (v *VCenter) refreshVcenter() error {
	refreshUrl, err := url.JoinPath(v.client.Client.rootVcdHref(), "api", "admin", "extension", "vimServer", extractUuid(v.VSphereVCenter.VcId), "action", "refresh")
	if err != nil {
		return fmt.Errorf("error building refresh path: %s", err)
	}

	resp, err := v.client.Client.executeJsonRequest(refreshUrl, http.MethodPost, nil, "error triggering vCenter refresh: %s")
	if err != nil {
		return err
	}
	defer closeBody(resp)
	task := NewTask(&v.client.Client)
	err = decodeBody(types.BodyTypeJSON, resp, task.Task)
	if err != nil {
		return fmt.Errorf("error triggering retrieving task: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting task completion: %s", err)
	}

	return nil
}

// RefreshStorageProfiles triggers a refresh operation on vCenter that syncs up vCenter components
// such as supervisors
// It uses legacy endpoint as there is no OpenAPI endpoint for this operation
func (v *VCenter) RefreshStorageProfiles() error {
	return runWithRetry(v.refreshStorageProfiles, vCenterEntityBusyRegexp, maximumVcenterRetryTime)
}

func (v *VCenter) refreshStorageProfiles() error {
	refreshUrl, err := url.JoinPath(v.client.Client.rootVcdHref(), "api", "admin", "extension", "vimServer", extractUuid(v.VSphereVCenter.VcId), "action", "refreshStorageProfiles")
	if err != nil {
		return fmt.Errorf("error building storage policy refresh path: %s", err)
	}

	resp, err := v.client.Client.executeJsonRequest(refreshUrl, http.MethodPost, nil, "error triggering vCenter refresh storage policy: %s")
	if err != nil {
		return err
	}
	defer closeBody(resp)
	task := NewTask(&v.client.Client)
	err = decodeBody(types.BodyTypeJSON, resp, task.Task)
	if err != nil {
		return fmt.Errorf("error triggering retrieving task: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("error waiting task completion: %s", err)
	}

	return nil
}

func runWithRetry(runOperation func() error, errRegexp *regexp.Regexp, duration time.Duration) error {
	startTime := time.Now()
	endTime := startTime.Add(duration)
	util.Logger.Printf("[DEBUG] runWithRetry - running with retry for %f seconds if error contains '%s' ", duration.Seconds(), errRegexp)
	count := 1
	for {
		err := runOperation()
		util.Logger.Printf("[DEBUG] runWithRetry - ran attempt %d, got error: %s ", count, err)
		// Operation had no error - it succeeded
		if err == nil {
			util.Logger.Printf("[DEBUG] runWithRetry - no error occurred after attempt %d, got error: %s ", count, err)
			return nil
		}
		// If there is an error, but it doesn't contain the retryIfErrContains value - exit it
		if !errRegexp.MatchString(err.Error()) {
			util.Logger.Printf("[DEBUG] runWithRetry - returning error after attempt %d, got error: %s ", count, err)
			return err
		}

		// If time limit is exceeded - return error containing statistics and original error
		if time.Now().After(endTime) {
			util.Logger.Printf("[DEBUG] runWithRetry - exceeded time after attempt %d, got error: %s ", count, err)
			return fmt.Errorf("error attempting to wait until error does not contain '%s' after %f seconds: %s", errRegexp, duration.Seconds(), err)
		}

		// Sleep and continue
		util.Logger.Printf("[DEBUG] runWithRetry - sleeping after attempt %d, will retry", count)
		// Sleep 2 seconds and attempt once more if the timeout is not excdeeded
		time.Sleep(2 * time.Second)
		count++
	}
}
