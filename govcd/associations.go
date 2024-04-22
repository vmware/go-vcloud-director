package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"net/url"
)

/*
This file contains methods to retrieve, set, and delete associations between VCD entities.

The associations come in two flavors:

1. Site association: will let one VCD use the other site entities as its own
2. Org associations:
   2a. Associates an organization with another on the same VCD
   2b. Associates an organization with another in a different VCD (requires a site association)

*/

// TODO:
// QueryAllSiteAssociations // gets list of all Site associations
// QueryAllOrgAssociations  // gets list of all Org associatons
// GetSiteAssociationById   // gets info on existing site association
// GetOrgAssociationById    // gets info on existing org association
// DownloadSiteData         // needed to create Site association
// DownloadOrgData          // needed to create Org association
// SetSiteAssociation       // creates new site association
// SetOrgAssociation        // creates new Org association

func (client Client) QueryAllSiteAssociations(params, notEncodedParams map[string]string) ([]*types.QueryResultSiteAssociationRecord, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("system administrator privileges are needed to handle site associations")
	}

	result, err := client.cumulativeQuery(types.QtSiteAssociation, params, notEncodedParams)
	if err != nil {
		return nil, err
	}

	return result.Results.SiteAssociationRecord, nil
}

func (client Client) QueryAllOrgAssociations(params, notEncodedParams map[string]string) ([]*types.QueryResultOrgAssociationRecord, error) {
	if !client.IsSysAdmin {
		return nil, fmt.Errorf("system administrator privileges are needed to handle Org associations")
	}

	result, err := client.cumulativeQuery(types.QtOrgAssociation, params, notEncodedParams)
	if err != nil {
		return nil, err
	}

	return result.Results.OrgAssociationRecord, nil
}

func (org *AdminOrg) GetOrgAssociations() ([]*types.OrgAssociationMember, error) {
	href := getUrlFromLink(org.AdminOrg.Link, "down", types.MimeOrgAssociation)
	if href == "" {
		return nil, fmt.Errorf("no HREF found to get Org associations for Org '%s'", org.AdminOrg.Name)
	}

	var associations types.OrgAssociations
	_, err := org.client.ExecuteRequest(href, http.MethodGet, types.MimeOrgAssociation,
		"error retrieving associations: %s", nil, &associations)
	if err != nil {
		return nil, err
	}

	return associations.OrgAssociationMember, nil
}

func (org *AdminOrg) GetOrgAssociationById(id string) (*types.OrgAssociationMember, error) {
	href := getUrlFromLink(org.AdminOrg.Link, "down", types.MimeOrgAssociation)
	if href == "" {
		return nil, fmt.Errorf("no HREF found to get Org associations for Org '%s'", org.AdminOrg.Name)
	}

	var err error
	href, err = url.JoinPath(href, id)
	if err != nil {
		return nil, fmt.Errorf("error joining URL path with ID: %s", err)
	}
	var association types.OrgAssociationMember
	_, err = org.client.ExecuteRequest(href, http.MethodGet, types.MimeOrgAssociation,
		"error retrieving association: %s", nil, &association)
	if err != nil {
		return nil, err
	}

	return &association, nil
}
