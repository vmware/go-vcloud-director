package types

import (
	"encoding/xml"
	"testing"
)

func TestAdminVDCResourcePoolUnmarshal(t *testing.T) {

	adminVdcXml := `
<AdminVdc xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:vmext="http://www.vmware.com/vcloud/extension/v1.5" xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" xmlns:common="http://schemas.dmtf.org/wbem/wscim/1/common" xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" xmlns:vmw="http://www.vmware.com/schema/ovf" xmlns:ovfenv="http://schemas.dmtf.org/ovf/environment/1" xmlns:ns9="http://www.vmware.com/vcloud/versions" status="1" isLegacyType="false" name="nl- - Production" id="urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/vnd.vmware.admin.vdc+xml">
    <VCloudExtension required="false">
        <vmext:VimObjectRef>
            <vmext:VimServerRef href="https://testcloud/api/admin/extension/vimServer/d5b16253-9f4b-4652-936c-bee560901797" id="urn:vcloud:vimserver:d5b16253-9f4b-4652-936c-bee560901797" type="application/vnd.vmware.admin.vmwvirtualcenter+xml" name="VC"/>
            <vmext:MoRef>resgroup-1696</vmext:MoRef>
            <vmext:VimObjectType>RESOURCE_POOL</vmext:VimObjectType>
        </vmext:VimObjectRef>
    </VCloudExtension>
    <Link rel="up" href="https://testcloud/api/admin/org/571dbc2f-a55c-44d7-937f-fa3f6f9ef554" type="application/vnd.vmware.admin.organization+xml"/>
    <Link rel="up" href="https://testcloud/api/admin/org/571dbc2f-a55c-44d7-937f-fa3f6f9ef554" type="application/vnd.vmware.admin.organization+json"/>
    <Link rel="down" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/metadata" type="application/vnd.vmware.vcloud.metadata+xml"/>
    <Link rel="down" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/metadata" type="application/vnd.vmware.vcloud.metadata+json"/>
    <Link rel="remove" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19"/>
    <Link rel="edit" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/vnd.vmware.admin.vdc+xml"/>
    <Link rel="edit" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/vnd.vmware.admin.vdc+json"/>
    <Link rel="disable" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/action/disable"/>
    <Link rel="add" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/action/registerVApp" type="application/vnd.vmware.admin.registerVAppParams+xml"/>
    <Link rel="add" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/action/registerVApp" type="application/vnd.vmware.admin.registerVAppParams+json"/>
    <Link rel="down" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/computePolicies" type="application/vnd.vmware.vcloud.vdcComputePolicyReferences+xml"/>
    <Link rel="down" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/computePolicies" type="application/vnd.vmware.vcloud.vdcComputePolicyReferences+json"/>
    <Link rel="down" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/controlAccess/" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link rel="down" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/controlAccess/" type="application/vnd.vmware.vcloud.controlAccess+json"/>
    <Link rel="controlAccess" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/action/controlAccess" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
    <Link rel="controlAccess" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/action/controlAccess" type="application/vnd.vmware.vcloud.controlAccess+json"/>
    <Link rel="down:vdcNetworkProfile" href="https://testcloud/cloudapi/1.0.0/vdcs/urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19/networkProfile" type="application/json"/>
    <Link rel="vdcCapabilities" href="https://testcloud/cloudapi/1.0.0/vdcs/urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19/capabilities" type="application/json"/>
    <Link rel="down:firewallGroups" href="https://testcloud/cloudapi/1.0.0/firewallGroups" type="application/json"/>
    <Link rel="down:appPortProfiles" href="https://testcloud/cloudapi/1.0.0/applicationPortProfiles" type="application/json"/>
    <Link rel="down" model="VdcComputePolicy" href="https://testcloud/cloudapi/1.0.0/vdcs/urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19/maxComputePolicy" type="application/json"/>
    <Link rel="edit" model="VdcComputePolicy" href="https://testcloud/cloudapi/1.0.0/vdcs/urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19/maxComputePolicy" type="application/json"/>
    <Link rel="down" model="VdcComputePolicies" href="https://testcloud/cloudapi/1.0.0/vdcs/urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19/computePolicies" type="application/json"/>
    <Link rel="down" model="VdcComputePolicies2" href="https://testcloud/cloudapi/2.0.0/vdcs/urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19/computePolicies" type="application/json"/>
    <Link rel="alternate" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/vnd.vmware.admin.vdc+xml"/>
    <Link rel="alternate" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/vnd.vmware.admin.vdc+json"/>
    <Link rel="add" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/vmAffinityRules/" type="application/vnd.vmware.vcloud.vmaffinityrule+xml"/>
    <Link rel="add" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/vmAffinityRules/" type="application/vnd.vmware.vcloud.vmaffinityrule+json"/>
    <Link rel="down" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/vmAffinityRules/" type="application/vnd.vmware.vcloud.vmaffinityrules+xml"/>
    <Link rel="down" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/vmAffinityRules/" type="application/vnd.vmware.vcloud.vmaffinityrules+json"/>
    <Link rel="add" model="EdgeGateway" href="https://testcloud/cloudapi/1.0.0/edgeGateways" type="application/json"/>
    <Link rel="alternate" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/vnd.vmware.vcloud.vdc+xml"/>
    <Link rel="alternate" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/vnd.vmware.vcloud.vdc+json"/>
    <Link rel="edgeGateways" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/edgeGateways" type="application/vnd.vmware.vcloud.query.records+xml"/>
    <Link rel="edgeGateways" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/edgeGateways" type="application/vnd.vmware.vcloud.query.records+json"/>
    <Link rel="add" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/networks" type="application/vnd.vmware.vcloud.orgVdcNetwork+xml"/>
    <Link rel="add" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/networks" type="application/vnd.vmware.vcloud.orgVdcNetwork+json"/>
    <Link rel="import:network" href="https://testcloud/network/orgvdcnetworks/import?orgVdc=c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/*+xml"/>
    <Link rel="down:importableSwitches" href="https://testcloud/network/orgvdcnetworks/importableswitches?orgVdc=c56ed4a4-9dec-4862-987a-5ebb601d7d19" type="application/*+xml"/>
    <Link rel="orgVdcNetworks" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/networks" type="application/vnd.vmware.vcloud.query.records+xml"/>
    <Link rel="orgVdcNetworks" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/networks" type="application/vnd.vmware.vcloud.query.records+json"/>
    <Link rel="down" href="https://testcloud/api/admin/extension/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/resourcePools" type="application/vnd.vmware.admin.OrganizationVdcResourcePoolSet+xml"/>
    <Link rel="down" href="https://testcloud/api/admin/extension/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/resourcePools" type="application/vnd.vmware.admin.OrganizationVdcResourcePoolSet+json"/>
    <Link rel="edit" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/vdcStorageProfiles" type="application/vnd.vmware.admin.updateVdcStorageProfiles+xml"/>
    <Link rel="edit" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/vdcStorageProfiles" type="application/vnd.vmware.admin.updateVdcStorageProfiles+json"/>
    <Link rel="down" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/extension" type="application/vnd.vmware.admin.extensibility.selectors+xml"/>
    <Link rel="down" href="https://testcloud/api/admin/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/extension" type="application/vnd.vmware.admin.extensibility.selectors+json"/>
    <Description>Customer nl- Production VDC</Description>
    <ComputeProviderScope>NL-AZ1-dc2-vcenter-t01</ComputeProviderScope>
    <NetworkProviderScope>dc2-nsxmgr</NetworkProviderScope>
    <AllocationModel>Flex</AllocationModel>
    <ComputeCapacity>
        <Cpu>
            <Units>MHz</Units>
            <Allocated>0</Allocated>
            <Limit>0</Limit>
            <Reserved>0</Reserved>
            <Used>62000</Used>
            <ReservationUsed>0</ReservationUsed>
        </Cpu>
        <Memory>
            <Units>MB</Units>
            <Allocated>0</Allocated>
            <Limit>0</Limit>
            <Reserved>0</Reserved>
            <Used>65536</Used>
            <ReservationUsed>0</ReservationUsed>
        </Memory>
    </ComputeCapacity>
    <ResourceEntities>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-4ee28475-65ca-4f13-944a-dc7fc6275c65" id="urn:vcloud:vapp:4ee28475-65ca-4f13-944a-dc7fc6275c65" type="application/vnd.vmware.vcloud.vApp+xml" name="cse313"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-002c0db6-e0c2-4f9d-839d-93c384b87d66" id="urn:vcloud:vapp:002c0db6-e0c2-4f9d-839d-93c384b87d66" type="application/vnd.vmware.vcloud.vApp+xml" name="test-native"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-46fb4453-ca37-4205-82e7-a2f2d54dfb36" id="urn:vcloud:vapp:46fb4453-ca37-4205-82e7-a2f2d54dfb36" type="application/vnd.vmware.vcloud.vApp+xml" name="scan"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-88b7453c-1e58-4f2f-a500-a4c3246c0911" id="urn:vcloud:vapp:88b7453c-1e58-4f2f-a500-a4c3246c0911" type="application/vnd.vmware.vcloud.vApp+xml" name="importtest3"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-a4424ee0-8245-4b23-b6a1-221313815978" id="urn:vcloud:vapp:a4424ee0-8245-4b23-b6a1-221313815978" type="application/vnd.vmware.vcloud.vApp+xml" name="usertest3"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-a0e6737a-69fb-48df-9151-3ee2791cedd7" id="urn:vcloud:vapp:a0e6737a-69fb-48df-9151-3ee2791cedd7" type="application/vnd.vmware.vcloud.vApp+xml" name="tkgm03"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-4b4cf0a4-8f15-4a1f-9f4d-9d2a7a0420c2" id="urn:vcloud:vapp:4b4cf0a4-8f15-4a1f-9f4d-9d2a7a0420c2" type="application/vnd.vmware.vcloud.vApp+xml" name="tanzu-management"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-6ef0e62b-fa6b-493c-b1fd-94633eff042f" id="urn:vcloud:vapp:6ef0e62b-fa6b-493c-b1fd-94633eff042f" type="application/vnd.vmware.vcloud.vApp+xml" name="cse313-2"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-bf130e31-1a52-4b50-9388-2f2ce9e1e252" id="urn:vcloud:vapp:bf130e31-1a52-4b50-9388-2f2ce9e1e252" type="application/vnd.vmware.vcloud.vApp+xml" name="poging2"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-143a37da-fff8-4a5f-a174-b0cb0c4d8210" id="urn:vcloud:vapp:143a37da-fff8-4a5f-a174-b0cb0c4d8210" type="application/vnd.vmware.vcloud.vApp+xml" name="importtest"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-2c0e2c93-9f2a-4b90-b47d-07112f06c990" id="urn:vcloud:vapp:2c0e2c93-9f2a-4b90-b47d-07112f06c990" type="application/vnd.vmware.vcloud.vApp+xml" name="tkgm02"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-b9e36481-6b3b-4591-89a3-105614bc992e" id="urn:vcloud:vapp:b9e36481-6b3b-4591-89a3-105614bc992e" type="application/vnd.vmware.vcloud.vApp+xml" name="test-no-dnsname"/>
        <ResourceEntity href="https://testcloud/api/vApp/vapp-9cba34e8-1d2e-4ac9-bd47-89daeeff1a0b" id="urn:vcloud:vapp:9cba34e8-1d2e-4ac9-bd47-89daeeff1a0b" type="application/vnd.vmware.vcloud.vApp+xml" name="importtest2"/>
        <ResourceEntity href="https://testcloud/api/disk/5d9c8173-8bd1-4f1e-9329-6fc52169d4b3" id="urn:vcloud:disk:5d9c8173-8bd1-4f1e-9329-6fc52169d4b3" type="application/vnd.vmware.vcloud.disk+xml" name="pvc-6e760f0e-07b1-4dc7-99b7-12395a3ee149"/>
        <ResourceEntity href="https://testcloud/api/disk/ae516e43-3f58-4992-9e12-0848d0a1d241" id="urn:vcloud:disk:ae516e43-3f58-4992-9e12-0848d0a1d241" type="application/vnd.vmware.vcloud.disk+xml" name="-c01-disk01"/>
        <ResourceEntity href="https://testcloud/api/disk/d6597a66-0cbf-40d5-bc47-8fbddaf7223e" id="urn:vcloud:disk:d6597a66-0cbf-40d5-bc47-8fbddaf7223e" type="application/vnd.vmware.vcloud.disk+xml" name="pvc-97aa2c91-2c31-4296-bddb-5afb47d72efc"/>
        <ResourceEntity href="https://testcloud/api/disk/6d0f0a0e-0ee0-4bfc-8c19-3244c9b990c8" id="urn:vcloud:disk:6d0f0a0e-0ee0-4bfc-8c19-3244c9b990c8" type="application/vnd.vmware.vcloud.disk+xml" name="pvc-dc00cba9-8fb5-44b8-8d32-bf84f356e2df"/>
    </ResourceEntities>
    <AvailableNetworks>
        <Network href="https://testcloud/api/admin/network/27bbca00-d7f0-4ad8-ae09-601bccdeb887" id="urn:vcloud:network:27bbca00-d7f0-4ad8-ae09-601bccdeb887" type="application/vnd.vmware.admin.network+xml" name="prod"/>
    </AvailableNetworks>
    <Capabilities>
        <SupportedHardwareVersions>
            <SupportedHardwareVersion name="vmx-04" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-04" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-04</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-07" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-07" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-07</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-08" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-08" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-08</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-09" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-09" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-09</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-10" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-10" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-10</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-11" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-11" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-11</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-12" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-12" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-12</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-13" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-13" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-13</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-14" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-14" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-14</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-15" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-15" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-15</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-16" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-16" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-16</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-17" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-17" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-17</SupportedHardwareVersion>
            <SupportedHardwareVersion name="vmx-18" href="https://testcloud/api/vdc/c56ed4a4-9dec-4862-987a-5ebb601d7d19/hwv/vmx-18" default="true" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-18</SupportedHardwareVersion>
        </SupportedHardwareVersions>
    </Capabilities>
    <NicQuota>0</NicQuota>
    <NetworkQuota>100</NetworkQuota>
    <UsedNetworkCount>0</UsedNetworkCount>
    <VmQuota>0</VmQuota>
    <IsEnabled>true</IsEnabled>
    <VdcStorageProfiles>
        <VdcStorageProfile href="https://testcloud/api/admin/vdcStorageProfile/67a9d4ec-fd09-40e0-9626-13b38260be71" id="urn:vcloud:vdcstorageProfile:67a9d4ec-fd09-40e0-9626-13b38260be71" type="application/vnd.vmware.admin.vdcStorageProfile+xml" name="VM Premium Storage"/>
    </VdcStorageProfiles>
    <DefaultComputePolicy href="https://testcloud/cloudapi/1.0.0/vdcComputePolicies/urn:vcloud:vdcComputePolicy:5fab95be-2790-44ed-9983-c03aced4b34f" id="urn:vcloud:vdcComputePolicy:5fab95be-2790-44ed-9983-c03aced4b34f" type="application/json" name="System Default"/>
    <MaxComputePolicy href="https://testcloud/cloudapi/1.0.0/vdcs/urn:vcloud:vdc:c56ed4a4-9dec-4862-987a-5ebb601d7d19/maxComputePolicy" type="application/json"/>
    <VCpuInMhz2>1000</VCpuInMhz2>
    <ResourceGuaranteedMemory>0.0</ResourceGuaranteedMemory>
    <ResourceGuaranteedCpu>0.0</ResourceGuaranteedCpu>
    <VCpuInMhz>1000</VCpuInMhz>
    <IsThinProvision>true</IsThinProvision>
    <NetworkPoolReference href="https://testcloud/api/admin/extension/networkPool/73e3748b-5ff5-475f-800f-0aa7bfe91f67" id="urn:vcloud:networkpool:73e3748b-5ff5-475f-800f-0aa7bfe91f67" type="application/vnd.vmware.admin.networkPool+xml" name=""/>
    <VendorServices/>
    <ProviderVdcReference href="https://testcloud/api/admin/providervdc/937e62a1-0670-4814-9b57-d281dfbd1192" id="urn:vcloud:providervdc:937e62a1-0670-4814-9b57-d281dfbd1192" type="application/vnd.vmware.admin.providervdc+xml" name="NL-AZ1-TEST-NIX-Provider"/>
    <ResourcePoolRefs>
        <vmext:VimObjectRef>
            <vmext:VimServerRef href="https://testcloud/api/admin/extension/vimServer/d5b16253-9f4b-4652-936c-bee560901797" id="urn:vcloud:vimserver:d5b16253-9f4b-4652-936c-bee560901797" type="application/vnd.vmware.admin.vmwvirtualcenter+xml" name="VC"/>
            <vmext:MoRef>resgroup-1696</vmext:MoRef>
            <vmext:VimObjectType>RESOURCE_POOL</vmext:VimObjectType>
        </vmext:VimObjectRef>
    </ResourcePoolRefs>
    <UsesFastProvisioning>false</UsesFastProvisioning>
    <VmDiscoveryEnabled>true</VmDiscoveryEnabled>
    <IsElastic>true</IsElastic>
    <IncludeMemoryOverhead>false</IncludeMemoryOverhead>
</AdminVdc>
`

	myAdminVDC := AdminVdc{}

	err := xml.Unmarshal([]byte(adminVdcXml), &myAdminVDC)
	if err != nil {
		t.FailNow()
	}
	if len(myAdminVDC.ResourcePoolRefs) != 1 {
		t.FailNow()
	}
	if myAdminVDC.ResourcePoolRefs[0].MoRef != "resgroup-1696" {
		t.FailNow()
	}
	if myAdminVDC.ResourcePoolRefs[0].VimObjectType != "RESOURCE_POOL" {
		t.FailNow()
	}
	if myAdminVDC.ResourcePoolRefs[0].VimServerRef == nil {
		t.FailNow()
	}

	expectedVIMServerRef := Reference{
		HREF: "https://testcloud/api/admin/extension/vimServer/d5b16253-9f4b-4652-936c-bee560901797",
		ID:   "urn:vcloud:vimserver:d5b16253-9f4b-4652-936c-bee560901797",
		Type: "application/vnd.vmware.admin.vmwvirtualcenter+xml",
		Name: "VC",
	}

	if *myAdminVDC.ResourcePoolRefs[0].VimServerRef != expectedVIMServerRef {
		t.FailNow()
	}
}
