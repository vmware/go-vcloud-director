// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

type AdminVdc struct {
	AdminVdc *types.AdminVdc
	client   *Client
	parent   organization
}

func NewAdminVdc(cli *Client) *AdminVdc {
	return &AdminVdc{
		AdminVdc: new(types.AdminVdc),
		client:   cli,
	}
}

// vdcVersionedFuncs is a generic representation of VDC CRUD operations across multiple versions
type vdcVersionedFuncs struct {
	SupportedVersion string
	CreateVdc        func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error)
	CreateVdcAsync   func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error)
	UpdateVdc        func(adminVdc *AdminVdc) (*AdminVdc, error)
	UpdateVdcAsync   func(adminVdc *AdminVdc) (Task, error)
}

// VDC function mapping for API version 32.0 (from vCD 9.7)
var vdcVersionedFuncsV97 = vdcVersionedFuncs{
	SupportedVersion: "32.0",
	CreateVdc:        createVdcV97,
	CreateVdcAsync:   createVdcAsyncV97,
	UpdateVdc:        updateVdcV97,
	UpdateVdcAsync:   updateVdcAsyncV97,
}

// vdcVersionedFuncsByVcdVersion is a map of VDC functions by vCD version
var vdcVersionedFuncsByVcdVersion = map[string]vdcVersionedFuncs{
	"vdc9.7": vdcVersionedFuncsV97,

	// If we add a new function to this list, we also need to update the "default" entry
	// The "default" entry will hold the highest currently available function
	"default": vdcVersionedFuncsV97,
}

// getVdcVersionedFuncsByVdcVersion is a wrapper function that retrieves the requested versioned VDC function
// When the wanted version does  not exist in the map, it returns the highest available one.
func getVdcVersionedFuncsByVdcVersion(version string) vdcVersionedFuncs {
	f, ok := vdcVersionedFuncsByVcdVersion[version]
	if ok {
		return f
	} else {
		return vdcVersionedFuncsByVcdVersion["default"]
	}
}

// GetAdminVdcByName function uses a valid VDC name and returns a admin VDC object.
// If no VDC is found, then it returns an empty VDC and no error.
// Otherwise it returns an empty VDC and an error.
// Deprecated: Use adminOrg.GetAdminVDCByName
func (adminOrg *AdminOrg) GetAdminVdcByName(vdcname string) (AdminVdc, error) {
	for _, vdcs := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if vdcs.Name == vdcname {
			adminVdc := NewAdminVdc(adminOrg.client)
			_, err := adminOrg.client.ExecuteRequest(vdcs.HREF, http.MethodGet,
				"", "error getting vdc: %s", nil, adminVdc.AdminVdc)
			return *adminVdc, err
		}
	}
	return AdminVdc{}, nil
}

// GetAdminVDCByHref retrieves a VDC using a direct call with the HREF
func (adminOrg *AdminOrg) GetAdminVDCByHref(vdcHref string) (*AdminVdc, error) {
	adminVdc := NewAdminVdc(adminOrg.client)
	adminVdc.parent = adminOrg
	_, err := adminOrg.client.ExecuteRequest(vdcHref, http.MethodGet,
		"", "error getting vdc: %s", nil, adminVdc.AdminVdc)

	if err != nil {
		return nil, err
	}
	return adminVdc, nil
}

