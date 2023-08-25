package govcd

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// QueryItem is an entity that is used to evaluate a Condition
type QueryItem interface {
	GetDate() string
	GetName() string
	GetType() string
	GetIp() string
	GetMetadataValue(key string) string
	GetParentName() string
	GetParentId() string
	GetHref() string
}

type (
	// All the Query* types are localizations of Query records that can be returned from a query.
	// Each one of these implements the QueryItem interface
	QueryVAppTemplate  types.QueryResultVappTemplateType
	QueryCatalogItem   types.QueryResultCatalogItemType
	QueryEdgeGateway   types.QueryResultEdgeGatewayRecordType
	QueryAdminCatalog  types.AdminCatalogRecord
	QueryCatalog       types.CatalogRecord
	QueryOrgVdcNetwork types.QueryResultOrgVdcNetworkRecordType
	QueryMedia         types.MediaRecordType
	QueryVapp          types.QueryResultVAppRecordType
	QueryVm            types.QueryResultVMRecordType
	QueryOrgVdc        types.QueryResultOrgVdcRecordType
	QueryTask          types.QueryResultTaskRecordType
	QueryAdminTask     types.QueryResultTaskRecordType
	QueryOrg           types.QueryResultOrgRecordType
)

// getMetadataValue is a generic metadata lookup for all query items
func getMetadataValue(metadata *types.Metadata, key string) string {
	if metadata == nil || len(metadata.MetadataEntry) == 0 {
		return ""
	}
	for _, x := range metadata.MetadataEntry {
		if key == x.Key {
			return x.TypedValue.Value
		}
	}
	return ""
}

// --------------------------------------------------------------
// Org VDC
// --------------------------------------------------------------
func (orgVdc QueryOrgVdc) GetHref() string       { return orgVdc.HREF }
func (orgVdc QueryOrgVdc) GetName() string       { return orgVdc.Name }
func (orgVdc QueryOrgVdc) GetType() string       { return "org_vdc" }
func (orgVdc QueryOrgVdc) GetIp() string         { return "" } // IP does not apply to VDC
func (orgVdc QueryOrgVdc) GetDate() string       { return "" } // Date does not aply to VDC
func (orgVdc QueryOrgVdc) GetParentName() string { return orgVdc.OrgName }
func (orgVdc QueryOrgVdc) GetParentId() string   { return orgVdc.Org }
func (orgVdc QueryOrgVdc) GetMetadataValue(key string) string {
	return getMetadataValue(orgVdc.Metadata, key)
}

// --------------------------------------------------------------
// vApp template
// --------------------------------------------------------------
func (vappTemplate QueryVAppTemplate) GetHref() string       { return vappTemplate.HREF }
func (vappTemplate QueryVAppTemplate) GetName() string       { return vappTemplate.Name }
func (vappTemplate QueryVAppTemplate) GetType() string       { return "vapp_template" }
func (vappTemplate QueryVAppTemplate) GetIp() string         { return "" }
func (vappTemplate QueryVAppTemplate) GetDate() string       { return vappTemplate.CreationDate }
func (vappTemplate QueryVAppTemplate) GetParentName() string { return vappTemplate.CatalogName }
func (vappTemplate QueryVAppTemplate) GetParentId() string   { return vappTemplate.Vdc }
func (vappTemplate QueryVAppTemplate) GetMetadataValue(key string) string {
	return getMetadataValue(vappTemplate.Metadata, key)
}

// --------------------------------------------------------------
// media item
// --------------------------------------------------------------
func (media QueryMedia) GetHref() string       { return media.HREF }
func (media QueryMedia) GetName() string       { return media.Name }
func (media QueryMedia) GetType() string       { return "catalog_media" }
func (media QueryMedia) GetIp() string         { return "" }
func (media QueryMedia) GetDate() string       { return media.CreationDate }
func (media QueryMedia) GetParentName() string { return media.CatalogName }
func (media QueryMedia) GetParentId() string   { return media.Catalog }
func (media QueryMedia) GetMetadataValue(key string) string {
	return getMetadataValue(media.Metadata, key)
}

// --------------------------------------------------------------
// catalog item
// --------------------------------------------------------------
func (catItem QueryCatalogItem) GetHref() string       { return catItem.HREF }
func (catItem QueryCatalogItem) GetName() string       { return catItem.Name }
func (catItem QueryCatalogItem) GetIp() string         { return "" }
func (catItem QueryCatalogItem) GetType() string       { return "catalog_item" }
func (catItem QueryCatalogItem) GetDate() string       { return catItem.CreationDate }
func (catItem QueryCatalogItem) GetParentName() string { return catItem.CatalogName }
func (catItem QueryCatalogItem) GetParentId() string   { return catItem.Catalog }
func (catItem QueryCatalogItem) GetMetadataValue(key string) string {
	return getMetadataValue(catItem.Metadata, key)
}

