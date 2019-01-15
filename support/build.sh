#!/usr/bin/env bash

echo "# Build $(date)"

# Print environment information, useful for troubleshooting
# NOTE: we're working inside a Docker container

go version

echo "## USER $USER"
echo "## HOME $HOME"
if [ -z "$HOME" ]
then
    echo "\$HOME is not set"
    exit 1
fi

if [ -z "$USER" ]
then
    echo "\$USER is not set"
    exit 1
fi

echo "## PWD $PWD"
echo "## GOROOT $GOROOT"
echo "## GOPATH $GOPATH"
echo "## OS $(uname -a)"
echo "## hostname $(hostname)"

echo "## GOVCD_CONFIG $GOVCD_CONFIG"
if [ -n "$GOVCD_CONFIG" ]
then
    if [ ! -f $GOVCD_CONFIG ]
    then
        echo "## $GOVCD_CONFIG not found"
        exit 1
    fi
else
    echo "## GOVCD_CONFIG not set"
    exit 1
fi

echo "## ls \$HOME"
ls -l $HOME

echo "## ls -l \$PWD"
ls -l 

# Creates a new $GOPATH, accessible to current user

#export GOPATH=$HOME/go
#destination=$HOME/go/src/github.com/vmware/go-vcloud-director
#mkdir -p $destination
#if [ ! -d $destination ]
#then
#    echo "# destination directory ($destination) not created"
#    exit 1
#fi
#
#echo "## NEW GOPATH $GOPATH"

# creates symbolic links for source code in the new GOPATH
#for item in Makefile scripts support govcd util types test-resources vendor
#do
#    ln -s $PWD/$item $destination/$item
#done

#cd $destination

echo "# Test $(date)"
make vet
make test
