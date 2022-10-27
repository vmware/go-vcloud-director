//go:build task || functional || ALL

/*
 * Copyright 2022 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

/**/
import (
	"fmt"
	"time"

	"github.com/kr/pretty"
	. "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_QueryTaskList(check *C) {
	fmt.Printf("Running: %s\n", check.TestName())

	adminOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)

	catalog, err := adminOrg.GetCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	adminCatalog, err := adminOrg.GetAdminCatalogByName(vcd.config.VCD.Catalog.Name, false)
	check.Assert(err, IsNil)
	startQuery := time.Now()
	allTasks, err := vcd.client.Client.QueryTaskList(map[string]string{
		//"orgName": vcd.config.VCD.Org + "-1",
		//"name": "catalogItemSync,catalogItemEnableDownload,catalogItemDelete",
		"status": "running,preRunning,queued",
	})
	check.Assert(err, IsNil)
	fmt.Printf("%# v\n%s\n", pretty.Formatter(allTasks), time.Since(startQuery))

	// search using a client, giving the org and catalog names
	resultByClient, err := vcd.client.Client.QueryTaskList(map[string]string{
		"orgName":    vcd.config.VCD.Org,
		"objectName": adminCatalog.AdminCatalog.Name,
		"name":       "catalogCreateCatalog"})
	check.Assert(err, IsNil)

	// search using an admin catalog, which will search by its HREF
	resultByAdminCatalog, err := adminCatalog.QueryTaskList(map[string]string{
		"name": "catalogCreateCatalog",
	})
	check.Assert(err, IsNil)
	// search using a catalog, which will search by its HREF
	resultByCatalog, err := catalog.QueryTaskList(map[string]string{
		"name": "catalogCreateCatalog",
	})
	check.Assert(err, IsNil)

	check.Assert(len(resultByClient), Equals, len(resultByAdminCatalog))
	check.Assert(len(resultByClient), Equals, len(resultByCatalog))
	if len(resultByAdminCatalog) > 0 {
		// there should be only one task for catalog creation
		check.Assert(len(resultByClient), Equals, 1)
		// check correspondence between task and its related object
		// and also that all sets have returned the same result
		catalogHref, err := adminCatalog.GetCatalogHref()
		check.Assert(err, IsNil)
		check.Assert(resultByAdminCatalog[0].HREF, Equals, resultByClient[0].HREF)
		check.Assert(resultByCatalog[0].HREF, Equals, resultByClient[0].HREF)
		check.Assert(resultByClient[0].Object, Equals, catalogHref)
		check.Assert(resultByAdminCatalog[0].ObjectName, Equals, adminCatalog.AdminCatalog.Name)
		check.Assert(resultByCatalog[0].ObjectName, Equals, adminCatalog.AdminCatalog.Name)
		check.Assert(resultByAdminCatalog[0].ObjectType, Equals, "catalog")
		check.Assert(resultByCatalog[0].ObjectType, Equals, "catalog")
	}
}

/**/
func init() {
	testingTags["task"] = "task_test.go"
}

/*
// Difference between QueryResultTaskRecordType and Task type
// TODO: remove this comment before merging
&types.QueryResultTaskRecordType{
    HREF:             "https://example.com/api/task/c1fd1d6a-30ec-4c44-bc2c-4861f2fe48fb",
    ID:               "",
    Type:             "",
    Org:              "https://example.com/api/org/fa7cd823-ee56-4be9-b57d-78400b6e8fcc",
    OrgName:          "datacloud",
    Name:             "catalogCreateCatalog",
    OperationFull:    "Created Catalog cat-datacloud(65637586-c703-48ae-a7e2-82605d18db57)",
    Message:          "",
    StartDate:        "2022-09-21T07:09:53.271Z",
    EndDate:          "2022-09-21T07:09:54.159Z",
    Status:           "success",
    Progress:         0,
    OwnerName:        "administrator",
    Object:           "https://example.com/api/catalog/65637586-c703-48ae-a7e2-82605d18db57",
    ObjectType:       "catalog",
    ObjectName:       "cat-datacloud",
    ServiceNamespace: "com.vmware.vcloud",
    Link:             (*types.Link)(nil),
    Metadata:         (*types.Metadata)(nil),
}
&types.Task{
    HREF:             "https://example.com/api/task/c1fd1d6a-30ec-4c44-bc2c-4861f2fe48fb",
    Type:             "application/vnd.vmware.vcloud.task+xml",
    ID:               "urn:vcloud:task:c1fd1d6a-30ec-4c44-bc2c-4861f2fe48fb",
    OperationKey:     "",
    Name:             "task",
    Status:           "success",
    Operation:        "Created Catalog cat-datacloud(65637586-c703-48ae-a7e2-82605d18db57)",
    OperationName:    "catalogCreateCatalog",
    ServiceNamespace: "com.vmware.vcloud",
    StartTime:        "2022-09-21T07:09:53.271Z",
    EndTime:          "2022-09-21T07:09:54.159Z",
    ExpiryTime:       "2022-12-20T07:09:53.271Z",
    CancelRequested:  false,
    Link:             &types.Link{HREF:"https://example.com/api/task/c1fd1d6a-30ec-4c44-bc2c-4861f2fe48fb", ID:"", Type:"application/vnd.vmware.vcloud.task+json", Name:"task", Rel:"edit"},
    Description:      "",
    Tasks:            (*types.TasksInProgress)(nil),
    Owner:            &types.Reference{HREF:"https://example.com/api/admin/catalog/65637586-c703-48ae-a7e2-82605d18db57", ID:"urn:vcloud:catalog:65637586-c703-48ae-a7e2-82605d18db57", Type:"application/vnd.vmware.admin.catalog+xml", Name:"cat-datacloud"},
    Error:            (*types.Error)(nil),
    User:             &types.Reference{HREF:"https://example.com/api/admin/user/a707f9c8-c219-424b-9d8e-505afccb33e9", ID:"urn:vcloud:user:a707f9c8-c219-424b-9d8e-505afccb33e9", Type:"application/vnd.vmware.admin.user+xml", Name:"administrator"},
    Organization:     &types.Reference{HREF:"https://example.com/api/org/fa7cd823-ee56-4be9-b57d-78400b6e8fcc", ID:"urn:vcloud:org:fa7cd823-ee56-4be9-b57d-78400b6e8fcc", Type:"application/vnd.vmware.vcloud.org+xml", Name:"datacloud"},
    Progress:         0,
    Details:          "",
}
*/
