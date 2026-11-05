//go:build tm || functional || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"net/url"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_TmNamespaceInventory(check *C) {
	requireTmTestVersion(vcd, check, "40.0")
	options := tmTestOptions(vcd, check, "40.0")

	namespaces, err := vcd.client.GetAllTmNamespaceSummaries(url.Values{"pageSize": {"1"}}, options)
	check.Assert(err, IsNil)
	if len(namespaces) == 0 {
		check.Skip("no namespaces available")
	}
	names := tmNameCounts(namespaces, func(e *types.TmNamespaceSummary) string { return e.Name })
	for _, summary := range namespaces {
		namespace, err := vcd.client.GetTmNamespaceById(summary.ID, options)
		check.Assert(err, IsNil)
		check.Assert(namespace.TmNamespace.ID, Equals, summary.ID)
		check.Assert(namespace.TmNamespace.Name, Equals, summary.Name)
		check.Assert(namespace.TmNamespace.ProjectAssignment, NotNil)
		if names[summary.Name] == 1 {
			byName, err := vcd.client.GetTmNamespaceByName(summary.Name, options)
			check.Assert(err, IsNil)
			check.Assert(byName.TmNamespace.ID, Equals, summary.ID)
			byNameOrId, err := vcd.client.GetTmNamespaceByNameOrId(summary.Name, options)
			check.Assert(err, IsNil)
			check.Assert(byNameOrId.TmNamespace.ID, Equals, summary.ID)
		}
	}
}

func (vcd *TestVCD) Test_TmNamespaceCapabilities(check *C) {
	requireTmTestVersion(vcd, check, "40.1")
	options := tmTestOptions(vcd, check, "40.1")

	namespaces, err := vcd.client.GetAllTmNamespaceSummaries(nil, options)
	check.Assert(err, IsNil)
	if len(namespaces) == 0 {
		check.Skip("no namespaces available")
	}
	namespace, err := vcd.client.GetTmNamespaceById(namespaces[0].ID, options)
	check.Assert(err, IsNil)
	capabilities, err := namespace.GetAllCapabilities(url.Values{"pageSize": {"1"}})
	check.Assert(err, IsNil)
	for _, capability := range capabilities {
		check.Assert(capability.Name, Not(Equals), "")
	}
	metrics, err := vcd.client.GetAllTmNamespaceStorageClassMetrics(url.Values{"pageSize": {"1"}, "filter": {"id==" + namespaces[0].ID}}, options)
	check.Assert(err, IsNil)
	for _, metric := range metrics {
		check.Assert(metric.ID, Equals, namespaces[0].ID)
		check.Assert(metric.NamespaceStorageClass, NotNil)
	}
}
