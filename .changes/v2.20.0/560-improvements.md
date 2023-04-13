* Add new field `TransparentModeEnabled` to `types.NsxtAlbVirtualService` which allows to preserve
  client IP for NSX-T ALB Virtual Service (VCD 10.4.1+) [GH-560]
* Add new field `MemberGroupRef` to `types.NsxtAlbPool` which allows to define NSX-T ALB Pool
  membership by using Edge Firewall Group (`NsxtFirewallGroup`) instead of plain IPs (VCD 10.4.1+)
  [GH-560]
