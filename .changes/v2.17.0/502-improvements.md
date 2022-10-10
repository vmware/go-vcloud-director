* Created new VDC Compute Policies CRUD methods using OpenAPI v2.0.0:
  `VCDClient.GetVdcComputePolicyV2ById`, `VCDClient.GetAllVdcComputePoliciesV2`, `VCDClient.CreateVdcComputePolicyV2`,
  `VdcComputePolicyV2.Update`, `VdcComputePolicyV2.Delete` and `AdminVdc.GetAllAssignedVdcComputePoliciesV2`.
  This version supports more filtering options like `isVgpuPolicy` [GH-502], [GH-504], [GH-507]
