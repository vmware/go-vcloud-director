/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package types

const (
	// PublicCatalog Name
	PublicCatalog = "Public Catalog"

	// DefaultCatalog Name
	DefaultCatalog = "Default Catalog"

	// JSONMimeV57 the json mime for version 5.7 of the API
	JSONMimeV57 = "application/json;version=5.7"
	// AnyXMLMime511 the wildcard xml mime for version 5.11 of the API
	AnyXMLMime511 = "application/*+xml;version=5.11"
	// Version511 the 5.11 version
	Version511 = "5.11"
	// Version is the default version number
	Version = Version511
)

const (
	// MimeOrgList mime for org list
	MimeOrgList = "application/vnd.vmware.vcloud.orgList+xml"
	// MimeOrg mime for org
	MimeOrg = "application/vnd.vmware.vcloud.org+xml"
	// MimeCatalog mime for catalog
	MimeCatalog = "application/vnd.vmware.vcloud.catalog+xml"
	// MimeCatalogItem mime for catalog item
	MimeCatalogItem = "application/vnd.vmware.vcloud.catalogItem+xml"
	// MimeVDC mime for a VDC
	MimeVDC = "application/vnd.vmware.vcloud.vdc+xml"
	// MimeVAppTemplate mime for a vapp template
	MimeVAppTemplate = "application/vnd.vmware.vcloud.vAppTemplate+xml"
	// MimeVApp mime for a vApp
	MimeVApp = "application/vnd.vmware.vcloud.vApp+xml"
	// MimeQueryRecords mime for the query records
	MimeQueryRecords = "application/vnd.vmware.vchs.query.records+xml"
	// MimeAPIExtensibility mime for api extensibility
	MimeAPIExtensibility = "application/vnd.vmware.vcloud.apiextensibility+xml"
	// MimeEntity mime for vcloud entity
	MimeEntity = "application/vnd.vmware.vcloud.entity+xml"
	// MimeQueryList mime for query list
	MimeQueryList = "application/vnd.vmware.vcloud.query.queryList+xml"
	// MimeSession mime for a session
	MimeSession = "application/vnd.vmware.vcloud.session+xml"
	// MimeTask mime for task
	MimeTask = "application/vnd.vmware.vcloud.task+xml"
	// MimeError mime for error
	MimeError = "application/vnd.vmware.vcloud.error+xml"
	// MimeNetwork mime for a network
	MimeNetwork = "application/vnd.vmware.vcloud.network+xml"
	//MimeDiskCreateParams mime for create independent disk
	MimeDiskCreateParams = "application/vnd.vmware.vcloud.diskCreateParams+xml"
	// Mime for VMs
	MimeVMs = "application/vnd.vmware.vcloud.vms+xml"
	// Mime for attach or detach independent disk
	MimeDiskAttachOrDetachParams = "application/vnd.vmware.vcloud.diskAttachOrDetachParams+xml"
	// Mime for Disk
	MimeDisk = "application/vnd.vmware.vcloud.disk+xml"
	// Mime for insert or eject media
	MimeMediaInsertOrEjectParams = "application/vnd.vmware.vcloud.mediaInsertOrEjectParams+xml"
	// Mime for catalog
	MimeAdminCatalog = "application/vnd.vmware.admin.catalog+xml"
	// Mime for networkConnectionSection
	MimeNetworkConnectionSection = "application/vnd.vmware.vcloud.networkConnectionSection+xml"
	// Mime for Item
	MimeRasdItem = "application/vnd.vmware.vcloud.rasdItem+xml"
	// Mime for guest customization section
	MimeGuestCustomizationSection = "application/vnd.vmware.vcloud.guestCustomizationSection+xml"
	// Mime for network config section
	MimeNetworkConfigSection = "application/vnd.vmware.vcloud.networkconfigsection+xml"
	// Mime for recompose vApp params
	MimeRecomposeVappParams = "application/vnd.vmware.vcloud.recomposeVAppParams+xml"
	// Mime for compose vApp params
	MimeComposeVappParams = "application/vnd.vmware.vcloud.composeVAppParams+xml"
	// Mime for undeploy vApp params
	MimeUndeployVappParams = "application/vnd.vmware.vcloud.undeployVAppParams+xml"
	// Mime for deploy vApp params
	MimeDeployVappParams = "application/vnd.vmware.vcloud.deployVAppParams+xml"
	// Mime for VM
	MimeVM = "application/vnd.vmware.vcloud.vm+xml"
	// Mime for instantiate vApp template params
	MimeInstantiateVappTemplateParams = "application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml"
	// Mime for product section
	MimeProductSection = "application/vnd.vmware.vcloud.productSections+xml"
	// Mime for metadata
	MimeMetaData = "application/vnd.vmware.vcloud.metadata+xml"
	// Mime for metadata value
	MimeMetaDataValue = "application/vnd.vmware.vcloud.metadata.value+xml"
)

const (
	VMsCDResourceSubType = "vmware.cdrom.iso"
)

// https://blogs.vmware.com/vapp/2009/11/virtual-hardware-in-ovf-part-1.html

const (
	ResourceTypeOther     int = 0
	ResourceTypeProcessor int = 3
	ResourceTypeMemory    int = 4
	ResourceTypeIDE       int = 5
	ResourceTypeSCSI      int = 6
	ResourceTypeEthernet  int = 10
	ResourceTypeFloppy    int = 14
	ResourceTypeCD        int = 15
	ResourceTypeDVD       int = 16
	ResourceTypeDisk      int = 17
	ResourceTypeUSB       int = 23
)

const (
	FenceModeIsolated = "isolated"
	FenceModeBridged  = "bridged"
	FenceModeNAT      = "natRouted"
)

const (
	IPAllocationModeDHCP   = "DHCP"
	IPAllocationModeManual = "MANUAL"
	IPAllocationModeNone   = "NONE"
	IPAllocationModePool   = "POOL"
)

const (
	XMLNamespaceVCloud = "http://www.vmware.com/vcloud/v1.5"
	XMLNamespaceOVF    = "http://schemas.dmtf.org/ovf/envelope/1"
	XMLNamespaceVMW    = "http://www.vmware.com/schema/ovf"
	XMLNamespaceXSI    = "http://www.w3.org/2001/XMLSchema-instance"
	XMLNamespaceRASD   = "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData"
	XMLNamespaceVSSD   = "http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData"
)
