//go:build catalog || functional || ALL

/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	. "gopkg.in/check.v1"
)

type subscriptionTestData struct {
	fromOrg                   *AdminOrg
	toOrg                     *AdminOrg
	ovaPath                   string
	mediaPath                 string
	localCopy                 bool
	storageProfile            types.CatalogStorageProfiles
	uploadWhen                string
	preservePublishingCatalog bool
	asynchronousSubscription  bool
}

// Test_SubscribedCatalog tests four scenarios of Catalog subscription
// All cases use a publishing catalog in one Org and a subscribing catalog
// in a different Org.
// The scenarios are a combination of these two facts:
// * whether the subscribing catalog was created before or after the publishing catalog was filled
// * whether the subscribing catalog enabled automatic downloads (localCopy)
//
// To see the inner working of the test components, you may run it as follows:
// $ export GOVCD_TASK_MONITOR=simple_show
// $ go test -tags catalog -check.f Test_SubscribedCatalog -vcd-verbose -check.vv -timeout 0
// When running this way, you will see the tasks originated by the catalogs and the ones started by the catalog items
func (vcd *TestVCD) Test_SubscribedCatalog(check *C) {

	fromOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org)
	check.Assert(err, IsNil)
	toOrg, err := vcd.client.GetAdminOrgByName(vcd.config.VCD.Org + "-1")
	check.Assert(err, IsNil)

	if toOrg.AdminOrg.Vdcs == nil || len(toOrg.AdminOrg.Vdcs.Vdcs) == 0 {
		check.Skip(fmt.Sprintf("receiving org %s does not have any storage", toOrg.AdminOrg.Name))
	}
	// TODO: remove this workaround when support for 10.3.3 is dropped
	// See Test_PublishToExternalOrganizations for details
	fromOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs = true
	fromOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally = true
	_, err = fromOrg.Update()

	check.Assert(err, IsNil)
	vdc, err := fromOrg.GetVDCByName(vcd.config.VCD.Nsxt.Vdc, false)
	check.Assert(err, IsNil)

	storageProfile, err := vdc.FindStorageProfileReference(vcd.config.VCD.StorageProfile.SP1)
	check.Assert(err, IsNil)
	createStorageProfiles := types.CatalogStorageProfiles{VdcStorageProfile: []*types.Reference{&storageProfile}}

	testSubscribedCatalog(subscriptionTestData{
		fromOrg:                  fromOrg,
		toOrg:                    toOrg,
		ovaPath:                  vcd.config.OVA.OvaPath,
		mediaPath:                vcd.config.Media.MediaPath,
		localCopy:                false,
		storageProfile:           createStorageProfiles,
		uploadWhen:               "after_subscription",
		asynchronousSubscription: true,
	}, check)

	testSubscribedCatalog(subscriptionTestData{
		fromOrg:                   fromOrg,
		toOrg:                     toOrg,
		ovaPath:                   vcd.config.OVA.OvaPath,
		mediaPath:                 vcd.config.Media.MediaPath,
		localCopy:                 true,
		storageProfile:            createStorageProfiles,
		uploadWhen:                "after_subscription",
		preservePublishingCatalog: true,
		asynchronousSubscription:  true,
	}, check)

	// For the tests where the items are uploaded before subscription, we can keep the publishing catalog
	// from the previous test
	testSubscribedCatalog(subscriptionTestData{
		fromOrg:                   fromOrg,
		toOrg:                     toOrg,
		ovaPath:                   vcd.config.OVA.OvaPath,
		mediaPath:                 vcd.config.Media.MediaPath,
		localCopy:                 false,
		storageProfile:            createStorageProfiles,
		uploadWhen:                "before_subscription",
		preservePublishingCatalog: true,
	}, check)
	testSubscribedCatalog(subscriptionTestData{
		fromOrg:                   fromOrg,
		toOrg:                     toOrg,
		ovaPath:                   vcd.config.OVA.OvaPath,
		mediaPath:                 vcd.config.Media.MediaPath,
		localCopy:                 true,
		storageProfile:            createStorageProfiles,
		uploadWhen:                "before_subscription",
		preservePublishingCatalog: false, // at the last subtest, we remove the publishing catalog
	}, check)

}

