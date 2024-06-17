package types

// type DseConfig struct {
// 	// ID is the Org ID that the Solution Landing Zone is configured for
// 	ID string `json:"id"`
// 	// Name is the Org name that the Solution Landing Zone is configured for
// 	Name     string                       `json:"name,omitempty"`
// 	Catalogs []SolutionLandingZoneCatalog `json:"catalogs"`
// 	Vdcs     []SolutionLandingZoneVdc     `json:"vdcs"`
// }

type DataSolution struct {
	Kind       string            `json:"kind"`
	Spec       DseConfigSpec     `json:"spec"`
	Metadata   DseConfigMetadata `json:"metadata"`
	APIVersion string            `json:"apiVersion"`
}

type DseConfigMetadata struct {
	Name string `json:"name"`
}

type DseConfigSpec struct {
	Artifacts    []ArtifactMap `json:"artifacts"`
	Description  string        `json:"description"`
	DockerConfig *DockerConfig `json:"dockerConfig,omitempty"`
	SolutionType string        `json:"solutionType"`
}

type ArtifactMap map[string]interface{}

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
	Auths Auths `json:"auths"`
}

type Auths map[string]Auth

type Auth struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

type DataSolutionOrgConfig struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
	// struct {
	// 	SolutionType string `json:"solutionType"`
	// 	PrivateData  struct {
	// 	} `json:"privateData"`
	// 	Data struct {
	// 		LicenseType string `json:"LicenseType"`
	// 	} `json:"data"`
	// 	PrivateSecureData struct {
	// 	} `json:"privateSecureData"`
	// } `json:"spec"`
}

type DataSolutionInstanceTemplate struct {
	Kind string `json:"kind"`
	Spec struct {
		Data struct {
			CPU              string `json:"cpu"`
			Name             string `json:"name"`
			Memory           string `json:"memory"`
			Storage          string `json:"storage"`
			Namespace        string `json:"namespace"`
			PvcPolicy        string `json:"pvcPolicy"`
			HighAvailability bool   `json:"highAvailability"`
		} `json:"data"`
		BuiltIn    bool   `json:"builtIn"`
		Content    string `json:"content"`
		Version    string `json:"version"`
		Featured   bool   `json:"featured"`
		DataSchema struct {
			Type string `json:"type"`
			Defs struct {
				Quantity struct {
					Type    string `json:"type"`
					Pattern string `json:"pattern"`
				} `json:"quantity"`
				LimitedQuantity struct {
					Type    string `json:"type"`
					Pattern string `json:"pattern"`
				} `json:"limitedQuantity"`
			} `json:"$defs"`
			Schema     string   `json:"$schema"`
			Required   []string `json:"required"`
			Properties struct {
				CPU struct {
					Ref         string `json:"$ref"`
					Title       string `json:"title"`
					Default     string `json:"default"`
					Description string `json:"description"`
				} `json:"cpu"`
				Memory struct {
					Ref         string `json:"$ref"`
					Title       string `json:"title"`
					Default     string `json:"default"`
					Description string `json:"description"`
				} `json:"memory"`
				Storage struct {
					Ref         string `json:"$ref"`
					Title       string `json:"title"`
					Default     string `json:"default"`
					Description string `json:"description"`
				} `json:"storage"`
				Namespace struct {
					Const string `json:"const"`
				} `json:"namespace"`
				PvcPolicy struct {
					Enum        []string `json:"enum"`
					Type        string   `json:"type"`
					Title       string   `json:"title"`
					Default     string   `json:"default"`
					Description string   `json:"description"`
				} `json:"pvcPolicy"`
				HighAvailability struct {
					Type        string `json:"type"`
					Title       string `json:"title"`
					Default     bool   `json:"default"`
					Description string `json:"description"`
				} `json:"highAvailability"`
			} `json:"properties"`
		} `json:"dataSchema"`
		ContentType    string `json:"contentType"`
		Description    string `json:"description"`
		SolutionType   string `json:"solutionType"`
		TemplateEngine string `json:"templateEngine"`
	} `json:"spec"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	APIVersion string `json:"apiVersion"`
}
