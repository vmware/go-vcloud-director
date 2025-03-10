package ccitypes

const (
	SupervisorNamespaceKind    = "SupervisorNamespace"
	SupervisorNamespaceAPI     = "infrastructure.cci.vmware.com"
	SupervisorNamespaceVersion = "v1alpha2"
	SupervisorNamespacesURL    = "/apis/" + SupervisorNamespaceAPI + "/" + SupervisorNamespaceVersion + "/namespaces/%s/supervisornamespaces"

	ProjectAPI     = "project.cci.vmware.com"
	ProjectKind    = "Project"
	ProjectVersion = "v1alpha2"
	ProjectsURL    = "/apis/" + ProjectAPI + "/" + ProjectVersion + "/projects"

	KubernetesSubpath = "%s://%s/cci/kubernetes"
)
