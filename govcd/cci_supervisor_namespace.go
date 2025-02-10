package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
)

const labelSupervisorNamespace = "Supervisor Namespace"

type SupervisorNamespace struct {
	TpClient            *CciClient
	SupervisorNamespace *ccitypes.SupervisorNamespace

	ProjectName             string
	SupervisorNamespaceName string
}

func (tpClient *CciClient) CreateSupervisorNamespace(projectName string, supervisorNamespace *ccitypes.SupervisorNamespace) (*SupervisorNamespace, error) {
	if projectName == "" {
		return nil, fmt.Errorf("project name must be specified")
	}

	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	urlRef, err := tpClient.GetCciUrl(urlSuffix)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	// Expected final entity URL is different and creation does not return any Location header to
	// detect it automatically so it must be computed manually
	resultUrlRef := copyUrlRef(urlRef)
	resultUrlRef.Path = fmt.Sprintf("%s/%s", resultUrlRef.Path, supervisorNamespace.GetName())

	returnObject := &SupervisorNamespace{
		TpClient:            tpClient,
		SupervisorNamespace: &ccitypes.SupervisorNamespace{},
	}

	if err := tpClient.PostItem(urlRef, resultUrlRef, nil, &supervisorNamespace, &returnObject.SupervisorNamespace); err != nil {
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

	if err := tpClient.VCDClient.Client.OpenApiGetItem("", addr, nil, returnObject.SupervisorNamespace, nil); err != nil {
		return nil, fmt.Errorf("error reading %s %s in Project %s: %s", labelSupervisorNamespace, supervisorNamespaceName, projectName, err)
	}
	return returnObject, nil

}

// TODO - not supported?
// func (sn *SupervisorNamespace) Update() error {
// 	return nil
// }

func (sn *SupervisorNamespace) Delete() error {
	urlSuffix := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, sn.ProjectName)
	addr, err := sn.TpClient.GetCciUrl(urlSuffix, "/", sn.SupervisorNamespaceName)
	if err != nil {
		return fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	if err := sn.TpClient.DeleteItem(addr, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s in Project %s: %s", labelSupervisorNamespace, sn.SupervisorNamespaceName, sn.ProjectName, err)
	}

	return nil
}