// GetAdminVDCByName finds an Admin VDC by Name
// On success, returns a pointer to the AdminVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminVDCByName(vdcName string, refresh bool) (*AdminVdc, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, vdc := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if vdc.Name == vdcName {
			return adminOrg.GetAdminVDCByHref(vdc.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetAdminVDCById finds an Admin VDC by ID
// On success, returns a pointer to the AdminVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminVDCById(vdcId string, refresh bool) (*AdminVdc, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, vdc := range adminOrg.AdminOrg.Vdcs.Vdcs {
		if equalIds(vdcId, vdc.ID, vdc.HREF) {
			return adminOrg.GetAdminVDCByHref(vdc.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetAdminVDCByNameOrId finds an Admin VDC by Name Or ID
// On success, returns a pointer to the AdminVdc structure and a nil error
// On failure, returns a nil pointer and an error
func (adminOrg *AdminOrg) GetAdminVDCByNameOrId(identifier string, refresh bool) (*AdminVdc, error) {
	getByName := func(name string, refresh bool) (interface{}, error) {
		return adminOrg.GetAdminVDCByName(name, refresh)
	}
	getById := func(id string, refresh bool) (interface{}, error) { return adminOrg.GetAdminVDCById(id, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*AdminVdc), err
}

// CreateVdc creates a VDC with the given params under the given organization.
// Returns an AdminVdc.
// API Documentation: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/POST-VdcConfiguration.html
// Deprecated in favor of adminOrg.CreateOrgVdcAsync
func (adminOrg *AdminOrg) CreateVdc(vdcConfiguration *types.VdcConfiguration) (Task, error) {
	err := validateVdcConfiguration(vdcConfiguration)
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

	_, err = adminOrg.client.ExecuteRequest(vdcCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.createVdcParams+xml", "error creating VDC: %s", vdcConfiguration, adminVdc.AdminVdc)
	if err != nil {
		return Task{}, err
	}

	// Return the task
	task := NewTask(adminOrg.client)
	if adminVdc.AdminVdc.Tasks == nil || len(adminVdc.AdminVdc.Tasks.Task) == 0 {
		return Task{}, fmt.Errorf("no task found after VDC %s creation", vdcConfiguration.Name)
	}
	task.Task = adminVdc.AdminVdc.Tasks.Task[0]
	return *task, nil
}

// Creates the VDC and waits for the asynchronous task to complete.
// Deprecated in favor of adminOrg.CreateOrgVdc
func (adminOrg *AdminOrg) CreateVdcWait(vdcDefinition *types.VdcConfiguration) error {
	task, err := adminOrg.CreateVdc(vdcDefinition)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("couldn't finish creating VDC %s", err)
	}
	return nil
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
	apiVersion, err := adminVdc.client.MaxSupportedVersion()
	if err != nil {
		return Task{}, err
	}
	vdcFunctions := getVdcVersionedFuncsByVdcVersion("vdc" + apiVersionToVcdVersion[apiVersion])
	if vdcFunctions.UpdateVdcAsync == nil {
		return Task{}, fmt.Errorf("function UpdateVdcAsync is not defined for %s", "vdc"+apiVersion)
	}
	util.Logger.Printf("[DEBUG] UpdateAsync call function for version %s", vdcFunctions.SupportedVersion)

	return vdcFunctions.UpdateVdcAsync(adminVdc)

}

// Update function updates an Admin VDC from current VDC struct contents.
// Any differences that may be legally applied will be updated.
// Returns an empty AdminVdc struct and error if the call to vCD fails.
// API Documentation: https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/operations/PUT-Vdc.html
func (adminVdc *AdminVdc) Update() (AdminVdc, error) {
	apiVersion, err := adminVdc.client.MaxSupportedVersion()
	if err != nil {
		return AdminVdc{}, err
	}

	vdcFunctions := getVdcVersionedFuncsByVdcVersion("vdc" + apiVersionToVcdVersion[apiVersion])
	if vdcFunctions.UpdateVdc == nil {
		return AdminVdc{}, fmt.Errorf("function UpdateVdc is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] Update call function for version %s", vdcFunctions.SupportedVersion)

	// Explicitly remove ResourcePoolRefs because it cannot be set and breaks Go marshaling bug
	// https://github.com/golang/go/issues/9519
	adminVdc.AdminVdc.ResourcePoolRefs = nil

	updatedAdminVdc, err := vdcFunctions.UpdateVdc(adminVdc)
	if err != nil {
		return AdminVdc{}, err
	}
	return *updatedAdminVdc, err
}

// CreateOrgVdc creates a VDC with the given params under the given organization
// and waits for the asynchronous task to complete.
// Returns an AdminVdc pointer and an error.
func (adminOrg *AdminOrg) CreateOrgVdc(vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	apiVersion, err := adminOrg.client.MaxSupportedVersion()
	if err != nil {
		return nil, err
	}
	vdcFunctions := getVdcVersionedFuncsByVdcVersion("vdc" + apiVersionToVcdVersion[apiVersion])
	if vdcFunctions.CreateVdc == nil {
		return nil, fmt.Errorf("function CreateVdc is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] CreateOrgVdc call function for version %s", vdcFunctions.SupportedVersion)
	return vdcFunctions.CreateVdc(adminOrg, vdcConfiguration)
}

// CreateOrgVdcAsync creates a VDC with the given params under the given organization.
// Returns a Task and an error.
func (adminOrg *AdminOrg) CreateOrgVdcAsync(vdcConfiguration *types.VdcConfiguration) (Task, error) {
	apiVersion, err := adminOrg.client.MaxSupportedVersion()
	if err != nil {
		return Task{}, err
	}
	vdcFunctions := getVdcVersionedFuncsByVdcVersion("vdc" + apiVersionToVcdVersion[apiVersion])
	if vdcFunctions.CreateVdcAsync == nil {
		return Task{}, fmt.Errorf("function CreateVdcAsync is not defined for %s", "vdc"+apiVersion)
	}

	util.Logger.Printf("[DEBUG] CreateOrgVdcAsync call function for version %s", vdcFunctions.SupportedVersion)

	return vdcFunctions.CreateVdcAsync(adminOrg, vdcConfiguration)
}

// updateVdcAsyncV97 updates a VDC with the given params. Supports Flex type allocation.
// Needs vCD 9.7+ to work. Returns a Task and an error.
func updateVdcAsyncV97(adminVdc *AdminVdc) (Task, error) {
	util.Logger.Printf("[TRACE] updateVdcAsyncV97 called %#v", *adminVdc)
	adminVdc.AdminVdc.Xmlns = types.XMLNamespaceVCloud

	// Return the task
	return adminVdc.client.ExecuteTaskRequest(adminVdc.AdminVdc.HREF, http.MethodPut,
		types.MimeAdminVDC, "error updating VDC: %s", adminVdc.AdminVdc)
}

// updateVdcV97 updates a VDC with the given params
// and waits for the asynchronous task to complete. Supports Flex type allocation.
// Needs vCD 9.7+ to work. Returns an AdminVdc pointer and an error.
func updateVdcV97(adminVdc *AdminVdc) (*AdminVdc, error) {
	util.Logger.Printf("[TRACE] updateVdcV97 called %#v", *adminVdc)
	task, err := updateVdcAsyncV97(adminVdc)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, err
	}
	err = adminVdc.Refresh()
	if err != nil {
		return nil, err
	}
	return adminVdc, nil
}

// createVdcV97 creates a VDC with the given params under the given organization
// and waits for the asynchronous task to complete. Supports Flex type allocation.
// Needs vCD 9.7+ to work. Returns a Vdc pointer and error.
func createVdcV97(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	util.Logger.Printf("[TRACE] createVdcV97 called %#v", *vdcConfiguration)
	task, err := createVdcAsyncV97(adminOrg, vdcConfiguration)
	if err != nil {
		return nil, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return nil, fmt.Errorf("couldn't finish creating VDC %s", err)
	}

	vdc, err := adminOrg.GetVDCByName(vdcConfiguration.Name, true)
	if err != nil {
		return nil, err
	}
	return vdc, nil
}

// createVdcAsyncV97 creates a VDC with the given params under the given organization. Supports Flex type allocation.
// Needs vCD 9.7+ to work. Returns a Task and an error
func createVdcAsyncV97(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error) {
	util.Logger.Printf("[TRACE] createVdcAsyncV97 called %#v", *vdcConfiguration)
	err := validateVdcConfigurationV97(*vdcConfiguration)
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

	_, err = adminOrg.client.ExecuteRequest(vdcCreateHREF.String(), http.MethodPost,
		"application/vnd.vmware.admin.createVdcParams+xml", "error creating VDC: %s",
		vdcConfiguration, adminVdc.AdminVdc)
	if err != nil {
		return Task{}, err
	}

	// Return the task
	task := NewTask(adminOrg.client)
	if adminVdc.AdminVdc.Tasks == nil || len(adminVdc.AdminVdc.Tasks.Task) == 0 {
		return Task{}, fmt.Errorf("no task found after VDC %s creation", vdcConfiguration.Name)
	}
	task.Task = adminVdc.AdminVdc.Tasks.Task[0]
	return *task, nil
}

// validateVdcConfigurationV97 uses validateVdcConfiguration and additionally checks Flex dependent values
func validateVdcConfigurationV97(vdcDefinition types.VdcConfiguration) error {
	err := validateVdcConfiguration(&vdcDefinition)
	if err != nil {
		return err
	}
	if vdcDefinition.AllocationModel == "Flex" && vdcDefinition.IsElastic == nil {
		return errors.New("VdcConfiguration missing required field: IsElastic")
	}
	if vdcDefinition.AllocationModel == "Flex" && vdcDefinition.IncludeMemoryOverhead == nil {
		return errors.New("VdcConfiguration missing required field: IncludeMemoryOverhead")
	}
	return nil
}

// GetVappList returns the list of vApps for an Admin VDC
func (vdc *AdminVdc) GetVappList() []*types.ResourceReference {
	var list []*types.ResourceReference
	for _, resourceEntities := range vdc.AdminVdc.ResourceEntities {
		for _, resourceReference := range resourceEntities.ResourceEntity {
			if resourceReference.Type == types.MimeVApp {
				list = append(list, resourceReference)
			}
		}
	}
	return list
}

// UpdateStorageProfile updates VDC storage profile and returns refreshed VDC or error.
func (vdc *AdminVdc) UpdateStorageProfile(storageProfileId string, storageProfile *types.AdminVdcStorageProfile) (*types.AdminVdcStorageProfile, error) {
	if vdc.client.VCDHREF.String() == "" {
		return nil, fmt.Errorf("cannot update VDC storage profile, VCD HREF is unset")
	}

	queryUrl := vdc.client.VCDHREF
	queryUrl.Path += "/admin/vdcStorageProfile/" + storageProfileId

	storageProfile.Xmlns = types.XMLNamespaceVCloud
	updateAdminVdcStorageProfile := &types.AdminVdcStorageProfile{}

	_, err := vdc.client.ExecuteRequest(queryUrl.String(), http.MethodPut,
		types.MimeStorageProfile, "error updating VDC storage profile: %s", storageProfile, updateAdminVdcStorageProfile)
	if err != nil {
		return nil, fmt.Errorf("cannot update VDC storage profile, error: %s", err)
	}

	return updateAdminVdcStorageProfile, err
}

// AddStorageProfile adds a storage profile to a VDC
func (vdc *AdminVdc) AddStorageProfile(storageProfile *types.VdcStorageProfileConfiguration, description string) (Task, error) {
	if vdc.client.VCDHREF.String() == "" {
		return Task{}, fmt.Errorf("cannot add VDC storage profile, VCD HREF is unset")
	}

	href := vdc.AdminVdc.HREF + "/vdcStorageProfiles"

	var updateStorageProfile = types.UpdateVdcStorageProfiles{
		Xmlns:                types.XMLNamespaceVCloud,
		Name:                 storageProfile.ProviderVdcStorageProfile.Name,
		Description:          description,
		AddStorageProfile:    storageProfile,
		RemoveStorageProfile: nil,
	}

	task, err := vdc.client.ExecuteTaskRequest(href, http.MethodPost,
		types.MimeUpdateVdcStorageProfiles, "error adding VDC storage profile: %s", &updateStorageProfile)
	if err != nil {
		return Task{}, fmt.Errorf("cannot add VDC storage profile, error: %s", err)
	}

	return task, nil
}

// AddStorageProfileWait adds a storage profile to a VDC and return a refreshed VDC
func (vdc *AdminVdc) AddStorageProfileWait(storageProfile *types.VdcStorageProfileConfiguration, description string) error {
	task, err := vdc.AddStorageProfile(storageProfile, description)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}
	return vdc.Refresh()
}

// RemoveStorageProfile remove a storage profile from a VDC
func (vdc *AdminVdc) RemoveStorageProfile(storageProfileName string) (Task, error) {
	if vdc.client.VCDHREF.String() == "" {
		return Task{}, fmt.Errorf("cannot remove VDC storage profile: VCD HREF is unset")
	}

	var storageProfile *types.Reference
	for _, sp := range vdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile {
		if sp.Name == storageProfileName {
			storageProfile = sp
		}
	}
	if storageProfile == nil {
		return Task{}, fmt.Errorf("cannot remove VDC storage profile: storage profile '%s' not found in VDC", storageProfileName)
	}

	vdcStorageProfileDetails, err := vdc.client.GetStorageProfileByHref(storageProfile.HREF)
	if err != nil {
		return Task{}, fmt.Errorf("cannot retrieve VDC storage profile '%s' details: %s", storageProfileName, err)
	}
	if vdcStorageProfileDetails.Enabled != nil && *vdcStorageProfileDetails.Enabled {
		_, err = vdc.UpdateStorageProfile(extractUuid(storageProfile.HREF), &types.AdminVdcStorageProfile{
			Name:    vdcStorageProfileDetails.Name,
			Units:   vdcStorageProfileDetails.Units,
			Limit:   vdcStorageProfileDetails.Limit,
			Default: false,
			Enabled: addrOf(false),
			ProviderVdcStorageProfile: &types.Reference{
				HREF: vdcStorageProfileDetails.ProviderVdcStorageProfile.HREF,
			},
		},
		)
		if err != nil {
			return Task{}, fmt.Errorf("cannot disable VDC storage profile '%s': %s", storageProfileName, err)
		}
	}

	href := vdc.AdminVdc.HREF + "/vdcStorageProfiles"

	var updateStorageProfile = types.UpdateVdcStorageProfiles{
		Xmlns:                types.XMLNamespaceVCloud,
		Name:                 vdcStorageProfileDetails.Name,
		Description:          "",
		RemoveStorageProfile: storageProfile,
	}

	task, err := vdc.client.ExecuteTaskRequest(href, http.MethodPost,
		types.MimeUpdateVdcStorageProfiles, "error removing VDC storage profile: %s", &updateStorageProfile)
	if err != nil {
		return Task{}, fmt.Errorf("cannot remove VDC storage profile, error: %s", err)
	}

	return task, nil
}

// RemoveStorageProfileWait removes a storege profile from a VDC and returns a refreshed VDC or an error
func (vdc *AdminVdc) RemoveStorageProfileWait(storageProfileName string) error {
	task, err := vdc.RemoveStorageProfile(storageProfileName)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return err
	}
	return vdc.Refresh()
}

// SetDefaultStorageProfile sets a given storage profile as default
// This operation will automatically unset the previous default storage profile.
func (vdc *AdminVdc) SetDefaultStorageProfile(storageProfileName string) error {
	if vdc.client.VCDHREF.String() == "" {
		return fmt.Errorf("cannot set VDC default storage profile: VCD HREF is unset")
	}

	var storageProfile *types.Reference
	for _, sp := range vdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile {
		if sp.Name == storageProfileName {
			storageProfile = sp
		}
	}
	if storageProfile == nil {
		return fmt.Errorf("cannot set VDC default storage profile: storage profile '%s' not found in VDC", storageProfileName)
	}

	vdcStorageProfileDetails, err := vdc.client.GetStorageProfileByHref(storageProfile.HREF)
	if err != nil {
		return fmt.Errorf("cannot retrieve VDC storage profile '%s' details: %s", storageProfileName, err)
	}
	_, err = vdc.UpdateStorageProfile(extractUuid(storageProfile.HREF), &types.AdminVdcStorageProfile{
		Name:    vdcStorageProfileDetails.Name,
		Units:   vdcStorageProfileDetails.Units,
		Limit:   vdcStorageProfileDetails.Limit,
		Default: true,
		Enabled: addrOf(true),
		ProviderVdcStorageProfile: &types.Reference{
			HREF: vdcStorageProfileDetails.ProviderVdcStorageProfile.HREF,
		},
	},
	)
	if err != nil {
		return fmt.Errorf("cannot set VDC default storage profile '%s': %s", storageProfileName, err)
	}
	return vdc.Refresh()
}

// GetDefaultStorageProfileReference finds the default storage profile for the VDC
func (adminVdc *AdminVdc) GetDefaultStorageProfileReference() (*types.Reference, error) {
	var defaultSp *types.Reference
	if adminVdc.AdminVdc.VdcStorageProfiles == nil || adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile == nil {
		return nil, fmt.Errorf("no storage profiles found in VDC %s", adminVdc.AdminVdc.Name)
	}
	for _, sp := range adminVdc.AdminVdc.VdcStorageProfiles.VdcStorageProfile {
		fullSp, err := adminVdc.client.GetStorageProfileByHref(sp.HREF)
		if err != nil {
			return nil, fmt.Errorf("error retrieving storage profile %s for VDC %s: %s", sp.Name, adminVdc.AdminVdc.Name, err)
		}
		if fullSp.Default {
			if defaultSp != nil {
				return nil, fmt.Errorf("more than one default storage profile found for VDC %s: '%s' and '%s'", adminVdc.AdminVdc.Name, sp.Name, defaultSp.Name)
			}
			defaultSp = sp
		}
	}
	if defaultSp != nil {
		return defaultSp, nil
	}
	return nil, fmt.Errorf("no default storage profile found for VDC %s", adminVdc.AdminVdc.Name)
}

// IsNsxv is a convenience function to check if the Admin VDC is backed by NSX-V Provider VDC
func (adminVdc *AdminVdc) IsNsxv() bool {
	vdc := NewVdc(adminVdc.client)
	vdc.Vdc = &adminVdc.AdminVdc.Vdc
	vdc.parent = adminVdc.parent
	return vdc.IsNsxv()
}
