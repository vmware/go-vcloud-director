package types

import "time"

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

// SolutionAddOnInstance represents the RDE Entity structure for Solution Add-On Instances
type SolutionAddOnInstance struct {
	Name                         string                          `json:"name"`
	Scope                        SolutionAddOnInstanceScope      `json:"scope"`
	Status                       string                          `json:"status"`
	Runtime                      SolutionAddOnInstanceRuntime    `json:"runtime"`
	Elements                     []any                           `json:"elements"`
	Requests                     []SolutionAddOnInstanceRequests `json:"requests"`
	Prototype                    string                          `json:"prototype"`
	Resources                    []any                           `json:"resources"`
	Properties                   map[string]any                  `json:"properties"`
	EncryptionKey                string                          `json:"encryptionKey"`
	AddonInstanceSolutionName    string                          `json:"addonInstanceSolutionName"`
	AddonInstanceSolutionVendor  string                          `json:"addonInstanceSolutionVendor"`
	AddonInstanceSolutionVersion string                          `json:"addonInstanceSolutionVersion"`
}
type SolutionAddOnInstanceScope struct {
	AllTenants     bool     `json:"allTenants"`
	TenantScoped   bool     `json:"tenantScoped"`
	Tenants        []string `json:"tenants"`
	ProviderScoped bool     `json:"providerScoped"`
}
type SolutionAddOnInstanceRuntime struct {
	GoVersion   string `json:"goVersion"`
	SdkVersion  string `json:"sdkVersion"`
	VcdVersion  string `json:"vcdVersion"`
	Environment string `json:"environment"`
}
type SolutionAddOnInstanceRequests struct {
	Error        string    `json:"error"`
	Status       string    `json:"status"`
	Operation    string    `json:"operation"`
	StartedBy    string    `json:"startedBy"`
	StartedOn    time.Time `json:"startedOn"`
	InvocationID string    `json:"invocationId"`
}

// SolutionAddOnInputField represents the schema that is defined for each field that is specified in
// a particular Solution Add-On
type SolutionAddOnInputField struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Values      map[string]string `json:"values,omitempty"`
	Default     any               `json:"default,omitempty"`
	Required    bool              `json:"required,omitempty"`
	Description string            `json:"description"`
	Secure      bool              `json:"secure,omitempty"`
	Validation  string            `json:"validation,omitempty"`
	View        string            `json:"view,omitempty"`
	Delete      bool              `json:"delete,omitempty"`
}

type SolutionAddOnInput struct {
	Inputs []SolutionAddOnInputField `json:"inputs"`
}
