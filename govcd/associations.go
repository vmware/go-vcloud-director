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

// GetSiteAssociationData retrieves the data needed to start an association with another site
func (client Client) GetSiteAssociationData() (*types.SiteAssociationMember, error) {
	href, err := url.JoinPath(client.VCDHREF.String(), "site", "associations", "site/associations/localAssociationData")
	if err != nil {
		return nil, fmt.Errorf("error setting the URL path for localAssociationData: %s", err)
	}
	var associationData types.SiteAssociationMember
	_, err = client.ExecuteRequest(href, http.MethodGet, types.MimeSiteAssociation,
		"error retrieving site associations: %s", nil, &associationData)
	if err != nil {
		return nil, err
	}

	return &associationData, nil
}

// GetSiteAssociations retrieves all current site associations
// If no associations are available, it returns an empty slice with no error
func (client Client) GetSiteAssociations() ([]*types.SiteAssociationMember, error) {

	href, err := url.JoinPath(client.VCDHREF.String(), "site", "associations")
	if err != nil {
		return nil, fmt.Errorf("error setting the URL path for site/associations: %s", err)
	}
	var associations types.SiteAssociations
	_, err = client.ExecuteRequest(href, http.MethodGet, types.MimeSiteAssociation,
		"error retrieving site associations: %s", nil, &associations)
	if err != nil {
		return nil, err
	}

	return associations.SiteAssociations, nil
}

// QueryAllOrgAssociations retrieve all site associations with optional search parameters
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

// GetOrgAssociations retrieves all associations available for the given Org
func (org AdminOrg) GetOrgAssociations() ([]*types.OrgAssociationMember, error) {
	href := getUrlFromLink(org.AdminOrg.Link, "down", types.MimeOrgAssociation)
	if href == "" {
		return nil, fmt.Errorf("no HREF found to get Org associations for Org '%s'", org.AdminOrg.Name)
	}

	var associations types.OrgAssociations
	_, err := org.client.ExecuteRequest(href, http.MethodGet, types.MimeOrgAssociation,
		"error retrieving org associations: %s", nil, &associations)
	if err != nil {
		return nil, err
	}

	return associations.OrgAssociations, nil
}

// GetOrgAssociationById retrieves a single Org association by its ID
func (org AdminOrg) GetOrgAssociationById(id string) (*types.OrgAssociationMember, error) {
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
