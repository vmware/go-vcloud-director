/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Definition of an OrgUser
type OrgUser struct {
	User     *types.User
	client   *Client
	adminOrg *AdminOrg // needed to be able to update, as the list of roles is found in the Org
}

// Simplified structure to insert or modify an organization user
type OrgUserConfiguration struct {
	Name            string // Mandatory
	Password        string // Mandatory
	RoleName        string // Mandatory
	ProviderType    string // Optional: defaults to "INTEGRATED"
	IsEnabled       bool   // Optional: defaults to false
	IsLocked        bool   // Only used for updates
	DeployedVmQuota int    // Optional: 0 means "unlimited"
	StoredVmQuota   int    // Optional: 0 means "unlimited"
	FullName        string
	Description     string
	EmailAddress    string
	Telephone       string
	TimeOut         time.Duration // Optional: default to 10 seconds
}

const (
	// Common role names and provider types are kept here to reduce hard-coded text and prevent mistakes
	// Roles that are added to the organization need to be entered as free text

	OrgUserRoleOrganizationAdministrator = "Organization Administrator"
	OrgUserRoleCatalogAuthor             = "Catalog Author"
	OrgUserRoleVappAuthor                = "vApp Author"
	OrgUserRoleVappUser                  = "vApp User"
	OrgUserRoleConsoleAccessOnly         = "Console Access Only"
	OrgUserRoleDeferToIdentityProvider   = "Defer to Identity Provider"

	// Allowed values for provider types
	OrgUserProviderIntegrated = "INTEGRATED" // The user is created locally or imported from LDAP
	OrgUserProviderSAML       = "SAML"       // The user is imported from a SAML identity provider.
	OrgUserProviderOAUTH      = "OAUTH"      // The user is imported from an OAUTH identity provider
)

// Used to check the validity of provider type on creation
var OrgUserProviderTypes = []string{
	OrgUserProviderIntegrated,
	OrgUserProviderSAML,
	OrgUserProviderOAUTH,
}

// NewUser creates an empty user
func NewUser(cli *Client, org *AdminOrg) *OrgUser {
	return &OrgUser{
		User:     new(types.User),
		client:   cli,
		adminOrg: org,
	}
}

// GetUserByNameOrId retrieves an user within an admin organization
// by either name or ID
// Returns a valid user if it exists. If it doesn't, returns nil and ErrorEntityNotFound
func (adminOrg *AdminOrg) GetUserByNameOrId(identifier string, willRefresh bool) (*OrgUser, error) {
	if willRefresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, orgUser := range adminOrg.AdminOrg.Users.User {
		if (orgUser.Name == identifier || orgUser.ID == identifier) && orgUser.Type == types.MimeAdminUser {

			user := NewUser(adminOrg.client, adminOrg)

			_, err := adminOrg.client.ExecuteRequest(orgUser.HREF, http.MethodGet,
				types.MimeAdminUser, "error getting user: %s", nil, user.User)

			return user, err
		}
	}
	return nil, ErrorEntityNotFound
}

// GetRole finds a role within the organization
func (adminOrg *AdminOrg) GetRole(roleName string) (*types.Reference, error) {

	// There is no need to refresh the AdminOrg, until we implement CRUD for roles
	for _, role := range adminOrg.AdminOrg.RoleReferences.RoleReference {
		if role.Name == roleName {
			return role, nil
		}
	}

	return &types.Reference{}, ErrorEntityNotFound
}

// CreateUser creates an OrgUser from a full configuration structure
// The timeOut variable is the maximum time we wait for the user to be ready
// (This operation does not return a task)
// Note that the timeout is not absolute, meaning that we only wait the
// full amount of the timeout if the user has not been created until then.
// This function returns as soon as the user has been created, which could be as
// little as 50ms or as much as TimeOut - 50 ms
// Mandatory fields are: Name, Role, Password.
// https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/POST-CreateUser.html
func (adminOrg *AdminOrg) CreateUser(userConfiguration *types.User, timeOut time.Duration) (*OrgUser, error) {
	err := validateUserForCreation(userConfiguration)
	if err != nil {
		return nil, err
	}

	userCreateHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parsing admin org url: %s", err)
	}
	userCreateHREF.Path += "/users"

	user := NewUser(adminOrg.client, adminOrg)

	_, err = adminOrg.client.ExecuteRequest(userCreateHREF.String(), http.MethodPost,
		types.MimeAdminUser, "error creating user: %s", userConfiguration, user.User)
	if err != nil {
		return nil, err
	}

	// Uncomment this line to show the user structure after creation
	// ShowUser(*user.OrgUser)

	// If there is a valid task, we try to follow through
	if user.User.Tasks != nil && len(user.User.Tasks.Task) > 0 {
		task := NewTask(adminOrg.client)
		task.Task = user.User.Tasks.Task[0]
		err = task.WaitTaskCompletion()

		if err != nil {
			return nil, err
		}
	}

	// Attempting to retrieve the user
	// Since there is no task to wait for, we try to retrieve using a max of 200 attempts
	// at intervals of 50 Ms
	// Experience shows that it usually takes less than one second to get the result
	var delayPerAttempt time.Duration = 50
	maxAttempts := 200
	if timeOut > 0 && timeOut > delayPerAttempt {
		maxAttempts = int(timeOut / delayPerAttempt)
	}

	// We make sure that the timeout is never less than 1 second
	if maxAttempts < 20 {
		maxAttempts = 20
	}

	startTime := time.Now()
	var newUser *OrgUser
	for N := 0; N <= int(maxAttempts); N++ {
		newUser, err = adminOrg.GetUserByNameOrId(userConfiguration.Name, true)
		if err == nil {
			break
		}
		time.Sleep(delayPerAttempt)
	}
	endTime := time.Now()

	elapsed := endTime.Sub(startTime)

	// Uncomment this line to show how long the creation takes
	// fmt.Printf("# %s: %s\n",user.OrgUser.Name,elapsed)

	// If the user was not retrieved within the allocated time, we inform the user about the failure
	// and the time it occurred to get to this point, so that they may try with a longer time
	if err != nil {
		return nil, fmt.Errorf("failure to retrieve a new user after %s : %s", elapsed, err)
	}

	return newUser, nil
}

