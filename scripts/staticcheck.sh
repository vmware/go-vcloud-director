#!/usr/bin/env bash
scripts_dir=$(dirname $0)
cd $scripts_dir
scripts_dir=$PWD
cd - > /dev/null

sc_exit_code=0

if [ ! -d ./govcd ]
then
    echo "source directory ./govcd not found"
    exit 1
fi

function exists_in_path {
    what=$1
    for dir in $(echo $PATH | tr ':' ' ')
    do
        wanted=$dir/$what
        if [ -x $wanted ]
        then
            echo $wanted
            return
        fi
    done
}

function get_check_static {
    static_check=$(exists_in_path staticcheck)
    if [  -z "$staticcheck" -a -n "$GITHUB_ACTIONS" ]
    then
        # Variables found in staticcheck-config.sh
        # STATICCHECK_URL
        # STATICCHECK_VERSION
        # STATICCHECK_FILE
        if [ -f $scripts_dir/staticcheck-config.sh ]
        then
            source $scripts_dir/staticcheck-config.sh
        else
            echo "File $scripts_dir/staticcheck-config.sh not found - Skipping check"
            exit 0
        fi
        download_name=$STATICCHECK_URL/$STATICCHECK_VERSION/$STATICCHECK_FILE
        wget=$(exists_in_path wget)
        if [ -z "$wget" ]
        then
            echo "'wget' executable not found - Skipping check"
            exit 0
        fi
        $wget $download_name
        if [ -n "$STATICCHECK_FILE" ]
        then
            tar -xzf $STATICCHECK_FILE
            executable=$PWD/staticcheck/staticcheck
            if [ ! -f $executable ]
            then
                echo "Extracted executable not available - Skipping check"
            fi
            chmod +x $executable
            static_check=$executable
        fi
    fi
    if [ -n "$static_check" ]
    then
        echo "## Found $static_check"
        echo -n "## "
        $static_check -version
    else
        echo "*** staticcheck executable not found - Check skipped"
        exit 0
    fi
}

function check_static {
    dir=$1
    if [ -n "$static_check" ]
    then
        cd $dir
        echo "## Checking $dir"
        $static_check -tags ALL .
        exit_code=$?
        if [ "$exit_code" != "0" ]
        then
            sc_exit_code=$exit_code
        fi
        cd - > /dev/null
    fi
    echo ""
}

get_check_static
echo ""

check_static govcd
check_static types/v56
check_static util
echo "Exit code: $sc_exit_code"
exit $sc_exit_code

