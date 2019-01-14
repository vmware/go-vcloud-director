#!/usr/bin/env bash

echo "# Build $(date)"
go version

echo "## USER $USER"
echo "## HOME $HOME"
if [  -z "$HOME" ]
then
    echo "\$HOME is not set"
    export HOME=/home/worker
    echo "## (2) HOME $HOME"
fi

if [  -z "$USER" ]
then
    echo "\$USER is not set"
    export USER=worker
    echo "## (2) USER $USER"
fi

echo "## PWD $PWD"
echo "## GOROOT $GOROOT"
echo "## GOPATH $GOPATH"
echo "## OS $(uname -a)"
echo "## hostname $(hostname)"

echo "## GOVCD_CONFIG $GOVCD_CONFIG"
if [ -n "$GOVCD_CONFIG" ]
then
    if [ -f $GOVCD_CONFIG ]
    then
        ls -l $GOVCD_CONFIG
        grep -vi password $GOVCD_CONFIG
    else
        echo "## $GOVCD_CONFIG not found"
    fi
else
    echo "## GOVCD_CONFIG not set"
fi

echo "## ls \$HOME"
ls -l $HOME

echo "## ls \$GOPATH"
ls -l $GOPATH

echo "## ls -l \$PWD"
ls -l 

destination=$HOME/go/src/github.com/vmware/go-vcloud-director
mkdir -p $destination
if [ ! -d $destination ]
then
    echo "# destination directory ($destination) not created"
    exit 1
fi

export GOPATH=$HOME/go
for item in Makefile scripts support govcd util types test-resources vendor
do
    ln -s $PWD/$item $destination/$item
done

cd $destination

echo "# Test $(date)"
make

