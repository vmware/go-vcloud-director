package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
)

const labelSupervisorNamespace = "Supervisor Namespace"

type SupervisorNamespace struct {
	TpClient            *CciClient
	SupervisorNamespace *ccitypes.SupervisorNamespace
}

func (tpClient *CciClient) CreateSupervisorNamespace(projectName string, supervisorNamespace *ccitypes.SupervisorNamespace) (*SupervisorNamespace, error) {
	if projectName == "" {
		return nil, fmt.Errorf("%s name must be specified", labelProject)
	}

	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	urlRef, err := tpClient.GetCciUrl(urlSuffix)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating %s", labelSupervisorNamespace)
	}

	returnObject := &SupervisorNamespace{
		TpClient:            tpClient,
		SupervisorNamespace: &ccitypes.SupervisorNamespace{},
	}

	resultUrl := func(t interface{}) (*url.URL, error) {
		entity := t.(*ccitypes.SupervisorNamespace)
		return tpClient.GetCciUrl(urlSuffix, "/", entity.GetName())
	}

	if err := tpClient.PostItemAsync(urlRef, resultUrl, nil, &supervisorNamespace, returnObject.SupervisorNamespace); err != nil {
		return nil, fmt.Errorf("error creating %s in Project %s: %s", labelSupervisorNamespace, projectName, err)
	}

	return returnObject, nil
}

func (tpClient *CciClient) GetSupervisorNamespaceByName(projectName, supervisorNamespaceName string) (*SupervisorNamespace, error) {
	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	addr, err := tpClient.GetCciUrl(urlSuffix, "/", supervisorNamespaceName)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	returnObject := &SupervisorNamespace{
		TpClient:            tpClient,
		SupervisorNamespace: &ccitypes.SupervisorNamespace{},
	}

	if err := tpClient.GetItem(addr, nil, returnObject.SupervisorNamespace, nil); err != nil {
		return nil, fmt.Errorf("error reading %s %s in Project %s: %s", labelSupervisorNamespace, supervisorNamespaceName, projectName, err)
	}
	return returnObject, nil

}

func (sn *SupervisorNamespace) Delete() error {
	projectName := sn.SupervisorNamespace.Namespace
	namespaceName := sn.SupervisorNamespace.GetName()
	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	addr, err := sn.TpClient.GetCciUrl(urlSuffix, "/", namespaceName)
	if err != nil {
		return fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	if err := sn.TpClient.DeleteItem(addr, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s in Project %s: %s", labelSupervisorNamespace, namespaceName, projectName, err)
	}

	return nil
}
