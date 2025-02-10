package ccitypes

const (
	SupervisorNamespaceKind    = "SupervisorNamespace"
	SupervisorNamespaceAPI     = "infrastructure.cci.vmware.com"
	SupervisorNamespaceVersion = "v1alpha1"
	SupervisorNamespacesURL    = "/apis/infrastructure.cci.vmware.com/v1alpha1/namespaces/%s/supervisornamespaces"

	CciKubernetesSubpath = "%s://%s/cci/kubernetes"
)
