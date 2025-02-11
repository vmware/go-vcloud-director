package ccitypes

const (
	SupervisorNamespaceKind = "SupervisorNamespace"
	ProjectKind             = "Project"
	InfrastructureCciAPI    = "infrastructure.cci.vmware.com"
	ProjectCciAPI           = "project.cci.vmware.com"
	ApiVersion              = "v1alpha1"
	SupervisorNamespacesURL = "/apis/infrastructure.cci.vmware.com/v1alpha1/namespaces/%s/supervisornamespaces"
	SupervisorProjectsURL   = "/apis/project.cci.vmware.com/v1alpha1/projects"

	CciKubernetesSubpath = "%s://%s/cci/kubernetes"
)

// /cci/kubernetes/apis/project.cci.vmware.com/v1alpha1/projects
