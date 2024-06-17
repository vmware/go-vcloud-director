package types

// type DseConfig struct {
// 	// ID is the Org ID that the Solution Landing Zone is configured for
// 	ID string `json:"id"`
// 	// Name is the Org name that the Solution Landing Zone is configured for
// 	Name     string                       `json:"name,omitempty"`
// 	Catalogs []SolutionLandingZoneCatalog `json:"catalogs"`
// 	Vdcs     []SolutionLandingZoneVdc     `json:"vdcs"`
// }

type DseConfig struct {
	Kind       string            `json:"kind"`
	Spec       DseConfigSpec     `json:"spec"`
	Metadata   DseConfigMetadata `json:"metadata"`
	APIVersion string            `json:"apiVersion"`
}

type DseConfigMetadata struct {
	Name string `json:"name"`
}

type DseConfigSpec struct {
	Artifacts    []Artifact   `json:"artifacts"`
	Description  string       `json:"description"`
	DockerConfig DockerConfig `json:"dockerConfig"`
	SolutionType string       `json:"solutionType"`
}

type Artifact struct {
	Name                        string `json:"name"`
	Type                        string `json:"type"`
	Image                       string `json:"image"`
	Version                     string `json:"version"`
	Manifests                   string `json:"manifests"`
	DefaultImage                string `json:"defaultImage"`
	DefaultVersion              string `json:"defaultVersion"`
	CompatibleVersions          string `json:"compatibleVersions"`
	RequireVersionCompatibility bool   `json:"requireVersionCompatibility"`
}

type DockerConfig struct {
	Auths map[string]Auth `json:"auths"`

	// Auths Auths `json:"auths"`
}

type Auth struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

// type Auths struct {
// 	RegistryTanzuVmwareCOM  RegistryTanzuVmwareCOM  `json:"registry.tanzu.vmware.com"`
// 	RregistryTanzuVmwareCOM RregistryTanzuVmwareCOM `json:"rregistry.tanzu.vmware.com"`
// }

// type RegistryTanzuVmwareCOM struct {
// 	Description string `json:"description"`
// }

// type RregistryTanzuVmwareCOM struct {
// 	Password    string `json:"password"`
// 	Username    string `json:"username"`
// 	Description string `json:"description"`
// }
