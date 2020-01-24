/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type AdminVdc struct {
	AdminVdc *types.AdminVdc
	client   *Client
}

func NewAdminVdc(cli *Client) *AdminVdc {
	return &AdminVdc{
		AdminVdc: new(types.AdminVdc),
		client:   cli,
	}
}

// Given an adminVdc with a valid HREF, the function refresh the adminVdc
// and updates the adminVdc data. Returns an error on failure
// Users should use refresh whenever they suspect
// a stale VDC due to the creation/update/deletion of a resource
// within the the VDC itself.
func (adminVdc *AdminVdc) Refresh() error {
	if *adminVdc == (AdminVdc{}) || adminVdc.AdminVdc.HREF == "" {
		return fmt.Errorf("cannot refresh, Object is empty or HREF is empty")
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	unmarshalledAdminVdc := &types.AdminVdc{}

	_, err := adminVdc.client.ExecuteRequest(adminVdc.AdminVdc.HREF, http.MethodGet,
		"", "error refreshing VDC: %s", nil, unmarshalledAdminVdc)
	if err != nil {
		return err
	}
	adminVdc.AdminVdc = unmarshalledAdminVdc

	return nil
}

// UpdateAsync updates VDC from current VDC struct contents.
// Any differences that may be legally applied will be updated.
// Returns an error if the call to vCD fails.
// API Documentation: https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/operations/PUT-Vdc.html
func (adminVdc *AdminVdc) UpdateAsync() (Task, error) {

	adminVdc.AdminVdc.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return adminVdc.client.ExecuteTaskRequest(adminVdc.AdminVdc.HREF, http.MethodPut,
		types.MimeAdminVDC, "error updating VDC: %s", adminVdc.AdminVdc)
}

// Update function updates an Admin VDC from current VDC struct contents.
// Any differences that may be legally applied will be updated.
// Returns an empty AdminVdc struct and error if the call to vCD fails.
// API Documentation: https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/operations/PUT-Vdc.html
func (adminVdc *AdminVdc) Update() (AdminVdc, error) {
	task, err := adminVdc.UpdateAsync()
	if err != nil {
		return AdminVdc{}, err
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return AdminVdc{}, err
	}

	err = adminVdc.Refresh()
	if err != nil {
		return AdminVdc{}, err
	}

	return *adminVdc, nil
}

type FunctionsVersions struct {
	Type             string
	supportedVersion string
	CreateVdc        func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error)
	CreateVdcAsync   func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error)
}

var VdcV290 = FunctionsVersions{
	Type:             "VDC",
	supportedVersion: "29.0",
	CreateVdc:        createVdc,
	CreateVdcAsync:   createVdcAsync,
}

var VdcV320 = FunctionsVersions{
	Type:             "VDC",
	supportedVersion: "32.0",
	CreateVdc:        createVdcV32,
	CreateVdcAsync:   createVdcAsyncV32,
}

var FunctionsByVersion = map[string]FunctionsVersions{
	"vdc29.0": VdcV290,
	"vdc30.0": VdcV290,
	"vdc31.0": VdcV290,
	"vdc32.0": VdcV320,
	"vdc33.0": VdcV320,
}

func (adminOrg *AdminOrg) CreateOrgVdc(vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	apiVersion, err := adminOrg.client.maxSupportedVersion()
	if err != nil {
		return nil, err
	}
	realFunction := FunctionsByVersion["vdc"+apiVersion]
	if realFunction.CreateVdc == nil {
		return nil, fmt.Errorf("function CreateVdc is not defined for %s", "vdc"+apiVersion)
	}
	return realFunction.CreateVdc(adminOrg, vdcConfiguration)
}

func (adminOrg *AdminOrg) CreateOrgVdcAsync(vdcConfiguration *types.VdcConfiguration) (Task, error) {
	apiVersion, err := adminOrg.client.maxSupportedVersion()
	if err != nil {
		return Task{}, err
	}
	realFunction := FunctionsByVersion["vdc"+apiVersion]
	if realFunction.CreateVdcAsync == nil {
		return Task{}, fmt.Errorf("function CreateVdcAsync is not defined for %s", "vdc"+apiVersion)
	}
	return realFunction.CreateVdcAsync(adminOrg, vdcConfiguration)
}

