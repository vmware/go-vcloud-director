# Contributing to go-vcloud-director

Welcome! We gladly accept contributions from the community. If you wish
to contribute code and you have not signed our contributor license
agreement (CLA), our bot will update the issue when you open a pull
request. For any questions about the CLA process, please refer to our
[FAQ](https://cla.vmware.com/faq).

## Community

vCloud Director Golang and Terraform contributors can be found here: 
https://vmwarecode.slack.com, vcd-terraform-dev#

## Logging Bugs

Anyone can log a bug using the GitHub 'New Issue' button.  Please use
a short title and give as much information as you can about what the
problem is, relevant software versions, and how to reproduce it.  If you
know of a fix or a workaround include that too.

## Code Contribution Flow

We use GitHub pull requests to incorporate code changes from external
contributors.  Typical contribution flow steps are:

- Fork the go-vcloud-director repo into a new repo on GitHub
- Clone the forked repo locally and set the original go-vcloud-director repo as the upstream repo
- Open an Issue in go-vcloud-director describing what you propose to do (unless the change is so trivial that an issue is not needed)
- Wait for discussion and possible direction hints in the issue thread
- Once you know  which steps to take in your intended contribution, make changes in a topic branch and commit (don't forget to add or modify tests too)
- Update Go modules files `go.mod` and `go.sum` if you're changing dependencies.
- Fetch changes from upstream and resolve any merge conflicts so that your topic branch is up-to-date
- Push all commits to the topic branch in your forked repo
- Submit a pull request to merge topic branch commits to upstream master

If this process sounds unfamiliar have a look at the
excellent [overview of collaboration via pull requests on
GitHub](https://help.github.com/categories/collaborating-with-issues-and-pull-requests) for more information. 

## Coding Style

Our standard for Golang contributions is to match the format of the [standard
Go package library](https://golang.org/pkg).  

- Run `go fmt` on all code with latest stable version of Go (`go fmt` results may vary between Go versions).  
- All public interfaces, functions, and structs must have complete, grammatically correct Godoc comments that explain their purpose and proper usage.
- Use self-explanatory names for all variables, functions, and interfaces.
- Add comments for non-obvious features of internal implementations but otherwise let the code explain itself.
- Include unit tests for new features and update tests for old ones. Refer to the [testing guide](TESTING.md) for more details.

Go is pretty readable so if you follow these rules most functions
will not need additional comments.

See **CODING_GUIDELINES.md** for more advice on how to write code for this project.

### Commit Message Format

We follow the conventions on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/).

Be sure to include any related GitHub
issue references in the commit message.  See [GFM
syntax](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown)
for referencing issues.

## Contribution Example

Here is a tutorial of adding a feature to fix the foo api using
GitHub account imahacker.  If you are an experienced git user feel free
to adapt it to your own work style.

### Sign the Contributor License Agreement (CLA)

VMware Apache-licensed projects require all contributors to sign a CLA. 
Visit https://cla.vmware.com and follow steps presented there. 

### Fork the Repo

Navigate to the [go-vcloud-director repo on
GitHub](https://github.com/vmware/go-vcloud-director) and use the 'Fork' button to
create a forked repository under your GitHub account.  This gives you a copy 
of the repo for pull requests back to go-vcloud-director. 

### Clone and Set Upstream Remote

Make a local clone of the forked repo and add the base go-vcloud-director
repo as the upstream remote repository.

The project uses Go modules so the path is up to you, but do not forget
to set `GO111MODULE=on` if you are in `GOPATH`

``` shell

cd $GOPATH/src/github.com/vmware
git clone https://github.com/imahacker/go-vcloud-director
cd go-vcloud-director
git remote add upstream https://github.com/vmware/go-vcloud-director.git
```

The last git command prepares your clone to pull changes from the
upstream repo and push them into the fork, which enables you to keep
the fork up to date. More on that shortly.

### Make Changes and Commit

Start a new topic branch from the current HEAD position on master and
commit your feature changes into that branch.  

``` shell
git checkout -b foo-api-fix-22 master
# (Make feature changes)
git commit -a --signoff
git push origin foo-api-fix-22
```

The --signoff puts your signature in the commit.  It's required by our CLA
bot. 

It is a git best practice to put work for each new feature in a separate
topic branch and use git checkout commands to jump between them.  This
makes it possible to have multiple active pull requests.  We can accept
pull requests from any branch, so it's up to you how to manage them.

### Stay in Sync with Upstream

From time to time you'll need to merge changes from the upstream
repo so your topic branch stays in sync with other checkins.  To
do so switch to your topic branch, pull from the upstream repo, and
push into the fork.  If there are conflicts you'll need to [merge
them now](https://stackoverflow.com/questions/161813/how-to-resolve-merge-conflicts-in-git).

``` shell
git checkout foo-api-fix-22
git fetch -a
git pull --rebase upstream master --tags
git push --force-with-lease origin foo-api-fix-22
```

The git pull and push options are important.  Here are some details if you 
need deeper understanding. 

- 'pull --rebase' eliminates unnecessary merges
by replaying your commit(s) into the log as if they happened
after the upstream changes.  Check out [What is a "merge
bubble"?](https://stackoverflow.com/questions/26239379/what-is-a-merge-bubble)
for why this is important.  
- --tags ensures that object tags are also pulled
- Depending on your git configuration push --force-with-lease is required to make git update your fork with commits from the upstream repo.


### Test Changes Locally

The last step before creating a Pull Request is to run the tests locally and
making sure they pass. Please see the [testing guide](TESTING.md) for more
details.

### Create a Pull Request

To contribute your feature, create a pull request by going to the [go-vcloud-director upstream repo on GitHub](https://github.com/vmware/go-vcloud-director) and pressing the 'New pull request' button. 

Select 'compare across forks' and select imahacker/go-vcloud-director as 'head fork'
and foo-api-fix-22 as the 'compare' branch.  Leave the base fork as 
vmware/go-vcloud-director and master. 

### Wait...

Your pull request will automatically build in [Travis
CI](https://travis-ci.org/vmware/go-vcloud-director/).  Have a look and correct
any failures.

Meanwhile a committer will look the request over and do one of three things: 

- accept it
- send back comments about things you need to fix
- or close the request without merging if we don't think it's a good addition.

### Updating Pull Requests with New Changes

If your pull request fails to pass Travis CI or needs changes based on
code review, you'll most likely want to squash the fixes into existing
commits.

If your pull request contains a single commit or your changes are related
to the most recent commit, you can simply amend the commit.

``` shell
git add .
git commit --amend
git push --force-with-lease origin foo-api-fix-22
```

If you need to squash changes into an earlier commit, you can use:

``` shell
git add .
git commit --fixup <commit>
git rebase -i --autosquash master
git push --force-with-lease origin foo-api-fix-22
```

Be sure to add a comment to the pull request indicating your new changes
are ready to review, as GitHub does not generate a notification when
you git push.

## Final Words

Thanks for helping us make the project better!
