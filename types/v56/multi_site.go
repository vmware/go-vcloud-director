package types

// This file defines structures used to retrieve amd modify associations between VCD entities.
// The entities could be:
//   * site (the whole VCD)
//   * Org (a tenant domain)

/*
	Note: every site or org can have as many associations as they want, but each association has only two members.
  Thus, an organization could be associated with 3 more, but we won't see one association with 4 members;
  we will see 3 associations of two members each
*/

// SiteAssociationMember describes the structure of one member of a site association
type SiteAssociationMember struct {
	Xmlns                   string   `xml:"xmlns,attr"`
	Href                    string   `xml:"href,attr"`
	Type                    string   `xml:"type,attr"`
	BaseUiEndpoint          string   `xml:"BaseUiEndpoint"`
	Link                    LinkList `xml:"Link,omitempty"`
	PublicKey               string   `xml:"PublicKey"`
	RestEndpoint            string   `xml:"RestEndpoint"`
	RestEndpointCertificate string   `xml:"RestEndpointCertificate"`
	SiteID                  string   `xml:"SiteId"`
	SiteName                string   `xml:"SiteName"`
}

type OrgAssociations struct {
	OrgAssociationMember []*OrgAssociationMember `xml:"OrgAssociationMember"`
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
}

// QueryResultSiteAssociationRecord defines a structure to retrieve site associations using a query
type QueryResultSiteAssociationRecord struct {
	AssociatedSiteName string   `xml:"associatedSiteName,attr"`
	AssociatedSiteId   string   `xml:"associatedSiteId,attr"`
	RestEndpoint       string   `xml:"restEndpoint,attr"`
	BaseUiEndpoint     string   `xml:"baseUiEndpoint,attr"`
	Href               string   `xml:"href,attr"`
	Status             string   `xml:"status,attr"`
	Link               LinkList `xml:"Link,omitempty"`
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
