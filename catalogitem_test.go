/*
* @Author: frapposelli
* @Date:   2014-10-22 13:45:30
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:10:55
 */

package govcloudair

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestCatalogItem(t *testing.T) {
	TestingT(t)
}

func (s *S) Test_RetrieveCatalogItem(c *C) {
	testServer.Response(200, nil, catalogitemExample)

	catalogitem, err := s.client.RetrieveCatalogItem("11111111-1111-1111-1111-111111111111")

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(catalogitem.Description, Equals, "id: cts-6.4-32bit")
	c.Assert(catalogitem.Entity.HREF, Equals, "https://p3v2-vcd.vchs.vmware.com:443/api/vAppTemplate/vappTemplate-40cb9721-5f1a-44f9-b5c3-98c5f518c4f5")
	c.Assert(catalogitem.VersionNumber, Equals, 4)

}

func (s *S) Test_FindVappTemplateId(c *C) {
	testServer.Response(200, nil, catalogitemExample)

	catalogitem, err := s.client.RetrieveCatalogItem("11111111-1111-1111-1111-111111111111")
	vappid := catalogitem.FindVappTemplateId()

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(vappid, Equals, "vappTemplate-40cb9721-5f1a-44f9-b5c3-98c5f518c4f5")

}

var catalogitemExample = `
	<?xml version="1.0" ?>
	<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae" id="urn:vcloud:catalogitem:1176e485-8858-4e15-94e5-ae4face605ae" name="CentOS64-32bit" size="0" type="application/vnd.vmware.vcloud.catalogItem+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854" rel="up" type="application/vnd.vmware.vcloud.catalog+xml"/>
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae/metadata" rel="down" type="application/vnd.vmware.vcloud.metadata+xml"/>
		<Description>id: cts-6.4-32bit</Description>
		<Entity href="https://p3v2-vcd.vchs.vmware.com:443/api/vAppTemplate/vappTemplate-40cb9721-5f1a-44f9-b5c3-98c5f518c4f5" name="CentOS64-32bit" type="application/vnd.vmware.vcloud.vAppTemplate+xml"/>
		<DateCreated>2014-06-04T21:06:43.750Z</DateCreated>
		<VersionNumber>4</VersionNumber>
	</CatalogItem>
`
