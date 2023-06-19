//go:build providervdc || functional || ALL

package govcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
	"strings"
)

func init() {
	testingTags["providervdc"] = "provider_vdc_test.go"
}

func (vcd *TestVCD) Test_GetProviderVdc(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	var providerVdcs []*ProviderVdc
	providerVdc, err := vcd.client.GetProviderVdcByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	providerVdcs = append(providerVdcs, providerVdc)
	providerVdc, err = vcd.client.GetProviderVdcById(providerVdc.ProviderVdc.ID)
	check.Assert(err, IsNil)
	providerVdcs = append(providerVdcs, providerVdc)
	providerVdc, err = vcd.client.GetProviderVdcByHref(providerVdc.ProviderVdc.HREF)
	check.Assert(err, IsNil)
	providerVdcs = append(providerVdcs, providerVdc)

	// Common asserts
	for _, providerVdc := range providerVdcs {
		check.Assert(providerVdc.ProviderVdc.Name, Equals, vcd.config.VCD.NsxtProviderVdc.Name)
		foundStorageProfile := false
		for _, storageProfile := range providerVdc.ProviderVdc.StorageProfiles.ProviderVdcStorageProfile {
			if storageProfile.Name == vcd.config.VCD.NsxtProviderVdc.StorageProfile {
				foundStorageProfile = true
				break
			}
		}
		check.Assert(foundStorageProfile, Equals, true)
		check.Assert(*providerVdc.ProviderVdc.IsEnabled, Equals, true)
		check.Assert(providerVdc.ProviderVdc.ComputeCapacity, NotNil)
		check.Assert(providerVdc.ProviderVdc.Status, Equals, 1)
		check.Assert(len(providerVdc.ProviderVdc.NetworkPoolReferences.NetworkPoolReference), Equals, 1)
		foundNetworkPool := false
		for _, networkPool := range providerVdc.ProviderVdc.NetworkPoolReferences.NetworkPoolReference {
			if networkPool.Name == vcd.config.VCD.NsxtProviderVdc.NetworkPool {
				foundNetworkPool = true
				break
			}
		}
		check.Assert(foundNetworkPool, Equals, true)
		check.Assert(providerVdc.ProviderVdc.Link, NotNil)
	}
}

func (vcd *TestVCD) Test_GetProviderVdcExtended(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	var providerVdcsExtended []*ProviderVdcExtended
	providerVdcExtended, err := vcd.client.GetProviderVdcExtendedByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	providerVdcsExtended = append(providerVdcsExtended, providerVdcExtended)
	providerVdcExtended, err = vcd.client.GetProviderVdcExtendedById(providerVdcExtended.VMWProviderVdc.ID)
	check.Assert(err, IsNil)
	providerVdcsExtended = append(providerVdcsExtended, providerVdcExtended)
	providerVdcExtended, err = vcd.client.GetProviderVdcExtendedByHref(providerVdcExtended.VMWProviderVdc.HREF)
	check.Assert(err, IsNil)
	providerVdcsExtended = append(providerVdcsExtended, providerVdcExtended)

	// Common asserts
	for _, providerVdcExtended := range providerVdcsExtended {
		// Basic PVDC asserts
		check.Assert(providerVdcExtended.VMWProviderVdc.Name, Equals, vcd.config.VCD.NsxtProviderVdc.Name)
		foundStorageProfile := false
		for _, storageProfile := range providerVdcExtended.VMWProviderVdc.StorageProfiles.ProviderVdcStorageProfile {
			if storageProfile.Name == vcd.config.VCD.NsxtProviderVdc.StorageProfile {
				foundStorageProfile = true
				break
			}
		}
		check.Assert(foundStorageProfile, Equals, true)
		check.Assert(*providerVdcExtended.VMWProviderVdc.IsEnabled, Equals, true)
		check.Assert(providerVdcExtended.VMWProviderVdc.ComputeCapacity, NotNil)
		check.Assert(providerVdcExtended.VMWProviderVdc.Status, Equals, 1)
		check.Assert(len(providerVdcExtended.VMWProviderVdc.NetworkPoolReferences.NetworkPoolReference), Equals, 1)
		foundNetworkPool := false
		for _, networkPool := range providerVdcExtended.VMWProviderVdc.NetworkPoolReferences.NetworkPoolReference {
			if networkPool.Name == vcd.config.VCD.NsxtProviderVdc.NetworkPool {
				foundNetworkPool = true
				break
			}
		}
		check.Assert(foundNetworkPool, Equals, true)
		check.Assert(providerVdcExtended.VMWProviderVdc.Link, NotNil)
		// Extended PVDC asserts
		check.Assert(providerVdcExtended.VMWProviderVdc.ComputeProviderScope, Equals, "vc1")
		check.Assert(len(providerVdcExtended.VMWProviderVdc.DataStoreRefs.VimObjectRef), Equals, 4)
		check.Assert(strings.HasPrefix(providerVdcExtended.VMWProviderVdc.HighestSupportedHardwareVersion, "vmx-"), Equals, true)
		check.Assert(providerVdcExtended.VMWProviderVdc.HostReferences, NotNil)
		check.Assert(providerVdcExtended.VMWProviderVdc.NsxTManagerReference, NotNil)
		check.Assert(providerVdcExtended.VMWProviderVdc.NsxTManagerReference.Name, Equals, vcd.config.VCD.Nsxt.Manager)
		check.Assert(providerVdcExtended.VMWProviderVdc.ResourcePoolRefs, NotNil)
		check.Assert(providerVdcExtended.VMWProviderVdc.VimServer, NotNil)
	}
}