// SimpleCreateUser creates an org user from a simplified structure
func (adminOrg *AdminOrg) SimpleCreateUser(userData OrgUserConfiguration) (*OrgUser, error) {

	if userData.Name == "" {
		return nil, fmt.Errorf("name is mandatory to create a user")
	}
	if userData.Password == "" {
		return nil, fmt.Errorf("password is mandatory to create a user")
	}
	if userData.RoleName == "" {
		return nil, fmt.Errorf("role is mandatory to create a user")
	}
	role, err := adminOrg.GetRole(userData.RoleName)
	if err != nil {
		return nil, fmt.Errorf("error finding a role named %s", userData.RoleName)
	}

	var userConfiguration = types.User{
		Xmlns:           types.XMLNamespaceVCloud,
		Type:            types.MimeAdminUser,
		ProviderType:    userData.ProviderType,
		Name:            userData.Name,
		IsEnabled:       userData.IsEnabled,
		Password:        userData.Password,
		DeployedVmQuota: userData.DeployedVmQuota,
		StoredVmQuota:   userData.StoredVmQuota,
		FullName:        userData.FullName,
		EmailAddress:    userData.EmailAddress,
		Description:     userData.Description,
		Role:            &types.Reference{HREF: role.HREF},
	}

	if userData.TimeOut == 0 {
		userData.TimeOut = 10 * time.Second
	}
	// ShowUser(userConfiguration)
	return adminOrg.CreateUser(&userConfiguration, userData.TimeOut)
}

// GetRoleName retrieves the name of the role currently assigned to the user
func (user *OrgUser) GetRoleName() string {
	if user.User.Role == nil {
		return ""
	}
	return user.User.Role.Name
}

// delete removes the user, returning an error if the call fails.
// if requested, it will attempt to take ownership before the removal.
// API Documentation: https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/DELETE-User.html
// Note: in the GUI we need to disable the user before deleting.
// There is no such constraint with the API.
func (user *OrgUser) delete(takeOwnership bool) error {
	util.Logger.Printf("[TRACE] Deleting user: %#v", user.User.Name)

	if takeOwnership {
		err := user.TakeOwnership()
		if err != nil {
			return err
		}
	}

	userHREF, err := url.ParseRequestURI(user.User.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for user %s : %v", user.User.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for deleting user : %#v and name: %s", userHREF, user.User.Name)

	return user.client.ExecuteRequestWithoutResponse(userHREF.String(), http.MethodDelete,
		types.MimeAdminUser, "error deleting user : %s", nil)
}

// UnconditionalDelete deletes the user, WITHOUT running a call to take ownership of the user objects
// This call will fail if the user has owns *running* vApps/VMs
func (user *OrgUser) UnconditionalDelete() error {
	return user.delete(false)
}

// SafeDelete deletes the user after taking ownership of its objects
func (user *OrgUser) SafeDelete() error {
	return user.delete(true)
}

// SimpleUpdate updates the user, using ALL the fields in userData structure
// returning an error if the call fails.
// Careful: DeployedVmQuota and StoredVmQuota use a `0` value to mean "unlimited"
func (user *OrgUser) SimpleUpdate(userData OrgUserConfiguration) error {
	util.Logger.Printf("[TRACE] Updating user: %#v", user.User.Name)

	if userData.Name != "" {
		user.User.Name = userData.Name
	}
	if userData.ProviderType != "" {
		user.User.ProviderType = userData.ProviderType
	}
	if userData.Description != "" {
		user.User.Description = userData.Description
	}
	if userData.FullName != "" {
		user.User.FullName = userData.FullName
	}
	if userData.EmailAddress != "" {
		user.User.EmailAddress = userData.EmailAddress
	}
	if userData.Telephone != "" {
		user.User.Telephone = userData.Telephone
	}
	if userData.Password != "" {
		user.User.Password = userData.Password
	}
	user.User.StoredVmQuota = userData.StoredVmQuota
	user.User.DeployedVmQuota = userData.DeployedVmQuota
	user.User.IsEnabled = userData.IsEnabled
	user.User.IsLocked = userData.IsLocked

	if userData.RoleName != "" && user.User.Role != nil && user.User.Role.Name != userData.RoleName {
		newRole, err := user.adminOrg.GetRole(userData.RoleName)
		if err != nil {
			return err
		}
		user.User.Role = newRole
	}
	return user.Update()
}

