//go:build tm || functional || ALL

package govcd

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
	. "gopkg.in/check.v1"
)

// vCenter task is sometimes unreliable and trying to refresh it immediately after it becomes
// connected causes a "BUSY_ENTITY" error (which has a few different messages)
var maximumVcenterRetryTime = 120 * time.Second                                         // The maximum time a single operation will be retried before giving up
var vCenterEntityBusyRegexp = regexp.MustCompile(`(is currently busy|400|BUSY_ENTITY)`) // Regexp to match entity busy error

func (vcd *TestVCD) Test_VCenter(check *C) {
	skipNonTm(vcd, check)
	sysadminOnly(vcd, check)

	if !vcd.config.Tm.CreateVcenter {
		check.Skip("Skipping vCenter creation")
	}

	cfg := &types.VSphereVirtualCenter{
		Name:      check.TestName() + "-vc",
		Username:  vcd.config.Tm.VcenterUsername,
		Password:  vcd.config.Tm.VcenterPassword,
		Url:       vcd.config.Tm.VcenterUrl,
		IsEnabled: true,
	}

	// Certificate must be trusted before adding vCenter
	url, err := url.Parse(cfg.Url)
	check.Assert(err, IsNil)
	trustedCert, err := vcd.client.AutoTrustCertificate(url)
	check.Assert(err, IsNil)
	if trustedCert != nil {
		AddToCleanupListOpenApi(trustedCert.TrustedCertificate.ID, check.TestName()+"trusted-cert", types.OpenApiPathVersion1_0_0+types.OpenApiEndpointTrustedCertificates+trustedCert.TrustedCertificate.ID)
	}

	v, err := vcd.client.CreateVcenter(cfg)
	check.Assert(err, IsNil)
	check.Assert(v, NotNil)

	err = v.Refresh()
	check.Assert(err, IsNil)

	// Add to cleanup list
	PrependToCleanupListOpenApi(v.VSphereVCenter.VcId, check.TestName(), types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVirtualCenters+v.VSphereVCenter.VcId)

	printVerbose("# Waiting for listener status to become 'CONNECTED'\n")
	err = waitForListenerStatusConnected(v)
	check.Assert(err, IsNil)

	// Sometimes the refresh fails with one of 'vCenterEntityBusyRegexp' errors
	err = runWithRetry(v.RefreshVcenter, vCenterEntityBusyRegexp, maximumVcenterRetryTime)
	check.Assert(err, IsNil)

	// Refresh storage policies
	err = runWithRetry(v.RefreshStorageProfiles, vCenterEntityBusyRegexp, maximumVcenterRetryTime)
	check.Assert(err, IsNil)

	// Get By Name
	byName, err := vcd.client.GetVCenterByName(cfg.Name)
	check.Assert(err, IsNil)
	check.Assert(byName, NotNil)

	// Get By ID
	byId, err := vcd.client.GetVCenterById(v.VSphereVCenter.VcId)
	check.Assert(err, IsNil)
	check.Assert(byId, NotNil)

	// TODO: TM: URLs should be the same from
	// check.Assert(byName.VSphereVCenter.Url, Equals, byId.VSphereVCenter.Url)

	// Get All
	allTmOrgs, err := vcd.client.GetAllVCenters(nil)
	check.Assert(err, IsNil)
	check.Assert(allTmOrgs, NotNil)
	check.Assert(len(allTmOrgs) > 0, Equals, true)

	// Update
	v.VSphereVCenter.IsEnabled = false
	updated, err := v.Update(v.VSphereVCenter)
	check.Assert(err, IsNil)
	check.Assert(updated, NotNil)

	// Delete
	err = v.Delete()
	check.Assert(err, IsNil)

	notFoundByName, err := vcd.client.GetVCenterByName(cfg.Name)
	check.Assert(ContainsNotFound(err), Equals, true)
	check.Assert(notFoundByName, IsNil)

	// Try to create async version
	task, err := vcd.client.CreateVcenterAsync(cfg)
	check.Assert(err, IsNil)
	check.Assert(task, NotNil)
	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	byIdAsync, err := vcd.client.GetVCenterById(task.Task.Owner.ID)
	check.Assert(err, IsNil)
	check.Assert(byIdAsync.VSphereVCenter.Name, Equals, cfg.Name)
	// Add to cleanup list
	PrependToCleanupListOpenApi(byIdAsync.VSphereVCenter.VcId, check.TestName()+"-async", types.OpenApiPathVersion1_0_0+types.OpenApiEndpointVirtualCenters+v.VSphereVCenter.VcId)

	err = byIdAsync.Disable()
	check.Assert(err, IsNil)
	err = byIdAsync.Delete()
	check.Assert(err, IsNil)

	// Remove trusted cert if it was created
	if trustedCert != nil {
		err = trustedCert.Delete()
		check.Assert(err, IsNil)
	}
}

func runWithRetry(runOperation func() error, errRegexp *regexp.Regexp, duration time.Duration) error {
	startTime := time.Now()
	endTime := startTime.Add(duration)
	util.Logger.Printf("[DEBUG] runWithRetry - running with retry for %f seconds if error contains '%s' ", duration.Seconds(), errRegexp)
	count := 1
	for {
		err := runOperation()
		util.Logger.Printf("[DEBUG] runWithRetry - ran attempt %d, got error: %s ", count, err)
		// Operation had no error - it succeeded
		if err == nil {
			util.Logger.Printf("[DEBUG] runWithRetry - no error occurred after attempt %d, got error: %s ", count, err)
			return nil
		}
		// If there is an error, but it doesn't contain the retryIfErrContains value - exit it
		if !errRegexp.MatchString(err.Error()) {
			util.Logger.Printf("[DEBUG] runWithRetry - returning error after attempt %d, got error: %s ", count, err)
			return err
		}

		// If time limit is exceeded - return error containing statistics and original error
		if time.Now().After(endTime) {
			util.Logger.Printf("[DEBUG] runWithRetry - exceeded time after attempt %d, got error: %s ", count, err)
			return fmt.Errorf("error attempting to wait until error does not contain '%s' after %f seconds: %s", errRegexp, duration.Seconds(), err)
		}

		// Sleep and continue
		util.Logger.Printf("[DEBUG] runWithRetry - sleeping after attempt %d, will retry", count)
		// Sleep 2 seconds and attempt once more if the timeout is not excdeeded
		time.Sleep(2 * time.Second)
		count++
	}
}
