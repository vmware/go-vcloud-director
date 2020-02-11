# Coding guidelines


( **Work in progress** )


## Principles

The functions, entities, and methods in this library have the wide goal of providing access to vCD functionality
using Go clients.
A more focused goal is to support the [Terraform Provider for vCD](https://github.com/terraform-providers/terraform-provider-vcd).
When in doubt about the direction of development, we should facilitate the path towards making the code usable and maintainable
in the above project.


## Create new entities

A new entity must have its type defined in `types/56/types.go`. If the type is not already there, it should be 
added using the [vCD API](https://code.vmware.com/apis/72/vcloud-director), and possibly reusing components already defined
in `types.go`.

The new entity should have a structure in `entity.go` as

```go
type Entity struct {
	Entity *types.Entity
	client *VCDClient
	// Optional, in some cases: Parent *Parent
}
```

The entity should have at least the following:

```
(parent *Parent) CreateEntityAsync(input *types.Entity) (Task, error)
(parent *Parent) CreateEntity(input *types.Entity) (*Entity, error)
```

The second form will invoke the `*Async` method, run task.WaitCompletion(), and then retrieving the new entity
from the parent and returning it.

If the API does not provide a task, the second method will be sufficient.

If the structure is exceedingly complex, we can use two approaches:

1. if the parameters needed to create the entity are less than 4, we can pass them as argument

```go
(parent *Parent) CreateEntityAsync(field1, field2 string, field3 bool) (Task, error)
```

2. If there are too many parameters to pass, we can create a simplified structure:

```go
type EntityInput struct {
	field1 string
	field2 string
	field3 bool
	field4 bool
	field5 int
	field6 string
	field7 []string
}

(parent *Parent) CreateEntityAsync(simple EntityInput) (Task, error)
```

The latter approach should be preferred when the simplified structure would be a one-to-one match with the corresponding
resource in Terraform.

## Calling the API

Calls to the vCD API should not be sent directly, but using one of the following functions from `api.go:

```go
// Helper function creates request, runs it, check responses and parses out interface from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// out - structure to be used for unmarshalling xml
// E.g. 	unmarshalledAdminOrg := &types.AdminOrg{}
// client.ExecuteRequest(adminOrg.AdminOrg.HREF, http.MethodGet, "", "error refreshing organization: %s", nil, unmarshalledAdminOrg)
func (client *Client) ExecuteRequest(pathURL, requestType, contentType, errorMessage string, payload, out interface{}) (*http.Response, error)
```

```go
// Helper function creates request, runs it, checks response and parses task from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// E.g. client.ExecuteTaskRequest(updateDiskLink.HREF, http.MethodPut, updateDiskLink.Type, "error updating disk: %s", xmlPayload)
func (client *Client) ExecuteTaskRequest(pathURL, requestType, contentType, errorMessage string, payload interface{}) (Task, error) 
```

```go
// Helper function creates request, runs it, checks response and do not expect any values from it.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// E.g. client.ExecuteRequestWithoutResponse(catalogItemHREF.String(), http.MethodDelete, "", "error deleting Catalog item: %s", nil)
func (client *Client) ExecuteRequestWithoutResponse(pathURL, requestType, contentType, errorMessage string, payload interface{}) error 
```

```go
// ExecuteRequestWithCustomError sends the request and checks for 2xx response. If the returned status code
// was not as expected - the returned error will be unmarshaled to `errType` which implements Go's standard `error`
// interface.
func (client *Client) ExecuteRequestWithCustomError(pathURL, requestType, contentType, errorMessage string,
	payload interface{}, errType error) (*http.Response, error) 
```

In addition to saving code and time by reducing the boilerplate, these functions also trigger debugging calls that make the code 
easier to monitor.
Using any of the above calls will result in the standard log i
(See [LOGGING.md](https://github.com/vmware/go-vcloud-director/blob/master/util/LOGGING.md)) recording all the requests and responses
on demand, and also triggering debug output for specific calls (see `enableDebugShowRequest` and `enableDebugShowResponse`
and the corresponding `disable*` in `api.go`).


## Implementing search methods

Each entity should have the following methods:

```
// OPTIONAL
(parent *Parent) GetEntityByHref(href string) (*Entity, error)

// ALWAYS
(parent *Parent) GetEntityByName(name string) (*Entity, error)
(parent *Parent) GetEntityById(id string) (*Entity, error)
(parent *Parent) GetEntityByNameOrId(identifier string) (*Entity, error)
```

For example, the parent for `Vdc` is `Org`, the parent for `EdgeGateway` is `Vdc`.
If the entity is at the top level (such as `Org`, `ExternalNetwork`), the parent is `VCDClient`.

These methods return a pointer to the entity's structure and a nil error when the search was successful,
a nil pointer and an error in every other case.
When the method can establish that the entity was not found because it did not appear in the
parent's list of entities, the method will return `ErrorEntityNotFound`.
In no cases we return a nil error when the method fails to find the entity.
The "ALWAYS" methods can optionally add a Boolean `refresh` argument, signifying that the parent should be refreshed
prior to attempting a search.

Note: We are in the process of replacing methods that don't adhere to the above principles (for example, return a
structure instead of a pointer, return a nil error on not-found, etc).

## Implementing functions to support different API versions

Functions dealing with different versions should use a matrix structure to identify which calls to run according to the 
highest API version supported by vCD. An example can be found in adminvdc.go.

```
type vdcVersionedFunc struct {
	SupportedVersion string
	CreateVdc        func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error)
	CreateVdcAsync   func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error)
	UpdateVdc        func(adminVdc *AdminVdc) (*AdminVdc, error)
	UpdateVdcAsync   func(adminVdc *AdminVdc) (Task, error)
}

var vdcVersionedFuncV90 = vdcVersionedFunc{
	SupportedVersion: "29.0",
	CreateVdc:        createVdc,
	CreateVdcAsync:   createVdcAsync,
	UpdateVdc:        updateVdc,
	UpdateVdcAsync:   updateVdcAsync,
}

var vdcVersionedFuncV97 = vdcVersionedFunc{
	SupportedVersion: "32.0",
	CreateVdc:        createVdcV97,
	CreateVdcAsync:   createVdcAsyncV97,
	UpdateVdc:        updateVdcV97,
	UpdateVdcAsync:   updateVdcAsyncV97,
}

var vdcVersionedFuncByVcdVersion = map[string]vdcVersionedFunc{
	"vdc9.0":  vdcCrudV90,
	"vdc9.1":  vdcCrudV90,
	"vdc9.5":  vdcCrudV90,
	"vdc9.7":  vdcCrudV97,
	"vdc10.0": vdcCrudV97,
}

func (adminOrg *AdminOrg) CreateOrgVdc(vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	apiVersion, err := adminOrg.client.maxSupportedVersion()
	if err != nil {
		return nil, err
	}
	producer, ok := vdcVersionedFuncByVcdVersion["vdc"+apiVersionToVcdVersion[apiVersion]]
	if !ok {
		return nil, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if producer.CreateVdc == nil {
		return nil, fmt.Errorf("function CreateVdc is not defined for %s", "vdc"+apiVersion)
	}
    util.Logger.Printf("[DEBUG] CreateOrgVdc call function for version %s", producer.SupportedVersion)
	return producer.CreateVdc(adminOrg, vdcConfiguration)
}
```

 

## Testing

Every feature in the library must include testing. See [TESTING.md](https://github.com/vmware/go-vcloud-director/blob/master/TESTING.md) for more info.
