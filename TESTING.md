# Testing in go-vcloud-director
To run tests in go-vcloud-director, users must use a yaml file specifying information about the users vcd. Users can set the `GOVCD_CONFIG` environmental variable with the path.

```
export GOVCD_CONFIG=your/path/to/test-configuration.yaml
```

If no environmental variable is set it will default to `govcd_test_config.yaml` in the same path where the test files are (`./govcd`.)

## Example Config file

See `./govcd/sample_govcd_test_config.yaml`.

Users must specify their username, password, API endpoint, vcd and org for any tests to run. Otherwise all tests get stopped. For more comprehensive testing the catalog, catalog item, storage profile, network, edge gateway, IP fields can be set using the format in the sample.
Note that all the entities included in the configuration file must exist already and will not be removed or left altered during the tests. Leaving a field blank will skip one or more corresponding tests.

If you are more comfortable with JSON, you can supply the configuration in that format. The field names are the same. See `./govcd/sample_govcd_test_config.json`.

## Running Tests
Once you have a config file setup, you can run tests with either the makefile or with go itself.

To run tests with go use these commands:

```bash
cd govcd
go test -tags "functional" -check.v .
```

If you want to see more details during the test run, use `-check.vv` instead of `-check.v`.

To run tests with the makefile:

```bash
make test
```

To run a specific test:

```bash
cd govcd
go test -check.f Test_SetOvf -check.vv .
```

The tests can run with several tags that define which components are tested.
Using the Makefile, you can run one of the following:

```bash
make testcatalog
make testnetwork
make testvapp
```

For more options, you can run manually in `./govcd`
When running `go test` without tags, we'll get a list of tags that are available.

```bash
$ go test -v .
=== RUN   TestTags
--- FAIL: TestTags (0.00s)
    api_test.go:59: # No tags were defined
    api_test.go:46:
        # -----------------------------------------------------
        # Tags are required to run the tests
        # -----------------------------------------------------

        At least one of the following tags should be defined:

           * ALL :       Runs all the tests (== functional + unit == all feature tests)

           * functional: Runs all the tests that use check.v1
           * unit:       Runs unit tests that do not need a live vCD

           * catalog:    Runs catalog related tests (also catalog_item, media)
           * disk:       Runs disk related tests
           * extnetwork: Runs external network related tests
           * lb:       	 Runs load balancer related tests
           * network:    Runs network related tests
           * gateway:    Runs edge gateway related tests
           * org:        Runs org related tests
           * query:      Runs query related tests
           * system:     Runs system related tests
           * task:       Runs task related tests
           * user:       Runs user related tests
           * vapp:       Runs vapp related tests
           * vdc:        Runs vdc related tests
           * vm:         Runs vm related tests

        Examples:

        go test -tags functional -check.vv -timeout=45m .
        go test -tags catalog -check.vv -timeout=45m .
        go test -tags "query extnetwork" -check.vv -timeout=45m .
FAIL
FAIL	github.com/vmware/go-vcloud-director/v2/govcd	0.011s
```

To run tests with `concurency` build tag (omitted by default) and Go race detector:

```bash
make testconcurrent
```
__Note__. At the moment they are failing because go-vcloud-director is not thread safe.

## How to write a test