func (vcd *TestVCD) Test_GetNonExistentProviderVdc(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	providerVdcExtended, err := vcd.client.GetProviderVdcExtendedByName("non-existent-pvdc")
	check.Assert(providerVdcExtended, IsNil)
	check.Assert(err, NotNil)
	providerVdcExtended, err = vcd.client.GetProviderVdcExtendedById("non-existent-pvdc")
	check.Assert(providerVdcExtended, IsNil)
	check.Assert(err, NotNil)
	providerVdcExtended, err = vcd.client.GetProviderVdcExtendedByHref("non-existent-pvdc")
	check.Assert(providerVdcExtended, IsNil)
	check.Assert(err, NotNil)
	providerVdc, err := vcd.client.GetProviderVdcByName("non-existent-pvdc")
	check.Assert(providerVdc, IsNil)
	check.Assert(err, NotNil)
	providerVdc, err = vcd.client.GetProviderVdcById("non-existent-pvdc")
	check.Assert(providerVdc, IsNil)
	check.Assert(err, NotNil)
	providerVdc, err = vcd.client.GetProviderVdcByHref("non-existent-pvdc")
	check.Assert(providerVdc, IsNil)
	check.Assert(err, NotNil)
}

func (vcd *TestVCD) Test_GetProviderVdcConvertFromExtendedToNormal(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}

	providerVdcExtended, err := vcd.client.GetProviderVdcExtendedByName(vcd.config.VCD.NsxtProviderVdc.Name)
	check.Assert(err, IsNil)
	providerVdc, err := providerVdcExtended.ToProviderVdc()
	check.Assert(err, IsNil)
	check.Assert(providerVdc.ProviderVdc.Name, Equals, vcd.config.VCD.NsxtProviderVdc.Name)
	foundStorageProfile := false
	for _, storageProfile := range providerVdc.ProviderVdc.StorageProfiles.ProviderVdcStorageProfile {
		if storageProfile.Name == vcd.config.VCD.NsxtProviderVdc.StorageProfile {
			foundStorageProfile = true
			break
		}
	}
	check.Assert(foundStorageProfile, Equals, true)
	check.Assert(*providerVdc.ProviderVdc.IsEnabled, Equals, true)
	check.Assert(providerVdc.ProviderVdc.Status, Equals, 1)
	check.Assert(len(providerVdc.ProviderVdc.NetworkPoolReferences.NetworkPoolReference), Equals, 1)
	check.Assert(providerVdc.ProviderVdc.NetworkPoolReferences.NetworkPoolReference[0].Name, Equals, vcd.config.VCD.NsxtProviderVdc.NetworkPool)
	check.Assert(providerVdc.ProviderVdc.Link, NotNil)
}