// --------------------------------------------------------------
// catalog
// --------------------------------------------------------------
func (catalog QueryCatalog) GetHref() string       { return catalog.HREF }
func (catalog QueryCatalog) GetName() string       { return catalog.Name }
func (catalog QueryCatalog) GetIp() string         { return "" }
func (catalog QueryCatalog) GetType() string       { return "catalog" }
func (catalog QueryCatalog) GetDate() string       { return catalog.CreationDate }
func (catalog QueryCatalog) GetParentName() string { return catalog.OrgName }
func (catalog QueryCatalog) GetParentId() string   { return "" }
func (catalog QueryCatalog) GetMetadataValue(key string) string {
	return getMetadataValue(catalog.Metadata, key)
}

func (catalog QueryAdminCatalog) GetHref() string       { return catalog.HREF }
func (catalog QueryAdminCatalog) GetName() string       { return catalog.Name }
func (catalog QueryAdminCatalog) GetIp() string         { return "" }
func (catalog QueryAdminCatalog) GetType() string       { return "catalog" }
func (catalog QueryAdminCatalog) GetDate() string       { return catalog.CreationDate }
func (catalog QueryAdminCatalog) GetParentName() string { return catalog.OrgName }
func (catalog QueryAdminCatalog) GetParentId() string   { return "" }
func (catalog QueryAdminCatalog) GetMetadataValue(key string) string {
	return getMetadataValue(catalog.Metadata, key)
}

// --------------------------------------------------------------
// edge gateway
// --------------------------------------------------------------
func (egw QueryEdgeGateway) GetHref() string       { return egw.HREF }
func (egw QueryEdgeGateway) GetName() string       { return egw.Name }
func (egw QueryEdgeGateway) GetIp() string         { return "" }
func (egw QueryEdgeGateway) GetType() string       { return "edge_gateway" }
func (egw QueryEdgeGateway) GetDate() string       { return "" }
func (egw QueryEdgeGateway) GetParentName() string { return egw.OrgVdcName }
func (egw QueryEdgeGateway) GetParentId() string   { return egw.Vdc }
func (egw QueryEdgeGateway) GetMetadataValue(key string) string {
	// Edge Gateway doesn't support metadata
	return ""
}

// --------------------------------------------------------------
// Org VDC network
// --------------------------------------------------------------
func (network QueryOrgVdcNetwork) GetHref() string { return network.HREF }
func (network QueryOrgVdcNetwork) GetName() string { return network.Name }
func (network QueryOrgVdcNetwork) GetIp() string   { return network.DefaultGateway }
func (network QueryOrgVdcNetwork) GetType() string {
	switch network.LinkType {
	case 0:
		return "network_direct"
	case 1:
		return "network_routed"
	case 2:
		return "network_isolated"
	default:
		// There are only three types so far, but just to make it future proof
		return "network"
	}
}
func (network QueryOrgVdcNetwork) GetDate() string       { return "" }
func (network QueryOrgVdcNetwork) GetParentName() string { return network.VdcName }
func (network QueryOrgVdcNetwork) GetParentId() string   { return network.Vdc }
func (network QueryOrgVdcNetwork) GetMetadataValue(key string) string {
	return getMetadataValue(network.Metadata, key)
}

// --------------------------------------------------------------
// Task
// --------------------------------------------------------------
func (task QueryTask) GetHref() string       { return task.HREF }
func (task QueryTask) GetName() string       { return task.Name }
func (task QueryTask) GetType() string       { return "Task" }
func (task QueryTask) GetIp() string         { return "" }
func (task QueryTask) GetDate() string       { return task.StartDate }
func (task QueryTask) GetParentName() string { return task.OwnerName }
func (task QueryTask) GetParentId() string   { return task.Org }
func (task QueryTask) GetMetadataValue(key string) string {
	return getMetadataValue(task.Metadata, key)
}

// --------------------------------------------------------------
// AdminTask
// --------------------------------------------------------------
func (task QueryAdminTask) GetHref() string       { return task.HREF }
func (task QueryAdminTask) GetName() string       { return task.Name }
func (task QueryAdminTask) GetType() string       { return "Task" }
func (task QueryAdminTask) GetIp() string         { return "" }
func (task QueryAdminTask) GetDate() string       { return task.StartDate }
func (task QueryAdminTask) GetParentName() string { return task.OwnerName }
func (task QueryAdminTask) GetParentId() string   { return task.Org }
func (task QueryAdminTask) GetMetadataValue(key string) string {
	return getMetadataValue(task.Metadata, key)
}

// --------------------------------------------------------------
// vApp
// --------------------------------------------------------------
func (vapp QueryVapp) GetHref() string       { return vapp.HREF }
func (vapp QueryVapp) GetName() string       { return vapp.Name }
func (vapp QueryVapp) GetType() string       { return "vApp" }
func (vapp QueryVapp) GetIp() string         { return "" }
func (vapp QueryVapp) GetDate() string       { return vapp.CreationDate }
func (vapp QueryVapp) GetParentName() string { return vapp.VdcName }
func (vapp QueryVapp) GetParentId() string   { return vapp.VdcHREF }
func (vapp QueryVapp) GetMetadataValue(key string) string {
	return getMetadataValue(vapp.MetaData, key)
}