go-vcloud-director tests are written using [check.v1](https://labix.org/gocheck), an auxiliary library for tests that provides several methods to help developers write comprehensive tests.


### Imports

The tests for `govcd` package must belong to the same package, so no need to import govcd.
The mandatory import is

```go
    checks "gopkg.in/check.v1"
```

We refer to this package as `checks`. There are examples online using a dot (".") instead of an explicit name, but please don't do it, as it pollutes the name space and makes code readability harder.

### Test function header

Every function that tests something related to govcd should have this header:

```go
func (vcd *TestVCD) Test_SomeFunctionality(check *checks.C) {
...
}
```

The `vcd` variable is our handler of the govcd functions. Its type is declared within `./govcd/api_vcd_test.go` and contains accessors to the main structures needed for running tests:

```go
type TestVCD struct {
	client *VCDClient
	org    Org
	vdc    Vdc
	vapp   VApp
	config TestConfig
}
```
This entity is initialized when the test starts. You can assume that the variable is ready for usage. If something happens during initialization, it should fail before it comes to your function.

The `check` variable is our interface with the `check.v1` package. We can do several things with it, such as skipping a test, probing a condition, declare success or failure, and more.


### Adding build tags.

All tests need to have a build tag. The tag should be the first line of the file, followed by a blank line

```go
// +build functional featurename ALL

package govcd
```

Tests that integrate in the functional suite use the tag `functional`. Using that tag, we can run all functional tests
at once.
We define as `functional` the tests that need a live vCD to run.

1. The test should always define the `ALL` tag:

* ALL :       Runs all the tests

2. The test should also always define either the `unit` or `functional` tag:

* functional: Runs all the tests that use check.v1
* unit:       Runs unit tests that do not need a live vCD

3. Finally, the test should always define the feature tag. For example:

* catalog:    Runs catalog related tests (also `catalog_item`, `media`)
* disk:       Runs disk related tests

The `ALL` tag includes tests that use a different framework. At the moment, this is useful to run a global compilation test.
Depending on which additional tests we will implement, we may change the dependency on the `ALL` tag if we detect
clashes between frameworks.

If the test file defines a new feature tag (i.e. one that has not been used before) the file should also implement an
`init` function that sets the tag in the global tag list.
This information is used by the main tag test in `api_test.go` to determine which tags were activated.

```go
func init() {
	testingTags["newtag"] = "filename_test.go"
}
```

**VERY IMPORTANT**: if we add a test that runs using a different tag (i.e. it is not included in `functional` tests), we need
to add such test to the Makefile under `make test`. **The general principle is that `make test` runs all tests**. If this can't be
achieved by adding the new test to the `functional` tag (perhaps because we foresee framework conflicts), we need to add the
new test as a separate command.
For example:

```
test: fmtcheck
	@echo "==> Running Tests"
	cd govcd && \
    go test -tags "MyNewTag" -timeout=10m -check.vv . && \ 
	go test -tags "functional" -timeout=60m -check.vv .
``` 

or we can encapsulate a complex test into a self containing script.

```
test: fmtcheck
	@echo "==> Running Tests"
	./scripts/my_complicated_test.sh
	cd govcd && go test -tags "functional" -timeout=60m -check.vv .
``` 

### Basic test function organization.

Within the testing function, you should perform four actions:

1. Run all operations that are needed for the test, such as finding a particular organization, or vDC, or vApp, deploying a VM, etc.
2. Run the operation being tested (such as deploy a VM, retrieve a VM or vApp, etc.)
3. Run the tests, which usually means using `check.Assert` or `check.Check` to compare known data with what was found during the test execution
4. Clean up all entities and data that was created or altered during the test. For example: delete a vApp that was deployed, or remove a setting that was added to a network.

An example:

```go
package govcd

import (
	"fmt"
	checks "gopkg.in/check.v1"
)

func (vcd *TestVCD) Test_GetVAppTemplate(check *checks.C) {

	fmt.Printf("Running: %s\n", check.TestName())

    // #0 Check that the data needed for this test is in the configuration
    if vcd.config.VCD.Catalog.Name == "" {
		check.Skip("Catalog name not provided. Test can't proceed")
    }

    // #1: preliminary data
	cat, err := vcd.org.FindCatalog(vcd.config.VCD.Catalog.Name)
    if err != nil {
		check.Skip("Catalog not found. Test can't proceed")
	}

    // #2: Run the operation being tested
	catitem, err := cat.FindCatalogItem(vcd.config.VCD.Catalog.Catalogitem)
	check.Assert(err, checks.IsNil)

	vapptemplate, err := catitem.GetVAppTemplate()

    // #3: Tests the object contents
	check.Assert(err, checks.IsNil)
	check.Assert(vapptemplate.VAppTemplate.Name, checks.Equals, vcd.config.VCD.Catalog.Catalogitem)
	if vcd.config.VCD.Catalog.Description != "" {
		check.Assert(vapptemplate.VAppTemplate.Description, checks.Equals, vcd.config.VCD.Catalog.CatalogItemDescription)
	}
    // #4 is not needed here, as the test did not modify anything
}
```

When step #4 is also used, we should make sure that the cleanup step is always reached.
Here's another simplified example

```go
func (vcd *TestVCD) Test_ComposeVApp(check *checks.C) {

	fmt.Printf("Running: %s\n", check.TestName())

	// Populate the input data
	// [ ... ]
	// See the full function in govcd/vcd_test.go

	// Run the main operation
	task, err := vcd.vdc.ComposeVApp(networks, vapptemplate, storageprofileref, temp_vapp_name, temp_vapp_description)

	// These tests can fail: we need to make sure that the new entity was created
	check.Assert(err, checks.IsNil)
	check.Assert(task.Task.OperationName, checks.Equals, composeVappName)
	// Get VApp
	vapp, err := vcd.vdc.FindVAppByName(temp_vapp_name)
	check.Assert(err, checks.IsNil)

	// After a successful creation, the entity is added to the cleanup list.
	// If something fails after this point, the entity will be removed
    AddToCleanupList(composeVappName, "vapp", "", "Test_ComposeVApp")

	// Once the operation is successful, we won't trigger a failure
	// until after the vApp deletion
	// Instead of "Assert" we run "Check". If one of the following
	// checks fails, it won't terminate the test, and we can reach the cleanup point.
	check.Check(vapp.VApp.Name, checks.Equals, temp_vapp_name)
	check.Check(vapp.VApp.Description, checks.Equals, temp_vapp_description)

	// [ ... ]
	// More checks follow

	// Here's the cleanup point
	// Deleting VApp
	task, err = vapp.Delete()
	task.WaitTaskCompletion()

	// Here we can fail again.
	check.Assert(err, checks.IsNil)
	no_such_vapp, err := vcd.vdc.FindVAppByName(temp_vapp_name)
	check.Assert(err, checks.NotNil)
	check.Assert(no_such_vapp.VApp, checks.IsNil)
    // If this deletion fails, a further attempt will be made 
    // by the cleanup function at the end of all tests
}
```

# Golden test files

In some tests (especially unit) there is a need for data samples (Golden files). There are a few
helpers in `api_vcd_test_unit.go` - `goldenString` and `goldenBytes`. These helpers are here to
unify data storage naming formats. All files will be stored in `test-resources/golden/`. File name
will be formatted as `t.Name() + "custompart" + ".golden"` (e.g.
"TestSamlAdfsAuthenticate_custompart.golden"). These functions allow to update existing data by
supplying actual data and setting `update=true`. As an example `TestSamlAdfsAuthenticate` test uses
golden data.

# Environment variables and corresponding flags

While running tests, the following environment variables can be used:

* `GOVCD_CONFIG=/path/file`: sets an alternative configuration file for the tests.
   e.g.:  `GOVCD_CONFIG=/some/path/govcd_test_config.yaml go test -tags functional -timeout 0 .`
* `GOVCD_DEBUG=1` (`-vcd-debug`): enable debug output on screen.
* `GOVCD_TEST_VERBOSE=1` (`-vcd-verbose`): shows execution details in some tests.
* `GOVCD_SKIP_VAPP_CREATION=1` (`-vcd-skip-vapp-creation`): will not create the initial vApp. All tests that
   depend on a vApp availability will be skipped.
* `GOVCD_TASK_MONITOR=show|log|simple_show|simple|log`: sets a task monitor function when running `task.WaitCompletion`
    * `show` : displays full task details on screen
    * `log` : writes full task details in the log
    * `simple_show` : displays a summary line for the task on screen
    * `simple_log` : writes a summary line for the task in the log
* `GOVCD_IGNORE_CLEANUP_FILE` (`-vcd-ignore-cleanup-file`): Ignore the cleanup file if it is left behind after a test failure.
    This could be useful after running a single test, when we need to check how the test behaves with the resource still
    in place.
* `GOVCD_SHOW_REQ` (`-vcd-show-request`): shows the API request on standard output
* `GOVCD_SHOW_RESP` (`-vcd-show-response`): shows the API response on standard output
* `VCD_TOKEN` : specifies the authorization token to use instead of username/password
   (Use `./scripts/get_token.sh` to retrieve one)
* `GOVCD_KEEP_TEST_OBJECTS` will skip deletion of objects created during tests.
* `GOVCD_API_VERSION` allows to select the API version to use. This must be used **for testing purposes only** as the SDK
   has been tested to use certain version of the API. Using this environment variable may lead to unexpected failures.
* `GOVCD_SKIP_LOG_TRACING` can disable sending 'X-VMWARE-VCLOUD-CLIENT-REQUEST-ID' header that is
  used for easier log correlation

When both the environment variable and the command line option are possible, the environment variable gets evaluated first.

# SAML auth testing with Active Directory Federation Services (ADFS) as Identity Provider (IdP)

This package supports SAML authentication with ADFS. It can be achieved by supplying
`WithSamlAdfs()` function to `NewVCDClient`. Testing framework also supports SAML auth and there are
a few ways to test it:
* There is a unit test `TestSamlAdfsAuthenticate` which spawns mock servers and does not require
  ADFS or SAML being configured. It tests the flow based on mock endpoints.
* Using regular `user` and `password` to supply SAML credentials and `true` for `useSamlAdfs`
  variable (optionally one can override Relaying Party Trust ID with variable `customAdfsRptId`).
  That way all tests would run using SAML authentication flow.
* Using `samlUser`, `samlPassword` and optionally `samlCustomRptId` variables will enable
  `Test_SamlAdfsAuth` test run. Test_SamlAdfsAuth will test and compare VDC retrieved using main
  authentication credentials vs the one retrieved using specific SAML credentials.

All these tests can run in combination.

# Final Words
Be careful about using our tests as these tests run on a real vcd. If you don't have 1 gb of ram and 2 vcpus available then you should not be running tests that deploy your vm/change memory and cpu. However everything created will be removed at the end of testing.

Have fun using our SDK!!
