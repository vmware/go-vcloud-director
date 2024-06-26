package types

// DataSolution represents RDE Entity for Data Solution in Data Solution Extension (DSE)
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
	Artifacts    []DseArtifactMap `json:"artifacts"`
	Description  string           `json:"description"`
	DockerConfig *DseDockerConfig `json:"dockerConfig,omitempty"`
	SolutionType string           `json:"solutionType"`
}

type DseArtifactMap map[string]interface{}

// DseDockerConfig provides registry auth configuration that is available for "VCD Data Solutions"
// Data Solution
type DseDockerConfig struct {
	Auths DseDockerAuths `json:"auths"`
}

type DseDockerAuths map[string]DseDockerAuth

type DseDockerAuth struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

// DataSolutionOrgConfig represents RDE Entity structure for Data Solution Org Configuration
type DataSolutionOrgConfig struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
}

// DataSolutionInstanceTemplate represents a read-only structure for Data Solution Instance
// Templates in Data Storage Extension (DSE)
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
				CPU       DataSolutionInstanceTemplateComputeProperty `json:"cpu"`
				Memory    DataSolutionInstanceTemplateComputeProperty `json:"memory"`
				Storage   DataSolutionInstanceTemplateComputeProperty `json:"storage"`
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

type DataSolutionInstanceTemplateComputeProperty struct {
	Ref         string `json:"$ref"`
	Title       string `json:"title"`
	Default     string `json:"default"`
	Description string `json:"description"`
}