func uploadTestItems(org *AdminOrg, catalogName, templatePath, mediaPath string, numTemplates, numMedia int) error {
	var taskList []*Task

	catalog, err := org.GetCatalogByName(catalogName, true)
	if err != nil {
		return fmt.Errorf("catalog %s not found: %s", catalogName, err)
	}

	for i := 1; i <= numTemplates; i++ {
		templateName := fmt.Sprintf("test-vt-%d", i)
		uploadTask, err := catalog.UploadOvf(templatePath, templateName, "upload from test", 1024)
		if err != nil {
			return err
		}
		taskList = append(taskList, uploadTask.Task)
	}
	for i := 1; i <= numMedia; i++ {
		mediaName := fmt.Sprintf("test_media-%d", i)
		uploadTask, err := catalog.UploadMediaImage(mediaName, "upload from test", mediaPath, 1024)
		if err != nil {
			return err
		}
		taskList = append(taskList, uploadTask.Task)
	}
	_, err = WaitTaskListCompletionMonitor(taskList, testMonitor)
	fmt.Println()
	return err
}

func testSubscribedCatalog(testData subscriptionTestData, check *C) {

	startSubtest := time.Now()
	drawHeader := func(char, msg string) {
		fmt.Println(strings.Repeat(char, 80))
		fmt.Printf("%s %s\n", char, msg)
	}
	drawHeader("*", fmt.Sprintf("START: upload %s - local copy: %v", testData.uploadWhen, testData.localCopy))

	fromOrg := testData.fromOrg
	toOrg := testData.toOrg

	publishingCatalogName := "Publisher"
	subscribingCatalogName := "Subscriber"

	var fromCatalog *AdminCatalog
	var err error
	fromCatalog, err = fromOrg.GetAdminCatalogByName(publishingCatalogName, true)
	if err == nil {
		drawHeader("-", "publishing catalog retrieved from previous test")
	} else {
		drawHeader("-", "creating publishing catalog")
		fromCatalog, err = fromOrg.CreateCatalogWithStorageProfile(publishingCatalogName, "publisher catalog", &testData.storageProfile)
		check.Assert(err, IsNil)
		AddToCleanupList(publishingCatalogName, "catalog", fromOrg.AdminOrg.Name, check.TestName())
	}

	subscriptionPassword := "superUnknown"
	err = fromCatalog.PublishToExternalOrganizations(types.PublishExternalCatalogParams{
		IsPublishedExternally:    takeBoolPointer(true),
		Password:                 subscriptionPassword,
		IsCachedEnabled:          takeBoolPointer(true),
		PreserveIdentityInfoFlag: takeBoolPointer(true),
	})
	check.Assert(err, IsNil)

	uploadItemsIf := func(wanted string) {
		if wanted != testData.uploadWhen {
			return
		}
		howManyTemplates := 3
		howManyMediaItems := 3
		publishedCatalogItems, err := fromCatalog.QueryCatalogItemList()
		if err == nil && len(publishedCatalogItems) == (howManyMediaItems+howManyTemplates) {
			return
		}
		drawHeader("-", fmt.Sprintf("uploading catalog items - %s", wanted))
		err = uploadTestItems(fromOrg, fromCatalog.AdminCatalog.Name, testData.ovaPath, testData.mediaPath, howManyTemplates, howManyMediaItems)
		check.Assert(err, IsNil)
	}
	err = fromCatalog.Refresh()
	check.Assert(err, IsNil)

	check.Assert(fromCatalog.AdminCatalog.PublishExternalCatalogParams, NotNil)
	check.Assert(fromCatalog.AdminCatalog.PublishExternalCatalogParams.CatalogPublishedUrl, Not(Equals), "")

	uploadItemsIf("before_subscription")
	err = fromCatalog.Refresh()
	check.Assert(err, IsNil)

	subscriptionUrl, err := fromCatalog.FullSubscriptionUrl()
	check.Assert(err, IsNil)

	subscriptionParams := types.ExternalCatalogSubscription{
		SubscribeToExternalFeeds: true,
		Location:                 subscriptionUrl,
		Password:                 subscriptionPassword,
		LocalCopy:                testData.localCopy,
	}

	var toCatalog *AdminCatalog
	testSubscribedCatalogWithInvalidParameters(toOrg, subscriptionParams, subscribingCatalogName, subscriptionPassword, testData.localCopy, check)
	if testData.asynchronousSubscription {
		drawHeader("-", "creating subscribed catalog asynchronously")
		// With asynchronous subscription the catalog starts the subscription but does not report its state, which is
		// monitored by its internal Task
		toCatalog, err = toOrg.CreateCatalogFromSubscriptionAsync(
			subscriptionParams,     // params
			nil,                    // storage profile
			subscribingCatalogName, // catalog name
			subscriptionPassword,   // password
			testData.localCopy)     // local copy
	} else {
		drawHeader("-", "creating subscribed catalog and waiting for completion")
		toCatalog, err = toOrg.CreateCatalogFromSubscription(
			subscriptionParams,     // params
			nil,                    // storage profile
			subscribingCatalogName, // catalog name
			subscriptionPassword,   // password
			testData.localCopy,     // local copy
			10*time.Minute)         // timeout
	}
	check.Assert(err, IsNil)
	AddToCleanupList(subscribingCatalogName, "catalog", toOrg.AdminOrg.Name, check.TestName())

	if testData.asynchronousSubscription {
		err = toCatalog.Refresh()
		check.Assert(err, IsNil)
		if ResourceInProgress(toCatalog.AdminCatalog.Tasks) {
			fmt.Println("catalog subscription tasks still in progress")
			for _, task := range toCatalog.AdminCatalog.Tasks.Task {
				testMonitor(task)
			}
		} else {
			fmt.Println("catalog subscription tasks complete")
		}
	}

	uploadItemsIf("after_subscription")

	// If the catalog items were uploaded before the catalog subscription, we don't need to
	// synchronise, as the subscription would have got at least the list of items
	if testData.uploadWhen != "before_subscription" {
		drawHeader("-", "synchronising catalog")
		err = toCatalog.Sync()
		check.Assert(err, IsNil)
	}

	publishedCatalogItems, err := fromCatalog.QueryCatalogItemList()
	check.Assert(err, IsNil)
	subscribedCatalogItems, err := toCatalog.QueryCatalogItemList()
	check.Assert(err, IsNil)
	fmt.Printf("Catalog items after catalog sync: %d\n", len(subscribedCatalogItems))
	publishedVappTemplates, err := fromCatalog.QueryVappTemplateList()
	check.Assert(err, IsNil)
	subscribedVappTemplates, err := toCatalog.QueryVappTemplateList()
	check.Assert(err, IsNil)
	publishedMediaItems, err := fromCatalog.QueryMediaList()
	check.Assert(err, IsNil)
	subscribedMediaItems, err := toCatalog.QueryMediaList()
	check.Assert(err, IsNil)

	fmt.Printf("vApp template after catalog sync %d\n", len(subscribedVappTemplates))
	fmt.Printf("media item after catalog sync %d\n", len(subscribedMediaItems))

	check.Assert(len(subscribedCatalogItems), Equals, len(publishedCatalogItems))
	check.Assert(len(subscribedVappTemplates), Equals, len(publishedVappTemplates))
	check.Assert(len(subscribedMediaItems), Equals, len(publishedMediaItems))

	if testData.localCopy && testData.uploadWhen == "before_subscription" {
		// we should have all the contents here if the data was available early
		// and the subscribed catalog uses automatic download
		retrieveCatalogItems(toCatalog, subscribedCatalogItems, check)
	}

	// Synchronising all vApp templates and media items. If the subscription includes local copy,
	// the synchronisation has alredy happened, and this extra call is very quick (~5 seconds)
	drawHeader("-", "synchronising vApp templates and media items")
	tasksVappTemplates, err := toCatalog.LaunchSynchronisationAllVappTemplates()
	check.Assert(err, IsNil)
	tasksMediaItems, err := toCatalog.LaunchSynchronisationAllMediaItems()
	check.Assert(err, IsNil)

	// Wait for all synchronisation tasks to end
	var allTasks []*Task
	allTasks = append(allTasks, tasksVappTemplates...)
	allTasks = append(allTasks, tasksMediaItems...)
	_, err = WaitTaskListCompletionMonitor(allTasks, testMonitor)
	if !testVerbose {
		fmt.Println()
	}
	check.Assert(err, IsNil)

	// after a full synchronisation, all data should be available	under every condition
	retrieveCatalogItems(toCatalog, subscribedCatalogItems, check)

	startDelete := time.Now()
	err = toCatalog.Delete(true, true)
	check.Assert(err, IsNil)
	fmt.Printf("subscribed catalog deletion done in %s\n", time.Since(startDelete))
	startDelete = time.Now()
	if !testData.preservePublishingCatalog {
		err = fromCatalog.Delete(true, true)
		check.Assert(err, IsNil)
		fmt.Printf("published catalog deletion done in %s\n", time.Since(startDelete))
	}
	drawHeader("=", fmt.Sprintf("END: upload %s - local copy: %v - Time taken: %s", testData.uploadWhen, testData.localCopy, time.Since(startSubtest)))
}

