* Added method `NsxtEdgeGateway.GetUsedAndUnusedExternalIPAddressCountWithLimit` to count used
  and unused IPs assigned to Edge Gateway. It supports a `limitTo` argument that can prevent
  exhausting system resources when counting IPs in assigned subnets [GH-682]
