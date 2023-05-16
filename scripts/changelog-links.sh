#!/usr/bin/env bash

# This script rewrites [GH-nnnn]-style references in the CHANGELOG.md file to
# be Markdown links to the given github issues.
#
# This is run during releases so that the issue references in all of the
# released items are presented as clickable links, but we can just use the
# easy [GH-nnnn] shorthand for quickly adding items to the "Unrelease" section
# while merging things between releases.

set -e

if [[ ! -f CHANGELOG.md ]]; then
  echo "ERROR: CHANGELOG.md not found in pwd."
  echo "Please run this from the root of the go-vcloud-director repository"
  exit 1
fi

if [[ `uname` == "Darwin" ]]; then
  echo "Using BSD sed"
  SED="sed -i.bak -E -e"
else
  echo "Using GNU sed"
  SED="sed -i.bak -r -e"
fi

GOVCD_URL="https:\/\/github.com\/vmware\/go-vcloud-director\/pull"

$SED "s/GH-([0-9]+)/\[#\1\]\($GOVCD_URL\/\1\)/g" -e 's/\[\[#(.+)([0-9])\)]$/(\[#\1\2))/g' CHANGELOG.md
if [ "$?" != "0" ] ; then exit 1 ; fi
rm CHANGELOG.md.bak
