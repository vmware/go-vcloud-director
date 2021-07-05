* Added NSX-T Firewall Group type (which represents a Security Group or an IP Set) support by using
  structures `NsxtFirewallGroup` and `NsxtFirewallGroupMemberVms`. The following methods are
  introduced for managing Security Groups and Ip Sets: `Vdc.CreateNsxtFirewallGroup`,
  `NsxtEdgeGateway.CreateNsxtFirewallGroup`, `Org.GetAllNsxtFirewallGroups`,
  `Vdc.GetAllNsxtFirewallGroups`, `Org.GetNsxtFirewallGroupByName`,
  `Vdc.GetNsxtFirewallGroupByName`, `NsxtEdgeGateway.GetNsxtFirewallGroupByName`,
  `Org.GetNsxtFirewallGroupById`, `Vdc.GetNsxtFirewallGroupById`,
  `NsxtEdgeGateway.GetNsxtFirewallGroupById`, `NsxtFirewallGroup.Update`,
  `NsxtFirewallGroup.Delete`, `NsxtFirewallGroup.GetAssociatedVms`,
  `NsxtFirewallGroup.IsSecurityGroup`, `NsxtFirewallGroup.IsIpSet`
  [GH-368]
* Added methods Org.QueryVmList and Org.QueryVmById to find VM by ID in an Org
  [GH-368]

