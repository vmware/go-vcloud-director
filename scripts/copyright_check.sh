#!/bin/bash
# This script will find code files that don't have a copyright notice
# or the ones with an outdated copyright.
#
# The checks will fail if:
# a) the source file does not have a copyright header
# b) the source file has a copyright header from last year, but it was modified this year


# This check will find files with a copyright for any year
vmware_any_copyright='Copyright \d\d\d\d VMware'

this_year=$(date +%Y)
last_year=$((this_year-1))

# This check will find files with a copyright for the current year
vmware_latest_copyright="Copyright $this_year VMware"

# This check will find files with a copyright for last year
vmware_last_year_copyright="Copyright $last_year VMware"
exit_code=0

modified_files=$(git status -uno | grep "modified:" | awk '{print $2}')

function is_modified {
    fname=$1
    for fn in $modified_files
    do
        if [ "$fname" == "$fn" -o "$fname" == "./$fn" ]
        then
            echo yes
        fi
    done
}

for F in $(find . -name '*.go' | grep -v '/vendor/' )
do
    modified_not_committed_yet=$(is_modified $F)
    copyright_found=""
    for line_num in 1 2 3
    do
        # Looks for copyright in the Nth line of the file
        has_any_copyright=$(head -n $line_num $F | tail -n 1 | grep "$vmware_any_copyright" )
        has_latest_copyright=$(head -n $line_num $F | tail -n 1 | grep "$vmware_latest_copyright" )
        has_last_year_copyright=$(head -n $line_num $F | tail -n 1 | grep "$vmware_last_year_copyright" )

        if [ -n "$has_latest_copyright" ]
        then
            if [ "$1" == "-v" -o "$1" == "--verbose" ]
            then
                echo "$F: latest copyright found in line $line_num"
            fi
            copyright_found=$line_num
        elif [ -n "$has_last_year_copyright" ]
        then
            commit_date=$(git log -1 --format="%cd" $F)
            update_label=committed

            # The file is updated this year if the commit date contains the current year
            committed_this_year=$(echo "$commit_date" | grep -w $this_year)

            # The file is also updated this year if it is in the list of modified files
            # (not committed yet)
            if [ -n "$modified_not_committed_yet" ]
            then
                update_label=modified
            fi

            if [ -n "$modified_not_committed_yet" -o "$committed_this_year" ]
            then
                echo "$F $update_label this year, but has last year's copyright"
                exit_code=1
            fi
            copyright_found=$line_num
        elif [ -n "$has_any_copyright" ]
        then
            echo "$F: older copyright found in line $line_num"
            echo "$has_any_copyright"
            copyright_found=$line_num
        fi
    done
    if [ -z "$copyright_found" ]
    then
        echo "File $F has no valid copyright"
        exit_code=1
    fi
done
exit $exit_code
