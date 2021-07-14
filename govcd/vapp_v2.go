package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type VAppV2 struct {
	VAppV2 *types.VAppV2
	client *Client
}

func NewVAppV2(cli *Client) *VAppV2 {
	return &VAppV2{
		VAppV2: new(types.VAppV2),
		client: cli,
	}
}

// ComposeVAppV2 creates a vApp with VMs
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
		types.MimeComposeVappParams, "error instantiating a new vApp:: %s", vAppDef, &vAppContents)
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

func vappV2ToVapp(vappV2 *VAppV2) *VApp {
	var convertedVapp types.VApp = types.VApp(*vappV2.VAppV2)
	vapp := NewVApp(vappV2.client)
	vapp.VApp = &convertedVapp
	return vapp
}

func (vappV2 *VAppV2) Refresh() error {
	return vappV2ToVapp(vappV2).Refresh()
}

func (vappV2 *VAppV2) Undeploy() (Task, error) {
	return vappV2ToVapp(vappV2).Undeploy()
}

func (vappV2 *VAppV2) Delete() (Task, error) {
	return vappV2ToVapp(vappV2).Delete()
}

func (vappV2 *VAppV2) RemoveAllNetworks() (Task, error) {
	return vappV2ToVapp(vappV2).RemoveAllNetworks()
}

func (vappV2 *VAppV2) PowerOn() (Task, error) {
	return vappV2ToVapp(vappV2).PowerOn()
}
