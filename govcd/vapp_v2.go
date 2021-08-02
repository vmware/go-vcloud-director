package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// VAppV2 is equivalent to a VApp, but with the additional capability of parallel VM creation
type VAppV2 struct {
	VAppV2 *types.VAppV2
	client *Client
}

// NewVAppV2 creates a VAppV2 pointer
func NewVAppV2(cli *Client) *VAppV2 {
	return &VAppV2{
		VAppV2: new(types.VAppV2),
		client: cli,
	}
}

// ComposeVAppV2 creates a vApp with one or more VMs
// []*SourcedItem is for VMs from template
// []*CreateItem is for VMs from VCD internal definitions
func (vdc *Vdc) ComposeVAppV2(vAppDef *types.ComposeVAppParamsV2) (*VAppV2, error) {

	vAppDef.Ovf = types.XMLNamespaceOVF
	vAppDef.Xsi = types.XMLNamespaceXSI
	vAppDef.Xmlns = types.XMLNamespaceVCloud
	if vAppDef.Name == "" {
		return nil, fmt.Errorf("empty vApp name provided")
	}

	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return nil, fmt.Errorf("error getting vdc href: %s", err)
	}
	vdcHref.Path += "/action/composeVApp"

	var vAppContents types.VAppV2

	_, err = vdc.client.ExecuteRequest(vdcHref.String(), http.MethodPost,
		types.MimeComposeVappParams, "error instantiating a new vApp: %s", vAppDef, &vAppContents)
	if err != nil {
		return nil, fmt.Errorf("error executing task request: %s", err)
	}

	if vAppContents.Tasks != nil {
		for _, innerTask := range vAppContents.Tasks.Task {
			if innerTask != nil {
				task := NewTask(vdc.client)
				task.Task = innerTask
				err = task.WaitTaskCompletion()
				if err != nil {
					return nil, fmt.Errorf("error performing task: %s", err)
				}
			}
		}
	}

	vapp := NewVAppV2(vdc.client)
	vapp.VAppV2 = &vAppContents

	err = vapp.Refresh()
	if err != nil {
		return nil, err
	}

	err = vdc.Refresh()
	if err != nil {
		return nil, err
	}
	return vapp, nil
}

// RecomposeVAppV2 modifies a vApp with one or more VMs
// []*SourcedItem is for VMs from template
// []*CreateItem is for VMs from VCD internal definitions
func (vappV2 *VAppV2) RecomposeVAppV2(vAppDef *types.ReComposeVAppParamsV2) error {

	vAppDef.Ovf = types.XMLNamespaceOVF
	vAppDef.Xsi = types.XMLNamespaceXSI
	vAppDef.Xmlns = types.XMLNamespaceVCloud
	if vAppDef.Name == "" {
		return fmt.Errorf("empty vApp name provided")
	}

	vappHref, err := url.ParseRequestURI(vappV2.VAppV2.HREF)
	if err != nil {
		return fmt.Errorf("[RecomposeVAppV2] error getting vapp href: %s", err)
	}
	vappHref.Path += "/action/recomposeVApp"

	task, err := vappV2.client.ExecuteTaskRequest(vappHref.String(), http.MethodPost,
		types.MimeRecomposeVappParams, "[RecomposeVAppV2] error recomposing vApp: %s", vAppDef)
	if err != nil {
		return fmt.Errorf("[RecomposeVAppV2] error executing task request: %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		return fmt.Errorf("[RecomposeVAppV2] error performing task: %s", err)
	}

	return vappV2.Refresh()
}

// The following functions are convenience shortcuts to manipulate a VAppV2 without
// explicitly converting it to a VApp

// vappV2ToVapp converts a VAppV2 to VApp
func vappV2ToVapp(vappV2 *VAppV2) *VApp {
	var convertedVapp types.VApp = types.VApp(*vappV2.VAppV2)
	vapp := NewVApp(vappV2.client)
	vapp.VApp = &convertedVapp
	return vapp
}

// vappToVappV2 converts a VApp to VAppV2
func vappToVappV2(vapp *VApp) *VAppV2 {
	var convertedVapp types.VAppV2 = types.VAppV2(*vapp.VApp)
	vappV2 := NewVAppV2(vapp.client)
	vappV2.VAppV2 = &convertedVapp
	return vappV2
}

// Refresh fetches up-to-date information about the vApp
func (vappV2 *VAppV2) Refresh() error {
	vapp := vappV2ToVapp(vappV2)
	err := vapp.Refresh()
	if err != nil {
		return err
	}
	innerVapp := types.VAppV2(*vapp.VApp)
	vappV2.VAppV2 = &innerVapp
	return nil
}

// Undeploy stops the vApp
func (vappV2 *VAppV2) Undeploy() (Task, error) {
	return vappV2ToVapp(vappV2).Undeploy()
}

// Delete deletes the vApp
func (vappV2 *VAppV2) Delete() (Task, error) {
	return vappV2ToVapp(vappV2).Delete()
}

// RemoveAllNetworks removes all networks from the vApp
func (vappV2 *VAppV2) RemoveAllNetworks() (Task, error) {
	return vappV2ToVapp(vappV2).RemoveAllNetworks()
}

// PowerOn starts the vApp
func (vappV2 *VAppV2) PowerOn() (Task, error) {
	return vappV2ToVapp(vappV2).PowerOn()
}

// GetVappV2ByName retrieves a vApp by name
func (vdc *Vdc) GetVappV2ByName(name string, refresh bool) (*VAppV2, error) {
	vapp, err := vdc.GetVAppByName(name, refresh)
	if err != nil {
		return nil, err
	}
	return vappToVappV2(vapp), nil
}

// GetVappV2ById retrieves a vApp by ID
func (vdc *Vdc) GetVappV2ById(id string, refresh bool) (*VAppV2, error) {
	vapp, err := vdc.GetVAppById(id, refresh)
	if err != nil {
		return nil, err
	}
	return vappToVappV2(vapp), nil
}

// GetVappV2ByNameOrId retrieves a vApp by Ime or D
func (vdc *Vdc) GetVappV2ByNameOrId(identifier string, refresh bool) (*VAppV2, error) {
	vapp, err := vdc.GetVAppByNameOrId(identifier, refresh)
	if err != nil {
		return nil, err
	}
	return vappToVappV2(vapp), nil
}

// GetVappV2ByHref retrieves a vApp by Href
func (vdc *Vdc) GetVappV2ByHref(href string) (*VAppV2, error) {
	vapp, err := vdc.GetVAppByHref(href)
	if err != nil {
		return nil, err
	}
	return vappToVappV2(vapp), nil
}

func (client *Client) GetVappV2ByHref(href string) (*VAppV2, error) {

	if href == "" {
		return nil, fmt.Errorf("cannot find VAppV2: HREF is empty")
	}

	vapp := NewVAppV2(client)
	_, err := vapp.client.ExecuteRequest(href, http.MethodGet,
		"", "error retrieving vApp: %s", nil, vapp.VAppV2)

	return vapp, err
}
