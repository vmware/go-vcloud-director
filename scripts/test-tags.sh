#!/usr/bin/env bash

# This test checks that all the build tags defined in api_vcd_test.go
# can run individually

if [ ! -d govcd ]
then
    echo "./govcd directory missing"
    exit 1
fi
cd govcd

if [ ! -f api_vcd_test.go ]
then
    echo "file ./govcd/api_vcd_test.go not found"
    exit 1
fi

start=$(date +%s)
tags=$(head -n 1 api_vcd_test.go | sed -e 's/^.*build //;s/|| //g')

echo "=== RUN TagsTest"
for tag in $tags
do
    
    go test -tags $tag -timeout 0 -count=0 -check.vv > /dev/null

    if [ "$?" == "0" ]
    then
        echo "  --- PASS: TagsTest/$tag"
    else
        echo "  --- FAIL: TagsTest/$tag"
        failed="$failed $tag"
    fi
done

end=$(date +%s)
elapsed=$((end-start))
if [ -n "$failed" ]
then
    echo "--- FAIL: TagsTest - Tests for tags [$failed] have failed (${elapsed}s)"
    exit 1
fi
echo "--- PASS: TagsTest (${elapsed}s)"
exit 0
