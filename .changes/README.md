# Handling of CHANGELOG entries

## PR operations

* All PRs, instead of updating CHANGELOG.md directly, will create individual files in a directory .changes/$version 
* The files will be named `xxx-section_name.md`, where xxx is the PR number, and `section_name` is one of 
    * features
    * improvements
    * bug-fixes
    * deprecations
    * notes
    * removals

* The changes files must NOT contain the header

* You can update the file `.changes/sections` to add more headers
* To see the full change set for the current version (or an old one), use `./scripts/make-changelog.sh [version]`


## Post release initialization
   
After a release, the changelog will be initialized with the following template:
 
```
    ## $VERSION (Unreleased)

    Changes in progress for v$VERSION are available at [.changes/v$VERSION](https://github.com/vmware/go-vcloud-director/tree/main/.changes/v$VERSION) until the release.
```

Run `.changes/init.sh version` to get the needed text
