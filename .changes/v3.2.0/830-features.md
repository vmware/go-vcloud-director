* Add `TmAviController`, `TmAviServiceEngineGroup`, `TmAviServiceEngineGroupAssignment`,
  `TmDistributedVxlanConnection`, `TmInfraPolicy`, `TmNamespace`, `TmProjectAssignment`,
  `TmVCenterComputePolicy`, and `TmVirtualNetworkApplianceCluster` resource wrappers, each with a
  property holding the underlying type and the retained `TmOperationOptions` [GH-830]
* Add `VCDClient.CreateTmAviController`, `VCDClient.CreateTmAviControllerAsync`,
  `VCDClient.GetAllTmAviControllers`, `VCDClient.GetTmAviControllerById`, `TmAviController.Update`,
  `TmAviController.UpdateAsync`, `TmAviController.Delete`, and `TmAviController.DeleteAsync`
  methods to manage Avi Load Balancer Controllers [GH-830]
* Add `TmAviController.Sync` and `TmAviController.SyncAsync` methods to sync an Avi Load Balancer
  Controller along with its associated Cloud and Service Engine Groups [GH-830]
* Add `VCDClient.GetAllTmAviServiceEngineGroups` and `VCDClient.GetTmAviServiceEngineGroupById`
  methods to retrieve Avi Load Balancer Service Engine Groups [GH-830]
* Add `VCDClient.CreateTmAviServiceEngineGroupAssignment`,
  `VCDClient.CreateTmAviServiceEngineGroupAssignmentAsync`,
  `VCDClient.GetAllTmAviServiceEngineGroupAssignments`,
  `VCDClient.GetTmAviServiceEngineGroupAssignmentById`,
  `TmAviServiceEngineGroupAssignment.Delete`, and `TmAviServiceEngineGroupAssignment.DeleteAsync`
  methods to manage Avi Service Engine Group Assignments [GH-830]
* Add `VCDClient.GetAllTmDistributedVxlanConnections`,
  `VCDClient.CreateTmDistributedVxlanConnection`,
  `VCDClient.CreateTmDistributedVxlanConnectionAsync`,
  `VCDClient.GetTmDistributedVxlanConnectionById`, `TmDistributedVxlanConnection.Update`,
  `TmDistributedVxlanConnection.UpdateAsync`, `TmDistributedVxlanConnection.Delete`, and
  `TmDistributedVxlanConnection.DeleteAsync` methods to manage Distributed VXLAN Connections
  [GH-830]
* Add `VCDClient.GetAllTmInfraPolicies`, `VCDClient.CreateTmInfraPolicy`,
  `VCDClient.CreateTmInfraPolicyAsync`, `VCDClient.GetTmInfraPolicyById`, `TmInfraPolicy.Update`,
  `TmInfraPolicy.Delete`, and `TmInfraPolicy.DeleteAsync` methods to manage infra policies
  [GH-830]
* Add `TmInfraPolicy.GetAllVCenterPolicies`, `TmInfraPolicy.SyncToVCenters`, and
  `TmInfraPolicy.SyncToVCentersAsync` methods to retrieve vCenter infra policies and sync an infra
  policy to vCenters [GH-830]
* Add `VCDClient.GetAllTmVCenterComputePolicies`, `VCDClient.GetTmVCenterComputePolicyById`,
  `TmVCenterComputePolicy.GetAllTags`, and `VCDClient.GetAllTmVCenterComputePolicyNames` methods to
  retrieve vCenter compute policies, their tags, and unique policy names [GH-830]
* Add `VCDClient.GetAllTmVirtualNetworkApplianceClusters`,
  `VCDClient.GetTmVirtualNetworkApplianceClusterById`, and
  `TmVirtualNetworkApplianceCluster.GetAllTransportNodeStatuses` methods to retrieve Virtual
  Network Appliance Clusters and the status of their transport nodes [GH-830]
* Add `VCDClient.SyncTmVirtualNetworkApplianceClusters` and
  `VCDClient.SyncTmVirtualNetworkApplianceClustersAsync` methods to sync Virtual Network Appliance
  Clusters from NSX Manager to Provider Management [GH-830]
