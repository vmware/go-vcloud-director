//go:build nsxt || alb || functional || ALL
// +build nsxt alb functional ALL

package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_AlbPool(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	skipNoNsxtAlbConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAlbEdgeGateway)

	// Setup prerequisites
	controller, cloud, seGroup, edge, assignment := setupAlbPoolPrerequisites(check, vcd)

	// Run various tests
	testMinimalPoolConfig(check, edge, vcd)
	testAdvancedPoolConfig(check, edge, vcd)
	testPoolWithCertNoPrivateKey(check, vcd, edge.EdgeGateway.ID)
	testPoolWithCertAndPrivateKey(check, vcd, edge.EdgeGateway.ID)

	// teardown prerequisites
	tearDownAlbPoolPrerequisites(check, assignment, edge, seGroup, cloud, controller)
}

func testMinimalPoolConfig(check *C, edge *NsxtEdgeGateway, vcd *TestVCD) {
	poolConfigMinimal := &types.NsxtAlbPool{
		Name:       check.TestName() + "Minimal",
		GatewayRef: types.OpenApiReference{ID: edge.EdgeGateway.ID},
	}

	poolConfigMinimalUpdated := &types.NsxtAlbPool{
		Name:       poolConfigMinimal.Name + "-updated",
		GatewayRef: types.OpenApiReference{ID: edge.EdgeGateway.ID},
	}

	testAlbPoolConfig(check, vcd, "Minimal", poolConfigMinimal, poolConfigMinimalUpdated)
}

func testAdvancedPoolConfig(check *C, edge *NsxtEdgeGateway, vcd *TestVCD) {
	poolConfigAdvanced := &types.NsxtAlbPool{
		Name:                     check.TestName() + "-Advanced",
		GatewayRef:               types.OpenApiReference{ID: edge.EdgeGateway.ID},
		Algorithm:                "FEWEST_SERVERS",
		DefaultPort:              takeIntAddress(8443),
		GracefulTimeoutPeriod:    takeIntAddress(1),
		PassiveMonitoringEnabled: takeBoolPointer(true),
		HealthMonitors:           nil,
		Members: []types.NsxtAlbPoolMember{
			types.NsxtAlbPoolMember{
				Enabled:   true,
				IpAddress: "1.1.1.1",
				Port:      8400,
				Ratio:     takeIntAddress(2),
			},
			types.NsxtAlbPoolMember{
				Enabled:   false,
				IpAddress: "1.1.1.2",
			},
			types.NsxtAlbPoolMember{
				Enabled:   true,
				IpAddress: "1.1.1.3",
			},
		},
		PersistenceProfile: &types.NsxtAlbPoolPersistenceProfile{
			Name:  "PersistenceProfile1",
			Type:  "CLIENT_IP",
			Value: "",
		},
	}

	poolConfigAdvancedUpdated := &types.NsxtAlbPool{
		Name:                     poolConfigAdvanced.Name + "-Updated",
		GatewayRef:               types.OpenApiReference{ID: edge.EdgeGateway.ID},
		Algorithm:                "LEAST_LOAD",
		GracefulTimeoutPeriod:    takeIntAddress(0),
		PassiveMonitoringEnabled: takeBoolPointer(false),
		HealthMonitors:           nil,
		Members: []types.NsxtAlbPoolMember{
			types.NsxtAlbPoolMember{
				Enabled:   true,
				IpAddress: "1.1.1.1",
				Port:      8300,
				Ratio:     takeIntAddress(3),
			},
			types.NsxtAlbPoolMember{
				Enabled:   true,
				IpAddress: "1.1.1.2",
			},
		},
		PersistenceProfile: nil,
	}

	testAlbPoolConfig(check, vcd, "Advanced", poolConfigAdvanced, poolConfigAdvancedUpdated)
}

