# Testing in go-vcloud-director
To run tests in go-vcloud-director, users must use a yaml file specifying information about the users vcd. Users can set the `VCLOUD_CONFIG` environmental variable with the path.

```
export GOVCD_CONFIG=your/path/to/test-configuration.yaml
```

If no environmental variable is set it will default to `govcd_test_config.yaml` in the same path where the test files are (`./govcd`.)

## Example Config file

See also `./govcd/sample_govcd_test_config.yaml`.

```yaml
# All items in this file must exist already
# (They will not be removed or left altered)
# The test will create a vApp and remove it at the end
#
provider:
    # vCD administrator credentials
    # (Providing org credentials will skip some tests)
    user: someuser
    password: somepassword
    #
    # The vCD address, in the format https://vCD_IP/api
    # or https://vCD_host_name/api
    url: https://11.111.1.111/api
    #
    # The organization you are authenticating with
    sysOrg: System
vcd:
    # Name of the organization (mandatory)
    org: myorg
    #
    # The virtual data center (mandatory)
    # The tests will create a vApp here
    #
    vdc: myvdc
    # An Org catalog, possibly containing at least one item
    catalog:
        name: mycat
        #
        # One item in the catalog. It will be used to compose test vApps
        catalogItem: myitem
        #
        # An optional description for the catalog. Its test will be skipped if omitted.
        # If provided, it must be the current description of the catalog
        description: mycat for loading
        #
        # An optional description for the catalog item
        catalogItemDescription: myitem to create vapps
    #
    network: mynet
    #
    # Storage profiles used in the vDC
    # One or two can be listed
    storageProfile:
        # First storage profile (mandatory)
        storageProfile1: Development
        # Second storage profile. If omitted, some tests will be skipped.
        storageProfile2: "*"
    # An edge gateway
    # (see https://pubs.vmware.com/vca/topic/com.vmware.vcloud.api.doc_56/GUID-18B0FB8B-385C-4B6D-982C-4B24D271C646.html)
    edgeGateway: myedgegw
    #
    # The IP of the gateway (must exist)
    externalIp: 10.150.10.10
    #
    # An existing IP in the vdc
    internalIp: 192.168.1.10
logging:
    # All items in this section are optional
    # Logging is disabled by default.
    # See ./util/LOGGING.md for more info
    #
    # Enables or disables logs
    enabled: true
    #
    # changes the log name
    logFileName: "go-vcloud-director.log"
    #
    # Defines whether we log the requests in HTTP operations
    logHttpRequests: true
    #
    # Defines whether we log the responses in HTTP operations
    logHttpResponses: true
```

Users must specify their username, password, api_endpoint, vcd and org for any tests to run. Otherwise all tests get aborted. For more comprehensive testing the catalog, catalogitem, storageprofile, network, edgegateway, ip field can be set using the format above. For comprehensive testing just replace each field with your vcd information. 
Note that all the entities included in the configuration file must exist already and will not be removed or left altered during the tests. Leaving a field blank will skip one or more corresponding tests.

If you are more comfortable with JSON, you can supply the configuration in that format. The field names are the same. See `./govcd/sample_govcd_test_config.json`.

## Running Tests
Once you have a config file setup, you can run tests with either the makefile or with go itself.

To run tests with go use these commands:

```bash
cd govcd
go test -check.v .
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
	check.Assert(task.Task.OperationName, checks.Equals, "vdcComposeVapp")
	// Get VApp
	vapp, err := vcd.vdc.FindVAppByName(temp_vapp_name)
	check.Assert(err, checks.IsNil)

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
}
```

# Final Words
Be careful about using our tests as these tests run on a real vcd. If you don't have 1 gb of ram and 2 vcpus available then you should not be running tests that deploy your vm/change memory and cpu. However everything created will be removed at the end of testing.

Have fun using our SDK!!
