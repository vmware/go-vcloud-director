package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/tpTypes"
)

const labelSupervisorNamespace = "Supervisor Namespace"

type SupervisorNamespace struct {
	TpClient            *TpClient
	SupervisorNamespace *tpTypes.SupervisorNamespace

	ProjectName             string
	SupervisorNamespaceName string
}

func (tpClient *TpClient) CreateSupervisorNamespace(projectName string, supervisorNamespace tpTypes.SupervisorNamespace) (*SupervisorNamespace, error) {
	if projectName == "" {
		return nil, fmt.Errorf("project name must be specified")
	}

	urlSuffix := fmt.Sprintf(tpTypes.SupervisorNamespacesURL, projectName)
	urlRef, err := tpClient.GetServerUrl(urlSuffix)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	// Expected final entity URL is different and creation does not return any Location header
	// so it must be computed manually
	resultUrlRef := copyUrlRef(urlRef)
	resultUrlRef.Path = resultUrlRef.Path + supervisorNamespace.GetName()

	returnObject := &SupervisorNamespace{
		TpClient:            tpClient,
		SupervisorNamespace: &tpTypes.SupervisorNamespace{},
	}

	if err := tpClient.PostItem(urlRef, resultUrlRef, nil, &supervisorNamespace, &returnObject.SupervisorNamespace); err != nil {
		return nil, fmt.Errorf("error creating %s in Project %s: %s", labelSupervisorNamespace, projectName, err)
	}

	// TODO  Need to wait until it is finalized

	return nil, nil
}

func (tpClient *TpClient) GetSupervisorNamespaceByName(projectName, supervisorNamespaceName string) (*SupervisorNamespace, error) {
	urlSuffix := fmt.Sprintf(tpTypes.SupervisorNamespacesURL, projectName)
	addr, err := tpClient.GetServerUrl(urlSuffix, "/", supervisorNamespaceName)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	returnObject := &SupervisorNamespace{
		TpClient:            tpClient,
		SupervisorNamespace: &tpTypes.SupervisorNamespace{},
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
	urlSuffix := fmt.Sprintf(tpTypes.SupervisorNamespacesURL, sn.ProjectName)
	addr, err := sn.TpClient.GetServerUrl(urlSuffix, "/", sn.SupervisorNamespaceName)
	if err != nil {
		return fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	if err := sn.TpClient.DeleteItem(addr, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s in Project %s: %s", labelSupervisorNamespace, sn.SupervisorNamespaceName, sn.ProjectName, err)
	}

	return nil
}

// stateChangeFunc := retry.StateChangeConf{
// 	Pending: []string{"CREATING", "WAITING"},
// 	Target:  []string{"CREATED"},
// 	Refresh: func() (any, string, error) {
// 		supervisorNamespace, err := readSupervisorNamespace(tmClient, projectName.(string), supervisorNamespaceOut.GetName())
// 		if err != nil {
// 			return nil, "", err
// 		}

// 		log.Printf("[DEBUG] %s %s current phase is %s", labelSupervisorNamespace, supervisorNamespaceOut.GetName(), supervisorNamespace.Status.Phase)
// 		if strings.ToUpper(supervisorNamespace.Status.Phase) == "ERROR" {
// 			return nil, "", fmt.Errorf("%s %s is in an ERROR state", labelSupervisorNamespace, supervisorNamespaceOut.GetName())
// 		}

// 		return supervisorNamespace, strings.ToUpper(supervisorNamespace.Status.Phase), nil
// 	},
// 	Timeout:    d.Timeout(schema.TimeoutDelete),
// 	Delay:      5 * time.Second,
// 	MinTimeout: 5 * time.Second,
// }
// if _, err = stateChangeFunc.WaitForStateContext(ctx); err != nil {
// 	return diag.Errorf("error waiting for %s %s in Project %s to be created: %s", labelSupervisorNamespace, supervisorNamespaceOut.GetName(), projectName, err)
// }