* Add `VCDClient.GetAllTmProjectAccessLevels`, `VCDClient.GetAllTmProjectAssignments`,
  `VCDClient.CreateTmProjectAssignment`, `VCDClient.CreateTmProjectAssignmentAsync`,
  `VCDClient.GetTmProjectAssignmentById`, `TmProjectAssignment.Update`,
  `TmProjectAssignment.UpdateAsync`, `TmProjectAssignment.Delete`, and
  `TmProjectAssignment.DeleteAsync` methods to manage Project Assignments and retrieve Project
  Access Levels [GH-830]
* Add `TmProjectAssignment.GetAllMembers`, `TmProjectAssignment.UpdateMembers`, and
  `TmProjectAssignment.UpdateMembersAsync` methods to retrieve and update Project Assignment
  Members [GH-830]
* Add `VCDClient.GetAllTmNamespaceSummaries`, `VCDClient.CreateTmNamespace`,
  `VCDClient.CreateTmNamespaceAsync`, `VCDClient.GetTmNamespaceById`, `TmNamespace.Update`,
  `TmNamespace.UpdateAsync`, `TmNamespace.Delete`, and `TmNamespace.DeleteAsync` methods to manage
  Namespaces [GH-830]
* Add `TmNamespace.GetAllCapabilities`, `VCDClient.ImportTmNamespace`,
  `VCDClient.ImportTmNamespaceAsync`, and `VCDClient.GetAllTmNamespaceStorageClassMetrics` methods
  to retrieve namespace capabilities, import an existing Namespace from vCenter, and retrieve
  namespace storage class usage [GH-830]
* Add `types.TmVCenterComputePolicies`, `types.TmVCenterComputePolicy`,
  `types.TmVCenterComputePolicyTags`, `types.TmVCenterComputePolicyTag`,
  `types.TmVCenterComputePolicyNames`, and `types.TmVCenterComputePolicyName` structures for vCenter
  compute policies [GH-830]
* Add `types.TmInfraPolicies`, `types.TmInfraPolicy`, `types.TmVCenterInfraPolicies`,
  `types.TmVCenterInfraPolicy`, `types.TmPolicyRule`, `types.TmRegionZones`,
  `types.TmWorkloadPolicyRule`, `types.TmGuestOsRule`, `types.TmGuestOsFamilyRule`, and
  `types.TmLabelSelectorRule` structures for infra policies [GH-830]
* Add `types.TmAviControllers`, `types.TmAviController`, `types.TmAviServiceEngineGroups`,
  `types.TmAviServiceEngineGroup`, `types.TmAviServiceEngineGroupAssignments`, and
  `types.TmAviServiceEngineGroupAssignment` structures for Avi Load Balancer resources [GH-830]
* Add `types.TmDistributedVxlanConnections`, `types.TmDistributedVxlanConnection`, and
  `types.TmVrfRouteTargets` structures for Distributed VXLAN Connections [GH-830]
* Add `types.TmVirtualNetworkApplianceClusters`, `types.TmVirtualNetworkApplianceCluster`, and
  `types.TmEdgeClusterTransportNodeStatusList` structures for Virtual Network Appliance Clusters
  [GH-830]
* Add `types.TmProjectAccessLevels`, `types.TmProjectAccessLevel`, `types.TmProjectAssignments`,
  `types.TmProjectAssignment`, `types.TmProjectAssignmentMembers`, and
  `types.TmProjectAssignmentMember` structures for Project Assignments [GH-830]
* Add `types.TmNamespaceSummaries`, `types.TmNamespace`, `types.TmCapabilities`,
  `types.TmNamespaceImportDetails`, `types.TmNamespaceStorageClassMetrics`,
  `types.TmNamespaceStorageClassMetric`, `types.TmNamespaceStorageClass`,
  `types.TmNamespaceSharedSubnet`, and `types.TmNamespaceInfraPolicy` structures for Namespaces
  [GH-830]
* Add `types.TmAssociation` structure with `EntityID` and `AssociationId` properties, used by the
  paginated list structures [GH-830]
