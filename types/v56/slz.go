package types

// SolutionLandingZoneType defines the configuration of Solution Landing Zone.
// It uses RDE so this body must be inserted into `types.DefinedEntity.State` field
type SolutionLandingZoneType struct {
	Name     string                       `json:"name,omitempty"`
	ID       string                       `json:"id"`
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
