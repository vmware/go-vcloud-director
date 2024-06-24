package types

// This file defines structures used to retrieve amd modify associations between VCD entities.
// The entities could be:
//   * site (the whole VCD)
//   * Org (a tenant domain)

/*
	Note: every site or org can have as many associations as they want, but each association has only two members.
  Thus, an organization could be associated with 3 more, but we won't see one association with 4 members;
  we will see instead 3 associations of two members each
*/

// MultiSiteStatus defines a type used for status constants
type MultiSiteStatus string

const (
	StatusActive      MultiSiteStatus = "ACTIVE"
	StatusAsymmetric  MultiSiteStatus = "ASYMMETRIC"
	StatusUnreachable MultiSiteStatus = "UNREACHABLE"
	StatusError       MultiSiteStatus = "ERROR"
)

type SiteAssociations struct {
	Href             string                   `xml:"href,attr,omitempty"`
	Type             string                   `xml:"type,attr,omitempty"`
	Link             LinkList                 `xml:"Link,omitempty"`
	SiteAssociations []*SiteAssociationMember `xml:"SiteAssociationMember"`
}

// SiteAssociationMember describes the structure of one member of a site association
type SiteAssociationMember struct {
	Xmlns                   string `xml:"xmlns,attr"`
	Href                    string `xml:"href,attr,omitempty"`
	Id                      string `xml:"id,attr,omitempty"`
	Type                    string `xml:"type,attr,omitempty"`
	Name                    string `xml:"name,attr"`
	Description             string `xml:"Description,omitempty"`             // Optional Description
	BaseUiEndpoint          string `xml:"BaseUiEndpoint"`                    // The base URI of the UI end-point for the site.
	PublicKey               string `xml:"PublicKey,omitempty"`               // PEM-encoded public key for the remote site.
	RestEndpoint            string `xml:"RestEndpoint"`                      // The URI of the REST API end-point for the site.
	RestEndpointCertificate string `xml:"RestEndpointCertificate,omitempty"` // Optional PEM-encoded certificate to use when connecting to the REST API end-point.
	SiteID                  string `xml:"SiteId"`                            // The URN of the remote site
	SiteName                string `xml:"SiteName"`                          // The name of the remote site
	// Current status of this association. One of:
	// ACTIVE (The association has been established by both members, and communication with the remote party succeeded.)
	// ASYMMETRIC (The association has been established at the local site, but the remote party has not yet reciprocated.)
	// UNREACHABLE (The association has been established by both members, but the remote member is currently unreachable.)
	Status string           `xml:"Status,omitempty"`
	Link   LinkList         `xml:"Link,omitempty"`
	Tasks  *TasksInProgress `xml:"task,omitempty"`
}

// Site is the definition of a VCD seen as an element in a collaborative pair
type Site struct {
	Xmlns                   string           `xml:"xmlns,attr,omitempty"`
	Href                    string           `xml:"href,attr,omitempty"`
	Type                    string           `xml:"type,attr,omitempty"`
	Id                      string           `xml:"id,attr,omitempty"`
	Name                    string           `xml:"name,attr"`
	OperationKey            string           `xml:"operationKey,attr,omitempty"`       // Optional unique identifier to support idempotent semantics for create and delete operations.
	Description             string           `xml:"description,omitempty"`             // Optional description
	RestEndpoint            string           `xml:"RestEndpoint"`                      //  The URI of the REST API end-point for the site.
	BaseUiEndpoint          string           `xml:"BaseUiEndpoint"`                    // The base URI of the UI end-point for the site.
	TenantUiEndpoint        string           `xml:"TenantUiEndpoint"`                  // The base URI of the UI end-point for the site.
	RestEndpointCertificate string           `xml:"RestEndpointCertificate,omitempty"` // Optional PEM-encoded certificate to use when connecting to the REST API end-point.
	MultiSiteUrl            string           `xml:"MultiSiteUrl,omitempty"`            //  The URL that represents the entire multisite setup.
	SiteAssociations        SiteAssociations `xml:"SiteAssociations,omitempty"`        //  List of sites associated with this site.
	Link                    LinkList         `xml:"Link,omitempty"`
	Tasks                   TasksInProgress  `xml:"Tasks,omitempty"`
}

// OrgAssociations is a collection of Org associations
type OrgAssociations struct {
	OrgAssociations []*OrgAssociationMember `xml:"OrgAssociationMember"`
}

// OrgAssociationMember describes the structure of one member of an Org association
type OrgAssociationMember struct {
	Xmlns        string   `xml:"xmlns,attr"`
	Href         string   `xml:"href,attr"`
	Type         string   `xml:"type,attr"`
	Link         LinkList `xml:"Link,omitempty"`
	OrgID        string   `xml:"OrgId"`
	OrgName      string   `xml:"OrgName"`
	OrgPublicKey string   `xml:"OrgPublicKey"`
	SiteID       string   `xml:"SiteId"`
	Status       string   `xml:"Status,omitempty"`
}

// QueryResultSiteAssociationRecord defines a structure to retrieve site associations using a query
type QueryResultSiteAssociationRecord struct {
	AssociatedSiteName string `xml:"associatedSiteName,attr"`
	AssociatedSiteId   string `xml:"associatedSiteId,attr"`
	RestEndpoint       string `xml:"restEndpoint,attr"`
	BaseUiEndpoint     string `xml:"baseUiEndpoint,attr"`
	Href               string `xml:"href,attr"`

	// Current status of this association. One of:
	// ACTIVE (The association has been established by both members, and communication with the remote party succeeded.)
	// ASYMMETRIC (The association has been established at the local site, but the remote party has not yet reciprocated.)
	// UNREACHABLE (The association has been established by both members, but the remote member is currently unreachable.)
	Status string   `xml:"status,attr"`
	Link   LinkList `xml:"Link,omitempty"`
}

// QueryResultOrgAssociationRecord defines a structure to retrieve Org associations using a query
type QueryResultOrgAssociationRecord struct {
	SiteId   string   `xml:"siteId,attr"`
	OrgId    string   `xml:"orgId,attr"`
	SiteName string   `xml:"siteName,attr"`
	OrgName  string   `xml:"orgName,attr"`
	Href     string   `xml:"href,attr"`
	Status   string   `xml:"status,attr"`
	Link     LinkList `xml:"Link,omitempty"`
}
