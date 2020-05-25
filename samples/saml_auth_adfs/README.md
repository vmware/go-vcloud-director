# SAML authentication with ADFS as IdP
This is an example how to use Active Directory Federation Services as SAML IdP for vCD.
`main()` function has an example how to setup vCD client with SAML auth. On successful login it will
list Edge Gateways.
To run this command please supply parameters as per below example:
```
go build -o auth
./auth --username test@test-forest.net --password my-password --org my-org --endpoint https://_YOUR_HOSTNAME_/api
```

Results should look similar to:
```
Found 1 Edge Gateways
my-edge-gw
```


## More details
Main trick for making SAML with ADFS work is to use configuration option function
`WithSamlAdfs(useSaml bool, customAdfsRptId string)` in `govcd.NewVCDClient()`.
At the moment ADFS WS-TRUST endpoint "/adfs/services/trust/13/usernamemixed" is the only one
supported and it must be enabled on ADFS server to work properly.

## Troubleshooting
Environment variable `GOVCD_LOG=1` can be used to enable API call logging. It should log all API
calls (including the ones to ADFS server) with obfuscated credentials to aid troubleshooting.