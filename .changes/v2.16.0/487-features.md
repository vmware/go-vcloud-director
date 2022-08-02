* Added support for Dynamic Security Groups in VCD 10.3 by expanding `types.NsxtFirewallGroup` to
  accommodate fields required for Dynamic Security Groups, implemented automatic API elevation to
  v36.0. Added New functions `VdcGroup.CreateNsxtFirewallGroup`,
  `NsxtFirewallGroup.IsDynamicSecurityGroup` [GH-487]