// Update updates the user, using its own configuration data
// returning an error if the call fails.
// API Documentation: https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/PUT-User.html
func (user *OrgUser) Update() error {
	util.Logger.Printf("[TRACE] Updating user: %s", user.User.Name)

	// Makes sure that GroupReferences is either properly filled or nil,
	// because otherwise vCD will complain that the payload is not well formatted when
	// the configuration contains a non-empty password.
	if user.User.GroupReferences != nil {
		if len(user.User.GroupReferences.GroupReference) == 0 {
			user.User.GroupReferences = nil
		}
	}

	userHREF, err := url.ParseRequestURI(user.User.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for user %s : %v", user.User.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for updating user : %#v and name: %s", userHREF, user.User.Name)

	_, err = user.client.ExecuteRequest(userHREF.String(), http.MethodPut,
		types.MimeAdminUser, "error updating user : %s", user.User, nil)
	return err
}

// Disable disables a user, if it is enabled. Fails otherwise.
func (user *OrgUser) Disable() error {
	util.Logger.Printf("[TRACE] Disabling user: %s", user.User.Name)

	if !user.User.IsEnabled {
		return fmt.Errorf("user %s is already disabled", user.User.Name)
	}
	user.User.IsEnabled = false

	return user.Update()
}

// ChangePassword changes user's password
// Constraints: the password must be non-empty, with a minimum of 6 characters
func (user *OrgUser) ChangePassword(newPass string) error {
	util.Logger.Printf("[TRACE] Changing user's password user: %s", user.User.Name)

	user.User.Password = newPass

	return user.Update()
}

// Enable enables a user if it was disabled. Fails otherwise.
func (user *OrgUser) Enable() error {
	util.Logger.Printf("[TRACE] Enabling user: %s", user.User.Name)

	if user.User.IsEnabled {
		return fmt.Errorf("user %s is already enabled", user.User.Name)
	}
	user.User.IsEnabled = true

	return user.Update()
}

// Unlock unlocks a user that was locked out by the system.
// Note that there is no procedure to LOCK a user: it is locked by the system when it exceeds the number of
// unauthorized access attempts
func (user *OrgUser) Unlock() error {
	util.Logger.Printf("[TRACE] Unlocking user: %s", user.User.Name)

	if !user.User.IsLocked {
		return fmt.Errorf("user %s is not locked", user.User.Name)
	}
	user.User.IsLocked = false

	return user.Update()
}

// ChangeRole changes a user's role
// Fails is we try to set the same role as the current one.
// Also fails if the provided role name is not found.
func (user *OrgUser) ChangeRole(roleName string) error {
	util.Logger.Printf("[TRACE] Changing user's role: %s", user.User.Name)

	if roleName == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	if user.User.Role != nil && user.User.Role.Name == roleName {
		return fmt.Errorf("new role is the same as current role")
	}

	newRole, err := user.adminOrg.GetRole(roleName)
	if err != nil {
		return err
	}
	user.User.Role = newRole

	return user.Update()
}

// TakeOwnership takes ownership of the user's objects.
// Ownership is transferred to the caller.
// This is a call to make before deleting. Calling user.SafeDelete() will
// run TakeOwnership before the actual user removal.
// API Documentation: https://code.vmware.com/apis/442/vcloud-director#/doc/doc/operations/POST-TakeOwnership.html
func (user *OrgUser) TakeOwnership() error {
	util.Logger.Printf("[TRACE] Taking ownership from user: %s", user.User.Name)

	userHREF, err := url.ParseRequestURI(user.User.Href + "/action/takeOwnership")
	if err != nil {
		return fmt.Errorf("error getting HREF for user %s : %v", user.User.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for taking ownership from user : %#v and name: %s", userHREF, user.User.Name)

	return user.client.ExecuteRequestWithoutResponse(userHREF.String(), http.MethodPost,
		types.MimeAdminUser, "error taking ownership from user : %s", nil)
}

// validateUserForInput makes sure that the minimum data
// needed for creating an org user has been included in the configuration
func validateUserForCreation(user *types.User) error {
	var missingField = "missing field %s"
	if user.Xmlns == "" {
		return fmt.Errorf(missingField, "Xmlns")
	}
	if user.Name == "" {
		return fmt.Errorf(missingField, "Name")
	}
	if user.Password == "" {
		return fmt.Errorf(missingField, "Password")
	}
	if user.ProviderType != "" {
		validProviderType := false
		for _, pt := range OrgUserProviderTypes {
			if user.ProviderType == pt {
				validProviderType = true
			}
		}
		if !validProviderType {
			return fmt.Errorf("'%s' is not a valid provider type", user.ProviderType)
		}
	}
	if user.Role.HREF == "" {
		return fmt.Errorf(missingField, "Role.HREF")
	}
	return nil
}
