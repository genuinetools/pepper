<p align="left"><img src="logo/horizontal.png" alt="pepper" height="160px"></p>

[![Travis CI](https://img.shields.io/travis/genuinetools/pepper.svg?style=for-the-badge)](https://travis-ci.org/genuinetools/pepper)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/genuinetools/pepper)
[![Github All Releases](https://img.shields.io/github/downloads/genuinetools/pepper/total.svg?style=for-the-badge)](https://github.com/genuinetools/pepper/releases)

Named after Pepper Potts. Set all your GitHub repos master branches to be
protected.

You can set which orgs to include and use `--dry-run` to see the
changes before they are actually made. Your user is automatically added to the
repositories it will consider.

Also see [genuinetools/audit](https://github.com/genuinetools/audit) for checking what
collaborators, hooks, deploy keys, and protected branched you have added on
all your GitHub repositories.

 * [Installation](README.md#installation)
      * [Binaries](README.md#binaries)
      * [Via Go](README.md#via-go)
 * [Usage](README.md#usage)

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
pepper -  Tool to set all GitHub repo master branches to be protected.

Usage: pepper <command>

Flags:

  -d        enable debug logging (default: false)
  -dry-run  do not change branch settings just print the changes that would occur (default: false)
  -nouser   do not include your user (default: false)
  -orgs     organizations to include (default: [])
  -token    GitHub API token (or env var GITHUB_TOKEN)
  -url      GitHub Enterprise URL (default: <none>)

Commands:

  version  Show the version information.
```

```console
$ pepper --dry-run --token 12345 --orgs jessconf --orgs maintainerati
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
