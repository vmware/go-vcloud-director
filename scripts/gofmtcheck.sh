#!/usr/bin/env bash

#
# © Broadcom. All Rights Reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
# SPDX-License-Identifier: Apache-2.0
#

# Check gofmt
echo "==> Checking that code complies with gofmt requirements..."
gofmt_files=$(gofmt -l `find . -name '*.go' | grep -v vendor`)
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "You can use the command: \`make fmt\` to reformat code."
    exit 1
fi

exit 0