func retrieveCatalogItems(toCatalog *AdminCatalog, subscribed []*types.QueryResultCatalogItemType, check *C) {
	for _, item := range subscribed {
		catalogItem, err := toCatalog.GetCatalogItemByHref(item.HREF)
		check.Assert(err, IsNil)
		switch catalogItem.CatalogItem.Entity.Type {
		case types.MimeVAppTemplate:
			vAppTemplate, err := catalogItem.GetVAppTemplate()
			check.Assert(err, IsNil)
			check.Assert(vAppTemplate.VAppTemplate.HREF, Equals, catalogItem.CatalogItem.Entity.HREF)
		case types.MimeMediaItem:
			mediaItem, err := toCatalog.GetMediaByHref(catalogItem.CatalogItem.Entity.HREF)
			check.Assert(err, IsNil)
			check.Assert(extractUuid(mediaItem.Media.ID), Equals, extractUuid(catalogItem.CatalogItem.Entity.HREF))
		}
	}
}

func testMonitor(task *types.Task) {
	if testVerbose {
		fmt.Printf("task %s - owner %s - operation %s -  status %s - progress %d\n", task.ID, task.Owner.Name, task.Operation, task.Status, task.Progress)
	} else {
		marker := "."
		if task.Status == "success" {
			marker = "+"
		}
		if task.Status == "error" {
			marker = "-"
		}
		fmt.Print(marker)
	}
}

