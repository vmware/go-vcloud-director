# Open API consumption using low level functions and RAW JSON structures

This example demonstrates how to consume new [OpenAPI](https://vdc-download.vmware.com/vmwb-repository/dcr-public/71f952e6-c14b-417d-8749-dbb5ff2dd48a/9b26a7c0-0cee-40a2-8c01-2f15472324cf/com.vmware.vmware_cloud_director.openapi_34_0.pdf) in VMware Cloud Director. 

OpenAPI low level functions consist of the following to match REST API:
* OpenApiGetAllItems (FIQL filtering can be applied to narrow down results)
* OpenApiPostItem
* OpenApiGetItem
* OpenApiPutItem
* OpenApiDeleteItem
* OpenApiIsSupported
* OpenApiBuildEndpoint

**Note** The endpoint `1.0.0/auditTrail` requires VCD API to support version 33.0 or higher. Version 33.0 was introduced
with VCD 10.0.

## Using mode 1 (Dump raw JSON message as string)
This command will dump JSON for audiTrail endpoint as string allowing to pipe it and process using
external tools like `jq`
```
./openapi --username my_user --password my_secret_password --org my-org --endpoint https://192.168.1.160/api --mode 1 | jq
```

Sample output:
```
[
  {
    "eventId": "urn:vcloud:audit:1df68f82-8e75-4bf9-94ce-c06e1cec66bc",
    "description": "User 'administrator' login failed",
    "operatingOrg": {
      "name": "username-custom",
      "id": "urn:vcloud:org:534625ff-7399-4368-8313-0b75d66bbb6e"
    },
    "user": {
      "name": "administrator",
      "id": "urn:vcloud:user:522d756c-ad00-3b4f-ae6a-6d241437b471"
    },
    "eventEntity": {
      "name": "administrator",
      "id": "urn:vcloud:user:522d756c-ad00-3b4f-ae6a-6d241437b471"
    },
    "taskId": null,
...
```

## Using mode 2 (Define custom struct with JSON tags and access fields)
This mode allows to use OpenAPI in regular Go way (by defining a struct with JSON field tags)

```
./openapi --username my_user --password my_secret_password --org my-org --endpoint https://192.168.1.160/api --mode 2
```

Sample output:

```
Got 30 results
2020-07-19T18:57:34.398+0000 - administrator, -com/vmware/vcloud/event/session/login
2020-07-19T18:58:04.211+0000 - administrator, -com/vmware/vcloud/event/user/create
2020-07-19T18:58:13.904+0000 - dainius, -com/vmware/vcloud/event/session/login
2020-07-19T18:58:14.026+0000 - dainius, -com/vmware/vcloud/event/session/authorize
2020-07-19T18:58:18.256+0000 - dainius, -com/vmware/vcloud/event/session/login
2020-07-19T18:58:18.353+0000 - dainius, -com/vmware/vcloud/event/session/authorize
2020-07-19T18:58:23.841+0000 - dainius, -com/vmware/vcloud/event/session/login
2020-07-19T18:58:23.934+0000 - dainius, -com/vmware/vcloud/event/session/authorize
2020-07-19T19:01:22.513+0000 - dainius, -com/vmware/vcloud/event/session/login
2020-07-19T19:01:22.619+0000 - dainius, -com/vmware/vcloud/event/session/authorize
2020-07-19T19:04:54.392+0000 - dainius, -com/vmware/vcloud/event/session/login
2020-07-19T19:04:54.517+0000 - dainius, -com/vmware/vcloud/event/session/authorize
2020-07-19T19:04:59.648+0000 - dainius, -com/vmware/vcloud/event/session/login
2020-07-19T19:04:59.758+0000 - dainius, -com/vmware/vcloud/event/session/authorize
```

## Troubleshooting
Environment variable `GOVCD_LOG=1` can be used to enable API call logging. It should log all API
calls with obfuscated credentials to aid troubleshooting.
