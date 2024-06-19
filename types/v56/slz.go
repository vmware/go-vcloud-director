package types

// SolutionLandingZoneType defines the configuration of Solution Landing Zone.
// It uses RDE so this body must be inserted into `types.DefinedEntity.State` field
type SolutionLandingZoneType struct {
	// ID is the Org ID that the Solution Landing Zone is configured for
	ID string `json:"id"`
	// Name is the Org name that the Solution Landing Zone is configured for
	Name string `json:"name,omitempty"`
	// Catalogs
	Catalogs []SolutionLandingZoneCatalog `json:"catalogs"`
	Vdcs     []SolutionLandingZoneVdc     `json:"vdcs"`
}

type SolutionLandingZoneCatalog struct {
	ID           string   `json:"id"`
	Name         string   `json:"name,omitempty"`
	Capabilities []string `json:"capabilities"`
}

type SolutionLandingZoneVdc struct {
	ID              string                        `json:"id"`
	Name            string                        `json:"name,omitempty"`
	Capabilities    []string                      `json:"capabilities"`
	IsDefault       bool                          `json:"isDefault"`
	Networks        []SolutionLandingZoneVdcChild `json:"networks"`
	StoragePolicies []SolutionLandingZoneVdcChild `json:"storagePolicies"`
	ComputePolicies []SolutionLandingZoneVdcChild `json:"computePolicies"`
}

type SolutionLandingZoneVdcChild struct {
	ID           string   `json:"id"`
	Name         string   `json:"name,omitempty"`
	IsDefault    bool     `json:"isDefault"`
	Capabilities []string `json:"capabilities"`
}

// SolutionAddOn defines structure of Solution Add-On that is deployed in the Solution Landing Zone
type SolutionAddOn struct {
	Eula     string              `json:"eula"`
	Icon     string              `json:"icon"`
	Manifest map[string]any      `json:"manifest"`
	Origin   SolutionAddOnOrigin `json:"origin"`
	Status   string              `json:"status"`
}

type SolutionAddOnOrigin struct {
	Type          string `json:"type"`
	AcceptedBy    string `json:"acceptedBy"`
	AcceptedOn    string `json:"acceptedOn"`
	CatalogItemId string `json:"catalogItemId"`
}
