//go:build network || nsxt || functional || openapi || ALL

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_NsxtApplicationPortProfileProvider(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAppPortProfiles)
	vcd.skipIfNotSysAdmin(check)

	appPortProfileConfig := getAppProfileProvider(vcd, check)
	testAppPortProfile(appPortProfileConfig, types.ApplicationPortProfileScopeProvider, vcd, check)
}

func (vcd *TestVCD) Test_NsxtApplicationPortProfileTenant(check *C) {
	vcd.skipIfNotSysAdmin(check)
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAppPortProfiles)

	appPortProfileConfig := getAppProfileTenant(vcd, check)
	testAppPortProfile(appPortProfileConfig, types.ApplicationPortProfileScopeTenant, vcd, check)
}

func (vcd *TestVCD) Test_NsxtApplicationPortProfileReadSystem(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAppPortProfiles)

	testApplicationProfilesForScope(types.ApplicationPortProfileScopeSystem, check, vcd)
}

func (vcd *TestVCD) Test_NsxtApplicationPortProfileReadProvider(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAppPortProfiles)

	testApplicationProfilesForScope(types.ApplicationPortProfileScopeProvider, check, vcd)
}

func (vcd *TestVCD) Test_NsxtApplicationPortProfileReadTenant(check *C) {
	skipNoNsxtConfiguration(vcd, check)
	skipOpenApiEndpointTest(vcd, check, types.OpenApiPathVersion1_0_0+types.OpenApiEndpointAppPortProfiles)

	testApplicationProfilesForScope(types.ApplicationPortProfileScopeTenant, check, vcd)
}

func getAppProfileProvider(vcd *TestVCD, check *C) *types.NsxtAppPortProfile {
	nsxtManager, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)

	nsxtManagerUuid, err := GetUuidFromHref(nsxtManager[0].HREF, true)
	check.Assert(err, IsNil)

	nsxtManagerUrn, err := BuildUrnWithUuid("urn:vcloud:nsxtmanager:", nsxtManagerUuid)
	check.Assert(err, IsNil)

	// For PROVIDER scope application port profile must have ContextEntityId set as NSX-T Managers URN and no Org
	appPortProfileConfig := &types.NsxtAppPortProfile{
		Name:        check.TestName() + "PROVIDER",
		Description: "Provider config",
		ApplicationPorts: []types.NsxtAppPortProfilePort{
			types.NsxtAppPortProfilePort{
				Protocol:         "TCP",
				DestinationPorts: []string{"11000-12000"},
			},
		},
		ContextEntityId: nsxtManagerUrn,
		Scope:           types.ApplicationPortProfileScopeProvider,
	}
	return appPortProfileConfig
}

func getAppProfileTenant(vcd *TestVCD, check *C) *types.NsxtAppPortProfile {
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	// For PROVIDER scope application port profile must have ContextEntityId set as NSX-T Managers URN and no Org
	appPortProfileConfig := &types.NsxtAppPortProfile{
		Name:        check.TestName() + "TENANT",
		Description: "Provider config",
		ApplicationPorts: []types.NsxtAppPortProfilePort{
			types.NsxtAppPortProfilePort{
				Protocol:         "ICMPv4",
				DestinationPorts: []string{"any"},
			},
		},
		OrgRef: &types.OpenApiReference{ID: org.Org.ID, Name: org.Org.Name},

		ContextEntityId: vcd.nsxtVdc.Vdc.ID, // VDC ID
		Scope:           types.ApplicationPortProfileScopeTenant,
	}
	return appPortProfileConfig
}

