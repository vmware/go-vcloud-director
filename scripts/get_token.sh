#!/bin/bash
# This script will connect to the vCD using username and password,
# and show the header that contains an authorization token.
#
user=$1
password=$2
org=$3
IP=$4

if [ -z "$IP" ]
then
    echo "Syntax $0 user password organization hostname_or_IP_address"
    exit 1
fi

auth=$(echo -n "$user@$org:$password" |base64)

curl -I -k --header "Accept: application/*;version=32.0" \
    --header "Authorization: Basic $auth" \
    --request POST https://$IP/api/sessions

# If successful, the output of this command will include lines like the following
# X-VCLOUD-AUTHORIZATION: 08a321735de84f1d9ec80c3b3e18fa8b
# X-VMWARE-VCLOUD-ACCESS-TOKEN: eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbmlzdHJhdG9yIiwiaXNzIjoiYTkzYzlkYjktNzQ3MS0zMTkyLThkMDktYThmN2VlZGE4NWY5QGY5MDZlODE1LTM0NjgtNGQ0ZS04MmJlLTcyYzFjMmVkMTBiMyIsImV4cCI6MTYwNzUxMjgyOCwidmVyc2lvbiI6InZjbG91ZF8xLjAiLCJqdGkiOiJjY2IwZjIwN2JjY2Y0NmYwYmEwNTcyNzgxZDQyNDg2MyJ9.SMjp5wsSd7CXGMdlj-weeCRdr5AazA74pwwx2w3Eqh3RdzyiEMvQfWQAuPAQjM1oOsEUnFOg2u0gYsnIyQg_p7kzXKPQwPNz3BPi0tm2DxxQtQVhOBRXCqUJ9OmRlMVu7FZZ6gKD4GhpbTkZyKMN_IgOFkkt8iXs1-weNZw5TmyVHeWiJdV0JFM45CV47jQNdQMy4OSsU-CqE2VVLOK83oJhRnlnc3O4OAAIfuVZ4SLWqgi1lIoc2vbZv0HYeWO7L_2pGfmja8CVzVhPrgIGEoDhXnvO29z1ToEXRnyMKh9cisiRkhUISpsh4aHRGUUzaZYeOejVX3PAO9aCX3iYWA
#
# The string after `X-VCLOUD-AUTHORIZATION:` is the old (deprecated) token.
# The string after `X-VMWARE-VCLOUD-ACCESS-TOKEN` is the bearer token

