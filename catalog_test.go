/*
* @Author: frapposelli
* @Date:   2014-10-22 12:18:30
* @Last Modified by:   frapposelli
* @Last Modified time: 2014-10-23 14:10:54
 */

package govcloudair

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestCatalog(t *testing.T) {
	TestingT(t)
}

func (s *S) Test_RetrieveCatalog(c *C) {
	testServer.Response(200, nil, catalogExample)

	catalog, err := s.client.RetrieveCatalog("11111111-1111-1111-1111-111111111111")

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(catalog.Description, Equals, "vCHS service catalog")
	c.Assert(catalog.IsPublished, Equals, true)
	c.Assert(catalog.VersionNumber, Equals, 60)

}

func (s *S) Test_FindCatalogItemId(c *C) {
	testServer.Response(200, nil, catalogExample)

	catalog, err := s.client.RetrieveCatalog("11111111-1111-1111-1111-111111111111")
	catalogitem := catalog.FindCatalogItemId("CentOS64-32bit")

	_ = testServer.WaitRequest()

	c.Assert(err, IsNil)
	c.Assert(catalogitem, Equals, "1176e485-8858-4e15-94e5-ae4face605ae")

}

var catalogExample = `
	<?xml version="1.0" ?>
	<Catalog href="https://p3v2-vcd.vchs.vmware.com:443/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854" id="urn:vcloud:catalog:e8a20fdf-8a78-440c-ac71-0420db59f854" name="Public Catalog" type="application/vnd.vmware.vcloud.catalog+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
		<Link href="https://p3v2-vcd.vchs.vmware.com:443/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854/metadata" rel="down" type="application/vnd.vmware.vcloud.metadata+xml"/>
		<Description>vCHS service catalog</Description>
		<CatalogItems>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/013d1994-f009-4c40-ac48-517fe7d952a0" id="013d1994-f009-4c40-ac48-517fe7d952a0" name="W2K12-STD-64BIT" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/05384603-e07e-4f00-a95e-776b427f22d9" id="05384603-e07e-4f00-a95e-776b427f22d9" name="W2K12-STD-R2-SQL2K14-WEB" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae" id="1176e485-8858-4e15-94e5-ae4face605ae" name="CentOS64-32bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/1a729040-71b6-412c-bda9-20b9085f9882" id="1a729040-71b6-412c-bda9-20b9085f9882" name="W2K8-STD-R2-64BIT-SQL2K8-STD-R2-SP2" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/222624b5-e62a-4f5b-a2af-b33a4664005e" id="222624b5-e62a-4f5b-a2af-b33a4664005e" name="W2K12-STD-64BIT-SQL2K12-STD-SP1" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/54cb2af1-4439-48fe-85b6-4c9524930ce6" id="54cb2af1-4439-48fe-85b6-4c9524930ce6" name="Ubuntu Server 12.04 LTS (amd64 20140619)" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/693f342b-d872-41d1-983b-fd5cc2c15f7c" id="693f342b-d872-41d1-983b-fd5cc2c15f7c" name="W2K8-STD-R2-64BIT" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/8d4edd11-393f-4cda-ace4-d5b8f1548928" id="8d4edd11-393f-4cda-ace4-d5b8f1548928" name="CentOS64-64bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/bfca201c-e8f3-49f8-a828-397e16fa6cfe" id="bfca201c-e8f3-49f8-a828-397e16fa6cfe" name="W2K12-STD-R2-64BIT" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/cb508cd9-664a-4fec-8eb1-ae5934aad6ad" id="cb508cd9-664a-4fec-8eb1-ae5934aad6ad" name="W2K12-STD-64BIT-SQL2K12-WEB-SP1" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/d0be59f3-ef80-4298-bd4c-f2258a3fec37" id="d0be59f3-ef80-4298-bd4c-f2258a3fec37" name="W2K8-STD-R2-64BIT-SQL2K8-WEB-R2-SP2" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/dbbf4633-64a3-4ac1-b9e0-7f923efa3f13" id="dbbf4633-64a3-4ac1-b9e0-7f923efa3f13" name="Ubuntu Server 12.04 LTS (i386 20140619)" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/ed996ae8-3081-4e16-a7b6-4bed1c462aa4" id="ed996ae8-3081-4e16-a7b6-4bed1c462aa4" name="CentOS63-64bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/f4dc0f92-74ae-413e-8e0f-25e6568a8195" id="f4dc0f92-74ae-413e-8e0f-25e6568a8195" name="W2K12-STD-R2-SQL2K14-STD" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="https://p3v2-vcd.vchs.vmware.com:443/api/catalogItem/ff9c9b63-ca3b-4e39-ab72-7eb9049f8b05" id="ff9c9b63-ca3b-4e39-ab72-7eb9049f8b05" name="CentOS63-32bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
		</CatalogItems>
		<IsPublished>true</IsPublished>
		<DateCreated>2013-10-15T01:14:22.370Z</DateCreated>
		<VersionNumber>60</VersionNumber>
	</Catalog>
`
