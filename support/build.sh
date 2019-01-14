#!/usr/bin/env bash

echo "# Build $(date)"
go version


echo "# Test $(date)"
make build

