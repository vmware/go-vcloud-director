* Added IP Space Uplink CRUD support via `IpSpaceUplink` and `types.IpSpaceUplink` and
  `VCDClient.CreateIpSpaceUplink`, `VCDClient.GetAllIpSpaceUplinks`,
  `VCDClient.GetIpSpaceUplinkById`, `VCDClient.GetIpSpaceUplinkByName`, `IpSpaceUplink.Update`,
  `IpSpaceUplink.Delete` [GH-579]
* Added IP Space Allocation CRUD support via `IpSpaceIpAllocation`, `types.IpSpaceIpAllocation`,
  `types.IpSpaceIpAllocationRequest`, `types.IpSpaceIpAllocationRequestResult`. Methods
  `IpSpace.AllocateIp`, `Org.IpSpaceAllocateIp`, `Org.GetIpSpaceAllocationByTypeAndValue`,
  `IpSpace.GetAllIpSpaceAllocations`, `Org.GetIpSpaceAllocationById`, `IpSpaceIpAllocation.Update`,
  `IpSpaceIpAllocation.Delete` [GH-579]
* Added IP Space Org assignment to support Custom Quotas via `IpSpaceOrgAssignment`,
  `types.IpSpaceOrgAssignment`, `IpSpace.GetAllOrgAssignments`, `IpSpace.GetOrgAssignmentById`,
  `IpSpace.GetOrgAssignmentByOrgName`,  `IpSpace.GetOrgAssignmentByOrgId`,
  `IpSpaceOrgAssignment.Update` [GH-579]