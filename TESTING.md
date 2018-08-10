# Testing in go-vcloud-director
To run tests in go-vcloud-director, users must use a yaml file specifying information about the users vcd. Users can set the VCLOUD_CONFIG environmental variable with the path.

```
export VCLOUD_CONFIG = $HOME/test.yaml
```

If no environmental variable is set it will default to $HOME/config.yaml.


# Example Config file

```
provider:
  user: root
  password: root
  url:  https://api.vcd.api/api

vcd:
  org: org
  vdc: org-vdc
  catalog:
    name: test
    description: test catalog
    catalogitem: ubuntu
  vapp: vapp
  storageprofile: Development
  network: net

```

Users must specify their username, password, api_endpoint, vcd and org for any tests to run. Otherwise all tests get aborted. For more comprehensive testing the catalog, vapp, storageprofile, and network field can be set using the format above. For comprehensive testing just replace each field with your vcd information. 

# Running Tests
Once you have a config file setup, you can run tests with either the makefile or with go itself.

To run tests with go use these commands:
```
cd govcd
go test -v 
```

To run tests with the makefile:
```
make test
```

# Final Words
Have fun using our SDK!! 