func (vcd *TestVCD) Test_CreateProviderVdc(check *C) {
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	providerVdcName := check.TestName()
	providerVdcDescription := check.TestName()
	storageProfileList, err := vcd.client.Client.QueryAllProviderVdcStorageProfiles()
	check.Assert(err, IsNil)
	check.Assert(len(storageProfileList) > 0, Equals, true)
	var storageProfile types.QueryResultProviderVdcStorageProfileRecordType
	for _, sp := range storageProfileList {
		if sp.Name == vcd.config.VCD.NsxtProviderVdc.StorageProfile {
			storageProfile = *sp
		}
	}
	check.Assert(storageProfile.HREF, Not(Equals), "")
	networkPools, err := QueryNetworkPoolByName(vcd.client, vcd.config.VCD.NsxtProviderVdc.NetworkPool)
	check.Assert(err, IsNil)
	check.Assert(len(networkPools) > 0, Equals, true)

	vcenter, err := vcd.client.GetVcenterByName(vcd.config.VCD.VimServer)
	check.Assert(err, IsNil)
	check.Assert(vcenter, NotNil)

	resourcePools, err := vcenter.GetAllAvailableResourcePools()
	check.Assert(err, IsNil)
	if len(resourcePools) == 0 {
		check.Skip("no available resource pools found for this deployment")
	}

	resourcePool := resourcePools[0]
	check.Assert(resourcePool, NotNil)

	nsxtManagers, err := vcd.client.QueryNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	check.Assert(err, IsNil)
	check.Assert(len(nsxtManagers), Equals, 1)

	/**/
	params := types.ProviderVdcCreationXml{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        providerVdcName,
		Description: providerVdcDescription,
		IsEnabled:   addrOf(true),
		NetworkPoolReferences: &types.NetworkPoolReferences{[]*types.Reference{
			{
				Name: networkPools[0].Name,
				HREF: networkPools[0].HREF,
				Type: "GENEVE",
			},
		}},
		StorageProfile:                  []string{storageProfile.Name},
		HighestSupportedHardwareVersion: "vmx-19",
		NsxTManagerReference: &types.Reference{
			HREF: nsxtManagers[0].HREF,
			ID:   extractUuid(nsxtManagers[0].HREF),
			Name: nsxtManagers[0].Name,
		},
		ResourcePoolRefs: &types.VimObjectRefs{[]*types.VimObjectRef{
			{
				VimServerRef: &types.Reference{
					HREF: vcenter.VSphereVcenter.Url,
					ID:   extractUuid(vcenter.VSphereVcenter.VcId),
					Name: vcenter.VSphereVcenter.Name,
				},
				MoRef:         resourcePool.ResourcePool.Moref,
				VimObjectType: "RESOURCE_POOL",
			},
		},
		},
		VimServer: &types.Reference{
			HREF: vcenter.VSphereVcenter.Url,
			ID:   extractUuid(vcenter.VSphereVcenter.VcId),
			Name: vcenter.VSphereVcenter.Name,
		},
	}

	providerVdc, err := vcd.client.CreateProviderVdc(&params)
	//check.Assert(err, IsNil)
	//check.Assert(providerVdc, NotNil)

	if err != nil {
		fmt.Printf("XML creation err %s\n", err)
		fmt.Printf("XML creation pvdc %v\n", providerVdc)
	}

	jsonparams := types.ProviderVdcCreation{
		Name:                            providerVdcName,
		Description:                     providerVdcDescription,
		HighestSupportedHardwareVersion: "vmx-19",
		IsEnabled:                       true,
		VimServer: []*types.Reference{
			{
				HREF: vcenter.VSphereVcenter.Url,
				ID:   extractUuid(vcenter.VSphereVcenter.VcId),
				Name: vcenter.VSphereVcenter.Name,
			},
		},
		ResourcePoolRefs: &types.VimObjectRefs{
			VimObjectRef: []*types.VimObjectRef{
				{
					VimServerRef: &types.Reference{
						HREF: vcenter.VSphereVcenter.Url,
						ID:   extractUuid(vcenter.VSphereVcenter.VcId),
						Name: vcenter.VSphereVcenter.Name,
					},
					MoRef:         resourcePool.ResourcePool.Moref,
					VimObjectType: "RESOURCE_POOL",
				},
			},
		},
		StorageProfile: []string{storageProfile.Name},
		NsxTManagerReference: types.Reference{
			HREF: nsxtManagers[0].HREF,
			ID:   extractUuid(nsxtManagers[0].HREF),
			Name: nsxtManagers[0].Name,
		},
		NetworkPool: types.Reference{
			Name: networkPools[0].Name,
			HREF: networkPools[0].HREF,
			Type: "GENEVE",
		},
		AutoCreateNetworkPool: false,
	}
	providerVdcJson, err := vcd.client.CreateProviderVdcJson(&jsonparams)
	check.Assert(err, IsNil)
	check.Assert(providerVdcJson, NotNil)
}