func testPoolWithCertNoPrivateKey(check *C, vcd *TestVCD, edgeGatewayId string) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	certificateConfigWithoutPrivateKey := &types.CertificateLibraryItem{
		Alias:       check.TestName(),
		Certificate: certificate,
	}
	openApiEndpoint, err := getEndpointByVersion(&vcd.client.Client)
	check.Assert(err, IsNil)
	createdCertificate, err := adminOrg.AddCertificateToLibrary(certificateConfigWithoutPrivateKey)
	check.Assert(err, IsNil)
	PrependToCleanupListOpenApi(createdCertificate.CertificateLibrary.Alias, check.TestName(), openApiEndpoint+createdCertificate.CertificateLibrary.Id)

	poolConfigWithCert := &types.NsxtAlbPool{
		Name:                   check.TestName() + "-complicated",
		GatewayRef:             types.OpenApiReference{ID: edgeGatewayId},
		Algorithm:              "FASTEST_RESPONSE",
		CaCertificateRefs:      []types.OpenApiReference{types.OpenApiReference{ID: createdCertificate.CertificateLibrary.Id}},
		CommonNameCheckEnabled: takeBoolPointer(true),
		DomainNames:            []string{"one", "two", "three"},
		DefaultPort:            takeIntAddress(1211),
	}

	testAlbPoolConfig(check, vcd, "CertificateWithNoPrivateKey", poolConfigWithCert, nil)

	err = createdCertificate.Delete()
	check.Assert(err, IsNil)
}

func testPoolWithCertAndPrivateKey(check *C, vcd *TestVCD, edgeGatewayId string) {
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.org.Org.Name)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)

	certificateConfigWithoutPrivateKey := &types.CertificateLibraryItem{
		Alias:                check.TestName(),
		Certificate:          certificate,
		PrivateKey:           privateKey,
		PrivateKeyPassphrase: "test",
	}

	openApiEndpoint, err := getEndpointByVersion(&vcd.client.Client)
	check.Assert(err, IsNil)
	createdCertificate, err := adminOrg.AddCertificateToLibrary(certificateConfigWithoutPrivateKey)
	check.Assert(err, IsNil)
	PrependToCleanupListOpenApi(createdCertificate.CertificateLibrary.Alias, check.TestName(), openApiEndpoint+createdCertificate.CertificateLibrary.Id)

	poolConfigWithCertAndKey := &types.NsxtAlbPool{
		Name:       check.TestName() + "-complicated",
		GatewayRef: types.OpenApiReference{ID: edgeGatewayId},

		Algorithm:         "FASTEST_RESPONSE",
		CaCertificateRefs: []types.OpenApiReference{types.OpenApiReference{ID: createdCertificate.CertificateLibrary.Id}},
		DefaultPort:       takeIntAddress(1211),
	}

	testAlbPoolConfig(check, vcd, "CertificateWithPrivateKey", poolConfigWithCertAndKey, nil)

	err = createdCertificate.Delete()
	check.Assert(err, IsNil)
}