func testSubscribedCatalogWithInvalidParameters(org *AdminOrg, subscription types.ExternalCatalogSubscription,
	name, password string, localCopy bool, check *C) {

	uuid := extractUuid(subscription.Location)
	params := subscription
	params.Location = strings.Replace(params.Location, uuid, "deadbeef-d72f-4a21-a4d2-4dc9e0b36555", 1)
	// Use a valid host with invalid UUID
	_, err := org.CreateCatalogFromSubscriptionAsync(params, nil, name, password, localCopy)
	check.Assert(err, ErrorMatches, ".*RESOURCE_NOT_FOUND.*")

	newUrl, err := url.Parse(subscription.Location)
	check.Assert(err, IsNil)

	params = subscription
	params.Location = strings.Replace(params.Location, newUrl.Host, "fake.example.com", 1)
	// use an invalid host
	_, err = org.CreateCatalogFromSubscriptionAsync(params, nil, name, password, localCopy)
	check.Assert(err, ErrorMatches, ".*INVALID_URL_OR_PASSWORD.*")

	params = subscription
	params.Location = "not-an-URL"
	// use an invalid URL
	_, err = org.CreateCatalogFromSubscriptionAsync(params, nil, name, password, localCopy)
	check.Assert(err, ErrorMatches, ".*UNKNOWN_ERROR.*")
}