func createVdc(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	err := adminOrg.CreateVdcWait(vdcConfiguration)
	if err != nil {
		return nil, err
	}

	vdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		return nil, err
	}
	return vdc, nil
}

func createVdcAsync(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error) {
	return adminOrg.CreateVdc(vdcConfiguration)
}

func createVdcV32(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	task, err := createVdcAsyncV32(adminOrg, vdcConfiguration)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("couldn't finish creating vdc %s", err)
	}

	vdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		return nil, err
	}
	return vdc, nil
}

func createVdcAsyncV32(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error) {
	err := validateVdcConfigurationV32(*vdcConfiguration)
	if err != nil {
		return Task{}, err
	}

	vdcConfiguration.Xmlns = types.XMLNamespaceVCloud

	vdcCreateHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("error parsing admin org url: %s", err)
	}
	vdcCreateHREF.Path += "/vdcsparams"

	adminVdc := NewAdminVdc(adminOrg.client)

	_, err = adminOrg.client.ExecuteRequestWithApiVersion(vdcCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.createVdcParams+xml", "error retrieving vdc: %s",
		vdcConfiguration, adminVdc.AdminVdc,
		adminOrg.client.GetSpecificApiVersionOnCondition(">= 32.0", "32.0"))
	if err != nil {
		return Task{}, err
	}

	// Return the task
	task := NewTask(adminOrg.client)
	task.Task = adminVdc.AdminVdc.Tasks.Task[0]
	return *task, nil
}

func validateVdcConfigurationV32(vdcDefinition types.VdcConfiguration) error {
	if vdcDefinition.Name == "" {
		return errors.New("VdcConfiguration missing required field: Name")
	}
	if vdcDefinition.AllocationModel == "" {
		return errors.New("VdcConfiguration missing required field: AllocationModel")
	}
	if vdcDefinition.ComputeCapacity == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity")
	}
	if len(vdcDefinition.ComputeCapacity) != 1 {
		return errors.New("VdcConfiguration invalid field: ComputeCapacity must only have one element")
	}
	if vdcDefinition.ComputeCapacity[0] == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0]")
	}
	if vdcDefinition.ComputeCapacity[0].CPU == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].CPU")
	}
	if vdcDefinition.ComputeCapacity[0].CPU.Units == "" {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].CPU.Units")
	}
	if vdcDefinition.ComputeCapacity[0].Memory == nil {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].Memory")
	}
	if vdcDefinition.ComputeCapacity[0].Memory.Units == "" {
		return errors.New("VdcConfiguration missing required field: ComputeCapacity[0].Memory.Units")
	}
	if vdcDefinition.VdcStorageProfile == nil || len(vdcDefinition.VdcStorageProfile) == 0 {
		return errors.New("VdcConfiguration missing required field: VdcStorageProfile")
	}
	if vdcDefinition.VdcStorageProfile[0].Units == "" {
		return errors.New("VdcConfiguration missing required field: VdcStorageProfile.Units")
	}
	if vdcDefinition.ProviderVdcReference == nil {
		return errors.New("VdcConfiguration missing required field: ProviderVdcReference")
	}
	if vdcDefinition.ProviderVdcReference.HREF == "" {
		return errors.New("VdcConfiguration missing required field: ProviderVdcReference.HREF")
	}
	if vdcDefinition.AllocationModel == "Flex" && vdcDefinition.IsElastic == nil {
		return errors.New("VdcConfiguration missing required field: IsElastic")
	}
	if vdcDefinition.AllocationModel == "Flex" && vdcDefinition.IncludeMemoryOverhead == nil {
		return errors.New("VdcConfiguration missing required field: IncludeMemoryOverhead")
	}
	return nil
}
