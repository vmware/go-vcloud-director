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
	AllTenants     bool `json:"allTenants"`
	TenantScoped   bool `json:"tenantScoped"`
	ProviderScoped bool `json:"providerScoped"`
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

/*
{
	"name": "ds-CwNjS",
	"scope": {
		"allTenants": false,
		"tenantScoped": false,
		"providerScoped": true
	},
	"status": "PENDING",
	"runtime": {
		"goVersion": "go1.21.5",
		"sdkVersion": "1.1.1.8617379",
		"vcdVersion": "10.5.1.23400185",
		"environment": "Development"
	},
	"elements": [],
	"requests": [
		{
			"error": "",
			"status": "PENDING",
			"operation": "CREATE",
			"startedBy": "administrator",
			"startedOn": "2024-05-15T11:58:32.606850307Z",
			"invocationId": "27901936-126f-4eb0-857d-1eba2d0de21e"
		}
	],
	"prototype": "urn:vcloud:entity:vmware:solutions_add_on:73d5cb45-c1ba-41ea-ab30-1f74f52b37a9",
	"resources": [],
	"properties": {
		"__vcdext_encrypted__": "secure@R7oGNaGm/OYBgdY7HvxHvBWfIG8Sj4N3WIw/lFrtoWWuduSWYqj1JztPj8huLgyFPHW9tfxqOkk9Z2upAT67XNJO/bnNWJyReaZl/JaY3Np0",
		"delete-previous-uiplugin-versions": false
	},
	"encryptionKey": "******",
	"addonInstanceSolutionName": "ds",
	"addonInstanceSolutionVendor": "vmware",
	"addonInstanceSolutionVersion": "1.4.0-23376809"
} */