func testAppPortProfile(appPortProfileConfig *types.NsxtAppPortProfile, scope string, vcd *TestVCD, check *C) {
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	appProfile, err := org.CreateNsxtAppPortProfile(appPortProfileConfig)
	check.Assert(err, IsNil)

	openApiEndpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles + appProfile.NsxtAppPortProfile.ID
	AddToCleanupListOpenApi(appProfile.NsxtAppPortProfile.Name, check.TestName(), openApiEndpoint)

	appPortProfileConfig.ID = appProfile.NsxtAppPortProfile.ID // Inject ID into original creation
	appPortProfileConfig.ContextEntityId = ""                  // Remove NSX-T Manager URN because read does not return it
	check.Assert(appProfile.NsxtAppPortProfile, DeepEquals, appPortProfileConfig)

	// Check update
	appProfile.NsxtAppPortProfile.Description = appProfile.NsxtAppPortProfile.Description + "-Update"
	updatedAppProfile, err := appProfile.Update(appProfile.NsxtAppPortProfile)
	check.Assert(err, IsNil)
	check.Assert(updatedAppProfile.NsxtAppPortProfile, DeepEquals, appProfile.NsxtAppPortProfile)

	// Check lookup
	foundAppProfileById, err := org.GetNsxtAppPortProfileById(appProfile.NsxtAppPortProfile.ID)
	check.Assert(err, IsNil)
	check.Assert(foundAppProfileById.NsxtAppPortProfile, DeepEquals, appProfile.NsxtAppPortProfile)

	foundAppProfileByName, err := org.GetNsxtAppPortProfileByName(appProfile.NsxtAppPortProfile.Name, scope)
	check.Assert(err, IsNil)
	check.Assert(foundAppProfileByName.NsxtAppPortProfile, DeepEquals, foundAppProfileById.NsxtAppPortProfile)

	// Check VDC and VDC Group lookup
	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	check.Assert(adminOrg, NotNil)
	vdc, vdcGroup := test_CreateVdcGroup(check, adminOrg, vcd)

	// Lookup by VDC
	foundAppProfileByNameInVdc, err := vdc.GetNsxtAppPortProfileByName(appProfile.NsxtAppPortProfile.Name, scope)
	check.Assert(err, IsNil)
	check.Assert(foundAppProfileByNameInVdc.NsxtAppPortProfile, DeepEquals, foundAppProfileById.NsxtAppPortProfile)

	foundAppProfileByNameInVdcGroup, err := vdcGroup.GetNsxtAppPortProfileByName(appProfile.NsxtAppPortProfile.Name, scope)
	check.Assert(err, IsNil)
	check.Assert(foundAppProfileByNameInVdcGroup.NsxtAppPortProfile, DeepEquals, foundAppProfileById.NsxtAppPortProfile)
	// Remove VDC group
	err = vdcGroup.Delete()
	check.Assert(err, IsNil)
	err = vdc.DeleteWait(true, true)
	check.Assert(err, IsNil)

	err = appProfile.Delete()
	check.Assert(err, IsNil)

	// Expect a not found error
	_, err = org.GetNsxtAppPortProfileById(appProfile.NsxtAppPortProfile.ID)
	check.Assert(ContainsNotFound(err), Equals, true)

	_, err = org.GetNsxtAppPortProfileByName(appProfile.NsxtAppPortProfile.Name, scope)
	check.Assert(ContainsNotFound(err), Equals, true)
}

func testApplicationProfilesForScope(scope string, check *C, vcd *TestVCD) {
	org, err := vcd.client.GetOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	resultCount := getResultCountByScope(scope, check, vcd)
	if testVerbose {
		fmt.Printf("# API shows results for scope '%s': %d\n", scope, resultCount)
	}

	appProfileSlice, err := org.GetAllNsxtAppPortProfiles(nil, scope)
	check.Assert(err, IsNil)

	if testVerbose {
		fmt.Printf("# Paginated item number for scope '%s': %d\n", scope, len(appProfileSlice))
	}

	// Ensure the amount of results is exactly the same as returned by getResultCountByScope which makes sure that
	// pagination is not broken.
	check.Assert(len(appProfileSlice), Equals, resultCount)
}

func getResultCountByScope(scope string, check *C, vcd *TestVCD) int {
	// Get element count by using a simple query and parse response directly to compare it against paginated list of items
	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointAppPortProfiles
	skipOpenApiEndpointTest(vcd, check, endpoint)
	apiVersion, err := vcd.client.Client.checkOpenApiEndpointCompatibility(endpoint)
	check.Assert(err, IsNil)

	urlRef, err := vcd.client.Client.OpenApiBuildEndpoint(endpoint)
	check.Assert(err, IsNil)

	// Limit search of audits trails to the last 12 hours so that it doesn't take too long and set pageSize to be 1 result
	// to force following pages
	queryParams := url.Values{}
	queryParams.Add("filter", "scope=="+scope)

	result := struct {
		Resulttotal int `json:"resultTotal"`
	}{}

	err = vcd.vdc.client.OpenApiGetItem(apiVersion, urlRef, queryParams, &result, nil)
	check.Assert(err, IsNil)
	return result.Resulttotal
}
