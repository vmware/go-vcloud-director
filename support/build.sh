#!/usr/bin/env bash

echo "# Build $(date)"
go version

echo "## PWD $PWD"
echo "## GOROOT $GOROOT"
echo "## GOPATH $GOPATH"
echo "## OS: $(uname -a)"
ls -l 

echo "# Test $(date)"
# make build


