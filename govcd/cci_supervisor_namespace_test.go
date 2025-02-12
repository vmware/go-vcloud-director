//go:build functional || cci || ALL

package govcd

import (
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	. "gopkg.in/check.v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (vcd *TestVCD) Test_SupervisorNamespace(check *C) {
	skipNonTm(vcd, check)
	vcd.skipIfSysAdmin(check) // The test is running in Org user mode

	regionName := vcd.config.Cci.Region
	vpcName := vcd.config.Cci.Vpc
	storagePolicy := vcd.config.Cci.StoragePolicy
	supervisorZoneName := vcd.config.Cci.SupervisorZone

	cciClient := vcd.client.GetCciClient()

	projectCfg := &ccitypes.Project{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.ProjectKind,
			APIVersion: ccitypes.ProjectCciAPI + "/" + ccitypes.ApiVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-supervisornamespace",
		},
		Spec: ccitypes.ProjectSpec{
			Description: check.TestName(),
		},
	}

	project, err := cciClient.CreateProject(projectCfg)
	check.Assert(err, IsNil)
	check.Assert(project, NotNil)
	AddToCleanupList(projectCfg.Name, "project", vcd.config.Tenants[0].SysOrg, check.TestName())

	defer func() {
		err = project.Delete()
		check.Assert(err, IsNil)
	}()

	supervisorNamespace := &ccitypes.SupervisorNamespace{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.SupervisorNamespaceKind,
			APIVersion: ccitypes.InfrastructureCciAPI + "/" + ccitypes.ApiVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "govcd-test-ns",
			Namespace:    project.Project.GetName(),
		},
		Spec: ccitypes.SupervisorNamespaceSpec{
			ClassName:   "small",
			Description: check.TestName(),
			InitialClassConfigOverrides: ccitypes.SupervisorNamespaceSpecInitialClassConfigOverrides{
				StorageClasses: []ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesStorageClass{
					{
						Name:     storagePolicy,
						LimitMiB: 256,
					},
				},
				Zones: []ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesZone{
					{
						CpuLimitMHz:          200,
						CpuReservationMHz:    1,
						MemoryLimitMiB:       256,
						MemoryReservationMiB: 1,
						Name:                 supervisorZoneName,
					},
				},
			},
			RegionName: regionName,
			VpcName:    vpcName,
		},
	}

	sn, err := cciClient.CreateSupervisorNamespace(project.Project.Name, supervisorNamespace)
	check.Assert(err, IsNil)
	check.Assert(sn, NotNil)
	check.Assert(sn.SupervisorNamespace.Spec, DeepEquals, supervisorNamespace.Spec)
	check.Assert(sn.SupervisorNamespace.Name != "", Equals, true)
	check.Assert(sn.SupervisorNamespace.Name, Not(Equals), sn.SupervisorNamespace.GenerateName)

	AddToCleanupList(sn.SupervisorNamespace.Name, "supervisorNamespace", projectCfg.Name, check.TestName())
	defer func() {
		err = sn.Delete()
		check.Assert(err, IsNil)
	}()
}
