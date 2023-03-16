//go:build nsxt || alb || functional || ALL

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_AlbClouds(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)

	albController := spawnAlbController(vcd, check)
	check.Assert(albController, NotNil)

	importableCloud, err := albController.GetAlbImportableCloudByName(vcd.config.VCD.Nsxt.NsxtAlbImportableCloud)
	check.Assert(err, IsNil)

	albCloudConfig := &types.NsxtAlbCloud{
		Name:        check.TestName(),
		Description: "alb-cloud-description",
		LoadBalancerCloudBacking: types.NsxtAlbCloudBacking{
			BackingId:   importableCloud.NsxtAlbImportableCloud.ID,
			BackingType: types.NsxtAlbCloudBackingTypeNsxtAlb,
			LoadBalancerControllerRef: types.OpenApiReference{
				ID: albController.NsxtAlbController.ID,
			},
		},
		NetworkPoolRef: &types.OpenApiReference{
			ID: importableCloud.NsxtAlbImportableCloud.NetworkPoolRef.ID,
		},
	}

	createdAlbCloud, err := vcd.client.CreateAlbCloud(albCloudConfig)
	check.Assert(err, IsNil)
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud + createdAlbCloud.NsxtAlbCloud.ID
	AddToCleanupListOpenApi(createdAlbCloud.NsxtAlbCloud.Name, check.TestName(), openApiEndpoint)

	// Get all clouds and ensure the needed one is found
	allClouds, err := vcd.client.GetAllAlbClouds(nil)
	check.Assert(err, IsNil)
	var foundCreatedCloud bool
	for cloudIndex := range allClouds {
		if allClouds[cloudIndex].NsxtAlbCloud.ID == createdAlbCloud.NsxtAlbCloud.ID {
			foundCreatedCloud = true
			break
		}
	}
	check.Assert(foundCreatedCloud, Equals, true)

	// Filter lookup by name
	filter := url.Values{}
	filter.Add("filter", "name=="+createdAlbCloud.NsxtAlbCloud.Name)
	allCloudsFiltered, err := vcd.client.GetAllAlbClouds(filter)
	check.Assert(err, IsNil)
	check.Assert(len(allCloudsFiltered), Equals, 1)
	check.Assert(allCloudsFiltered[0].NsxtAlbCloud.ID, Equals, createdAlbCloud.NsxtAlbCloud.ID)

	// Get by Name
	albCloudByName, err := vcd.client.GetAlbCloudByName(createdAlbCloud.NsxtAlbCloud.Name)
	check.Assert(err, IsNil)
	check.Assert(albCloudByName.NsxtAlbCloud.Name, Equals, createdAlbCloud.NsxtAlbCloud.Name)

	// Get by ID
	albCloudById, err := vcd.client.GetAlbCloudById(createdAlbCloud.NsxtAlbCloud.ID)
	check.Assert(err, IsNil)
	check.Assert(albCloudById.NsxtAlbCloud.Name, Equals, createdAlbCloud.NsxtAlbCloud.Name)

	// Cleanup
	err = createdAlbCloud.Delete()
	check.Assert(err, IsNil)

	_, err = vcd.client.GetAlbCloudByName(createdAlbCloud.NsxtAlbCloud.Name)
	check.Assert(ContainsNotFound(err), Equals, true)

	err = albController.Delete()
	check.Assert(err, IsNil)
}

// spawnAlbControllerAndCloud is a helper function to spawn NSX-T ALB Controller and Cloud
// It automatically adds these artefacts to clean up list
func spawnAlbControllerAndCloud(vcd *TestVCD, check *C) (*NsxtAlbController, *NsxtAlbCloud) {
	skipNoNsxtAlbConfiguration(vcd, check)

	albController := spawnAlbController(vcd, check)
	check.Assert(albController, NotNil)

	importableCloud, err := albController.GetAlbImportableCloudByName(vcd.config.VCD.Nsxt.NsxtAlbImportableCloud)
	check.Assert(err, IsNil)

	albCloudConfig := &types.NsxtAlbCloud{
		Name:        check.TestName(),
		Description: "alb-cloud-description",
		LoadBalancerCloudBacking: types.NsxtAlbCloudBacking{
			BackingId: importableCloud.NsxtAlbImportableCloud.ID,
			//BackingType: types.NsxtAlbCloudBackingTypeNsxtAlb,
			LoadBalancerControllerRef: types.OpenApiReference{
				ID: albController.NsxtAlbController.ID,
			},
		},
		NetworkPoolRef: &types.OpenApiReference{
			ID: importableCloud.NsxtAlbImportableCloud.NetworkPoolRef.ID,
		},
	}

	createdAlbCloud, err := vcd.client.CreateAlbCloud(albCloudConfig)
	check.Assert(err, IsNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbCloud + createdAlbCloud.NsxtAlbCloud.ID
	PrependToCleanupListOpenApi(createdAlbCloud.NsxtAlbCloud.Name, check.TestName(), openApiEndpoint)

	return albController, createdAlbCloud
}
