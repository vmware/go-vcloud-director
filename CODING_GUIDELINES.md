# Coding guidelines


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
(See [LOGGING.md](https://github.com/vmware/go-vcloud-director/blob/main/util/LOGGING.md)) recording all the requests and responses
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

## Implementing functions to support different vCD API versions

Functions dealing with different versions should use a matrix structure to identify which calls to run according to the 
highest API version supported by vCD. An example can be found in adminvdc.go.

Note: use this pattern for adding new vCD functionality, which is not available in the earliest API version supported 
by the code base (as indicated by `Client.APIVersion`).

```
type vdcVersionedFunc struct {
	SupportedVersion string
	CreateVdc        func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (*Vdc, error)
	CreateVdcAsync   func(adminOrg *AdminOrg, vdcConfiguration *types.VdcConfiguration) (Task, error)
	UpdateVdc        func(adminVdc *AdminVdc) (*AdminVdc, error)
	UpdateVdcAsync   func(adminVdc *AdminVdc) (Task, error)
}

var vdcVersionedFuncsV95 = vdcVersionedFuncs{
	SupportedVersion: "31.0",
	CreateVdc:        createVdc,
	CreateVdcAsync:   createVdcAsync,
	UpdateVdc:        updateVdc,
	UpdateVdcAsync:   updateVdcAsync,
}

var vdcVersionedFuncsV97 = vdcVersionedFuncs{
	SupportedVersion: "32.0",
	CreateVdc:        createVdcV97,
	CreateVdcAsync:   createVdcAsyncV97,
	UpdateVdc:        updateVdcV97,
	UpdateVdcAsync:   updateVdcAsyncV97,
}

var vdcVersionedFuncsByVcdVersion = map[string]vdcVersionedFuncs{
	"vdc9.5":  vdcVersionedFuncsV95,
	"vdc9.7":  vdcVersionedFuncsV97,
	"vdc10.0": vdcVersionedFuncsV97
}

func (adminOrg *AdminOrg) CreateOrgVdc(vdcConfiguration *types.VdcConfiguration) (*Vdc, error) {
	apiVersion, err := adminOrg.client.MaxSupportedVersion()
	if err != nil {
		return nil, err
	}
	vdcFunctions, ok := vdcVersionedFuncsByVcdVersion["vdc"+apiVersionToVcdVersion[apiVersion]]
	if !ok {
		return nil, fmt.Errorf("no entity type found %s", "vdc"+apiVersion)
	}
	if vdcFunctions.CreateVdc == nil {
		return nil, fmt.Errorf("function CreateVdc is not defined for %s", "vdc"+apiVersion)
	}
    util.Logger.Printf("[DEBUG] CreateOrgVdc call function for version %s", vdcFunctions.SupportedVersion)
	return vdcFunctions.CreateVdc(adminOrg, vdcConfiguration)
}
```

 ## Query engine
 
The query engine is a search engine that is based on queries (see `query.go`) with additional filters.

The query runs through the function `client.SearchByFilter` (`filter_engine.go`), which requires a `queryType` (string),
and a set of criteria (`*FilterDef`).

We can search by one of the types handled by `queryFieldsOnDemand` (`query_metadata.go`), such as 

```go
const (
	QtVappTemplate      = "vappTemplate"      // vApp template
	QtAdminVappTemplate = "adminVAppTemplate" // vApp template as admin
	QtEdgeGateway       = "edgeGateway"       // edge gateway
	QtOrgVdcNetwork     = "orgVdcNetwork"     // Org VDC network
	QtAdminCatalog      = "adminCatalog"      // catalog
	QtCatalogItem       = "catalogItem"       // catalog item
	QtAdminCatalogItem  = "adminCatalogItem"  // catalog item as admin
	QtAdminMedia        = "adminMedia"        // media item as admin
	QtMedia             = "media"             // media item
)
```
There are two reasons for this limitation:

* If we want to include metadata, we need to add the metadata fields to the list of fields we want the query to fetch.
* Unfortunately, not all fields defined in the corresponding type is accepted by the `fields` parameter in a query.
The fields returned by `queryFieldsOnDemand` are the one that have been proven to be accepted.


The `FilterDef` type is defined as follows (`filter_utils.go`)
```go
type FilterDef struct {
	// A collection of filters (with keys from SupportedFilters)
	Filters map[string]string

	// A list of metadata filters
	Metadata []MetadataDef

	// If true, the query will include metadata fields and search for exact values.
	// Otherwise, the engine will collect metadata fields and search by regexp
	UseMetadataApiFilter bool
}
```

A `FilterDef` may contain several filters, such as:

```go
criteria := &govcd.FilterDef{
    Filters:  {
        "name":   "^Centos",
        "date":   "> 2020-02-02",
        "latest": "true",
    },
    Metadata: {
        {
            Key:      "dept",
            Type:     "STRING",
            Value:    "ST\\w+",
            IsSystem: false,
        },
    },
    UseMetadataApiFilter: false,
}
```

The set of criteria above will find an item with name starting with "Centos", created after February 2nd, 2020, with
a metadata key "dept" associated with a value starting with "ST". If more than one item is found, the engine will return
the newest one (because of `"latest": "true"`)
The argument `UseMetadataApiFilter`, when true, instructs the engine to run the search with metadata values. Meaning that
the query will contain a clause `filter=metadata:KeyName==TYPE:Value`. If `IsSystem` is true, the clause will become
`filter=metadata@SYSTEM:KeyName==TYPE:Value`. This search can't evaluate regular expressions, because it goes directly 
to vCD.

An example of `SYSTEM` metadata values is the set of annotations that the vCD adds to a vApp template when we save a
vApp to a catalog.

```
  "metadata" = {
    "vapp.origin.id" = "deadbeef-2913-4ed7-b943-79a91620fd52" // vApp ID
    "vapp.origin.name" = "my_vapp_name"
    "vapp.origin.type" = "com.vmware.vcloud.entity.vapp"
  }
```

The engine returns a list of `QueryItem`, and interface that defines several methods used to help evaluate the search
conditions.

### How to use the query engine

Here is an example of how to retrieve a media item.
The criteria ask for the newest item created after the 2nd of February 2020, containing a metadata field named "abc",
with a non-empty value.

```go
            criteria := &govcd.FilterDef{
                Filters:  map[string]string{
                    "date":"> 2020-02-02", 
                    "latest": "true",
                 },
                Metadata: []govcd.MetadataDef{
                    {
                        Key:      "abc",
                        Type:     "STRING",
                        Value:    "\\S+",
                        IsSystem: false,
                    },
                },
                UseMetadataApiFilter: false,
            }
			queryType := govcd.QtMedia
			if vcdClient.Client.IsSysAdmin {
				queryType = govcd.QtAdminMedia
			}
			queryItems, explanation, err := vcdClient.Client.SearchByFilter(queryType, criteria)
			if err != nil {
				return err
			}
			if len(queryItems) == 0 {
				return fmt.Errorf("no media found with given criteria (%s)", explanation)
			}
			if len(queryItems) > 1 {
                // deal with several items
				var itemNames = make([]string, len(queryItems))
				for i, item := range queryItems {
					itemNames[i] = item.GetName()
				}
				return fmt.Errorf("more than one media item found by given criteria: %v", itemNames)
			}
            // retrieve the full entity for the item found
			media, err = catalog.GetMediaByHref(queryItems[0].GetHref())
```

The `explanation` returned by `SearchByFilter` contains the details of the criteria as they were understood by the
engine, and the detail of how each comparison with other items was evaluated. This is useful to create meaningful error
messages.

### Supporting a new type in the query engine

To add a type to the search engine, we need the following:

1. Add the type to `types.QueryResultRecordsType` (`types.go`), or, if the type exists, make sure it includes `Metadata`
2. Add the list of supported fields to `queryFieldsOnDemand` (`query_metadata.go`)
3. Implement the interface `QueryItem` (`filter_interface.go`), which requires a type localization (such as 
`type QueryMedia  types.MediaRecordType`)
4. Add a clause to `resultToQueryItems` (`filter_interface.go`)

## Data inspection checkpoints

Logs should not be cluttered with excessive detail.
However, sometimes we need to provide such detail when hunting for bugs.

We can introduce data inspection points, regulated by the environment variable `GOVCD_INSPECT`, which uses a convenient
code to activate the inspection at different points.

For example, we can mark the inspection points in the query engine with labels "QE1", "QE2", etc., in the network creation
they will be "NET1", "NET2", etc, and then activate them using
`GOVCD_INSPECT=QE2,NET1`.

In the code, we use the function `dataInspectionRequested(code)` that will check whether the environment variable contains
the  given code.

## Tenant Context

Tenant context is a mechanism in the VCD API to run calls as a tenant when connected as a system administrator.
It is used, for example, in the UI, to start a session as tenant administrator without having credentials for such a user,
or even when there is no such user yet.
The context change works by adding a header to the API call, containing these fields:

```
X-Vmware-Vcloud-Tenant-Context: [604cf889-b01e-408b-95ae-67b02a0ecf33]
X-Vmware-Vcloud-Auth-Context:   [org-name]
```

The field `X-Vmware-Vcloud-Tenant-Context` contains the bare ID of the organization (it's just the UUID, without the
prefix `urn:vcloud:org:`).
The field `X-Vmware-Vcloud-Auth-Context` contains the organization name.

### tenant context: data availability

From the SDK standpoint, finding the data needed to put together the tenant context is relatively easy when the originator
of the API call is the organization itself (such as `org.GetSomeEntityByName`).
When we deal with objects down the hierarchy, however, things are more difficult. Running a call from a VDC means that
we need to retrieve the parent organization, and extract ID and name. The ID is available through the `Link` structure
of the VDC, but for the name we need to retrieve the organization itself.

The approach taken in the SDK is to save the tenant context (or a pointer to the parent) in the object that we have just
created. For example, when we create a VDC, we save the organization as a pointer in the `parent` field, and the organization 
itself has a field `TenantContext` with the needed information.

Here are the types that are needed for tenant context manipulation
```go

// tenant_context.go
type TenantContext struct {
	OrgId   string // The bare ID (without prefix) of an organization
	OrgName string // The organization name
}

// tenant_context.go
type organization interface {
	orgId() string
	orgName() string
	tenantContext() (*TenantContext, error)
	fullObject() interface{}
}

// org.go
type Org struct {
	Org           *types.Org
	client        *Client
	TenantContext *TenantContext
}

// adminorg.go
type AdminOrg struct {
	AdminOrg      *types.AdminOrg
	client        *Client
	TenantContext *TenantContext
}

// vdc.go
type Vdc struct {
	Vdc    *types.Vdc
	client *Client
	parent organization
}
```

The `organization` type is an abstraction to include both `Org` and `AdminOrg`. Thus, the VDC object has a pointer to its
parent that is only needed to get the tenant context quickly.

Each object has a way to get the tenant context by means of a `entity.getTenantContext()`. The information
trickles down from the hierarchy:

* a VDC gets the tenant context directly from its `parent` field, which has a method `tenantContext()`
* similarly, a Catalog has a `parent` field with the same functionality.
* a vApp will get the tenant context by first retrieving its parent (`vapp.getParentVdc()`) and then asking the parent
for the tenant context.

### tenant context: usage

Once we have the tenant context, we need to pass the information along to the HTTP request that builds the request header,
so that our API call will run in the desired context.

The basic OpenAPI methods (`Client.OpenApiDeleteItem`, `Client.OpenApiGetAllItems`, `Client.OpenApiGetItem`,
`Client.OpenApiPostItem`, `Client.OpenApiPutItem`, `Client.OpenApiPutItemAsync`, `Client.OpenApiPutItemSync`)  all include
a parameter `additionalHeader map[string]string` containing the information needed to build the tenant context header elements.

Inside the function where we want to use tenant context, we do these two steps:

1. retrieve the tenant context
2. add the additional header to the API call.

For example:

```go
func (adminOrg *AdminOrg) GetAllRoles(queryParameters url.Values) ([]*Role, error) {
	tenantContext, err := adminOrg.getTenantContext()
	if err != nil {
		return nil, err
	}
	return getAllRoles(adminOrg.client, queryParameters, getTenantContextHeader(tenantContext))
}
```
The function `getTenantContextHeader` takes a tenant context and returns a map of strings containing the right header
keys. In the example above, the header is passed to `getAllRoles`, which in turn calls `Client.OpenApiGetAllItems`,
which passes the additional header until it reaches `newOpenApiRequest`, where the tenent context data is inserted in
the request header.

When the tenant context is not needed (system administration calls), we just pass `nil` as `additionalHeader`.



## Testing

Every feature in the library must include testing. See [TESTING.md](https://github.com/vmware/go-vcloud-director/blob/main/TESTING.md) for more info.
