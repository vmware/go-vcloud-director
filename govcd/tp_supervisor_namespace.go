package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/tptypes"
)

const labelSupervisorNamespace = "Supervisor Namespace"

type SupervisorNamespace struct {
	TpClient            *TpClient
	SupervisorNamespace *tptypes.SupervisorNamespace

	ProjectName             string
	SupervisorNamespaceName string
}

func (tpClient *TpClient) CreateSupervisorNamespace(projectName string, supervisorNamespace *tptypes.SupervisorNamespace) (*SupervisorNamespace, error) {
	if projectName == "" {
		return nil, fmt.Errorf("project name must be specified")
	}

	urlSuffix := fmt.Sprintf(tptypes.SupervisorNamespacesURL, projectName)
	urlRef, err := tpClient.GetServerUrl(urlSuffix)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	// Expected final entity URL is different and creation does not return any Location header to
	// detect it automatically so it must be computed manually
	resultUrlRef := copyUrlRef(urlRef)
	resultUrlRef.Path = fmt.Sprintf("%s/%s", resultUrlRef.Path, supervisorNamespace.GetName())

	returnObject := &SupervisorNamespace{
		TpClient:            tpClient,
		SupervisorNamespace: &tptypes.SupervisorNamespace{},
	}

	if err := tpClient.PostItem(urlRef, resultUrlRef, nil, &supervisorNamespace, &returnObject.SupervisorNamespace); err != nil {
		return nil, fmt.Errorf("error creating %s in Project %s: %s", labelSupervisorNamespace, projectName, err)
	}

	return returnObject, nil
}

func (tpClient *TpClient) GetSupervisorNamespaceByName(projectName, supervisorNamespaceName string) (*SupervisorNamespace, error) {
	urlSuffix := fmt.Sprintf(tptypes.SupervisorNamespacesURL, projectName)
	addr, err := tpClient.GetServerUrl(urlSuffix, "/", supervisorNamespaceName)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	returnObject := &SupervisorNamespace{
		TpClient:            tpClient,
		SupervisorNamespace: &tptypes.SupervisorNamespace{},
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
	urlSuffix := fmt.Sprintf(tptypes.SupervisorNamespacesURL, sn.ProjectName)
	addr, err := sn.TpClient.GetServerUrl(urlSuffix, "/", sn.SupervisorNamespaceName)
	if err != nil {
		return fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	if err := sn.TpClient.DeleteItem(addr, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s in Project %s: %s", labelSupervisorNamespace, sn.SupervisorNamespaceName, sn.ProjectName, err)
	}

	return nil
}
