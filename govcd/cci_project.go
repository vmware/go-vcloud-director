package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
)

const cciLabelProject = "Project"

// Project manages a VCFA project
type Project struct {
	CciClient *CciClient
	Project   *ccitypes.Project
}

// CreateProject instantiates new project with a given configuration
func (cciClient *CciClient) CreateProject(projectCfg *ccitypes.Project) (*Project, error) {
	urlRef, err := cciClient.GetCciUrl(ccitypes.SupervisorProjectsURL)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating supervisor namespace")
	}

	returnObject := &Project{
		CciClient: cciClient,
		Project:   &ccitypes.Project{},
	}

	if err := cciClient.PostItemSync(urlRef, nil, &projectCfg, returnObject.Project); err != nil {
		return nil, fmt.Errorf("error creating %s in Project %s: %s", cciLabelProject, projectCfg.GetName(), err)
	}

	return returnObject, nil
}

// GetProjectByName retrieves a project by name
func (cciClient *CciClient) GetProjectByName(name string) (*Project, error) {
	addr, err := cciClient.GetCciUrl(ccitypes.SupervisorProjectsURL, "/", name)
	if err != nil {
		return nil, fmt.Errorf("error getting URL for creating %s: %s", cciLabelProject, err)
	}

	returnObject := &Project{
		CciClient: cciClient,
		Project:   &ccitypes.Project{},
	}

	if err := cciClient.GetItem(addr, nil, returnObject.Project, nil); err != nil {
		return nil, fmt.Errorf("error reading %s %s : %s", cciLabelProject, name, err)
	}
	return returnObject, nil

}

// Delete project
func (p *Project) Delete() error {
	addr, err := p.CciClient.GetCciUrl(ccitypes.SupervisorProjectsURL, "/", p.Project.GetName())
	if err != nil {
		return fmt.Errorf("error getting URL for deleting %s: %s", cciLabelProject, err)
	}

	if err := p.CciClient.DeleteItem(addr, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s : %s", cciLabelProject, p.Project.GetName(), err)
	}

	return nil
}
