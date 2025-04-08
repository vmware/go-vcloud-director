#!/usr/bin/env bash

missing_license_files=$(find . -name "*.go" -type f -exec sh -c 'head -n 10 "$0" | grep -q "// © Broadcom. All Rights Reserved." || echo "License missing in: $0"' {} \;)

if [ -z "$missing_license_files" ]; then
  exit 0
else 
  echo ${missing_license_files}
  exit 1
fi
