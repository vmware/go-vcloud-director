/*
* @Author: frapposelli
* @Date:   2014-10-21 17:47:56
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:10:56
 */

package govcloudair

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestOrg(t *testing.T) {
	TestingT(t)
}

func (s *S) Test_RetrieveOrg(c *C) {
	testServer.Response(200, nil, orgExample)

	org, err := s.client.RetrieveOrg("11111111-1111-1111-1111-111111111111")

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(org.Description, Equals, "")
	c.Assert(org.FullName, Equals, "OrganizationName")

}

func (s *S) Test_FindOrgCatalogId(c *C) {
	testServer.Response(200, nil, orgExample)

	org, err := s.client.RetrieveOrg("11111111-1111-1111-1111-111111111111")
	net := org.FindOrgCatalogId("Public Catalog")

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(net, Equals, "e8a20fdf-8a78-440c-ac71-0420db59f854")

}

func (s *S) Test_FindOrgVDCId(c *C) {
	testServer.Response(200, nil, orgExample)

	org, err := s.client.RetrieveOrg("11111111-1111-1111-1111-111111111111")
	net := org.FindOrgVDCId("M916272752-5793")

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(net, Equals, "214cd6b2-3f7a-4ee5-9b0a-52b4001a4a84")

}

func (s *S) Test_FindOrgNetworkId(c *C) {
	testServer.Response(200, nil, orgExample)

	org, err := s.client.RetrieveOrg("11111111-1111-1111-1111-111111111111")
	net := org.FindOrgNetworkId("M916272752-5793-default-isolated")

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(net, Equals, "8d0cbfe2-25b3-4a1f-b608-5ffeabc7a53d")

}

var orgExample = `
	<?xml version="1.0" ?>
	<Org href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57" id="urn:vcloud:org:23bd2339-c55f-403c-baf3-13109e8c8d57" name="M916272752-5793" type="application/vnd.vmware.vcloud.org+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/vdc/214cd6b2-3f7a-4ee5-9b0a-52b4001a4a84" name="M916272752-5793" rel="down" type="application/vnd.vmware.vcloud.vdc+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/tasksList/23bd2339-c55f-403c-baf3-13109e8c8d57" rel="down" type="application/vnd.vmware.vcloud.tasksList+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854" name="Public Catalog" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/catalog/5cb6451b-8091-4c89-930d-1ff9653cb12d" name="PSE" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalog/5cb6451b-8091-4c89-930d-1ff9653cb12d/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalog/5cb6451b-8091-4c89-930d-1ff9653cb12d/action/controlAccess" rel="controlAccess" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/catalog/8715ed97-348a-4bdd-aea3-e13e84213e6f" name="GCoE" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalog/8715ed97-348a-4bdd-aea3-e13e84213e6f/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalog/8715ed97-348a-4bdd-aea3-e13e84213e6f/action/controlAccess" rel="controlAccess" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/catalog/92d4ad5a-217e-4bd1-8bcf-be9cc70dcfa6" name="Vagrant" rel="down" type="application/vnd.vmware.vcloud.catalog+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalog/92d4ad5a-217e-4bd1-8bcf-be9cc70dcfa6/controlAccess/" rel="down" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalog/92d4ad5a-217e-4bd1-8bcf-be9cc70dcfa6/action/controlAccess" rel="controlAccess" type="application/vnd.vmware.vcloud.controlAccess+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/admin/org/23bd2339-c55f-403c-baf3-13109e8c8d57/catalogs" rel="add" type="application/vnd.vmware.admin.catalog+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/network/8d0cbfe2-25b3-4a1f-b608-5ffeabc7a53d" name="M916272752-5793-default-isolated" rel="down" type="application/vnd.vmware.vcloud.orgNetwork+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/network/cb0f4c9e-1a46-49d4-9fcb-d228000a6bc1" name="M916272752-5793-default-routed" rel="down" type="application/vnd.vmware.vcloud.orgNetwork+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/supportedSystemsInfo/" rel="down" type="application/vnd.vmware.vcloud.supportedSystemsInfo+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/metadata" rel="down" type="application/vnd.vmware.vcloud.metadata+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57/hybrid" rel="down" type="application/vnd.vmware.vcloud.hybridOrg+xml"/>
		<Description/>
		<FullName>OrganizationName</FullName>
	</Org>
	`
