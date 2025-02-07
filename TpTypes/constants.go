package tpTypes

const (
	labelSupervisorNamespace   = "Supervisor Namespace"
	SupervisorNamespaceKind    = "SupervisorNamespace"
	SupervisorNamespaceAPI     = "infrastructure.cci.vmware.com"
	SupervisorNamespaceVersion = "v1alpha"
	SupervisorNamespacesURL    = "/apis/infrastructure.cci.vmware.com/v1alpha1/namespaces/%s/supervisornamespaces"

	CciKubernetesSubpath = "%s://%s/cci/kubernetes"
)
