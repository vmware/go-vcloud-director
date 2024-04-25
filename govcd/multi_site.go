package govcd

import (
	"encoding/xml"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"net/http"
	"net/url"
	"os"
	"path"
)

/*
This file contains methods to retrieve, set, and delete associations between VCD entities.

The associations come in two flavors:

1. Site association: will let one VCD use the other site entities as its own
2. Org associations:
   2a. Associates an organization with another on the same VCD
   2b. Associates an organization with another in a different VCD (requires a site association)

*/

// -----------------------------------------------------------------------------------------------------------------
// Site association read operations
// -----------------------------------------------------------------------------------------------------------------

// QueryAllSiteAssociations retrieves all site associations for the current site
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

// GetSiteAssociationData retrieves the structured data needed to start an association with another site
// This is useful when we have control of both sites from the same client
func (client Client) GetSiteAssociationData() (*types.SiteAssociationMember, error) {
	href, err := url.JoinPath(client.VCDHREF.String(), "site", "associations", "localAssociationData")
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

// GetSiteRawAssociationData retrieves the raw (XML) data needed to start an association with another site
// This is useful when we want to save this data to a file for future use
func (client Client) GetSiteRawAssociationData() ([]byte, error) {
	href, err := url.JoinPath(client.VCDHREF.String(), "site", "associations", "localAssociationData")
	if err != nil {
		return nil, fmt.Errorf("error setting the URL path for site/associations/localAssociationData: %s", err)
	}
	return client.RetrieveRemoteDocument(href)
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

// GetSiteAssociationBySiteId retrieves a single site association by the ID of the associated site
// Note that there could be only one association between two sites
func (client Client) GetSiteAssociationBySiteId(siteId string) (*types.SiteAssociationMember, error) {
	associations, err := client.GetSiteAssociations()
	if err != nil {
		return nil, fmt.Errorf("error retrieving associations for current site: %s", err)
	}

	for _, a := range associations {
		if equalIds(siteId, a.SiteID, "") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("no association found for site ID %s", siteId)
}

// -----------------------------------------------------------------------------------------------------------------
// Site association modifying operations
// -----------------------------------------------------------------------------------------------------------------

// SetSiteAssociationAsync sets a new site association without waiting for completion
func (client Client) SetSiteAssociationAsync(associationData types.SiteAssociationMember) (Task, error) {
	href, err := url.JoinPath(client.VCDHREF.String(), "site", "associations")
	if err != nil {
		return Task{}, fmt.Errorf("error setting the URL path for site/associations: %s", err)
	}
	associationData.Xmlns = types.XMLNamespaceVCloud
	task, err := client.ExecuteTaskRequest(href, http.MethodPost, "application/*+xml",
		"error setting site association: %s", &associationData)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

// SetSiteAssociation sets a new site association, waiting for completion
func (client Client) SetSiteAssociation(associationData types.SiteAssociationMember) error {
	task, err := client.SetSiteAssociationAsync(associationData)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// RemoveSiteAssociationAsync removes a site association without waiting for completion
func (client Client) RemoveSiteAssociationAsync(associationHref string) (Task, error) {
	task, err := client.ExecuteTaskRequest(associationHref, http.MethodDelete, "",
		"error removing site association: %s", nil)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

// RemoveSiteAssociation removes a site association, waiting for completion
func (client Client) RemoveSiteAssociation(associationHref string) error {
	task, err := client.RemoveSiteAssociationAsync(associationHref)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// -----------------------------------------------------------------------------------------------------------------
// Org association read operations
// -----------------------------------------------------------------------------------------------------------------

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
	href, err := org.getAssociationLink(false)
	if err != nil {
		return nil, fmt.Errorf("error retrieving association URL: %s", err)
	}
	var associations types.OrgAssociations
	_, err = org.client.ExecuteRequest(href, http.MethodGet, types.MimeOrgAssociation,
		"error retrieving org associations: %s", nil, &associations)
	if err != nil {
		return nil, err
	}

	return associations.OrgAssociations, nil
}

// GetOrgAssociationByOrgId retrieves a single Org association by the ID of the associated Org
// Note that there could be only one association between two organization
func (org AdminOrg) GetOrgAssociationByOrgId(orgId string) (*types.OrgAssociationMember, error) {
	associations, err := org.GetOrgAssociations()
	if err != nil {
		return nil, fmt.Errorf("error retrieving associations for org '%s': %s", org.AdminOrg.Name, err)
	}

	for _, a := range associations {
		if equalIds(orgId, a.OrgID, "") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("no association found for Org ID %s", orgId)
}

// GetOrgAssociationData retrieves the structured data needed to start an association with another Org
// This is useful when we have control of both Orgs from the same client
func (org AdminOrg) GetOrgAssociationData() (*types.OrgAssociationMember, error) {
	href, err := org.getAssociationLink(true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving association URL: %s", err)
	}
	var associationData types.OrgAssociationMember
	_, err = org.client.ExecuteRequest(href, http.MethodGet, types.MimeOrgAssociation,
		"error retrieving org association data: %s", nil, &associationData)
	if err != nil {
		return nil, err
	}

	return &associationData, nil
}

// GetOrgRawAssociationData retrieves the raw (XML) data needed to start an association with another Org
// This is useful when we want to save this data to a file for future use
func (org AdminOrg) GetOrgRawAssociationData() ([]byte, error) {
	href, err := org.getAssociationLink(true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving association URL: %s", err)
	}
	return org.client.RetrieveRemoteDocument(href)
}

// -----------------------------------------------------------------------------------------------------------------
// Org association modifying operations
// -----------------------------------------------------------------------------------------------------------------

// SetOrgAssociationAsync sets a new Org association without waiting for completion
func (org *AdminOrg) SetOrgAssociationAsync(associationData types.OrgAssociationMember) (Task, error) {
	href, err := org.getAssociationLink(false)
	if err != nil {
		return Task{}, fmt.Errorf("error retrieving association URL: %s", err)
	}
	associationData.Xmlns = types.XMLNamespaceVCloud
	task, err := org.client.ExecuteTaskRequest(href, http.MethodPost, "application/*+xml",
		"error setting org association: %s", &associationData)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

// SetOrgAssociation sets a new Org association, waiting for completion
func (org *AdminOrg) SetOrgAssociation(associationData types.OrgAssociationMember) error {
	task, err := org.SetOrgAssociationAsync(associationData)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// RemoveOrgAssociationAsync removes an Org association without waiting for completion
func (org *AdminOrg) RemoveOrgAssociationAsync(associationHref string) (Task, error) {
	task, err := org.client.ExecuteTaskRequest(associationHref, http.MethodDelete, "",
		"error removing org association: %s", nil)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

// RemoveOrgAssociation removes an Org association, waiting for completion
func (org *AdminOrg) RemoveOrgAssociation(associationHref string) error {
	task, err := org.RemoveOrgAssociationAsync(associationHref)
	if err != nil {
		return err
	}
	return task.WaitTaskCompletion()
}

// -----------------------------------------------------------------------------------------------------------------
// Miscellaneous
// -----------------------------------------------------------------------------------------------------------------

// getAssociationLink retrieves the URL needed to run associations operations with an Org.
// If the 'localData' parameter is true, it returns the URL needed to download the association
// data needed to create a new association
func (org AdminOrg) getAssociationLink(localData bool) (string, error) {
	href := getUrlFromLink(org.AdminOrg.Link, "down", types.MimeOrgAssociation)
	if href == "" {
		return "", fmt.Errorf("no HREF found to get Org association data for Org '%s'", org.AdminOrg.Name)
	}

	if localData {
		var err error
		href, err = url.JoinPath(href, "localAssociationData")
		if err != nil {
			return "", err
		}
	}
	return href, nil
}

// ReadXmlDataFromFile reads the contents of a file and attempts decoding an expected data type
// Examples:
// orgSettingData, err := ReadXmlDataFromFile[types.OrgAssociationMember]("./data/org1-association-data.xml")
// siteSettingData, err := ReadXmlDataFromFile[types.SiteAssociationMember]("./data/site1-association-data.xml")
func ReadXmlDataFromFile[dataType any](fileName string) (*dataType, error) {
	contents, err := os.ReadFile(path.Clean(fileName))
	if err != nil {
		return nil, fmt.Errorf("error reading file '%s': %s", fileName, err)
	}
	var localData dataType
	err = xml.Unmarshal(contents, &localData)
	if err != nil {
		return nil, fmt.Errorf("error decoding data from file '%s': %s", fileName, err)
	}
	return &localData, nil
}
