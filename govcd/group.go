/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// OrgGroup defines groups
type OrgGroup struct {
	Group    *types.Group
	client   *Client
	AdminOrg *AdminOrg // needed to be able to update, as the list of roles is found in the Org
}

// NewGroup creates a new group
func NewGroup(cli *Client, org *AdminOrg) *OrgGroup {
	return &OrgGroup{
		Group:    new(types.Group),
		client:   cli,
		AdminOrg: org,
	}
}

func (adminOrg *AdminOrg) CreateGroup(group *types.Group) (*OrgGroup, error) {
	// err := validateUserForCreation(userConfiguration)
	// if err != nil {
	// 	return nil, err
	// }

	userCreateHREF, err := url.ParseRequestURI(adminOrg.AdminOrg.HREF)
	if err != nil {
		return nil, fmt.Errorf("error parsing admin org url: %s", err)
	}
	userCreateHREF.Path += "/groups"

	grpgroup := NewGroup(adminOrg.client, adminOrg)
	group.Xmlns = types.XMLNamespaceVCloud
	group.Type = types.MimeAdminGroup

	_, err = adminOrg.client.ExecuteRequest(userCreateHREF.String(), http.MethodPost,
		types.MimeAdminGroup, "error creating group: %s", group, grpgroup.Group)
	if err != nil {
		return nil, err
	}

	// If there is a valid task, we try to follow through
	// A valid task exists if the Task object in the user structure
	// is not nil and contains at least a task
	// if grpgroup.Group.Tasks != nil && len(grpgroup.Group.Tasks.Task) > 0 {
	// 	task := NewTask(adminOrg.client)
	// 	task.Task = grpgroup.Group.Tasks.Task[0]
	// 	err = task.WaitTaskCompletion()

	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return grpgroup, nil
}

func (adminOrg *AdminOrg) GetGroupByHref(href string) (*OrgGroup, error) {
	orgGroup := NewGroup(adminOrg.client, adminOrg)

	_, err := adminOrg.client.ExecuteRequest(href, http.MethodGet,
		types.MimeAdminUser, "error getting group: %s", nil, orgGroup.Group)

	if err != nil {
		return nil, err
	}
	return orgGroup, nil
}

func (adminOrg *AdminOrg) GetGroupByName(name string, refresh bool) (*OrgGroup, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, group := range adminOrg.AdminOrg.Groups.Group {
		if group.Name == name {
			return adminOrg.GetGroupByHref(group.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

func (adminOrg *AdminOrg) GetGroupById(id string, refresh bool) (*OrgGroup, error) {
	if refresh {
		err := adminOrg.Refresh()
		if err != nil {
			return nil, err
		}
	}

	for _, group := range adminOrg.AdminOrg.Groups.Group {
		if group.ID == id {
			return adminOrg.GetGroupByHref(group.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

func (adminOrg *AdminOrg) GetGroupByNameOrId(identifier string, refresh bool) (*OrgGroup, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetGroupByName(name, refresh) }
	getById := func(name string, refresh bool) (interface{}, error) { return adminOrg.GetGroupById(name, refresh) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*OrgGroup), err
}

func (group *OrgGroup) Delete() error {
	// util.Logger.Printf("[TRACE] Deleting user: %#v (take ownership: %v)", user.User.Name, takeOwnership)

	// if takeOwnership {
	// 	err := user.TakeOwnership()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	groupHREF, err := url.ParseRequestURI(group.Group.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for user %s : %s", group.Group.Name, err)
	}
	util.Logger.Printf("[TRACE] Url for deleting user : %#v and name: %s", groupHREF, group.Group.Name)

	return group.client.ExecuteRequestWithoutResponse(groupHREF.String(), http.MethodDelete,
		types.MimeAdminGroup, "error deleting group : %s", nil)
}

func (group *OrgGroup) Update() error {
	util.Logger.Printf("[TRACE] Updating group: %s", group.Group.Name)

	// Makes sure that GroupReferences is either properly filled or nil,
	// because otherwise vCD will complain that the payload is not well formatted when
	// the configuration contains a non-empty password.
	// if group.Group.GroupReferences != nil {
	// 	if len(user.User.GroupReferences.GroupReference) == 0 {
	// 		user.User.GroupReferences = nil
	// 	}
	// }

	groupHREF, err := url.ParseRequestURI(group.Group.Href)
	if err != nil {
		return fmt.Errorf("error getting HREF for group %s : %s", group.Group.Href, err)
	}
	util.Logger.Printf("[TRACE] Url for updating group : %#v and name: %s", groupHREF, group.Group.Name)

	_, err = group.client.ExecuteRequest(groupHREF.String(), http.MethodPut,
		types.MimeAdminGroup, "error updating group : %s", group.Group, nil)
	return err
}
