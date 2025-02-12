package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
)

const cciLabelSupervisorNamespace = "Supervisor Namespace"

type SupervisorNamespace struct {
	CciClient           *CciClient
	SupervisorNamespace *ccitypes.SupervisorNamespace
}

// CreateSupervisorNamespace instantiates new Supervisor Namespace in a given project
func (cciClient *CciClient) CreateSupervisorNamespace(projectName string, supervisorNamespace *ccitypes.SupervisorNamespace) (*SupervisorNamespace, error) {
	if projectName == "" {
		return nil, fmt.Errorf("%s name must be specified", cciLabelProject)
	}

	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	urlRef, err := cciClient.GetCciUrl(urlSuffix)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating %s", cciLabelSupervisorNamespace)
	}

	returnObject := &SupervisorNamespace{
		CciClient:           cciClient,
		SupervisorNamespace: &ccitypes.SupervisorNamespace{},
	}

	resultUrl := func(t interface{}) (*url.URL, error) {
		entity := t.(*ccitypes.SupervisorNamespace)
		return cciClient.GetCciUrl(urlSuffix, "/", entity.GetName())
	}

	if err := cciClient.PostItemAsync(urlRef, resultUrl, nil, &supervisorNamespace, returnObject.SupervisorNamespace); err != nil {
		return nil, fmt.Errorf("error creating %s in Project %s: %s", cciLabelSupervisorNamespace, projectName, err)
	}

	return returnObject, nil
}

// GetSupervisorNamespaceByName retrieves Supervisor Namespace from a given project
func (cciClient *CciClient) GetSupervisorNamespaceByName(projectName, supervisorNamespaceName string) (*SupervisorNamespace, error) {
	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	addr, err := cciClient.GetCciUrl(urlSuffix, "/", supervisorNamespaceName)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	returnObject := &SupervisorNamespace{
		CciClient:           cciClient,
		SupervisorNamespace: &ccitypes.SupervisorNamespace{},
	}

	if err := cciClient.GetItem(addr, nil, returnObject.SupervisorNamespace, nil); err != nil {
		return nil, fmt.Errorf("error reading %s %s in Project %s: %s", cciLabelSupervisorNamespace, supervisorNamespaceName, projectName, err)
	}
	return returnObject, nil

}

// Delete removes Supervisor Namespace
func (sn *SupervisorNamespace) Delete() error {
	projectName := sn.SupervisorNamespace.Namespace
	namespaceName := sn.SupervisorNamespace.GetName()
	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	addr, err := sn.CciClient.GetCciUrl(urlSuffix, "/", namespaceName)
	if err != nil {
		return fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	if err := sn.CciClient.DeleteItem(addr, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s in Project %s: %s", cciLabelSupervisorNamespace, namespaceName, projectName, err)
	}

	return nil
}
