#!/usr/bin/env bash
# This script will connect to the vCD using username and password,
# and show the headers that contain a bearer or authorization token.
#
user=$1
password=$2
org=$3
IP=$4
api_version=$5

if [ -z "$IP" ]
then
    echo "Syntax $0 user password organization hostname_or_IP_address [API version]"
    exit 1
fi

auth=$(echo -n "$user@$org:$password" |base64)

[ -z "$api_version" ] && api_version=32
operation=api/sessions

# if the requested version is greater than 32 (VCD 10.0+), we can use cloudapi
if [[ $api_version -ge 33  ]]
then
    # endpoint for system administrator
    operation=cloudapi/1.0.0/sessions/provider
    if [ "$org" != "System" -a "$org" != "system" ]
    then
        # endpoint for org users
        operation=cloudapi/1.0.0/sessions
    fi
fi

set -x
curl -I -k --header "Accept: application/*;version=${api_version}.0" \
    --header "Authorization: Basic $auth" \
    --request POST https://$IP/$operation

# If successful, the output of this command will include lines like the following
# X-VCLOUD-AUTHORIZATION: 08a321735de84f1d9ec80c3b3e18fa8b
# X-VMWARE-VCLOUD-ACCESS-TOKEN: eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbmlzdHJhdG9yI[562 more characters]
#
# The string after `X-VCLOUD-AUTHORIZATION:` is the old (deprecated) token.
# The 612-character string after `X-VMWARE-VCLOUD-ACCESS-TOKEN` is the bearer token
#
# Note that using cloudapi we will only get the bearer token
