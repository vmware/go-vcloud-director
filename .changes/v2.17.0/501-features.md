* Added `Client.TestConnection` method to check remote VCD endpoints [GH-447], [GH-501]
* Added `Client.TestConnectionWithDefaults` method that uses `Client.TestConnection` with some
  default values  [GH-447], [GH-501]
* Changed behavior of `Client.OpenApiPostItem` and `Client.OpenApiPostItemSync` so they accept
  response code 200 OK as valid. The reason is `TestConnection` endpoint requires a POST request and
  returns a 200OK when successful  [GH-447], [GH-501]
