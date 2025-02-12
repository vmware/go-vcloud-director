//go:build functional || cci || ALL

package govcd

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_KubeConfig(check *C) {
	skipNonTm(vcd, check)
	vcd.skipIfSysAdmin(check) // The test is running in Org user mode

	cciClient := vcd.client.GetCciClient()
	kubeCfg, kubeConfigValues, err := cciClient.GetKubeConfig(vcd.config.Tenants[0].SysOrg, "", "")
	check.Assert(err, IsNil)
	check.Assert(kubeCfg, NotNil)
	check.Assert(kubeConfigValues, NotNil)

	check.Assert(kubeConfigValues.ContextName, Equals, vcd.config.Tenants[0].SysOrg)
	check.Assert(kubeConfigValues.ClusterName, Equals, fmt.Sprintf("%s:%s", vcd.config.Tenants[0].SysOrg, cciClient.VCDClient.Client.VCDHREF.Host))
	check.Assert(kubeConfigValues.ClusterServer, Equals, fmt.Sprintf("%s://%s/%s", cciClient.VCDClient.Client.VCDHREF.Scheme, cciClient.VCDClient.Client.VCDHREF.Host, "cci/kubernetes"))
	check.Assert(kubeConfigValues.UserName, Equals, fmt.Sprintf("%s:%s@%s", vcd.config.Tenants[0].SysOrg, vcd.config.Tenants[0].User, cciClient.VCDClient.Client.VCDHREF.Host))
	check.Assert(kubeConfigValues.Token, NotNil)

	check.Assert(len(kubeCfg.Clusters) == 1, Equals, true)
	check.Assert(kubeCfg.Clusters[kubeConfigValues.ClusterName].Server, Equals, kubeConfigValues.ClusterServer)
	check.Assert(len(kubeCfg.AuthInfos) == 1, Equals, true)
	check.Assert(kubeCfg.AuthInfos[kubeConfigValues.UserName].Token, Not(Equals), "")
	check.Assert(kubeCfg.Contexts[kubeConfigValues.ContextName].AuthInfo, Equals, kubeConfigValues.UserName)
	check.Assert(kubeCfg.Contexts[kubeConfigValues.ContextName].Cluster, Equals, kubeConfigValues.ClusterName)
}
