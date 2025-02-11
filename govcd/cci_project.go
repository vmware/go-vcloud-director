package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
)

const cciLabelProject = "Project"

type Project struct {
	TpClient *CciClient
	Project  *ccitypes.Project
}

func (tpClient *CciClient) CreateProject(project *ccitypes.Project) (*Project, error) {
	urlRef, err := tpClient.GetCciUrl(ccitypes.SupervisorProjectsURL)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	returnObject := &Project{
		TpClient: tpClient,
		Project:  &ccitypes.Project{},
	}

	if err := tpClient.PostItemSync(urlRef, nil, &project, returnObject.Project); err != nil {
		return nil, fmt.Errorf("error creating %s in Project %s: %s", cciLabelProject, project.GetName(), err)
	}

	return returnObject, nil
}

func (tpClient *CciClient) GetProjectByName(name string) (*Project, error) {
	addr, err := tpClient.GetCciUrl(ccitypes.SupervisorProjectsURL, "/", name)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating %s: %s", cciLabelProject, err)
	}

	returnObject := &Project{
		TpClient: tpClient,
		Project:  &ccitypes.Project{},
	}

	if err := tpClient.GetItem(addr, nil, returnObject.Project, nil); err != nil {
		return nil, fmt.Errorf("error reading %s %s : %s", cciLabelProject, name, err)
	}
	return returnObject, nil

}

func (sn *Project) Delete() error {
	addr, err := sn.TpClient.GetCciUrl(ccitypes.SupervisorProjectsURL, "/", sn.Project.GetName())
	if err != nil {
		return fmt.Errorf("error getting URL for deleting %s: %s", cciLabelProject, err)
	}

	if err := sn.TpClient.DeleteItem(addr, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s : %s", cciLabelProject, sn.Project.GetName(), err)
	}

	return nil
}