func testAlbPoolConfig(check *C, vcd *TestVCD, name string, setupConfig *types.NsxtAlbPool, updateConfig *types.NsxtAlbPool) {
	fmt.Printf("# Running ALB Pool test with config %s ", name)

	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	createdPool, err := vcd.client.CreateNsxtAlbPool(setupConfig)
	check.Assert(err, IsNil)

	// Verify mandatory fields
	check.Assert(createdPool.NsxtAlbPool.ID, NotNil)
	check.Assert(createdPool.NsxtAlbPool.Name, NotNil)
	check.Assert(createdPool.NsxtAlbPool.GatewayRef.ID, NotNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbPools + createdPool.NsxtAlbPool.ID
	PrependToCleanupListOpenApi(createdPool.NsxtAlbPool.Name, check.TestName(), openApiEndpoint)

	// Get By ID
	poolById, err := vcd.client.GetAlbPoolById(createdPool.NsxtAlbPool.ID)
	check.Assert(err, IsNil)
	check.Assert(poolById.NsxtAlbPool.ID, Equals, createdPool.NsxtAlbPool.ID)

	// Get By Name
	poolByName, err := vcd.client.GetAlbPoolByName(edge.EdgeGateway.ID, createdPool.NsxtAlbPool.Name)
	check.Assert(err, IsNil)
	check.Assert(poolByName.NsxtAlbPool.ID, Equals, createdPool.NsxtAlbPool.ID)

	// Get All Pool summaries
	allPoolSummaries, err := vcd.client.GetAllAlbPoolSummaries(edge.EdgeGateway.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(len(allPoolSummaries) > 0, Equals, true)

	// Get All Pools
	allPools, err := vcd.client.GetAllAlbPools(edge.EdgeGateway.ID, nil)
	check.Assert(err, IsNil)
	check.Assert(len(allPools) > 0, Equals, true)

	check.Assert(len(allPoolSummaries), Equals, len(allPools))

	// Attempt an update if config is provided
	if updateConfig != nil {
		updateConfig.ID = createdPool.NsxtAlbPool.ID
		updatedPool, err := createdPool.Update(updateConfig)
		check.Assert(err, IsNil)
		check.Assert(createdPool.NsxtAlbPool.ID, Equals, updatedPool.NsxtAlbPool.ID)
		check.Assert(updatedPool.NsxtAlbPool.Name, NotNil)
		check.Assert(updatedPool.NsxtAlbPool.GatewayRef.ID, NotNil)
	}

	err = createdPool.Delete()
	check.Assert(err, IsNil)
	fmt.Printf("Done.\n")
}

func setupAlbPoolPrerequisites(check *C, vcd *TestVCD) (*NsxtAlbController, *NsxtAlbCloud, *NsxtAlbServiceEngineGroup, *NsxtEdgeGateway, *NsxtAlbServiceEngineGroupAssignment) {
	controller, cloud, seGroup := spawnAlbControllerCloudServiceEngineGroup(vcd, check, "SHARED")
	edge, err := vcd.nsxtVdc.GetNsxtEdgeGatewayByName(vcd.config.VCD.Nsxt.EdgeGateway)
	check.Assert(err, IsNil)

	// Enable ALB on Edge Gateway with default ServiceNetworkDefinition
	albSettingsConfig := &types.NsxtAlbConfig{
		Enabled: true,
	}
	enabledSettings, err := edge.UpdateAlbSettings(albSettingsConfig)
	check.Assert(err, IsNil)
	check.Assert(enabledSettings.Enabled, Equals, true)
	PrependToCleanupList(check.TestName()+"-ALB-settings", "OpenApiEntityAlbSettingsDisable", edge.EdgeGateway.Name, check.TestName())

	serviceEngineGroupAssignmentConfig := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: edge.EdgeGateway.ID},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: seGroup.NsxtAlbServiceEngineGroup.ID},
		MaxVirtualServices:    takeIntAddress(89),
		MinVirtualServices:    takeIntAddress(20),
	}

	assignment, err := vcd.client.CreateAlbServiceEngineGroupAssignment(serviceEngineGroupAssignmentConfig)
	check.Assert(err, IsNil)
	check.Assert(assignment.NsxtAlbServiceEngineGroupAssignment.ID, Not(Equals), "")
	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAlbServiceEngineGroupAssignments + assignment.NsxtAlbServiceEngineGroupAssignment.ID
	PrependToCleanupListOpenApi(assignment.NsxtAlbServiceEngineGroupAssignment.ServiceEngineGroupRef.Name, check.TestName(), openApiEndpoint)
	return controller, cloud, seGroup, edge, assignment
}

func tearDownAlbPoolPrerequisites(check *C, assignment *NsxtAlbServiceEngineGroupAssignment, edge *NsxtEdgeGateway, seGroup *NsxtAlbServiceEngineGroup, cloud *NsxtAlbCloud, controller *NsxtAlbController) {
	err := assignment.Delete()
	check.Assert(err, IsNil)
	err = edge.DisableAlb()
	check.Assert(err, IsNil)
	err = seGroup.Delete()
	check.Assert(err, IsNil)
	err = cloud.Delete()
	check.Assert(err, IsNil)
	err = controller.Delete()
	check.Assert(err, IsNil)
}