// --------------------------------------------------------------
// VM
// --------------------------------------------------------------
func (vm QueryVm) GetHref() string       { return vm.HREF }
func (vm QueryVm) GetName() string       { return vm.Name }
func (vm QueryVm) GetType() string       { return "Vm" }
func (vm QueryVm) GetIp() string         { return vm.IpAddress }
func (vm QueryVm) GetDate() string       { return vm.DateCreated }
func (vm QueryVm) GetParentName() string { return vm.ContainerName }
func (vm QueryVm) GetParentId() string   { return vm.VdcHREF }
func (vm QueryVm) GetMetadataValue(key string) string {
	return getMetadataValue(vm.MetaData, key)
}

// --------------------------------------------------------------
// Organization
// --------------------------------------------------------------
func (org QueryOrg) GetHref() string       { return org.HREF }
func (org QueryOrg) GetName() string       { return org.Name }
func (org QueryOrg) GetType() string       { return "organization" }
func (org QueryOrg) GetDate() string       { return "" }
func (org QueryOrg) GetIp() string         { return "" }
func (org QueryOrg) GetParentId() string   { return "" }
func (org QueryOrg) GetParentName() string { return "" }
func (org QueryOrg) GetMetadataValue(key string) string {
	return getMetadataValue(org.Metadata, key)
}

// --------------------------------------------------------------
// result conversion
// --------------------------------------------------------------
// resultToQueryItems converts a set of query results into a list of query items
func resultToQueryItems(queryType string, results Results) ([]QueryItem, error) {
	resultSize := int64(results.Results.Total)
	if resultSize < 1 {
		return nil, nil
	}
	var items = make([]QueryItem, resultSize)
	switch queryType {
	case types.QtAdminCatalogItem:
		for i, item := range results.Results.AdminCatalogItemRecord {
			items[i] = QueryCatalogItem(*item)
		}
	case types.QtCatalogItem:
		for i, item := range results.Results.CatalogItemRecord {
			items[i] = QueryCatalogItem(*item)
		}
	case types.QtMedia:
		for i, item := range results.Results.MediaRecord {
			items[i] = QueryMedia(*item)
		}
	case types.QtAdminMedia:
		for i, item := range results.Results.AdminMediaRecord {
			items[i] = QueryMedia(*item)
		}
	case types.QtVappTemplate:
		for i, item := range results.Results.VappTemplateRecord {
			items[i] = QueryVAppTemplate(*item)
		}
	case types.QtAdminVappTemplate:
		for i, item := range results.Results.AdminVappTemplateRecord {
			items[i] = QueryVAppTemplate(*item)
		}
	case types.QtEdgeGateway:
		for i, item := range results.Results.EdgeGatewayRecord {
			items[i] = QueryEdgeGateway(*item)
		}
	case types.QtOrgVdcNetwork:
		for i, item := range results.Results.OrgVdcNetworkRecord {
			items[i] = QueryOrgVdcNetwork(*item)
		}
	case types.QtOrg:
		for i, item := range results.Results.OrgRecord {
			items[i] = QueryOrg(*item)
		}
	case types.QtCatalog:
		for i, item := range results.Results.CatalogRecord {
			items[i] = QueryCatalog(*item)
		}
	case types.QtAdminCatalog:
		for i, item := range results.Results.AdminCatalogRecord {
			items[i] = QueryAdminCatalog(*item)
		}
	case types.QtVm:
		for i, item := range results.Results.VMRecord {
			items[i] = QueryVm(*item)
		}
	case types.QtAdminVm:
		for i, item := range results.Results.AdminVMRecord {
			items[i] = QueryVm(*item)
		}
	case types.QtVapp:
		for i, item := range results.Results.VAppRecord {
			items[i] = QueryVapp(*item)
		}
	case types.QtAdminVapp:
		for i, item := range results.Results.AdminVAppRecord {
			items[i] = QueryVapp(*item)
		}
	case types.QtOrgVdc:
		for i, item := range results.Results.OrgVdcRecord {
			items[i] = QueryOrgVdc(*item)
		}
	case types.QtAdminOrgVdc:
		for i, item := range results.Results.OrgVdcAdminRecord {
			items[i] = QueryOrgVdc(*item)
		}
	case types.QtTask:
		for i, item := range results.Results.TaskRecord {
			items[i] = QueryTask(*item)
		}
	case types.QtAdminTask:
		for i, item := range results.Results.TaskRecord {
			items[i] = QueryAdminTask(*item)
		}

	}
	if len(items) > 0 {
		return items, nil
	}
	return nil, fmt.Errorf("unsupported query type %s", queryType)
}

// GetQueryType is an utility function to get the appropriate query type depending on
// the user's role
func (client Client) GetQueryType(queryType string) string {
	if client.IsSysAdmin {
		adminType, ok := types.AdminQueryTypes[queryType]
		if ok {
			return adminType
		} else {
			panic(fmt.Sprintf("no corresponding admin type found for type %s", queryType))
		}
	}
	return queryType
}
