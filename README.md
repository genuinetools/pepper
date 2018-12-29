<p align="left"><img src="logo/horizontal.png" alt="pepper" height="160px"></p>

[![Travis CI](https://img.shields.io/travis/genuinetools/pepper.svg?style=for-the-badge)](https://travis-ci.org/genuinetools/pepper)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/pepper)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/pepper/total.svg?style=for-the-badge)](https://github.com/genuinetools/pepper/releases)

Named after Pepper Potts.  A tool for performing actions on GitHub repos or a single repo. 

Actions include:

- [Protecting all master branches](#protect)
- [Adding a collaborator](#collaborators)
- [Setting merge settings](#merge)

You can set which orgs to include and use `--dry-run` to see the
changes before they are actually made. Your user is automatically added to the
repositories it will consider.

Also see [genuinetools/audit](https://github.com/genuinetools/audit) for checking what
collaborators, hooks, deploy keys, and protected branched you have added on
all your GitHub repositories.

**Table of Contents**

<!-- toc -->

<!-- tocstop -->

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/genuinetools/pepper/releases).

#### Via Go

```console
$ go get github.com/genuinetools/pepper
```

## Usage

```console
$ pepper -h
pepper -  A tool for performing actions on GitHub repos or a single repo.

Usage: pepper <command>

Flags:

  -d, --debug  enable debug logging (default: false)
  --dry-run    do not change settings just print the changes that would occur (default: false)
  --nouser     do not include your user (default: false)
  --orgs       organizations to include (default: [])
  -r, --repo   specific repo (e.g. 'genuinetools/img') (default: <none>)
  -t, --token  GitHub API token (or env var GITHUB_TOKEN) (default: <none>)
  -u, --url    GitHub Enterprise URL (default: <none>)

Commands:

  audit          Audit collaborators, branches, hooks, deploy keys etc.
  collaborators  Add a collaborator to all the repositories.
  merge          Update all merge settings to allow specific types only.
  protect        Protect the master branch.
  release        Update the release body information.
  version        Show the version information.
```

### Protect

Protect all master branches.

```console
$ pepper protect --dry-run --token 12345 --orgs jessconf --orgs maintainerati
[OK] jessconf/jessconf:master is already protected
[OK] genuinetools/.vim:master is already protected
[OK] genuinetools/anonymail:master is already protected
[OK] genuinetools/apk-file:master is already protected
[UPDATE] genuinetools/certok:master will be changed to protected
...
[OK] genuinetools/weather:master is already protected
[OK] genuinetools/ykpiv:master is already protected
[OK] maintainerati/wontfix-cabal-site:master is already protected
```

### Audit

Audit collaborators, branches, hooks, deploy keys etc.

```console
$ pepper audit -r genuinetools/img
genuinetools/img -> 
        Collaborators (7):
                Admin (2):
                        jessfraz
                        j3ssb0t
                Write (5):
                        bketelsen
                        gabrtv
                        bacongobbler
                        sajayantony
                        AkihiroSuda
                Read (0):

        Hooks (4):
                travis - active:true (https://api.github.com/repos/genuinetools/img/hooks/22351842)
                jenkins - active:true (https://api.github.com/repos/genuinetools/img/hooks/22351967)
                web - active:true (https://api.github.com/repos/genuinetools/img/hooks/38652766)
                web - active:true (https://api.github.com/repos/genuinetools/img/hooks/38654028)
        Protected Branches (1): master
        Merge Methods: squash
```


### Collaborators

Add a collaborator to all the repositories.

```console
$ pepper collaborators -h
Usage: pepper collaborators [OPTIONS] COLLABORATOR

Add a collaborator to all the repositories.

Flags:

  --admin      Team members can pull, push and administer this repository (default: false)
  -d, --debug  enable debug logging (default: false)
  --dry-run    do not change settings just print the changes that would occur (default: false)
  --nouser     do not include your user (default: false)
  --orgs       organizations to include (default: [])
  --pull       Team members can pull, but not push to or administer this repository (default: false)
  --push       Team members can pull and push, but not administer this repository (default: false)
  -r, --repo   specific repo (e.g. 'genuinetools/img') (default: <none>)
  -t, --token  GitHub API token (or env var GITHUB_TOKEN) (default: <none>)
  -u, --url    GitHub Enterprise URL (default: <none>)
```

```console
$ pepper collaborators --dry-run --admin j3ssb0t
...t 
[UPDATE] jessfraz/blog will have j3ssb0t added as a collaborator (admin)
[UPDATE] jessfraz/bolt will have j3ssb0t added as a collaborator (admin)
[UPDATE] jessfraz/boulder will have j3ssb0t added as a collaborator (admin)
[UPDATE] jessfraz/buildkit will have j3ssb0t added as a collaborator (admin)
[UPDATE] jessfraz/cadvisor will have j3ssb0t added as a collaborator (admin)
...
```

### Merge

Update all merge settings to allow specific types only.

```console
$ pepper merge -h
Usage: pepper merge [OPTIONS]

Update all merge settings to allow specific types only.

Flags:

  --commits    Allow merge commits, add all commits from the head branch to the base branch with a merge commit (default: false)
  -d, --debug  enable debug logging (default: false)
  --dry-run    do not change settings just print the changes that would occur (default: false)
  --nouser     do not include your user (default: false)
  --orgs       organizations to include (default: [])
  -r, --repo   specific repo (e.g. 'genuinetools/img') (default: <none>)
  --rebase     Allow rebase merging, add all commits from the head branch onto the base branch individually (default: false)
  --squash     Allow squash merging, combine all commits from the head branch into a single commit in the base branch (default: false)
  -t, --token  GitHub API token (or env var GITHUB_TOKEN) (default: <none>)
  -u, --url    GitHub Enterprise URL (default: <none>)
```

```console
$ pepper merge --dry-run --squash
UPDATE] jessfraz/.vim will be changed to squash
[OK] jessfraz/1up is already set to squash
[UPDATE] jessfraz/acs-engine will be changed to squash
[UPDATE] jessfraz/acs-ignite-demos will be changed to squash
[UPDATE] jessfraz/apparmor-docs will be changed to squash
[UPDATE] jessfraz/baselayout will be changed to squash
...
```

### Update Release

Update the release body according to the template.

```console
$ pepper release --repo genuinetools/img
Updated release v0.5.0/v0.5.0 for repo: genuinetools/img
```
