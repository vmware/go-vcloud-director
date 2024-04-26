* Fixed an issue that prevented CSE Kubernetes clusters from being upgraded to an OVA with higher Kubernetes version but same TKG version,
  and to an OVA with a higher patch version of Kubernetes [GH-663] 
* Fixed an issue that prevented CSE Kubernetes clusters from being upgraded to TKG v2.5.0 with Kubernetes v1.26.11 as it
  performed an invalid upgrade of CoreDNS [GH-663] 
* Fixed an issue that prevented reading the SSH Public Key from provisioned CSE Kubernetes clusters [GH-663] 
